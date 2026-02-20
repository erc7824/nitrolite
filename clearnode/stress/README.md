# Clearnode Stress Testing Tool

Built-in stress testing tool for validating clearnode performance, correctness, and stability under load.

## Quick Start

```bash
# Set target
export STRESS_WS_URL=ws://localhost:7824/ws

# Read-only test (no wallet needed)
clearnode stress-test ping:1000:10

# State-mutating test (funded wallet required)
export STRESS_PRIVATE_KEY=<hex-encoded-private-key>
clearnode stress-test transfer-roundtrip:10:20:usdc
```

## Architecture

The stress tool is compiled into the clearnode binary and invoked via the `stress-test` subcommand. It connects to a running clearnode instance over WebSocket using the Go SDK.

```
clearnode stress-test <spec>
        │
        ├── Read-only methods ──► Connection pool (N parallel WebSocket clients)
        │                              │
        │                              └── Distributes requests round-robin across connections
        │
        └── State-mutating methods ──► Custom orchestration
                                          │
                                          ├── transfer-roundtrip: 3-phase fund/stress/collect
                                          └── app-session-lifecycle: create/deposit/operate/close
```

**Key design decisions:**
- Each WebSocket connection sends requests sequentially (waits for response before sending next)
- Parallelism is achieved through multiple connections
- Connection pool tolerates individual failures — test runs with whatever connections succeeded
- Results include per-request latency, percentile distribution, and error breakdown

## Configuration

All configuration is via environment variables.

| Variable | Required | Default | Description |
|---|---|---|---|
| `STRESS_WS_URL` | Yes | - | WebSocket URL of the target clearnode |
| `STRESS_PRIVATE_KEY` | No | ephemeral | Hex-encoded ECDSA private key |
| `STRESS_CONNECTIONS` | No | `10` | Default parallel connections per test |
| `STRESS_TIMEOUT` | No | `10m` | Overall test timeout |
| `STRESS_MAX_ERROR_RATE` | No | `0.01` | Error rate threshold (0.01 = 1%) |

When `STRESS_PRIVATE_KEY` is not set, an ephemeral key is generated. This works for read-only methods but state-mutating methods require a funded wallet.

## Spec Format

```
method:total_requests[:connections[:extra_params...]]
```

- `method` — test method name
- `total_requests` — total number of operations to execute
- `connections` — parallel WebSocket connections (optional, falls back to `STRESS_CONNECTIONS`)
- `extra_params` — method-specific parameters (asset, amount, wallet address, etc.)

## Available Methods

### Read-Only Methods

These methods test read path performance. They use a shared connection pool and do not modify server state. An ephemeral wallet is used if `STRESS_PRIVATE_KEY` is not set.

| Method | Extra Params | Description |
|---|---|---|
| `ping` | none | WebSocket ping/pong roundtrip |
| `get-config` | none | Fetch server configuration |
| `get-blockchains` | none | List available blockchains |
| `get-assets` | `[chain_id]` | List assets, optionally filtered by chain |
| `get-balances` | `[wallet]` | Get wallet balances |
| `get-transactions` | `[wallet]` | Fetch transactions (paginated, limit 20) |
| `get-home-channel` | `asset` or `wallet:asset` | Get home channel for wallet+asset |
| `get-escrow-channel` | `channel_id` | Get escrow channel by ID |
| `get-latest-state` | `asset` or `wallet:asset` | Get latest channel state |
| `get-channel-key-states` | `[wallet]` | Get last channel key states |
| `get-app-sessions` | `[wallet]` | Query app sessions (paginated) |
| `get-app-key-states` | `[wallet]` | Get last app key states |

**Examples:**

```bash
# 1000 pings over 10 connections
clearnode stress-test ping:1000:10

# 500 config fetches over 5 connections
clearnode stress-test get-config:500:5

# 2000 balance queries over 20 connections
clearnode stress-test get-balances:2000:20:0x1234...

# 1000 home channel lookups
clearnode stress-test get-home-channel:1000:10:usdc

# Asset queries filtered by chain ID
clearnode stress-test get-assets:500:5:84532
```

### State-Mutating Methods

These methods test write path performance. They require `STRESS_PRIVATE_KEY` with a funded wallet.

#### `transfer-roundtrip`

Spec: `transfer-roundtrip:rounds:wallets:asset[:amount]`

| Param | Description |
|---|---|
| `rounds` | Back-and-forth transfer rounds per wallet pair |
| `wallets` | Number of derived wallets (rounded up to even) |
| `asset` | Asset symbol (e.g., `usdc`) |
| `amount` | Transfer amount per operation (default: `0.000001`) |

**Three-phase execution:**

1. **Fund** — Sender distributes `amount` to each derived wallet
2. **Stress** — Wallet pairs (0,1), (2,3), ... transfer back and forth in parallel for `rounds` iterations
3. **Collect** — All wallets return funds to sender

Wallet keys are deterministically derived from the master key using SHA-256: `masterKey:receiver:<index>`.

Total measured operations = `wallets * rounds` (phase 2 only).

