# Protocol Extensions

Previous: [Security and Limitations](../security-and-limitations.md) | Next: [Application Sessions](app-sessions.md)

---

This document explains the extension model used by the Nitrolite protocol.

## Purpose

Extensions allow additional functionality to be built on top of the core protocol without modifying its rules.

The core protocol defines channels, states, and enforcement. Extensions provide higher-level operations such as application sessions, enabling new use cases while preserving protocol safety guarantees.

## Extension Model

Extensions interact with the core protocol through defined interfaces:

- Extensions receive assets from channels through **commit** operations
- Extensions return assets to channels through **release** operations
- Extension state is maintained separately from channel state
- Extension operations require participant signatures, enforced by the core protocol

An extension does not modify channel protocol rules. It operates within the boundaries defined by commit and release transitions.

## Extension Lifecycle

**Registration**
An extension type is recognized by the protocol. Registration defines the rules and interfaces the extension must follow.

**Activation**
An extension instance is created when participants commit assets from a channel. The extension becomes active and maintains its own state.

**Operation**
The extension processes application-specific logic off-chain. Participants interact with the extension through signed state updates.

**Termination**
The extension releases assets back to the channel. Once all assets are released, the extension instance is terminated.

## Integration with Channels

Extensions integrate with channels through the following mechanisms:

- **Commit transition** — a channel state update that moves assets into an extension. The channel's allocations decrease by the committed amount.
- **Release transition** — a channel state update that returns assets from an extension. The channel's allocations increase by the released amount.
- **State coordination** — extension state updates and channel state updates are coordinated to maintain overall asset consistency.

Asset totals across the channel and all active extensions must remain constant.

## Extension Safety

Extensions must follow these constraints:

- Extensions cannot create or destroy assets
- Extensions cannot modify channel state directly; they interact only through commit and release transitions
- Extension operations require valid signatures from authorized participants
- Extension state must be independently verifiable by all participants
- A malfunctioning extension cannot compromise the safety of the underlying channel

---

Previous: [Security and Limitations](../security-and-limitations.md) | Next: [Application Sessions](app-sessions.md)
