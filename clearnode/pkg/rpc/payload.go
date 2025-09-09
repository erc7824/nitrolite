package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

// Payload represents the core data structure for RPC communication.
// It contains all the information needed to process an RPC call or response.
//
// Payloads are encoded as JSON arrays for compact transmission:
// [RequestID, Method, Params, Timestamp]
//
// This encoding reduces message size while maintaining human readability
// and allows for efficient parsing. The array format is automatically
// handled by the custom JSON marshaling methods.
type Payload struct {
	// RequestID is a unique identifier for tracking requests and matching responses.
	// Clients should generate unique IDs to prevent collisions and enable proper
	// request-response correlation.
	RequestID uint64 `json:"request_id"`

	// Method specifies the RPC method to be invoked (e.g., "wallet_transfer").
	// Method names should follow a consistent naming convention, typically
	// using lowercase with underscores (e.g., "module_action").
	Method string `json:"method"`

	// Params contains the method-specific parameters as a flexible map.
	// This allows different methods to have different parameter structures
	// while maintaining type safety through the Translate method.
	Params Params `json:"params"`

	// Timestamp is the Unix timestamp in milliseconds when the payload was created.
	// This is used for replay protection and request expiration checks.
	// Servers should validate that timestamps are within an acceptable time window.
	Timestamp uint64 `json:"ts"`
}

// NewPayload creates a new Payload with the given request ID, method, and parameters.
// The timestamp is automatically set to the current time in Unix milliseconds.
//
// Example:
//
//	params, _ := NewParams(map[string]string{"address": "0x123"})
//	payload := NewPayload(12345, "wallet_getBalance", params)
//
// The resulting payload will have the current timestamp and can be used
// in either a Request or Response.
func NewPayload(id uint64, method string, params Params) Payload {
	if params == nil {
		params = Params{}
	}

	return Payload{
		RequestID: id,
		Method:    method,
		Params:    params,
		Timestamp: uint64(time.Now().UnixMilli()),
	}
}

// Hash computes and returns the cryptographic hash of the payload.
// This hash is used for signature creation and verification.
//
// The method works by:
// 1. Marshaling the payload to its JSON representation (compact array format)
// 2. Computing the Keccak256 hash of the JSON bytes
//
// Returns:
//   - The 32-byte Keccak256 hash of the payload
//   - An error if JSON marshaling fails
//
// This hash is deterministic - the same payload will always produce the same hash,
// which is essential for signature verification across different systems.
//
// Example usage:
//
//	hash, err := payload.Hash()
//	if err != nil {
//	    return fmt.Errorf("failed to hash payload: %w", err)
//	}
//	signature := sign.Sign(hash, privateKey)
func (p Payload) Hash() ([]byte, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	hash := crypto.Keccak256(data) // TODO: Replace with unified hashing function across the project
	return hash, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for Payload.
// It expects data in the compact array format: [RequestID, Method, Params, Timestamp]
//
// This custom unmarshaling ensures backward compatibility with the array-based
// protocol format while providing a clean struct-based API for Go code.
//
// The method validates that:
// - The input is a valid JSON array
// - The array contains exactly 4 elements
// - Each element has the correct type
//
// Returns an error if the JSON format is invalid or any element has the wrong type.
func (p *Payload) UnmarshalJSON(data []byte) error {
	var rawArr []json.RawMessage
	if err := json.Unmarshal(data, &rawArr); err != nil {
		return fmt.Errorf("error reading RPCData as array: %w", err)
	}
	if len(rawArr) != 4 {
		return errors.New("invalid RPCData: expected 4 elements in array")
	}

	// Element 0: uint64 RequestID - Must be a valid unsigned integer
	if err := json.Unmarshal(rawArr[0], &p.RequestID); err != nil {
		return fmt.Errorf("invalid request_id: %w", err)
	}

	// Element 1: string Method - Must be a non-empty string
	if err := json.Unmarshal(rawArr[1], &p.Method); err != nil {
		return fmt.Errorf("invalid method: %w", err)
	}

	// Element 2: Params - Must be a JSON object (can be empty {})
	if err := json.Unmarshal(rawArr[2], &p.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Element 3: uint64 Timestamp - Unix milliseconds timestamp
	if err := json.Unmarshal(rawArr[3], &p.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	return nil
}

// MarshalJSON implements the json.Marshaler interface for Payload.
// It always emits the compact array format: [RequestID, Method, Params, Timestamp]
//
// This ensures consistent wire format regardless of how the Payload struct
// is modified in the future, maintaining protocol compatibility.
//
// Example output:
//
//	[12345, "wallet_transfer", {"to": "0xabc", "amount": "100"}, 1634567890123]
func (p Payload) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		p.RequestID,
		p.Method,
		p.Params,
		p.Timestamp,
	})
}