**Examples:**

```bash
# 10 rounds, 20 wallets (10 pairs), usdc, default amount
clearnode stress-test transfer-roundtrip:10:20:usdc

# 50 rounds, 100 wallets (50 pairs), custom amount
clearnode stress-test transfer-roundtrip:50:100:usdc:0.0001
```

#### `app-session-lifecycle`

Spec: `app-session-lifecycle:sessions:participants:operates:asset[:amount]`

| Param | Description |
|---|---|
| `sessions` | Number of concurrent app session lifecycles |
| `participants` | Wallets per session (quorum = all must sign) |
| `operates` | Number of operate state updates per session |
| `asset` | Asset symbol (e.g., `usdc`) |
| `amount` | Deposit amount per session (default: `0.000003`) |

**Per-session lifecycle:**

1. **Create** — Create app session with all participants
2. **Deposit** — First participant deposits funds into session
3. **Operate** — Submit `N` state updates with rotating fund allocations
4. **Close** — Close session with final allocation matching last operate

All signatures are pre-generated before the stress phase begins. Each session is driven by its first participant ("pipe lead") over a dedicated WebSocket connection.

Wallet keys are derived using SHA-256: `masterKey:appsession:<pipeIdx>:<walletIdx>`.

Total measured operations = `sessions * (operates + 3)`.

**Examples:**

```bash
# 10 sessions, 5 participants each, 3 operates per session
clearnode stress-test app-session-lifecycle:10:5:3:usdc

# 50 sessions, 3 participants, 10 operates, custom amount
clearnode stress-test app-session-lifecycle:50:3:10:usdc:0.000005
```

## Report Output

Every test produces a standardized report:

```
Stress Test Report
==================
Method:          ping
Total Requests:  1000
Connections:     10
Duration:        2.345s

Results
-------
Successful:      998 (99.8%)
Failed:          2 (0.2%)
Requests/sec:    426.44

Latency
-------
Min:             1.2ms
Max:             45.3ms
Average:         2.3ms
Median (p50):    2.1ms
P95:             4.5ms
P99:             12.8ms

Errors
------
  context deadline exceeded                                          2
```

**Pass/fail criteria:** The test exits with code 0 (PASS) if the error rate is within `STRESS_MAX_ERROR_RATE`, or code 1 (FAIL) if exceeded.

## Helm Integration

The stress tool is integrated as a Helm test. When enabled, `helm test` creates a Pod that runs the stress spec against the in-cluster clearnode service.

**values.yaml:**

```yaml
stressTest:
  enabled: true
  specs:
    - "ping:100000:100"
  privateKey: "<hex-key>"  # optional, for state-mutating tests
  connections: 10
  timeout: "10m"
  maxErrorRate: "0.01"
```

**Run:**

```bash
helm test <release-name>
```

The WebSocket URL defaults to the in-cluster service (`ws://<release>-clearnode:7824/ws`). Override with `stressTest.wsURL` for external targets.

## Testing Strategy

### Phase 1: Read Path Baseline

Validate read performance under increasing load. No funded wallet needed.

```bash
export STRESS_WS_URL=ws://target:7824/ws

# Baseline latency
clearnode stress-test ping:100:1
clearnode stress-test get-config:100:1

# Scale connections
clearnode stress-test ping:10000:10
clearnode stress-test ping:100000:100
clearnode stress-test get-balances:10000:50:0xWALLET
```

### Phase 2: Write Path Stress

Test state mutation throughput. Requires funded wallet.

```bash
export STRESS_PRIVATE_KEY=<key>

# Small scale
clearnode stress-test transfer-roundtrip:5:4:usdc

# Production scale
clearnode stress-test transfer-roundtrip:50:100:usdc:0.0001
```

### Phase 3: App Session Lifecycle

Test multi-participant coordination.

```bash
# Small scale
clearnode stress-test app-session-lifecycle:5:3:3:usdc

# Production scale
clearnode stress-test app-session-lifecycle:50:5:10:usdc:0.000005
```

### Phase 4: Sustained Load

Run extended tests to detect resource leaks and degradation.

```bash
# High volume read
clearnode stress-test ping:1000000:100

# Extended write
clearnode stress-test transfer-roundtrip:500:100:usdc:0.000001
```

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `STRESS_WS_URL is required` | Missing environment variable | Set `STRESS_WS_URL` |
| `failed to open any connections` | Target unreachable or refusing connections | Verify URL and that clearnode is running |
| `WARNING: Only N/M connections established` | Server under load or connection limits | Reduce connection count or check server capacity |
| `STRESS_PRIVATE_KEY is required` | State-mutating method without key | Set `STRESS_PRIVATE_KEY` with a funded wallet |
| `fund wallet X: insufficient balance` | Sender wallet not funded | Transfer funds to the wallet address printed at startup |
| High error rate in transfer tests | Database contention or deadlocks | Check server logs for deadlock traces |
| `context deadline exceeded` | Test exceeded `STRESS_TIMEOUT` | Increase timeout or reduce test scope |
