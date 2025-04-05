// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Test} from "lib/forge-std/src/Test.sol";
import {Vm} from "lib/forge-std/src/Vm.sol";

import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";
import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";

import {TestUtils} from "../TestUtils.sol";

import {IAdjudicator} from "../../src/interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature} from "../../src/interfaces/Types.sol";
import {NitroRPC} from "../../src/adjudicators/NitroRPC.sol";
import {Utils} from "../../src/Utils.sol";

contract NitroRPCTest is Test {
    NitroRPC public nitroRpcAdjudicator;

    // Test accounts
    address public client;
    address public server;
    uint256 public clientPrivateKey;
    uint256 public serverPrivateKey;

    // Channel parameters
    Channel public channel;

    // Constants for participant ordering
    uint256 private constant CLIENT_IDX = 0;
    uint256 private constant SERVER_IDX = 1;

    function setUp() public {
        // Deploy the adjudicator
        nitroRpcAdjudicator = new NitroRPC();

        // Set private keys and corresponding addresses
        clientPrivateKey = 0x1;
        serverPrivateKey = 0x2;
        client = vm.addr(clientPrivateKey);
        server = vm.addr(serverPrivateKey);

        // Set up the channel with the two participants
        address[] memory participants = new address[](2);
        participants[CLIENT_IDX] = client;
        participants[SERVER_IDX] = server;
        channel = Channel({
            participants: participants,
            adjudicator: address(nitroRpcAdjudicator),
            challenge: 3600, // e.g., 1-hour challenge period
            nonce: 1
        });
    }

    // Helper function to create an RPCMessage state
    function createRPCState(
        uint64 requestID,
        uint64 timestamp,
        string memory method,
        bytes memory params,
        bytes memory result
    ) internal pure returns (State memory) {
        State memory state;

        // Create RPCMessage and encode it
        NitroRPC.RPCMessage memory rpcMessage = NitroRPC.RPCMessage({
            requestID: requestID,
            timestamp: timestamp,
            method: method,
            params: params,
            result: result
        });

        state.data = abi.encode(rpcMessage);
        state.sigs = new Signature[](0);
        state.allocations = new Allocation[](0);
        return state;
    }

    // Helper to sign a state using a given private key
    function signState(State memory state, uint256 privateKey) internal view returns (Signature memory) {
        bytes32 stateHash = Utils.getStateHash(channel, state);
        (uint8 v, bytes32 r, bytes32 s) = TestUtils.sign(vm, privateKey, stateHash);
        return Signature({v: v, r: r, s: s});
    }

    // Helper to sign an RPC request from client (special NitroRPC method)
    function signRequest(State memory state, uint256 privateKey) internal view returns (Signature memory) {
        bytes32 channelId = Utils.getChannelId(channel);
        NitroRPC.RPCMessage memory message = abi.decode(state.data, (NitroRPC.RPCMessage));

        // Hash allocations separately (as done in the contract)
        bytes32 allocationsHash = keccak256(abi.encode(state.allocations));

        // Create request hash following NitroRPC's getReqStateHash logic
        bytes32 reqStateHash = keccak256(
            abi.encode(channelId, allocationsHash, message.requestID, message.method, message.params, message.timestamp)
        );

        (uint8 v, bytes32 r, bytes32 s) = TestUtils.sign(vm, privateKey, reqStateHash);
        return Signature({v: v, r: r, s: s});
    }

    // Helper to sign an RPC response from server (special NitroRPC method)
    function signResponse(State memory state, uint256 privateKey) internal view returns (Signature memory) {
        bytes32 channelId = Utils.getChannelId(channel);
        NitroRPC.RPCMessage memory message = abi.decode(state.data, (NitroRPC.RPCMessage));

        // Hash allocations separately (as done in the contract)
        bytes32 allocationsHash = keccak256(abi.encode(state.allocations));

        // Create response hash following NitroRPC's getResStateHash logic
        bytes32 resStateHash = keccak256(
            abi.encode(
                channelId,
                allocationsHash,
                message.requestID,
                message.method,
                message.params,
                message.result,
                message.timestamp
            )
        );

        (uint8 v, bytes32 r, bytes32 s) = TestUtils.sign(vm, privateKey, resStateHash);
        return Signature({v: v, r: r, s: s});
    }

    // -------------------- INITIAL STATE TESTS --------------------

    // Valid initial state with two valid signatures (client request and server response)
    function test_ValidRPCState() public view {
        uint64 requestID = 1;
        uint64 timestamp = 1684155125000; // Example timestamp in milliseconds
        string memory method = "echo";
        bytes memory params = abi.encode("Hello");
        bytes memory result = abi.encode("Hello");

        State memory state = createRPCState(requestID, timestamp, method, params, result);
        state.allocations = new Allocation[](1);
        state.allocations[0] = Allocation({destination: client, token: address(0), amount: 1 ether});

        // Add client signature for request and server signature for response
        state.sigs = new Signature[](2);
        state.sigs[CLIENT_IDX] = signRequest(state, clientPrivateKey);
        state.sigs[SERVER_IDX] = signResponse(state, serverPrivateKey);

        // Create a previous state with earlier timestamp
        uint64 prevRequestID = 0;
        uint64 prevTimestamp = timestamp - 1000; // 1 second earlier
        string memory prevMethod = "echo";
        bytes memory prevParams = abi.encode("Hi");
        bytes memory prevResult = abi.encode("Hi");

        State memory prevState = createRPCState(prevRequestID, prevTimestamp, prevMethod, prevParams, prevResult);
        prevState.allocations = new Allocation[](1);
        prevState.allocations[0] = Allocation({destination: client, token: address(0), amount: 1 ether});

        prevState.sigs = new Signature[](2);
        prevState.sigs[CLIENT_IDX] = signRequest(prevState, clientPrivateKey);
        prevState.sigs[SERVER_IDX] = signResponse(prevState, serverPrivateKey);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        bool valid = nitroRpcAdjudicator.adjudicate(channel, state, proofs);
        assertTrue(valid, "Valid RPC state should be accepted");
    }

    // State with insufficient signatures should fail
    function test_InsufficientSignatures() public view {
        uint64 requestID = 1;
        uint64 timestamp = 1684155125000;
        string memory method = "echo";
        bytes memory params = abi.encode("Hello");
        bytes memory result = abi.encode("Hello");

        State memory state = createRPCState(requestID, timestamp, method, params, result);

        // Only one signature provided (client's)
        state.sigs = new Signature[](1);
        state.sigs[CLIENT_IDX] = signRequest(state, clientPrivateKey);

        // Create a previous state
        uint64 prevRequestID = 0;
        uint64 prevTimestamp = timestamp - 1000;
        State memory prevState = createRPCState(prevRequestID, prevTimestamp, method, params, result);
        prevState.sigs = new Signature[](2);
        prevState.sigs[CLIENT_IDX] = signRequest(prevState, clientPrivateKey);
        prevState.sigs[SERVER_IDX] = signResponse(prevState, serverPrivateKey);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        bool valid = nitroRpcAdjudicator.adjudicate(channel, state, proofs);
        assertFalse(valid, "State with insufficient signatures should be rejected");
    }

    // State with invalid client signature should fail
    function test_InvalidClientSignature() public view {
        uint64 requestID = 1;
        uint64 timestamp = 1684155125000;
        string memory method = "echo";
        bytes memory params = abi.encode("Hello");
        bytes memory result = abi.encode("Hello");

        State memory state = createRPCState(requestID, timestamp, method, params, result);

        // Client signature is invalid (signed by server's key)
        state.sigs = new Signature[](2);
        state.sigs[CLIENT_IDX] = signRequest(state, serverPrivateKey); // Wrong signer
        state.sigs[SERVER_IDX] = signResponse(state, serverPrivateKey);

        // Create a previous state
        uint64 prevRequestID = 0;
        uint64 prevTimestamp = timestamp - 1000;
        State memory prevState = createRPCState(prevRequestID, prevTimestamp, method, params, result);
        prevState.sigs = new Signature[](2);
        prevState.sigs[CLIENT_IDX] = signRequest(prevState, clientPrivateKey);
        prevState.sigs[SERVER_IDX] = signResponse(prevState, serverPrivateKey);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        bool valid = nitroRpcAdjudicator.adjudicate(channel, state, proofs);
        assertFalse(valid, "State with invalid client signature should be rejected");
    }

    // State with invalid server signature should fail
    function test_InvalidServerSignature() public view {
        uint64 requestID = 1;
        uint64 timestamp = 1684155125000;
        string memory method = "echo";
        bytes memory params = abi.encode("Hello");
        bytes memory result = abi.encode("Hello");

        State memory state = createRPCState(requestID, timestamp, method, params, result);

        // Server signature is invalid (signed by client's key)
        state.sigs = new Signature[](2);
        state.sigs[CLIENT_IDX] = signRequest(state, clientPrivateKey);
        state.sigs[SERVER_IDX] = signResponse(state, clientPrivateKey); // Wrong signer

        // Create a previous state
        uint64 prevRequestID = 0;
        uint64 prevTimestamp = timestamp - 1000;
        State memory prevState = createRPCState(prevRequestID, prevTimestamp, method, params, result);
        prevState.sigs = new Signature[](2);
        prevState.sigs[CLIENT_IDX] = signRequest(prevState, clientPrivateKey);
        prevState.sigs[SERVER_IDX] = signResponse(prevState, serverPrivateKey);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        bool valid = nitroRpcAdjudicator.adjudicate(channel, state, proofs);
        assertFalse(valid, "State with invalid server signature should be rejected");
    }

    // -------------------- TIMESTAMP TESTS --------------------

    // State with timestamp not greater than previous state should fail
    function test_InvalidTimestamp_NotGreater() public view {
        uint64 requestID = 1;
        uint64 timestamp = 1684155125000;
        string memory method = "echo";
        bytes memory params = abi.encode("Hello");
        bytes memory result = abi.encode("Hello");

        State memory state = createRPCState(requestID, timestamp, method, params, result);
        state.sigs = new Signature[](2);
        state.sigs[CLIENT_IDX] = signRequest(state, clientPrivateKey);
        state.sigs[SERVER_IDX] = signResponse(state, serverPrivateKey);

        // Create a previous state with SAME timestamp
        uint64 prevRequestID = 0;
        uint64 prevTimestamp = timestamp; // Same timestamp
        State memory prevState = createRPCState(prevRequestID, prevTimestamp, method, params, result);
        prevState.sigs = new Signature[](2);
        prevState.sigs[CLIENT_IDX] = signRequest(prevState, clientPrivateKey);
        prevState.sigs[SERVER_IDX] = signResponse(prevState, serverPrivateKey);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        bool valid = nitroRpcAdjudicator.adjudicate(channel, state, proofs);
        assertFalse(valid, "State with timestamp not greater than previous should be rejected");
    }

    // State with timestamp earlier than previous state should fail
    function test_InvalidTimestamp_Earlier() public view {
        uint64 requestID = 1;
        uint64 timestamp = 1684155125000;
        string memory method = "echo";
        bytes memory params = abi.encode("Hello");
        bytes memory result = abi.encode("Hello");

        State memory state = createRPCState(requestID, timestamp, method, params, result);
        state.sigs = new Signature[](2);
        state.sigs[CLIENT_IDX] = signRequest(state, clientPrivateKey);
        state.sigs[SERVER_IDX] = signResponse(state, serverPrivateKey);

        // Create a previous state with later timestamp
        uint64 prevRequestID = 0;
        uint64 prevTimestamp = timestamp + 1000; // Later timestamp
        State memory prevState = createRPCState(prevRequestID, prevTimestamp, method, params, result);
        prevState.sigs = new Signature[](2);
        prevState.sigs[CLIENT_IDX] = signRequest(prevState, clientPrivateKey);
        prevState.sigs[SERVER_IDX] = signResponse(prevState, serverPrivateKey);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        bool valid = nitroRpcAdjudicator.adjudicate(channel, state, proofs);
        assertFalse(valid, "State with earlier timestamp than previous should be rejected");
    }

    // -------------------- REQUEST ID TESTS --------------------

    // State with request ID not greater than previous state should fail
    function test_InvalidRequestID_NotGreater() public view {
        uint64 requestID = 5;
        uint64 timestamp = 1684155130000;
        string memory method = "echo";
        bytes memory params = abi.encode("Hello");
        bytes memory result = abi.encode("Hello");

        State memory state = createRPCState(requestID, timestamp, method, params, result);
        state.sigs = new Signature[](2);
        state.sigs[CLIENT_IDX] = signRequest(state, clientPrivateKey);
        state.sigs[SERVER_IDX] = signResponse(state, serverPrivateKey);

        // Create a previous state with SAME request ID
        uint64 prevRequestID = requestID; // Same request ID
        uint64 prevTimestamp = timestamp - 1000; // Earlier timestamp (valid)
        State memory prevState = createRPCState(prevRequestID, prevTimestamp, method, params, result);
        prevState.sigs = new Signature[](2);
        prevState.sigs[CLIENT_IDX] = signRequest(prevState, clientPrivateKey);
        prevState.sigs[SERVER_IDX] = signResponse(prevState, serverPrivateKey);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        bool valid = nitroRpcAdjudicator.adjudicate(channel, state, proofs);
        assertFalse(valid, "State with request ID not greater than previous should be rejected");
    }

    // State with request ID less than previous state should fail
    function test_InvalidRequestID_Less() public view {
        uint64 requestID = 5;
        uint64 timestamp = 1684155130000;
        string memory method = "echo";
        bytes memory params = abi.encode("Hello");
        bytes memory result = abi.encode("Hello");

        State memory state = createRPCState(requestID, timestamp, method, params, result);
        state.sigs = new Signature[](2);
        state.sigs[CLIENT_IDX] = signRequest(state, clientPrivateKey);
        state.sigs[SERVER_IDX] = signResponse(state, serverPrivateKey);

        // Create a previous state with higher request ID
        uint64 prevRequestID = requestID + 1; // Higher request ID
        uint64 prevTimestamp = timestamp - 1000; // Earlier timestamp (valid)
        State memory prevState = createRPCState(prevRequestID, prevTimestamp, method, params, result);
        prevState.sigs = new Signature[](2);
        prevState.sigs[CLIENT_IDX] = signRequest(prevState, clientPrivateKey);
        prevState.sigs[SERVER_IDX] = signResponse(prevState, serverPrivateKey);

        State[] memory proofs = new State[](1);
        proofs[0] = prevState;

        bool valid = nitroRpcAdjudicator.adjudicate(channel, state, proofs);
        assertFalse(valid, "State with request ID less than previous should be rejected");
    }

    // -------------------- PROOF TESTS --------------------

    // State with no proofs should fail
    function test_NoProofs() public view {
        uint64 requestID = 1;
        uint64 timestamp = 1684155125000;
        string memory method = "echo";
        bytes memory params = abi.encode("Hello");
        bytes memory result = abi.encode("Hello");

        State memory state = createRPCState(requestID, timestamp, method, params, result);
        state.sigs = new Signature[](2);
        state.sigs[CLIENT_IDX] = signRequest(state, clientPrivateKey);
        state.sigs[SERVER_IDX] = signResponse(state, serverPrivateKey);

        // No proofs
        State[] memory proofs = new State[](0);

        bool valid = nitroRpcAdjudicator.adjudicate(channel, state, proofs);
        assertFalse(valid, "State with no proofs should be rejected");
    }

    // State with too many proofs should fail
    function test_TooManyProofs() public view {
        uint64 requestID = 1;
        uint64 timestamp = 1684155125000;
        string memory method = "echo";
        bytes memory params = abi.encode("Hello");
        bytes memory result = abi.encode("Hello");

        State memory state = createRPCState(requestID, timestamp, method, params, result);
        state.sigs = new Signature[](2);
        state.sigs[CLIENT_IDX] = signRequest(state, clientPrivateKey);
        state.sigs[SERVER_IDX] = signResponse(state, serverPrivateKey);

        // Create previous states
        uint64 prevRequestID1 = 0;
        uint64 prevTimestamp1 = timestamp - 2000;
        State memory prevState1 = createRPCState(prevRequestID1, prevTimestamp1, method, params, result);
        prevState1.sigs = new Signature[](2);
        prevState1.sigs[CLIENT_IDX] = signRequest(prevState1, clientPrivateKey);
        prevState1.sigs[SERVER_IDX] = signResponse(prevState1, serverPrivateKey);

        uint64 prevRequestID2 = 0;
        uint64 prevTimestamp2 = timestamp - 1000;
        State memory prevState2 = createRPCState(prevRequestID2, prevTimestamp2, method, params, result);
        prevState2.sigs = new Signature[](2);
        prevState2.sigs[CLIENT_IDX] = signRequest(prevState2, clientPrivateKey);
        prevState2.sigs[SERVER_IDX] = signResponse(prevState2, serverPrivateKey);

        // Two proofs (too many)
        State[] memory proofs = new State[](2);
        proofs[0] = prevState1;
        proofs[1] = prevState2;

        bool valid = nitroRpcAdjudicator.adjudicate(channel, state, proofs);
        assertFalse(valid, "State with too many proofs should be rejected");
    }

    // -------------------- ADDITIONAL FUNCTIONALITY TESTS --------------------

    // Test the compare function directly
    function test_Compare() public view {
        // State with later timestamp
        uint64 laterRequestID = 2;
        uint64 laterTimestamp = 1684155130000;
        string memory method = "echo";
        bytes memory params = abi.encode("Hello");
        bytes memory result = abi.encode("Hello");
        State memory laterState = createRPCState(laterRequestID, laterTimestamp, method, params, result);

        // State with earlier timestamp
        uint64 earlierRequestID = 1;
        uint64 earlierTimestamp = 1684155125000;
        State memory earlierState = createRPCState(earlierRequestID, earlierTimestamp, method, params, result);

        // Later state should be greater (return 1)
        int8 compareResult = nitroRpcAdjudicator.compare(laterState, earlierState);
        assertEq(compareResult, 1, "Later state should be greater than earlier state");

        // Earlier state should be less (return -1)
        compareResult = nitroRpcAdjudicator.compare(earlierState, laterState);
        assertEq(compareResult, -1, "Earlier state should be less than later state");

        // Same timestamp should be equal (return 0)
        State memory sameTimestampState = createRPCState(laterRequestID, laterTimestamp, method, params, result);
        compareResult = nitroRpcAdjudicator.compare(laterState, sameTimestampState);
        assertEq(compareResult, 0, "States with same timestamp should be equal");
    }
}
