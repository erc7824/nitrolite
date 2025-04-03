// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";

/**
 * @title Counter Adjudicator
 * @notice Implements a strict turn‐taking counter game.
 * @dev Host sets the initial counter value. After funding the channel, the state is ACTIVE only if counter > 0.
 *      Host and Guest take strict alternating turns to increment the counter.
 *      When the counter reaches the target, the game ends with FINAL status.
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

    uint256 private constant HOST = 0;
    uint256 private constant GUEST = 1;

    /**
     * @dev CounterApp represents the game state.
     * @param counter Current counter value.
     * @param target  Target counter value at which the game ends.
     * @param version State version number starting from 0.
     */
    struct CounterApp {
        uint256 counter;
        uint256 target;
        uint256 version;
    }

    /**
     * @notice Validates that the counter state transition is valid with strict turn‐taking.
     * @param chan The channel configuration.
     * @param candidate The proposed new state.
     * @param proofs Array containing the previous state.
     * @return valid True if the state transition is valid, false otherwise.
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        pure
        override
        returns (bool valid)
    {
        // Ensure the candidate state is signed by both participants.
        if (candidate.sigs.length != 2) {
            return false;
        }

        // Compute the state hash for signature verification.
        bytes32 stateHash = Utils.getStateHash(chan, candidate);

        // Decode the candidate state data.
        CounterApp memory candidateState = abi.decode(candidate.data, (CounterApp));

        // INITIAL STATE: version 0 requires both signatures.
        if (candidateState.version == 0) {
            // true if both signatures are valid
            return Utils.verifySignature(stateHash, candidate.sigs[0], chan.participants[HOST])
                && Utils.verifySignature(stateHash, candidate.sigs[1], chan.participants[GUEST]);
        }

        // For non-initial states, ensure a previous state is provided.
        if (proofs.length == 0) {
            return false;
        }

        // Decode the previous state.
        CounterApp memory previousState = abi.decode(proofs[0].data, (CounterApp));

        // Validate that the counter increment is exactly 1.
        if (candidateState.counter != previousState.counter + 1) {
            return false;
        }

        // Validate that the version increment is exactly 1.
        if (candidateState.version != previousState.version + 1) {
            return false;
        }

        // Ensure the rules of the game are consistent, for simplisity.
        if (candidateState.target != previousState.target) {
            return false;
        }

        // Ensure the candidate counter does not exceed its target.
        if (candidateState.counter > candidateState.target) {
            return false;
        }

        // All validations passed.
        return true;
    }
}
