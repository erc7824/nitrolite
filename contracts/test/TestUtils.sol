// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {Vm} from "lib/forge-std/src/Vm.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";

import {DEFAULT_SIG_VALIDATOR_ID, State} from "../src/interfaces/Types.sol";
import {SessionKeyAuthorization, toSigningData} from "../src/sigValidators/SessionKeyValidator.sol";
import {Utils} from "../src/Utils.sol";

uint8 constant SESSION_KEY_VALIDATOR_ID = 1;

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
        return abi.encodePacked(DEFAULT_SIG_VALIDATOR_ID, signature);
    }

    function signStateEip191WithSkValidator(
        Vm vm,
        bytes32 channelId,
        State memory state,
        uint256 skPk,
        SessionKeyAuthorization memory skAuth
    ) internal pure returns (bytes memory) {
        bytes memory packedState = Utils.pack(state, channelId);
        bytes memory skSig = TestUtils.signEip191(vm, skPk, packedState);
        bytes memory skModuleSig = abi.encode(skAuth, skSig);
        return abi.encodePacked(SESSION_KEY_VALIDATOR_ID, skModuleSig);
    }

    function buildAndSignSkAuth(Vm vm, address sessionKey, bytes32 metadataHash, uint256 authorizerPk)
        internal
        pure
        returns (SessionKeyAuthorization memory)
    {
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({sessionKey: sessionKey, metadataHash: metadataHash, authSignature: ""})
        );
        bytes memory signature = TestUtils.signEip191(vm, authorizerPk, authMessage);
        return SessionKeyAuthorization({sessionKey: sessionKey, metadataHash: metadataHash, authSignature: signature});
    }

    function buildAndSignValidatorRegistration(Vm vm, uint8 validatorId, address validatorAddress, uint256 nodePk)
        internal
        view
        returns (bytes memory)
    {
        bytes memory message = abi.encode(validatorId, validatorAddress, block.chainid);
        return signEip191(vm, nodePk, message);
    }
}
