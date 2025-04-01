// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";

/**
 * @title MicroPayment Adjudicator
 * @notice An adjudicator that implements a micro-payment channel where only the Host makes deposits
 * @dev Host deposits funds and progressively signs states to transfer value to the Guest
 * Each state.data contains a single uint256 representing the amount to be transferred to Guest
 * Only the Host signs states, and the amount can only increase (not decrease)
 */
contract MicroPayment is IAdjudicator {
    /// @notice Error thrown when signature verification fails
    error InvalidSignature();
    /// @notice Error thrown when state is not signed by Host
    error NotSignedByHost();
    /// @notice Error thrown when payment amount decreases in a new state
    error DecreasingPayment();
    /// @notice Error thrown when payment amount exceeds Host's deposit
    error PaymentExceedsDeposit();
    /// @notice Error thrown when insufficient signatures are provided
    error InsufficientSignatures();

    uint256 private constant HOST = 0;
    uint256 private constant GUEST = 1;

    /**
     * @notice Validates payment channel states where Host progressively signs increasing payments to Guest
     * @param chan The channel configuration
     * @param candidate The proposed payment state
     * @param proofs Array containing previous states (optional)
     * @return decision The status of the channel after adjudication
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        pure
        override
        returns (Status decision)
    {
        // Ensure at least one signature exists
        if (candidate.sigs.length == 0) return Status.INVALID;

        // Get the state hash for signature verification
        bytes32 stateHash = Utils.getStateHash(chan, candidate);

        // Verify that the Host signed the state
        if (!Utils.verifySignature(stateHash, candidate.sigs[0], chan.participants[HOST])) {
            return Status.INVALID;
        }

        // Decode the payment amount from candidate state.data
        uint256 paymentAmount = abi.decode(candidate.data, (uint256));

        // If we have a previous state, verify payment amount is not decreasing
        if (proofs.length > 0 && proofs[0].sigs.length > 0) {
            // Validate previous state is also signed by Host
            bytes32 proofStateHash = Utils.getStateHash(chan, proofs[0]);
            if (!Utils.verifySignature(proofStateHash, proofs[0].sigs[0], chan.participants[HOST])) {
                return Status.INVALID;
            }

            // Decode the payment amount from proof state
            uint256 previousAmount = abi.decode(proofs[0].data, (uint256));

            // Payment amount must not decrease
            if (paymentAmount < previousAmount) {
                return Status.INVALID;
            }
        }

        // Calculate the total deposited by Host
        // Note: We are still reading from allocations for validation purposes
        // even though we don't return them anymore
        uint256 hostDeposit = candidate.allocations[HOST].amount + candidate.allocations[GUEST].amount;

        // Payment amount cannot exceed Host's deposit
        if (paymentAmount > hostDeposit) {
            return Status.INVALID;
        }

        // For micro-payment channels, states are always ACTIVE until participants decide to close
        // A payment of the full deposit amount makes the channel FINAL
        if (paymentAmount == hostDeposit) {
            return Status.FINAL;
        }

        return Status.ACTIVE;
    }
}
