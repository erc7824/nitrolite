# Security and Limitations

Previous: [Interactions](interactions.md) | Next: [Extensions Overview](extensions/overview.md)

---

This document describes the security guarantees of the Nitrolite protocol and the limitations of the current version.

## Security Goals

The protocol aims to guarantee:

- **Asset safety** — participants cannot lose assets without signing a state that authorizes the loss
- **State finality** — the latest mutually signed state can always be enforced on-chain
- **Non-repudiation** — a participant cannot deny having signed a state
- **Censorship resistance** — any participant can independently enforce state on the settlement layer

## Off-Chain Safety

The protocol protects against invalid or malicious state submissions through:

**Signature requirements**
Every state update requires valid signatures from all required participants. No participant can unilaterally change the state.

**Version ordering**
State versions are strictly increasing. Old states cannot replace newer states.

**Asset conservation**
State transitions must preserve total asset amounts. No assets can be created or destroyed through state updates.

**Transition validation**
Each state update must satisfy transition-specific rules. Invalid transitions are rejected.

## Enforcement Guarantees

The settlement layer provides the following guarantees:

- Any participant can submit the latest signed state at any time
- The settlement contract accepts only states with valid signatures and a higher version than the current on-chain state
- After the challenge period, the enforced state becomes final
- Final state allocations determine asset distribution

## Node Responsibilities

Nodes are expected to:

- Remain available to process state updates during channel operation
- Respond to checkpoint requests in a timely manner
- Maintain accurate records of the latest channel states
- Submit enforcement transactions when requested by participants

If a node becomes unavailable, participants retain the ability to enforce the last mutually signed state directly on the settlement layer.

## Liquidity Requirements

The protocol assumes:

- Nodes maintain sufficient liquidity to fund their side of channel allocations
- Deposit assets must be available on the settlement layer before channel creation
- Cross-chain operations require liquidity on each involved chain

Insufficient node liquidity may prevent channel creation or state updates but does not compromise the safety of existing channels.

## Current Trust Assumptions

In the current protocol version, participants must trust nodes for:

- **Liveness** — nodes must be online to facilitate state advancement
- **Cross-chain relay** — nodes relay cross-chain state updates; trustless cross-chain enforcement is not yet implemented
- **Timely enforcement** — nodes are expected to submit checkpoints when requested; delayed enforcement may affect user experience but not asset safety

Participants do not need to trust nodes for:

- **Asset custody** — assets can always be recovered through on-chain enforcement
- **State validity** — invalid states are rejected by signature and validation rules

## Known Limitations

The following capabilities are not yet implemented:

- Trustless cross-chain state enforcement
- Multi-party channels (more than two participants)
- Watchtower services for automated enforcement
- State compression for long-lived channels
- Formal verification of protocol rules

## Future Improvements

Planned protocol improvements include:

- Trustless cross-chain bridges for enforcement without relay trust
- Support for multi-party channels
- Watchtower integration for monitoring and automated enforcement
- State pruning mechanisms for long-running channels
- Extended extension framework for additional application types

---

Previous: [Interactions](interactions.md) | Next: [Extensions Overview](extensions/overview.md)
