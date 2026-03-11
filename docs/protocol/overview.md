# Nitrolite Protocol Overview

## Table of Contents

1. [Overview](overview.md) — high-level protocol description and design goals
2. [Terminology](terminology.md) — canonical definitions of all protocol terms
3. [Cryptography](cryptography.md) — encoding, hashing, signing, and replay protection
4. [State Model](state-model.md) — state structure, versioning, and consistency rules
5. [Channel Protocol](channel-protocol.md) — channel lifecycle, transitions, and advancement rules
6. [Enforcement and Settlement](enforcement-and-settlement.md) — checkpoints, on-chain validation, and settlement
7. [Cross-Chain and Assets](cross-chain-and-assets.md) — unified asset model and cross-chain operations
8. [Interactions](interactions.md) — message envelope, core operations, and events
9. [Security and Limitations](security-and-limitations.md) — security guarantees, trust assumptions, and known limitations
10. [Extensions Overview](extensions/overview.md) — extension model, lifecycle, and safety constraints
11. [Application Sessions](extensions/app-sessions.md) — app session entity, state, session keys, and commit/release

## Purpose

Nitrolite is a state channel protocol that enables off-chain interactions between users while preserving on-chain security guarantees.

Network participants exchange signed state updates off-chain. Any user can enforce the latest agreed state on the blockchain layer at any time.

## Design Goals

The protocol is designed to achieve:

- **Off-chain scalability** — minimize on-chain transactions by moving state advancement off-chain
- **Blockchain security guarantees** — any user can fall back to the blockchain layer to enforce the latest state
- **Cross-chain asset interaction** — operate on assets across multiple blockchains through a unified model
- **Extensibility** — support additional functionality through protocol extensions without modifying the core protocol

## System Roles

The protocol defines the following roles.

**User**
An entity that opens channels, signs state updates, and holds assets within the protocol.

**Node**
An entity that facilitates off-chain state advancement, manages channels, and interacts with the blockchain layer.

**Blockchain**
The on-chain system that stores enforceable channel states and resolves disputes.

## High-Level Architecture

The system operates in three conceptual layers:

1. **Off-chain layer** — participants exchange signed state updates directly
2. **Protocol layer** — defines rules for state validity, advancement, and enforcement
3. **Blockchain layer** — blockchain contracts that hold assets and enforce states

## Core Protocol Concepts

**Channels**
A channel is a state container shared between a Node and a User. It holds asset allocations and supports off-chain updates.

**States**
A state represents the current agreed configuration of a channel, including asset allocations and metadata.

**State Advancement**
User updates channel states off-chain by exchanging signed state transitions.

**State Enforcement**
Any participant can submit the latest signed state to the settlement layer for on-chain enforcement.

**Unified Assets**
Assets from multiple blockchains are represented in a unified model, enabling cross-chain operations within a single channel.

**Extensions**
Additional protocol functionality, such as application sessions, is provided through the extension layer without modifying core protocol rules.

## Protocol Layers

The protocol separates responsibilities into distinct layers.

**Core Protocol**
Defines channels, states, state advancement rules, and enforcement mechanisms.

**Extension Layer**
Provides additional functionality such as application sessions. Extensions interact with the core protocol through defined interfaces.

**Settlement Layer**
Blockchain contracts that create channels, hold deposits, accept state checkpoints, and release funds.

## Protocol Version

This documentation describes Nitrolite protocol version 1.

Compatibility expectations:

- State structures and signing rules defined in this version are stable
- Extension interfaces may evolve in future versions
- Settlement layer contracts are version-specific

---

Next: [Terminology](terminology.md)
