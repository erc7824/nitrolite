// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IChannel} from "../interfaces/IChannel.sol";
import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";

contract ChannelMock is IChannel {
    IAdjudicator public adjudicator;
    mapping(bytes32 channelId => Channel ch) public channels;

    constructor(IAdjudicator _adjudicator) {
        adjudicator = _adjudicator;
    }

    function open(Channel calldata ch, State calldata deposit) public returns (bytes32 channelId) {
        State[] memory emptyProofs = new State[](0);
        IAdjudicator.Status status = adjudicator.adjudicate(ch, deposit, emptyProofs);

        if (status == IAdjudicator.Status.ACTIVE) {
            channels[channelId] = ch;
            emit ChannelOpened(channelId, ch);
        } else {
            revert("Invalid Channel State");
        }

        return Utils.getChannelId(ch);
    }

    function close(bytes32 channelId, State calldata candidate, State[] calldata proofs) public {
        Channel memory ch = channels[channelId];
        IAdjudicator.Status status = adjudicator.adjudicate(ch, candidate, proofs);

        if (status == IAdjudicator.Status.FINAL) {
            emit ChannelClosed(channelId);
        } else {
            revert("Invalid Channel State");
        }
    }

    function reset(
        bytes32 channelId,
        State calldata candidate,
        State[] calldata proofs,
        Channel calldata newCh,
        State calldata newDeposit
    ) external {
        open(newCh, newDeposit);
        close(channelId, candidate, proofs);
    }

    function challenge(bytes32 channelId, State calldata candidate, State[] calldata proofs) external {
        Channel memory ch = channels[channelId];
        IAdjudicator.Status status = adjudicator.adjudicate(ch, candidate, proofs);

        if (status == IAdjudicator.Status.ACTIVE) {
            emit ChannelChallenged(channelId, block.timestamp + ch.challenge);
        } else {
            revert("Invalid Channel State");
        }
    }

    function checkpoint(bytes32 channelId, State calldata candidate, State[] calldata proofs) external {
        Channel memory ch = channels[channelId];
        IAdjudicator.Status status = adjudicator.adjudicate(ch, candidate, proofs);

        if (status == IAdjudicator.Status.ACTIVE) {
            emit ChannelCheckpointed(channelId);
        } else {
            revert("Invalid Channel State");
        }
    }

    function reclaim(bytes32 channelId) external {
        emit ChannelClosed(channelId);
    }
}
