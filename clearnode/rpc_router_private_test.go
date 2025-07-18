package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestRPCRouterHandleGetLedgerBalances(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	t.Cleanup(cleanup)

	participant1 := newTestCommonAddress("0xParticipant1")
	participant1AccountID := NewAccountID(participant1.Hex())
	ledger := GetWalletLedger(router.DB, participant1)
	err := ledger.Record(participant1AccountID, "usdc", decimal.NewFromInt(1000))
	require.NoError(t, err)

	params := map[string]string{"account_id": participant1AccountID.String()}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	c := &RPCContext{
		Context: context.TODO(),
		UserID:  participant1.Hex(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 1,
				Method:    "get_ledger_balances",
				Params:    []any{json.RawMessage(paramsJSON)},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []Signature{Signature([]byte("0xdummySignature"))},
		},
	}

	router.HandleGetLedgerBalances(c)
	res := c.Message.Res
	require.NotNil(t, res)

	require.NotEmpty(t, res.Params)
	balancesArray, ok := res.Params[0].([]Balance)
	require.True(t, ok, "Response should contain an array of Balance")
	require.Equal(t, 1, len(balancesArray), "Should have 1 balance entry")

	expectedAssets := map[string]decimal.Decimal{"usdc": decimal.NewFromInt(1000)}
	for _, balance := range balancesArray {
		expectedBalance, exists := expectedAssets[balance.Asset]
		require.True(t, exists, "Unexpected asset in response: %s", balance.Asset)
		require.Equal(t, expectedBalance, balance.Amount, "Incorrect balance for asset %s", balance.Asset)
		delete(expectedAssets, balance.Asset)
	}
	require.Empty(t, expectedAssets, "Not all expected assets were found")
}

func TestRPCRouterHandleGetRPCHistory(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	t.Cleanup(cleanup)

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

	require.Equal(t, "get_rpc_history", res.Method)
	require.Equal(t, uint64(100), res.RequestID)

	require.Len(t, res.Params, 1, "Response should contain RPCEntry entries")
	rpcHistory, ok := res.Params[0].([]RPCEntry)
	require.True(t, ok, "Response parameter should be a slice of RPCEntry")
	require.Len(t, rpcHistory, 3, "Should return 3 records for the participant")

	require.Equal(t, uint64(3), rpcHistory[0].ReqID, "First record should be the newest")
	require.Equal(t, uint64(2), rpcHistory[1].ReqID, "Second record should be the middle one")
	require.Equal(t, uint64(1), rpcHistory[2].ReqID, "Third record should be the oldest")
}

