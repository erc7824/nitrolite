// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature, StateIntent} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";

/**
 * @title Simple mutual consent Adjudicator
 * @notice An adjudicator that validates state based on mutual signatures from both participants.
 * @dev Any state is considered valid as long as it's signed by both participants.
 */
contract SimpleConsensus is IAdjudicator {
    using Utils for State;

    /**
     * @notice Validates that the state is signed by both participants
     * @param chan The channel configuration
     * @param candidate The proposed state
     * @param proofs Array of previous states (unused in this implementation)
     * @return valid True if the state is valid, false otherwise
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        view
        override
        returns (bool valid)
    {
        if (proofs.length != 0) {
            return false;
        }

        if (candidate.version == 0) {
            return candidate.validateInitialState(chan);
        }

        // proof is Operate or Resize State (both have same validation)
        return candidate.intent != StateIntent.INITIALIZE && candidate.validateUnanimousSignatures(chan);
    }
}
