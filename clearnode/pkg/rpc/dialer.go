package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/log"
	"github.com/gorilla/websocket"
)

// Dialer is the interface for RPC client connections.
// It provides methods to establish connections, send requests, and receive responses.
type Dialer interface {
	// Dial establishes a connection to the specified URL.
	// This method is designed to be called in a goroutine as it blocks until the connection is closed.
	// The handleClosure callback is invoked when the connection is closed, with an error if any.
	Dial(ctx context.Context, url string, handleClosure func(err error)) error

	// IsConnected returns true if the dialer has an active connection.
	IsConnected() bool

	// Call sends an RPC request and waits for a response.
	// It returns an error if the request cannot be sent or no response is received.
	// The context can be used to cancel the request.
	Call(ctx context.Context, req *Request) (*Response, error)

	// EventCh returns a read-only channel for receiving unsolicited events from the server.
	// Events are responses that don't match any pending request ID.
	EventCh() <-chan *Response
}

// dialCtx holds the connection context and resources
type dialCtx struct {
	ctx  context.Context // Connection context for lifecycle management
	conn *websocket.Conn // WebSocket connection
	lg   log.Logger      // Logger for this connection
}

// WebsocketDialerConfig contains configuration options for the WebSocket dialer
type WebsocketDialerConfig struct {
	// HandshakeTimeout is the duration to wait for the WebSocket handshake to complete
	HandshakeTimeout time.Duration

	// PingInterval is how often to send ping messages to keep the connection alive
	PingInterval time.Duration

	// PingRequestID is the request ID used for ping messages
	// This should be a reserved ID that won't conflict with regular requests
	PingRequestID uint64

	// EventChanSize is the buffer size for the event channel
	// A larger buffer prevents blocking when processing many unsolicited events
	EventChanSize int
}

// DefaultWebsocketDialerConfig provides sensible defaults for WebSocket connections
var DefaultWebsocketDialerConfig = WebsocketDialerConfig{
	HandshakeTimeout: 5 * time.Second,
	PingInterval:     5 * time.Second,
	PingRequestID:    100,
	EventChanSize:    100,
}

// WebsocketDialer implements the Dialer interface using WebSocket connections.
// It provides thread-safe RPC communication with automatic ping/pong handling.
type WebsocketDialer struct {
	cfg           WebsocketDialerConfig
	dialCtx       *dialCtx                  // Connection context and resources
	eventCh       chan *Response            // Channel for unsolicited events
	responseSinks map[uint64]chan *Response // Map of request IDs to response channels
	mu            sync.RWMutex              // Protects dialCtx and responseSinks
	writeMu       sync.Mutex                // Serializes WebSocket write operations
}

// Ensure WebsocketDialer implements the Dialer interface
var _ Dialer = (*WebsocketDialer)(nil)

// NewWebsocketDialer creates a new WebSocket dialer with the given configuration
func NewWebsocketDialer(cfg WebsocketDialerConfig) *WebsocketDialer {
	return &WebsocketDialer{
		cfg:           cfg,
		eventCh:       make(chan *Response, cfg.EventChanSize),
		responseSinks: make(map[uint64]chan *Response),
	}
}

// Dial establishes a WebSocket connection to the specified URL.
// This method blocks until the connection is closed, so it should typically be called in a goroutine.
// It starts three background goroutines:
// - One to handle context cancellation
// - One to read and route incoming messages
// - One to send periodic ping messages
//
// Example:
//
//	dialer := NewWebsocketDialer(DefaultWebsocketDialerConfig)
//	go dialer.Dial(ctx, "ws://localhost:8080/ws", func(err error) {
//	    if err != nil {
//	        log.Error("Connection closed", "error", err)
//	    }
//	})
func (d *WebsocketDialer) Dial(ctx context.Context, url string, handleClosure func(err error)) error {
	if d.IsConnected() {
		return ErrAlreadyConnected
	}

	dialer := websocket.Dialer{
		HandshakeTimeout:  d.cfg.HandshakeTimeout,
		EnableCompression: true,
	}

	// Establish WebSocket connection
	conn, _, err := dialer.DialContext(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDialingWebsocket, err)
	}

	// Create a cancelable context for managing goroutines
	parentCtx, cancel := context.WithCancel(ctx)
	wg := sync.WaitGroup{}
	wg.Add(3) // We'll start 3 goroutines

	// Track the first error that causes closure
	var closureErr error
	childHandleClosure := func(err error) {
		if err != nil && closureErr == nil {
			closureErr = err
		}

		cancel() // Cancel context to stop other goroutines
		wg.Done()
	}

	// Store connection context
	d.mu.Lock()
	d.dialCtx = &dialCtx{
		ctx:  parentCtx,
		conn: conn,
		lg:   log.FromContext(ctx).WithName("ws-dialer"),
	}
	d.eventCh = make(chan *Response, d.cfg.EventChanSize)
	d.mu.Unlock()

	// Start background goroutines
	go d.closeOnContextDone(parentCtx, childHandleClosure)
	go d.readMessages(parentCtx, childHandleClosure)
	go d.pingPeriodically(parentCtx, childHandleClosure)

	// Wait for all goroutines to finish before calling the closure handler
	go func() {
		wg.Wait()
		handleClosure(closureErr)
	}()

	return nil
}

// IsConnected returns true if the dialer has an active connection
func (d *WebsocketDialer) IsConnected() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.dialCtx != nil && d.dialCtx.ctx.Err() == nil
}

