# Clearnode TypeScript SDK

[![npm version](https://img.shields.io/npm/v/@erc7824/nitrolite.svg)](https://www.npmjs.com/package/@erc7824/nitrolite)
[![License](https://img.shields.io/npm/l/@erc7824/nitrolite.svg)](https://github.com/erc7824/nitrolite/blob/main/LICENSE)
[![Documentation](https://img.shields.io/badge/docs-website-blue)](https://erc7824.org/quick_start)

TypeScript SDK for Clearnode payment channels providing both high-level and low-level operations in a unified client:
- **High-Level Operations**: `deposit()`, `withdraw()`, `transfer()`, `closeHomeChannel()` with automatic state management
- **Low-Level Operations**: Direct RPC access for custom flows and advanced use cases
- **Full Feature Parity**: 100% compatibility with Go SDK functionality

## Method Cheat Sheet

### High-Level Operations (Blockchain Interaction)
```typescript
client.deposit(blockchainId, asset, amount)       // Deposit to channel
client.withdraw(blockchainId, asset, amount)      // Withdraw from channel
client.transfer(recipientWallet, asset, amount)   // Off-chain transfer
client.closeHomeChannel(asset)                    // Close and finalize channel
```

### Node Information
```typescript
client.ping()                        // Health check
client.getConfig()                   // Node configuration
client.getBlockchains()              // Supported blockchains
client.getAssets(blockchainId?)      // Supported assets
```

### User Queries
```typescript
client.getBalances(wallet)              // User balances
client.getTransactions(wallet, opts)    // Transaction history
```

### Channel Queries
```typescript
client.getHomeChannel(wallet, asset)            // Home channel info
client.getEscrowChannel(escrowChannelId)        // Escrow channel info
client.getLatestState(wallet, asset, onlySigned) // Latest state
```

### App Sessions
```typescript
client.getAppSessions(opts)                                     // List sessions
client.getAppDefinition(appSessionId)                           // Session definition
client.createAppSession(definition, sessionData, sigs)          // Create session
client.submitAppSessionDeposit(update, sigs, asset, amount)     // Deposit to session
client.submitAppState(update, sigs)                             // Update session
client.rebalanceAppSessions(signedUpdates)                      // Atomic rebalance
```

### Session Keys
```typescript
client.submitSessionKeyState(state)                             // Register/update session key
client.getLastKeyStates(userAddress, sessionKey?)               // Get active session key states
```

### Shared Utilities
```typescript
client.close()                              // Close connection
client.waitForClose()                       // Connection monitor promise
client.signState(state)                     // Sign a state (advanced)
client.getUserAddress()                     // Get signer's address
client.setHomeBlockchain(asset, chainId)    // Set default blockchain for asset
```

## Installation

```bash
npm install @erc7824/nitrolite
# or
yarn add @erc7824/nitrolite
# or
pnpm add @erc7824/nitrolite
```

## Quick Start

### Unified Client (High-Level + Low-Level)

```typescript
import { Client, createSigners, withBlockchainRPC } from '@erc7824/nitrolite';
import Decimal from 'decimal.js';

async function main() {
  // Create signers from private key
  const { stateSigner, txSigner } = createSigners(
    process.env.PRIVATE_KEY as `0x${string}`
  );

  // Create unified client
  const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner,
    withBlockchainRPC(80002n, 'https://polygon-amoy.alchemy.com/v2/KEY')
  );

  try {
    // High-level operations - SDK handles everything
    const txHash = await client.deposit(80002n, 'usdc', new Decimal(100));
    const txId = await client.transfer('0xRecipient...', 'usdc', new Decimal(50));
    const withdrawTx = await client.withdraw(80002n, 'usdc', new Decimal(25));

    // Low-level operations - same client
    const config = await client.getConfig();
    const balances = await client.getBalances(client.getUserAddress());
    const state = await client.getLatestState(client.getUserAddress(), 'usdc', false);
  } finally {
    await client.close();
  }
}

main().catch(console.error);
```

## Architecture

```
sdk/ts/src/
├── client.ts         # Core client, constructors, high-level operations
├── signers.ts        # EthereumMsgSigner and EthereumRawSigner
├── config.ts         # Configuration options
├── asset_store.ts    # Asset metadata caching
├── utils.ts          # Type transformations
├── core/             # State management, transitions, types
├── rpc/              # WebSocket RPC client
├── blockchain/       # EVM blockchain interactions
└── app/              # App session types and logic
```

## Client API

### Creating a Client

```typescript
import { Client, createSigners, withBlockchainRPC } from '@erc7824/nitrolite';

// Step 1: Create signers from private key
const { stateSigner, txSigner } = createSigners('0x1234...');

// Step 2: Create unified client
const client = await Client.create(
  wsURL,
  stateSigner,  // For signing channel states
  txSigner,     // For signing blockchain transactions
  withBlockchainRPC(chainId, rpcURL), // Required for Deposit/Withdraw
  withHandshakeTimeout(10000),         // Optional: connection timeout
  withPingInterval(5000)               // Optional: keepalive interval
);

// Step 3: (Optional) Set home blockchain for assets
// Required for Transfer operations that may trigger channel creation
await client.setHomeBlockchain('usdc', 80002n);
```

### Signer Implementations

The SDK provides two signer types matching the Go SDK patterns:

#### EthereumMsgSigner (for channel states)

Signs channel state updates with EIP-191 "Ethereum Signed Message" prefix.

```typescript
import { EthereumMsgSigner } from '@erc7824/nitrolite';
import { privateKeyToAccount } from 'viem/accounts';

// From private key
const signer1 = new EthereumMsgSigner('0x...');

// From viem account
const account = privateKeyToAccount('0x...');
const signer2 = new EthereumMsgSigner(account);
```

**When to use**: All off-chain operations (state signatures, transfers)

#### EthereumRawSigner (for blockchain transactions)

Signs raw hashes directly without prefix for on-chain operations.

```typescript
import { EthereumRawSigner } from '@erc7824/nitrolite';

const signer = new EthereumRawSigner('0x...');
```

**When to use**: On-chain operations (deposits, withdrawals, channel creation)

#### Helper: createSigners()

```typescript
import { createSigners } from '@erc7824/nitrolite';

// Creates both signers at once
const { stateSigner, txSigner } = createSigners('0x...');
const client = await Client.create(wsURL, stateSigner, txSigner);
```

### Configuring Home Blockchain

#### `setHomeBlockchain(asset, blockchainId)`

Sets the default blockchain network for a specific asset. Required for `transfer()` operations that may trigger channel creation.

```typescript
await client.setHomeBlockchain('usdc', 80002n);
```

**Important Notes:**
- This mapping is immutable once set for the client instance
- The asset must be supported on the specified blockchain
- Required before calling `transfer()` on a new channel

### High-Level Operations

#### `deposit(blockchainId, asset, amount)`

Deposits funds into channel. Automatically handles:
- Channel creation if needed
- Checkpointing to existing channel
- State building and signing
- Blockchain transaction

```typescript
const txHash = await client.deposit(
  80002n,             // Blockchain ID
  'usdc',             // Asset symbol
  new Decimal(100)    // Amount
);
```

**Requirements:**
- Blockchain RPC configured via `withBlockchainRPC()`
- Token approval for contract address
- Sufficient token balance and gas

**Scenarios:**
1. **No channel exists**: Creates new channel with initial deposit
2. **Channel exists**: Checkpoints deposit to existing channel

#### `transfer(recipientWallet, asset, amount)`

Off-chain transfer to another wallet. Instant, no gas required.

```typescript
const txId = await client.transfer(
  '0xRecipient...',   // Recipient address
  'usdc',             // Asset symbol
  new Decimal(50)     // Amount
);
```

**Requirements:**
- Existing channel with sufficient balance OR
- Home blockchain configured via `setHomeBlockchain()` (for new channels)

#### `withdraw(blockchainId, asset, amount)`

Withdraws funds from channel to blockchain wallet.

```typescript
const txHash = await client.withdraw(
  80002n,             // Blockchain ID
  'usdc',             // Asset symbol
  new Decimal(25)     // Amount
);
```

**Requirements:**
- Existing channel with sufficient balance
- Blockchain RPC configured
- Sufficient gas for transaction

#### `closeHomeChannel(asset)`

Finalizes and closes the user's channel for a specific asset.

```typescript
const txHash = await client.closeHomeChannel('usdc');
```

**Requirements:**
- Existing channel (user must have deposited first)
- Blockchain RPC configured
- Sufficient gas for transaction

## Low-Level API

All low-level RPC methods are available on the same Client instance.

### Node Information

```typescript
await client.ping();
const config = await client.getConfig();
const blockchains = await client.getBlockchains();
const assets = await client.getAssets(); // or client.getAssets(blockchainId)
```

### User Data

```typescript
const balances = await client.getBalances(wallet);
const { transactions, metadata } = await client.getTransactions(wallet, {
  page: 1,
  pageSize: 50,
});
```

### Channel Queries

```typescript
const channel = await client.getHomeChannel(wallet, asset);
const escrow = await client.getEscrowChannel(escrowChannelId);
const state = await client.getLatestState(wallet, asset, onlySigned);
```

**Note:** State submission and channel creation are handled internally by high-level operations (`deposit()`, `withdraw()`, `transfer()`).

### App Sessions (Low-Level)

```typescript
// Query sessions
const { sessions, metadata } = await client.getAppSessions(opts);
const definition = await client.getAppDefinition(appSessionId);

// Create and manage sessions
const { appSessionId, version, status } = await client.createAppSession(
  definition,
  sessionData,
  signatures
);

const nodeSig = await client.submitAppSessionDeposit(
  appUpdate,
  quorumSigs,
  asset,
  depositAmount
);

await client.submitAppState(appUpdate, quorumSigs);

const batchId = await client.rebalanceAppSessions(signedUpdates);
```

### Session Keys

```typescript
// Submit a session key state for registration or update
await client.submitSessionKeyState({
  user_address: '0x1234...',
  session_key: '0xabcd...',
  version: '1',
  application_id: ['app1'],
  app_session_id: [],
  expires_at: String(Math.floor(Date.now() / 1000) + 86400),
  user_sig: '0x...',
});

// Query active session key states
const states = await client.getLastKeyStates('0x1234...');
const filtered = await client.getLastKeyStates('0x1234...', '0xSessionKey...');
```

## Key Concepts

### State Management

Payment channels use versioned states signed by both user and node:

```typescript
// High-level operations handle state management automatically
client.deposit(...)   // Creates/updates state, signs, submits
client.withdraw(...)  // Updates state, signs, submits
client.transfer(...)  // Updates state, signs, submits
```

**State Flow (Internal):**
1. Get latest state with `getLatestState()`
2. Create next state with `nextState()`
3. Apply transition (deposit, withdraw, transfer, etc.)
4. Calculate state ID
5. Sign state with `signState()`
6. Submit to node (internal method)

State submission and channel creation are handled automatically by high-level operations.

### Signing

States are signed using ECDSA with EIP-191/EIP-155:

```typescript
// Create signers from private key
const { stateSigner, txSigner } = createSigners('0x...');

// Get address
const address = stateSigner.getAddress();
```

**Signing Process:**
1. State → ABI Encode (via `packState`)
2. Packed State → Keccak256 Hash
3. Hash → ECDSA Sign (via signer)
4. Result: 65-byte signature (R || S || V)

**Two Signer Types:**
- `EthereumMsgSigner`: Signs channel state updates (off-chain signatures) with EIP-191 prefix
- `EthereumRawSigner`: Signs blockchain transactions (on-chain operations) without prefix

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

```typescript
try {
  const txHash = await client.deposit(80002n, 'usdc', amount);
} catch (error) {
  // Error: "failed to create channel on blockchain: insufficient balance"
  console.error('Deposit failed:', error);
}
```

### Common Errors

| Error Message | Cause | Solution |
|--------------|-------|----------|
| `"channel not created, deposit first"` | Transfer before deposit | Deposit funds first |
| `"home blockchain not set for asset"` | Missing `setHomeBlockchain()` | Call `setHomeBlockchain()` before transfer |
| `"blockchain client not configured"` | Missing `withBlockchainRPC()` | Add `withBlockchainRPC()` configuration |
| `"insufficient balance"` | Not enough funds | Deposit more funds |
| `"failed to sign state"` | Invalid private key or state | Check signer configuration |

### Custom Error Handler

```typescript
const client = await Client.create(
  wsURL,
  stateSigner,
  txSigner,
  withErrorHandler((error) => {
    console.error('[Connection Error]', error);
    // Custom error handling logic
  })
);
```

## Configuration Options

```typescript
import {
  withBlockchainRPC,
  withHandshakeTimeout,
  withPingInterval,
  withErrorHandler
} from '@erc7824/nitrolite';

const client = await Client.create(
  wsURL,
  stateSigner,
  txSigner,
  withBlockchainRPC(chainId, rpcURL),  // Configure blockchain RPC
  withHandshakeTimeout(10000),          // Connection timeout (ms, default: 5000)
  withPingInterval(5000),               // Keepalive interval (ms, default: 5000)
  withErrorHandler(func)                // Connection error handler
);
```

## Complete Examples

### Example 1: Basic Deposit and Transfer

```typescript
import { Client, createSigners, withBlockchainRPC } from '@erc7824/nitrolite';
import Decimal from 'decimal.js';

async function basicExample() {
  const { stateSigner, txSigner } = createSigners(process.env.PRIVATE_KEY!);

  const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner,
    withBlockchainRPC(80002n, process.env.RPC_URL!)
  );

  try {
    console.log('User:', client.getUserAddress());

    // Set home blockchain
    await client.setHomeBlockchain('usdc', 80002n);

    // Deposit 100 USDC
    const depositTx = await client.deposit(80002n, 'usdc', new Decimal(100));
    console.log('Deposited:', depositTx);

    // Check balance
    const balances = await client.getBalances(client.getUserAddress());
    console.log('Balances:', balances);

    // Transfer 50 USDC
    const transferId = await client.transfer(
      '0xRecipient...',
      'usdc',
      new Decimal(50)
    );
    console.log('Transfer ID:', transferId);
  } finally {
    await client.close();
  }
}

basicExample().catch(console.error);
```

### Example 2: Multi-Chain Operations

```typescript
import { Client, createSigners, withBlockchainRPC } from '@erc7824/nitrolite';
import Decimal from 'decimal.js';

async function multiChainExample() {
  const { stateSigner, txSigner } = createSigners(process.env.PRIVATE_KEY!);

  const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner,
    withBlockchainRPC(80002n, process.env.POLYGON_RPC!), // Polygon Amoy
    withBlockchainRPC(11155111n, process.env.SEPOLIA_RPC!) // Sepolia
  );

  try {
    // Set home blockchains
    await client.setHomeBlockchain('usdc', 80002n);
    await client.setHomeBlockchain('eth', 11155111n);

    // Deposit on different chains
    await client.deposit(80002n, 'usdc', new Decimal(100));
    await client.deposit(11155111n, 'eth', new Decimal(0.1));

    // Check balances across all chains
    const balances = await client.getBalances(client.getUserAddress());
    balances.forEach(b => console.log(`${b.asset}: ${b.balance}`));
  } finally {
    await client.close();
  }
}

multiChainExample().catch(console.error);
```

### Example 3: Transaction History with Pagination

```typescript
import { Client, createSigners } from '@erc7824/nitrolite';

async function queryTransactions() {
  const { stateSigner, txSigner } = createSigners(process.env.PRIVATE_KEY!);
  const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner
  );

  try {
    const wallet = client.getUserAddress();

    // Get paginated transactions
    const result = await client.getTransactions(wallet, {
      page: 1,
      pageSize: 10,
    });

    console.log(`Total: ${result.metadata.totalCount}`);
    console.log(`Page ${result.metadata.page} of ${result.metadata.pageCount}`);

    result.transactions.forEach((tx, i) => {
      console.log(`${i + 1}. ${tx.txType}: ${tx.amount} ${tx.asset}`);
    });
  } finally {
    await client.close();
  }
}

queryTransactions().catch(console.error);
```

### Example 4: App Session Workflow

```typescript
import { Client, createSigners, withBlockchainRPC } from '@erc7824/nitrolite';
import Decimal from 'decimal.js';

async function appSessionExample() {
  const { stateSigner, txSigner } = createSigners(process.env.PRIVATE_KEY!);
  const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner,
    withBlockchainRPC(80002n, process.env.RPC_URL!)
  );

  try {
    // Create app session
    const definition = {
      application: 'chess-v1',
      participants: [
        { walletAddress: client.getUserAddress(), signatureWeight: 1 },
        { walletAddress: '0xOpponent...', signatureWeight: 1 },
      ],
      quorum: 2,
      nonce: 1n,
    };

    const { appSessionId } = await client.createAppSession(
      definition,
      '{}',
      ['sig1', 'sig2']
    );
    console.log('Session created:', appSessionId);

    // Deposit to app session
    const appUpdate = {
      appSessionId,
      intent: 1, // Deposit
      version: 1n,
      allocations: [{
        participant: client.getUserAddress(),
        asset: 'usdc',
        amount: new Decimal(50),
      }],
      sessionData: '{}',
    };

    const nodeSig = await client.submitAppSessionDeposit(
      appUpdate,
      ['sig1'],
      'usdc',
      new Decimal(50)
    );
    console.log('Deposit signature:', nodeSig);

    // Query sessions
    const { sessions } = await client.getAppSessions({
      wallet: client.getUserAddress(),
    });
    console.log(`Found ${sessions.length} sessions`);
  } finally {
    await client.close();
  }
}

appSessionExample().catch(console.error);
```

### Example 5: Connection Monitoring

```typescript
import { Client, createSigners, withErrorHandler, withPingInterval } from '@erc7824/nitrolite';

async function monitorConnection() {
  const { stateSigner, txSigner } = createSigners(process.env.PRIVATE_KEY!);

  const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner,
    withPingInterval(3000),
    withErrorHandler((error) => {
      console.error('Connection error:', error);
    })
  );

  // Monitor connection
  client.waitForClose().then(() => {
    console.log('Connection closed, reconnecting...');
    // Reconnection logic here
  });

  // Perform operations
  const config = await client.getConfig();
  console.log('Connected to:', config.nodeAddress);

  // Keep alive...
  await new Promise(resolve => setTimeout(resolve, 30000));
  await client.close();
}

monitorConnection().catch(console.error);
```

## TypeScript-Specific Notes

### Type Imports

```typescript
import type {
  State,
  Channel,
  Transaction,
  BalanceEntry,
  Asset,
  Blockchain,
  AppSessionInfoV1,
  AppDefinitionV1,
  AppSessionKeyStateV1,
  PaginationMetadata,
} from '@erc7824/nitrolite';
```

### BigInt for Chain IDs

```typescript
// Use 'n' suffix for bigint literals
const polygonAmoy = 80002n;
const ethereum = 1n;

await client.deposit(polygonAmoy, 'usdc', amount);
```

### Decimal.js for Amounts

```typescript
import Decimal from 'decimal.js';

const amount1 = new Decimal(100);
const amount2 = new Decimal('123.456');
const amount3 = Decimal.div(1000, 3);

await client.deposit(chainId, 'usdc', amount1);
```

### Viem Integration

```typescript
import { privateKeyToAccount } from 'viem/accounts';
import type { Address } from 'viem';

const account = privateKeyToAccount('0x...');
const stateSigner = new EthereumMsgSigner(account);

const wallet: Address = '0x1234...';
const balances = await client.getBalances(wallet);
```

### Async/Await

```typescript
// All SDK methods are async
const txHash = await client.deposit(chainId, asset, amount);

// Or with .then()
client.deposit(chainId, asset, amount)
  .then(txHash => console.log('Deposited:', txHash))
  .catch(error => console.error('Error:', error));
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

### CloseHomeChannel Flow
1. Get latest state
2. Verify channel exists
3. Create next state
4. Apply finalize transition
5. Sign state
6. Submit to node
7. Close channel on blockchain
8. Return transaction hash

## Requirements

- **Node.js**: 20.0.0 or later
- **TypeScript**: 5.3.0 or later (for development)
- **Running Clearnode instance** or access to public node
- **Blockchain RPC endpoint** (for on-chain operations)

## Documentation

For complete documentation, visit [https://erc7824.org](https://erc7824.org)

### Documentation Links

- [Quick Start Guide](https://erc7824.org/quick_start)
- [Channel Creation](https://erc7824.org/quick_start/initializing_channel)
- [ClearNode Connection](https://erc7824.org/quick_start/connect_to_the_clearnode)
- [Application Sessions](https://erc7824.org/quick_start/application_session)
- [Session Closure](https://erc7824.org/quick_start/close_session)

## Build with AI

We have generated a [llms-full.txt](https://erc7824.org/llms-full.txt) file that converts all our documentation into a single markdown document following the [llmstxt.org](https://llmstxt.org/) standard.

## License

Part of the Nitrolite project. See [LICENSE](../../LICENSE) for details.

## Related Projects

- [Nitrolite Go SDK](../go/README.md) - Go implementation with same API
- [ERC-7824 Specification](https://eips.ethereum.org/EIPS/eip-7824) - Standard specification
- [Nitrolite Smart Contracts](../../contract/) - On-chain contracts

---

**Built with Nitrolite** - Powering the next generation of scalable blockchain applications.
