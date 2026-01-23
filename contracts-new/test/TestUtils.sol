// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {Vm} from "lib/forge-std/src/Vm.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";

import {State} from "../src/interfaces/Types.sol";
import {Utils} from "../src/Utils.sol";

library TestUtils {
    function signEIP191(Vm vm, uint256 privateKey, bytes memory message) internal pure returns (bytes memory) {
        // Apply EIP-191 prefix and sign
        bytes32 ethSignedMessageHash = MessageHashUtils.toEthSignedMessageHash(message);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, ethSignedMessageHash);
        return abi.encodePacked(r, s, v);
    }

    function signStateEIP191(Vm vm, bytes32 channelId, State memory state, uint256 privateKey)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory packedState = Utils.pack(state, channelId);
        return TestUtils.signEIP191(vm, privateKey, packedState);
    }
}