// closeOnContextDone waits for the context to be done and then closes the connection
func (d *WebsocketDialer) closeOnContextDone(ctx context.Context, handleClosure func(err error)) {
	<-ctx.Done()

	// Close the websocket connection
	d.mu.RLock()
	conn := d.dialCtx.conn
	d.mu.RUnlock()

	err := conn.Close()

	// Clean up response sinks to prevent goroutine leaks
	d.mu.Lock()
	for _, sink := range d.responseSinks {
		close(sink)
	}
	d.responseSinks = make(map[uint64]chan *Response)
	d.mu.Unlock()

	handleClosure(err)
}

// readMessages continuously reads messages from the WebSocket connection
// and routes them to the appropriate response channel or event channel
func (d *WebsocketDialer) readMessages(ctx context.Context, handleClosure func(err error)) {
	// Cache connection and logger to avoid repeated mutex access
	d.mu.RLock()
	conn := d.dialCtx.conn
	lg := d.dialCtx.lg
	d.mu.RUnlock()

	for {
		// Read next message from WebSocket
		_, messageBytes, err := conn.ReadMessage()
		if ctx.Err() != nil {
			handleClosure(nil)
			lg.Info("Websocket read loop exiting due to context done")
			return
		} else if _, ok := err.(net.Error); ok {
			handleClosure(fmt.Errorf("%w: %w", ErrConnectionTimeout, err))
			lg.Error("Websocket connection timeout", "error", err)
			return
		} else if err != nil {
			handleClosure(fmt.Errorf("%w: %w", ErrReadingMessage, err))
			lg.Error("Websocket read error", "error", err)
			return
		}

		// Parse the response
		var msg Response
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			lg.Warn("Malformed message", "message", string(messageBytes), "error", err)
			continue
		}

		// Route the response to the appropriate channel
		d.mu.Lock()
		responseSink, exists := d.responseSinks[msg.Res.RequestID]
		d.mu.Unlock()

		if !exists {
			// No pending request for this ID, treat as an unsolicited event
			responseSink = d.eventCh
		}

		// Try to send the response, but don't block
		select {
		case <-ctx.Done():
			handleClosure(nil)
			return
		case responseSink <- &msg:
			// Successfully sent
		default:
			// Channel full, drop the message
			lg.Warn("Response channel full, dropping message", "requestID", msg.Res.RequestID)
		}
	}
}

// Call sends an RPC request and waits for a response.
// The request must have a unique RequestID that identifies this call.
// The method is thread-safe and can be called concurrently.
//
// The context can be used to set a timeout for the request:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	resp, err := dialer.Call(ctx, request)
func (d *WebsocketDialer) Call(ctx context.Context, req *Request) (*Response, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	// Check connection and register response channel atomically
	d.mu.Lock()
	if d.dialCtx == nil || d.dialCtx.ctx.Err() != nil {
		d.mu.Unlock()
		return nil, ErrNotConnected
	}
	conn := d.dialCtx.conn
	connCtx := d.dialCtx.ctx
	responseSink := make(chan *Response, 1) // Buffered to prevent blocking in readMessages
	d.responseSinks[req.Req.RequestID] = responseSink
	d.mu.Unlock()

	// Marshal the request
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMarshalingRequest, err)
	}

	// Send the request (WebSocket writes must be serialized)
	d.writeMu.Lock()
	err = conn.WriteMessage(websocket.TextMessage, reqJSON)
	d.writeMu.Unlock()

	if err != nil {
		// Clean up on send failure
		d.mu.Lock()
		delete(d.responseSinks, req.Req.RequestID)
		d.mu.Unlock()
		return nil, fmt.Errorf("%w: %w", ErrSendingRequest, err)
	}

	// Wait for response or timeout
	var res *Response
	select {
	case <-ctx.Done():
		// Request context cancelled
	case <-connCtx.Done():
		// Connection closed
	case res = <-responseSink:
		// Got response
	}

	// Clean up response channel
	d.mu.Lock()
	delete(d.responseSinks, req.Req.RequestID)
	d.mu.Unlock()

	if res == nil {
		return nil, fmt.Errorf("%w for request %d", ErrNoResponse, req.Req.RequestID)
	}
	return res, nil
}

// pingPeriodically sends ping requests at regular intervals to keep the connection alive
func (d *WebsocketDialer) pingPeriodically(ctx context.Context, handleClosure func(err error)) {
	// Cache logger to avoid repeated mutex access
	d.mu.RLock()
	lg := d.dialCtx.lg
	d.mu.RUnlock()

	ticker := time.NewTicker(d.cfg.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			handleClosure(nil)
			lg.Info("Ping loop exiting due to context done")
			return
		case <-ticker.C:
			// Send ping request
			var params Params
			payload := NewPayload(d.cfg.PingRequestID, "ping", params)
			req := NewRequest(payload)

			// Use the connection context for ping requests
			res, err := d.Call(ctx, &req)
			if err != nil {
				handleClosure(fmt.Errorf("%w: %w", ErrSendingPing, err))
				lg.Error("Error sending ping", "error", err)
				return
			}

			// Verify we got a pong response
			if res.Res.Method != "pong" {
				lg.Warn("Unexpected response to ping", "method", res.Res.Method)
			}
		}
	}
}

// EventCh returns a read-only channel for receiving unsolicited events.
// Events are responses that don't match any pending request ID.
// The channel will receive nil when the connection is closed.
//
// Example:
//
//	for event := range dialer.EventCh() {
//	    if event == nil {
//	        // Connection closed
//	        break
//	    }
//	    // Handle event
//	    log.Info("Received event", "method", event.Res.Method)
//	}
func (d *WebsocketDialer) EventCh() <-chan *Response {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.eventCh
}
