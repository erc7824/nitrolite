# Nitrolite SDK for TypeScript

A streamlined TypeScript SDK for building custom state channel applications with the Nitrolite framework. The SDK provides a simple client interface that allows developers to create and manage channels with their own application logic.

## Overview

Nitrolite SDK provides a framework for developing scalable blockchain applications using state channels. State channels allow transactions to occur off-chain while maintaining the security guarantees of the underlying blockchain, resulting in:

- ⚡ **Instant Finality**: Transactions settle immediately between parties
- 💰 **Reduced Gas Costs**: Most interactions happen off-chain, with minimal on-chain footprint
- 🔄 **High Throughput**: Support for thousands of transactions per second
- 🛡️ **Security Guarantees**: Same security as on-chain, with cryptographic proofs
- 🌐 **Chain Agnostic**: Works with any EVM-compatible blockchain

## Installation

```bash
npm install @erc7824/nitrolite
```

## Quick Start

```typescript
import { NitroliteClient, AppDataTypes } from '@erc7824/nitrolite';
import { createPublicClient, createWalletClient, http, encodeAbiParameters, Hex } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { mainnet } from 'viem/chains';

// Setup clients
const publicClient = createPublicClient({
  chain: mainnet,
  transport: http('https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY')
});

const account = privateKeyToAccount('0x...');
const walletClient = createWalletClient({
  account,
  chain: mainnet,
  transport: http('https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY')
});

// Initialize Nitrolite client with required configuration
const client = new NitroliteClient({
  publicClient,
  walletClient,
  account,
  chainId: 1, // The chain ID your contracts are deployed on
  addresses: {
    custody: '0xYOUR_CUSTODY_CONTRACT_ADDRESS',
    adjudicators: {
      base: '0xYOUR_BASE_ADJUDICATOR_ADDRESS',
      numeric: '0xYOUR_NUMERIC_ADJUDICATOR_ADDRESS',
      sequential: '0xYOUR_SEQUENTIAL_ADJUDICATOR_ADDRESS'
    }
  }
});

// Create a custom application
interface MyAppState {
  value: bigint;
  sequence: bigint;
  metadata: string;
  isComplete: boolean;
}

const channel = client.createCustomChannel<MyAppState>({
  participants: ['0xALICE_ADDRESS', '0xBOB_ADDRESS'],
  adjudicatorAddress: '0xADJUDICATOR_ADDRESS',
  challenge: BigInt(86400), // 1 day challenge period
  
  // Encode app state to bytes
  encode: (state: MyAppState): Hex => {
    return encodeAbiParameters(
      [
        { type: 'uint256', name: 'value' },
        { type: 'uint256', name: 'sequence' },
        { type: 'string', name: 'metadata' },
        { type: 'bool', name: 'isComplete' }
      ],
      [state.value, state.sequence, state.metadata, state.isComplete]
    );
  },
  
  // Decode bytes back to app state
  decode: (encoded: Hex): MyAppState => {
    // Implementation would decode the bytes
    return { value: BigInt(0), sequence: BigInt(0), metadata: "", isComplete: false };
  },
  
  // Define your application logic
  validateTransition: (prevState, nextState, signer) => {
    // Only allow increasing values and sequence numbers
    return nextState.sequence > prevState.sequence && 
           nextState.value >= prevState.value;
  },
  
  // Define when the application state is final
  isFinal: (state) => state.isComplete,
  
  // Initial state
  initialState: { value: BigInt(0), sequence: BigInt(0), metadata: "", isComplete: false }
});

// Open the channel with initial funding
await channel.open(
  '0xTOKEN_ADDRESS', // ERC20 token address
  [BigInt(100), BigInt(100)] // Both participants fund with 100 tokens
);

// Update application state
await channel.updateAppState({
  value: BigInt(50),
  sequence: BigInt(1),
  metadata: "First update",
  isComplete: false
});

// Close the channel when done
await channel.close();
```

## Core Features

### 🔄 State Channel Management

- **Open/close channels** with configurable challenge periods
- **Challenge resolution** for uncooperative counterparties
- **Checkpointing** to prevent disputes

