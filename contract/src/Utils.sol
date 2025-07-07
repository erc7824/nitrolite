// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";
import {EIP712} from "lib/openzeppelin-contracts/contracts/utils/cryptography/EIP712.sol";
import {Channel, State, Signature, StateIntent} from "./interfaces/Types.sol";

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
     * @notice Verifies that a state is signed by the specified participant
     * @param domainSeparator The EIP-712 domain separator
     * @param typeHash The EIP-712 type hash
     * @param msgHash The hash of the message to verify the signature against
     * @param sig The signature to verify
     * @param signer The address of the expected signer
     * @return True if the signature is valid, false otherwise
     */
    function verifySignature(bytes32 domainSeparator, bytes32 typeHash, bytes32 msgHash, Signature memory sig, address signer) internal pure returns (bool) {
        address rawECDSASigner = recoverRawECDSASigner(msgHash, sig);
        if (rawECDSASigner == signer) {
            return true;
        }

        address eip191Signer = recoverEIP191Signer(msgHash, sig);
        if (eip191Signer == signer) {
            return true;
        }

        // NOTE: The goal of EIP-712 is user clarity of what they sign, which is shadowed here, because `stateHash` and not the `State` is used.
        address eip712Signer = recoverEIP712Signer(domainSeparator, keccak256(abi.encode(typeHash, msgHash)), sig);
        if (eip712Signer == signer) {
            return true;
        }

        return false;
    }

    /**
     * @notice Checks if any of the provided addresses have signed the given state
     * @param domainSeparator The EIP-712 domain separator
     * @param typeHash The EIP-712 type hash
     * @param msgHash The message hash to verify the signature against
     * @param sig The signature to verify
     * @param possibleSigners The list of addresses to check
     * @return True if any address has signed the state, false otherwise
     */
    function verifySignatureOneOf(bytes32 domainSeparator, bytes32 typeHash, bytes32 msgHash, Signature memory sig, address[] memory possibleSigners) internal pure returns (bool) {
        for (uint256 i = 0; i < possibleSigners.length; i++) {
            if (verifySignature(domainSeparator, typeHash, msgHash, sig, possibleSigners[i])) {
                return true;
            }
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
     * @notice Validates that a state is a valid initial state for a channel
     * @dev Initial states must have version 0 and INITIALIZE intent
     * @param domainSeparator The EIP-712 domain separator
     * @param stateTypeHash The EIP-712 state-specific type hash
     * @param state The state to validate
     * @param chan The channel configuration
     * @return True if the state is a valid initial state, false otherwise
     */
    function validateInitialState(bytes32 domainSeparator, bytes32 stateTypeHash, State memory state, Channel memory chan) internal view returns (bool) {
        if (state.version != 0) {
            return false;
        }

        if (state.intent != StateIntent.INITIALIZE) {
            return false;
        }

        return validateUnanimousSignatures(domainSeparator, stateTypeHash, state, chan);
    }

    /**
     * @notice Validates that a state has signatures from both participants
     * @dev For 2-participant channels, both must sign to establish unanimous consent
     * @param domainSeparator The EIP-712 domain separator
     * @param stateTypeHash The EIP-712 state-specific type hash
     * @param state The state to validate
     * @param chan The channel configuration
     * @return True if the state has valid signatures from both participants, false otherwise
     */
    function validateUnanimousSignatures(bytes32 domainSeparator, bytes32 stateTypeHash, State memory state, Channel memory chan) internal view returns (bool) {
        if (state.sigs.length != 2) {
            return false;
        }

        bytes32 stateHash = getStateHash(chan, state);

        return verifySignature(domainSeparator, stateTypeHash, stateHash, state.sigs[0], chan.participants[CLIENT])
            && verifySignature(domainSeparator, stateTypeHash, stateHash, state.sigs[1], chan.participants[SERVER]);
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
