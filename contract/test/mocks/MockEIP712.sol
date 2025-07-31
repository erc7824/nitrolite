// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {EIP712} from "lib/openzeppelin-contracts/contracts/utils/cryptography/EIP712.sol";

contract MockEIP712 is EIP712 {
    constructor(string memory name, string memory version) EIP712(name, version) {}

    function domainSeparator() external view returns (bytes32) {
        return _domainSeparatorV4();
    }
}
