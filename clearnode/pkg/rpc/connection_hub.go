package rpc

import (
	"fmt"
	"sync"
)

// ConnectionHub provides centralized management of all active RPC connections.
// It maintains thread-safe mappings between connection IDs and Connection instances,
// as well as user IDs and their associated connections. This enables efficient
// message routing and connection lifecycle management.
//
// Key features:
//   - Thread-safe connection storage and retrieval
//   - User-to-connection mapping for authenticated sessions
//   - Automatic cleanup of auth mappings when connections close
//   - Support for re-authentication (updating user associations)
//   - Broadcast capabilities to all connections for a specific user
type ConnectionHub struct {
	// connections maps connection IDs to RPCConnection instances
	connections map[string]Connection
	// authMapping maps UserIDs to their active connections.
	authMapping map[string]map[string]bool
	// mu protects concurrent access to the maps
	mu sync.RWMutex
}

// NewConnectionHub creates a new ConnectionHub instance with initialized maps.
// The hub is typically used internally by Node implementations to manage
// the lifecycle of all active connections.
func NewConnectionHub() *ConnectionHub {
	return &ConnectionHub{
		connections: make(map[string]Connection),
		authMapping: make(map[string]map[string]bool),
	}
}

// Add registers a new connection with the hub.
// The connection is indexed by its ConnectionID for fast retrieval.
// If the connection has an associated UserID (is authenticated),
// it also updates the user-to-connection mapping.
//
// Returns an error if:
//   - The connection is nil
//   - A connection with the same ID already exists
func (hub *ConnectionHub) Add(conn Connection) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	connID := conn.ConnectionID()
	userID := conn.UserID()

	hub.mu.Lock()
	defer hub.mu.Unlock()

	// If the connection already exists, return an error
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

// Reauthenticate updates the UserID association for an existing connection.
// This method handles the complete re-authentication process:
//   - Removes the connection from the old user's mapping (if any)
//   - Updates the connection's UserID
//   - Adds the connection to the new user's mapping
//
// This is typically called when a user logs in or switches accounts
// on an existing connection.
//
// Returns an error if the specified connection doesn't exist.
func (hub *ConnectionHub) Reauthenticate(connID, userID string) error {
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
				delete(hub.authMapping, oldUserID) // Remove auth mapping if no connections left
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

// Get retrieves a connection by its unique connection ID.
// Returns the Connection instance if found, or nil if no connection
// with the specified ID exists in the hub.
//
// This method is safe for concurrent access.
func (hub *ConnectionHub) Get(connID string) Connection {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	conn, ok := hub.connections[connID]
	if !ok {
		return nil
	}

	return conn
}

// Remove unregisters a connection from the hub.
// This method:
//   - Removes the connection from the main connection map
//   - Cleans up any user-to-connection mappings
//   - Removes empty user entries to prevent memory leaks
//
// If the connection doesn't exist, this method does nothing (no-op).
// This method is safe for concurrent access.
func (hub *ConnectionHub) Remove(connID string) {
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
			delete(hub.authMapping, userID) // Remove auth mapping if no connections left
		}
	}
}

// Publish broadcasts a message to all active connections for a specific user.
// This enables server-initiated notifications to be sent to all of a user's
// connected clients (e.g., multiple browser tabs or devices).
//
// The method:
//   - Looks up all connections associated with the user
//   - Attempts to send the message to each connection
//   - Silently skips any connections that fail to accept the message
//
// If the user has no active connections, the message is silently dropped.
// This method is safe for concurrent access.
func (hub *ConnectionHub) Publish(userID string, response []byte) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	connIDs, ok := hub.authMapping[userID]
	if !ok {
		return
	}

	// Iterate over all connections for this user and send the message
	for connID := range connIDs {
		conn := hub.connections[connID]
		if conn == nil {
			continue // Skip if connection is nil or write sink is not set
		}

		// Write the response to the connection's write sink
		conn.WriteRawResponse(response)
	}
}
