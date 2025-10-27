# RPC Package

The `pkg/rpc` package provides the core data structures and utilities for the Clearnode RPC protocol. This package implements a secure, signature-based RPC communication protocol suitable for blockchain and distributed systems.

## Overview

### Client Features
- **High-Level Client**: Type-safe methods for all RPC operations
- **WebSocket Transport**: Persistent connection with automatic reconnection
- **Event Handling**: Asynchronous notifications for balance updates, transfers, and channel changes
- **Authentication**: Wallet-based authentication with JWT token support
- **Channel Management**: Create, resize, and close payment channels
- **Application Sessions**: Multi-party state channel applications
- **Off-chain Transfers**: Instant, gas-free transfers within ClearNode

### Server Features
- **RPC Server**: Complete server implementation with WebSocket transport
- **Handler Registration**: Simple method-based request routing
- **Middleware Support**: Composable request processing pipeline
- **Handler Groups**: Organize endpoints with shared middleware
- **Connection Management**: Automatic connection lifecycle handling
- **Authentication**: Built-in support for user authentication and re-authentication
- **Server Push**: Send notifications to specific users or connections

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

### High-Level Client API

The `Client` type provides convenient methods for all server operations:

#### Public Methods (No Authentication Required)
- `Ping()` - Check server connectivity
- `GetConfig()` - Get server configuration and supported networks
- `GetAssets()` - List supported tokens/assets
- `GetChannels()` - Query payment channels
- `GetAppSessions()` - List application sessions
- `GetLedgerEntries()` - View ledger entries
- `GetLedgerTransactions()` - View transaction history

#### Authentication
- `AuthWithSig()` - Authenticate using wallet signature
- `AuthJWTVerify()` - Verify existing JWT token

#### Authenticated Methods
- `GetUserTag()` - Get user's human-readable tag
- `GetLedgerBalances()` - View account balances
- `GetRPCHistory()` - View RPC call history
- `CreateChannel()` - Request new payment channel
- `ResizeChannel()` - Modify channel funding
- `CloseChannel()` - Close payment channel
- `Transfer()` - Transfer funds between ClearNode accounts
- `CreateAppSession()` - Start multi-party application
- `SubmitAppState()` - Update application state
- `CloseAppSession()` - Close application session

### Event Handling

The client supports real-time event notifications:

```go
// Register event handlers
client.HandleBalanceUpdateEvent(func(ctx context.Context, notif BalanceUpdateNotification, sigs []sign.Signature) {
    // Handle balance changes
})

client.HandleChannelUpdateEvent(func(ctx context.Context, notif ChannelUpdateNotification, sigs []sign.Signature) {
    // Handle channel state changes
})

client.HandleTransferEvent(func(ctx context.Context, notif TransferNotification, sigs []sign.Signature) {
    // Handle incoming/outgoing transfers
})

// Event listening starts automatically when you call Start()
```

### Transport Layer

The package includes a `Dialer` interface with a WebSocket implementation:
- **Thread-safe**: Supports concurrent RPC calls
- **Automatic reconnection**: Built-in ping/pong mechanism
- **Event handling**: Separate channel for unsolicited server events
- **Context support**: Full context cancellation and timeout support

## Installation

```go
import "github.com/erc7824/nitrolite/clearnode/pkg/rpc"
```

## Server Usage

### Creating an RPC Server

```go
import (
    "github.com/erc7824/nitrolite/clearnode/pkg/rpc"
    "github.com/erc7824/nitrolite/clearnode/pkg/sign"
    "github.com/erc7824/nitrolite/clearnode/pkg/log"
)

// Create server configuration
config := rpc.WebsocketNodeConfig{
    Signer: serverSigner,  // Required: signs all responses
    Logger: logger,        // Required: for structured logging
    
    // Connection lifecycle callbacks
    OnConnectHandler: func(send rpc.SendResponseFunc) {
        log.Info("New connection established")
        // Optionally send welcome message
        params, _ := rpc.NewParams(map[string]interface{}{
            "message": "Connected to RPC server",
        })
        send("welcome", params)
    },
    OnDisconnectHandler: func(userID string) {
        log.Info("Connection closed", "userID", userID)
    },
    OnAuthenticatedHandler: func(userID string, send rpc.SendResponseFunc) {
        log.Info("User authenticated", "userID", userID)
        // Send personalized greeting
        params, _ := rpc.NewParams(map[string]interface{}{"userID": userID})
        send("authenticated", params)
    },
}

// Create the RPC node
node, err := rpc.NewWebsocketNode(config)
if err != nil {
    log.Fatal("Failed to create node", "error", err)
}

// Register handlers
node.Handle("ping", handlePing)           // Built-in, responds with "pong"
node.Handle("get_config", handleGetConfig)
node.Handle("auth_request", handleAuthRequest)

// Add global middleware
node.Use(loggingMiddleware)
node.Use(rateLimitMiddleware)

// Create groups for organized endpoints
publicGroup := node.NewGroup("public")
publicGroup.Handle("get_assets", handleGetAssets)
publicGroup.Handle("get_channels", handleGetChannels)

privateGroup := node.NewGroup("private")
privateGroup.Use(authRequiredMiddleware)  // Only for this group
privateGroup.Handle("get_balance", handleGetBalance)
privateGroup.Handle("transfer", handleTransfer)

// Start the server
http.Handle("/ws", node)
log.Fatal(http.ListenAndServe(":8080", nil))
```

