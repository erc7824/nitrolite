// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ISignatureValidator} from "./ISignatureValidator.sol";

// ========= Channel Types ==========

struct ChannelDefinition {
    uint32 challengeDuration;
    address user;
    address node;
    uint64 nonce;
    ISignatureValidator signatureValidator;
    bytes32 metadata;
}

enum ChannelStatus {
    VOID,
    OPERATING,
    DISPUTED,
    CLOSED,
    MIGRATING_IN,
    MIGRATED_OUT
}

enum EscrowStatus {
    VOID,
    INITIALIZED,
    DISPUTED,
    FINALIZED
}

enum StateIntent {
    OPERATE,
    CLOSE,
    DEPOSIT,
    WITHDRAW,
    INITIATE_ESCROW_DEPOSIT,
    FINALIZE_ESCROW_DEPOSIT,
    INITIATE_ESCROW_WITHDRAWAL,
    FINALIZE_ESCROW_WITHDRAWAL,
    INITIATE_MIGRATION,
    FINALIZE_MIGRATION
}

/**
 * @notice Signature validator type selector for pluggable signature validation
 * @dev Signature encoding format:
 *      bytes signature = abi.encodePacked(uint8(SigValidatorType), bytes sigBody)
 *
 *      The first byte indicates which validator to use:
 *      - 0x00 (DEFAULT) -> routes to defaultSigValidator
 *      - 0x01 (CHANNEL) -> routes to channel's signatureValidator from definition
 *
 *      ChannelHub logic reads the first byte to determine routing.
 *      The remainder (sigBody) is passed to the selected validator's validateSignature
 *      or validateChallengerSignature method.
 */
enum SigValidatorType {
    DEFAULT,
    CHANNEL
}

struct State {
    uint64 version;
    StateIntent intent;
    bytes32 metadata;

    // to be added for fees logic:
    // bytes data;

    Ledger homeState;
    Ledger nonHomeState;

    bytes userSig;
    bytes nodeSig;
}

struct Ledger {
    uint64 chainId;
    address token;
    uint8 decimals;

    uint256 userAllocation; // FIXME: investigate whether naming the same thing differently in different components is good
    int256 userNetFlow; // can be negative as user can withdraw funds without depositing them (e.g., on a non-home chain)

    uint256 nodeAllocation;
    int256 nodeNetFlow; // can be negative as node can withdraw user funds
}
