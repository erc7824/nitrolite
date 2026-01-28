# Clearnode CLI - Developer Tool

> 🚀 Comprehensive command-line interface for the Clearnode Go SDK

A powerful, interactive CLI tool providing access to all Clearnode SDK features - both high-level smart operations and low-level RPC methods. Perfect for developers, testing, and exploring the Clearnode protocol.

## Features

### 🎯 High-Level Operations (Smart Client)
- **deposit** - Automatic channel creation and deposit handling
- **withdraw** - Seamless withdrawal from channels
- **transfer** - Instant off-chain transfers

### 🔧 Low-Level Operations (Base Client)
- **Node queries** - Ping, info, chains, assets
- **User data** - Balances, channels, transactions
- **State management** - View and explore channel states
- **App sessions** - Manage application sessions

### 💡 Smart Features
- **Interactive autocomplete** - Intelligent command suggestions
- **Wallet management** - Secure local storage of private keys
- **RPC configuration** - Multi-chain RPC management
- **Colorful output** - Easy-to-read formatted responses
- **Error handling** - Clear, helpful error messages

## Installation

```bash
cd examples/cli
go build -o clearnode-cli
```

Or install directly:

```bash
go install github.com/erc7824/nitrolite/sdk/go/examples/cli@latest
```

## Quick Start

```bash
# Connect to Clearnode
./clearnode-cli wss://clearnode.example.com/ws

# Setup your wallet (import existing or generate new)
clearnode> import wallet
Choose an option:
  1. Import existing private key
  2. Generate new wallet
Enter choice (1 or 2): 2

# Import blockchain RPC (for deposits/withdrawals)
clearnode> import rpc 80002 https://polygon-amoy.g.alchemy.com/v2/YOUR_KEY

# Check configuration
clearnode> config

# Deposit to channel (creates channel if needed)
clearnode> deposit 80002 usdc 100

# Transfer to another wallet
clearnode> transfer 0xRecipient... usdc 50

# Withdraw from channel
clearnode> withdraw 80002 usdc 25
```

## Commands Reference

### Setup Commands

```bash
help                          # Show detailed help
config                        # Show current configuration
import wallet                 # Setup wallet (import or generate new)
import rpc <chain_id> <url>   # Import blockchain RPC URL
```

**Wallet Setup Options:**

When you run `import wallet`, you can choose between:
1. **Import existing** - Enter your existing private key
2. **Generate new** - Create a brand new wallet with secure random key

**Example:**
```bash
clearnode> import wallet

🔑 Wallet Setup
===============

Choose an option:
  1. Import existing private key
  2. Generate new wallet

Enter choice (1 or 2): 2

🆕 Generate New Wallet

⚠️  IMPORTANT: Save your private key securely!
=====================================
Private Key: 0x1234567890abcdef...
=====================================

Type 'I have saved my private key' to continue: I have saved my private key

✅ Wallet setup completed successfully
📍 Address: 0xYourNewAddress...

💡 Tips:
   - Store your private key in a secure location
   - Never share your private key with anyone
   - Consider using a hardware wallet for large amounts
```

### High-Level Operations (Smart Client)

These commands use the Smart Client for automatic state management:

```bash
deposit <chain_id> <asset> <amount>     # Deposit to channel
withdraw <chain_id> <asset> <amount>    # Withdraw from channel
transfer <recipient> <asset> <amount>   # Transfer to another wallet
```

**Examples:**
```bash
deposit 80002 usdc 100         # Deposit 100 USDC on Polygon Amoy
withdraw 84532 eth 0.1         # Withdraw 0.1 ETH on Base Sepolia
transfer 0x123... usdc 50      # Transfer 50 USDC to address
```

### Node Information (Base Client)

Query node and network information:

```bash
ping                  # Test connection to node
node info             # Get node configuration and version
chains                # List all supported blockchains
assets                # List all supported assets
assets <chain_id>     # List assets on specific chain
```

### User Queries (Base Client)

Query user data:

```bash
balances <wallet>            # Get user balances
channels <wallet>            # List user channels with details
transactions <wallet>        # Get recent transaction history
```

### Low-Level State Management (Base Client)

Advanced state inspection:

```bash
state <wallet> <asset>       # Get latest state for wallet/asset
states <wallet> <asset>      # Get state history (last 10)
```

### Advanced State Management

Interactive state building and submission:

```bash
submit-state                 # Interactively build and submit a state transition
```

This command provides a guided workflow to:
1. Select a wallet and asset
2. View the current state
3. Choose a transition type (transfer, deposit, withdrawal, finalize)
4. Enter transition-specific parameters
5. Preview the new state
6. Sign and submit to the node

**Supported Transitions:**
- **Transfer Send** - Send funds to another wallet (off-chain)
- **Home Deposit** - Record a deposit from blockchain
- **Home Withdrawal** - Record a withdrawal to blockchain
- **Finalize** - Close and finalize the channel
- **Commit** - Commit funds to an app session (requires app state update data and quorum signatures)

