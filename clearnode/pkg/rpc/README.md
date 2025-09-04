# RPC Package

The `pkg/rpc` package provides the core data structures and utilities for the Clearnode RPC protocol. This package implements a secure, signature-based RPC communication protocol suitable for blockchain and distributed systems.

## Overview

The RPC package defines the fundamental building blocks for RPC communication:
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

### Client Communication

The package includes a `Dialer` interface and a WebSocket implementation for client-side RPC communication:
- **Thread-safe**: Supports concurrent RPC calls
- **Automatic reconnection**: Built-in ping/pong mechanism
- **Event handling**: Separate channel for unsolicited server events
- **Context support**: Full context cancellation and timeout support

## Installation

```go
import "github.com/erc7824/nitrolite/clearnode/pkg/rpc"
```

## Basic Usage

### Creating a Request

```go
// Create parameters for the RPC method
params, err := rpc.NewParams(map[string]interface{}{
    "address": "0x1234567890abcdef",
    "amount": "1000000000000000000",
})
if err != nil {
    return err
}

// Create a payload
payload := rpc.NewPayload(
    12345,              // Request ID
    "wallet_transfer",  // Method name
    params,             // Parameters
)

// Create a request (signatures would be added by the transport layer)
request := rpc.NewRequest(payload)
```

### Creating a Response

```go
// Create response parameters
resultParams, err := rpc.NewParams(map[string]interface{}{
    "txHash": "0xabcdef123456",
    "status": "confirmed",
})
if err != nil {
    return err
}

// Create response payload
responsePayload := rpc.NewPayload(
    12345,           // Same Request ID as the request
    "wallet_transfer", // Method name
    resultParams,      // Result parameters
)

// Create response
response := rpc.NewResponse(responsePayload)
```

### Error Handling

```go
// Creating client-facing errors
if amount < 0 {
    return rpc.Errorf("invalid amount: cannot be negative")
}

if balance < amount {
    return rpc.Errorf("insufficient balance: need %d but have %d", amount, balance)
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

params, err := rpc.NewParams(transferReq)
if err != nil {
    return err
}

// Extracting parameters into a struct
var received TransferParams
if err := params.Translate(&received); err != nil {
    return rpc.Errorf("invalid parameters: %v", err)
}
```

## Advanced Usage

### Multi-Signature Requests

```go
// Create a request with multiple signatures
request := rpc.NewRequest(
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
    return rpc.Errorf("invalid signatures: %v", err)
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
payload := rpc.NewPayload(123, "test_method", params)
data, _ := json.Marshal(payload)
// Output: [123,"test_method",{...params...},1634567890123]
```

### Timestamp Validation

```go
// Validate request timestamp (example: 5-minute window)
maxAge := 5 * time.Minute
requestTime := time.Unix(0, int64(payload.Timestamp)*int64(time.Millisecond))
if time.Since(requestTime) > maxAge {
    return rpc.Errorf("request expired: timestamp too old")
}
```

### WebSocket Client Usage

```go
// Create and configure a WebSocket dialer
cfg := rpc.DefaultWebsocketDialerConfig
cfg.EventChanSize = 100  // Buffer for unsolicited events
dialer := rpc.NewWebsocketDialer(cfg)

// Connect to server (runs in background)
ctx := context.Background()
go dialer.Dial(ctx, "ws://localhost:8080/ws", func(err error) {
    if err != nil {
        log.Error("Connection closed", "error", err)
    }
})

// Wait for connection to establish
for !dialer.IsConnected() {
    time.Sleep(100 * time.Millisecond)
}

// Make RPC calls with timeout
callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

params, _ := rpc.NewParams(map[string]string{"key": "value"})
request := rpc.NewRequest(rpc.NewPayload(1, "get_status", params))
response, err := dialer.Call(callCtx, &request)
if err != nil {
    log.Error("RPC call failed", "error", err)
    return
}

// Process response
var result map[string]interface{}
if err := response.Res.Params.Translate(&result); err != nil {
    log.Error("Invalid response", "error", err)
}

// Handle unsolicited events in the background
go func() {
    for event := range dialer.EventCh() {
        if event == nil {
            // Connection closed
            break
        }
        log.Info("Received event", 
            "method", event.Res.Method,
            "requestID", event.Res.RequestID)
    }
}()
```

### Concurrent RPC Calls

```go
// The dialer supports concurrent calls from multiple goroutines
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        
        params, _ := rpc.NewParams(map[string]int{"id": id})
        request := rpc.NewRequest(rpc.NewPayload(uint64(id), "process", params))
        
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()
        
        resp, err := dialer.Call(ctx, &request)
        if err != nil {
            log.Error("Call failed", "id", id, "error", err)
            return
        }
        
        log.Info("Call succeeded", "id", id, "response", resp.Res.Method)
    }(i)
}

wg.Wait()
```

## Security Considerations

1. **Signature Verification**: Always verify signatures before processing requests
2. **Timestamp Validation**: Implement replay protection using timestamps
3. **Parameter Validation**: Thoroughly validate all parameters before processing
4. **Error Messages**: Use `rpc.Errorf()` for safe client-facing errors
5. **Request IDs**: Use unique request IDs to prevent duplicate processing

## Configuration Options

### WebSocketDialerConfig

The WebSocket dialer can be configured with the following options:

```go
type WebsocketDialerConfig struct {
    // Duration to wait for WebSocket handshake (default: 5s)
    HandshakeTimeout time.Duration
    
    // How often to send ping messages (default: 5s)
    PingInterval time.Duration
    
    // Request ID used for ping messages (default: 100)
    PingRequestID uint64
    
    // Buffer size for event channel (default: 100)
    EventChanSize int
}
```

## Testing

The package includes comprehensive tests for all components. Run tests with:

```bash
go test -race ./pkg/rpc
```

## Dependencies

- `github.com/erc7824/nitrolite/clearnode/pkg/sign`: Signature types and interfaces
- Standard library: `encoding/json`, `errors`, `fmt`, `time`

## See Also

- [Clearnode Protocol Specification](../../docs/Clearnode.protocol.md)
- [API Documentation](../../docs/API.md)
- [Entity Documentation](../../docs/Entities.md)
