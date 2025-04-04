// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";
import {Vm} from "lib/forge-std/src/Vm.sol";

library TestUtils {
    using MessageHashUtils for bytes32;

    function sign(Vm vm, uint256 privateKey, bytes32 digest) external pure returns (uint8 v, bytes32 r, bytes32 s) {
        (v, r, s) = vm.sign(privateKey, digest.toEthSignedMessageHash());
    }
}