### Low-Level App Sessions (Base Client)

Application session management:

```bash
app-sessions                 # List all app sessions
```

### Scenarios (Automation)

Automate sequences of operations:

```bash
scenario <file>              # Execute a scenario from YAML file
scenario-template [file]     # Generate a scenario template
```

Scenarios allow you to define complex workflows as YAML files and execute them automatically. Perfect for:
- Automated testing
- Reproducible demos
- CI/CD integration
- Quick state setup

**Key Features:**
- Variable substitution (`${VAR}`)
- Retry logic for resilience
- Timing controls (wait_before, wait_after)
- Comprehensive assertions for testing
  - Balance checks (exact, >, <)
  - Channel validation (exists, status)
  - State version tracking
  - Transaction count verification
- Error handling options

See `scenarios/README.md` for detailed documentation and examples.

## Configuration

### Environment Variables

- `CLEARNODE_CLI_CONFIG_DIR` - Custom config directory (default: OS config dir)

### Config File Location

- **Linux**: `~/.config/clearnode-cli/config.db`
- **macOS**: `~/Library/Application Support/clearnode-cli/config.db`
- **Windows**: `%APPDATA%\clearnode-cli\config.db`

## Usage Patterns

### First Time Setup

```bash
# 1. Connect
./clearnode-cli wss://clearnode.example.com/ws

# 2. Setup wallet
clearnode> import wallet
# Choose option 1 to import existing key, or option 2 to generate new wallet

# 3. Check available chains
clearnode> chains

# 4. Import RPC for chains you'll use
clearnode> import rpc 80002 https://polygon-amoy.g.alchemy.com/v2/KEY
clearnode> import rpc 84532 https://base-sepolia.g.alchemy.com/v2/KEY

# 5. Verify setup
clearnode> config
```

### Making Your First Deposit

```bash
# Check supported assets
clearnode> assets 80002

# Deposit (will create channel if needed)
clearnode> deposit 80002 usdc 100

# Check your balance
clearnode> balances 0xYourAddress...

# View your channels
clearnode> channels 0xYourAddress...
```

### Transferring Funds

```bash
# Transfer to another user (instant, off-chain)
clearnode> transfer 0xRecipient... usdc 50

# View transaction
clearnode> transactions 0xYourAddress...
```

### Withdrawing Funds

```bash
# Withdraw back to blockchain
clearnode> withdraw 80002 usdc 25

# Transaction will be submitted on-chain
```

## Command Tips

### Autocomplete

The CLI provides intelligent autocomplete:

- Type a command and press `Tab` to see options
- Chain IDs and assets are suggested based on node data
- Wallet addresses from recent commands are suggested

### Navigation

- `Tab` - Show/cycle through suggestions
- `↑/↓` - Navigate command history
- `Ctrl+C` - Exit the CLI
- `Ctrl+D` - (disabled for safety)

### Best Practices

1. **Always import wallet first** - Most commands require authentication
2. **Import RPCs for chains you'll use** - Deposits/withdrawals need blockchain access
3. **Use `config` to verify setup** - Ensure wallet and RPCs are configured
4. **Check `chains` and `assets`** - Know what's supported before transacting
5. **Use autocomplete** - Press Tab to discover available options

## Architecture

### Files

```
cli/
├── main.go         # Entry point, terminal setup
├── operator.go     # Command routing and completion
├── commands.go     # Command implementations
├── storage.go      # SQLite storage for config
├── scenario.go     # Scenario loading and execution
├── scenarios/      # Example scenario files
│   ├── README.md
│   ├── demo_workflow.yaml
│   ├── quick_test.yaml
│   └── balance_verification.yaml
├── go.mod          # Dependencies
└── README.md       # This file
```

### Design

- **Interactive prompt** - Uses `go-prompt` for rich CLI experience
- **Layered approach** - Smart Client for high-level, Base Client for low-level
- **Local storage** - SQLite for secure wallet and RPC storage
- **Context-aware** - All operations use context for timeout/cancellation
- **Error resilient** - Clear error messages with recovery suggestions
- **Automation-ready** - YAML-based scenarios for automated workflows

## Troubleshooting

### "No wallet imported"

```bash
# Setup your wallet first
clearnode> import wallet
# Choose to import existing or generate new
```

### "No RPC configured for chain X"

```bash
# Import RPC for that chain
clearnode> import rpc <chain_id> <rpc_url>
```

### "Failed to connect"

- Check WebSocket URL is correct
- Ensure Clearnode is running and accessible
- Verify network connectivity

### "Deposit failed: insufficient balance"

- Check you have enough tokens in your wallet
- Verify token approval for the contract
- Ensure enough gas (native token)

## Examples

### Complete Workflow

