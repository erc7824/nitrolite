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

struct Amount {
  address token; // ERC-20 token contract address
  uint256 amount; // Token amount allocated
}

struct Allocation {
    address destination; // Where funds are sent on channel closure
    address token; // ERC-20 token contract address
    uint256 amount; // Token amount allocated
    //TODO: consider using Amount
}

struct Channel {
    address[] participants; // List of participants in the channel [Host, Guest]
    address adjudicator; // Address of the contract that validates final states
    uint64 challenge; // Duration in second, Participants can dispute by submitting newer valid state during challenge
    uint64 nonce; // Unique per channel with same participants and adjudicator
}

struct State {
    bytes data; // Application data encoded, decoded by the adjudicator for business logic
    Allocation[] allocations; // Combined asset allocation and destination for each participant
    Signature[] sigs; // stateHash signatures
}

enum Status {
    VOID, // VOID Channel was not created
    INITIAL, // Channel is created and in funding process
    ACTIVE, // Channel fully funded and valid state
    DISPUTE,
    FINAL, // This is the FINAL state, channel can be closed
}

constant uint256 CREATOR = 0; // participant index for the channel creator
constant uint256 BROKER = 1; // participant index for the broker

// Funding protocol use State with expected deposits
constant uint32 CHANOPEN = 7877; // State.data value for funding stateHash
constant uint32 CHANCLOSE = 7879; // State.data value for closing stateHash

// Reference storage type for IChannel implementation to be cdustomized
struct Metadata {
    Channel chan; // Opener define channel configuration
    Status stage;
    address creator;
    Amount[] expectedDeposits; // Creator defines Token per participant
    Amount[] actualDeposits; // Tracks deposits made by each participant
    uint256 challengeExpire; // If non-zero channel will resolve to lastValidState when challenge Expires
    State lastValidState; // Last valid state when adjudicator was called
}
