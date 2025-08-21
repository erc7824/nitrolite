---
sidebar_position: 4
title: Supported Signature Formats
description: Documentation for supported signature formats in NitroliteClient
keywords: [erc7824, statechannels, state channels, nitrolite, ethereum scaling, layer 2, off-chain, advanced, signature, format, ECDSA, EIP-191, EIP-712, EIP-1271, EIP-6492]
---

# Supported Signature Formats

The nitrolite smart contract supports multiple signature formats over a State to accommodate various use cases and compatibility with different wallets and applications.

The message being signed is a channelId and State, formatted in a specific way. The most common is a `packedState`, which is calculated as follows:

```solidity
abi.encode(channelId, state.intent, state.version, state.data, state.allocations)
```

## EOA signatures

Externally Owned Accounts (EOAs) can sign messages with their private key using the ECDSA.

Based on how the message is handled before signing, the following formats are supported:

### Raw ECDSA Signature

The message is a `packedState`, that is hashed with `keccak256` before signing. The signature is a 65-byte ECDSA signature.

### EIP-191 Signature

You can read more about EIP-191 in the [EIP-191 specification](https://eips.ethereum.org/EIPS/eip-191).

The message is a `packedState` prefixed with `"\x19Ethereum Signed Message:\n" + len(packedState)` and hashed with `keccak256` before signing. The signature is a 65-byte ECDSA signature.

### EIP-712 Signature

You can read more about EIP-712 in the [EIP-712 specification](https://eips.ethereum.org/EIPS/eip-712).

The message is an `AllowStateHash` typed data, calculated as follows:

```solidity
abi.encode(
    typeHash,
    channelId,
    state.intent,
    state.version,
    keccak256(state.data),
    keccak256(abi.encode(state.allocations))
);
```

Where `typeHash` is `AllowStateHash(bytes32 channelId,uint8 intent,uint256 version,bytes data,Allocation[] allocations)Allocation(address destination,address token,uint256 amount)`.

The message is then hashed with `keccak256`, appended to `"\x19\x01" || domainSeparator` and signed. The signature is a 65-byte ECDSA signature.

`||` is a concatenation operator, and `domainSeparator` is calculated as follows:

```solidity
keccak256(
  abi.encode(
        EIP712_TYPE_HASH,
        keccak256(name),
        keccak256(version),
        chainId,
        verifyingContract
    )
);
```

`EIP712_TYPE_HASH` is `keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)")`.

Additionally, `name`, `version` are the name and version of the Custody contract, `chainId` is the chain ID of the network, and `verifyingContract` is the address of the contract.

## Smart Contract Signatures

Smart Contracts that support [EIP-1271](https://eips.ethereum.org/EIPS/eip-1271) or [EIP-6492](https://eips.ethereum.org/EIPS/eip-6492) can sign messages using their own logic. When checking such signatures, the nitrolite smart contract will pass the `keccak256` hash of the `packedState` as a message hash for verification.

See the aforementioned EIP standards for details on how these signatures are structured and verified. If you want to add support for such signatures in your client, you probably need to look at how signature verification logic is implemented in the Smart Contract (Smart Wallet, etc) that will use them.

## Challenge Signatures

The aforementioned signature formats are used to sign States, however to submit a challenge, the user must provide a `challengerSignature`, which proves that the user has the right to challenge a Channel.

Depending on a signature format, the `challengerSignature` is calculated differently from the common State signature:

- **Raw ECDSA, EIP-191, EIP-1271 and EIP-6492**: The message (`packedState`) is suffixed with a `challenge` string (`abi.encodePacked(packedState, "challenge")`).
- **EIP-712**: The `typeHash` name is `AllowChallengeStateHash`, while type format remains the same.
