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

### 1. Creating a Signer and signing a message.

```go
package main

import (
    "fmt"
    "log"

    "your-module-path/sign"
    "your-module-path/sign/ethereum"
)

func main() {
    pkHex := "0x..." // Your private key

    // Create a new signer. It returns the generic sign.Signer interface.
    signer, err := ethereum.NewEthereumSigner(pkHex)
    if err != nil {
        log.Fatal(err)
    }

    // You can now use the signer for generic operations.
    fmt.Println("Address:", signer.Address())
    
    message := []byte("hello world")
    signature, err := signer.PrivateKey().Sign(message)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Signature:", signature)
}
```

### 2. Using Blockchain-Specific Features

For functions that are unique to a blockchain, like EIP-712 recovery for Ethereum, call them directly from the implementation package.

```go
import (
    "your-module-path/sign"
    "your-module-path/sign/ethereum"
    "github.com/ethereum/go-ethereum/signer/core/apitypes"
)

func recoverSignature(typedData apitypes.TypedData, sig sign.Signature) (string, error) {
    // Call the function directly from the `ethereum` package.
    return ethereum.RecoverAddressEip712(typedData, sig)
}
```

## Extending the Library

To add support for a new blockchain (e.g., Solana):

1.  Create a new package (e.g., `sign/solana`).
2.  Implement the interfaces defined in the `sign` package.
3.  Provide a constructor (e.g., `NewSolanaSigner(...)`) that returns a `sign.Signer`.