### 📱 Application Framework

Generic application interface that you can extend with your own logic:

- **Custom Applications**: Build any application logic on top of state channels
- **Application Logic Interface**: Define your own rules for state transitions
- **Built-in Helpers**: Utility functions for common application patterns
- **Example Applications**: Counter, MicroPayment, and more examples provided

### 🧩 Custom Adjudicators

Support for custom state transition validators (adjudicators):

- **Use Standard Adjudicators**: Built-in support for common patterns
- **Custom Adjudicator ABIs**: Provide your own adjudicator contract ABIs
- **Adjudicator Registry**: Register and reference adjudicators by type
- **Type-safe Interface**: TypeScript generics for your application states

### 🌐 Off-Chain Communication

- **Message Types** for protocol communication
- **Flexible Design** - implement your own communication layer
- **Type Definitions** for state proposals, signatures, and notifications

### 🔐 Cryptographic Utilities

- **State hashing** and verification
- **Signature generation** and validation
- **Channel ID computation**

## Documentation

For detailed API documentation, see [API.md](docs/API.md).

### Key Concepts

#### State Channels

A state channel is a relationship between participants that allows them to exchange state updates off-chain, with the blockchain serving as the ultimate arbiter in case of disputes.

```
+---------+                    +---------+
|         |   Off-chain state  |         |
| Alice   |  <-------------→   | Bob     |
|         |      updates       |         |
+---------+                    +---------+
     ↑                              ↑
     |      On-chain resolution     |
     +------------+  +---------------+
                  |  |
             +----+--+----+
             |            |
             | Blockchain |
             |            |
             +------------+
```

#### Applications

Applications define the rules for state transitions in a channel. The Nitrolite SDK provides a framework for building custom applications or using pre-built ones.

```typescript
// Creating a simple channel with a single numeric value
const numericChannel = client.createNumericChannel({
  participants: [aliceAddress, bobAddress],
  // The SDK will use the 'numeric' adjudicator from your addresses config
  // Or you can explicitly provide an address:
  // adjudicatorAddress: '0xSPECIFIC_ADJUDICATOR_ADDRESS'
});

// Creating a channel with sequential state updates
const sequentialChannel = client.createSequentialChannel({
  participants: [aliceAddress, bobAddress],
  // The SDK will use the 'sequential' adjudicator from your addresses config
  // Or you can explicitly provide an address:
  // adjudicatorAddress: '0xSPECIFIC_ADJUDICATOR_ADDRESS'
});

// Creating a custom channel with your own application logic
const customChannel = client.createCustomChannel<MyAppState>({
  participants: [aliceAddress, bobAddress],
  // You can reference a registered adjudicator by type
  adjudicatorType: 'base', // Will use the base adjudicator from your addresses config
  // Or provide a specific address
  // adjudicatorAddress: '0xYOUR_CUSTOM_ADJUDICATOR_ADDRESS',
  // You can also provide a custom adjudicator ABI
  // adjudicatorAbi: myCustomAdjudicatorAbi,
  encode: myStateEncoder,
  decode: myStateDecoder,
  validateTransition: myStateValidator,
  isFinal: myFinalStateChecker
});
```

#### Off-Chain Communication

The SDK provides message type definitions for off-chain communication between participants, but lets you implement the transport layer yourself.

```typescript
import { MessageType, ProposeStateMessage } from '@erc7824/nitrolite';

// Example: Creating your own message transport
class MyChannelMessenger {
  async sendMessage(message: NitroliteMessage) {
    // Your implementation - could use WebSockets, HTTP, etc.
    await this.socket.send(JSON.stringify(message));
  }

  async proposeState(channelId, state) {
    const stateHash = getStateHash(this.channel, state);
    
    // Create a properly formatted message
    const message: ProposeStateMessage = {
      type: MessageType.PROPOSE_STATE,
      channelId,
      timestamp: Date.now(),
      state,
      stateHash
    };
    
    await this.sendMessage(message);
  }
}
```

## Examples

See the [examples](examples/) directory for examples of using the Nitrolite SDK:

