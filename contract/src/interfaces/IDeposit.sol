// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

/**
 * @title Deposit Interface
 * @notice Interface for contracts that manage token deposits and withdrawals
 * @dev Handles funds that can be allocated to state channels
 */
interface IDeposit {
    /**
     * @notice Emitted when tokens are deposited into the contract
     * @param wallet Address of the account whose ledger is changed
     * @param token Token address (use address(0) for native tokens)
     * @param amount Amount of tokens deposited
     */
    event Deposited(address indexed wallet, address indexed token, uint256 amount);

    /**
     * @notice Emitted when tokens are withdrawn from the contract
     * @param wallet Address of the account whose ledger is changed
     * @param token Token address (use address(0) for native tokens)
     * @param amount Amount of tokens withdrawn
     */
    event Withdrawn(address indexed wallet, address indexed token, uint256 amount);

    /**
     * @notice Gets the balance of a specific account for a given token
     * @param account Address of the account to check
     * @param token Token address (use address(0) for native tokens)
     * @return balance of the account in the specified token
     */
    function getAccountBalance(address account, address token) external view returns (uint256);

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
}
