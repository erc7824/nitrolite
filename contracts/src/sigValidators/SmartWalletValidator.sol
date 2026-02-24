// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {
    ISignatureValidator,
    ValidationResult,
    VALIDATION_SUCCESS,
    VALIDATION_FAILURE
} from "../interfaces/ISignatureValidator.sol";
import {SwSignatureUtils} from "./SwSignatureUtils.sol";
import {Utils} from "../Utils.sol";

/**
 * @title SmartWalletValidator
 * @notice Signature validator supporting smart contract wallets via ERC-4337 standards
 * @dev Supports ERC-1271 and ERC-6492 signatures.
 *
 * The validator prepends channelId to the signingData to construct the full message,
 * then attempts validation in the order listed above.
 */
contract SmartWalletValidator is ISignatureValidator {
    /**
     * @notice Validates a signature from either an EOA or smart contract wallet
     * @dev Constructs the full message by prepending channelId to signingData,
     *      then tries validation in order: ERC-1271, ERC-6492
     * @param channelId The channel identifier to include in the signed message
     * @param signingData The encoded state data (without channelId or signatures)
     * @param signature The signature to validate (format varies by wallet type)
     * @param participant The expected signer's address
     * @return result VALIDATION_SUCCESS if valid, VALIDATION_FAILURE otherwise
     */
    function validateSignature(
        bytes32 channelId,
        bytes calldata signingData,
        bytes calldata signature,
        address participant
    ) external returns (ValidationResult) {
        bytes memory message = Utils.pack(channelId, signingData);

        if (SwSignatureUtils.validateSmartWalletSigner(message, signature, participant)) {
            return VALIDATION_SUCCESS;
        } else {
            return VALIDATION_FAILURE;
        }
    }
}
