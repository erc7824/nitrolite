# State Enforcement and Settlement

Previous: [Channel Protocol](channel-protocol.md) | Next: [Cross-Chain and Assets](cross-chain-and-assets.md)

---

This document describes how channel states are enforced on the settlement layer.

## Purpose

Enforcement ensures that any participant can fall back to the blockchain to protect their assets if off-chain cooperation fails.

The settlement layer acts as the ultimate arbiter of channel state, providing security guarantees that do not depend on participant cooperation.

## Enforcement Model

Off-chain states and enforcement states are related as follows:

- Participants advance state off-chain through signed updates
- At any time, a participant may submit the latest signed state to the settlement layer
- The settlement layer validates the submitted state and updates its record
- On-chain state always reflects the latest successfully checkpointed state

The on-chain state may lag behind the off-chain state. This is expected during normal operation.

## Checkpoint Operation

A checkpoint submits a signed state to the settlement layer.

The checkpoint process:

1. A participant constructs the enforcement representation of the latest signed state
2. The participant submits the enforcement representation along with all required signatures to the settlement contract
3. The settlement contract validates the submission
4. If valid, the on-chain state is updated

## Channel Creation

Channels are created through an enforcement operation.

The creation process:

1. Participants agree on a channel definition and initial state off-chain
2. A participant submits the channel definition, initial state, and asset deposits to the settlement contract
3. The settlement contract creates the channel record and locks the deposited assets
4. The channel is now active on both the off-chain and on-chain layers

## Enforcement Validation

The settlement layer applies the following validation rules when processing a checkpoint:

1. The channel must exist on-chain
2. The submitted state must reference the correct channel identifier
3. The state version must be strictly greater than the currently recorded version
4. All required signatures must be present and valid
5. The signature validation mode must be supported
6. Asset allocations must be consistent with the channel's total deposited assets

## Settlement State Update

When enforcement validation succeeds:

1. The on-chain channel state is updated to the submitted version
2. The on-chain allocations are updated to match the submitted state
3. A challenge period may begin depending on the transition type

After the challenge period expires without a higher-version state being submitted, the enforced state becomes final.

## Settlement Layer Interaction

The protocol interacts with the blockchain settlement layer through smart contracts.

The settlement contract provides the following capabilities:

- **Create channel** — register a new channel and lock deposits
- **Checkpoint** — update the on-chain state to a newer signed version
- **Challenge** — submit a higher-version state during a challenge period
- **Finalize** — release assets according to the final state after the challenge period

The protocol does not prescribe a specific smart contract implementation. Any settlement contract that satisfies these capabilities is compatible.

## Failure Conditions

Enforcement may fail in the following situations:

- **Invalid signatures** — one or more signatures cannot be verified
- **Stale version** — the submitted state version is not greater than the current on-chain version
- **Inconsistent allocations** — total allocations do not match channel deposits
- **Unknown channel** — the channel identifier does not correspond to a registered channel
- **Expired challenge** — a challenge was submitted after the challenge period ended

---

Previous: [Channel Protocol](channel-protocol.md) | Next: [Cross-Chain and Assets](cross-chain-and-assets.md)
