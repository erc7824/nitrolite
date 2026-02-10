// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "lib/forge-std/src/Test.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";

import {TestUtils} from "../TestUtils.sol";

import {BaseValidator} from "../../src/sigValidators/BaseValidator.sol";

/**
 * @notice Test harness that exposes internal BaseValidator functions for testing
 */
contract TestBaseValidator is BaseValidator {
    function exposed_validateEcdsaSigner(
        bytes memory message,
        bytes memory signature,
        address expectedSigner
    ) external pure returns (bool) {
        return validateEcdsaSigner(message, signature, expectedSigner);
    }

    function exposed_validateEcdsaSignerIsEither(
        bytes memory message,
        bytes memory signature,
        address addr1,
        address addr2
    ) external pure returns (bool) {
        return validateEcdsaSignerIsEither(message, signature, addr1, addr2);
    }
}

/**
 * @title BaseValidatorTest_Base
 * @notice Base contract for BaseValidator tests with common setup and utilities
 */
contract BaseValidatorTest_Base is Test {
    TestBaseValidator public validator;

    uint256 constant SIGNER1_PK = 1;
    uint256 constant SIGNER2_PK = 2;
    uint256 constant OTHER_SIGNER_PK = 3;

    address signer1;
    address signer2;
    address otherSigner;

    bytes constant TEST_MESSAGE = "Test message for signature validation";

    function setUp() public virtual {
        validator = new TestBaseValidator();

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
 * @title BaseValidatorTest_validateEcdsaSigner
 * @notice Tests for the validateEcdsaSigner function
 */
contract BaseValidatorTest_validateEcdsaSigner is BaseValidatorTest_Base {
    function test_success_withCorrectEip191Sig() public view {
        bytes memory signature = TestUtils.signEip191(vm, SIGNER1_PK, TEST_MESSAGE);

        bool result = validator.exposed_validateEcdsaSigner(TEST_MESSAGE, signature, signer1);
        assertTrue(result, "Should validate correct EIP-191 signature");
    }

    function test_success_withCorrectRawEcdsaSig() public view {
        bytes memory signature = TestUtils.signRaw(vm, SIGNER1_PK, TEST_MESSAGE);

        bool result = validator.exposed_validateEcdsaSigner(TEST_MESSAGE, signature, signer1);
        assertTrue(result, "Should validate correct raw ECDSA signature");
    }

    function test_failure_withFaultyEip191Sig() public view {
        bytes memory validSignature = TestUtils.signEip191(vm, SIGNER1_PK, TEST_MESSAGE);
        bytes memory faultySignature = createFaultySignature(validSignature);

        bool result = validator.exposed_validateEcdsaSigner(TEST_MESSAGE, faultySignature, signer1);
        assertFalse(result, "Should reject faulty EIP-191 signature");
    }

    function test_failure_withFaultyRawEcdsaSig() public view {
        bytes memory validSignature = TestUtils.signRaw(vm, SIGNER1_PK, TEST_MESSAGE);
        bytes memory faultySignature = createFaultySignature(validSignature);

        bool result = validator.exposed_validateEcdsaSigner(TEST_MESSAGE, faultySignature, signer1);
        assertFalse(result, "Should reject faulty raw ECDSA signature");
    }

    function test_failure_withIncorrectEip191Sig() public view {
        bytes memory signature = TestUtils.signEip191(vm, OTHER_SIGNER_PK, TEST_MESSAGE);
        bool result = validator.exposed_validateEcdsaSigner(TEST_MESSAGE, signature, signer1);
        assertFalse(result, "Should reject EIP-191 signature from wrong signer");
    }

    function test_failure_withIncorrectRawEcdsaSig() public view {
        bytes memory signature = TestUtils.signRaw(vm, OTHER_SIGNER_PK, TEST_MESSAGE);
        bool result = validator.exposed_validateEcdsaSigner(TEST_MESSAGE, signature, signer1);
        assertFalse(result, "Should reject raw ECDSA signature from wrong signer");
    }
}

/**
 * @title BaseValidatorTest_validateEcdsaSignerIsEither
 * @notice Tests for the validateEcdsaSignerIsEither function
 */
contract BaseValidatorTest_validateEcdsaSignerIsEither is BaseValidatorTest_Base {
    function test_success_withEip191SigFromFirstAddress() public view {
        bytes memory signature = TestUtils.signEip191(vm, SIGNER1_PK, TEST_MESSAGE);
        bool result = validator.exposed_validateEcdsaSignerIsEither(TEST_MESSAGE, signature, signer1, signer2);
        assertTrue(result, "Should validate EIP-191 signature from first address");
    }

    function test_success_withEip191SigFromSecondAddress() public view {
        bytes memory signature = TestUtils.signEip191(vm, SIGNER2_PK, TEST_MESSAGE);
        bool result = validator.exposed_validateEcdsaSignerIsEither(TEST_MESSAGE, signature, signer1, signer2);
        assertTrue(result, "Should validate EIP-191 signature from second address");
    }

    function test_success_withRawEcdsaSigFromFirstAddress() public view {
        bytes memory signature = TestUtils.signRaw(vm, SIGNER1_PK, TEST_MESSAGE);
        bool result = validator.exposed_validateEcdsaSignerIsEither(TEST_MESSAGE, signature, signer1, signer2);
        assertTrue(result, "Should validate raw ECDSA signature from first address");
    }

    function test_success_withRawEcdsaSigFromSecondAddress() public view {
        bytes memory signature = TestUtils.signRaw(vm, SIGNER2_PK, TEST_MESSAGE);
        bool result = validator.exposed_validateEcdsaSignerIsEither(TEST_MESSAGE, signature, signer1, signer2);
        assertTrue(result, "Should validate raw ECDSA signature from second address");
    }

    function test_failure_withEip191SigFromOtherSigner() public view {
        bytes memory signature = TestUtils.signEip191(vm, OTHER_SIGNER_PK, TEST_MESSAGE);

        bool result = validator.exposed_validateEcdsaSignerIsEither(TEST_MESSAGE, signature, signer1, signer2);

        assertFalse(result, "Should reject EIP-191 signature from other signer");
    }

    function test_failure_withRawEcdsaSigFromOtherSigner() public view {
        bytes memory signature = TestUtils.signRaw(vm, OTHER_SIGNER_PK, TEST_MESSAGE);
        bool result = validator.exposed_validateEcdsaSignerIsEither(TEST_MESSAGE, signature, signer1, signer2);
        assertFalse(result, "Should reject raw ECDSA signature from other signer");
    }

    function test_failure_withFaultyEip191SigFromFirstSigner() public view {
        bytes memory validSignature = TestUtils.signEip191(vm, SIGNER1_PK, TEST_MESSAGE);
        bytes memory faultySignature = createFaultySignature(validSignature);

        bool result = validator.exposed_validateEcdsaSignerIsEither(TEST_MESSAGE, faultySignature, signer1, signer2);
        assertFalse(result, "Should reject faulty EIP-191 signature from first signer");
    }

    function test_failure_withFaultyRawEcdsaSigFromSecondSigner() public view {
        bytes memory validSignature = TestUtils.signRaw(vm, SIGNER2_PK, TEST_MESSAGE);
        bytes memory faultySignature = createFaultySignature(validSignature);

        bool result = validator.exposed_validateEcdsaSignerIsEither(TEST_MESSAGE, faultySignature, signer1, signer2);
        assertFalse(result, "Should reject faulty raw ECDSA signature from second signer");
    }
}
