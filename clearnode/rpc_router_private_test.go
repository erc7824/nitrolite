package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
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

func TestRPCRouterHandleTransfer(t *testing.T) {
	// Create signers
	senderKey, _ := crypto.GenerateKey()
	senderSigner := Signer{privateKey: senderKey}
	senderAddr := senderSigner.GetAddress().Hex()
	recipientAddr := "0x" + strings.Repeat("1", 40) // Valid ethereum address with 1s

	t.Run("SuccessfulTransfer", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(1000)))
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "eth", decimal.NewFromInt(5)))

		// Create transfer parameters
		transferParams := Transfer{
			Destination: recipientAddr,
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
				{AssetSymbol: "eth", Amount: decimal.NewFromInt(2)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr,
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 42,
					Method:    "transfer",
					Params:    []any{transferParams},
					Timestamp: ts,
				},
			},
		}

		// 1) Marshal c.Message.Req exactly as a JSON array
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Verify response
		assert.Equal(t, "transfer", res.Method)
		assert.Equal(t, uint64(42), res.RequestID)
		// Verify response structure
		transferResp, ok := res.Params[0].(TransferResponse)
		require.True(t, ok, "Response should be a TransferResponse")
		assert.Equal(t, senderAddr, transferResp.From)
		assert.Equal(t, recipientAddr, transferResp.To)
		assert.False(t, transferResp.CreatedAt.IsZero(), "CreatedAt should be set")

		// Check balances were updated correctly
		// Sender should have 500 USDC and 3 ETH left
		senderUSDC, err := GetWalletLedger(db, senderAddr).Balance(senderAddr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(500).String(), senderUSDC.String())

		senderETH, err := GetWalletLedger(db, senderAddr).Balance(senderAddr, "eth")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(3).String(), senderETH.String())

		// Recipient should have 500 USDC and 2 ETH
		recipientUSDC, err := GetWalletLedger(db, recipientAddr).Balance(recipientAddr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(500).String(), recipientUSDC.String())

		recipientETH, err := GetWalletLedger(db, recipientAddr).Balance(recipientAddr, "eth")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(2).String(), recipientETH.String())
	})

	t.Run("ErrorInvalidDestinationAddress", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(1000)))

		// Create transfer with invalid destination
		transferParams := Transfer{
			Destination: "not-a-valid-address",
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr,
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 43,
					Method:    "transfer",
					Params:    []any{transferParams},
					Timestamp: ts,
				},
			},
		}

		// 1) Marshal c.Message.Req exactly as a JSON array
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "invalid destination account")
	})

	t.Run("ErrorTransferToSelf", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(1000)))

		// Create transfer to self
		transferParams := Transfer{
			Destination: senderAddr,
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr,
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 44,
					Method:    "transfer",
					Params:    []any{transferParams},
					Timestamp: ts,
				},
			},
		}

		// 1) Marshal c.Message.Req exactly as a JSON array
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "invalid destination")
	})

	t.Run("ErrorInsufficientFunds", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account with a small amount
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(100)))

		// Create transfer for more than available
		transferParams := Transfer{
			Destination: recipientAddr,
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr,
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 45,
					Method:    "transfer",
					Params:    []any{transferParams},
					Timestamp: ts,
				},
			},
		}

		// 1) Marshal c.Message.Req exactly as a JSON array
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "insufficient funds")
	})

	t.Run("ErrorEmptyAllocations", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Create transfer with empty allocations
		transferParams := Transfer{
			Destination: recipientAddr,
			Allocations: []TransferAllocation{},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr,
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 46,
					Method:    "transfer",
					Params:    []any{transferParams},
					Timestamp: ts,
				},
			},
		}

		// 1) Marshal c.Message.Req exactly as a JSON array
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "empty allocations")
	})

	t.Run("ErrorZeroAmount", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(1000)))

		// Create transfer with zero amount
		transferParams := Transfer{
			Destination: recipientAddr,
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr,
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 49,
					Method:    "transfer",
					Params:    []any{transferParams},
					Timestamp: ts,
				},
			},
		}

		// 1) Marshal c.Message.Req exactly as a JSON array
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "invalid allocation")
	})

	t.Run("ErrorNegativeAmount", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(1000)))

		// Create transfer with negative amount
		transferParams := Transfer{
			Destination: recipientAddr,
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(-500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr,
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 47,
					Method:    "transfer",
					Params:    []any{transferParams},
					Timestamp: ts,
				},
			},
		}

		// 1) Marshal c.Message.Req exactly as a JSON array
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "invalid allocation")
	})

	t.Run("ErrorInvalidSignature", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(1000)))

		// Create transfer parameters
		transferParams := Transfer{
			Destination: recipientAddr,
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr,
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 48,
					Method:    "transfer",
					Params:    []any{transferParams},
					Timestamp: ts,
				},
			},
		}

		// 1) Marshal c.Message.Req exactly as a JSON array
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq with wrong key
		wrongKey, _ := crypto.GenerateKey()
		wrongSigner := Signer{privateKey: wrongKey}
		sigBytes, err := wrongSigner.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "invalid signature")
	})
}

