// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, StateIntent} from "../interfaces/Types.sol";
import {EIP712AdjudicatorBase} from "./EIP712AdjudicatorBase.sol";
import {Utils} from "../Utils.sol";

/**
 * @title Simple mutual consent Adjudicator
 * @notice An adjudicator that validates state based on mutual signatures from both participants.
 * @dev Any state is considered valid as long as it's signed by both participants.
 */
contract SimpleConsensus is IAdjudicator, EIP712AdjudicatorBase {
    using Utils for State;

    /**
     * @notice Constructor for the SimpleConsensus adjudicator.
     * @param owner The owner of the adjudicator contract.
     * @param channelImpl The address of the channel implementation contract.
     */
    constructor(address owner, address channelImpl) EIP712AdjudicatorBase(owner, channelImpl) {}

    /**
     * @notice Validates that the state is signed by both participants.
     * @param chan The channel configuration.
     * @param candidate The proposed state.
     * @param proofs Array of previous states (unused in this implementation).
     * @return valid True if the state is valid, false otherwise.
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        override
        returns (bool valid)
    {
        if (proofs.length != 0) {
            return false;
        }

        bytes32 channelImplDomainSeparator = getChannelImplDomainSeparator();

        if (candidate.version == 0) {
            return candidate.validateInitialState(chan, channelImplDomainSeparator);
        }

        // proof is Operate or Resize State (both have same validation)
        return candidate.intent != StateIntent.INITIALIZE
            && candidate.validateUnanimousStateSignatures(chan, channelImplDomainSeparator);
    }
}