- **NextJS TypeScript Example** - A complete frontend application demonstrating how to use the SDK with a React-based web application.
- **Nitrolite RPC Example** - A simple example demonstrating how to use the NitroliteRPC protocol with WebSockets.

More examples are coming soon! Check the [examples README](examples/README.md) for details.

### Multi-Chain Support

Nitrolite SDK works with any EVM-compatible blockchain. Here's how to use it with different chains:

```typescript
import { NitroliteClient } from '@erc7824/nitrolite';
import { createPublicClient, http } from 'viem';
import { mainnet, optimism, arbitrum, base, polygon } from 'viem/chains';

// Example: Initialize client for Optimism
const optimismClient = new NitroliteClient({
  publicClient: createPublicClient({
    chain: optimism,
    transport: http('https://optimism.example.com')
  }),
  // Chain ID is automatically detected from the publicClient
  addresses: {
    custody: '0xOPTIMISM_CUSTODY_ADDRESS',
    adjudicators: {
      base: '0xOPTIMISM_BASE_ADJUDICATOR',
      // Add other adjudicators as needed
    }
  }
});

// Example: Initialize client for Arbitrum
const arbitrumClient = new NitroliteClient({
  publicClient: createPublicClient({
    chain: arbitrum,
    transport: http('https://arbitrum.example.com')
  }),
  // Explicitly provide chain ID if needed
  chainId: arbitrum.id,
  addresses: {
    custody: '0xARBITRUM_CUSTODY_ADDRESS',
    adjudicators: {
      base: '0xARBITRUM_BASE_ADJUDICATOR',
      // Add other adjudicators as needed
    }
  }
});

// The same SDK code works across all chains
// Just initialize with the appropriate client for your target chain
```

### Using Custom Adjudicator ABIs

```typescript
import { NitroliteClient } from '@erc7824/nitrolite';
import { Abi } from 'viem';

// Your custom adjudicator ABI
const myGameAdjudicatorAbi: Abi = [
  {
    type: 'function',
    name: 'adjudicate',
    inputs: [
      // Channel structure
      {
        name: 'chan',
        type: 'tuple',
        components: [
          { name: 'participants', type: 'address[2]' },
          { name: 'adjudicator', type: 'address' },
          { name: 'challenge', type: 'uint64' },
          { name: 'nonce', type: 'uint64' }
        ]
      },
      // Candidate state
      {
        name: 'candidate',
        type: 'tuple',
        components: [
          { name: 'data', type: 'bytes' },
          // Rest of the structure...
        ]
      },
      // Proof states
      {
        name: 'proofs',
        type: 'tuple[]',
        components: [
          // State structure...
        ]
      }
    ],
    outputs: [
      // Define outputs...
    ],
    stateMutability: 'view'
  }
  // Other function definitions...
];

// Initialize client with custom adjudicator ABIs
const client = new NitroliteClient({
  publicClient,
  walletClient,
  account,
  chainId: 1, // Required - specify the chain ID your contracts are deployed on
  addresses: {
    custody: '0xCUSTODY_ADDRESS', // Required
    adjudicators: {
      // You must provide at least a 'base' adjudicator
      base: '0xBASE_ADJUDICATOR_ADDRESS',
      // And you can register any custom adjudicators
      myGame: '0xMY_GAME_ADJUDICATOR_ADDRESS'
    }
  },
  // Optionally provide custom ABIs for your adjudicators
  adjudicatorAbis: {
    myGame: myGameAdjudicatorAbi
  }
});

// Or register an adjudicator ABI after initialization
client.registerAdjudicatorAbi('myOtherGame', myOtherGameAdjudicatorAbi);

// Create a channel using the custom adjudicator
const channel = client.createCustomChannel<MyGameState>({
  participants: [player1Address, player2Address],
  adjudicatorType: 'myGame', // References the registered ABI
  encode: gameStateEncoder,
  decode: gameStateDecoder,
  // Rest of config...
});
```

## Development

```bash
# Install dependencies
npm install

# Build the SDK
npm run build

# Run tests
npm test

# Type checking
npm run typecheck

# Lint code
npm run lint

# Clean build artifacts
npm run clean
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.