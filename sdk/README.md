# Nitrolite SDK

[![npm version](https://img.shields.io/npm/v/@erc7824/nitrolite.svg)](https://www.npmjs.com/package/@erc7824/nitrolite)
[![License](https://img.shields.io/npm/l/@erc7824/nitrolite.svg)](https://github.com/erc7824/nitrolite/blob/main/LICENSE)
[![Documentation](https://img.shields.io/badge/docs-website-blue)](https://erc7824.org/nitrolite/nitrolite_client/)

A TypeScript SDK for building scalable blockchain applications using ERC-7824. The SDK provides a simple client interface that allows developers to create and manage channels with custom application logic.

## Features

- **Instant Finality**: Transactions settle immediately between parties
- **Reduced Gas Costs**: Most interactions happen off-chain, with minimal on-chain footprint
- **High Throughput**: Support for thousands of transactions per second
- **Security Guarantees**: Same security as on-chain, with cryptographic proofs
- **Framework Agnostic**: Works with any JavaScript framework (Vue, Angular, React, etc.)

## Installation

```bash
npm install @erc7824/nitrolite
```

## Quick Start

```typescript
import { NitroliteClient } from '@erc7824/nitrolite';
import { createPublicClient, createWalletClient, http } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { mainnet } from 'viem/chains';

// Setup clients
const publicClient = createPublicClient({
  chain: mainnet,
  transport: http('https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY')
});

const account = privateKeyToAccount('0x...');
const walletClient = createWalletClient({
  account,
  chain: mainnet,
  transport: http('https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY')
});

// Initialize Nitrolite client
const client = new NitroliteClient({
  publicClient,
  walletClient,
  addresses: {
    custody: '0xCustodyAddress',
    adjudicator: '0xAdjudicatorAddress',
    guestAddress: '0xGuestAddress',
    tokenAddress: '0xTokenAddress'
  },
  challengeDuration: 100n,
  chainId: 1
});

// 1. Deposit funds
const depositTxHash = await client.deposit(1000000n);

// 2. Create a channel
const { channelId, initialState, txHash } = await client.createChannel({
  initialAllocationAmounts: [700000n, 300000n],
  stateData: '0x1234'
});

// 3. Resize the channel when needed
const resizeTxHash = await client.resizeChannel({
  channelId,
  candidateState: updatedState
});

// 4. Close the channel
const closeTxHash = await client.closeChannel({
  finalState: finalState
});
```

## Core Methods

The `NitroliteClient` class provides the following key methods:

| Method | Description |
|--------|-------------|
| `deposit` | Deposits tokens or ETH into the custody contract |
| `createChannel` | Creates a new state channel |
| `depositAndCreateChannel` | Combines deposit and channel creation in one operation |
| `challengeChannel` | Challenges a channel when the counterparty is non-responsive |
| `resizeChannel` | Updates a channel's allocation |
| `closeChannel` | Closes a channel with a mutually agreed final state |
| `withdrawal` | Withdraws tokens from the custody contract |
| `getAccountChannels` | Gets a list of channels for the account |
| `getAccountInfo` | Gets account balance information |
| `approveTokens` | Approves tokens for the custody contract |
| `getTokenAllowance` | Gets current token allowance |
| `getTokenBalance` | Gets token balance for an account |

## RPC Communication

Nitrolite includes a built-in RPC system for communication between channel participants:

```typescript
import { NitroliteRPC } from '@erc7824/nitrolite';

// Create a WebSocket connection to the RPC server
const socket = new WebSocket('wss://rpc.nitrolite.io');

// Create a message for authentication
const authMsg = await NitroliteRPC.createAuthRequestMessage(
  signer, 
  account.address
);

// Send the authentication message
socket.send(authMsg);

// Create application-specific messages
const appMsg = await NitroliteRPC.createAppMessage(
  signer,
  channelId,
  ["param1", "param2"]
);

// Send the application message
socket.send(appMsg);
```

### RPC Features

- **Authentication**: Secure authentication using wallet signatures
- **Message Signing**: All RPC messages are cryptographically signed
- **Channel-Specific Communication**: Messages can target specific channels
- **Application Messages**: Support for custom application-specific messages
- **Intent Communication**: Facilitate state updates with off-chain intents

### Core RPC Methods

| Method | Description |
|--------|-------------|
| `createAuthRequestMessage` | Creates a message for authentication |
| `createAuthVerifyMessage` | Verifies an authentication challenge |
| `createPingMessage` | Creates a heartbeat message |
| `createGetConfigMessage` | Retrieves channel configuration |
| `createApplicationMessage` | Creates an application-specific message |
| `createCloseApplicationMessage` | Initiates application closure |
| `parseResponse` | Parses and validates incoming messages |

## Framework Compatibility

Nitrolite SDK is designed to be framework-agnostic, working seamlessly with:

- **Vue.js** - Integrate with Pinia/Vuex for state management
- **Angular** - Use with Angular services and RxJS
- **React** - Works with any state management (Redux, Context API, Zustand)
- **Svelte** - Integrate with Svelte stores
- **Vanilla JavaScript** - Use directly with no framework dependencies

## Account Abstraction Support

Nitrolite works with account abstraction through transaction preparation methods that can be used with any AA system:

```typescript
// Get the transaction request for a deposit
const depositRequest = await client.txPreparer.prepareDeposit(100000n);

// Execute with your AA provider
const userOp = await aaProvider.prepareUserOperation({
  target: depositRequest.to,
  data: depositRequest.data,
  value: depositRequest.value
});
```

## Documentation

For complete documentation, visit [https://erc7824.org/nitrolite/nitrolite_client/](https://erc7824.org/nitrolite/nitrolite_client/)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
