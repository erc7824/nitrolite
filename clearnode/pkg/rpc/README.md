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

The package provides two levels of client functionality:

#### High-Level Client

The `Client` type provides a type-safe, convenient interface for all RPC operations:
- **Type-safe methods**: Dedicated methods for each RPC operation
- **Event handling**: Simple registration of event handlers
- **Automatic serialization**: Handles request/response marshaling
- **Thread-safe**: Safe for concurrent use from multiple goroutines

#### Low-Level Dialer

The `Dialer` interface and WebSocket implementation provide lower-level communication:
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

## Using the RPC Client

The `Client` type provides a high-level interface for interacting with Clearnode RPC servers. It handles all the low-level details of request construction, signature management, and response parsing.

### Client Setup

```go
// 1. Create a dialer configuration
cfg := rpc.DefaultWebsocketDialerConfig
cfg.EventChanSize = 100  // Adjust based on expected event volume

// 2. Create the dialer
dialer := rpc.NewWebsocketDialer(cfg)

// 3. Create the client
client := rpc.NewClient(dialer)

// 4. Register event handlers (before connecting)
client.HandleBalanceUpdateEvent(func(ctx context.Context, notif rpc.BalanceUpdateNotification, sigs []sign.Signature) {
    for _, update := range notif.BalanceUpdates {
        log.Info("Balance changed", 
            "account", update.Account,
            "asset", update.Asset,
            "amount", update.Amount)
    }
})

client.HandleChannelUpdateEvent(func(ctx context.Context, notif rpc.ChannelUpdateNotification, sigs []sign.Signature) {
    log.Info("Channel updated",
        "channelID", notif.ChannelID,
        "status", notif.Status)
})

client.HandleTransferEvent(func(ctx context.Context, notif rpc.TransferNotification, sigs []sign.Signature) {
    for _, tx := range notif.Transactions {
        log.Info("Transfer",
            "from", tx.FromAccount,
            "to", tx.ToAccount,
            "amount", tx.Amount)
    }
})

// 5. Connect to the server
ctx := context.Background()
go dialer.Dial(ctx, "ws://localhost:8080/ws", func(err error) {
    if err != nil {
        log.Error("Connection closed", "error", err)
    }
})

// 6. Start event listener
go client.ListenEvents(ctx, func(err error) {
    log.Info("Event listener stopped", "error", err)
})

// 7. Wait for connection to be established
for !dialer.IsConnected() {
    time.Sleep(100 * time.Millisecond)
}
```

### Making RPC Calls

#### Query Operations

```go
// Get server configuration
config, sigs, err := client.GetConfig(ctx)
if err != nil {
    return fmt.Errorf("failed to get config: %w", err)
}

// Get asset information
assetsResp, sigs, err := client.GetAssets(ctx, rpc.GetAssetsRequest{
    AssetIDs: []string{"ETH", "USDC"},
})

// Get account balances
balances, sigs, err := client.GetLedgerBalances(ctx, rpc.GetLedgerBalancesRequest{
    Account: myAddress,
    Assets: []string{"ETH", "USDC"},
})

// Get channels
channels, sigs, err := client.GetChannels(ctx, rpc.GetChannelsRequest{
    Participant: myAddress,
    Status: "open",
})
```

#### Authentication

```go
// Request authentication challenge
authReq := rpc.AuthRequestRequest{
    Address: myAddress,
    Scope: []string{"read", "write"},
}
challenge, sigs, err := client.AuthRequest(ctx, authReq)
if err != nil {
    return fmt.Errorf("auth request failed: %w", err)
}

// Sign the challenge (using your signing library)
signature := signChallenge(privateKey, challenge.Challenge)

// Verify the signature
verifyReq := rpc.AuthSigVerifyRequest{
    Challenge: challenge.Challenge,
    Address: myAddress,
}
authResp, sigs, err := client.AuthSigVerify(ctx, verifyReq, signature)
if err != nil {
    return fmt.Errorf("auth verify failed: %w", err)
}

// Use the auth token for subsequent requests
log.Info("Authenticated", "token", authResp.Token)
```

#### Transfers

```go
// Simple transfer
transferReq := rpc.TransferRequest{
    From: myAddress,
    To: recipientAddress,
    Amount: "1000000000000000000", // 1 ETH
    Asset: "ETH",
    Memo: "Payment for services",
}

txResp, sigs, err := client.Transfer(ctx, transferReq)
if err != nil {
    return fmt.Errorf("transfer failed: %w", err)
}

log.Info("Transfer completed", 
    "txID", txResp.TransactionID,
    "status", txResp.Status)
```

#### Channel Operations

