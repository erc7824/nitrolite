// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {Channel, State, ChannelStatus, Amount} from "./Types.sol";

interface IChannelReader {
    /**
     * @notice Get a list of all channel IDs associated with the caller
     * @return Array of channel IDs
     */
    function getAccountChannels(address account) external view returns (bytes32[] memory);

    /**
     * @notice Get detailed information about a specific channel
     * @param channelId The unique identifier of the channel
     * @return exists Whether the channel exists
     * @return channel The channel configuration
     * @return status Current channel status
     * @return creator Address that created the channel
     * @return expectedDeposits Expected deposits for each participant
     * @return actualDeposits Actual deposits made by each participant
     * @return challengeExpiry Timestamp when challenge period expires (0 if no active challenge)
     * @return lastValidState The last valid state of the channel
     */
    function getChannelInfo(bytes32 channelId)
        external
        view
        returns (
            bool exists,
            Channel memory channel,
            ChannelStatus status, 
            address creator,
            Amount[2] memory expectedDeposits,
            Amount[2] memory actualDeposits,
            uint256 challengeExpiry,
            State memory lastValidState
        );

    /**
     * @notice Get token balance information for a specific channel
     * @param channelId The unique identifier of the channel
     * @param token The token address (zero address for native token)
     * @return balance The current balance of the specified token in the channel
     */
    function getChannelBalance(bytes32 channelId, address token) external view returns (uint256 balance);

    /**
     * @notice Get account information for a specific token
     * @param user The address of the user
     * @param token The token address (zero address for native token)
     * @return balance The available balance of the specified token for the user
     */
    function getAccountBalance(address user, address token) external view returns (uint256 balance);
}
