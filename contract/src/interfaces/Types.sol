// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

/**
 * @title State Channel Type Definitions
 * @notice Shared types used in the state channel system
 */
struct Signature {
    uint8 v;
    bytes32 r;
    bytes32 s;
}

struct Allocation {
    address destination; // Where funds are sent on channel closure
    address token; // ERC-20 token contract address
    uint256 amount; // Token amount allocated
}

struct Channel {
    address[2] participants; // List of participants in the channel [Host, Guest]
    address adjudicator; // Address of the contract that validates final states
    uint64 challenge; // Duration in second, Participants can dispute by submitting newer valid state during challenge
    uint64 nonce; // Unique per channel with same participants and adjudicator
}

struct State {
    bytes data; // Application data encoded, decoded by the adjudicator for business logic
    Allocation[2] allocations; // Combined asset allocation and destination for each participant
    Signature[] sigs; // stateHash signatures
}

enum Status {
    VOID, // Channel was never active (zero-initialized)
    PARTIAL, // Partial funding waiting for other participants
    ACTIVE, // Channel fully funded and valid state
    FINAL, // This is the FINAL state, channel can be closed
    INVALID // Channel state is invalid

}

// This struct has been moved to Custody.sol with additional fields
// Kept here for backward compatibility, but should be migrated to use the Custody.sol version
struct Metadata {
    Channel chan; // Opener define channel configuration
    Status status; // Current channel status
    uint256 challengeExpire; // If non-zero channel will resolve to lastValidState when challenge Expires
    State lastValidState; // Last valid state when adjudicator was called
}
