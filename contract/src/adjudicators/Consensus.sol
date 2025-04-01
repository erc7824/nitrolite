// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";

/**
 * @title MutualConsent Adjudicator
 * @notice An adjudicator that validates state based on mutual signatures from both participants
 * @dev Any state is considered valid as long as it's signed by both participants
 */
contract Consensus is IAdjudicator {
    uint256 constant HOST = 0;
    uint256 constant GUEST = 1;

    /// @notice Error thrown when signature verification fails
    error InvalidSignature();

    /// @notice Error thrown when insufficient signatures are provided
    error InsufficientSignatures();

    enum AppStatus {
        Starting,
        Ready,
        Finish
    }

    struct AppData {
        bytes appData; // Application-specific data
        AppStatus status; // Application-specific Status
    }

    /**
     * @notice Validates that the state is signed by both participants
     * @param chan The channel configuration
     * @param candidate The proposed state
     * @param proofs Array of previous states (unused in this implementation)
     * @return decision The status of the channel after adjudication
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        view
        override
        returns (IAdjudicator.Status decision)
    {
        // Check for insufficient signatures
        if (candidate.sigs.length == 0) return Status.INVALID;

        // Get the state hash for signature verification
        bytes32 stateHash = Utils.getStateHash(chan, candidate);

        // Decode application data
        AppData memory appData = abi.decode(candidate.data, (AppData));

        // Check if we have at least one signature (host)
        // For the initial state (with no proofs)
        if (proofs.length == 0 && appData.status == AppStatus.Starting) {
            // Verify Host's signature (first participant)
            if (!Utils.verifySignature(stateHash, candidate.sigs[HOST], chan.participants[HOST])) {
                return Status.INVALID;
            }

            // Initial state is PARTIAL until Guest joins
            return IAdjudicator.Status.PARTIAL;
        }

        // For normal state transitions and final state
        // Check if we have signatures from both participants

        // Verify Host's signature (first participant)
        if (!Utils.verifySignature(stateHash, candidate.sigs[HOST], chan.participants[HOST])) {
            return Status.INVALID;
        }

        // Verify Guest's signature (second participant)
        if (!Utils.verifySignature(stateHash, candidate.sigs[GUEST], chan.participants[GUEST])) {
            return Status.INVALID;
        }

        // If both signatures are valid, check the application status to determine channel status
        // Return ACTIVE if app status is Ready and FINAL if app status is Finish
        if (appData.status == AppStatus.Finish) {
            return IAdjudicator.Status.FINAL;
        } else if (appData.status == AppStatus.Ready) {
            return IAdjudicator.Status.ACTIVE;
        } else {
            // Default to PARTIAL for Starting status
            return IAdjudicator.Status.PARTIAL;
        }
    }
}
