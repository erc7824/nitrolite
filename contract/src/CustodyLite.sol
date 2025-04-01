// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IChannel} from "./interfaces/IChannel.sol";
import {IAdjudicator} from "./interfaces/IAdjudicator.sol";
import {Channel, State, Allocation} from "./interfaces/Types.sol";
import {Utils} from "./Utils.sol";
import {IERC20} from "lib/openzeppelin-contracts/contracts/interfaces/IERC20.sol";

/**
 * @title CustodyLite
 * @notice A simple custody contract for state channels that delegates most state transition logic to an adjudicator
 */
contract CustodyLite is IChannel {
    // Errors
    error ChannelNotFound();
    error InvalidParticipant();
    error InvalidChannel();
    error InvalidState();
    error InvalidAdjudicator();
    error InvalidChallengePeriod();
    error ChannelAlreadyExists();
    error TransferFailed();
    error ChallengeNotExpired();
    error ChannelNotFinal();

    // Events
    event ChannelPartiallyFunded(bytes32 indexed channelId, Channel channel);
    event ChannelOpened(bytes32 indexed channelId, Channel channel);
    event ChannelChallenged(bytes32 indexed channelId, uint256 expiration);
    event ChannelCheckpointed(bytes32 indexed channelId);
    event ChannelClosed(bytes32 indexed channelId);

    // Index in the array of participants
    uint256 constant HOST = 0;
    uint256 constant GUEST = 1;

    // Recommended structure to keep track of states
    struct Metadata {
        Channel chan; // Opener define channel configuration
        uint256 challengeExpire; // If non-zero channel will resolve to lastValidState when challenge Expires
        State lastValidState; // Last valid state when adjudicator was called
    }

    mapping(bytes32 channelId => Metadata chMeta) private channels;

    /**
     * @notice Open or join a channel by depositing assets
     * @param ch Channel configuration
     * @param deposit is the initial State defined by the opener, it contains the expected allocation
     * @return channelId Unique identifier for the channel
     */
    function open(Channel calldata ch, State calldata deposit) public returns (bytes32 channelId) {
        // Validate input parameters
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
            Allocation memory allocation = deposit.allocations[HOST];
            if (allocation.amount > 0) {
                bool success = IERC20(allocation.token).transferFrom(msg.sender, address(this), allocation.amount);
                require(success, TransferFailed());
            }

            Metadata memory newCh = Metadata({chan: ch, challengeExpire: 0, lastValidState: deposit});

            channels[channelId] = newCh;
            emit ChannelPartiallyFunded(channelId, ch);
        } else {
            Allocation memory allocation = deposit.allocations[GUEST];
            if (allocation.amount > 0) {
                bool success = IERC20(allocation.token).transferFrom(msg.sender, address(this), allocation.amount);
                require(success, TransferFailed());
            }

            // Get adjudicator's validation of the state
            State[] memory emptyProofs = new State[](0);
            IAdjudicator.Status status = _adjudicate(meta.chan, deposit, emptyProofs);

            // Update channel state based on adjudicator decision
            if (status == IAdjudicator.Status.ACTIVE) {
                meta.lastValidState = deposit;
                emit ChannelOpened(channelId, ch);
            } else if (status == IAdjudicator.Status.INVALID) {
                revert InvalidState();
            } else if (status == IAdjudicator.Status.PARTIAL) {
                // For Counter adjudicator, PARTIAL means counter = 0
                revert InvalidState();
            } else {
                // Handle other statuses like VOID
                meta.lastValidState = deposit;
                emit ChannelOpened(channelId, ch);
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
        Metadata storage meta = channels[channelId];
        require(meta.chan.adjudicator != address(0), ChannelNotFound());

        // Get adjudicator's validation of the candidate state
        IAdjudicator.Status status = _adjudicate(meta.chan, candidate, proofs);

        // Only proceed if adjudicator determines the state is FINAL
        if (status == IAdjudicator.Status.FINAL) {
            // Set last valid state
            meta.lastValidState = candidate;
            _closeChannel(channelId, meta);
        } else if (status == IAdjudicator.Status.INVALID) {
            revert InvalidState();
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
        Metadata storage meta = channels[channelId];
        require(meta.chan.adjudicator != address(0), ChannelNotFound());

        // Get adjudicator's validation of the candidate state
        IAdjudicator.Status status = _adjudicate(meta.chan, candidate, proofs);

        if (status == IAdjudicator.Status.ACTIVE || status == IAdjudicator.Status.PARTIAL) {
            // Valid challenge, save state and start challenge period
            meta.lastValidState = candidate;
            meta.challengeExpire = block.timestamp + meta.chan.challenge;

            emit ChannelChallenged(channelId, meta.challengeExpire);
        } else if (status == IAdjudicator.Status.INVALID) {
            revert InvalidState();
        } else if (status == IAdjudicator.Status.FINAL) {
            // If state is final, close the channel directly
            meta.lastValidState = candidate;
            _closeChannel(channelId, meta);
        } else {
            // For other statuses like PARTIAL or VOID
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
        Metadata storage meta = channels[channelId];
        require(meta.chan.adjudicator != address(0), ChannelNotFound());

        // Get adjudicator's validation of the candidate state
        IAdjudicator.Status status = _adjudicate(meta.chan, candidate, proofs);

        if (status == IAdjudicator.Status.ACTIVE) {
            // Valid state, save it without starting challenge
            meta.lastValidState = candidate;
            emit ChannelCheckpointed(channelId);
        } else if (status == IAdjudicator.Status.INVALID) {
            revert InvalidState();
        } else if (status == IAdjudicator.Status.FINAL) {
            // If state is final, checkpoint it
            meta.lastValidState = candidate;
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
        Metadata storage meta = channels[channelId];
        require(meta.chan.adjudicator != address(0), ChannelNotFound());

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

            // FIXME: withdrawals when partially funded for challenges on this state to make sense. For that TODO: specify assets distribution per channel
            if (allocation.amount > 0) {
                if (allocation.token == address(0)) {
                    // Transfer ETH
                    (bool success,) = allocation.destination.call{value: allocation.amount}("");
                    require(success, TransferFailed());
                } else {
                    // Transfer ERC20
                    bool success = IERC20(allocation.token).transfer(allocation.destination, allocation.amount);
                    require(success, TransferFailed());
                }
            }
        }

        // Mark channel as closed by removing it
        delete channels[channelId];

        emit ChannelClosed(channelId);
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
}
