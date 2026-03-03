// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHubTest_Base} from "./ChannelHub_Base.t.sol";

import {Utils} from "../src/Utils.sol";
import {State, ChannelDefinition, StateIntent, Ledger, ChannelStatus} from "../src/interfaces/Types.sol";
import {SessionKeyAuthorization} from "../src/sigValidators/SessionKeyValidator.sol";
import {TestUtils, SESSION_KEY_VALIDATOR_ID} from "./TestUtils.sol";

contract ChannelHubTest_Challenge_NonHomeChain_EscrowDeposit is ChannelHubTest_Base {
    /*
    - escrow deposit can be challenged until `unlockAt` time has NOT passed
    - escrow deposit can not be challenged after `unlockAt` time has passed
    - challenged escrow deposit funds can be withdrawn after `challengeExpireAt` time passes
    - challenged escrow deposit can be resolved until `challengeExpireAt` time has passed with a newer finalization state, which removes challenge and unlock funds
    - challenged escrow deposit can not be resolved if `challengeExpireAt` has passed
    */

    }

contract ChannelHubTest_Challenge_NonHomeChain_EscrowWithdrawal is ChannelHubTest_Base {
    /*
    - escrow withdrawal can be challenged
    - challenged escrow withdrawal funds can be withdrawn after `challengeExpireAt` time passes
    - challenged escrow withdrawal can be resolved until `challengeExpireAt` time has passed with a newer finalization state, which removes challenge and unlock funds
    - challenged escrow withdrawal can not be resolved if `challengeExpireAt` has passed
    */

    }

contract ChannelHubTest_Challenge_NonHomeChain_Migration is ChannelHubTest_Base {
    /*
    - a channel in earlier state can be challenged with initiated migration state
    - a channel in initiated migration state can be challenged with it
    - a channel in earlier state can be challenged with finalize migration state
    - a channel in finalize migration state can be challenged with it
    */

    }
