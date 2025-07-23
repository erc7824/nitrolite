// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, StateIntent} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";

/**
 * @title Counter Adjudicator
 * @notice Implements a strict turn‐taking counter game.
 * @dev Host sets the initial counter value. After funding the channel, the state is ACTIVE only if counter > 0.
 *      Host and Guest take strict alternating turns to increment the counter.
 *      When the counter reaches the target, the game ends with FINAL status.
 */
contract Counter is IAdjudicator {
    using Utils for State;

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
            return proofs[0].validateTransitionTo(candidate) && _validateAppTransitionTo(proofs[0].data, candidate.data)
                && proofs[0].validateInitialState(chan, Utils.NO_EIP712_SUPPORT) && _validateStateSig(chan, candidate);
        }

        bytes memory proofData = proofs[0].data;

        if (proofs[0].intent == StateIntent.RESIZE) {
            // NOTE: this approach requires double encoding of Data: `abi.encode(resizeAmounts,abi.encode(Data))`
            (, proofData) = abi.decode(proofs[0].data, (int256[], bytes));
        }

        // proof is Operate or Resize State
        return proofs[0].validateTransitionTo(candidate) && _validateAppTransitionTo(proofData, candidate.data)
            && _validateStateSig(chan, proofs[0]) && _validateStateSig(chan, candidate);
    }

    function _validateAppTransitionTo(bytes memory previousData, bytes memory candidateData)
        internal
        pure
        returns (bool)
    {
        Data memory candidateDataDecoded = abi.decode(candidateData, (Data));
        Data memory previousDataDecoded = abi.decode(previousData, (Data));

        return candidateDataDecoded.target == previousDataDecoded.target;
    }

    function _validateStateSig(Channel calldata chan, State calldata state) internal returns (bool) {
        if (state.sigs.length != 1) {
            return false;
        }

        // NOTE: 0th state is unanimously signed, 1st - by host, 2nd - by guest and so on
        uint256 signerIdx = 0; // host signer by default

        if (state.version % 2 == 0) {
            signerIdx = 1; // guest signer
        }

        return state.verifyStateSignature(
            Utils.getChannelId(chan), Utils.NO_EIP712_SUPPORT, state.sigs[0], chan.participants[signerIdx]
        );
    }
}
