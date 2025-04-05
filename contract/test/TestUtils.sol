// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Vm} from "lib/forge-std/src/Vm.sol";

library TestUtils {
    function sign(Vm vm, uint256 privateKey, bytes32 digest) external pure returns (uint8 v, bytes32 r, bytes32 s) {
        // Sign the digest directly without applying EIP-191 prefix
        (v, r, s) = vm.sign(privateKey, digest);
    }
}
