// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {ERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/ERC20.sol";

/**
 * @title Mock ERC20 Token
 * @notice A simple ERC20 token mock for testing purposes
 * @dev Includes features to test failure cases
 */
contract MockERC20 is ERC20 {
    uint8 private _decimals;
    bool private _failTransfers;

    constructor(string memory name, string memory symbol, uint8 decimals_) ERC20(name, symbol) {
        _decimals = decimals_;
    }

    function mint(address to, uint256 amount) public {
        _mint(to, amount);
    }

    function burn(address from, uint256 amount) public {
        _burn(from, amount);
    }

    function decimals() public view override returns (uint8) {
        return _decimals;
    }

    function setFailTransfers(bool fail) public {
        _failTransfers = fail;
    }

    function transfer(address to, uint256 amount) public override returns (bool) {
        if (_failTransfers) {
            return false;
        }
        return super.transfer(to, amount);
    }

    function transferFrom(address from, address to, uint256 amount) public override returns (bool) {
        if (_failTransfers) {
            return false;
        }
        return super.transferFrom(from, to, amount);
    }
}
