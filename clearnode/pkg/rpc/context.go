package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

// Handler defines the function signature for RPC request processors.
// Handlers receive a Context containing the request and all necessary
// information to process it. They can call c.Next() to delegate to
// the next handler in the middleware chain, enabling composable
// request processing pipelines.
type Handler func(c *Context)

// SendResponseFunc is a function type for sending server-initiated RPC notifications.
// Unlike regular responses that reply to client requests, these functions enable
// the server to push unsolicited messages to clients (e.g., balance updates,
// connection events). The method parameter specifies the notification type,
// and params contains the notification data.
type SendResponseFunc func(method string, params Params)

// Context encapsulates all information related to an RPC request and provides
// methods for handlers to process and respond. It implements a middleware
// pattern where handlers can be chained together, each having the ability
// to process the request, modify the context, or delegate to the next handler.
//
// The Context serves multiple purposes:
//   - Request/response container: Holds the incoming request and outgoing response
//   - Middleware chain management: Tracks and executes the handler chain
//   - Session state: Provides per-connection storage for maintaining state
//   - Authentication context: Carries the authenticated user ID
//   - Response helpers: Convenient methods for success and error responses
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

// Next advances the middleware chain by executing the next handler.
// This enables handlers to perform pre-processing, call Next() to
// delegate to subsequent handlers, then perform post-processing.
// If there are no more handlers in the chain, Next() returns
// immediately without error.
//
// Example middleware pattern:
//
//	func authMiddleware(c *Context) {
//	    // Pre-processing: check authentication
//	    if c.UserID == "" {
//	        c.Fail(nil, "authentication required")
//	        return
//	    }
//	    c.Next() // Continue to next handler
//	    // Post-processing: log the response
//	    log.Info("Request processed", "user", c.UserID)
//	}
func (c *Context) Next() {
	if len(c.handlers) == 0 {
		return
	}

	handler := c.handlers[0]
	c.handlers = c.handlers[1:]
	handler(c)
}

// Succeed sets a successful response for the RPC request.
// This method should be called by handlers when the request has been
// processed successfully. The method parameter typically matches the
// request method, and params contains the result data.
//
// Example:
//
//	func handleGetBalance(c *Context) {
//	    balance := getBalanceForUser(c.UserID)
//	    c.Succeed("get_balance", Params{"balance": balance})
//	}
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
		message = defaultNodeErrorMessage
	}

	c.Response = NewErrorResponse(
		c.Request.Req.RequestID,
		message,
	)
}

// GetRawResponse returns the signed response message as raw bytes.
// This is called internally after handler processing to prepare the response.
func (c *Context) GetRawResponse() ([]byte, error) {
	if c.Response.Res.Method == "" {
		c.Fail(nil, "internal server error: no response from handler")
	}

	return prepareRawResponse(c.Signer, c.Response.Res)
}

// prepareRawResponse creates a complete, signed RPC response message.
// This internal helper:
//  1. Computes the hash of the response payload
//  2. Signs the hash with the provided signer
//  3. Constructs a Response with the payload and signature
//  4. Marshals the complete response to JSON bytes
//
// Returns an error if hashing, signing, or marshaling fails.
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
// Each connection gets its own SafeStorage instance that persists for the
// connection's lifetime. This enables handlers to store and retrieve session
// state, authentication tokens, rate limiting counters, or any other
// per-connection data across multiple requests.
//
// Common use cases:
//   - Storing authentication state and policies
//   - Caching frequently accessed data
//   - Maintaining request counters for rate limiting
//   - Storing connection-specific configuration
type SafeStorage struct {
	// mu protects concurrent access to the storage map
	mu sync.RWMutex
	// storage holds the key-value pairs
	storage map[string]any
}

// NewSafeStorage creates a new thread-safe storage instance.
// The storage starts empty and can be used immediately for
// storing connection-specific data.
func NewSafeStorage() *SafeStorage {
	return &SafeStorage{
		storage: make(map[string]any),
	}
}

// Set stores a value with the given key in the storage.
// If the key already exists, its value is overwritten.
// The value can be of any type. This method is thread-safe
// and can be called concurrently from multiple goroutines.
//
// Example:
//
//	storage.Set("auth_token", "bearer-xyz123")
//	storage.Set("rate_limit_count", 42)
//	storage.Set("user_preferences", userPrefs)
func (s *SafeStorage) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.storage[key] = value
}

// Get retrieves a value by key from the storage.
// Returns the value and true if the key exists, or nil and false
// if the key is not found. The caller must type-assert the returned
// value to the expected type.
//
// Example:
//
//	if val, ok := storage.Get("auth_token"); ok {
//	    token := val.(string)
//	    // Use token...
//	}
func (s *SafeStorage) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.storage[key]
	return value, exists
}
