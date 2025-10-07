package rpc

import (
	"context"
	"sync"
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/log"
	"github.com/gorilla/websocket"
)

var (
	defaultResponseWriteDuration = 5 * time.Second // Default timeout for writing responses to WebSocket
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
	// Write sends a message to the connection's write sink
	WriteRawResponse(message []byte)
	// Serve starts the connection's lifecycle, handling reads and writes
	Serve(parentCtx context.Context, handleClosure func(error))
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
	websocketConn *websocket.Conn
	// logger is used for logging events related to this connection
	logger log.Logger
	// onMessageSentHandlers are callbacks that are called when a message is sent
	onMessageSentHandlers []func()

	// writeSink is the channel for sending messages to this connection
	writeSink chan []byte
	// processSink is the channel for processing incoming messages
	processSink chan []byte
	// closeConnCh is a channel that can be used to signal connection closure
	closeConnCh chan struct{}

	// mu is a mutex to protect access to user-related data
	mu sync.RWMutex
}

// NewWebsocketConnection creates a new Connection instance.
func NewWebsocketConnection(connID, userID string, websocketConn *websocket.Conn, logger log.Logger, onMessageSentHandlers ...func()) *WebsocketConnection {
	if onMessageSentHandlers == nil {
		onMessageSentHandlers = []func(){}
	}

	return &WebsocketConnection{
		connectionID:          connID,
		userID:                userID,
		websocketConn:         websocketConn,
		logger:                logger.WithKV("connectionID", connID),
		onMessageSentHandlers: onMessageSentHandlers,

		writeSink:   make(chan []byte, 10),
		processSink: make(chan []byte, 10),
		closeConnCh: make(chan struct{}, 1),
	}
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
	wg.Add(2)

	var closureErr error
	childHandleClosure := func(err error) {
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
func (conn *WebsocketConnection) WriteRawResponse(message []byte) {
	timer := time.NewTimer(defaultResponseWriteDuration)
	defer timer.Stop()

	select {
	case <-timer.C:
		select {
		case conn.closeConnCh <- struct{}{}:
		default:
		}
		return
	case conn.writeSink <- message:
		return
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

			// Call all message sent handlers
			for _, handler := range conn.onMessageSentHandlers {
				handler()
			}
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
