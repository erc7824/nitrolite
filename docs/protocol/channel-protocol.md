# Channel Protocol

Previous: [State Model](state-model.md) | Next: [Enforcement and Settlement](enforcement-and-settlement.md)

---

This document describes how channels operate and how states evolve through off-chain state advancement.

## Purpose

Channels are the primary mechanism for off-chain interaction in the Nitrolite protocol. They allow participants to exchange assets and update state without on-chain transactions.

## Channel Definition

A channel is defined by a set of immutable parameters fixed at creation time.

```
ChannelDefinition {
  Participants:   []address     // ordered list of participant addresses
  Adjudicator:    address       // settlement contract address
  Challenge:      uint64        // challenge period duration
  Nonce:          uint64        // unique nonce to distinguish channels with identical parameters
}
```

The channel definition cannot change after creation.

## Channel Identifier

The channel identifier is derived deterministically from the channel definition.

```
ChannelId = Keccak256(AbiEncode(ChannelDefinition))
```

This ensures that:

- each unique channel definition produces a unique identifier
- the identifier can be independently computed by any participant
- no central authority is required to assign identifiers

## Channel Lifecycle

A channel progresses through the following stages.

**Creation**
A channel is created by submitting its initial state to the settlement layer along with asset deposits.

**Active Operation**
Participants exchange signed state updates off-chain. The channel remains active as long as participants cooperate.

**Settlement**
A participant submits the latest signed state to the settlement layer. After the challenge period, assets are released according to the final state allocations.

## State Advancement Rules

When a new state is proposed, the following general rules apply:

1. The state must reference a valid channel identifier
2. The state version must be strictly greater than the current version
3. The transition type must be valid for the current channel phase
4. Transition-specific validation rules must be satisfied
5. All required participants must sign the new state
6. Total asset allocations must remain consistent

## Transition Types

The protocol supports the following channel transition types:

- **Fund** — initial deposit of assets into the channel
- **Update** — general state update modifying allocations
- **Commit** — move assets from the channel into an extension
- **Release** — return assets from an extension back to the channel
- **Close** — cooperative channel closure

## Multi-State Advancement Rules

Some operations require multiple coordinated state updates across different protocol entities.

Rules for multi-state operations:

- All related state updates must be signed atomically (all or none)
- Cross-entity operations must maintain overall asset consistency
- The ordering of related updates must be deterministic

## Transition-Specific Rules

### Fund

- Valid only as the initial transition when creating a channel
- Allocations must match the deposited amounts
- All participants must sign

### Update

- Modifies asset allocations between participants
- Total assets per asset type must remain unchanged
- All participants must sign

### Commit

- Moves assets from channel allocations into an extension
- Channel allocations decrease by the committed amount
- The extension must be recognized by the protocol
- All participants must sign

### Release

- Returns assets from an extension back to channel allocations
- Channel allocations increase by the released amount
- The extension state must authorize the release
- All participants must sign

### Close

- Indicates cooperative intent to settle the channel
- Final allocations become the settlement distribution
- All participants must sign

## Checkpoint-Relevant Transitions

The following transitions may require or trigger a checkpoint to the settlement layer:

- **Fund** — requires an on-chain transaction to create the channel and deposit assets
- **Close** — may trigger a checkpoint to initiate settlement
- Any transition where a participant wishes to enforce the current state on-chain

---

Previous: [State Model](state-model.md) | Next: [Enforcement and Settlement](enforcement-and-settlement.md)