### Writing Handlers

Handlers process RPC requests and generate responses:

```go
func handleGetBalance(c *rpc.Context) {
    // Check authentication (if not done by middleware)
    if c.UserID == "" {
        c.Fail(nil, "authentication required")
        return
    }
    
    // Extract and validate parameters
    var req GetBalanceRequest
    if err := c.Request.Req.Params.Translate(&req); err != nil {
        c.Fail(nil, "invalid parameters")
        return
    }
    
    // Access connection storage
    if lastCheck, ok := c.Storage.Get("last_balance_check"); ok {
        log.Debug("Last balance check", "time", lastCheck)
    }
    c.Storage.Set("last_balance_check", time.Now())
    
    // Process the request
    balance, err := ledger.GetBalance(c.UserID, req.Asset)
    if err != nil {
        // Internal error - generic message sent to client
        log.Error("Failed to get balance", "error", err)
        c.Fail(err, "failed to retrieve balance")
        return
    }
    
    // Send successful response
    params, _ := rpc.NewParams(map[string]interface{}{
        "asset": req.Asset,
        "balance": balance.String(),
    })
    c.Succeed("get_balance", params)
}
```

### Writing Middleware

Middleware can process requests before/after handlers:

```go
func loggingMiddleware(c *rpc.Context) {
    start := time.Now()
    method := c.Request.Req.Method
    
    // Pre-processing
    log.Info("Request started", 
        "method", method,
        "userID", c.UserID,
        "requestID", c.Request.Req.RequestID)
    
    // Continue to next handler
    c.Next()
    
    // Post-processing
    duration := time.Since(start)
    log.Info("Request completed",
        "method", method,
        "requestID", c.Request.Req.RequestID,
        "duration", duration)
}

func authRequiredMiddleware(c *rpc.Context) {
    // Check if already authenticated
    if c.UserID != "" {
        c.Next()
        return
    }
    
    // Try to authenticate from request
    var auth AuthData
    if err := c.Request.Req.Params.Translate(&auth); err != nil {
        c.Fail(nil, "authentication required")
        return
    }
    
    userID, err := authenticateUser(auth.Token)
    if err != nil {
        c.Fail(nil, "invalid authentication")
        return
    }
    
    // Set the user ID for subsequent handlers
    c.UserID = userID
    c.Next()
}

func rateLimitMiddleware(c *rpc.Context) {
    key := fmt.Sprintf("rate_limit_%s", c.UserID)
    
    // Get current count
    count := 0
    if val, ok := c.Storage.Get(key); ok {
        count = val.(int)
    }
    
    if count >= 100 {
        c.Fail(nil, "rate limit exceeded")
        return
    }
    
    // Increment and continue
    c.Storage.Set(key, count+1)
    c.Next()
}
```

### Server-Initiated Notifications

Send notifications to specific users:

```go
// Send notification to a specific user
params, _ := rpc.NewParams(map[string]interface{}{
    "asset": "ETH",
    "old_balance": "100",
    "new_balance": "150",
    "reason": "deposit",
})
node.Notify(userID, "balance_update", params)
```

## Client Usage

### Quick Start

```go
import "github.com/erc7824/nitrolite/clearnode/pkg/rpc"

// Create client
dialer := rpc.NewWebsocketDialer(rpc.DefaultWebsocketDialerConfig)
client := rpc.NewClient(dialer)

// Set up event handlers
client.HandleBalanceUpdateEvent(handleBalanceUpdate)

// Example server URL (use your actual server)
serverRpcWsUrl := "wss://clearnet-sandbox.yellow.com/ws"

// Connect to server and start listening for events
ctx := context.Background()
err := client.Start(ctx, serverRpcWsUrl, func(err error) {
    if err != nil {
        log.Error("Connection closed", "error", err)
    }
})
if err != nil {
    log.Fatal("Failed to start client", "error", err)
}

// Get server configuration
config, _, err := client.GetConfig(ctx)
if err != nil {
    log.Fatal("Failed to get config", "error", err)
}

// Authenticate
walletSigner, _ := sign.NewEthereumSigner(walletPrivateKey)
sessionSigner, _ := sign.NewEthereumSigner(sessionPrivateKey)

authReq := rpc.AuthRequestRequest{
    Address:            walletSigner.PublicKey().Address().String(),
    SessionKey:         sessionSigner.PublicKey().Address().String(), // Different from Address
    Application:            "MyApp",
}

authResp, _, err := client.AuthWithSig(ctx, authReq, walletSigner)
if err != nil {
    log.Fatal("Authentication failed", "error", err)
}
jwtToken := authResp.JwtToken // Store for future use

// Make authenticated calls
balances, _, err := client.GetLedgerBalances(ctx, rpc.GetLedgerBalancesRequest{})
```

