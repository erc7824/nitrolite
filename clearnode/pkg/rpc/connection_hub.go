package rpc

import (
	"fmt"
	"sync"
)

// ConnectionHub manages all active WebSocket connections.
// It provides thread-safe operations for connection tracking and auth mapping.
type ConnectionHub struct {
	// connections maps connection IDs to RPCConnection instances
	connections map[string]Connection
	// authMapping maps UserIDs to their active connections.
	authMapping map[string]map[string]bool
	// mu protects concurrent access to the maps
	mu sync.RWMutex
}

// NewConnectionHub creates a new instance of ConnectionHub.
// The hub is used internally by Node to manage connections.
func NewConnectionHub() *ConnectionHub {
	return &ConnectionHub{
		connections: make(map[string]Connection),
		authMapping: make(map[string]map[string]bool),
	}
}

// Add adds a connection to the hub.
// If the connection has a UserID, it also updates the auth mapping.
func (hub *ConnectionHub) Add(conn Connection) error {
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

// Reauthenticate updates the UserID for an existing connection.
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

// Get retrieves a connection by its connection ID.
// Returns nil if the connection doesn't exist.
func (hub *ConnectionHub) Get(connID string) Connection {
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

// Publish sends a response to a specific authenticated user.
// If the user is not connected, the response is silently dropped.
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
			delete(connIDs, connID)
			continue // Skip if connection is nil or write sink is not set
		}

		// Write the response to the connection's write sink
		conn.WriteRawResponse(response)
	}
}