func TestRPCRouterHandleCreateAppSession(t *testing.T) {
	rawA, _ := crypto.GenerateKey()
	rawB, _ := crypto.GenerateKey()
	signerA := Signer{privateKey: rawA}
	signerB := Signer{privateKey: rawB}
	addrA := signerA.GetAddress().Hex()
	addrB := signerB.GetAddress().Hex()
	t.Run("SuccessfulCreateAppSession", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		token := "0xTokenXYZ"
		for i, p := range []string{addrA, addrB} {
			ch := &Channel{
				ChannelID:   fmt.Sprintf("0xChannel%c", 'A'+i),
				Wallet:      p,
				Participant: p,
				Status:      ChannelStatusOpen,
				Token:       token,
				Nonce:       1,
			}
			require.NoError(t, db.Create(ch).Error)
			require.NoError(t, db.Create(&SignerWallet{
				Signer: p, Wallet: p,
			}).Error)
		}

		require.NoError(t, GetWalletLedger(db, addrA).Record(addrA, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, addrB).Record(addrB, "usdc", decimal.NewFromInt(200)))

		ts := uint64(time.Now().Unix())
		def := AppDefinition{
			Protocol:           "test-proto",
			ParticipantWallets: []string{addrA, addrB},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Challenge:          60,
			Nonce:              ts,
		}
		createParams := CreateAppSessionParams{
			Definition: def,
			Allocations: []AppAllocation{
				{ParticipantWallet: addrA, AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: addrB, AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
		}

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 42,
					Method:    "create_app_session",
					Params:    []any{createParams},
					Timestamp: ts,
				},
			},
		}

		// 1) Marshal c.Message.Req exactly as a JSON array
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq with both participants
		sigA, err := signerA.Sign(rawReq)
		require.NoError(t, err)
		sigB, err := signerB.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigA), hexutil.Encode(sigB)}

		// Call handler
		router.HandleCreateApplication(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "create_app_session", res.Method)
		appResp, ok := res.Params[0].(AppSessionResponse)
		require.True(t, ok)
		assert.Equal(t, string(ChannelStatusOpen), appResp.Status)
		assert.Equal(t, uint64(1), appResp.Version)

		var vApp AppSession
		require.NoError(t, db.Where("session_id = ?", appResp.AppSessionID).First(&vApp).Error)
		assert.ElementsMatch(t, []string{addrA, addrB}, vApp.ParticipantWallets)
		assert.Equal(t, uint64(1), vApp.Version)

		// Participant accounts drained
		partBalA, _ := GetWalletLedger(db, addrA).Balance(addrA, "usdc")
		partBalB, _ := GetWalletLedger(db, addrB).Balance(addrB, "usdc")
		assert.True(t, partBalA.IsZero(), "Participant A balance should be zero")
		assert.True(t, partBalB.IsZero(), "Participant B balance should be zero")

		// Virtual-app funded
		vBalA, _ := GetWalletLedger(db, addrA).Balance(appResp.AppSessionID, "usdc")
		vBalB, _ := GetWalletLedger(db, addrB).Balance(appResp.AppSessionID, "usdc")
		assert.Equal(t, decimal.NewFromInt(100).String(), vBalA.String())
		assert.Equal(t, decimal.NewFromInt(200).String(), vBalB.String())
	})
	t.Run("ErrorChallengedChannel", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		token := "0xTokenXYZ"
		for i, p := range []string{addrA, addrB} {
			ch := &Channel{
				ChannelID:   fmt.Sprintf("0xChannel%c", 'A'+i),
				Wallet:      p,
				Participant: p,
				Status:      ChannelStatusChallenged,
				Token:       token,
				Nonce:       1,
			}
			require.NoError(t, db.Create(ch).Error)
			require.NoError(t, db.Create(&SignerWallet{
				Signer: p, Wallet: p,
			}).Error)
		}

		require.NoError(t, GetWalletLedger(db, addrA).Record(addrA, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, addrB).Record(addrB, "usdc", decimal.NewFromInt(200)))

		ts := uint64(time.Now().Unix())
		def := AppDefinition{
			Protocol:           "test-proto",
			ParticipantWallets: []string{addrA, addrB},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Challenge:          60,
			Nonce:              ts,
		}
		createParams := CreateAppSessionParams{
			Definition: def,
			Allocations: []AppAllocation{
				{ParticipantWallet: addrA, AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: addrB, AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
		}

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 42,
					Method:    "create_app_session",
					Params:    []any{createParams},
					Timestamp: ts,
				},
			},
		}

		// 1) Marshal c.Message.Req exactly as a JSON array
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq with both participants
		sigA, err := signerA.Sign(rawReq)
		require.NoError(t, err)
		sigB, err := signerB.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigA), hexutil.Encode(sigB)}

		// Call handler
		router.HandleCreateApplication(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "has challenged channels")
	})
}

