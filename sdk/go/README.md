# Clearnode Go SDK

Go SDK for Clearnode payment channels providing both high-level and low-level operations in a unified client:
- **High-Level Operations**: `Deposit`, `Withdraw`, `Transfer` with automatic state management
- **Low-Level Operations**: Direct RPC access for custom flows and advanced use cases

## Method Cheat Sheet

### High-Level Operations (Blockchain Interaction)
```go
client.Deposit(ctx, blockchainID, asset, amount)      // Deposit to channel
client.Withdraw(ctx, blockchainID, asset, amount)     // Withdraw from channel
client.Transfer(ctx, recipientWallet, asset, amount)  // Off-chain transfer
```

### Node Information
```go
client.Ping(ctx)                    // Health check
client.GetConfig(ctx)               // Node configuration
client.GetBlockchains(ctx)          // Supported blockchains
client.GetAssets(ctx, blockchainID) // Supported assets
```

### User Queries
```go
client.GetBalances(ctx, wallet)             // User balances
client.GetTransactions(ctx, wallet, opts)   // Transaction history
```

### Channel Queries
```go
client.GetHomeChannel(ctx, wallet, asset)       // Home channel info
client.GetEscrowChannel(ctx, escrowChannelID)   // Escrow channel info
client.GetLatestState(ctx, wallet, asset, onlySigned) // Latest state
```

### App Sessions
```go
client.GetAppSessions(ctx, opts)                              // List sessions
client.GetAppDefinition(ctx, appSessionID)                    // Session definition
client.CreateAppSession(ctx, definition, sessionData, sigs)   // Create session
client.SubmitAppSessionDeposit(ctx, update, sigs, userState)  // Deposit to session
client.SubmitAppState(ctx, update, sigs)                      // Update session
client.RebalanceAppSessions(ctx, signedUpdates)               // Atomic rebalance
```

### Shared Utilities
```go
client.Close()              // Close connection
client.WaitCh()             // Connection monitor channel
client.SignState(state)     // Sign a state (advanced)
client.GetUserAddress()     // Get signer's address
```

## Installation

```bash
go get github.com/erc7824/nitrolite/sdk/go
```

## Quick Start

### Unified Client (High-Level + Low-Level)

```go
package main

import (
    "context"
    "github.com/erc7824/nitrolite/pkg/sign"
    sdk "github.com/erc7824/nitrolite/sdk/go"
    "github.com/shopspring/decimal"
)

func main() {
    // Create signers from private key
    stateSigner, _ := sign.NewEthereumMsgSigner(privateKeyHex)
    txSigner, _ := sign.NewEthereumRawSigner(privateKeyHex)

    // Create unified client
    client, _ := sdk.NewClient(
        "wss://clearnode.example.com/ws",
        stateSigner,
        txSigner,
        sdk.WithBlockchainRPC(80002, "https://polygon-amoy.alchemy.com/v2/KEY"),
    )
    defer client.Close()

    ctx := context.Background()

    // High-level operations - SDK handles everything
    txHash, _ := client.Deposit(ctx, 80002, "usdc", decimal.NewFromInt(100))
    txID, _ := client.Transfer(ctx, "0xRecipient...", "usdc", decimal.NewFromInt(50))
    txHash, _ = client.Withdraw(ctx, 80002, "usdc", decimal.NewFromInt(25))

    // Low-level operations - same client
    config, _ := client.GetConfig(ctx)
    balances, _ := client.GetBalances(ctx, walletAddress)
    state, _ := client.GetLatestState(ctx, wallet, asset, false)
}
```

## Architecture

```
sdk/go/
├── client.go         # Core client, constructors, high-level operations
├── node.go           # Node information methods
├── user.go           # User query methods
├── channel.go        # Channel and state management
├── app_session.go    # App session methods
├── config.go         # Configuration options
└── transform.go      # Type conversions
```

**Design Principles:**
- Zero SDK-specific types - all types from `pkg/core`, `pkg/app`, and `pkg/sign`
- Single unified client with both high-level and low-level methods
- Uses `pkg/sign` for state signing (no SDK-specific signing wrapper)
- Automatic flow management for high-level operations (channel creation, state building, signing)
- Direct RPC access for low-level operations
- Code organized by domain for readability

## Client API

### Creating a Client

