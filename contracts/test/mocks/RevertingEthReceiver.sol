// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

/**
 * @title RevertingEthReceiver
 * @notice Mock contract that reverts on ETH receive
 */
contract RevertingEthReceiver {
    error EthReceiveBlocked();

    receive() external payable {
        revert EthReceiveBlocked();
    }

    fallback() external payable {
        revert EthReceiveBlocked();
    }
}
