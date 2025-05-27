# Go-Nitrolite

[![GoDoc](https://godoc.org/github.com/erc7824/go-nitrolite?status.svg)](https://godoc.org/github.com/erc7824/go-nitrolite)

A lightweight Go implementation of Ethereum State Channels based on ERC-7824 (Nitro protocol).

## Features

- Channel management with participants, adjudicator, challenge periods, and nonces
- State representation with custom data, allocations, and signatures
- Fund allocation tracking with destination, token, and amount
- Cryptographic signing and verification using keccak256
- Standalone signature utilities for arbitrary data

## Installation

```bash
go get github.com/erc7824/go-nitrolite
```

## Usage

### Basic Types

```go
import (
    "github.com/erc7824/go-nitrolite"
    "github.com/ethereum/go-ethereum/common"
    "math/big"
)

// Create a new channel
channel := nitrolite.Channel{
    Participants: [2]common.Address{addr1, addr2},
    Adjudicator:  adjAddr,
    Challenge:    86400, // 24 hours
    Nonce:        1,
}

// Create a state
state := nitrolite.State{
    Data: []byte("channel data"),
    Allocations: [2]nitrolite.Allocation{
        {Destination: addr1, Token: tokenAddr, Amount: big.NewInt(100)},
        {Destination: addr2, Token: tokenAddr, Amount: big.NewInt(200)},
    },
}
```

### Signing and Verification

```go
// To sign a state, first serialize it to bytes
stateData := []byte("serialized state data")

// Sign the state data
signature, err := nitrolite.Sign(stateData, privateKey)
if err != nil {
    // Handle error
}

// Add signature to the state
state.Sigs = append(state.Sigs, signature)

// Verifying a signature
isValid, err := nitrolite.Verify(stateData, signature, address)
if err != nil {
    // Handle error
}
if isValid {
    // Signature is valid
}
```

### Signature Functions

```go
// Sign arbitrary data
data := []byte("data to sign")
signature, err := nitrolite.Sign(data, privateKey)
if err != nil {
    // Handle error
}

// Verify a signature on arbitrary data
isValid, err := nitrolite.Verify(data, signature, address)
if err != nil {
    // Handle error
}
```

## Developer Notes

### Signatures

- Signatures follow Ethereum's format with V, R, S components
- V is adjusted by +27 per Ethereum convention
- Signatures are created using Keccak256 hashing
- The library uses go-ethereum's crypto package for signing operations
- The standalone signing utilities can be used with any data, not just channel states

### Testing

Run the tests with:

```bash
go test ./...
```

## License

MIT