func TestRPCRouterHandleGetUserTag(t *testing.T) {
	t.Parallel()
	// Create signers
	userKey, _ := crypto.GenerateKey()
	userSigner := Signer{privateKey: userKey}
	userAddr := userSigner.GetAddress().Hex()

	t.Run("Succesfully retrieve the user tag", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: userAddr, Wallet: userAddr,
		}).Error)

		// Setup user tag
		userTag, err := GenerateOrRetrieveUserTag(db, userAddr)
		require.NoError(t, err)

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  userAddr,
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 42,
					Method:    "get_user_tag",
					Params:    []any{},
					Timestamp: ts,
				},
			},
		}

		// 1) Marshal c.Message.Req exactly as a JSON array
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := userSigner.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleGetUserTag(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Verify response
		require.Equal(t, "get_user_tag", res.Method)
		require.Equal(t, uint64(42), res.RequestID)
		// Verify response structure
		getTagResponse, ok := res.Params[0].(GetUserTagResponse)
		require.True(t, ok, "Response should be a GetUserTagResponse")
		require.Equal(t, userTag.Tag, getTagResponse.Tag)
	})
	t.Run("Error when there is no tag", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: userAddr, Wallet: userAddr,
		}).Error)

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  userAddr,
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 42,
					Method:    "get_user_tag",
					Params:    []any{},
					Timestamp: ts,
				},
			},
		}

		// 1) Marshal c.Message.Req exactly as a JSON array
		rawReq, err := json.Marshal(c.Message.Req)
		require.NoError(t, err)
		c.Message.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := userSigner.Sign(rawReq)
		require.NoError(t, err)
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleGetUserTag(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "failed to get user tag")
	})
}
func TestRPCRouterHandleTransfer(t *testing.T) {
	// Create signers
	senderKey, _ := crypto.GenerateKey()
	senderSigner := Signer{privateKey: senderKey}
	senderAddr := senderSigner.GetAddress()
	senderAccountID := NewAccountID(senderAddr.Hex())
	recipientAddr := newTestCommonAddress("0xRecipient")
	recipientAccountID := NewAccountID(recipientAddr.Hex())

	t.Run("SuccessfulTransfer", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr.Hex(), Wallet: senderAddr.Hex(),
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "eth", decimal.NewFromInt(5)))

		// Create transfer parameters
		transferParams := TransferParams{
			Destination: recipientAddr.Hex(),
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
				{AssetSymbol: "eth", Amount: decimal.NewFromInt(2)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Verify response
		require.Equal(t, "transfer", res.Method)
		require.Equal(t, uint64(42), res.RequestID)
		// Verify response structure
		transferResp, ok := res.Params[0].([]TransactionResponse)
		require.True(t, ok, "Response should be a slice of TransactionResponse")
		require.Equal(t, senderAddr.Hex(), transferResp[0].FromAccount)
		require.Equal(t, recipientAddr.Hex(), transferResp[0].ToAccount)
		require.False(t, transferResp[0].CreatedAt.IsZero(), "CreatedAt should be set")

		// Check balances were updated correctly
		// Sender should have 500 USDC and 3 ETH left
		senderUSDC, err := GetWalletLedger(db, senderAddr).Balance(senderAccountID, "usdc")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(500).String(), senderUSDC.String())

		senderETH, err := GetWalletLedger(db, senderAddr).Balance(senderAccountID, "eth")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(3).String(), senderETH.String())

		// Recipient should have 500 USDC and 2 ETH
		recipientUSDC, err := GetWalletLedger(db, recipientAddr).Balance(recipientAccountID, "usdc")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(500).String(), recipientUSDC.String())

		recipientETH, err := GetWalletLedger(db, recipientAddr).Balance(recipientAccountID, "eth")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(2).String(), recipientETH.String())

		// Verify transactions were recorded to the database
		var transactions []LedgerTransaction
		err = db.Where("from_account = ? AND to_account = ?", senderAddr.Hex(), recipientAddr.Hex()).Find(&transactions).Error
		require.NoError(t, err)
		require.Len(t, transactions, 2, "Should have 2 transactions recorded (one for each asset)")

		// Verify transaction details
		for _, tx := range transactions {
			require.Equal(t, TransactionTypeTransfer, tx.Type, "Transaction type should be transfer")
			require.Equal(t, senderAddr.Hex(), tx.FromAccount, "From account should match")
			require.Equal(t, recipientAddr.Hex(), tx.ToAccount, "To account should match")
			require.False(t, tx.CreatedAt.IsZero(), "CreatedAt should be set")

			// Check asset-specific amounts
			if tx.AssetSymbol == "usdc" {
				require.Equal(t, decimal.NewFromInt(500), tx.Amount, "USDC amount should match")
			} else if tx.AssetSymbol == "eth" {
				require.Equal(t, decimal.NewFromInt(2), tx.Amount, "ETH amount should match")
			} else {
				t.Errorf("Unexpected asset symbol: %s", tx.AssetSymbol)
			}
		}

		// Verify response transactions match database transactions
		require.Len(t, transferResp, 2, "Response should contain 2 transaction objects")
		for _, responseTx := range transferResp {
			// Find matching transaction in database
			var dbTx LedgerTransaction
			err = db.Where("id = ?", responseTx.Id).First(&dbTx).Error
			require.NoError(t, err, "Response transaction should exist in database")

			require.Equal(t, dbTx.Type.String(), responseTx.TxType, "Transaction type should match")
			require.Equal(t, dbTx.FromAccount, responseTx.FromAccount, "From account should match")
			require.Equal(t, dbTx.ToAccount, responseTx.ToAccount, "To account should match")
			require.Equal(t, dbTx.AssetSymbol, responseTx.Asset, "Asset should match")
			require.Equal(t, dbTx.Amount, responseTx.Amount, "Amount should match")
		}
	})

	t.Run("Successful Transfer by destination user tag", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr.Hex(), Wallet: senderAddr.Hex(),
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "eth", decimal.NewFromInt(5)))

		// Setup user tag for recipient
		recipientTag, err := GenerateOrRetrieveUserTag(db, recipientAddr.Hex())
		require.NoError(t, err)

		// Create transfer parameters
		transferParams := TransferParams{
			DestinationUserTag: recipientTag.Tag,
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
				{AssetSymbol: "eth", Amount: decimal.NewFromInt(2)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Verify response
		require.Equal(t, "transfer", res.Method)
		require.Equal(t, uint64(42), res.RequestID)
		// Verify response structure
		transactionResponse, ok := res.Params[0].([]TransactionResponse)
		require.True(t, ok, "Response should be a TransactionResponse")

		targetTransaction := transactionResponse[0]

		require.Len(t, transactionResponse, 2, "Should have 2 transaction entries for the transfer")
		require.Equal(t, senderAddr.Hex(), targetTransaction.FromAccount)
		require.Equal(t, recipientAddr.Hex(), targetTransaction.ToAccount)
		require.False(t, targetTransaction.CreatedAt.IsZero(), "CreatedAt should be set")

		// Check balances were updated correctly
		// Sender should have 500 USDC and 3 ETH left
		senderUSDC, err := GetWalletLedger(db, senderAddr).Balance(senderAccountID, "usdc")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(500).String(), senderUSDC.String())

		senderETH, err := GetWalletLedger(db, senderAddr).Balance(senderAccountID, "eth")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(3).String(), senderETH.String())

		// Recipient should have 500 USDC and 2 ETH
		recipientUSDC, err := GetWalletLedger(db, recipientAddr).Balance(recipientAccountID, "usdc")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(500).String(), recipientUSDC.String())

		recipientETH, err := GetWalletLedger(db, recipientAddr).Balance(recipientAccountID, "eth")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(2).String(), recipientETH.String())
	})
	t.Run("ErrorInvalidDestinationAddress", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr.Hex(), Wallet: senderAddr.Hex(),
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))

		// Create transfer with invalid destination
		transferParams := TransferParams{
			Destination: "not-a-valid-address",
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "invalid destination account")
	})

	t.Run("ErrorTransferToSelf", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr.Hex(), Wallet: senderAddr.Hex(),
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))

		// Create transfer to self
		transferParams := TransferParams{
			Destination: senderAddr.Hex(),
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "cannot transfer to self")
	})

	t.Run("ErrorInsufficientFunds", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr.Hex(), Wallet: senderAddr.Hex(),
		}).Error)

		// Fund sender's account with a small amount
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(100)))

		// Create transfer for more than available
		transferParams := TransferParams{
			Destination: recipientAddr.Hex(),
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "insufficient funds")
	})

	t.Run("ErrorEmptyAllocations", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr.Hex(), Wallet: senderAddr.Hex(),
		}).Error)

		// Create transfer with empty allocations
		transferParams := TransferParams{
			Destination: recipientAddr.Hex(),
			Allocations: []TransferAllocation{},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "allocations cannot be empty")
	})

	t.Run("ErrorZeroAmount", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr.Hex(), Wallet: senderAddr.Hex(),
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))

		// Create transfer with zero amount
		transferParams := TransferParams{
			Destination: recipientAddr.Hex(),
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "invalid allocation")
	})

	t.Run("ErrorNegativeAmount", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr.Hex(), Wallet: senderAddr.Hex(),
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))

		// Create transfer with negative amount
		transferParams := TransferParams{
			Destination: recipientAddr.Hex(),
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(-500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "invalid allocation")
	})

	t.Run("ErrorInvalidSignature", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr.Hex(), Wallet: senderAddr.Hex(),
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))

		// Create transfer parameters
		transferParams := TransferParams{
			Destination: recipientAddr.Hex(),
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		c := &RPCContext{
			Context: context.TODO(),
			UserID:  senderAddr.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleTransfer(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "invalid signature")
	})
}

