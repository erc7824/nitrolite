---
sidebar_position: 3
title: Erc20Service
description: Documentation for the ERC20Service class
keywords: [erc7824, statechannels, state channels, nitrolite, ethereum scaling, layer 2, off-chain, advanced, erc20, tokens]
---

# Erc20Service

The `Erc20Service` class provides a convenient interface for interacting with ERC20 tokens. It handles token approvals, allowance checks, and balance inquiries that are essential for deposit and withdrawal operations in the Nitrolite system.

## Initialization

```typescript
import { Erc20Service } from '@erc7824/nitrolite';

const erc20Service = new Erc20Service(
  publicClient,  // viem PublicClient
  walletClient   // viem WalletClient
);
```

## Core Methods

| Method | Description | Parameters | Return Type |
|--------|-------------|------------|------------|
| `approve` | Approves a spender to use tokens. | `tokenAddress: Address, spender: Address, amount: bigint` | `Promise<Hash>` |
| `getTokenAllowance` | Gets token allowance for a spender. | `tokenAddress: Address, owner: Address, spender: Address` | `Promise<bigint>` |
| `getTokenBalance` | Gets token balance for an account. | `tokenAddress: Address, account: Address` | `Promise<bigint>` |

## Method Details

### Approve Tokens

Approves a spender (typically the Custody contract) to transfer tokens on behalf of the owner.

```typescript
// Approve the custody contract to spend 1000 tokens
const txHash = await erc20Service.approve(
  tokenAddress,  // ERC20 token address
  spenderAddress, // Custody contract address
  1000000000000000000n // Amount to approve (1 token with 18 decimals)
);
```

**Important notes:**
- For security reasons, always specify the exact amount you want to approve
- Consider using the ERC20 token decimals for the amount calculation
- The transaction will fail if the owner has insufficient balance

### Get Token Allowance

Retrieves the current allowance granted by an owner to a spender.

```typescript
// Check current allowance
const allowance = await erc20Service.getTokenAllowance(
  tokenAddress,  // ERC20 token address
  ownerAddress,  // Owner's address
  spenderAddress // Spender's address (custody contract)
);

console.log(`Current allowance: ${allowance}`);

// Check if allowance is sufficient
if (allowance < requiredAmount) {
  console.log('Need to approve more tokens');
}
```

### Get Token Balance

Retrieves the token balance for a specific account.

```typescript
// Check token balance
const balance = await erc20Service.getTokenBalance(
  tokenAddress, // ERC20 token address
  accountAddress // Account to check
);

console.log(`Account balance: ${balance}`);

// Check if balance is sufficient
if (balance < requiredAmount) {
  console.log('Insufficient token balance');
}
```

## Transaction Preparation

For Account Abstraction support, `Erc20Service` provides a transaction preparation method:

| Method | Description | Parameters | Return Type |
|--------|-------------|------------|------------|
| `prepareApprove` | Prepares an approval transaction. | `tokenAddress: Address, spender: Address, amount: bigint` | `Promise<PreparedTransaction>` |

Example:
```typescript
// Prepare approval transaction
const tx = await erc20Service.prepareApprove(
  tokenAddress,
  spenderAddress,
  amount
);

// Use with your Account Abstraction provider
const userOp = await aaProvider.buildUserOperation({
  target: tx.to,
  data: tx.data,
  value: 0n // ERC20 approvals don't require ETH
});
```

## Implementation Details

The `Erc20Service` uses the standard ERC20 interface methods:

- `approve`: Allows a spender to withdraw tokens from the owner's account, up to the specified amount
- `allowance`: Returns the remaining tokens that a spender is allowed to withdraw
- `balanceOf`: Returns the token balance of the specified account

## Working with Token Decimals

ERC20 tokens typically have decimal places (most commonly 18). When working with token amounts, you should account for these decimals:

```typescript
import { parseUnits } from 'viem';

// For a token with 18 decimals
const tokenDecimals = 18;

// Convert 1.5 tokens to the smallest unit
const amount = parseUnits('1.5', tokenDecimals);

// Approve the amount
await erc20Service.approve(tokenAddress, spenderAddress, amount);
```

## Error Handling

The `Erc20Service` throws specific error types:

- `TokenError`: For token-specific errors (insufficient balance, approval failures)
- `ContractCallError`: When calls to the contract fail
- `WalletClientRequiredError`: When wallet client is needed but not provided

Example:
```typescript
try {
  await erc20Service.approve(tokenAddress, spenderAddress, amount);
} catch (error) {
  if (error instanceof TokenError) {
    console.error(`Token error: ${error.message}`);
    console.error(`Suggestion: ${error.suggestion}`);
    
    // Check for specific token error conditions
    if (error.details?.errorName === 'InsufficientBalance') {
      console.log(`Available balance: ${error.details.available}`);
    }
  }
}
```

## Common Patterns

### Checking and Approving Tokens

A common pattern is to check if the current allowance is sufficient before approving more tokens:

```typescript
// Get current allowance
const allowance = await erc20Service.getTokenAllowance(
  tokenAddress,
  ownerAddress,
  spenderAddress
);

// If allowance is insufficient, approve more tokens
if (allowance < requiredAmount) {
  await erc20Service.approve(tokenAddress, spenderAddress, requiredAmount);
}

// Now proceed with the operation that requires the approval
// (e.g., deposit into custody contract)
```

### Handling Multiple Tokens

If your application works with multiple tokens, you can reuse the same `Erc20Service` instance:

```typescript
// Same service instance for different tokens
const erc20Service = new Erc20Service(publicClient, walletClient);

// Work with token A
const balanceA = await erc20Service.getTokenBalance(tokenAddressA, accountAddress);

// Work with token B
const balanceB = await erc20Service.getTokenBalance(tokenAddressB, accountAddress);
```

## Integration with NitroliteClient

When using the `NitroliteClient`, you typically don't need to interact with `Erc20Service` directly, as the client handles these operations for you:

```typescript
// NitroliteClient handles token approvals automatically during deposit
await nitroliteClient.deposit(amount);

// For explicit approval without deposit
await nitroliteClient.approveTokens(amount);

// Get token balance through the client
const balance = await nitroliteClient.getTokenBalance();
```

However, for advanced use cases or custom token interaction, direct access to `Erc20Service` can be useful.