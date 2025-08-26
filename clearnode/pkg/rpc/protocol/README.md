# RPC Protocol Package

The `pkg/rpc/protocol` package provides the core data structures and utilities for the Clearnode RPC protocol. This package implements a secure, signature-based RPC communication protocol suitable for blockchain and distributed systems.

## Overview

The protocol package defines the fundamental building blocks for RPC communication:
- **Request/Response Messages**: Structured messages with cryptographic signatures
- **Payload Format**: Compact array-based JSON encoding for efficient transmission
- **Error Handling**: Distinction between client-facing and internal errors
- **Type Safety**: Strong typing with flexible parameter handling

## Core Components

### Messages

The protocol uses two main message types:

- **Request**: Contains a payload and one or more signatures
- **Response**: Contains a payload and one or more signatures

Both message types support multiple signatures, enabling multi-signature authorization scenarios.

### Payload Structure

Payloads are the core data containers in the protocol, containing:
- `RequestID` (uint64): Unique identifier for request tracking
- `Method` (string): The RPC method to invoke
- `Params` (map): Flexible parameter object
- `Timestamp` (uint64): Unix millisecond timestamp

The payload uses a compact JSON array encoding: `[id, method, params, timestamp]`

### Error Handling

The package provides a specialized `Error` type for client-facing errors:
- Protocol errors are explicitly marked for client communication
- Internal errors remain hidden from external clients
- Clear API for creating formatted error messages

## Installation

```go
import "github.com/erc7824/nitrolite/clearnode/pkg/rpc/protocol"
```

## Basic Usage

### Creating a Request

```go
// Create parameters for the RPC method
params, err := protocol.NewParams(map[string]interface{}{
    "address": "0x1234567890abcdef",
    "amount": "1000000000000000000",
})
if err != nil {
    return err
}

// Create a payload
payload := protocol.NewPayload(
    12345,              // Request ID
    "wallet_transfer",  // Method name
    params,             // Parameters
)

// Create a request (signatures would be added by the transport layer)
request := protocol.NewRequest(payload)
```

### Creating a Response

```go
// Create response parameters
resultParams, err := protocol.NewParams(map[string]interface{}{
    "txHash": "0xabcdef123456",
    "status": "confirmed",
})
if err != nil {
    return err
}

// Create response payload
responsePayload := protocol.NewPayload(
    12345,           // Same Request ID as the request
    "wallet_transfer", // Method name
    resultParams,      // Result parameters
)

// Create response
response := protocol.NewResponse(responsePayload)
```

### Error Handling

```go
// Creating client-facing errors
if amount < 0 {
    return protocol.Errorf("invalid amount: cannot be negative")
}

if balance < amount {
    return protocol.Errorf("insufficient balance: need %d but have %d", amount, balance)
}

// Internal errors (not exposed to clients) use standard Go errors
if err := db.Save(tx); err != nil {
    return fmt.Errorf("database error: %w", err) // Client won't see details
}
```

### Working with Parameters

```go
// Creating parameters from a struct
type TransferParams struct {
    From   string `json:"from"`
    To     string `json:"to"`
    Amount string `json:"amount"`
}

transferReq := TransferParams{
    From:   "0x111...",
    To:     "0x222...",
    Amount: "1000000000000000000",
}

params, err := protocol.NewParams(transferReq)
if err != nil {
    return err
}

// Extracting parameters into a struct
var received TransferParams
if err := params.Translate(&received); err != nil {
    return protocol.Errorf("invalid parameters: %v", err)
}
```

## Advanced Usage

### Multi-Signature Requests

```go
// Create a request with multiple signatures
request := protocol.NewRequest(
    payload,
    signature1,  // Primary signer
    signature2,  // Co-signer
    signature3,  // Additional authorization
)
```

### Signature Verification

```go
// Verify request signatures
signers, err := request.GetSigners()
if err != nil {
    return protocol.Errorf("invalid signatures: %v", err)
}

// For responses, verify server signature
responseSigners, err := response.GetSigners()
if err != nil {
    return fmt.Errorf("invalid response signature: %w", err)
}
```

### Custom JSON Marshaling

The payload automatically marshals to the compact array format:

```go
payload := protocol.NewPayload(123, "test_method", params)
data, _ := json.Marshal(payload)
// Output: [123,"test_method",{...params...},1634567890123]
```

### Timestamp Validation

```go
// Validate request timestamp (example: 5-minute window)
maxAge := 5 * time.Minute
requestTime := time.Unix(0, int64(payload.Timestamp)*int64(time.Millisecond))
if time.Since(requestTime) > maxAge {
    return protocol.Errorf("request expired: timestamp too old")
}
```

## Security Considerations

1. **Signature Verification**: Always verify signatures before processing requests
2. **Timestamp Validation**: Implement replay protection using timestamps
3. **Parameter Validation**: Thoroughly validate all parameters before processing
4. **Error Messages**: Use `protocol.Errorf()` for safe client-facing errors
5. **Request IDs**: Use unique request IDs to prevent duplicate processing

## Testing

The package includes comprehensive tests for all components. Run tests with:

```bash
go test ./pkg/rpc/protocol
```

## Dependencies

- `github.com/erc7824/nitrolite/clearnode/pkg/sign`: Signature types and interfaces
- Standard library: `encoding/json`, `errors`, `fmt`, `time`

## See Also

- [Clearnode Protocol Specification](../../docs/Clearnode.protocol.md)
- [API Documentation](../../docs/API.md)
- [Entity Documentation](../../docs/Entities.md)
