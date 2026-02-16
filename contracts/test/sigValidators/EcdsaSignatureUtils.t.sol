// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "lib/forge-std/src/Test.sol";

import {TestUtils} from "../TestUtils.sol";

import {EcdsaSignatureUtils} from "../../src/sigValidators/EcdsaSignatureUtils.sol";

/**
 * @title EcdsaSignatureUtilsTest_Base
 * @notice Base contract for EcdsaSignatureUtils tests with common setup and utilities
 */
contract EcdsaSignatureUtilsTest_Base is Test {
    uint256 constant SIGNER1_PK = 1;
    uint256 constant SIGNER2_PK = 2;
    uint256 constant OTHER_SIGNER_PK = 3;

    address signer1;
    address signer2;
    address otherSigner;

    bytes constant TEST_MESSAGE = "Test message for signature validation";

    function setUp() public virtual {
        signer1 = vm.addr(SIGNER1_PK);
        signer2 = vm.addr(SIGNER2_PK);
        otherSigner = vm.addr(OTHER_SIGNER_PK);
    }

    function createFaultySignature(bytes memory validSignature) internal pure returns (bytes memory) {
        require(validSignature.length == 65, "Invalid signature length");
        bytes memory faulty = new bytes(65);
        for (uint256 i = 0; i < 65; i++) {
            faulty[i] = validSignature[i];
        }
        // Corrupt the signature by modifying the last byte of the s component
        // This creates a signature that's still 65 bytes but will not recover correctly
        faulty[63] = bytes1(uint8(faulty[63]) ^ 0x01);
        return faulty;
    }
}

/**
 * @title EcdsaSignatureUtilsTest_validateEcdsaSigner
 * @notice Tests for the validateEcdsaSigner function
 */
contract EcdsaSignatureUtilsTest_validateEcdsaSigner is EcdsaSignatureUtilsTest_Base {
    function test_success_withCorrectEip191Sig() public view {
        bytes memory signature = TestUtils.signEip191(vm, SIGNER1_PK, TEST_MESSAGE);

        bool result = EcdsaSignatureUtils.validateEcdsaSigner(TEST_MESSAGE, signature, signer1);
        assertTrue(result, "Should validate correct EIP-191 signature");
    }

    function test_success_withCorrectRawEcdsaSig() public view {
        bytes memory signature = TestUtils.signRaw(vm, SIGNER1_PK, TEST_MESSAGE);

        bool result = EcdsaSignatureUtils.validateEcdsaSigner(TEST_MESSAGE, signature, signer1);
        assertTrue(result, "Should validate correct raw ECDSA signature");
    }

    function test_failure_withFaultyEip191Sig() public view {
        bytes memory validSignature = TestUtils.signEip191(vm, SIGNER1_PK, TEST_MESSAGE);
        bytes memory faultySignature = createFaultySignature(validSignature);

        bool result = EcdsaSignatureUtils.validateEcdsaSigner(TEST_MESSAGE, faultySignature, signer1);
        assertFalse(result, "Should reject faulty EIP-191 signature");
    }

    function test_failure_withFaultyRawEcdsaSig() public view {
        bytes memory validSignature = TestUtils.signRaw(vm, SIGNER1_PK, TEST_MESSAGE);
        bytes memory faultySignature = createFaultySignature(validSignature);

        bool result = EcdsaSignatureUtils.validateEcdsaSigner(TEST_MESSAGE, faultySignature, signer1);
        assertFalse(result, "Should reject faulty raw ECDSA signature");
    }

    function test_failure_withIncorrectEip191Sig() public view {
        bytes memory signature = TestUtils.signEip191(vm, OTHER_SIGNER_PK, TEST_MESSAGE);
        bool result = EcdsaSignatureUtils.validateEcdsaSigner(TEST_MESSAGE, signature, signer1);
        assertFalse(result, "Should reject EIP-191 signature from wrong signer");
    }

    function test_failure_withIncorrectRawEcdsaSig() public view {
        bytes memory signature = TestUtils.signRaw(vm, OTHER_SIGNER_PK, TEST_MESSAGE);
        bool result = EcdsaSignatureUtils.validateEcdsaSigner(TEST_MESSAGE, signature, signer1);
        assertFalse(result, "Should reject raw ECDSA signature from wrong signer");
    }
}