func TestRPCRouterHandleCreateAppSession(t *testing.T) {
	rawA, _ := crypto.GenerateKey()
	rawB, _ := crypto.GenerateKey()
	signerA := Signer{privateKey: rawA}
	signerB := Signer{privateKey: rawB}
	signerAddressA := signerA.GetAddress()
	signerAddressB := signerB.GetAddress()
	accountIDA := NewAccountID(signerAddressA.Hex())
	accountIDB := NewAccountID(signerAddressB.Hex())

	t.Run("SuccessfulCreateAppSession", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		token := "0xTokenXYZ"
		for i, p := range []string{signerAddressA.Hex(), signerAddressB.Hex()} {
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

		require.NoError(t, GetWalletLedger(db, signerAddressA).Record(accountIDA, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, signerAddressB).Record(accountIDB, "usdc", decimal.NewFromInt(200)))

		ts := uint64(time.Now().Unix())
		def := AppDefinition{
			Protocol:           "test-proto",
			ParticipantWallets: []string{signerAddressA.Hex(), signerAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Challenge:          60,
			Nonce:              ts,
		}
		data := `{"state":"initial"}`
		createParams := CreateAppSessionParams{
			Definition: def,
			Allocations: []AppAllocation{
				{ParticipantWallet: signerAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: signerAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
			SessionData: &data,
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
		c.Message.Sig = []Signature{sigA, sigB}

		// Call handler
		router.HandleCreateApplication(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "create_app_session", res.Method)
		appResp, ok := res.Params[0].(AppSessionResponse)
		require.True(t, ok)
		require.Equal(t, string(ChannelStatusOpen), appResp.Status)
		require.Equal(t, uint64(1), appResp.Version)
		require.Empty(t, appResp.SessionData, "session data should not be returned in response")

		var vApp AppSession
		require.NoError(t, db.Where("session_id = ?", appResp.AppSessionID).First(&vApp).Error)
		require.ElementsMatch(t, []string{signerAddressA.Hex(), signerAddressB.Hex()}, vApp.ParticipantWallets)
		require.Equal(t, uint64(1), vApp.Version)
		require.Equal(t, data, vApp.SessionData, "session data should be stored in the database")

		// Participant accounts drained
		partBalA, _ := GetWalletLedger(db, signerAddressA).Balance(accountIDA, "usdc")
		partBalB, _ := GetWalletLedger(db, signerAddressB).Balance(accountIDB, "usdc")
		require.True(t, partBalA.IsZero(), "Participant A balance should be zero")
		require.True(t, partBalB.IsZero(), "Participant B balance should be zero")

		// Virtual-app funded
		sessionAccountID := NewAccountID(appResp.AppSessionID)
		vBalA, _ := GetWalletLedger(db, signerAddressA).Balance(sessionAccountID, "usdc")
		vBalB, _ := GetWalletLedger(db, signerAddressB).Balance(sessionAccountID, "usdc")
		require.Equal(t, decimal.NewFromInt(100).String(), vBalA.String())
		require.Equal(t, decimal.NewFromInt(200).String(), vBalB.String())
	})
	t.Run("ErrorChallengedChannel", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		token := "0xTokenXYZ"
		for i, p := range []string{signerAddressA.Hex(), signerAddressB.Hex()} {
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

		require.NoError(t, GetWalletLedger(db, signerAddressA).Record(accountIDA, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, signerAddressB).Record(accountIDB, "usdc", decimal.NewFromInt(200)))

		ts := uint64(time.Now().Unix())
		def := AppDefinition{
			Protocol:           "test-proto",
			ParticipantWallets: []string{signerAddressA.Hex(), signerAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Challenge:          60,
			Nonce:              ts,
		}
		createParams := CreateAppSessionParams{
			Definition: def,
			Allocations: []AppAllocation{
				{ParticipantWallet: signerAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: signerAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
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
		c.Message.Sig = []Signature{sigA, sigB}

		// Call handler
		router.HandleCreateApplication(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "has challenged channels")
	})
}

func TestRPCRouterHandleSubmitAppState(t *testing.T) {
	t.Run("SuccessfulSubmitAppState", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddressA := signer.GetAddress()
		userAddressB := newTestCommonAddress("0xUserB")

		tokenAddress := "0xToken123"
		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChannelA",
			Participant: userAddressA.Hex(),
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			Nonce:       1,
		}).Error)
		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChannelB",
			Participant: userAddressB.Hex(),
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			Nonce:       1,
		}).Error)

		vAppID := newTestCommonHash("0xVApp123")
		sessionAccountID := NewAccountID(vAppID.Hex())
		require.NoError(t, db.Create(&AppSession{
			SessionID:          vAppID.Hex(),
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			SessionData:        `{"state":"initial"}`,
			Status:             ChannelStatusOpen,
			Challenge:          60,
			Weights:            []int64{100, 0},
			Quorum:             100,
			Version:            1,
		}).Error)

		assetSymbol := "usdc"
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, assetSymbol, decimal.NewFromInt(200)))
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(sessionAccountID, assetSymbol, decimal.NewFromInt(300)))

		data := `{"state":"updated"}`
		submitAppStateParams := SubmitAppStateParams{
			AppSessionID: vAppID.Hex(),
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
			},
			SessionData: &data,
		}

		// Create RPC request
		paramsJSON, err := json.Marshal(submitAppStateParams)
		require.NoError(t, err)
		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 1,
					Method:    "submit_app_state",
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleSubmitAppState(c)
		res := c.Message.Res
		require.NotNil(t, res)

		appResp, ok := res.Params[0].(AppSessionResponse)
		require.True(t, ok)
		require.Equal(t, string(ChannelStatusOpen), appResp.Status)
		require.Equal(t, uint64(2), appResp.Version)
		require.Empty(t, appResp.SessionData, "session data should not be returned in response")

		var updated AppSession
		require.NoError(t, db.Where("session_id = ?", vAppID.Hex()).First(&updated).Error)
		require.Equal(t, ChannelStatusOpen, updated.Status)
		require.Equal(t, uint64(2), updated.Version)
		require.Equal(t, data, updated.SessionData, "session data should be stored in the database")

		// Check balances redistributed
		balA, _ := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "usdc")
		balB, _ := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "usdc")
		require.Equal(t, decimal.NewFromInt(250), balA)
		require.Equal(t, decimal.NewFromInt(250), balB)
	})
}

