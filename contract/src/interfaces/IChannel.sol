// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Channel, State} from "./Types.sol";

/**
 * @title State Channel Interface
 * @notice Main interface for the state channel system
 */
interface IChannel {
    event Created(bytes32 indexed channelId, Channel channel, Amount[] expected);
    event Joined(bytes32 indexed channelId, uint256 index);
    event Opened(bytes32 indexed channelId);
    event Challenged(bytes32 indexed channelId, uint256 expiration);
    event Checkpointed(bytes32 indexed channelId);
    event Closed(bytes32 indexed channelId);

    /**
     * @notice Create a channel and allocate assets
     * @param ch Channel configuration
     * @param initial is the initial State defined by the opener, it contains the expected allocation
     * @return channelId Unique identifier for the channel
     */
    function create(Channel calldata ch, State calldata initial) external returns (bytes32 channelId);

    /**
     * @notice Join a channel and allocate assets
     * @param channelId Channel hash
     * @param index of participant funding
     * @return channelId Unique identifier for the channel
     */
    function join(bytes32 channelId, uint256 index, Signature sig) external returns (bytes32 channelId);

    /**
     * @notice Finalize the channel with a mutually signed state
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function close(bytes32 channelId, State calldata candidate, State[] calldata proofs) external;

    /**
     * @notice Reset will close and open channel for resizing allocations
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     * @param ch Channel configuration
     * @param initial is the initial State defined by the opener, it contains the expected allocation
     */
    function reset(
        bytes32 channelId,
        State calldata candidate,
        State[] calldata proofs,
        Channel calldata ch,
        State calldata initial
    ) external;

    /**
     * @notice Unilaterally post a state when the other party is uncooperative
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function challenge(bytes32 channelId, State calldata candidate, State[] calldata proofs) external;

    /**
     * @notice Unilaterally post a state to store it on-chain to prevent future disputes
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function checkpoint(bytes32 channelId, State calldata candidate, State[] calldata proofs) external;

}