```go
// Create a channel
createReq := rpc.CreateChannelRequest{
    Participants: []string{myAddress, counterpartyAddress},
    InitialBalances: map[string]string{
        myAddress: "5000000000000000000", // 5 ETH
        counterpartyAddress: "0",
    },
    Asset: "ETH",
}

// Sign the channel creation
signature := signChannelCreation(privateKey, createReq)

channelResp, sigs, err := client.CreateChannel(ctx, createReq, signature)
if err != nil {
    return fmt.Errorf("channel creation failed: %w", err)
}

log.Info("Channel created", "channelID", channelResp.ChannelID)

// Later, close the channel
closeReq := rpc.CloseChannelRequest{
    ChannelID: channelResp.ChannelID,
    FinalBalances: map[string]string{
        myAddress: "3000000000000000000",
        counterpartyAddress: "2000000000000000000",
    },
    Cooperative: true,
}

// Sign the close request
closeSig := signChannelClose(privateKey, closeReq)

closeResp, sigs, err := client.CloseChannel(ctx, closeReq, closeSig)
if err != nil {
    return fmt.Errorf("channel close failed: %w", err)
}
```

#### Application Sessions

```go
// Create an app session (e.g., for a game)
createSessionReq := rpc.CreateAppSessionRequest{
    AppID: gameAppID,
    Participants: []string{player1, player2, player3},
    InitialState: initialGameState,
    Timeout: 3600, // 1 hour
}

// Collect signatures from all participants
sigs := []sign.Signature{player1Sig, player2Sig, player3Sig}

session, serverSigs, err := client.CreateAppSession(ctx, createSessionReq, sigs)
if err != nil {
    return fmt.Errorf("session creation failed: %w", err)
}

// Submit state updates during the game
stateUpdateReq := rpc.SubmitAppStateRequest{
    SessionID: session.SessionID,
    NewState: updatedGameState,
    SequenceNumber: 42,
}

// All players must sign state updates
updateSigs := []sign.Signature{player1Sig, player2Sig, player3Sig}

stateResp, serverSigs, err := client.SubmitAppState(ctx, stateUpdateReq, updateSigs)
if err != nil {
    return fmt.Errorf("state update failed: %w", err)
}

// Close the session when done
closeSessionReq := rpc.CloseAppSessionParams{
    SessionID: session.SessionID,
    FinalState: finalGameState,
    Outcome: map[string]string{
        player1: "2000000000000000000", // Player 1 wins 2 ETH
        player2: "500000000000000000",   // Player 2 gets 0.5 ETH
        player3: "500000000000000000",   // Player 3 gets 0.5 ETH
    },
}

closeSigs := []sign.Signature{player1Sig, player2Sig, player3Sig}
closeResp, serverSigs, err := client.CloseAppSession(ctx, closeSessionReq, closeSigs)
```

### Error Handling

The client returns errors in several cases:

```go
resp, sigs, err := client.Transfer(ctx, transferReq)
if err != nil {
    // Check if it's an RPC protocol error
    var rpcErr *rpc.Error
    if errors.As(err, &rpcErr) {
        // This is a client-facing error from the server
        log.Error("RPC error", "message", rpcErr.Error())
        return err
    }
    
    // Network or other transport error
    log.Error("Transport error", "error", err)
    return err
}
```

### Best Practices

1. **Always register event handlers before connecting** - Events may arrive immediately after connection
2. **Use contexts with timeouts** - Prevent hanging on network issues
3. **Handle reconnection** - The dialer handles reconnection, but you may need to re-authenticate
4. **Verify signatures** - Always verify server signatures on sensitive operations
5. **Keep event handlers lightweight** - Don't block the event processing loop

### Complete Example

```go
func runClient(ctx context.Context) error {
    // Setup
    cfg := rpc.DefaultWebsocketDialerConfig
    dialer := rpc.NewWebsocketDialer(cfg)
    client := rpc.NewClient(dialer)
    
    // Event handlers
    client.HandleBalanceUpdateEvent(handleBalanceUpdate)
    client.HandleChannelUpdateEvent(handleChannelUpdate)
    client.HandleTransferEvent(handleTransfer)
    
    // Connect
    errCh := make(chan error, 2)
    go func() {
        dialer.Dial(ctx, "ws://localhost:8080/ws", func(err error) {
            errCh <- fmt.Errorf("connection closed: %w", err)
        })
    }()
    
    go func() {
        client.ListenEvents(ctx, func(err error) {
            errCh <- fmt.Errorf("event listener stopped: %w", err)
        })
    }()
    
    // Wait for connection
    for !dialer.IsConnected() {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case err := <-errCh:
            return err
        case <-time.After(100 * time.Millisecond):
            // Continue waiting
        }
    }
    
    // Authenticate
    if err := authenticate(ctx, client); err != nil {
        return fmt.Errorf("authentication failed: %w", err)
    }
    
    // Run main logic
    if err := runMainLogic(ctx, client); err != nil {
        return fmt.Errorf("main logic failed: %w", err)
    }
    
    // Wait for shutdown
    select {
    case <-ctx.Done():
        return ctx.Err()
    case err := <-errCh:
        return err
    }
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
