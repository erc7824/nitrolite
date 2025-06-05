# consensus Contract

Complete reference for the consensus smart contract with 1 functions, 0 events, and 3 custom errors.

## Functions

### `adjudicate`

Determines the validity of a state transition or resolves disputes between channel participants.

**Type:** `view`

**Parameters:**

- **`chan`**: Unique identifier for the state channel
- **`candidate`**: The proposed new state for the channel
- **`proofs`**: Supporting states that prove the transition is valid

**Returns:**

- **`valid`** (`bool`): bool

**Example Usage:**

```typescript
// Read adjudicate from contract
const result = await publicClient.readContract({
    address: contractAddress,
    abi: contractAbi,
    functionName: 'adjudicate',
    args: [chan, candidate, proofs],
});
```

## Events

No events defined.

## Errors

### `ECDSAInvalidSignature`

### `ECDSAInvalidSignatureLength`

**Parameters:**

- **`length`** (`uint256`): uint256

### `ECDSAInvalidSignatureS`

**Parameters:**

- **`s`** (`bytes32`): bytes32

## Type Safety

This contract is fully type-safe when used with the generated TypeScript types:

```typescript
import { consensusAbi } from '@erc7824/nitrolite';

// Full type safety with autocomplete
const result = await publicClient.readContract({
  address: contractAddress,
  abi: consensusAbi,
  functionName: 'functionName', // ✅ Autocomplete available
  args: [...], // ✅ Type-checked arguments
});
```
