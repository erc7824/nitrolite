// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {Vm} from "lib/forge-std/src/Vm.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";

import {State, SigValidatorType} from "../src/interfaces/Types.sol";
import {Utils} from "../src/Utils.sol";

library TestUtils {
    function signEip191(Vm vm, uint256 privateKey, bytes memory message) internal pure returns (bytes memory) {
        // Apply EIP-191 prefix and sign
        bytes32 ethSignedMessageHash = MessageHashUtils.toEthSignedMessageHash(message);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, ethSignedMessageHash);
        return abi.encodePacked(r, s, v);
    }

    function signStateEip191WithDefaultValidator(Vm vm, bytes32 channelId, State memory state, uint256 privateKey)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory packedState = Utils.pack(state, channelId);
        bytes memory signature = TestUtils.signEip191(vm, privateKey, packedState);
        return abi.encodePacked(uint8(SigValidatorType.DEFAULT), signature);
    }

    function signStateEip191WithChannelValidator(Vm vm, bytes32 channelId, State memory state, uint256 privateKey)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory packedState = Utils.pack(state, channelId);
        bytes memory signature = TestUtils.signEip191(vm, privateKey, packedState);
        return abi.encodePacked(uint8(SigValidatorType.CHANNEL), signature);
    }
}
