// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Test} from "lib/forge-std/src/Test.sol";
import {Vm} from "lib/forge-std/src/Vm.sol";

import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";

import {TestUtils} from "../TestUtils.sol";
import {MockERC20} from "../mocks/MockERC20.sol";

import {IAdjudicator} from "../../src/interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, StateIntent} from "../../src/interfaces/Types.sol";
import {ConsensusTransition} from "../../src/adjudicators/ConsensusTransition.sol";
import {Utils} from "../../src/Utils.sol";

contract ConsensusTransitionTest is Test {
    using ECDSA for bytes32;

    ConsensusTransition public adjudicator;

    // Mockup constructor parameters
    address mockedOwner = address(0x456);
    address mockedChannelImpl = address(0x123);

    // Test accounts
    address public host;
    address public guest;
    uint256 public hostPrivateKey;
    uint256 public guestPrivateKey;

    // Channel parameters
    Channel public channel;
    MockERC20 public token;

    // Constants for participant ordering
    uint256 private constant HOST = 0;
    uint256 private constant GUEST = 1;

    function setUp() public {
        // Deploy the adjudicator contract
        adjudicator = new ConsensusTransition(mockedOwner, mockedChannelImpl);

        // Generate private keys and addresses for the participants
        hostPrivateKey = 0x1;
        guestPrivateKey = 0x2;
        host = vm.addr(hostPrivateKey);
        guest = vm.addr(guestPrivateKey);

        // Deploy the mock token
        token = new MockERC20("Test Token", "TEST", 18);

        // Set up the channel
        address[] memory participants = new address[](2);
        participants[HOST] = host;
        participants[GUEST] = guest;
        channel = Channel({
            participants: participants,
            adjudicator: address(adjudicator),
            challenge: 3600, // 1 hour challenge period
            nonce: 1
        });
    }

    // Helper function to create test allocations
    function createAllocations(uint256 hostAmount, uint256 guestAmount) internal view returns (Allocation[2] memory) {
        Allocation[2] memory allocations;

        allocations[HOST] = Allocation({destination: host, token: address(token), amount: hostAmount});

        allocations[GUEST] = Allocation({destination: guest, token: address(token), amount: guestAmount});

        return allocations;
    }

    // Helper function to create an initial state
    function _createInitialState(string memory data) internal view returns (State memory) {
        // Create allocations
        Allocation[2] memory allocations = createAllocations(50, 50);

        // Create the state
        State memory state;
        state.intent = StateIntent.INITIALIZE;
        state.version = 0;
        state.data = bytes(data);
        state.allocations = new Allocation[](2);
        state.allocations[HOST] = allocations[HOST];
        state.allocations[GUEST] = allocations[GUEST];
        state.sigs = new bytes[](0);

        return state;
    }

    // Helper function to create an operation state
    function _createOperateState(string memory data, uint256 version) internal view returns (State memory) {
        // Create allocations
        Allocation[2] memory allocations = createAllocations(50, 50);

        // Create the state
        State memory state;
        state.intent = StateIntent.OPERATE;
        state.version = version;
        state.data = bytes(data);
        state.allocations = new Allocation[](2);
        state.allocations[HOST] = allocations[HOST];
        state.allocations[GUEST] = allocations[GUEST];
        state.sigs = new bytes[](0);

        return state;
    }

    // Helper function to create a resize state
    function _createResizeState(string memory data, uint256 version, int256[] memory resizeAmounts)
        internal
        view
        returns (State memory)
    {
        // Create allocations
        Allocation[2] memory allocations = createAllocations(50, 50);

        // Create the state
        State memory state;
        state.intent = StateIntent.RESIZE;
        state.version = version;
        state.data = abi.encode(resizeAmounts, bytes(data));
        state.allocations = new Allocation[](2);
        state.allocations[HOST] = allocations[HOST];
        state.allocations[GUEST] = allocations[GUEST];
        state.sigs = new bytes[](0);

        return state;
    }

    // Helper to sign a state
    function _signState(State memory state, uint256 privateKey) internal view returns (bytes memory) {
        bytes memory packedState = Utils.getPackedState(Utils.getChannelId(channel), state);
        return TestUtils.sign(vm, privateKey, packedState);
    }

    // -------------------- FIRST STATE TRANSITION TESTS --------------------

    function test_adjudicate_firstState_valid() public {
        // Create initial state with both signatures
        State memory initialState = _createInitialState("initial state");
        initialState.sigs = new bytes[](2);
        initialState.sigs[HOST] = _signState(initialState, hostPrivateKey);
        initialState.sigs[GUEST] = _signState(initialState, guestPrivateKey);

        // Create first operate state with both signatures
        State memory firstState = _createOperateState("first state", 1);
        firstState.sigs = new bytes[](2);
        firstState.sigs[HOST] = _signState(firstState, hostPrivateKey);
        firstState.sigs[GUEST] = _signState(firstState, guestPrivateKey);

        // Provide the initial state as proof
        State[] memory proofs = new State[](1);
        proofs[0] = initialState;

        // Verify that the transition is valid
        bool valid = adjudicator.adjudicate(channel, firstState, proofs);
        assertTrue(valid, "Valid first state transition should be accepted");
    }

    function test_adjudicate_firstState_revert_whenMissingParticipantSignature() public {
        // Create initial state with both signatures
        State memory initialState = _createInitialState("initial state");
        initialState.sigs = new bytes[](2);
        initialState.sigs[HOST] = _signState(initialState, hostPrivateKey);
        initialState.sigs[GUEST] = _signState(initialState, guestPrivateKey);

        // Create first operate state with only one signature
        State memory firstState = _createOperateState("first state", 1);
        firstState.sigs = new bytes[](1);
        firstState.sigs[0] = _signState(firstState, hostPrivateKey);

        // Provide the initial state as proof
        State[] memory proofs = new State[](1);
        proofs[0] = initialState;

        // Verify that the transition is rejected
        bool valid = adjudicator.adjudicate(channel, firstState, proofs);
        assertFalse(valid, "First state without both signatures should be rejected");
    }

    function test_adjudicate_firstState_revert_whenIncorrectVersion() public {
        // Create initial state with both signatures
        State memory initialState = _createInitialState("initial state");
        initialState.sigs = new bytes[](2);
        initialState.sigs[HOST] = _signState(initialState, hostPrivateKey);
        initialState.sigs[GUEST] = _signState(initialState, guestPrivateKey);

        // Create first operate state with incorrect version (2 instead of 1)
        State memory firstState = _createOperateState("first state", 2);
        firstState.sigs = new bytes[](2);
        firstState.sigs[HOST] = _signState(firstState, hostPrivateKey);
        firstState.sigs[GUEST] = _signState(firstState, guestPrivateKey);

        // Provide the initial state as proof
        State[] memory proofs = new State[](1);
        proofs[0] = initialState;

        // Verify that the transition is rejected
        bool valid = adjudicator.adjudicate(channel, firstState, proofs);
        assertFalse(valid, "First state with incorrect version should be rejected");
    }

    // -------------------- LATER STATE TRANSITION TESTS --------------------

    function test_adjudicate_laterState_valid() public {
        // Create state 1 with both signatures
        State memory state1 = _createOperateState("state 1", 1);
        state1.sigs = new bytes[](2);
        state1.sigs[HOST] = _signState(state1, hostPrivateKey);
        state1.sigs[GUEST] = _signState(state1, guestPrivateKey);

        // Create state 2 with both signatures
        State memory state2 = _createOperateState("state 2", 2);
        state2.sigs = new bytes[](2);
        state2.sigs[HOST] = _signState(state2, hostPrivateKey);
        state2.sigs[GUEST] = _signState(state2, guestPrivateKey);

        // Provide state 1 as proof for state 2
        State[] memory proofs = new State[](1);
        proofs[0] = state1;

        // Verify that the transition is valid
        bool valid = adjudicator.adjudicate(channel, state2, proofs);
        assertTrue(valid, "Valid state transition from 1 to 2 should be accepted");
    }

    function test_adjudicate_laterState_revert_whenIncorrectVersionIncrement() public {
        // Create state 1 with both signatures
        State memory state1 = _createOperateState("state 1", 1);
        state1.sigs = new bytes[](2);
        state1.sigs[HOST] = _signState(state1, hostPrivateKey);
        state1.sigs[GUEST] = _signState(state1, guestPrivateKey);

        // Create state 3 with both signatures (skipping version 2)
        State memory state3 = _createOperateState("state 3", 3);
        state3.sigs = new bytes[](2);
        state3.sigs[HOST] = _signState(state3, hostPrivateKey);
        state3.sigs[GUEST] = _signState(state3, guestPrivateKey);

        // Provide state 1 as proof for state 3
        State[] memory proofs = new State[](1);
        proofs[0] = state1;

        // Verify that the transition is rejected
        bool valid = adjudicator.adjudicate(channel, state3, proofs);
        assertFalse(valid, "State with non-sequential version should be rejected");
    }

    function test_adjudicate_laterState_revert_whenAllocationSumChanged() public {
        // Create state 1 with both signatures
        State memory state1 = _createOperateState("state 1", 1);
        state1.sigs = new bytes[](2);
        state1.sigs[HOST] = _signState(state1, hostPrivateKey);
        state1.sigs[GUEST] = _signState(state1, guestPrivateKey);

        // Create state 2 with different allocation sums
        State memory state2 = _createOperateState("state 2", 2);
        state2.allocations[HOST].amount = 60; // Changed from 50
        state2.allocations[GUEST].amount = 50; // Kept at 50, total sum is now 110 instead of 100
        state2.sigs = new bytes[](2);
        state2.sigs[HOST] = _signState(state2, hostPrivateKey);
        state2.sigs[GUEST] = _signState(state2, guestPrivateKey);

        // Provide state 1 as proof for state 2
        State[] memory proofs = new State[](1);
        proofs[0] = state1;

        // Verify that the transition is rejected
        bool valid = adjudicator.adjudicate(channel, state2, proofs);
        assertFalse(valid, "State with changed allocation sum should be rejected");
    }

    function test_adjudicate_revert_whenNoStateProof() public {
        // Create state 2 without providing a proof
        State memory state2 = _createOperateState("state 2", 2);
        state2.sigs = new bytes[](2);
        state2.sigs[HOST] = _signState(state2, hostPrivateKey);
        state2.sigs[GUEST] = _signState(state2, guestPrivateKey);

        // Provide empty proofs array
        State[] memory emptyProofs = new State[](0);

        // Verify that the adjudication is rejected
        bool valid = adjudicator.adjudicate(channel, state2, emptyProofs);
        assertFalse(valid, "State without proof should be rejected");
    }

    function test_adjudicate_revert_whenTooManyProofs() public {
        // Create state 1
        State memory state1 = _createOperateState("state 1", 1);
        state1.sigs = new bytes[](2);
        state1.sigs[HOST] = _signState(state1, hostPrivateKey);
        state1.sigs[GUEST] = _signState(state1, guestPrivateKey);

        // Create state 2
        State memory state2 = _createOperateState("state 2", 2);
        state2.sigs = new bytes[](2);
        state2.sigs[HOST] = _signState(state2, hostPrivateKey);
        state2.sigs[GUEST] = _signState(state2, guestPrivateKey);

        // Provide both initial state and state 1 as proofs (too many)
        State[] memory tooManyProofs = new State[](2);
        tooManyProofs[0] = _createInitialState("initial state");
        tooManyProofs[1] = state1;

        // Verify that the adjudication is rejected
        bool valid = adjudicator.adjudicate(channel, state2, tooManyProofs);
        assertFalse(valid, "State with too many proofs should be rejected");
    }

    // -------------------- RESIZE STATE TRANSITION TESTS --------------------

    function test_adjudicate_afterResize_valid() public {
        // Create state 1 with both signatures
        State memory state1 = _createOperateState("state 1", 1);
        state1.sigs = new bytes[](2);
        state1.sigs[HOST] = _signState(state1, hostPrivateKey);
        state1.sigs[GUEST] = _signState(state1, guestPrivateKey);

        // Create resize state 2
        int256[] memory resizeAmounts = new int256[](2);
        resizeAmounts[HOST] = 10;
        resizeAmounts[GUEST] = -10;
        State memory resizeState = _createResizeState("resize state", 2, resizeAmounts);
        resizeState.sigs = new bytes[](2);
        resizeState.sigs[HOST] = _signState(resizeState, hostPrivateKey);
        resizeState.sigs[GUEST] = _signState(resizeState, guestPrivateKey);

        // Create state 3 after resize
        State memory state3 = _createOperateState("state 3", 3);
        state3.sigs = new bytes[](2);
        state3.sigs[HOST] = _signState(state3, hostPrivateKey);
        state3.sigs[GUEST] = _signState(state3, guestPrivateKey);

        // Provide resize state as proof for state 3
        State[] memory proofs = new State[](1);
        proofs[0] = resizeState;

        // Verify that the transition is valid
        bool valid = adjudicator.adjudicate(channel, state3, proofs);
        assertTrue(valid, "Valid state transition after resize should be accepted");
    }

    // Test signature validation using a non-corrupt signature but wrong signer
    function test_WrongSignerRejected() public {
        // Create state with signatures from wrong participants
        State memory state = _createOperateState("state", 1);
        state.sigs = new bytes[](2);

        // Use guest's signature for both slots
        state.sigs[HOST] = _signState(state, guestPrivateKey); // Should be host, but using guest
        state.sigs[GUEST] = _signState(state, guestPrivateKey);

        // Create a valid initial state as proof
        State memory initialState = _createInitialState("initial state");
        initialState.sigs = new bytes[](2);
        initialState.sigs[HOST] = _signState(initialState, hostPrivateKey);
        initialState.sigs[GUEST] = _signState(initialState, guestPrivateKey);

        State[] memory proofs = new State[](1);
        proofs[0] = initialState;

        // Verify that the signature from wrong participant is rejected
        bool valid = adjudicator.adjudicate(channel, state, proofs);
        assertFalse(valid, "State with wrong signer should be rejected");
    }
}
