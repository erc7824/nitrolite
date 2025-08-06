// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {EIP712} from "lib/openzeppelin-contracts/contracts/utils/cryptography/EIP712.sol";

import {Vm} from "lib/forge-std/src/Vm.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";

import {State, Channel} from "../src/interfaces/Types.sol";
import {Utils} from "../src/Utils.sol";

library TestUtils {
    bytes32 public constant TYPE_HASH =
        keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)");

    function buildDomainSeparator(string memory name, string memory version, uint256 chainId, address verifyingContract)
        internal
        pure
        returns (bytes32)
    {
        return keccak256(
            abi.encode(TYPE_HASH, keccak256(bytes(name)), keccak256(bytes(version)), chainId, verifyingContract)
        );
    }

    function buildDomainSeparatorForContract(EIP712 eip712Contract) internal view returns (bytes32) {
        (, string memory name, string memory version, uint256 chainId, address verifyingContract,,) =
            eip712Contract.eip712Domain();
        return buildDomainSeparator(name, version, chainId, verifyingContract);
    }

    function sign(Vm vm, uint256 privateKey, bytes memory message) internal pure returns (bytes memory) {
        // Sign the message directly without applying EIP-191 prefix
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, keccak256(message));
        return abi.encodePacked(r, s, v);
    }

    function signEIP191(Vm vm, uint256 privateKey, bytes memory message) internal pure returns (bytes memory) {
        // Apply EIP-191 prefix and sign
        bytes32 ethSignedMessageHash = MessageHashUtils.toEthSignedMessageHash(message);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, ethSignedMessageHash);
        return abi.encodePacked(r, s, v);
    }

    function signEIP712(Vm vm, uint256 privateKey, bytes32 domainSeparator, bytes32 structHash)
        internal
        pure
        returns (bytes memory)
    {
        // Apply EIP-712 prefix and sign
        bytes32 typedDataHash = MessageHashUtils.toTypedDataHash(domainSeparator, structHash);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, typedDataHash);
        return abi.encodePacked(r, s, v);
    }

    function signStateRaw(Vm vm, bytes32 channelId, State memory state, uint256 privateKey)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory packedState = Utils.getPackedState(channelId, state);
        return sign(vm, privateKey, packedState);
    }

    function signStateEIP191(Vm vm, bytes32 channelId, State memory state, uint256 privateKey)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory packedState = Utils.getPackedState(channelId, state);
        return TestUtils.signEIP191(vm, privateKey, packedState);
    }

    function signStateEIP712(
        Vm vm,
        bytes32 channelId,
        State memory state,
        bytes32 stateTypehash,
        bytes32 domainSeparator,
        uint256 privateKey
    ) internal pure returns (bytes memory) {
        bytes32 structHash = keccak256(
            abi.encode(
                stateTypehash,
                channelId,
                state.intent,
                state.version,
                keccak256(state.data),
                keccak256(abi.encode(state.allocations))
            )
        );
        return TestUtils.signEIP712(vm, privateKey, domainSeparator, structHash);
    }
}
