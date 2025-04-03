// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";
import {Channel, State, Signature} from "./interfaces/Types.sol";

/**
 * @title Channel Utilities
 * @notice Library providing utility functions for state channel operations
 */
library Utils {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;

    /**
     * @notice Compute the unique identifier for a channel
     * @param ch The channel struct
     * @return The channel identifier as bytes32
     */
    function getChannelId(Channel memory ch) internal pure returns (bytes32) {
        return keccak256(abi.encode(ch.participants, ch.adjudicator, ch.challenge, ch.nonce));
    }

    /**
     * @notice Compute the hash of a channel state in a canonical way (ignoring the signature)
     * @param ch The channel struct
     * @param state The state struct
     * @return The state hash as bytes32
     * @dev The state hash is computed according to the specification in the README, using channelId, data, and allocations
     */
    function getStateHash(Channel memory ch, State memory state) internal pure returns (bytes32) {
        bytes32 channelId = getChannelId(ch).toEthSignedMessageHash();
        return keccak256(abi.encode(channelId, state.data, state.allocations));
    }

    /**
     * @notice Verifies that a state is signed by the specified participant
     * @param stateHash The hash of the state to verify (computed using the canonical form)
     * @param sig The signature to verify
     * @param signer The address of the expected signer
     * @return True if the signature is valid, false otherwise
     */
    function verifySignature(bytes32 stateHash, Signature memory sig, address signer) internal pure returns (bool) {
        // Verify the signature
        address recoveredSigner = stateHash.recover(sig.v, sig.r, sig.s);
        return recoveredSigner == signer;
    }
}