func TestRPCRouterHandleCloseApplication(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	db := router.DB
	t.Cleanup(cleanup)

	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: rawKey}
	userAddressA := signer.GetAddress()
	userAddressB := newTestCommonAddress("0xParticipantB")
	accountIDA := NewAccountID(userAddressA.Hex())
	accountIDB := NewAccountID(userAddressB.Hex())

	tokenAddress := "0xToken123"
	require.NoError(t, db.Create(&Channel{
		ChannelID:   "0xChannelA",
		Participant: userAddressA.Hex(),
		Status:      ChannelStatusOpen,
		Token:       tokenAddress,
		Nonce:       1,
	}).Error)
	require.NoError(t, db.Create(&Channel{
		ChannelID:   "0xChannelB",
		Participant: userAddressB.Hex(),
		Status:      ChannelStatusOpen,
		Token:       tokenAddress,
		Nonce:       1,
	}).Error)

	vAppID := newTestCommonHash("0xVApp123")
	sessionAccountID := NewAccountID(vAppID.Hex())
	require.NoError(t, db.Create(&AppSession{
		SessionID:          vAppID.Hex(),
		ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
		SessionData:        `{"state":"initial"}`,
		Status:             ChannelStatusOpen,
		Challenge:          60,
		Weights:            []int64{100, 0},
		Quorum:             100,
		Version:            2,
	}).Error)

	assetSymbol := "usdc"
	require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, assetSymbol, decimal.NewFromInt(200)))
	require.NoError(t, GetWalletLedger(db, userAddressB).Record(sessionAccountID, assetSymbol, decimal.NewFromInt(300)))

	data := `{"state":"closed"}`
	closeParams := CloseAppSessionParams{
		AppSessionID: vAppID.Hex(),
		Allocations: []AppAllocation{
			{ParticipantWallet: userAddressA.Hex(), AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
			{ParticipantWallet: userAddressB.Hex(), AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
		},
		SessionData: &data,
	}

	// Create RPC request
	paramsJSON, err := json.Marshal(closeParams)
	require.NoError(t, err)
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
	c.Message.Sig = []Signature{sigBytes}

	// Call handler
	router.HandleCloseApplication(c)
	res := c.Message.Res
	require.NotNil(t, res)

	appResp, ok := res.Params[0].(AppSessionResponse)
	require.True(t, ok)
	require.Equal(t, string(ChannelStatusClosed), appResp.Status)
	require.Equal(t, uint64(3), appResp.Version)
	require.Empty(t, "", appResp.SessionData, "session data should not be returned in response")

	var updated AppSession
	require.NoError(t, db.Where("session_id = ?", vAppID.Hex()).First(&updated).Error)
	require.Equal(t, ChannelStatusClosed, updated.Status)
	require.Equal(t, uint64(3), updated.Version)
	require.Equal(t, data, updated.SessionData, "session data should be stored in the database")

	// Check balances redistributed
	balA, _ := GetWalletLedger(db, userAddressA).Balance(accountIDA, "usdc")
	balB, _ := GetWalletLedger(db, userAddressB).Balance(accountIDB, "usdc")
	require.Equal(t, decimal.NewFromInt(250), balA)
	require.Equal(t, decimal.NewFromInt(250), balB)

	// v-app accounts drained
	vBalA, _ := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "usdc")
	vBalB, _ := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "usdc")
	require.True(t, vBalA.IsZero(), "Participant A vApp balance should be zero")
	require.True(t, vBalB.IsZero(), "Participant B vApp balance should be zero")
}

