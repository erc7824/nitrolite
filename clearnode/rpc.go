package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// RPCMessage represents a complete message in the RPC protocol, including data and signatures
type RPCMessage struct {
	Req          *RPCData    `json:"req,omitempty" validate:"required_without=Res,excluded_with=Res"`
	Res          *RPCData    `json:"res,omitempty" validate:"required_without=Req,excluded_with=Req"`
	AppSessionID string      `json:"sid,omitempty"`
	Sig          []Signature `json:"sig"`
}

// ParseRPCMessage parses a JSON string into an RPCMessage
func ParseRPCMessage(data []byte) (RPCMessage, error) {
	var req RPCMessage
	if err := json.Unmarshal(data, &req); err != nil {
		return RPCMessage{}, fmt.Errorf("failed to parse request: %w", err)
	}
	return req, nil
}

// GetRequestSignersMap returns map with request signers public adresses
func (r RPCMessage) GetRequestSignersMap() (map[string]struct{}, error) {
	recoveredAddresses := make(map[string]struct{}, len(r.Sig))
	for _, sigHex := range r.Sig {
		recovered, err := RecoverAddress(r.Req.rawBytes, sigHex)
		if err != nil {
			return nil, err
		}
		recoveredAddresses[recovered] = struct{}{}
	}

	return recoveredAddresses, nil
}

// TODO: ensure that it accepts only structs or maps, and prevent passing primitive (and other DS) types
type RPCDataParams = any

// RPCData represents the common structure for both requests and responses
// Format: [request_id, method, params, ts]
type RPCData struct {
	RequestID uint64        `json:"request_id" validate:"required"`
	Method    string        `json:"method" validate:"required"`
	Params    RPCDataParams `json:"params" validate:"required"`
	Timestamp uint64        `json:"ts" validate:"required"`
	rawBytes  []byte
}

// UnmarshalJSON implements the json.Unmarshaler interface for RPCMessage
func (m *RPCData) UnmarshalJSON(data []byte) error {
	var rawArr []json.RawMessage
	if err := json.Unmarshal(data, &rawArr); err != nil {
		return fmt.Errorf("error reading RPCData as array: %w", err)
	}
	if len(rawArr) != 4 {
		return errors.New("invalid RPCData: expected 4 elements in array")
	}

	// Element 0: uint64 RequestID
	if err := json.Unmarshal(rawArr[0], &m.RequestID); err != nil {
		return fmt.Errorf("invalid request_id: %w", err)
	}
	// Element 1: string Method
	if err := json.Unmarshal(rawArr[1], &m.Method); err != nil {
		return fmt.Errorf("invalid method: %w", err)
	}
	// Element 2: RPCDataParams Params
	if err := json.Unmarshal(rawArr[2], &m.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	// Element 3: uint64 Timestamp
	if err := json.Unmarshal(rawArr[3], &m.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	// Store raw bytes for signature verification
	m.rawBytes = data

	return nil
}

// MarshalJSON for RPCData always emits the array‐form [RequestID, Method, Params, Timestamp].
func (m RPCData) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		m.RequestID,
		m.Method,
		m.Params,
		m.Timestamp,
	})
}

// CreateResponse is unchanged. It simply constructs an RPCMessage with a "res" array.
func CreateResponse(id uint64, method string, responseParams RPCDataParams) *RPCMessage {
	return &RPCMessage{
		Res: &RPCData{
			RequestID: id,
			Method:    method,
			Params:    responseParams,
			Timestamp: uint64(time.Now().UnixMilli()),
		},
		Sig: []Signature{},
	}
}

// RPCError represents an error in the RPC protocol that should be sent back to the client
// in the RPC response. Unlike generic errors, RPCError messages are guaranteed to be
// included in the error response sent to the client.
//
// Use RPCError when you want to provide specific, user-facing error messages in RPC responses.
// For internal errors that should not be exposed to clients, use regular errors instead.
//
// Example:
//
//	// Client will receive this exact error message
//	return RPCErrorf("invalid wallet address: %s", addr)
//
//	// Client will receive a generic error message
//	return fmt.Errorf("database connection failed")
type RPCError struct {
	err error
}

// RPCErrorf creates a new RPCError with a formatted error message that will be sent
// to the client in the RPC response. This is the preferred way to create client-facing
// errors in RPC handlers.
//
// The error message should be clear, actionable, and safe to expose to external clients.
// Avoid including sensitive information like internal system details, file paths, or
// database specifics.
//
// Usage in RPC handlers:
//
//	// In a handler function that returns an error
//	if amount.IsNegative() {
//		return RPCErrorf("invalid amount: cannot be negative")
//	}
//
//	// With formatting for specific details
//	if balance.LessThan(amount) {
//		return RPCErrorf("insufficient balance: need %s but have %s", amount, balance)
//	}
func RPCErrorf(format string, args ...any) RPCError {
	return RPCError{
		err: fmt.Errorf(format, args...),
	}
}

// Error implements the error interface for RPCError
func (e RPCError) Error() string {
	return e.err.Error()
}
