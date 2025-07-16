// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Vm} from "lib/forge-std/src/Vm.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";

library TestUtils {
    function sign(Vm vm, uint256 privateKey, bytes32 digest) external pure returns (bytes memory) {
        // Sign the digest directly without applying EIP-191 prefix
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, digest);
        return abi.encodePacked(r, s, v);
    }

    function signEIP191(Vm vm, uint256 privateKey, bytes32 messageHash) external pure returns (bytes memory) {
        // Apply EIP-191 prefix and sign
        bytes32 ethSignedMessageHash = MessageHashUtils.toEthSignedMessageHash(messageHash);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, ethSignedMessageHash);
        return abi.encodePacked(r, s, v);
    }

    function signEIP712(Vm vm, uint256 privateKey, bytes32 domainSeparator, bytes32 structHash) external pure returns (bytes memory) {
        // Apply EIP-712 prefix and sign
        bytes32 typedDataHash = MessageHashUtils.toTypedDataHash(domainSeparator, structHash);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, typedDataHash);
        return abi.encodePacked(r, s, v);
    }
}
