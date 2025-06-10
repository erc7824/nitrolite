// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";
import {Channel, State, Signature, StateIntent} from "./interfaces/Types.sol";

/**
 * @title Channel Utilities
 * @notice Library providing utility functions for state channel operations
 */
library Utils {
    using ECDSA for bytes32;

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
     * @param stateHash The hash of the state to verify (computed using the canonical form)
     * @param sig The signature to verify
     * @return The address of the signer
     */
    function recoverSigner(bytes32 stateHash, Signature memory sig) internal pure returns (address) {
        // Verify the signature directly on the stateHash without using EIP-191
        return stateHash.recover(sig.v, sig.r, sig.s);
    }

    /**
     * @notice Verifies that a state is signed by the specified participant
     * @param stateHash The hash of the state to verify (computed using the canonical form)
     * @param sig The signature to verify
     * @param signer The address of the expected signer
     * @return True if the signature is valid, false otherwise
     */
    function verifySignature(bytes32 stateHash, Signature memory sig, address signer) internal pure returns (bool) {
        // Verify the signature directly on the stateHash without using EIP-191
        address recoveredSigner = ECDSA.recover(stateHash, sig.v, sig.r, sig.s);
        return recoveredSigner == signer;
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
     * @param state The state to validate
     * @param chan The channel configuration
     * @return True if the state is a valid initial state, false otherwise
     */
    function validateInitialState(State memory state, Channel memory chan) internal view returns (bool) {
        if (state.version != 0) {
            return false;
        }

        if (state.intent != StateIntent.INITIALIZE) {
            return false;
        }

        return validateUnanimousSignatures(state, chan);
    }

    /**
     * @notice Validates that a state has signatures from both participants
     * @dev For 2-participant channels, both must sign to establish unanimous consent
     * @param state The state to validate
     * @param chan The channel configuration
     * @return True if the state has valid signatures from both participants, false otherwise
     */
    function validateUnanimousSignatures(State memory state, Channel memory chan) internal view returns (bool) {
        if (state.sigs.length != 2) {
            return false;
        }

        // Compute the state hash for signature verification.
        bytes32 stateHash = getStateHash(chan, state);

        return verifySignature(stateHash, state.sigs[0], chan.participants[CLIENT])
            && verifySignature(stateHash, state.sigs[1], chan.participants[SERVER]);
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
