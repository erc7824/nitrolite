// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature, StateIntent} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";

/**
 * @title MutualConsent Adjudicator
 * @notice An adjudicator that validates state based on mutual signatures from both participants
 * @dev Any state is considered valid as long as it's signed by both participants
 */
contract Consensus is IAdjudicator {
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
        // FIXME: add `resize` handling
        // NOTE: candidate is never initial state, as this can only happen during challenge or checkpoint, in which case
        // initial state is handled in the protocol layer
        // NOTE: However, initial state can be proofs[0], in which case it should contain signatures from all participants
        // (which can be obtained from blockchain events as all participants are required to join the channel)

        if (proofs.length != 1) {
            return false;
        }

        if (candidate.version == 1) {
            return _validateStateTransition(candidate, proofs[0]) &&
                    _validateInitialState(chan, proofs[0]) &&
                    _validateStateSigs(chan, candidate);
        }

        return _validateStateTransition(candidate, proofs[0]) &&
                _validateStateSigs(chan, proofs[0]) &&
                _validateStateSigs(chan, candidate);
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

    function _validateStateSigs(Channel calldata chan, State calldata state) internal pure returns (bool) {
        if (state.sigs.length != 2) {
            return false;
        }

        bytes32 stateHash = Utils.getStateHash(chan, state);

        return Utils.verifySignature(stateHash, state.sigs[HOST], chan.participants[HOST]) &&
                Utils.verifySignature(stateHash, state.sigs[GUEST], chan.participants[GUEST]);
    }
}
