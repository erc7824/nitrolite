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
    mapping(address account => Ledger) internal _accounts;

    function deposit(address token, uint256 amount) external payable {
        address account = msg.sender;
        if (token == address(0)) {
            require(msg.value == amount, InvalidAmount());
        } else {
            bool success = IERC20(token).transferFrom(
                account,
                address(this),
                amount
            );
            require(success, TransferFailed(token, address(this), amount));
        }
        _accounts[msg.sender].tokens[token].available += amount;
    }

    function withdraw(address token, uint256 amount) external {
        Ledger storage account = _accounts[msg.sender];
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
    function getAccountChannels(
        address account
    ) public view returns (bytes32[] memory) {
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
    function getAccountInfo(
        address account,
        address token
    )
        public
        view
        returns (uint256 available, uint256 locked, uint256 channelCount)
    {
        Ledger storage accountInfo = _accounts[account];
        Account storage funds = accountInfo.tokens[token];
        return (funds.available, funds.locked, accountInfo.channels.length());
    }

    /**
     * @notice Open or join a channel by depositing assets
     * @param ch Channel configuration
     * @param depositState is the initial State defined by the opener, it contains the expected allocation
     * @return channelId Unique identifier for the channel
     */
    function open(
        Channel calldata ch,
        State calldata depositState
    ) public returns (bytes32 channelId) {
        // Validate input parameters
        // FIXME: lift this restriction when supporting N-party channels
        require(ch.participants.length == 2, InvalidParticipant());
        require(
            ch.participants[0] != address(0) &&
                ch.participants[1] != address(0),
            InvalidParticipant()
        );
        require(ch.adjudicator != address(0), InvalidAdjudicator());
        require(ch.challenge != 0, InvalidChallengePeriod());

        // Generate channel ID
        channelId = Utils.getChannelId(ch);

        // Check if channel doesn't exist and create new one (HOST deposit)
        Metadata storage meta = _channels[channelId];
        if (meta.chan.adjudicator == address(0)) {
            // This is the first participant (HOST) creating the channel
            Allocation memory allocation = depositState.allocations[HOST];

            // TODO: Verify by calling adjudicator with empty proofs
            // TODO: Verify stateHash is signed by HOST
           
            // Initialize channel metadata
            Metadata storage newCh = _channels[channelId];
            newCh.chan = ch;
            newCh.status = Status.PARTIAL; // Set initial status to PARTIAL (partially funded)
            newCh.challengeExpire = 0;
            newCh.lastValidState = depositState;

            // Transfer funds from account to channel
            _lockAccountFundsToChannel(
                ch.participants[HOST],
                channelId,
                allocation.token,
                allocation.amount
            );
            _accounts[ch.participants[HOST]].channels.add(channelId);

            emit ChannelPartiallyFunded(channelId, ch);
        } else if (meta.status == Status.PARTIAL) {
            // This is the second participant (GUEST) joining an existing partially funded channel
            Allocation memory allocation = depositState.allocations[GUEST];

            // Lock funds from the GUEST to the channel
            _lockAccountFundsToChannel(
                ch.participants[GUEST],
                channelId,
                allocation.token,
                allocation.amount
            );

            // Verify the guest's signature
            bytes32 stateHash = Utils.getStateHash(ch, depositState);
            bool validSignature = Utils.verifySignature(
                stateHash,
                depositState.sigs[0],
                ch.participants[GUEST]
            );

            if (validSignature) {
                // Store the new state with GUEST's signature
                meta.lastValidState = depositState;

                // Now that both participants have funded, change status to ACTIVE
                meta.status = Status.ACTIVE;

                // Add channel to GUEST's account
                _accounts[ch.participants[GUEST]].channels.add(channelId);

                emit ChannelOpened(channelId, ch);
            } else {
                revert InvalidState();
            }
        } else {
            // Channel exists but is not in PARTIAL state
            revert InvalidState();
        }

        return channelId;
    }

    /**
     * @notice Finalize the channel with a mutually signed state
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function close(
        bytes32 channelId,
        State calldata candidate,
        State[] calldata proofs
    ) public {
        Metadata storage meta = _requireValidChannel(channelId);

        // Channel must be in a valid state to close (ACTIVE, PARTIAL, or already FINAL)
        if (meta.status == Status.VOID || meta.status == Status.INVALID) {
            revert InvalidState();
        }

        // If already FINAL, we can proceed to close without adjudicator validation
        if (meta.status != Status.FINAL) {
            // Validate the candidate state using the adjudicator
            // Adjudicator only validates state transitions, not channel status
            bool isValid = _validateState(meta, candidate, proofs);

            // For cooperative close, we need:
            // 1. The state must be valid according to the adjudicator
            // 2. The state must have signatures from all participants
            if (
                isValid &&
                candidate.sigs.length == meta.chan.participants.length
            ) {
                // All requirements met, set status to FINAL
                meta.status = Status.FINAL;
            } else {
                revert ChannelNotFinal();
            }
        }

        // At this point, the channel is in FINAL state, so we can close it
        _closeChannel(channelId, meta);
    }

    /**
     * @notice Unilaterally post a state when the other party is uncooperative
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function challenge(
        bytes32 channelId,
        State calldata candidate,
        State[] calldata proofs
    ) external {
        Metadata storage meta = _requireValidChannel(channelId);

        // Only allow challenges on funded channels
        if (meta.status != Status.ACTIVE && meta.status != Status.PARTIAL) {
            revert InvalidState();
        }

        // Validate the candidate state using adjudicator
        // Adjudicators only validate state transitions, not channel status
        bool isValid = _validateState(meta, candidate, proofs);

        if (isValid) {
            // If state has signatures from all participants, consider it final
            if (candidate.sigs.length == meta.chan.participants.length) {
                // All parties have signed - cooperative close
                meta.status = Status.FINAL;
                _closeChannel(channelId, meta);
            } else {
                // Start challenge period - this is handled by Custody, not adjudicator
                meta.challengeExpire = block.timestamp + meta.chan.challenge;

                // Keep the channel status as is (ACTIVE or PARTIAL)
                // The status will be updated to FINAL when reclaim is called after challenge period

                emit ChannelChallenged(channelId, meta.challengeExpire);
            }
        } else {
            // Invalid state according to adjudicator
            revert InvalidState();
        }
    }

    /**
     * @notice Unilaterally post a state to store it on-chain to prevent future disputes
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function checkpoint(
        bytes32 channelId,
        State calldata candidate,
        State[] calldata proofs
    ) external {
        Metadata storage meta = _requireValidChannel(channelId);

        // Only allow checkpoints on funded channels
        if (meta.status != Status.ACTIVE && meta.status != Status.PARTIAL) {
            revert InvalidState();
        }

        // Validate the candidate state
        // Adjudicators only validate state transitions, not channel status
        bool isValid = _validateState(meta, candidate, proofs);

        if (isValid) {
            // Valid state is stored for future reference

            // If all participants signed, treat as a cooperative final state
            if (candidate.sigs.length == meta.chan.participants.length) {
                meta.status = Status.FINAL;
                // Don't close the channel yet - user needs to call close explicitly
            }

            // Otherwise, keep the channel status as is (ACTIVE or PARTIAL)
            // This allows the channel to continue operating with the checkpoint as the latest valid state

            emit ChannelCheckpointed(channelId);
        } else {
            // Invalid state according to adjudicator
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
        require(
            meta.challengeExpire != 0 &&
                block.timestamp >= meta.challengeExpire,
            ChallengeNotExpired()
        );

        // Set the status to FINAL before closing
        // This is the custody contract's responsibility, not the adjudicator's
        meta.status = Status.FINAL;

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
            _unlockChannelFundsToAccount(
                channelId,
                allocation.destination,
                allocation.token,
                allocation.amount
            );
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
    function _requireValidChannel(
        bytes32 channelId
    ) internal view returns (Metadata storage meta) {
        meta = _channels[channelId];
        require(
            meta.chan.adjudicator != address(0),
            ChannelNotFound(channelId)
        );
        return meta;
    }

    /**
     * @notice Internal function to validate a candidate state and handle the result
     * @param meta The channel's metadata
     * @param candidate The state to be adjudicated
     * @param proofs Additional proof states if required
     * @return valid True if the state is valid according to the adjudicator
     */
    function _validateState(
        Metadata storage meta,
        State memory candidate,
        State[] memory proofs
    ) internal returns (bool valid) {
        IAdjudicator adjudicator = IAdjudicator(meta.chan.adjudicator);

        // Adjudicators only validate state transitions, not channel status
        try adjudicator.adjudicate(meta.chan, candidate, proofs) returns (
            bool result
        ) {
            valid = result;

            if (valid) {
                // Store the valid state
                meta.lastValidState = candidate;

                // Channel status handling is managed by Custody contract
                // Don't automatically change status to ACTIVE here

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
            (success, ) = to.call{value: amount}("");
        } else {
            success = IERC20(token).transfer(to, amount);
        }
        require(success, TransferFailed(token, to, amount));
    }

    function _lockAccountFundsToChannel(
        address account,
        bytes32 channelId,
        address token,
        uint256 amount
    ) internal {
        if (amount == 0) return;

        Ledger storage accountInfo = _accounts[account];
        uint256 available = accountInfo.tokens[token].available;
        require(available >= amount, InsufficientBalance(available, amount));

        accountInfo.tokens[token].available -= amount;
        accountInfo.tokens[token].locked += amount;

        Metadata storage meta = _channels[channelId];
        meta.tokenBalances[token] += amount;
    }

    // Does not perform checks to allow transferring partial balances in case of partial deposit
    function _unlockChannelFundsToAccount(
        bytes32 channelId,
        address account,
        address token,
        uint256 amount
    ) internal {
        if (amount == 0) return;

        Metadata storage meta = _channels[channelId];
        uint256 channelBalance = meta.tokenBalances[token];
        if (channelBalance == 0) return;

        uint256 correctedAmount = channelBalance > amount
            ? amount
            : channelBalance;
        meta.tokenBalances[token] -= correctedAmount;

        Ledger storage accountInfo = _accounts[account];

        // Check locked amount before subtracting to prevent underflow
        uint256 lockedAmount = accountInfo.tokens[token].locked;
        uint256 amountToUnlock = lockedAmount > correctedAmount
            ? correctedAmount
            : lockedAmount;

        if (amountToUnlock > 0) {
            accountInfo.tokens[token].locked -= amountToUnlock;
            accountInfo.tokens[token].available += amountToUnlock;
        }
    }
}
