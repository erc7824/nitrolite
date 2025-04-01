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

    uint256 private constant FINAL_COUNTER = 1000;
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
     * @return decision The status of the channel after adjudication
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        view
        override
        returns (Status decision)
    {
        // Check if we have at least one signature
        if (candidate.sigs.length == 0) return Status.INVALID;

        // Get the state hash for signature verification
        bytes32 stateHash = Utils.getStateHash(chan, candidate);

        // Decode the counter from candidate state.data
        CounterData memory candidateCounterData = abi.decode(candidate.data, (CounterData));

        // INITIAL STATE ACTIVATION: No proofs provided
        if (proofs.length == 0) {
            // First signature must be from HOST who sets initial counter
            if (!Utils.verifySignature(stateHash, candidate.sigs[0], chan.participants[HOST])) {
                return Status.VOID;
            }

            // If only Host has signed, channel is PARTIAL
            if (candidate.sigs.length < 2) {
                return Status.PARTIAL;
            }

            // Both signatures provided, verify Guest's signature
            if (!Utils.verifySignature(stateHash, candidate.sigs[1], chan.participants[GUEST])) {
                return Status.VOID;
            }

            // Channel becomes ACTIVE only if counter > 0
            if (candidateCounterData.counter > 0) {
                return Status.ACTIVE;
            } else {
                return Status.PARTIAL;
            }
        }

        // NORMAL STATE TRANSITION: Proof provided.
        // Ensure proof state has at least one signature
        if (proofs[0].sigs.length == 0) return Status.INVALID;

        CounterData memory proofCounterData = abi.decode(proofs[0].data, (CounterData));

        // Verify the increment is exactly 1
        if (candidateCounterData.counter != proofCounterData.counter + 1) {
            return Status.INVALID;
        }

        bytes32 proofStateHash = Utils.getStateHash(chan, proofs[0]);

        // When Host is the signer of candidate, Guest must have signed the proof
        if (Utils.verifySignature(stateHash, candidate.sigs[0], chan.participants[HOST])) {
            // Verify Guest signed the proof
            if (!Utils.verifySignature(proofStateHash, proofs[0].sigs[0], chan.participants[GUEST])) {
                return Status.INVALID;
            }
        }
        // When Guest is the signer of candidate, Host must have signed the proof
        else if (Utils.verifySignature(stateHash, candidate.sigs[0], chan.participants[GUEST])) {
            // Verify Host signed the proof
            if (!Utils.verifySignature(proofStateHash, proofs[0].sigs[0], chan.participants[HOST])) {
                return Status.INVALID;
            }
        }
        // Invalid signature on candidate
        else {
            return Status.INVALID;
        }

        // Check if counter has reached or exceeded the final value
        if (candidateCounterData.counter >= FINAL_COUNTER) {
            return Status.FINAL;
        }

        // Valid state transition, channel remains ACTIVE
        return Status.ACTIVE;
    }
}
