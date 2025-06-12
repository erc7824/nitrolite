package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

	rootMwKey := "root_mw_executed"
	rootMethod := "root.test"
	groupAMwKey := "group_a_mw_executed"
	groupMethodA := "group.test1"
	groupBMwKey := "group_b_mw_executed"
	groupMethodB := "group.test2"
	previousExecMethodKey := "previous_exec_method"

	createDummyHandler := func(method string) func(c *RPCContext) {
		return func(c *RPCContext) {
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
			c.Succeed(prevMethod, method, rootMwValue, groupAMwValue, groupBMwValue)
			c.Storage.Set(previousExecMethodKey, method)
		}
	}

	// 2) Add one middleware and one handler to the root
	node.Use(func(c *RPCContext) {
		logger.Debug("executing root middleware")

		c.Storage.Set(rootMwKey, true)
		c.Storage.Set(groupAMwKey, false) // Reset group A middleware state
		c.Storage.Set(groupBMwKey, false) // Reset group B middleware state
		c.Next()
	})

	node.Handle(rootMethod, createDummyHandler(rootMethod))

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

		// Read response
		var respMsg RPCMessage
		err = conn.ReadJSON(&respMsg)
		require.NoError(t, err)

		fmt.Printf("Received response: %+v\n", respMsg.Res.Params)
		return &respMsg
	}

	// Test root handler
	t.Run("root handler", func(t *testing.T) {
		resp := sendAndReceive(t, 1, rootMethod)

		require.NotNil(t, resp.Res)
		assert.Equal(t, rootMethod, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 5)
		assert.Len(t, resp.Sig, 1)
		assert.Equal(t, "", resp.Res.Params[0])         // previous method empty
		assert.Equal(t, rootMethod, resp.Res.Params[1]) // this method
		assert.Equal(t, true, resp.Res.Params[2])       // root middleware executed
		assert.Equal(t, false, resp.Res.Params[3])      // group A middleware not executed
		assert.Equal(t, false, resp.Res.Params[4])      // group B middleware not executed
	})

	// Test group handler 1
	t.Run("group handler 1", func(t *testing.T) {
		resp := sendAndReceive(t, 2, groupMethodA)

		require.NotNil(t, resp.Res)
		assert.Equal(t, groupMethodA, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 5)
		assert.Len(t, resp.Sig, 1)
		assert.Equal(t, rootMethod, resp.Res.Params[0])   // previous method empty
		assert.Equal(t, groupMethodA, resp.Res.Params[1]) // this method
		assert.Equal(t, true, resp.Res.Params[2])         // root middleware executed
		assert.Equal(t, true, resp.Res.Params[3])         // group A middleware executed
		assert.Equal(t, false, resp.Res.Params[4])        // group B middleware not executed
	})

	// Test group handler 2
	t.Run("group handler 2", func(t *testing.T) {
		resp := sendAndReceive(t, 3, groupMethodB)

		require.NotNil(t, resp.Res)
		assert.Equal(t, groupMethodB, resp.Res.Method)
		assert.Len(t, resp.Res.Params, 5)
		assert.Len(t, resp.Sig, 1)
		assert.Equal(t, groupMethodA, resp.Res.Params[0]) // previous method empty
		assert.Equal(t, groupMethodB, resp.Res.Params[1]) // this method
		assert.Equal(t, true, resp.Res.Params[2])         // root middleware executed
		assert.Equal(t, true, resp.Res.Params[3])         // group A middleware executed
		assert.Equal(t, true, resp.Res.Params[4])         // group B middleware executed
	})

	// Test unknown method
	t.Run("unknown method", func(t *testing.T) {
		resp := sendAndReceive(t, 4, "unknown.method")

		require.NotNil(t, resp.Res)
		assert.Equal(t, "error", resp.Res.Method)
		assert.Len(t, resp.Res.Params, 1)
		assert.Contains(t, resp.Res.Params[0], "unknown method")
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

		require.NotNil(t, respMsg.Res)
		assert.Equal(t, "error", respMsg.Res.Method)
		assert.Contains(t, respMsg.Res.Params[0], "invalid message format")
	})
}
