// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IChannel} from "./interfaces/IChannel.sol";
import {IDeposit} from "./interfaces/IDeposit.sol";
import {IAdjudicator} from "./interfaces/IAdjudicator.sol";
import {Channel, State, Allocation} from "./interfaces/Types.sol";
import {Utils} from "./Utils.sol";
import {IERC20} from "lib/openzeppelin-contracts/contracts/interfaces/IERC20.sol";

/**
 * @title Custody
 * @notice A simple custody contract for state channels that delegates most state transition logic to an adjudicator
 */
contract Custody is IChannel, IDeposit {
    // Errors
    error ChannelNotFound(bytes32 channelId);
    error InvalidParticipant();
    error InvalidState();
    error InvalidAdjudicator();
    error InvalidChallengePeriod();
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
        uint256 challengeExpire; // If non-zero channel will resolve to lastValidState when challenge Expires
        State lastValidState; // Last valid state when adjudicator was called
    }

    struct Account {
        uint256 deposited; // Total amount for this token
        uint256 locked; // Amount currently allocated to channels
        bytes32[] channels; // List of user ChannelId
    }

    mapping(bytes32 channelId => Metadata chMeta) internal channels;
    mapping(address account => mapping(address token => Account)) internal balances;
    mapping(bytes32 channelId => mapping(address token => uint256 balance)) internal channelBalances;

    function deposit(address token, uint256 amount) external {
        _transferFrom(token, msg.sender, address(this), amount);
        balances[msg.sender][token].deposited += amount;
    }

    function withdraw(address token, uint256 amount) external {
        Account storage account = balances[msg.sender][token];
        uint256 available = getAvailableBalance(msg.sender, token);
        require(available >= amount, InsufficientBalance(available, amount));
        _transfer(token, msg.sender, amount);
        account.deposited -= amount;
    }

    /**
     * @notice Get available balance for an account for a specific token
     * @param account The account address
     * @param token The token address
     * @return available The available balance that can be withdrawn or allocated to channels
     */
    function getAvailableBalance(address account, address token) public view returns (uint256 available) {
        Account storage accountInfo = balances[account][token];
        return accountInfo.deposited - accountInfo.locked;
    }

    /**
     * @notice Get channels associated with an account for a specific token
     * @param account The account address
     * @param token The token address
     * @return List of channel IDs associated with the account for the token
     */
    function getAccountChannels(address account, address token) public view returns (bytes32[] memory) {
        return balances[account][token].channels;
    }

    /**
     * @notice Get account information for a specific token
     * @param account The account address
     * @param token The token address
     * @return deposited Total deposited amount
     * @return locked Amount locked in channels
     * @return channelCount Number of associated channels
     */
    function getAccountInfo(address account, address token)
        public
        view
        returns (uint256 deposited, uint256 locked, uint256 channelCount)
    {
        Account storage accountInfo = balances[account][token];
        return (accountInfo.deposited, accountInfo.locked, accountInfo.channels.length);
    }

    /**
     * @notice Open or join a channel by depositing assets
     * @param ch Channel configuration
     * @param depositSt is the initial State defined by the opener, it contains the expected allocation
     * @return channelId Unique identifier for the channel
     */
    function open(Channel calldata ch, State calldata depositSt) public returns (bytes32 channelId) {
        // Validate input parameters
        // FIXME: lift this restriction
        require(ch.participants.length == 2, InvalidParticipant());
        require(ch.participants[0] != address(0) && ch.participants[1] != address(0), InvalidParticipant());
        require(ch.adjudicator != address(0), InvalidAdjudicator());
        require(ch.challenge != 0, InvalidChallengePeriod());

        // Generate channel ID
        channelId = Utils.getChannelId(ch);

        // Check if channel doesn't exists and create new one
        Metadata storage meta = channels[channelId];
        if (meta.chan.adjudicator == address(0)) {
            // Validate deposits and transfer funds
            Allocation memory allocation = depositSt.allocations[HOST];
            _checkAndTransferAccountToChannel(ch.participants[HOST], channelId, allocation.token, allocation.amount);

            Metadata memory newCh = Metadata({chan: ch, challengeExpire: 0, lastValidState: depositSt});

            channels[channelId] = newCh;
            emit ChannelPartiallyFunded(channelId, ch);
        } else {
            Allocation memory allocation = depositSt.allocations[GUEST];
            _checkAndTransferAccountToChannel(ch.participants[GUEST], channelId, allocation.token, allocation.amount);

            // Validate the state with an empty proof
            State[] memory emptyProofs = new State[](0);
            IAdjudicator.Status status = _validateState(meta, depositSt, emptyProofs);

            // Update channel state based on adjudicator decision
            if (status == IAdjudicator.Status.ACTIVE || status == IAdjudicator.Status.VOID) {
                emit ChannelOpened(channelId, ch);
            } else if (status == IAdjudicator.Status.PARTIAL) {
                // For Counter adjudicator, PARTIAL means counter = 0
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
        IAdjudicator.Status status = _validateState(meta, candidate, proofs);

        // Only proceed if adjudicator determines the state is FINAL
        if (status == IAdjudicator.Status.FINAL) {
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
        IAdjudicator.Status status = _validateState(meta, candidate, proofs);

        if (status == IAdjudicator.Status.ACTIVE || status == IAdjudicator.Status.PARTIAL) {
            // Valid challenge, start challenge period
            meta.challengeExpire = block.timestamp + meta.chan.challenge;
            emit ChannelChallenged(channelId, meta.challengeExpire);
        } else if (status == IAdjudicator.Status.FINAL) {
            // If state is final, close the channel directly
            _closeChannel(channelId, meta);
        } else {
            // For other statuses like VOID
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
        IAdjudicator.Status status = _validateState(meta, candidate, proofs);

        if (status == IAdjudicator.Status.ACTIVE || status == IAdjudicator.Status.FINAL) {
            // Valid state, checkpoint without starting challenge
            emit ChannelCheckpointed(channelId);
        } else {
            // For other statuses like PARTIAL or VOID
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
        for (uint256 i = 0; i < meta.lastValidState.allocations.length; i++) {
            Allocation memory allocation = meta.lastValidState.allocations[i];
            _checkAndTransferChannelToAccount(channelId, allocation.destination, allocation.token, allocation.amount);
        }

        // Mark channel as closed by removing it
        delete channels[channelId];

        emit ChannelClosed(channelId);
    }

    /**
     * @notice Internal function to ensure a channel exists
     * @param channelId The channel identifier
     * @return meta The channel's metadata
     */
    function _requireValidChannel(bytes32 channelId) internal view returns (Metadata storage meta) {
        meta = channels[channelId];
        require(meta.chan.adjudicator != address(0), ChannelNotFound(channelId));
        return meta;
    }

    /**
     * @notice Internal function to validate a candidate state and handle the result
     * @param meta The channel's metadata
     * @param candidate The state to be adjudicated
     * @param proofs Additional proof states if required
     * @return status The adjudication status
     */
    function _validateState(Metadata storage meta, State memory candidate, State[] memory proofs)
        internal
        returns (IAdjudicator.Status)
    {
        IAdjudicator.Status status = _adjudicate(meta.chan, candidate, proofs);

        if (status == IAdjudicator.Status.INVALID) {
            revert InvalidState();
        }

        // For valid states, update the lastValidState
        if (
            status == IAdjudicator.Status.ACTIVE || status == IAdjudicator.Status.PARTIAL
                || status == IAdjudicator.Status.FINAL
        ) {
            meta.lastValidState = candidate;
        }

        return status;
    }

    /**
     * @notice Internal function to adjudicate a state
     * @param ch The channel configuration
     * @param candidate The state to be adjudicated
     * @param proofs Additional proof states if required
     * @return status The adjudication status
     */
    function _adjudicate(Channel memory ch, State memory candidate, State[] memory proofs)
        internal
        view
        returns (IAdjudicator.Status status)
    {
        IAdjudicator adjudicator = IAdjudicator(ch.adjudicator);
        // Convert to calldata by passing individual parameters
        return adjudicator.adjudicate(ch, candidate, proofs);
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

    function _transferFrom(address token, address from, address to, uint256 amount) internal {
        bool success;
        if (token == address(0)) {
            (success,) = to.call{value: amount}("");
        } else {
            success = IERC20(token).transferFrom(from, to, amount);
        }
        require(success, TransferFailed(token, to, amount));
    }

    function _checkAndTransferAccountToChannel(address account, bytes32 channelId, address token, uint256 amount)
        internal
    {
        if (amount == 0) return;

        Account storage accountInfo = balances[account][token];
        uint256 available = getAvailableBalance(account, token);
        require(available >= amount, InsufficientBalance(available, amount));

        accountInfo.locked += amount;
        // Add channelId to the list if it's not already there
        if (!_containsValue(accountInfo.channels, channelId)) {
            accountInfo.channels.push(channelId);
        }

        channelBalances[channelId][token] += amount;
    }

    // Does not perform checks to allow transferring partial balances in case of partial deposit
    function _checkAndTransferChannelToAccount(bytes32 channelId, address account, address token, uint256 amount)
        internal
    {
        if (amount == 0) return;

        uint256 channelBalance = channelBalances[channelId][token];
        if (channelBalance == 0) return;

        uint256 correctedAmount = channelBalance > amount ? amount : channelBalance;
        channelBalances[channelId][token] -= correctedAmount;

        Account storage accountInfo = balances[account][token];
        accountInfo.deposited += correctedAmount;

        // Check if we need to update locked amount and possibly remove channel from list
        bool shouldRemoveChannel = channelBalance <= correctedAmount;
        if (shouldRemoveChannel) {
            // Update locked amount
            if (accountInfo.locked >= correctedAmount) {
                accountInfo.locked -= correctedAmount;
            } else {
                // This shouldn't happen, but as a safety measure
                accountInfo.locked = 0;
            }

            // Remove channelId from the list
            _removeFromArray(accountInfo.channels, channelId);
        }
    }

    /**
     * @notice Utility function to remove an element from a bytes32 array
     * @param array The array to modify
     * @param value The value to remove
     * @return found Whether the value was found and removed
     */
    function _removeFromArray(bytes32[] storage array, bytes32 value) internal returns (bool found) {
        uint256 length = array.length;
        for (uint256 i = 0; i < length; i++) {
            if (array[i] == value) {
                // Replace with the last element and remove the last one
                if (i < length - 1) {
                    array[i] = array[length - 1];
                }
                array.pop();
                return true;
            }
        }
        return false;
    }

    /**
     * @notice Utility function to check if an array contains a value
     * @param array The array to check
     * @param value The value to look for
     * @return exists Whether the value exists in the array
     */
    function _containsValue(bytes32[] storage array, bytes32 value) internal view returns (bool exists) {
        uint256 length = array.length;
        for (uint256 i = 0; i < length; i++) {
            if (array[i] == value) {
                return true;
            }
        }
        return false;
    }
}
