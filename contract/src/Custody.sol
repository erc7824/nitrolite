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
    // TODO: sort errors
    error ChannelNotFound(bytes32 channelId);
    error ChannelNotFinal();
    error InvalidParticipant();
    error InvalidStatus();
    error InvalidState();
    error InvalidAllocations();
    error DepositAlreadyFulfilled();
    error DepositsNotFulfilled(uint256 expectedFulfilled, uint256 actualFulfilled);
    error InvalidStateSignatures();
    error InvalidAdjudicator();
    error InvalidChallengerSignature();
    error InvalidChallengePeriod();
    error InvalidValue();
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
        Amount[2] expectedDeposits; // Creator defines Token per participant
        Amount[2] actualDeposits; // Tracks deposits made by each participant
        uint256 challengeExpire; // If non-zero channel will resolve to lastValidState when challenge Expires
        State lastValidState; // Last valid state when adjudicator was called
        mapping(address token => uint256 balance) tokenBalances; // Token balances for the channel
    }

    struct Ledger {
        mapping(address token => uint256 available) tokens; // Available amount that can be withdrawn or allocated to channels
        EnumerableSet.Bytes32Set channels; // Set of user ChannelId
    }

    // Custody contract restricts number of participants to 2
    uint256 constant PART_NUM = 2;
    uint256 constant CLIENT_IDX = 0; // Participant index for the channel creator
    uint256 constant SERVER_IDX = 1; // Participant index for the server in clearnet context


    mapping(bytes32 channelId => Metadata chMeta) internal _channels;
    mapping(address account => Ledger ledger) internal _ledgers;

    function deposit(address token, uint256 amount) external payable {
        address account = msg.sender;
        if (token == address(0)) {
            if (msg.value != amount) revert InvalidValue();
        } else {
            if (msg.value != 0) revert InvalidValue();
            if (!IERC20(token).transferFrom(account, address(this), amount)) {
                revert TransferFailed(token, address(this), amount);
            }
        }

        _ledgers[msg.sender].tokens[token] += amount;
    }

    function withdraw(address token, uint256 amount) external {
        Ledger storage ledger = _ledgers[msg.sender];
        uint256 available = ledger.tokens[token];
        if (available < amount) revert InsufficientBalance(available, amount);

        ledger.tokens[token] -= amount;

        _transfer(token, msg.sender, amount);
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

    /**
     * @notice Create a channel by depositing assets
     * @dev CLIENT and SERVER had NO prior agreement on the channel parameters. SERVER itself decides whether to join or not.
     * In the same time, it should be noted that SERVER should have a manifest of channel parameters it agrees to join automatically.
     * @param ch Channel configuration
     * @param initial is the initial State defined by the opener, it contains the expected allocation
     * @return channelId Unique identifier for the channel
     */
    function create(Channel calldata ch, State calldata initial) public returns (bytes32 channelId) {
        // Validate channel configuration
        if (ch.participants.length != PART_NUM ||
                ch.participants[CLIENT_IDX] == address(0) ||
                ch.participants[SERVER_IDX] == address(0) ||
                ch.participants[CLIENT_IDX] == ch.participants[SERVER_IDX]
        ) revert InvalidParticipant();
        if (ch.adjudicator == address(0)) revert InvalidAdjudicator();
        if (ch.challenge == 0) revert InvalidChallengePeriod();

        // Validate initial state for funding protocol
        uint32 magicNumber = abi.decode(initial.data, (uint32));
        // TODO: replace with `require(...)`
        if (magicNumber != CHANOPEN) revert InvalidState();

        // Generate channel ID and check it doesn't exist
        channelId = Utils.getChannelId(ch);
        if (_channels[channelId].stage != Status.VOID) revert InvalidStatus();

        // Verify creator's signature
        bytes32 stateHash = Utils.getStateHash(ch, initial);
        if (initial.sigs.length != 1) revert InvalidStateSignatures();
        // TODO: later we can lift the restriction that first sig must be from CLIENT
        if (!Utils.verifySignature(stateHash, initial.sigs[CLIENT_IDX], ch.participants[CLIENT_IDX])) {
            revert InvalidStateSignatures();
        }

        // NOTE: even if there is not allocation planned, it should be present as `Allocation{address(0), 0}`
        if (initial.allocations.length != ch.participants.length) revert InvalidAllocations();

        // Initialize channel metadata
        Metadata storage meta = _channels[channelId];
        meta.chan = ch;
        meta.stage = Status.INITIAL;
        meta.creator = msg.sender;
        meta.lastValidState = initial;

        // NOTE: allocations MUST come in the same order as participants in deposit
        for (uint256 i = 0; i < PART_NUM; i++) {
            address token = initial.allocations[i].token;
            uint256 amount = initial.allocations[i].amount;

            // even if participant does not have an allocation, still track that
            meta.expectedDeposits[i] = Amount({token: token, amount: amount});
            meta.actualDeposits[i] = Amount({token: address(0), amount: 0}); // Initialize actual deposits to zero
        }

        // NOTE: it is allowed for depositor (and msg.sender) to be different from channel creator (participant)
        // This enables logic of "session keys" where a user can create a channel on behalf of another account, but will lock their own funds
        // if (ch.participants[CLIENT_IDX]; != msg.sender) revert InvalidParticipant();

        Amount memory creatorDeposit = meta.expectedDeposits[CLIENT_IDX];
        _lockAccountFundsToChannel(msg.sender, channelId, creatorDeposit.token, creatorDeposit.amount);

        // Record actual deposit
        meta.actualDeposits[CLIENT_IDX] = creatorDeposit;

        // Add channel to the creator's ledger
        _ledgers[ch.participants[CLIENT_IDX]].channels.add(channelId);

        // Emit event
        emit Created(channelId, ch, initial);

        return channelId;
    }

    /**
     * @notice Allows a SERVER to join a channel by signing the funding state
     * @dev CLIENT and SERVER had NO prior agreement on the channel parameters. SERVER itself decides whether to join or not.
     * In the same time, it should be noted that SERVER should have a manifest of channel parameters it agrees to join automatically.
     * @param channelId Unique identifier for the channel
     * @param index Index of the participant in the channel's participants array (must be 1 for SERVER)
     * @param sig Signature of SERVER on the funding state
     * @return The channelId of the joined channel
     */
    function join(bytes32 channelId, uint256 index, Signature calldata sig) external returns (bytes32) {
        Metadata storage meta = _channels[channelId];

        // Verify channel exists and is in INITIAL state
        if (meta.stage == Status.VOID) revert ChannelNotFound(channelId);
        // allow joining after previous participant has either joined or challenged the fact that the next participant is not joining
        if (meta.stage != Status.INITIAL && meta.stage != Status.DISPUTE) revert InvalidStatus();

        // Verify index is a SERVER index
        if (index != SERVER_IDX) revert InvalidParticipant();
        // forbid joining several times
        if (meta.actualDeposits[SERVER_IDX].amount != 0) revert DepositAlreadyFulfilled();

        // Verify SERVER signature on funding stateHash
        bytes32 stateHash = Utils.getStateHash(meta.chan, meta.lastValidState);
        bool validSig = Utils.verifySignature(stateHash, sig, meta.chan.participants[SERVER_IDX]);
        if (!validSig) revert InvalidStateSignatures();
        // add signature to the state
        meta.lastValidState.sigs.push(sig);

        // Lock SERVER's funds according to expected deposit
        Amount memory expectedDeposit = meta.expectedDeposits[SERVER_IDX];
        _lockAccountFundsToChannel(msg.sender, channelId, expectedDeposit.token, expectedDeposit.amount);

        // Record actual deposit
        meta.actualDeposits[SERVER_IDX] = expectedDeposit;
        meta.challengeExpire = 0; // Reset challenge expiration

        // Add channel to participant's ledger
        _ledgers[meta.chan.participants[SERVER_IDX]].channels.add(channelId);

        meta.stage = Status.ACTIVE;

        // Emit joined event
        emit Joined(channelId, SERVER_IDX);
        emit Opened(channelId);

        return channelId;
    }

    /**
     * @notice Finalize the channel with a mutually signed state
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * NOTE: Custody implementation does NOT require the `proofs` parameter for the close function.
     */
    function close(bytes32 channelId, State calldata candidate, State[] calldata) public {
        Metadata storage meta = _channels[channelId];

        if (meta.stage == Status.VOID) revert ChannelNotFound(channelId);
        if (meta.stage == Status.INITIAL || meta.stage == Status.FINAL) revert InvalidStatus();

        // Channel must not be in INITIAL stage (participants should close the channel with challenge then)
        if (meta.stage == Status.ACTIVE) {
            uint32 magicNumber = abi.decode(candidate.data, (uint32));
            if (magicNumber != CHANCLOSE) revert InvalidState();

            if (!_verifyAllSignatures(meta.chan, candidate)) revert InvalidStateSignatures();

            meta.lastValidState = candidate;
            meta.stage = Status.FINAL;
        } else { //meta.stage == Status.DISPUTE
            // Can overwrite any challenge state with a valid final state
            if (block.timestamp < meta.challengeExpire) {
                uint32 magicNumber = abi.decode(candidate.data, (uint32));
                if (magicNumber != CHANCLOSE) revert InvalidState();

                if (!_verifyAllSignatures(meta.chan, candidate)) revert InvalidStateSignatures();

                meta.challengeExpire = 0;
                meta.lastValidState = candidate;
                meta.stage = Status.FINAL;
            } else {
                // Already in DISPUTE with an expired challenge - can proceed to finalization
                meta.stage = Status.FINAL;
            }
        }

        _unlockAllocations(channelId, candidate.allocations);

        // TODO: implement a better way for this
        // remove sender's channel in case they are a different account then participant
        _ledgers[msg.sender].channels.remove(channelId);
        for (uint256 i = 0; i < PART_NUM; i++) {
            address participant = meta.chan.participants[i];
            _ledgers[participant].channels.remove(channelId);
        }

        delete _channels[channelId];

        emit ChannelClosed(channelId);
    }

    /**
     * @notice Unilaterally post a state when the other party is uncooperative
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     * @param challengerSig Challenger signature over `keccak256(abi.encode(stateHash, "challenge"))` to disallow 3rd party
     * to challenge with a stolen state and its signature
     */
    function challenge(bytes32 channelId, State calldata candidate, State[] calldata proofs, Signature calldata challengerSig) external {
        Metadata storage meta = _channels[channelId];

        // Verify channel exists and is in a valid state for challenge
        if (meta.stage == Status.VOID) revert ChannelNotFound(channelId);
        if (meta.stage == Status.FINAL) revert InvalidStatus();

        _requireChallengerIsParticipant(
            Utils.getStateHash(meta.chan, candidate),
            [meta.chan.participants[CLIENT_IDX], meta.chan.participants[SERVER_IDX]],
            challengerSig
        );

        uint32 candidateMagicNumber = abi.decode(candidate.data, (uint32));
        if (candidateMagicNumber == CHANCLOSE) revert InvalidState();

        uint32 lastValidStateMagicNumber = abi.decode(meta.lastValidState.data, (uint32));

        if (meta.stage == Status.INITIAL) {
            // main goal: verify Candidate is valid and >= LastValidState
            if (candidateMagicNumber == CHANOPEN) {
                if (proofs.length != 0) revert InvalidState();
                if (candidate.sigs.length < meta.lastValidState.sigs.length) revert InvalidState();
                _requireValidSignatures(meta.chan, candidate);
                _requireChannelHasNFulfilledDeposits(channelId, candidate.sigs.length);
            } else {
                _requireChannelHasNFulfilledDeposits(channelId, PART_NUM);
                if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) revert InvalidState();
            }
        } else if (meta.stage == Status.ACTIVE) {

            // main goal: verify Candidate is valid and >= LastValidState
            if (lastValidStateMagicNumber == CHANOPEN) {
                if (candidateMagicNumber == CHANOPEN) {
                    if (proofs.length != 0) revert InvalidState();
                    if (candidate.sigs.length == PART_NUM) revert InvalidState();
                    _requireValidSignatures(meta.chan, candidate);
                } else {
                    if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) revert InvalidState();
                }
            } else {
                if (candidateMagicNumber == CHANOPEN) revert InvalidState();
                if (_isMoreRecent(meta.chan.adjudicator, meta.lastValidState, candidate)) revert InvalidState();
                if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) revert InvalidState();
            }
        } else { // meta.stage == DISPUTE
            // main goal: verify Candidate is valid and > LastValidState
            if (lastValidStateMagicNumber == CHANOPEN) {
                if (meta.lastValidState.sigs.length != PART_NUM) {
                    _requireChannelHasNFulfilledDeposits(channelId, PART_NUM);
                }

                if (candidateMagicNumber == CHANOPEN) {
                    if (proofs.length != 0) revert InvalidState();
                    if (candidate.sigs.length != PART_NUM) revert InvalidState();
                    _requireValidSignatures(meta.chan, candidate);
                } else {
                    if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) revert InvalidState();
                }
            } else {
                if (candidateMagicNumber == CHANOPEN) revert InvalidState();
                if (!_isMoreRecent(meta.chan.adjudicator, candidate, meta.lastValidState)) revert InvalidState();
                if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) revert InvalidState();
            }
        }

        meta.lastValidState = candidate;
        meta.challengeExpire = block.timestamp + meta.chan.challenge;
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

        uint32 candidateMagicNumber = abi.decode(candidate.data, (uint32));
        if (candidateMagicNumber == CHANOPEN || // disallow checkpointing a funding state as `join` already does that
            candidateMagicNumber == CHANCLOSE) { // disallow checkpointing a closing state as `close` already does that
                revert InvalidState();
        }

        uint32 lastValidStateMagicNumber = abi.decode(meta.lastValidState.data, (uint32));

        // main goal: verify Candidate is valid and > LastValidState

        if (meta.stage == Status.INITIAL) {
            // LastValidState can only be a CHANOPEN and where NOT ALL parties have joined

            _requireChannelHasNFulfilledDeposits(channelId, PART_NUM);
            if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) revert InvalidState();
        } else if (meta.stage == Status.ACTIVE) {
            if (lastValidStateMagicNumber != CHANOPEN) {
                if (!_isMoreRecent(meta.chan.adjudicator, candidate, meta.lastValidState)) revert InvalidState();
            }

            if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) revert InvalidState();
        } else { // meta.stage == DISPUTE
            if (lastValidStateMagicNumber == CHANOPEN) {
                if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) revert InvalidState();

                if (meta.lastValidState.sigs.length != PART_NUM) {
                    _requireChannelHasNFulfilledDeposits(channelId, PART_NUM);
                }
            } else {
                if (!_isMoreRecent(meta.chan.adjudicator, candidate, meta.lastValidState)) revert InvalidState();
                if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) revert InvalidState();
            }

            meta.challengeExpire = 0;
        }

        meta.stage = Status.ACTIVE;
        meta.lastValidState = candidate;


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
     * @param allocations The allocations to distribute
     */
    function _unlockAllocations(bytes32 channelId, Allocation[] memory allocations) internal {
        uint256 allocsLength = allocations.length;
        for (uint256 i = 0; i < allocsLength; i++) {
            _unlockAllocation(channelId, allocations[i]);
        }
    }

    /**
     * @notice Helper function to compare two states for recency
     * @param adjudicator The adjudicator contract address
     * @param stateA The first state to compare
     * @param stateB The second state to compare against
     * @return True if state A is more recent than state B state
     */
    function _isMoreRecent(address adjudicator, State memory stateA, State memory stateB)
        internal
        view
        returns (bool)
    {
        return IComparable(adjudicator).compare(stateA, stateB) > 0;
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
        uint256 available = ledger.tokens[token];
        if (available < amount) revert InsufficientBalance(available, amount);

        ledger.tokens[token] = available - amount; // avoiding "-=" saves gas on a storage lookup
        _channels[channelId].tokenBalances[token] += amount;
    }

    // Does not perform checks to allow transferring partial balances in case of partial deposit
    function _unlockAllocation(bytes32 channelId, Allocation memory alloc) internal {
        if (alloc.amount == 0) return;

        Metadata storage meta = _channels[channelId];
        uint256 channelBalance = meta.tokenBalances[alloc.token];
        if (channelBalance == 0) return;

        uint256 correctedAmount = channelBalance > alloc.amount ? alloc.amount : channelBalance;
        meta.tokenBalances[alloc.token] = channelBalance - correctedAmount; // avoiding "-=" saves gas on a storage lookup
        _ledgers[alloc.destination].tokens[alloc.token] += correctedAmount;
    }

    /**
     * @notice Verifies that all provided signatures are valid for the given state
     * @param chan The channel configuration
     * @param state The state to verify signatures for
     * @return valid True if all provided signatures are valid
     */
    function _verifyAllSignatures(Channel memory chan, State memory state) internal pure returns (bool valid) {
        bytes32 stateHash = Utils.getStateHash(chan, state);

        if (state.sigs.length != PART_NUM) {
            return false;
        }

        for (uint256 i = 0; i < PART_NUM; i++) {
            if (!Utils.verifySignature(stateHash, state.sigs[i], chan.participants[i])) return false;
        }

        return true;
    }

    function _requireValidSignatures(Channel memory chan, State memory state) internal pure {
        bytes32 stateHash = Utils.getStateHash(chan, state);
        uint256 sigsLength = state.sigs.length;

        for (uint256 i = 0; i < sigsLength; i++) {
            if (!Utils.verifySignature(stateHash, state.sigs[i], chan.participants[i])) {
                revert InvalidStateSignatures();
            }
        }
    }

    function _requireChannelHasNFulfilledDeposits(bytes32 channelId, uint256 n) internal view {
        Metadata storage meta = _channels[channelId];

        uint256 fulfilledDeposits = 0;
        for (uint256 i = 0; i < PART_NUM; i++) {
            if (meta.actualDeposits[i].amount != meta.expectedDeposits[i].amount) {
                fulfilledDeposits++;
            }
        }

        if (fulfilledDeposits < n) revert DepositsNotFulfilled(n, fulfilledDeposits);
    }

    function _requireChallengerIsParticipant(
        bytes32 challengedStateHash,
        address[2] memory participants,
        Signature memory challengerSig
    ) internal pure {
        address challenger = Utils.recoverSigner(
			keccak256(abi.encode(challengedStateHash, 'challenge')),
			challengerSig
		);

        if (challenger != participants[CLIENT_IDX] && challenger != participants[SERVER_IDX]) {
            revert InvalidChallengerSignature();
        }
    }
}
