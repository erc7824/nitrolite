# Nitrolite Protocol Documentation Structure

This document defines the **structure and organization of the Nitrolite protocol specification**.

Its goal is to ensure that the protocol is described:

* clearly and consistently
* independently from implementation languages
* independently from specific transports
* independently from specific storage systems
* in a form that is easy for developers and AI systems to understand

The protocol specification should remain **concise but precise**, avoiding unnecessary verbosity while still describing all required rules.

---

# Protocol Documentation Directory

The protocol documentation must be located under:

```
docs/protocol/
```

Recommended structure:

```
docs/protocol/
  overview.md
  terminology.md
  cryptography.md
  state-model.md
  channel-protocol.md
  enforcement-and-settlement.md
  cross-chain-and-assets.md
  interactions.md
  security-and-limitations.md
  extensions/
    overview.md
    app-sessions.md
```

The documents are listed below in the **recommended writing order**, since later documents depend on concepts introduced earlier.

---

# 1. overview.md

## Purpose

Provide a high-level explanation of the Nitrolite protocol without technical detail.

This document helps readers understand the **problem being solved** and the **overall design approach**.

## Sections

```
# Nitrolite Protocol Overview

## Purpose
What Nitrolite is designed to achieve.

## Design Goals
Main protocol objectives, such as:
- off-chain scalability
- blockchain security guarantees
- cross-chain asset interaction
- extensibility

## System Roles
Participants in the protocol:
- User
- Node
- Blockchain settlement layer

## High-Level Architecture
Conceptual layers of the system.

## Core Protocol Concepts
Brief introduction to:
- channels
- states
- state advancement
- state enforcement
- unified assets
- extensions

## Protocol Layers
Explanation of the separation between:
- core protocol
- extension layer
- settlement layer

## Protocol Version
Current protocol version and compatibility expectations.
```

---

# 2. terminology.md

## Purpose

Define all protocol terms used throughout the documentation.

Terminology must be **strictly defined once** and reused consistently.

## Sections

```
# Terminology

## Naming Conventions
General naming rules used in the protocol.

## Core Entities
Definitions of:
- Channel
- Channel Definition
- Channel State
- Participant
- Asset

## State Concepts
Definitions of:
- State
- State Version
- State Advancement
- State Enforcement
- Transition
- Intent

## Cryptographic Concepts
Definitions of:
- Signature
- Signer
- Session Key
- Signature Validation Mode

## Ledger Concepts
Definitions of:
- Ledger
- Home Ledger
- Non-Home Ledger

## Protocol Operations
Definitions of:
- Checkpoint
- Commit
- Release

## Extension Concepts
Definitions of:
- Extension
- Application Session
- Application State
```

---

# 3. cryptography.md

## Purpose

Define how protocol objects are encoded, hashed, and signed.

This document must describe **algorithms and canonical rules**, not specific programming language implementations.

## Sections

```
# Cryptography

## Purpose
Role of cryptography in the protocol.

## Cryptographic Algorithms
Algorithms used by the protocol, such as:
- signature algorithm
- hash function

## Canonical Encoding
Definition of the canonical binary encoding used for signable protocol objects.

## Message Digest Construction
Rules for generating the hash of a signable payload.

## Signature Envelope
Structure of the protocol signature envelope.

## Signature Validation Modes
Mechanism allowing multiple signature verification methods.

## Signable Object Classes
Classes of protocol objects that require signatures.

## Session Key Authorization
Rules allowing delegated signing.

## Replay Protection
Mechanisms preventing replay attacks.
```

---

# 4. state-model.md

## Purpose

Describe the abstract structure of protocol states.

This document explains **how states are defined and structured**, but does not describe operational flows.

## Sections

```
# State Model

## Purpose
Why states exist and what they represent.

## Common State Fields
Shared properties of protocol states such as:
- entity identifier
- version
- entity-specific payload

## State Identification and Versioning
Rules governing state identity and version progression.

## Channel State
Logical structure of the primary protocol state.

## Off-Chain Representation
Operational representation used during off-chain state advancement.

## Enforcement Representation
Representation used when enforcing a state on the settlement layer.

## Representation Mapping
How off-chain state representations are converted into enforcement representations.

## Transition Field
Role of transition information in state updates.

## Ledger Components
Description of home and non-home ledger structures.

## State Consistency Rules
Conditions required for states to be considered valid.
```

---

# 5. channel-protocol.md

## Purpose

Describe how channels operate and how states evolve through off-chain state advancement.

## Sections

