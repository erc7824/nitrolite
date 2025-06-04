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
- **[`/docs`](/docs)**: Protocol specifications and documentation

## Protocol

Nitrolite implements a state channel protocol that enables secure off-chain communication with minimal on-chain operations. The protocol includes:

- **Channel Creation**: A funding protocol where participants lock assets in the custody contract
- **Off-Chain Updates**: A mechanism for exchanging and signing state updates off-chain
- **Channel Closure**: Multiple resolution paths including cooperative close and challenge-response
- **Checkpointing**: The ability to record valid states on-chain without closing the channel
- **Reset Capability**: Support for resizing allocations by closing and reopening channels

See the [protocol specification](/docs/PROTOCOL.md) for complete details.

## Smart Contracts

The Nitrolite contract system provides:

- **Custody** of ERC-20 tokens for each channel
- **Mutual close** when participants agree on a final state
- **Challenge/response** mechanism for unilateral finalization
- **Checkpointing** for recording valid states without closing

### Deployments

- CELO Custody: <https://celoscan.io/address/0xDB33fEC4e2994a675133320867a6439Da4A5acD8>
- CELO NitroRPC: `0x6C68440eF55deecE7532CDa3b52D379d0Bb19cF5`
- CELO Dummy: `0xC2BA5c5E2c4848F64187Aa1F3f32a331b0C031b9`
- Polygon Custody: <https://polygonscan.com/address/0xDB33fEC4e2994a675133320867a6439Da4A5acD8>
- Polygon NitroRPC: `0x6C68440eF55deecE7532CDa3b52D379d0Bb19cF5`
- Polygon Dummy: `0xC2BA5c5E2c4848F64187Aa1F3f32a331b0C031b9`

### Interface Structure

The core interfaces include:

- **IChannel**: Main interface for channel creation, joining, closing, and dispute resolution
- **IAdjudicator**: Interface for state validation contracts
- **IDeposit**: Interface for token deposits and withdrawals
- **IComparable**: Interface for determining the ordering between states

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

### Channel Lifecycle

1. **Creation**: Creator constructs channel config, defines initial state with `CHANOPEN` magic number
2. **Joining**: Participants verify the channel and sign the same funding state
3. **Active**: Once fully funded, the channel transitions to active state for off-chain operation
4. **Off-chain Updates**: Participants exchange and sign state updates according to application logic
5. **Resolution**:
   - **Cooperative Close**: All parties sign a final state with `CHANCLOSE` magic number
   - **Challenge-Response**: Participant can post a state on-chain and initiate challenge period
   - **Checkpoint**: Record valid state on-chain without closing for future dispute resolution
   - **Reset**: Close and reopen a channel to resize allocations

### Data Structures

- **Channel**: Configuration with participants, adjudicator, challenge period, and nonce
- **State**: Application data, asset allocations, and signatures
- **Allocation**: Destination address, token, and amount for each participant
- **Status**: Channel lifecycle stages (VOID, INITIAL, ACTIVE, DISPUTE, FINAL)

## Quick Start with Docker Compose

Get started quickly with the local development environment using Docker Compose:

```bash
# Start the environment
docker-compose up -d

# This will:
# 1. Start a local Anvil blockchain on port 8545
# 2. Deploy the Custody, ERC20, and FlagAdjudicator contracts
# 3. Seed the database with test tokens information
# 4. Start the Clearnode service on ports 8000 (HTTP) and 4242 (WebSocket)
```

To check the status:
```bash
docker-compose ps
```

To view logs:
```bash
docker-compose logs -f clearnode
```

To stop the environment:
```bash
docker-compose down
```

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
