// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {IChannel} from "./interfaces/IChannel.sol";
import {IDeposit} from "./interfaces/IDeposit.sol";
import {IAdjudicator} from "./interfaces/IAdjudicator.sol";
import {IComparable} from "./interfaces/IComparable.sol";
import {Channel, State, Allocation, ChannelStatus, StateIntent, Signature, Amount} from "./interfaces/Types.sol";
import {Utils} from "./Utils.sol";
import {IERC20} from "lib/openzeppelin-contracts/contracts/interfaces/IERC20.sol";
import {SafeERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/utils/SafeERC20.sol";
import {EnumerableSet} from "lib/openzeppelin-contracts/contracts/utils/structs/EnumerableSet.sol";

/**
 * @title Custody
 * @notice A simple custody contract for state channels that delegates most state transition logic to an adjudicator
 * @dev This implementation currently only supports 2 participant channels (CLIENT and SERVER)
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

    // Constants for participant indices
    uint256 constant CLIENT = 0; // Participant index for the channel creator
    uint256 constant SERVER = 1; // Participant index for the server in clearnet context

    // Recommended structure to keep track of states
    struct Metadata {
        Channel chan; // Opener define channel configuration
        ChannelStatus stage;
        address creator;
        // Fixed arrays for exactly 2 participants (CLIENT and SERVER)
        // TODO: store `uint256` instead of `Amount`, as tokens are the same
        Amount[2] expectedDeposits; // CLIENT defines Token per participant
        Amount[2] actualDeposits; // Tracks deposits made by each participant
        uint256 challengeExpire; // If non-zero channel will resolve to lastValidState when challenge Expires
        State lastValidState; // Last valid state when adjudicator was called
        mapping(address token => uint256 balance) tokenBalances; // Token balances for the channel
    }

    struct Ledger {
        mapping(address token => uint256 available) tokens; // Available amount that can be withdrawn or allocated to channels
        EnumerableSet.Bytes32Set channels; // Set of user ChannelId
    }

    mapping(bytes32 channelId => Metadata chMeta) internal _channels;
    mapping(address account => Ledger ledger) internal _ledgers;

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
     * @return channelCount Number of associated channels
     */
    function getAccountInfo(address user, address token)
        public
        view
        returns (uint256 available, uint256 channelCount)
    {
        Ledger storage ledger = _ledgers[user];
        return (ledger.tokens[token], ledger.channels.length());
    }

    function deposit(address token, uint256 amount) external payable {
        address account = msg.sender;
        if (token == address(0)) {
            if (msg.value != amount) revert InvalidAmount();
        } else {
            IERC20(token).safeTransferFrom(account, address(this), amount);
        }
        _ledgers[msg.sender].tokens[token] += amount;
    }

    function withdraw(address token, uint256 amount) external {
        Ledger storage ledger = _ledgers[msg.sender];
        uint256 available = ledger.tokens[token];
        if (available < amount) revert InsufficientBalance(available, amount);
        _transfer(token, msg.sender, amount);
        ledger.tokens[token] -= amount;
    }

    /**
     * @notice Create a channel by depositing assets
     * @param ch Channel configuration
     * @param initial is the initial State defined by the opener, it contains the expected allocation
     * @return channelId Unique identifier for the channel
     */
    function create(Channel calldata ch, State calldata initial) public returns (bytes32 channelId) {
        // TODO: add checks that there are only 2 allocations, they have the same token (here and throughout the code)
        // Validate channel configuration
        if (ch.participants.length != 2) revert InvalidParticipant();
        if (ch.adjudicator == address(0)) revert InvalidAdjudicator();
        if (ch.challenge == 0) revert InvalidChallengePeriod();

        // TODO: replace with `require(...)`
        if (initial.intent != StateIntent.INITIALIZE) revert InvalidState();

        // Validate version must be 0 for INITIAL state
        if (initial.version != 0) revert InvalidState();

        // Generate channel ID and check it doesn't exist
        channelId = Utils.getChannelId(ch);
        if (_channels[channelId].stage != ChannelStatus.VOID) revert InvalidStatus();

        // Verify creator's signature
        bytes32 stateHash = Utils.getStateHash(ch, initial);
        if (initial.sigs.length != 1) revert InvalidStateSignatures();
        // TODO: later we can lift the restriction that first sig must be from participant 0
        bool validSig = Utils.verifySignature(stateHash, initial.sigs[CLIENT], ch.participants[CLIENT]);
        if (!validSig) revert InvalidStateSignatures();

        // NOTE: even if there is not allocation planned, it should be present as `Allocation{address(0), 0}`
        if (initial.allocations.length != ch.participants.length) revert InvalidAllocations();

        // Initialize channel metadata
        Metadata storage meta = _channels[channelId];
        meta.chan = ch;
        meta.stage = ChannelStatus.INITIAL;
        meta.creator = msg.sender;
        meta.lastValidState = initial;

        // NOTE: allocations MUST come in the same order as participants in deposit
        uint256 participantCount = ch.participants.length;
        for (uint256 i = 0; i < participantCount; i++) {
            address token = initial.allocations[i].token;
            uint256 amount = initial.allocations[i].amount;

            // even if participant does not have an allocation, still track that
            meta.expectedDeposits[i] = Amount(token, amount);
            meta.actualDeposits[i] = Amount(address(0), 0); // Initialize actual deposits to zero
        }

        // NOTE: it is allowed for depositor (and msg.sender) to be different from channel creator (participant)
        // This enables logic of "session keys" where a user can create a channel on behalf of another account, but will lock their own funds
        // if (ch.participants[0]; != msg.sender) revert InvalidParticipant();

        Amount memory creatorDeposit = meta.expectedDeposits[CLIENT];
        _lockAccountFundsToChannel(msg.sender, channelId, creatorDeposit.token, creatorDeposit.amount);

        // Record actual deposit
        meta.actualDeposits[CLIENT] = creatorDeposit;

        // NOTE: channel is added to PARTICIPANT's ledger
        _ledgers[ch.participants[CLIENT]].channels.add(channelId);

        // Emit event
        emit Created(channelId, ch, initial);

        return channelId;
    }

    /**
     * @notice Allows a participant to join a channel by signing the funding state
     * @param channelId Unique identifier for the channel
     * @param index Index of the participant in the channel's participants array (must be 1 for SERVER)
     * @param sig Signature of the participant on the funding state
     * @return The channelId of the joined channel
     */
    function join(bytes32 channelId, uint256 index, Signature calldata sig) external returns (bytes32) {
        Metadata storage meta = _channels[channelId];

        // Verify channel exists and is in INITIAL state
        if (meta.stage == ChannelStatus.VOID) revert ChannelNotFound(channelId);
        if (meta.stage != ChannelStatus.INITIAL) revert InvalidStatus();

        // Verify index is valid and participant has not already joined
        // For 2-participant channels, index can only be SERVER (second participant)
        if (index != SERVER) revert InvalidParticipant();
        if (meta.actualDeposits[index].amount != 0) revert InvalidParticipant();

        // Get participant address from channel config
        address participant = meta.chan.participants[index];

        // Verify signature on funding stateHash
        bytes32 stateHash = Utils.getStateHash(meta.chan, meta.lastValidState);
        if (!Utils.verifySignature(stateHash, sig, participant)) revert InvalidStateSignatures();

        // NOTE: lock SENDER's funds according to expected deposit
        Amount memory expectedDeposit = meta.expectedDeposits[index];
        _lockAccountFundsToChannel(msg.sender, channelId, expectedDeposit.token, expectedDeposit.amount);

        // Record actual deposit
        meta.actualDeposits[index] = expectedDeposit;

        // NOTE: channel is added to PARTICIPANT's ledger
        _ledgers[meta.chan.participants[SERVER]].channels.add(channelId);

        // Emit joined event
        emit Joined(channelId, index);

        // For 2-participant channels, just check if the second participant has joined
        // since we know the first participant (CLIENT) has already joined
        bool allJoined = meta.actualDeposits[SERVER].amount == meta.expectedDeposits[SERVER].amount;

        // If all participants have joined, set channel to ACTIVE
        if (allJoined) {
            meta.stage = ChannelStatus.ACTIVE;
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
        if (meta.stage == ChannelStatus.VOID) revert ChannelNotFound(channelId);

        // Case 1: Mutual closing with CHANCLOSE magic number
        // Channel must not be in INITIAL stage (participants should close the channel with challenge then)
        if (meta.stage == ChannelStatus.ACTIVE) {
            // Check that this is a closing state with CHANCLOSE magic number
            if (candidate.intent != StateIntent.FINALIZE) revert InvalidState();

            // For ACTIVE channels, version must be greater than 0
            if (candidate.version == 0) revert InvalidState();

            // Verify all participants have signed the closing state
            // For our 2-participant channels, we need exactly 2 signatures
            if (candidate.sigs.length != 2) revert InvalidStateSignatures();
            if (!_verifyAllSignatures(meta.chan, candidate)) revert InvalidStateSignatures();

            // Store the final state
            meta.lastValidState = candidate;
            meta.stage = ChannelStatus.FINAL;
        }
        // Case 2: Challenge resolution (after challenge period expires)
        else if (meta.stage == ChannelStatus.DISPUTE) {
            // Ensure challenge period has expired
            if (block.timestamp < meta.challengeExpire) revert ChallengeNotExpired();

            // Already in DISPUTE with an expired challenge - can proceed to finalization
            meta.stage = ChannelStatus.FINAL;
        } else {
            revert InvalidStatus();
        }

        // At this point, the channel is in FINAL state, so we can close it
        _unlockAllocations(channelId, candidate.allocations);

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

        emit Closed(channelId);
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
        if (meta.stage == ChannelStatus.VOID) revert ChannelNotFound(channelId);
        if (meta.stage == ChannelStatus.FINAL) revert InvalidStatus();

        // Verify that at least one participant signed the state
        if (candidate.sigs.length == 0) revert InvalidStateSignatures();

        // Validate version based on channel status
        if (meta.stage == ChannelStatus.INITIAL && candidate.version != 0) revert InvalidState();

        if (candidate.data.length != 0) {
            if (candidate.intent == StateIntent.INITIALIZE) {
                // TODO:
            } else if (candidate.intent == StateIntent.RESIZE) {
                uint256 deposited = meta.expectedDeposits[CLIENT].amount + meta.expectedDeposits[SERVER].amount;
                uint256 expected = candidate.allocations[CLIENT].amount + candidate.allocations[SERVER].amount;
                if (deposited != expected) {
                    revert InvalidAllocations();
                }
            }
        }

        // ([("0xe601dfaFbDBc16a4d997d04fCEE655Caab831798","0x3c93C321634a80FB3657CFAC707718A11cA57cBf"],"0xC2BA5c5E2c4848F64187Aa1F3f32a331b0C031b9","1","2987423159340149059"),("1","0","0x",[("0x21F7D1F35979B125f6F7918fC16Cb9101e5882d7","0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359","2000"),("0x3c93C321634a80FB3657CFAC707718A11cA57cBf","0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359","0")],[("28","0x7b112e1433ec0146568c95d98f6a087d9db2c5c9b63731c05ebaefbcfd68afc4","0x7fd74770a1189b5e1f3b0fc03eadaaf8dd78f942700a62d98b1674ccd38d48a8")]),[])

        if (
            candidate.data.length == 0
                || (candidate.intent != StateIntent.INITIALIZE && candidate.intent != StateIntent.RESIZE)
        ) {
            // if no state data or magic number is not CHANOPEN or CHANRESIZE, assume this is a normal state

            // Verify the state is valid according to the adjudicator
            bool isValid = IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs);
            if (!isValid) revert InvalidState();

            // Reject states with equal version
            if (candidate.version == meta.lastValidState.version) {
                // Explicitly check for equal versions and reject them
                revert InvalidState();
            }

            // Revert if trying to challenge with an older state that is already known
            if (!_isMoreRecent(meta.chan.adjudicator, candidate, meta.lastValidState)) {
                revert InvalidState();
            }
        }

        // Store the candidate as the last valid state
        meta.lastValidState = candidate;
        // Set or reset the challenge expiration
        meta.challengeExpire = block.timestamp + meta.chan.challenge;
        // Set the channel status to DISPUTE
        meta.stage = ChannelStatus.DISPUTE;

        emit Challenged(channelId, meta.challengeExpire);
    }

    /**
     * @notice Unilaterally post a state to store it on-chain to prevent future disputes
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    // TODO: add responding to CHANOPEN, CHANRESIZE challenge (should NOT call `adjudicate`)
    function checkpoint(bytes32 channelId, State calldata candidate, State[] calldata proofs) external {
        Metadata storage meta = _channels[channelId];

        // Verify channel exists and is not VOID or FINAL
        if (meta.stage == ChannelStatus.VOID) revert ChannelNotFound(channelId);
        if (meta.stage == ChannelStatus.FINAL) revert InvalidStatus();

        // Verify that at least one participant signed the state
        if (candidate.sigs.length == 0) revert InvalidStateSignatures();

        // Validate version based on channel status
        if (meta.stage == ChannelStatus.INITIAL && candidate.version != 0) revert InvalidState();

        // Verify the state is valid according to the adjudicator
        bool isValid = IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs);
        if (!isValid) revert InvalidState();

        // Verify this state is more recent than the current stored state
        if (candidate.version == meta.lastValidState.version) {
            // Explicitly check for equal versions and reject them
            revert InvalidState();
        }

        if (!_isMoreRecent(meta.chan.adjudicator, candidate, meta.lastValidState)) {
            revert InvalidState();
        }

        // Store the candidate as the last valid state
        meta.lastValidState = candidate;

        // If there's an ongoing challenge and this state is newer, cancel the challenge
        if (meta.stage == ChannelStatus.DISPUTE) {
            meta.stage = ChannelStatus.ACTIVE;
            meta.challengeExpire = 0;
        }

        emit Checkpointed(channelId);
    }

    /**
     * @notice All participants agree in setting a new allocation resulting in locking or unlocking funds
     * @dev Used for resizing channel allocations without withdrawing funds
     * @param channelId Unique identifier for the channel to resize
     * @param candidate The state that is to be true after resizing, containing the delta allocations
     * @param proofs An array of states supporting the claim that the candidate is true
     * NOTE: proof is needed to improve UX and allow resized state to follow any state (no need for consensus)
     */
    function resize(bytes32 channelId, State calldata candidate, State[] calldata proofs) external {
        Metadata storage meta = _channels[channelId];

        // Verify channel exists and is ACTIVE
        if (meta.stage == ChannelStatus.VOID) revert ChannelNotFound(channelId);
        if (meta.stage != ChannelStatus.ACTIVE) revert InvalidStatus();

        if (proofs.length == 0) revert InvalidState();
        State memory precedingState = proofs[0];
        // NOTE: this is required as `proofs[0:]` over arrays of dynamic types (State is dynamic) is not supported by Solidity compiler as of 0.8.29.
        State[] memory precedingProofs = new State[](proofs.length - 1);
        for (uint256 i = 1; i < proofs.length; i++) {
            precedingProofs[i - 1] = proofs[i];
        }

        // Verify the preceding state is valid according to the adjudicator
        bool isValid = IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, precedingState, precedingProofs);
        if (!isValid) revert InvalidState();

        // resized state should be the successor of the preceding state
        if (candidate.version != precedingState.version + 1) revert InvalidState();

        _requireCorrectAllocations(precedingState.allocations);
        _requireCorrectAllocations(candidate.allocations);

        // Verify all participants have signed the resize state
        if (!_verifyAllSignatures(meta.chan, candidate)) revert InvalidStateSignatures();

        // Decode the magic number and resize amounts
        // TODO: extract `int256[]` into an alias type
        int256[] memory resizeAmounts = abi.decode(candidate.data, (int256[]));

        if (candidate.intent != StateIntent.RESIZE) revert InvalidState();

        _requireCorrectDelta(precedingState.allocations, candidate.allocations, resizeAmounts);

        _processResize(channelId, meta, resizeAmounts, candidate.allocations);

        // Update the latest valid state
        meta.lastValidState = candidate;

        emit Resized(channelId, resizeAmounts);
    }

    function _requireCorrectAllocations(Allocation[] memory allocations) internal pure {
        if (allocations.length != 2) revert InvalidState();
        if (allocations[CLIENT].token != allocations[SERVER].token) revert InvalidState();
    }

    function _transfer(address token, address to, uint256 amount) internal {
        bool success;
        if (token == address(0)) {
            (success,) = to.call{value: amount}("");
            if (!success) revert TransferFailed(token, to, amount);
        } else {
            IERC20(token).safeTransfer(to, amount);
        }
    }

    /**
     * @notice Lock funds from an account to a channel
     * @dev Used during channel creation and joining for 2-participant channels
     */
    function _lockAccountFundsToChannel(address account, bytes32 channelId, address token, uint256 amount) internal {
        if (amount == 0) return;

        Ledger storage ledger = _ledgers[account];
        uint256 available = ledger.tokens[token];
        if (available < amount) revert InsufficientBalance(available, amount);

        ledger.tokens[token] -= amount;
        _channels[channelId].tokenBalances[token] += amount;
    }

    /**
     * @notice Internal function to close a channel and distribute funds
     * @param channelId The channel identifier
     * @param allocations The allocations to distribute
     */
    function _unlockAllocations(bytes32 channelId, Allocation[] memory allocations) internal {
        // Distribute funds according to allocations
        uint256 allocsLength = allocations.length;
        for (uint256 i = 0; i < allocsLength; i++) {
            _unlockAllocation(channelId, allocations[i]);
        }
    }

    // Does not perform checks to allow transferring partial balances in case of partial deposit
    function _unlockAllocation(bytes32 channelId, Allocation memory alloc) internal {
        if (alloc.amount == 0) return;

        Metadata storage meta = _channels[channelId];
        uint256 channelBalance = meta.tokenBalances[alloc.token];
        if (channelBalance == 0) return;

        uint256 correctedAmount = channelBalance > alloc.amount ? alloc.amount : channelBalance;
        meta.tokenBalances[alloc.token] -= correctedAmount;
        _ledgers[alloc.destination].tokens[alloc.token] += correctedAmount;
    }

    /**
     * @notice Verifies that both signatures are valid for the given state in a 2-participant channel
     * @param chan The channel configuration
     * @param state The state to verify signatures for
     * @return valid True if both signatures are valid
     */
    function _verifyAllSignatures(Channel memory chan, State memory state) internal view returns (bool valid) {
        // Calculate the state hash once
        bytes32 stateHash = Utils.getStateHash(chan, state);

        // Check if we have exactly 2 signatures for our 2-participant channels
        if (state.sigs.length != 2 || chan.participants.length != 2) {
            return false;
        }

        // Verify client's signature
        bool isClientValid = Utils.verifySignature(stateHash, state.sigs[CLIENT], chan.participants[CLIENT]);
        if (!isClientValid) {
            return false;
        }

        // Verify server's signature
        bool isServerValid = Utils.verifySignature(stateHash, state.sigs[SERVER], chan.participants[SERVER]);
        if (!isServerValid) {
            return false;
        }

        return true;
    }

    /**
     * @notice Helper function to compare two states for recency
     * @param adjudicator The adjudicator contract address
     * @param candidate The candidate state
     * @param previous The previous state to compare against
     * @return True if the candidate state is strictly more recent than the previous state
     * @dev Returns false if states have equal version numbers or if candidate is older
     */
    function _isMoreRecent(address adjudicator, State calldata candidate, State memory previous)
        internal
        view
        returns (bool)
    {
        // TODO: add support to ERC-165
        // Try to use IComparable if the adjudicator implements it
        try IComparable(adjudicator).compare(candidate, previous) returns (int8 result) {
            // Must return strictly positive result (>0), equal versions (==0) are not considered more recent
            return result > 0;
        } catch {
            // If IComparable is not implemented, fall back to comparing version numbers
            // Must be strictly greater, equal versions are not considered more recent
            return candidate.version > previous.version;
        }
    }

    /// @notice Allows "implicit transfer" between CLIENT and SERVER, which is useful in situations where
    /// a participant wants to top-up a channel only to transfer funds to the other participant, so they can withdraw it
    /// @dev "implicit transfer" means that only the sum of "initial + resize == final" is checked, not the individual amounts_channels
    /// Explicit delta can be calculated as |final[i] - initial[i] - resize[i]|, where i can be CLIENT or SERVER
    function _requireCorrectDelta(
        Allocation[] memory initialAllocations,
        Allocation[] memory finalAllocations,
        int256[] memory delta
    ) internal pure {
        if (delta.length != 2) revert InvalidState();

        uint256 sumBefore = initialAllocations[CLIENT].amount + initialAllocations[SERVER].amount;
        int256 sumDelta = delta[CLIENT] + delta[SERVER];
        uint256 sumAfter = finalAllocations[CLIENT].amount + finalAllocations[SERVER].amount;

        if (int256(sumBefore) + sumDelta != int256(sumAfter)) {
            revert InvalidAllocations();
        }
    }

    /// @notice Supports "implicit transfer"
    /// @dev Positive deltas must be processed first as they add more funds to the channel that the negative delta may want to withdraw
    function _processResize(bytes32 channelId, Metadata storage chMeta, int256[] memory resizeAmounts, Allocation[] memory finalAllocations) internal {
        // NOTE: all tokens are the same
        address token = chMeta.expectedDeposits[CLIENT].token;

        // First pass: Process all positive resizes
        for (uint256 i = 0; i < 2; i++) {
            if (resizeAmounts[i] > 0) {
                _lockAccountFundsToChannel(chMeta.chan.participants[i], channelId, token, uint256(resizeAmounts[i]));
            }
        }

        // Second pass: Process all negative resizes
        for (uint256 i = 0; i < 2; i++) {
            if (resizeAmounts[i] < 0) {
                _unlockAllocation(channelId, Allocation(chMeta.chan.participants[i], token, uint256(-resizeAmounts[i])));
            }
        }

        for (uint256 i = 0; i < 2; i++) {
            chMeta.expectedDeposits[i].amount = finalAllocations[i].amount;
            chMeta.actualDeposits[i].amount = finalAllocations[i].amount;
        }
    }
}
