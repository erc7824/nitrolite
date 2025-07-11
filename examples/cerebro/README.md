# Cerebro

> *"Cerebro is a device that unlocks clearnet's cross-chain features and enables global decentralized networks to operate with one another"* - Just like Professor X's legendary device connects mutants worldwide, Cerebro CLI connects blockchain networks seamlessly.

A powerful Go-based CLI tool for interacting with Clearnode networks and orchestrating cross-chain operations. Originally designed as a simple bridge, Cerebro has evolved into a comprehensive interface for the Clearnode protocol.

## Usage

```bash
go run . <clearnode_ws_url>
```

### Environment Variables

- `CEREBRO_STORE_PATH` - Path to the database file (optional, defaults to `bridge.db`)

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

- **Cross-Platform Configuration**
  - Store local database in OS-appropriate config directories:
    - Linux: `~/.config/cerebro/`
    - macOS: `~/Library/Application Support/cerebro/`
    - Windows: `%APPDATA%\cerebro\`

- **Enhanced Security**
  - Encrypted local database storage
  - Password-protected wallet access

- **Comprehensive Balance Tracking**
  - Show on-chain balances
  - Display custody balances
  - View custody channel balances
  - Unified balance view across Clearnode network
