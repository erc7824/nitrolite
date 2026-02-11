# Nitrolite Client Module

This module provides on-chain interaction capabilities for the Nitrolite V1 SDK, allowing clients to manage state channels, deposits, withdrawals, and channel lifecycle operations through smart contracts.

## Overview

The Nitrolite Client module offers three layers of abstraction:

1. **NitroliteClient**: High-level convenience API for common operations
2. **NitroliteService**: Mid-level service for direct contract interactions
3. **NitroliteTransactionPreparer**: Low-level transaction preparation for Account Abstraction and batching

Choose the layer that best fits your needs. All layers interact with the same V1 Custody contract.

## Architecture

```
┌─────────────────────────────────────────────┐
│           NitroliteClient                    │
│  High-level API with automatic signing      │
│  (deposit, createChannel, closeChannel)     │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│        NitroliteService                      │
│  Direct contract interactions                │
│  (prepare + execute methods)                 │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│   NitroliteTransactionPreparer               │
│  Transaction preparation only                │
│  (for AA, batching, custom flows)            │
└──────────────────────────────────────────────┘
```

## Quick Start

### Basic Setup

```typescript
import { NitroliteClient } from '@erc7824/nitrolite';
import { createPublicClient, createWalletClient, http } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { mainnet } from 'viem/chains';

// Create viem clients
const account = privateKeyToAccount('0x...');
const publicClient = createPublicClient({
    chain: mainnet,
    transport: http(),
});
const walletClient = createWalletClient({
    account,
    chain: mainnet,
    transport: http(),
});

// Create state signer
const stateSigner = {
    signState: async (stateHash) => {
        return account.signMessage({ message: { raw: stateHash } });
    },
};

// Initialize client
const client = new NitroliteClient({
    publicClient,
    walletClient,
    stateSigner,
    addresses: {
        custody: '0x...', // V1 Custody contract address
    },
    chainId: 1,
    challengeDuration: 3600, // 1 hour in seconds
});
```

### Common Operations

```typescript
import { zeroAddress } from 'viem';

// 1. Deposit ETH to vault
const depositTx = await client.deposit(
    nodeAddress,
    zeroAddress, // ETH
    parseEther('1.0')
);
await client.waitForTransaction(depositTx);

// 2. Create a channel
const createTx = await client.createChannel({
    definition: {
        challengeDuration: 3600,
        user: account.address,
        node: nodeAddress,
        nonce: 1n,
        metadata: '0x0000000000000000000000000000000000000000000000000000000000000000',
    },
    state: {
        version: 0n,
        intent: 0,
        metadata: '0x0000000000000000000000000000000000000000000000000000000000000000',
        homeState: {
            chainId: 1n,
            token: zeroAddress,
            decimals: 18,
            userAllocation: parseEther('1.0'),
            userNetFlow: 0n,
            nodeAllocation: 0n,
            nodeNetFlow: 0n,
        },
        nonHomeState: {
            chainId: 1n,
            token: zeroAddress,
            decimals: 18,
            userAllocation: 0n,
            userNetFlow: 0n,
            nodeAllocation: 0n,
            nodeNetFlow: 0n,
        },
    },
    counterpartySig: '0x...', // Node signature
});

// 3. Combined deposit and create
const combinedTx = await client.depositAndCreateChannel(
    nodeAddress,
    zeroAddress,
    parseEther('1.0'),
    channelParams
);

// 4. Challenge a channel
const challengeTx = await client.challengeChannel({
    channelId: '0x...',
    state: updatedState,
    proofs: [],
});

// 5. Close a channel
const closeTx = await client.closeChannel({
    channelId: '0x...',
    state: finalState,
    proofs: [],
    counterpartySig: '0x...',
});

// 6. Withdraw from vault
const withdrawTx = await client.withdraw(
    recipientAddress,
    zeroAddress,
    parseEther('1.0')
);
```

## API Reference

### NitroliteClient

High-level client with automatic state signing and transaction execution.

#### Methods

