# Terminology

Previous: [Overview](overview.md) | Next: [Cryptography](cryptography.md)

---

This document defines all protocol terms used throughout the Nitrolite protocol documentation.

Each term is defined once. All other documents must use these terms consistently.

## Naming Conventions

- Protocol entities use CamelCase (e.g., ChannelState, AppSession)
- Field names use CamelCase (e.g., ChannelId, StateVersion)
- Operations use lowercase with hyphens in document references (e.g., state-advancement)

## Core Entities

### Channel

A state container shared between participants that allows off-chain updates while maintaining on-chain security guarantees.

Channels enable fast off-chain execution while preserving the ability to settle on-chain if necessary.

### Channel Definition

The immutable parameters that define a channel. A channel definition is fixed at creation time and cannot change during the channel lifecycle.

### Channel State

The current agreed configuration of a channel, including asset allocations and metadata. Channel state evolves through off-chain state advancement.

### Participant

An entity that holds a signing key and participates in a channel. Each channel has a fixed set of participants defined at creation.

### Asset

A representation of value within the protocol. Assets are identified independently of any specific blockchain.

## State Concepts

### State

An abstract data structure representing the current configuration of a protocol entity at a specific version.

### State Version

A monotonically increasing integer that identifies the order of state updates. Each new state must have a version strictly greater than the previous state.

### State Advancement

The process of updating a protocol entity's state off-chain through signed messages exchanged between participants.

### State Enforcement

The process of submitting a signed state to the settlement layer for on-chain validation and enforcement.

### Transition

A typed operation that describes the reason and parameters for a state update.

### Intent

A signed message from a participant indicating their agreement to a proposed state update.

## Cryptographic Concepts

### Signature

A cryptographic proof that a specific key holder authorized a specific message.

### Signer

An entity capable of producing signatures. Each signer is associated with a specific key.

### Session Key

A delegated signing key authorized by a participant's primary key to sign specific types of state updates on their behalf.

### Signature Validation Mode

A mechanism that determines how a signature is verified. Different validation modes support different key types and authorization schemes.

## Ledger Concepts

### Ledger

A record of asset allocations within a channel, associated with a specific blockchain.

### Home Ledger

The primary ledger of a channel, located on the blockchain where the channel was created. The home ledger is the authoritative source for channel state enforcement.

### Non-Home Ledger

A secondary ledger tracking asset allocations on a blockchain other than the home chain.

## Protocol Operations

### Checkpoint

The operation of submitting a signed state to the settlement layer. A checkpoint records the latest agreed state on-chain.

### Commit

The operation of moving assets from a channel into an extension, such as an application session.

### Release

The operation of returning assets from an extension back to the channel.

## Extension Concepts

### Extension

An additional protocol module that provides functionality beyond the core channel protocol. Extensions interact with channels through defined interfaces.

### Application Session

An extension that enables off-chain application functionality. Application sessions hold committed assets and maintain their own state.

### Application State

The state associated with an application session, tracking committed assets and application-specific data.

---

Previous: [Overview](overview.md) | Next: [Cryptography](cryptography.md)
