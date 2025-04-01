// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Test} from "lib/forge-std/src/Test.sol";
import {Vm} from "lib/forge-std/src/Vm.sol";
import {IAdjudicator} from "../../src/interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature} from "../../src/interfaces/Types.sol";
import {MicroPayment} from "../../src/adjudicators/MicroPayment.sol";
import {Utils} from "../../src/Utils.sol";
import {MockERC20} from "../mocks/MockERC20.sol";

contract MicroPaymentTest is Test {
    MicroPayment public adjudicator;

    // Test accounts
    address public host;
    address public guest;
    uint256 public hostPrivateKey;
    uint256 public guestPrivateKey;

    // Channel parameters
    Channel public channel;
    MockERC20 public token;

    // Constants
    uint256 private constant HOST = 0;
    uint256 private constant GUEST = 1;

    function setUp() public {
        // Deploy the adjudicator contract
        adjudicator = new MicroPayment();

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

    // Helper function to create a payment state
    function createPaymentState(uint256 paymentAmount, Allocation[2] memory allocations)
        internal
        pure
        returns (State memory)
    {
        State memory state;
        state.data = abi.encode(paymentAmount);
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
    function test_BasicSignatureVerification() public {
        bytes32 message = keccak256("test message");
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(hostPrivateKey, message);

        address recovered = ecrecover(message, v, r, s);
        assertEq(recovered, host);

        Signature memory signature = Signature({v: v, r: r, s: s});
        bool isValid = Utils.verifySignature(message, signature, host);
        assertTrue(isValid);
    }

    // Test: Valid payment state, host can send payment to guest
    function test_ValidPaymentState() public {
        // Initial allocation: host has 100 tokens, guest has 0
        Allocation[2] memory allocations = createAllocations(100, 0);

        // Create a state with 30 tokens payment to guest
        State memory state = createPaymentState(30, allocations);

        // Add host signature
        state.sigs = new Signature[](1);
        state.sigs[0] = signState(state, hostPrivateKey);

        // Adjudicate
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, state, new State[](0));

        // Check the status is ACTIVE
        assertEq(uint256(decision), uint256(IAdjudicator.Status.ACTIVE));

        // With the updated interface, we no longer get allocations returned
        // The adjudicator still processes them internally, but we just verify the Status
    }

    // Test: Guest trying to sign a payment is invalid
    function test_GuestSignatureInvalid() public {
        // Initial allocation: host has 100 tokens, guest has 0
        Allocation[2] memory allocations = createAllocations(100, 0);

        // Create a state with 30 tokens payment to guest
        State memory state = createPaymentState(30, allocations);

        // Add guest signature (this should fail)
        state.sigs = new Signature[](1);
        state.sigs[0] = signState(state, guestPrivateKey);

        // Adjudicate and expect INVALID status instead of revert
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, state, new State[](0));
        assertEq(uint256(decision), uint256(IAdjudicator.Status.INVALID));
    }

    // Test: Missing signature
    function test_MissingSignature() public {
        // Initial allocation: host has 100 tokens, guest has 0
        Allocation[2] memory allocations = createAllocations(100, 0);

        // Create a state with 30 tokens payment to guest
        State memory state = createPaymentState(30, allocations);

        // No signatures added

        // Adjudicate and expect INVALID status instead of revert
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, state, new State[](0));
        assertEq(uint256(decision), uint256(IAdjudicator.Status.INVALID));
    }

    // Test: Invalid host signature (corrupted)
    function test_InvalidHostSignature() public {
        // Initial allocation: host has 100 tokens, guest has 0
        Allocation[2] memory allocations = createAllocations(100, 0);

        // Create a state with 30 tokens payment to guest
        State memory state = createPaymentState(30, allocations);

        // Add corrupted host signature
        state.sigs = new Signature[](1);
        Signature memory sig = signState(state, hostPrivateKey);
        sig.r = bytes32(0); // Corrupt the signature
        state.sigs[0] = sig;

        // Adjudicate and expect INVALID status instead of revert
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, state, new State[](0));
        assertEq(uint256(decision), uint256(IAdjudicator.Status.INVALID));
    }

    // Test: Payment exceeds deposit
    function test_PaymentExceedsDeposit() public {
        // Initial allocation: host has 100 tokens, guest has 0
        Allocation[2] memory allocations = createAllocations(100, 0);

        // Create a state with 150 tokens payment (exceeds the 100 token deposit)
        State memory state = createPaymentState(150, allocations);

        // Add host signature
        state.sigs = new Signature[](1);
        state.sigs[0] = signState(state, hostPrivateKey);

        // Adjudicate and expect INVALID status instead of revert
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, state, new State[](0));
        assertEq(uint256(decision), uint256(IAdjudicator.Status.INVALID));
    }

    // Test: Decreasing payment amount
    function test_DecreasingPayment() public {
        // Initial allocation: host has 100 tokens, guest has 0
        Allocation[2] memory allocations = createAllocations(100, 0);

        // Create a previous state with 50 tokens payment
        State memory prevState = createPaymentState(50, allocations);
        prevState.sigs = new Signature[](1);
        prevState.sigs[0] = signState(prevState, hostPrivateKey);

        // Create a new state with 30 tokens payment (decreasing)
        State memory newState = createPaymentState(30, allocations);
        newState.sigs = new Signature[](1);
        newState.sigs[0] = signState(newState, hostPrivateKey);

        // Create proofs array with previous state
        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        // Adjudicate and expect INVALID status instead of revert
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, newState, proofs);
        assertEq(uint256(decision), uint256(IAdjudicator.Status.INVALID));
    }

    // Test: Increasing payment amount
    function test_IncreasingPayment() public {
        // Initial allocation: host has 100 tokens, guest has 0
        Allocation[2] memory allocations = createAllocations(100, 0);

        // Create a previous state with 30 tokens payment
        State memory prevState = createPaymentState(30, allocations);
        prevState.sigs = new Signature[](1);
        prevState.sigs[0] = signState(prevState, hostPrivateKey);

        // Create a new state with 50 tokens payment (increasing)
        State memory newState = createPaymentState(50, allocations);
        newState.sigs = new Signature[](1);
        newState.sigs[0] = signState(newState, hostPrivateKey);

        // Create proofs array with previous state
        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        // Adjudicate
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, newState, proofs);

        // Check the status is ACTIVE
        assertEq(uint256(decision), uint256(IAdjudicator.Status.ACTIVE));

        // Check the allocations have been updated properly
        // With the updated interface, we no longer get allocations returned
        // The adjudicator still processes them internally, but we just verify the Status
    }

    // Test: Final payment (full amount)
    function test_FinalPayment() public {
        // Initial allocation: host has 100 tokens, guest has 0
        Allocation[2] memory allocations = createAllocations(100, 0);

        // Create a state with full 100 tokens payment
        State memory state = createPaymentState(100, allocations);

        // Add host signature
        state.sigs = new Signature[](1);
        state.sigs[0] = signState(state, hostPrivateKey);

        // Adjudicate
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, state, new State[](0));

        // Check the status is FINAL
        assertEq(uint256(decision), uint256(IAdjudicator.Status.FINAL));
    }

    // Test: Invalid signature in proof state
    function test_InvalidProofSignature() public {
        // Initial allocation: host has 100 tokens, guest has 0
        Allocation[2] memory allocations = createAllocations(100, 0);

        // Create a previous state with 30 tokens payment but corrupt signature
        State memory prevState = createPaymentState(30, allocations);
        prevState.sigs = new Signature[](1);
        Signature memory prevSig = signState(prevState, hostPrivateKey);
        prevSig.r = bytes32(0); // Corrupt the signature
        prevState.sigs[0] = prevSig;

        // Create a new state with 50 tokens payment
        State memory newState = createPaymentState(50, allocations);
        newState.sigs = new Signature[](1);
        newState.sigs[0] = signState(newState, hostPrivateKey);

        // Create proofs array with corrupted previous state
        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        // Adjudicate and expect INVALID status instead of revert
        IAdjudicator.Status decision = adjudicator.adjudicate(channel, newState, proofs);
        assertEq(uint256(decision), uint256(IAdjudicator.Status.INVALID));
    }
}
