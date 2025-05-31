# Clearnode Testing Tools

This directory contains tools for testing the Clearnode API by creating properly signed RPC messages and interacting with the Clearnode server.

The tool automatically handles authentication, private key management, and message signing according to the Clearnode protocol.

## Quick Start

```bash
# Generate a signer key
go run . --genkey 1    # Creates signer_key_1.hex

# Generate a signed message only
go run . --method ping

# Send the message to the server and get a response
go run . --method ping --send --server ws://localhost:8000/ws

# TESTNET: wss://canarynet.yellow.com/ws
# PROD: wss://clearnet.yellow.com/ws

# Or set the server URL via environment variable
export SERVER=wss://canarynet.yellow.com/ws
go run . --method ping --send
```

## Private Key Management

### Key Generation and Storage

- All signers are treated equally in the system
- Keys are stored as `signer_key_N.hex` files (where N is the signer number)
- To generate signer keys, use the `--genkey` parameter with a number:

```bash
# Generate signer keys
go run . --genkey 1    # Creates signer_key_1.hex
go run . --genkey 2    # Creates signer_key_2.hex
go run . --genkey 3    # Creates signer_key_3.hex
```

These commands will display:
- The file path where the key is stored
- The Ethereum address associated with the key
- The private key in hex format (for importing into wallets)

### Multiple Signers Support

The client supports using multiple signers for request signing:

- When running commands, the client automatically detects and uses all signer keys in the current directory
- Messages will be signed by all available signers by default
- You can specify which signers to use with the `--signers` flag:
  ```bash
  # Use only signer #1
  go run . --method ping --signers 1 --send
  
  # Use signers #1 and #3
  go run . --method ping --signers 1,3 --send
  
  # Use only signers #2 and #4
  go run . --method ping --signers 2,4 --send
  ```
- Signatures are collected in the `sig` field of the RPC message
- By default, the first signer in the list is used as the address for authentication, but you can specify a different signer with the `--auth` flag

### Using Your Key with MetaMask

To import your testing key into MetaMask:

1. Open `signer_key_N.hex` to view your current key
2. In MetaMask, click Account → Import Account
3. Select "Private Key" and paste with "0x" prefix: `0x<key-from-output>`

**Security Note**: Use dedicated keys for testing and avoid storing significant funds on them.

## Command Line Usage

### Basic Syntax

```bash
go run . --method <method_name> [options]
```

### Common Options

| Option | Description | Default |
|--------|-------------|---------|
| `--method` | RPC method name (required) | |
| `--id` | Request ID | 1 |
| `--params` | JSON array of parameters | `[]` |
| `--send` | Send to server (omit to only create signed message) | false |
| `--server` | WebSocket server URL | ws://localhost:8000/ws |
| `--genkey` | Generate a new key and exit. Use a signer number (e.g., '1', '2', '3') | "" |
| `--signers` | Comma-separated list of signer numbers to use (e.g., "1,2,3") | All available |
| `--auth` | Specify which signer to authenticate with (e.g., "1") | First signer |
| `--nosign` | Make a request without signatures | false |
| `--noauth` | Skip authentication (for public endpoints) | false |

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER` | WebSocket server URL (overrides default but is overridden by `--server` flag) | ws://localhost:8000/ws |

## Common Test Scenarios

### Authentication & Signature Options

```bash
# Authenticate with a specific signer
go run . --method ping --send --auth 2

# Send a request with no signatures (useful for testing)
go run . --method ping --send --nosign

# Skip authentication completely (for public endpoints)
go run . --method ping --send --noauth

# Authenticate with signer #2 but sign the message with signers #1 and #3
go run . --method ping --send --auth 2 --signers 1,3
```

### Public Endpoints (No Authentication Required)

These endpoints can be accessed without authentication by using the `--noauth` flag:

```bash
# Simple connectivity check
go run . --method ping --send --noauth --server wss://canarynet.yellow.com/ws

# Get server configuration and supported networks
go run . --method get_config --send --noauth --server wss://canarynet.yellow.com/ws

