# Yellow Network State Definition

This directory contains the protobuf definitions for Yellow Network's state architecture, supporting multi-chain operations.

## Overview

The Yellow Network implements a unified state management system that enables users to interact across multiple blockchains through a single, coherent state representation. 

It provides users with a unified account, which aggregates their asset holdings across all integrated chains into a single `State`. This architecture enables high-speed, low-cost off-chain transactions, deep liquidity, and a seamless cross-chain user experience.

All state transitions within the network are authorized by users and validated by a dynamic, weighted quorum of **Ledger Nodes**. A master smart contract on Ethereum maintains the network configuration, while on-chain **Custody** and **Adjudicator** contracts on each supported chain handle deposits, withdrawals, and dispute resolution.

## Entities

- **User**: The owner of the funds and the `State`, identified by a unique Yellow Network `UserID`. The user authorizes all actions by signing intents or states with their primary key or a delegated session key.
- **Ledger Nodes (Clearnodes)**: A permissioned set of off-chain nodes responsible for validating user intents, issuing new states, reaching consensus, and attesting to state validity with their weighted signatures. Their registry and weights are managed by the Adjudicator contract.
- **Master Registry Contract (on Ethereum)**: The primary on-chain contract defining the network's configuration, including approved ledger node keys, supported blockchains and the addresses of their respective Custody and Adjudicator contracts.
- **Custody Contract (on each chain)**: Holds the pooled funds for all users on a specific blockchain, processes deposits, and executes withdrawals upon receiving valid proofs.
- **Adjudicator Contract (on each chain)**: A companion to the Custody contract that validates state proofs. It verifies signatures, and resolves disputes by computationally verifying forced state transitions.

## Architecture Components

### Network Configuration

The master smart contract on Ethereum serves as the central registry for network configuration:

- **Registry of supported blockchains**: Defines which blockchain networks are supported
- **Network definitions**: Maps each blockchain to its specific networks and smart contracts  
- **Ledger node registry**: Manages approved validator nodes with signature weights

```protobuf
message NetworkConfig {
  repeated Blockchain blockchains = 1;
  repeated LedgerNode ledger_nodes = 2;
}

message Blockchain {
  string name = 1;
  repeated Network networks = 2;
}

message Network {
  string chain_id = 1;
  string name = 2;
  bytes contract_address = 3;
}

message LedgerNode {
  string node_id = 1;
  SessionKey session_key = 2;
  uint32 signature_weight = 3;
}
```

### Per-Chain Configuration

Each blockchain in its contract:

- Maps Yellow Network asset IDs to local chain tokens
- Maintains a registry of supported tokens
- Mirrors the approved ledger node registry with signature weights for funds recovery

```protobuf
message ChainConfig {
  string chain_id = 1;
  repeated AssetToken asset_tokens = 2;
  repeated LedgerNode ledger_nodes = 3; // Mirrors the main registry (for funds recovery)
}

message AssetToken {
  string asset_id = 1; // Asset ID on Yellow Network
  bytes address = 2;
  uint32 decimals = 3;
}
```

### User Account System

Since Yellow Network supports non-EVM chains, users receive a unique Yellow Network account identifier rather than relying solely on blockchain addresses.

**User Key Management**:

```protobuf
message UserAccount {
  bytes user_id = 1;
  repeated UserKey user_keys = 2;
}

message UserKey {
  bytes user_id = 1;
  string key_type = 2;
  bytes address = 3;
  KeyPermissions permissions = 4;
}

message KeyPermissions {
  repeated string target_chains = 1;
  repeated string allowed_apps = 2;
  repeated string allowed_actions = 3;
}
```

## State Management

### Core State Structure

- **UnsignedState**: Contains user's nonce, data, chain allocations, balances, and active session keys
- **State**: UnsignedState + owner signature + network validator signatures
- **SessionKeys**: Delegated keys with spending limits and expiry times

```protobuf
message State {
  UnsignedState state = 1;
  Signature owner_sig = 2;
  repeated Signature network_sigs = 3;
}

message UnsignedState {
  uint64 nonce = 1;
  bytes data = 2;
  repeated AssetAmount balances = 4;
  repeated SessionKey active_session_keys = 5;
}

message AssetAmount {
  string asset = 1;
  bytes amount = 2;
}

message SessionKey {
  bytes address = 1;
  uint64 nonce = 2;
  SessionKeyPermissions permissions = 3;
}

message SessionKeyPermissions {
  repeated AssetAmount spending_limits = 1;
  uint64 expiry = 2;
}
```

### State Transitions

