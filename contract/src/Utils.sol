// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";
import {EIP712} from "lib/openzeppelin-contracts/contracts/utils/cryptography/EIP712.sol";
import {STATE_TYPEHASH, Channel, State, Signature, StateIntent} from "./interfaces/Types.sol";

/**
 * @title Channel Utilities
 * @notice Library providing utility functions for state channel operations
 */
library Utils {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;

    uint256 constant CLIENT = 0;
    uint256 constant SERVER = 1;

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
     * @notice Compute the hash of a channel state in a canonical way (ignoring the signature)
     * @param ch The channel struct
     * @param state The state struct
     * @return The state hash as bytes32
     * @dev The state hash is computed according to the specification in the README, using channelId, data, version, and allocations
     */
    function getStateHash(Channel memory ch, State memory state) internal view returns (bytes32) {
        bytes32 channelId = getChannelId(ch);
        return keccak256(abi.encode(channelId, state.intent, state.version, state.data, state.allocations));
    }

    /**
     * @notice Compute the hash of a channel state in a canonical way (ignoring the signature)
     * @param channelId The unique identifier for the channel
     * @param state The state struct
     * @return The state hash as bytes32
     * @dev The state hash is computed according to the specification in the README, using channelId, data, version, and allocations
     */
    function getStateHashShort(bytes32 channelId, State memory state) internal pure returns (bytes32) {
        return keccak256(abi.encode(channelId, state.intent, state.version, state.data, state.allocations));
    }

    /**
     * @notice Recovers the signer of a state hash from a signature
     * @param msgHash The hash of the message to verify the signature against
     * @param sig The signature to verify
     * @return The address of the signer
     */
    function recoverRawECDSASigner(bytes32 msgHash, Signature memory sig) internal pure returns (address) {
        // Verify the signature directly on the stateHash without using EIP-191
        return msgHash.recover(sig.v, sig.r, sig.s);
    }

    /**
     * @notice Recovers the signer of a state hash using EIP-191 format
     * @dev NOTE: FIXME: inconsistent with EIP-712 state recovery, which receives channelId and state as "message", whereas EIP-191 receives
     * stateHash directly. This breaks the principle of least astonishment, as in EIP-191 and EIP-712 contexts the message is different.
     * @param msgHash The hash of the message to verify the signature against
     * @param sig The signature to verify
     * @return The address of the signer
     */
    function recoverEIP191Signer(bytes32 msgHash, Signature memory sig) internal pure returns (address) {
        return msgHash.toEthSignedMessageHash().recover(sig.v, sig.r, sig.s);
    }

    /**
     * @notice Recovers the signer of a state hash using the EIP-712 format
     * @param domainSeparator The EIP-712 domain separator
     * @param structHash The hash of the struct to verify the signature against
     * @param sig The signature to verify
     * @return The address of the signer
     */
    function recoverEIP712Signer(bytes32 domainSeparator, bytes32 structHash, Signature memory sig)
        internal
        pure
        returns (address)
    {
        return domainSeparator.toTypedDataHash(structHash).recover(sig.v, sig.r, sig.s);
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
    function recoverStateEIP712Signer(bytes32 domainSeparator, bytes32 typeHash, bytes32 channelId, State memory state, Signature memory sig)
        internal
        pure
        returns (address)
    {
        return Utils.recoverEIP712Signer(domainSeparator, keccak256(abi.encode(typeHash, channelId, state.intent, state.version, state.data, state.allocations)), sig);
    }

        /**
     * @notice Verifies that a message hash is signed by the specified participant
     * @param channelId The ID of the channel
     * @param state The state to verify
     * @param sig The signature to verify
     * @param signer The address of the expected signer
     * @return True if the signature is valid, false otherwise
     */
    function verifyStateSignature(bytes32 domainSeparator, bytes32 channelId, State memory state, Signature memory sig, address signer) internal pure returns (bool) {
        bytes32 stateHash = Utils.getStateHashShort(channelId, state);

        address rawECDSASigner = Utils.recoverRawECDSASigner(stateHash, sig);
        if (rawECDSASigner == signer) {
            return true;
        }

        address eip191Signer = Utils.recoverEIP191Signer(stateHash, sig);
        if (eip191Signer == signer) {
            return true;
        }

        address eip712Signer = Utils.recoverStateEIP712Signer(domainSeparator, STATE_TYPEHASH, channelId, state, sig);
        if (eip712Signer == signer) {
            return true;
        }

        return false;
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

    function addressArrayIncludes(address[] memory arr, address addr) internal pure returns (bool) {
        for (uint256 i = 0; i < arr.length; i++) {
            if (arr[i] == addr) {
                return true;
            }
        }
        return false;
    }
}