# List all supported assets
go run . --method get_assets --send --noauth --server wss://canarynet.yellow.com/ws

# Get assets for a specific blockchain
go run . --method get_assets --params '[{"chain_id":137}]' --send --noauth --server wss://canarynet.yellow.com/ws

# Get application definition for a specific app session
go run . --method get_app_definition --params '[{"app_session_id":"0xAppSessionID"}]' --send --noauth --server wss://canarynet.yellow.com/ws

# List virtual applications for a participant
go run . --method get_app_sessions  --send --noauth --server wss://canarynet.yellow.com/ws

# Get channels for a participant
go run . --method get_channels --send --noauth --server wss://canarynet.yellow.com/ws
```

### Authenticated Endpoints

These endpoints require authentication:

```bash
# Get RPC message history
go run . --method get_rpc_history --send --server ws://localhost:8000/ws

# Create a virtual app session between two participants
go run . --method create_app_session --params '[{
  "definition": {
    "protocol": "NitroRPC/0.2",
    "participants": ["0x70997970C51812dc3A010C7d01b50e0d17dc79C8", "0xF13EFAA6aEc3301Def39A23B8B15a1a43140191F"],
    "weights": [100, 0],
    "quorum": 100,
    "challenge": 86400,
    "nonce": 1
  },
  "allocations": [
    {"participant": "0x70997970C51812dc3A010C7d01b50e0d17dc79C8", "asset": "usd", "amount": "0.000000000000000004"},
    {"participant": "0xF13EFAA6aEc3301Def39A23B8B15a1a43140191F", "asset": "usd", "amount": "0.0"}
  ]
}]' --auth 10 --signers 10 --send --server ws://localhost:8000/ws

# Close an app session
go run . --method close_app_session --params '[{
  "app_session_id": "0x2b1843390eef1ed7406826b01fa95135e71ea2266222761ffa2efeaad6b81f84",
  "allocations": [
    {"participant": "0x70997970C51812dc3A010C7d01b50e0d17dc79C8", "asset": "usd", "amount": "0.0"},
    {"participant": "0xF13EFAA6aEc3301Def39A23B8B15a1a43140191F", "asset": "usd", "amount": "0.000000000000000004"}
  ]
}]' --auth 10 --signers 10 --send --server ws://localhost:8000/ws

# Resize a channel (increase or decrease allocation)
go run . --method resize_channel --params '[{
  "channel_id": "0xYourChannelID",
  "resize_amount": "10.0",
  "allocate_amount": "0.0",
  "funds_destination": "0xYourAddress"
}]' --send --server ws://localhost:8000/ws

# Close a channel
go run . --method close_channel --params '[{
  "channel_id": "0xYourChannelID",
  "funds_destination": "0xYourAddress"
}]' --send --server ws://localhost:8000/ws
```

## Interactive Testing Script

The `test_api.sh` script provides a menu-driven interface for common operations:

- Generate/manage keys
- Test server connectivity
- Query balances and ledger entries
- Manage channels and app sessions
- List supported assets
- Run custom commands

## Troubleshooting

- **Connection Errors**: Verify the server URL with `--server` parameter
- **Authentication Errors**: The tool handles auth automatically, but requires correct server URL
- **Parameter Format Errors**: Ensure JSON parameters follow the required format
- **Key-Related Issues**: Delete `signer_key.hex` to generate a new key automatically

## Technical Details

- The client uses WebSocket connections to communicate with the Clearnode server
- Messages are signed using ECDSA with the Ethereum secp256k1 curve
- Authentication follows the challenge-response pattern required by Clearnode
- The testing tools automatically handle the entire authentication flow
- Multi-signature support allows testing scenarios where messages must be signed by multiple parties
- The `sig` field in RPC messages contains an array of signatures from all signers
- The `--nosign` flag allows sending requests without signatures, useful for testing
- The `--noauth` flag lets you skip authentication completely for public endpoints
- The `--auth` flag lets you specify which signer to use for authentication when using multiple signers