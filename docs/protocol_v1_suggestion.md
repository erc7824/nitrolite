# Nitrolite Protocol Documentation - Cross-Chain Architecture

## Overview

The Nitrolite protocol implements a cross-chain state channel system where contracts are deployed on multiple EVM chains. A broker server facilitates users' channel state updates in a coordinated manner across these chains. The protocol enables most operations to happen off-chain, with the ability to submit states on-chain when needed to reflect the latest agreements. The broker acts as a coordinator ensuring correct state progression across the multi-chain ecosystem.

### Key Differences from Current Protocol

This document describes a new cross-chain protocol that differs from the current implementation in several key ways:
- **Participants**: Changed from generic CLIENT/SERVER to specific User/Broker roles
- **State Management**: Replaced intent-based states with transition-based states
- **Fund Management**: Introduced two-tier system (custody ledger + custody liquidity)
- **Cross-chain**: Designed for multi-chain coordination from the ground up
- **Allocations**: Enhanced with available/locked/netValue tracking instead of simple amounts

## Core Concepts

### Channel

A channel is a cryptographic construct that establishes a relationship between a user and the broker for coordinated state management across chains. Each channel consists of:

```solidity
struct Channel {
    address userAddress;     // The user's address
    address brokerAddress;   // The broker's address (facilitator/coordinator)
    address adjudicator;     // Contract that validates state transitions
    uint64 challenge;        // Dispute resolution period in seconds (min 1 hour)
    uint64 nonce;           // Unique identifier for channels with same participants
}
```

**Channel ID**: A unique identifier computed from the channel parameters and chain ID:
```
channelId = keccak256(userAddress, brokerAddress, adjudicator, challenge, nonce, chainId)
```

### Participants

The protocol involves exactly 2 participants:
- **User**: The end user who owns assets and initiates channel operations
- **Broker**: The server entity that facilitates correct state updates across chains

### Channel State

The channel state represents the current agreed-upon status between user and broker:

```solidity
struct State {
    uint256 version;         // Incremental version number
    bytes data;             // Application-specific data
    Allocation[] allocations; // Fund management for each participant (one per token)
    bytes userSignature;     // User's signature on the state
    bytes brokerSignature;   // Broker's signature on the state
    bool is_final;          // True if this is a final state for closing
}
```

#### Allocations

Track fund availability and net balance per token:
```solidity
struct Allocation {
    address token;          // Token address (address(0) for native tokens)
    uint256 available;      // Amount available for use
    uint256 locked;         // Amount locked/reserved
    uint256 netValue;       // Absolute value of (deposits - withdrawals)
    bool isNetPositive;     // True if (deposits - withdrawals) >= 0
}
```

**Multi-asset Support**: The channel supports multiple tokens by having one allocation per token in the allocations array.

The allocation structure maintains:
- **available**: Funds that can be freely used in state transitions
- **locked**: Funds that are reserved/locked for specific purposes
- **netValue + isNetPositive**: Tracks the net flow (deposits minus withdrawals), which can be negative

**Note**: `available + locked` may differ from `netValue` due to cross-chain transfers:
- After sending funds: `available` decreases but `netValue` remains unchanged
- After receiving funds: `available` increases but `netValue` remains unchanged
- Only deposit/withdraw operations affect `netValue`

### Channel Lifecycle States

```solidity
enum ChannelStatus {
    VOID,     // Channel doesn't exist
    ACTIVE,   // Channel is operational
    DISPUTE,  // In challenge period
    FINAL     // Closed (ephemeral state)
}
```

### Fund Management

The protocol maintains two separate fund pools:

1. **Custody Ledger**: User's deposited funds that can be withdrawn directly
2. **Custody Liquidity**: Contract's liquidity pool that changes only during state updates (cannot be withdrawn directly)

#### State Update Rules

For each allocation in a state update, the following rules are enforced:

1. **Net Value Changes**:
   - If `netValueDiff = newNetValue - previousNetValue > 0`: Move `netValueDiff` from custody ledger to custody liquidity
   - If `netValueDiff = newNetValue - previousNetValue < 0`: Move `|netValueDiff|` from custody liquidity to custody ledger