func TestRPCRouterHandleSubmitState(t *testing.T) {
	t.Run("SuccessfulSubmitState", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		participantA := signer.GetAddress().Hex()
		participantB := "0xParticipantB"

		tokenAddress := "0xToken123"
		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChannelA",
			Participant: participantA,
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			Nonce:       1,
		}).Error)
		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChannelB",
			Participant: participantB,
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			Nonce:       1,
		}).Error)

		vAppID := "0xVApp123"
		require.NoError(t, db.Create(&AppSession{
			SessionID:          vAppID,
			ParticipantWallets: []string{participantA, participantB},
			Status:             ChannelStatusOpen,
			Challenge:          60,
			Weights:            []int64{100, 0},
			Quorum:             100,
			Version:            1,
		}).Error)

		assetSymbol := "usdc"
		require.NoError(t, GetWalletLedger(db, participantA).Record(vAppID, assetSymbol, decimal.NewFromInt(200)))
		require.NoError(t, GetWalletLedger(db, participantB).Record(vAppID, assetSymbol, decimal.NewFromInt(300)))

		submitStateParams := SubmitStateParams{
			AppSessionID: vAppID,
			Allocations: []AppAllocation{
				{ParticipantWallet: participantA, AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
				{ParticipantWallet: participantB, AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
			},
		}

		// Create RPC request
		paramsJSON, _ := json.Marshal(submitStateParams)
		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 1,
					Method:    "submit_state",
					Params:    []any{json.RawMessage(paramsJSON)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get the exact raw bytes of [request_id, method, params, timestamp]
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleSubmitState(c)
		res := c.Message.Res
		require.NotNil(t, res)

		appResp, ok := res.Params[0].(AppSessionResponse)
		require.True(t, ok)
		assert.Equal(t, string(ChannelStatusOpen), appResp.Status)
		assert.Equal(t, uint64(2), appResp.Version)

		var updated AppSession
		require.NoError(t, db.Where("session_id = ?", vAppID).First(&updated).Error)
		assert.Equal(t, ChannelStatusOpen, updated.Status)
		assert.Equal(t, uint64(2), updated.Version)

		// Check balances redistributed
		balA, _ := GetWalletLedger(db, participantA).Balance(vAppID, "usdc")
		balB, _ := GetWalletLedger(db, participantB).Balance(vAppID, "usdc")
		assert.Equal(t, decimal.NewFromInt(250), balA)
		assert.Equal(t, decimal.NewFromInt(250), balB)
	})
}

