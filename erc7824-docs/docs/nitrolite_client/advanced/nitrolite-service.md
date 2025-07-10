---
sidebar_position: 2
title: NitroliteService
description: Documentation for the core NitroliteService class
keywords: [erc7824, statechannels, state channels, nitrolite, ethereum scaling, layer 2, off-chain, advanced]
---

# NitroliteService

The `NitroliteService` class is the core service that directly interacts with the Nitrolite Custody smart contract. It handles channel management, deposits, withdrawals, and all other channel-specific operations following the channel lifecycle.

## Initialization

```typescript
import { NitroliteService } from '@erc7824/nitrolite';

const nitroliteService = new NitroliteService(
  publicClient,  // viem PublicClient
  addresses,     // ContractAddresses
  walletClient,  // viem WalletClient
  account        // Account address
);
```

## Channel Lifecycle Methods

### 1. Deposit Operations

| Method | Description | Parameters | Return Type |
|--------|-------------|------------|------------|
| `deposit` | Deposits tokens/ETH into the custody contract. | `tokenAddress: Address, amount: bigint` | `Promise<Hash>` |
| `withdraw` | Withdraws tokens from the custody contract. | `tokenAddress: Address, amount: bigint` | `Promise<Hash>` |

Example:
```typescript
// Deposit ETH or token
const txHash = await nitroliteService.deposit(tokenAddress, amount);

// Withdraw ETH or token
const txHash = await nitroliteService.withdraw(tokenAddress, amount);
```

### 2. Channel Creation

| Method | Description | Parameters | Return Type |
|--------|-------------|------------|------------|
| `createChannel` | Creates a new channel with the given parameters. | `channel: Channel, initialState: State` | `Promise<Hash>` |

Example:
```typescript
// Create a channel
const txHash = await nitroliteService.createChannel(channel, initialState);
```

Where:
- `channel` defines the participants, adjudicator, challenge period, and nonce
- `initialState` contains the initial allocation of funds and state data

### 3. Channel Operations

| Method | Description | Parameters | Return Type |
|--------|-------------|------------|------------|
| `checkpoint` | Checkpoints a state on-chain. | `channelId: ChannelId, candidateState: State, proofStates?: State[]` | `Promise<Hash>` |
| `challenge` | Challenges a channel with a candidate state. | `channelId: ChannelId, candidateState: State, proofStates?: State[]` | `Promise<Hash>` |
| `resize` | Resizes a channel with a candidate state. | `channelId: ChannelId, candidateState: State, proofStates?: State[]` | `Promise<Hash>` |

Example:
```typescript
// Checkpoint a channel state
const txHash = await nitroliteService.checkpoint(channelId, candidateState);

// Challenge a channel
const txHash = await nitroliteService.challenge(channelId, candidateState);

// Resize a channel
const txHash = await nitroliteService.resize(channelId, candidateState);
```

### 4. Channel Closing

| Method | Description | Parameters | Return Type |
|--------|-------------|------------|------------|
| `close` | Closes a channel using a final state. | `channelId: ChannelId, finalState: State` | `Promise<Hash>` |

Example:
```typescript
// Close a channel
const txHash = await nitroliteService.close(channelId, finalState);
```

### 5. Account Information

| Method | Description | Parameters | Return Type |
|--------|-------------|------------|------------|
| `getAccountChannels` | Gets channel IDs for an account. | `accountAddress: Address` | `Promise<ChannelId[]>` |
| `getAccountInfo` | Gets account info for a token. | `accountAddress: Address, tokenAddress: Address` | `Promise<AccountInfo>` |

Example:
```typescript
// Get all channels for an account
const channels = await nitroliteService.getAccountChannels(accountAddress);

// Get detailed account info
const info = await nitroliteService.getAccountInfo(accountAddress, tokenAddress);
console.log(`Available: ${info.available}, Locked: ${info.locked}`);
```

## Transaction Preparation Methods

For Account Abstraction support, NitroliteService provides transaction preparation methods that return transaction data without executing it:

| Method | Description | Parameters | Return Type |
|--------|-------------|------------|------------|
| `prepareDeposit` | Prepares deposit transaction. | `tokenAddress: Address, amount: bigint` | `Promise<PreparedTransaction>` |
| `prepareCreateChannel` | Prepares channel creation transaction. | `channel: Channel, initialState: State` | `Promise<PreparedTransaction>` |
| `prepareCheckpoint` | Prepares checkpoint transaction. | `channelId: ChannelId, candidateState: State, proofStates?: State[]` | `Promise<PreparedTransaction>` |
| `prepareChallenge` | Prepares challenge transaction. | `channelId: ChannelId, candidateState: State, proofStates?: State[]` | `Promise<PreparedTransaction>` |
| `prepareResize` | Prepares resize transaction. | `channelId: ChannelId, candidateState: State, proofStates?: State[]` | `Promise<PreparedTransaction>` |
| `prepareClose` | Prepares close transaction. | `channelId: ChannelId, finalState: State` | `Promise<PreparedTransaction>` |
| `prepareWithdraw` | Prepares withdraw transaction. | `tokenAddress: Address, amount: bigint` | `Promise<PreparedTransaction>` |

Example:
```typescript
// Prepare deposit transaction
const tx = await nitroliteService.prepareDeposit(tokenAddress, amount);

// Use with your Account Abstraction provider
const userOp = await aaProvider.buildUserOperation({
  target: tx.to,
  data: tx.data,
  value: tx.value || 0n
});
```

## Implementation Details

The `NitroliteService` connects to the Custody contract using:

- A viem `PublicClient` for read operations
- A viem `WalletClient` for write operations and signing
- The contract address specified in the configuration

The service handles:
- Contract interaction
- Parameter validation
- Error handling
- Transaction preparation

## Error Handling

The `NitroliteService` throws specific error types:

- `ContractCallError`: When calls to the contract fail
- `InvalidParameterError`: When parameters are invalid
- `MissingParameterError`: When required parameters are missing
- `WalletClientRequiredError`: When wallet client is needed but not provided
- `AccountRequiredError`: When account is needed but not provided

Example:
```typescript
try {
  await nitroliteService.deposit(tokenAddress, amount);
} catch (error) {
  if (error instanceof ContractCallError) {
    console.error(`Contract call failed: ${error.message}`);
    console.error(`Suggestion: ${error.suggestion}`);
  }
}
```

## Advanced Usage

### Custom Contract Interaction

For advanced use cases, you might need to interact directly with the contract:

```typescript
// Get the custody contract address
const custodyAddress = nitroliteService.custodyAddress;

// Use with your own custom contract interaction
const customContract = getContract({
  address: custodyAddress,
  abi: custodyAbi,
  // Additional configuration...
});
```

### Multiple Channel Management

For applications managing multiple channels:

```typescript
// Get all channels for the account
const channels = await nitroliteService.getAccountChannels(accountAddress);

// Process each channel
for (const channelId of channels) {
  // Get channel info from contract
  // Process channel state
}
```