// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Utils} from "../src/Utils.sol";
import {Channel, State, StateIntent} from "../src/interfaces/Types.sol";

contract TestUtilsContract {
    using Utils for *;

    function getChannelId(Channel memory ch) external view returns (bytes32) {
        return Utils.getChannelId(ch);
    }

    function getStateHash(Channel memory ch, State memory state) external view returns (bytes32) {
        return Utils.getStateHash(ch, state);
    }

    function getStateHashShort(bytes32 channelId, State memory state) external pure returns (bytes32) {
        return Utils.getStateHashShort(channelId, state);
    }

    function recoverRawECDSASigner(bytes32 msgHash, bytes memory sig) external pure returns (address) {
        return Utils.recoverRawECDSASigner(msgHash, sig);
    }

    function recoverEIP191Signer(bytes32 msgHash, bytes memory sig) external pure returns (address) {
        return Utils.recoverEIP191Signer(msgHash, sig);
    }

    function recoverEIP712Signer(bytes32 domainSeparator, bytes32 structHash, bytes memory sig)
        external
        pure
        returns (address)
    {
        return Utils.recoverEIP712Signer(domainSeparator, structHash, sig);
    }

    function recoverStateEIP712Signer(
        bytes32 typeHash,
        bytes32 channelId,
        bytes32 domainSeparator,
        State memory state,
        bytes memory sig
    ) external pure returns (address) {
        return Utils.recoverStateEIP712Signer(typeHash, channelId, domainSeparator, state, sig);
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
        view
        returns (bool)
    {
        return Utils.validateInitialState(state, chan, domainSeparator);
    }

    function validateUnanimousStateSignatures(State memory state, Channel memory chan, bytes32 domainSeparator)
        external
        view
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

    // Constants
    function CLIENT() external pure returns (uint256) {
        return Utils.CLIENT;
    }

    function SERVER() external pure returns (uint256) {
        return Utils.SERVER;
    }

    function NO_EIP712_SUPPORT() external pure returns (bytes32) {
        return Utils.NO_EIP712_SUPPORT;
    }

    function ERC6492_DETECTION_SUFFIX() external pure returns (bytes32) {
        return Utils.ERC6492_DETECTION_SUFFIX;
    }

    function ERC1271_SUCCESS() external pure returns (bytes4) {
        return Utils.ERC1271_SUCCESS;
    }
}