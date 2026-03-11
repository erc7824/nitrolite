# Cryptography

Previous: [Terminology](terminology.md) | Next: [State Model](state-model.md)

---

This document defines how protocol objects are encoded, hashed, and signed.

All rules are described as algorithms and canonical procedures, independent of any specific programming language.

## Purpose

Cryptography in the Nitrolite protocol serves three functions:

1. **Authentication** — proving that a specific participant authorized a state update
2. **Integrity** — ensuring that signed data has not been modified
3. **Replay protection** — preventing previously signed states from being reused in unintended contexts

## Cryptographic Algorithms

The protocol uses the following cryptographic primitives.

**Signature Algorithm**
ECDSA over the secp256k1 curve, producing a 65-byte signature (r, s, v).

**Hash Function**
Keccak-256, producing a 32-byte digest.

## Canonical Encoding

Protocol objects that require signing must be encoded into a canonical binary representation before hashing.

The canonical encoding uses ABI encoding as defined by the Ethereum ABI specification. This ensures deterministic byte sequences regardless of implementation language.

Rules:

- All fields are encoded in the order defined by the protocol structure
- Dynamic types (byte arrays, strings) follow ABI encoding rules for dynamic types
- Encoding must be deterministic — the same logical object must always produce the same byte sequence

## Message Digest Construction

The digest of a signable payload is constructed as follows:

1. Encode the object using canonical encoding
2. Compute the Keccak-256 hash of the encoded bytes

The resulting 32-byte digest is the value that is signed.

## Signature Envelope

A protocol signature consists of:

```
Signature {
  V: byte       // recovery identifier
  R: bytes32    // ECDSA r component
  S: bytes32    // ECDSA s component
}
```

The signer's address is recovered from the signature and the message digest. The protocol does not transmit the signer's public key or address alongside the signature.

## Signature Validation Modes

The protocol supports multiple signature validation modes to allow different key types and authorization schemes.

Each signature includes a mode byte prefix that determines how the signature is validated.

**Default Mode (0x00)**
Standard ECDSA signature validation. The signer's address is recovered directly from the signature. The recovered address must match the expected participant address.

**Session Key Mode (0x01)**
Delegated signature validation. The signer's address is recovered from the signature and verified against a registered session key authorization.

## Signable Object Classes

The following protocol objects require signatures:

- **Channel State** — the primary state of a channel, signed by all participants
- **Application State** — the state of an application session, signed by session participants
- **Session Key Authorization** — a delegation granting signing authority to a session key

## Session Key Authorization

A participant may delegate signing authority to a session key.

The authorization is constructed as follows:

1. The participant signs a message containing:
   - the session key address
   - authorization metadata (scope, expiration)
2. The authorization signature is produced using the participant's primary key
3. The session key may then produce signatures on behalf of the participant within the authorized scope

Session key signatures must include the authorization proof alongside the session key signature.

## Replay Protection

The protocol prevents replay attacks through the following mechanisms:

**Channel Identifier**
Each channel has a unique identifier derived from its definition. States are bound to a specific channel.

**State Version**
Each state includes a monotonically increasing version number. The settlement layer rejects states with a version less than or equal to the currently enforced version.

**Chain Identifier**
States include chain-specific identifiers preventing cross-chain replay.

---

Previous: [Terminology](terminology.md) | Next: [State Model](state-model.md)
