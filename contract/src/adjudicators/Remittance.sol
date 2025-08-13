// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {IComparable} from "../interfaces/IComparable.sol";
import {Channel, State, Allocation, Amount, StateIntent} from "../interfaces/Types.sol";
import {EIP712AdjudicatorBase} from "./EIP712AdjudicatorBase.sol";
import {Utils} from "../Utils.sol";

/**
 * @title Remittance Adjudicator
 * @notice An adjudicator that validates payment state transfers requiring only the sender's signature
 * @dev Validates that the participant who is decreasing their allocation has signed the state
 *      This prevents forging "more balance out of thin air" since only the person giving up funds
 *      needs to sign, making it secure for two-party channels with single allocations per party
 */
contract RemittanceAdjudicator is IAdjudicator, IComparable, EIP712AdjudicatorBase {
    using Utils for State;

    uint8 constant CREATOR = 0;
    uint8 constant BROKER = 1;

    /**
     * @dev Remittance represents a payment transfer from one participant to another.
     * @param sender Index of the participant sending funds (0 for CREATOR, 1 for BROKER).
     * @param amount Amount and token being transferred.
     */
    struct Remittance {
        uint8 sender; // Index of the participant sending funds
        Amount amount; // Amount and token being transferred
    }

    /**
     * @notice Constructor for the Remittance adjudicator.
     * @param owner The owner of the adjudicator contract.
     * @param channelImpl The address of the channel implementation contract.
     */
    constructor(address owner, address channelImpl) EIP712AdjudicatorBase(owner, channelImpl) {}

    /**
     * @notice Validates state transitions based on the principle that only the sender needs to sign.
     * @param chan The channel configuration.
     * @param candidate The proposed state.
     * @param proofs Array containing previous states in increasing order up to a starting state, which is either INITIALIZE, RESIZE or state signed by the other party.
     * I.e. if the same party has consequently been signing Remittances, they should supply all their subsequent states, based on a starting state.
     * @return valid True if the state transition is valid, false otherwise.
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        override
        returns (bool valid)
    {
        // Must have at least one proof
        if (proofs.length == 0) {
            return false;
        }

        bytes32 channelImplDomainSeparator = getChannelImplDomainSeparator();

        Remittance memory candidateRemittance = abi.decode(candidate.data, (Remittance));

        // The last proof must be either INITIALIZE, RESIZE, or signed by the other party
        State memory earliestProof = proofs[0];

        // Check if the last proof is a valid starting point
        if (earliestProof.intent == StateIntent.INITIALIZE) {
            if (!earliestProof.validateInitialState(chan, channelImplDomainSeparator)) {
                return false;
            }
        } else if (earliestProof.intent == StateIntent.RESIZE) {
            if (!earliestProof.validateUnanimousStateSignatures(chan, channelImplDomainSeparator)) {
                return false;
            }
            // NOTE: "extract" resize amounts, keep only the RESIZE state data
            // Remittance is encoded in "bytes"
            (, earliestProof.data) = abi.decode(earliestProof.data, (int256[], bytes));
        } else if (earliestProof.intent == StateIntent.OPERATE) {
            // Otherwise, it must be signed by the other party (not the sender in candidate)
            // The last proof must have exactly one signature
            if (earliestProof.sigs.length != 1) {
                return false;
            }

            // Get the other participant who is not the sender in candidate
            uint8 otherParty = candidateRemittance.sender == CREATOR ? BROKER : CREATOR;

            // Verify the other party signed the last proof
            if (
                !earliestProof.verifyStateSignature(
                    Utils.getChannelId(chan),
                    channelImplDomainSeparator,
                    earliestProof.sigs[0],
                    chan.participants[otherParty]
                )
            ) {
                return false;
            }
        } else {
            // Invalid intent
            return false;
        }

        // FIXME: INITIALIZE state does NOT have a Remittance packed
        uint256 proofsLength = proofs.length;

        if (proofsLength == 0) {
            if (!_validateRemittanceState(channelImplDomainSeparator, chan, candidate)) {
                return false;
            }

            return _validateRemittanceTransition(proofs[0], candidate);
        }

        State memory previousState;
        State memory currentState = proofs[0];
        for (uint256 currIdx = 1; currIdx < proofsLength; currIdx++) {
            previousState = currentState;
            currentState = proofs[currIdx];

            if (!_validateRemittanceState(channelImplDomainSeparator, chan, currentState)) {
                return false;
            }

            if (!_validateRemittanceTransition(previousState, currentState)) {
                return false;
            }
        }

        return true;
    }

    function _validateRemittanceState(bytes32 domainSeparator, Channel calldata chan, State memory state)
        internal
        returns (bool)
    {
        if (state.intent != StateIntent.OPERATE) {
            return false;
        }

        if (state.sigs.length != 1) {
            return false;
        }

        Remittance memory currentRemittance = abi.decode(state.data, (Remittance));

        // NOTE: token must be the same
        if (
            currentRemittance.amount.token != state.allocations[0].token
                || currentRemittance.amount.token != state.allocations[1].token
        ) {
            return false;
        }

        // Verify signature is from the sender
        return state.verifyStateSignature(
            Utils.getChannelId(chan), domainSeparator, state.sigs[0], chan.participants[currentRemittance.sender]
        );
    }

    function _validateRemittanceTransition(State memory previousState, State memory currentState)
        internal
        pure
        returns (bool)
    {
        // basic validations: version, allocations sum
        if (!Utils.validateTransitionTo(previousState, currentState)) {
            return false;
        }

        Remittance memory previousRemittance = abi.decode(previousState.data, (Remittance));
        Remittance memory currentRemittance = abi.decode(currentState.data, (Remittance));

        if (currentRemittance.sender != previousRemittance.sender) {
            return false;
        }

        // Verify prev and curr tokens match
        /// @dev prev and curr states are already confirmed to have only one token, so check if this is the same token
        if (currentState.allocations[0].token != previousState.allocations[0].token) {
            return false;
        }

        return _validateRemittanceIsApplied(currentRemittance, previousState.allocations, currentState.allocations);
    }

    function _validateRemittanceIsApplied(
        Remittance memory remittance,
        Allocation[] memory prevAllocations,
        Allocation[] memory currAllocations
    ) internal pure returns (bool) {
        // Verify sender's balance is decreasing
        if (prevAllocations[remittance.sender].amount <= currAllocations[remittance.sender].amount) {
            return false;
        }

        // Verify receiver's balance is increasing by the same amount
        uint8 receiver = remittance.sender == CREATOR ? BROKER : CREATOR;
        uint256 senderDecrease = prevAllocations[remittance.sender].amount - currAllocations[remittance.sender].amount;
        uint256 receiverIncrease = currAllocations[receiver].amount - prevAllocations[receiver].amount;

        if (senderDecrease != receiverIncrease) {
            return false;
        }

        // Verify that the remittance amount matches the actual balance changes
        return remittance.amount.amount == senderDecrease;
    }

    /**
     * @notice Compares two states to determine their relative ordering.
     * @param candidate The state being evaluated.
     * @param previous The reference state to compare against.
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
