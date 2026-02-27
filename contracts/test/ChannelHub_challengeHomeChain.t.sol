// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "./ChannelHub_Base.t.sol";

import {Utils} from "../src/Utils.sol";
import {
    State,
    ChannelDefinition,
    StateIntent,
    Ledger,
    ChannelStatus,
    ParticipantIndex,
    DEFAULT_SIG_VALIDATOR_ID
} from "../src/interfaces/Types.sol";
import {SessionKeyAuthorization} from "../src/sigValidators/SessionKeyValidator.sol";
import {TestUtils, SESSION_KEY_VALIDATOR_ID} from "./TestUtils.sol";
import {ChannelHub} from "../src/ChannelHub.sol";
import {ChannelEngine} from "../src/ChannelEngine.sol";

contract ChannelHubTest_Challenge_HomeChain_NormalOperation is ChannelHubTest_Base {
    /*
    - a channel can be challenged with a newer state, which is enforced during challenge
    - a channel can be challenged with existing state, which is NOT enforced the second time during challenge
    - challenge is finalized (funds can be withdrawn) after `challengeExpireAt` time expires
    - challenged "operating" state can be resolved with a newer state until `challengeExpireAt` time has NOT passed
    - challenged state can NOT be resolved after `challengeExpireAt` time has passed
    - it is not possible to challenge an already challenged channel
    - a channel can NOT be challenged with an earlier state
    */

    ChannelDefinition def;
    bytes32 channelId;
    State initState;

    function setUp() public override {
        super.setUp();

        def = ChannelDefinition({
            challengeDuration: CHALLENGE_DURATION,
            user: alice,
            node: node,
            nonce: NONCE,
            approvedSignatureValidators: 0,
            metadata: bytes32(0)
        });

        channelId = Utils.getChannelId(def, CHANNEL_HUB_VERSION);

        initState = State({
            version: 0,
            intent: StateIntent.DEPOSIT,
            metadata: bytes32(0),
            homeLedger: Ledger({
                chainId: uint64(block.chainid),
                token: address(token),
                decimals: 18,
                userAllocation: 1000,
                userNetFlow: 1000,
                nodeAllocation: 0,
                nodeNetFlow: 0
            }),
            nonHomeLedger: Ledger({
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
        initState = mutualSignStateBothWithEcdsaValidator(initState, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.createChannel(def, initState);

    }

    function signChallengeEip191WithEcdsaValidator(
        bytes32 channelId_,
        State memory state,
        uint256 privateKey
    ) internal pure returns (bytes memory) {
        bytes memory signingData = Utils.toSigningData(state);
        bytes memory challengerSigningData = abi.encodePacked(signingData, "challenge");
        bytes memory message = Utils.pack(channelId_, challengerSigningData);
        bytes memory signature = TestUtils.signEip191(vm, privateKey, message);
        return abi.encodePacked(DEFAULT_SIG_VALIDATOR_ID, signature);
    }

    function test_challengeWithNewerState_enforcesState() public {
        // Off-chain: user transfers 100 to node
        State memory stateV1 = nextState(initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]);
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        // Off-chain: user transfers another 50 to node
        State memory stateV2 = nextState(stateV1, StateIntent.OPERATE, [uint256(850), uint256(0)], [int256(1000), int256(-150)]);
        stateV2 = mutualSignStateBothWithEcdsaValidator(stateV2, channelId, ALICE_PK);

        // Node challenges with newer state V2, which should be enforced during challenge
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, stateV2, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, stateV2, challengerSig, ParticipantIndex.NODE);

        verifyChannelData(channelId, ChannelStatus.DISPUTED, 2, block.timestamp + CHALLENGE_DURATION, "State V2 should be enforced during challenge");
        verifyChannelState(channelId, 850, 1000, 0, -150, "State V2 should be enforced during challenge");
    }

    function test_challengeWithExistingState_notEnforcedAgain() public {
        // Checkpoint a new state
        State memory stateV1 = nextState(initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]);
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        vm.prank(alice);
        cHub.checkpointChannel(channelId, stateV1);

        // Verify state V1 is on-chain
        (,, State memory latestStateBefore,,) = cHub.getChannelData(channelId);
        assertEq(latestStateBefore.version, 1, "State version should be 1 before challenge");

        // Node challenges with the same state V1 (already on-chain)
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, stateV1, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, stateV1, challengerSig, ParticipantIndex.NODE);

        verifyChannelData(channelId, ChannelStatus.DISPUTED, 1, block.timestamp + CHALLENGE_DURATION, "State V1 should be enforced during challenge");
        verifyChannelState(channelId, 900, 1000, 0, -100, "State V1 should be enforced during challenge");
    }

    function test_challengeFinalization_afterTimeout() public {
        State memory stateV1 = nextState(initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]);
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        // Challenge with current state
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, stateV1, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, stateV1, challengerSig, ParticipantIndex.NODE);

        vm.warp(block.timestamp + CHALLENGE_DURATION + 1);

        uint256 aliceBalanceBefore = token.balanceOf(alice);
        uint256 nodeBalanceBefore = cHub.getAccountBalance(node, address(token));

        // Finalize challenge by closing the channel (unilateral closure)
        // When doing unilateral closure after timeout, any state works
        vm.prank(alice);
        cHub.closeChannel(channelId, initState);

        // Verify channel is CLOSED and funds were distributed according to last enforced state (V1)
        verifyChannelData(channelId, ChannelStatus.CLOSED, 1, 0, "Channel should be CLOSED after challenge finalization");

        uint256 aliceBalanceAfter = token.balanceOf(alice);
        uint256 nodeBalanceAfter = cHub.getAccountBalance(node, address(token));

        assertEq(aliceBalanceAfter, aliceBalanceBefore + 900, "Alice should receive her allocation");
        // Node balance should remain unchanged because:
        // 1. The node already received its 100 when the challenge was processed (nodeNetFlow -100 released funds)
        // 2. During unilateral closure, node gets nodeAllocation (0)
        assertEq(nodeBalanceAfter, nodeBalanceBefore, "Node balance should remain unchanged (already received net flow during challenge)");
    }

    function test_resolveChallenge_withNewerState_beforeTimeout() public {
        // State V1: user transfers 100
        State memory stateV1 = nextState(initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]);
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        // Challenge with stateV1
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, stateV1, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, stateV1, challengerSig, ParticipantIndex.NODE);

        verifyChannelData(channelId, ChannelStatus.DISPUTED, 1, block.timestamp + CHALLENGE_DURATION, "Channel should be DISPUTED after challenge");

        // State V2: user transfers another 50 (newer state to resolve challenge)
        State memory stateV2 = nextState(stateV1, StateIntent.OPERATE, [uint256(850), uint256(0)], [int256(1000), int256(-150)]);
        stateV2 = mutualSignStateBothWithEcdsaValidator(stateV2, channelId, ALICE_PK);

        // Resolve challenge by checkpointing newer state (before timeout)
        vm.prank(alice);
        cHub.checkpointChannel(channelId, stateV2);

        verifyChannelData(channelId, ChannelStatus.OPERATING, 2, 0, "Channel should be OPERATING after resolving challenge with newer state");
        verifyChannelState(channelId, 850, 1000, 0, -150, "State V2 should be enforced after resolving challenge with newer state");
    }

    function test_revert_resolveChallenge_withOlderState_beforeTimeout() public {
        // State V1: user transfers 100
        State memory stateV1 = nextState(initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]);
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        // State V2: user receives 50 back
        State memory stateV2 = nextState(stateV1, StateIntent.OPERATE, [uint256(950), uint256(0)], [int256(1000), int256(-50)]);
        stateV2 = mutualSignStateBothWithEcdsaValidator(stateV2, channelId, ALICE_PK);

        // Challenge with stateV2
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, stateV2, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, stateV2, challengerSig, ParticipantIndex.NODE);

        verifyChannelData(channelId, ChannelStatus.DISPUTED, 2, block.timestamp + CHALLENGE_DURATION, "Channel should be DISPUTED after challenge");

        // Try to resolve with older state V1 (should fail)
        vm.expectRevert(ChannelEngine.IncorrectStateVersion.selector);
        vm.prank(alice);
        cHub.checkpointChannel(channelId, stateV1);
    }

    function test_revert_resolveChallenge_withNewerState_afterTimeout() public {
        // State V1
        State memory stateV1 = nextState(initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]);
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        // Challenge
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, stateV1, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, stateV1, challengerSig, ParticipantIndex.NODE);

        vm.warp(block.timestamp + CHALLENGE_DURATION + 1);

        // State V2: user transfers another 50 (newer state to resolve challenge)
        State memory stateV2 = nextState(stateV1, StateIntent.OPERATE, [uint256(850), uint256(0)], [int256(1000), int256(-150)]);
        stateV2 = mutualSignStateBothWithEcdsaValidator(stateV2, channelId, ALICE_PK);

        // Cannot resolve challenge after timeout - must close channel instead
        vm.expectRevert(ChannelEngine.ChallengeExpired.selector);
        vm.prank(alice);
        cHub.checkpointChannel(channelId, stateV2);
    }

    function test_revert_challengeAlreadyChallengedChannel() public {
        // First challenge
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, initState, NODE_PK);

        vm.prank(node);
        cHub.challengeChannel(channelId, initState, challengerSig, ParticipantIndex.NODE);

        // Verify channel is DISPUTED
        verifyChannelData(channelId, ChannelStatus.DISPUTED, 0, block.timestamp + CHALLENGE_DURATION, "Channel should be DISPUTED after first challenge");

        // Try to challenge again (should fail)
        State memory stateV1 = nextState(initState, StateIntent.OPERATE, [uint256(850), uint256(0)], [int256(1000), int256(-150)]);
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        bytes memory challengerSig2 = signChallengeEip191WithEcdsaValidator(channelId, stateV1, NODE_PK);

        vm.prank(node);
        vm.expectRevert(ChannelHub.IncorrectChannelStatus.selector);
        cHub.challengeChannel(channelId, stateV1, challengerSig2, ParticipantIndex.NODE);
    }

    function test_revert_challengeWithOlderState() public {
        // State V1
        State memory stateV1 = nextState(initState, StateIntent.OPERATE, [uint256(900), uint256(0)], [int256(1000), int256(-100)]);
        stateV1 = mutualSignStateBothWithEcdsaValidator(stateV1, channelId, ALICE_PK);

        // Checkpoint V1
        vm.prank(alice);
        cHub.checkpointChannel(channelId, stateV1);

        // Try to challenge with older state (initial) (should fail)
        bytes memory challengerSig = signChallengeEip191WithEcdsaValidator(channelId, initState, NODE_PK);

        vm.prank(node);
        vm.expectRevert(ChannelHub.ChallengerVersionTooLow.selector);
        cHub.challengeChannel(channelId, initState, challengerSig, ParticipantIndex.NODE);
    }
}
