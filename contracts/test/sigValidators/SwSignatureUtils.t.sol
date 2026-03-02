// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";

import {TestUtils} from "../TestUtils.sol";
import {MockSmartWallet, MockSmartWalletFactory, FailingFactory, EmptyFactory, SwTestUtils} from "./SwTestUtils.sol";

import {SwSignatureUtils} from "../../src/sigValidators/SwSignatureUtils.sol";

/**
 * @title TestSwSignatureUtils
 * @notice Wrapper contract to test library functions properly
 */
contract TestSwSignatureUtils {
    function validateSmartWalletSigner(bytes memory message, bytes memory signature, address expectedSigner)
        external
        returns (bool)
    {
        return SwSignatureUtils.validateSmartWalletSigner(message, signature, expectedSigner);
    }
}

/**
 * @title SwSignatureUtilsTest_Base
 * @notice Base contract for SwSignatureUtils tests with common setup and utilities
 */
contract SwSignatureUtilsTest_Base is Test {
    uint256 constant OWNER_PK = 1;
    uint256 constant OTHER_PK = 2;

    address owner;
    address otherAccount;

    MockSmartWallet wallet;
    MockSmartWalletFactory factory;
    TestSwSignatureUtils swSigUtils;

    bytes constant TEST_MESSAGE = "Test message for smart wallet signature validation";

    function setUp() public virtual {
        owner = vm.addr(OWNER_PK);
        otherAccount = vm.addr(OTHER_PK);

        wallet = new MockSmartWallet(owner);
        factory = new MockSmartWalletFactory();
        swSigUtils = new TestSwSignatureUtils();
    }
}

/**
 * @title SwSignatureUtilsTest_validateSmartWalletSigner_ERC1271
 * @notice Tests for ERC-1271 signature validation (deployed wallets)
 */
contract SwSignatureUtilsTest_validateSmartWalletSigner_ERC1271 is SwSignatureUtilsTest_Base {
    function test_success_withCorrectSignature() public {
        bytes memory signature = TestUtils.signRaw(vm, OWNER_PK, TEST_MESSAGE);

        bool result = swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, signature, address(wallet));
        assertTrue(result, "Should validate correct ERC-1271 signature");
    }

    function test_failure_withIncorrectSigner() public {
        bytes memory signature = TestUtils.signRaw(vm, OTHER_PK, TEST_MESSAGE);

        bool result = swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, signature, address(wallet));
        assertFalse(result, "Should reject signature from wrong signer");
    }

    function test_failure_withFaultySignature() public {
        bytes memory validSignature = TestUtils.signRaw(vm, OWNER_PK, TEST_MESSAGE);
        validSignature[0] = 0x42; // Corrupt the signature to make it invalid

        bool result = swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, validSignature, address(wallet));
        assertFalse(result, "Should reject faulty signature");
    }

    function test_failure_withWrongMessage() public {
        bytes memory signature = TestUtils.signRaw(vm, OWNER_PK, "Different message");

        bool result = swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, signature, address(wallet));
        assertFalse(result, "Should reject signature for different message");
    }

    function test_failure_withNonERC1271Contract() public {
        // Deploy a contract that doesn't implement ERC-1271
        address nonERC1271 = address(new MockSmartWalletFactory());
        bytes memory signature = TestUtils.signRaw(vm, OWNER_PK, TEST_MESSAGE);

        bool result = swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, signature, nonERC1271);
        assertFalse(result, "Should reject signature from non-ERC1271 contract");
    }

    function test_failure_withEOA() public {
        bytes memory signature = TestUtils.signRaw(vm, OWNER_PK, TEST_MESSAGE);

        bool result = swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, signature, owner);
        assertFalse(result, "Should reject signature when expected signer is EOA");
    }

    function test_failure_whenWalletRejectsSignature() public {
        bytes memory signature = TestUtils.signRaw(vm, OWNER_PK, TEST_MESSAGE);

        // Make the wallet reject all signatures
        wallet.setValidation(false);

        bool result = swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, signature, address(wallet));
        assertFalse(result, "Should reject when wallet returns invalid magic value");
    }
}

/**
 * @title SwSignatureUtilsTest_validateSmartWalletSigner_ERC6492
 * @notice Tests for ERC-6492 signature validation (non-deployed wallets)
 */
