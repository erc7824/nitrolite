// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {EIP712AdjudicatorBase} from "../../src/adjudicators/EIP712AdjudicatorBase.sol";

/**
 * @title Test EIP712 Adjudicator
 * @notice Test contract that inherits from EIP712AdjudicatorBase for testing purposes.
 */
contract TestEIP712Adjudicator is EIP712AdjudicatorBase {
    constructor(address owner, address channelImpl_) EIP712AdjudicatorBase(owner, channelImpl_) {}
}