### Off-chain Transfers

```go
// Transfer funds between ClearNode accounts (no blockchain interaction)
transferReq := rpc.TransferRequest{
    Destination: recipientAddress,
    Allocations: []rpc.TransferAllocation{
        {AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
    },
}

response, _, err := client.Transfer(ctx, transferReq)
if err != nil {
    log.Fatal("Transfer failed", "error", err)
}
```

### Application Sessions

```go
// Create multi-party application session
createSessReq := rpc.CreateAppSessionRequest{
    Definition: rpc.AppDefinition{
        Protocol:           rpc.VersionNitroRPCv0_4, // Required: "NitroRPC/0.4"
        ParticipantWallets: []string{player1, player2},
        Weights:            []int64{1, 1},
        Quorum:             2,
        Challenge:          3600,
        Nonce:              uint64(uuid.New().ID()),
    },
    Allocations: []rpc.AppAllocation{
        {ParticipantWallet: player1, AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
        {ParticipantWallet: player2, AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
    },
}

payload, _ := client.PreparePayload(rpc.CreateAppSessionMethod, createSessReq)
hash, _ := payload.Hash()

// Both participants must sign
sig1, _ := player1Signer.Sign(hash)
sig2, _ := player2Signer.Sign(hash)
fullReq := rpc.NewRequest(payload, sig1, sig2)

response, _, err := client.CreateAppSession(ctx, &fullReq)
```

### Channel Operations

```go
// Create a payment channel
amount := decimal.NewFromInt(1000000)
createReq := rpc.CreateChannelRequest{
    ChainID:    1,
    Token:      "0xUSDC",
    Amount:     &amount,
    SessionKey: &sessionKeyAddress, // Required
}

// Prepare and sign request
payload, _ := client.PreparePayload(rpc.CreateChannelMethod, createReq)
hash, _ := payload.Hash()
sig, _ := sessionSigner.Sign(hash)
fullReq := rpc.NewRequest(payload, sig)

// Server returns signed channel state
response, _, err := client.CreateChannel(ctx, &fullReq)
if err != nil {
    log.Fatal("Failed to create channel", "error", err)
}

// You must sign the state and submit to blockchain yourself
stateHash := computeStateHash(response.State)
mySignature, _ := sessionSigner.Sign(stateHash)
// Submit response.StateSignature and mySignature to blockchain
```

## Low-Level Usage

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

### WebSocket Client Usage

```go
// Create and configure a WebSocket dialer
cfg := rpc.DefaultWebsocketDialerConfig
cfg.EventChanSize = 100  // Buffer for unsolicited events
dialer := rpc.NewWebsocketDialer(cfg)

// Create client for high-level API
client := rpc.NewClient(dialer)

// Connect to server and start event handling
ctx := context.Background()
err := client.Start(ctx, "ws://localhost:8080/ws", func(err error) {
    if err != nil {
        log.Error("Connection closed", "error", err)
    }
})
if err != nil {
    log.Fatal("Failed to start client", "error", err)
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

The package includes comprehensive tests:

### Unit Tests
```bash
go test -race ./pkg/rpc
```

### Integration Tests
The `client_manual_test.go` file contains integration tests that connect to a real ClearNode server. These tests demonstrate real-world usage patterns and verify the client works correctly with the current server deployment.

```bash
# Run manual tests (requires test credentials)
TEST_WALLET_PK=<wallet_private_key> TEST_SESSION_PK=<session_private_key> \
    go test -run TestManualClient ./pkg/rpc
```

The manual test serves two purposes:
1. Verifies the client implementation works with the live server
2. Provides working examples of all major client operations

## Dependencies

- `github.com/erc7824/nitrolite/clearnode/pkg/sign`: Signature types and interfaces
- Standard library: `encoding/json`, `errors`, `fmt`, `time`

## See Also

- [Clearnode Protocol Specification](../../docs/Clearnode.protocol.md)
- [API Documentation](../../docs/API.md)
- [Entity Documentation](../../docs/Entities.md)
