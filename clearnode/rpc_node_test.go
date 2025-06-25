package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCNode(t *testing.T) {
	// Setup
	// Use a test private key
	privateKeyHex := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	signer, err := NewSigner(privateKeyHex)
	require.NoError(t, err)

	logger := NewLoggerIPFS("root")

	// 1) Create an instance of RPCNode
	node := NewRPCNode(signer, logger)
	require.NotNil(t, node)

	mu := sync.Mutex{}

	rootMwKey := "root_mw_executed"
	rootMethod := "root.test"
	groupAMwKey := "group_a_mw_executed"
	groupMethodA := "group.test1"
	groupBMwKey := "group_b_mw_executed"
	groupMethodB := "group.test2"
	previousExecMethodKey := "previous_exec_method"
	authMethod := "auth.test"

	onConnectMethod := "on_connect.test"
	onConnectCounts := 0
	node.OnConnect(func(send SendRPCMessageFunc) {
		mu.Lock()
		defer mu.Unlock()

		onConnectCounts++
		send(onConnectMethod, onConnectCounts)
	})

	onDisconnectCounts := 0
	disconnectedUserID := ""
	node.OnDisconnect(func(userID string) {
		mu.Lock()
		defer mu.Unlock()

		onDisconnectCounts++
		disconnectedUserID = userID
	})

	onAuthenticatedMethod := "on_authenticated.test"
	onAuthenticatedCounts := 0
	authenticatedUserID := "user.test"
	node.OnAuthenticated(func(userID string, send SendRPCMessageFunc) {
		mu.Lock()
		defer mu.Unlock()

		onAuthenticatedCounts++
		send(onAuthenticatedMethod, onAuthenticatedCounts, userID)
	})

	onMessageSentCounts := 0
	node.OnMessageSent(func() {
		mu.Lock()
		defer mu.Unlock()

		onMessageSentCounts++
	})

	createDummyHandler := func(method string) func(c *RPCContext) {
		return func(c *RPCContext) {
			mu.Lock()
			defer mu.Unlock()

			logger.Debug("executing handler", "method", method)

			var prevMethod string
			if prevMethodVal, ok := c.Storage.Get(previousExecMethodKey); ok {
				prevMethod, ok = prevMethodVal.(string)
				if !ok {
					prevMethod = "non_string"
				}
			}

			var rootMwValue, groupAMwValue, groupBMwValue bool
			if rootMwVal, ok := c.Storage.Get(rootMwKey); ok {
				rootMwValue, ok = rootMwVal.(bool)
				if !ok {
					rootMwValue = false
				}
			}
			if groupMwVal, ok := c.Storage.Get(groupAMwKey); ok {
				groupAMwValue, ok = groupMwVal.(bool)
				if !ok {
					groupAMwValue = false
				}
			}
			if groupMwVal, ok := c.Storage.Get(groupBMwKey); ok {
				groupBMwValue, ok = groupMwVal.(bool)
				if !ok {
					groupBMwValue = false
				}
			}
			c.Succeed(method, c.UserID, prevMethod, rootMwValue, groupAMwValue, groupBMwValue)
			c.Storage.Set(previousExecMethodKey, method)
		}
	}

	// 2) Add one middleware and 2 handlers to the root
	node.Use(func(c *RPCContext) {
		logger.Debug("executing root middleware")

		c.Storage.Set(rootMwKey, true)
		c.Storage.Set(groupAMwKey, false) // Reset group A middleware state
		c.Storage.Set(groupBMwKey, false) // Reset group B middleware state
		c.Next()
	})

	node.Handle(rootMethod, createDummyHandler(rootMethod))
	node.Handle(authMethod, func(c *RPCContext) {
		logger.Debug("executing auth handler")
		c.Succeed(authMethod, authenticatedUserID)
		c.UserID = authenticatedUserID // Simulate authenticated user
	})

	// 3) Add 2 groups with 2 middlewares and 2 handlers
	testGroupA := node.NewGroup("testGroupA")

	testGroupA.Use(func(c *RPCContext) {
		logger.Debug("executing group A middleware")
		c.Storage.Set(groupAMwKey, true)
		c.Storage.Set(groupBMwKey, false)
		c.Next()
	})

	testGroupA.Handle(groupMethodA, createDummyHandler(groupMethodA))

	testGroupB := testGroupA.NewGroup("testGroupB")
	testGroupB.Use(func(c *RPCContext) {
		logger.Debug("executing group B middleware")
		c.Storage.Set(groupBMwKey, true)
		c.Next()
	})

	testGroupB.Handle(groupMethodB, createDummyHandler(groupMethodB))

	// 4) Start server
	server := httptest.NewServer(http.HandlerFunc(node.HandleConnection))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// 5) Call each of methods and verify that they work as expected

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Receive message
	receive := func(t *testing.T) *RPCMessage {
		var respMsg RPCMessage
		err = conn.ReadJSON(&respMsg)
		require.NoError(t, err)

		return &respMsg
	}

	// Helper function to send request and receive response
	sendAndReceive := func(t *testing.T, RequestID uint64, method string, params ...interface{}) *RPCMessage {
		if params == nil {
			params = []interface{}{}
		}
		// Create request
		reqData := &RPCData{
			RequestID: RequestID,
			Method:    method,
			Params:    params,
			Timestamp: uint64(time.Now().UnixMilli()),
		}

		reqMsg := &RPCMessage{
			Req: reqData,
			Sig: []string{},
		}

		// Send request
		err = conn.WriteJSON(reqMsg)
		require.NoError(t, err)

		return receive(t)
	}

	// Test connect
	t.Run("connect", func(t *testing.T) {
		resp := receive(t)

		mu.Lock()
		defer mu.Unlock()

		require.NotNil(t, resp.Res)
		assert.Equal(t, onConnectMethod, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 1)
		assert.Len(t, resp.Sig, 1)
		assert.Equal(t, 1, onConnectCounts)     // on connect counts
		assert.Equal(t, 1, onMessageSentCounts) // number of messages sent
	})

	// Test root handler
	t.Run("root handler", func(t *testing.T) {
		resp := sendAndReceive(t, 1, rootMethod)

		mu.Lock()
		defer mu.Unlock()

		require.NotNil(t, resp.Res)
		assert.Equal(t, rootMethod, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 5)
		assert.Len(t, resp.Sig, 1)
		assert.Equal(t, "", resp.Res.Params[0])    // not authenticated
		assert.Equal(t, "", resp.Res.Params[1])    // previous dummy method empty
		assert.Equal(t, true, resp.Res.Params[2])  // root middleware executed
		assert.Equal(t, false, resp.Res.Params[3]) // group A middleware not executed
		assert.Equal(t, false, resp.Res.Params[4]) // group B middleware not executed
		assert.Equal(t, 2, onMessageSentCounts)    // number of messages sent
	})

	// Test auth handler
	t.Run("auth handler", func(t *testing.T) {
		resp := sendAndReceive(t, 1, authMethod)

		// So we definitely receive both authMethod and onAuthenticatedMethod
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		require.NotNil(t, resp.Res)
		assert.Equal(t, authMethod, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 1)
		assert.Len(t, resp.Sig, 1)
		assert.Equal(t, authenticatedUserID, resp.Res.Params[0]) // authenticated user ID
		assert.Equal(t, 4, onMessageSentCounts)                  // number of messages sent
		mu.Unlock()

		// on authenticated method executed
		resp = receive(t)

		mu.Lock()
		require.NotNil(t, resp.Res)
		assert.Equal(t, onAuthenticatedMethod, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 2)
		assert.Len(t, resp.Sig, 1)
		assert.Equal(t, 1, onAuthenticatedCounts)                // on authenticated counts
		assert.Equal(t, authenticatedUserID, resp.Res.Params[1]) // authenticated user ID
		assert.Equal(t, 4, onMessageSentCounts)                  // number of messages sent
		mu.Unlock()
	})

	// Test group handler 1
	t.Run("group handler 1", func(t *testing.T) {
		resp := sendAndReceive(t, 2, groupMethodA)

		mu.Lock()
		defer mu.Unlock()

		require.NotNil(t, resp.Res)
		assert.Equal(t, groupMethodA, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 5)
		assert.Len(t, resp.Sig, 1)
		assert.Equal(t, authenticatedUserID, resp.Res.Params[0]) // this method
		assert.Equal(t, rootMethod, resp.Res.Params[1])          // previous dummy method root
		assert.Equal(t, true, resp.Res.Params[2])                // root middleware executed
		assert.Equal(t, true, resp.Res.Params[3])                // group A middleware executed
		assert.Equal(t, false, resp.Res.Params[4])               // group B middleware not executed
		assert.Equal(t, 5, onMessageSentCounts)                  // number of messages sent
	})

	// Test group handler 2
	t.Run("group handler 2", func(t *testing.T) {
		resp := sendAndReceive(t, 3, groupMethodB)

		mu.Lock()
		defer mu.Unlock()

		require.NotNil(t, resp.Res)
		assert.Equal(t, groupMethodB, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 5)
		assert.Len(t, resp.Sig, 1)
		assert.Equal(t, authenticatedUserID, resp.Res.Params[0]) // this method
		assert.Equal(t, groupMethodA, resp.Res.Params[1])        // previous dummy method root
		assert.Equal(t, true, resp.Res.Params[2])                // root middleware executed
		assert.Equal(t, true, resp.Res.Params[3])                // group A middleware executed
		assert.Equal(t, true, resp.Res.Params[4])                // group B middleware executed
		assert.Equal(t, 6, onMessageSentCounts)                  // number of messages sent
	})

	// Test unknown method
	t.Run("unknown method", func(t *testing.T) {
		resp := sendAndReceive(t, 4, "unknown.method")

		mu.Lock()
		defer mu.Unlock()

		require.NotNil(t, resp.Res)
		assert.Equal(t, "error", resp.Res.Method)
		assert.Len(t, resp.Res.Params, 1)
		assert.Contains(t, resp.Res.Params[0], "unknown method")
		assert.Equal(t, 7, onMessageSentCounts) // number of messages sent
	})

	// Test invalid message format
	t.Run("invalid message format", func(t *testing.T) {
		// Send invalid JSON
		err := conn.WriteMessage(websocket.TextMessage, []byte("{invalid json"))
		require.NoError(t, err)

		// Read error response
		var respMsg RPCMessage
		err = conn.ReadJSON(&respMsg)
		require.NoError(t, err)

		mu.Lock()
		defer mu.Unlock()

		require.NotNil(t, respMsg.Res)
		assert.Equal(t, "error", respMsg.Res.Method)
		assert.Contains(t, respMsg.Res.Params[0], "invalid message format")
		assert.Equal(t, 8, onMessageSentCounts) // number of messages sent
	})

	// Test disconnect
	t.Run("disconnect", func(t *testing.T) {
		// Close the connection
		err = conn.Close()
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond) // Give some time for the disconnect handler to be called

		mu.Lock()
		defer mu.Unlock()

		// Verify onDisconnect handler was called
		assert.Equal(t, 1, onDisconnectCounts)                   // number of disconnects
		assert.Equal(t, authenticatedUserID, disconnectedUserID) // disconnected user ID
		assert.Equal(t, 8, onMessageSentCounts)                  // number of messages sent
	})
}

