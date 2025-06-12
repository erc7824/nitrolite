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
     * @notice Gets the balances of multiple accounts for multiple tokens
     * @dev Returns a 2D array where each inner array corresponds to the balances of the tokens for each account
     * @param accounts Array of account addresses to check balances for
     * @param tokens Array of token addresses to check balances for (use address(0) for native tokens)
     * @return A 2D array of balances, where each inner array corresponds to the balances of the tokens for each account
     */
    function getAccountsBalances(address[] calldata accounts, address[] calldata tokens)
        external
        view
        returns (uint256[][] memory);

    /**
     * @notice Deposits tokens into the contract
     * @dev For native tokens, the value should be sent with the transaction
     * @param account Address of the account whose ledger is changed
     * @param token Token address (use address(0) for native tokens)
     * @param amount Amount of tokens to deposit
     */
    function deposit(address account, address token, uint256 amount) external payable;

    /**
     * @notice Withdraws tokens from the contract
     * @dev Can only withdraw available (not locked in channels) funds
     * @param token Token address (use address(0) for native tokens)
     * @param amount Amount of tokens to withdraw
     */
    function withdraw(address token, uint256 amount) external;
}
