# State Model

Previous: [Cryptography](cryptography.md) | Next: [Channel Protocol](channel-protocol.md)

---

This document describes the abstract structure of protocol states.

It explains how states are defined and structured. Operational flows are described in separate documents.

## Purpose

States represent the current agreed configuration of protocol entities. The state model defines:

- what information a state contains
- how states are identified and versioned
- how states are represented for off-chain and on-chain use

## Common State Fields

All protocol states share the following common properties:

```
EntityId:  bytes32    // unique identifier of the entity this state belongs to
Version:   uint64     // monotonically increasing version number
Payload:   bytes      // entity-specific state data
```

## State Identification and Versioning

Each state is identified by the combination of its entity identifier and version number.

Rules:

- EntityId is derived from the entity definition and is immutable
- Version starts at 0 for the initial state
- Each subsequent state must have a version strictly greater than the previous state
- Version numbers do not need to be sequential, only strictly increasing

## Channel State

The channel state is the primary protocol state. It represents the current configuration of a channel.

A channel state contains:

```
ChannelState {
  ChannelId:     bytes32         // derived from the channel definition
  Version:       uint64          // state version
  Allocations:   []Allocation    // asset allocations per participant
  Transition:    Transition      // describes the operation that produced this state
}
```

### Allocation

An allocation describes how assets are distributed among participants.

```
Allocation {
  Participant:  address    // participant address
  Asset:        AssetId    // asset identifier
  Amount:       uint256    // allocated amount
}
```

## Off-Chain Representation

During off-chain state advancement, channel states use an operational representation optimized for:

- human readability
- ease of validation
- efficient signature generation

The off-chain representation contains the full channel state including all allocations, transition data, and metadata needed for validation.

## Enforcement Representation

When a state is submitted to the settlement layer, it uses an enforcement representation optimized for:

- on-chain verification
- gas efficiency
- deterministic encoding

The enforcement representation contains the minimum data required for the settlement layer to validate and enforce the state.

## Representation Mapping

The off-chain representation must be convertible to the enforcement representation through a deterministic mapping.

Rules:

- The mapping must be lossless for enforcement-relevant fields
- The enforcement representation must be derivable from the off-chain representation without additional information
- Both representations must produce the same message digest when signed

## Transition Field

Each state update includes a transition that describes the operation that produced the new state.

The transition contains:

```
Transition {
  Type:    uint8    // transition type identifier
  Data:    bytes    // transition-specific parameters
}
```

The transition type determines the validation rules applied to the state update.

## Ledger Components

A channel state may include multiple ledger components.

**Home Ledger**
The primary ledger on the channel's home chain. Contains the authoritative asset allocations.

**Non-Home Ledgers**
Secondary ledgers on other chains. Track asset allocations for cross-chain operations.

Each ledger component follows the same allocation structure but is associated with a specific chain identifier.

## State Consistency Rules

A state is considered valid only if all of the following conditions hold:

- The channel identifier matches the channel definition
- The version is strictly greater than the previously accepted version
- Total allocations per asset are consistent (no assets created or destroyed)
- The transition type is recognized and its specific rules are satisfied
- All required signatures are present and valid

---

Previous: [Cryptography](cryptography.md) | Next: [Channel Protocol](channel-protocol.md)
