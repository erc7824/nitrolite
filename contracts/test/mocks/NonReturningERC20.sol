// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title NonReturningERC20
 * @notice Mock ERC20 token that does NOT return bool on transfer (like USDT on some chains)
 */
contract NonReturningERC20 is ERC20 {
    constructor() ERC20("Non-Returning Token", "NRT") {
        _mint(msg.sender, 1000000 * 10 ** 18);
    }

    function transfer(address to, uint256 amount) public override returns (bool) {
        _transfer(msg.sender, to, amount);
        // Simulate non-returning behavior by using assembly to remove return value
        assembly {
            return(0, 0)
        }
    }

    function transferFrom(address from, address to, uint256 amount) public override returns (bool) {
        _spendAllowance(from, msg.sender, amount);
        _transfer(from, to, amount);
        // Simulate non-returning behavior
        assembly {
            return(0, 0)
        }
    }

    function mint(address to, uint256 amount) public {
        _mint(to, amount);
    }
}