| Method | Description | Returns |
|--------|-------------|---------|
| `deposit(node, token, amount)` | Deposit tokens/ETH to vault | `Promise<Hash>` |
| `createChannel(params)` | Create a new channel | `Promise<Hash>` |
| `depositAndCreateChannel(node, token, amount, params)` | Deposit and create channel in one flow | `Promise<Hash>` |
| `checkpointChannel(params)` | Checkpoint state on-chain | `Promise<Hash>` |
| `challengeChannel(params)` | Challenge a channel state | `Promise<Hash>` |
| `closeChannel(params)` | Close channel cooperatively or after challenge | `Promise<Hash>` |
| `withdraw(to, token, amount)` | Withdraw from vault | `Promise<Hash>` |
| `getOpenChannels()` | Get user's open channel IDs | `Promise<ChannelId[]>` |
| `getAccountBalance(node, token)` | Get vault balance | `Promise<bigint>` |
| `getChannelData(channelId)` | Get channel data | `Promise<ChannelData>` |
| `approveTokens(token, amount)` | Approve tokens for custody contract | `Promise<Hash>` |
| `getTokenAllowance(token)` | Get current token allowance | `Promise<bigint>` |
| `getTokenBalance(token)` | Get user token balance | `Promise<bigint>` |

### NitroliteService

Mid-level service for direct contract interactions.

#### Methods

All methods come in two variants:
- `prepare*()` - Returns prepared transaction data (for batching/AA)
- Direct method - Executes transaction immediately

Example:
```typescript
// Prepare transaction
const tx = await service.prepareDeposit(node, token, amount);

// Execute transaction
const hash = await service.deposit(node, token, amount);
```

Available operations:
- `deposit` / `prepareDeposit`
- `withdraw` / `prepareWithdraw`
- `createChannel` / `prepareCreateChannel`
- `depositToChannel` / `prepareDepositToChannel`
- `withdrawFromChannel` / `prepareWithdrawFromChannel`
- `checkpointChannel` / `prepareCheckpointChannel`
- `challengeChannel` / `prepareChallengeChannel`
- `closeChannel` / `prepareCloseChannel`

### NitroliteTransactionPreparer

Low-level transaction preparation for custom flows.

#### Methods

| Method | Description | Returns |
|--------|-------------|---------|
| `prepareDepositTransactions(node, token, amount)` | Prepare deposit with approval if needed | `Promise<PreparedTransaction[]>` |
| `prepareCreateChannelTransaction(params)` | Prepare channel creation | `Promise<PreparedTransaction>` |
| `prepareCheckpointChannelTransaction(params)` | Prepare checkpoint | `Promise<PreparedTransaction>` |
| `prepareChallengeChannelTransaction(params)` | Prepare challenge | `Promise<PreparedTransaction>` |
| `prepareCloseChannelTransaction(params)` | Prepare close | `Promise<PreparedTransaction>` |
| `prepareWithdrawTransaction(to, token, amount)` | Prepare withdrawal | `Promise<PreparedTransaction>` |
| `prepareDepositAndCreateChannelTransactions(...)` | Prepare deposit + create in batch | `Promise<PreparedTransaction[]>` |

## Breaking Changes from V0

### Contract Structure Changes

**V0 Structure:**
```typescript
// Channel was identified by user + asset + node
interface Allocation {
    user: bigint;
    node: bigint;
}

interface Channel {
    user: Address;
    node: Address;
    asset: Address;
    challengeDuration: number;
    allocation: Allocation;
}
```

**V1 Structure:**
```typescript
// Channel identified by unique ID from definition hash
interface ChannelDefinition {
    challengeDuration: number;
    user: Address;
    node: Address;
    nonce: bigint;
    metadata: Hex; // bytes32
}

interface Ledger {
    chainId: bigint;
    token: Address;
    decimals: number;
    userAllocation: bigint;
    userNetFlow: bigint;
    nodeAllocation: bigint;
    nodeNetFlow: bigint;
}

interface State {
    version: bigint;
    intent: StateIntent;
    metadata: Hex;
    homeState: Ledger;
    nonHomeState: Ledger;
    userSig: Hex;
    nodeSig: Hex;
}
```

### Removed Methods

The following V0 methods are no longer available:

