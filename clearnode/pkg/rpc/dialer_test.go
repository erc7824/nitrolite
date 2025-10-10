package rpc_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebsocketDialer_BasicConnection(t *testing.T) {
	t.Parallel()

	// Create mock server
	server := createEchoServer(t, nil)
	defer server.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := rpc.DefaultWebsocketDialerConfig
	cfg.EventChanSize = 10
	dialer := rpc.NewWebsocketDialer(cfg)

	errorCh := connectDialer(t, ctx, dialer, server.Listener.Addr().String())

	// Test basic call
	params, err := rpc.NewParams(map[string]interface{}{"key": "value"})
	require.NoError(t, err)
	req := rpc.NewRequest(rpc.NewPayload(1, "test", params))
	resp, err := dialer.Call(ctx, &req)
	require.NoError(t, err)
	assert.Equal(t, "response_test", resp.Res.Method)
	assert.Equal(t, req.Req.RequestID, resp.Res.RequestID)

	// Ensure no errors occurred
	select {
	case err := <-errorCh:
		require.NoError(t, err)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestWebsocketDialer_ConnectionFailure(t *testing.T) {
	t.Parallel()

	cfg := rpc.DefaultWebsocketDialerConfig
	dialer := rpc.NewWebsocketDialer(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errorCh := make(chan error, 1)
	err := dialer.Dial(ctx, "ws://invalid-url-that-does-not-exist:12345", func(err error) {
		errorCh <- err
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error dialing websocket server")
	assert.False(t, dialer.IsConnected())
}

func TestWebsocketDialer_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Create a server that accepts connections but doesn't respond
	server := createEchoServer(t, nil)
	defer server.Close()

	cfg := rpc.DefaultWebsocketDialerConfig
	dialer := rpc.NewWebsocketDialer(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	errorCh := connectDialer(t, ctx, dialer, server.Listener.Addr().String())
	cancel() // Cancel the context to trigger closure

	// Verify not connected after cancel
	time.Sleep(100 * time.Millisecond)
	assert.False(t, dialer.IsConnected())

	// Ensure no errors occurred
	select {
	case err := <-errorCh:
		require.NoError(t, err)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestWebsocketDialer_MultipleRequests(t *testing.T) {
	t.Parallel()

	server := createEchoServer(t, nil)
	defer server.Close()

	ctx := context.Background()
	cfg := rpc.DefaultWebsocketDialerConfig
	cfg.EventChanSize = 10
	dialer := rpc.NewWebsocketDialer(cfg)

	errorCh := connectDialer(t, ctx, dialer, server.Listener.Addr().String())

	// Send multiple requests concurrently
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			params, err := rpc.NewParams(map[string]interface{}{
				"index": idx,
			})
			require.NoError(t, err)
			req := rpc.NewRequest(rpc.NewPayload(uint64(idx), "test", params))

			callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			resp, err := dialer.Call(callCtx, &req)
			require.NoError(t, err)

			assert.Equal(t, uint64(idx), resp.Res.RequestID)
			assert.Equal(t, "response_test", resp.Res.Method)
		}(i)
	}

	wg.Wait()

	// Ensure no errors occurred
	select {
	case err := <-errorCh:
		require.NoError(t, err)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestWebsocketDialer_RequestTimeout(t *testing.T) {
	t.Parallel()

	// Create server that delays responses
	extraHandlers := map[string]func(*rpc.Request) *rpc.Response{
		"slow_request": func(req *rpc.Request) *rpc.Response {
			time.Sleep(10 * time.Second) // Delay response
			return &rpc.Response{
				Res: rpc.Payload{
					RequestID: req.Req.RequestID,
					Method:    "response_slow_request",
					Timestamp: uint64(time.Now().UnixMilli()),
				},
			}
		},
	}
	server := createEchoServer(t, extraHandlers)
	defer server.Close()

	ctx := context.Background()
	cfg := rpc.DefaultWebsocketDialerConfig
	dialer := rpc.NewWebsocketDialer(cfg)

	errorCh := connectDialer(t, ctx, dialer, server.Listener.Addr().String())

	// Test timeout
	var params rpc.Params
	req := rpc.NewRequest(rpc.NewPayload(1, "slow_request", params))

	callCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	_, err := dialer.Call(callCtx, &req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no response received")

	// Ensure no errors occurred
	select {
	case err := <-errorCh:
		require.NoError(t, err)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestWebsocketDialer_UnsolicitedEvents(t *testing.T) {
	t.Parallel()

	// Create server that sends unsolicited events
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()

		// Send unsolicited event immediately
		params, err := rpc.NewParams(map[string]interface{}{"type": "notification"})
		require.NoError(t, err)

		event := rpc.Response{
			Res: rpc.Payload{
				RequestID: 9999,
				Method:    "unsolicited_event",
				Params:    params,
				Timestamp: uint64(time.Now().UnixMilli()),
			},
		}
		eventJSON, _ := json.Marshal(event)
		conn.WriteMessage(websocket.TextMessage, eventJSON)

		// Then handle requests normally
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req rpc.Request
			json.Unmarshal(msg, &req)

			method := req.Req.Method
			if method == "ping" {
				method = "pong"
			}

			resp := rpc.Response{
				Res: rpc.Payload{
					RequestID: req.Req.RequestID,
					Method:    method,
					Timestamp: uint64(time.Now().UnixMilli()),
				},
			}
			respJSON, _ := json.Marshal(resp)
			conn.WriteMessage(websocket.TextMessage, respJSON)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	cfg := rpc.DefaultWebsocketDialerConfig
	cfg.EventChanSize = 10
	dialer := rpc.NewWebsocketDialer(cfg)

	errorCh := connectDialer(t, ctx, dialer, server.Listener.Addr().String())

	// Check event channel for unsolicited event
	select {
	case event := <-dialer.EventCh():
		assert.NotNil(t, event)
		assert.Equal(t, "unsolicited_event", event.Res.Method)
		assert.Equal(t, uint64(9999), event.Res.RequestID)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for unsolicited event")
	}

	// Ensure no errors occurred
	select {
	case err := <-errorCh:
		require.NoError(t, err)
	case <-time.After(100 * time.Millisecond):
	}
}

// Helper functions

func createEchoServer(t *testing.T, extraHandlers map[string]func(*rpc.Request) *rpc.Response) *httptest.Server {
	if extraHandlers == nil {
		extraHandlers = make(map[string]func(*rpc.Request) *rpc.Response)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req rpc.Request
			err = json.Unmarshal(msg, &req)
			if err != nil {
				continue
			}

			method := req.Req.Method
			var res *rpc.Response
			if handler, exists := extraHandlers[method]; exists {
				res = handler(&req)
			} else {
				if method == "ping" {
					method = "pong"
				} else {
					method = "response_" + method
				}

				resp := rpc.NewResponse(rpc.NewPayload(req.Req.RequestID, method, req.Req.Params))
				res = &resp
			}

			respJSON, err := json.Marshal(res)
			require.NoError(t, err)
			conn.WriteMessage(websocket.TextMessage, respJSON)
		}
	}))
}

func connectDialer(t *testing.T, ctx context.Context, dialer *rpc.WebsocketDialer, addr string) <-chan error {
	errorCh := make(chan error, 1)

	err := dialer.Dial(ctx, "ws://"+addr, func(err error) {
		if err != nil {
			errorCh <- err
		}
	})
	require.NoError(t, err)
	require.True(t, dialer.IsConnected())

	return errorCh
}