2. **Available Balance Changes**:
   - If `availableDiff = newAvailable - previousAvailable > 0`: Subtract `availableDiff` from custody liquidity
   - If `availableDiff = newAvailable - previousAvailable < 0`: Add `|availableDiff|` to custody liquidity

3. **Locked Balance Changes**:
   - If `lockedDiff = newLocked - previousLocked > 0`: Add `lockedDiff` to custody liquidity
   - If `lockedDiff = newLocked - previousLocked < 0`: Subtract `lockedDiff` from custody liquidity

## Protocol Workflows

### 1. Channel Creation

**Steps:**

```
create(Channel, State) → channelId
```
- Anyone can submit the transaction
- Initial state must have both user and broker signatures
- Channel immediately becomes ACTIVE

**Cross-Chain Architecture**:
- Each channel is unique to its chain (channelId includes chainId)
- Broker coordinates related channels across different chains
- Each chain maintains its own custody liquidity

### 2. State Transitions

Each state update must include transition data explaining the change from the previous state:

#### 2.1 Deposit
- **Data**: `{amount}` for each deposited asset
- **Allocation Changes**: `netValueDiff = availableDiff = amount`
- **Flow**: User submits to broker for validation, then submits on-chain
- **Broker Action**: Blocks newer state updates until this state is checkpointed on-chain

#### 2.2 Withdraw
- **Data**: `{amount}` for each withdrawn asset
- **Allocation Changes**: `netValueDiff = availableDiff = -amount`
- **Flow**: User submits to broker for validation
- **Broker Action**: Signs if valid, user can then withdraw on-chain

#### 2.3 Send
- **Data**: `{chainId, receiver, amount[], transferId}`
  - `amount[]` array for each sent asset
- **Allocation Changes**: `availableDiff = -amount` for each asset
- **Flow**: User submits to broker for validation
- **Broker Action**: Signs if valid, initiates cross-chain coordination

#### 2.4 Receive
- **Data**: `{chainId, sender, amount[], transferId}`
  - `amount[]` array for each received asset
- **Allocation Changes**: `availableDiff = amount` for each asset
- **Flow**: Broker publishes to user
- **User Action**: Acknowledges by providing signature

#### 2.5 Lock
- **Data**: `{amount}` for each locked asset
- **Allocation Changes**: `availableDiff = -amount`, `lockedDiff = amount`
- **Flow**: User submits to broker for validation
- **Broker Action**: Signs if valid

#### 2.6 Unlock
- **Data**: `{amount}` for each unlocked asset
- **Allocation Changes**: `availableDiff = amount`, `lockedDiff = -amount`
- **Flow**: User submits to broker for validation
- **Broker Action**: Signs if valid

### 4. Checkpointing

Store a state on-chain without closing:

```
checkpoint(channelId, State, proofs)
```

**Use Cases:**
- Prevent future disputes
- Clear a challenge without closing
- Update on-chain state record

### 5. Dispute Resolution

When cooperation fails, use the challenge mechanism:

```
challenge(channelId, State, proofs, challengerSig)
```

**Challenge Rules:**
- Must provide challenger signature to prevent griefing
- Starts challenge period (min 1 hour)
- Can be countered with newer valid state
- After expiry, channel closes with last valid state

**Special Cases:**
- States with specific transition types may have different handling

### 6. Cooperative Closing

Mutually agreed closure:

```
close(channelId, State, proofs)
```

**Requirements:**
- State must have `is_final = true`
- Both user and broker must sign
- Funds distributed according to final allocations

### 7. Emergency Closing

After challenge period expires:
```
close(channelId, lastValidState, [])
```

## Contract Operations

The protocol maintains a clear separation between channel lifecycle operations and state updates:

### Channel Lifecycle Operations

1. **open** (replaces `create`):
   - Creates a new channel with initial state
   - Requires both user and broker signatures on initial state
   - Channel immediately becomes ACTIVE

2. **close**:
   - Finalizes the channel
   - Distributes funds according to final allocations
   - Can be cooperative (with agreed final state) or forced (after challenge timeout)

