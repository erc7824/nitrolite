# NitroliteClient API Documentation

## Table of Contents

- [API Reference](#api-reference)
  - [NitroliteClient Methods](#nitrolieclient-methods)
  - [Transaction Preparer Methods](#transaction-preparer-methods)
  - [State Preparation Methods](#state-preparation-methods)
  - [NitroliteService Methods](#nitroliteservice-methods)
  - [Erc20Service Methods](#erc20service-methods)
- [Architecture](#architecture)
- [Initialization](#initialization)
- [Core Methods](#core-methods)
  - [Channel Management](#channel-management)
  - [Fund Management](#fund-management)
  - [Account Information](#account-information)
  - [Token Operations](#token-operations)
- [Services](#services)
  - [NitroliteService](#nitroliteservice)
  - [Erc20Service](#erc20service)
- [Transaction Preparation](#transaction-preparation)
- [Advanced Usage](#advanced-usage)
  - [Custom State Signing](#custom-state-signing)
  - [Transaction Batching](#transaction-batching)
  - [Account Abstraction](#account-abstraction)
- [Type Definitions](#type-definitions)

## API Reference

### NitroliteClient Methods

| Method | Description | Parameters | Returns |
|--------|-------------|------------|---------|
| `createChannel` | Creates a new state channel | `params: CreateChannelParams` | `{ channelId: ChannelId; initialState: State; txHash: Hash }` |
| `depositAndCreateChannel` | Deposits tokens and creates a channel | `depositAmount: bigint`, `params: CreateChannelParams` | `{ channelId: ChannelId; initialState: State; depositTxHash: Hash; createChannelTxHash: Hash }` |
| `checkpointChannel` | Records a state on-chain | `params: CheckpointChannelParams` | `Hash` |
| `challengeChannel` | Challenges a channel with a state | `params: ChallengeChannelParams` | `Hash` |
| `closeChannel` | Closes a channel with final state | `params: CloseChannelParams` | `Hash` |
| `deposit` | Deposits tokens into custody contract | `amount: bigint` | `Hash` |
| `withdrawal` | Withdraws available tokens | `amount: bigint` | `Hash` |
| `getAccountChannels` | Gets channels for client's account | - | `Array<ChannelId>` |
| `getAccountInfo` | Gets account balance information | - | `AccountInfo` |
| `approveTokens` | Approves token spending | `amount: bigint` | `Hash` |
| `getTokenAllowance` | Gets current token allowance | - | `bigint` |
| `getTokenBalance` | Gets current token balance | - | `bigint` |

### Transaction Preparer Methods

| Method | Description | Parameters | Returns |
|--------|-------------|------------|---------|
| `prepareDepositTransactions` | Prepares deposit transactions (includes approval if needed) | `amount: bigint` | `Array<PreparedTransaction>` |
| `prepareCreateChannelTransaction` | Prepares channel creation transaction | `params: CreateChannelParams` | `PreparedTransaction` |
| `prepareDepositAndCreateChannelTransactions` | Prepares deposit and channel creation | `depositAmount: bigint`, `params: CreateChannelParams` | `Array<PreparedTransaction>` |
| `prepareCheckpointChannelTransaction` | Prepares checkpoint transaction | `params: CheckpointChannelParams` | `PreparedTransaction` |
| `prepareChallengeChannelTransaction` | Prepares challenge transaction | `params: ChallengeChannelParams` | `PreparedTransaction` |
| `prepareCloseChannelTransaction` | Prepares channel closure transaction | `params: CloseChannelParams` | `PreparedTransaction` |
| `prepareWithdrawalTransaction` | Prepares withdrawal transaction | `amount: bigint` | `PreparedTransaction` |
| `prepareApproveTokensTransaction` | Prepares token approval transaction | `amount: bigint` | `PreparedTransaction` |

### State Preparation Methods

| Method | Description | Parameters | Returns |
|--------|-------------|------------|---------|
| `prepareAndSignInitialState` | Creates and signs initial state | `params: CreateChannelParams` | `{ channel: Channel; initialState: State; channelId: ChannelId }` |
| `prepareAndSignFinalState` | Creates and signs final state | `params: CloseChannelParams` | `{ finalStateWithSigs: State; channelId: ChannelId }` |

### NitroliteService Methods

| Method | Description | Parameters | Returns |
|--------|-------------|------------|---------|
| `deposit` | Deposits tokens/ETH into custody contract | `tokenAddress: Address`, `amount: bigint` | `Hash` |
| `createChannel` | Creates a new channel | `channel: Channel`, `initialState: State` | `Hash` |
| `joinChannel` | Joins an existing channel | `channelId: ChannelId`, `index: bigint`, `signature: Signature` | `Hash` |
| `checkpoint` | Records state on-chain | `channelId: ChannelId`, `candidate: State`, `proofs?: State[]` | `Hash` |
| `challenge` | Challenges with a state | `channelId: ChannelId`, `candidate: State`, `proofs?: State[]` | `Hash` |
| `close` | Closes a channel | `channelId: ChannelId`, `candidate: State`, `proofs?: State[]` | `Hash` |
| `reset` | Closes and creates a new channel | `channelId: ChannelId`, `candidate: State`, `proofs: State[]`, `newChannel: Channel`, `newDeposit: State` | `Hash` |
| `withdraw` | Withdraws available tokens | `tokenAddress: Address`, `amount: bigint` | `Hash` |
| `getAccountChannels` | Gets channels for an account | `account: Address` | `Array<ChannelId>` |
| `getAccountInfo` | Gets account balance information | `user: Address`, `token: Address` | `{ available: bigint; locked: bigint; channelCount: bigint }` |

### Erc20Service Methods

| Method | Description | Parameters | Returns |
|--------|-------------|------------|---------|
| `getTokenBalance` | Gets token balance | `tokenAddress: Address`, `account: Address` | `bigint` |
| `getTokenAllowance` | Gets token allowance | `tokenAddress: Address`, `owner: Address`, `spender: Address` | `bigint` |
| `approve` | Approves token spending | `tokenAddress: Address`, `spender: Address`, `amount: bigint` | `Hash` |
| `prepareApprove` | Prepares approval transaction | `tokenAddress: Address`, `spender: Address`, `amount: bigint` | `PreparedTransaction` |

## Architecture

The `NitroliteClient` is the primary class for interacting with Nitrolite smart contracts. It provides a high-level API that abstracts the complexity of state channel operations, while offering access to the underlying services for advanced use cases.

The NitroliteClient follows a layered architecture:

1. **Client Layer** (`NitroliteClient`) - High-level API with error handling and business logic
2. **Service Layer** (`NitroliteService`, `Erc20Service`) - Mid-level API for contract interactions
3. **Transaction Preparation Layer** (`NitroliteTransactionPreparer`) - Low-level API for advanced use cases including Account Abstraction support

This design allows for both simplicity for common use cases and flexibility for advanced scenarios.

## Initialization

Create a NitroliteClient instance with your wallet and contract configuration:

```typescript
import { NitroliteClient } from '@erc7824/nitrolite';
import { createPublicClient, createWalletClient, http } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { mainnet } from 'viem/chains';

// Initialize viem clients
const publicClient = createPublicClient({
  chain: mainnet,
  transport: http('https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY')
});

const account = privateKeyToAccount('0xYOUR_PRIVATE_KEY');
const walletClient = createWalletClient({
  account,
  chain: mainnet,
  transport: http('https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY')
});

// Create a NitroliteClient instance
const client = new NitroliteClient({
  publicClient,
  walletClient,
  addresses: {
    custody: '0xYOUR_CUSTODY_CONTRACT_ADDRESS',
    adjudicators: {
      base: '0xBASE_ADJUDICATOR_ADDRESS',
      numeric: '0xNUMERIC_ADJUDICATOR_ADDRESS'
    },
    guestAddress: '0xGUEST_ADDRESS',  // Address of the counterparty
    tokenAddress: '0xTOKEN_ADDRESS'   // Address of the token (use '0x0' for ETH)
  },
  challengeDuration: BigInt(86400),  // 1 day challenge period
  
  // Optional: For using a different wallet for state signing
  stateWalletClient: customStateWalletClient
});
```

### Configuration Options

```typescript
interface NitroliteClientConfig {
  // Required: The viem public client for reading blockchain data
  publicClient: PublicClient;
  
  // Required: The viem wallet client for transactions and default state signing
  walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;
  
  // Optional: Separate wallet client for signing off-chain state updates
  // Useful for hot wallets or different security models
  stateWalletClient?: WalletClient<Transport, Chain, ParseAccount<Account>>;
  
  // Required: Contract addresses for the SDK
  addresses: ContractAddresses;
  
  // Optional: Default challenge duration in seconds for new channels
  challengeDuration?: bigint;
}

interface ContractAddresses {
  // Address of the Custody contract
  custody: Address;
  
  // Map of adjudicator types to their contract addresses
  adjudicators: {
    [key: string]: Address;
  };
  
  // Address of the guest participant
  guestAddress: Address;
  
  // Address of the token (use zeroAddress for ETH)
  tokenAddress: Address;
}
```

## Core Methods

### Channel Management

#### Create Channel

Creates a new state channel on-chain with initial funding allocations.

```typescript
const { channelId, initialState, txHash } = await client.createChannel({
  initialAllocationAmounts: [BigInt(1000), BigInt(0)],  // [Host amount, Guest amount]
  stateData: '0x00'  // Optional initial application state data
});

console.log(`Channel ID: ${channelId}`);
console.log(`Initial State:`, initialState);
console.log(`Transaction: ${txHash}`);
```

**Parameters:**
- `params: CreateChannelParams` - Channel creation parameters
  - `initialAllocationAmounts: [bigint, bigint]` - Token amounts for Host and Guest
  - `stateData?: Hex` - Optional initial application state

**Returns:**
- `channelId: ChannelId` - The unique identifier for the created channel
- `initialState: State` - The initial state object with signatures
- `txHash: Hash` - Transaction hash for the on-chain operation

#### Deposit and Create Channel

Convenience method to deposit tokens and create a channel in one logical operation.

```typescript
const result = await client.depositAndCreateChannel(
  BigInt(1000),  // Amount to deposit
  {
    initialAllocationAmounts: [BigInt(1000), BigInt(0)],
    stateData: '0x00'
  }
);

console.log(`Channel ID: ${result.channelId}`);
console.log(`Deposit TX: ${result.depositTxHash}`);
console.log(`Create TX: ${result.createChannelTxHash}`);
```

**Parameters:**
- `depositAmount: bigint` - Amount of tokens to deposit
- `params: CreateChannelParams` - Channel creation parameters

**Returns:**
- `channelId: ChannelId` - The unique identifier for the created channel
- `initialState: State` - The initial state object with signatures
- `depositTxHash: Hash` - Transaction hash for the deposit operation
- `createChannelTxHash: Hash` - Transaction hash for the channel creation

#### Checkpoint Channel

Records a state on-chain to prevent future disputes.

```typescript
const txHash = await client.checkpointChannel({
  channelId: '0xCHANNEL_ID',
  candidateState: signedState,
  proofStates: [] // Optional previous states if required by adjudicator
});
```

**Parameters:**
- `params: CheckpointChannelParams` - Checkpoint parameters
  - `channelId: ChannelId` - Channel identifier
  - `candidateState: State` - State to checkpoint (must be signed by both participants)
  - `proofStates?: State[]` - Optional proof states for complex adjudicators

**Returns:**
- `txHash: Hash` - Transaction hash for the checkpoint operation

#### Challenge Channel

Used when the counterparty is unresponsive, starts the challenge process.

```typescript
const txHash = await client.challengeChannel({
  channelId: '0xCHANNEL_ID',
  candidateState: mySignedState,
  proofStates: [] // Optional supporting proofs
});
```

**Parameters:**
- `params: ChallengeChannelParams` - Challenge parameters
  - `channelId: ChannelId` - Channel identifier
  - `candidateState: State` - State to use for the challenge (must be signed by at least the challenger)
  - `proofStates?: State[]` - Optional proof states

**Returns:**
- `txHash: Hash` - Transaction hash for the challenge operation

#### Close Channel

Finalizes the channel with a mutually signed state.

```typescript
const txHash = await client.closeChannel({
  finalState: {
    channelId: '0xCHANNEL_ID',
    stateData: '0x00',                           // Final app state
    allocations: [
      { destination: aliceAddress, token: tokenAddress, amount: BigInt(800) },
      { destination: bobAddress, token: tokenAddress, amount: BigInt(200) }
    ],
    serverSignature: [counterpartySignature]     // Guest's signature
  }
});
```

**Parameters:**
- `params: CloseChannelParams` - Close parameters
  - `finalState: { channelId, stateData, allocations, serverSignature }` - Final state information
  - `stateData?: Hex` - Optional override for application state data

**Returns:**
- `txHash: Hash` - Transaction hash for the close operation

### Fund Management

#### Deposit

Deposits tokens or ETH into the custody contract, handling ERC20 approval automatically if needed.

```typescript
const txHash = await client.deposit(BigInt(1000));
```

**Parameters:**
- `amount: bigint` - Amount of tokens/ETH to deposit

**Returns:**
- `txHash: Hash` - Transaction hash for the deposit

#### Withdraw

Withdraws available funds (not locked in active channels).

```typescript
const txHash = await client.withdrawal(BigInt(500));
```

**Parameters:**
- `amount: bigint` - Amount of tokens/ETH to withdraw

**Returns:**
- `txHash: Hash` - Transaction hash for the withdrawal

### Account Information

#### Get Account Channels

Retrieves all channels associated with the client's account.

```typescript
const channels = await client.getAccountChannels();
console.log(`You have ${channels.length} channels`);
channels.forEach(channel => console.log(`Channel: ${channel}`));
```

**Returns:**
- `Array<ChannelId>` - Array of channel identifiers

#### Get Account Info

Gets detailed information about the account's deposits and locked funds.

```typescript
const info = await client.getAccountInfo();
console.log(`Available: ${info.available}`);
console.log(`Locked in channels: ${info.locked}`);
console.log(`Channel count: ${info.channelCount}`);
```

**Returns:**
- `AccountInfo` - Object containing:
  - `available: bigint` - Available (unlocked) balance
  - `locked: bigint` - Balance locked in active channels
  - `channelCount: bigint` - Number of active channels

### Token Operations

#### Approve Tokens

Approves the custody contract to spend tokens on behalf of the account.

```typescript
const txHash = await client.approveTokens(BigInt(5000));
```

**Parameters:**
- `amount: bigint` - Amount of tokens to approve

**Returns:**
- `txHash: Hash` - Transaction hash for the approval

#### Get Token Allowance

Checks current token allowance for the custody contract.

```typescript
const allowance = await client.getTokenAllowance();
console.log(`Current allowance: ${allowance}`);
```

**Returns:**
- `bigint` - Current token allowance

#### Get Token Balance

Checks the token balance of the account.

```typescript
const balance = await client.getTokenBalance();
console.log(`Token balance: ${balance}`);
```

**Returns:**
- `bigint` - Current token balance

## Services

The NitroliteClient uses two primary services which are accessible through the client instance for advanced usage scenarios.

### NitroliteService

The `NitroliteService` provides direct access to the Custody contract functionality.

```typescript
// Accessible through the client instance's internals
const nitroliteService = client.txPreparer.deps.nitroliteService;

// Get list of channels for a specific account (not the one used to initialize the client)
const channels = await nitroliteService.getAccountChannels("0xOTHER_ADDRESS");

// Get account info for a specific account and token
const info = await nitroliteService.getAccountInfo(
  "0xSOME_ADDRESS",
  "0xTOKEN_ADDRESS"
);
```

### Erc20Service

The `Erc20Service` handles ERC20 token interactions:

```typescript
// Accessible through the client instance's internals
const erc20Service = client.txPreparer.deps.erc20Service;

// Check balance of a specific account
const balance = await erc20Service.getTokenBalance(
  "0xTOKEN_ADDRESS",
  "0xACCOUNT_ADDRESS"
);

// Check allowance between specific owner and spender
const allowance = await erc20Service.getTokenAllowance(
  "0xTOKEN_ADDRESS",
  "0xOWNER_ADDRESS",
  "0xSPENDER_ADDRESS"
);
```

## Transaction Preparation

For Account Abstraction (AA) support or transaction batching, NitroliteClient exposes the `txPreparer` property with methods that prepare transactions without executing them.

```typescript
// Prepare deposit transaction without executing it
const depositRequest = await client.txPreparer.prepareDepositTransactions(BigInt(1000));

// Prepare channel creation transaction
const createRequest = await client.txPreparer.prepareCreateChannelTransaction({
  initialAllocationAmounts: [BigInt(1000), BigInt(0)],
  stateData: '0x00'
});

// Use with Account Abstraction provider
const userOp = await aaProvider.buildUserOperation([
  depositRequest[0],  // ERC20 Approval (if needed)
  depositRequest[1],  // Deposit transaction
  createRequest       // Channel creation transaction
]);
await aaProvider.sendUserOperation(userOp);
```

## Advanced Usage

### Custom State Signing

By default, NitroliteClient uses the provided `walletClient` for both on-chain transactions and off-chain state signing. For advanced security models, you can provide a separate `stateWalletClient`:

```typescript
// Using a "hot" wallet for state signing and a "cold" wallet for on-chain transactions
import { privateKeyToAccount } from 'viem/accounts';
import { createWalletClient } from 'viem';

// Hot wallet for frequent state signing operations
const hotWalletAccount = privateKeyToAccount('0xHOT_KEY');
const stateWalletClient = createWalletClient({
  account: hotWalletAccount,
  chain,
  transport: http()
});

// Cold/hardware wallet for on-chain operations
const client = new NitroliteClient({
  publicClient,
  walletClient: hardwareWalletClient, // Less frequent on-chain transactions
  stateWalletClient,                  // Frequent state signing
  // ...other configuration
});
```

### Transaction Batching

For efficiency, you can batch multiple contract interactions into a single transaction:

```typescript
// Get request objects without executing them
const depositReq = await client.txPreparer.prepareWithdrawalTransaction(BigInt(500));
const approveReq = await client.txPreparer.prepareApproveTokensTransaction(BigInt(1000));

// Use viem's multicall
const multicallResult = await publicClient.multicall({
  contracts: [
    {
      ...depositReq,
      functionName: 'withdraw'
    },
    {
      ...approveReq,
      functionName: 'approve'
    }
  ]
});
```

### Account Abstraction

For Account Abstraction, you can prepare all the transaction components and send them as a UserOperation:

```typescript
// 1. Prepare transactions without executing them
const createChannelParams = {
  initialAllocationAmounts: [BigInt(1000), BigInt(0)],
  stateData: '0x00'
};

// Get multiple transactions in one call (approve + deposit + create)
const txs = await client.txPreparer.prepareDepositAndCreateChannelTransactions(
  BigInt(1000),
  createChannelParams
);

// 2. Convert to UserOperation format for your AA provider
const userOp = await aaProvider.buildUserOperation(txs);

// 3. Send the UserOperation
const userOpHash = await aaProvider.sendUserOperation(userOp);
console.log(`UserOperation hash: ${userOpHash}`);

// 4. Wait for the transaction to be included
const receipt = await aaProvider.waitForUserOperationReceipt(userOpHash);
console.log(`Transaction success: ${receipt.success}, gas used: ${receipt.gasUsed}`);
```

## Type Definitions

### Core Types

```typescript
// Channel identifier
type ChannelId = Hex;

// Channel configuration
interface Channel {
  participants: [Address, Address]; // [Host, Guest]
  adjudicator: Address;             // Contract that validates states
  challenge: bigint;                // Challenge duration in seconds
  nonce: bigint;                    // Unique identifier for the channel
}

// Channel state
interface State {
  data: Hex;                        // Application-specific state data
  allocations: [Allocation, Allocation]; // Fund distribution
  sigs: Signature[];                // Signatures approving this state
}

// Fund allocation
interface Allocation {
  destination: Address;             // Recipient address
  token: Address;                   // Token address (zeroAddress for ETH) 
  amount: bigint;                   // Amount allocated
}

// ECDSA signature
interface Signature {
  v: number;                        // Recovery value
  r: Hex;                           // First 32 bytes of signature
  s: Hex;                           // Second 32 bytes of signature
}

// Channel status
enum AdjudicatorStatus {
  VOID = 0,     // Never active or has anomaly
  PARTIAL = 1,  // Partially funded, waiting for other participants
  ACTIVE = 2,   // Fully funded and active
  INVALID = 3,  // State is invalid
  FINAL = 4     // Final state, can be closed
}

// Transaction preparation result
type PreparedTransaction = SimulateContractReturnType["request"];
```

### Method Parameters

```typescript
// Parameters for creating a new channel
interface CreateChannelParams {
  initialAllocationAmounts: [bigint, bigint]; // [Host, Guest] amounts
  stateData?: Hex;                            // Optional application state
}

// Parameters for closing a channel
interface CloseChannelParams {
  stateData?: Hex;                            // Optional override state
  finalState: {
    channelId: ChannelId;                     // Channel to close
    stateData: Hex;                           // Final app state
    allocations: [Allocation, Allocation];    // Final fund distribution
    serverSignature: Signature;             // Guest's signature
  };
}

// Parameters for challenging a channel
interface ChallengeChannelParams {
  channelId: ChannelId;                       // Channel to challenge
  candidateState: State;                      // Candidate state
  proofStates?: State[];                      // Optional proof states
}

// Parameters for checkpointing a channel
interface CheckpointChannelParams {
  channelId: ChannelId;                       // Channel to checkpoint
  candidateState: State;                      // State to checkpoint
  proofStates?: State[];                      // Optional proof states
}

// Account information
interface AccountInfo {
  available: bigint;                          // Available (unlocked) balance
  locked: bigint;                             // Balance locked in channels
  channelCount: bigint;                       // Number of active channels
}
```