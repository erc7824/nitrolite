// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature, StateIntent} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";
import {AdjudicatorUtils} from "../adjudicators/AdjudicatorUtils.sol";

/**
 * @title MutualConsent Adjudicator
 * @notice An adjudicator that validates state based on mutual signatures from both participants
 * @dev Any state is considered valid as long as it's signed by both participants
 */
contract Consensus is IAdjudicator {
    using AdjudicatorUtils for State;

    // TODO: replace with constants from Custody
    uint256 constant HOST = 0;
    uint256 constant GUEST = 1;

    /**
     * @notice Validates that the state is signed by both participants
     * @param chan The channel configuration
     * @param candidate The proposed state
     * @param proofs Array of previous states (unused in this implementation)
     * @return valid True if the state is valid, false otherwise
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        pure
        override
        returns (bool valid)
    {
        // NOTE: candidate is never initial state, as this can only happen during challenge or checkpoint, in which case
        // initial state is handled in the protocol layer
        // NOTE: However, initial state can be proofs[0], in which case it should contain signatures from all participants
        // (which can be obtained from blockchain events as all participants are required to join the channel)

        if (proofs.length != 1) {
            return false;
        }

        // proof is Initialize State
        if (candidate.version == 1) {
            return _validateStateTransition(candidate, proofs[0]) &&
                    proofs[0].validateInitialState(chan) &&
                    candidate.validateUnanimousSignatures(chan);
        }

        // proof is Operate or Resize State (both have same validation)
         return _validateStateTransition(candidate, proofs[0]) &&
                proofs[0].validateUnanimousSignatures(chan) &&
                candidate.validateUnanimousSignatures(chan);
    }


    function _validateInitialState(Channel calldata chan, State calldata state) internal pure returns (bool) {
        if (state.version != 0 ||  state.sigs.length != 2) {
            return false;
        }

        if (state.intent != StateIntent.INITIALIZE) {
            return false;
        }

        // Compute the state hash for signature verification.
        bytes32 stateHash = Utils.getStateHash(chan, state);

        return Utils.verifySignature(stateHash, state.sigs[0], chan.participants[HOST])
            && Utils.verifySignature(stateHash, state.sigs[1], chan.participants[GUEST]);
    }

    function _validateStateTransition(State calldata candidate, State calldata previous) internal pure returns (bool) {
        if (candidate.version != previous.version + 1) {
            return false;
        }

        uint256 candidateSum = candidate.allocations[0].amount + candidate.allocations[1].amount;
        uint256 previousSum = previous.allocations[0].amount + previous.allocations[1].amount;

        if (candidateSum != previousSum) {
            return false;
        }

        return true;
    }
}
