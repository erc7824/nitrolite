# Ethereum Signing Package (`ethereum`)

This package provides an implementation of the `sign` interfaces for Ethereum and other EVM-compatible blockchains.

It handles all Ethereum-specific cryptographic details, such as Keccak256 hashing, ECDSA `secp256k1` keys, and signature recovery.


## Usage

### 1. Creating a Signer

Create a signer using a hex-encoded private key. The constructor returns a generic `sign.Signer` interface, which allows it to be used in blockchain-agnostic code.

```go
import (
    "fmt"
    "your-module-path/sign"
    "your-module-path/sign/ethereum"
)

func main() {
    pkHex := "0x..." // Your private key

    // NewEthereumSigner returns a generic sign.Signer
    signer, err := ethereum.NewEthereumSigner(pkHex)
    if err != nil {
        // handle error
    }
    fmt.Println("Signer Address:", signer.Address())
}
```


### 2. Signing Data

The `Sign` method follows the standard Ethereum practice of first hashing the data with Keccak256.

```go
message := []byte("some data to sign")

// The Sign method handles the hashing automatically
signature, err := signer.PrivateKey().Sign(message)

fmt.Println("Signature:", signature)
```


### 3. Signature Recovery

This package provides utility functions to recover the signer's address from a signature, which is a common requirement in Ethereum applications.

#### EIP-191 (Standard Message)

```go
// Recover the address from the message and signature
recoveredAddr, err := ethereum.RecoverAddress(message, signature)

// recoveredAddr should match signer.Address().String()
fmt.Println("Recovered Address:", recoveredAddr)
```

#### EIP-712 (Typed Data)

```go
import "github.com/ethereum/go-ethereum/signer/core/apitypes"

var typedData apitypes.TypedData
var eip712Signature sign.Signature

// Recover the address from the EIP-712 typed data
recoveredAddr, err := ethereum.RecoverAddressEip712(typedData, eip712Signature)
```
