// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {IComparable} from "../interfaces/IComparable.sol";
import {Channel, State, Allocation, Signature, Amount} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";

/**
 * @title Remittance Adjudicator
 * @notice An adjudicator that validates payment intent state transfers between participants
 * @dev Validates that allocation transfers are valid from offchain signed stateHash
 */
contract Remittance is IAdjudicator, IComparable {
    uint256 constant CREATOR = 0;
    uint256 constant BROKER = 1;

    /// @notice Error thrown when signature verification fails
    error InvalidSignature();
    /// @notice Error thrown when insufficient signatures are provided
    error InsufficientSignatures();
    /// @notice Error thrown when transfer intent is invalid
    error InvalidTransfer();
    /// @notice Error thrown when allocations do not match
    error InvalidAllocations();

    /**
     * @dev Intent represents a payment transfer from one participant to another
     * @param payer Index of the paying participant (0 for CREATOR, 1 for BROKER)
     * @param transfer Amount and token being transferred
     */
    struct Intent {
        uint8 payer; // Index of the Payer
        Amount transfer; // Amount and token should match the updated allocation
    }

    /**
     * @notice Validates state transitions based on payment intents
     * @param chan The channel configuration
     * @param candidate The proposed state
     * @param proofs Array containing previous states (proofs[0] is funding state, proofs[1] is last valid state if needed)
     * @return valid True if the state transition is valid, false otherwise
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        pure
        override
        returns (bool valid)
    {
        // Check that we have at least one signature
        if (candidate.sigs.length == 0) {
            return false;
        }

        // First state after funding (version == 1) only requires funding proof
        if (candidate.version == 1) {
            // Must have funding proof
            if (proofs.length == 0) {
                return false;
            }

            // Decode the intent
            Intent memory intent = abi.decode(candidate.data, (Intent));

            // Verify payer's signature
            bytes32 stateHash = Utils.getStateHash(chan, candidate);
            if (!Utils.verifySignature(stateHash, candidate.sigs[0], chan.participants[intent.payer])) {
                return false;
            }

            // Check allocations based on funding state (proofs[0])
            // Calculate expected allocations after applying intent
            Allocation[] memory expectedAllocations = new Allocation[](2);

            // Start with the funding allocations
            expectedAllocations[CREATOR] = proofs[0].allocations[CREATOR];
            expectedAllocations[BROKER] = proofs[0].allocations[BROKER];

            // Apply the intent transfer
            if (intent.payer == CREATOR) {
                // CREATOR is paying, reduce CREATOR's amount and increase BROKER's amount
                if (expectedAllocations[CREATOR].amount < intent.transfer.amount) {
                    return false; // Insufficient funds
                }
                expectedAllocations[CREATOR].amount -= intent.transfer.amount;
                expectedAllocations[BROKER].amount += intent.transfer.amount;
            } else {
                // BROKER is paying, reduce BROKER's amount and increase CREATOR's amount
                if (expectedAllocations[BROKER].amount < intent.transfer.amount) {
                    return false; // Insufficient funds
                }
                expectedAllocations[BROKER].amount -= intent.transfer.amount;
                expectedAllocations[CREATOR].amount += intent.transfer.amount;
            }

            // Verify candidate allocations match expected allocations
            if (candidate.allocations.length != 2) {
                return false;
            }

            // Check that tokens and amounts match for both allocations
            if (
                candidate.allocations[CREATOR].token != expectedAllocations[CREATOR].token
                    || candidate.allocations[CREATOR].amount != expectedAllocations[CREATOR].amount
                    || candidate.allocations[BROKER].token != expectedAllocations[BROKER].token
                    || candidate.allocations[BROKER].amount != expectedAllocations[BROKER].amount
            ) {
                return false;
            }

            return true;
        }
        // For states after the first transition (version > 1)
        else if (candidate.version > 1) {
            // Must have at least two proofs: funding state and last valid state
            if (proofs.length < 2) {
                return false;
            }

            // Decode the intent
            Intent memory intent = abi.decode(candidate.data, (Intent));

            // Verify payer's signature
            bytes32 stateHash = Utils.getStateHash(chan, candidate);
            if (!Utils.verifySignature(stateHash, candidate.sigs[0], chan.participants[intent.payer])) {
                return false;
            }

            // Check that version is incremented properly
            if (candidate.version != proofs[1].version + 1) {
                return false;
            }

            // Calculate expected allocations after applying intent to the last valid state
            Allocation[] memory expectedAllocations = new Allocation[](2);

            // Start with the last valid state allocations
            expectedAllocations[CREATOR] = proofs[1].allocations[CREATOR];
            expectedAllocations[BROKER] = proofs[1].allocations[BROKER];

            // Apply the intent transfer
            if (intent.payer == CREATOR) {
                // CREATOR is paying, reduce CREATOR's amount and increase BROKER's amount
                if (expectedAllocations[CREATOR].amount < intent.transfer.amount) {
                    return false; // Insufficient funds
                }
                expectedAllocations[CREATOR].amount -= intent.transfer.amount;
                expectedAllocations[BROKER].amount += intent.transfer.amount;
            } else {
                // BROKER is paying, reduce BROKER's amount and increase CREATOR's amount
                if (expectedAllocations[BROKER].amount < intent.transfer.amount) {
                    return false; // Insufficient funds
                }
                expectedAllocations[BROKER].amount -= intent.transfer.amount;
                expectedAllocations[CREATOR].amount += intent.transfer.amount;
            }

            // Verify candidate allocations match expected allocations
            if (candidate.allocations.length != 2) {
                return false;
            }

            // Check that tokens and amounts match for both allocations
            if (
                candidate.allocations[CREATOR].token != expectedAllocations[CREATOR].token
                    || candidate.allocations[CREATOR].amount != expectedAllocations[CREATOR].amount
                    || candidate.allocations[BROKER].token != expectedAllocations[BROKER].token
                    || candidate.allocations[BROKER].amount != expectedAllocations[BROKER].amount
            ) {
                return false;
            }

            return true;
        }

        // Any other state is invalid
        return false;
    }

    /**
     * @notice Compares two states to determine their relative ordering
     * @param candidate The state being evaluated
     * @param previous The reference state to compare against
     * @return result The comparison result:
     *         -1: candidate < previous (candidate is older)
     *          0: candidate == previous (same recency)
     *          1: candidate > previous (candidate is newer)
     */
    function compare(State calldata candidate, State calldata previous) external pure returns (int8 result) {
        if (candidate.version < previous.version) {
            return -1; // Candidate is older
        } else if (candidate.version > previous.version) {
            return 1; // Candidate is newer
        } else {
            return 0; // Same version
        }
    }
}
