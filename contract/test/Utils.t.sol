// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {Test} from "lib/forge-std/src/Test.sol";

import {TestUtils} from "./TestUtils.sol";
import {MockEIP712} from "./mocks/MockEIP712.sol";
import {MockERC20} from "./mocks/MockERC20.sol";
import {MockFlagERC1271} from "./mocks/MockFlagERC1271.sol";
import {MockERC4337Factory} from "./mocks/MockERC4337Factory.sol";
import {Utils} from "../src/Utils.sol";
import {TestUtilsContract} from "./TestUtilsContract.sol";
import {Channel, State, Allocation, StateIntent, STATE_TYPEHASH} from "../src/interfaces/Types.sol";

contract UtilsTest_Signatures is Test {
    MockEIP712 public mockEIP712;
    MockERC20 public token;
    MockERC4337Factory public factory;
    TestUtilsContract public testUtils;

    address public signer;
    uint256 public signerPrivateKey;
    address public wrongSigner;
    uint256 public wrongSignerPrivateKey;

    Channel public channel;
    State public testState;

    function setUp() public {
        mockEIP712 = new MockEIP712("TestDomain", "1.0");
        token = new MockERC20("Test Token", "TEST", 18);
        factory = new MockERC4337Factory();
        testUtils = new TestUtilsContract();

        signerPrivateKey = vm.createWallet("signer").privateKey;
        wrongSignerPrivateKey = vm.createWallet("wrongSigner").privateKey;
        signer = vm.addr(signerPrivateKey);
        wrongSigner = vm.addr(wrongSignerPrivateKey);

        // Create test channel
        address[] memory participants = new address[](2);
        participants[0] = signer;
        participants[1] = wrongSigner;

        channel = Channel({participants: participants, adjudicator: address(0x123), challenge: 3600, nonce: 1});

        // Create test state
        Allocation[] memory allocations = new Allocation[](2);
        allocations[0] = Allocation({destination: signer, token: address(token), amount: 100});
        allocations[1] = Allocation({destination: wrongSigner, token: address(token), amount: 200});

        testState = State({
            intent: StateIntent.INITIALIZE,
            version: 0,
            data: bytes("test data"),
            allocations: allocations,
            sigs: new bytes[](0)
        });
    }

    // ==================== recoverRawECDSASigner TESTS ====================

    function test_recoverRawECDSASigner_returnsCorrectSigner() public view {
        bytes32 messageHash = keccak256("test message");
        bytes memory sig = TestUtils.sign(vm, signerPrivateKey, messageHash);

        address recoveredSigner = testUtils.recoverRawECDSASigner(messageHash, sig);

        assertEq(recoveredSigner, signer, "Should recover correct signer");
    }

    function test_recoverRawECDSASigner_returnsWrongSignerForDifferentMessage() public view {
        bytes32 messageHash = keccak256("test message");
        bytes32 differentMessageHash = keccak256("different message");
        bytes memory sig = TestUtils.sign(vm, signerPrivateKey, messageHash);

        address recoveredSigner = testUtils.recoverRawECDSASigner(differentMessageHash, sig);

        assertNotEq(recoveredSigner, signer, "Should not recover correct signer for different message");
    }

    // ==================== recoverEIP191Signer TESTS ====================

    function test_recoverEIP191Signer_returnsCorrectSigner() public view {
        bytes32 messageHash = keccak256("test message");
        bytes memory sig = TestUtils.signEIP191(vm, signerPrivateKey, messageHash);

        address recoveredSigner = testUtils.recoverEIP191Signer(messageHash, sig);

        assertEq(recoveredSigner, signer, "Should recover correct signer with EIP191");
    }

    function test_recoverEIP191Signer_returnsWrongSigner_forDifferentMessage() public view {
        bytes32 messageHash = keccak256("test message");
        bytes32 differentMessageHash = keccak256("different message");
        bytes memory sig = TestUtils.signEIP191(vm, signerPrivateKey, messageHash);

        address recoveredSigner = testUtils.recoverEIP191Signer(differentMessageHash, sig);

        assertNotEq(recoveredSigner, signer, "Should not recover correct signer for different message");
    }

    function test_recoverEIP191Signer_returnsWrongSigner_forRawSignature() public view {
        bytes32 messageHash = keccak256("test message");
        // Sign with raw ECDSA instead of EIP191
        bytes memory sig = TestUtils.sign(vm, signerPrivateKey, messageHash);

        address recoveredSigner = testUtils.recoverEIP191Signer(messageHash, sig);

        assertNotEq(recoveredSigner, signer, "Should not recover correct signer when using raw signature for EIP191");
    }

    // ==================== recoverEIP712Signer TESTS ====================

    function test_recoverEIP712Signer_returnsCorrectSigner() public view {
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 structHash = keccak256("test struct");
        bytes memory sig = TestUtils.signEIP712(vm, signerPrivateKey, domainSeparator, structHash);

        address recoveredSigner = testUtils.recoverEIP712Signer(domainSeparator, structHash, sig);

        assertEq(recoveredSigner, signer, "Should recover correct signer with EIP712");
    }

    function test_recoverEIP712Signer_returnsWrongSignerForDifferentDomain() public view {
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 differentDomainSeparator = keccak256("different domain");
        bytes32 structHash = keccak256("test struct");
        bytes memory sig = TestUtils.signEIP712(vm, signerPrivateKey, domainSeparator, structHash);

        address recoveredSigner = testUtils.recoverEIP712Signer(differentDomainSeparator, structHash, sig);

        assertNotEq(recoveredSigner, signer, "Should not recover correct signer for different domain");
    }

    // ==================== recoverStateEIP712Signer TESTS ====================

    function test_recoverStateEIP712Signer_returnsCorrectSigner() public view {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 structHash = keccak256(
            abi.encode(
                STATE_TYPEHASH,
                channelId,
                testState.intent,
                testState.version,
                keccak256(testState.data),
                keccak256(abi.encode(testState.allocations))
            )
        );
        bytes memory sig = TestUtils.signEIP712(vm, signerPrivateKey, domainSeparator, structHash);

        address recoveredSigner =
            testUtils.recoverStateEIP712Signer(STATE_TYPEHASH, channelId, domainSeparator, testState, sig);

        assertEq(recoveredSigner, signer, "Should recover correct signer for state EIP712");
    }

    function test_recoverStateEIP712Signer_returnsWrongSignerForDifferentState() public view {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();

        // Sign original state
        bytes32 structHash = keccak256(
            abi.encode(
                STATE_TYPEHASH,
                channelId,
                testState.intent,
                testState.version,
                keccak256(testState.data),
                keccak256(abi.encode(testState.allocations))
            )
        );
        bytes memory sig = TestUtils.signEIP712(vm, signerPrivateKey, domainSeparator, structHash);

        // Create different state
        State memory differentState = testState;
        differentState.data = bytes("different data");

        address recoveredSigner =
            testUtils.recoverStateEIP712Signer(STATE_TYPEHASH, channelId, domainSeparator, differentState, sig);

        assertNotEq(recoveredSigner, signer, "Should not recover correct signer for different state");
    }

    // ==================== verifyStateEOASignature TESTS ====================

    function test_verifyStateEOASignature_returnsTrue_forRawECDSASignature() public view {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 stateHash = testUtils.getStateHashShort(channelId, testState);
        bytes memory sig = TestUtils.sign(vm, signerPrivateKey, stateHash);

        bool isValid = testUtils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, signer);

        assertTrue(isValid, "Should verify raw ECDSA signature");
    }

    function test_verifyStateEOASignature_returnsTrue_forEIP191Signature() public view {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 stateHash = testUtils.getStateHashShort(channelId, testState);
        bytes memory sig = TestUtils.signEIP191(vm, signerPrivateKey, stateHash);

        bool isValid = testUtils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, signer);

        assertTrue(isValid, "Should verify EIP191 signature");
    }

    function test_verifyStateEOASignature_returnsTrue_forEIP712Signature() public view {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 structHash = keccak256(
            abi.encode(
                STATE_TYPEHASH,
                channelId,
                testState.intent,
                testState.version,
                keccak256(testState.data),
                keccak256(abi.encode(testState.allocations))
            )
        );
        bytes memory sig = TestUtils.signEIP712(vm, signerPrivateKey, domainSeparator, structHash);

        bool isValid = testUtils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, signer);

        assertTrue(isValid, "Should verify EIP712 signature");
    }

    function test_verifyStateEOASignature_returnsFalse_forWrongSigner_rawECDSA() public view {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 stateHash = testUtils.getStateHashShort(channelId, testState);
        bytes memory sig = TestUtils.sign(vm, signerPrivateKey, stateHash);

        bool isValid = testUtils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, wrongSigner);

        assertFalse(isValid, "Should not verify signature for wrong signer");
    }

    function test_verifyStateEOASignature_returnsFalse_forWrongSigner_EIP191() public view {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 stateHash = testUtils.getStateHashShort(channelId, testState);
        bytes memory sig = TestUtils.signEIP191(vm, signerPrivateKey, stateHash);

        bool isValid = testUtils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, wrongSigner);

        assertFalse(isValid, "Should not verify EIP191 signature for wrong signer");
    }

    function test_verifyStateEOASignature_returnsFalse_forWrongSigner_EIP712() public view {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 structHash = keccak256(
            abi.encode(
                STATE_TYPEHASH,
                channelId,
                testState.intent,
                testState.version,
                keccak256(testState.data),
                keccak256(abi.encode(testState.allocations))
            )
        );
        bytes memory sig = TestUtils.signEIP712(vm, signerPrivateKey, domainSeparator, structHash);

        bool isValid = testUtils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, wrongSigner);

        assertFalse(isValid, "Should not verify EIP712 signature for wrong signer");
    }

    function test_verifyStateEOASignature_returnsFalse_whenNoEIP712Support() public view {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = testUtils.NO_EIP712_SUPPORT();
        bytes32 structHash = keccak256(
            abi.encode(
                STATE_TYPEHASH,
                channelId,
                testState.intent,
                testState.version,
                keccak256(testState.data),
                keccak256(abi.encode(testState.allocations))
            )
        );
        bytes memory sig = TestUtils.signEIP712(vm, signerPrivateKey, mockEIP712.domainSeparator(), structHash);

        bool isValid = testUtils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, signer);

        assertFalse(isValid, "Should not verify EIP712 signature when NO_EIP712_SUPPORT is set");
    }

    function test_verifyStateEOASignature_returnsTrue_forRawECDSAWhenNoEIP712Support() public view {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = testUtils.NO_EIP712_SUPPORT();
        bytes32 stateHash = testUtils.getStateHashShort(channelId, testState);
        bytes memory sig = TestUtils.sign(vm, signerPrivateKey, stateHash);

        bool isValid = testUtils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, signer);

        assertTrue(isValid, "Should verify raw ECDSA signature even when NO_EIP712_SUPPORT is set");
    }

    function test_verifyStateEOASignature_returnsFalse_forEIP712WhenNoEIP712Support() public view {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 structHash = keccak256(
            abi.encode(
                STATE_TYPEHASH,
                channelId,
                testState.intent,
                testState.version,
                keccak256(testState.data),
                keccak256(abi.encode(testState.allocations))
            )
        );
        bytes memory sig = TestUtils.signEIP712(vm, signerPrivateKey, domainSeparator, structHash);

        bool isValid = testUtils.verifyStateEOASignature(testState, channelId, testUtils.NO_EIP712_SUPPORT(), sig, signer);

        assertFalse(isValid, "Should not verify EIP712 signature when NO_EIP712_SUPPORT is set");
    }

    // ==================== isValidERC1271Signature TESTS ====================

    function test_isValidERC1271Signature_returnsTrue_forValidSignature() public {
        MockFlagERC1271 mockContract = new MockFlagERC1271(true);
        bytes32 msgHash = keccak256("test message");
        bytes memory sig = "dummy signature";

        bool isValid = testUtils.isValidERC1271Signature(msgHash, sig, address(mockContract));

        assertTrue(isValid, "Should return true when flag is true");
    }

    function test_isValidERC1271Signature_returnsFalse_forInvalidSignature() public {
        MockFlagERC1271 mockContract = new MockFlagERC1271(false);
        bytes32 msgHash = keccak256("test message");
        bytes memory sig = "dummy signature";

        bool isValid = testUtils.isValidERC1271Signature(msgHash, sig, address(mockContract));

        assertFalse(isValid, "Should return false when flag is false");
    }

    // ==================== isValidERC6492Signature TESTS ====================

    function getERC6492SignatureAndSigner(
        bool flag,
        bytes32 salt,
        bytes memory originalSig
    ) internal view returns (bytes memory, address) {
        address expectedSigner = factory.getAddress(flag, salt);
        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockERC4337Factory.createAccount.selector,
            flag,
            salt
        );

        bytes memory erc6492Sig = abi.encode(address(factory), factoryCalldata, originalSig);
        bytes memory sigWithSuffix = abi.encodePacked(erc6492Sig, testUtils.ERC6492_DETECTION_SUFFIX());

        return (sigWithSuffix, expectedSigner);
    }

    function test_isValidERC6492Signature_returnsTrue_forValidSignature_notDeployed() public {
        bytes32 msgHash = keccak256("test message");
        (bytes memory signature, address expectedSigner) =
            getERC6492SignatureAndSigner(true, keccak256("test salt"), "dummy signature");

        bool isValid = testUtils.isValidERC6492Signature(msgHash, signature, expectedSigner);

        assertTrue(isValid, "Should return true for valid signature with undeployed contract");
    }

    function test_isValidERC6492Signature_returnsTrue_forValidSignature_deployed() public {
        bytes32 salt = keccak256("test salt");
        bool flag = true;

        address expectedSigner = factory.createAccount(flag, salt);
        bytes32 msgHash = keccak256("test message");
        (bytes memory signature, ) =
            getERC6492SignatureAndSigner(flag, salt, "dummy signature");

        bool isValid = testUtils.isValidERC6492Signature(msgHash, signature, expectedSigner);

        assertTrue(isValid, "Should return true for valid signature with deployed contract");
    }

    function test_isValidERC6492Signature_returnsFalse_forInvalidSignature_notDeployed() public {
        bytes32 msgHash = keccak256("test message");
        (bytes memory signature, address expectedSigner) =
            getERC6492SignatureAndSigner(false, keccak256("test salt"), "dummy signature");

        bool isValid = testUtils.isValidERC6492Signature(msgHash, signature, expectedSigner);

        assertFalse(isValid, "Should return false for invalid signature with undeployed contract");
    }

    function test_isValidERC6492Signature_reverts_forValidSignature_errorOnDeployment() public {
        address expectedSigner = address(0x12345678); // Simulate a contract that will not deploy correctly
        bytes32 msgHash = keccak256("test message");
        bytes memory originalSig = "dummy signature";

        bytes memory factoryCalldata = abi.encodeWithSelector(
            MockERC4337Factory.createAccount.selector,
            "corrupted data",
            keccak256("test salt")
        );

        bytes memory erc6492Sig = abi.encode(address(factory), factoryCalldata, originalSig);
        bytes memory sigWithSuffix = abi.encodePacked(erc6492Sig, testUtils.ERC6492_DETECTION_SUFFIX());

        vm.expectRevert(abi.encodeWithSelector(
            Utils.ERC6492DeploymentFailed.selector,
            factory,
            factoryCalldata
        ));
        testUtils.isValidERC6492Signature(msgHash, sigWithSuffix, expectedSigner);
    }

    function test_isValidERC6492Signature_reverts_forWrongExpectedSigner_thatIsNotContract() public {
        address wrongExpectedSigner = address(0x33231234); // Wrong address
        bytes32 msgHash = keccak256("test message");

        (bytes memory signature, ) = getERC6492SignatureAndSigner(true, keccak256("test salt"), "dummy signature");


        vm.expectRevert(abi.encodeWithSelector(
            Utils.ERC6492NoCode.selector,
            wrongExpectedSigner
        ));
        testUtils.isValidERC6492Signature(msgHash, signature, wrongExpectedSigner);
    }

    function test_isValidERC6492Signature_returnsFalse_forWrongExpectedSigner_thatIsContract() public {
        address wrongExpectedSigner = address(new MockFlagERC1271(false)); // Different deployed contract
        bytes32 msgHash = keccak256("test message");

        (bytes memory signature, ) =
            getERC6492SignatureAndSigner(true, keccak256("test salt"), "dummy signature");

        bool isValid = testUtils.isValidERC6492Signature(msgHash, signature, wrongExpectedSigner);

        assertFalse(isValid, "Should return false when expectedSigner is a different deployed contract");
    }

    // ==================== verifyStateSignature TESTS ====================

    // EOA Signature Tests
    function test_verifyStateSignature_returnsTrue_forEOA_rawECDSASignature() public {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 stateHash = testUtils.getStateHashShort(channelId, testState);
        bytes memory sig = TestUtils.sign(vm, signerPrivateKey, stateHash);

        bool isValid = testUtils.verifyStateSignature(testState, channelId, domainSeparator, sig, signer);

        assertTrue(isValid, "Should verify raw ECDSA signature for EOA");
    }

    function test_verifyStateSignature_returnsTrue_forEOA_EIP191Signature() public {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 stateHash = testUtils.getStateHashShort(channelId, testState);
        bytes memory sig = TestUtils.signEIP191(vm, signerPrivateKey, stateHash);

        bool isValid = testUtils.verifyStateSignature(testState, channelId, domainSeparator, sig, signer);

        assertTrue(isValid, "Should verify EIP191 signature for EOA");
    }

    function test_verifyStateSignature_returnsTrue_forEOA_EIP712Signature() public {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 structHash = keccak256(
            abi.encode(
                STATE_TYPEHASH,
                channelId,
                testState.intent,
                testState.version,
                keccak256(testState.data),
                keccak256(abi.encode(testState.allocations))
            )
        );
        bytes memory sig = TestUtils.signEIP712(vm, signerPrivateKey, domainSeparator, structHash);

        bool isValid = testUtils.verifyStateSignature(testState, channelId, domainSeparator, sig, signer);

        assertTrue(isValid, "Should verify EIP712 signature for EOA");
    }

    function test_verifyStateSignature_returnsFalse_forEOA_wrongSigner() public {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 stateHash = testUtils.getStateHashShort(channelId, testState);
        bytes memory sig = TestUtils.sign(vm, signerPrivateKey, stateHash);

        bool isValid = testUtils.verifyStateSignature(testState, channelId, domainSeparator, sig, wrongSigner);

        assertFalse(isValid, "Should not verify signature for wrong EOA signer");
    }

    // ERC-1271 Contract Signature Tests
    function test_verifyStateSignature_returnsTrue_forERC1271Contract_validSignature() public {
        MockFlagERC1271 mockContract = new MockFlagERC1271(true);
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes memory sig = "dummy signature";

        bool isValid = testUtils.verifyStateSignature(testState, channelId, domainSeparator, sig, address(mockContract));

        assertTrue(isValid, "Should verify valid ERC1271 contract signature");
    }

    function test_verifyStateSignature_returnsFalse_forERC1271Contract_invalidSignature() public {
        MockFlagERC1271 mockContract = new MockFlagERC1271(false);
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes memory sig = "dummy signature";

        bool isValid = testUtils.verifyStateSignature(testState, channelId, domainSeparator, sig, address(mockContract));

        assertFalse(isValid, "Should not verify invalid ERC1271 contract signature");
    }

    // ERC-6492 Signature Tests
    function test_verifyStateSignature_returnsTrue_forERC6492_validSignature_notDeployed() public {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        
        (bytes memory signature, address expectedSigner) =
            getERC6492SignatureAndSigner(true, keccak256("test salt"), "dummy signature");

        bool isValid = testUtils.verifyStateSignature(testState, channelId, domainSeparator, signature, expectedSigner);

        assertTrue(isValid, "Should verify valid ERC6492 signature for undeployed contract");
    }

    function test_verifyStateSignature_returnsTrue_forERC6492_validSignature_deployed() public {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 salt = keccak256("test salt");
        bool flag = true;
        
        address expectedSigner = factory.createAccount(flag, salt);
        (bytes memory signature, ) =
            getERC6492SignatureAndSigner(flag, salt, "dummy signature");

        bool isValid = testUtils.verifyStateSignature(testState, channelId, domainSeparator, signature, expectedSigner);

        assertTrue(isValid, "Should verify valid ERC6492 signature for deployed contract");
    }

    function test_verifyStateSignature_returnsFalse_forERC6492_invalidSignature_notDeployed() public {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        
        (bytes memory signature, address expectedSigner) =
            getERC6492SignatureAndSigner(false, keccak256("test salt"), "dummy signature");

        bool isValid = testUtils.verifyStateSignature(testState, channelId, domainSeparator, signature, expectedSigner);

        assertFalse(isValid, "Should not verify invalid ERC6492 signature for undeployed contract");
    }

    function test_verifyStateSignature_returnsFalse_forERC6492_invalidSignature_deployed() public {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 salt = keccak256("test salt");
        bool flag = false;
        
        address expectedSigner = factory.createAccount(flag, salt);
        (bytes memory signature, ) =
            getERC6492SignatureAndSigner(flag, salt, "dummy signature");

        bool isValid = testUtils.verifyStateSignature(testState, channelId, domainSeparator, signature, expectedSigner);

        assertFalse(isValid, "Should not verify invalid ERC6492 signature for deployed contract");
    }

    function test_verifyStateSignature_returnsFalse_forERC6492_wrongExpectedSigner() public {
        bytes32 channelId = testUtils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        address wrongExpectedSigner = address(new MockFlagERC1271(false));
        
        (bytes memory signature, ) =
            getERC6492SignatureAndSigner(true, keccak256("test salt"), "dummy signature");

        bool isValid = testUtils.verifyStateSignature(testState, channelId, domainSeparator, signature, wrongExpectedSigner);

        assertFalse(isValid, "Should not verify ERC6492 signature for wrong expected signer");
    }
}
