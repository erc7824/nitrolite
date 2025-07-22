# Cerebro

> *"Cerebro is a device that unlocks clearnet's cross-chain features and enables global decentralized networks to operate with one another"* - Just like Professor X's legendary device connects mutants worldwide, Cerebro CLI connects blockchain networks seamlessly.

A powerful Go-based CLI tool for interacting with Clearnode networks and orchestrating cross-chain operations. Originally designed as a simple bridge, Cerebro has evolved into a comprehensive interface for the Clearnode protocol.

## Installation

```bash
go install github.com/erc7824/nitrolite/examples/cerebro@latest
```

## Usage

```bash
cerebro <clearnode_ws_url>
```

### Environment Variables

- `CEREBRO_CONFIG_DIR` - Path to the configuration directory (optional, defaults to OS-specific config directory)

## Features

- **Wallet Management**: Import and manage wallets and signers across multiple chains
- **Authentication**: Secure authentication with Clearnode using wallet private keys and signers
- **Custody Operations**: Deposit and withdraw assets to/from the custody ledger
- **Payment Channels**: 
  - Open channels for specific assets on supported chains
  - Close channels to withdraw locked funds
  - Resize channels to adjust liquidity between custody and channel balances
- **Real-time Events**: WebSocket-based event handling for live updates on assets, channels, and balances
- **Balance Tracking**: View custody, channel, and unified balances across the Clearnode network
- **Interactive CLI**: User-friendly command-line interface with intelligent auto-completion
- **Protocol Extensions**: Expandable architecture for future Clearnode protocol features

## Commands

- `import` - Import a wallet, signer or chain RPC URL
- `list` - List available chains, wallets, signers, or channels
- `authenticate` - Authenticate to the Clearnode using your wallet private key and signer
- `deposit custody` - Deposit assets from your wallet to the custody ledger
- `withdraw custody` - Withdraw assets from the custody ledger to your wallet
- `open channel` - Open a payment channel for a specific asset on a chain
- `close channel` - Close a payment channel and unlock funds
- `resize channel` - Resize payment channels by adjusting allocations between custody ledger, channel, and unified balance
- `transfer` - Transfer assets to another Clearnode user
- `exit` - Exit the application

## Channel Operations Workflow

1. **Open a Channel**
   ```bash
   open channel <chain_name> <token_symbol>
   ```
   Creates a new payment channel for the specified asset on the given chain.

2. **Resize a Channel**
   ```bash
   resize channel <chain_name> <token_symbol>
   ```
   Adjusts channel allocations by:
   - Moving funds from custody ledger to channel (resize amount)
   - Moving funds from unified balance to channel (allocate amount)
   - Displays current balances before operation for clarity

3. **Close a Channel**
   ```bash
   close channel <chain_name> <token_symbol>
   ```
   Requests channel closure and withdraws unlocked funds back to custody ledgers.

## Dependencies

- Go 1.24.3+
- go-ethereum for Ethereum chain interactions
- SQLite for local storage
- WebSocket support for real-time events

## Roadmap

### Planned Features

- **Enhanced Logging**
  - Proper logger/printer with error redirection to stderr
  - Structured logging with different verbosity levels

- **Improved User Experience**
  - Comprehensive `help` command with detailed documentation for each command
  - Visual demonstration GIF in README showing CLI in action

- **Extended Import/Export Functionality**
  - Export capabilities for all importable items (wallets, signers, chain configurations)
  - Remove functionality for imported items with proper cleanup

- **Simplified Bridge Operations**
  - Single `bridge` command that executes the full bridging flow:
    1. Deposit to custody
    2. Resize channel (+1)
    3. Resize channel (-1)
    4. Withdraw from custody

- **Enhanced Security**
  - Encrypted local database storage
  - Password-protected wallet access

- **Operator Architecture Refactoring**
  - Refactor `operator.Complete` and `operator.Execute` to use `gin.Router` pattern
  - Treat user commands like HTTP endpoints with middleware support
  - Enable pre/post action hooks for common command patterns

- **Improved Balance Visibility**
  - `list custodies` showing: `Chain`, `Custody Address`, `Asset`, `Balance`
  - `list unified-balances` showing: `Asset`, `Balance`
