// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {Vm} from "lib/forge-std/src/Vm.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";

import {State, SigValidatorType} from "../src/interfaces/Types.sol";
import {SessionKeyAuthorization} from "../src/sigValidators/SessionKeyValidator.sol";
import {Utils} from "../src/Utils.sol";

library TestUtils {
    function signRaw(Vm vm, uint256 privateKey, bytes memory message) internal pure returns (bytes memory) {
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, keccak256(message));
        return abi.encodePacked(r, s, v);
    }

    function signEip191(Vm vm, uint256 privateKey, bytes memory message) internal pure returns (bytes memory) {
        bytes32 ethSignedMessageHash = MessageHashUtils.toEthSignedMessageHash(message);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, ethSignedMessageHash);
        return abi.encodePacked(r, s, v);
    }

    function signStateEip191WithEcdsaValidator(Vm vm, bytes32 channelId, State memory state, uint256 privateKey)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory packedState = Utils.pack(state, channelId);
        bytes memory signature = TestUtils.signEip191(vm, privateKey, packedState);
        return abi.encodePacked(uint8(SigValidatorType.DEFAULT), signature);
    }

    function signStateEip191WithSkValidator(Vm vm, bytes32 channelId, State memory state, uint256 skPk, SessionKeyAuthorization memory skAuth)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory packedState = Utils.pack(state, channelId);
        bytes memory skSig = TestUtils.signEip191(vm, skPk, packedState);
        bytes memory skModuleSig = abi.encode(skAuth, skSig);
        return abi.encodePacked(uint8(SigValidatorType.CHANNEL), skModuleSig);
    }

    function buildAndSignSkAuth(Vm vm, address sessionKey, bytes32 metadataHash, uint256 authorizerPk) internal pure returns (SessionKeyAuthorization memory) {
        bytes memory authMessage = abi.encode(sessionKey, metadataHash);
        bytes memory signature = TestUtils.signEip191(vm, authorizerPk, authMessage);
        return SessionKeyAuthorization({
            sessionKey: sessionKey,
            metadataHash: metadataHash,
            authSignature: signature
        });
    }

    function buildSkSig(Vm vm, SessionKeyAuthorization memory skAuth, bytes32 channelId, bytes memory signingData, uint256 sessionKeyPk) internal pure returns (bytes memory) {
        bytes memory stateMessage = Utils.pack(channelId, signingData);
        bytes memory signature = TestUtils.signEip191(vm, sessionKeyPk, stateMessage);
        return abi.encode(skAuth, signature);
    }
}
