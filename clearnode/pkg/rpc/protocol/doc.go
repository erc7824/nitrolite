// Package protocol provides the core data structures and utilities for the Clearnode RPC protocol.
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
//	    return protocol.Errorf("invalid amount: cannot be negative")
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
//	params, err := protocol.NewParams(struct{
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
//  3. Use protocol.Errorf() for safe client-facing errors
//  4. Thoroughly validate all parameters
//  5. Use unique request IDs to prevent duplicate processing
//
// # Example Usage
//
// Creating and sending a request:
//
//	// Create request
//	params, _ := protocol.NewParams(map[string]string{"key": "value"})
//	payload := protocol.NewPayload(12345, "method_name", params)
//	request := protocol.NewRequest(payload, signature)
//
//	// Marshal and send
//	data, _ := json.Marshal(request)
//	// ... send data over transport ...
//
// Processing a request:
//
//	// Unmarshal request
//	var request protocol.Request
//	err := json.Unmarshal(data, &request)
//
//	// Verify signatures using GetSigners
//	signers, err := request.GetSigners()
//	if err != nil {
//	    return protocol.Errorf("invalid signatures: %v", err)
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
//	    return protocol.Errorf("unauthorized request")
//	}
//
//	// Process based on method
//	switch request.Req.Method {
//	case "wallet_transfer":
//	    var params TransferParams
//	    if err := request.Req.Params.Translate(&params); err != nil {
//	        return protocol.Errorf("invalid parameters: %v", err)
//	    }
//	    // ... handle transfer ...
//	}
package protocol
