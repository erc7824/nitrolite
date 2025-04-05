// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {State} from "./Types.sol";

/**
 * @title Comparable Interface
 * @notice Interface for contracts that can determine ordering between states
 * @dev Used to determine which state is more recent during challenge resolution
 */
interface IComparable {
    /**
     * @notice Compares two states to determine their relative ordering
     * @dev Implementations should return:
     *      -1 if candidate is less recent than previous
     *       0 if candidate is equally recent as previous
     *       1 if candidate is more recent than previous
     * @param candidate The state being evaluated
     * @param previous The reference state to compare against
     * @return result The comparison result:
     *         -1: candidate < previous (candidate is older)
     *          0: candidate == previous (same recency)
     *          1: candidate > previous (candidate is newer)
     */
    function compare(State calldata candidate, State calldata previous) external view returns (int8 result);
}
