[![Release Policy](https://img.shields.io/badge/release%20policy-v1.0-blue)](https://github.com/layer-3/release-process/blob/master/README.md)
[![codecov](https://codecov.io/github/erc7824/nitrolite/graph/badge.svg?token=XASM4CIEFO)](https://codecov.io/github/erc7824/nitrolite)
[![Go Reference](https://pkg.go.dev/badge/github.com/erc7824/nitrolite/clearnode.svg)](https://pkg.go.dev/github.com/erc7824/nitrolite/clearnode)

# Nitrolite: State Channel Framework

Nitrolite is a lightweight, efficient state channel framework for Ethereum and other EVM-compatible blockchains, enabling off-chain interactions while maintaining on-chain security guarantees.

## Overview

Nitrolite is a complete state channel infrastructure consisting of three main components:

1. **Smart Contracts**: On-chain infrastructure for state channel management
2. **Clearnode**: A broker providing ledger services for the Clearnet protocol
3. **TypeScript SDK**: Client-side library for building custom state channel applications

### Key Benefits

- **Instant Finality**: Transactions settle immediately between parties
- **Reduced Gas Costs**: Most interactions happen off-chain, with minimal on-chain footprint
- **High Throughput**: Support for thousands of transactions per second
- **Security Guarantees**: Same security as on-chain, with cryptographic proofs
- **Chain Agnostic**: Works with any EVM-compatible blockchain

## Project Structure

This repository contains:

- **[`/contract`](/contract)**: Solidity smart contracts for the state channel framework
- **[`/clearnode`](/clearnode)**: Message broker implementation for the Clearnet protocol
- **[`/sdk`](/sdk)**: TypeScript SDK for building applications with Nitrolite
- **[`/docs`](/docs)**: Protocol specifications and documentation
- **[`/examples`](/examples)**: Sample applications built with Nitrolite

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

For the most up-to-date contract addresses on all supported networks, see the [contract deployments directory](/contract/deployments/).

Each network directory contains deployment information with timestamps and contract addresses.

### Interface Structure

The core interfaces include:

- **IChannel**: Main interface for channel creation, joining, closing, and dispute resolution
- **IAdjudicator**: Interface for state validation contracts
- **IDeposit**: Interface for token deposits and withdrawals
- **IComparable**: Interface for determining the ordering between states

See the [contract README](/contract/README.md) for detailed contract documentation.

## Clearnode

Clearnode is an implementation of a message broker node providing ledger services for the Clearnet protocol. It enables efficient off-chain payment channels with on-chain settlement capabilities.

### Features

- **Multi-Chain Support**: Connect to multiple EVM blockchains (Polygon, Celo, Base)
- **Off-Chain Payments**: Efficient payment channels for high-throughput transactions
- **Virtual Applications**: Create multi-participant applications
- **Message Forwarding**: Bi-directional message routing between application participants
- **Flexible Database**: Support for both PostgreSQL and SQLite
- **Prometheus Metrics**: Built-in monitoring and telemetry
- **Quorum-Based Signatures**: Support for multi-signature schemes with weight-based quorums

See the [Clearnode Documentation](/clearnode/README.md) for detailed documentation.

## TypeScript SDK

The SDK provides a simple client interface that allows developers to create and manage channels with their own application logic.

### Installation

```bash
npm install @erc7824/nitrolite
```

### Quick Start

**[Check Quick Start Guide](https://erc7824.org/quick_start)**


See the [SDK README](/sdk/README.md) for detailed SDK documentation.

## Examples

The repository includes several example applications built with Nitrolite:

### Snake Game

A multiplayer snake game that uses state channels for secure, off-chain gameplay payments:

- **Real-time multiplayer**: WebSocket-based gameplay
- **State channel integration**: Secure payments and state signing
- **Fair fund distribution**: Based on game outcome

[Learn more about Snake](/examples/snake)

### Tic Tac Toe

A simple tic-tac-toe game demonstrating the fundamentals of state channels:

- **React frontend**: Simple, clean UI
- **WebSocket backend**: For game coordination
- **State channel integration**: For secure payments

[Learn more about Tic Tac Toe](/examples/tictactoe)

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
# 4. Start the Clearnode service.

# To check the status:
docker-compose ps

# To view logs:
docker-compose logs -f clearnode

# To stop the environment:
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
