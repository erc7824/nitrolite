// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

/**
 * @title Deposit Interface
 * @notice Interface for contracts that allow users to deposit and withdraw token funds
 */
interface IDeposit {
    /**
     * @notice Deposits tokens into the contract
     * @dev Any user can deposit tokens
     * @param token Address of the ERC20 token to deposit
     * @param amount Amount of tokens to deposit
     */
    function deposit(address token, uint256 amount) external;

    /**
     * @notice Withdraws tokens from the contract
     * @dev Any user can withdraw their previously deposited tokens
     * @param token Address of the ERC20 token to withdraw
     * @param amount Amount of tokens to withdraw
     */
    function withdraw(address token, uint256 amount) external;
}
