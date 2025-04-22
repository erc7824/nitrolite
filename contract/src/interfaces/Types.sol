// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

/**
 * @title State Channel Type Definitions
 * @notice Shared types used in the Nitrolite state channel system
 */

/**
 * @notice Signature structure for digital signatures
 * @dev Used for off-chain signatures verification in the state channel protocol
 */
struct Signature {
    uint8 v; // Recovery ID
    bytes32 r; // R component of the signature
    bytes32 s; // S component of the signature
}

/**
 * @notice Amount structure for token value storage
 * @dev Used to represent a token and its associated amount
 */
struct Amount {
    address token; // ERC-20 token contract address (address(0) for native tokens)
    uint256 amount; // Token amount
}

/**
 * @notice Allocation structure for channel fund distribution
 * @dev Specifies where funds should be sent when a channel is closed
 */
struct Allocation {
    address destination; // Where funds are sent on channel closure
    address token; // ERC-20 token contract address (address(0) for native tokens)
    uint256 amount; // Token amount allocated
}

/**
 * @notice Channel configuration structure
 * @dev Defines the parameters of a state channel
 */
struct Channel {
    address[] participants; // List of participants in the channel
    address adjudicator; // Address of the contract that validates state transitions
    uint64 challenge; // Duration in seconds for dispute resolution period
    uint64 nonce; // Unique per channel with same participants and adjudicator
}

/**
 * @notice State structure for channel state representation
 * @dev Contains application data, asset allocations, and signatures
 */
struct State {
    bytes data; // Application data encoded, decoded by the adjudicator for business logic
    Allocation[] allocations; // Combined asset allocation and destination for each participant
    Signature[] sigs; // stateHash signatures from participants
}

/**
 * @notice Status enum representing the lifecycle of a channel
 * @dev Tracks the current state of a channel
 */
enum Status {
    VOID, // Channel was not created
    INITIAL, // Channel is created and in funding process
    ACTIVE, // Channel fully funded and operational
    DISPUTE, // Challenge period is active
    FINAL // Final state, channel can be closed

}

// Magic numbers for funding protocol
uint32 constant CHANOPEN = 7877; // State.data value for funding stateHash
uint32 constant CHANCLOSE = 7879; // State.data value for closing stateHash
uint32 constant CHANRESIZE = 7883; // State.data value for resize stateHash
