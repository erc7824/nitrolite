# custody Contract

Complete reference for the custody smart contract with 11 functions, 9 events, and 18 custom errors.

## Functions

### `challenge`

Initiates a challenge against a state channel, disputing the current state with evidence of a more recent valid state.

**Type:** `nonpayable`

**Parameters:**

- **`channelId`**: Unique 32-byte identifier
- **`candidate`**: The proposed new state for the channel
- **`proofs`**: Supporting states that prove the transition is valid

**Example Usage:**

```typescript
// Execute challenge transaction
const { request } = await publicClient.simulateContract({
    address: contractAddress,
    abi: contractAbi,
    functionName: 'challenge',
    args: [channelId, candidate, proofs],
});

// Execute the transaction
const hash = await walletClient.writeContract(request);
```

### `checkpoint`

Updates the channel to a new agreed-upon state, typically used to progress the channel without disputes.

**Type:** `nonpayable`

**Parameters:**

- **`channelId`**: Unique 32-byte identifier
- **`candidate`**: The proposed new state for the channel
- **`proofs`**: Supporting states that prove the transition is valid

**Example Usage:**

```typescript
// Execute checkpoint transaction
const { request } = await publicClient.simulateContract({
    address: contractAddress,
    abi: contractAbi,
    functionName: 'checkpoint',
    args: [channelId, candidate, proofs],
});

// Execute the transaction
const hash = await walletClient.writeContract(request);
```

### `close`

