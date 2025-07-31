// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Utils} from "../src/Utils.sol";
import {Channel, State, StateIntent} from "../src/interfaces/Types.sol";

contract UtilsHarness {
    using Utils for *;

    function getChannelId(Channel memory ch) external view returns (bytes32) {
        return Utils.getChannelId(ch);
    }

    function getPackedState(bytes32 channelId, State memory state) external pure returns (bytes memory) {
        return Utils.getPackedState(channelId, state);
    }

    function recoverRawECDSASigner(bytes memory message, bytes memory sig) external pure returns (address) {
        return Utils.recoverRawECDSASigner(message, sig);
    }

    function recoverEIP191Signer(bytes memory message, bytes memory sig) external pure returns (address) {
        return Utils.recoverEIP191Signer(message, sig);
    }

    function recoverEIP712Signer(bytes32 domainSeparator, bytes32 structHash, bytes memory sig)
        external
        pure
        returns (address)
    {
        return Utils.recoverEIP712Signer(domainSeparator, structHash, sig);
    }

    function recoverStateEIP712Signer(
        bytes32 domainSeparator,
        bytes32 typeHash,
        bytes32 channelId,
        State memory state,
        bytes memory sig
    ) external pure returns (address) {
        return Utils.recoverStateEIP712Signer(domainSeparator, typeHash, channelId, state, sig);
    }

    function verifyStateEOASignature(
        State memory state,
        bytes32 channelId,
        bytes32 domainSeparator,
        bytes memory sig,
        address signer
    ) external pure returns (bool) {
        return Utils.verifyStateEOASignature(state, channelId, domainSeparator, sig, signer);
    }

    function isValidERC1271Signature(bytes32 msgHash, bytes memory sig, address expectedSigner)
        external
        view
        returns (bool)
    {
        return Utils.isValidERC1271Signature(msgHash, sig, expectedSigner);
    }

    function isValidERC6492Signature(bytes32 msgHash, bytes memory sig, address expectedSigner)
        external
        returns (bool)
    {
        return Utils.isValidERC6492Signature(msgHash, sig, expectedSigner);
    }

    function verifyStateSignature(
        State memory state,
        bytes32 channelId,
        bytes32 domainSeparator,
        bytes memory sig,
        address signer
    ) external returns (bool) {
        return Utils.verifyStateSignature(state, channelId, domainSeparator, sig, signer);
    }

    function validateInitialState(State memory state, Channel memory chan, bytes32 domainSeparator)
        external
        returns (bool)
    {
        return Utils.validateInitialState(state, chan, domainSeparator);
    }

    function validateUnanimousStateSignatures(State memory state, Channel memory chan, bytes32 domainSeparator)
        external
        returns (bool)
    {
        return Utils.validateUnanimousStateSignatures(state, chan, domainSeparator);
    }

    function statesAreEqual(State memory a, State memory b) external pure returns (bool) {
        return Utils.statesAreEqual(a, b);
    }

    function validateTransitionTo(State memory previous, State memory candidate) external pure returns (bool) {
        return Utils.validateTransitionTo(previous, candidate);
    }

    function trailingBytes32(bytes memory data) external pure returns (bytes32) {
        return Utils.trailingBytes32(data);
    }
}
