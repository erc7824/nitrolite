// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Test, console} from "lib/forge-std/src/Test.sol";
import {IERC20} from "lib/openzeppelin-contracts/contracts/interfaces/IERC20.sol";

import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";
import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";

import {TestUtils} from "./TestUtils.sol";
import {Custody} from "../src/Custody.sol";
import {Channel, State, Allocation, Signature, Status, Amount, CHANOPEN, CHANCLOSE} from "../src/interfaces/Types.sol";
import {Utils} from "../src/Utils.sol";

import {FlagAdjudicator} from "./mocks/FlagAdjudicator.sol";
import {MockERC20} from "./mocks/MockERC20.sol";

contract CustodyTest is Test {
    Custody public custody;
    FlagAdjudicator public adjudicator;
    MockERC20 public token;

    // Private keys for testing
    uint256 constant hostPrivKey = 1;
    uint256 constant guestPrivKey = 2;
    uint256 constant nonParticipantPrivKey = 3;
    uint256 constant depositorPrivKey = 4;
    uint256 constant hostParticipantPrivKey = 5;
    uint256 constant guestParticipantPrivKey = 6;

    // Test users
    address public host;
    address public guest;
    address public nonParticipant;
    address public depositor;
    address public hostParticipant;
    address public guestParticipant;

    // Common test values
    uint64 constant CHALLENGE_DURATION = 3600; // 1 hour
    uint64 constant NONCE = 1;
    uint256 constant DEPOSIT_AMOUNT = 1000;
    uint256 constant INITIAL_BALANCE = 10000;

    function setUp() public {
        // Set up user addresses from private keys
        host = vm.addr(hostPrivKey);
        guest = vm.addr(guestPrivKey);
        nonParticipant = vm.addr(nonParticipantPrivKey);
        depositor = vm.addr(depositorPrivKey);
        hostParticipant = vm.addr(hostParticipantPrivKey);
        guestParticipant = vm.addr(guestParticipantPrivKey);

        // Deploy contracts
        custody = new Custody();
        adjudicator = new FlagAdjudicator();
        token = new MockERC20("Test Token", "TST", 18);

        // Fund accounts
        token.mint(host, INITIAL_BALANCE);
        token.mint(guest, INITIAL_BALANCE);
        token.mint(nonParticipant, INITIAL_BALANCE);
        token.mint(depositor, INITIAL_BALANCE);

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

        vm.startPrank(depositor);
        token.approve(address(custody), INITIAL_BALANCE);
        vm.stopPrank();
    }

    // Helper to create a standard test channel
    function createTestChannel() internal view returns (Channel memory) {
        address[] memory participants = new address[](2);
        participants[0] = host;
        participants[1] = guest;

        return Channel({
            participants: participants,
            adjudicator: address(adjudicator),
            challenge: CHALLENGE_DURATION,
            nonce: NONCE
        });
    }

    // Helper to create a test channel with separate participant addresses
    function createTestChannelWithParticipants() internal view returns (Channel memory) {
        address[] memory participants = new address[](2);
        participants[0] = hostParticipant;
        participants[1] = guestParticipant;

        return Channel({
            participants: participants,
            adjudicator: address(adjudicator),
            challenge: CHALLENGE_DURATION,
            nonce: NONCE
        });
    }

    // Helper to create an initial state for testing
    function createInitialState() internal view returns (State memory) {
        // Create allocations for both participants
        Allocation[] memory allocations = new Allocation[](2);

        allocations[0] = Allocation({destination: host, token: address(token), amount: DEPOSIT_AMOUNT});

        allocations[1] = Allocation({destination: guest, token: address(token), amount: DEPOSIT_AMOUNT});

        // Create openChannel magic number data
        bytes memory data = abi.encode(CHANOPEN);

        // Create unsigned state
        return State({
            data: data,
            allocations: allocations,
            sigs: new Signature[](0) // Empty initially
        });
    }

    // Helper to create an initial state for testing with separate participants
    function createInitialStateWithParticipants() internal view returns (State memory) {
        // Create allocations for both participants
        Allocation[] memory allocations = new Allocation[](2);

        allocations[0] = Allocation({destination: hostParticipant, token: address(token), amount: DEPOSIT_AMOUNT});

        allocations[1] = Allocation({destination: guestParticipant, token: address(token), amount: DEPOSIT_AMOUNT});

        // Create openChannel magic number data
        bytes memory data = abi.encode(CHANOPEN);

        // Create unsigned state
        return State({
            data: data,
            allocations: allocations,
            sigs: new Signature[](0) // Empty initially
        });
    }

    // Helper to create a closing state
    function createClosingState() internal view returns (State memory) {
        // Create allocations for both participants
        Allocation[] memory allocations = new Allocation[](2);

        allocations[0] = Allocation({destination: host, token: address(token), amount: DEPOSIT_AMOUNT});

        allocations[1] = Allocation({destination: guest, token: address(token), amount: DEPOSIT_AMOUNT});

        // Create closeChannel magic number data
        bytes memory data = abi.encode(CHANCLOSE);

        // Create unsigned state
        return State({
            data: data,
            allocations: allocations,
            sigs: new Signature[](0) // Empty initially
        });
    }

    // Helper to create a closing state with separate participants
    function createClosingStateWithParticipants() internal view returns (State memory) {
        // Create allocations for both participants
        Allocation[] memory allocations = new Allocation[](2);

        allocations[0] = Allocation({destination: hostParticipant, token: address(token), amount: DEPOSIT_AMOUNT});

        allocations[1] = Allocation({destination: guestParticipant, token: address(token), amount: DEPOSIT_AMOUNT});

        // Create closeChannel magic number data
        bytes memory data = abi.encode(CHANCLOSE);

        // Create unsigned state
        return State({
            data: data,
            allocations: allocations,
            sigs: new Signature[](0) // Empty initially
        });
    }

    // Helper to sign a state
    function signState(Channel memory chan, State memory state, uint256 privateKey)
        internal
        pure
        returns (Signature memory)
    {
        bytes32 stateHash = Utils.getStateHash(chan, state);
        (uint8 v, bytes32 r, bytes32 s) = TestUtils.sign(vm, privateKey, stateHash);
        return Signature({v: v, r: r, s: s});
    }

    // Helper to sign a challenge
    function signChallenge(Channel memory chan, State memory state, uint256 privateKey)
        internal
        pure
        returns (Signature memory)
    {
        bytes32 stateHash = Utils.getStateHash(chan, state);
        bytes32 challengeHash = keccak256(abi.encode(stateHash, "challenge"));
        (uint8 v, bytes32 r, bytes32 s) = TestUtils.sign(vm, privateKey, challengeHash);
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

    // ==== 1. Channel Creation and Joining ====

    function test_ChannelCreation() public {
        // 1. Prepare channel and initial state
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState();

        // 2. Sign the state by the host
        vm.deal(host, 1 ether); // Ensure host has ETH for gas

        // Sign the state
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature[] memory sigs = new Signature[](1);
        sigs[0] = hostSig;
        initialState.sigs = sigs;

        // 3. Deposit tokens for the host
        vm.prank(host);
        custody.deposit(address(token), DEPOSIT_AMOUNT * 2);

        // 4. Create the channel as host
        vm.prank(host);
        bytes32 channelId = custody.create(chan, initialState);

        // Verify the channel is created and in INITIAL state
        (uint256 available, uint256 channelCount) = custody.getAccountInfo(host, address(token));
        assertEq(available, DEPOSIT_AMOUNT, "Host should have correct available balance");
        assertEq(channelCount, 1, "Host should have 1 channel");

        // Also check that the channelId is consistent
        bytes32 expectedChannelId = Utils.getChannelId(chan);
        assertEq(channelId, expectedChannelId, "Channel ID is incorrect");
    }

    function test_CompleteChannelFunding() public {
        // 1. Create channel with host
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState();

        // 2. Sign the state by the host
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;
        initialState.sigs = hostSigs;

        // 3. Host creates channel with initial funding
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.create(chan, initialState);

        // 4. Guest joins the channel
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);
        depositTokens(guest, DEPOSIT_AMOUNT * 2);

        vm.prank(guest);
        custody.join(channelId, 1, guestSig);

        // Verify channel is now ACTIVE
        bytes32[] memory hostChannels = custody.getAccountChannels(host);
        assertEq(hostChannels.length, 1, "Host should have 1 channel");

        bytes32[] memory guestChannels = custody.getAccountChannels(guest);
        assertEq(guestChannels.length, 1, "Guest should have 1 channel");

        // Check available amounts
        (uint256 hostAvailable,) = custody.getAccountInfo(host, address(token));
        (uint256 guestAvailable,) = custody.getAccountInfo(guest, address(token));

        assertEq(hostAvailable, DEPOSIT_AMOUNT, "Host should have correct available balance");
        assertEq(guestAvailable, DEPOSIT_AMOUNT, "Guest should have correct available balance");
    }

    function test_InvalidChannelCreation() public {
        // Create channel with invalid participant (empty array)
        address[] memory invalidParticipants = new address[](1);
        invalidParticipants[0] = host;
        Channel memory chanWithInvalidParticipants = Channel({
            participants: invalidParticipants,
            adjudicator: address(adjudicator),
            challenge: CHALLENGE_DURATION,
            nonce: NONCE
        });

        State memory initialState = createInitialState();
        Signature memory hostSig = signState(chanWithInvalidParticipants, initialState, hostPrivKey);
        Signature[] memory sigs = new Signature[](1);
        sigs[0] = hostSig;
        initialState.sigs = sigs;

        depositTokens(host, DEPOSIT_AMOUNT * 2);

        vm.startPrank(host);
        vm.expectRevert(Custody.InvalidParticipant.selector);
        custody.create(chanWithInvalidParticipants, initialState);
        vm.stopPrank();

        // Test with zero address as adjudicator
        address[] memory validParticipants = new address[](2);
        validParticipants[0] = host;
        validParticipants[1] = guest;

        Channel memory chanWithZeroAdjudicator = Channel({
            participants: validParticipants,
            adjudicator: address(0),
            challenge: CHALLENGE_DURATION,
            nonce: NONCE
        });

        vm.startPrank(host);
        vm.expectRevert(Custody.InvalidAdjudicator.selector);
        custody.create(chanWithZeroAdjudicator, initialState);
        vm.stopPrank();

        // Test with zero challenge period
        Channel memory chanWithZeroChallenge =
            Channel({participants: validParticipants, adjudicator: address(adjudicator), challenge: 0, nonce: NONCE});

        vm.startPrank(host);
        vm.expectRevert(Custody.InvalidChallengePeriod.selector);
        custody.create(chanWithZeroChallenge, initialState);
        vm.stopPrank();
    }

    // ==== 2. Channel Closing ====

    function test_ChannelCooperativeClose() public {
        // 1. First create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState();

        // Host signs initial state
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;
        initialState.sigs = hostSigs;

        // Create channel with host
        depositTokens(host, DEPOSIT_AMOUNT);
        vm.prank(host);
        bytes32 channelId = custody.create(chan, initialState);

        // Guest joins the channel
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);
        depositTokens(guest, DEPOSIT_AMOUNT);
        vm.prank(guest);
        custody.join(channelId, 1, guestSig);

        // 2. Create a final state that both participants sign
        State memory finalState = createClosingState();

        // Both sign the final state
        hostSig = signState(chan, finalState, hostPrivKey);
        guestSig = signState(chan, finalState, guestPrivKey);

        Signature[] memory bothSigs = new Signature[](2);
        bothSigs[0] = hostSig;
        bothSigs[1] = guestSig;
        finalState.sigs = bothSigs;

        // 3. Close the channel cooperatively
        vm.prank(host);
        custody.close(channelId, finalState, new State[](0));

        // 4. Verify channel is closed and funds returned
        bytes32[] memory hostChannels = custody.getAccountChannels(host);
        assertEq(hostChannels.length, 0, "Host should have no channels after close");

        (uint256 hostAvailable,) = custody.getAccountInfo(host, address(token));
        (uint256 guestAvailable,) = custody.getAccountInfo(guest, address(token));

        assertEq(hostAvailable, DEPOSIT_AMOUNT, "Host's available balance incorrect");
        assertEq(guestAvailable, DEPOSIT_AMOUNT, "Guest's available balance incorrect");
    }

    function test_InvalidChannelClose() public {
        // 1. Create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState();

        // Set up signatures
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;
        initialState.sigs = hostSigs;

        // Create channel with host
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.create(chan, initialState);

        // Guest joins the channel
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);
        depositTokens(guest, DEPOSIT_AMOUNT * 2);
        vm.prank(guest);
        custody.join(channelId, 1, guestSig);

        // 2. Try to close with invalid close state (missing CHANCLOSE magic number)
        State memory invalidState = initialState; // Not a closing state

        Signature[] memory bothSigs = new Signature[](2);
        bothSigs[0] = hostSig;
        bothSigs[1] = guestSig;
        invalidState.sigs = bothSigs;

        vm.startPrank(host);
        vm.expectRevert(Custody.InvalidState.selector);
        custody.close(channelId, invalidState, new State[](0));
        vm.stopPrank();

        // 3. Try to close non-existent channel
        bytes32 nonExistentChannelId = bytes32(uint256(1234));
        State memory closingState = createClosingState();
        closingState.sigs = bothSigs;

        vm.startPrank(host);
        vm.expectRevert(abi.encodeWithSelector(Custody.ChannelNotFound.selector, nonExistentChannelId));
        custody.close(nonExistentChannelId, closingState, new State[](0));
        vm.stopPrank();
    }

    // ==== 3. Challenge Mechanism ====

    function test_ChannelChallenge() public {
        // 1. Create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState();

        // Set up signatures
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;
        initialState.sigs = hostSigs;

        // Create channel with host
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.create(chan, initialState);

        // Guest joins the channel
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);
        depositTokens(guest, DEPOSIT_AMOUNT * 2);
        vm.prank(guest);
        custody.join(channelId, 1, guestSig);

        // 2. Create a challenge state
        State memory challengeState = initialState;
        challengeState.data = abi.encode(42);

        // Host signs the challenge state
        Signature memory hostChallengeSig = signState(chan, challengeState, hostPrivKey);
        Signature[] memory challengeSigs = new Signature[](1);
        challengeSigs[0] = hostChallengeSig;
        challengeState.sigs = challengeSigs;

        // 3. Host challenges with this state and signs the challenge
        Signature memory hostChallengeSigFinal = signChallenge(chan, challengeState, hostPrivKey);
        vm.prank(host);
        custody.challenge(channelId, challengeState, new State[](0), hostChallengeSigFinal);

        // 4. Create a counter-challenge state (more signatures = "newer")
        State memory counterChallengeState = initialState;
        counterChallengeState.data = abi.encode(4242);

        // Both sign the counter-challenge
        Signature memory hostCounterSig = signState(chan, counterChallengeState, hostPrivKey);
        Signature memory guestCounterSig = signState(chan, counterChallengeState, guestPrivKey);

        Signature[] memory counterChallengeSigs = new Signature[](2);
        counterChallengeSigs[0] = hostCounterSig;
        counterChallengeSigs[1] = guestCounterSig;
        counterChallengeState.sigs = counterChallengeSigs;

        // 5. Guest counter-challenges with their own signature
        Signature memory guestChallengeSignature = signChallenge(chan, counterChallengeState, guestPrivKey);
        vm.prank(guest);
        custody.challenge(channelId, counterChallengeState, new State[](0), guestChallengeSignature);

        // 6. Skip time and close the channel
        skipChallengeTime();

        vm.prank(host);
        custody.close(channelId, counterChallengeState, new State[](0));

        // 7. Verify channel is closed and funds returned
        bytes32[] memory hostChannels = custody.getAccountChannels(host);
        assertEq(hostChannels.length, 0, "Host should have no channels after challenge resolution");

        (uint256 hostAvailable,) = custody.getAccountInfo(host, address(token));
        (uint256 guestAvailable,) = custody.getAccountInfo(guest, address(token));

        assertEq(hostAvailable, DEPOSIT_AMOUNT * 2, "Host's available balance incorrect");
        assertEq(guestAvailable, DEPOSIT_AMOUNT * 2, "Guest's available balance incorrect");
    }

    function test_InvalidChallenge() public {
        // 1. Create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState();

        // Set up signatures
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;
        initialState.sigs = hostSigs;

        // Create channel with host
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.create(chan, initialState);

        // Guest joins the channel
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);
        depositTokens(guest, DEPOSIT_AMOUNT * 2);
        vm.prank(guest);
        custody.join(channelId, 1, guestSig);

        // 2. Try to challenge with invalid state (adjudicator rejects)
        State memory invalidState = initialState;
        invalidState.data = abi.encode(42);
        adjudicator.setAdjudicateReturnValue(false); // Set adjudicate return value to false for invalid state

        // Host signs the invalid state
        Signature memory hostInvalidSig = signState(chan, invalidState, hostPrivKey);
        Signature[] memory invalidSigs = new Signature[](1);
        invalidSigs[0] = hostInvalidSig;
        invalidState.sigs = invalidSigs;

        // Attempt to challenge with invalid state
        Signature memory hostInvalidChallengeSig = signChallenge(chan, invalidState, hostPrivKey);
        vm.prank(host);
        vm.expectRevert(Custody.InvalidState.selector);
        custody.challenge(channelId, invalidState, new State[](0), hostInvalidChallengeSig);

        // 3. Try to challenge non-existent channel
        bytes32 nonExistentChannelId = bytes32(uint256(1234));
        adjudicator.setAdjudicateReturnValue(true); // Set flag back to true

        vm.prank(host);
        vm.expectRevert(abi.encodeWithSelector(Custody.ChannelNotFound.selector, nonExistentChannelId));
        custody.challenge(nonExistentChannelId, invalidState, new State[](0), hostInvalidChallengeSig);
    }

    function test_InvalidChallengerSignature() public {
        // 1. Create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState();

        // Set up signatures
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;
        initialState.sigs = hostSigs;

        // Create channel with host
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.create(chan, initialState);

        // Guest joins the channel
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);
        depositTokens(guest, DEPOSIT_AMOUNT * 2);
        vm.prank(guest);
        custody.join(channelId, 1, guestSig);

        // 2. Create a challenge state
        State memory challengeState = initialState;
        challengeState.data = abi.encode(42);

        // Host signs the challenge state
        Signature memory hostChallengeSig = signState(chan, challengeState, hostPrivKey);
        Signature[] memory challengeSigs = new Signature[](1);
        challengeSigs[0] = hostChallengeSig;
        challengeState.sigs = challengeSigs;

        // 3. Non-participant tries to challenge with a signature from non-participant
        Signature memory nonParticipantSig = signChallenge(chan, challengeState, nonParticipantPrivKey);
        adjudicator.setAdjudicateReturnValue(true); // Make sure adjudicator allows the state

        vm.prank(nonParticipant);
        vm.expectRevert(Custody.InvalidChallengerSignature.selector);
        custody.challenge(channelId, challengeState, new State[](0), nonParticipantSig);
    }

    // ==== 4. Checkpoint Mechanism ====

    function test_Checkpoint() public {
        // 1. Create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState();

        // Set up signatures
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;
        initialState.sigs = hostSigs;

        // Create channel with host
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.create(chan, initialState);

        // Guest joins the channel
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);
        depositTokens(guest, DEPOSIT_AMOUNT * 2);
        vm.prank(guest);
        custody.join(channelId, 1, guestSig);

        // 2. Create a new state to checkpoint
        State memory checkpointState = initialState;
        checkpointState.data = abi.encode(42);

        // Both sign the checkpoint state
        Signature memory hostCheckpointSig = signState(chan, checkpointState, hostPrivKey);
        Signature memory guestCheckpointSig = signState(chan, checkpointState, guestPrivKey);

        Signature[] memory checkpointSigs = new Signature[](2);
        checkpointSigs[0] = hostCheckpointSig;
        checkpointSigs[1] = guestCheckpointSig;
        checkpointState.sigs = checkpointSigs;

        // 3. Checkpoint the state
        vm.prank(host);
        custody.checkpoint(channelId, checkpointState, new State[](0));

        // 4. Start a challenge with single-signed state
        State memory challengeState = initialState;
        challengeState.data = abi.encode(21);
        Signature memory hostChallengeSig = signState(chan, challengeState, hostPrivKey);
        Signature[] memory challengeSigs = new Signature[](1);
        challengeSigs[0] = hostChallengeSig;
        challengeState.sigs = challengeSigs;

        Signature memory hostChallengeForCheckpoint = signChallenge(chan, challengeState, hostPrivKey);
        adjudicator.setCompareReturnValue(false); // make sure adjudicator allows the state
        vm.prank(host);
        custody.challenge(channelId, challengeState, new State[](0), hostChallengeForCheckpoint);
        adjudicator.setCompareReturnValue(true); // set value back

        // 5. Checkpoint should resolve the challenge
        vm.prank(guest);
        custody.checkpoint(channelId, checkpointState, new State[](0));

        // Close with checkpointed state
        skipChallengeTime();

        // Try to close normally - should succeed because challenge timer expired
        State memory closeState = createClosingState();
        // Add signatures
        Signature memory hostCloseSig = signState(chan, closeState, hostPrivKey);
        Signature memory guestCloseSig = signState(chan, closeState, guestPrivKey);
        Signature[] memory closeSigs = new Signature[](2);
        closeSigs[0] = hostCloseSig;
        closeSigs[1] = guestCloseSig;
        closeState.sigs = closeSigs;

        vm.prank(host);
        custody.close(channelId, closeState, new State[](0));
    }

    // ==== 5. Fund Management ====

    function test_DepositAndWithdraw() public {
        // 1. Test deposit
        vm.startPrank(host);
        custody.deposit(address(token), DEPOSIT_AMOUNT);

        (uint256 available,) = custody.getAccountInfo(host, address(token));
        assertEq(available, DEPOSIT_AMOUNT, "Deposit not recorded correctly");

        // 2. Test withdrawal
        custody.withdraw(address(token), DEPOSIT_AMOUNT / 2);

        (available,) = custody.getAccountInfo(host, address(token));
        assertEq(available, DEPOSIT_AMOUNT / 2, "Withdrawal not processed correctly");

        // 3. Test insufficient balance for withdrawal
        vm.expectRevert(
            abi.encodeWithSelector(Custody.InsufficientBalance.selector, DEPOSIT_AMOUNT / 2, DEPOSIT_AMOUNT)
        );
        custody.withdraw(address(token), DEPOSIT_AMOUNT);

        vm.stopPrank();
    }

    // ==== 6. Reset Function ====

    function test_Reset() public {
        // 1. Create and fund a channel
        Channel memory chan = createTestChannel();
        State memory initialState = createInitialState();

        // Set up signatures
        Signature memory hostSig = signState(chan, initialState, hostPrivKey);
        Signature[] memory hostSigs = new Signature[](1);
        hostSigs[0] = hostSig;
        initialState.sigs = hostSigs;

        // Create channel with host
        depositTokens(host, DEPOSIT_AMOUNT * 2);
        vm.prank(host);
        bytes32 channelId = custody.create(chan, initialState);

        // Guest joins the channel
        Signature memory guestSig = signState(chan, initialState, guestPrivKey);
        depositTokens(guest, DEPOSIT_AMOUNT * 2);
        vm.prank(guest);
        custody.join(channelId, 1, guestSig);

        // 2. Create a final state for closing
        State memory finalState = createClosingState();

        // Both sign the final state
        Signature memory hostFinalSig = signState(chan, finalState, hostPrivKey);
        Signature memory guestFinalSig = signState(chan, finalState, guestPrivKey);

        Signature[] memory finalSigs = new Signature[](2);
        finalSigs[0] = hostFinalSig;
        finalSigs[1] = guestFinalSig;
        finalState.sigs = finalSigs;

        // 3. Create new channel config with different nonce
        address[] memory participants = new address[](2);
        participants[0] = host;
        participants[1] = guest;

        Channel memory newChan = Channel({
            participants: participants,
            adjudicator: address(adjudicator),
            challenge: CHALLENGE_DURATION,
            nonce: NONCE + 1
        });

        // 4. Create new initial state for the new channel
        State memory newInitialState = createInitialState();

        // Host signs new initial state
        Signature memory hostNewSig = signState(newChan, newInitialState, hostPrivKey);
        Signature[] memory newHostSigs = new Signature[](1);
        newHostSigs[0] = hostNewSig;
        newInitialState.sigs = newHostSigs;

        // 5. Reset the channel
        vm.prank(host);
        custody.reset(channelId, finalState, new State[](0), newChan, newInitialState);

        // 6. Verify old channel is closed and new one is open
        bytes32[] memory hostChannels = custody.getAccountChannels(host);
        assertEq(hostChannels.length, 1, "Host should have 1 channel after reset");

        bytes32 newChannelId = Utils.getChannelId(newChan);
        assertEq(hostChannels[0], newChannelId, "Host's channel should be the new one");
    }

    // ==== 7. Separate Depositor and Participant Addresses ====

    function test_SeparateDepositorAndParticipant() public {
        // 1. Prepare channel with different participant addresses
        Channel memory chan = createTestChannelWithParticipants();
        State memory initialState = createInitialStateWithParticipants();
        // NOTE: depositor is specified instead of host in allocations
        initialState.allocations[0].destination = depositor;

        // 2. Sign the state by the host participant (not the depositor/creator)
        Signature memory hostPartSig = signState(chan, initialState, hostParticipantPrivKey);
        Signature[] memory sigs = new Signature[](1);
        sigs[0] = hostPartSig;
        initialState.sigs = sigs;

        // 3. Depositor deposits tokens into the participant accounts first
        vm.startPrank(depositor);
        custody.deposit(address(token), DEPOSIT_AMOUNT);
        vm.stopPrank();

        // 4. Create the channel as host participant
        vm.prank(depositor);
        bytes32 channelId = custody.create(chan, initialState);

        // 5. Verify the channel is created
        (uint256 available, uint256 channelCount) = custody.getAccountInfo(depositor, address(token));
        assertEq(available, 0, "Depositor should have no available balance after locking");
        assertEq(channelCount, 0, "Depositor should have 0 channels");
        (uint256 hostParticipantAvailable, uint256 hostParticipantChannelCount) = custody.getAccountInfo(hostParticipant, address(token));
        assertEq(hostParticipantAvailable, 0, "hostParticipant should not have available tokens");
        assertEq(hostParticipantChannelCount, 1, "hostParticipant should have 1 channel after creation");

        // 6. Guest participant joins the channel
        vm.startPrank(guestParticipant);
        token.mint(guestParticipant, INITIAL_BALANCE);
        token.approve(address(custody), INITIAL_BALANCE);
        custody.deposit(address(token), DEPOSIT_AMOUNT);
        vm.stopPrank();

        // Sign the state by guest participant
        Signature memory guestPartSig = signState(chan, initialState, guestParticipantPrivKey);

        // Guest participant joins with their own signature
        vm.prank(guestParticipant);
        custody.join(channelId, 1, guestPartSig);

        // 7. Verify channel is ACTIVE
        bytes32[] memory hostParticipantChannels = custody.getAccountChannels(hostParticipant);
        assertEq(hostParticipantChannels.length, 1, "hostParticipant should have 1 channel");

        bytes32[] memory guestChannels = custody.getAccountChannels(guestParticipant);
        assertEq(guestChannels.length, 1, "Guest participant should have 1 channel");

        // 8. Create a checkpoint state
        State memory checkpointState = initialState;
        checkpointState.data = abi.encode(42);

        // Both participants sign the checkpoint state
        Signature memory hostPartCheckpointSig = signState(chan, checkpointState, hostParticipantPrivKey);
        Signature memory guestPartCheckpointSig = signState(chan, checkpointState, guestParticipantPrivKey);

        Signature[] memory checkpointSigs = new Signature[](2);
        checkpointSigs[0] = hostPartCheckpointSig;
        checkpointSigs[1] = guestPartCheckpointSig;
        checkpointState.sigs = checkpointSigs;

        // 9. Checkpoint the state by the host participant
        vm.prank(depositor);
        custody.checkpoint(channelId, checkpointState, new State[](0));

        // 10. Create a closing state
        State memory finalState = createClosingStateWithParticipants();
        finalState.allocations[0].destination = depositor;

        // Both participants sign the final state
        Signature memory hostPartFinalSig = signState(chan, finalState, hostParticipantPrivKey);
        Signature memory guestPartFinalSig = signState(chan, finalState, guestParticipantPrivKey);

        Signature[] memory finalSigs = new Signature[](2);
        finalSigs[0] = hostPartFinalSig;
        finalSigs[1] = guestPartFinalSig;
        finalState.sigs = finalSigs;

        // 11. Close the channel cooperatively
        vm.prank(depositor);
        custody.close(channelId, finalState, new State[](0));

        // 12. Verify funds are returned correctly
        bytes32[] memory hostParticipantChannelsAfter = custody.getAccountChannels(hostParticipant);
        assertEq(hostParticipantChannelsAfter.length, 0, "hostParticipant should have no channels after close");

        bytes32[] memory guestChannelsAfter = custody.getAccountChannels(guestParticipant);
        assertEq(guestChannelsAfter.length, 0, "Guest participant should have no channels after close");

        (uint256 depositorAvailable,) = custody.getAccountInfo(depositor, address(token));
        (uint256 guestAvailable,) = custody.getAccountInfo(guestParticipant, address(token));

        // In this flow, the funds go back to participants (who are also depositors)
        assertEq(depositorAvailable, DEPOSIT_AMOUNT, "Depositor available balance incorrect");
        assertEq(guestAvailable, DEPOSIT_AMOUNT, "Guest available balance incorrect");
    }
}