func TestRPCRouterHandleResizeChannel(t *testing.T) {
	t.Run("SuccessfulAllocation", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		// Create asset
		asset := Asset{Token: "0xTokenResize", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with initial amount 1000
		initialRawAmount := decimal.NewFromInt(1000)
		ch := Channel{
			ChannelID:   "0xChanResize",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger with 1500 USDC (enough for resize)
		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, "usdc", decimal.NewFromInt(1500)))

		// Verify initial balance
		initialBalance, err := ledger.Balance(userAccountID, "usdc")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(1500), initialBalance)

		// Prepare allocation params: allocate 200 to channel (does not change user's total balance)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(200),
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Validate response
		require.Equal(t, "resize_channel", res.Method)
		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok, "Response should be ResizeChannelResponse")
		require.Equal(t, ch.ChannelID, resObj.ChannelID)
		require.Equal(t, ch.Version+1, resObj.Version)

		// New channel amount should be initial + 200
		expected := new(big.Int).Add(initialRawAmount.BigInt(), big.NewInt(200))
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(expected), "Allocated amount mismatch")
		require.Equal(t, 0, resObj.Allocations[1].RawAmount.Cmp(big.NewInt(0)), "Broker allocation should be zero")

		// Verify channel state in database remains unchanged (no update until blockchain confirmation)
		var unchangedChannel Channel
		require.NoError(t, db.Where("channel_id = ?", ch.ChannelID).First(&unchangedChannel).Error)
		require.Equal(t, initialRawAmount, unchangedChannel.RawAmount) // Should remain unchanged
		require.Equal(t, ch.Version, unchangedChannel.Version)         // Should remain unchanged
		require.Equal(t, ChannelStatusOpen, unchangedChannel.Status)

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(userAccountID, "usdc")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(1500), finalBalance) // Should remain unchanged
	})

	t.Run("SuccessfulDeallocation", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		asset := Asset{Token: "0xTokenResize2", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		initialRawAmount := decimal.NewFromInt(1000)
		ch := Channel{
			ChannelID:   "0xChanResize2",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, "usdc", decimal.NewFromInt(500)))

		// Prepare resize params: decrease by 300
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(-300),
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok)

		// Channel amount should decrease
		expected := new(big.Int).Sub(initialRawAmount.BigInt(), big.NewInt(300))
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(expected), "Decreased amount mismatch")

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(userAccountID, "usdc")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(500), finalBalance) // Should remain unchanged
	})

	t.Run("ErrorInvalidChannelID", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		resizeParams := ResizeChannelParams{
			ChannelID:        "0xNonExistentChannel",
			AllocateAmount:   big.NewInt(100),
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "channel 0xNonExistentChannel not found")
	})

	t.Run("ErrorChannelClosed", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		asset := Asset{Token: "0xTokenClosed", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanClosed",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusClosed,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "channel 0xChanClosed is not open: closed")
	})

	t.Run("ErrorChannelJoining", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		asset := Asset{Token: "0xTokenJoining", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanJoining",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusJoining,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "channel 0xChanJoining is not open: joining")
	})

	t.Run("ErrorOtherChallengedChannel", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		asset := Asset{Token: "0xToken", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChanChallenged",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusChallenged,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}).Error)

		ch := Channel{
			ChannelID:   "0xChan",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "has challenged channels")
	})

	t.Run("ErrorInsufficientFunds", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		asset := Asset{Token: "0xTokenInsufficient", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanInsufficient",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund with very small amount (0.000001 USDC), but try to allocate 200 raw units
		// This will create insufficient balance when converted to raw units
		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, "usdc", decimal.NewFromFloat(0.000001)))

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(200),
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "insufficient unified balance")
	})

	t.Run("ErrorZeroAmounts", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		asset := Asset{Token: "0xTokenZero", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanZero",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, "usdc", decimal.NewFromInt(1500)))

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(0),
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Zero allocation should now be rejected as it's a wasteful no-op operation
		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "resize operation requires non-zero ResizeAmount or AllocateAmount")
	})

	t.Run("SuccessfulResizeDeposit", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		asset := Asset{Token: "0xTokenResizeOnly", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		initialRawAmount := decimal.NewFromInt(1000)
		ch := Channel{
			ChannelID:   "0xChanResizeOnly",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund the ledger to pass balance validation
		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, "usdc", decimal.NewFromInt(1500)))

		// Resize operation: deposit 100 into channel (changes user's total balance)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			ResizeAmount:     big.NewInt(100),
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok)

		// Should be initial amount (1000) + allocate amount (0) + resize amount (100) = 1100
		expected := new(big.Int).Add(initialRawAmount.BigInt(), big.NewInt(100))
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(expected))
	})

	t.Run("SuccessfulResizeWithdrawal", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		asset := Asset{Token: "0xTokenResizeOnly", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		initialRawAmount := decimal.NewFromInt(1000)
		ch := Channel{
			ChannelID:   "0xChanResizeOnly",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund the ledger to pass balance validation
		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, "usdc", decimal.NewFromInt(1500)))

		// Resize operation: withdraw 100 from channel (changes user's total balance)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			ResizeAmount:     big.NewInt(-100),
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok)

		// Should be initial amount (1000) + allocate amount (0) - resize amount (100) = 900
		expected := new(big.Int).Add(initialRawAmount.BigInt(), big.NewInt(-100))
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(expected))
	})

	t.Run("ErrorExcessiveDeallocation", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		asset := Asset{Token: "0xTokenExcessive", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanExcessive",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Try to decrease by more than channel amount
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(-1500),
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "new channel amount must be positive")
	})

	t.Run("ErrorInvalidSignature", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		// Create a different signer for invalid signature
		wrongKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		wrongSigner := Signer{privateKey: wrongKey}

		asset := Asset{Token: "0xTokenSig", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanSig",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "invalid signature")
	})

	t.Run("BoundaryLargeAllocation", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		asset := Asset{Token: "0xTokenLarge", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanLarge",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund with a very large amount
		ledger := GetWalletLedger(db, userAddress)
		largeAmount := decimal.NewFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil), 0) // 10^18
		require.NoError(t, ledger.Record(userAccountID, "usdc", largeAmount))

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil), // 10^15
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok)

		// Verify the large allocation was processed correctly
		expectedAmount := new(big.Int).Add(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil))
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(expectedAmount))
	})

	t.Run("SuccessfulAllocationWithResizeDeposit", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		// Create asset
		asset := Asset{Token: "0xTokenMixed", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with initial amount 1000
		initialRawAmount := decimal.NewFromInt(1000)
		ch := Channel{
			ChannelID:   "0xChanMixed",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger with 2000 USDC (enough for both operations)
		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, "usdc", decimal.NewFromInt(2000)))

		// Verify initial balance
		initialBalance, err := ledger.Balance(userAccountID, "usdc")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(2000), initialBalance)

		// Combined operation: allocate 150 to channel + resize (deposit) 100 more
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(150), // Allocation: moves funds from user balance to channel
			ResizeAmount:     big.NewInt(100), // Resize: deposits additional funds into channel
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Validate response
		require.Equal(t, "resize_channel", res.Method)
		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok, "Response should be ResizeChannelResponse")
		require.Equal(t, ch.ChannelID, resObj.ChannelID)
		require.Equal(t, ch.Version+1, resObj.Version)

		// New channel amount should be initial + AllocateAmount + ResizeAmount = 1000 + 150 + 100 = 1250
		expected := new(big.Int).Add(initialRawAmount.BigInt(), big.NewInt(250))
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(expected), "Combined allocation+resize amount mismatch")
		require.Equal(t, 0, resObj.Allocations[1].RawAmount.Cmp(big.NewInt(0)), "Broker allocation should be zero")

		// Verify channel state in database remains unchanged (no update until blockchain confirmation)
		var unchangedChannel Channel
		require.NoError(t, db.Where("channel_id = ?", ch.ChannelID).First(&unchangedChannel).Error)
		require.Equal(t, initialRawAmount, unchangedChannel.RawAmount) // Should remain unchanged
		require.Equal(t, ch.Version, unchangedChannel.Version)         // Should remain unchanged
		require.Equal(t, ChannelStatusOpen, unchangedChannel.Status)

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(userAccountID, "usdc")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(2000), finalBalance) // Should remain unchanged
	})

	t.Run("SuccessfulAllocationWithResizeWithdrawal", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		// Create asset
		asset := Asset{Token: "0xTokenMixed", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with initial amount 0
		initialRawAmount := decimal.NewFromInt(0)
		ch := Channel{
			ChannelID:   "0xChanMixed",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger with 2000 USDC (enough for both operations)
		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, "usdc", decimal.NewFromInt(2000)))

		// Verify initial balance
		initialBalance, err := ledger.Balance(userAccountID, "usdc")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(2000), initialBalance)

		// Combined operation: allocate 150 to channel + resize (deposit) 100 more
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),  // Allocation: moves funds from user balance to channel
			ResizeAmount:     big.NewInt(-100), // Resize: immediately withdraws allocated funds from channel
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleResizeChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Validate response
		require.Equal(t, "resize_channel", res.Method)
		resObj, ok := res.Params[0].(ResizeChannelResponse)
		require.True(t, ok, "Response should be ResizeChannelResponse")
		require.Equal(t, ch.ChannelID, resObj.ChannelID)
		require.Equal(t, ch.Version+1, resObj.Version)

		// New channel amount should be initial + AllocateAmount + ResizeAmount = 0 + 100 - 100 = 0
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(big.NewInt(0)), "Combined allocation+resize amount mismatch")
		require.Equal(t, 0, resObj.Allocations[1].RawAmount.Cmp(big.NewInt(0)), "Broker allocation should be zero")

		// Verify channel state in database remains unchanged (no update until blockchain confirmation)
		var unchangedChannel Channel
		require.NoError(t, db.Where("channel_id = ?", ch.ChannelID).First(&unchangedChannel).Error)
		require.Equal(t, initialRawAmount, unchangedChannel.RawAmount) // Should remain unchanged
		require.Equal(t, ch.Version, unchangedChannel.Version)         // Should remain unchanged
		require.Equal(t, ChannelStatusOpen, unchangedChannel.Status)

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(userAccountID, "usdc")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(2000), finalBalance) // Should remain unchanged
	})
}

