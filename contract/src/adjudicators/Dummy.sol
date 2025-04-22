// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Channel, State} from "../interfaces/Types.sol";
import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {IComparable} from "../interfaces/IComparable.sol";

/**
 * @title Dummy Adjudicator
 * @notice A simple adjudicator that always validates states as true and considers newer states as more recent
 * @dev This is a minimal implementation for testing or simple channels where all states are valid
 */
contract Dummy is IAdjudicator, IComparable {
    /**
     * @notice Always validates candidate states as true
     * @dev This implementation accepts any state regardless of content
     * @param chan The channel configuration (unused in this implementation)
     * @param candidate The proposed state to be validated (unused in this implementation)
     * @param proofs Array of previous states (unused in this implementation)
     * @return valid Always returns true
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        pure
        returns (bool valid)
    {
        // Always return true regardless of inputs
        return true;
    }

    /**
     * @notice Always considers candidate state as newer than previous state
     * @dev This implementation always returns 1 to indicate candidate is more recent
     * @param candidate The state being evaluated (unused in this implementation)
     * @param previous The reference state to compare against (unused in this implementation)
     * @return result Always returns 1 (candidate is newer)
     */
    function compare(State calldata candidate, State calldata previous) external pure returns (int8 result) {
        // Always indicate that the candidate state is newer
        return 1;
    }
}
