// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Channel, State, Allocation} from "./Types.sol";

/**
 * @title Adjudicator Interface
 * @notice Interface for state validation and outcome determination
 */
interface IAdjudicator {
    enum Status {
        VOID, // Channel was never active or have an anomaly
        PARTIAL, // Partial funding waiting for other participants
        ACTIVE, // Channel fully funded using open or state are valid
        INVALID, // Channel state is invalid
        FINAL // This is the FINAL State channel can be closed
    }

    /**
     * @notice Validates the application state and determines the outcome of a channel
     * @dev This function evaluates the validity of a candidate state against provided proofs
     * @param chan The channel information containing participants, adjudicator, nonce, and challenge period
     * @param candidate The proposed state to be validated
     * @param proofs Array of previous states that may be used to validate the candidate state
     * @return decision The status of the channel after adjudication
     * @return allocations The final allocations for participants based on the adjudication
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        view
        returns (Status decision, Allocation[2] memory allocations);
}