func TestRPCRouterHandleCloseApplication(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	db := router.DB
	defer cleanup()

	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: rawKey}
	participantA := signer.GetAddress().Hex()
	participantB := "0xParticipantB"

	tokenAddress := "0xToken123"
	require.NoError(t, db.Create(&Channel{
		ChannelID:   "0xChannelA",
		Participant: participantA,
		Status:      ChannelStatusOpen,
		Token:       tokenAddress,
		Nonce:       1,
	}).Error)
	require.NoError(t, db.Create(&Channel{
		ChannelID:   "0xChannelB",
		Participant: participantB,
		Status:      ChannelStatusOpen,
		Token:       tokenAddress,
		Nonce:       1,
	}).Error)

	vAppID := "0xVApp123"
	require.NoError(t, db.Create(&AppSession{
		SessionID:          vAppID,
		ParticipantWallets: []string{participantA, participantB},
		Status:             ChannelStatusOpen,
		Challenge:          60,
		Weights:            []int64{100, 0},
		Quorum:             100,
		Version:            2,
	}).Error)

	assetSymbol := "usdc"
	require.NoError(t, GetWalletLedger(db, participantA).Record(vAppID, assetSymbol, decimal.NewFromInt(200)))
	require.NoError(t, GetWalletLedger(db, participantB).Record(vAppID, assetSymbol, decimal.NewFromInt(300)))

	closeParams := CloseAppSessionParams{
		AppSessionID: vAppID,
		Allocations: []AppAllocation{
			{ParticipantWallet: participantA, AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
			{ParticipantWallet: participantB, AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
		},
	}

	// Create RPC request
	paramsJSON, _ := json.Marshal(closeParams)
	c := &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 1,
				Method:    "close_app_session",
				Params:    []any{json.RawMessage(paramsJSON)},
				Timestamp: uint64(time.Now().Unix()),
			},
		},
	}

	// 1) Marshal c.Message.Req to get the exact raw bytes of [request_id, method, params, timestamp]
	rawReq, err := json.Marshal(c.Message.Req)
	require.NoError(t, err)
	c.Message.Req.rawBytes = rawReq

	// 2) Sign rawReq directly
	sigBytes, err := signer.Sign(rawReq)
	require.NoError(t, err)
	c.Message.Sig = []string{hexutil.Encode(sigBytes)}

	// Call handler
	router.HandleCloseApplication(c)
	res := c.Message.Res
	require.NotNil(t, res)

	appResp, ok := res.Params[0].(AppSessionResponse)
	require.True(t, ok)
	assert.Equal(t, string(ChannelStatusClosed), appResp.Status)
	assert.Equal(t, uint64(3), appResp.Version)

	var updated AppSession
	require.NoError(t, db.Where("session_id = ?", vAppID).First(&updated).Error)
	assert.Equal(t, ChannelStatusClosed, updated.Status)
	assert.Equal(t, uint64(3), updated.Version)

	// Check balances redistributed
	balA, _ := GetWalletLedger(db, participantA).Balance(participantA, "usdc")
	balB, _ := GetWalletLedger(db, participantB).Balance(participantB, "usdc")
	assert.Equal(t, decimal.NewFromInt(250), balA)
	assert.Equal(t, decimal.NewFromInt(250), balB)

	// v-app accounts drained
	vBalA, _ := GetWalletLedger(db, participantA).Balance(vAppID, "usdc")
	vBalB, _ := GetWalletLedger(db, participantB).Balance(vAppID, "usdc")
	assert.True(t, vBalA.IsZero(), "Participant A vApp balance should be zero")
	assert.True(t, vBalB.IsZero(), "Participant B vApp balance should be zero")
}

