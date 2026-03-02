// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";

import {TestUtils} from "../TestUtils.sol";
import {MockSmartWallet, MockSmartWalletFactory, SwTestUtils} from "./SwTestUtils.sol";

import {
    SessionKeyValidator,
    SessionKeyAuthorization,
    toSigningData
} from "../../src/sigValidators/SessionKeyValidator.sol";
import {ValidationResult, VALIDATION_SUCCESS, VALIDATION_FAILURE} from "../../src/interfaces/ISignatureValidator.sol";
import {Utils} from "../../src/Utils.sol";

contract SessionKeyValidatorTest_Base is Test {
    SessionKeyValidator public validator;

    uint256 constant USER_PK = 1;
    uint256 constant NODE_PK = 2;
    uint256 constant OTHER_SIGNER_PK = 3;
    uint256 constant SESSION_KEY1_PK = 4;
    uint256 constant SESSION_KEY2_PK = 5;

    address user;
    address node;
    address otherSigner;
    address sessionKey1;
    address sessionKey2;

    bytes32 constant CHANNEL_ID = keccak256("test-channel");
    bytes32 constant OTHER_CHANNEL_ID = keccak256("other-channel");
    bytes constant SIGNING_DATA = hex"1234567890abcdef";
    bytes constant OTHER_SIGNING_DATA = hex"abcdef1234567890";
    bytes32 constant METADATA_HASH = keccak256("metadata");
    bytes32 constant OTHER_METADATA_HASH = keccak256("other-metadata");

    function setUp() public virtual {
        validator = new SessionKeyValidator();

        user = vm.addr(USER_PK);
        node = vm.addr(NODE_PK);
        otherSigner = vm.addr(OTHER_SIGNER_PK);
        sessionKey1 = vm.addr(SESSION_KEY1_PK);
        sessionKey2 = vm.addr(SESSION_KEY2_PK);
    }

    function createSkAuth(address sessionKey, bytes32 metadataHash, uint256 authorizerPk, bool useEip191)
        internal
        pure
        returns (SessionKeyAuthorization memory)
    {
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({sessionKey: sessionKey, metadataHash: metadataHash, authSignature: ""})
        );
        bytes memory authSignature;

        if (useEip191) {
            authSignature = TestUtils.signEip191(vm, authorizerPk, authMessage);
        } else {
            authSignature = TestUtils.signRaw(vm, authorizerPk, authMessage);
        }

        return
            SessionKeyAuthorization({sessionKey: sessionKey, metadataHash: metadataHash, authSignature: authSignature});
    }

    function signStateWithSk(bytes32 channelId, bytes memory signingData, uint256 skPk, bool useEip191)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory stateMessage = Utils.pack(channelId, signingData);

        if (useEip191) {
            return TestUtils.signEip191(vm, skPk, stateMessage);
        } else {
            return TestUtils.signRaw(vm, skPk, stateMessage);
        }
    }

    function signChallengeWithSk(bytes32 channelId, bytes memory signingData, uint256 skPk, bool useEip191)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory challengeMessage = abi.encodePacked(Utils.pack(channelId, signingData), "challenge");

        if (useEip191) {
            return TestUtils.signEip191(vm, skPk, challengeMessage);
        } else {
            return TestUtils.signRaw(vm, skPk, challengeMessage);
        }
    }
}

contract SessionKeyValidatorTest_validateSignature is SessionKeyValidatorTest_Base {
    function test_success_withBothEip191() public {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, true);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_success_withBothRaw() public {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, false);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, false);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_success_withAuthEip191SkSigRaw() public {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, true);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, false);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_success_withAuthRawSkSigEip191() public {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, false);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_failure_withSkAuthNotSignedByParticipant_eip191() public {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, OTHER_SIGNER_PK, true);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withSkAuthNotSignedByParticipant_raw() public {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, OTHER_SIGNER_PK, false);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, false);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withSigningDataNotSignedBySessionKey_eip191() public {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, true);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY2_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withSigningDataNotSignedBySessionKey_raw() public {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, false);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY2_PK, false);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withOtherMetadataHash_eip191() public {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, OTHER_METADATA_HASH, USER_PK, true);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        skAuth.metadataHash = METADATA_HASH;
        signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withOtherMetadataHash_raw() public {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, OTHER_METADATA_HASH, USER_PK, false);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, false);
        bytes memory signature = abi.encode(skAuth, skSignature);

        skAuth.metadataHash = METADATA_HASH;
        signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withOtherSigningData_eip191() public {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, true);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, OTHER_SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_withOtherSigningData_raw() public {
        SessionKeyAuthorization memory skAuth = createSkAuth(sessionKey1, METADATA_HASH, USER_PK, false);
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, OTHER_SIGNING_DATA, SESSION_KEY1_PK, false);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }
}

