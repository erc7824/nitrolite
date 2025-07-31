// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

/**
 * @title State Channel Type Definitions
 * @notice Shared types used in the Nitrolite state channel system
 */

/// @dev EIP-712 domain separator type hash for state channel protocol
bytes32 constant STATE_TYPEHASH = keccak256(
    "AllowStateHash(bytes32 channelId,uint8 intent,uint256 version,bytes data,Allocation[] allocations)Allocation(address destination,address token,uint256 amount)"
);

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
 * @notice Status enum representing the lifecycle of a channel
 * @dev Tracks the current state of a channel
 */
enum ChannelStatus {
    VOID, // Channel was not created, State.version must be 0
    INITIAL, // Channel is created and in funding process, State.version must be 0
    ACTIVE, // Channel fully funded and operational, State.version is greater than 0
    DISPUTE, // Challenge period is active
    FINAL // Final state, channel can be closed

}

/**
 * @notice Intent enum representing the purpose of a state
 * @dev Used to indicate the action to be taken with the state
 */
enum StateIntent {
    OPERATE, // Operate is the app state
    INITIALIZE, // Initial funding state
    RESIZE, // Resize state
    FINALIZE // Final closing state

}

/**
 * @notice State structure for channel state representation
 * @dev Contains application data, asset allocations, and signatures
 */
struct State {
    StateIntent intent; // Intent of the state
    uint256 version; // State version incremental number to compare most recent
    bytes data; // Application data encoded, decoded by the adjudicator for business logic
    Allocation[] allocations; // Combined asset allocation and destination for each participant
    bytes[] sigs; // stateHash signatures from participants
}

/**
 * @notice App structure for virtual ledger layer
 * @dev vApp definition for off-chain quorum. Is used off-chain
 */
struct App {
    string protocol; // String protocol/version "NitroRPC/0.2"
    address[] participants; // Array of participants in the app
    uint8[] weights; // Signers weights for this app [50, 50, 80, 20, 20]
    uint64 quorum; // Example value 100 would be the signature threshold
    uint64 challenge; // Duration in seconds for dispute resolution period
    uint64 nonce; // Unique per channel with same participants and adjudicator
}

/**
 * @notice Allowance structure for policy asset permissions
 * @dev Defines allowed asset amounts for session key operations
 */
struct Allowance {
    string asset; // Asset identifier (e.g., "usdc")
    uint256 amount; // Maximum allowed amount
}

/**
 * @notice Policy structure for EIP-712 session key authorization and registration
 * @dev Defines permissions and constraints for delegated operations
 */
struct Policy {
    string challenge; // Unique challenge identifier (UUID format)
    string scope; // Permission scope (e.g., "app.create", "ledger.readonly")
    address wallet; // Main wallet address authorizing the session
    address application; // Application public address
    address participant; // Delegated session key address
    uint256 expire; // Expiration timestamp
    Allowance[] allowances; // Array of asset allowances
}
