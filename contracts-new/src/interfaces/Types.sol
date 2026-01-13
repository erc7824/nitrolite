// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

// ========= Channel Types ==========

struct Definition {
    uint32 challengeDuration;
    address participant;
    address node;
    uint64 nonce;
    bytes32 metadata;
    // to be added later:
    // address executionModule;
    // address signatureValidator;
}

enum ChannelStatus {
    VOID,
    OPERATING,
    DISPUTED,
    CLOSED
}

enum StateIntent {
    OPERATE,
    CREATE,
    CLOSE,
    DEPOSIT,
    WITHDRAW,
    MIGRATE_HOME,
    LOCK,
    UNLOCK
}

struct CrossChainState {
    uint64 version;

    StateIntent intent;

    // to be added for fees logic:
    // bytes data;

    State homeState;
    State nonHomeState;

    bytes participantSig;
    bytes nodeSig;
}

struct State {
    uint64 chainId;
    address token;

    uint256 userAllocation;
    int256 userNetFlow; // can be negative as user can withdraw funds without depositing them (e.g., on a non-home chain)

    uint256 nodeAllocation;
    int256 nodeNetFlow; // can be negative as node can withdraw user funds
}