func TestRPCRouterHandleCloseChannel(t *testing.T) {
	t.Run("SuccessfulCloseChannel", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		// Create asset
		asset := Asset{Token: "0xTokenClose", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with amount 500
		initialRawAmount := decimal.NewFromInt(500)
		ch := Channel{
			ChannelID:   "0xChanClose",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     2,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger so that raw units match channel.Amount
		require.NoError(t, GetWalletLedger(db, userAddress).Record(
			userAccountID,
			"usdc",
			rawToDecimal(initialRawAmount.BigInt(), asset.Decimals),
		))

		// Prepare close params
		closeParams := CloseChannelParams{
			ChannelID:        ch.ChannelID,
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleCloseChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		// Validate response
		require.Equal(t, "close_channel", res.Method)
		resObj, ok := res.Params[0].(CloseChannelResponse)
		require.True(t, ok, "Response should be CloseChannelResponse")
		require.Equal(t, ch.ChannelID, resObj.ChannelID)
		require.Equal(t, ch.Version+1, resObj.Version)

		// Final allocation should send full balance to destination
		require.Equal(t, 0, resObj.FinalAllocations[0].RawAmount.Cmp(initialRawAmount.BigInt()), "Primary allocation mismatch")
		require.Equal(t, 0, resObj.FinalAllocations[1].RawAmount.Cmp(big.NewInt(0)), "Broker allocation should be zero")
	})
	t.Run("ErrorOtherChallengedChannel", func(t *testing.T) {
		router, cleanup := setupTestRPCRouter(t)
		db := router.DB
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		// Create asset
		asset := Asset{Token: "0xTokenClose", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with amount 500
		initialRawAmount := decimal.NewFromInt(500)

		// Seed other challenged channel
		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChanChallenged",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusChallenged,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     2,
		}).Error)

		ch := Channel{
			ChannelID:   "0xChanClose",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     2,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger so that raw units match channel.Amount
		require.NoError(t, GetWalletLedger(db, userAddress).Record(
			userAccountID,
			"usdc",
			rawToDecimal(initialRawAmount.BigInt(), asset.Decimals),
		))

		// Prepare close params
		closeParams := CloseChannelParams{
			ChannelID:        ch.ChannelID,
			FundsDestination: userAddress.Hex(),
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
		c.Message.Sig = []Signature{sigBytes}

		// Call handler
		router.HandleCloseChannel(c)
		res := c.Message.Res
		require.NotNil(t, res)

		require.Equal(t, "error", res.Method)
		require.Len(t, res.Params, 1)
		require.Contains(t, res.Params[0], "has challenged channels")
	})
}