contract SwSignatureUtilsTest_validateSmartWalletSigner_ERC6492 is SwSignatureUtilsTest_Base {
    function test_success_withUndeployedWallet() public {
        bytes32 salt = keccak256("test_salt");
        address expectedAddress = factory.getAddress(owner, salt);

        assertEq(expectedAddress.code.length, 0, "Wallet should not be deployed yet");

        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            owner,
            salt
        );

        bytes memory signature = TestUtils.signRaw(vm, OWNER_PK, TEST_MESSAGE);
        bytes memory erc6492Signature = SwTestUtils.createERC6492Signature(address(factory), factoryCalldata, signature);

        bool result = swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, erc6492Signature, expectedAddress);
        assertTrue(result, "Should validate ERC-6492 signature and deploy wallet");
        assertTrue(expectedAddress.code.length > 0, "Wallet should be deployed after validation");
    }

    function test_success_withAlreadyDeployedWallet() public {
        bytes32 salt = keccak256("test_salt_deployed");
        // Deploy the wallet first
        address deployedAddress = factory.deploy(owner, salt);

        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            owner,
            salt
        );

        bytes memory signature = TestUtils.signRaw(vm, OWNER_PK, TEST_MESSAGE);
        bytes memory erc6492Signature = SwTestUtils.createERC6492Signature(address(factory), factoryCalldata, signature);

        bool result = swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, erc6492Signature, deployedAddress);
        assertTrue(result, "Should validate ERC-6492 signature for already deployed wallet");
    }

    function test_failure_withInvalidERC6492Signature_nonDeployed() public {
        bytes32 salt = keccak256("test_salt_invalid");
        address expectedAddress = factory.getAddress(owner, salt);

        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            owner,
            salt
        );

        // Use wrong private key
        bytes memory signature = TestUtils.signRaw(vm, OTHER_PK, TEST_MESSAGE);
        bytes memory erc6492Signature = SwTestUtils.createERC6492Signature(address(factory), factoryCalldata, signature);

        // Should deploy the contract but reject the invalid signature
        bool result = swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, erc6492Signature, expectedAddress);
        assertFalse(result, "Should reject invalid ERC-6492 signature");
        assertTrue(expectedAddress.code.length > 0, "Wallet should be deployed even if signature is invalid");
    }

    function test_failure_withInvalidERC6492Signature_deployed() public {
        bytes32 salt = keccak256("test_salt_wrong_sig");
        // Deploy the wallet first
        address deployedAddress = factory.deploy(owner, salt);

        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockSmartWalletFactory.deploy.selector,
            owner,
            salt
        );

        // Sign with wrong private key
        bytes memory signature = TestUtils.signRaw(vm, OTHER_PK, TEST_MESSAGE);
        bytes memory erc6492Signature = SwTestUtils.createERC6492Signature(address(factory), factoryCalldata, signature);

        // Should fail because signature is from wrong signer
        bool result = swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, erc6492Signature, deployedAddress);
        assertFalse(result, "Should reject ERC-6492 signature with wrong signer for deployed wallet");
    }

    function test_failure_withShortSignature() public {
        // Create a signature shorter than 32 bytes (won't be detected as ERC-6492)
        bytes memory shortSig = new bytes(16);

        bool result = swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, shortSig, address(wallet));
        assertFalse(result, "Should reject signature shorter than 32 bytes");
    }

    function test_revert_ERC6492DeploymentFailed() public {
        FailingFactory failFactory = new FailingFactory();
        bytes32 salt = keccak256("test_salt_deployment_failed");
        address expectedAddress = address(1); // The address that would be deployed (not important since deployment fails)

        // Create calldata for failing factory
        bytes memory failingCalldata = abi.encodeWithSelector(
            FailingFactory.deploy.selector,
            owner,
            salt
        );

        bytes memory signature = TestUtils.signRaw(vm, OWNER_PK, TEST_MESSAGE);
        bytes memory erc6492Signature = SwTestUtils.createERC6492Signature(address(failFactory), failingCalldata, signature);

        vm.expectRevert(
            abi.encodeWithSelector(
                SwSignatureUtils.ERC6492DeploymentFailed.selector,
                address(failFactory),
                failingCalldata
            )
        );
        swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, erc6492Signature, expectedAddress);
    }

    function test_revert_ERC6492NoCode() public {
        // Deploy a special factory that doesn't revert but also doesn't deploy to expected address
        EmptyFactory emptyFactory = new EmptyFactory();
        bytes32 salt = keccak256("test_salt_no_code");

        // Create a counterfactual address that won't have code
        address expectedAddress = address(0x1234567890123456789012345678901234567890);

        bytes memory factoryCalldata = abi.encodeWithSelector(
            EmptyFactory.deploy.selector,
            owner,
            salt
        );

        bytes memory signature = TestUtils.signRaw(vm, OWNER_PK, TEST_MESSAGE);
        bytes memory erc6492Signature = SwTestUtils.createERC6492Signature(address(emptyFactory), factoryCalldata, signature);

        // Should revert with ERC6492NoCode because the expected address doesn't have code after deployment
        vm.expectRevert(
            abi.encodeWithSelector(
                SwSignatureUtils.ERC6492NoCode.selector,
                expectedAddress
            )
        );
        swSigUtils.validateSmartWalletSigner(TEST_MESSAGE, erc6492Signature, expectedAddress);
    }
}