```go
// Step 1: Create signers from private key
stateSigner, err := sign.NewEthereumMsgSigner("0x1234...")
if err != nil {
    log.Fatal(err)
}

txSigner, err := sign.NewEthereumRawSigner("0x1234...")
if err != nil {
    log.Fatal(err)
}

// Step 2: Create unified client
client, err := sdk.NewClient(
    wsURL,
    stateSigner,  // For signing channel states
    txSigner,     // For signing blockchain transactions
    sdk.WithBlockchainRPC(chainID, rpcURL), // Required for Deposit/Withdraw
    sdk.WithHandshakeTimeout(10*time.Second),
    sdk.WithPingInterval(5*time.Second),
)
```

### High-Level Operations

#### `Deposit(ctx, blockchainID, asset, amount) (txHash, error)`

Deposits funds into channel. Automatically handles:
- Channel creation if needed
- Checkpointing to existing channel
- State building and signing
- Blockchain transaction

```go
txHash, err := client.Deposit(ctx, 80002, "usdc", decimal.NewFromInt(100))
```

**Requirements:**
- Blockchain RPC configured via `WithBlockchainRPC`
- Token approval for contract address
- Sufficient token balance and gas

#### `Transfer(ctx, recipientWallet, asset, amount) (txID, error)`

Off-chain transfer to another wallet. Instant, no gas required.

```go
txID, err := client.Transfer(ctx, "0xRecipient...", "usdc", decimal.NewFromInt(50))
```

**Requirements:**
- Existing channel with sufficient balance

#### `Withdraw(ctx, blockchainID, asset, amount) (txHash, error)`

Withdraws funds from channel to blockchain wallet.

```go
txHash, err := client.Withdraw(ctx, 80002, "usdc", decimal.NewFromInt(25))
```

**Requirements:**
- Existing channel with sufficient balance
- Blockchain RPC configured
- Sufficient gas for transaction

## Low-Level API

All low-level RPC methods are available on the same Client instance.

### Node Information

```go
err := client.Ping(ctx)
config, err := client.GetConfig(ctx)
blockchains, err := client.GetBlockchains(ctx)
assets, err := client.GetAssets(ctx, &blockchainID) // or nil for all
```

### User Data

```go
balances, err := client.GetBalances(ctx, wallet)
txs, meta, err := client.GetTransactions(ctx, wallet, opts)
```

### Channel Queries

```go
channel, err := client.GetHomeChannel(ctx, wallet, asset)
channel, err := client.GetEscrowChannel(ctx, escrowChannelID)
state, err := client.GetLatestState(ctx, wallet, asset, onlySigned)
```

**Note:** State submission and channel creation are handled internally by high-level operations (Deposit, Withdraw, Transfer).

### App Sessions (Low-Level)

```go
sessions, meta, err := client.GetAppSessions(ctx, opts)
def, err := client.GetAppDefinition(ctx, appSessionID)
sessionID, version, status, err := client.CreateAppSession(ctx, def, data, sigs)
nodeSig, err := client.SubmitAppSessionDeposit(ctx, update, sigs, userState)
err := client.SubmitAppState(ctx, update, sigs)
batchID, err := client.RebalanceAppSessions(ctx, signedUpdates)
```

## Key Concepts

### State Management

Payment channels use versioned states signed by both user and node:

```go
// High-level operations handle state management automatically
client.Deposit(...)   // Creates/updates state, signs, submits
client.Withdraw(...)  // Updates state, signs, submits
client.Transfer(...)  // Updates state, signs, submits
```

**State Flow (Internal):**
1. Get latest state with `GetLatestState()`
2. Create next state with `state.NextState()`
3. Apply transition (deposit, withdraw, transfer, etc.)
4. Calculate state ID with `core.GetStateID()`
5. Sign state with `SignState()`
6. Submit to node (internal method)

State submission and channel creation are handled automatically by high-level operations.

### Signing

States are signed using ECDSA with EIP-155 via `pkg/sign`:

```go
// Create signers from private key
stateSigner, err := sign.NewEthereumMsgSigner(privateKeyHex)  // For channel states
txSigner, err := sign.NewEthereumRawSigner(privateKeyHex)     // For blockchain transactions

// Get address
address := txSigner.PublicKey().Address().String()
```