/**
 * @title SessionKeyValidatorTest_validateSignature_SmartWallet_ERC1271
 * @notice Tests for ERC-1271 signature validation with smart wallets
 */
contract SessionKeyValidatorTest_validateSignature_SmartWallet_ERC1271 is SessionKeyValidatorTest_Base {
    MockSmartWallet userSw;
    MockSmartWallet sessionKeySw;

    function setUp() public override {
        super.setUp();
        userSw = new MockSmartWallet(user);
        sessionKeySw = new MockSmartWallet(sessionKey1);
    }

    function userSwSign(bytes memory message) internal pure returns (bytes memory) {
        return TestUtils.signRaw(vm, USER_PK, message);
    }

    function sessionKeySwSign(bytes memory message) internal pure returns (bytes memory) {
        return TestUtils.signRaw(vm, SESSION_KEY1_PK, message);
    }

    function test_success_participantIsSmartWallet_sessionKeyIsEOA() public {
        // Participant (user) is a smart wallet, session key is EOA
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({sessionKey: sessionKey1, metadataHash: METADATA_HASH, authSignature: ""})
        );
        bytes memory authSignature = userSwSign(authMessage);
        SessionKeyAuthorization memory skAuth =
            SessionKeyAuthorization({sessionKey: sessionKey1, metadataHash: METADATA_HASH, authSignature: authSignature});

        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, address(userSw));
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_success_participantIsEOA_sessionKeyIsSmartWallet() public {
        // Participant is EOA, session key is a smart wallet
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({
                sessionKey: address(sessionKeySw),
                metadataHash: METADATA_HASH,
                authSignature: ""
            })
        );
        bytes memory authSignature = TestUtils.signEip191(vm, USER_PK, authMessage);
        SessionKeyAuthorization memory skAuth = SessionKeyAuthorization({
            sessionKey: address(sessionKeySw),
            metadataHash: METADATA_HASH,
            authSignature: authSignature
        });

        bytes memory stateMessage = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory skSignature = sessionKeySwSign(stateMessage);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_success_bothAreSmartWallets() public {
        // Both participant and session key are smart wallets
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({
                sessionKey: address(sessionKeySw),
                metadataHash: METADATA_HASH,
                authSignature: ""
            })
        );
        bytes memory authSignature = userSwSign(authMessage);
        SessionKeyAuthorization memory skAuth = SessionKeyAuthorization({
            sessionKey: address(sessionKeySw),
            metadataHash: METADATA_HASH,
            authSignature: authSignature
        });

        bytes memory stateMessage = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory skSignature = sessionKeySwSign(stateMessage);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, address(userSw));
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
    }

    function test_failure_participantSmartWallet_wrongAuthSigner() public {
        // Participant is a smart wallet, but auth signature is from wrong signer
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({sessionKey: sessionKey1, metadataHash: METADATA_HASH, authSignature: ""})
        );
        // sign by other signer instead of user
        bytes memory authSignature = TestUtils.signRaw(vm, OTHER_SIGNER_PK, authMessage);
        SessionKeyAuthorization memory skAuth =
            SessionKeyAuthorization({sessionKey: sessionKey1, metadataHash: METADATA_HASH, authSignature: authSignature});

        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, address(userSw));
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_sessionKeySmartWallet_wrongSessionKeySigner() public {
        // Session key is a smart wallet, but state signature is from wrong signer
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({
                sessionKey: address(sessionKeySw),
                metadataHash: METADATA_HASH,
                authSignature: ""
            })
        );
        bytes memory authSignature = TestUtils.signEip191(vm, USER_PK, authMessage);
        SessionKeyAuthorization memory skAuth = SessionKeyAuthorization({
            sessionKey: address(sessionKeySw),
            metadataHash: METADATA_HASH,
            authSignature: authSignature
        });

        bytes memory stateMessage = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        // sign by other signer instead of sessionKey1
        bytes memory skSignature = TestUtils.signRaw(vm, SESSION_KEY2_PK, stateMessage);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }

    function test_failure_participantSmartWallet_walletRejectsSignature() public {
        // Participant wallet is configured to reject all signatures
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({sessionKey: sessionKey1, metadataHash: METADATA_HASH, authSignature: ""})
        );
        bytes memory authSignature = TestUtils.signRaw(vm, USER_PK, authMessage);
        SessionKeyAuthorization memory skAuth =
            SessionKeyAuthorization({sessionKey: sessionKey1, metadataHash: METADATA_HASH, authSignature: authSignature});

        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        // Make wallet reject all signatures
        userSw.setValidation(false);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, address(userSw));
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
    }
}