contract Custody is IChannel, IDeposit {

**Type:** `nonpayable`

**Parameters:**

- **`channelId`**: Unique 32-byte identifier
- **`candidate`**: The proposed new state for the channel
- **``**: The token contract address (use address(0) for ETH)

**Example Usage:**

```typescript
// Execute close transaction
const { request } = await publicClient.simulateContract({
    address: contractAddress,
    abi: contractAbi,
    functionName: 'close',
    args: [channelId, candidate],
});

// Execute the transaction
const hash = await walletClient.writeContract(request);
```

### `create`

contract Custody is IChannel, IDeposit {

**Type:** `nonpayable`

**Parameters:**

- **`ch`**: Unique identifier for the state channel
- **`initial`**: The starting state when creating a new channel

**Returns:**

- **`channelId`** (`bytes32`): bytes32

**Example Usage:**

```typescript
// Execute create transaction
const { request } = await publicClient.simulateContract({
    address: contractAddress,
    abi: contractAbi,
    functionName: 'create',
    args: [ch, initial],
});

// Execute the transaction
const hash = await walletClient.writeContract(request);
```

### `deposit`

Deposits tokens or ETH into the custody contract for use in state channels.

**Type:** `payable`

**Parameters:**

- **`token`**: The token contract address (use address(0) for ETH)
- **`amount`**: Amount in the token's smallest unit (wei for ETH, etc.)

**Real Usage Examples:**

```typescript
// Log client methods for deposit
console.log('[depositToChannel] Available client methods:', Object.keys(client));

const amountBigInt =
    typeof amount === 'string' && !amount.startsWith('0x') ? parseTokenUnits(tokenAddress, amount) : BigInt(amount);
```

### `getAccountChannels`

contract Custody is IChannel, IDeposit {

**Type:** `view`

**Parameters:**

- **`account`**: Ethereum address to query information for

**Returns:**

- **``** (`bytes32[]`): bytes32[]

**Real Usage Examples:**

```typescript
}

    async getAccountChannels() {
        if (!this.client || !this.isConnected) {
            console.error("ClearNet client not initialized");
```

### `getAccountInfo`

contract Custody is IChannel, IDeposit {

**Type:** `view`

**Parameters:**

- **`user`**: Ethereum address of the user account
- **`token`**: The token contract address (use address(0) for ETH)

**Returns:**

- **`available`** (`uint256`): uint256
- **`channelCount`** (`uint256`): uint256

**Example Usage:**

```typescript
// Read getAccountInfo from contract
const result = await publicClient.readContract({
    address: contractAddress,
    abi: contractAbi,
    functionName: 'getAccountInfo',
    args: [user, token],
});
```

### `getContractInfo`

/ Newly added function to test type chain

**Type:** `pure`

**Parameters:**
None

**Returns:**

- **`version`** (`string`): string
- **`maxParticipants`** (`uint256`): uint256
- **`minChallengePeriod`** (`uint256`): uint256

**Example Usage:**

```typescript
// Read getContractInfo from contract
const result = await publicClient.readContract({
    address: contractAddress,
    abi: contractAbi,
    functionName: 'getContractInfo',
});
```

### `join`

contract Custody is IChannel, IDeposit {

**Type:** `nonpayable`

**Parameters:**

- **`channelId`**: Unique 32-byte identifier
- **`index`**: Numeric value (in smallest units)
- **`sig`**: tuple value

**Returns:**

- **``** (`bytes32`): bytes32

**Real Usage Examples:**

```typescript
console.error("Failed to hash state:", error);
            // Return a mock hash if there's an error
            return "0x" + Array(64).fill("0").join("") as Hex;
        }
    }
```

### `resize`

contract Custody is IChannel, IDeposit {

**Type:** `nonpayable`

**Parameters:**

- **`channelId`**: Unique 32-byte identifier
- **`candidate`**: The proposed new state for the channel
- **`proofs`**: Supporting states that prove the transition is valid

**Example Usage:**

```typescript
// Execute resize transaction
const { request } = await publicClient.simulateContract({
    address: contractAddress,
    abi: contractAbi,
    functionName: 'resize',
    args: [channelId, candidate, proofs],
});

// Execute the transaction
const hash = await walletClient.writeContract(request);
```

### `withdraw`

Withdraws available funds from the custody contract back to the user's wallet.

**Type:** `nonpayable`

**Parameters:**

- **`token`**: The token contract address (use address(0) for ETH)
- **`amount`**: Amount in the token's smallest unit (wei for ETH, etc.)

**Example Usage:**

```typescript
// Execute withdraw transaction
const { request } = await publicClient.simulateContract({
    address: contractAddress,
    abi: contractAbi,
    functionName: 'withdraw',
    args: [token, amount],
});

// Execute the transaction
const hash = await walletClient.writeContract(request);
```

## Events

### `Challenged`

**Parameters:**

- **`channelId`** (`bytes32`): bytes32
- **`expiration`** (`uint256`): uint256

### `Checkpointed`

**Parameters:**

- **`channelId`** (`bytes32`): bytes32

### `Closed`

**Parameters:**

- **`channelId`** (`bytes32`): bytes32
- **`finalState`** (`tuple (intent: uint8, version: uint256, data: bytes, allocations: tuple[], sigs: tuple[])`): struct State

### `Created`

**Parameters:**

- **`channelId`** (`bytes32`): bytes32
- **`wallet`** (`address`): address
- **`channel`** (`tuple (participants: address[], adjudicator: address, challenge: uint64, nonce: uint64)`): struct Channel
- **`initial`** (`tuple (intent: uint8, version: uint256, data: bytes, allocations: tuple[], sigs: tuple[])`): struct State

### `Deposited`

**Parameters:**

- **`wallet`** (`address`): address
- **`token`** (`address`): address
- **`amount`** (`uint256`): uint256

### `Joined`

**Parameters:**

- **`channelId`** (`bytes32`): bytes32
- **`index`** (`uint256`): uint256

### `Opened`

**Parameters:**

- **`channelId`** (`bytes32`): bytes32

### `Resized`

**Parameters:**

- **`channelId`** (`bytes32`): bytes32
- **`deltaAllocations`** (`int256[]`): int256[]

### `Withdrawn`

**Parameters:**

- **`wallet`** (`address`): address
- **`token`** (`address`): address
- **`amount`** (`uint256`): uint256

## Errors

### `ChallengeNotExpired`

### `ChannelNotFinal`

### `ChannelNotFound`

**Parameters:**

- **`channelId`** (`bytes32`): bytes32

### `ECDSAInvalidSignature`

### `ECDSAInvalidSignatureLength`

**Parameters:**

- **`length`** (`uint256`): uint256

### `ECDSAInvalidSignatureS`

**Parameters:**

- **`s`** (`bytes32`): bytes32

### `InsufficientBalance`

**Parameters:**

- **`available`** (`uint256`): uint256
- **`required`** (`uint256`): uint256

### `InvalidAdjudicator`

### `InvalidAllocations`

### `InvalidAmount`

### `InvalidChallengePeriod`

### `InvalidParticipant`

### `InvalidState`

### `InvalidStateSignatures`

### `InvalidStatus`

### `InvalidValue`

### `SafeERC20FailedOperation`

**Parameters:**

- **`token`** (`address`): address

### `TransferFailed`

**Parameters:**

- **`token`** (`address`): address
- **`to`** (`address`): address
- **`amount`** (`uint256`): uint256

## Type Safety

This contract is fully type-safe when used with the generated TypeScript types:

```typescript
import { custodyAbi } from '@erc7824/nitrolite';

// Full type safety with autocomplete
const result = await publicClient.readContract({
  address: contractAddress,
  abi: custodyAbi,
  functionName: 'functionName', // ✅ Autocomplete available
  args: [...], // ✅ Type-checked arguments
});
```