```
# Channel Protocol

## Purpose
Role of channels in the protocol.

## Channel Definition
Immutable parameters defining a channel.

## Channel Identifier
How channel identifiers are derived.

## Channel Lifecycle
Stages such as:
- creation
- active operation
- settlement

## State Advancement Rules
General rules for validating and accepting new channel states.

## Transition Types
List of supported channel transition types.

## Multi-State Advancement Rules
Protocol rules governing operations that require multiple state updates.

## Transition-Specific Rules
Detailed validation and side-effect rules for each transition type.

## Checkpoint-Relevant Transitions
Transitions that require or may trigger state enforcement.
```

---

# 6. enforcement-and-settlement.md

## Purpose

Describe how channel states are enforced on the settlement layer.

## Sections

```
# State Enforcement and Settlement

## Purpose
Role of enforcement in maintaining protocol guarantees.

## Enforcement Model
Relationship between off-chain states and enforcement states.

## Checkpoint Operation
Submitting a state to the settlement layer.

## Channel Creation
How channels are created through enforcement.

## Enforcement Validation
Rules applied when validating an enforcement request.

## Settlement State Update
How enforcement updates the on-chain channel state.

## Settlement Layer Interaction
General description of blockchain interaction.

## Failure Conditions
Situations in which enforcement may fail.
```

---

# 7. cross-chain-and-assets.md

## Purpose

Describe the unified asset model and cross-chain functionality.

## Sections

```
# Cross-Chain and Asset Model

## Purpose
Why unified assets exist.

## Unified Asset Concept
Operating on assets independent of specific blockchains.

## Home Chain
Definition and role of the home chain.

## Home and Non-Home Ledger Roles
Purpose and responsibilities of each ledger.

## Cross-Chain Deposit Rules
Procedure for depositing assets from non-home chains.

## Cross-Chain Withdrawal Rules
Procedure for withdrawing assets to non-home chains.

## Ledger Swap Rules
Rules governing asset transfers between ledgers.

## Home Chain Migration
Procedure for changing the home chain of an asset.

## Cross-Chain Replay Protection
Mechanisms preventing replay across multiple chains.

## Current Version Notes
Special considerations for the current protocol version.
```

---

# 8. interactions.md

## Purpose

Define the logical communication protocol between participants.

This document defines **semantic protocol operations**, independent of transport technologies such as WebSocket or gRPC.

## Sections

```
# Interaction Model

## Purpose
How participants exchange protocol messages.

## Connection Assumptions
Assumptions about the communication channel.

## Message Envelope
Common structure shared by all protocol messages.

## Core Operations
Overview of supported protocol operations.

## Operation: <name>
### Purpose
### Request
### Successful Result
### Failure Result
### Related Events

## Event Messages
Asynchronous notifications generated by the protocol.

## Correlation and Identifiers
How responses are associated with requests.

## Error Handling
Rules for communicating failures.

## Message Ordering
Requirements governing message sequencing.
```

---

# 9. security-and-limitations.md

## Purpose

Describe the security guarantees of the protocol and the limitations of the current version.

This document must clearly state **what the protocol guarantees and what it does not guarantee**.

## Sections

```
# Security and Limitations

## Security Goals
What the protocol aims to guarantee.

## Off-Chain Safety
Protection against invalid or malicious state submissions.

## Enforcement Guarantees
Protection provided by checkpointing and settlement.

## Node Responsibilities
Operational expectations for nodes.

## Liquidity Requirements
Assumptions about node liquidity.

## Current Trust Assumptions
Areas where users must trust nodes.

## Known Limitations
Capabilities not yet implemented.

## Future Improvements
Planned protocol improvements.
```

---

# 10. extensions/overview.md

## Purpose

Explain the extension model used by the protocol.

Extensions allow additional functionality without modifying the core protocol.

## Sections

```
# Protocol Extensions

## Purpose
Why extensions exist.

## Extension Model
How extensions interact with the core protocol.

## Extension Lifecycle
Creation and use of extensions.

## Integration with Channels
How extensions interact with channel states.

## Extension Safety
Constraints extensions must follow.
```

---

# 11. extensions/app-sessions.md

## Purpose

Describe the application session extension currently supported by the protocol.

## Sections

```
# Application Sessions

## Purpose
Providing off-chain application functionality.

## Application Session Entity
Structure of an application session.

## Application State
State associated with an application.

## Application Session Keys
Delegated signing for application sessions.

## Commit Operation
Moving assets from a channel into an extension.

## Release Operation
Returning assets from an extension back to a channel.

## Interaction with Channel Protocol
How application sessions coordinate with channel states.

## Current Limitations
Limitations of the current extension implementation.
```

---

# Summary

This documentation structure provides:

* clear separation of protocol responsibilities
* independence from implementation technologies
* support for protocol evolution
* a manageable number of documents
* compatibility with future extensions

Following this structure ensures that the Nitrolite protocol specification remains **structured, understandable, and maintainable**.
