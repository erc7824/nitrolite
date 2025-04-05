// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Status, Allocation, Signature} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";

/**
 * @title NitroRPC Adjudicator
 * @notice Implements an adjudicator for NitroRPC message format designed for state channels.
 * @dev NitroRPC is an asynchronous RPC message format where requests are signed by the initiator
 *      and responses are countersigned. The use of a server-side universal millisecond timestamp
 *      builds a tamper-proof history. This adjudicator enforces the rules defined in the NitroRPC
 *      specification.
 */
contract NitroRPC is IAdjudicator {
    /// @notice Error thrown when signature verification fails
    error InvalidSignature();
    /// @notice Error thrown when the timestamp is invalid
    error InvalidTimestamp();
    /// @notice Error thrown when request ID doesn't match
    error InvalidRequestId();
    /// @notice Error thrown when insufficient signatures are provided
    error InsufficientSignatures();
    /// @notice Error thrown when the open channel magic number is invalid
    error InvalidOpenChannelMagic();

    //FIXME: find coherent naming
    uint256 private constant HOST = 0; // Client
    uint256 private constant GUEST = 1; // Server

    /**
     * @dev RPCError represents an error returned by the RPC server.
     * @param code Error code.
     * @param message Error message.
     */
    struct RPCError {
        uint32 code;
        string message;
    }

    /**
     * @dev RPCMessage represents a NitroRPC message.
     * @param requestID Unique identifier for the request.
     * @param method Remote method name to be invoked.
     * @param params Method parameters (client).
     * @param result Method result (server).
     * @param timestamp Server timestamp in milliseconds.
     */
    struct RPCMessage {
        uint64 requestID;
        string method;
        bytes[] params;
        bytes[] result;
        uint64 timestamp;
    }

    /**
     * @notice Computes the state hash for an RPCMessage
     * @param message The RPCMessage
     * @return The hash of the RPCMessage
     */
    function getReqStateHash(bytes32 channelId, State memory state, RPCMessage memory message)
        internal
        pure
        returns (bytes32)
    {
        return keccak256(
            abi.encode(
                channelId, state.allocations, message.requestID, message.method, message.params, message.timestamp
            )
        );
    }

    function getResStateHash(bytes32 channelId, State memory state, RPCMessage memory message)
        internal
        pure
        returns (bytes32)
    {
        return keccak256(
            abi.encode(
                channelId,
                state.allocations,
                message.requestID,
                message.method,
                message.params,
                message.result,
                message.timestamp
            )
        );
    }

    /**
     * @notice Validates that the NitroRPC state transition is valid according to the rules.
     * @param chan The channel configuration.
     * @param candidate The proposed new state.
     * @param proofs Array containing the previous state(s).
     * @return valid True if the state transition is valid, false otherwise.
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        pure
        override
        returns (bool valid)
    {
        // Ensure the candidate state is signed by both participants.
        if (candidate.sigs.length != 2) return false;
        if (candidate.stage != Status.ACTIVE) return false;
        // Compute the state hash for signature verification.
        RPCMessage memory candidateState = abi.decode(candidate.data, (RPCMessage));

        bytes32 channelId = Utils.getChannelId(chan);
        bytes32 reqStateHash = getReqStateHash(channelId, candidate, candidateState);
        bytes32 resStateHash = getResStateHash(channelId, candidate, candidateState);

        // Ensure the client signature is valid for the request
        if (!Utils.verifySignature(reqStateHash, candidate.sigs[HOST], chan.participants[HOST])) {
            return false;
        }

        // Ensure the server signature is valid for the response
        if (!Utils.verifySignature(resStateHash, candidate.sigs[GUEST], chan.participants[GUEST])) {
            return false;
        }

        if (proofs.length != 1) {
            return false;
        }

        // Decode the previous state.
        RPCMessage memory previousState = abi.decode(proofs[0].data, (RPCMessage));

        // Validate that the timestamp in the new state is greater than in the previous state
        if (candidateState.timestamp <= previousState.timestamp) {
            return false;
        }

        // Validate that server response contains the same request ID as the request
        if (candidateState.requestID <= previousState.requestID) {
            return false;
        }

        // All validations passed.
        return true;
    }
}
