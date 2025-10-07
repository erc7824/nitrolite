package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

// Handler is a function that processes an RPC request.
// Handlers can call c.Next() to pass control to the next handler in the chain.
type Handler func(c *Context)

// SendRResponseFunc is a function type for sending RPC notifications to a connection.
// It's provided to event handlers to allow server-initiated messages.
type SendResponseFunc func(method string, params Params)

// Context contains all the information about an RPC request and provides
// methods for handlers to process and respond to the request.
type Context struct {
	// Context is the standard Go context for the request
	Context context.Context
	// UserID is the authenticated user's identifier (empty if not authenticated)
	UserID string
	// Signer is used to sign the response message
	Signer sign.Signer
	// Request is the original RPC request message
	Request Request
	// Response is the response message to be sent back to the client
	Response Response
	// Storage provides per-connection storage for session data
	Storage *SafeStorage

	// handlers is the remaining handler chain to execute
	handlers []Handler
}

// Next executes the next handler in the middleware chain.
// If there are no more handlers, it returns without doing anything.
func (c *Context) Next() {
	if len(c.handlers) == 0 {
		return
	}

	handler := c.handlers[0]
	c.handlers = c.handlers[1:]
	handler(c)
}

// Succeed sets a successful response with the given method and parameters.
// This should be called by handlers to indicate successful processing.
func (c *Context) Succeed(method string, params Params) {
	c.Response.Res = NewPayload(
		c.Request.Req.RequestID,
		method,
		params,
	)
}

// Fail sets an error response for the RPC request. This method should be called by handlers
// when an error occurs during request processing.
//
// Error handling behavior:
//   - If err is an RPCError: The exact error message is sent to the client
//   - If err is any other error type: The fallbackMessage is sent to the client
//   - If both err is nil/non-RPCError AND fallbackMessage is empty: A generic error message is sent
//
// This design allows handlers to control what error information is exposed to clients:
//   - Use RPCError for client-safe, descriptive error messages
//   - Use regular errors with a fallbackMessage to hide internal error details
//
// Usage examples:
//
//	// Hide internal error details from client
//	balance, err := ledger.GetBalance(account)
//	if err != nil {
//		c.Fail(err, "failed to retrieve balance")
//		return
//	}
//
//	// Validation error with no internal error
//	if len(params) < 3 {
//		c.Fail(nil, "invalid parameters: expected at least 3")
//		return
//	}
//
// The response will have Method="error" and Params containing the error message.
func (c *Context) Fail(err error, fallbackMessage string) {
	message := fallbackMessage
	if _, ok := err.(Error); ok {
		message = err.Error()
	}
	if message == "" {
		message = nodeDefaultErrorMessage
	}

	c.Response = NewErrorResponse(
		c.Request.Req.RequestID,
		message,
	)
}

// GetRawResponse returns the signed response message as raw bytes.
// This is called internally after handler processing to prepare the response.
func (c *Context) GetRawResponse() ([]byte, error) {
	return prepareRawResponse(c.Signer, c.Response.Res)
}

// prepareRawResponse creates a signed RPC response message from the given data.
// It marshals the data, signs it, and returns the complete message as bytes.
func prepareRawResponse(signer sign.Signer, payload Payload) ([]byte, error) {
	payloadHash, err := payload.Hash()
	if err != nil {
		return nil, fmt.Errorf("failed to hash response payload: %w", err)
	}

	signature, err := signer.Sign(payloadHash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign response data: %w", err)
	}

	responseMessage := &Response{
		Res: payload,
		Sig: []sign.Signature{signature},
	}
	resMessageBytes, err := json.Marshal(responseMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response message: %w", err)
	}

	return resMessageBytes, nil
}

// SafeStorage provides thread-safe key-value storage for connection-specific data.
// It's used to store session information, authentication policies, and other per-connection state.
type SafeStorage struct {
	// mu protects concurrent access to the storage map
	mu sync.RWMutex
	// storage holds the key-value pairs
	storage map[string]any
}

// NewSafeStorage creates a new thread-safe storage instance.
func NewSafeStorage() *SafeStorage {
	return &SafeStorage{
		storage: make(map[string]any),
	}
}

// Set stores a value with the given key.
// If the key already exists, its value is overwritten.
func (s *SafeStorage) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.storage[key] = value
}

// Get retrieves a value by key.
// Returns the value and true if found, or nil and false if not found.
func (s *SafeStorage) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.storage[key], s.storage[key] != nil
}
