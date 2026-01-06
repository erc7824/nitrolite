# dummy Contract

Complete reference for the dummy smart contract with 2 functions, 0 events, and 0 custom errors.

## Functions

### `adjudicate`

Determines the validity of a state transition or resolves disputes between channel participants.

**Type:** `pure`

**Parameters:**

- **``**: The token contract address (use address(0) for ETH)
- **``**: The token contract address (use address(0) for ETH)
- **``**: The token contract address (use address(0) for ETH)

**Returns:**

- **`valid`** (`bool`): bool

**Example Usage:**

```typescript
// Read adjudicate from contract
const result = await publicClient.readContract({
    address: contractAddress,
    abi: contractAbi,
    functionName: 'adjudicate',
    args: [, ,],
});
```

### `compare`

Compares two states to determine which is more recent or valid.

**Type:** `pure`

**Parameters:**

- **``**: The token contract address (use address(0) for ETH)
- **``**: The token contract address (use address(0) for ETH)

**Returns:**

- **`result`** (`int8`): int8

**Example Usage:**

```typescript
// Read compare from contract
const result = await publicClient.readContract({
    address: contractAddress,
    abi: contractAbi,
    functionName: 'compare',
    args: [,],
});
```

## Events

No events defined.

## Errors

No custom errors defined.

## Type Safety

This contract is fully type-safe when used with the generated TypeScript types:

```typescript
import { dummyAbi } from '@erc7824/nitrolite';

// Full type safety with autocomplete
const result = await publicClient.readContract({
  address: contractAddress,
  abi: dummyAbi,
  functionName: 'functionName', // ✅ Autocomplete available
  args: [...], // ✅ Type-checked arguments
});
```
