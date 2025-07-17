# Cerebro

> *"Cerebro is a device that unlocks clearnet's cross-chain features and enables global decentralized networks to operate with one another"* - Just like Professor X's legendary device connects mutants worldwide, Cerebro CLI connects blockchain networks seamlessly.

A powerful Go-based CLI tool for interacting with Clearnode networks and orchestrating cross-chain operations. Originally designed as a simple bridge, Cerebro has evolved into a comprehensive interface for the Clearnode protocol.

## Usage

```bash
go run . <clearnode_ws_url>
```

### Environment Variables

- `CEREBRO_CONFIG_DIR` - Path to the configuration directory (optional, defaults to OS-specific config directory)

## Features

- **Wallet Management**: Import and manage wallets and signers across multiple chains
- **Chain Operations**: Enable/disable chains for cross-chain operations
- **Authentication**: Secure authentication with Clearnode using wallet private keys
- **Asset Bridging**: Seamlessly transfer assets between supported blockchain networks
- **Custody Operations**: Deposit and withdraw assets to/from the custody ledger
- **Channel Resizing**: Dynamically resize payment channels for optimal liquidity management
- **Interactive CLI**: User-friendly command-line interface with intelligent auto-completion
- **Protocol Extensions**: Expandable architecture for future Clearnode protocol features

## Commands

- `import` - Import a wallet, signer or chain RPC URL
- `list` - List available chains, wallets, or signers
- `authenticate` - Authenticate to the Clearnode using your wallet private key
- `enable` - Enable a chain for the current wallet
- `disable` - Disable a chain for the current wallet
- `deposit` - Deposit assets from your wallet to the custody ledger
- `withdraw` - Withdraw assets from the custody ledger to your wallet
- `resize` - Resize payment channels by moving funds from ledger to channel
- `transfer` - Transfer assets to another Clearnode user
- `exit` - Exit the application

## Dependencies

- Go 1.24.3+
- Ethereum client libraries
- SQLite for local storage

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

- **Comprehensive Balance Tracking**
  - Show on-chain balances
  - Display custody balances
  - View custody channel balances
  - Unified balance view across Clearnode network

- **Operator Architecture Refactoring**
  - Refactor `operator.Complete` and `operator.Execute` to use `gin.Router` pattern
  - Treat user commands like HTTP endpoints with middleware support
  - Enable pre/post action hooks for common command patterns

- **Chains and Channels Separation**
  - Separate chain and channel concepts for clearer user experience
  - Replace enable/disable chain with open/close channel operations
  - Update `list chains` command:
    - Remove `enabled` column
    - Add `balance` column showing current user balance of chain asset
  - New `list channels` command with columns: `Chain`, `Asset`, `ChannelID`, `Status`, `Balance`

- **Enhanced Deposit/Withdraw Commands**
  - Clarify deposit/withdraw destination with more intuitive naming like `deposit/withdraw custody`
  - Make it clear which balances are affected by each operation

- **Improved Balance Visibility**
  - `list custodies` showing: `Chain`, `Custody Address`, `Asset`, `Balance`
  - `list unified-balances` showing: `Asset`, `Balance`

- **Resize Command Redesign**
  - Review and simplify the `resize` command
  - Clarify how it affects custody, channel, and unified balances
  - Provide better usage guidance and examples
