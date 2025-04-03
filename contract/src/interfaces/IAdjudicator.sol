// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Channel, State, Allocation} from "./Types.sol";

/**
 * @title Adjudicator Interface
 * @notice Interface for state validation and outcome determination
 */
interface IAdjudicator {
    /**
     * @notice Validates the application state and determines the outcome of a channel
     * @dev This function evaluates the validity of a candidate state against provided proofs
     * @param chan The channel information containing participants, adjudicator, nonce, and challenge period
     * @param candidate The proposed state to be validated
     * @param proofs Array of previous states that may be used to validate the candidate state
     * @return valid is true if the candidate is approved
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        view
        returns (bool valid);
}