func TestRPCRouterHandleResizeChannel(t *testing.T) {
	t.Run("SuccessfulAllocation", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		// Create asset
		asset := Asset{Token: "0xTokenResize", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with initial amount 1000
		initialAmount := uint64(1000)
		ch := Channel{
			ChannelID:   "0xChanResize",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger with 1500 USDC (enough for resize)
		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(1500)))

		// Verify initial balance
		initialBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(1500), initialBalance)

		// Prepare allocation params: allocate 200 to channel (does not change user's total balance)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(200),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 1,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Validate response
		assert.Equal(t, "resize_channel", res.Method)
		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok, "Response should be ResizeChannelResponse")
		assert.Equal(t, ch.ChannelID, resObj.ChannelID)
		assert.Equal(t, ch.Version+1, resObj.Version)

		// New channel amount should be initial + 200
		expected := new(big.Int).Add(new(big.Int).SetUint64(initialAmount), big.NewInt(200))
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(expected), "Allocated amount mismatch")
		assert.Equal(t, 0, resObj.Allocations[1].Amount.Cmp(big.NewInt(0)), "Broker allocation should be zero")

		// Verify channel state in database remains unchanged (no update until blockchain confirmation)
		var unchangedChannel Channel
		require.NoError(t, db.Where("channel_id = ?", ch.ChannelID).First(&unchangedChannel).Error)
		assert.Equal(t, initialAmount, unchangedChannel.Amount) // Should remain unchanged
		assert.Equal(t, ch.Version, unchangedChannel.Version)   // Should remain unchanged
		assert.Equal(t, ChannelStatusOpen, unchangedChannel.Status)

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(1500), finalBalance) // Should remain unchanged
	})

	t.Run("SuccessfulDeallocation", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		asset := Asset{Token: "0xTokenResize2", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		initialAmount := uint64(1000)
		ch := Channel{
			ChannelID:   "0xChanResize2",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(500)))

		// Prepare resize params: decrease by 300
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(-300),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 2,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok)

		// Channel amount should decrease
		expected := new(big.Int).Sub(new(big.Int).SetUint64(initialAmount), big.NewInt(300))
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(expected), "Decreased amount mismatch")

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(500), finalBalance) // Should remain unchanged
	})

	t.Run("ErrorInvalidChannelID", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		resizeParams := ResizeChannelParams{
			ChannelID:        "0xNonExistentChannel",
			AllocateAmount:   big.NewInt(100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 3,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "channel 0xNonExistentChannel not found")
	})

	t.Run("ErrorChannelClosed", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		asset := Asset{Token: "0xTokenClosed", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanClosed",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusClosed,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 4,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "channel 0xChanClosed is not open: closed")
	})

	t.Run("ErrorChannelJoining", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		asset := Asset{Token: "0xTokenJoining", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanJoining",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusJoining,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 10,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "channel 0xChanJoining is not open: joining")
	})

	t.Run("ErrorOtherChallengedChannel", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		asset := Asset{Token: "0xToken", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChanChallenged",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusChallenged,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}).Error)

		ch := Channel{
			ChannelID:   "0xChan",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 10,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "has challenged channels")
	})

	t.Run("ErrorInsufficientFunds", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		asset := Asset{Token: "0xTokenInsufficient", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanInsufficient",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund with very small amount (0.000001 USDC), but try to allocate 200 raw units
		// This will create insufficient balance when converted to raw units
		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromFloat(0.000001)))

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(200),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 5,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "insufficient unified balance")
	})

	t.Run("ErrorZeroAmounts", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		asset := Asset{Token: "0xTokenZero", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanZero",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(1500)))

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(0),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 6,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Zero allocation should now be rejected as it's a wasteful no-op operation
		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "resize operation requires non-zero ResizeAmount or AllocateAmount")
	})

	t.Run("SuccessfulResizeDeposit", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		asset := Asset{Token: "0xTokenResizeOnly", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		initialAmount := uint64(1000)
		ch := Channel{
			ChannelID:   "0xChanResizeOnly",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund the ledger to pass balance validation
		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(1500)))

		// Resize operation: deposit 100 into channel (changes user's total balance)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			ResizeAmount:     big.NewInt(100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 11,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok)

		// Should be initial amount (1000) + allocate amount (0) + resize amount (100) = 1100
		expected := new(big.Int).Add(new(big.Int).SetUint64(initialAmount), big.NewInt(100))
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(expected))
	})

	t.Run("SuccessfulResizeWithdrawal", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		asset := Asset{Token: "0xTokenResizeOnly", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		initialAmount := uint64(1000)
		ch := Channel{
			ChannelID:   "0xChanResizeOnly",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund the ledger to pass balance validation
		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(1500)))

		// Resize operation: withdraw 100 from channel (changes user's total balance)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			ResizeAmount:     big.NewInt(-100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 11,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok)

		// Should be initial amount (1000) + allocate amount (0) - resize amount (100) = 900
		expected := new(big.Int).Add(new(big.Int).SetUint64(initialAmount), big.NewInt(-100))
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(expected))
	})

	t.Run("ErrorExcessiveDeallocation", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		asset := Asset{Token: "0xTokenExcessive", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanExcessive",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Try to decrease by more than channel amount
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(-1500),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 7,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "new channel amount must be positive")
	})

	t.Run("ErrorInvalidSignature", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		// Create a different signer for invalid signature
		wrongKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		wrongSigner := Signer{privateKey: wrongKey}

		asset := Asset{Token: "0xTokenSig", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanSig",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 8,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly with wrong signer
		sigBytes, err := wrongSigner.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "invalid signature")
	})

	t.Run("BoundaryLargeAllocation", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		asset := Asset{Token: "0xTokenLarge", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanLarge",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund with a very large amount
		ledger := GetWalletLedger(db, addr)
		largeAmount := decimal.NewFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil), 0) // 10^18
		require.NoError(t, ledger.Record(addr, "usdc", largeAmount))

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil), // 10^15
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 9,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok)

		// Verify the large allocation was processed correctly
		expectedAmount := new(big.Int).Add(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil))
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(expectedAmount))
	})

	t.Run("SuccessfulAllocationWithResizeDeposit", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		// Create asset
		asset := Asset{Token: "0xTokenMixed", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with initial amount 1000
		initialAmount := uint64(1000)
		ch := Channel{
			ChannelID:   "0xChanMixed",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger with 2000 USDC (enough for both operations)
		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(2000)))

		// Verify initial balance
		initialBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(2000), initialBalance)

		// Combined operation: allocate 150 to channel + resize (deposit) 100 more
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(150), // Allocation: moves funds from user balance to channel
			ResizeAmount:     big.NewInt(100), // Resize: deposits additional funds into channel
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 12,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Validate response
		assert.Equal(t, "resize_channel", res.Method)
		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok, "Response should be ResizeChannelResponse")
		assert.Equal(t, ch.ChannelID, resObj.ChannelID)
		assert.Equal(t, ch.Version+1, resObj.Version)

		// New channel amount should be initial + AllocateAmount + ResizeAmount = 1000 + 150 + 100 = 1250
		expected := new(big.Int).Add(new(big.Int).SetUint64(initialAmount), big.NewInt(250))
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(expected), "Combined allocation+resize amount mismatch")
		assert.Equal(t, 0, resObj.Allocations[1].Amount.Cmp(big.NewInt(0)), "Broker allocation should be zero")

		// Verify channel state in database remains unchanged (no update until blockchain confirmation)
		var unchangedChannel Channel
		require.NoError(t, db.Where("channel_id = ?", ch.ChannelID).First(&unchangedChannel).Error)
		assert.Equal(t, initialAmount, unchangedChannel.Amount) // Should remain unchanged
		assert.Equal(t, ch.Version, unchangedChannel.Version)   // Should remain unchanged
		assert.Equal(t, ChannelStatusOpen, unchangedChannel.Status)

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(2000), finalBalance) // Should remain unchanged
	})

	t.Run("SuccessfulAllocationWithResizeWithdrawal", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		// Create asset
		asset := Asset{Token: "0xTokenMixed", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with initial amount 0
		initialAmount := uint64(0)
		ch := Channel{
			ChannelID:   "0xChanMixed",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger with 2000 USDC (enough for both operations)
		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(2000)))

		// Verify initial balance
		initialBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(2000), initialBalance)

		// Combined operation: allocate 150 to channel + resize (deposit) 100 more
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),  // Allocation: moves funds from user balance to channel
			ResizeAmount:     big.NewInt(-100), // Resize: immediately withdraws allocated funds from channel
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 12,
					Method:    "resize_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Validate response
		assert.Equal(t, "resize_channel", res.Method)
		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok, "Response should be ResizeChannelResponse")
		assert.Equal(t, ch.ChannelID, resObj.ChannelID)
		assert.Equal(t, ch.Version+1, resObj.Version)

		// New channel amount should be initial + AllocateAmount + ResizeAmount = 0 + 100 - 100 = 0
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(big.NewInt(0)), "Combined allocation+resize amount mismatch")
		assert.Equal(t, 0, resObj.Allocations[1].Amount.Cmp(big.NewInt(0)), "Broker allocation should be zero")

		// Verify channel state in database remains unchanged (no update until blockchain confirmation)
		var unchangedChannel Channel
		require.NoError(t, db.Where("channel_id = ?", ch.ChannelID).First(&unchangedChannel).Error)
		assert.Equal(t, initialAmount, unchangedChannel.Amount) // Should remain unchanged
		assert.Equal(t, ch.Version, unchangedChannel.Version)   // Should remain unchanged
		assert.Equal(t, ChannelStatusOpen, unchangedChannel.Status)

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(2000), finalBalance) // Should remain unchanged
	})
}

