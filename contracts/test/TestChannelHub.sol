// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHub} from "../src/ChannelHub.sol";
import {ISignatureValidator} from "../src/interfaces/ISignatureValidator.sol";

/**
 * @title TestChannelHub
 * @notice Test harness contract that exposes internal ChannelHub functions for testing
 */
contract TestChannelHub is ChannelHub {
    constructor(ISignatureValidator _defaultSigValidator) ChannelHub(_defaultSigValidator) {}

    /**
     * @notice Exposed version of _pushFunds for testing
     */
    function exposed_pushFunds(address to, address token, uint256 amount) external payable {
        _pushFunds(to, token, amount);
    }

    /**
     * @notice Exposed version of _pullFunds for testing
     */
    function exposed_pullFunds(address from, address token, uint256 amount) external payable {
        _pullFunds(from, token, amount);
    }
}
