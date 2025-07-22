// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";
import {EIP712} from "lib/openzeppelin-contracts/contracts/utils/cryptography/EIP712.sol";
import {STATE_TYPEHASH, Channel, State, StateIntent} from "./interfaces/Types.sol";

/**
 * @title Channel Utilities
 * @notice Library providing utility functions for state channel operations
 */
library Utils {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;

    uint256 constant CLIENT = 0;
    uint256 constant SERVER = 1;

    bytes32 constant NO_EIP712_SUPPORT = keccak256("NoEIP712Support");

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
    function recoverRawECDSASigner(bytes32 msgHash, bytes memory sig) internal pure returns (address) {
        // Verify the signature directly on the stateHash without using EIP-191
        return msgHash.recover(sig);
    }

    /**
     * @notice Recovers the signer of a state hash using EIP-191 format
     * @dev NOTE: FIXME: inconsistent with EIP-712 state recovery, which receives channelId and state as "message", whereas EIP-191 receives
     * stateHash directly. This breaks the principle of least astonishment, as in EIP-191 and EIP-712 contexts the message is different.
     * @param msgHash The hash of the message to verify the signature against
     * @param sig The signature to verify
     * @return The address of the signer
     */
    function recoverEIP191Signer(bytes32 msgHash, bytes memory sig) internal pure returns (address) {
        return msgHash.toEthSignedMessageHash().recover(sig);
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
     * @param typeHash The type hash for the state structure
     * @param channelId The unique identifier for the channel
     * @param domainSeparator The EIP-712 domain separator
     * @param state The state to verify
     * @param sig The signature to verify
     * @return The address of the signer
     */
    function recoverStateEIP712Signer(
        bytes32 typeHash,
        bytes32 channelId,
        bytes32 domainSeparator,
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
     * @notice Verifies that a message hash is signed by the specified participant
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
    ) internal pure returns (bool) {
        bytes32 stateHash = Utils.getStateHashShort(channelId, state);

        address rawECDSASigner = Utils.recoverRawECDSASigner(stateHash, sig);
        if (rawECDSASigner == signer) {
            return true;
        }

        address eip191Signer = Utils.recoverEIP191Signer(stateHash, sig);
        if (eip191Signer == signer) {
            return true;
        }

        if (domainSeparator == NO_EIP712_SUPPORT) {
            return false;
        }

        address eip712Signer = Utils.recoverStateEIP712Signer(STATE_TYPEHASH, channelId, domainSeparator, state, sig);
        if (eip712Signer == signer) {
            return true;
        }

        return false;
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
        view
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
        view
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

    function addressArrayIncludes(address[] memory arr, address addr) internal pure returns (bool) {
        for (uint256 i = 0; i < arr.length; i++) {
            if (arr[i] == addr) {
                return true;
            }
        }
        return false;
    }
}
