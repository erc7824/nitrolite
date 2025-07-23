// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {Test} from "lib/forge-std/src/Test.sol";

import {TestUtils} from "./TestUtils.sol";
import {MockEIP712} from "./mocks/MockEIP712.sol";
import {MockERC20} from "./mocks/MockERC20.sol";
import {Utils} from "../src/Utils.sol";
import {Channel, State, Allocation, StateIntent, STATE_TYPEHASH} from "../src/interfaces/Types.sol";

contract UtilsTest_Signatures is Test {
    MockEIP712 public mockEIP712;
    MockERC20 public token;

    address public signer;
    uint256 public signerPrivateKey;
    address public wrongSigner;
    uint256 public wrongSignerPrivateKey;

    Channel public channel;
    State public testState;

    function setUp() public {
        mockEIP712 = new MockEIP712("TestDomain", "1.0");
        token = new MockERC20("Test Token", "TEST", 18);

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

        address recoveredSigner = Utils.recoverRawECDSASigner(messageHash, sig);

        assertEq(recoveredSigner, signer, "Should recover correct signer");
    }

    function test_recoverRawECDSASigner_returnsWrongSignerForDifferentMessage() public view {
        bytes32 messageHash = keccak256("test message");
        bytes32 differentMessageHash = keccak256("different message");
        bytes memory sig = TestUtils.sign(vm, signerPrivateKey, messageHash);

        address recoveredSigner = Utils.recoverRawECDSASigner(differentMessageHash, sig);

        assertNotEq(recoveredSigner, signer, "Should not recover correct signer for different message");
    }

    // ==================== recoverEIP191Signer TESTS ====================

    function test_recoverEIP191Signer_returnsCorrectSigner() public view {
        bytes32 messageHash = keccak256("test message");
        bytes memory sig = TestUtils.signEIP191(vm, signerPrivateKey, messageHash);

        address recoveredSigner = Utils.recoverEIP191Signer(messageHash, sig);

        assertEq(recoveredSigner, signer, "Should recover correct signer with EIP191");
    }

    function test_recoverEIP191Signer_returnsWrongSigner_forDifferentMessage() public view {
        bytes32 messageHash = keccak256("test message");
        bytes32 differentMessageHash = keccak256("different message");
        bytes memory sig = TestUtils.signEIP191(vm, signerPrivateKey, messageHash);

        address recoveredSigner = Utils.recoverEIP191Signer(differentMessageHash, sig);

        assertNotEq(recoveredSigner, signer, "Should not recover correct signer for different message");
    }

    function test_recoverEIP191Signer_returnsWrongSigner_forRawSignature() public view {
        bytes32 messageHash = keccak256("test message");
        // Sign with raw ECDSA instead of EIP191
        bytes memory sig = TestUtils.sign(vm, signerPrivateKey, messageHash);

        address recoveredSigner = Utils.recoverEIP191Signer(messageHash, sig);

        assertNotEq(recoveredSigner, signer, "Should not recover correct signer when using raw signature for EIP191");
    }

    // ==================== recoverEIP712Signer TESTS ====================

    function test_recoverEIP712Signer_returnsCorrectSigner() public view {
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 structHash = keccak256("test struct");
        bytes memory sig = TestUtils.signEIP712(vm, signerPrivateKey, domainSeparator, structHash);

        address recoveredSigner = Utils.recoverEIP712Signer(domainSeparator, structHash, sig);

        assertEq(recoveredSigner, signer, "Should recover correct signer with EIP712");
    }

    function test_recoverEIP712Signer_returnsWrongSignerForDifferentDomain() public view {
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 differentDomainSeparator = keccak256("different domain");
        bytes32 structHash = keccak256("test struct");
        bytes memory sig = TestUtils.signEIP712(vm, signerPrivateKey, domainSeparator, structHash);

        address recoveredSigner = Utils.recoverEIP712Signer(differentDomainSeparator, structHash, sig);

        assertNotEq(recoveredSigner, signer, "Should not recover correct signer for different domain");
    }

    // ==================== recoverStateEIP712Signer TESTS ====================

    function test_recoverStateEIP712Signer_returnsCorrectSigner() public view {
        bytes32 channelId = Utils.getChannelId(channel);
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
            Utils.recoverStateEIP712Signer(STATE_TYPEHASH, channelId, domainSeparator, testState, sig);

        assertEq(recoveredSigner, signer, "Should recover correct signer for state EIP712");
    }

    function test_recoverStateEIP712Signer_returnsWrongSignerForDifferentState() public view {
        bytes32 channelId = Utils.getChannelId(channel);
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
            Utils.recoverStateEIP712Signer(STATE_TYPEHASH, channelId, domainSeparator, differentState, sig);

        assertNotEq(recoveredSigner, signer, "Should not recover correct signer for different state");
    }

    // ==================== verifyStateEOASignature TESTS ====================

    function test_verifyStateEOASignature_returnsTrue_forRawECDSASignature() public view {
        bytes32 channelId = Utils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 stateHash = Utils.getStateHashShort(channelId, testState);
        bytes memory sig = TestUtils.sign(vm, signerPrivateKey, stateHash);

        bool isValid = Utils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, signer);

        assertTrue(isValid, "Should verify raw ECDSA signature");
    }

    function test_verifyStateEOASignature_returnsTrue_forEIP191Signature() public view {
        bytes32 channelId = Utils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 stateHash = Utils.getStateHashShort(channelId, testState);
        bytes memory sig = TestUtils.signEIP191(vm, signerPrivateKey, stateHash);

        bool isValid = Utils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, signer);

        assertTrue(isValid, "Should verify EIP191 signature");
    }

    function test_verifyStateEOASignature_returnsTrue_forEIP712Signature() public view {
        bytes32 channelId = Utils.getChannelId(channel);
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

        bool isValid = Utils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, signer);

        assertTrue(isValid, "Should verify EIP712 signature");
    }

    function test_verifyStateEOASignature_returnsFalse_forWrongSigner_rawECDSA() public view {
        bytes32 channelId = Utils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 stateHash = Utils.getStateHashShort(channelId, testState);
        bytes memory sig = TestUtils.sign(vm, signerPrivateKey, stateHash);

        bool isValid = Utils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, wrongSigner);

        assertFalse(isValid, "Should not verify signature for wrong signer");
    }

    function test_verifyStateEOASignature_returnsFalse_forWrongSigner_EIP191() public view {
        bytes32 channelId = Utils.getChannelId(channel);
        bytes32 domainSeparator = mockEIP712.domainSeparator();
        bytes32 stateHash = Utils.getStateHashShort(channelId, testState);
        bytes memory sig = TestUtils.signEIP191(vm, signerPrivateKey, stateHash);

        bool isValid = Utils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, wrongSigner);

        assertFalse(isValid, "Should not verify EIP191 signature for wrong signer");
    }

    function test_verifyStateEOASignature_returnsFalse_forWrongSigner_EIP712() public view {
        bytes32 channelId = Utils.getChannelId(channel);
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

        bool isValid = Utils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, wrongSigner);

        assertFalse(isValid, "Should not verify EIP712 signature for wrong signer");
    }

    function test_verifyStateEOASignature_returnsFalse_whenNoEIP712Support() public view {
        bytes32 channelId = Utils.getChannelId(channel);
        bytes32 domainSeparator = Utils.NO_EIP712_SUPPORT;
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

        bool isValid = Utils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, signer);

        assertFalse(isValid, "Should not verify EIP712 signature when NO_EIP712_SUPPORT is set");
    }

    function test_verifyStateEOASignature_returnsTrue_forRawECDSAWhenNoEIP712Support() public view {
        bytes32 channelId = Utils.getChannelId(channel);
        bytes32 domainSeparator = Utils.NO_EIP712_SUPPORT;
        bytes32 stateHash = Utils.getStateHashShort(channelId, testState);
        bytes memory sig = TestUtils.sign(vm, signerPrivateKey, stateHash);

        bool isValid = Utils.verifyStateEOASignature(testState, channelId, domainSeparator, sig, signer);

        assertTrue(isValid, "Should verify raw ECDSA signature even when NO_EIP712_SUPPORT is set");
    }

    function test_verifyStateEOASignature_returnsFalse_forEIP712WhenNoEIP712Support() public view {
        bytes32 channelId = Utils.getChannelId(channel);
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

        bool isValid = Utils.verifyStateEOASignature(testState, channelId, Utils.NO_EIP712_SUPPORT, sig, signer);

        assertFalse(isValid, "Should not verify EIP712 signature when NO_EIP712_SUPPORT is set");
    }
}
