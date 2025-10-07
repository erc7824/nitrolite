package rpc

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/log"
	"github.com/gorilla/websocket"
)

var (
	// defaultWsConnWriteTimeout is the default maximum duration to wait for a write to complete.
	defaultWsConnWriteTimeout = 5 * time.Second
	// defaultWsConnProcessBufferSize is the default size of the buffer for processing incoming messages.
	defaultWsConnProcessBufferSize = 10
	// defaultWsConnWriteBufferSize is the default size of the buffer for outgoing messages.
	defaultWsConnWriteBufferSize = 10
)

type Connection interface {
	// ConnectionID returns the unique identifier for this connection
	ConnectionID() string
	// UserID returns the authenticated user's identifier for this connection
	UserID() string
	// SetUserID sets the UserID for this connection
	SetUserID(userID string)
	// RawRequests returns the channel for processing incoming requests
	RawRequests() <-chan []byte
	// WriteRawResponse sends a raw response message to the connection's write sink
	// Returns true if the message was successfully queued, false if it timed out
	WriteRawResponse(message []byte) bool
	// Serve starts the connection's lifecycle, handling reads and writes
	Serve(parentCtx context.Context, handleClosure func(error))
}

// GorillaWsConnectionAdapter abstracts the methods of a WebSocket connection needed by WebsocketConnection.
type GorillaWsConnectionAdapter interface {
	// ReadMessage reads a message from the WebSocket connection.
	ReadMessage() (messageType int, p []byte, err error)
	// NextWriter returns a writer for the next message to be sent on the WebSocket connection.
	NextWriter(messageType int) (io.WriteCloser, error)
	// Close closes the WebSocket connection.
	Close() error
}

// WebsocketConnection represents an active WebSocket connection.
// It tracks the authentication, stores session data, and provides communication channels.
type WebsocketConnection struct {
	// ctx is the parent context for managing goroutines
	ctx context.Context
	// connectionID is a unique identifier for this connection
	connectionID string
	// UserID is the authenticated user's identifier (empty if not authenticated)
	userID string
	// websocketConn is the underlying WebSocket connection
	websocketConn GorillaWsConnectionAdapter
	// writeTimeout is the maximum duration to wait for a write to complete
	writeTimeout time.Duration

	// logger is used for logging events related to this connection
	logger log.Logger
	// onMessageSentHandler is called when a message is sent
	onMessageSentHandler func([]byte)
	// writeSink is the channel for sending messages to this connection
	writeSink chan []byte
	// processSink is the channel for processing incoming messages
	processSink chan []byte
	// closeConnCh is a channel that can be used to signal connection closure
	closeConnCh chan struct{}

	// mu is a mutex to protect access to user-related data
	mu sync.RWMutex
}

type WebsocketConnectionConfig struct {
	ConnectionID  string
	UserID        string
	WebsocketConn GorillaWsConnectionAdapter

	WriteTimeout         time.Duration
	WriteBufferSize      int
	ProcessBufferSize    int
	Logger               log.Logger
	OnMessageSentHandler func([]byte)
}

// NewWebsocketConnection creates a new Connection instance.
func NewWebsocketConnection(config WebsocketConnectionConfig) (*WebsocketConnection, error) {
	if config.ConnectionID == "" {
		return nil, fmt.Errorf("connection ID cannot be empty")
	}
	if config.WebsocketConn == nil {
		return nil, fmt.Errorf("websocket connection cannot be nil")
	}
	if config.Logger == nil {
		config.Logger = log.NewNoopLogger()
	}
	if config.WriteTimeout <= 0 {
		config.WriteTimeout = defaultWsConnWriteTimeout
	}
	if config.WriteBufferSize <= 0 {
		config.WriteBufferSize = defaultWsConnWriteBufferSize
	}
	if config.ProcessBufferSize <= 0 {
		config.ProcessBufferSize = defaultWsConnProcessBufferSize
	}
	if config.OnMessageSentHandler == nil {
		config.OnMessageSentHandler = func([]byte) {}
	}

	return &WebsocketConnection{
		connectionID:  config.ConnectionID,
		userID:        config.UserID,
		websocketConn: config.WebsocketConn,
		writeTimeout:  config.WriteTimeout,

		logger:               config.Logger.WithKV("connectionID", config.ConnectionID),
		onMessageSentHandler: config.OnMessageSentHandler,
		writeSink:            make(chan []byte, config.WriteBufferSize),
		processSink:          make(chan []byte, config.ProcessBufferSize),
		closeConnCh:          make(chan struct{}, 1),
	}, nil
}

