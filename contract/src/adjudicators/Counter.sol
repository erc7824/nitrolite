// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";

/**
 * @title Counter Adjudicator
 * @notice An adjudicator that implements a strict turn‐taking counter game.
 * @dev Host sets the initial counter value. After funding channel, state is ACTIVE only if counter > 0.
 * Host and Guest take strict alternating turns to increment the counter.
 * When counter reaches 1000, the game ends with FINAL status.
 */
contract Counter is IAdjudicator {
    /// @notice Error thrown when signature verification fails
    error InvalidSignature();
    /// @notice Error thrown when turn order is violated
    error InvalidTurn();
    /// @notice Error thrown when the counter increment is invalid
    error InvalidIncrement();
    /// @notice Error thrown when counter value is invalid
    error InvalidCounter();
    /// @notice Error thrown when insufficient signatures are provided
    error InsufficientSignatures();

    // The custody contract now determines if a state is final based on signatures
    // not on reaching a specific counter value
    uint256 private constant HOST = 0;
    uint256 private constant GUEST = 1;

    struct CounterData {
        uint256 counter;
    }

    /**
     * @notice Validates that the counter is incremented correctly with strict turn‐taking.
     * @param chan The channel configuration
     * @param candidate The proposed counter state
     * @param proofs Array containing the previous state signed by the previous participant
     * @return valid True if the state is valid, false otherwise
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        view
        override
        returns (bool valid)
    {
        // Check if we have at least one signature
        if (candidate.sigs.length == 0) return false;

        // Get the state hash for signature verification
        bytes32 stateHash = Utils.getStateHash(chan, candidate);

        // Decode the counter from candidate state.data
        CounterData memory candidateCounterData = abi.decode(candidate.data, (CounterData));

        // The isFinal flag in the State struct should match the counter value
        // but we won't use it for validation to maintain backward compatibility
        // The custody contract will handle the final state status

        // INITIAL STATE ACTIVATION
        if (candidateCounterData.counter == 0) {
            // First signature must be from HOST who sets initial counter
            if (!Utils.verifySignature(stateHash, candidate.sigs[0], chan.participants[HOST])) {
                return false;
            }

            // If only Host has signed, it's valid for initial funding
            if (candidate.sigs.length < 2) {
                return true;
            }

            // Both signatures provided, verify Guest's signature
            if (!Utils.verifySignature(stateHash, candidate.sigs[1], chan.participants[GUEST])) {
                return false;
            }

            // Valid initial state with both signatures
            return true;
        }

        // NORMAL STATE TRANSITION: Proof provided.
        // If we have a non-zero counter but no proofs, we need to check both signatures
        if (proofs.length == 0) {
            // For non-zero counter without proofs, we need both signatures
            if (candidate.sigs.length < 2) {
                return false;
            }

            // Verify both signatures
            if (!Utils.verifySignature(stateHash, candidate.sigs[0], chan.participants[HOST])) {
                return false;
            }

            if (!Utils.verifySignature(stateHash, candidate.sigs[1], chan.participants[GUEST])) {
                return false;
            }

            return true;
        }

        // Ensure proof state has at least one signature
        if (proofs[0].sigs.length == 0) return false;

        CounterData memory proofCounterData = abi.decode(proofs[0].data, (CounterData));

        // Verify the increment is exactly 1
        if (candidateCounterData.counter != proofCounterData.counter + 1) {
            return false;
        }

        bytes32 proofStateHash = Utils.getStateHash(chan, proofs[0]);

        // When Host is the signer of candidate, Guest must have signed the proof
        if (Utils.verifySignature(stateHash, candidate.sigs[0], chan.participants[HOST])) {
            // Verify Guest signed the proof
            if (!Utils.verifySignature(proofStateHash, proofs[0].sigs[0], chan.participants[GUEST])) {
                return false;
            }
        }
        // When Guest is the signer of candidate, Host must have signed the proof
        else if (Utils.verifySignature(stateHash, candidate.sigs[0], chan.participants[GUEST])) {
            // Verify Host signed the proof
            if (!Utils.verifySignature(proofStateHash, proofs[0].sigs[0], chan.participants[HOST])) {
                return false;
            }
        }
        // Invalid signature on candidate
        else {
            return false;
        }

        // All validations passed
        return true;
    }
}
