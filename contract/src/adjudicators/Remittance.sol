// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {IComparable} from "../interfaces/IComparable.sol";
import {Amount, Channel, State, Status, Allocation, Signature} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";

/**
 * @title Remittance Adjudicator
 * @notice Implements an adjudicator for simple remittance payments in state channels
 * @dev Validates transfers are signed by the payer and uses versioning to establish ordering.
 *      The data field contains a version counter that must increase with each allocation change.
 */
contract Remittance is IAdjudicator, IComparable {
    /// @notice Error thrown when signature verification fails
    error InvalidSignature();
    /// @notice Error thrown when version isn't higher than previous version
    error InvalidVersion();
    /// @notice Error thrown when payer hasn't signed the state
    error PayerSignatureRequired();
    /// @notice Error thrown when required proofs aren't provided
    error InsufficientProofs();
    /// @notice Error thrown when allocations are invalid
    error InvalidAllocations();

    /**
     * @dev Remittance data structure
     * @param version State version counter - must increase with each allocation change
     */
    struct Voucher {
        uint8 payer; // Index of the Payer
        Amount transfer; // Amount and token should match the updated allocation
        uint256 version;
    }

    /**
     * @notice Compares two states based on their version numbers
     * @dev Uses the version field from Voucher to determine state ordering
     * @param candidate The state being evaluated
     * @param previous The reference state to compare against
     * @return result The comparison result:
     *         -1: candidate < previous (candidate is older)
     *          0: candidate == previous (same recency)
     *          1: candidate > previous (candidate is newer)
     */
    function compare(State calldata candidate, State calldata previous) external pure returns (int8 result) {
        Voucher memory candidateData = abi.decode(candidate.data, (Voucher));
        Voucher memory previousData = abi.decode(previous.data, (Voucher));

        if (candidateData.version < previousData.version) return -1;
        if (candidateData.version > previousData.version) return 1;
        return 0;
    }

    /**
     * @notice Identifies which participant's allocation is decreasing (the payer)
     * @param current Current allocations
     * @param previous Previous allocations
     * @return payerIndex The index of the payer in the participants array
     */
    function identifyPayer(Allocation[] memory current, Allocation[] memory previous)
        internal
        pure
        returns (uint256 payerIndex)
    {
        // Simple case with 2 participants
        if (current.length != 2 || previous.length != 2) {
            revert InvalidAllocations();
        }

        // Search for participant with decreased allocation
        for (uint256 i = 0; i < 2; i++) {
            if (
                current[i].token == previous[i].token && current[i].destination == previous[i].destination
                    && current[i].amount < previous[i].amount
            ) {
                return i; // Found the payer
            }
        }

        // If no allocations decreased, revert
        revert InvalidAllocations();
    }

    /**
     * @notice Validates remittance state transitions
     * @dev Ensures the payer has signed states with decreased allocations and version is incremented
     * @param chan The channel configuration
     * @param candidate The proposed state to be validated
     * @param proofs Array of previous states for validation context
     * @return valid True if the state transition is valid
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        pure
        override
        returns (bool valid)
    {
        // Need both funding proof and previous state for validation
        if (proofs.length < 2) {
            return false;
        }

        // Decode the version from the candidate state
        Voucher memory candidateData = abi.decode(candidate.data, (Voucher));

        // Get the previous state (most recent proof)
        State memory previousState = proofs[0];
        Voucher memory previousData = abi.decode(previousState.data, (Voucher));

        // Version must increase with each allocation change
        if (candidateData.version <= previousData.version) {
            return false;
        }

        // Compute state hashes for verification
        bytes32 stateHash = Utils.getStateHash(chan, candidate);

        // Identify the payer (participant whose allocation is decreasing)
        uint256 payerIndex = identifyPayer(candidate.allocations, previousState.allocations);

        // Payer must have signed the candidate state
        if (
            candidate.sigs.length <= payerIndex
                || !Utils.verifySignature(stateHash, candidate.sigs[payerIndex], chan.participants[payerIndex])
        ) {
            return false;
        }

        // All validations passed
        return true;
    }
}
