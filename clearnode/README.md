# Clearnode

[![codecov](https://codecov.io/github/erc7824/nitrolite/graph/badge.svg)](https://codecov.io/github/erc7824/nitrolite)
[![Go Reference](https://pkg.go.dev/badge/github.com/erc7824/nitrolite/clearnode.svg)](https://pkg.go.dev/github.com/erc7824/nitrolite/clearnode)
[![Go Report Card](https://goreportcard.com/badge/github.com/erc7824/nitrolite/clearnode)](https://goreportcard.com/report/github.com/erc7824/nitrolite/clearnode)

Clearnode is the off-chain node implementation for the Nitrolite V1 protocol. It manages state channels, processes off-chain transactions, and coordinates state-channel state updates between users and to enable fast, low-cost payment channels and application sessions.

## Overview

Clearnode provides a WebSocket-based RPC service that allows users to:
- Create and manage home and escrow channels on multiple blockchains
- Perform instant off-chain transfers between users
- Run application sessions with multiple participants
- Deposit and withdraw assets across chains
- Track balances and transaction history

The node listens to blockchain events, validates state transitions, and ensures secure coordination between on-chain channels and off-chain state.

## Architecture

Clearnode consists of several key components:

- **RPC Server**: WebSocket server (`:7824`) handling client requests
- **Blockchain Listeners**: Monitor on-chain events from nitrolite contracts
- **Event Handlers**: Process blockchain events and update internal state
- **Database Store**: Persistent storage for channels, states, and transactions
- **Memory Store**: In-memory configuration for blockchains and assets
- **Blockchain Workers**: Coordinate on-chain operations (future use)
- **Metrics Server**: Prometheus metrics endpoint (`:4242`)

### API Groups

Clearnode exposes four main API groups via WebSocket RPC:

1. **channel_v1**: Channel creation and state management
   - `get_home_channel` - Retrieve home channel information
   - `get_escrow_channel` - Retrieve escrow channel information
   - `get_latest_state` - Get latest user state
   - `request_creation` - Request channel creation signature
   - `submit_state` - Submit signed state update

2. **app_session_v1**: Application session management
   - `create_app_session` - Create new application session
   - `get_app_definition` - Get application session definition
   - `get_app_sessions` - List user's application sessions
   - `submit_deposit_state` - Submit deposit to app session
   - `submit_app_state` - Submit app session state update
   - `rebalance_app_sessions` - Rebalance across app sessions

3. **user_v1**: User account queries
   - `get_balances` - Retrieve user balances by asset
   - `get_transactions` - Get transaction history

4. **node_v1**: Node information
   - `get_config` - Get node configuration
   - `get_assets` - List supported assets

For detailed API specifications, see [../docs/api.yaml](../docs/api.yaml).

## Configuration

Clearnode uses YAML configuration files for blockchain and asset setup, combined with environment variables for runtime configuration.

### Blockchain Configuration

Create a `config/blockchains.yaml` file to define supported blockchains:

```yaml
default_contract_address: "0x019B65A265EB3363822f2752141b3dF16131b262"

blockchains:
  - name: polygon_amoy
    id: 80002
    contract_address: "0x9d1E88627884e066B81A02d69BCB2437a520534C"

  - name: base_sepolia
    id: 84532
    contract_address: "0x33e57a8900882B8D5A038eC3Aa844c19Acfc539A"
```

Configuration options:
- `default_contract_address`: Default Nitrolite contract address for blockchains without overrides

- `blockchains`: Array of blockchain configurations
  - `name`: Blockchain identifier (lowercase, underscores allowed)
  - `id`: Chain ID for validation
  - `disabled`: Set to `true` to skip this blockchain
  - `block_step`: Block range for event scanning (default: 10000)
  - `contract_address`: Override default address per chain

See [config/compose/example/blockchains.yaml](config/compose/example/blockchains.yaml) for a complete example.

### Asset Configuration

Create a `config/assets.yaml` file to define supported assets and their token implementations:

```yaml
assets:
  - name: "USD Coin"
    decimals: 6
    symbol: "USDC"
    tokens:
      - blockchain_id: 80002
        address: "0xDB9F293e3898c9E5536A3be1b0C56c89d2b32DEb"
        decimals: 6
      - blockchain_id: 84532
        address: "0xaf88d065e77c8cC2239327C5EDb3A432268e5831"
        decimals: 6
```

Configuration options:
- `name`: Human-readable asset name (optional, defaults to symbol)
- `symbol`: Asset ticker symbol (required, case-sensitive)
- `decimals`: Number of decimal places in YN (required)
- `disabled`: Set to `true` to skip this asset
- `tokens`: Array of token implementations across chains
  - `blockchain_id`: Chain ID where token is deployed
  - `address`: Token contract address
  - `decimals`: Number of decimal places
  - `name`: Token-specific name (optional, inherits from asset)
  - `symbol`: Token-specific symbol (optional, inherits from asset)
  - `disabled`: Set to `true` to skip this token

See [config/compose/example/assets.yaml](config/compose/example/assets.yaml) for a complete example.

### Environment Variables

Configure Clearnode using environment variables:

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `CLEARNODE_SIGNER_KEY` | Private key for signing node messages | Yes | - |
| `CLEARNODE_DATABASE_DRIVER` | Database driver (`postgres` or `sqlite`) | No | `sqlite` |
| `CLEARNODE_DATABASE_URL` | Database connection string | No | `clearnode.db` |
| `CLEARNODE_LOG_LEVEL` | Logging level (`debug`, `info`, `warn`, `error`) | No | `info` |
| `CLEARNODE_CONFIG_DIR_PATH` | Path to configuration directory | No | `.` |
| `CLEARNODE_CHANNEL_MIN_CHALLENGE_DURATION` | Minimum channel challenge period (seconds) | No | `86400` |
| `CLEARNODE_BLOCKCHAIN_RPC_<NAME>` | RPC endpoint for each blockchain | Yes (per blockchain) | - |

#### Blockchain RPC Configuration

For each enabled blockchain in `blockchains.yaml`, set an RPC endpoint:

```bash
# Format: CLEARNODE_BLOCKCHAIN_RPC_<BLOCKCHAIN_NAME_UPPERCASE>
CLEARNODE_BLOCKCHAIN_RPC_POLYGON_AMOY=https://rpc-amoy.polygon.technology
CLEARNODE_BLOCKCHAIN_RPC_BASE_SEPOLIA=https://sepolia.base.org
```

The blockchain name is converted to uppercase and prefixed with `CLEARNODE_BLOCKCHAIN_RPC_`.

#### Database Configuration

**SQLite** (default):
```bash
CLEARNODE_DATABASE_DRIVER=sqlite
CLEARNODE_DATABASE_URL=clearnode.db
```

**PostgreSQL**:
```bash
CLEARNODE_DATABASE_DRIVER=postgres
CLEARNODE_DATABASE_URL=postgresql://user:password@localhost:5432/clearnode?sslmode=disable
```

## Running Clearnode

### Prerequisites

- Go 1.25 or later
- SQLite (for default database) or PostgreSQL
- RPC access to configured blockchains

### Local Development

1. Create configuration directory:

```bash
mkdir -p config
```

2. Create `config/blockchains.yaml` and `config/assets.yaml` (see examples above)

3. Create `config/.env` with required variables:

```bash
# Required
CLEARNODE_SIGNER_KEY=0xYOUR_PRIVATE_KEY_HERE

# Blockchain RPCs (add for each enabled blockchain)
CLEARNODE_BLOCKCHAIN_RPC_POLYGON_AMOY=https://your-rpc-url
CLEARNODE_BLOCKCHAIN_RPC_BASE_SEPOLIA=https://your-rpc-url

# Optional
CLEARNODE_LOG_LEVEL=debug
```

4. Run the server:

```bash
export CLEARNODE_CONFIG_DIR_PATH=./config
go run .
```

The server will start on:
- **WebSocket RPC**: `ws://localhost:7824/ws`
- **Metrics**: `http://localhost:4242/metrics`

### Docker

Build and run with Docker:

```bash
# Build
docker build -t clearnode:latest .

# Run
docker run -p 7824:7824 -p 4242:4242 \
  -v $(pwd)/config:/config \
  -e CLEARNODE_CONFIG_DIR_PATH=/config \
  -e CLEARNODE_SIGNER_KEY=0xYOUR_KEY \
  -e CLEARNODE_BLOCKCHAIN_RPC_POLYGON_AMOY=https://rpc-url \
  clearnode:latest
```

### Docker Compose

Use the provided `docker-compose.yml` for a complete setup with PostgreSQL:

```bash
# Copy example configuration
cp -r config/compose/example config/compose/local

# Edit configuration files
vim config/compose/local/blockchains.yaml
vim config/compose/local/assets.yaml
vim config/compose/local/.env

# Start services
docker-compose up
```

### Communication Flows

Clearnode supports several key flows:

- **Off-chain transfers**: Balance transfers between users
- **Home channel creation**: Establish user's primary channel
- **Home channel deposits/withdrawals**: Fund and withdraw from channels
- **Escrow channel operations**: Cross-chain escrow management
- **App session lifecycle**: Multi-party application sessions

For detailed sequence diagrams, see [../docs/communication_flows/](../docs/communication_flows/).

## Development

### Project Structure

```
clearnode/
├── api/                    # RPC API handlers
│   ├── app_session_v1/    # App session endpoints
│   ├── channel_v1/        # Channel endpoints
│   ├── node_v1/           # Node info endpoints
│   ├── user_v1/           # User query endpoints
│   └── rpc_router.go      # RPC method routing
├── config/                # Configuration files
│   ├── compose/           # Docker compose configs
│   └── migrations/        # Database migrations
├── event_handlers/        # Blockchain event processing
├── metrics/               # Prometheus metrics
│   └── prometheus/
├── nitrolite/             # Smart contract bindings
├── store/                 # Data storage layer
│   ├── database/          # Persistent storage (PostgreSQL/SQLite)
│   └── memory/            # In-memory configuration store
├── blockchain_worker.go   # Blockchain interaction coordinator
├── main.go               # Application entry point
└── runtime.go            # Initialization and configuration
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./api/channel_v1/...
```

### Building

```bash
# Build binary
go build -o bin/clearnode

# Build with version
go build -o bin/clearnode -ldflags "-X main.Version=1.0.0"

# Build for production
CGO_ENABLED=1 go build -o bin/clearnode -ldflags "-X main.Version=1.0.0"
```

## Documentation

- [API Specification](../docs/api.yaml) - Complete RPC API reference
- [Data Models](../docs/data_models.mmd) - Core data structures
- [Communication Flows](../docs/communication_flows/) - Sequence diagrams
- [Nitrolite V1 Specs](../docs/README.md) - Protocol specifications

## License

See the main repository [LICENSE](../LICENSE) file for details.

## Support

For issues and questions:
- GitHub Issues: [github.com/erc7824/nitrolite/issues](https://github.com/erc7824/nitrolite/issues)
- Documentation: [github.com/erc7824/nitrolite/docs](https://github.com/erc7824/nitrolite/tree/main/docs)
