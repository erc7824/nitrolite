// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {Channel, State, ChannelStatus, Amount} from "./Types.sol";

interface IChannelReader {
    /**
     * @notice Get the list of open channels for a list of accounts
     * @param accounts Array of account addresses to check for open channels
     * @return Array of arrays, where each inner array contains channel IDs for the corresponding account
     */
    function getOpenChannels(address[] memory accounts) external view returns (bytes32[][] memory);

    /**
     * @notice Get the config of a specific channel
     * @param channelId The unique identifier of the channel
     * @return Channel configuration for the specified channel
     */
    function getChannelConfig(bytes32 channelId) external view returns (Channel memory);

    /**
     * @notice Get the status of a specific channel
     * @param channelId The unique identifier of the channel
     * @return ChannelStatus representing the current status of the channel
     */
    function getChannelStatus(bytes32 channelId) external view returns (ChannelStatus);

    /**
     * @notice Get the last valid state of a specific channel
     * @param channelId The unique identifier of the channel
     * @return State representing the last valid state of the channel
     */
    function getChannelLastValidState(bytes32 channelId) external view returns (State memory);

    /**
     * @notice Get the challenge expiry time for a specific channel
     * @param channelId The unique identifier of the channel
     * @return uint256 representing the challenge expiry timestamp
     */
    function getChannelChallengeExpiry(bytes32 channelId) external view returns (uint256);

    /**
     * @notice Get detailed information about a specific channel
     * @param channelId The unique identifier of the channel
     * @return channel The Channel configuration
     * @return status The current status of the channel
     * @return challengeExpiry The challenge expiry timestamp
     * @return lastValidState The last valid state of the channel
     */
    function getChannelData(bytes32 channelId)
        external
        view
        returns (
            Channel memory channel,
            ChannelStatus status,
            uint256 challengeExpiry,
            State memory lastValidState
        );

    /**
     * @notice Get the balance of a channel for a list of tokens
     * @param channelId The unique identifier of the channel
     * @param tokens Array of token addresses to check balances for (use address(0) for native tokens)
     * @return balances Array of balances corresponding to the provided tokens
     */
    function getChannelBalances(bytes32 channelId, address[] memory tokens) external view returns (uint256[] memory balances);
}
