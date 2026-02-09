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

### Two-Tier System

**Default Validator:** Set in ChannelHub constructor, used when signature's first byte is `0x00`

**Channel-Specific Validator:** Specified in Channel Definition, used when signature's first byte is `0x01`

### Signature Format

```txt
[validator_type: 1 byte][signature_data: variable length]
```

- `0x00` = Use default validator
- `0x01` = Use channel validator

The validator address is part of the Channel Definition and cannot be changed after channel creation.

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

- Validator choice is immutable (part of channelId)
- Each participant chooses default (0x00) or channel (0x01) validator per signature
- SessionKeyValidator metadata is enforced off-chain by Clearnode
- Nodes should only use ECDSAValidator until a node-specific session key validator exists
