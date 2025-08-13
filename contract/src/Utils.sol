// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";
import {EIP712} from "lib/openzeppelin-contracts/contracts/utils/cryptography/EIP712.sol";
import {IERC1271} from "lib/openzeppelin-contracts/contracts/interfaces/IERC1271.sol";
import {STATE_TYPEHASH, Channel, State, StateIntent} from "./interfaces/Types.sol";

/**
 * @title Channel Utilities
 * @notice Library providing utility functions for state channel operations
 */
library Utils {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes;
    using {MessageHashUtils.toTypedDataHash} for bytes32;

    error ERC6492DeploymentFailed(address factory, bytes calldata_);
    error ERC6492NoCode(address expectedSigner);

    uint256 public constant CLIENT = 0;
    uint256 public constant SERVER = 1;

    bytes32 public constant NO_EIP712_SUPPORT = keccak256("NoEIP712Support");

    bytes32 public constant ERC6492_DETECTION_SUFFIX =
        0x6492649264926492649264926492649264926492649264926492649264926492;
    bytes4 public constant ERC1271_SUCCESS = 0x1626ba7e;

    /**
     * @notice Compute the unique identifier for a channel
     * @param ch The channel struct
     * @return The channel identifier as bytes32
     */
    function getChannelId(Channel memory ch) internal view returns (bytes32) {
        uint256 chainId;
        assembly {
            chainId := chainid()
        }
        return keccak256(abi.encode(ch.participants, ch.adjudicator, ch.challenge, ch.nonce, chainId));
    }

    /**
     * @notice Packs the channelId and the state into a byte array for signing
     * @param channelId The unique identifier for the channel
     * @param state The state struct to pack
     * @return The packed channelId and state as bytes
     */
    function getPackedState(bytes32 channelId, State memory state) internal pure returns (bytes memory) {
        return abi.encode(channelId, state.intent, state.version, state.data, state.allocations);
    }

    /**
     * @notice Recovers the signer of a state hash from a signature
     * @param message The message to verify the signature against
     * @param sig The signature to verify
     * @return The address of the signer
     */
    function recoverRawECDSASigner(bytes memory message, bytes memory sig) internal pure returns (address) {
        // Verify the signature directly on the message hash without using EIP-191
        return keccak256(message).recover(sig);
    }

    /**
     * @notice Recovers the signer of a state hash using EIP-191 format
     * @param message The message to verify the signature against
     * @param sig The signature to verify
     * @return The address of the signer
     */
    function recoverEIP191Signer(bytes memory message, bytes memory sig) internal pure returns (address) {
        return message.toEthSignedMessageHash().recover(sig);
    }

    /**
     * @notice Recovers the signer of a state hash using the EIP-712 format
     * @param domainSeparator The EIP-712 domain separator
     * @param structHash The hash of the struct to verify the signature against
     * @param sig The signature to verify
     * @return The address of the signer
     */
    function recoverEIP712Signer(bytes32 domainSeparator, bytes32 structHash, bytes memory sig)
        internal
        pure
        returns (address)
    {
        return domainSeparator.toTypedDataHash(structHash).recover(sig);
    }

    /**
     * @notice Recovers the signer of a state using EIP-712 format
     * @param domainSeparator The EIP-712 domain separator
     * @param typeHash The type hash for the state structure
     * @param channelId The unique identifier for the channel
     * @param state The state to verify
     * @param sig The signature to verify
     * @return The address of the signer
     */
    function recoverStateEIP712Signer(
        bytes32 domainSeparator,
        bytes32 typeHash,
        bytes32 channelId,
        State memory state,
        bytes memory sig
    ) internal pure returns (address) {
        return Utils.recoverEIP712Signer(
            domainSeparator,
            keccak256(
                abi.encode(
                    typeHash,
                    channelId,
                    state.intent,
                    state.version,
                    keccak256(state.data),
                    keccak256(abi.encode(state.allocations))
                )
            ),
            sig
        );
    }

    /**
     * @notice Verifies that a state is signed by the specified EOA participant in either raw ECDSA, EIP-191, or EIP-712 format
     * @param state The state to verify
     * @param channelId The ID of the channel
     * @param domainSeparator The EIP-712 domain separator for the channel
     * @param sig The signature to verify
     * @param signer The address of the expected signer
     * @return True if the signature is valid, false otherwise
     */
    function verifyStateEOASignature(
        State memory state,
        bytes32 channelId,
        bytes32 domainSeparator,
        bytes memory sig,
        address signer
    ) internal pure returns (bool) {
        bytes memory packedState = Utils.getPackedState(channelId, state);

        address rawECDSASigner = Utils.recoverRawECDSASigner(packedState, sig);
        if (rawECDSASigner == signer) {
            return true;
        }

        address eip191Signer = Utils.recoverEIP191Signer(packedState, sig);
        if (eip191Signer == signer) {
            return true;
        }

        if (domainSeparator == NO_EIP712_SUPPORT) {
            return false;
        }

        address eip712Signer = Utils.recoverStateEIP712Signer(domainSeparator, STATE_TYPEHASH, channelId, state, sig);
        if (eip712Signer == signer) {
            return true;
        }

        return false;
    }

    /**
     * @notice Checks if a signature is valid by calling the expected signer contract according to the ERC-1271 standard
     * @param msgHash The hash of the message to verify the signature against
     * @param sig The signature to verify
     * @param expectedSigner The address of the expected signer
     * @return True if the signature is valid, false otherwise or if signer is not a contract
     */
    function isValidERC1271Signature(bytes32 msgHash, bytes memory sig, address expectedSigner)
        internal
        view
        returns (bool)
    {
        return IERC1271(expectedSigner).isValidSignature(msgHash, sig) == ERC1271_SUCCESS;
    }

    /**
     * @notice Checks the validity of a smart contract signature. If the expected signer has no code, it is deployed using the provided factory and calldata from the signature.
     * Otherwise, it checks the signature using the ERC-1271 standard.
     * @param msgHash The hash of the message to verify the signature against
     * @param sig The signature to verify
     * @param expectedSigner The address of the expected signer
     * @return True if the signature is valid, false otherwise or if signer is not a contract
     */
    function isValidERC6492Signature(bytes32 msgHash, bytes memory sig, address expectedSigner)
        internal
        returns (bool)
    {
        (address create2Factory, bytes memory factoryCalldata, bytes memory originalSig) =
            abi.decode(sig, (address, bytes, bytes));

        if (expectedSigner.code.length == 0) {
            (bool success,) = create2Factory.call(factoryCalldata);
            require(success, ERC6492DeploymentFailed(create2Factory, factoryCalldata));
            require(expectedSigner.code.length != 0, ERC6492NoCode(expectedSigner));
        }

        return IERC1271(expectedSigner).isValidSignature(msgHash, originalSig) == ERC1271_SUCCESS;
    }

    /**
     * @notice Returns the last 32 bytes of a byte array
     * @param data The byte array to extract from
     * @return result The last 32 bytes of the byte array
     */
    function trailingBytes32(bytes memory data) internal pure returns (bytes32 result) {
        if (data.length < 32) {
            return bytes32(0);
        }
        assembly {
            result := mload(add(data, mload(data)))
        }
    }

    /**
     * @notice Verifies that a state is signed by the specified participant as an EOA or a Smart Contract
     * @param state The state to verify
     * @param channelId The ID of the channel
     * @param domainSeparator The EIP-712 domain separator for the channel
     * @param sig The signature to verify
     * @param signer The address of the expected signer
     * @return True if the signature is valid, false otherwise
     */
    function verifyStateSignature(
        State memory state,
        bytes32 channelId,
        bytes32 domainSeparator,
        bytes memory sig,
        address signer
    ) internal returns (bool) {
        // NOTE: both EIP-1271 and EIP-6492 signatures use message hash
        bytes32 stateHash = keccak256(Utils.getPackedState(channelId, state));

        if (trailingBytes32(sig) == ERC6492_DETECTION_SUFFIX) {
            return isValidERC6492Signature(stateHash, sig, signer);
        }

        if (signer.code.length != 0) {
            return isValidERC1271Signature(stateHash, sig, signer);
        }

        return Utils.verifyStateEOASignature(state, channelId, domainSeparator, sig, signer);
    }

    /**
     * @notice Validates that a state is a valid initial state for a channel
     * @dev Initial states must have version 0 and INITIALIZE intent
     * @param state The state to validate
     * @param chan The channel configuration
     * @param domainSeparator The EIP-712 domain separator for the channel
     * @return True if the state is a valid initial state, false otherwise
     */
    function validateInitialState(State memory state, Channel memory chan, bytes32 domainSeparator)
        internal
        returns (bool)
    {
        if (state.version != 0) {
            return false;
        }

        if (state.intent != StateIntent.INITIALIZE) {
            return false;
        }

        return validateUnanimousStateSignatures(state, chan, domainSeparator);
    }

    /**
     * @notice Validates that a state has signatures from both participants
     * @dev For 2-participant channels, both must sign to establish unanimous consent
     * @param state The state to validate
     * @param chan The channel configuration
     * @param domainSeparator The EIP-712 domain separator for the channel
     * @return True if the state has valid signatures from both participants, false otherwise
     */
    function validateUnanimousStateSignatures(State memory state, Channel memory chan, bytes32 domainSeparator)
        internal
        returns (bool)
    {
        if (state.sigs.length != 2) {
            return false;
        }

        bytes32 channelId = getChannelId(chan);

        return Utils.verifyStateSignature(state, channelId, domainSeparator, state.sigs[0], chan.participants[CLIENT])
            && Utils.verifyStateSignature(state, channelId, domainSeparator, state.sigs[1], chan.participants[SERVER]);
    }

    /**
     * @notice Compares two states for equality
     * @param a The first state to compare
     * @param b The second state to compare
     * @return True if the states are equal, false otherwise
     */
    function statesAreEqual(State memory a, State memory b) internal pure returns (bool) {
        return keccak256(abi.encode(a)) == keccak256(abi.encode(b));
    }

    /**
     * @notice Validates that a state transition is valid according to basic rules
     * @dev Ensures version increments by 1 and total allocation sum remains constant
     * @param previous The previous state
     * @param candidate The candidate new state
     * @return True if the transition is valid, false otherwise
     */
    function validateTransitionTo(State memory previous, State memory candidate) internal pure returns (bool) {
        if (candidate.version != previous.version + 1) {
            return false;
        }

        uint256 candidateSum = candidate.allocations[0].amount + candidate.allocations[1].amount;
        uint256 previousSum = previous.allocations[0].amount + previous.allocations[1].amount;

        if (candidateSum != previousSum) {
            return false;
        }

        return true;
    }
}
