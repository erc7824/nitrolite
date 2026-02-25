// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

/**
 * @title GasConsumingEthReceiver
 * @notice Mock contract that consumes all available gas on ETH receive
 */
contract GasConsumingEthReceiver {
    receive() external payable {
        // Consume all remaining gas
        while (true) {
            keccak256(abi.encodePacked(block.timestamp, msg.sender));
        }
    }

    fallback() external payable {
        // Consume all remaining gas
        while (true) {
            keccak256(abi.encodePacked(block.timestamp, msg.sender));
        }
    }
}
