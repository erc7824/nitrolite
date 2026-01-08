package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type MsgType uint8

var (
	MsgTypeReq   MsgType = 1
	MsgTypeResp  MsgType = 2
	MsgTypeEvent MsgType = 3
)

// Message represents the core data structure for RPC communication.
// It contains all the information needed to process an RPC call, response, or event.
//
// Messages are encoded as JSON arrays for compact transmission:
// [Type, RequestID, Method, Params, Timestamp]
//
// This encoding reduces message size while maintaining human readability
// and allows for efficient parsing. The array format is automatically
// handled by the custom JSON marshaling methods.
type Message struct {
	Type MsgType `json:"type"`

	// RequestID is a unique identifier for tracking requests and matching responses.
	// Clients should generate unique IDs to prevent collisions and enable proper
	// request-response correlation.
	RequestID uint64 `json:"request_id"`

	// Method specifies the RPC method to be invoked (e.g., "wallet_transfer").
	// Method names should follow a consistent naming convention, typically
	// using lowercase with underscores (e.g., "module_action").
	Method string `json:"method"`

	// Payload contains the method-specific parameters as a flexible map.
	// This allows different methods to have different parameter structures
	// while maintaining type safety through the Translate method.
	Payload Payload `json:"payload"`

	// Timestamp is the Unix timestamp in milliseconds when the payload was created.
	// This is used for replay protection and request expiration checks.
	// Servers should validate that timestamps are within an acceptable time window.
	Timestamp uint64 `json:"ts"`
}

// NewMessage creates a new Message with the given request ID, type, method, and parameters.
// The timestamp is automatically set to the current time in Unix milliseconds.
//
// Example:
//
//	params, _ := NewParams(map[string]string{"address": "0x123"})
//	message := NewMessage(12345, MsgTypeReq, "wallet_getBalance", params)
//
// The resulting message will have the current timestamp and can be used
// for requests, responses, or events depending on the type specified.
func NewMessage(typ MsgType, id uint64, method string, params Payload) Message {
	if params == nil {
		params = Payload{}
	}

	return Message{
		Type:      typ,
		RequestID: id,
		Method:    method,
		Payload:   params,
		Timestamp: uint64(time.Now().UnixMilli()),
	}
}

// NewRequest creates a new Request message with the given request ID, method, and parameters.
// The message type is automatically set to MsgTypeReq and the timestamp is set to the current time.
//
// Example usage:
//
//	// Create a request message
//	params, _ := NewParams(map[string]string{"address": "0x123"})
//	request := NewRequest(12345, "wallet_getBalance", params)
//
//	// Create a request with complex parameters
//	type TransferParams struct {
//	    To     string `json:"to"`
//	    Amount string `json:"amount"`
//	}
//	params, _ := NewParams(TransferParams{To: "0xabc", Amount: "100"})
//	request := NewRequest(67890, "wallet_transfer", params)
func NewRequest(requestID uint64, method string, params Payload) Message {
	return NewMessage(MsgType(MsgTypeReq), requestID, method, params)
}

// NewResponse creates a new Response message with the given request ID, method, and parameters.
// The message type is automatically set to MsgTypeResp and the timestamp is set to the current time.
//
// Example usage:
//
//	// Create a success response
//	params, _ := NewParams(map[string]string{"balance": "1000"})
//	response := NewResponse(12345, "wallet_getBalance", params)
//
//	// Create an error response (use NewErrorResponse for convenience)
//	errParams := NewErrorParams("insufficient balance")
//	response := NewResponse(12345, ErrorMethod.String(), errParams)
func NewResponse(requestID uint64, method string, params Payload) Message {
	return NewMessage(MsgType(MsgTypeResp), requestID, method, params)
}

func NewEvent(requestID uint64, method string, params Payload) Message {
	return NewMessage(MsgType(MsgTypeEvent), requestID, method, params)
}

// NewErrorResponse creates a Response message containing an error message.
// This is a convenience function that combines error parameter creation
// and response construction in a single call.
//
// Parameters:
//   - requestID: The ID from the original request
//   - errMsg: The error message to send to the client
//
// Example usage:
//
//	// In an RPC handler when an error occurs
//	if err := validateRequest(request); err != nil {
//	    return NewErrorResponse(request.RequestID, err.Error())
//	}
//
//	// Creating an error response
//	errorResponse := NewErrorResponse(12345, "insufficient balance")
//
// The resulting response will have the method set to the error method constant
// and params in the format: {"error": "<errMsg>"}
func NewErrorResponse(requestID uint64, errMsg string) Message {
	errParams := NewErrorPayload(errMsg)
	return NewMessage(MsgType(MsgTypeResp), requestID, ErrorMethod.String(), errParams)
}

