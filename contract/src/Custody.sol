// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {IChannel} from "./interfaces/IChannel.sol";
import {IDeposit} from "./interfaces/IDeposit.sol";
import {IAdjudicator} from "./interfaces/IAdjudicator.sol";
import {IComparable} from "./interfaces/IComparable.sol";
import {Channel, State, Allocation, Status, Signature, Amount, CHANOPEN, CHANCLOSE} from "./interfaces/Types.sol";
import {Utils} from "./Utils.sol";
import {IERC20} from "lib/openzeppelin-contracts/contracts/interfaces/IERC20.sol";
import {SafeERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/utils/SafeERC20.sol";
import {EnumerableSet} from "lib/openzeppelin-contracts/contracts/utils/structs/EnumerableSet.sol";

/**
 * @title Custody
 * @notice A simple custody contract for state channels that delegates most state transition logic to an adjudicator
 */
contract Custody is IChannel, IDeposit {
    using EnumerableSet for EnumerableSet.Bytes32Set;
    using SafeERC20 for IERC20;

    // Errors
    error ChannelNotFound(bytes32 channelId);
    error ChannelNotFinal();
    error InvalidParticipant();
    error InvalidStatus();
    error InvalidState();
    error InvalidAllocations();
    error InvalidStateSignatures();
    error InvalidAdjudicator();
    error InvalidChallengePeriod();
    error InvalidAmount();
    error TransferFailed(address token, address to, uint256 amount);
    error ChallengeNotExpired();
    error InsufficientBalance(uint256 available, uint256 required);

    // Recommended structure to keep track of states
    struct Metadata {
        Channel chan; // Opener define channel configuration
        Status stage;
        address creator;
        // TODO: replace 2 Amount[] arrays with EnumerableSet of participants that have joined
        Amount[] expectedDeposits; // Creator defines Token per participant
        Amount[] actualDeposits; // Tracks deposits made by each participant
        uint256 challengeExpire; // If non-zero channel will resolve to lastValidState when challenge Expires
        State lastValidState; // Last valid state when adjudicator was called
        mapping(address token => uint256 balance) tokenBalances; // Token balances for the channel
    }

    // Account is a ledger account per unique depositor and token
    struct Account {
        uint256 available; // Available amount that can be withdrawn or allocated to channels
        uint256 locked; // Amount currently allocated to channels
    }

    struct Ledger {
        mapping(address token => Account funds) tokens; // Token balances
        EnumerableSet.Bytes32Set channels; // Set of user ChannelId
    }

    mapping(bytes32 channelId => Metadata chMeta) internal _channels;
    mapping(address account => Ledger ledger) internal _ledgers;

    function deposit(address token, uint256 amount) external payable {
        address account = msg.sender;
        if (token == address(0)) {
            if (msg.value != amount) revert InvalidAmount();
        } else {
            bool success = IERC20(token).transferFrom(account, address(this), amount);
            if (!success) revert TransferFailed(token, address(this), amount);
        }
        _ledgers[msg.sender].tokens[token].available += amount;
    }

    function withdraw(address token, uint256 amount) external {
        Ledger storage ledger = _ledgers[msg.sender];
        uint256 available = ledger.tokens[token].available;
        if (available < amount) revert InsufficientBalance(available, amount);
        _transfer(token, msg.sender, amount);
        ledger.tokens[token].available -= amount;
    }

    /**
     * @notice Get channels associated with an account
     * @param account The account address
     * @return List of channel IDs associated with the account
     */
    function getAccountChannels(address account) public view returns (bytes32[] memory) {
        return _ledgers[account].channels.values();
    }

    /**
     * @notice Get account information for a specific token
     * @param user The account address
     * @param token The token address
     * @return available Amount available for withdrawal or allocation
     * @return locked Amount locked in channels
     * @return channelCount Number of associated channels
     */
    function getAccountInfo(address user, address token)
        public
        view
        returns (uint256 available, uint256 locked, uint256 channelCount)
    {
        Ledger storage ledger = _ledgers[user];
        Account storage account = ledger.tokens[token];
        return (account.available, account.locked, ledger.channels.length());
    }

    /**
     * @notice Create a channel by depositing assets
     * @param ch Channel configuration
     * @param initial is the initial State defined by the opener, it contains the expected allocation
     * @return channelId Unique identifier for the channel
     */
    function create(Channel calldata ch, State calldata initial) public returns (bytes32 channelId) {
        // Validate channel configuration
        if (ch.participants.length < 2) revert InvalidParticipant();
        if (ch.adjudicator == address(0)) revert InvalidAdjudicator();
        if (ch.challenge == 0) revert InvalidChallengePeriod();

        // Validate initial state for funding protocol
        (uint32 magicNumber) = abi.decode(initial.data, (uint32));
        // TODO: replace with `require(...)`
        if (magicNumber != CHANOPEN) revert InvalidState();

        // Generate channel ID and check it doesn't exist
        channelId = Utils.getChannelId(ch);
        if (_channels[channelId].stage != Status.VOID) revert InvalidStatus();

        // Verify creator's signature
        bytes32 stateHash = Utils.getStateHash(ch, initial);
        if (initial.sigs.length != 1) revert InvalidStateSignatures();
        // TODO: later we can lift the restriction that first sig must be from participant 0
        bool validSig = Utils.verifySignature(stateHash, initial.sigs[0], ch.participants[0]);
        if (!validSig) revert InvalidStateSignatures();

        // NOTE: even if there is not allocation planned, it should be present as `Allocation{address(0), 0}`
        if (initial.allocations.length != ch.participants.length) revert InvalidAllocations();

        // Initialize channel metadata
        Metadata storage meta = _channels[channelId];
        meta.chan = ch;
        meta.stage = Status.INITIAL;
        meta.creator = msg.sender;
        meta.lastValidState = initial;

        // NOTE: allocations MUST come in the same order as participants in deposit
        uint256 participantCount = ch.participants.length;
        for (uint256 i = 0; i < participantCount; i++) {
            address token = initial.allocations[i].token;
            uint256 amount = initial.allocations[i].amount;

            // even if participant does not have an allocation, still track that
            meta.expectedDeposits.push(Amount(token, amount));
            meta.actualDeposits.push(Amount(address(0), 0)); // Initialize actual deposits to zero
        }

        // NOTE: it is allowed for depositor (and msg.sender) to be different from channel creator (participant)
        // This enables logic of "session keys" where a user can create a channel on behalf of another account, but will lock their own funds
        // if (ch.participants[0]; != msg.sender) revert InvalidParticipant();

        Amount memory creatorDeposit = meta.expectedDeposits[0];
        _lockAccountFundsToChannel(msg.sender, channelId, creatorDeposit.token, creatorDeposit.amount);

        // Record actual deposit
        meta.actualDeposits[0] = creatorDeposit;

        // Add channel to the creator's ledger
        _ledgers[msg.sender].channels.add(channelId);

        // Emit event
        emit Created(channelId, ch, initial);

        return channelId;
    }

    /**
     * @notice Allows a participant to join a channel by signing the funding state
     * @param channelId Unique identifier for the channel
     * @param index Index of the participant in the channel's participants array
     * @param sig Signature of the participant on the funding state
     * @return The channelId of the joined channel
     */
    function join(bytes32 channelId, uint256 index, Signature calldata sig) external returns (bytes32) {
        Metadata storage meta = _channels[channelId];

        // Verify channel exists and is in INITIAL state
        if (meta.stage == Status.VOID) revert ChannelNotFound(channelId);
        if (meta.stage != Status.INITIAL) revert InvalidStatus();

        // Verify index is valid and participant has not already joined
        if (index == 0 || index >= meta.chan.participants.length) revert InvalidParticipant();
        if (meta.actualDeposits[index].amount != 0) revert InvalidParticipant();

        // Get participant address from channel config
        address participant = meta.chan.participants[index];

        // Verify signature on funding stateHash
        bytes32 stateHash = Utils.getStateHash(meta.chan, meta.lastValidState);
        bool validSig = Utils.verifySignature(stateHash, sig, participant);
        if (!validSig) revert InvalidStateSignatures();

        // Lock participant's funds according to expected deposit
        Amount memory expectedDeposit = meta.expectedDeposits[index];
        _lockAccountFundsToChannel(msg.sender, channelId, expectedDeposit.token, expectedDeposit.amount);

        // Record actual deposit
        meta.actualDeposits[index] = expectedDeposit;

        // Add channel to participant's ledger
        _ledgers[msg.sender].channels.add(channelId);

        // Emit joined event
        emit Joined(channelId, index);

        // Check if all participants have joined
        bool allJoined = true;
        for (uint256 i = 0; i < meta.actualDeposits.length; i++) {
            if (meta.actualDeposits[i].amount != meta.expectedDeposits[i].amount) {
                allJoined = false;
                break;
            }
        }

        // If all participants have joined, set channel to ACTIVE
        if (allJoined) {
            meta.stage = Status.ACTIVE;
            emit Opened(channelId);
        }

        return channelId;
    }

    /**
     * @notice Finalize the channel with a mutually signed state
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     */
    function close(bytes32 channelId, State calldata candidate, State[] calldata) public {
        Metadata storage meta = _channels[channelId];

        // Verify channel exists and is not VOID
        if (meta.stage == Status.VOID) revert ChannelNotFound(channelId);

        // Case 1: Mutual closing with CHANCLOSE magic number
        // Channel must not be in INITIAL stage (participants should close the channel with challenge then)
        if (meta.stage == Status.ACTIVE) {
            // Check that this is a closing state with CHANCLOSE magic number
            (uint32 magicNumber) = abi.decode(candidate.data, (uint32));
            if (magicNumber != CHANCLOSE) revert InvalidState();

            // Verify all participants have signed the closing state
            if (candidate.sigs.length != meta.chan.participants.length) revert InvalidStateSignatures();
            if (!_verifyAllSignatures(meta.chan, candidate)) revert InvalidStateSignatures();

            // Store the final state
            meta.lastValidState = candidate;
            meta.stage = Status.FINAL;
        }
        // Case 2: Challenge resolution (after challenge period expires)
        else if (meta.stage == Status.DISPUTE) {
            // Ensure challenge period has expired
            if (block.timestamp < meta.challengeExpire) revert ChallengeNotExpired();

            // Already in DISPUTE with an expired challenge - can proceed to finalization
            meta.stage = Status.FINAL;
        } else {
            revert InvalidStatus();
        }

        // At this point, the channel is in FINAL state, so we can close it
        _distributeAllocation(channelId, meta);

        // TODO: implement a better way for this
        // remove sender's channel in case they are a different account then participant
        _ledgers[msg.sender].channels.remove(channelId);
        uint256 participantsLength = meta.chan.participants.length;
        for (uint256 i = 0; i < participantsLength; i++) {
            address participant = meta.chan.participants[i];
            _ledgers[participant].channels.remove(channelId);
        }

        // Mark channel as closed by removing it
        delete _channels[channelId];

        emit ChannelClosed(channelId);
    }

    /**
     * @notice Unilaterally post a state when the other party is uncooperative
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    // TODO: add a challengerSig and check that it signed by either participant of the channel to disallow non-channel participants to challenge with stolen state
    function challenge(bytes32 channelId, State calldata candidate, State[] calldata proofs) external {
        Metadata storage meta = _channels[channelId];

        // Verify channel exists and is in a valid state for challenge
        if (meta.stage == Status.VOID) revert ChannelNotFound(channelId);
        if (meta.stage == Status.FINAL) revert InvalidStatus();

        // Verify that at least one participant signed the state
        if (candidate.sigs.length == 0) revert InvalidStateSignatures();

        // Verify the state is valid according to the adjudicator
        bool isValid = IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs);
        if (!isValid) revert InvalidState();

        // Revert if trying to challenge with an older state that is already known
        if (!_isMoreRecent(meta.chan.adjudicator, candidate, meta.lastValidState)) {
            revert InvalidState();
        }

        // Store the candidate as the last valid state
        meta.lastValidState = candidate;
        // Set or reset the challenge expiration
        meta.challengeExpire = block.timestamp + meta.chan.challenge;
        // Set the channel status to DISPUTE
        meta.stage = Status.DISPUTE;

        emit Challenged(channelId, meta.challengeExpire);
    }

    /**
     * @notice Unilaterally post a state to store it on-chain to prevent future disputes
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function checkpoint(bytes32 channelId, State calldata candidate, State[] calldata proofs) external {
        Metadata storage meta = _channels[channelId];

        // Verify channel exists and is not VOID or FINAL
        if (meta.stage == Status.VOID) revert ChannelNotFound(channelId);
        if (meta.stage == Status.FINAL) revert InvalidStatus();

        // Verify that at least one participant signed the state
        if (candidate.sigs.length == 0) revert InvalidStateSignatures();

        // Verify the state is valid according to the adjudicator
        bool isValid = IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs);
        if (!isValid) revert InvalidState();

        // Verify this state is more recent than the current stored state
        if (!_isMoreRecent(meta.chan.adjudicator, candidate, meta.lastValidState)) {
            revert InvalidState();
        }

        // Store the candidate as the last valid state
        meta.lastValidState = candidate;

        // If there's an ongoing challenge and this state is newer, cancel the challenge
        if (meta.stage == Status.DISPUTE) {
            meta.stage = Status.ACTIVE;
            meta.challengeExpire = 0;
        }

        emit Checkpointed(channelId);
    }

    /**
     * @notice Reset will close and open channel for resizing allocations
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs An array of valid state required by the adjudicator
     * @param newChannel New channel configuration
     * @param newDeposit Initial State defined by the opener, containing the expected allocation
     */
    function reset(
        bytes32 channelId,
        State calldata candidate,
        State[] calldata proofs,
        Channel calldata newChannel,
        State calldata newDeposit
    ) external {
        // First close the existing channel
        close(channelId, candidate, proofs);

        // Then open a new channel with the provided configuration
        create(newChannel, newDeposit);
    }

    /**
     * @notice Internal function to close a channel and distribute funds
     * @param channelId The channel identifier
     * @param meta The channel's metadata
     */
    function _distributeAllocation(bytes32 channelId, Metadata storage meta) internal {
        // Distribute funds according to allocations
        uint256 allocsLength = meta.lastValidState.allocations.length;
        for (uint256 i = 0; i < allocsLength; i++) {
            Allocation memory allocation = meta.lastValidState.allocations[i];
            _unlockChannelFundsToAccount(channelId, allocation.destination, allocation.token, allocation.amount);
        }
    }

    /**
     * @notice Helper function to compare two states for recency
     * @param adjudicator The adjudicator contract address
     * @param candidate The candidate state
     * @param previous The previous state to compare against
     * @return True if the candidate state is more recent than the previous state
     */
    function _isMoreRecent(address adjudicator, State calldata candidate, State memory previous)
        internal
        view
        returns (bool)
    {
        return IComparable(adjudicator).compare(candidate, previous) > 0;
    }

    function _transfer(address token, address to, uint256 amount) internal {
        bool success;
        if (token == address(0)) {
            (success,) = to.call{value: amount}("");
        } else {
            success = IERC20(token).transfer(to, amount);
        }
        if (!success) revert TransferFailed(token, to, amount);
    }

    function _lockAccountFundsToChannel(address account, bytes32 channelId, address token, uint256 amount) internal {
        if (amount == 0) return;

        Ledger storage ledger = _ledgers[account];
        uint256 available = ledger.tokens[token].available;
        if (available < amount) revert InsufficientBalance(available, amount);

        ledger.tokens[token].available -= amount;
        ledger.tokens[token].locked += amount;

        Metadata storage meta = _channels[channelId];
        meta.tokenBalances[token] += amount;
    }

    // Does not perform checks to allow transferring partial balances in case of partial deposit
    function _unlockChannelFundsToAccount(bytes32 channelId, address account, address token, uint256 amount) internal {
        if (amount == 0) return;

        Metadata storage meta = _channels[channelId];
        uint256 channelBalance = meta.tokenBalances[token];
        if (channelBalance == 0) return;

        uint256 correctedAmount = channelBalance > amount ? amount : channelBalance;
        meta.tokenBalances[token] -= correctedAmount;

        Ledger storage ledger = _ledgers[account];

        // Check locked amount before subtracting to prevent underflow
        uint256 lockedAmount = ledger.tokens[token].locked;
        uint256 amountToUnlock = lockedAmount > correctedAmount ? correctedAmount : lockedAmount;

        if (amountToUnlock > 0) {
            ledger.tokens[token].locked -= amountToUnlock;
        }
        ledger.tokens[token].available += amountToUnlock;
    }

    /**
     * @notice Verifies that all provided signatures are valid for the given state
     * @param chan The channel configuration
     * @param state The state to verify signatures for
     * @return valid True if all provided signatures are valid
     */
    function _verifyAllSignatures(Channel memory chan, State memory state) internal pure returns (bool valid) {
        // Calculate the state hash once
        bytes32 stateHash = Utils.getStateHash(chan, state);

        // Check if we have the right number of signatures
        if (state.sigs.length > chan.participants.length) {
            return false;
        }

        // Verify each signature
        for (uint256 i = 0; i < state.sigs.length; i++) {
            bool isValid = Utils.verifySignature(stateHash, state.sigs[i], chan.participants[i]);
            if (!isValid) {
                return false;
            }
        }

        return true;
    }
}
