// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";

/**
 * @title EcdsaSignatureUtils
 * @notice Utility library for ECDSA signature validation
 * @dev Provides flexible ECDSA signature recovery that tries both EIP-191 and raw ECDSA formats.
 *      Used by validators and ChannelHub for signature verification.
 */
library EcdsaSignatureUtils {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes;

    /**
     * @notice Validates that a signature was created by an expected signer
     * @dev Tries EIP-191 recovery first (with Ethereum signed message prefix), then raw ECDSA if that fails.
     *      Hashes the message internally before recovery.
     *      Returns false for invalid signature formats (e.g., wrong length).
     * @param message The message that was signed (will be hashed internally)
     * @param signature The signature to validate (65 bytes ECDSA: r, s, v)
     * @param expectedSigner The address that should have signed the message
     * @return bool True if signature is valid and from expectedSigner, false otherwise
     */
    function validateEcdsaSigner(bytes memory message, bytes memory signature, address expectedSigner)
        internal
        pure
        returns (bool)
    {
        // ECDSA signatures must be exactly 65 bytes (r: 32, s: 32, v: 1)
        // Return false for other lengths (e.g., ERC-6492 wrapped signatures)
        if (signature.length != 65) {
            return false;
        }

        bytes32 eip191Digest = message.toEthSignedMessageHash();
        address recovered = eip191Digest.recover(signature);

        if (recovered == expectedSigner) {
            return true;
        }

        recovered = keccak256(message).recover(signature);

        if (recovered == expectedSigner) {
            return true;
        }

        return false;
    }
}
