// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {
    ISignatureValidator,
    ValidationResult,
    VALIDATION_SUCCESS,
    VALIDATION_FAILURE
} from "../interfaces/ISignatureValidator.sol";
import {EcdsaSignatureUtils} from "./EcdsaSignatureUtils.sol";
import {SwSignatureUtils} from "./SwSignatureUtils.sol";
import {Utils} from "../Utils.sol";

/**
 * @notice Authorization struct for delegating signing authority to a session key
 * @dev The participant signs this authorization to allow the session key to sign on their behalf
 * @param sessionKey The address of the delegated session key
 * @param metadataHash Hashed application-specific data (e.g., expiration timestamp, nonce, permissions)
 * @param authSignature The participant's signature authorizing this session key (65 bytes ECDSA)
 */
struct SessionKeyAuthorization {
    address sessionKey;
    bytes32 metadataHash;
    bytes authSignature;
}

function toSigningData(SessionKeyAuthorization memory skAuth) pure returns (bytes memory) {
    return abi.encode(
        skAuth.sessionKey,
        skAuth.metadataHash
        // omit authSignature
    );
}

/**
 * @title SessionKeyValidator
 * @notice Validator supporting session key delegation for temporary signing authority
 * @dev Enables a participant to delegate signing authority to a session key with metadata.
 *      Useful for hot wallets, time-limited access, or gasless transactions.
 *      Supports both EOA (ECDSA) and smart contract wallet (ERC-1271/ERC-6492) signatures.
 *
 * Authorization Flow:
 * 1. Participant signs a SessionKeyAuthorization to delegate to a session key
 * 2. Session key signs the actual state data
 * 3. Both signatures are validated on-chain
 *
 * Signature Format:
 * bytes sigBody = abi.encode(SessionKeyAuthorization skAuthorization, bytes signature)
 *
 * Where signature can be:
 * - Standard 65-byte EIP-191 or raw ECDSA signature (for EOAs)
 * - ERC-1271 signature (for deployed smart wallets)
 * - ERC-6492 signature (for undeployed smart wallets)
 *
 * Security Model:
 * - Off-chain enforcement (Clearnode) should validate session key expiration and usage limits
 * - On-chain validation only checks cryptographic validity
 * - Participants are responsible for session key management
 */
contract SessionKeyValidator is ISignatureValidator {
    /**
     * @notice Validates a signature using a delegated session key
     * @dev Validates:
     *      1. participant signed the SessionKeyAuthorization (with channelId binding)
     *      2. sessionKey signed the full state message (channelId + signingData)
     *      Supports both EOA (ECDSA) and smart wallet (ERC-1271/ERC-6492) signatures.
     *      Tries ECDSA validation first, then smart wallet validation for both signatures.
     * @param channelId The channel identifier to include in state messages
     * @param signingData The encoded state data (without channelId or signatures)
     * @param signature Encoded as abi.encode(SessionKeyAuthorization, bytes signature)
     * @param participant The expected authorizing participant's address
     * @return result VALIDATION_SUCCESS if valid, VALIDATION_FAILURE otherwise
     */
    function validateSignature(
        bytes32 channelId,
        bytes calldata signingData,
        bytes calldata signature,
        address participant
    ) external returns (ValidationResult) {
        (SessionKeyAuthorization memory skAuth, bytes memory skSignature) =
            abi.decode(signature, (SessionKeyAuthorization, bytes));

        // Step 1: Verify participant authorized this session key
        bytes memory authMessage = toSigningData(skAuth);
        bool authResult = _validateSigner(authMessage, skAuth.authSignature, participant);

        if (!authResult) {
            return VALIDATION_FAILURE;
        }

        // Step 2: Verify session key signed the full state message
        bytes memory stateMessage = Utils.pack(channelId, signingData);
        if (_validateSigner(stateMessage, skSignature, skAuth.sessionKey)) {
            return VALIDATION_SUCCESS;
        } else {
            return VALIDATION_FAILURE;
        }
    }

    /**
     * @notice Validates a signature for a given signer (EOA or smart wallet)
     * @dev Tries ECDSA validation first, then smart wallet validation
     * @param message The message that was signed
     * @param signature The signature to validate
     * @param signer The expected signer's address
     * @return bool True if signature is valid, false otherwise
     */
    function _validateSigner(bytes memory message, bytes memory signature, address signer)
        private
        returns (bool)
    {
        if (signer.code.length != 0 || SwSignatureUtils.isERC6492Signature(signature)) {
            // If signer has code or signature is ERC-6492, treat as smart wallet
            return SwSignatureUtils.validateSmartWalletSigner(message, signature, signer);
        }

        return EcdsaSignatureUtils.validateEcdsaSigner(message, signature, signer);
    }
}