**Signing Process:**
1. State → ABI Encode (via `core.PackState`)
2. Packed State → Keccak256 Hash
3. Hash → ECDSA Sign (via `signer.Sign`)
4. Result: 65-byte signature (R || S || V)

**Two Signer Types:**
- `EthereumMsgSigner`: Signs channel state updates (off-chain signatures)
- `EthereumRawSigner`: Signs blockchain transactions (on-chain operations)

### Channel Lifecycle

1. **Void**: No channel exists
2. **Create**: Deposit creates channel on-chain
3. **Open**: Channel active, can deposit/withdraw/transfer
4. **Challenged**: Dispute initiated (advanced)
5. **Closed**: Channel finalized (advanced)

## When to Use High-Level vs Low-Level Operations

### Use High-Level Operations When:
- Building user-facing applications
- Need simple deposit/withdraw/transfer
- Want automatic state management
- Don't need custom flows

### Use Low-Level Operations When:
- Building infrastructure/tooling
- Implementing custom state transitions
- Need fine-grained control
- Working with app sessions directly

## Error Handling

All errors include context:

```go
txHash, err := client.Deposit(ctx, 80002, "usdc", amount)
if err != nil {
    // Error: "failed to create channel on blockchain: insufficient balance"
    log.Printf("Error: %v", err)
}
```

Common errors:
- `"channel not created, deposit first"` - Transfer before deposit
- `"blockchain client not configured"` - Missing `WithBlockchainRPC`
- `"insufficient balance"` - Not enough funds in channel/wallet
- `"failed to sign state"` - Invalid private key or state

## Configuration Options

```go
sdk.WithBlockchainRPC(chainID, rpcURL)    // Configure blockchain RPC
sdk.WithHandshakeTimeout(duration)         // Connection timeout (default: 5s)
sdk.WithPingInterval(duration)             // Keepalive interval (default: 5s)
sdk.WithErrorHandler(func(error))          // Connection error handler
```

## Examples

### 1. Interactive CLI Tool (Recommended for Development)

Comprehensive command-line interface with autocomplete, perfect for development and testing.

See [examples/cli/](examples/cli/)

```bash
cd examples/cli
go build -o clearnode-cli
./clearnode-cli wss://clearnode.example.com/ws
```

Features:
- Interactive prompt with autocomplete
- All high-level operations (deposit, withdraw, transfer)
- All low-level operations (states, channels, balances)
- Wallet and RPC management
- Colorful, easy-to-use interface

### 2. Programmatic Example

Basic usage of both high-level and low-level operations.

See [examples/basic/main.go](examples/basic/main.go) for examples of:
- High-level operations (Deposit, Withdraw, Transfer)
- Low-level operations (GetConfig, GetBalances, GetLatestState, etc.)

## Types

All types are imported from `pkg/core` and `pkg/app`:

```go
// Core types
core.State           // Channel state
core.Channel         // Channel info
core.Transition      // State transition
core.Transaction     // Transaction record
core.Asset           // Asset info
core.Token           // Token implementation
core.Blockchain      // Blockchain info

// App session types
app.AppSessionInfoV1      // Session info
app.AppDefinitionV1       // Session definition
app.AppStateUpdateV1      // Session update
```

## High-Level Operation Internals

For understanding how high-level operations work:

### Transfer Flow
1. Get latest state
2. Create next state
3. Apply transfer transition
4. Calculate state ID
5. Sign state
6. Submit to node
7. Return transaction ID

### Deposit Flow (New Channel)
1. Create channel definition
2. Create void state
3. Set home ledger (token, chain)
4. Calculate channel ID
5. Apply deposit transition
6. Sign state
7. Request channel creation from node
8. Create channel on blockchain
9. Return transaction hash

### Deposit Flow (Existing Channel)
1. Get latest state
2. Create next state
3. Apply deposit transition
4. Sign state
5. Submit to node
6. Checkpoint on blockchain
7. Return transaction hash

### Withdraw Flow
1. Get latest state
2. Create next state
3. Apply withdrawal transition
4. Sign state
5. Submit to node
6. Checkpoint on blockchain
7. Return transaction hash

## Requirements

- Go 1.21+
- Running Clearnode instance
- Blockchain RPC endpoint (for Smart Client deposits/withdrawals)

## License

Part of the Nitrolite project.
