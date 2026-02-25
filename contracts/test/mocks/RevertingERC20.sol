// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title RevertingERC20
 * @notice Mock ERC20 token that always reverts on transfer (simulates blacklist behavior)
 */
contract RevertingERC20 is ERC20 {
    error TransferBlocked();

    constructor() ERC20("Reverting Token", "RVT") {
        _mint(msg.sender, 1000000 * 10 ** 18);
    }

    function transfer(address, uint256) public pure override returns (bool) {
        revert TransferBlocked();
    }

    function transferFrom(address, address, uint256) public pure override returns (bool) {
        revert TransferBlocked();
    }

    function mint(address to, uint256 amount) public {
        _mint(to, amount);
    }
}
