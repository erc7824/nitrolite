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
    error InvalidParticipant();
    error InvalidState();
    error InvalidAdjudicator();
    error InvalidChallengePeriod();
    error InvalidAmount();
    error TransferFailed(address token, address to, uint256 amount);
    error ChallengeNotExpired();
    error ChannelNotFinal();
    error InsufficientBalance(uint256 available, uint256 required);

    // Index in the array of participants
    uint256 constant HOST = 0;
    uint256 constant GUEST = 1;

    // Recommended structure to keep track of states
    struct Metadata {
        Channel chan; // Opener define channel configuration
        Status status; // Current channel status
        uint256 challengeExpire; // If non-zero channel will resolve to lastValidState when challenge Expires
        State lastValidState; // Last valid state when adjudicator was called
        mapping(address token => uint256 balance) tokenBalances; // Token balances for the channel
    }

    struct Funds {
        uint256 available; // Available amount that can be withdrawn or allocated to channels
        uint256 locked; // Amount currently allocated to channels
    }

    struct Account {
        mapping(address token => Funds funds) tokens; // Token balances
        EnumerableSet.Bytes32Set channels; // Set of user ChannelId
    }

    mapping(bytes32 channelId => Metadata chMeta) internal _channels;
    mapping(address account => Account) internal _accounts;

    function deposit(address token, uint256 amount) external payable {
        address account = msg.sender;
        if (token == address(0)) {
            require(msg.value == amount, InvalidAmount());
        } else {
            bool success = IERC20(token).transferFrom(account, address(this), amount);
            require(success, TransferFailed(token, address(this), amount));
        }

        _accounts[msg.sender].tokens[token].available += amount;
    }

    function withdraw(address token, uint256 amount) external {
        Account storage account = _accounts[msg.sender];
        uint256 available = account.tokens[token].available;
        require(available >= amount, InsufficientBalance(available, amount));
        _transfer(token, msg.sender, amount);
        account.tokens[token].available -= amount;
    }

    /**
     * @notice Get channels associated with an account
     * @param account The account address
     * @return List of channel IDs associated with the account
     */
    function getAccountChannels(address account) public view returns (bytes32[] memory) {
        return _accounts[account].channels.values();
    }

    /**
     * @notice Get account information for a specific token
     * @param account The account address
     * @param token The token address
     * @return available Amount available for withdrawal or allocation
     * @return locked Amount locked in channels
     * @return channelCount Number of associated channels
     */
    function getAccountInfo(address account, address token)
        public
        view
        returns (uint256 available, uint256 locked, uint256 channelCount)
    {
        Account storage accountInfo = _accounts[account];
        Funds storage funds = accountInfo.tokens[token];
        return (funds.available, funds.locked, accountInfo.channels.length());
    }

    /**
     * @notice Open or join a channel by depositing assets
     * @param ch Channel configuration
     * @param depositState is the initial State defined by the opener, it contains the expected allocation
     * @return channelId Unique identifier for the channel
     */
    function open(Channel calldata ch, State calldata depositState) public returns (bytes32 channelId) {
        // Validate input parameters
        // FIXME: lift this restriction
        require(ch.participants.length == 2, InvalidParticipant());
        require(ch.participants[0] != address(0) && ch.participants[1] != address(0), InvalidParticipant());
        require(ch.adjudicator != address(0), InvalidAdjudicator());
        require(ch.challenge != 0, InvalidChallengePeriod());

        // Generate channel ID
        channelId = Utils.getChannelId(ch);

        // Check if channel doesn't exists and create new one
        Metadata storage meta = _channels[channelId];
        if (meta.chan.adjudicator == address(0)) {
            // Validate deposits and transfer funds
            Allocation memory allocation = depositState.allocations[HOST];

            // Initialize channel metadata
            Metadata storage newCh = _channels[channelId];
            newCh.chan = ch;
            newCh.status = Status.PARTIAL; // Set initial status to PARTIAL
            newCh.challengeExpire = 0;
            newCh.lastValidState = depositState;

            // Transfer funds from account to channel
            _lockAccountFundsToChannel(ch.participants[HOST], channelId, allocation.token, allocation.amount);
            _accounts[ch.participants[HOST]].channels.add(channelId);

            emit ChannelPartiallyFunded(channelId, ch);
        } else {
            Allocation memory allocation = depositState.allocations[GUEST];
            _lockAccountFundsToChannel(ch.participants[GUEST], channelId, allocation.token, allocation.amount);

            // For opening a channel, we just need to verify the guest's signature
            // but don't need to validate with the adjudicator
            bytes32 stateHash = Utils.getStateHash(ch, depositState);
            bool validSignature = Utils.verifySignature(stateHash, depositState.sigs[0], ch.participants[GUEST]);

            if (validSignature) {
                meta.lastValidState = depositState;
                meta.status = Status.ACTIVE; // Set status to ACTIVE
                _accounts[ch.participants[GUEST]].channels.add(channelId);
                emit ChannelOpened(channelId, ch);
            } else {
                revert InvalidState();
            }
        }

        return channelId;
    }

    /**
     * @notice Finalize the channel with a mutually signed state
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function close(bytes32 channelId, State calldata candidate, State[] calldata proofs) public {
        Metadata storage meta = _requireValidChannel(channelId);

        // Validate the candidate state
        bool isValid = _validateState(meta, candidate, proofs);

        // For cooperative close, we need:
        // 1. The state must be valid
        // 2. The state must have signatures from all participants
        if (isValid && candidate.sigs.length == meta.chan.participants.length) {
            // Set status to FINAL
            meta.status = Status.FINAL;
            _closeChannel(channelId, meta);
        } else {
            revert ChannelNotFinal();
        }
    }

    /**
     * @notice Unilaterally post a state when the other party is uncooperative
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function challenge(bytes32 channelId, State calldata candidate, State[] calldata proofs) external {
        Metadata storage meta = _requireValidChannel(channelId);

        // Validate the candidate state
        bool isValid = _validateState(meta, candidate, proofs);

        if (isValid) {
            if (meta.status == Status.ACTIVE || meta.status == Status.PARTIAL) {
                // If state has signatures from all participants, consider it final
                if (candidate.sigs.length == meta.chan.participants.length) {
                    meta.status = Status.FINAL;
                    _closeChannel(channelId, meta);
                } else {
                    // Valid challenge, start challenge period
                    meta.challengeExpire = block.timestamp + meta.chan.challenge;
                    emit ChannelChallenged(channelId, meta.challengeExpire);
                }
            } else {
                // For other statuses like VOID or INVALID
                revert InvalidState();
            }
        } else {
            // This shouldn't happen since _validateState would revert
            revert InvalidState();
        }
    }

    /**
     * @notice Unilaterally post a state to store it on-chain to prevent future disputes
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function checkpoint(bytes32 channelId, State calldata candidate, State[] calldata proofs) external {
        Metadata storage meta = _requireValidChannel(channelId);

        // Validate the candidate state
        bool isValid = _validateState(meta, candidate, proofs);

        if (isValid) {
            if (meta.status == Status.ACTIVE || meta.status == Status.PARTIAL) {
                // Valid state, checkpoint without starting challenge
                emit ChannelCheckpointed(channelId);

                // If all participants signed, update to FINAL
                if (candidate.sigs.length == meta.chan.participants.length) {
                    meta.status = Status.FINAL;
                }
            } else {
                // For other statuses like VOID or INVALID
                revert InvalidState();
            }
        } else {
            // This shouldn't happen since _validateState would revert
            revert InvalidState();
        }
    }

    /**
     * @notice Conclude the channel after challenge period expires
     * @param channelId Unique identifier for the channel
     */
    function reclaim(bytes32 channelId) external {
        Metadata storage meta = _requireValidChannel(channelId);

        // Ensure challenge period has expired
        require(meta.challengeExpire != 0 && block.timestamp >= meta.challengeExpire, ChallengeNotExpired());

        // Close the channel with last valid state
        _closeChannel(channelId, meta);
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
    function _closeChannel(bytes32 channelId, Metadata storage meta) internal {
        // Distribute funds according to allocations
        uint256 allocsLength = meta.lastValidState.allocations.length;
        for (uint256 i = 0; i < allocsLength; i++) {
            Allocation memory allocation = meta.lastValidState.allocations[i];
            _unlockChannelFundsToAccount(channelId, allocation.destination, allocation.token, allocation.amount);
        }

        uint256 participantsLength = meta.chan.participants.length;
        for (uint256 i = 0; i < participantsLength; i++) {
            address participant = meta.chan.participants[i];
            _accounts[participant].channels.remove(channelId);
        }

        // Mark channel as closed by removing it
        delete _channels[channelId];

        emit ChannelClosed(channelId);
    }

    /**
     * @notice Internal function to ensure a channel exists
     * @param channelId The channel identifier
     * @return meta The channel's metadata
     */
    function _requireValidChannel(bytes32 channelId) internal view returns (Metadata storage meta) {
        meta = _channels[channelId];
        require(meta.chan.adjudicator != address(0), ChannelNotFound(channelId));
        return meta;
    }

    /**
     * @notice Internal function to validate a candidate state and handle the result
     * @param meta The channel's metadata
     * @param candidate The state to be adjudicated
     * @param proofs Additional proof states if required
     * @return valid True if the state is valid according to the adjudicator
     */
    function _validateState(Metadata storage meta, State memory candidate, State[] memory proofs)
        internal
        returns (bool valid)
    {
        IAdjudicator adjudicator = IAdjudicator(meta.chan.adjudicator);

        try adjudicator.adjudicate(meta.chan, candidate, proofs) returns (bool result) {
            valid = result;

            if (valid) {
                meta.lastValidState = candidate;
                meta.status = Status.ACTIVE;
                return valid;
            }

            meta.status = Status.INVALID;
            return valid;
        } catch {
            // If the adjudicator call reverts, treat as invalid state
            meta.status = Status.INVALID;
            return false;
        }
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

        Account storage accountInfo = _accounts[account];
        uint256 available = accountInfo.tokens[token].available;
        require(available >= amount, InsufficientBalance(available, amount));

        accountInfo.tokens[token].available -= amount;
        accountInfo.tokens[token].locked += amount;

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

        Account storage accountInfo = _accounts[account];

        // Check locked amount before subtracting to prevent underflow
        uint256 lockedAmount = accountInfo.tokens[token].locked;
        uint256 amountToUnlock = lockedAmount > correctedAmount ? correctedAmount : lockedAmount;

        if (amountToUnlock > 0) {
            accountInfo.tokens[token].locked -= amountToUnlock;
            accountInfo.tokens[token].available += amountToUnlock;
        }
    }
}
