package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCRouterHandleGetLedgerBalances(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	ledger := GetWalletLedger(router.DB, "0xParticipant1")
	err := ledger.Record("0xParticipant1", "usdc", decimal.NewFromInt(1000))
	require.NoError(t, err)

	params := map[string]string{"account_id": "0xParticipant1"}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	c := &RPCContext{
		Context: context.TODO(),
		UserID:  "0xParticipant1",
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 1,
				Method:    "get_ledger_balances",
				Params:    []any{json.RawMessage(paramsJSON)},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	router.HandleGetLedgerBalances(c)
	res := c.Message.Res
	require.NotNil(t, res)

	require.NotEmpty(t, res.Params)
	balancesArray, ok := res.Params[0].([]Balance)
	require.True(t, ok, "Response should contain an array of Balance")
	assert.Equal(t, 1, len(balancesArray), "Should have 1 balance entry")

	expectedAssets := map[string]decimal.Decimal{"usdc": decimal.NewFromInt(1000)}
	for _, balance := range balancesArray {
		expectedBalance, exists := expectedAssets[balance.Asset]
		assert.True(t, exists, "Unexpected asset in response: %s", balance.Asset)
		assert.Equal(t, expectedBalance, balance.Amount, "Incorrect balance for asset %s", balance.Asset)
		delete(expectedAssets, balance.Asset)
	}
	assert.Empty(t, expectedAssets, "Not all expected assets were found")
}

func TestRPCRouterHandleGetRPCHistory(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: rawKey}
	participantAddr := signer.GetAddress().Hex()
	rpcStore := NewRPCStore(router.DB)
	timestamp := uint64(time.Now().Unix())

	router.RPCStore = rpcStore
	records := []RPCRecord{
		{
			Sender:    participantAddr,
			ReqID:     1,
			Method:    "ping",
			Params:    []byte(`[null]`),
			Timestamp: timestamp - 3600,
			ReqSig:    []string{"sig1"},
			Response:  []byte(`{"res":[1,"pong",[],1621234567890]}`),
			ResSig:    []string{},
		},
		{
			Sender:    participantAddr,
			ReqID:     2,
			Method:    "get_config",
			Params:    []byte(`[]`),
			Timestamp: timestamp - 1800,
			ReqSig:    []string{"sig2"},
			Response:  []byte(`{"res":[2,"get_config",[{"broker_address":"0xBroker"}],1621234597890]}`),
			ResSig:    []string{},
		},
		{
			Sender:    participantAddr,
			ReqID:     3,
			Method:    "get_channels",
			Params:    []byte(fmt.Sprintf(`[{"participant":"%s"}]`, participantAddr)),
			Timestamp: timestamp - 900,
			ReqSig:    []string{"sig3"},
			Response:  []byte(`{"res":[3,"get_channels",[[]],1621234627890]}`),
			ResSig:    []string{},
		},
	}

	for _, record := range records {
		require.NoError(t, router.DB.Create(&record).Error)
	}

	otherRecord := RPCRecord{
		Sender:    "0xOtherParticipant",
		ReqID:     4,
		Method:    "ping",
		Params:    []byte(`[null]`),
		Timestamp: timestamp,
		ReqSig:    []string{"sig4"},
		Response:  []byte(`{"res":[4,"pong",[],1621234657890]}`),
		ResSig:    []string{},
	}
	require.NoError(t, router.DB.Create(&otherRecord).Error)

	c := &RPCContext{
		Context: context.TODO(),
		UserID:  participantAddr,
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 100,
				Method:    "get_rpc_history",
				Params:    []any{},
				Timestamp: timestamp,
			},
		},
	}

	// Call handler
	router.HandleGetRPCHistory(c)
	res := c.Message.Res
	require.NotNil(t, res)

	assert.Equal(t, "get_rpc_history", res.Method)
	assert.Equal(t, uint64(100), res.RequestID)

	require.Len(t, res.Params, 1, "Response should contain RPCEntry entries")
	rpcHistory, ok := res.Params[0].([]RPCEntry)
	require.True(t, ok, "Response parameter should be a slice of RPCEntry")
	assert.Len(t, rpcHistory, 3, "Should return 3 records for the participant")

	assert.Equal(t, uint64(3), rpcHistory[0].ReqID, "First record should be the newest")
	assert.Equal(t, uint64(2), rpcHistory[1].ReqID, "Second record should be the middle one")
	assert.Equal(t, uint64(1), rpcHistory[2].ReqID, "Third record should be the oldest")
}
