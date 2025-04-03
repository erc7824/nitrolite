// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Test} from "lib/forge-std/src/Test.sol";
import {Vm} from "lib/forge-std/src/Vm.sol";
import {Custody} from "../src/Custody.sol";
import {IAdjudicator} from "../src/interfaces/IAdjudicator.sol";
import {IChannel} from "../src/interfaces/IChannel.sol";
import {IDeposit} from "../src/interfaces/IDeposit.sol";
import {Channel, State, Allocation, Signature} from "../src/interfaces/Types.sol";
import {Utils} from "../src/Utils.sol";
import {Counter} from "../src/adjudicators/Counter.sol";
import {MockERC20} from "./mocks/MockERC20.sol";

contract CustodyTest is Test {
    // Main contracts
    Custody public custody;
    Counter public adjudicator;

    // Test accounts
    address public host;
    address public guest;
    address public alice; // Additional user who is not a channel participant
    uint256 public hostPrivateKey;
    uint256 public guestPrivateKey;

    // Mock tokens
    MockERC20 public token;

    // Channel parameters
    Channel public channel;
    bytes32 public channelId;

    // Constants
    uint256 private constant HOST = 0;
    uint256 private constant GUEST = 1;
    uint64 private constant CHALLENGE_PERIOD = 3600; // 1 hour

    // Amounts
    uint256 private constant INITIAL_BALANCE = 1000 ether;
    uint256 private constant DEPOSIT_AMOUNT = 100 ether;

    function setUp() public {
        // Generate private keys and addresses for the participants
        hostPrivateKey = 0x1;
        guestPrivateKey = 0x2;
        host = vm.addr(hostPrivateKey);
        guest = vm.addr(guestPrivateKey);
        alice = address(0xABCD);

        // Deploy the contracts
        adjudicator = new Counter();
        custody = new Custody();

        // Deploy the mock token
        token = new MockERC20("Test Token", "TEST", 18);

        // Mint tokens for test accounts
        token.mint(host, INITIAL_BALANCE);
        token.mint(guest, INITIAL_BALANCE);
        token.mint(alice, INITIAL_BALANCE);

        // Set up the channel parameters
        address[2] memory participants = [host, guest];
        channel = Channel({
            participants: participants,
            adjudicator: address(adjudicator),
            challenge: CHALLENGE_PERIOD,
            nonce: 1
        });

        // Get the channelId
        channelId = Utils.getChannelId(channel);

        // Approve token transfers
        vm.startPrank(host);
        token.approve(address(custody), INITIAL_BALANCE);
        custody.deposit(address(token), DEPOSIT_AMOUNT);
        vm.stopPrank();

        vm.startPrank(guest);
        token.approve(address(custody), INITIAL_BALANCE);
        custody.deposit(address(token), DEPOSIT_AMOUNT);
        vm.stopPrank();

        vm.startPrank(alice);
        token.approve(address(custody), INITIAL_BALANCE);
        custody.deposit(address(token), DEPOSIT_AMOUNT);
        vm.stopPrank();
    }

    // ========================== Helper Functions ==========================

    function createAllocations(uint256 hostAmount, uint256 guestAmount) internal view returns (Allocation[2] memory) {
        Allocation[2] memory allocations;
        allocations[HOST] = Allocation({destination: host, token: address(token), amount: hostAmount});
        allocations[GUEST] = Allocation({destination: guest, token: address(token), amount: guestAmount});
        return allocations;
    }

    function createCounterState(uint256 counter, Allocation[2] memory allocations)
        internal
        pure
        returns (State memory)
    {
        State memory state;
        // Update the version based on counter value to ensure valid transitions
        uint256 version = counter > 0 ? counter - 1 : 0;
        
        Counter.CounterApp memory CounterApp = Counter.CounterApp({
            counter: counter,
            target: 10, // Default target
            version: version // Version should match counter-1 for validity
        });
        state.data = abi.encode(CounterApp);
        state.allocations = allocations;
        state.sigs = new Signature[](0); // Empty signatures to be filled later
        return state;
    }

    function signState(State memory state, uint256 privateKey) internal view returns (Signature memory) {
        bytes32 stateHash = Utils.getStateHash(channel, state);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, stateHash);
        return Signature({v: v, r: r, s: s});
    }

    function addSignatures(State memory state, bool includeHost, bool includeGuest)
        internal
        view
        returns (State memory)
    {
        uint8 count = 0;
        if (includeHost) count++;
        if (includeGuest) count++;

        state.sigs = new Signature[](count);

        uint8 index = 0;
        if (includeHost) {
            state.sigs[index] = signState(state, hostPrivateKey);
            index++;
        }

        if (includeGuest) {
            state.sigs[index] = signState(state, guestPrivateKey);
        }

        return state;
    }

    // ========================== 1. Channel Creation and Opening Tests ==========================

    function test_ChannelCreation() public {
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory initialState = createCounterState(0, allocations);
        initialState = addSignatures(initialState, true, false);

        vm.startPrank(host);
        bytes32 newChannelId = custody.open(channel, initialState);
        vm.stopPrank();

        assertEq(newChannelId, channelId, "Channel ID should match the calculated ID");

        // Verify channel is in PARTIAL status (by joining it)
        allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Verify channels are associated with participants
        bytes32[] memory hostChannels = custody.getAccountChannels(host);
        bytes32[] memory guestChannels = custody.getAccountChannels(guest);

        assertEq(hostChannels.length, 1, "Host should have 1 channel");
        assertEq(guestChannels.length, 1, "Guest should have 1 channel");
        assertEq(hostChannels[0], channelId, "Host's channel ID should match");
        assertEq(guestChannels[0], channelId, "Guest's channel ID should match");
    }

    function test_ChannelWithDifferentNonce() public {
        // Create a channel with a different nonce
        Channel memory channelWithNewNonce = channel;
        channelWithNewNonce.nonce = 2;

        bytes32 newChannelId = Utils.getChannelId(channelWithNewNonce);

        // Ensure channel IDs are different
        assertTrue(channelId != newChannelId, "Channel IDs should be different with different nonces");

        // Create the channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory initialState = createCounterState(0, allocations);
        initialState = addSignatures(initialState, true, false);

        vm.startPrank(host);
        bytes32 createdChannelId = custody.open(channelWithNewNonce, initialState);
        vm.stopPrank();

        assertEq(createdChannelId, newChannelId, "Created channel ID should match the calculated ID");
    }

    function test_PartialToActiveTransition() public {
        // First deposit from host creates PARTIAL channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory initialState = createCounterState(0, allocations);
        initialState = addSignatures(initialState, true, false);

        vm.startPrank(host);
        custody.open(channel, initialState);
        vm.stopPrank();

        // Second deposit from guest transitions to ACTIVE
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Verify funds are allocated correctly
        (uint256 hostAvailable, uint256 hostLocked,) = custody.getAccountInfo(host, address(token));
        (uint256 guestAvailable, uint256 guestLocked,) = custody.getAccountInfo(guest, address(token));

        assertEq(hostAvailable, DEPOSIT_AMOUNT - DEPOSIT_AMOUNT / 2, "Host available funds incorrect");
        assertEq(hostLocked, DEPOSIT_AMOUNT / 2, "Host locked funds incorrect");
        assertEq(guestAvailable, DEPOSIT_AMOUNT - DEPOSIT_AMOUNT / 2, "Guest available funds incorrect");
        assertEq(guestLocked, DEPOSIT_AMOUNT / 2, "Guest locked funds incorrect");
    }

    function test_OpenRevertsWithInvalidParticipants() public {
        // Create a channel with invalid participant count
        // Should revert if participants array length != 2
        address[2] memory participants = [host, address(0)];
        Channel memory invalidChannel = Channel({
            participants: participants,
            adjudicator: address(adjudicator),
            challenge: CHALLENGE_PERIOD,
            nonce: 1
        });

        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory state = createCounterState(0, allocations);
        state = addSignatures(state, true, false);

        vm.startPrank(host);
        vm.expectRevert(Custody.InvalidParticipant.selector);
        custody.open(invalidChannel, state);
        vm.stopPrank();
    }

    function test_OpenRevertsWithZeroAddressAdjudicator() public {
        // Create a channel with zero address adjudicator
        address[2] memory participants = [host, guest];
        Channel memory invalidChannel =
            Channel({participants: participants, adjudicator: address(0), challenge: CHALLENGE_PERIOD, nonce: 1});

        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory state = createCounterState(0, allocations);
        state = addSignatures(state, true, false);

        vm.startPrank(host);
        vm.expectRevert(Custody.InvalidAdjudicator.selector);
        custody.open(invalidChannel, state);
        vm.stopPrank();
    }

    function test_OpenRevertsWithZeroChallengePeriod() public {
        // Create a channel with zero challenge period
        address[2] memory participants = [host, guest];
        Channel memory invalidChannel =
            Channel({participants: participants, adjudicator: address(adjudicator), challenge: 0, nonce: 1});

        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory state = createCounterState(0, allocations);
        state = addSignatures(state, true, false);

        vm.startPrank(host);
        vm.expectRevert(Custody.InvalidChallengePeriod.selector);
        custody.open(invalidChannel, state);
        vm.stopPrank();
    }

    function test_OpenRevertsWithNonParticipant() public {
        // Adapted test for non-participant opening
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory state = createCounterState(0, allocations);
        state = addSignatures(state, true, false);

        // First, let the host open the channel
        vm.startPrank(host);
        custody.open(channel, state);
        vm.stopPrank();

        // Then have alice (not participant) try to join
        State memory aliceState = createCounterState(0, allocations);

        // Generate alice signature
        bytes32 stateHash = Utils.getStateHash(channel, aliceState);
        uint256 alicePrivateKey = uint256(keccak256(abi.encodePacked("alice")));
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(alicePrivateKey, stateHash);
        aliceState.sigs = new Signature[](1);
        aliceState.sigs[0] = Signature({v: v, r: r, s: s});

        vm.startPrank(alice);
        vm.expectRevert();
        custody.open(channel, aliceState);
        vm.stopPrank();
    }

    function test_OpenRevertsWithInsufficientFunds() public {
        // Try to deposit more than available funds
        Allocation[2] memory allocations = createAllocations(INITIAL_BALANCE + 1, DEPOSIT_AMOUNT / 2);
        State memory state = createCounterState(0, allocations);
        state = addSignatures(state, true, false);

        vm.startPrank(host);
        vm.expectRevert();
        custody.open(channel, state);
        vm.stopPrank();
    }

    // ========================== 2. Channel Closing Tests ==========================

    function test_CooperativeClose() public {
        // First set up a channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory hostState = createCounterState(0, allocations);
        hostState = addSignatures(hostState, true, false);

        vm.startPrank(host);
        custody.open(channel, hostState);
        vm.stopPrank();

        // Guest joins
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Create a final state with both signatures (version 1 for valid transition)
        Allocation[2] memory finalAllocations = createAllocations(DEPOSIT_AMOUNT / 4, (DEPOSIT_AMOUNT * 3) / 4);
        State memory finalState = createCounterState(1, finalAllocations);
        finalState = addSignatures(finalState, true, true);

        // We need to provide the previous state as proof for close to work properly
        State[] memory proofs = new State[](1);
        proofs[0] = hostState;

        // Close the channel with the final state
        vm.startPrank(host);
        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelClosed(channelId);
        custody.close(channelId, finalState, proofs);
        vm.stopPrank();

        // Verify funds are distributed correctly
        (uint256 hostAvailable, uint256 hostLocked,) = custody.getAccountInfo(host, address(token));
        (uint256 guestAvailable, uint256 guestLocked,) = custody.getAccountInfo(guest, address(token));

        assertEq(
            hostAvailable, DEPOSIT_AMOUNT - DEPOSIT_AMOUNT / 2 + DEPOSIT_AMOUNT / 4, "Host available funds incorrect"
        );
        assertEq(hostLocked, 0, "Host should have no locked funds");
        assertEq(
            guestAvailable,
            DEPOSIT_AMOUNT - DEPOSIT_AMOUNT / 2 + (DEPOSIT_AMOUNT * 3) / 4,
            "Guest available funds incorrect"
        );
        assertEq(guestLocked, 0, "Guest should have no locked funds");

        // Channel should no longer exist
        bytes32[] memory hostChannels = custody.getAccountChannels(host);
        assertEq(hostChannels.length, 0, "Host should have no channels after close");
    }

    function test_CloseRevertsWithNonFinalState() public {
        // First set up a channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory hostState = createCounterState(0, allocations);
        hostState = addSignatures(hostState, true, false);

        vm.startPrank(host);
        custody.open(channel, hostState);
        vm.stopPrank();

        // Guest joins
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Try to close with state that doesn't have both signatures
        State memory nonFinalState = createCounterState(1, allocations);
        nonFinalState = addSignatures(nonFinalState, true, false);

        // We need to provide the previous state as proof
        State[] memory proofs = new State[](1);
        proofs[0] = hostState;

        vm.startPrank(host);
        vm.expectRevert(Custody.ChannelNotFinal.selector);
        custody.close(channelId, nonFinalState, proofs);
        vm.stopPrank();
    }

    function test_CloseRevertsWithNonExistentChannel() public {
        // Try to close a non-existent channel
        bytes32 nonExistentChannelId = bytes32(uint256(123456));

        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory finalState = createCounterState(1, allocations);
        finalState = addSignatures(finalState, true, true);

        State[] memory proofs = new State[](0); // No proofs needed for non-existent channel test

        vm.startPrank(host);
        vm.expectRevert(abi.encodeWithSelector(Custody.ChannelNotFound.selector, nonExistentChannelId));
        custody.close(nonExistentChannelId, finalState, proofs);
        vm.stopPrank();
    }

    // ========================== 3. Challenge Mechanism Tests ==========================

    function test_Challenge() public {
        // First set up a channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory hostState = createCounterState(0, allocations);
        hostState = addSignatures(hostState, true, false);

        vm.startPrank(host);
        custody.open(channel, hostState);
        vm.stopPrank();

        // Guest joins
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Create a challengeable state (counter = 1, host signature)
        Allocation[2] memory challengeAllocations = createAllocations(DEPOSIT_AMOUNT / 3, (DEPOSIT_AMOUNT * 2) / 3);
        State memory challengeState = createCounterState(1, challengeAllocations);
        challengeState = addSignatures(challengeState, true, false);

        // Use initial state as proof
        State[] memory proofs = new State[](1);
        proofs[0] = hostState;

        // Submit challenge
        vm.startPrank(host);
        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelChallenged(channelId, block.timestamp + CHALLENGE_PERIOD);
        custody.challenge(channelId, challengeState, proofs);
        vm.stopPrank();

        // Challenge submitted successfully
    }

    function test_CounterChallenge() public {
        // First set up a channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory hostState = createCounterState(0, allocations);
        hostState = addSignatures(hostState, true, false);

        vm.startPrank(host);
        custody.open(channel, hostState);
        vm.stopPrank();

        // Guest joins
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Host challenges with state 1
        Allocation[2] memory hostChallengeAllocations = createAllocations(DEPOSIT_AMOUNT / 3, (DEPOSIT_AMOUNT * 2) / 3);
        State memory hostChallengeState = createCounterState(1, hostChallengeAllocations);
        hostChallengeState = addSignatures(hostChallengeState, true, false);

        State[] memory proofs = new State[](1);
        proofs[0] = hostState; // Use initial state (0) as proof for state 1

        vm.startPrank(host);
        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelChallenged(channelId, block.timestamp + CHALLENGE_PERIOD);
        custody.challenge(channelId, hostChallengeState, proofs);
        vm.stopPrank();

        // Guest counter-challenges with state 2
        Allocation[2] memory guestChallengeAllocations = createAllocations(DEPOSIT_AMOUNT / 4, (DEPOSIT_AMOUNT * 3) / 4);
        State memory guestChallengeState = createCounterState(2, guestChallengeAllocations);
        guestChallengeState = addSignatures(guestChallengeState, false, true);

        // Use host's challenge state as proof
        State[] memory counterProofs = new State[](1);
        counterProofs[0] = hostChallengeState;

        vm.startPrank(guest);
        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelChallenged(channelId, block.timestamp + CHALLENGE_PERIOD);
        custody.challenge(channelId, guestChallengeState, counterProofs);
        vm.stopPrank();
    }

    function test_ChallengeWithFinalState() public {
        // First set up a channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory hostState = createCounterState(0, allocations);
        hostState = addSignatures(hostState, true, false);

        vm.startPrank(host);
        custody.open(channel, hostState);
        vm.stopPrank();

        // Guest joins
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Challenge with state signed by both parties
        Allocation[2] memory finalAllocations = createAllocations(DEPOSIT_AMOUNT / 4, (DEPOSIT_AMOUNT * 3) / 4);
        State memory finalState = createCounterState(5, finalAllocations);
        finalState = addSignatures(finalState, true, true);

        // Previous state for proof
        State memory prevState = createCounterState(4, allocations);
        prevState = addSignatures(prevState, false, true);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        // Challenge with final state should close immediately
        vm.startPrank(host);
        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelClosed(channelId);
        custody.challenge(channelId, finalState, proofs);
        vm.stopPrank();

        // Channel should be closed and funds distributed
        (uint256 hostAvailable, uint256 hostLocked,) = custody.getAccountInfo(host, address(token));
        (uint256 guestAvailable, uint256 guestLocked,) = custody.getAccountInfo(guest, address(token));

        assertEq(
            hostAvailable, DEPOSIT_AMOUNT - DEPOSIT_AMOUNT / 2 + DEPOSIT_AMOUNT / 4, "Host available funds incorrect"
        );
        assertEq(hostLocked, 0, "Host should have no locked funds");
        assertEq(
            guestAvailable,
            DEPOSIT_AMOUNT - DEPOSIT_AMOUNT / 2 + (DEPOSIT_AMOUNT * 3) / 4,
            "Guest available funds incorrect"
        );
        assertEq(guestLocked, 0, "Guest should have no locked funds");
    }

    function test_ChallengeWithInvalidState() public {
        // First set up a channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory hostState = createCounterState(0, allocations);
        hostState = addSignatures(hostState, true, false);

        vm.startPrank(host);
        custody.open(channel, hostState);
        vm.stopPrank();

        // Guest joins
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Invalid state (counter jumps by 2)
        State memory invalidState = createCounterState(2, allocations);
        invalidState = addSignatures(invalidState, true, false);

        // Previous state for proof
        State memory prevState = createCounterState(0, allocations);
        prevState = addSignatures(prevState, false, true);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        // Challenge with invalid state should revert with InvalidState
        vm.startPrank(host);
        vm.expectRevert(Custody.InvalidState.selector);
        custody.challenge(channelId, invalidState, proofs);
        vm.stopPrank();
    }

    // ========================== 4. Checkpoint Tests ==========================

    function test_Checkpoint() public {
        // First set up a channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory hostState = createCounterState(0, allocations);
        hostState = addSignatures(hostState, true, false);

        vm.startPrank(host);
        custody.open(channel, hostState);
        vm.stopPrank();

        // Guest joins
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Create a valid state for checkpoint
        Allocation[2] memory checkpointAllocations = createAllocations(DEPOSIT_AMOUNT / 3, (DEPOSIT_AMOUNT * 2) / 3);
        State memory checkpointState = createCounterState(5, checkpointAllocations);
        checkpointState = addSignatures(checkpointState, true, true);

        // Previous state for proof
        State memory prevState = createCounterState(4, allocations);
        prevState = addSignatures(prevState, false, true);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        // Submit checkpoint
        vm.startPrank(host);
        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelCheckpointed(channelId);
        custody.checkpoint(channelId, checkpointState, proofs);
        vm.stopPrank();

        // Checkpoint was successful
    }

    function test_CheckpointWithFinalState() public {
        // First set up a channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory hostState = createCounterState(0, allocations);
        hostState = addSignatures(hostState, true, false);

        vm.startPrank(host);
        custody.open(channel, hostState);
        vm.stopPrank();

        // Guest joins
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Create a final state for checkpoint (using version 1 for valid transition)
        Allocation[2] memory finalAllocations = createAllocations(DEPOSIT_AMOUNT / 4, (DEPOSIT_AMOUNT * 3) / 4);
        State memory finalState = createCounterState(1, finalAllocations);
        finalState = addSignatures(finalState, true, true);

        // For test simplicity, we'll use an initial state as proof
        State[] memory proofs = new State[](1);
        proofs[0] = hostState;

        // Submit checkpoint with final state
        vm.startPrank(host);
        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelCheckpointed(channelId);
        custody.checkpoint(channelId, finalState, proofs);
        vm.stopPrank();

        // Checkpoint was successful
    }

    function test_CheckpointWithInvalidState() public {
        // First set up a channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory hostState = createCounterState(0, allocations);
        hostState = addSignatures(hostState, true, false);

        vm.startPrank(host);
        custody.open(channel, hostState);
        vm.stopPrank();

        // Guest joins
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Invalid state (counter jumps by 2)
        State memory invalidState = createCounterState(2, allocations);
        invalidState = addSignatures(invalidState, true, false);

        // Previous state for proof
        State memory prevState = createCounterState(0, allocations);
        prevState = addSignatures(prevState, false, true);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        // Checkpoint with invalid state should not revert anymore
        vm.startPrank(host);
        vm.expectRevert(Custody.InvalidState.selector);
        custody.checkpoint(channelId, invalidState, proofs);
        vm.stopPrank();
    }

    // ========================== 5. Reclaim Tests ==========================

    function test_Reclaim() public {
        // First set up a channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory hostState = createCounterState(0, allocations);
        hostState = addSignatures(hostState, true, false);

        vm.startPrank(host);
        custody.open(channel, hostState);
        vm.stopPrank();

        // Guest joins
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Host challenges with an intermediate state
        Allocation[2] memory challengeAllocations = createAllocations(DEPOSIT_AMOUNT / 3, (DEPOSIT_AMOUNT * 2) / 3);
        State memory challengeState = createCounterState(1, challengeAllocations);
        challengeState = addSignatures(challengeState, true, false);

        State[] memory proofs = new State[](1);
        proofs[0] = hostState; // Use initial state (0) as proof for state 1

        vm.startPrank(host);
        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelChallenged(channelId, block.timestamp + CHALLENGE_PERIOD);
        custody.challenge(channelId, challengeState, proofs);
        vm.stopPrank();

        // Warp time past challenge period
        vm.warp(block.timestamp + CHALLENGE_PERIOD + 1);

        // Reclaim funds
        vm.startPrank(host);
        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelClosed(channelId);
        custody.reclaim(channelId);
        vm.stopPrank();

        // Verify funds are distributed correctly
        (uint256 hostAvailable, uint256 hostLocked,) = custody.getAccountInfo(host, address(token));
        (uint256 guestAvailable, uint256 guestLocked,) = custody.getAccountInfo(guest, address(token));

        assertEq(
            hostAvailable, DEPOSIT_AMOUNT - DEPOSIT_AMOUNT / 2 + DEPOSIT_AMOUNT / 3, "Host available funds incorrect"
        );
        assertEq(hostLocked, 0, "Host should have no locked funds");
        assertEq(
            guestAvailable,
            DEPOSIT_AMOUNT - DEPOSIT_AMOUNT / 2 + (DEPOSIT_AMOUNT * 2) / 3,
            "Guest available funds incorrect"
        );
        assertEq(guestLocked, 0, "Guest should have no locked funds");

        // Channel should no longer exist
        bytes32[] memory hostChannels = custody.getAccountChannels(host);
        assertEq(hostChannels.length, 0, "Host should have no channels after reclaim");
    }

    function test_ReclaimRevertsBeforeChallengeExpiry() public {
        // First set up a channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory hostState = createCounterState(0, allocations);
        hostState = addSignatures(hostState, true, false);

        vm.startPrank(host);
        custody.open(channel, hostState);
        vm.stopPrank();

        // Guest joins
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Host challenges with an intermediate state
        Allocation[2] memory challengeAllocations = createAllocations(DEPOSIT_AMOUNT / 3, (DEPOSIT_AMOUNT * 2) / 3);
        State memory challengeState = createCounterState(1, challengeAllocations);
        challengeState = addSignatures(challengeState, true, false);

        State[] memory proofs = new State[](1);
        proofs[0] = hostState; // Use initial state (0) as proof for state 1

        vm.startPrank(host);
        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelChallenged(channelId, block.timestamp + CHALLENGE_PERIOD);
        custody.challenge(channelId, challengeState, proofs);
        vm.stopPrank();

        // Warp time but not past challenge period
        vm.warp(block.timestamp + CHALLENGE_PERIOD / 2);

        // Try to reclaim before challenge expires
        vm.startPrank(host);
        vm.expectRevert(Custody.ChallengeNotExpired.selector);
        custody.reclaim(channelId);
        vm.stopPrank();
    }

    // ========================== 6. Fund Management Tests ==========================

    function test_DepositAndWithdraw() public {
        // Verify initial deposit
        (uint256 hostAvailable, uint256 hostLocked,) = custody.getAccountInfo(host, address(token));
        assertEq(hostAvailable, DEPOSIT_AMOUNT, "Initial deposit amount incorrect");
        assertEq(hostLocked, 0, "Should have no locked funds initially");

        // Withdraw half
        uint256 withdrawAmount = DEPOSIT_AMOUNT / 2;
        vm.startPrank(host);
        custody.withdraw(address(token), withdrawAmount);
        vm.stopPrank();

        // Verify after withdrawal
        (hostAvailable, hostLocked,) = custody.getAccountInfo(host, address(token));
        assertEq(hostAvailable, DEPOSIT_AMOUNT - withdrawAmount, "Available amount after withdrawal incorrect");

        // Deposit more
        uint256 additionalDeposit = 50 ether;
        vm.startPrank(host);
        token.approve(address(custody), additionalDeposit);
        custody.deposit(address(token), additionalDeposit);
        vm.stopPrank();

        // Verify after additional deposit
        (hostAvailable, hostLocked,) = custody.getAccountInfo(host, address(token));
        assertEq(
            hostAvailable,
            DEPOSIT_AMOUNT - withdrawAmount + additionalDeposit,
            "Available amount after additional deposit incorrect"
        );
    }

    function test_WithdrawRevertsWithInsufficientFunds() public {
        // Try to withdraw more than available
        vm.startPrank(host);
        vm.expectRevert(
            abi.encodeWithSelector(Custody.InsufficientBalance.selector, DEPOSIT_AMOUNT, DEPOSIT_AMOUNT + 1)
        );
        custody.withdraw(address(token), DEPOSIT_AMOUNT + 1);
        vm.stopPrank();
    }

    function test_FailedTransfer() public {
        // Setup token to fail transfers
        vm.startPrank(host);
        token.approve(address(custody), DEPOSIT_AMOUNT);
        custody.deposit(address(token), DEPOSIT_AMOUNT);
        vm.stopPrank();

        // Make token transfers fail
        token.setFailTransfers(true);

        // Try to withdraw - should revert
        vm.startPrank(host);
        vm.expectRevert();
        custody.withdraw(address(token), 1);
        vm.stopPrank();
    }

    // ========================== 7. Reset Channel Tests ==========================

    function test_ResetChannel() public {
        // First set up a channel
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory hostState = createCounterState(0, allocations);
        hostState = addSignatures(hostState, true, false);

        vm.startPrank(host);
        custody.open(channel, hostState);
        vm.stopPrank();

        // Guest joins
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Create final state for closing
        Allocation[2] memory finalAllocations = createAllocations(DEPOSIT_AMOUNT / 4, (DEPOSIT_AMOUNT * 3) / 4);
        State memory finalState = createCounterState(1000, finalAllocations);
        finalState = addSignatures(finalState, true, true);

        // Create new channel config with different nonce
        Channel memory newChannel = channel;
        newChannel.nonce = 2;
        bytes32 newChannelId = Utils.getChannelId(newChannel);

        // Create new deposit allocations
        Allocation[2] memory newAllocations = createAllocations(DEPOSIT_AMOUNT / 3, (DEPOSIT_AMOUNT * 2) / 3);
        State memory newState = createCounterState(0, newAllocations);
        newState = addSignatures(newState, true, false);

        // We need state proofs for reset to work
        State[] memory proofs = new State[](1);
        proofs[0] = hostState; // Use initial state (0) as proof for state 1000

        // Reset channel
        vm.startPrank(host);
        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelClosed(channelId);
        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelPartiallyFunded(newChannelId, newChannel);
        custody.reset(channelId, finalState, proofs, newChannel, newState);
        vm.stopPrank();

        // Verify old channel is closed
        bytes32[] memory hostChannels = custody.getAccountChannels(host);
        bool foundOldChannel = false;
        bool foundNewChannel = false;

        for (uint256 i = 0; i < hostChannels.length; i++) {
            if (hostChannels[i] == channelId) foundOldChannel = true;
            if (hostChannels[i] == newChannelId) foundNewChannel = true;
        }

        assertFalse(foundOldChannel, "Old channel should be closed");
        assertTrue(foundNewChannel, "New channel should be opened");

        // Have guest join new channel
        State memory newGuestState = createCounterState(0, newAllocations);
        newGuestState = addSignatures(newGuestState, false, true);

        vm.startPrank(guest);
        custody.open(newChannel, newGuestState);
        vm.stopPrank();

        // Verify channel status through fund allocation
        (uint256 hostAvailable, uint256 hostLocked,) = custody.getAccountInfo(host, address(token));
        (uint256 guestAvailable, uint256 guestLocked,) = custody.getAccountInfo(guest, address(token));

        // Verify locked amounts for new channel
        assertEq(hostLocked, DEPOSIT_AMOUNT / 3, "Host locked amount incorrect for new channel");
        assertEq(guestLocked, (DEPOSIT_AMOUNT * 2) / 3, "Guest locked amount incorrect for new channel");
    }

    // ========================== 8. Event Tests ==========================

    function test_EventEmission() public {
        // Test ChannelPartiallyFunded event
        Allocation[2] memory allocations = createAllocations(DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT / 2);
        State memory hostState = createCounterState(0, allocations);
        hostState = addSignatures(hostState, true, false);

        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelPartiallyFunded(channelId, channel);

        vm.startPrank(host);
        custody.open(channel, hostState);
        vm.stopPrank();

        // Test ChannelOpened event
        State memory guestState = createCounterState(0, allocations);
        guestState = addSignatures(guestState, false, true);

        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelOpened(channelId, channel);

        vm.startPrank(guest);
        custody.open(channel, guestState);
        vm.stopPrank();

        // Test ChannelChallenged event - using valid state transitions this time
        State memory challengeState = createCounterState(1, allocations);
        challengeState = addSignatures(challengeState, true, false);

        State[] memory proofs = new State[](1);
        proofs[0] = hostState; // Use initial state as proof

        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelChallenged(channelId, block.timestamp + CHALLENGE_PERIOD);

        vm.startPrank(host);
        custody.challenge(channelId, challengeState, proofs);
        vm.stopPrank();

        // Test ChannelClosed event from reclaim
        vm.warp(block.timestamp + CHALLENGE_PERIOD + 1);

        vm.expectEmit(true, false, false, false);
        emit IChannel.ChannelClosed(channelId);

        vm.startPrank(host);
        custody.reclaim(channelId);
        vm.stopPrank();
    }
}
