package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCConnectionWrite(t *testing.T) {
	logger := NewLoggerIPFS("test")

	t.Run("successful write", func(t *testing.T) {
		connID := "conn1"
		userID := "user1"
		writeChan := make(chan []byte, 1)
		processSink := make(chan []byte, 1)
		closeChan := make(chan struct{}, 1)
		conn := &RPCConnection{
			connectionID: connID,
			userID:       userID,
			logger:       logger.With("connectionID", connID),
			writeSink:    writeChan,
			processSink:  processSink,
			closeConnCh:  closeChan,
		}

		message := []byte("test message")
		conn.Write(message)

		select {
		case received := <-writeChan:
			assert.Equal(t, message, received)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("message not received")
		}

		assert.Empty(t, closeChan, "close channel should be empty")
	})

	t.Run("write timeout triggers connection close", func(t *testing.T) {
		connID := "conn1"
		userID := "user1"
		writeChan := make(chan []byte)
		processSink := make(chan []byte, 1)
		closeChan := make(chan struct{}, 1)
		conn := &RPCConnection{
			connectionID: connID,
			userID:       userID,
			logger:       logger.With("connectionID", connID),
			writeSink:    writeChan,
			processSink:  processSink,
			closeConnCh:  closeChan,
		}

		originalTimeout := defaultRPCMessageWriteDuration
		defaultRPCMessageWriteDuration = 50 * time.Millisecond
		defer func() { defaultRPCMessageWriteDuration = originalTimeout }()

		message := []byte("test message")
		conn.Write(message)

		select {
		case <-closeChan:
			// Success - connection close was triggered
		case <-time.After(100 * time.Millisecond):
			t.Fatal("connection close not triggered")
		}
	})
}

func TestRPCConnectionHub(t *testing.T) {
	logger := NewLoggerIPFS("test")
	hub := newRPCConnectionHub()
	var err error

	userID1 := "user1"
	userID2 := "user2"

	// Add connections
	connID1 := "conn1"
	writeChan1 := make(chan []byte, 10)
	processSink1 := make(chan []byte, 10)
	closeChan1 := make(chan struct{}, 1)
	conn1 := &RPCConnection{
		connectionID: connID1,
		userID:       userID1,
		logger:       logger,
		writeSink:    writeChan1,
		processSink:  processSink1,
		closeConnCh:  closeChan1,
	}
	err = hub.Add(conn1)
	require.NoError(t, err)

	connID2 := "conn2"
	writeChan2 := make(chan []byte, 10)
	processSink2 := make(chan []byte, 10)
	closeChan2 := make(chan struct{}, 1)
	conn2 := &RPCConnection{
		connectionID: connID2,
		userID:       userID1, // Same user as conn1
		logger:       logger,
		writeSink:    writeChan2,
		processSink:  processSink2,
		closeConnCh:  closeChan2,
	}
	err = hub.Add(conn2)
	require.NoError(t, err)

	connID3 := "conn3"
	writeChan3 := make(chan []byte, 10)
	processSink3 := make(chan []byte, 10)
	closeChan3 := make(chan struct{}, 1)
	conn3 := &RPCConnection{
		connectionID: connID3,
		userID:       userID2, // Same user as conn1
		logger:       logger,
		writeSink:    writeChan3,
		processSink:  processSink3,
		closeConnCh:  closeChan3,
	}
	err = hub.Add(conn3)
	require.NoError(t, err)

	// Verify connections
	assert.Equal(t, conn1, hub.Get(connID1))
	assert.Equal(t, conn2, hub.Get(connID2))
	assert.Equal(t, conn3, hub.Get(connID3))

	// Publish to user1
	message1 := []byte("message for user1")
	hub.Publish(userID1, message1)

	// Both user1 connections should receive
	require.Equal(t, message1, <-writeChan1)
	require.Equal(t, message1, <-writeChan2)

	// user2 should not receive
	assert.Empty(t, writeChan3)

	// Remove one connection for user1
	hub.Remove(connID1)
	assert.Nil(t, hub.Get(connID1))

	// Publish again
	message2 := []byte("second message")
	hub.Publish(userID1, message2)

	// Only conn2 should receive now
	require.Equal(t, message2, <-writeChan2)
	assert.Empty(t, writeChan1)

	// Remove all connections
	hub.Remove(connID2)
	hub.Remove(connID3)

	assert.Empty(t, hub.connections)
	assert.Empty(t, hub.authMapping)
}