### State Update Operation

**checkpoint**:
- Used for ALL state updates between open and close
- Validates state transition rules
- Updates on-chain state record
- Required transitions: deposit, withdraw, send, receive, lock, unlock

Example flow:
```
open(channel, initialState) → channelId
checkpoint(channelId, depositState)
checkpoint(channelId, sendState)
checkpoint(channelId, receiveState)
close(channelId, finalState)
```

This design ensures:
- Clear operation boundaries
- Consistent state validation
- Single method for all intermediate updates

## Security Features

### Signature Verification

The protocol supports multiple signature types:
- Raw ECDSA
- EIP-191 (Ethereum Signed Message)
- EIP-712 (Structured Data)
- ERC-1271 (Smart Contract Wallets)

### Adjudicator

An external contract that validates state transitions:
- Enforces application-specific rules
- Prevents invalid state progressions
- Can implement custom comparison logic

### Fund Management

**Account Ledger System:**
- Users deposit funds into their account
- Funds locked to channels during operation
- Automatic unlocking on channel closure

**Safety Features:**
- Minimum 1-hour challenge period
- Participant-only challenges (with signature)
- Implicit transfer support for flexibility

## Working with Channels

### Best Practices

1. **State Management**:
   - Always increment version numbers
   - Keep signed states for dispute resolution
   - Verify counterparty signatures before accepting

2. **Fund Safety**:
   - Deposit sufficient funds before creating channels
   - Monitor for challenges
   - Checkpoint important states

3. **Dispute Handling**:
   - Respond to challenges promptly
   - Keep latest signed states available
   - Use checkpoint to prevent disputes

### Integration Guidelines

1. **Opening Channels**:
   ```solidity
   // 1. Deposit funds
   custody.deposit(account, token, amount);
   
   // 2. Open channel with initial state (requires both signatures)
   custody.open(channel, initialState);
   ```

2. **Operating Channels**:
   - Exchange signed states off-chain
   - Validate with adjudicator rules
   - Checkpoint periodically

3. **Closing Channels**:
   - Prefer cooperative closing
   - Have challenge response ready
   - Monitor challenge periods

### Error Handling

Common errors and their meanings:
- `ChannelNotFound`: Invalid channel ID
- `InvalidStatus`: Operation not allowed in current state
- `InvalidStateSignatures`: Missing or invalid signatures
- `InsufficientBalance`: Not enough deposited funds
- `ChallengeNotExpired`: Must wait for challenge period

## Implementation Notes

### Smart Allocations

The protocol uses "smart allocations" that enable:
- Validation of non-consecutive state versions (e.g., version 5 → 10)
- Mathematical verification of state transitions
- Automatic fund rebalancing between ledger and liquidity

### Broker Responsibilities

The broker must:
1. Validate all state transitions before signing
2. Coordinate cross-chain transfers (send/receive pairs)
3. Block new states during deposit checkpoint pending
4. Maintain consistent state across all chains
5. Provide state data to users on request

### Security Considerations

1. **Trust Model**: Similar to current protocol - broker (like SERVER) can block progress but cannot steal funds
2. **Dispute Resolution**: Users can always challenge and force close after timeout
3. **Cross-chain Risks**: Each chain operates independently; broker coordinates but cannot guarantee atomicity
4. **State Validation**: Contract validates allocation math, not business logic

### Future Enhancements

1. **Non-EVM Support**: Update encoding, hashing and signatures for cross-chain compatibility with non-EVM chains
2. **Batch Checkpoint**: Add `batchCheckpoint` method to allow workers to submit multiple channel updates in a single transaction, reducing gas costs
3. **Additional State Transitions**: Easily extensible to support new transition types beyond the current six

## Summary

The new Nitrolite protocol evolves the current implementation to support:
- **Cross-chain operations** through broker coordination
- **Multi-asset channels** with one allocation per token
- **Enhanced fund tracking** with available/locked/netValue
- **Unified state updates** via checkpoint method
- **Flexible transitions** replacing rigid intent system

While maintaining the same security guarantees as the current protocol, it enables new use cases across multiple blockchains.
