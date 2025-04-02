# TODO: REMOVE
# Nitrolite SDK for TypeScript - API Documentation

## Table of Contents
- [Overview](#overview)
- [Core Concepts](#core-concepts)
- [Installation](#installation)
- [Main SDK Class](#main-sdk-class)
- [Client API](#client-api)
- [Custom Adjudicators](#custom-adjudicators)
- [Applications](#applications)
- [RPC Protocol](#rpc-protocol)
  - [Basic Message Types](#basic-message-types-rpcrelay)
  - [Complete RPC Protocol](#complete-rpc-protocol)
  - [Virtual Channels](#virtual-channel-support)
- [Utilities](#utilities)
- [Types](#types)

## Overview

The Nitrolite SDK provides a complete toolkit for building applications using state channels. It abstracts the complexity of channel management, cryptographic operations, and off-chain communication to enable developers to create scalable blockchain applications.

## Core Concepts

### State Channels

State channels allow participants to execute transactions off-chain while maintaining the security guarantees of the underlying blockchain. The key components are:

1. **Channel**: A relationship between participants with defined rules
2. **State**: The current state of the application running in the channel
3. **Signatures**: Cryptographic proofs of participant agreement on state
4. **Adjudicator**: A contract that validates state transitions and resolves disputes

### Channel Lifecycle

1. **Open**: Participants fund the channel and establish initial state
2. **Update**: Participants exchange signed state updates off-chain
3. **Dispute**: If needed, participants can submit states on-chain to resolve disagreements
4. **Close**: When complete, the final state determines fund allocation

## Installation

```bash
npm install @ethtaipei/Nitrolite-sdk-ts
```

## Main SDK Components

### NitroliteClient

The `NitroliteClient` is the primary entry point for interacting with on-chain contracts.

```typescript
import { 
  NitroliteClient, 
  AppDataTypes 
} from '@ethtaipei/Nitrolite-sdk-ts';
import { createPublicClient, createWalletClient, http } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { mainnet } from 'viem/chains';

// Initialize clients
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
  account,
  addresses: {
    custody: '0xYOUR_CUSTODY_CONTRACT_ADDRESS',
    adjudicators: {
      base: '0xBASE_ADJUDICATOR_ADDRESS',
      numeric: '0xNUMERIC_ADJUDICATOR_ADDRESS',
      sequential: '0xSEQUENTIAL_ADJUDICATOR_ADDRESS'
    }
  },
  logger: customLogger // Optional custom logger
});

// Create a channel
const channel = client.createNumericChannel({
  participants: [aliceAddress, bobAddress],
  initialValue: BigInt(0)
});
```

### RPC Communication

For off-chain communication between participants, use the `RPCClient` with a provider:

```typescript
import { 
  RPCClient, 
  RPCChannelManager, 
  createRPCChannelContext,
  MemoryRPCProvider 
} from '@ethtaipei/Nitrolite-sdk-ts';
import { signMessage } from 'viem';

// Create a memory provider for testing
// In production, you'd implement your own provider
const provider = new MemoryRPCProvider(myAddress);

// Create an RPC client for communication
const rpcClient = new RPCClient({
  provider,
  address: myAddress,
  signer: (message) => signMessage({ message, privateKey: privateKey })
});

// Connect to the network
await rpcClient.connect();

// Create a channel manager
const channelManager = new RPCChannelManager(rpcClient);

// Enhance a channel with RPC capabilities
const rpcChannel = createRPCChannelContext(channel, rpcClient, channelManager);

// Now you can update state off-chain
await rpcChannel.updateAppState({ value: BigInt(10) });
```

#### Configuration

```typescript
// NitroliteClient configuration
interface NitroliteClientConfig {
  publicClient: PublicClient;
  walletClient?: WalletClient;
  account?: Account;
  chainId?: number;
  custodyAddress?: Address; // Legacy, prefer addresses
  addresses?: ContractAddresses;
  adjudicatorAbis?: Record<string, Abi>;
  logger?: Logger; // Custom logger instance
}

// RPC Client configuration
interface RPCClientConfig extends SDKConfig {
  provider: RPCProvider;
  address: Address;
  signer: (message: Hex) => Promise<Hex>;
  requestTimeoutMs?: number;
  maxRequestRetries?: number;
}
```

## Client API

### NitroliteClient

The `NitroliteClient` provides direct methods for interacting with the on-chain contracts.

```typescript
const client = Nitrolite.getClient();

// Open a channel
const channelId = await client.openChannel(channel, initialState);

// Close a channel
await client.closeChannel(channelId, finalState);
```

#### Methods

| Method | Description |
|--------|-------------|
| `openChannel(channel, deposit)` | Opens a new channel or joins an existing one |
| `closeChannel(channelId, candidate, proofs)` | Finalizes the channel with a mutually signed state |
| `challengeChannel(channelId, candidate, proofs)` | Submits a unilateral state when counterparty is unresponsive |
| `checkpointChannel(channelId, candidate, proofs)` | Records state on-chain to prevent future disputes |
| `reclaimChannel(channelId)` | Concludes channel after challenge period expires |
| `approveTokens(tokenAddress, amount, spender)` | Approves tokens for the custody contract |
| `getTokenAllowance(tokenAddress, owner, spender)` | Gets token allowance |
| `getTokenBalance(tokenAddress, account)` | Gets token balance |
| `registerAdjudicatorAbi(type, abi)` | Registers a custom adjudicator ABI |
| `getAdjudicatorAbi(type)` | Gets an adjudicator ABI by type |
| `getAdjudicatorAddress(type)` | Gets an adjudicator address by type |
| `createCustomChannel(params)` | Creates a channel with custom application logic |
| `createNumericChannel(params)` | Creates a channel with numeric state application |
| `createSequentialChannel(params)` | Creates a channel with sequential state application |

## Custom Adjudicators

The SDK allows users to provide custom adjudicator ABIs for their specific applications.

### Registering a Custom Adjudicator ABI

```typescript
// During client initialization
const client = new NitroliteClient({
  // Other configuration...
  addresses: {
    custody: '0xCUSTODY_ADDRESS',
    adjudicators: {
      myCustomType: '0xADJUDICATOR_ADDRESS'
    }
  },
  adjudicatorAbis: {
    myCustomType: customAdjudicatorAbi
  }
});

// After client initialization
client.registerAdjudicatorAbi('myCustomType', customAdjudicatorAbi);
```

### Creating a Channel with Custom Adjudicator

```typescript
// Create a channel with a custom adjudicator
const channel = client.createCustomChannel<MyAppState>({
  participants: [aliceAddress, bobAddress],
  adjudicatorType: 'myCustomType', // References the registered ABI
  adjudicatorAbi: myAdjudicatorAbi, // Optional if already registered
  encode: myStateEncoder,
  decode: myStateDecoder,
  // Other parameters...
});
```

### Custom Adjudicator Interface

The minimum interface for an adjudicator contract is:

```solidity
function adjudicate(
  Channel calldata chan,
  State calldata candidate,
  State[] calldata proofs
) external view returns (AdjudicatorStatus, Allocation[2] memory);
```

## Applications

### BaseApp

`BaseApp` is the abstract base class for all application implementations.

```typescript
abstract class BaseApp {
  // Methods to override
  abstract createInitialState(tokenAddress: Address, amounts: [bigint, bigint]): Promise<State>;
  
  // Common methods
  getChannelId(): ChannelId;
  getChannel(): Channel;
  getCurrentState(): State | undefined;
  getRole(): Role;
  getCounterparty(): Address;
  
  // Channel lifecycle methods
  async open(tokenAddress: Address, amounts: [bigint, bigint]): Promise<ChannelId>;
  async close(proofs?: State[]): Promise<void>;
  async challenge(proofs?: State[]): Promise<void>;
  async checkpoint(proofs?: State[]): Promise<void>;
  async reclaim(): Promise<void>;
  
  // State operations
  getStateHash(): StateHash | undefined;
}
```

### CounterApp

`CounterApp` implements a turn-taking counter game.

```typescript
const counterApp = Nitrolite.createApp({
  appType: AppType.COUNTER,
  participants: [aliceAddress, bobAddress],
  challenge: BigInt(86400) // 1 day
});

// Open channel with funding
await counterApp.open(tokenAddress, [BigInt(100), BigInt(0)]);

// Increment counter (make a move)
await counterApp.incrementCounter();

// Check if game is complete
if (counterApp.isGameComplete()) {
  await counterApp.close();
}
```

#### Methods

| Method | Description |
|--------|-------------|
| `getCounter()` | Returns current counter value |
| `incrementCounter()` | Increments the counter and returns new state |
| `isGameComplete()` | Checks if counter has reached 1000 |
| `processReceivedState(state)` | Processes a state received from counterparty |

### MicroPaymentApp

`MicroPaymentApp` implements a one-way payment channel.

```typescript
const paymentApp = Nitrolite.createApp({
  appType: AppType.MICROPAYMENT,
  participants: [aliceAddress, bobAddress],
  challenge: BigInt(86400) // 1 day
});

// Open channel with funding
await paymentApp.open(tokenAddress, [BigInt(1000), BigInt(0)]);

// Make a payment
await paymentApp.makePayment(BigInt(10));

// Close the channel
await paymentApp.close();
```

#### Methods

| Method | Description |
|--------|-------------|
| `getPaymentNonce()` | Returns current payment nonce |
| `getPaymentAmount()` | Returns current payment amount |
| `makePayment(amount)` | Makes a payment and returns new state |
| `processReceivedState(state)` | Processes a state received from counterparty |

## RPC Protocol

Nitrolite SDK provides two levels of off-chain communication:

1. **Basic Message Types** (`RPCRelay`) - Type definitions for message formats with no implementation
2. **Complete RPC Protocol** - Full implementation with client, server, virtual channels, and integration with channel state management

### Basic Message Types (RPCRelay)

```typescript
import { MessageType, ProposeStateMessage } from '@ethtaipei/Nitrolite-sdk-ts';

// Message types
const messageTypes = [
  MessageType.PROPOSE_STATE,
  MessageType.ACCEPT_STATE,
  MessageType.REJECT_STATE,
  MessageType.SIGN_STATE,
  MessageType.CHALLENGE_NOTIFICATION,
  MessageType.CLOSURE_NOTIFICATION
];

// Create a propose state message
const message: ProposeStateMessage = {
  type: MessageType.PROPOSE_STATE,
  channelId,
  timestamp: Date.now(),
  state,
  stateHash
};
```

### Complete RPC Protocol

The RPC Protocol layer enables comprehensive off-chain communication between participants:

```typescript
import { 
  RPCClient, 
  RPCChannelManager, 
  createRPCChannelContext,
  MemoryRPCProvider,
  LVCI
} from '@ethtaipei/Nitrolite-sdk-ts';

// Create a provider (MemoryRPCProvider for testing)
const provider = new MemoryRPCProvider(myAddress);

// Create an RPC client
const rpcClient = new RPCClient({
  provider,
  address: myAddress,
  signer: (message) => signMessage({ message, privateKey: PRIVATE_KEY })
});

// Connect the client
await rpcClient.connect();

// Create a channel manager
const channelManager = new RPCChannelManager(rpcClient);

// Enhance an existing channel context with RPC
const rpcChannel = createRPCChannelContext(
  existingChannel,
  rpcClient,
  channelManager
);

// Updates will automatically propagate to the counterparty
await rpcChannel.updateAppState(newAppState);

// Listen for state updates
rpcChannel.onStateUpdate((state) => {
  console.log('Received state update:', state);
});
```

### RPC Client API

The RPC client handles communication between participants:

| Method | Description |
|--------|-------------|
| `connect()` | Connect to the provider network |
| `disconnect()` | Disconnect from the provider network |
| `registerMethod(name, handler)` | Register a method handler |
| `unregisterMethod(name)` | Unregister a method handler |
| `sendRequest(recipient, method, params)` | Send a request to another participant |
| `sendNotification(recipient, type, data)` | Send a notification to another participant |
| `sendStateUpdate(recipient, channelId, state)` | Send a state update to a participant |
| `requestStateSignature(recipient, channelId, state)` | Request a signature for a state |
| `notifyChallenge(recipient, channelId, time, state)` | Notify a participant about a challenge |
| `notifyClosure(recipient, channelId, finalState)` | Notify a participant about channel closure |

### Custom Transport Providers

The SDK is designed to be transport-agnostic. To create your own transport provider:

```typescript
import { RPCProvider, RPCMessage } from '@ethtaipei/Nitrolite-sdk-ts';

// Implement your own transport provider
class MyCustomProvider implements RPCProvider {
  // Implementation of the RPCProvider interface
  async connect(): Promise<void> {
    // Connect to your transport
  }
  
  async disconnect(): Promise<void> {
    // Disconnect from your transport
  }
  
  async send(recipient: Address, message: RPCMessage): Promise<void> {
    // Send a message to the recipient
  }
  
  onMessage(handler: (from: Address, message: RPCMessage) => void): () => void {
    // Register a message handler
    // Return a function that unregisters the handler
  }
}

// Use your custom provider with the RPC client
const client = new RPCClient({
  provider: new MyCustomProvider(),
  address: myAddress,
  signer: mySignerFunction
});
```

### Virtual Channel Support

Virtual channels allow two participants to interact without having a direct on-chain channel. Instead, they connect through one or more intermediaries, who relay messages and state updates between them.

#### Virtual Channel Architecture

```
┌─────────┐                 ┌─────────┐                 ┌─────────┐
│         │   Ledger        │         │   Ledger        │         │
│  Alice  │◄────Channel────►│   Bob   │◄────Channel────►│ Charlie │
│         │                 │         │                 │         │
└────┬────┘                 └────┬────┘                 └────┬────┘
     │                           │                           │
     │                           │                           │
     │        Virtual Channel    │                           │
     └───────────────────────────┼───────────────────────────┘
                                 │
                                 ▼
                          ┌─────────────┐
                          │  Blockchain │
                          │   Network   │
                          └─────────────┘
```

In this structure:
1. Alice and Bob have a direct ledger channel
2. Bob and Charlie have a direct ledger channel
3. Alice and Charlie have a virtual channel through Bob
4. Bob serves as the relay/intermediary between Alice and Charlie

#### Virtual Channel Flow Diagram

```
┌─────────┐                 ┌─────────┐                 ┌─────────┐
│  Alice  │                 │   Bob   │                 │ Charlie │
└────┬────┘                 └────┬────┘                 └────┬────┘
     │                           │                           │
     │     1. Create LVCI        │                           │
     ├───────────────────────────┼───────────────────────────┤
     │                           │                           │
     │  2. Request VC Creation   │                           │
     │ ─────────────────────────►│                           │
     │                           │  3. Relay VC Creation     │
     │                           │ ─────────────────────────►│
     │                           │                           │
     │                           │     4. Accept VC          │
     │                           │ ◄─────────────────────────┤
     │       5. VC Created       │                           │
     │ ◄─────────────────────────┤                           │
     │                           │                           │
     │   6. Update State (S1)    │                           │
     │ ─────────────────────────►│                           │
     │                           │   7. Relay State (S1)     │
     │                           │ ─────────────────────────►│
     │                           │                           │
     │                           │   8. Update State (S2)    │
     │                           │ ◄─────────────────────────┤
     │   9. Relay State (S2)     │                           │
     │ ◄─────────────────────────┤                           │
     │                           │                           │
     │                           │                           │
```

#### Creating and Using Virtual Channels

The RPC protocol supports virtual channels through intermediaries, allowing parties to transact even without a direct channel:

```typescript
// Create a virtual channel through Bob
const lvci = LVCI.create(
  aliceAddress,  // Origin
  charlieAddress, // Destination
  [bobAddress]   // Intermediary
);

// Create the virtual channel
const virtualChannel = await rpcClient.createVirtualChannel(lvci, initialState);

// Update state through the virtual channel
// State updates flow from origin to destination and back, creating a round-trip
// through all participants to ensure everyone has the latest state
const updatedState = await rpcClient.relayStateUpdate(lvci, newState);

// You can include metadata to help with flow control in complex virtual channels
const stateWithMetadata = {
  ...newState,
  metadata: {
    isInbound: false, // Used in relay logic to determine flow direction
    timestamp: Date.now()
  }
};
const relayedState = await rpcClient.relayStateUpdate(lvci, stateWithMetadata);
```

#### Multi-Hop Virtual Channels

Nitrolite SDK supports channels through multiple intermediaries:

```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│         │     │         │     │         │     │         │
│  Alice  │◄───►│  Bob    │◄───►│  Carol  │◄───►│  Dave   │
│         │     │         │     │         │     │         │
└────┬────┘     └─────────┘     └─────────┘     └────┬────┘
     │                                                │
     │                                                │
     │               Virtual Channel                  │
     └────────────────────────────────────────────────┘
```

Creating a multi-hop virtual channel:

```typescript
// Create a multi-hop virtual channel through Bob and Carol
const multiHopLvci = LVCI.create(
  aliceAddress,     // Origin
  daveAddress,      // Destination
  [bobAddress, carolAddress]  // Multiple intermediaries
);

// The rest of the API remains the same
const virtualChannel = await rpcClient.createVirtualChannel(multiHopLvci, initialState);
```

### RPCChannelManager API

The channel manager integrates with the RPC layer:

| Method | Description |
|--------|-------------|
| `registerChannel(channel, counterparty, initialState)` | Register a channel for tracking |
| `unregisterChannel(channelId)` | Unregister a channel |
| `getChannel(channelId)` | Get a registered channel |
| `getChannelIds()` | Get all registered channel IDs |
| `updateState(channelId, newState)` | Update the state of a channel |
| `requestSignature(channelId, state)` | Request a signature for a state |

### Message Types

The protocol supports the following message types:

| Message Type | Description |
|--------------|-------------|
| `REQUEST` | Request a method call |
| `RESPONSE` | Response to a method call |
| `ERROR` | Error response |
| `NOTIFICATION` | Server-initiated notification |

### Standard RPC Methods

| Method | Description |
|--------|-------------|
| `open_channel` | Open a channel |
| `update_state` | Update channel state |
| `sign_state` | Sign a state |
| `close_channel` | Close a channel |
| `challenge_channel` | Challenge a channel |
| `checkpoint_channel` | Checkpoint a channel state |
| `create_virtual_channel` | Create a virtual channel |
| `relay_state_update` | Relay a state update |
| `relay_signature` | Relay a signature |
| `close_virtual_channel` | Close a virtual channel |
| `ping` | Check connectivity |
| `get_time` | Get server time |
| `get_channels` | Get available channels |

## Utilities

### Cryptographic Operations

```typescript
import { 
  getChannelId, 
  getStateHash, 
  verifySignature, 
  generateChannelNonce 
} from '@ethtaipei/Nitrolite-sdk-ts';

// Generate channel ID
const channelId = getChannelId(channel);

// Generate robust nonce for channel creation
const nonce = generateChannelNonce(userAddress);

// Generate state hash
const stateHash = getStateHash(channel, state);

// Verify state signature
const isValid = await verifySignature(stateHash, signature, signer);
```

### Data Encoding/Decoding

```typescript
import { appEncoders, appDecoders } from '@ethtaipei/Nitrolite-sdk-ts';

// Encode counter data
const data = appEncoders.counter(BigInt(42));

// Decode counter data
const { counter } = appDecoders.counter(data);
```

## Types

### Error Handling

The SDK provides a comprehensive error handling system:

```typescript
import { 
  NitroliteError, 
  ValidationError, 
  ConnectionError, 
  ContractError,
  TokenError,
  InsufficientBalanceError,
  InsufficientAllowanceError,
  TransactionError, 
  StateError,
  VirtualChannelError
} from '@ethtaipei/Nitrolite-sdk-ts';

try {
  await Nitrolite.client.openChannel(channel, initialState);
} catch (error) {
  if (error instanceof TokenError) {
    // Handle token-specific errors
    console.error(`Token error: ${error.message}, Suggestion: ${error.suggestion}`);
    if (error.code === 'INSUFFICIENT_BALANCE') {
      // Handle insufficient balance
      console.log('Not enough tokens to complete the transaction');
      // Show the required amount vs. actual balance
      if (error.details?.required && error.details?.actual) {
        console.log(`Required: ${error.details.required}, Available: ${error.details.actual}`);
      }
    } else if (error.code === 'INSUFFICIENT_ALLOWANCE') {
      // Handle insufficient allowance
      console.log('The contract needs permission to use your tokens');
      // Prompt user to approve tokens
      const tokenAddress = error.details?.tokenAddress;
      const spender = error.details?.spender;
      const requiredAmount = error.details?.required;
      
      if (tokenAddress && spender && requiredAmount) {
        await Nitrolite.client.approveTokens(tokenAddress, requiredAmount, spender);
        console.log(`Approved ${requiredAmount} tokens for ${spender}`);
      }
    }
  } else if (error instanceof TransactionError) {
    // Handle transaction errors
    console.error(`Transaction failed: ${error.message}`);
  } else if (error instanceof NitroliteError) {
    // Handle other Nitrolite-specific errors
    console.error(`Error: ${error.message}, Code: ${error.code}`);
  } else {
    // Handle unknown errors
    console.error(`Unknown error: ${error}`);
  }
}
```

### Core Types

```typescript
// Channel configuration
interface Channel {
  participants: [Address, Address];
  adjudicator: Address;
  challenge: bigint;
  nonce: bigint; // Use generateChannelNonce() for robust, collision-resistant values
}

// Channel state
interface State {
  data: Hex;
  allocations: [Allocation, Allocation];
  sigs: Signature[];
}

// Fund allocation
interface Allocation {
  destination: Address;
  token: Address;
  amount: bigint;
}

// ECDSA signature
interface Signature {
  v: number;
  r: Hex;
  s: Hex;
}

// Adjudicator status
enum AdjudicatorStatus {
  VOID = 0,
  PARTIAL = 1,
  ACTIVE = 2,
  INVALID = 3,
  FINAL = 4
}

// Application types
// Application logic interface
interface AppLogic<T = unknown> {
  encode: (data: T) => Hex;
  decode: (encoded: Hex) => T;
  validateTransition?: (prevState: T, nextState: T, signer: Address) => boolean;
  isFinal?: (state: T) => boolean;
  getAdjudicatorAddress: () => Address;
  getAdjudicatorType?: () => string;
}

// Contract addresses configuration
interface ContractAddresses {
  custody: Address;
  adjudicators: {
    base?: Address;
    [key: string]: Address | undefined;
  };
}
```

### Application Data Types

```typescript
// Counter app data
interface CounterData {
  counter: bigint;
}

// TicTacToe app data
interface TicTacToeData {
  board: number[][];
  nextPlayer: number;
  winner: number;
}

// MicroPayment app data
interface MicroPaymentData {
  nonce: bigint;
  amount: bigint;
}
```