// Serve starts the connection's lifecycle.
// It handles reading and writing messages, and waits for the connection to close.
func (conn *WebsocketConnection) Serve(parentCtx context.Context, handleClosure func(error)) {
	conn.mu.Lock()
	if conn.ctx != nil {
		conn.mu.Unlock()
		handleClosure(nil) // Connection is already running
		return
	}
	conn.ctx = parentCtx
	conn.mu.Unlock()

	// Create a child context that can be cancelled to stop all goroutines
	childCtx, cancel := context.WithCancel(parentCtx)
	wg := &sync.WaitGroup{}
	wg.Add(3)

	var closureErr error
	var closureErrMu sync.Mutex
	childHandleClosure := func(err error) {
		closureErrMu.Lock()
		defer closureErrMu.Unlock()

		// Capture the first error encountered
		if err != nil && closureErr == nil {
			closureErr = err
		}

		cancel()  // Trigger exit on other goroutines
		wg.Done() // Decrement the wait group counter
	}

	// Start reading messages from the WebSocket connection
	go conn.readMessages(childHandleClosure)

	// Start writing messages to the WebSocket connection
	go conn.writeMessages(childCtx, childHandleClosure)

	// Wait for the WebSocket connection to close
	go conn.waitForConnClose(childCtx, childHandleClosure)

	go func() {
		// Wait for all goroutines to finish
		wg.Wait()

		closureErrMu.Lock()
		defer closureErrMu.Unlock()

		// Invoke the closure handler with any error that occurred
		handleClosure(closureErr)

		// Close the WebSocket connection
		if err := conn.websocketConn.Close(); err != nil {
			conn.logger.Error("error closing WebSocket connection", "error", err)
		}
	}()
}

// ConnectionID returns the unique identifier for this connection.
func (conn *WebsocketConnection) ConnectionID() string {
	return conn.connectionID
}

// UserID returns the authenticated user's identifier for this connection.
func (conn *WebsocketConnection) UserID() string {
	conn.mu.RLock()
	defer conn.mu.RUnlock()
	return conn.userID
}

// SetUserID sets the UserID for this connection.
func (conn *WebsocketConnection) SetUserID(userID string) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.userID = userID
}

// RawRequests returns the channel for processing incoming requests.
func (conn *WebsocketConnection) RawRequests() <-chan []byte {
	return conn.processSink
}

// WriteRawResponse sends a message to the connection's write sink.
// If the write operation takes too long, it signals the connection to close.
// This is useful for preventing hangs if the client is unresponsive.
func (conn *WebsocketConnection) WriteRawResponse(message []byte) bool {
	timer := time.NewTimer(conn.writeTimeout)
	defer timer.Stop()

	select {
	case <-timer.C:
		select {
		case conn.closeConnCh <- struct{}{}:
		default:
		}
		return false
	case conn.writeSink <- message:
		return true
	}
}

// readMessages listens for incoming messages on the WebSocket connection.
// It reads messages and sends them to the processSink channel for further processing.
func (conn *WebsocketConnection) readMessages(handleClosure func(error)) {
	defer close(conn.processSink) // Close the processing channel when done

	for {
		_, messageBytes, err := conn.websocketConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				conn.logger.Error("WebSocket connection closed with unexpected reason", "error", err)
				handleClosure(err)
			} else {
				handleClosure(nil) // Normal closure
			}
			return
		}

		if len(messageBytes) == 0 {
			conn.logger.Debug("received empty message, skipping")
			continue // Skip empty messages
		}
		conn.processSink <- messageBytes // Send message to processing channel
	}
}

// writeMessages handles outgoing messages to the WebSocket connection.
// It reads from the message sink channel and writes to the WebSocket.
func (conn *WebsocketConnection) writeMessages(ctx context.Context, handleClosure func(error)) {
	defer handleClosure(nil) // Stop other goroutines

	for {
		select {
		case <-ctx.Done():
			conn.logger.Debug("context done, stopping message writing")
			return
		case messageBytes := <-conn.writeSink:
			if len(messageBytes) == 0 {
				continue // Skip empty messages
			}

			w, err := conn.websocketConn.NextWriter(websocket.TextMessage)
			if err != nil {
				conn.logger.Error("error getting writer for response", "error", err)
				continue
			}

			if _, err := w.Write(messageBytes); err != nil {
				conn.logger.Error("error writing response", "error", err)
				w.Close()
				continue
			}

			if err := w.Close(); err != nil {
				conn.logger.Error("error closing writer for response", "error", err)
				continue
			}

			conn.onMessageSentHandler(messageBytes)
		}
	}
}

// waitForConnClose waits for the WebSocket connection to close.
// It listens for the close signal and logs the closure event.
func (conn *WebsocketConnection) waitForConnClose(ctx context.Context, handleClosure func(error)) {
	defer handleClosure(nil) // Stop other goroutines when done

	select {
	case <-ctx.Done():
		conn.logger.Debug("context done, stopping connection close wait")
	case <-conn.closeConnCh:
		conn.logger.Info("WebSocket connection closed by server", "connectionID", conn.ConnectionID())
	}
}