#### From NitroliteClient
```typescript
// Removed
joinChannel()           // No longer needed - channels created directly
depositAndCreate()      // Replaced with depositAndCreateChannel()
resizeChannel()         // Use state transitions instead
getChannelBalance()     // Use getAccountBalance() for vault balance

// Use Instead
createChannel()         // Direct channel creation
depositAndCreateChannel() // Combined deposit + create
getAccountBalance()     // Vault balance by node + token
```

#### From Contract Functions
```typescript
// V0 Functions (removed)
join(channelId, allocation)
depositAndCreate(user, node, asset, userDeposit, nodeDeposit)
resize(channelId, newAllocation)

// V1 Functions (current)
createChannel(definition, initialState)
depositToVault(node, token, amount)
withdrawFromVault(to, token, amount)
depositToChannel(channelId, amount)
withdrawFromChannel(channelId, amount)
```

### New Concepts in V1

#### 1. Channel Definition with Nonce
Channels now require a unique nonce for identification:

```typescript
import { generateChannelNonce, getChannelId } from '@erc7824/nitrolite';

// Generate unique nonce
const nonce = generateChannelNonce(userAddress);

// Create definition
const definition: ChannelDefinition = {
    challengeDuration: 3600,
    user: userAddress,
    node: nodeAddress,
    nonce,
    metadata: '0x0000000000000000000000000000000000000000000000000000000000000000',
};

// Calculate channel ID
const channelId = getChannelId(definition, chainId);
```

#### 2. Ledger-Based States
States now contain separate home and non-home ledgers:

```typescript
const state: State = {
    version: 0n,
    intent: StateIntent.None,
    metadata: '0x0000000000000000000000000000000000000000000000000000000000000000',
    homeState: {
        chainId: 1n,
        token: tokenAddress,
        decimals: 18,
        userAllocation: parseEther('1.0'),
        userNetFlow: 0n,
        nodeAllocation: 0n,
        nodeNetFlow: 0n,
    },
    nonHomeState: {
        chainId: 1n,
        token: tokenAddress,
        decimals: 18,
        userAllocation: 0n,
        userNetFlow: 0n,
        nodeAllocation: 0n,
        nodeNetFlow: 0n,
    },
    userSig: '0x...',
    nodeSig: '0x...',
};
```

#### 3. Vault-Based Balance Management
Balances are now tracked per node + token in a vault:

```typescript
// V0: Channel-specific balance
const balance = await client.getChannelBalance(channelId);

// V1: Vault balance by node + token
const balance = await client.getAccountBalance(nodeAddress, tokenAddress);
```

#### 4. State Intent
States now include an intent field:

```typescript
enum StateIntent {
    None = 0,
    Challenge = 1,
    Final = 2,
}
```

### Migration Guide

#### Before (V0):
```typescript
// Create channel with initial deposit
const tx = await client.depositAndCreate(
    userAddress,
    nodeAddress,
    tokenAddress,
    parseEther('1.0'),
    0n
);

// Get channel balance
const balance = await client.getChannelBalance(channelId);
```

#### After (V1):
```typescript
// Option 1: Separate deposit and create
await client.deposit(nodeAddress, tokenAddress, parseEther('1.0'));
await client.createChannel({
    definition: { challengeDuration, user, node, nonce, metadata },
    state: initialState,
    counterpartySig: nodeSig,
});

// Option 2: Combined deposit and create
await client.depositAndCreateChannel(
    nodeAddress,
    tokenAddress,
    parseEther('1.0'),
    { definition, state, counterpartySig }
);

// Get vault balance
const balance = await client.getAccountBalance(nodeAddress, tokenAddress);
```

### Type Changes

**Removed Types:**
- `Allocation`
- `Channel`
- `UnsignedState` (V0 version)
- `LegacyState`

**New Types:**
- `ChannelDefinition`
- `Ledger`
- `State` (V1 version)
- `UnsignedStateV1`
- `ChannelId` (now a branded type)
- `StateHash` (branded type)

**Changed Types:**
- `ChannelData` now includes:
  ```typescript
  interface ChannelData {
      status: number;
      definition: ChannelDefinition; // was Channel
      lastState: State;              // new structure
      challengeExpiry: bigint;
  }
  ```

