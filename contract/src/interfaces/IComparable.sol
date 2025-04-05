// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {State} from "./Types.sol";

/**
 * @title Comparable Interface
 * @notice Interface for contracts that can compare two states
 * @dev Implementations should return:
 *      -1 if candidate is less than previous
 *       0 if candidate is equal to previous
 *       1 if candidate is greater than previous
 */
interface IComparable {
    /**
     * @notice Compare two states and determine their relative ordering
     * @param candidate The state being evaluated
     * @param previous The reference state to compare against
     * @return result The comparison result:
     *         -1: candidate < previous
     *          0: candidate == previous
     *          1: candidate > previous
     */
    function compare(State calldata candidate, State calldata previous) external view returns (uint8 result);
}

