package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// RPCConnection represents an active WebSocket connection.
// It tracks the authentication, stores session data, and provides communication channels.
type RPCConnection struct {
	// connectionID is a unique identifier for this connection
	connectionID string
	// UserID is the authenticated user's identifier (empty if not authenticated)
	userID string
	// websocketConn is the underlying WebSocket connection
	websocketConn *websocket.Conn
	// logger is used for logging events related to this connection
	logger Logger
	// onMessageSentHandlers are callbacks that are called when a message is sent
	onMessageSentHandlers []func()

	// writeSink is the channel for sending messages to this connection
	writeSink chan []byte
	// processSink is the channel for processing incoming messages
	processSink chan []byte
	// closeConnCh is a channel that can be used to signal connection closure
	closeConnCh chan struct{}

	// userMu is a mutex to protect access to user-related data
	userMu sync.RWMutex
}

// NewRPCConnection creates a new RPCConnection instance.
func NewRPCConnection(connID, userID string, websocketConn *websocket.Conn, logger Logger, onMessageSentHandlers ...func()) *RPCConnection {
	if onMessageSentHandlers == nil {
		onMessageSentHandlers = []func(){}
	}

	return &RPCConnection{
		connectionID:          connID,
		userID:                userID,
		websocketConn:         websocketConn,
		logger:                logger.With("connectionID", connID),
		onMessageSentHandlers: onMessageSentHandlers,

		writeSink:   make(chan []byte, 10),
		processSink: make(chan []byte, 10),
		closeConnCh: make(chan struct{}),
	}
}

// Serve starts the connection's lifecycle.
// It handles reading and writing messages, and waits for the connection to close.
func (conn *RPCConnection) Serve(parentCtx context.Context, abortParents func()) {
	defer abortParents() // Stop parent goroutines when done

	ctx, cancel := context.WithCancel(parentCtx)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	abortOthers := func() {
		cancel()  // Trigger exit on other goroutines
		wg.Done() // Decrement the wait group counter
	}

	// Start reading messages from the WebSocket connection
	go conn.readMessages(cancel)

	// Start writing messages to the WebSocket connection
	go conn.writeMessages(ctx, abortOthers)

	// Wait for the WebSocket connection to close
	go conn.waitForConnClose(ctx, abortOthers)

	// Wait for all goroutines to finish
	wg.Wait()
	// Close the WebSocket connection
	if err := conn.websocketConn.Close(); err != nil {
		conn.logger.Error("error closing WebSocket connection", "error", err)
	}
}

// ConnectionID returns the unique identifier for this connection.
func (conn *RPCConnection) ConnectionID() string {
	return conn.connectionID
}

// UserID returns the authenticated user's identifier for this connection.
func (conn *RPCConnection) UserID() string {
	conn.userMu.RLock()
	defer conn.userMu.RUnlock()
	return conn.userID
}

// SetUserID sets the UserID for this connection.
func (conn *RPCConnection) SetUserID(userID string) {
	conn.userMu.Lock()
	defer conn.userMu.Unlock()
	conn.userID = userID
}

// ProcessSink returns the channel for processing incoming messages.
func (conn *RPCConnection) ProcessSink() <-chan []byte {
	return conn.processSink
}

