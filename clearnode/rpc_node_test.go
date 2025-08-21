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

	paramsUserIDKey := "userID"
	paramsPrevMethodKey := "prev"
	paramsRootMwKey := "rootMw"
	paramsGroupAMwKey := "groupA"
	paramsGroupBMwKey := "groupB"
	paramsErrorKey := "error"
	paramsOnConnectCounts := "onConnectCounts"
	paramsOnAuthCounts := "onAuthCounts"

	onConnectMethod := "on_connect.test"
	onConnectCounts := 0
	node.OnConnect(func(send SendRPCMessageFunc) {
		mu.Lock()
		defer mu.Unlock()

		onConnectCounts++
		params := map[string]any{
			paramsOnConnectCounts: onConnectCounts,
		}
		send(onConnectMethod, params)
	})

	onDisconnectSignal := newTestSignal()
	disconnectedUserID := ""
	node.OnDisconnect(func(userID string) {
		mu.Lock()
		defer mu.Unlock()

		disconnectedUserID = userID
		onDisconnectSignal.trigger()
	})

	onAuthenticatedMethod := "on_authenticated.test"
	onAuthenticatedCounts := 0
	authenticatedUserID := "user.test"
	node.OnAuthenticated(func(userID string, send SendRPCMessageFunc) {
		mu.Lock()
		defer mu.Unlock()

		onAuthenticatedCounts++
		params := map[string]any{
			paramsOnAuthCounts: onAuthenticatedCounts,
			paramsUserIDKey:    userID,
		}
		send(onAuthenticatedMethod, params)
	})

	onMessageSentSignal := newTestSignal()
	node.OnMessageSent(func() {
		mu.Lock()
		defer mu.Unlock()

		onMessageSentSignal.trigger()
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
			params := map[string]any{
				paramsUserIDKey:     c.UserID,
				paramsPrevMethodKey: prevMethod,
				paramsRootMwKey:     rootMwValue,
				paramsGroupAMwKey:   groupAMwValue,
				paramsGroupBMwKey:   groupBMwValue,
			}
			c.Succeed(method, params)
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
		params := map[string]any{
			paramsUserIDKey: authenticatedUserID,
		}
		c.Succeed(authMethod, params)
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
	sendAndReceive := func(t *testing.T, RequestID uint64, method string, params RPCDataParams) *RPCMessage {
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
			Sig: []Signature{},
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
		assert.Equal(t, 1, onConnectCounts)         // on connect counts
		assert.True(t, onMessageSentSignal.await()) // on message sent signal
	})

	// Test root handler
	t.Run("root handler", func(t *testing.T) {
		resp := sendAndReceive(t, 1, rootMethod, nil)

		mu.Lock()
		defer mu.Unlock()

		require.NotNil(t, resp.Res)
		assert.Equal(t, rootMethod, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 5)
		assert.Len(t, resp.Sig, 1)

		params, ok := resp.Res.Params.(map[string]any)
		require.True(t, ok, "params should be a map[string]any")

		assert.Equal(t, "", params[paramsUserIDKey])      // not authenticated
		assert.Equal(t, "", params[paramsPrevMethodKey])  // previous dummy method empty
		assert.Equal(t, true, params[paramsRootMwKey])    // root middleware executed
		assert.Equal(t, false, params[paramsGroupAMwKey]) // group A middleware not executed
		assert.Equal(t, false, params[paramsGroupBMwKey]) // group B middleware not executed
		assert.True(t, onMessageSentSignal.await())       // on message sent signal
	})

	// Test auth handler
	t.Run("auth handler", func(t *testing.T) {
		resp := sendAndReceive(t, 1, authMethod, nil)

		// So we definitely receive both authMethod and onAuthenticatedMethod
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		require.NotNil(t, resp.Res)
		assert.Equal(t, authMethod, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 1)
		assert.Len(t, resp.Sig, 1)

		params, ok := resp.Res.Params.(map[string]any)
		require.True(t, ok, "params should be a map[string]any")

		assert.Equal(t, authenticatedUserID, params[paramsUserIDKey]) // authenticated user ID
		assert.True(t, onMessageSentSignal.await())                   // on message sent signal
		mu.Unlock()

		// on authenticated method executed
		resp = receive(t)

		mu.Lock()
		require.NotNil(t, resp.Res)
		assert.Equal(t, onAuthenticatedMethod, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 2)
		assert.Len(t, resp.Sig, 1)
		assert.Equal(t, 1, onAuthenticatedCounts) // on authenticated counts

		params, ok = resp.Res.Params.(map[string]any)
		require.True(t, ok, "params should be a map[string]any")

		assert.Equal(t, authenticatedUserID, params[paramsUserIDKey]) // authenticated user ID
		assert.True(t, onMessageSentSignal.await())                   // on message sent signal

		mu.Unlock()
	})

	// Test group handler 1
	t.Run("group handler 1", func(t *testing.T) {
		resp := sendAndReceive(t, 2, groupMethodA, nil)

		mu.Lock()
		defer mu.Unlock()

		require.NotNil(t, resp.Res)
		assert.Equal(t, groupMethodA, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 5)
		assert.Len(t, resp.Sig, 1)

		params, ok := resp.Res.Params.(map[string]any)
		require.True(t, ok, "params should be a map[string]any")

		assert.Equal(t, authenticatedUserID, params[paramsUserIDKey]) // this method
		assert.Equal(t, rootMethod, params[paramsPrevMethodKey])      // previous dummy method root
		assert.Equal(t, true, params[paramsRootMwKey])                // root middleware executed
		assert.Equal(t, true, params[paramsGroupAMwKey])              // group A middleware executed
		assert.Equal(t, false, params[paramsGroupBMwKey])             // group B middleware not executed
		assert.True(t, onMessageSentSignal.await())                   // on message sent signal
	})

	// Test group handler 2
	t.Run("group handler 2", func(t *testing.T) {
		resp := sendAndReceive(t, 3, groupMethodB, nil)

		mu.Lock()
		defer mu.Unlock()

		require.NotNil(t, resp.Res)
		assert.Equal(t, groupMethodB, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 5)
		assert.Len(t, resp.Sig, 1)

		params, ok := resp.Res.Params.(map[string]any)
		require.True(t, ok, "params should be a map[string]any")

		assert.Equal(t, authenticatedUserID, params[paramsUserIDKey]) // this method
		assert.Equal(t, groupMethodA, params[paramsPrevMethodKey])    // previous dummy method root
		assert.Equal(t, true, params[paramsRootMwKey])                // root middleware executed
		assert.Equal(t, true, params[paramsGroupAMwKey])              // group A middleware executed
		assert.Equal(t, true, params[paramsGroupBMwKey])              // group B middleware executed
		assert.True(t, onMessageSentSignal.await())                   // on message sent signal
	})

	// Test unknown method
	t.Run("unknown method", func(t *testing.T) {
		resp := sendAndReceive(t, 4, "unknown.method", nil)

		mu.Lock()
		defer mu.Unlock()

		require.NotNil(t, resp.Res)
		assert.Equal(t, "error", resp.Res.Method)
		assert.Len(t, resp.Res.Params, 1)

		params, ok := resp.Res.Params.(map[string]any)
		require.True(t, ok, "params should be a map[string]any")

		assert.Contains(t, params[paramsErrorKey], "unknown method")
		assert.True(t, onMessageSentSignal.await()) // on message sent signal
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

		params, ok := respMsg.Res.Params.(map[string]any)
		require.True(t, ok, "params should be a map[string]any")

		assert.Contains(t, params[paramsErrorKey], "invalid message format")
		assert.True(t, onMessageSentSignal.await()) // on message sent signal
	})

	// Test disconnect
	t.Run("disconnect", func(t *testing.T) {
		// Close the connection
		err = conn.Close()
		require.NoError(t, err)

		// Verify onDisconnect handler was called
		assert.True(t, onDisconnectSignal.await()) // on disconnect signal

		mu.Lock()
		defer mu.Unlock()

		assert.Equal(t, authenticatedUserID, disconnectedUserID) // disconnected user ID
		assert.False(t, onMessageSentSignal.await())             // on message sent signal should not be called after disconnect
	})
}

type testSignal struct {
	signalCh chan struct{}
}

func newTestSignal() *testSignal {
	return &testSignal{
		signalCh: make(chan struct{}, 5), // Buffered channel to avoid blocking
	}
}

func (ts *testSignal) trigger() {
	ts.signalCh <- struct{}{}
}

func (ts *testSignal) await() bool {
	maxSignalWait := 500 * time.Millisecond

	select {
	case <-ts.signalCh:
		return true
	case <-time.After(maxSignalWait):
		return false
	}
}
