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
// The package includes a Dialer interface and WebSocket implementation for client-side RPC.
// For most use cases, the high-level Client type provides a more convenient interface:
//
//	// Create and configure a dialer
//	cfg := rpc.DefaultWebsocketDialerConfig
//	cfg.EventChanSize = 100  // Buffer for unsolicited events
//	dialer := rpc.NewWebsocketDialer(cfg)
//
//	// Create a client
//	client := rpc.NewClient(dialer)
//
//	// Register event handlers before connecting
//	client.HandleBalanceUpdateEvent(func(ctx context.Context, notif rpc.BalanceUpdateNotification, sigs []sign.Signature) {
//	    log.Info("Balance updated", "balances", notif.BalanceUpdates)
//	})
//
//	client.HandleChannelUpdateEvent(func(ctx context.Context, notif rpc.ChannelUpdateNotification, sigs []sign.Signature) {
//	    log.Info("Channel updated", "channelID", notif.ChannelID, "status", notif.Status)
//	})
//
//	// Connect to server (in a goroutine as it blocks)
//	go dialer.Dial(ctx, "ws://localhost:8080/ws", func(err error) {
//	    if err != nil {
//	        log.Error("Connection closed", "error", err)
//	    }
//	})
//
//	// Start listening for events
//	go client.ListenEvents(ctx, func(err error) {
//	    log.Info("Event listener stopped", "error", err)
//	})
//
//	// Wait for connection
//	for !dialer.IsConnected() {
//	    time.Sleep(100 * time.Millisecond)
//	}
//
//	// Use type-safe client methods
//	config, sigs, err := client.GetConfig(ctx)
//	if err != nil {
//	    log.Error("Failed to get config", "error", err)
//	}
//
//	// Make transfers
//	transferResp, sigs, err := client.Transfer(ctx, rpc.TransferRequest{
//	    From: myAddress,
//	    To: recipientAddress,
//	    Amount: "1000000000000000000",
//	    Asset: "ETH",
//	})
//
// For lower-level control, you can use the dialer directly:
//
//	// Send RPC requests manually
//	params, _ := rpc.NewParams(map[string]string{"key": "value"})
//	payload := rpc.NewPayload(1, "method_name", params)
//	request := rpc.NewRequest(payload)
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	response, err := dialer.Call(ctx, &request)
//	if err != nil {
//	    log.Error("RPC call failed", "error", err)
//	}
//
//	// Handle unsolicited events manually
//	go func() {
//	    for event := range dialer.EventCh() {
//	        if event == nil {
//	            // Connection closed
//	            break
//	        }
//	        log.Info("Received event", "method", event.Res.Method)
//	    }
//	}()
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
package rpc