```bash
# Connect to testnet
./clearnode-cli wss://testnet.clearnode.example.com/ws

# Setup
clearnode> import wallet
# (enter private key)
clearnode> import rpc 80002 https://polygon-amoy.g.alchemy.com/v2/KEY

# Check what's available
clearnode> chains
clearnode> assets

# Deposit
clearnode> deposit 80002 usdc 1000

# Check balance
clearnode> balances 0xYourAddress...

# Make transfers
clearnode> transfer 0xAlice... usdc 100
clearnode> transfer 0xBob... usdc 50

# View activity
clearnode> transactions 0xYourAddress...
clearnode> channels 0xYourAddress...

# Withdraw
clearnode> withdraw 80002 usdc 500

# Check final state
clearnode> balances 0xYourAddress...
```

### Automated Workflow with Scenarios

Instead of manual commands, automate everything:

```bash
# Connect
./clearnode-cli wss://testnet.clearnode.example.com/ws

# Setup (still manual)
clearnode> import wallet
clearnode> import rpc 80002 https://polygon-amoy.g.alchemy.com/v2/KEY

# Run entire workflow automatically
clearnode> scenario scenarios/demo_workflow.yaml

# Output:
# 🎬 Starting Scenario: Demo Workflow
# ▶️  Step 1/13: Connection Check
#    ✅ Success
# ▶️  Step 2/13: Initial Deposit
#    🔧 Executing: deposit 80002 usdc 1000
#    ✅ Success
# ... (continues automatically)
# 🎉 Scenario completed!
# 📊 Results: 13 succeeded, 0 failed, 13 total
```

### Development & Testing

```bash
# Explore node capabilities
clearnode> node info
clearnode> chains
clearnode> assets

# Test a user's channels
clearnode> channels 0xTestAddress...
clearnode> state 0xTestAddress... usdc

# View recent activity
clearnode> transactions 0xTestAddress...

# Check app sessions
clearnode> app-sessions
```

### Advanced: Manual State Construction

For advanced users who want full control over state transitions:

```bash
# Start interactive state builder
clearnode> submit-state

# Follow the prompts:
# 1. Enter wallet address (or use your own)
# 2. Enter asset symbol
# 3. Review current state
# 4. Select transition type
# 5. Enter transition parameters
# 6. Review and confirm
# 7. State is signed and submitted

# Example session (Transfer):
clearnode> submit-state
Enter wallet address: 0x1234...
Enter asset symbol: usdc

Current State:
  Version:      5
  User Balance: 100

Select transition type (1-5): 1  # Transfer
Recipient address: 0x5678...
Amount to transfer: 25

Next State to Submit:
  Version:      6
  User Balance: 75

Submit this state? (yes/no): yes
✅ State submitted successfully!

# Example session (Commit to App Session):
clearnode> submit-state
Enter wallet address: (press Enter for your own)
Enter asset symbol: usdc

Current State:
  Version:      8
  User Balance: 200

Select transition type (1-5): 5  # Commit

🎮 Commit to App Session
App Session ID: app_session_abc123
Amount to commit: 50

📋 App State Update Information:
App session version: 2
Allocations (enter participant allocations):
Allocation: 0xParticipant1... 30
Allocation: 0xParticipant2... 20
Allocation: done

✍️  Quorum Signatures:
Signature: 0x1234...abcd
Signature: 0x5678...efgh
Signature: done

Submit this commit state? (yes/no): yes
✅ Commit state submitted successfully!
🎮 App session updated to version: 2
```

### Working with App Sessions (Commit Transitions)

The commit transition allows you to move funds from your home channel into an app session. This is useful for applications like gaming, state channels, or multi-party computations.

**Prerequisites for Commit:**
1. An existing app session (created via API or other means)
2. App session ID
3. New allocations for participants
4. Quorum signatures from participants approving the state update

**Workflow:**

```bash
# First, check available app sessions
clearnode> app-sessions

# Use submit-state to commit funds
clearnode> submit-state

# Enter your wallet and asset
Enter wallet address: (your wallet)
Enter asset symbol: usdc

# Select commit transition
Select transition type (1-5): 5

# Provide commit details
App Session ID: app_abc123...
Amount to commit: 100

# Provide app state update
App session version (new version): 3
Allocations:
  0xPlayer1... 60
  0xPlayer2... 40
  done

Session data (JSON): {"round": 5, "state": "active"}

# Provide quorum signatures
Signature: 0x1234...
Signature: 0x5678...
done

# Review and confirm
Submit this commit state? yes

✅ Commit state submitted successfully!
```

**Important Notes:**
- The sum of allocations must equal the commit amount
- Quorum signatures must be from app session participants
- App session version must increment by 1
- The intent is always "deposit" for commit transitions
- Allocations must include all existing balances (not just the new deposits)

## Security Notes

- Private keys are stored in local SQLite database (unencrypted)
- When generating a new wallet, **save the private key immediately** - it cannot be recovered
- For production use, consider hardware wallets or key management services
- Never commit or share your config database
- Never share your private key with anyone
- Config directory is user-specific and protected by OS permissions
- When providing quorum signatures, ensure they are from trusted participants
- **Backup your private key** - if you lose it, you lose access to your funds

## Contributing

This CLI is part of the Clearnode SDK examples. Contributions welcome!

## License

Part of the Nitrolite project.
