---
sidebar_position: 3
title: Advanced Usage
description: Advanced topics for working with the NitroliteClient
keywords: [erc7824, statechannels, state channels, nitrolite, ethereum scaling, layer 2, off-chain, advanced]
---

# Advanced Usage

This section covers advanced topics for working with the `@erc7824/nitrolite` SDK. These are intended for developers who need deeper control over the state channel operations or want to integrate with specialized systems.

## Available Topics

### Low-level Services

The NitroliteClient is built on top of specialized services that can be used directly for more fine-grained control:

- [NitroliteService](./nitrolite-service.md) - Core service for interacting with the custody contract and managing channels
- [Erc20Service](./erc20-service.md) - Service for interacting with ERC20 tokens

### Account Abstraction Integration

- [Abstract Accounts](./abstract-accounts.md) - Using NitroliteClient with ERC-4337 Account Abstraction for smart contract wallets

## When to Use Advanced Features

Consider using the advanced features when:

1. **Building custom workflows** that require more control than the high-level NitroliteClient API provides
2. **Integrating with smart contract wallets** or other account abstraction systems
3. **Implementing specialized monitoring or management systems** for state channels
4. **Developing cross-chain applications** that require custom handling of state channel operations
5. **Optimizing gas usage** through transaction batching and other techniques

## Example: Direct Service Usage

While using the high-level NitroliteClient is recommended for most applications, here's how you can work directly with the services:

```typescript
import { 
  NitroliteService, 
  Erc20Service, 
  getChannelId,
  signState 
} from '@erc7824/nitrolite';

// Initialize services
const nitroliteService = new NitroliteService(
  publicClient,
  addresses,
  walletClient,
  account.address
);

const erc20Service = new Erc20Service(
  publicClient,
  walletClient
);

// Check token allowance
const allowance = await erc20Service.getTokenAllowance(
  tokenAddress,
  account.address,
  addresses.custody
);

// Approve tokens if needed
if (allowance < depositAmount) {
  await erc20Service.approve(tokenAddress, addresses.custody, depositAmount);
}

// Deposit funds
await nitroliteService.deposit(tokenAddress, depositAmount);

// Create channel with custom parameters
const channelNonce = generateChannelNonce(account.address);
const channel = {
  participants: [account.address, counterpartyAddress],
  adjudicator: addresses.adjudicator,
  challenge: 100n, // Challenge duration in seconds
  nonce: channelNonce
};

// Prepare initial state
const initialState = {
  intent: StateIntent.INITIALIZE,
  version: 0n,
  data: '0x1234', // Application-specific data
  allocations: [
    { destination: account.address, token: tokenAddress, amount: 700000n },
    { destination: counterpartyAddress, token: tokenAddress, amount: 300000n }
  ],
  sigs: [] // Will be filled with signatures
};

// Sign the state
const channelId = getChannelId(channel);
const stateHash = getStateHash(channelId, initialState);
const signature = await signState(walletClient, stateHash);
initialState.sigs = [signature];

// Create the channel
await nitroliteService.createChannel(channel, initialState);
```

## Example: Complex Transaction Preparation

For applications using Account Abstraction, you can prepare complex transaction sequences:

```typescript
import { NitroliteClient, NitroliteTransactionPreparer } from '@erc7824/nitrolite';

// Initialize client
const client = new NitroliteClient({/* config */});

// Access the transaction preparer directly
const txPreparer = client.txPreparer;

// Prepare a complete sequence (approve, deposit, create channel)
const allTxs = [];

// 1. Check and prepare token approval if needed
const allowance = await client.getTokenAllowance();
if (allowance < depositAmount) {
  const approveTx = await txPreparer.prepareApproveTokensTransaction(depositAmount);
  allTxs.push(approveTx);
}

// 2. Prepare deposit
const depositTx = await txPreparer.prepareDepositTransactions(depositAmount);
allTxs.push(...depositTx);

// 3. Prepare channel creation
const createChannelTx = await txPreparer.prepareCreateChannelTransaction({
  initialAllocationAmounts: [700000n, 300000n],
  stateData: '0x1234'
});
allTxs.push(createChannelTx);

// 4. Use with your Account Abstraction provider
await aaProvider.sendUserOperation({
  userOperations: allTxs.map(tx => ({
    target: tx.to,
    data: tx.data,
    value: tx.value || 0n
  }))
});
```

These advanced techniques give you greater flexibility and control over the state channel operations, but they also require a deeper understanding of the underlying protocol.