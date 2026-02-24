// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";

import {TestUtils} from "../TestUtils.sol";
import {MockSmartWallet, MockSmartWalletFactory, SwTestUtils} from "./SwTestUtils.sol";

import {SmartWalletValidator} from "../../src/sigValidators/SmartWalletValidator.sol";
import {ValidationResult, VALIDATION_SUCCESS, VALIDATION_FAILURE} from "../../src/interfaces/ISignatureValidator.sol";
import {Utils} from "../../src/Utils.sol";

contract SmartWalletValidatorTest_Base is Test {
    SmartWalletValidator public validator;

    uint256 constant USER_PK = 1;
    uint256 constant OTHER_SIGNER_PK = 2;

    address user;
    address otherSigner;

    MockSmartWallet userSw;
    MockSmartWallet otherSw;
    MockSmartWalletFactory factory;

    bytes32 constant CHANNEL_ID = keccak256("test-channel");
    bytes32 constant OTHER_CHANNEL_ID = keccak256("other-channel");
    bytes constant SIGNING_DATA = hex"1234567890abcdef";

    function setUp() public virtual {
        validator = new SmartWalletValidator();

        user = vm.addr(USER_PK);
        otherSigner = vm.addr(OTHER_SIGNER_PK);

        userSw = new MockSmartWallet(user);
        otherSw = new MockSmartWallet(otherSigner);
        factory = new MockSmartWalletFactory();
    }
}

/**
 * @title SmartWalletValidatorTest_validateSignature_ERC1271
 * @notice Tests for ERC-1271 signature validation (deployed wallets)
 */
contract SmartWalletValidatorTest_validateSignature_ERC1271 is SmartWalletValidatorTest_Base {
    function test_success_withCorrectSignature() public {
        bytes memory message = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signRaw(vm, USER_PK, message);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, address(userSw));
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_failure_withIncorrectSigner() public {
        bytes memory message = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        // use other signer
        bytes memory signature = TestUtils.signRaw(vm, OTHER_SIGNER_PK, message);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, address(userSw));
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withWrongChannelId() public {
        // use wrong channel Id
        bytes memory message = Utils.pack(OTHER_CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signRaw(vm, USER_PK, message);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, address(userSw));
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withFaultySignature() public {
        bytes memory message = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory validSignature = TestUtils.signRaw(vm, USER_PK, message);
        validSignature[0] = 0x42; // Corrupt the signature

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, validSignature, address(userSw));
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withEOA() public {
        bytes memory message = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signRaw(vm, USER_PK, message);

        // Try to validate against EOA address instead of wallet
        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_whenWalletRejectsSignature() public {
        bytes memory message = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signRaw(vm, USER_PK, message);

        // Make the wallet reject all signatures
        userSw.setValidation(false);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, address(userSw));
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }
}

/**
 * @title SmartWalletValidatorTest_validateSignature_ERC6492
 * @notice Tests for ERC-6492 signature validation (non-deployed wallets)
 */
contract SmartWalletValidatorTest_validateSignature_ERC6492 is SmartWalletValidatorTest_Base {
    function test_success_withNonDeployedWallet() public {
        bytes32 salt = keccak256("test_salt");
        address expectedAddress = factory.getAddress(user, salt);

        // Ensure wallet is not deployed
        assertEq(expectedAddress.code.length, 0, "Wallet should not be deployed yet");

        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            user,
            salt
        );

        bytes memory message = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signRaw(vm, USER_PK, message);
        bytes memory erc6492Signature = SwTestUtils.createERC6492Signature(address(factory), factoryCalldata, signature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, erc6492Signature, expectedAddress);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
        assertTrue(expectedAddress.code.length > 0, "Wallet should be deployed after validation");
    }

    function test_success_withAlreadyDeployedWallet() public {
        bytes32 salt = keccak256("test_salt_deployed");

        // Deploy the wallet first
        address deployedAddress = factory.deploy(user, salt);

        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            user,
            salt
        );

        bytes memory message = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signRaw(vm, USER_PK, message);
        bytes memory erc6492Signature = SwTestUtils.createERC6492Signature(address(factory), factoryCalldata, signature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, erc6492Signature, deployedAddress);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_failure_withInvalidSignature() public {
        bytes32 salt = keccak256("test_salt_invalid");
        address expectedAddress = factory.getAddress(user, salt);

        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            user,
            salt
        );

        // Use wrong private key
        bytes memory message = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signRaw(vm, OTHER_SIGNER_PK, message);
        bytes memory erc6492Signature = SwTestUtils.createERC6492Signature(address(factory), factoryCalldata, signature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, erc6492Signature, expectedAddress);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
        assertTrue(expectedAddress.code.length > 0, "Wallet should be deployed even if signature is invalid");
    }

    function test_failure_withWrongChannelId() public {
        bytes32 salt = keccak256("test_salt_wrong_channel");
        address expectedAddress = factory.getAddress(user, salt);

        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            user,
            salt
        );

        // Sign with wrong channel ID
        bytes memory message = Utils.pack(OTHER_CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signRaw(vm, USER_PK, message);
        bytes memory erc6492Signature = SwTestUtils.createERC6492Signature(address(factory), factoryCalldata, signature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, erc6492Signature, expectedAddress);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }
}
