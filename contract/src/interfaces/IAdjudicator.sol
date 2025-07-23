// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Channel, State} from "./Types.sol";

/**
 * @title Adjudicator Interface
 * @notice Interface for state validation and outcome determination
 * @dev Implementations validate state transitions according to application-specific rules
 */
interface IAdjudicator {
    /**
     * @notice Validates a candidate state based on application-specific rules
     * @dev Used to determine if a state is valid during challenges or checkpoints
     * @param chan The channel configuration with participants, adjudicator, challenge period, and nonce
     * @param candidate The proposed state to be validated
     * @param proofs Array of previous states that provide context for validation
     * @return valid True if the candidate state is valid according to application rules
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        returns (bool valid);
}
