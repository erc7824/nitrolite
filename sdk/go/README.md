# Clearnode Go SDK

Go SDK for Clearnode payment channels with two levels of abstraction:
- **Smart Client**: High-level operations (`Deposit`, `Withdraw`, `Transfer`) with automatic state management
- **Base Client**: Low-level RPC access for custom flows and advanced use cases

## Installation

```bash
go get github.com/erc7824/nitrolite/sdk/go
```

## Quick Start

### Smart Client (Recommended for Most Users)

```go
package main

import (
    "context"
    "github.com/erc7824/nitrolite/pkg/sign"
    sdk "github.com/erc7824/nitrolite/sdk/go"
    "github.com/shopspring/decimal"
)

func main() {
    // Create signer from private key
    signer, _ := sign.NewEthereumSigner(privateKeyHex)

    // Create smart client
    client, _ := sdk.NewSmartClient(
        "wss://clearnode.example.com/ws",
        signer,
        sdk.WithBlockchainRPC(80002, "https://polygon-amoy.alchemy.com/v2/KEY"),
    )
    defer client.Close()

    ctx := context.Background()

    // Simple operations - SDK handles everything
    txHash, _ := client.Deposit(ctx, 80002, "usdc", decimal.NewFromInt(100))
    txID, _ := client.Transfer(ctx, "0xRecipient...", "usdc", decimal.NewFromInt(50))
    txHash, _ = client.Withdraw(ctx, 80002, "usdc", decimal.NewFromInt(25))
}
```

### Base Client (For Low-Level Access)

```go
// Create base client
client, _ := sdk.NewClient("ws://localhost:7824/ws")
defer client.Close()

// Low-level operations
config, _ := client.GetConfig(ctx)
balances, _ := client.GetBalances(ctx, walletAddress)
state, _ := client.GetLatestState(ctx, wallet, asset, false)
```

## Architecture

```
sdk/go/
├── base_client.go    # 18 low-level RPC methods
├── smart_client.go   # High-level: Deposit/Withdraw/Transfer
├── config.go         # Configuration options
└── transform.go      # Type conversions
```

**Design Principles:**
- Zero SDK-specific types - all types from `pkg/core`, `pkg/app`, and `pkg/sign`
- Smart Client embeds Base Client (all methods available)
- Uses `pkg/sign` for state signing (no SDK-specific signing wrapper)
- Automatic flow management (channel creation, state building, signing)

## Smart Client API

### Creating a Client

```go
// Step 1: Create signer from private key
signer, err := sign.NewEthereumSigner("0x1234...")
if err != nil {
    log.Fatal(err)
}

// Step 2: Create smart client
client, err := sdk.NewSmartClient(
    wsURL,
    signer,
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

## Base Client API

All base client methods are available on Smart Client through embedding.

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
channels, meta, err := client.GetChannels(ctx, wallet, opts)
txs, meta, err := client.GetTransactions(ctx, wallet, opts)
```

### State Management (Low-Level)

```go
state, err := client.GetLatestState(ctx, wallet, asset, onlySigned)
states, meta, err := client.GetStates(ctx, wallet, asset, opts)
nodeSig, err := client.SubmitState(ctx, state)
nodeSig, err := client.RequestChannelCreation(ctx, state, channelDef)
```

### App Sessions (Low-Level)

```go
sessions, meta, err := client.GetAppSessions(ctx, opts)
def, err := client.GetAppDefinition(ctx, appSessionID)
sessionID, version, status, err := client.CreateAppSession(ctx, def, data, sigs)
nodeSig, err := client.SubmitDepositState(ctx, update, sigs, userState)
err := client.SubmitAppState(ctx, update, sigs)
batchID, err := client.RebalanceAppSessions(ctx, signedUpdates)
```

## Key Concepts

### State Management

Payment channels use versioned states signed by both user and node:

```go
// Smart Client handles this automatically
client.Deposit(...)  // Creates/updates state, signs, submits

// Base Client requires manual state building
state, _ := client.GetLatestState(ctx, wallet, asset, false)
nextState := state.NextState()
transition, _ := nextState.ApplyHomeDepositTransition(amount)
nextState.ID = core.GetStateID(nextState.UserWallet, nextState.Asset,
                                nextState.Epoch, nextState.Version)
sig, _ := signer.SignState(nextState)
nextState.UserSig = &sig
nodeSig, _ := client.SubmitState(ctx, *nextState)
```

### Signing

States are signed using ECDSA with EIP-155 via `pkg/sign`:

```go
// Create signer from private key
signer, err := sign.NewEthereumSigner(privateKeyHex)

// Get address
address := signer.PublicKey().Address().String()
```

**Signing Process:**
1. State → ABI Encode (via `core.PackState`)
2. Packed State → Keccak256 Hash
3. Hash → ECDSA Sign (via `signer.Sign`)
4. Result: 65-byte signature (R || S || V)

### Channel Lifecycle

1. **Void**: No channel exists
2. **Create**: Deposit creates channel on-chain
3. **Open**: Channel active, can deposit/withdraw/transfer
4. **Challenged**: Dispute initiated (advanced)
5. **Closed**: Channel finalized (advanced)

## When to Use Which Client

### Use Smart Client When:
- Building user-facing applications
- Need simple deposit/withdraw/transfer
- Want automatic state management
- Don't need custom flows

### Use Base Client When:
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

### 2. Smart Client Example

Programmatic usage of high-level operations.

See [examples/smart/main.go](examples/smart/main.go)

```bash
cd examples/smart
PRIVATE_KEY=0x... CLEARNODE_WS=wss://... POLYGON_RPC=https://... go run main.go
```

### 3. Base Client Example

Low-level RPC operations.

See [examples/basic/main.go](examples/basic/main.go)

```bash
cd examples/basic
go run main.go
```

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

## Smart Client Internals

For understanding how operations work:

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
