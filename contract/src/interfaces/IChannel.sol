// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Channel, State, Signature, Amount} from "./Types.sol";

/**
 * @title State Channel Interface
 * @notice Main interface for the Nitrolite state channel system
 * @dev Defines the core functions for creating, managing, and resolving state channels
 */
interface IChannel {
    /**
     * @notice Emitted when a new channel is created
     * @param channelId Unique identifier for the channel
     * @param channel Channel configuration including participants and adjudicator
     * @param expected Expected token deposits from each participant
     */
    event Created(bytes32 indexed channelId, Channel channel, Amount[] expected);

    /**
     * @notice Emitted when a participant joins a channel
     * @param channelId Unique identifier for the channel
     * @param index Index of the participant who joined
     */
    event Joined(bytes32 indexed channelId, uint256 index);

    /**
     * @notice Emitted when a channel becomes fully funded and active
     * @param channelId Unique identifier for the channel
     */
    event Opened(bytes32 indexed channelId);

    /**
     * @notice Emitted when a channel enters the challenge period
     * @param channelId Unique identifier for the channel
     * @param expiration Timestamp when the challenge period expires
     */
    event Challenged(bytes32 indexed channelId, uint256 expiration);

    /**
     * @notice Emitted when a state is checkpointed on-chain
     * @param channelId Unique identifier for the channel
     */
    event Checkpointed(bytes32 indexed channelId);

    /**
     * @notice Emitted when a channel is closed and funds are distributed
     * @param channelId Unique identifier for the channel
     */
    event ChannelClosed(bytes32 indexed channelId);

    /**
     * @notice Creates a new channel and initializes funding
     * @dev The creator must sign the funding state containing the CHANOPEN magic number
     * @param ch Channel configuration with participants, adjudicator, challenge period, and nonce
     * @param initial Initial state with CHANOPEN magic number and expected allocations
     * @return channelId Unique identifier for the created channel
     */
    function create(Channel calldata ch, State calldata initial) external returns (bytes32 channelId);

    /**
     * @notice Allows a participant to join a channel by signing the funding state
     * @dev Participant must provide signature on the same funding state with CHANOPEN magic number
     * @param channelId Unique identifier for the channel
     * @param index Index of the participant in the channel's participants array
     * @param sig Signature of the participant on the funding state
     * @return channelId Unique identifier for the joined channel
     */
    function join(bytes32 channelId, uint256 index, Signature calldata sig) external returns (bytes32);

    /**
     * @notice Finalizes a channel with a mutually signed closing state
     * @dev Requires all participants' signatures on a state with CHANCLOSE magic number,
     *      or can be called after challenge period expires with the last valid state
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state to be finalized
     * @param proofs Additional states required by the adjudicator to validate the candidate
     */
    function close(bytes32 channelId, State calldata candidate, State[] calldata proofs) external;

    /**
     * @notice Closes an existing channel and creates a new one with updated parameters
     * @dev Used for resizing channel allocations without fully withdrawing funds
     * @param channelId Unique identifier for the channel to close
     * @param candidate The latest known valid state for closing the current channel
     * @param proofs Additional states required by the adjudicator for closing
     * @param ch New channel configuration for the replacement channel
     * @param initial Initial state for the new channel with CHANOPEN magic number
     */
    function reset(
        bytes32 channelId,
        State calldata candidate,
        State[] calldata proofs,
        Channel calldata ch,
        State calldata initial
    ) external;

    /**
     * @notice Initiates or updates a challenge with a signed state
     * @dev Starts a challenge period during which participants can respond with newer states
     * @param channelId Unique identifier for the channel
     * @param candidate The state being submitted as the latest valid state
     * @param proofs Additional states required by the adjudicator to validate the candidate
     */
    function challenge(bytes32 channelId, State calldata candidate, State[] calldata proofs) external;

    /**
     * @notice Records a valid state on-chain without initiating a challenge
     * @dev Used to establish on-chain proof of the latest state to prevent future disputes
     * @param channelId Unique identifier for the channel
     * @param candidate The state to checkpoint
     * @param proofs Additional states required by the adjudicator to validate the candidate
     */
    function checkpoint(bytes32 channelId, State calldata candidate, State[] calldata proofs) external;
}
