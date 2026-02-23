// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {IERC1271} from "@openzeppelin/contracts/interfaces/IERC1271.sol";
import {MessageHashUtils} from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";

/**
 * @title SwSignatureUtils
 * @notice Utility library for Smart Wallet signature validation
 */
library SwSignatureUtils {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes;

    bytes4 private constant ERC1271_MAGIC_VALUE = 0x1626ba7e;

    bytes32 private constant ERC6492_MAGIC_SUFFIX = 0x6492649264926492649264926492649264926492649264926492649264926492;

    /**
     * @notice Thrown when ERC-6492 contract deployment fails
     * @param factory The factory address that failed to deploy
     * @param factoryCalldata The calldata that was used for deployment
     */
    error ERC6492DeploymentFailed(address factory, bytes factoryCalldata);

    /**
     * @notice Thrown when deployed contract has no code after ERC-6492 deployment
     * @param expectedSigner The address that should have been deployed
     */
    error ERC6492NoCode(address expectedSigner);

    /**
     * @notice Validates a signature from a smart contract wallet only (no ECDSA)
     * @dev Attempts validation in order: ERC-1271 (deployed), ERC-6492 (non-deployed)
     * @param message The message that was signed (will be hashed internally)
     * @param signature The signature to validate (format varies by wallet type)
     * @param expectedSigner The smart contract wallet address
     * @return bool True if signature is valid according to ERC-1271 or ERC-6492
     */
    function validateSmartWalletSigner(bytes memory message, bytes memory signature, address expectedSigner)
        internal
        returns (bool)
    {
        bytes32 messageHash = keccak256(message);

        if (expectedSigner.code.length > 0) {
            if (isValidErc1271Signature(expectedSigner, messageHash, signature)) {
                return true;
            }
        }

        if (isERC6492Signature(signature)) {
            return isValidERC6492Signature(messageHash, signature, expectedSigner);
        }

        return false;
    }

    /**
     * @notice Validates an ERC-1271 signature for a deployed smart contract wallet
     * @param signer The smart contract wallet address
     * @param hash The hash of the signed message
     * @param signature The signature bytes
     * @return bool True if the contract returns the ERC-1271 magic value
     */
    function isValidErc1271Signature(address signer, bytes32 hash, bytes memory signature) internal view returns (bool) {
        try IERC1271(signer).isValidSignature(hash, signature) returns (bytes4 magicValue) {
            return magicValue == ERC1271_MAGIC_VALUE;
        } catch {
            return false;
        }
    }

    /**
     * @notice Checks if a signature is wrapped in ERC-6492 format
     * @dev ERC-6492 signatures end with the magic suffix
     * @param signature The signature to check
     * @return bool True if signature contains ERC-6492 magic suffix
     */
    function isERC6492Signature(bytes memory signature) internal pure returns (bool) {
        if (signature.length < 32) {
            return false;
        }

        bytes32 suffix;
        assembly {
            suffix := mload(add(signature, mload(signature)))
        }

        return suffix == ERC6492_MAGIC_SUFFIX;
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
        // Extract the signature data (remove the magic suffix)
        uint256 dataLength = sig.length - 32;
        bytes memory signatureData = new bytes(dataLength);

        for (uint256 i = 0; i < dataLength; i++) {
            signatureData[i] = sig[i];
        }

        (address create2Factory, bytes memory factoryCalldata, bytes memory originalSig) =
            abi.decode(signatureData, (address, bytes, bytes));

        if (expectedSigner.code.length == 0) {
            (bool success,) = create2Factory.call(factoryCalldata);
            require(success, ERC6492DeploymentFailed(create2Factory, factoryCalldata));
            require(expectedSigner.code.length != 0, ERC6492NoCode(expectedSigner));
        }

        return IERC1271(expectedSigner).isValidSignature(msgHash, originalSig) == ERC1271_MAGIC_VALUE;
    }
}
