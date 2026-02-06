// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "lib/forge-std/src/Test.sol";

import {MockERC20} from "./mocks/MockERC20.sol";
import {TestUtils} from "./TestUtils.sol";

import {ChannelHub} from "../src/ChannelHub.sol";
import {State, StateIntent, Ledger} from "../src/interfaces/Types.sol";

// forge-lint: disable-next-item(unsafe-typecast)
contract ChannelHubTest_Base is Test {
    ChannelHub public cHub;
    MockERC20 public token;

    uint256 constant NODE_PK = 1;
    uint256 constant ALICE_PK = 2;
    uint256 constant BOB_PK = 3;

    address node;
    address alice;
    address bob;

    uint8 constant CHANNEL_HUB_VERSION = 1;
    uint32 constant CHALLENGE_DURATION = 86400; // 1 day
    uint64 constant NONCE = 1;
    uint256 constant DEPOSIT_AMOUNT = 1000;
    uint256 constant INITIAL_BALANCE = 10000;

    function setUp() public virtual {
        // Deploy contracts
        cHub = new ChannelHub();
        token = new MockERC20("Test Token", "TST", 18);

        node = vm.addr(NODE_PK);
        alice = vm.addr(ALICE_PK);
        bob = vm.addr(BOB_PK);

        token.mint(node, INITIAL_BALANCE);
        token.mint(alice, INITIAL_BALANCE);
        token.mint(bob, INITIAL_BALANCE);

        vm.startPrank(node);
        token.approve(address(cHub), INITIAL_BALANCE);
        cHub.depositToVault(node, address(token), INITIAL_BALANCE);
        vm.stopPrank();

        vm.prank(alice);
        token.approve(address(cHub), INITIAL_BALANCE);

        vm.prank(bob);
        token.approve(address(cHub), INITIAL_BALANCE);
    }

    function nextState(State memory state, StateIntent intent, uint256[2] memory allocations, int256[2] memory netFlows)
        internal
        pure
        returns (State memory)
    {
        return State({
            version: state.version + 1,
            intent: intent,
            metadata: state.metadata,
            homeState: Ledger({
                chainId: state.homeState.chainId,
                token: state.homeState.token,
                decimals: state.homeState.decimals,
                userAllocation: allocations[0],
                userNetFlow: netFlows[0],
                nodeAllocation: allocations[1],
                nodeNetFlow: netFlows[1]
            }),
            nonHomeState: Ledger({
                chainId: 0,
                token: address(0),
                decimals: 0,
                userAllocation: 0,
                userNetFlow: 0,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            userSig: "",
            nodeSig: ""
        });
    }

    function nextState(
        State memory state,
        StateIntent intent,
        uint256[2] memory allocations,
        int256[2] memory netFlows,
        uint64 nonHomeChainId,
        address nonHomeChainToken,
        uint256[2] memory nonHomeAllocations,
        int256[2] memory nonHomeNetFlows
    ) internal pure returns (State memory) {
        return State({
            version: state.version + 1,
            intent: intent,
            metadata: state.metadata,
            homeState: Ledger({
                chainId: state.homeState.chainId,
                token: state.homeState.token,
                decimals: state.homeState.decimals,
                userAllocation: allocations[0],
                userNetFlow: netFlows[0],
                nodeAllocation: allocations[1],
                nodeNetFlow: netFlows[1]
            }),
            nonHomeState: Ledger({
                chainId: nonHomeChainId,
                token: nonHomeChainToken,
                decimals: 18,
                userAllocation: nonHomeAllocations[0],
                userNetFlow: nonHomeNetFlows[0],
                nodeAllocation: nonHomeAllocations[1],
                nodeNetFlow: nonHomeNetFlows[1]
            }),
            userSig: "",
            nodeSig: ""
        });
    }

    function nextState(
        State memory state,
        StateIntent intent,
        uint256[2] memory allocations,
        int256[2] memory netFlows,
        uint64 nonHomeChainId,
        address nonHomeChainToken,
        uint8 nonHomeDecimals,
        uint256[2] memory nonHomeAllocations,
        int256[2] memory nonHomeNetFlows
    ) internal pure returns (State memory) {
        return State({
            version: state.version + 1,
            intent: intent,
            metadata: state.metadata,
            homeState: Ledger({
                chainId: state.homeState.chainId,
                token: state.homeState.token,
                decimals: state.homeState.decimals,
                userAllocation: allocations[0],
                userNetFlow: netFlows[0],
                nodeAllocation: allocations[1],
                nodeNetFlow: netFlows[1]
            }),
            nonHomeState: Ledger({
                chainId: nonHomeChainId,
                token: nonHomeChainToken,
                decimals: nonHomeDecimals,
                userAllocation: nonHomeAllocations[0],
                userNetFlow: nonHomeNetFlows[0],
                nodeAllocation: nonHomeAllocations[1],
                nodeNetFlow: nonHomeNetFlows[1]
            }),
            userSig: "",
            nodeSig: ""
        });
    }

    function signStateWithBothParties(State memory state, bytes32 channelId, uint256 userPk)
        internal
        pure
        returns (State memory)
    {
        state.userSig = TestUtils.signStateEip191(vm, channelId, state, userPk);
        state.nodeSig = TestUtils.signStateEip191(vm, channelId, state, NODE_PK);
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
        (,, State memory latestState,,) = cHub.getChannelData(channelId);
        assertEq(
            latestState.homeState.userAllocation, expectedUserAllocation, string.concat("User allocation ", description)
        );
        assertEq(latestState.homeState.userNetFlow, expectedUserNetFlow, string.concat("User net flow ", description));
        assertEq(
            latestState.homeState.nodeAllocation, expectedNodeAllocation, string.concat("Node allocation ", description)
        );
        assertEq(latestState.homeState.nodeNetFlow, expectedNodeNetFlow, string.concat("Node net flow ", description));

        uint256 nodeBalance = cHub.getAccountBalance(node, address(token));
        uint256 expectedNodeBalance = expectedNodeNetFlow < 0
            ? INITIAL_BALANCE + uint256(-expectedNodeNetFlow)
            : INITIAL_BALANCE - uint256(expectedNodeNetFlow);
        assertEq(nodeBalance, expectedNodeBalance, string.concat("Node vault balance ", description));
    }
}
