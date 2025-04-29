// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";
import {Channel, State, Signature, StateIntent} from "../interfaces/Types.sol";
import {CREATOR, BROKER} from "../Custody.sol";
import {Utils} from "../Utils.sol";

library AdjudicatorUtils {
    function validateInitialState(State calldata state, Channel calldata chan) internal pure returns (bool) {
        if (state.version != 0) {
            return false;
        }

        if (state.intent != StateIntent.INITIALIZE) {
            return false;
        }

        return validateUnanimousSignatures(state, chan);
    }

    function validateUnanimousSignatures(State calldata state, Channel calldata chan) internal pure returns (bool) {
        if (state.sigs.length != 2) {
            return false;
        }

        // Compute the state hash for signature verification.
        bytes32 stateHash = Utils.getStateHash(chan, state);

        return Utils.verifySignature(stateHash, state.sigs[0], chan.participants[CREATOR])
            && Utils.verifySignature(stateHash, state.sigs[1], chan.participants[BROKER]);
    }
}
