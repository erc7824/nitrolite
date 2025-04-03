# Nitrolite: State Channel Framework

Nitrolite is a lightweight, efficient state channel framework for Ethereum and other EVM-compatible blockchains, enabling off-chain interactions while maintaining on-chain security guarantees.

## Overview

The Nitrolite framework consists of two main components:

1. **Smart Contracts**: On-chain infrastructure for state channel management
2. **TypeScript SDK**: Client-side library for building custom state channel applications

### Key Benefits

- **Instant Finality**: Transactions settle immediately between parties
- **Reduced Gas Costs**: Most interactions happen off-chain, with minimal on-chain footprint
- **High Throughput**: Support for thousands of transactions per second
- **Security Guarantees**: Same security as on-chain, with cryptographic proofs
- **Chain Agnostic**: Works with any EVM-compatible blockchain

## Project Structure

This repository contains:

- **[`/contract`](/contract)**: Solidity smart contracts for the state channel framework
- **[`/sdk`](/sdk)**: TypeScript SDK for building applications with Nitrolite

## Smart Contracts

The Nitrolite contract system provides:

- **Custody** of ERC-20 tokens for each channel
- **Mutual close** when participants agree on a final state
- **Challenge/response** mechanism for unilateral finalization

### Interface Structure

The core interfaces include:

- **IChannel**: Main interface for channel management
- **IAdjudicator**: Interface for state validation contracts
- **IDeposit**: Interface for token deposits and withdrawals

See the [contract README](/contract/README.md) for detailed contract documentation.

## TypeScript SDK

The SDK provides a simple client interface that allows developers to create and manage channels with their own application logic.

### Installation

```bash
npm install @erc7824/nitrolite
```

### Quick Start

```typescript
import { NitroliteClient } from '@erc7824/nitrolite';
import { createPublicClient, createWalletClient, http } from 'viem';
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

// Initialize Nitrolite client
const client = new NitroliteClient({
  publicClient,
  walletClient,
  account,
  chainId: 1,
  addresses: {
    custody: '0xYOUR_CUSTODY_CONTRACT_ADDRESS',
    adjudicators: {
      base: '0xYOUR_BASE_ADJUDICATOR_ADDRESS'
    }
  }
});

// Create a channel
const channel = client.createCustomChannel({
  // Channel configuration
});

// Open the channel with initial funding
await channel.open(
  '0xTOKEN_ADDRESS',
  [BigInt(100), BigInt(100)]
);

// Update state off-chain
await channel.updateAppState({
  // Your application state
});

// Close the channel when done
await channel.close();
```

See the [SDK README](/sdk/README.md) for detailed SDK documentation.

## Examples

Check out the examples in the [`/sdk/examples`](/sdk/examples) directory:

- **NextJS TypeScript Example**: A complete frontend application demonstrating the SDK
- **Nitrolite RPC Example**: Sample code for the WebSocket-based RPC protocol

## Key Concepts

### State Channels

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

### High-Level Flow

1. **Channel Creation**: Participants deposit ERC20 tokens into the contract
2. **Off-Chain Updates**: Parties exchange and co-sign states off-chain
3. **Happy Path**: Both parties agree on a final state and close cooperatively
4. **Unhappy Path**: If a party stops responding, the other can use the challenge mechanism to finalize

## Development

```bash
# Install dependencies
npm install

# Build the SDK
cd sdk && npm run build

# Run tests
cd contract && forge test
cd sdk && npm test
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.