package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRPCRouter(t *testing.T) (*RPCRouter, func()) {
	db, dbCleanup := setupTestDB(t)

	// Use a test private key
	privateKeyHex := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	signer, err := NewSigner(privateKeyHex)
	require.NoError(t, err)

	logger := NewLoggerIPFS("root.test")

	// Create an instance of RPCRouter
	router := &RPCRouter{
		Signer: signer,
		DB:     db,
		lg:     logger.NewSystem("rpc-router"),
	}

	return router, func() {
		dbCleanup()
	}
}

func TestRPCRouterHandlePing(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	c := &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 1,
				Method:    "ping",
				Params:    []any{nil},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	router.HandlePing(c)
	res := c.Message.Res
	require.NotNil(t, res)

	assert.Equal(t, "pong", res.Method)
}
