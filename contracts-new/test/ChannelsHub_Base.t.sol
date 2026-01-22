// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Vm} from "lib/forge-std/src/Vm.sol";
import {Test} from "lib/forge-std/src/Test.sol";

import {MockERC20} from "./mocks/MockERC20.sol";
import {TestUtils} from "./TestUtils.sol";

import {ChannelsHub} from "../src/ChannelsHub.sol";
import {CrossChainState, StateIntent, State} from "../src/interfaces/Types.sol";

contract ChannelsHubTest_Base is Test {
    ChannelsHub public cHub;
    MockERC20 public token;

    uint256 constant nodePK = 1;
    uint256 constant alicePK = 2;
    uint256 constant bobPK = 3;

    address node;
    address alice;
    address bob;

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
        bob = vm.addr(bobPK);

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
            userSig: "",
            nodeSig: ""
        });
    }

    function nextCrossChainState(
        CrossChainState memory state,
        StateIntent intent,
        uint256[2] memory allocations,
        int256[2] memory netFlows,
        uint64 nonHomeChainId,
        address nonHomeChainToken,
        uint256[2] memory nonHomeAllocations,
        int256[2] memory nonHomeNetFlows
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
                chainId: nonHomeChainId,
                token: nonHomeChainToken,
                userAllocation: nonHomeAllocations[0],
                userNetFlow: nonHomeNetFlows[0],
                nodeAllocation: nonHomeAllocations[1],
                nodeNetFlow: nonHomeNetFlows[1]
            }),
            userSig: "",
            nodeSig: ""
        });
    }

    function signStateWithBothParties(CrossChainState memory state, bytes32 channelId, uint256 userPK) internal pure returns (CrossChainState memory) {
        state.userSig = TestUtils.signStateEIP191(vm, channelId, state, userPK);
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
}
