// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature, OPENCHAN} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";

/**
 * @title NitroRPC Adjudicator
 * @notice Implements validation for RPC-based state channel communication
 * @dev Validates RPC request/response pairs with signatures from both client and server
 */
contract NitroRPC is IAdjudicator {
    /// @notice Error thrown when signature verification fails
    error InvalidSignature();
    /// @notice Error thrown when timestamp validation fails
    error InvalidTimestamp();
    /// @notice Error thrown when request ID validation fails
    error InvalidRequestID();
    /// @notice Error thrown when required proofs are missing
    error InsufficientProofs();
    /// @notice Error thrown when the format of RPC message is invalid
    error InvalidRPCFormat();

    uint256 private constant HOST = 0; // Client
    uint256 private constant GUEST = 1; // Server

    /**
     * @dev RPCMessage represents an RPC communication
     * @param requestID Unique identifier for the request
     * @param method Method name being called
     * @param params Parameters for the request
     * @param result Results from the response
     * @param timestamp Server timestamp in milliseconds
     */
    struct RPCMessage {
        uint64 requestID;
        string method;
        bytes[] params;
        bytes[] result;
        uint64 timestamp;
    }

    /**
     * @notice Validates an RPC request/response pair with signatures from both parties
     * @param chan The channel configuration
     * @param candidate The proposed new state
     * @param proofs Array containing previous states
     * @return valid True if the state transition is valid, false otherwise
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        pure
        override
        returns (bool valid)
    {
        // Ensure the candidate state is signed by both participants
        if (candidate.sigs.length != 2) {
            return false;
        }

        // Compute the state hash for signature verification
        bytes32 stateHash = Utils.getStateHash(chan, candidate);

        // Verify both signatures
        bool hostSigValid = Utils.verifySignature(stateHash, candidate.sigs[HOST], chan.participants[HOST]);
        bool guestSigValid = Utils.verifySignature(stateHash, candidate.sigs[GUEST], chan.participants[GUEST]);
        
        if (!hostSigValid || !guestSigValid) {
            return false;
        }

        // Decode the candidate state data
        RPCMessage memory candidateRPC = abi.decode(candidate.data, (RPCMessage));

        // Initial funding state validation
        if (proofs.length == 0) {
            // First state must contain the OPENCHAN magic number in the first parameter
            if (candidateRPC.params.length == 0) {
                return false;
            }
            
            uint16 magicNumber;
            if (candidateRPC.params[0].length == 2) {
                magicNumber = uint16(bytes2(candidateRPC.params[0]));
                return magicNumber == OPENCHAN;
            }
            
            return false;
        }

        // For subsequent states, ensure a previous state is provided
        if (proofs.length == 0) {
            return false;
        }

        // Decode the previous state
        RPCMessage memory previousRPC = abi.decode(proofs[0].data, (RPCMessage));

        // Validate timestamp progression
        if (candidateRPC.timestamp <= previousRPC.timestamp) {
            return false;
        }

        // Validate request ID is unique and valid
        if (candidateRPC.requestID <= previousRPC.requestID) {
            return false;
        }

        // All validations passed
        return true;
    }
}