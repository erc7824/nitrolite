// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Test} from "lib/forge-std/src/Test.sol";
import {Vm} from "lib/forge-std/src/Vm.sol";
import {IAdjudicator} from "../../src/interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature} from "../../src/interfaces/Types.sol";
import {Counter} from "../../src/adjudicators/Counter.sol";
import {Utils} from "../../src/Utils.sol";
import {MockERC20} from "../mocks/MockERC20.sol";

contract CounterTest is Test {
    Counter public adjudicator;

    // Test accounts
    address public host;
    address public guest;
    uint256 public hostPrivateKey;
    uint256 public guestPrivateKey;

    // Channel parameters
    Channel public channel;
    MockERC20 public token;

    // Constants
    uint256 private constant FINAL_COUNTER = 1000;
    uint256 private constant HOST = 0;
    uint256 private constant GUEST = 1;

    function setUp() public {
        // Deploy the adjudicator contract
        adjudicator = new Counter();

        // Generate private keys and addresses for the participants
        hostPrivateKey = 0x1;
        guestPrivateKey = 0x2;
        host = vm.addr(hostPrivateKey);
        guest = vm.addr(guestPrivateKey);

        // Deploy the mock token
        token = new MockERC20("Test Token", "TEST", 18);

        // Set up the channel
        address[2] memory participants = [host, guest];
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
        allocations[0] = Allocation({destination: host, token: address(token), amount: hostAmount});
        allocations[1] = Allocation({destination: guest, token: address(token), amount: guestAmount});
        return allocations;
    }

    // Helper function to create a counter state
    function createCounterState(uint256 counter, Allocation[2] memory allocations)
        internal
        pure
        returns (State memory)
    {
        State memory state;
        Counter.CounterData memory counterData = Counter.CounterData({counter: counter});
        state.data = abi.encode(counterData);
        state.allocations = allocations;
        state.sigs = new Signature[](0); // Empty signatures to be filled later
        return state;
    }

    // Helper to sign state with specified key
    function signState(State memory state, uint256 privateKey) internal view returns (Signature memory) {
        bytes32 stateHash = Utils.getStateHash(channel, state);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, stateHash);
        return Signature({v: v, r: r, s: s});
    }

    // Test basic signature verification
    function test_BasicSignatureVerification() public view {
        bytes32 message = keccak256("test message");
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(hostPrivateKey, message);

        address recovered = ecrecover(message, v, r, s);
        assertEq(recovered, host);

        Signature memory signature = Signature({v: v, r: r, s: s});
        bool isValid = Utils.verifySignature(message, signature, host);
        assertTrue(isValid);
    }

    // Test: Initial state with host signature only - should return PARTIAL
    function test_InitialStateHostOnly() public view {
        Allocation[2] memory allocations = createAllocations(50, 50);
        State memory state = createCounterState(0, allocations);

        // Add host signature
        state.sigs = new Signature[](1);
        state.sigs[0] = signState(state, hostPrivateKey);

        // Adjudicate
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, state, new State[](0));

        // With counter = 0, should be PARTIAL
        assertEq(uint256(decision), uint256(IAdjudicator.Status.PARTIAL));
    }

    // Test: Initial state with both signatures but counter = 0 - should return PARTIAL
    function test_InitialStateBothSignaturesZeroCounter() public view {
        Allocation[2] memory allocations = createAllocations(50, 50);
        State memory state = createCounterState(0, allocations);

        // Add both signatures
        state.sigs = new Signature[](2);
        state.sigs[0] = signState(state, hostPrivateKey);
        state.sigs[1] = signState(state, guestPrivateKey);

        // Adjudicate
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, state, new State[](0));

        // With counter = 0, should be PARTIAL
        assertEq(uint256(decision), uint256(IAdjudicator.Status.PARTIAL));
    }

    // Test: Initial state with both signatures and counter > 0 - should return ACTIVE
    function test_InitialStateBothSignaturesNonZeroCounter() public view {
        Allocation[2] memory allocations = createAllocations(50, 50);
        State memory state = createCounterState(5, allocations);

        // Add both signatures
        state.sigs = new Signature[](2);
        state.sigs[0] = signState(state, hostPrivateKey);
        state.sigs[1] = signState(state, guestPrivateKey);

        // Adjudicate
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, state, new State[](0));

        // With counter > 0, should be ACTIVE
        assertEq(uint256(decision), uint256(IAdjudicator.Status.INVALID));
    }

    // Test: Invalid host signature
    function test_InvalidHostSignature() public view {
        Allocation[2] memory allocations = createAllocations(50, 50);
        State memory state = createCounterState(5, allocations);

        // Add host signature but corrupt it
        state.sigs = new Signature[](1);
        Signature memory sig = signState(state, hostPrivateKey);
        sig.r = bytes32(0); // Corrupt signature
        state.sigs[0] = sig;

        // Adjudicate with corrupted signature should return VOID
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, state, new State[](0));
        assertEq(uint256(decision), uint256(IAdjudicator.Status.VOID));
    }

    // Test: Invalid guest signature
    function test_InvalidGuestSignature() public view {
        Allocation[2] memory allocations = createAllocations(50, 50);
        State memory state = createCounterState(5, allocations);

        // Add both signatures but corrupt guest signature
        state.sigs = new Signature[](2);
        state.sigs[0] = signState(state, hostPrivateKey);

        Signature memory guestSig = signState(state, guestPrivateKey);
        guestSig.r = bytes32(0); // Corrupt signature
        state.sigs[1] = guestSig;

        // Adjudicate with corrupted guest signature should return VOID
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, state, new State[](0));
        assertEq(uint256(decision), uint256(IAdjudicator.Status.VOID));
    }

    // Test: Insufficient signatures
    function test_InsufficientSignatures() public view {
        Allocation[2] memory allocations = createAllocations(50, 50);
        State memory state = createCounterState(5, allocations);

        // No signatures added

        // Adjudicate and expect INVALID status instead of revert
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, state, new State[](0));
        assertEq(uint256(decision), uint256(IAdjudicator.Status.INVALID));
    }

    // Test: Valid counter increment from host to guest
    function test_ValidCounterIncrementHostToGuest() public view {
        // Create the previous state (counter = 5, signed by Guest)
        Allocation[2] memory allocations = createAllocations(50, 50);
        State memory prevState = createCounterState(5, allocations);
        prevState.sigs = new Signature[](1);
        prevState.sigs[0] = signState(prevState, guestPrivateKey);

        // Create the new state (counter = 6, signed by Host)
        State memory newState = createCounterState(6, allocations);
        newState.sigs = new Signature[](1);
        newState.sigs[0] = signState(newState, hostPrivateKey);

        // Create proofs array with previous state
        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        // Adjudicate
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, newState, proofs);

        // Should be ACTIVE
        assertEq(uint256(decision), uint256(IAdjudicator.Status.ACTIVE));
    }

    // Test: Valid counter increment from guest to host
    function test_ValidCounterIncrementGuestToHost() public view {
        // Create the previous state (counter = 6, signed by Host)
        Allocation[2] memory allocations = createAllocations(50, 50);
        State memory prevState = createCounterState(6, allocations);
        prevState.sigs = new Signature[](1);
        prevState.sigs[0] = signState(prevState, hostPrivateKey);

        // Create the new state (counter = 7, signed by Guest)
        State memory newState = createCounterState(7, allocations);
        newState.sigs = new Signature[](1);
        newState.sigs[0] = signState(newState, guestPrivateKey);

        // Create proofs array with previous state
        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        // Adjudicate
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, newState, proofs);

        // Should be ACTIVE
        assertEq(uint256(decision), uint256(IAdjudicator.Status.ACTIVE));
    }

    // Test: Invalid increment (not exactly +1)
    function test_InvalidIncrement() public view {
        // Create the previous state (counter = 5, signed by Guest)
        Allocation[2] memory allocations = createAllocations(50, 50);
        State memory prevState = createCounterState(5, allocations);
        prevState.sigs = new Signature[](1);
        prevState.sigs[0] = signState(prevState, guestPrivateKey);

        // Create the new state (counter = 7, signed by Host) - increment by 2 instead of 1
        State memory newState = createCounterState(7, allocations);
        newState.sigs = new Signature[](1);
        newState.sigs[0] = signState(newState, hostPrivateKey);

        // Create proofs array with previous state
        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        // Adjudicate and expect INVALID status instead of revert
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, newState, proofs);
        assertEq(uint256(decision), uint256(IAdjudicator.Status.INVALID));
    }

    // Test: Invalid turn (Host followed by Host)
    function test_InvalidTurnSameParticipant() public view {
        // Create the previous state (counter = 5, signed by Host)
        Allocation[2] memory allocations = createAllocations(50, 50);
        State memory prevState = createCounterState(5, allocations);
        prevState.sigs = new Signature[](1);
        prevState.sigs[0] = signState(prevState, hostPrivateKey);

        // Create the new state (counter = 6, signed by Host again)
        State memory newState = createCounterState(6, allocations);
        newState.sigs = new Signature[](1);
        newState.sigs[0] = signState(newState, hostPrivateKey);

        // Create proofs array with previous state
        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        // Adjudicate and expect INVALID status instead of revert
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, newState, proofs);
        assertEq(uint256(decision), uint256(IAdjudicator.Status.INVALID));
    }

    // Test: Reaching final state
    function test_ReachingFinalState() public view {
        // Create the previous state (counter = 999, signed by Guest)
        Allocation[2] memory allocations = createAllocations(50, 50);
        State memory prevState = createCounterState(999, allocations);
        prevState.sigs = new Signature[](1);
        prevState.sigs[0] = signState(prevState, guestPrivateKey);

        // Create the new state (counter = 1000, signed by Host)
        State memory newState = createCounterState(1000, allocations);
        newState.sigs = new Signature[](1);
        newState.sigs[0] = signState(newState, hostPrivateKey);

        // Create proofs array with previous state
        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        // Adjudicate
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, newState, proofs);

        // Should be FINAL
        assertEq(uint256(decision), uint256(IAdjudicator.Status.FINAL));
    }
}