### Parameter Changes

#### createChannel()

**Before (V0):**
```typescript
createChannel(
    user: Address,
    node: Address,
    asset: Address,
    userDeposit: bigint,
    nodeDeposit: bigint
)
```

**After (V1):**
```typescript
createChannel({
    definition: ChannelDefinition,
    state: UnsignedStateV1,
    counterpartySig: Hex
})
```

#### deposit()

**Before (V0):**
```typescript
deposit(amount: bigint, asset: Address)
```

**After (V1):**
```typescript
deposit(node: Address, tokenAddress: Address, amount: bigint)
```

## Advanced Usage

### Account Abstraction with Prepared Transactions

```typescript
import { NitroliteTransactionPreparer } from '@erc7824/nitrolite';

// Create preparer
const preparer = new NitroliteTransactionPreparer({
    nitroliteService,
    erc20Service,
    addresses,
    account,
    walletClient,
    stateSigner,
    challengeDuration,
    chainId,
});

// Prepare multiple transactions
const depositTxs = await preparer.prepareDepositTransactions(
    nodeAddress,
    tokenAddress,
    amount
);

const createTx = await preparer.prepareCreateChannelTransaction(params);

// Batch into UserOperation
const userOp = createUserOp([...depositTxs, createTx]);
await sendUserOperation(userOp);
```

### Custom State Signing

```typescript
import { WalletStateSigner, SessionKeyStateSigner } from '@erc7824/nitrolite';

// Option 1: Wallet-based signer
const walletSigner = new WalletStateSigner(walletClient);

// Option 2: Session key signer
const sessionKeySigner = new SessionKeyStateSigner(sessionKeyAccount);

// Option 3: Custom signer
const customSigner = {
    signState: async (stateHash: Hex): Promise<Hex> => {
        // Your custom signing logic
        return signature;
    },
};
```

### State Utilities

```typescript
import {
    getPackedState,
    getStateHash,
    getChallengeHash,
    verifySignature,
} from '@erc7824/nitrolite';

// Pack state for hashing
const packed = getPackedState(channelId, unsignedState);

// Get state hash
const stateHash = getStateHash(channelId, unsignedState);

// Get challenge hash
const challengeHash = getChallengeHash(channelId, unsignedState);

// Verify signature
const isValid = await verifySignature(
    channelId,
    unsignedState,
    signature,
    expectedSigner
);
```

## Error Handling

The SDK provides typed errors:

```typescript
import * as Errors from '@erc7824/nitrolite';

try {
    await client.createChannel(params);
} catch (error) {
    if (error instanceof Errors.ContractCallError) {
        console.error('Contract call failed:', error.operation, error.context);
    } else if (error instanceof Errors.TransactionError) {
        console.error('Transaction failed:', error.message);
    } else if (error instanceof Errors.InvalidParameterError) {
        console.error('Invalid parameter:', error.message);
    }
}
```

## Security Considerations

1. **Challenge Duration**: Set appropriate challenge duration (minimum 3600 seconds)
2. **State Validation**: Always verify state signatures before submitting
3. **Nonce Management**: Use unique nonces for channel creation to prevent collisions
4. **Allowance Management**: Approve only necessary amounts for token operations
5. **State Intent**: Ensure correct intent when creating challenge/final states

## Testing

```typescript
import { NitroliteClient } from '@erc7824/nitrolite';
import { expect } from 'chai';

describe('NitroliteClient', () => {
    it('should create a channel', async () => {
        const tx = await client.createChannel(params);
        expect(tx).to.match(/^0x[a-fA-F0-9]{64}$/);

        const receipt = await client.waitForTransaction(tx);
        expect(receipt.status).to.equal('success');
    });
});
```

## Related Documentation

- [RPC Module](../rpc/README.md) - Off-chain communication with Clearnode
- [Main SDK Documentation](../../README.md) - Overview and quick start

## Support

For issues or questions:
- GitHub Issues: [github.com/erc7824/nitrolite/issues](https://github.com/erc7824/nitrolite/issues)
- Documentation: [erc7824.org](https://erc7824.org)
