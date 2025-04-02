// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation} from "../interfaces/Types.sol";

contract AND_Adjudicator is IAdjudicator {
    struct AND_State {
        bool isFinal;
        bool flag;
    }

    function adjudicate(Channel calldata, State calldata candidate, State[] calldata proofs)
        public
        pure
        returns (Status)
    {
        if (proofs.length > 1) {
            return Status.INVALID;
        }

        // starting state
        if (proofs.length == 0) {
            return Status.ACTIVE;
        }

        // proofs.length == 1
        AND_State memory candidateState = abi.decode(candidate.data, (AND_State));
        if (candidateState.isFinal) {
            return Status.FINAL;
        }

        State memory proof = proofs[0];
        AND_State memory proofState = abi.decode(proof.data, (AND_State));

        if (candidateState.flag && proofState.flag) {
            return Status.ACTIVE;
        }

        return Status.INVALID;
    }
}
