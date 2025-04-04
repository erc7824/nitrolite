// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Test} from "lib/forge-std/src/Test.sol";
import {Vm} from "lib/forge-std/src/Vm.sol";

import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";

import {TestUtils} from "../TestUtils.sol";
import {MockERC20} from "../mocks/MockERC20.sol";

import {IAdjudicator} from "../../src/interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature} from "../../src/interfaces/Types.sol";
import {Consensus} from "../../src/adjudicators/Consensus.sol";
import {Utils} from "../../src/Utils.sol";

contract ConsensusTest is Test {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;

    Consensus public adjudicator;

    // Test accounts
    address public host;
    address public guest;
    uint256 public hostPrivateKey;
    uint256 public guestPrivateKey;

    // Channel parameters
    Channel public channel;
    MockERC20 public token;

    function setUp() public {
        // Deploy the adjudicator contract
        adjudicator = new Consensus();

        // Generate private keys and addresses for the participants
        hostPrivateKey = 0x1;
        guestPrivateKey = 0x2;
        host = vm.addr(hostPrivateKey);
        guest = vm.addr(guestPrivateKey);

        // Deploy the mock token
        token = new MockERC20("Test Token", "TEST", 18);

        // Set up the channel
        address[2] memory participants = [host, guest];
        channel = Channel({
            participants: participants,
            adjudicator: address(adjudicator),
            challenge: 3600, // 1 hour challenge period
            nonce: 1
        });
    }

    // Test case: Basic signing test to verify our signature approach
    // using foundry's vm.sign and correctly formatted message for ecrecover
    function test_BasicSignatureVerification() public view {
        // Simple message to sign
        bytes32 message = keccak256("test message");

        // Sign with the prefixed hash as required by Ethereum
        (uint8 v, bytes32 r, bytes32 s) = TestUtils.sign(vm, hostPrivateKey, message);

        // Directly recover using ecrecover (without prefix since we already signed prefixed hash)
        address recovered = message.toEthSignedMessageHash().recover(v, r, s);

        // This should pass - verifying our signing approach
        assertEq(recovered, host);

        // Now test with Utils.verifySignature
        Signature memory signature = Signature({v: v, r: r, s: s});
        bool isValid = Utils.verifySignature(message, signature, host);

        // This should also pass as Utils.verifySignature adds the prefix
        assertTrue(isValid);
    }

    // Helper function to create test allocations
    function createAllocations(uint256 hostAmount, uint256 guestAmount) internal view returns (Allocation[2] memory) {
        Allocation[2] memory allocations;

        allocations[0] = Allocation({destination: host, token: address(token), amount: hostAmount});

        allocations[1] = Allocation({destination: guest, token: address(token), amount: guestAmount});

        return allocations;
    }

    // Create an initial state signed by host only
    function test_DirectInitialState() public view {
        // Create allocations
        Allocation[2] memory allocations = createAllocations(50, 50);

        // Create the app data
        Consensus.AppData memory appData;
        appData.appData = "initial state";
        appData.status = Consensus.AppStatus.Starting;

        // Create the state
        State memory state;
        state.data = abi.encode(appData);
        state.allocations = allocations;
        state.sigs = new Signature[](1); // Create a dynamic array with 1 element

        // Calculate the state hash
        bytes32 stateHash = Utils.getStateHash(channel, state);

        // Sign the Ethereum signed message hash with host key
        (uint8 v, bytes32 r, bytes32 s) = TestUtils.sign(vm, hostPrivateKey, stateHash);

        // Create the signature
        state.sigs[0] = Signature({v: v, r: r, s: s});

        // Verify the signature directly with Utils
        bool hostSigValid = Utils.verifySignature(stateHash, state.sigs[0], host);

        // This should pass if our approach is correct
        assertTrue(hostSigValid, "Host signature verification failed in Utils");

        // Adjudicate using the contract
        bool isValid = adjudicator.adjudicate(channel, state, new State[](0));

        // Check the state is valid
        assertTrue(isValid, "State should be valid");
    }

    // Test corrupt host signature
    function test_CorruptHostSignature() public {
        // Create allocations
        Allocation[2] memory allocations = createAllocations(50, 50);

        // Create the app data
        Consensus.AppData memory appData;
        appData.appData = "initial state";
        appData.status = Consensus.AppStatus.Starting;

        // Create the state
        State memory state;
        state.data = abi.encode(appData);
        state.allocations = allocations;
        state.sigs = new Signature[](1); // Create a dynamic array with 1 element

        // Calculate the state hash
        bytes32 stateHash = Utils.getStateHash(channel, state);

        // Sign the state hash with host key
        (uint8 v, bytes32 r, bytes32 s) = TestUtils.sign(vm, hostPrivateKey, stateHash);

        // Corrupt the signature by changing r
        r = bytes32(0);

        // Create the signature
        state.sigs[0] = Signature({v: v, r: r, s: s});

        // Adjudicate and expect invalid result
        vm.expectRevert(ECDSA.ECDSAInvalidSignature.selector);
        adjudicator.adjudicate(channel, state, new State[](0));
    }

    // Test corrupt guest signature
    function test_CorruptGuestSignature() public {
        // Create allocations
        Allocation[2] memory allocations = createAllocations(50, 50);

        // Create the app data
        Consensus.AppData memory appData;
        appData.appData = "ready state";
        appData.status = Consensus.AppStatus.Ready;

        // Create the state
        State memory state;
        state.data = abi.encode(appData);
        state.allocations = allocations;
        state.sigs = new Signature[](2); // Create a dynamic array with 2 elements

        // Calculate the state hash
        bytes32 stateHash = Utils.getStateHash(channel, state);

        // Sign the state hash with both keys
        (uint8 v, bytes32 r, bytes32 s) = TestUtils.sign(vm, hostPrivateKey, stateHash);
        state.sigs[0] = Signature({v: v, r: r, s: s});

        // Corrupt the guest signature
        (v, r, s) = vm.sign(guestPrivateKey, stateHash);
        r = bytes32(0); // Corrupt r
        state.sigs[1] = Signature({v: v, r: r, s: s});

        // Adjudicate and expect invalid result
        vm.expectRevert(ECDSA.ECDSAInvalidSignature.selector);
        bool isValid = adjudicator.adjudicate(channel, state, new State[](0));
        assertFalse(isValid, "State should be invalid");
    }

    // Test insufficient signatures
    function test_InsufficientSignatures() public view {
        // Create allocations
        Allocation[2] memory allocations = createAllocations(50, 50);

        // Create the app data
        Consensus.AppData memory appData;
        appData.appData = "initial state";
        appData.status = Consensus.AppStatus.Starting;

        // Create the state with empty signatures
        State memory state;
        state.data = abi.encode(appData);
        state.allocations = allocations;
        // State.sigs defaults to empty values

        // Adjudicate and expect invalid result
        bool isValid = adjudicator.adjudicate(channel, state, new State[](0));
        assertFalse(isValid, "State should be invalid");
    }
}
