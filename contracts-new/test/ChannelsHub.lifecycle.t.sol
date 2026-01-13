// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Vm} from "lib/forge-std/src/Vm.sol";
import {Test} from "lib/forge-std/src/Test.sol";

import {MockERC20} from "./mocks/MockERC20.sol";
import {TestUtils} from "./TestUtils.sol";

import {ChannelsHub} from "../src/ChannelsHub.sol";
import {CrossChainState, Definition, StateIntent, State, ChannelStatus} from "../src/interfaces/Types.sol";

contract ChannelsHubLifecycleTest is Test {
    ChannelsHub public cHub;
    MockERC20 public token;

    uint256 constant nodePK = 1;
    uint256 constant alicePK = 2;

    address node;
    address alice;

    uint32 constant CHALLENGE_DURATION = 86400; // 1 day
    uint64 constant NONCE = 1;
    uint256 constant DEPOSIT_AMOUNT = 1000;
    uint256 constant INITIAL_BALANCE = 10000;

    function setUp() public virtual {
        // Deploy contracts
        cHub = new ChannelsHub();
        token = new MockERC20("Test Token", "TST", 18);

        node = vm.addr(nodePK);
        alice = vm.addr(alicePK);

        token.mint(node, INITIAL_BALANCE);
        token.mint(alice, INITIAL_BALANCE);

        vm.startPrank(node);
        token.approve(address(cHub), INITIAL_BALANCE);
        cHub.depositToVault(node, address(token), INITIAL_BALANCE);
        vm.stopPrank();

        vm.startPrank(alice);
        token.approve(address(cHub), INITIAL_BALANCE);
        vm.stopPrank();
    }

    function nextState(
        CrossChainState memory state,
        StateIntent intent,
        uint256[2] memory allocations,
        int256[2] memory netFlows
    ) internal pure returns (CrossChainState memory) {
        return CrossChainState({
            version: state.version + 1,
            intent: intent,
            homeState: State({
                chainId: state.homeState.chainId,
                token: state.homeState.token,
                userAllocation: allocations[0],
                userNetFlow: netFlows[0],
                nodeAllocation: allocations[1],
                nodeNetFlow: netFlows[1]
            }),
            nonHomeState: State({
                chainId: 0,
                token: address(0),
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            participantSig: "",
            nodeSig: ""
        });
    }

    function signStateWithBothParties(CrossChainState memory state, bytes32 channelId) internal pure returns (CrossChainState memory) {
        state.participantSig = TestUtils.signStateEIP191(vm, channelId, state, alicePK);
        state.nodeSig = TestUtils.signStateEIP191(vm, channelId, state, nodePK);
        return state;
    }

    function verifyChannelState(
        bytes32 channelId,
        uint256 expectedUserAllocation,
        int256 expectedUserNetFlow,
        uint256 expectedNodeAllocation,
        int256 expectedNodeNetFlow,
        string memory description
    ) internal view {
        (,, CrossChainState memory latestState,,) = cHub.getChannelData(channelId);
        assertEq(latestState.homeState.userAllocation, expectedUserAllocation, string.concat("User allocation ", description));
        assertEq(latestState.homeState.userNetFlow, expectedUserNetFlow, string.concat("User net flow ", description));
        assertEq(latestState.homeState.nodeAllocation, expectedNodeAllocation, string.concat("Node allocation ", description));
        assertEq(latestState.homeState.nodeNetFlow, expectedNodeNetFlow, string.concat("Node net flow ", description));

        uint256 nodeBalance = cHub.getVaultBalance(node, address(token));
        uint256 expectedNodeBalance = expectedNodeNetFlow < 0
            ? INITIAL_BALANCE + uint256(-expectedNodeNetFlow)
            : INITIAL_BALANCE - uint256(expectedNodeNetFlow);
        assertEq(nodeBalance, expectedNodeBalance, string.concat("Node vault balance ", description));
    }

    function test_happyPath() public {
        Definition memory def = Definition({
            challengeDuration: CHALLENGE_DURATION,
            participant: alice,
            node: node,
            nonce: NONCE,
            metadata: bytes32(0)
        });

        bytes32 channelId = keccak256(abi.encode(def.challengeDuration, def.participant, def.node, def.nonce));

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
                userAllocation: 1000,
                userNetFlow: 1000,
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
            participantSig: "",
            nodeSig: ""
        });

        state = signStateWithBothParties(state, channelId);

        vm.prank(alice);
        cHub.createChannel(def, state);

        // Verify user balance after channel creation (deposited 1000)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1000, "User balance after channel creation");

        // transfer 42 (allocation decreases by 42, node net flow decreases by 42)
        state = nextState(state, StateIntent.OPERATE, [uint256(958), uint256(0)], [int256(1000), int256(-42)]);
        state = signStateWithBothParties(state, channelId);

        // invoke a checkpoint
        // Expected: user allocation = 958, user net flow = 1000, node allocation = 0, node net flow = -42
        vm.prank(alice);
        cHub.checkpointChannel(channelId, state, new CrossChainState[](0));
        verifyChannelState(channelId, 958, 1000, 0, -42, "after checkpoint");

        // receive 24 (allocation increases by 24, node net flow increases by 24)
        state = nextState(state, StateIntent.OPERATE, [uint256(982), uint256(0)], [int256(1000), int256(-18)]);
        state = signStateWithBothParties(state, channelId);

        // invoke a deposit (500)
        // Expected: user allocation = 1482, user net flow = 1500, node allocation = 0, node net flow = -18
        state = nextState(state, StateIntent.DEPOSIT, [uint256(1482), uint256(0)], [int256(1500), int256(-18)]);
        state = signStateWithBothParties(state, channelId);

        vm.prank(alice);
        cHub.depositToChannel(channelId, state);
        verifyChannelState(channelId, 1482, 1500, 0, -18, "after deposit");

        // Verify user balance after first deposit (deposited 500 more)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1500, "User balance after first deposit");

        // transfer 1
        state = nextState(state, StateIntent.OPERATE, [uint256(1481), uint256(0)], [int256(1500), int256(-19)]);
        state = signStateWithBothParties(state, channelId);

        // transfer 2
        state = nextState(state, StateIntent.OPERATE, [uint256(1479), uint256(0)], [int256(1500), int256(-21)]);
        state = signStateWithBothParties(state, channelId);

        // invoke a withdrawal (100)
        // Expected: user allocation = 1379, user net flow = 1400, node allocation = 0, node net flow = -21
        state = nextState(state, StateIntent.WITHDRAW, [uint256(1379), uint256(0)], [int256(1400), int256(-21)]);
        state = signStateWithBothParties(state, channelId);

        vm.prank(alice);
        cHub.withdrawFromChannel(channelId, state);
        verifyChannelState(channelId, 1379, 1400, 0, -21, "after withdrawal");

        // Verify user balance after first withdrawal (withdrew 100)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1400, "User balance after first withdrawal");

        // transfer 3
        state = nextState(state, StateIntent.OPERATE, [uint256(1376), uint256(0)], [int256(1400), int256(-24)]);
        state = signStateWithBothParties(state, channelId);

        // receive 10
        state = nextState(state, StateIntent.OPERATE, [uint256(1386), uint256(0)], [int256(1400), int256(-14)]);
        state = signStateWithBothParties(state, channelId);

        // invoke a deposit (200)
        // Expected: user allocation = 1586, user net flow = 1600, node allocation = 0, node net flow = -14
        state = nextState(state, StateIntent.DEPOSIT, [uint256(1586), uint256(0)], [int256(1600), int256(-14)]);
        state = signStateWithBothParties(state, channelId);

        vm.prank(alice);
        cHub.depositToChannel(channelId, state);
        verifyChannelState(channelId, 1586, 1600, 0, -14, "after second deposit");

        // Verify user balance after second deposit (deposited 200 more)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1600, "User balance after second deposit");

        // receive 1
        state = nextState(state, StateIntent.OPERATE, [uint256(1587), uint256(0)], [int256(1600), int256(-13)]);
        state = signStateWithBothParties(state, channelId);

        // transfer 2
        state = nextState(state, StateIntent.OPERATE, [uint256(1585), uint256(0)], [int256(1600), int256(-15)]);
        state = signStateWithBothParties(state, channelId);

        // receive 3
        state = nextState(state, StateIntent.OPERATE, [uint256(1588), uint256(0)], [int256(1600), int256(-12)]);
        state = signStateWithBothParties(state, channelId);

        // transfer 4
        state = nextState(state, StateIntent.OPERATE, [uint256(1584), uint256(0)], [int256(1600), int256(-16)]);
        state = signStateWithBothParties(state, channelId);

        // receive 5
        state = nextState(state, StateIntent.OPERATE, [uint256(1589), uint256(0)], [int256(1600), int256(-11)]);
        state = signStateWithBothParties(state, channelId);

        // withdraw (300)
        // Expected: user allocation = 1289, user net flow = 1300, node allocation = 0, node net flow = -11
        state = nextState(state, StateIntent.WITHDRAW, [uint256(1289), uint256(0)], [int256(1300), int256(-11)]);
        state = signStateWithBothParties(state, channelId);

        vm.prank(alice);
        cHub.withdrawFromChannel(channelId, state);
        verifyChannelState(channelId, 1289, 1300, 0, -11, "after second withdrawal");

        // Verify user balance after second withdrawal (withdrew 300)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1300, "User balance after second withdrawal");

        // transfer 1
        // Expected: user allocation = 1288, user net flow = 1300, node allocation = 0, node net flow = -12
        state = nextState(state, StateIntent.OPERATE, [uint256(1288), uint256(0)], [int256(1300), int256(-12)]);
        state = signStateWithBothParties(state, channelId);

        // receive 2
        // Expected: user allocation = 1290, user net flow = 1300, node allocation = 0, node net flow = -10
        state = nextState(state, StateIntent.OPERATE, [uint256(1290), uint256(0)], [int256(1300), int256(-10)]);
        state = signStateWithBothParties(state, channelId);

        // transfer 3
        // Expected: user allocation = 1287, user net flow = 1300, node allocation = 0, node net flow = -13
        state = nextState(state, StateIntent.OPERATE, [uint256(1287), uint256(0)], [int256(1300), int256(-13)]);
        state = signStateWithBothParties(state, channelId);

        // close channel
        // Expected: user allocation = 1287, node allocation = 0
        state = nextState(state, StateIntent.CLOSE, [uint256(1287), uint256(0)], [int256(1300), int256(-13)]);
        state = signStateWithBothParties(state, channelId);

        vm.prank(alice);
        cHub.closeChannel(channelId, state, new CrossChainState[](0));

        // Check VOID status after channel closure
        (ChannelStatus finalStatus,,,,) = cHub.getChannelData(channelId);
        assertEq(uint8(finalStatus), uint8(ChannelStatus.VOID), "Channel should be VOID after closure");

        // Verify user balance after channel closure (received back 1287)
        assertEq(token.balanceOf(alice), INITIAL_BALANCE - 1300 + 1287, "User balance after channel closure");
    }
}
