# Signature Validators

This document describes the pluggable signature validation system in the Nitrolite protocol.

---

## Overview

The protocol supports flexible signature validation through the `ISignatureValidator` interface. All validators implement two methods:

- `validateSignature(channelId, signingData, signature, participant)` - Validates a participant's signature
- `validateChallengerSignature(channelId, signingData, signature, user, node)` - Validates a challenger's signature

Validators receive the core state data (`signingData`) and `channelId` separately, allowing them to construct the full message according to their signing scheme.

---

## Validator Selection

### Node Validator Registry

The protocol uses a **per-node validator registry** system. Each node can register signature validators and assign them 1-byte identifiers (0x01-0xFF). Both users and nodes select validators from the node's registry when signing channel states.

**Design rationale:** In the Nitrolite off-chain protocol, the node acts as the orchestrator and decides which signature validators are supported. Users select from the node-approved validators. This ensures:

- Nodes can enforce their security requirements
- Users benefit from node-vetted validator implementations
- Cross-chain compatibility (validator addresses don't affect channelId or signature verification)

### Validator Registration

Nodes register validators by providing a signature over the validator configuration. This allows node operators to use cold storage or hardware wallets without exposing private keys to send transactions.

**Registration message:**

```solidity
bytes memory message = abi.encode(validatorId, validatorAddress, block.chainid);
```

The signature is verified using ECDSA recovery:

1. Try EIP-191 recovery first (standard for wallet software)
2. Fall back to raw ECDSA if needed
3. Verify recovered address matches the node address

**Key properties:**

- Includes `block.chainid` for cross-chain replay protection (registrations are chain-specific)

- Anyone can submit a registration transaction (relayer-friendly)
- Node's private key only signs, never sends transactions
- Validator ID 0x00 is reserved for the default validator
- Registration is immutable (cannot change once set)
- 255 validators per node (0x01-0xFF)

### Signature Format

All signatures in the protocol follow this structure:

```txt
[validator_id: 1 byte][signature_data: variable length]
```

- `0x00` = Use ChannelHub's default validator
- `0x01-0xFF` = Look up validator in node's registry

The first byte determines which validator verifies the signature. The remaining bytes are passed to the selected validator for verification.

---

## Domain Separation: ChannelHub vs Validators

The protocol maintains clear separation between protocol concerns and cryptographic concerns:

### ChannelHub Responsibilities

- Define protocol message structure (when and how channelId binds to states)
- Manage channel lifecycle and state transitions
- Select appropriate validators based on signature first byte
- Handle validator registration (infrastructure concern)

### Validator Responsibilities

- Verify cryptographic signatures using specific schemes (ECDSA, multi-sig, session keys, etc.)
- Support different signature formats and recovery mechanisms
- Remain agnostic to protocol-level message structure

### Why This Matters

**State validation** requires channelId binding for security. Validators receive `channelId` and `signingData` separately because ChannelHub controls *when* and *how* channelId is included in signed messages. This is a protocol-level security requirement, not a cryptographic concern.

**Validator registration** is an infrastructure operation that happens outside the channel state validation flow. It uses direct ECDSA recovery in ChannelHub rather than going through the validator abstraction, because:

- Registration has no channelId (different domain)
- Registration is operational setup, not protocol state transition
- All node operators use ECDSA-capable wallets for registration
- Keeps `ISignatureValidator` focused on its primary purpose

This separation ensures validators remain pluggable for state verification while keeping protocol-level concerns within ChannelHub.

---

## Cross-Chain Compatibility

The node validator registry design solves a critical cross-chain problem: validator contracts may not deploy to the same address on all chains.
This enables true cross-chain operation without requiring deterministic deployment (CREATE2) across all chains.

---

## ECDSAValidator

**Location:** `src/sigValidators/ECDSAValidator.sol`

### Description

Default validator supporting standard ECDSA signatures. Automatically tries both EIP-191 (with Ethereum prefix) and raw ECDSA formats.

### Signature Format

65 bytes: `[r: 32 bytes][s: 32 bytes][v: 1 byte]`

### Validation Logic

1. Try EIP-191 recovery first (most wallets use this)
2. If fails, try raw ECDSA recovery
3. Return `VALIDATION_SUCCESS` if recovered address matches participant, `VALIDATION_FAILURE` otherwise

### Use Cases

- Standard wallet signatures (MetaMask, WalletConnect, hardware wallets)
- Most common validator for all channels
- Recommended for both users and nodes

---

## SessionKeyValidator

**Location:** `src/sigValidators/SessionKeyValidator.sol`

### Description

Enables delegation to temporary session keys. Useful for hot wallets, time-limited access, and gasless transactions.

### Signature Format

```solidity
struct SessionKeyAuthorization {
    address sessionKey;      // Delegated signer
    bytes32 metadataHash;    // Hashed expiration, permissions, etc.
    bytes authSignature;     // Participant's authorization (65 bytes)
}

bytes sigBody = abi.encode(SessionKeyAuthorization, bytes sessionKeySignature)
```

### Validation Logic

**Two checks:**

1. Participant authorized the session key: `authData = abi.encode(sessionKey, metadataHash)`
2. Session key signed the state

Both use EIP-191 first, then raw ECDSA if that fails.

### Metadata

Application-defined data encoding expiration, allowed channels, and permissions. **Validated off-chain by Clearnode, not on-chain.**

---

## SECURITY: SessionKeyValidator Usage

⚠️ **CRITICAL: SessionKeyValidator is for USER usage only, NOT for nodes.**

### Users: Safe ✅

- Clearnode validates metadata (expiration, scope, permissions)
- Node must countersign (provides protection)
- Limited damage if compromised (Clearnode rejects invalid requests)
- Revocable (switch to main key anytime)

### Nodes: Unsafe ⚠️

- User has no off-chain validation mechanism
- User cannot verify metadata constraints (only hash is checked on-chain)
- If node's session key is compromised: unlimited, irrevocable authority
- User must blindly trust node's session key

---

## Validator Comparison

| Feature | ECDSAValidator | SessionKeyValidator |
| --------- | --------------- | --------------------- |
| **Signature Size** | 65 bytes | ~200+ bytes |
| **Gas Cost** | Low | Medium-High |
| **Hot Wallet Safe** | No | Yes |
| **Time-Limited** | No | Yes |
| **User Usage** | ✅ Recommended | ✅ Recommended |
| **Node Usage** | ✅ Recommended | ⚠️ Not recommended |

---

## Key Points

- Validators are registered per-node with 1-byte IDs
- Both users and nodes select from the node's validator registry
- Validator ID 0x00 is reserved for ChannelHub's default validator
- Registration uses signature-based authorization (node key signs, anyone can relay)
- Registration is immutable (255 validators per node, cannot change once set)
- Validator addresses don't affect channelId (enables cross-chain compatibility)
- SessionKeyValidator metadata is enforced off-chain by Clearnode
- Nodes should only use ECDSAValidator until a node-specific session key validator exists