// Params represents method-specific parameters as a map of JSON raw messages.
// This design allows maximum flexibility while maintaining type safety:
// - Parameters are stored as raw JSON until needed
// - The Translate method provides type-safe extraction into Go structs
// - Supports optional parameters and forward compatibility
//
// Example usage:
//
//	// Creating params from a struct
//	params, _ := NewParams(TransferRequest{To: "0x123", Amount: "100"})
//
//	// Accessing individual parameters
//	var amount string
//	json.Unmarshal(params["amount"], &amount)
//
//	// Translating to a struct
//	var req TransferRequest
//	params.Translate(&req)
type Params map[string]json.RawMessage

// NewParams creates a Params map from any JSON-serializable value.
// This is typically used with structs or maps to create method parameters.
//
// The function works by:
// 1. Marshaling the input value to JSON
// 2. Unmarshaling it into a Params map
// 3. Each field becomes a key with its JSON representation as the value
//
// Example:
//
//	type TransferRequest struct {
//	    From   string `json:"from"`
//	    To     string `json:"to"`
//	    Amount string `json:"amount"`
//	}
//
//	req := TransferRequest{
//	    From:   "0x111...",
//	    To:     "0x222...",
//	    Amount: "1000000000000000000",
//	}
//
//	params, err := NewParams(req)
//	// params now contains: {"from": "0x111...", "to": "0x222...", "amount": "1000000000000000000"}
//
// Returns an error if the value cannot be marshaled to JSON or is not a valid object.
func NewParams(v any) (Params, error) {
	if v == nil {
		return Params{}, nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("error marshalling params: %w", err)
	}
	var params Params
	if err := json.Unmarshal(data, &params); err != nil {
		return nil, fmt.Errorf("error unmarshalling params: %w", err)
	}
	return params, nil
}

// Translate extracts the parameters into the provided value (typically a struct).
// This provides type-safe parameter extraction with automatic JSON unmarshaling.
//
// The method works by:
// 1. Marshaling the Params map back to JSON
// 2. Unmarshaling that JSON into the target value
// 3. Go's JSON unmarshaling handles type conversion and validation
//
// Example:
//
//	type BalanceRequest struct {
//	    Address string `json:"address"`
//	    Block   string `json:"block,omitempty"`
//	}
//
//	// In an RPC handler:
//	var req BalanceRequest
//	if err := payload.Params.Translate(&req); err != nil {
//	    return rpc.Errorf("invalid parameters: %v", err)
//	}
//	// req.Address and req.Block are now populated
//
// The target value should be a pointer to the desired type.
// Returns an error if the parameters don't match the expected structure.
func (p Params) Translate(v any) error {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("error marshalling params: %w", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("error unmarshalling params: %w", err)
	}
	return nil
}

// Error extracts and returns an error from the Params if one exists.
// This method checks for the standard "error" key in the params and
// attempts to unmarshal its value as a string error message.
//
// Returns:
//   - An error with the message if the "error" key exists and contains a valid string
//   - nil if no error key exists or if the value cannot be unmarshaled
//
// This is typically used when processing response payloads to check for errors:
//
//	// In a client processing a response
//	if err := response.Res.Params.Error(); err != nil {
//	    // The server returned an error
//	    return fmt.Errorf("RPC error: %w", err)
//	}
//
//	// Process successful response
//	var result TransferResult
//	response.Res.Params.Translate(&result)
//
// This method is designed to work with error params created by NewErrorParams.
func (p Params) Error() error {
	if errMsgRaw, ok := p[errorParamKey]; ok {
		var errMsg string
		if err := json.Unmarshal(errMsgRaw, &errMsg); err == nil {
			return fmt.Errorf("%s", errMsg)
		}
	}
	return nil
}
