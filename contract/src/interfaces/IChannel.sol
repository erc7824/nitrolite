// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Channel, State} from "./Types.sol";

/**
 * @title State Channel Interface
 * @notice Main interface for the state channel system
 */
interface IChannel {
    event ChannelPartiallyFunded(bytes32 indexed channelId, Channel channel);
    event ChannelOpened(bytes32 indexed channelId, Channel channel);
    event ChannelChallenged(bytes32 indexed channelId, uint256 expiration);
    event ChannelCheckpointed(bytes32 indexed channelId);
    event ChannelClosed(bytes32 indexed channelId);

    /**
     * @notice Open or join a channel by depositing assets
     * @param ch Channel configuration
     * @param deposit is the initial State defined by the opener, it contains the expected allocation
     * @return channelId Unique identifier for the channel
     */
    function open(Channel calldata ch, State calldata deposit) external returns (bytes32 channelId);

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
     * @param deposit is the initial State defined by the opener, it contains the expected allocation
     */
    function reset(
        bytes32 channelId,
        State calldata candidate,
        State[] calldata proofs,
        Channel calldata ch,
        State calldata deposit
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

    /**
     * @notice Conclude the channel after challenge period expires
     * @param channelId Unique identifier for the channel
     */
    function reclaim(bytes32 channelId) external;
}
