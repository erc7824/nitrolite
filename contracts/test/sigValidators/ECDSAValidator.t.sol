// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "lib/forge-std/src/Test.sol";

import {ECDSAValidator} from "../../src/sigValidators/ECDSAValidator.sol";
import {ValidationResult, VALIDATION_SUCCESS, VALIDATION_FAILURE} from "../../src/interfaces/ISignatureValidator.sol";
import {Utils} from "../../src/Utils.sol";
import {TestUtils} from "../TestUtils.sol";

contract ECDSAValidatorTest_Base is Test {
    ECDSAValidator public validator;

    uint256 constant USER_PK = 1;
    uint256 constant NODE_PK = 2;
    uint256 constant OTHER_SIGNER_PK = 3;

    address user;
    address node;
    address otherSigner;

    bytes32 constant CHANNEL_ID = keccak256("test-channel");
    bytes32 constant OTHER_CHANNEL_ID = keccak256("other-channel");
    bytes constant SIGNING_DATA = hex"1234567890abcdef";

    function setUp() public virtual {
        validator = new ECDSAValidator();

        user = vm.addr(USER_PK);
        node = vm.addr(NODE_PK);
        otherSigner = vm.addr(OTHER_SIGNER_PK);
    }
}

contract ECDSAValidatorTest_validateSignature is ECDSAValidatorTest_Base {
    function test_success_withEip191Sig() public view {
        bytes memory message = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signEip191(vm, USER_PK, message);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_success_withRawEcdsaSig() public view {
        bytes memory message = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signRaw(vm, USER_PK, message);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_failure_withEip191SigFromOtherSigner() public view {
        bytes memory message = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signEip191(vm, OTHER_SIGNER_PK, message);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withRawEcdsaSigFromOtherSigner() public view {
        bytes memory message = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signRaw(vm, OTHER_SIGNER_PK, message);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withEip191SigWithOtherChannelId() public view {
        bytes memory message = Utils.pack(OTHER_CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signEip191(vm, USER_PK, message);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withRawEcdsaSigWithOtherChannelId() public view {
        bytes memory message = Utils.pack(OTHER_CHANNEL_ID, SIGNING_DATA);
        bytes memory signature = TestUtils.signRaw(vm, USER_PK, message);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }
}