func TestRPCRouterHandleCloseChannel(t *testing.T) {
	t.Run("SuccessfulCloseChannel", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		// Create asset
		asset := Asset{Token: "0xTokenClose", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with amount 500
		initialAmount := uint64(500)
		ch := Channel{
			ChannelID:   "0xChanClose",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     2,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger so that raw units match channel.Amount
		require.NoError(t, GetWalletLedger(db, addr).Record(
			addr,
			"usdc",
			decimal.NewFromBigInt(big.NewInt(int64(initialAmount)), -int32(asset.Decimals)),
		))

		// Prepare close params
		closeParams := CloseChannelParams{
			ChannelID:        ch.ChannelID,
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(closeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 10,
					Method:    "close_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleCloseChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Validate response
		assert.Equal(t, "close_channel", res.Method)
		resObj, ok := res.Params[0].(CloseChannelResponse)
		require.True(t, ok, "Response should be CloseChannelResponse")
		assert.Equal(t, ch.ChannelID, resObj.ChannelID)
		assert.Equal(t, ch.Version+1, resObj.Version)

		// Final allocation should send full balance to destination
		assert.Equal(t, 0, resObj.FinalAllocations[0].Amount.Cmp(new(big.Int).SetUint64(initialAmount)), "Primary allocation mismatch")
		assert.Equal(t, 0, resObj.FinalAllocations[1].Amount.Cmp(big.NewInt(0)), "Broker allocation should be zero")
	})
	t.Run("ErrorOtherChallengedChannel", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		// Create asset
		asset := Asset{Token: "0xTokenClose", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with amount 500
		initialAmount := uint64(500)

		// Seed other challenged channel
		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChanChallenged",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusChallenged,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     2,
		}).Error)

		ch := Channel{
			ChannelID:   "0xChanClose",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     2,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger so that raw units match channel.Amount
		require.NoError(t, GetWalletLedger(db, addr).Record(
			addr,
			"usdc",
			decimal.NewFromBigInt(big.NewInt(int64(initialAmount)), -int32(asset.Decimals)),
		))

		// Prepare close params
		closeParams := CloseChannelParams{
			ChannelID:        ch.ChannelID,
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(closeParams)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 10,
					Method:    "close_channel",
					Params:    []any{json.RawMessage(paramsBytes)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// 1) Marshal c.Message.Req to get rawReq
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		router.HandleCloseChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		assert.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		assert.Contains(t, res.Params[0], "has challenged channels")
	})
}
