// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {IChannel} from "./interfaces/IChannel.sol";
import {IDeposit} from "./interfaces/IDeposit.sol";
import {IAdjudicator} from "./interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Status} from "./interfaces/Types.sol";
import {Utils} from "./Utils.sol";
import {IERC20} from "lib/openzeppelin-contracts/contracts/interfaces/IERC20.sol";
import {EnumerableSet} from "lib/openzeppelin-contracts/contracts/utils/structs/EnumerableSet.sol";

/**
 * @title Custody
 * @notice A simple custody contract for state channels that delegates most state transition logic to an adjudicator
 */
contract Custody is IChannel, IDeposit {
    using EnumerableSet for EnumerableSet.Bytes32Set;

    // Errors
    error ChannelNotFound(bytes32 channelId);
    error ChannelNotFinal();
    error InvalidParticipant();
    error InvalidStatus();
    error InvalidState();
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
            require(msg.value == amount, InvalidAmount());
        } else {
            //TODO: Support native token
            bool success = IERC20(token).transferFrom(account, address(this), amount);
            require(success, TransferFailed(token, address(this), amount));
        }
        _ledgers[msg.sender].tokens[token].available += amount;
    }

    function withdraw(address token, uint256 amount) external {
        Ledger storage ledger = _ledgers[msg.sender];
        uint256 available = ledger.tokens[token].available;
        require(available >= amount, InsufficientBalance(available, amount));
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
    function create(Channel calldata ch, State calldata depositState) public returns (bytes32 channelId) {
        // TODO Generate the create implementation from CREATOR following the funding protocol
        // require State.data == CHANOPEN magic number
        // Verify CREATOR participant has signed funding stateHash
        // Construct channel Metadata and initialize expectedDeposits
        // Add creator address at actualDeposits index 0
        // If valid Transfer User funds to channel Account ledger
        // Set channel metadata.stage = Status.INITIAL
        // NOTE: a participant Allocation can be zero but will still be required to join
        // return channelId;
    }

    // TODO: implement join
    function join(bytes32 channelId, uint256 index, Signature sig) external returns (bytes32 channelId) {
        //TODO enerate the join implementation for other participants
        // Verify participant has signed funding stateHash and his recovered public address
        // is the same as the participant address at index
        // If valid Transfer User funds to channel Account ledger
        // add the participant to actualDeposits at index
        // emit Joined event
        // If actualDeposits is equal to expectedDeposits for all participant
        // Set meta.stage = ACTIVE and emit event Opened
        // Channel is ready for off-chain protocols
    }

    /**
     * @notice Finalize the channel with a mutually signed state
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function close(bytes32 channelId, State calldata candidate, State[] calldata proofs) public {
        // Require all participant signatures on funding protocol stateHash
        // require State.data == CHANCLOSE magic number
        //
        // In the case ot meta.stage == Status.DISPUTE and blocktime is higher than challengeExpire ts
        // you can distribute and close

        // At this point, the channel is in FINAL state, so we can close it
        // _distributeAllocation(channelId, meta);
    }

    /**
     * @notice Unilaterally post a state when the other party is uncooperative
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function challenge(bytes32 channelId, State calldata candidate, State[] calldata proofs) external {
        // TODO call adjudicator
        // if valid compare candidate with lastValidState, if more recent store candidate
        // in lastValidState; Calculate challengeExpire time to start challenge set stage = DISPUTE
        // If other participant submit valid newer states, overwrite lastValidState and reset challengeExpire
        // If meta.stage == INITAL, validate funding protocol stateHash and start challenge
    }

    /**
     * @notice Unilaterally post a state to store it on-chain to prevent future disputes
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    // TODO: checkpoint should remove ongoing challenge if checkpointed state is newer then the challenged one
    function checkpoint(bytes32 channelId, State calldata candidate, State[] calldata proofs) external {
        // 1. validate candidate if adjudicator
        // 2. Compare with IComparable if candidate is more recent than lastValidState
        // 3. Store candidate in lastValidState
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
        open(newChannel, newDeposit);
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

        uint256 participantsLength = meta.chan.participants.length;
        for (uint256 i = 0; i < participantsLength; i++) {
            address participant = meta.chan.participants[i];
            _ledgers[participant].channels.remove(channelId);
        }

        // Mark channel as closed by removing it
        delete _channels[channelId];

        emit ChannelClosed(channelId);
    }

    function _transfer(address token, address to, uint256 amount) internal {
        bool success;
        if (token == address(0)) {
            (success,) = to.call{value: amount}("");
        } else {
            success = IERC20(token).transfer(to, amount);
        }
        require(success, TransferFailed(token, to, amount));
    }

    function _lockAccountFundsToChannel(address account, bytes32 channelId, address token, uint256 amount) internal {
        if (amount == 0) return;

        Ledger storage ledger = _ledgers[account];
        uint256 available = ledger.tokens[token].available;
        require(available >= amount, InsufficientBalance(available, amount));

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