func TestRPCConnectionWrite(t *testing.T) {
	t.Run("successful write", func(t *testing.T) {
		writeChan := make(chan []byte, 1)
		closeChan := make(chan struct{}, 1)
		conn := NewRPCConnection("conn1", "user1", writeChan, closeChan)

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
		writeChan := make(chan []byte)
		closeChan := make(chan struct{}, 1)
		conn := NewRPCConnection("conn1", "user1", writeChan, closeChan)

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
	hub := newRPCConnectionHub()

	// Add connections
	writeChan1 := make(chan []byte, 10)
	closeChan1 := make(chan struct{}, 1)
	conn1 := NewRPCConnection("conn1", "user1", writeChan1, closeChan1)
	hub.Set(conn1)

	writeChan2 := make(chan []byte, 10)
	closeChan2 := make(chan struct{}, 1)
	conn2 := NewRPCConnection("conn2", "user1", writeChan2, closeChan2)
	hub.Set(conn2)

	writeChan3 := make(chan []byte, 10)
	closeChan3 := make(chan struct{}, 1)
	conn3 := NewRPCConnection("conn3", "user2", writeChan3, closeChan3)
	hub.Set(conn3)

	// Verify connections
	assert.Equal(t, conn1, hub.Get("conn1"))
	assert.Equal(t, conn2, hub.Get("conn2"))
	assert.Equal(t, conn3, hub.Get("conn3"))

	// Publish to user1
	message1 := []byte("message for user1")
	hub.Publish("user1", message1)

	// Both user1 connections should receive
	require.Equal(t, message1, <-writeChan1)
	require.Equal(t, message1, <-writeChan2)

	// user2 should not receive
	assert.Empty(t, writeChan3)

	// Remove one connection for user1
	hub.Remove("conn1")
	assert.Nil(t, hub.Get("conn1"))

	// Publish again
	message2 := []byte("second message")
	hub.Publish("user1", message2)

	// Only conn2 should receive now
	require.Equal(t, message2, <-writeChan2)
	assert.Empty(t, writeChan1)

	// Remove all connections
	hub.Remove("conn2")
	hub.Remove("conn3")

	assert.Empty(t, hub.connections)
	assert.Empty(t, hub.authMapping)
}
