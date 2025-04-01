// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation} from "../interfaces/Types.sol";

/**
 * @title Trivial Adjudicator
 * @notice A simple adjudicator that always returns ACTIVE status
 * @dev Used primarily for testing and demonstration purposes
 */
contract Trivial is IAdjudicator {
    /**
     * @notice Always returns ACTIVE status regardless of inputs
     * @param chan The channel parameters
     * @param candidate The candidate state to adjudicate
     * @param proofs Previous state proofs (unused in this implementation)
     * @return decision Always returns Status.ACTIVE
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        pure
        override
        returns (Status decision)
    {
        return Status.ACTIVE;
    }
}
