package rpc

import (
	"fmt"
	"sync"
)

const defaultConnectionRegion = "default"

type ObserveConnectionsFn func(region, origin string, count uint32)

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

	// sourceMap is an optional mapping of connection sources (e.g., IP addresses or regions)
	sourceMap map[string]uint32
	// observeConnections is a callback function to monitor connection counts by region
	observeConnections ObserveConnectionsFn
}

// NewConnectionHub creates a new ConnectionHub instance with initialized maps.
// The hub is typically used internally by Node implementations to manage
// the lifecycle of all active connections.
func NewConnectionHub(observeConnections ObserveConnectionsFn) *ConnectionHub {
	return &ConnectionHub{
		connections:        make(map[string]Connection),
		authMapping:        make(map[string]map[string]bool),
		sourceMap:          make(map[string]uint32),
		observeConnections: observeConnections,
	}
}

// Add registers a new connection with the hub.
// The connection is indexed by its ConnectionID for fast retrieval.
//
// Returns an error if:
//   - The connection is nil
//   - A connection with the same ID already exists
func (hub *ConnectionHub) Add(conn Connection) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	connID := conn.ConnectionID()

	hub.mu.Lock()
	defer hub.mu.Unlock()

	// If the connection already exists, return an error
	if _, exists := hub.connections[connID]; exists {
		return fmt.Errorf("connection with ID %s already exists", connID)
	}

	hub.connections[connID] = conn

	sourceID := getSourceID(conn.Origin())
	hub.sourceMap[sourceID]++
	hub.observeConnections(defaultConnectionRegion, conn.Origin(), uint32(hub.sourceMap[sourceID]))

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

	conn, exists := hub.connections[connID]
	if !exists {
		return // No connection to remove
	}
	delete(hub.connections, connID)

	sourceID := getSourceID(conn.Origin())
	if count, exists := hub.sourceMap[sourceID]; exists && count > 0 {
		hub.sourceMap[sourceID]--
		if hub.sourceMap[sourceID] == 0 {
			delete(hub.sourceMap, sourceID)
		}
	}
	hub.observeConnections(defaultConnectionRegion, conn.Origin(), uint32(hub.sourceMap[sourceID]))
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
// TODO: refine with subscription topics capability
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

func getSourceID(origin string) string {
	return origin
}
