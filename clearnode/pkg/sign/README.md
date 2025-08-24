# Blockchain-Agnostic Signing Library

A blockchain-agnostic library for cryptographic operations.

## Core Design

This library separates generic interfaces from specific blockchain implementations.

## Features

* **Blockchain-Agnostic Interfaces**: Defines a standard set of interfaces for cryptographic operations.
* **EVM Implementation**: Includes a ready-to-use implementation for Ethereum and other EVM-compatible chains.
* **Easily Extensible**: Simple to add support for new blockchains like Solana, Bitcoin, etc.
* **Type-Safe**: Provides distinct types for `Address`, `Signature`, `PublicKey`, and `PrivateKey`.

## Usage

See the Go package documentation and examples by running:
```bash
go doc -all github.com/erc7824/nitrolite/clearnode/pkg/sign
```

Or view examples in your IDE or on pkg.go.dev.

## Extending the Library

To add support for a new blockchain (e.g., Solana):

1.  Create a new package (e.g., `sign/solana`).
2.  Implement the interfaces defined in the `sign` package.
3.  Provide a constructor (e.g., `NewSolanaSigner(...)`) that returns a `sign.Signer`.
