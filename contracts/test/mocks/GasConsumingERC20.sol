// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title GasConsumingERC20
 * @notice Mock ERC20 token that consumes all available gas during transfer
 */
contract GasConsumingERC20 is ERC20 {
    constructor() ERC20("Gas Consuming Token", "GCT") {
        _mint(msg.sender, 1000000 * 10 ** 18);
    }

    function transfer(address to, uint256 amount) public override returns (bool) {
        _transfer(msg.sender, to, amount);

        // Consume all remaining gas
        while (true) {
            // Infinite loop until out of gas
            keccak256(abi.encodePacked(block.timestamp, msg.sender));
        }

        return true;
    }

    function transferFrom(address from, address to, uint256 amount) public override returns (bool) {
        _spendAllowance(from, msg.sender, amount);
        _transfer(from, to, amount);

        // Consume all remaining gas
        while (true) {
            keccak256(abi.encodePacked(block.timestamp, msg.sender));
        }

        return true;
    }

    function mint(address to, uint256 amount) public {
        _mint(to, amount);
    }
}
