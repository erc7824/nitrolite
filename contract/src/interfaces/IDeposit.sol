// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

/**
 * @title Deposit Interface
 * @notice Interface for contracts that manage token deposits and withdrawals
 * @dev Handles funds that can be allocated to state channels
 */
interface IDeposit {
    /**
     * @notice Deposits tokens into the contract
     * @dev For native tokens, the value should be sent with the transaction
     * @param token Token address (use address(0) for native tokens)
     * @param amount Amount of tokens to deposit
     */
    function deposit(address token, uint256 amount) external payable;

    /**
     * @notice Withdraws tokens from the contract
     * @dev Can only withdraw available (not locked in channels) funds
     * @param token Token address (use address(0) for native tokens)
     * @param amount Amount of tokens to withdraw
     */
    function withdraw(address token, uint256 amount) external;

    /**
     * @notice Returns the list of channels associated with an account
     * @param account The account address
     * @return Array of channel identifiers linked to the account
     */
    function getAccountChannels(address account) external view returns (bytes32[] memory);

    /**
     * @notice Returns the account information for a specific token
     * @param user The account address
     * @param token The token address
     * @return available Amount available for withdrawal or allocation
     * @return locked Amount locked in active channels
     * @return channelCount Number of associated channels
     */
    function getAccountInfo(address user, address token)
        external
        view
        returns (uint256 available, uint256 locked, uint256 channelCount);
}
