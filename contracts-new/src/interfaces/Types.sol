// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

// ========= Channel Types ==========

struct Definition {
    uint32 challengeDuration;
    address participant;
    address node;
    uint256 nonce;
    // to be added later:
    // address executionModule;
    // address signatureValidator;
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
    uint256 version;

    uint64 homeChainId;
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

    Balance participantBalance;
    Balance nodeBalance;
}

struct Balance {
    uint256 allocation;
    int256 netFlow;
}
