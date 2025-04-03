// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Test, console} from "lib/forge-std/src/Test.sol";
import {Custody} from "../src/Custody.sol";
import {Counter} from "../src/adjudicators/Counter.sol";
import {IERC20} from "lib/openzeppelin-contracts/contracts/interfaces/IERC20.sol";
import {Channel, State, Allocation, Signature, Status} from "../src/interfaces/Types.sol";
import {Utils} from "../src/Utils.sol";
import {MockERC20} from "./mocks/MockERC20.sol";

contract CustodyTest is Test {
    Custody public custody;
    Counter public adjudicator;
    MockERC20 public token;

    // Private keys for testing
    uint256 constant hostPrivKey = 1;
    uint256 constant guestPrivKey = 2;
    uint256 constant nonParticipantPrivKey = 3;

    // Test users
    address public host;
    address public guest;
    address public nonParticipant;

    // Common test values
    uint64 constant CHALLENGE_DURATION = 3600; // 1 hour
    uint64 constant NONCE = 1;
    uint256 constant DEPOSIT_AMOUNT = 1000;
    uint256 constant INITIAL_BALANCE = 10000;

    // For Counter adjudicator
    uint256 constant TARGET = 10;

    function setUp() public {
        // Set up user addresses from private keys
        host = vm.addr(hostPrivKey);
        guest = vm.addr(guestPrivKey);
        nonParticipant = vm.addr(nonParticipantPrivKey);

        // Deploy contracts
        custody = new Custody();
        adjudicator = new Counter();
        token = new MockERC20("Test Token", "TST", 18);

        // Fund accounts
        token.mint(host, INITIAL_BALANCE);
        token.mint(guest, INITIAL_BALANCE);
        token.mint(nonParticipant, INITIAL_BALANCE);

        // Approve token transfers
        vm.startPrank(host);
        token.approve(address(custody), INITIAL_BALANCE);
        vm.stopPrank();

        vm.startPrank(guest);
        token.approve(address(custody), INITIAL_BALANCE);
        vm.stopPrank();

        vm.startPrank(nonParticipant);
        token.approve(address(custody), INITIAL_BALANCE);
        vm.stopPrank();
    }

    // Helper to create a standard test channel
    function createTestChannel() internal view returns (Channel memory) {
        address[2] memory participants = [host, guest];
        return Channel({
            participants: participants,
            adjudicator: address(adjudicator),
            challenge: CHALLENGE_DURATION,
            nonce: NONCE
        });
    }

    // Helper to create an initial state for Counter adjudicator
    function createInitialState(address[2] memory participants, uint256 counter) internal view returns (State memory) {
        // Create allocations for both participants
        Allocation memory hostAllocation =
            Allocation({destination: participants[0], token: address(token), amount: DEPOSIT_AMOUNT});

        Allocation memory guestAllocation =
            Allocation({destination: participants[1], token: address(token), amount: DEPOSIT_AMOUNT});

        Allocation[2] memory allocations = [hostAllocation, guestAllocation];

        // Create the CounterApp struct and encode it
        Counter.CounterApp memory counterApp = Counter.CounterApp({counter: counter, target: TARGET, version: 0});
        bytes memory data = abi.encode(counterApp);

        // Create unsigned state
        State memory state = State({
            data: data,
            allocations: allocations,
            sigs: new Signature[](0) // Empty initially
        });

        return state;
    }

    // Helper to sign a state
    function signState(Channel memory chan, State memory state, uint256 privateKey)
        internal
        view
        returns (Signature memory)
    {
        bytes32 stateHash = Utils.getStateHash(chan, state);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, stateHash);
        return Signature({v: v, r: r, s: s});
    }

    // Helper to deposit tokens using prank instead of startPrank
    function depositTokens(address user, uint256 amount) internal {
        vm.prank(user);
        custody.deposit(address(token), amount);
    }

    // Helper to skip time for challenge testing
    function skipChallengeTime() internal {
        skip(CHALLENGE_DURATION + 1);
    }

    // ==================== TEST CASES ====================

    // ==== 1. Channel Creation and Opening ====

    function test_ChannelCreation() public {
        // 1. Prepare channel and initial state
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState([host, guest], 1);

        // 2. Sign the state by the host
        vm.deal(host, 1 ether); // Ensure host has ETH for gas

        // Sign the state
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature[] memory sigs = new Signature[](1);
        sigs[0] = hostSig;
        initialState.sigs = sigs;

        // 3. Deposit tokens for the host - using direct prank instead of helper that uses startPrank
        vm.prank(host);
        custody.deposit(address(token), DEPOSIT_AMOUNT * 2);

        // 4. Open the channel (partial funding by host)
        vm.prank(host);
        bytes32 channelId = custody.open(chan, initialState);

        // Verify the channel is created and in PARTIAL state
        (uint256 available, uint256 locked, uint256 channelCount) = custody.getAccountInfo(host, address(token));
        assertEq(locked, DEPOSIT_AMOUNT, "Host's tokens not locked correctly");
        assertEq(channelCount, 1, "Host should have 1 channel");

        // Also check that the channelId is consistent
        bytes32 expectedChannelId = Utils.getChannelId(chan);
        assertEq(channelId, expectedChannelId, "Channel ID is incorrect");
    }

    function test_CompleteChannelOpening() public {
        // 1. Create channel with host
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState([host, guest], 1);

        // 2. Sign the state by both participants
        // Host signs
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);

        // Guest signs
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);

        // Add signatures to state
        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;

        Signature[] memory bothSigs = new Signature[](2);
        bothSigs[0] = hostSig;
        bothSigs[1] = guestSig;

        // 3. Host opens channel with partial funding
        initialState.sigs = hostSigs;
        vm.prank(host);
        custody.deposit(address(token), DEPOSIT_AMOUNT * 2);

        vm.prank(host);
        bytes32 channelId = custody.open(chan, initialState);

        // 4. Guest completes channel funding
        State memory fullState = initialState;
        fullState.sigs = bothSigs;

        vm.prank(guest);
        custody.deposit(address(token), DEPOSIT_AMOUNT * 2);

        vm.prank(guest);
        custody.open(chan, fullState);

        // Verify channel is now ACTIVE
        bytes32[] memory hostChannels = custody.getAccountChannels(host);
        assertEq(hostChannels.length, 1, "Host should have 1 channel");

        bytes32[] memory guestChannels = custody.getAccountChannels(guest);
        assertEq(guestChannels.length, 1, "Guest should have 1 channel");

        // Check locked amounts
        (uint256 hostAvailable, uint256 hostLocked,) = custody.getAccountInfo(host, address(token));
        (uint256 guestAvailable, uint256 guestLocked,) = custody.getAccountInfo(guest, address(token));

        assertEq(hostLocked, DEPOSIT_AMOUNT, "Host's tokens not locked correctly");
        assertEq(guestLocked, DEPOSIT_AMOUNT, "Guest's tokens not locked correctly");
    }

    function test_InvalidChannelCreation() public {
        // Test with zero address as participant
        address[2] memory invalidParticipants = [host, address(0)];
        Channel memory chanWithZeroAddress = Channel({
            participants: invalidParticipants,
            adjudicator: address(adjudicator),
            challenge: CHALLENGE_DURATION,
            nonce: NONCE
        });

        State memory initialState = createInitialState([host, address(0)], 1);
        Signature memory hostSig = signState(chanWithZeroAddress, initialState, hostPrivKey);
        Signature[] memory sigs = new Signature[](1);
        sigs[0] = hostSig;
        initialState.sigs = sigs;

        depositTokens(host, DEPOSIT_AMOUNT * 2);

        vm.startPrank(host);
        vm.expectRevert(Custody.InvalidParticipant.selector);
        custody.open(chanWithZeroAddress, initialState);
        vm.stopPrank();

        // Test with zero address as adjudicator
        Channel memory chanWithZeroAdjudicator =
            Channel({participants: [host, guest], adjudicator: address(0), challenge: CHALLENGE_DURATION, nonce: NONCE});

        vm.startPrank(host);
        vm.expectRevert(Custody.InvalidAdjudicator.selector);
        custody.open(chanWithZeroAdjudicator, initialState);
        vm.stopPrank();

        // Test with zero challenge period
        Channel memory chanWithZeroChallenge =
            Channel({participants: [host, guest], adjudicator: address(adjudicator), challenge: 0, nonce: NONCE});

        vm.startPrank(host);
        vm.expectRevert(Custody.InvalidChallengePeriod.selector);
        custody.open(chanWithZeroChallenge, initialState);
        vm.stopPrank();
    }

    // ==== 2. Channel Closing ====

    function test_ChannelCooperativeClose() public {
        // 1. First create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState([host, guest], 1);

        // Host signs initial state
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;
        initialState.sigs = hostSigs;

        // Both participants sign full state
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);
        Signature[] memory bothSigs = new Signature[](2);
        bothSigs[0] = hostSig;
        bothSigs[1] = guestSig;

        // Open channel with both participants
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.open(chan, initialState);

        depositTokens(guest, DEPOSIT_AMOUNT * 2);
        State memory fullState = initialState;
        fullState.sigs = bothSigs;
        vm.prank(guest);
        custody.open(chan, fullState);

        // 2. Create a final state that both participants sign
        // Create a new state with updated counter
        Counter.CounterApp memory finalCounterApp = Counter.CounterApp({
            counter: TARGET, // Set to target to make it a final state
            target: TARGET,
            version: 1
        });

        State memory finalState = initialState;
        finalState.data = abi.encode(finalCounterApp);

        // Both sign the final state
        hostSig = signState(chan, finalState, hostPrivKey);
        guestSig = signState(chan, finalState, guestPrivKey);
        bothSigs[0] = hostSig;
        bothSigs[1] = guestSig;
        finalState.sigs = bothSigs;

        // 3. Close the channel cooperatively
        vm.prank(host);
        custody.close(channelId, finalState, new State[](0));

        // 4. Verify channel is closed and funds returned
        bytes32[] memory hostChannels = custody.getAccountChannels(host);
        assertEq(hostChannels.length, 0, "Host should have no channels after close");

        (uint256 hostAvailable, uint256 hostLocked,) = custody.getAccountInfo(host, address(token));
        (uint256 guestAvailable, uint256 guestLocked,) = custody.getAccountInfo(guest, address(token));

        assertEq(hostLocked, 0, "Host's tokens should be unlocked");
        assertEq(guestLocked, 0, "Guest's tokens should be unlocked");
        assertEq(hostAvailable, INITIAL_BALANCE, "Host's available balance incorrect");
        assertEq(guestAvailable, INITIAL_BALANCE, "Guest's available balance incorrect");
    }

    function test_InvalidChannelClose() public {
        // 1. Create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState([host, guest], 1);

        // Set up signatures
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);

        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;

        Signature[] memory bothSigs = new Signature[](2);
        bothSigs[0] = hostSig;
        bothSigs[1] = guestSig;

        // Open the channel
        initialState.sigs = hostSigs;
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.open(chan, initialState);

        State memory fullState = initialState;
        fullState.sigs = bothSigs;
        depositTokens(guest, DEPOSIT_AMOUNT * 2);
        vm.prank(guest);
        custody.open(chan, fullState);

        // 2. Try to close with only one signature
        State memory invalidState = initialState;
        invalidState.sigs = hostSigs;

        vm.startPrank(host);
        vm.expectRevert(Custody.ChannelNotFinal.selector);
        custody.close(channelId, invalidState, new State[](0));
        vm.stopPrank();

        // 3. Try to close non-existent channel
        bytes32 nonExistentChannelId = bytes32(uint256(1234));

        vm.startPrank(host);
        vm.expectRevert(abi.encodeWithSelector(Custody.ChannelNotFound.selector, nonExistentChannelId));
        custody.close(nonExistentChannelId, fullState, new State[](0));
        vm.stopPrank();
    }

    // ==== 3. Challenge Mechanism ====

    function test_ChannelChallenge() public {
        // 1. Create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState([host, guest], 1);

        // Set up signatures
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);

        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;

        Signature[] memory bothSigs = new Signature[](2);
        bothSigs[0] = hostSig;
        bothSigs[1] = guestSig;

        // Open channel
        initialState.sigs = hostSigs;
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.open(chan, initialState);

        State memory fullState = initialState;
        fullState.sigs = bothSigs;
        depositTokens(guest, DEPOSIT_AMOUNT * 2);
        vm.prank(guest);
        custody.open(chan, fullState);

        // 2. Create a challenge state
        Counter.CounterApp memory challengeCounterApp = Counter.CounterApp({counter: 2, target: TARGET, version: 1});

        State memory challengeState = initialState;
        challengeState.data = abi.encode(challengeCounterApp);

        // Host signs the challenge state
        Signature memory hostChallengeSig = signState(chan, challengeState, hostPrivKey);
        Signature[] memory challengeSigs = new Signature[](1);
        challengeSigs[0] = hostChallengeSig;
        challengeState.sigs = challengeSigs;

        // 3. Host challenges with this state
        vm.prank(host);
        custody.challenge(channelId, challengeState, new State[](1));

        // 4. Create a counter-challenge state
        Counter.CounterApp memory counterChallengeApp = Counter.CounterApp({counter: 3, target: TARGET, version: 2});

        State memory counterChallengeState = initialState;
        counterChallengeState.data = abi.encode(counterChallengeApp);

        // Guest signs the counter-challenge state
        Signature memory guestChallengeSig = signState(chan, counterChallengeState, guestPrivKey);
        Signature[] memory counterChallengeSigs = new Signature[](1);
        counterChallengeSigs[0] = guestChallengeSig;
        counterChallengeState.sigs = counterChallengeSigs;

        // 5. Guest counter-challenges
        vm.prank(guest);
        custody.challenge(channelId, counterChallengeState, new State[](1));

        // 6. Skip time and reclaim
        skipChallengeTime();

        vm.prank(host);
        custody.reclaim(channelId);

        // 7. Verify channel is closed and funds returned
        bytes32[] memory hostChannels = custody.getAccountChannels(host);
        assertEq(hostChannels.length, 0, "Host should have no channels after reclaim");

        (uint256 hostAvailable, uint256 hostLocked,) = custody.getAccountInfo(host, address(token));
        (uint256 guestAvailable, uint256 guestLocked,) = custody.getAccountInfo(guest, address(token));

        assertEq(hostLocked, 0, "Host's tokens should be unlocked");
        assertEq(guestLocked, 0, "Guest's tokens should be unlocked");
    }

    function test_InvalidChallenge() public {
        // 1. Create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState([host, guest], 1);

        // Set up signatures
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);

        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;

        Signature[] memory bothSigs = new Signature[](2);
        bothSigs[0] = hostSig;
        bothSigs[1] = guestSig;

        // Open channel
        initialState.sigs = hostSigs;
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.open(chan, initialState);

        State memory fullState = initialState;
        fullState.sigs = bothSigs;
        depositTokens(guest, DEPOSIT_AMOUNT * 2);
        vm.prank(guest);
        custody.open(chan, fullState);

        // 2. Try to challenge with invalid state (non-sequential version)
        Counter.CounterApp memory invalidCounterApp = Counter.CounterApp({
            counter: 2,
            target: TARGET,
            version: 3 // Skipping version 1 and 2
        });

        State memory invalidState = initialState;
        invalidState.data = abi.encode(invalidCounterApp);

        // Host signs the invalid state
        Signature memory hostInvalidSig = signState(chan, invalidState, hostPrivKey);
        Signature[] memory invalidSigs = new Signature[](1);
        invalidSigs[0] = hostInvalidSig;
        invalidState.sigs = invalidSigs;

        // Attempt to challenge with invalid state
        vm.prank(host);
        vm.expectRevert(Custody.InvalidState.selector);
        custody.challenge(channelId, invalidState, new State[](0));

        // 3. Try to challenge non-existent channel
        bytes32 nonExistentChannelId = bytes32(uint256(1234));

        vm.prank(host);
        vm.expectRevert(abi.encodeWithSelector(Custody.ChannelNotFound.selector, nonExistentChannelId));
        custody.challenge(nonExistentChannelId, invalidState, new State[](0));
    }

    // ==== 4. Checkpoint Mechanism ====

    function test_Checkpoint() public {
        // 1. Create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState([host, guest], 1);

        // Set up signatures
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);

        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;

        Signature[] memory bothSigs = new Signature[](2);
        bothSigs[0] = hostSig;
        bothSigs[1] = guestSig;

        // Open channel
        initialState.sigs = hostSigs;
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.open(chan, initialState);

        State memory fullState = initialState;
        fullState.sigs = bothSigs;
        depositTokens(guest, DEPOSIT_AMOUNT * 2);
        vm.prank(guest);
        custody.open(chan, fullState);

        // 2. Create a new state to checkpoint
        Counter.CounterApp memory checkpointCounterApp = Counter.CounterApp({counter: 2, target: TARGET, version: 1});

        State memory checkpointState = initialState;
        checkpointState.data = abi.encode(checkpointCounterApp);

        // Both sign the checkpoint state
        Signature memory hostCheckpointSig = signState(chan, checkpointState, hostPrivKey);
        Signature memory guestCheckpointSig = signState(chan, checkpointState, guestPrivKey);

        Signature[] memory checkpointSigs = new Signature[](2);
        checkpointSigs[0] = hostCheckpointSig;
        checkpointSigs[1] = guestCheckpointSig;
        checkpointState.sigs = checkpointSigs;

        // 3. Checkpoint the state
        vm.prank(host);
        custody.checkpoint(channelId, checkpointState, new State[](1));

        // 4. Try to challenge with an older state (should fail)
        vm.prank(guest);
        vm.expectRevert(); // Should revert as initial state is older
        custody.challenge(channelId, initialState, new State[](0));
    }

    // ==== 5. Reclaim Function ====

    function test_ReclaimAfterChallenge() public {
        // 1. Create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState([host, guest], 1);

        // Set up signatures
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);

        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;

        Signature[] memory bothSigs = new Signature[](2);
        bothSigs[0] = hostSig;
        bothSigs[1] = guestSig;

        // Open channel
        initialState.sigs = hostSigs;
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.open(chan, initialState);

        State memory fullState = initialState;
        fullState.sigs = bothSigs;
        depositTokens(guest, DEPOSIT_AMOUNT * 2);
        vm.prank(guest);
        custody.open(chan, fullState);

        // 2. Create a challenge state
        Counter.CounterApp memory challengeCounterApp = Counter.CounterApp({counter: 2, target: TARGET, version: 1});

        State memory challengeState = initialState;
        challengeState.data = abi.encode(challengeCounterApp);

        // Host signs the challenge state
        Signature memory hostChallengeSig = signState(chan, challengeState, hostPrivKey);
        Signature[] memory challengeSigs = new Signature[](1);
        challengeSigs[0] = hostChallengeSig;
        challengeState.sigs = challengeSigs;

        // 3. Host challenges
        vm.prank(host);
        custody.challenge(channelId, challengeState, new State[](1));

        // 4. Try to reclaim before challenge period expires (should fail)
        vm.prank(host);
        vm.expectRevert(Custody.ChallengeNotExpired.selector);
        custody.reclaim(channelId);

        // 5. Skip time and reclaim
        skipChallengeTime();

        vm.prank(host);
        custody.reclaim(channelId);

        // 6. Verify channel is closed and funds returned
        bytes32[] memory hostChannels = custody.getAccountChannels(host);
        assertEq(hostChannels.length, 0, "Host should have no channels after reclaim");

        (uint256 hostAvailable, uint256 hostLocked,) = custody.getAccountInfo(host, address(token));
        (uint256 guestAvailable, uint256 guestLocked,) = custody.getAccountInfo(guest, address(token));

        assertEq(hostLocked, 0, "Host's tokens should be unlocked");
        assertEq(guestLocked, 0, "Guest's tokens should be unlocked");
    }

    // ==== 6. Fund Management ====

    function test_DepositAndWithdraw() public {
        // 1. Test deposit
        vm.startPrank(host);
        custody.deposit(address(token), DEPOSIT_AMOUNT);

        (uint256 available, uint256 locked,) = custody.getAccountInfo(host, address(token));
        assertEq(available, DEPOSIT_AMOUNT, "Deposit not recorded correctly");
        assertEq(locked, 0, "No funds should be locked initially");

        // 2. Test withdrawal
        custody.withdraw(address(token), DEPOSIT_AMOUNT / 2);

        (available, locked,) = custody.getAccountInfo(host, address(token));
        assertEq(available, DEPOSIT_AMOUNT / 2, "Withdrawal not processed correctly");

        // 3. Test insufficient balance for withdrawal
        vm.expectRevert(
            abi.encodeWithSelector(Custody.InsufficientBalance.selector, DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT)
        );
        custody.withdraw(address(token), DEPOSIT_AMOUNT);

        vm.stopPrank();
    }

    // ==== 7. Reset Function ====

    function test_Reset() public {
        // 1. Create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState([host, guest], 1);

        // Set up signatures
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);

        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;

        Signature[] memory bothSigs = new Signature[](2);
        bothSigs[0] = hostSig;
        bothSigs[1] = guestSig;

        // Open channel
        initialState.sigs = hostSigs;
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.open(chan, initialState);

        State memory fullState = initialState;
        fullState.sigs = bothSigs;
        depositTokens(guest, DEPOSIT_AMOUNT * 2);
        vm.prank(guest);
        custody.open(chan, fullState);

        // 2. Create a final state for closing
        Counter.CounterApp memory finalCounterApp = Counter.CounterApp({counter: TARGET, target: TARGET, version: 1});

        State memory finalState = initialState;
        finalState.data = abi.encode(finalCounterApp);

        // Both sign the final state
        Signature memory hostFinalSig = signState(chan, finalState, hostPrivKey);
        Signature memory guestFinalSig = signState(chan, finalState, guestPrivKey);

        Signature[] memory finalSigs = new Signature[](2);
        finalSigs[0] = hostFinalSig;
        finalSigs[1] = guestFinalSig;
        finalState.sigs = finalSigs;

        // 3. Create new channel config with different nonce
        Channel memory newChan = Channel({
            participants: [host, guest],
            adjudicator: address(adjudicator),
            challenge: CHALLENGE_DURATION,
            nonce: NONCE + 1
        });

        // 4. Create new initial state for the new channel
        State memory newInitialState = createInitialState([host, guest], 1);

        // Host signs new initial state
        Signature memory hostNewSig = signState(newChan, newInitialState, hostPrivKey);
        Signature[] memory newHostSigs = new Signature[](1);
        newHostSigs[0] = hostNewSig;
        newInitialState.sigs = newHostSigs;

        // 5. Reset the channel
        vm.prank(host);
        custody.reset(channelId, finalState, new State[](0), newChan, newInitialState);

        // 6. Verify old channel is closed and new one is in PARTIAL state
        bytes32[] memory hostChannels = custody.getAccountChannels(host);
        assertEq(hostChannels.length, 1, "Host should have 1 channel after reset");

        bytes32 newChannelId = Utils.getChannelId(newChan);
        assertEq(hostChannels[0], newChannelId, "Host's channel should be the new one");
    }
}
