# remittanceAdjudicator Contract

Complete reference for the remittanceAdjudicator smart contract with 2 functions, 0 events, and 3 custom errors.

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

### `compare`

Compares two states to determine which is more recent or valid.

**Type:** `pure`

**Parameters:**

- **`candidate`**: The proposed new state for the channel
- **`previous`**: tuple value

**Returns:**

- **`result`** (`int8`): int8

**Example Usage:**

```typescript
// Read compare from contract
const result = await publicClient.readContract({
    address: contractAddress,
    abi: contractAbi,
    functionName: 'compare',
    args: [candidate, previous],
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
import { remittanceadjudicatorAbi } from '@erc7824/nitrolite';

// Full type safety with autocomplete
const result = await publicClient.readContract({
  address: contractAddress,
  abi: remittanceadjudicatorAbi,
  functionName: 'functionName', // ✅ Autocomplete available
  args: [...], // ✅ Type-checked arguments
});
```
