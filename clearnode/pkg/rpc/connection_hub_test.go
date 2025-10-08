package rpc_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
)

func TestConnectionHub(t *testing.T) {
	t.Parallel()

	hub := rpc.NewConnectionHub()
	var err error

	userID1 := "user1"
	userID2 := "user2"

	// Add connections
	connID1 := "conn1"
	conn1 := newMockConnection(connID1, userID1)
	err = hub.Add(conn1)
	require.NoError(t, err)

	connID2 := "conn2"
	conn2 := newMockConnection(connID2, userID1)
	err = hub.Add(conn2)
	require.NoError(t, err)

	connID3 := "conn3"
	conn3 := newMockConnection(connID3, userID2)
	err = hub.Add(conn3)
	require.NoError(t, err)

	err = hub.Add(conn1) // Duplicate
	require.Equal(t, "connection with ID conn1 already exists", err.Error())

	// Verify connections
	assert.Equal(t, conn1, hub.Get(connID1))
	assert.Equal(t, conn2, hub.Get(connID2))
	assert.Equal(t, conn3, hub.Get(connID3))

	// Publish to user1
	message1 := []byte("message for user1")
	hub.Publish(userID1, message1)

	// Both user1 connections should receive
	require.Equal(t, message1, conn1.getLastResponse())
	require.Equal(t, message1, conn2.getLastResponse())

	// user2 should not receive
	assert.Empty(t, conn3.getLastResponse())

	// Remove one connection for user1
	hub.Remove(connID1)
	assert.Nil(t, hub.Get(connID1))

	// Reauthenticate conn2 to user2
	conn2.SetUserID(userID2)
	err = hub.Reauthenticate(connID2, userID2)
	require.NoError(t, err)

	// Reauthenticate non-existent connection
	err = hub.Reauthenticate("nonexistent", "userX")
	require.Equal(t, "connection with ID nonexistent does not exist", err.Error())

	// Publish again
	message2 := []byte("second message")
	hub.Publish(userID2, message2)

	// Only conn2 should receive now
	require.Equal(t, message2, conn2.getLastResponse())
	require.Equal(t, message2, conn3.getLastResponse())
	require.NotEqual(t, message2, conn1.getLastResponse())

	// Remove all connections
	hub.Remove(connID2)
	hub.Remove(connID3)
}

type mockConnection struct {
	connectionID string
	userID       string

	rawRequests   chan []byte
	lastResponse  []byte
	handleClosure func(error)
	mu            sync.RWMutex
}

func newMockConnection(connID, userID string) *mockConnection {
	return &mockConnection{
		connectionID: connID,
		userID:       userID,
		rawRequests:  make(chan []byte, 10),
	}
}

func (mc *mockConnection) ConnectionID() string {
	return mc.connectionID
}

func (mc *mockConnection) UserID() string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.userID
}

func (mc *mockConnection) SetUserID(userID string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.userID = userID
}

func (mc *mockConnection) RawRequests() <-chan []byte {
	return mc.rawRequests
}

func (mc *mockConnection) WriteRawResponse(response []byte) bool {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.lastResponse = response
	return true
}

func (mc *mockConnection) getLastResponse() []byte {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.lastResponse
}

func (mc *mockConnection) Serve(_ context.Context, handleClosure func(error)) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.handleClosure = handleClosure
}

func (mc *mockConnection) closeWithError(err error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.handleClosure != nil {
		mc.handleClosure(err)
	}
}
