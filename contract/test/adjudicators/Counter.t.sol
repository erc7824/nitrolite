// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Test} from "lib/forge-std/src/Test.sol";
import {Vm} from "lib/forge-std/src/Vm.sol";
import {IAdjudicator} from "../../src/interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature} from "../../src/interfaces/Types.sol";
import {Counter} from "../../src/adjudicators/Counter.sol";
import {Utils} from "../../src/Utils.sol";

contract CounterTest is Test {
    Counter public counterAdjudicator;

    // Test accounts
    address public host;
    address public guest;
    uint256 public hostPrivateKey;
    uint256 public guestPrivateKey;

    // Channel parameters
    Channel public channel;

    // Constants for participant ordering
    uint256 private constant HOST_IDX = 0;
    uint256 private constant GUEST_IDX = 1;

    function setUp() public {
        // Deploy the adjudicator
        counterAdjudicator = new Counter();

        // Set private keys and corresponding addresses
        hostPrivateKey = 0x1;
        guestPrivateKey = 0x2;
        host = vm.addr(hostPrivateKey);
        guest = vm.addr(guestPrivateKey);

        // Set up the channel with the two participants
        address[2] memory participants;
        participants[HOST_IDX] = host;
        participants[GUEST_IDX] = guest;
        channel = Channel({
            participants: participants,
            adjudicator: address(counterAdjudicator),
            challenge: 3600, // e.g., 1-hour challenge period
            nonce: 1
        });
    }

    // Helper function to create a Counter state.
    // The CounterApp struct is: { uint256 counter, uint256 target, uint256 version }
    function createCounterState(uint256 _counter, uint256 _target, uint256 _version)
        internal
        pure
        returns (State memory)
    {
        State memory state;
        // Encoding must match the order in the CounterApp struct.
        state.data = abi.encode(_counter, _target, _version);
        state.sigs = new Signature[](0);
        return state;
    }

    // Helper to sign a state using a given private key.
    function signState(State memory state, uint256 privateKey) internal view returns (Signature memory) {
        bytes32 stateHash = Utils.getStateHash(channel, state);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, stateHash);
        return Signature({v: v, r: r, s: s});
    }

    // -------------------- INITIAL STATE TESTS --------------------

    // Valid initial state: version 0 with two valid signatures (host, then guest)
    function test_ValidInitialState() public {
        uint256 counterVal = 0;
        uint256 target = 10;
        uint256 version = 0;
        State memory state = createCounterState(counterVal, target, version);

        // Provide exactly two valid signatures in the proper order.
        state.sigs = new Signature[](2);
        state.sigs[HOST_IDX] = signState(state, hostPrivateKey);
        state.sigs[GUEST_IDX] = signState(state, guestPrivateKey);

        State[] memory emptyProofs = new State[](0);
        bool valid = counterAdjudicator.adjudicate(channel, state, emptyProofs);
        assertTrue(valid, "Valid initial state should be accepted");
    }

    // Initial state with insufficient signatures should fail.
    function test_InitialStateInsufficientSignatures() public {
        uint256 counterVal = 0;
        uint256 target = 10;
        uint256 version = 0;
        State memory state = createCounterState(counterVal, target, version);

        // Only one signature provided.
        state.sigs = new Signature[](1);
        state.sigs[HOST_IDX] = signState(state, hostPrivateKey);

        State[] memory emptyProofs = new State[](0);
        bool valid = counterAdjudicator.adjudicate(channel, state, emptyProofs);
        assertFalse(valid, "Initial state with insufficient signatures should be rejected");
    }

    // Initial state with an invalid (corrupted) signature should fail.
    function test_InitialStateInvalidSignature() public {
        uint256 counterVal = 0;
        uint256 target = 10;
        uint256 version = 0;
        State memory state = createCounterState(counterVal, target, version);

        state.sigs = new Signature[](2);
        Signature memory sigHost = signState(state, hostPrivateKey);
        Signature memory sigGuest = signState(state, guestPrivateKey);
        // Corrupt the host signature.
        sigHost.r = bytes32(0);
        state.sigs[HOST_IDX] = sigHost;
        state.sigs[GUEST_IDX] = sigGuest;

        State[] memory emptyProofs = new State[](0);
        bool valid = counterAdjudicator.adjudicate(channel, state, emptyProofs);
        assertFalse(valid, "Initial state with an invalid signature should be rejected");
    }

    // -------------------- NON-INITIAL STATE TESTS --------------------

    // Valid non-initial state transition:
    // previous state has version 0 and candidate state has version 1,
    // counter increments by 1 and target remains the same.
    function test_ValidNonInitialState() public {
        // Create previous (initial) state.
        uint256 prevCounter = 3;
        uint256 target = 10;
        uint256 prevVersion = 0;
        State memory prevState = createCounterState(prevCounter, target, prevVersion);
        prevState.sigs = new Signature[](2);
        prevState.sigs[HOST_IDX] = signState(prevState, hostPrivateKey);
        prevState.sigs[GUEST_IDX] = signState(prevState, guestPrivateKey);

        // Create candidate (non-initial) state.
        uint256 newCounter = prevCounter + 1;
        uint256 newVersion = prevVersion + 1;
        State memory newState = createCounterState(newCounter, target, newVersion);
        // Even for non-initial states, exactly two signatures are required.
        newState.sigs = new Signature[](2);
        newState.sigs[HOST_IDX] = signState(newState, hostPrivateKey);
        newState.sigs[GUEST_IDX] = signState(newState, guestPrivateKey);

        // Provide the previous state as proof.
        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        bool valid = counterAdjudicator.adjudicate(channel, newState, proofs);
        assertTrue(valid, "Valid non-initial state transition should be accepted");
    }

    // Non-initial state with missing proof(s) should fail.
    function test_NonInitialStateMissingProofs() public {
        uint256 counterVal = 1;
        uint256 target = 10;
        uint256 version = 1; // Non-initial state
        State memory state = createCounterState(counterVal, target, version);
        state.sigs = new Signature[](2);
        state.sigs[HOST_IDX] = signState(state, hostPrivateKey);
        state.sigs[GUEST_IDX] = signState(state, guestPrivateKey);

        // No proofs provided.
        State[] memory proofs = new State[](0);

        bool valid = counterAdjudicator.adjudicate(channel, state, proofs);
        assertFalse(valid, "Non-initial state without proof should be rejected");
    }

    // Non-initial state with an incorrect counter increment (not exactly +1) should fail.
    function test_InvalidCounterIncrement() public {
        // Previous state: counter = 3.
        uint256 prevCounter = 3;
        uint256 target = 10;
        uint256 prevVersion = 0;
        State memory prevState = createCounterState(prevCounter, target, prevVersion);
        prevState.sigs = new Signature[](2);
        prevState.sigs[HOST_IDX] = signState(prevState, hostPrivateKey);
        prevState.sigs[GUEST_IDX] = signState(prevState, guestPrivateKey);

        // Candidate state: counter jumps by 2 (should be 4, not 5).
        uint256 newCounter = prevCounter + 2;
        uint256 newVersion = prevVersion + 1;
        State memory newState = createCounterState(newCounter, target, newVersion);
        newState.sigs = new Signature[](2);
        newState.sigs[HOST_IDX] = signState(newState, hostPrivateKey);
        newState.sigs[GUEST_IDX] = signState(newState, guestPrivateKey);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        bool valid = counterAdjudicator.adjudicate(channel, newState, proofs);
        assertFalse(valid, "State with incorrect counter increment should be rejected");
    }

    // Non-initial state with an incorrect version increment (not exactly +1) should fail.
    function test_InvalidVersionIncrement() public {
        // Previous state: version = 0.
        uint256 prevCounter = 3;
        uint256 target = 10;
        uint256 prevVersion = 0;
        State memory prevState = createCounterState(prevCounter, target, prevVersion);
        prevState.sigs = new Signature[](2);
        prevState.sigs[HOST_IDX] = signState(prevState, hostPrivateKey);
        prevState.sigs[GUEST_IDX] = signState(prevState, guestPrivateKey);

        // Candidate state: version jumps by 2 instead of 1.
        uint256 newCounter = prevCounter + 1;
        uint256 newVersion = prevVersion + 2;
        State memory newState = createCounterState(newCounter, target, newVersion);
        newState.sigs = new Signature[](2);
        newState.sigs[HOST_IDX] = signState(newState, hostPrivateKey);
        newState.sigs[GUEST_IDX] = signState(newState, guestPrivateKey);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        bool valid = counterAdjudicator.adjudicate(channel, newState, proofs);
        assertFalse(valid, "State with incorrect version increment should be rejected");
    }

    // Non-initial state where the target changes should fail.
    function test_InvalidTargetChange() public {
        // Previous state: target = 10.
        uint256 prevCounter = 3;
        uint256 target = 10;
        uint256 prevVersion = 0;
        State memory prevState = createCounterState(prevCounter, target, prevVersion);
        prevState.sigs = new Signature[](2);
        prevState.sigs[HOST_IDX] = signState(prevState, hostPrivateKey);
        prevState.sigs[GUEST_IDX] = signState(prevState, guestPrivateKey);

        // Candidate state: target changes (e.g., to 12) while counter and version increment correctly.
        uint256 newCounter = prevCounter + 1;
        uint256 newVersion = prevVersion + 1;
        State memory newState = createCounterState(newCounter, 12, newVersion);
        newState.sigs = new Signature[](2);
        newState.sigs[HOST_IDX] = signState(newState, hostPrivateKey);
        newState.sigs[GUEST_IDX] = signState(newState, guestPrivateKey);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        bool valid = counterAdjudicator.adjudicate(channel, newState, proofs);
        assertFalse(valid, "State with a changed target should be rejected");
    }

    // Non-initial state where the candidate counter exceeds the target should fail.
    function test_CandidateCounterExceedsTarget() public {
        // Previous state: counter equals the target.
        uint256 prevCounter = 10;
        uint256 target = 10;
        uint256 prevVersion = 0;
        State memory prevState = createCounterState(prevCounter, target, prevVersion);
        prevState.sigs = new Signature[](2);
        prevState.sigs[HOST_IDX] = signState(prevState, hostPrivateKey);
        prevState.sigs[GUEST_IDX] = signState(prevState, guestPrivateKey);

        // Candidate state: counter increments to 11, which exceeds target.
        uint256 newCounter = prevCounter + 1;
        uint256 newVersion = prevVersion + 1;
        State memory newState = createCounterState(newCounter, target, newVersion);
        newState.sigs = new Signature[](2);
        newState.sigs[HOST_IDX] = signState(newState, hostPrivateKey);
        newState.sigs[GUEST_IDX] = signState(newState, guestPrivateKey);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        bool valid = counterAdjudicator.adjudicate(channel, newState, proofs);
        assertFalse(valid, "State with counter exceeding target should be rejected");
    }
}