// readMessages listens for incoming messages on the WebSocket connection.
// It reads messages and sends them to the processSink channel for further processing.
func (conn *RPCConnection) readMessages(abortOthers func()) {
	defer abortOthers()           // Stop other goroutines when done
	defer close(conn.processSink) // Close the processing channel when done

	for {
		_, messageBytes, err := conn.websocketConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				conn.logger.Error("WebSocket connection closed with unexpected reason", "error", err)
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
func (conn *RPCConnection) writeMessages(ctx context.Context, abortOthers context.CancelFunc) {
	defer abortOthers() // Stop other goroutines

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
func (conn *RPCConnection) waitForConnClose(ctx context.Context, abortOthers context.CancelFunc) {
	defer abortOthers() // Stop other goroutines when done

	select {
	case <-ctx.Done():
		conn.logger.Debug("context done, stopping connection close wait")
	case <-conn.closeConnCh:
		conn.logger.Info("WebSocket connection closed by server", "connectionID", conn.ConnectionID)
	}
}

// Write sends a message to the connection's write sink.
// If the write operation takes too long, it signals the connection to close.
// This is useful for preventing hangs if the client is unresponsive.
func (conn *RPCConnection) Write(message []byte) {
	select {
	case <-time.After(defaultRPCMessageWriteDuration):
		conn.closeConnCh <- struct{}{} // Signal connection closure if write times out
		return
	case conn.writeSink <- message:
		return
	}
}

// rpcConnectionHub manages all active WebSocket connections.
// It provides thread-safe operations for connection tracking and user mapping.
type rpcConnectionHub struct {
	// connections maps connection IDs to RPCConnection instances
	connections map[string]*RPCConnection
	// authMapping maps UserIDs to their active connections.
	authMapping map[string]map[string]bool
	// mu protects concurrent access to the maps
	mu sync.RWMutex
}

// newRPCConnectionHub creates a new instance of rpcConnectionHub.
// The hub is used internally by RPCNode to manage connections.
func newRPCConnectionHub() *rpcConnectionHub {
	return &rpcConnectionHub{
		connections: make(map[string]*RPCConnection),
		authMapping: make(map[string]map[string]bool),
	}
}

// Set adds or updates a connection in the hub.
// If the connection has a UserID, it also updates the user mapping.
func (hub *rpcConnectionHub) Add(conn *RPCConnection) error {
	connID := conn.ConnectionID()
	userID := conn.UserID()

	hub.mu.Lock()
	defer hub.mu.Unlock()

	// If the connection already exists, remove it first
	if _, exists := hub.connections[connID]; exists {
		return fmt.Errorf("connection with ID %s already exists", connID)
	}

	hub.connections[connID] = conn

	if userID == "" {
		return nil
	}

	// If the connection has a userID, update the auth mapping
	if _, exists := hub.authMapping[userID]; !exists {
		hub.authMapping[userID] = make(map[string]bool)
	}

	// Update the mapping for this user
	hub.authMapping[userID][connID] = true
	return nil
}

// Reauthenticate updates the UserID for an existing connection.
func (hub *rpcConnectionHub) Reauthenticate(connID, userID string) error {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	conn, exists := hub.connections[connID]
	if !exists {
		return fmt.Errorf("connection with ID %s does not exist", connID)
	}

	// Remove the old user mapping if it exists
	oldUserID := conn.UserID()
	if oldUserID != "" {
		if userConns, ok := hub.authMapping[oldUserID]; ok {
			delete(userConns, connID)
			if len(userConns) == 0 {
				delete(hub.authMapping, oldUserID) // Remove user mapping if no connections left
			}
		}
	}

	// Set the new UserID
	conn.SetUserID(userID)

	// Update the auth mapping for the new UserID
	if _, ok := hub.authMapping[userID]; !ok {
		hub.authMapping[userID] = make(map[string]bool)
	}
	hub.authMapping[userID][connID] = true

	return nil
}

// Get retrieves a connection by its connection ID.
// Returns nil if the connection doesn't exist.
func (hub *rpcConnectionHub) Get(connID string) *RPCConnection {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	conn, ok := hub.connections[connID]
	if !ok {
		return nil
	}

	return conn
}

// Remove deletes a connection from the hub.
// It also removes any associated user mapping.
func (hub *rpcConnectionHub) Remove(connID string) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	conn, ok := hub.connections[connID]
	if !ok {
		return
	}

	delete(hub.connections, connID)
	userID := conn.UserID()
	if userID == "" {
		return
	}

	// If the connection has a UserID, remove it from the auth mapping
	if userConns, exists := hub.authMapping[userID]; exists {
		delete(userConns, connID)
		if len(userConns) == 0 {
			delete(hub.authMapping, userID) // Remove user mapping if no connections left
		}
	}
}

// Publish sends a message to a specific authenticated user.
// If the user is not connected, the message is silently dropped.
func (hub *rpcConnectionHub) Publish(userID string, message []byte) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	connIDs, ok := hub.authMapping[userID]
	if !ok {
		return
	}

	// Iterate over all connections for this user and send the message
	for connID := range connIDs {
		conn := hub.connections[connID]
		if conn == nil || conn.writeSink == nil {
			continue // Skip if connection is nil or write sink is not set
		}

		// Write the message to the connection's write sink
		conn.Write(message)
	}
}