Users initiate state changes through **Intents**:

1. **Transfer Intent**: User creates and signs a transfer request
2. **Validation**: Network validators verify and sign new states for both sender and receiver (processed through mempool)
3. **Settlement**: Users can submit signed states to custody contracts at any time

### Deposits

The deposit process is automatic:

1. Network validators monitor deposit events from custody contracts
2. Upon detecting a deposit, validators create new `UnsignedState` with:
   - Incremented nonce
   - Updated state balance reflecting deposited funds
3. Validators sign and credit the user's account
4. No separate user signature required - the deposit event serves as authorization

## Transaction Flow

### Transfers

1. User creates `BatchTransferIntent` with current state nonce and allocation details
2. Validators process the intent and create new states for both parties
3. Both sender and receiver receive validator-signed state updates
4. States can be settled on-chain when needed

```protobuf
message BatchTransferIntent {
  UnsignedBatchTransferIntent batch_transfer_intent = 1;
  Signature signature = 2;
}

message UnsignedBatchTransferIntent {
  uint64 state_nonce = 1;
  bytes destination = 2;
  repeated AssetAmount allocations = 3;
}
```

### Withdrawals

Users have two withdrawal options:

#### 1. Cooperative Withdrawal (Immediate)
- Submit withdrawal intent to Yellow Network
- Obtain validator signatures
- Submit signed withdrawal state to custody contract

#### 2. Non-Cooperative Withdrawal (Delayed)
- Submit withdrawal intent without validator signatures
- Triggers cooldown period (similar to challenge period)
- Validators can submit newer valid state during this window

```protobuf
message SignedBatchWithdrawalIntent {
  BatchWithdrawalIntent batch_withdrawal_intent = 1;
  Signature signature = 2;
  repeated Signature network_sigs = 3;
}

message BatchWithdrawalIntent {
  uint64 state_nonce = 1;
  string chain_id = 2;
  bytes destination = 3;
  repeated AssetAmount withdrawals = 4;
}
```


## Mempool Architecture

The Yellow Network operates a distributed mempool system where intents and state transitions are processed before reaching finality:

### Intent Processing Flow

1. **Intent Submission**: Users submit signed intents (transfers, withdrawals) to any Ledger Node
2. **Mempool Entry**: Valid intents enter the network-wide mempool where they await processing
3. **Validation Queue**: Ledger Nodes validate intents in the mempool, checking signatures, nonces, and balances
4. **State Generation**: Upon successful validation, new states are created and added to the mempool
5. **Signature Collection**: States remain in mempool while collecting required validator signatures
6. **Finality Threshold**: Once signature quorum is reached, states achieve finality and are delivered to users
7. **Mempool Cleanup**: Processed intents and finalized states are removed from the mempool

### Mempool Properties

- **Distributed**: All Ledger Nodes maintain synchronized copies of the mempool
- **Ordered**: Intents are processed in nonce order to prevent double-spending
- **Temporary**: Items are removed once processed or expired
- **Fault Tolerant**: Network can continue operating even if some nodes are offline

The mempool serves as a temporary holding area ensuring atomic state transitions and preventing race conditions in the distributed system.


## Adjudication Process

When the network needs to enforce state transitions on-chain, the adjudicator contract:

1. Accepts the last user-signed state (State A)
2. Verifies an array of signed transfer intents as proofs
3. Validates the final validator-signed state (State C)
4. Confirms that provided intents create a valid transition from A to C
5. Ensures validator signatures meet the required quorum threshold

## File Structure

- `state.proto` - Core state definitions and transfer/withdrawal intents
- `signature.proto` - Signature algorithms and structures
  ```protobuf
  enum SignatureAlgorithm {
    SIG_ALG_UNSPECIFIED = 0;
    SIG_ALG_SECP256K1 = 1;
    SIG_ALG_ED25519 = 2;
  }

  message Signature {
    SignatureAlgorithm alg = 1;
    bytes data = 2;
  }
  ```
- `config.proto` - Network and blockchain configuration
- `user.proto` - User account and key management
- `proto/` - Generated Go structs (auto-generated, do not edit)

## Usage

Generate Go structs from protobuf definitions:

```bash
make proto
```

Clean generated files:

```bash
make clean-proto
```

## Key Benefits

- **Multi-chain Support**: Seamless operations across different blockchain networks
- **High Liquidity**: Shared pool approach maximizes fund efficiency
- **Flexible Settlement**: Users choose when to settle on-chain
- **Session Keys**: Delegated authorization with spending limits and expiry
- **Robust Validation**: Quorum-based validator signatures ensure security