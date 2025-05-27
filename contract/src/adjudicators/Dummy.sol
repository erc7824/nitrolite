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
     */
    function adjudicate(Channel calldata, State calldata, State[] calldata) external pure returns (bool valid) {
        // Always return true regardless of inputs
        return true;
    }

    /**
     * @notice Always considers candidate state as newer than previous state
     * @dev This implementation always returns 1 to indicate candidate is more recent
     */
    function compare(State calldata, State calldata) external pure returns (int8 result) {
        // Always indicate that the candidate state is newer
        return 1;
    }
}