// Error checks if the Message contains an error and returns it.
// This method extracts any error stored in the message's payload
// under the standard "error" key.
//
// Returns:
//   - An error if the message contains an error message
//   - nil if the message represents a successful operation
//
// This is typically used by clients to check if an RPC call failed:
//
//	// After receiving and unmarshaling a response message
//	var message Message
//	if err := json.Unmarshal(data, &message); err != nil {
//	    return fmt.Errorf("failed to unmarshal message: %w", err)
//	}
//
//	// Check if the message contains an error
//	if err := message.Error(); err != nil {
//	    return fmt.Errorf("RPC call failed: %w", err)
//	}
//
//	// Process successful response
//	var result TransferResult
//	message.Payload.Translate(&result)
//
// This method is designed to work with error responses created by NewErrorResponse
// or any response where errors are stored using NewErrorParams.
func (r Message) Error() error {
	if r.Method != ErrorMethod.String() {
		return nil
	}

	return r.Payload.Error()
}

// UnmarshalJSON implements the json.Unmarshaler interface for Message.
// It expects data in the compact array format: [Type, RequestID, Method, Params, Timestamp]
//
// This custom unmarshaling ensures backward compatibility with the array-based
// protocol format while providing a clean struct-based API for Go code.
//
// The method validates that:
// - The input is a valid JSON array
// - The array contains exactly 5 elements
// - Each element has the correct type
//
// Returns an error if the JSON format is invalid or any element has the wrong type.
func (p *Message) UnmarshalJSON(data []byte) error {
	var rawArr []json.RawMessage
	if err := json.Unmarshal(data, &rawArr); err != nil {
		return fmt.Errorf("error reading RPCData as array: %w", err)
	}
	if len(rawArr) != 5 {
		return errors.New("invalid RPCData: expected 5 elements in array")
	}

	// Element 0: uint8 Type - Message type (request, response, or event)
	if err := json.Unmarshal(rawArr[0], &p.Type); err != nil {
		return fmt.Errorf("invalid type: %w", err)
	}

	// Element 1: uint64 RequestID - Must be a valid unsigned integer
	if err := json.Unmarshal(rawArr[1], &p.RequestID); err != nil {
		return fmt.Errorf("invalid request_id: %w", err)
	}

	// Element 2: string Method - Must be a non-empty string
	if err := json.Unmarshal(rawArr[2], &p.Method); err != nil {
		return fmt.Errorf("invalid method: %w", err)
	}

	// Element 3: Params - Must be a JSON object (can be empty {})
	if err := json.Unmarshal(rawArr[3], &p.Payload); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Element 4: uint64 Timestamp - Unix milliseconds timestamp
	if err := json.Unmarshal(rawArr[4], &p.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	return nil
}

// MarshalJSON implements the json.Marshaler interface for Message.
// It always emits the compact array format: [Type, RequestID, Method, Params, Timestamp]
//
// This ensures consistent wire format regardless of how the Message struct
// is modified in the future, maintaining protocol compatibility.
//
// Example output:
//
//	[1, 12345, "wallet_transfer", {"to": "0xabc", "amount": "100"}, 1634567890123]
func (p Message) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		p.Type,
		p.RequestID,
		p.Method,
		p.Payload,
		p.Timestamp,
	})
}

// Payload represents method-specific parameters as a map of JSON raw messages.
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
type Payload map[string]json.RawMessage

// NewPayload creates a Payload map from any JSON-serializable value.
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
//	payload, err := NewPayload(req)
//	// payload now contains: {"from": "0x111...", "to": "0x222...", "amount": "1000000000000000000"}
//
// Returns an error if the value cannot be marshaled to JSON or is not a valid object.
func NewPayload(v any) (Payload, error) {
	if v == nil {
		return Payload{}, nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("error marshalling params: %w", err)
	}
	var params Payload
	if err := json.Unmarshal(data, &params); err != nil {
		return nil, fmt.Errorf("error unmarshalling params: %w", err)
	}
	return params, nil
}

// Translate extracts the parameters into the provided value (typically a struct).
// This provides type-safe parameter extraction with automatic JSON unmarshaling.
//
// The method works by:
// 1. Marshaling the Payload map back to JSON
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
//	if err := message.Payload.Translate(&req); err != nil {
//	    return rpc.Errorf("invalid parameters: %v", err)
//	}
//	// req.Address and req.Block are now populated
//
// The target value should be a pointer to the desired type.
// Returns an error if the parameters don't match the expected structure.
func (p Payload) Translate(v any) error {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("error marshalling params: %w", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("error unmarshalling params: %w", err)
	}
	return nil
}

// Error extracts and returns an error from the Payload if one exists.
// This method checks for the standard "error" key in the payload and
// attempts to unmarshal its value as a string error message.
//
// Returns:
//   - An error with the message if the "error" key exists and contains a valid string
//   - nil if no error key exists or if the value cannot be unmarshaled
//
// This is typically used when processing response payloads to check for errors:
//
//	// In a client processing a response message
//	if err := message.Payload.Error(); err != nil {
//	    // The server returned an error
//	    return fmt.Errorf("RPC error: %w", err)
//	}
//
//	// Process successful response
//	var result TransferResult
//	message.Payload.Translate(&result)
//
// This method is designed to work with error params created by NewErrorParams.
func (p Payload) Error() error {
	if errMsgRaw, ok := p[errorParamKey]; ok {
		var errMsg string
		if err := json.Unmarshal(errMsgRaw, &errMsg); err == nil {
			return fmt.Errorf("%s", errMsg)
		}
	}
	return nil
}
