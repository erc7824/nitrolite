// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Vm} from "lib/forge-std/src/Vm.sol";
import {Test} from "lib/forge-std/src/Test.sol";

import {ChannelsHubTest_Base} from "./ChannelsHub_Base.t.sol";
import {MockERC20} from "./mocks/MockERC20.sol";
import {TestUtils} from "./TestUtils.sol";

import {ChannelsHub} from "../src/ChannelsHub.sol";
import {Utils} from "../src/Utils.sol";
import {CrossChainState, Definition, StateIntent, State, ChannelStatus, EscrowStatus} from "../src/interfaces/Types.sol";

contract ChannelsHubTest_CrossChain_Lifecycle is ChannelsHubTest_Base {
    bytes32 bobChannelId;
    Definition bobDef;

    function setUp() public override {
        super.setUp();

        bobDef = Definition({
            challengeDuration: CHALLENGE_DURATION,
            user: bob,
            node: node,
            nonce: NONCE,
            metadata: bytes32(0)
        });

        bobChannelId = Utils.getChannelId(bobDef);
    }

    function test_happyPath_homeChain() public {
        Definition memory def = Definition({
            challengeDuration: CHALLENGE_DURATION,
            user: alice,
            node: node,
            nonce: NONCE,
            metadata: bytes32(0)
        });

        bytes32 channelId = Utils.getChannelId(def);

        // Check VOID status before channel creation
        (ChannelStatus status,,,,) = cHub.getChannelData(channelId);
        assertEq(uint8(status), uint8(ChannelStatus.VOID), "Channel should be VOID before creation");

        // Verify user balance before channel creation
        assertEq(token.balanceOf(alice), INITIAL_BALANCE, "User balance before channel creation");

        // Initial state: alice deposits 1000
        // Expected: user allocation = 1000, user net flow = 1000, node allocation = 0, node net flow = 0
        CrossChainState memory state = CrossChainState({
            version: 0,
            intent: StateIntent.CREATE,
            homeState: State({
                chainId: uint64(block.chainid),
                token: address(token),
                userAllocation: DEPOSIT_AMOUNT,
                userNetFlow: int256(DEPOSIT_AMOUNT),
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            nonHomeState: State({
                chainId: 0,
                token: address(0),
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });

        state = signStateWithBothParties(state, channelId, alicePK);

        vm.prank(alice);
        cHub.createChannel(def, state);

        // Verify user balance after channel creation (deposited 1000)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1000, "User balance after channel creation");

        // transfer 42 (allocation decreases by 42, node net flow decreases by 42)
        state = nextState(state, StateIntent.OPERATE, [uint256(958), uint256(0)], [int256(1000), int256(-42)]);
        state = signStateWithBothParties(state, channelId, alicePK);

        // deposit from another chain
        state = nextCrossChainState(
            state,
            StateIntent.INITIATE_ESCROW_DEPOSIT,
            // user amounts stay the same, node amounts increase by 500
            [uint256(958), uint256(500)], [int256(1000), int256(458)],
            42, address(42), // chainId 42, token address 42 for simplicity
            // user deposit amount appear in allocation and net flow on non-home side
            [uint256(500), uint256(0)], [int256(500), int256(0)]
        );
        state = signStateWithBothParties(state, channelId, alicePK);

        // on chainId 42:
        // channelsHub.initiateEscrowDeposit(channelId, state)
        // NOTE: see a `test_depositEscrow_nonHomeChain` test for that

        // initiate escrow deposit on home chain
        // Expected: user allocation = 958, user net flow = 1000, node allocation = 0, node net flow = -42
        vm.prank(alice);
        cHub.initiateEscrowDeposit(def, state);
        verifyChannelState(channelId, 958, 1000, 500, 458, "after cross chain deposit");

        // finalize escrow deposit
        state = nextCrossChainState(
            state,
            StateIntent.FINALIZE_ESCROW_DEPOSIT,
            // user allocation amount increases by cross-chain deposit, node allocation goes to 0
            [uint256(1458), uint256(0)], [int256(1000), int256(458)],
            42, address(42),
            // user deposit amount is zeroed, and withdrawn (unlocked) by node via net flow on non-home side
            [uint256(0), uint256(0)], [int256(500), int256(-500)]
        );
        state = signStateWithBothParties(state, channelId, alicePK);

        // receive 24 (allocation increases by 24, node net flow increases by 24)
        state = nextState(state, StateIntent.OPERATE, [uint256(1482), uint256(0)], [int256(1000), int256(482)]);
        state = signStateWithBothParties(state, channelId, alicePK);

        // send 12 (allocation decreases by 12, node net flow decreases by 12)
        state = nextState(state, StateIntent.OPERATE, [uint256(1470), uint256(0)], [int256(1000), int256(470)]);
        state = signStateWithBothParties(state, channelId, alicePK);

        // withdraw 250 on home chain
        // Expected: user allocation = 1220, user net flow = 750, node allocation = 0, node net flow = 470
        state = nextState(state, StateIntent.WITHDRAW, [uint256(1220), uint256(0)], [int256(750), int256(470)]);
        state = signStateWithBothParties(state, channelId, alicePK);

        vm.prank(alice);
        cHub.withdrawFromChannel(channelId, state);
        verifyChannelState(channelId, 1220, 750, 0, 470, "after withdrawal");

        // Verify user balance after withdrawal (withdrew 250)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 750, "User balance after withdrawal");

        // send 2 (allocation decreases by 2, node net flow decreases by 2)
        state = nextState(state, StateIntent.OPERATE, [uint256(1218), uint256(0)], [int256(750), int256(468)]);
        state = signStateWithBothParties(state, channelId, alicePK);

        // receive 3 (allocation increases by 3, node net flow increases by 3)
        state = nextState(state, StateIntent.OPERATE, [uint256(1221), uint256(0)], [int256(750), int256(471)]);
        state = signStateWithBothParties(state, channelId, alicePK);

        // send 4 (allocation decreases by 4, node net flow decreases by 4)
        state = nextState(state, StateIntent.OPERATE, [uint256(1217), uint256(0)], [int256(750), int256(467)]);
        state = signStateWithBothParties(state, channelId, alicePK);

        // withdrawal to another chain
        state = nextCrossChainState(
            state,
            StateIntent.INITIATE_ESCROW_WITHDRAWAL,
            // home chain stays the same
            [uint256(1217), uint256(0)], [int256(750), int256(467)],
            42, address(42), // chainId 42, token address 42 for simplicity
            // node deposits withdrawal amount
            [uint256(0), uint256(750)], [int256(0), int256(750)]
        );
        state.nodeSig = TestUtils.signStateEIP191(vm, channelId, state, nodePK);

        // on chainId 42:
        // channelsHub.initiateEscrowWithdrawal(channelId, state)
        // NOTE: see a `test_withdrawalEscrow_nonHomeChain` test for that

        // finalize escrow withdrawal on another chain
        state = nextCrossChainState(
            state,
            StateIntent.FINALIZE_ESCROW_WITHDRAWAL,
            // user allocation decreases by withdrawal amount, node allocation stays 0, node net flow decreases by withdrawal amount
            [uint256(467), uint256(0)], [int256(750), int256(-283)],
            42, address(42), // chainId 42, token address 42 for simplicity
            // user withdraws the amount (negative net flow)
            [uint256(0), uint256(0)], [int256(-750), int256(750)]
        );
        state = signStateWithBothParties(state, channelId, alicePK);

        // receive 10 (allocation increases by 10, node net flow increases by 10)
        state = nextState(state, StateIntent.OPERATE, [uint256(477), uint256(0)], [int256(750), int256(-273)]);
        state = signStateWithBothParties(state, channelId, alicePK);

        // checkpoint on home chain
        vm.prank(alice);
        cHub.checkpointChannel(channelId, state, new CrossChainState[](0));
        verifyChannelState(channelId, 477, 750, 0, -273, "after checkpoint");

        // Verify user balance hasn't changed
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 750, "User balance after checkpoint");

        // send 9 (allocation decreases by 9, node net flow decreases by 9)
        state = nextState(state, StateIntent.OPERATE, [uint256(468), uint256(0)], [int256(750), int256(-282)]);
        state = signStateWithBothParties(state, channelId, alicePK);

        // receive 8 (allocation increases by 8, node net flow increases by 8)
        state = nextState(state, StateIntent.OPERATE, [uint256(476), uint256(0)], [int256(750), int256(-274)]);
        state = signStateWithBothParties(state, channelId, alicePK);

        // send 7 (allocation decreases by 7, node net flow decreases by 7)
        state = nextState(state, StateIntent.OPERATE, [uint256(469), uint256(0)], [int256(750), int256(-281)]);
        state = signStateWithBothParties(state, channelId, alicePK);

        // migrate channel
        state = nextCrossChainState(
            state,
            StateIntent.INITIATE_MIGRATION,
            // home chain stays the same
            [uint256(469), uint256(0)], [int256(750), int256(-281)],
            42, address(42), // chainId 42, token address 42 for simplicity
            // node deposits full user allocation amount
            [uint256(0), uint256(469)], [int256(0), int256(469)]
        );
        state.nodeSig = TestUtils.signStateEIP191(vm, channelId, state, nodePK);

        // on chainId 42:
        // channelsHub.initiateMigration(channelId, state)
        // NOTE: see a `test_migration_nonHomeChain` test for that

        // finalize migration on home chain
        state = nextCrossChainState(
            state,
            StateIntent.FINALIZE_MIGRATION,
            // channel closes on home chain, allocations go to 0, net flows balance out
            [uint256(0), uint256(0)], [int256(750), int256(-750)],
            42, address(42), // chainId 42, token address 42 for simplicity
            // user receives allocation on new home chain
            [uint256(469), uint256(0)], [int256(469), int256(-469)]
        );
        // home state and non-home state are swapped
        State memory temp = state.homeState;
        state.homeState = state.nonHomeState;
        state.nonHomeState = temp;
        state = signStateWithBothParties(state, channelId, alicePK);

        // vm.prank(node);
        // cHub.finalizeMigration(channelId, state);

        // // Verify channel is migrated out
        // verifyChannelState(channelId, 0, 750, 0, -750, "after migration");

        // // Verify user balance hasn't changed (migration doesn't move funds on home chain)
        // assertEq(token.balanceOf(alice), INITIAL_BALANCE - DEPOSIT_AMOUNT, "User balance after migration");

        // // Check MIGRATED_OUT status after channel was migrated
        // (ChannelStatus finalStatus,,,,) = cHub.getChannelData(channelId);
        // assertEq(uint8(finalStatus), uint8(ChannelStatus.MIGRATED_OUT), "Channel should be MIGRATED_OUT after migration");
    }

    function test_depositEscrow_nonHomeChain() public {
        // Check VOID status
        (ChannelStatus status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.VOID), "Channel should be VOID on non-home chain");

        // Verify user balance before deposit
        assertEq(token.balanceOf(bob), INITIAL_BALANCE, "User balance before escrow deposit");

        // state from the "happyPath" test, but with home and nonHome states swapped
        CrossChainState memory state = CrossChainState({
            version: 42,
            intent: StateIntent.INITIATE_ESCROW_DEPOSIT,
            homeState: State({
                chainId: 42,
                token: address(42),
                userAllocation: 958,
                userNetFlow: 1000,
                nodeAllocation: 500,
                nodeNetFlow: 458
            }),
            nonHomeState: State({
                chainId: uint64(block.chainid),
                token: address(token),
                userAllocation: 500,
                userNetFlow: 500,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });
        state = signStateWithBothParties(state, bobChannelId, bobPK);

        bytes32 escrowId = Utils.getEscrowId(bobChannelId, state);

        // verify no escrow struct exists yet
        (EscrowStatus escrowStatus,,,,) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(escrowStatus), uint8(EscrowStatus.VOID), "Escrow should be VOID");

        vm.prank(bob);
        cHub.initiateEscrowDeposit(bobDef, state);

        // Verify user balance after deposit (deposited 500)
        assertEq(token.balanceOf(bob), INITIAL_BALANCE - 500, "User balance after escrow deposit");

        // Verify escrow struct is updated on ChannelsHub
        (EscrowStatus finalEscrowStatus, uint64 unlockAt, uint64 challengeExpiresAt, uint256 lockedAmount, CrossChainState memory initState) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(finalEscrowStatus), uint8(EscrowStatus.INITIALIZED), "Escrow should be INITIALIZED");
        uint64 expectedUnlockAt = uint64(block.timestamp + cHub.ESCROW_DEPOSIT_UNLOCK_DELAY());
        assertEq(unlockAt, expectedUnlockAt, "Escrow unlockAt is incorrect");
        assertEq(challengeExpiresAt, 0, "Escrow challengeExpiresAt should be zero");
        assertEq(lockedAmount, 500, "Escrow locked amount is incorrect");
        assertEq(initState.version, state.version, "Escrow initState version is incorrect");

        // ====== finalize escrow deposit ======
        // this is an explicit action
        // escrow deposit locked funds should also be unlocked after `unlockAt` time passes alongside any other on-chain call
        vm.warp(block.timestamp + cHub.ESCROW_DEPOSIT_UNLOCK_DELAY() + 1);

        uint256 nodeBalanceBefore = cHub.getVaultBalance(node, address(token));

        // state from the "happyPath" test, but with home and nonHome states swapped
        state = CrossChainState({
            version: 43,
            intent: StateIntent.FINALIZE_ESCROW_DEPOSIT,
            homeState: State({
                chainId: 42,
                token: address(42),
                userAllocation: 1458,
                userNetFlow: 1000,
                nodeAllocation: 0,
                nodeNetFlow: 458
            }),
            nonHomeState: State({
                chainId: uint64(block.chainid),
                token: address(token),
                userAllocation: 0,
                userNetFlow: 500,
                nodeAllocation: 0,
                nodeNetFlow: -500
            }),
            userSig: "",
            nodeSig: ""
        });
        state = signStateWithBothParties(state, bobChannelId, bobPK);

        vm.prank(node);
        cHub.finalizeEscrowDeposit(escrowId, state);

        // Verify user balance after deposit finalized has NOT changed
        assertEq(token.balanceOf(bob), INITIAL_BALANCE - 500, "User balance after escrow deposit finalized");

        uint256 nodeBalanceAfter = cHub.getVaultBalance(node, address(token));
        assertEq(nodeBalanceAfter, nodeBalanceBefore + 500, "Node balance after escrow deposit finalized");

        // Verify escrow struct is updated on ChannelsHub
        (finalEscrowStatus, unlockAt, challengeExpiresAt, lockedAmount, initState) = cHub.getEscrowDepositData(escrowId);
        assertEq(uint8(finalEscrowStatus), uint8(EscrowStatus.FINALIZED), "Escrow should be FINALIZED");
        assertEq(unlockAt, expectedUnlockAt, "Escrow unlockAt should remain unchanged");
        assertEq(challengeExpiresAt, 0, "Escrow challengeExpiresAt should be zero");
        assertEq(lockedAmount, 0, "Escrow locked amount should have been zeroed");
        assertEq(initState.version, 42, "Escrow initState version should remain unchanged");
    }

    function test_withdrawalEscrow_nonHomeChain() public {
        // Check VOID status
        (ChannelStatus status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.VOID), "Channel should be VOID on non-home chain");

        uint256 nodeBalanceBefore = cHub.getVaultBalance(node, address(token));

        // state from the "happyPath" test, but with home and nonHome states swapped
        CrossChainState memory state = CrossChainState({
            version: 42,
            intent: StateIntent.INITIATE_ESCROW_WITHDRAWAL,
            homeState: State({
                chainId: 42,
                token: address(42),
                userAllocation: 1217,
                userNetFlow: 750,
                nodeAllocation: 0,
                nodeNetFlow: 467
            }),
            nonHomeState: State({
                chainId: uint64(block.chainid),
                token: address(token),
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 750,
                nodeNetFlow: 750
            }),
            userSig: "",
            nodeSig: ""
        });
        state.nodeSig = TestUtils.signStateEIP191(vm, bobChannelId, state, nodePK);

        bytes32 escrowId = Utils.getEscrowId(bobChannelId, state);

        // verify no escrow struct exists yet
        (EscrowStatus escrowStatus,,,) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(escrowStatus), uint8(EscrowStatus.VOID), "Escrow should be VOID");

        vm.prank(bob);
        cHub.initiateEscrowWithdrawal(bobDef, state);

        // Verify user node's after deposit (deposited 500)
        uint256 nodeBalanceAfter = cHub.getVaultBalance(node, address(token));
        assertEq(nodeBalanceAfter, nodeBalanceBefore - 750, "Node balance after escrow withdrawal");

        // Verify escrow struct is updated on ChannelsHub: escrow data exists, `locked` equals to withdrawalAmount
        (EscrowStatus finalEscrowStatus, uint64 challengeExpireAt, uint256 lockedAmount, CrossChainState memory initState) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(finalEscrowStatus), uint8(EscrowStatus.INITIALIZED), "Escrow should be INITIALIZED");
        assertEq(challengeExpireAt, 0, "Escrow challengeExpireAt should be zero");
        assertEq(lockedAmount, 750, "Escrow locked amount is incorrect");
        assertEq(initState.version, state.version, "Escrow initState version is incorrect");

        uint256 bobBalanceBefore = token.balanceOf(bob);

        // finalize escrow withdrawal on another chain
        state = nextCrossChainState(
            state,
            StateIntent.FINALIZE_ESCROW_WITHDRAWAL,
            [uint256(467), uint256(0)], [int256(750), int256(-283)],
            uint64(block.chainid), address(token),
            [uint256(0), uint256(0)], [int256(-750), int256(750)]
        );
        state = signStateWithBothParties(state, bobChannelId, bobPK);

        vm.prank(node);
        cHub.finalizeEscrowWithdrawal(escrowId, state);

        // Verify user balance after withdrawal (withdrew 750)
        uint256 bobBalanceAfter = token.balanceOf(bob);
        assertEq(bobBalanceAfter, bobBalanceBefore + 750, "User balance after escrow withdrawal");

        // Verify escrow struct is updated on ChannelsHub: escrow data exists, has status "finalized", `locked` equals to 0
        (finalEscrowStatus, challengeExpireAt, lockedAmount, initState) = cHub.getEscrowWithdrawalData(escrowId);
        assertEq(uint8(finalEscrowStatus), uint8(EscrowStatus.FINALIZED), "Escrow should be FINALIZED");
        assertEq(challengeExpireAt, 0, "Escrow challengeExpireAt should be zero");
        assertEq(lockedAmount, 0, "Escrow locked amount is incorrect");
        assertEq(initState.version, 42, "Escrow initState  should not have changed");
    }

    function test_migration_nonHomeChain() public {
        vm.skip(true);

        // Check VOID status
        (ChannelStatus status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.VOID), "Channel should be VOID on non-home chain");

        uint256 nodeBalanceBefore = cHub.getVaultBalance(node, address(token));
        uint256 userBalanceBefore = cHub.getVaultBalance(bob, address(token));

        // state from the "happyPath" test, but with home and nonHome states swapped
        CrossChainState memory state = CrossChainState({
            version: 42,
            intent: StateIntent.INITIATE_ESCROW_DEPOSIT,
            homeState: State({
                chainId: 42,
                token: address(42),
                userAllocation: 469,
                userNetFlow: 750,
                nodeAllocation: 0,
                nodeNetFlow: -281
            }),
            nonHomeState: State({
                chainId: uint64(block.chainid),
                token: address(token),
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 469,
                nodeNetFlow: 469
            }),
            userSig: "",
            nodeSig: ""
        });
        state.nodeSig = TestUtils.signStateEIP191(vm, bobChannelId, state, nodePK);

        vm.prank(bob);
        cHub.initiateMigration(bobDef, state);
        // TODO: in ChannelEngine it should be checked that no channel exists with such channelId, and that nonHomeState only includes the same `userAllocation` as in the homeState

        // Verify user node's after migration (should have deposited 469)
        uint256 nodeBalanceAfter = cHub.getVaultBalance(node, address(token));
        assertEq(nodeBalanceAfter, nodeBalanceBefore - 750, "Node balance after escrow withdrawal");

        // user balance should not have changed
        uint256 userBalanceAfter = token.balanceOf(bob);
        assertEq(userBalanceAfter, userBalanceBefore, "User balance after escrow withdrawal");

        // Check MIGRATING_IN status
        (status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.MIGRATING_IN), "Channel should be MIGRATING_IN after migration");

        // sign finalize migration state by swapping homeState and nonHomeState, and swapping allocations
        state = CrossChainState({
            version: 42,
            intent: StateIntent.INITIATE_ESCROW_DEPOSIT,
            nonHomeState: State({
                chainId: 42,
                token: address(42),
                userAllocation: 0,
                userNetFlow: 750,
                nodeAllocation: 469,
                nodeNetFlow: -281
            }),
            homeState: State({
                chainId: uint64(block.chainid),
                token: address(token),
                userAllocation: 469,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: 469
            }),
            userSig: "",
            nodeSig: ""
        });
        state = signStateWithBothParties(state, bobChannelId, bobPK);

        // perform some operations to verify channel is operating as normal
        // send 9 (allocation decreases by 9, node net flow decreases by 9)
        state = nextState(state, StateIntent.OPERATE, [uint256(460), uint256(0)], [int256(0), int256(460)]);
        state = signStateWithBothParties(state, bobChannelId, bobPK);

        // receive 8 (allocation increases by 8, node net flow increases by 8)
        state = nextState(state, StateIntent.OPERATE, [uint256(468), uint256(0)], [int256(0), int256(468)]);
        state = signStateWithBothParties(state, bobChannelId, bobPK);

        // send 7 (allocation decreases by 7, node net flow decreases by 7)
        state = nextState(state, StateIntent.OPERATE, [uint256(461), uint256(0)], [int256(0), int256(461)]);
        state = signStateWithBothParties(state, bobChannelId, bobPK);

        // withdraw 400 on home chain
        // Expected: user allocation = 61, user net flow = -400, node allocation = 0, node net flow = 461
        state = nextState(state, StateIntent.WITHDRAW, [uint256(61), uint256(0)], [int256(-400), int256(461)]);
        state = signStateWithBothParties(state, bobChannelId, bobPK);

        vm.prank(bob);
        cHub.withdrawFromChannel(bobChannelId, state);
        verifyChannelState(bobChannelId, 61, -400, 0, 461, "after withdrawal");

        // Verify user balance after withdrawal (withdrew 400)
        assertEq(token.balanceOf(bob), userBalanceBefore + 400, "User balance after withdrawal");

        // Verify channel is still operating after migration and withdrawal
        (status,,,,) = cHub.getChannelData(bobChannelId);
        assertEq(uint8(status), uint8(ChannelStatus.OPERATING), "Channel should be OPERATING after withdrawal");
    }
}