/**
 * @title SessionKeyValidatorTest_validateSignature_SmartWallet_ERC6492
 * @notice Tests for ERC-6492 signature validation with undeployed smart wallets
 */
contract SessionKeyValidatorTest_validateSignature_SmartWallet_ERC6492 is SessionKeyValidatorTest_Base {
    MockSmartWalletFactory factory;

    function setUp() public override {
        super.setUp();
        factory = new MockSmartWalletFactory();
    }

    function test_success_participantIsNonDeployedSmartWallet_sessionKeyIsEoa() public {
        bytes32 salt = keccak256("user_wallet_salt");
        address userWalletAddress = factory.getAddress(user, salt);

        // Ensure wallet is not deployed
        assertEq(userWalletAddress.code.length, 0, "Wallet should not be deployed yet");

        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            user,
            salt
        );

        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({sessionKey: sessionKey1, metadataHash: METADATA_HASH, authSignature: ""})
        );
        bytes memory authSig = TestUtils.signRaw(vm, USER_PK, authMessage);
        bytes memory authSignature = SwTestUtils.createERC6492Signature(address(factory), factoryCalldata, authSig);

        SessionKeyAuthorization memory skAuth =
            SessionKeyAuthorization({sessionKey: sessionKey1, metadataHash: METADATA_HASH, authSignature: authSignature});

        // Session key signs the state (EOA)
        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, userWalletAddress);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
        assertTrue(userWalletAddress.code.length > 0, "Wallet should be deployed after validation");
    }

    function test_success_participantIsEoa_sessionKeyIsNonDeployedSmartWallet() public {
        bytes32 salt = keccak256("sk_wallet_salt");
        address skWalletAddress = factory.getAddress(sessionKey1, salt);

        // Ensure wallet is not deployed
        assertEq(skWalletAddress.code.length, 0, "Wallet should not be deployed yet");

        // Create factory calldata for deployment
        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            sessionKey1,
            salt
        );

        // Participant (EOA) signs the authorization
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({
                sessionKey: skWalletAddress,
                metadataHash: METADATA_HASH,
                authSignature: ""
            })
        );
        bytes memory authSignature = TestUtils.signEip191(vm, USER_PK, authMessage);
        SessionKeyAuthorization memory skAuth = SessionKeyAuthorization({
            sessionKey: skWalletAddress,
            metadataHash: METADATA_HASH,
            authSignature: authSignature
        });


        bytes memory stateMessage = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory stateSig = TestUtils.signRaw(vm, SESSION_KEY1_PK, stateMessage);
        bytes memory skSignature = SwTestUtils.createERC6492Signature(address(factory), factoryCalldata, stateSig);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
        assertTrue(skWalletAddress.code.length > 0, "Wallet should be deployed after validation");
    }

    function test_success_bothAreNonDeployedSmartWallets() public {
        bytes32 userSalt = keccak256("user_wallet_salt_both");
        bytes32 skSalt = keccak256("sk_wallet_salt_both");
        address userWalletAddress = factory.getAddress(user, userSalt);
        address skWalletAddress = factory.getAddress(sessionKey1, skSalt);

        // Ensure wallets are not deployed
        assertEq(userWalletAddress.code.length, 0, "User wallet should not be deployed yet");
        assertEq(skWalletAddress.code.length, 0, "SK wallet should not be deployed yet");

        // Create factory calldata for both
        bytes memory userFactoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            user,
            userSalt
        );
        bytes memory skFactoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            sessionKey1,
            skSalt
        );

        // Participant smart wallet signs authorization with ERC-6492
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({
                sessionKey: skWalletAddress,
                metadataHash: METADATA_HASH,
                authSignature: ""
            })
        );
        bytes memory authSig = TestUtils.signRaw(vm, USER_PK, authMessage);
        bytes memory authSignature = SwTestUtils.createERC6492Signature(address(factory), userFactoryCalldata, authSig);

        SessionKeyAuthorization memory skAuth = SessionKeyAuthorization({
            sessionKey: skWalletAddress,
            metadataHash: METADATA_HASH,
            authSignature: authSignature
        });

        // Session key smart wallet signs state with ERC-6492
        bytes memory stateMessage = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory stateSig = TestUtils.signRaw(vm, SESSION_KEY1_PK, stateMessage);
        bytes memory skSignature = SwTestUtils.createERC6492Signature(address(factory), skFactoryCalldata, stateSig);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, userWalletAddress);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_SUCCESS));
        assertTrue(userWalletAddress.code.length > 0, "User wallet should be deployed after validation");
        assertTrue(skWalletAddress.code.length > 0, "SK wallet should be deployed after validation");
    }

    function test_failure_nonDeployedParticipantWallet_wrongSigner() public {
        bytes32 salt = keccak256("user_wallet_wrong_signer");
        address userWalletAddress = factory.getAddress(user, salt);

        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            user,
            salt
        );

        // Use wrong signer for auth signature
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({sessionKey: sessionKey1, metadataHash: METADATA_HASH, authSignature: ""})
        );
        bytes memory authSig = TestUtils.signRaw(vm, OTHER_SIGNER_PK, authMessage); // Wrong signer
        bytes memory authSignature = SwTestUtils.createERC6492Signature(address(factory), factoryCalldata, authSig);

        SessionKeyAuthorization memory skAuth =
            SessionKeyAuthorization({sessionKey: sessionKey1, metadataHash: METADATA_HASH, authSignature: authSignature});

        bytes memory skSignature = signStateWithSk(CHANNEL_ID, SIGNING_DATA, SESSION_KEY1_PK, true);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, userWalletAddress);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
        assertTrue(userWalletAddress.code.length > 0, "Wallet should be deployed even with invalid signature");
    }

    function test_failure_nonDeployedSessionKeyWallet_wrongSigner() public {
        bytes32 salt = keccak256("sk_wallet_wrong_signer");
        address skWalletAddress = factory.getAddress(sessionKey1, salt);

        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            sessionKey1,
            salt
        );

        // Use wrong signer for state signature
        bytes memory authMessage = toSigningData(
            SessionKeyAuthorization({
                sessionKey: skWalletAddress,
                metadataHash: METADATA_HASH,
                authSignature: ""
            })
        );
        bytes memory authSignature = TestUtils.signEip191(vm, USER_PK, authMessage);
        SessionKeyAuthorization memory skAuth = SessionKeyAuthorization({
            sessionKey: skWalletAddress,
            metadataHash: METADATA_HASH,
            authSignature: authSignature
        });

        bytes memory stateMessage = Utils.pack(CHANNEL_ID, SIGNING_DATA);
        bytes memory stateSig = TestUtils.signRaw(vm, OTHER_SIGNER_PK, stateMessage); // Wrong signer
        bytes memory skSignature = SwTestUtils.createERC6492Signature(address(factory), factoryCalldata, stateSig);
        bytes memory signature = abi.encode(skAuth, skSignature);

        ValidationResult result = validator.validateSignature(CHANNEL_ID, SIGNING_DATA, signature, user);
        assertEq(ValidationResult.unwrap(result), ValidationResult.unwrap(VALIDATION_FAILURE));
        assertTrue(skWalletAddress.code.length > 0, "SK Wallet should be deployed even with invalid signature");
    }
}
