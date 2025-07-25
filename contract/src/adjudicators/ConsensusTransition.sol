// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, StateIntent} from "../interfaces/Types.sol";
import {EIP712AdjudicatorBase} from "./EIP712AdjudicatorBase.sol";
import {Utils} from "../Utils.sol";

/**
 * @title Mutual consent transition Adjudicator
 * @notice An adjudicator that validates state based on mutual signatures from both participants and the transition from the previous state.
 * @dev Any state is considered valid as long as it's signed by both participants and is a valid transition from the previous state.
 */
contract ConsensusTransition is IAdjudicator, EIP712AdjudicatorBase {
    using Utils for State;

    /**
     * @notice Constructor for the ConsensusTransition adjudicator.
     * @param owner The owner of the adjudicator contract.
     * @param channelImpl The address of the channel implementation contract.
     */
    constructor(address owner, address channelImpl) EIP712AdjudicatorBase(owner, channelImpl) {}

    /**
     * @notice Validates that the state is signed by both participants.
     * @param chan The channel configuration.
     * @param candidate The proposed state.
     * @param proofs Array of previous states.
     * @return valid True if the state is valid, false otherwise.
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
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

        bytes32 channelImplDomainSeparator = getChannelImplDomainSeparator();

        // proof is Initialize State
        if (candidate.version == 1) {
            return proofs[0].validateTransitionTo(candidate)
                && proofs[0].validateInitialState(chan, channelImplDomainSeparator)
                && candidate.validateUnanimousStateSignatures(chan, channelImplDomainSeparator);
        }

        // proof is Operate or Resize State (both have same validation)
        return proofs[0].validateTransitionTo(candidate)
            && proofs[0].validateUnanimousStateSignatures(chan, channelImplDomainSeparator)
            && candidate.validateUnanimousStateSignatures(chan, channelImplDomainSeparator);
    }
}
