// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature, StateIntent} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";
import {AdjudicatorUtils} from "../adjudicators/AdjudicatorUtils.sol";

/**
 * @title Counter Adjudicator
 * @notice Implements a strict turn‐taking counter game.
 * @dev Host sets the initial counter value. After funding the channel, the state is ACTIVE only if counter > 0.
 *      Host and Guest take strict alternating turns to increment the counter.
 *      When the counter reaches the target, the game ends with FINAL status.
 */
contract Counter is IAdjudicator {
    using AdjudicatorUtils for State;

    // TODO: replace with constants from Custody
    uint256 private constant HOST = 0;
    uint256 private constant GUEST = 1;

    /**
     * @dev Data represents the game state.
     * @param target  Target counter value at which the game ends.
     */
    struct Data {
        uint256 target;
    }

    /**
     * @notice Validates that the counter state transition is valid with strict turn‐taking.
     * @param chan The channel configuration.
     * @param candidate The proposed new state.
     * @param proofs Array containing the previous state.
     * @return valid True if the state transition is valid, false otherwise.
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        pure
        override
        returns (bool valid)
    {
        // NOTE: Another reason why Adjudicator cares about "resize" state is when it enters the states chain.
        // NOTE: candidate is never initial state, as this can only happen during challenge or checkpoint, in which case
        // initial state is handled in the protocol layer
        // NOTE: However, initial state can be proofs[0], in which case it should contain signatures from all participants
        // (which can be obtained from blockchain events as all participants are required to join the channel)

        if (proofs.length != 1) {
            return false;
        }

        // for state 1+ validate it does NOT exceed the target
        Data memory candidateData = abi.decode(candidate.data, (Data));
        if (candidate.version > candidateData.target) {
            return false;
        }

        if (candidate.intent != StateIntent.OPERATE) {
            return false;
        }

        // proof is Initialize State
        if (candidate.version == 1) {
            return _validateStateTransition(candidate, proofs[0]) &&
                     proofs[0].validateInitialState(chan) &&
                    _validateStateSig(chan, candidate);
        }

        // proof is Resize State
        if (proofs[0].intent == StateIntent.RESIZE) {
            return _validateStateTransition(candidate, proofs[0]) &&
                 proofs[0].validateUnanimousSignatures(chan) &&
                _validateStateSig(chan, candidate);
        }

        // proof is Operate State
        return _validateStateTransition(candidate, proofs[0]) &&
                _validateStateSig(chan, proofs[0]) &&
                _validateStateSig(chan, candidate);
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

        Data memory candidateData = abi.decode(candidate.data, (Data));
        Data memory previousData;

        // candidate can be based on a RESIZE proof, which contains resize amounts apart from the AppData
        if (candidate.intent == StateIntent.RESIZE) {
           (,previousData) = abi.decode(previous.data, (int256[], Data));
        } else {
            previousData = abi.decode(candidate.data, (Data));
        }

        if (candidateData.target != previousData.target) {
            return false;
        }

        return true;
    }

    function _validateStateSig(Channel calldata chan, State calldata state) internal pure returns (bool) {
        if (state.sigs.length != 1) {
            return false;
        }

        // NOTE: 0th state is unanimously signed, 1st - by host, 2nd - by guest and so on
        uint256 signerIdx = 0; // host signer by default

        if (state.version % 2 == 0) {
            signerIdx = 1; // guest signer
        }

        return Utils.verifySignature(Utils.getStateHash(chan, state), state.sigs[0], chan.participants[signerIdx]);
    }
}
