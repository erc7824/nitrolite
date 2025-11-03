// Package rpc provides the core data structures and utilities for the Clearnode RPC protocol.
//
// This package implements a secure, signature-based RPC communication protocol designed for
// blockchain and distributed systems. It provides strong typing, efficient encoding, and
// clear separation between client-facing and internal errors.
//
// # Protocol Overview
//
// The protocol uses a request-response pattern with cryptographic signatures:
//
//   - Requests contain a payload and one or more signatures
//   - Responses mirror the request structure with their own signatures
//   - Payloads use a compact array-based JSON encoding for efficiency
//   - All messages include timestamps for replay protection
//
// # Core Types
//
// Request and Response types wrap payloads with signatures:
//
//	type Request struct {
//	    Req Payload          // The request payload
//	    Sig []sign.Signature // One or more signatures
//	}
//
//	type Response struct {
//	    Res Payload          // The response payload
//	    Sig []sign.Signature // One or more signatures
//	}
//
// Payloads contain the actual RPC data:
//
//	type Payload struct {
//	    RequestID uint64 // Unique request identifier
//	    Method    string // RPC method name
//	    Params    Params // Method parameters
//	    Timestamp uint64 // Unix milliseconds timestamp
//	}
//
// # JSON Encoding
//
// Payloads use a compact array encoding for efficiency. A payload like:
//
//	Payload{
//	    RequestID: 12345,
//	    Method: "wallet_transfer",
//	    Params: {"to": "0xabc", "amount": "100"},
//	    Timestamp: 1634567890123,
//	}
//
// Encodes to:
//
//	[12345, "wallet_transfer", {"to": "0xabc", "amount": "100"}, 1634567890123]
//
// This format reduces message size while maintaining readability and compatibility.
//
// # Error Handling
//
// The package provides explicit error types for client communication:
//
//	// Client-facing error - will be sent in response
//	if amount < 0 {
//	    return rpc.Errorf("invalid amount: cannot be negative")
//	}
//
//	// Internal error - generic message sent to client
//	if err := db.Save(); err != nil {
//	    return fmt.Errorf("database error: %w", err)
//	}
//
// # Parameter Handling
//
// The Params type provides flexible parameter handling with type safety:
//
//	// Creating parameters from a struct
//	params, err := rpc.NewParams(struct{
//	    Address string `json:"address"`
//	    Amount  string `json:"amount"`
//	}{
//	    Address: "0x123...",
//	    Amount:  "1000000000000000000",
//	})
//
//	// Extracting parameters into a struct
//	var req TransferRequest
//	err := params.Translate(&req)
//
// # Security Considerations
//
// When using this protocol:
//
//  1. Always verify signatures before processing requests
//  2. Validate timestamps to prevent replay attacks
//  3. Use rpc.Errorf() for safe client-facing errors
//  4. Thoroughly validate all parameters
//  5. Use unique request IDs to prevent duplicate processing
//
// # Client Communication
//
// The package provides two levels of client APIs:
//
// 1. Low-level Dialer interface for direct RPC communication
// 2. High-level Client type with methods for all ClearNode operations
//
// ## High-Level Client (Recommended)
//
// The Client type provides convenient methods for all RPC operations:
//
//	// Create client with WebSocket dialer
//	dialer := rpc.NewWebsocketDialer(rpc.DefaultWebsocketDialerConfig)
//	client := rpc.NewClient(dialer)
//
//	// Set up event handlers
//	client.HandleBalanceUpdateEvent(func(ctx context.Context, notif rpc.BalanceUpdateNotification, sigs []sign.Signature) {
//	    log.Info("Balance updated", "balances", notif.BalanceUpdates)
//	})
//
//	// Connect to server and start event handling
//	err := client.Start(ctx, "wss://server.example.com/ws", func(err error) {
//	    if err != nil {
//	        log.Error("Connection closed", "error", err)
//	    }
//	})
//	if err != nil {
//	    log.Fatal("Failed to start client", "error", err)
//	}
//
//	// Make RPC calls
//	config, _, err := client.GetConfig(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Authenticate
//	authReq := rpc.AuthRequestRequest{
//	    Address:            walletAddress,
//	    SessionKey:         sessionKeyAddress,
//	    Application:        "MyApp",
//	}
//	authResp, _, err := client.AuthWithSig(ctx, authReq, walletSigner)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	jwtToken := authResp.JwtToken
//
//	// Make authenticated calls
//	balances, _, err := client.GetLedgerBalances(ctx, rpc.GetLedgerBalancesRequest{})
//
// ## Low-Level Dialer
//
// For direct RPC communication without the convenience methods:
//
//	params, _ := rpc.NewParams(map[string]string{"key": "value"})
//	payload := rpc.NewPayload(1, "method_name", params)
//	request := rpc.NewRequest(payload)
//
//	response, err := dialer.Call(ctx, &request)
//	if err != nil {
//	    log.Error("RPC call failed", "error", err)
//	}
//
//	// Handle events manually
//	go func() {
//	    for event := range dialer.EventCh() {
//	        if event == nil {
//	            break
//	        }
//	        log.Info("Received event", "method", event.Res.Method)
//	    }
//	}()
//
// # API Types
//
// The package includes comprehensive type definitions for the ClearNode RPC API:
//
// - Request/Response types for all RPC methods
// - Asset and network configuration types
// - Payment channel state and operations
// - Application session management
// - Ledger and transaction types
// - Event notification types
//
// All monetary values use decimal.Decimal for arbitrary precision arithmetic.
//
// # Server Implementation
//
// The package provides a complete RPC server implementation through the Node interface:
//
//	// Create and configure the server
//	config := rpc.WebsocketNodeConfig{
//	    Signer: signer,
//	    Logger: logger,
//	    OnConnectHandler: func(send rpc.SendResponseFunc) {
//	        // Handle new connections
//	    },
//	    OnAuthenticatedHandler: func(userID string, send rpc.SendResponseFunc) {
//	        // Handle authentication
//	    },
//	}
//
//	node, err := rpc.NewWebsocketNode(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Register handlers
//	node.Handle("get_balance", handleGetBalance)
//	node.Handle("transfer", handleTransfer)
//
//	// Add middleware
//	node.Use(loggingMiddleware)
//	node.Use(authMiddleware)
//
//	// Create handler groups
//	privateGroup := node.NewGroup("private")
//	privateGroup.Use(requireAuthMiddleware)
//	privateGroup.Handle("create_channel", handleCreateChannel)
//
//	// Start the server
//	http.Handle("/ws", node)
//	http.ListenAndServe(":8080", nil)
//
// Writing handlers:
//
//	func handleGetBalance(c *rpc.Context) {
//	    // Extract parameters
//	    var req GetBalanceRequest
//	    if err := c.Request.Req.Params.Translate(&req); err != nil {
//	        c.Fail(nil, "invalid parameters")
//	        return
//	    }
//
//	    // Process request
//	    balance := getBalanceForUser(c.UserID, req.Asset)
//
//	    // Send response
//	    c.Succeed("get_balance", rpc.Params{"balance": balance})
//	}
//
// Writing middleware:
//
//	func authMiddleware(c *rpc.Context) {
//	    // Check if connection is authenticated
//	    if c.UserID == "" {
//	        // Try to authenticate from request
//	        token := extractToken(c.Request)
//	        userID, err := validateToken(token)
//	        if err != nil {
//	            c.Fail(nil, "authentication required")
//	            return
//	        }
//	        c.UserID = userID
//	    }
//
//	    // Continue to next handler
//	    c.Next()
//	}
//
// # Example Usage
//
// Creating and sending a request:
//
//	// Create request
//	params, _ := rpc.NewParams(map[string]string{"key": "value"})
//	payload := rpc.NewPayload(12345, "method_name", params)
//	request := rpc.NewRequest(payload, signature)
//
//	// Marshal and send
//	data, _ := json.Marshal(request)
//	// ... send data over transport ...
//
// Processing a request:
//
//	// Unmarshal request
//	var request rpc.Request
//	err := json.Unmarshal(data, &request)
//
//	// Verify signatures using GetSigners
//	signers, err := request.GetSigners()
//	if err != nil {
//	    return rpc.Errorf("invalid signatures: %v", err)
//	}
//
//	// Check if request is from a known address
//	authorized := false
//	for _, signer := range signers {
//	    if signer == trustedAddress {
//	        authorized = true
//	        break
//	    }
//	}
//	if !authorized {
//	    return rpc.Errorf("unauthorized request")
//	}
//
//	// Process based on method
//	switch request.Req.Method {
//	case "transfer":
//	    var params TransferParams
//	    if err := request.Req.Params.Translate(&params); err != nil {
//	        return rpc.Errorf("invalid parameters: %v", err)
//	    }
//	    // ... handle transfer ...
//	}
//
// # Testing
//
// The package includes a comprehensive test suite with mock implementations:
//
// - client_test.go: Unit tests for all client methods
// - client_internal_test.go: Tests for internal authentication methods
// - client_manual_test.go: Integration tests against live server (requires credentials)
//
// The manual test demonstrates real-world usage patterns and can be run with:
//
//	TEST_WALLET_PK=<wallet_private_key> TEST_SESSION_PK=<session_private_key> go test -run TestManualClient
package rpc
