---
sidebar_position: 1
title: Abstract Accounts
description: Using NitroliteClient with Account Abstraction
keywords: [erc7824, statechannels, state channels, nitrolite, ethereum scaling, layer 2, off-chain, account abstraction, ERC-4337]
---

import MethodDetails from '@site/src/components/MethodDetails';
import { Card, CardGrid } from '@site/src/components/Card';

# Using with Abstract Accounts

The `NitroliteClient` provides special support for [ERC-4337 Account Abstraction](https://eips.ethereum.org/EIPS/eip-4337) through the `txPreparer` object. This allows dApps using smart contract wallets to prepare transactions without executing them, enabling batching and other advanced patterns.

## Transaction Preparer Overview

The `txPreparer` is a property of the `NitroliteClient` that provides methods for preparing transaction data without sending it to the blockchain. Each method returns one or more [`PreparedTransaction`](../types.md#preparedtransaction) objects that can be used with Account Abstraction providers.

```typescript
import { NitroliteClient } from '@erc7824/nitrolite';

const client = new NitroliteClient({/* config */});

// Instead of: await client.deposit(amount)
const txs = await client.txPreparer.prepareDepositTransactions(amount);
```

## Transaction Preparation Methods

These methods allow you to prepare transactions for the entire channel lifecycle without executing them.

### 1. Deposit Methods

<MethodDetails
  name="prepareDepositTransactions"
  description="Prepares deposit transactions for sending tokens to the custody contract. May include an ERC-20 approval transaction if the current allowance is insufficient. Returns an array of transactions that must all be executed for the deposit to succeed."
  params={[{ name: "amount", type: "bigint" }]}
  returns={`Promise<[PreparedTransaction](../types.md#preparedtransaction)[]>`}
  example={`// Prepare deposit transaction(s) - may include ERC20 approval
const txs = await client.txPreparer.prepareDepositTransactions(1000000n);

// For each prepared transaction
for (const tx of txs) {
  await aaProvider.sendUserOperation({
    target: tx.to,
    data: tx.data,
    value: tx.value || 0n
  });
}`}
/>

<MethodDetails
  name="prepareApproveTokensTransaction"
  description="Prepares a transaction to approve the custody contract to spend ERC-20 tokens. This is useful when you want to separate approval from the actual deposit operation or when implementing a custom approval flow."
  params={[{ name: "amount", type: "bigint" }]}
  returns={`Promise<[PreparedTransaction](../types.md#preparedtransaction)>`}
  example={`// Prepare approval transaction
const tx = await client.txPreparer.prepareApproveTokensTransaction(2000000n);

// Send through your AA provider
await aaProvider.sendUserOperation({
  target: tx.to,
  data: tx.data
});`}
/>

### 2. Channel Creation Methods

<MethodDetails
  name="prepareCreateChannelTransaction"
  description="Prepares a transaction for creating a new state channel with the specified initial allocation. This transaction calls the custody contract to establish a new channel with the given parameters without executing it immediately."
  params={[{ name: "params", type: "[CreateChannelParams](../types.md#2-channel-creation)" }]}
  returns={`Promise<[PreparedTransaction](../types.md#preparedtransaction)>`}
  example={`// Prepare channel creation transaction
const tx = await client.txPreparer.prepareCreateChannelTransaction({
  initialAllocationAmounts: [700000n, 300000n],
  stateData: '0x1234'
});

// Send it through your Account Abstraction provider
await aaProvider.sendUserOperation({
  target: tx.to,
  data: tx.data,
  value: tx.value || 0n
});`}
/>

<MethodDetails
  name="prepareDepositAndCreateChannelTransactions"
  description="Combines deposit and channel creation into a sequence of prepared transactions. This is ideal for batching with Account Abstraction to create a streamlined onboarding experience. Returns an array of transactions that may include token approval, deposit, and channel creation."
  params={[
    { name: "depositAmount", type: "bigint" },
    { name: "params", type: "[CreateChannelParams](../types.md#2-channel-creation)" }
  ]}
  returns={`Promise<[PreparedTransaction](../types.md#preparedtransaction)[]>`}
  example={`// Prepare deposit + channel creation (potentially 3 txs: approve, deposit, create)
const txs = await client.txPreparer.prepareDepositAndCreateChannelTransactions(
  1000000n,
  {
    initialAllocationAmounts: [700000n, 300000n],
    stateData: '0x1234'
  }
);

// Bundle these transactions into a single UserOperation
await aaProvider.sendUserOperation({
  userOperations: txs.map(tx => ({
    target: tx.to,
    data: tx.data,
    value: tx.value || 0n
  }))
});`}
/>

### 3. Channel Operation Methods

<MethodDetails
  name="prepareCheckpointChannelTransaction"
  description="Prepares a transaction to checkpoint a channel state on-chain. This creates an immutable record of the channel state that both parties have agreed to, which is useful for security and dispute resolution."
  params={[{ name: "params", type: "[CheckpointChannelParams](../types.md#3-channel-operations)" }]}
  returns={`Promise<[PreparedTransaction](../types.md#preparedtransaction)>`}
  example={`// Prepare checkpoint transaction
const tx = await client.txPreparer.prepareCheckpointChannelTransaction({
  channelId: '0x...',
  candidateState: state
});

await aaProvider.sendUserOperation({
  target: tx.to,
  data: tx.data
});`}
/>

<MethodDetails
  name="prepareChallengeChannelTransaction"
  description="Prepares a transaction to challenge a channel with a candidate state. This is used when the counterparty is unresponsive, allowing you to force progress in the dispute resolution process."
  params={[{ name: "params", type: "[ChallengeChannelParams](../types.md#3-channel-operations)" }]}
  returns={`Promise<[PreparedTransaction](../types.md#preparedtransaction)>`}
  example={`// Prepare challenge transaction
const tx = await client.txPreparer.prepareChallengeChannelTransaction({
  channelId: '0x...',
  candidateState: state
});

await aaProvider.sendUserOperation({
  target: tx.to,
  data: tx.data
});`}
/>

<MethodDetails
  name="prepareResizeChannelTransaction"
  description="Prepares a transaction to adjust the total funds allocated to a channel. This allows you to add more funds to a channel that's running low or reduce locked funds when less capacity is needed."
  params={[{ name: "params", type: "[ResizeChannelParams](../types.md#3-channel-operations)" }]}
  returns={`Promise<[PreparedTransaction](../types.md#preparedtransaction)>`}
  example={`// Prepare resize transaction
const tx = await client.txPreparer.prepareResizeChannelTransaction({
  channelId: '0x...',
  candidateState: state
});

await aaProvider.sendUserOperation({
  target: tx.to,
  data: tx.data
});`}
/>

### 4. Channel Closing Methods

<MethodDetails
  name="prepareCloseChannelTransaction"
  description="Prepares a transaction to close a channel on-chain using a mutually agreed final state. This transaction unlocks funds according to the agreed allocations and makes them available for withdrawal."
  params={[{ name: "params", type: "[CloseChannelParams](../types.md#4-channel-closing)" }]}
  returns={`Promise<[PreparedTransaction](../types.md#preparedtransaction)>`}
  example={`// Prepare close channel transaction
const tx = await client.txPreparer.prepareCloseChannelTransaction({
  finalState: {
    channelId: '0x...',
    stateData: '0x...',
    allocations: [allocation1, allocation2],
    version: 10n,
    serverSignature: signature
  }
});

await aaProvider.sendUserOperation({
  target: tx.to,
  data: tx.data
});`}
/>

### 5. Withdrawal Methods

<MethodDetails
  name="prepareWithdrawalTransaction"
  description="Prepares a transaction to withdraw tokens previously deposited into the custody contract back to the user's wallet. This is the final step in the channel lifecycle, allowing users to reclaim their funds after channels have been closed."
  params={[{ name: "amount", type: "bigint" }]}
  returns={`Promise<[PreparedTransaction](../types.md#preparedtransaction)>`}
  example={`// Prepare withdrawal transaction
const tx = await client.txPreparer.prepareWithdrawalTransaction(500000n);

await aaProvider.sendUserOperation({
  target: tx.to,
  data: tx.data
});`}
/>


## Understanding PreparedTransaction

The [`PreparedTransaction`](../types.md#preparedtransaction) type is the core data structure returned by all transaction preparation methods. It contains all the information needed to construct a transaction or UserOperation:

```typescript
type PreparedTransaction = {
  // Target contract address
  to: Address;

  // Contract call data
  data?: Hex;

  // ETH value to send (0n for token operations)
  value?: bigint;
};
```

Each `PreparedTransaction` represents a single contract call that can be:

1. **Executed directly** - If you're using a standard EOA wallet
2. **Bundled into a UserOperation** - For account abstraction providers
3. **Batched with other transactions** - For advanced use cases

## Integration Examples

The Nitrolite transaction preparer can be integrated with any Account Abstraction provider. Here are examples with popular AA SDKs:

### With Safe Account Abstraction SDK

```typescript
import { NitroliteClient } from '@erc7824/nitrolite';
import { AccountAbstraction } from '@safe-global/account-abstraction-kit-poc';

// Initialize clients
const client = new NitroliteClient({/* config */});
const aaKit = new AccountAbstraction(safeProvider);

// Prepare transaction
const tx = await client.txPreparer.prepareCreateChannelTransaction({
  initialAllocationAmounts: [700000n, 300000n],
  stateData: '0x1234'
});

// Send through AA provider
const safeTransaction = await aaKit.createTransaction({
  transactions: [{
    to: tx.to,
    data: tx.data,
    value: tx.value?.toString() || '0'
  }]
});

const txResponse = await aaKit.executeTransaction(safeTransaction);
```

### With Biconomy SDK

```typescript
import { NitroliteClient } from '@erc7824/nitrolite';
import { BiconomySmartAccountV2 } from "@biconomy/account";

// Initialize clients
const client = new NitroliteClient({/* config */});
const smartAccount = await BiconomySmartAccountV2.create({/* config */});

// Prepare transaction
const txs = await client.txPreparer.prepareDepositAndCreateChannelTransactions(
  1000000n,
  {
    initialAllocationAmounts: [700000n, 300000n],
    stateData: '0x1234'
  }
);

// Build user operation
const userOp = await smartAccount.buildUserOp(
  txs.map(tx => ({
    to: tx.to,
    data: tx.data,
    value: tx.value || 0n
  }))
);

// Send user operation
const userOpResponse = await smartAccount.sendUserOp(userOp);
await userOpResponse.wait();
```

## Advanced Use Cases

The transaction preparer is especially powerful when combined with advanced Account Abstraction features.

### Batching Multiple Operations

One of the main advantages of Account Abstraction is the ability to batch multiple operations into a single transaction:

```typescript
// Collect prepared transactions from different operations
const preparedTxs = [];

// 1. Add token approval if needed
const allowance = await client.getTokenAllowance();
if (allowance < totalNeeded) {
  const approveTx = await client.txPreparer.prepareApproveTokensTransaction(totalNeeded);
  preparedTxs.push(approveTx);
}

// 2. Add deposit
const depositTx = await client.txPreparer.prepareDepositTransactions(amount);
preparedTxs.push(...depositTx);

// 3. Add channel creation
const createChannelTx = await client.txPreparer.prepareCreateChannelTransaction(params);
preparedTxs.push(createChannelTx);

// 4. Execute all as a batch with your AA provider
await aaProvider.sendUserOperation({
  userOperations: preparedTxs.map(tx => ({
    target: tx.to,
    data: tx.data,
    value: tx.value || 0n
  }))
});
```

### Gas Sponsoring

Account Abstraction enables gas sponsorship, where someone else pays for the transaction gas:

```typescript
// Prepare transaction
const tx = await client.txPreparer.prepareCreateChannelTransaction(params);

// Use a sponsored transaction
await paymasterProvider.sponsorTransaction({
  target: tx.to,
  data: tx.data,
  value: tx.value || 0n,
  user: userAddress
});
```

### Session Keys

Some AA wallets support session keys, which are temporary keys with limited permissions:

```typescript
// Create a session key with permissions only for specific operations
const sessionKeyData = await aaWallet.createSessionKey({
  permissions: [
    {
      target: client.addresses.custody,
      // Only allow specific functions
      functionSelector: [
        "0xdeposit(...)",
        "0xwithdraw(...)"
      ]
    }
  ],
  expirationTime: Date.now() + 3600 * 1000 // 1 hour
});

// Use the session key to prepare and send transactions
const tx = await client.txPreparer.prepareDepositTransactions(amount);
await aaWallet.executeWithSessionKey(sessionKeyData, {
  target: tx.to,
  data: tx.data,
  value: tx.value || 0n
});
```

## Best Practices

<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
  <div className="p-4 border rounded-lg shadow-sm bg-gray-50 dark:bg-gray-800">
    <h3 className="text-lg font-medium">Batch Related Operations</h3>
    <p>Use <code>prepareDepositAndCreateChannelTransactions</code> to batch deposit and channel creation into a single user operation.</p>
  </div>
  
  <div className="p-4 border rounded-lg shadow-sm bg-gray-50 dark:bg-gray-800">
    <h3 className="text-lg font-medium">Handle Approvals</h3>
    <p>For ERC20 tokens, <code>prepareDepositTransactions</code> will include an approval transaction if needed. Always process all returned transactions.</p>
  </div>
  
  <div className="p-4 border rounded-lg shadow-sm bg-gray-50 dark:bg-gray-800">
    <h3 className="text-lg font-medium">State Signing</h3>
    <p>Even when using Account Abstraction, state signatures are handled separately using the <code>stateWalletClient</code> (or <code>walletClient</code> if not specified).</p>
  </div>
  
  <div className="p-4 border rounded-lg shadow-sm bg-gray-50 dark:bg-gray-800">
    <h3 className="text-lg font-medium">Error Handling</h3>
    <p>The preparation methods throw the same errors as their execution counterparts, so use the same error handling patterns.</p>
  </div>

  <div className="p-4 border rounded-lg shadow-sm bg-gray-50 dark:bg-gray-800">
    <h3 className="text-lg font-medium">Check Token Allowances</h3>
    <p>Before preparing token operations, you can check if approval is needed:</p>
    <pre className="bg-gray-100 dark:bg-gray-700 p-2 rounded-md mt-2 text-sm">
      <code className="language-typescript">
const allowance = await client.getTokenAllowance();
if (allowance < amount) {
  // Need approval
}
      </code>
    </pre>
  </div>
  
  <div className="p-4 border rounded-lg shadow-sm bg-gray-50 dark:bg-gray-800">
    <h3 className="text-lg font-medium">Gas Estimation</h3>
    <p>When using Account Abstraction, gas estimation is typically handled by the AA provider, but you can request estimates if needed.</p>
  </div>
</div>

## Limitations

:::caution Important
- The transaction preparer **doesn't handle sequencing or nonce management** - that's the responsibility of your AA provider.
- Some operations (like checkpointing) require signatures from all participants, which must be collected separately from the transaction preparation.
:::