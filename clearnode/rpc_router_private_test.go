package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createSignedRPCContext(id int, method string, params any, signers ...Signer) *RPCContext {
	ctx := createRPCContext(id, method, params)

	if len(signers) > 0 {
		ctx.UserID = signers[0].GetAddress().Hex()
	}

	rawReq, _ := json.Marshal(ctx.Message.Req)
	ctx.Message.Req.rawBytes = rawReq

	ctx.Message.Sig = make([]Signature, 0, len(signers))
	for _, signer := range signers {
		sigBytes, _ := signer.Sign(rawReq)
		ctx.Message.Sig = append(ctx.Message.Sig, sigBytes)
	}

	return ctx
}

func assertResponse(t *testing.T, ctx *RPCContext, expectedMethod string) *RPCData {
	res := ctx.Message.Res
	require.NotNil(t, res)
	require.Equal(t, expectedMethod, res.Method)
	return res
}

func assertErrorResponse(t *testing.T, ctx *RPCContext, expectedContains string) {
	res := ctx.Message.Res
	require.NotNil(t, res)
	require.Equal(t, "error", res.Method)

	errorParams, ok := res.Params.(ErrorResponse)
	require.True(t, ok, "Response parameter should be an ErrorResponse")

	require.Contains(t, errorParams.Error, expectedContains)
}

func TestRPCRouterHandleGetRPCHistory(t *testing.T) {
	key, _ := crypto.GenerateKey()
	userSigner := Signer{privateKey: key}
	userAddress := userSigner.GetAddress().Hex()

	t.Run("Succesfully retrieve rpc history", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		router.RPCStore = NewRPCStore(db)

		baseTime := uint64(time.Now().Unix())
		// Create 11 test records for pagination testing
		records := []RPCRecord{
			{Sender: userAddress, Method: "ping", Params: []byte(`[null]`), Response: []byte(`{"res":[1,"pong",[],1621234567890]}`)},
			{Sender: userAddress, Method: "get_config", Params: []byte(`[]`), Response: []byte(`{"res":[2,"get_config",[{"broker_address":"0xBroker"}],1621234597890]}`)},
			{Sender: userAddress, Method: "get_channels", Params: []byte(fmt.Sprintf(`[{"participant":"%s"}]`, userAddress)), Response: []byte(`{"res":[3,"get_channels",[[]],1621234627890]}`)},
			{Sender: userAddress, Method: "transfer", Params: []byte(`[{"destination":"0xDest","allocations":[{"asset":"USDC","amount":"100"}]}]`), Response: []byte(`{"res":[4,"transfer",[],1621234657890]}`)},
			{Sender: userAddress, Method: "get_ledger_balances", Params: []byte(`[]`), Response: []byte(`{"res":[5,"get_ledger_balances",[],1621234687890]}`)},
			{Sender: userAddress, Method: "create_application", Params: []byte(`[{"definition":{"protocol":"test"}}]`), ReqSig: []string{"0x0006"}, Response: []byte(`{"res":[6,"create_application",[],1621234717890]}`)},
			{Sender: userAddress, Method: "submit_app_state", Params: []byte(`[{"app_session_id":"123"}]`), Response: []byte(`{"res":[7,"submit_app_state",[],1621234747890]}`)},
			{Sender: userAddress, Method: "close_application", Params: []byte(`[{"app_session_id":"123"}]`), Response: []byte(`{"res":[8,"close_application",[],1621234777890]}`)},
			{Sender: userAddress, Method: "resize_channel", Params: []byte(`[{"channel_id":"ch123"}]`), Response: []byte(`{"res":[9,"resize_channel",[],1621234807890]}`)},
			{Sender: userAddress, Method: "close_channel", Params: []byte(`[{"channel_id":"ch123"}]`), Response: []byte(`{"res":[10,"close_channel",[],1621234837890]}`)},
			{Sender: userAddress, Method: "get_user_tag", Params: []byte(`[]`), Response: []byte(`{"res":[11,"get_user_tag",[],1621234867890]}`)},
			{Sender: "0xOtherParticipant", Method: "ping", Params: []byte(`[null]`), Response: []byte(`{"res":[12,"pong",[],1621234897890]}`)},
		}

		numOfTestRecords := len(records)

		for i := range records {
			records[i].ReqID = uint64(i + 1)
			records[i].Timestamp = baseTime - uint64(numOfTestRecords-i)
			records[i].ReqSig = []string{fmt.Sprintf("0x%04X", i+1)}
			records[i].ResSig = []string{}
		}

		require.NoError(t, router.DB.Create(records).Error)

		// Expected record IDs in descending order (newest first)
		expectedReqIDs := []uint64{11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

		testCases := []struct {
			name                string
			params              map[string]interface{}
			expectedReqIDs      []uint64
			expectedRecordCount int
		}{
			{
				name:                "No params (default pagination)",
				params:              map[string]interface{}{},
				expectedReqIDs:      expectedReqIDs[:10], // Default limit is 10
				expectedRecordCount: 10,
			},
			{
				name:                "Offset only",
				params:              map[string]interface{}{"offset": float64(2)},
				expectedReqIDs:      expectedReqIDs[2:], // Skip first 2
				expectedRecordCount: 9,
			},
			{
				name:                "Limit only",
				params:              map[string]interface{}{"limit": float64(5)},
				expectedReqIDs:      expectedReqIDs[:5], // First 5 records
				expectedRecordCount: 5,
			},
			{
				name:                "Offset and limit",
				params:              map[string]interface{}{"offset": float64(2), "limit": float64(3)},
				expectedReqIDs:      expectedReqIDs[2:5], // Skip 2, take 3
				expectedRecordCount: 3,
			},
			{
				name:                "Pagination with sort asc",
				params:              map[string]interface{}{"offset": float64(1), "limit": float64(3), "sort": "asc"},
				expectedReqIDs:      []uint64{2, 3, 4}, // Ascending order, skip 1, take 3
				expectedRecordCount: 3,
			},
			{
				name:                "Pagination with sort desc (default)",
				params:              map[string]interface{}{"offset": float64(1), "limit": float64(3), "sort": "desc"},
				expectedReqIDs:      expectedReqIDs[1:4], // Descending order, skip 1, take 3
				expectedRecordCount: 3,
			},
			{
				name:                "Offset beyond available records",
				params:              map[string]interface{}{"offset": float64(20)},
				expectedReqIDs:      []uint64{}, // No records
				expectedRecordCount: 0,
			},
			{
				name:                "Limit exceeds max limit",
				params:              map[string]interface{}{"limit": float64(200)},
				expectedReqIDs:      expectedReqIDs, // Should be capped at MaxLimit (100), but we only have 11 records
				expectedRecordCount: 11,
			},
		}

		for idx, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ctx := createSignedRPCContext(idx+100, "get_rpc_history", tc.params, userSigner)

				router.HandleGetRPCHistory(ctx)

				res := assertResponse(t, ctx, "get_rpc_history")

				require.Equal(t, uint64(idx+100), res.RequestID)
				rpcHistory, ok := res.Params.(GetRPCHistoryResponse)
				require.True(t, ok, "Response parameter should be a GetRPCHistoryResponse")
				assert.Len(t, rpcHistory.RPCEntries, tc.expectedRecordCount, "Should return expected number of records")

				// Check records are in expected order
				for i, record := range rpcHistory.RPCEntries {
					if i < len(tc.expectedReqIDs) {
						assert.Equal(t, tc.expectedReqIDs[i], record.ReqID, "Record %d should have expected ReqID", i)
						assert.Equal(t, userAddress, record.Sender, "All records should belong to the requesting participant")
					}
				}
			})
		}
	})
}

func TestRPCRouterHandleGetLedgerBalances(t *testing.T) {
	key, _ := crypto.GenerateKey()
	userSigner := Signer{privateKey: key}

	t.Run("Succesfully retrieve ledger balance", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		ledger := GetWalletLedger(db, userSigner.GetAddress())
		userAccountID := NewAccountID(userSigner.GetAddress().Hex())
		require.NoError(t, ledger.Record(userAccountID, "usdc", decimal.NewFromInt(1000)))

		ctx := createSignedRPCContext(1, "get_ledger_balances", map[string]string{"account_id": userAccountID.String()}, userSigner)
		router.HandleGetLedgerBalances(ctx)

		res := assertResponse(t, ctx, "get_ledger_balances")
		balancesResp, ok := res.Params.(GetLedgerBalancesResponse)
		balancesArray := balancesResp.LedgerBalances
		require.True(t, ok)
		require.Len(t, balancesArray, 1)
		require.Equal(t, "usdc", balancesArray[0].Asset)
		require.Equal(t, decimal.NewFromInt(1000), balancesArray[0].Amount)
	})
}

func TestRPCRouterHandleGetUserTag(t *testing.T) {
	userKey, _ := crypto.GenerateKey()
	userSigner := Signer{privateKey: userKey}
	userAddr := userSigner.GetAddress().Hex()

	t.Run("Succesfully retrieve the user tag", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		require.NoError(t, db.Create(&SignerWallet{Signer: userAddr, Wallet: userAddr}).Error)
		t.Cleanup(cleanup)

		userTag, err := GenerateOrRetrieveUserTag(db, userAddr)
		require.NoError(t, err)

		ctx := createSignedRPCContext(42, "get_user_tag", nil, userSigner)
		router.HandleGetUserTag(ctx)

		assertResponse(t, ctx, "get_user_tag")
		getTagResponse, ok := ctx.Message.Res.Params.(GetUserTagResponse)
		require.True(t, ok, "Response should be a GetUserTagResponse")
		require.Equal(t, userTag.Tag, getTagResponse.Tag)
	})
	t.Run("Error when there is no tag", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		require.NoError(t, db.Create(&SignerWallet{Signer: userAddr, Wallet: userAddr}).Error)
		t.Cleanup(cleanup)

		ctx := createSignedRPCContext(1, "get_user_tag", nil, userSigner)
		router.HandleGetUserTag(ctx)

		assertErrorResponse(t, ctx, "failed to get user tag")
	})
}

func TestRPCRouterHandleTransfer(t *testing.T) {
	senderKey, _ := crypto.GenerateKey()
	senderSigner := Signer{privateKey: senderKey}

	senderAddr := senderSigner.GetAddress()
	senderAccountID := NewAccountID(senderAddr.Hex())
	recipientAddr := newTestCommonAddress("0xRecipient")
	recipientAccountID := NewAccountID(recipientAddr.Hex())

	t.Run("SuccessfulTransfer", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		require.NoError(t, db.Create(&SignerWallet{Signer: senderAddr.Hex(), Wallet: senderAddr.Hex()}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "eth", decimal.NewFromInt(5)))

		transferParams := TransferParams{
			Destination: recipientAddr.Hex(),
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
				{AssetSymbol: "eth", Amount: decimal.NewFromInt(2)},
			},
		}

		ctx := createSignedRPCContext(1, "transfer", transferParams, senderSigner)
		router.HandleTransfer(ctx)

		res := assertResponse(t, ctx, "transfer")
		transferResp, ok := res.Params.(TransferResponse)
		require.Len(t, transferResp.Transactions, 2, "Response should contain 2 transaction objects")

		transferTransaction := transferResp.Transactions[0]
		require.True(t, ok, "Response should be a slice of TransactionResponse")
		require.Equal(t, senderAddr.Hex(), transferTransaction.FromAccount)
		require.Equal(t, recipientAddr.Hex(), transferTransaction.ToAccount)
		require.False(t, transferTransaction.CreatedAt.IsZero(), "CreatedAt should be set")

		// Verify user tags are empty (since no tags were created for these wallets)
		require.Empty(t, transferTransaction.FromAccountTag, "FromAccountTag should be empty when no tag exists")
		require.Empty(t, transferTransaction.ToAccountTag, "ToAccountTag should be empty when no tag exists")

		// Verify that all transactions in response have the tag fields
		for _, tx := range transferResp.Transactions {
			require.Equal(t, senderAddr.Hex(), tx.FromAccount)
			require.Equal(t, recipientAddr.Hex(), tx.ToAccount)
			require.Empty(t, tx.FromAccountTag, "FromAccountTag should be empty when no tag exists")
			require.Empty(t, tx.ToAccountTag, "ToAccountTag should be empty when no tag exists")
		}

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
			switch tx.AssetSymbol {
			case "usdc":
				require.Equal(t, decimal.NewFromInt(500), tx.Amount, "USDC amount should match")
			case "eth":
				require.Equal(t, decimal.NewFromInt(2), tx.Amount, "ETH amount should match")
			default:
				t.Errorf("Unexpected asset symbol: %s", tx.AssetSymbol)
			}
		}

		// Verify response transactions match database transactions
		for _, responseTx := range transferResp.Transactions {
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
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		require.NoError(t, db.Create(&SignerWallet{Signer: senderAddr.Hex(), Wallet: senderAddr.Hex()}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "eth", decimal.NewFromInt(5)))

		// Setup user tag for recipient
		recipientTag, err := GenerateOrRetrieveUserTag(db, recipientAddr.Hex())
		require.NoError(t, err)

		transferParams := TransferParams{
			DestinationUserTag: strings.ToLower(recipientTag.Tag), // Verify that tag is case-insensitive
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
				{AssetSymbol: "eth", Amount: decimal.NewFromInt(2)},
			},
		}

		ctx := createSignedRPCContext(1, "transfer", transferParams, senderSigner)
		router.HandleTransfer(ctx)

		res := assertResponse(t, ctx, "transfer")
		transactionResponse, ok := res.Params.(TransferResponse)
		require.True(t, ok, "Response should be a TransactionResponse")
		require.Len(t, transactionResponse.Transactions, 2, "Should have 2 transaction entries for the transfer")

		targetTransaction := transactionResponse.Transactions[0]

		require.Equal(t, senderAddr.Hex(), targetTransaction.FromAccount)
		require.Equal(t, recipientAddr.Hex(), targetTransaction.ToAccount)
		require.False(t, targetTransaction.CreatedAt.IsZero(), "CreatedAt should be set")

		// Verify user tag fields in transaction response
		require.Empty(t, targetTransaction.FromAccountTag, "FromAccountTag should be empty since sender has no tag")
		require.Equal(t, recipientTag.Tag, targetTransaction.ToAccountTag, "ToAccountTag should match recipient's tag")

		// Verify all transactions have correct tag information
		for _, tx := range transactionResponse.Transactions {
			require.Equal(t, senderAddr.Hex(), tx.FromAccount)
			require.Equal(t, recipientAddr.Hex(), tx.ToAccount)
			require.Empty(t, tx.FromAccountTag, "FromAccountTag should be empty since sender has no tag")
			require.Equal(t, recipientTag.Tag, tx.ToAccountTag, "ToAccountTag should match recipient's tag")
		}

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
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		require.NoError(t, db.Create(&SignerWallet{Signer: senderAddr.Hex(), Wallet: senderAddr.Hex()}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))

		// Create transfer with invalid destination
		transferParams := TransferParams{
			Destination: "not-a-valid-address",
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		ctx := createSignedRPCContext(48, "transfer", transferParams, senderSigner)
		router.HandleTransfer(ctx)

		assertErrorResponse(t, ctx, "invalid destination account")
	})

	t.Run("ErrorTransferToSelf", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		require.NoError(t, db.Create(&SignerWallet{Signer: senderAddr.Hex(), Wallet: senderAddr.Hex()}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))

		// Create transfer to self
		transferParams := TransferParams{
			Destination: senderAddr.Hex(),
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		ctx := createSignedRPCContext(48, "transfer", transferParams, senderSigner)
		router.HandleTransfer(ctx)

		assertErrorResponse(t, ctx, "cannot transfer to self")
	})

	t.Run("ErrorInsufficientFunds", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		require.NoError(t, db.Create(&SignerWallet{Signer: senderAddr.Hex(), Wallet: senderAddr.Hex()}).Error)

		// Fund sender's account with a small amount
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(100)))

		// Create transfer for more than available
		transferParams := TransferParams{
			Destination: recipientAddr.Hex(),
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		ctx := createSignedRPCContext(48, "transfer", transferParams, senderSigner)
		router.HandleTransfer(ctx)

		assertErrorResponse(t, ctx, "insufficient funds")
	})

	t.Run("ErrorEmptyAllocations", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		require.NoError(t, db.Create(&SignerWallet{Signer: senderAddr.Hex(), Wallet: senderAddr.Hex()}).Error)

		// Create transfer with empty allocations
		transferParams := TransferParams{
			Destination: recipientAddr.Hex(),
			Allocations: []TransferAllocation{},
		}

		ctx := createSignedRPCContext(48, "transfer", transferParams, senderSigner)
		router.HandleTransfer(ctx)

		assertErrorResponse(t, ctx, "allocations cannot be empty")
	})

	t.Run("ErrorZeroAmount", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		require.NoError(t, db.Create(&SignerWallet{Signer: senderAddr.Hex(), Wallet: senderAddr.Hex()}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))

		// Create transfer with zero amount
		transferParams := TransferParams{
			Destination: recipientAddr.Hex(),
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
			},
		}

		ctx := createSignedRPCContext(48, "transfer", transferParams, senderSigner)
		router.HandleTransfer(ctx)

		assertErrorResponse(t, ctx, "invalid allocation")
	})

	t.Run("ErrorNegativeAmount", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		require.NoError(t, db.Create(&SignerWallet{Signer: senderAddr.Hex(), Wallet: senderAddr.Hex()}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))

		// Create transfer with negative amount
		transferParams := TransferParams{
			Destination: recipientAddr.Hex(),
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(-500)},
			},
		}

		ctx := createSignedRPCContext(48, "transfer", transferParams, senderSigner)
		router.HandleTransfer(ctx)

		assertErrorResponse(t, ctx, "invalid allocation")
	})

	t.Run("ErrorInvalidSignature", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		require.NoError(t, db.Create(&SignerWallet{Signer: senderAddr.Hex(), Wallet: senderAddr.Hex()}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAccountID, "usdc", decimal.NewFromInt(1000)))

		transferParams := TransferParams{
			Destination: recipientAddr.Hex(),
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		wrongKey, _ := crypto.GenerateKey()
		wrongSigner := Signer{privateKey: wrongKey}

		ctx := createSignedRPCContext(48, "transfer", transferParams, wrongSigner)
		ctx.UserID = senderAddr.Hex() // Ensure user ID is still the sender's address

		router.HandleTransfer(ctx)

		assertErrorResponse(t, ctx, "invalid signature")
	})
}

func TestRPCRouterHandleCreateAppSession(t *testing.T) {
	keyA, _ := crypto.GenerateKey()
	keyB, _ := crypto.GenerateKey()
	userA := Signer{privateKey: keyA}
	userB := Signer{privateKey: keyB}
	userAddressA := userA.GetAddress()
	userAddressB := userB.GetAddress()
	accountIDA := NewAccountID(userAddressA.Hex())
	accountIDB := NewAccountID(userAddressB.Hex())

	t.Run("SuccessfulCreateAppSession", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		for i, p := range []string{userAddressA.Hex(), userAddressB.Hex()} {
			ch := &Channel{
				ChannelID:   fmt.Sprintf("0xChannel%ctx", 'A'+i),
				Wallet:      p,
				Participant: p,
				Status:      ChannelStatusOpen,
				Token:       "0xTokenXYZ",
				Nonce:       1,
			}
			require.NoError(t, db.Create(ch).Error)
			require.NoError(t, db.Create(&SignerWallet{
				Signer: p, Wallet: p,
			}).Error)
		}

		require.NoError(t, GetWalletLedger(db, userAddressA).Record(accountIDA, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(accountIDB, "usdc", decimal.NewFromInt(200)))

		ts := uint64(time.Now().Unix())
		def := AppDefinition{
			Protocol:           "test-proto",
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Challenge:          60,
			Nonce:              ts,
		}
		data := `{"state":"initial"}`
		createParams := CreateAppSessionParams{
			Definition: def,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
			SessionData: &data,
		}

		ctx := createSignedRPCContext(1, "create_app_session", createParams, userA, userB)
		router.HandleCreateApplication(ctx)

		res := assertResponse(t, ctx, "create_app_session")
		appResp, ok := res.Params.(AppSessionResponse)
		require.True(t, ok)
		require.Equal(t, string(ChannelStatusOpen), appResp.Status)
		require.Equal(t, uint64(1), appResp.Version)
		require.Empty(t, appResp.SessionData, "session data should not be returned in response")

		var vApp AppSession
		require.NoError(t, db.Where("session_id = ?", appResp.AppSessionID).First(&vApp).Error)
		require.ElementsMatch(t, []string{userAddressA.Hex(), userAddressB.Hex()}, vApp.ParticipantWallets)
		require.Equal(t, uint64(1), vApp.Version)
		require.Equal(t, data, vApp.SessionData, "session data should be stored in the database")

		// Participant accounts drained
		partBalA, _ := GetWalletLedger(db, userAddressA).Balance(accountIDA, "usdc")
		partBalB, _ := GetWalletLedger(db, userAddressB).Balance(accountIDB, "usdc")
		require.True(t, partBalA.IsZero(), "Participant A balance should be zero")
		require.True(t, partBalB.IsZero(), "Participant B balance should be zero")

		// Virtual-app funded
		sessionAccountID := NewAccountID(appResp.AppSessionID)
		vBalA, _ := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "usdc")
		vBalB, _ := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "usdc")
		require.Equal(t, decimal.NewFromInt(100).String(), vBalA.String())
		require.Equal(t, decimal.NewFromInt(200).String(), vBalB.String())
	})
	t.Run("ErrorChallengedChannel", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		for i, p := range []string{userAddressA.Hex(), userAddressB.Hex()} {
			ch := &Channel{
				ChannelID:   fmt.Sprintf("0xChannel%ctx", 'A'+i),
				Wallet:      p,
				Participant: p,
				Status:      ChannelStatusChallenged,
				Token:       "0xTokenXYZ",
				Nonce:       1,
			}
			require.NoError(t, db.Create(ch).Error)
			require.NoError(t, db.Create(&SignerWallet{
				Signer: p, Wallet: p,
			}).Error)
		}

		require.NoError(t, GetWalletLedger(db, userAddressA).Record(accountIDA, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(accountIDB, "usdc", decimal.NewFromInt(200)))

		ts := uint64(time.Now().Unix())
		def := AppDefinition{
			Protocol:           "test-proto",
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Challenge:          60,
			Nonce:              ts,
		}
		createParams := CreateAppSessionParams{
			Definition: def,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
		}

		ctx := createSignedRPCContext(1, "create_app_session", createParams, userA, userB)
		router.HandleCreateApplication(ctx)

		assertErrorResponse(t, ctx, "has challenged channels")
	})
}

func TestRPCRouterHandleSubmitAppState(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	userSigner := Signer{privateKey: key}

	userAddressA := userSigner.GetAddress()
	userAddressB := newTestCommonAddress("0xParticipantB")

	t.Run("SuccessfulSubmitAppState", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		channels := []Channel{
			{
				ChannelID:   "0xChannelA",
				Participant: userAddressA.Hex(),
				Status:      ChannelStatusOpen,
				Token:       "0xToken123",
				Nonce:       1,
			},
			{
				ChannelID:   "0xChannelB",
				Participant: userAddressB.Hex(),
				Status:      ChannelStatusOpen,
				Token:       "0xToken123",
				Nonce:       1,
			},
		}

		require.NoError(t, db.Create(channels).Error)

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

		ctx := createSignedRPCContext(1, "submit_app_state", submitAppStateParams, userSigner)
		router.HandleSubmitAppState(ctx)

		res := assertResponse(t, ctx, "submit_app_state")
		appResp, ok := res.Params.(AppSessionResponse)
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

func TestRPCRouterHandleCloseAppSession(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	userSigner := Signer{privateKey: key}

	userAddressA := userSigner.GetAddress()
	accountIDA := NewAccountID(userAddressA.Hex())

	userAddressB := newTestCommonAddress("0xParticipantB")
	accountIDB := NewAccountID(userAddressB.Hex())

	t.Run("SuccessfulCloseAppSession", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		channels := []Channel{
			{
				ChannelID:   "0xChannelA",
				Participant: userAddressA.Hex(),
				Status:      ChannelStatusOpen,
				Token:       "0xToken123",
				Nonce:       1,
			},
			{
				ChannelID:   "0xChannelB",
				Participant: userAddressB.Hex(),
				Status:      ChannelStatusOpen,
				Token:       "0xToken123",
				Nonce:       1,
			},
		}

		require.NoError(t, db.Create(channels).Error)

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

		ctx := createSignedRPCContext(1, "close_app_session", closeParams, userSigner)
		router.HandleCloseApplication(ctx)

		res := assertResponse(t, ctx, "close_app_session")
		appResp, ok := res.Params.(AppSessionResponse)
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
	})

}

func TestRPCRouterHandleResizeChannel(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	userSigner := Signer{privateKey: key}
	userAddress := userSigner.GetAddress()
	userAccountID := NewAccountID(userAddress.Hex())

	t.Run("SuccessfulAllocation", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		// Seed asset
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
		allocateAmount := decimal.NewFromInt(200)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(1, "resize_channel", resizeParams, userSigner)
		router.HandleResizeChannel(ctx)

		res := assertResponse(t, ctx, "resize_channel")
		resObj, ok := res.Params.(ResizeChannelResponse)
		require.True(t, ok, "Response should be ResizeChannelResponse")
		require.Equal(t, ch.ChannelID, resObj.ChannelID)
		require.Equal(t, ch.Version+1, resObj.Version)

		// New channel amount should be initial + 200
		expected := initialRawAmount.Add(allocateAmount)
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(expected), "Allocated amount mismatch")
		require.Equal(t, 0, resObj.Allocations[1].RawAmount.Cmp(decimal.Zero), "Broker allocation should be zero")

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
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

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
		allocateAmount := decimal.NewFromInt(-300)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(1, "resize_channel", resizeParams, userSigner)
		router.HandleResizeChannel(ctx)

		res := assertResponse(t, ctx, "resize_channel")
		resObj, ok := res.Params.(ResizeChannelResponse)
		require.True(t, ok)

		// Channel amount should decrease
		expected := initialRawAmount.Add(allocateAmount)
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(expected), "Decreased amount mismatch")

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(userAccountID, "usdc")
		require.NoError(t, err)
		require.Equal(t, decimal.NewFromInt(500), finalBalance) // Should remain unchanged
	})

	t.Run("ErrorInvalidChannelID", func(t *testing.T) {
		t.Parallel()

		router, _, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		allocateAmount := decimal.NewFromInt(100)
		resizeParams := ResizeChannelParams{
			ChannelID:        "0xNonExistentChannel",
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(1, "resize_channel", resizeParams, userSigner)
		router.HandleResizeChannel(ctx)

		assertErrorResponse(t, ctx, "channel 0xNonExistentChannel not found")
	})

	t.Run("ErrorChannelClosed", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

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

		allocateAmount := decimal.NewFromInt(100)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(1, "resize_channel", resizeParams, userSigner)
		router.HandleResizeChannel(ctx)

		assertErrorResponse(t, ctx, "channel 0xChanClosed is not open: closed")
	})

	t.Run("ErrorChannelJoining", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
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

		allocateAmount := decimal.NewFromInt(100)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(1, "resize_channel", resizeParams, userSigner)
		router.HandleResizeChannel(ctx)

		assertErrorResponse(t, ctx, "channel 0xChanJoining is not open: joining")
	})

	t.Run("ErrorOtherChallengedChannel", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
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

		allocateAmount := decimal.NewFromInt(100)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(1, "resize_channel", resizeParams, userSigner)
		router.HandleResizeChannel(ctx)

		assertErrorResponse(t, ctx, "has challenged channels")
	})

	t.Run("ErrorInsufficientFunds", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

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
		require.NoError(t, GetWalletLedger(db, userAddress).Record(userAccountID, "usdc", decimal.NewFromFloat(0.000001)))

		allocateAmount := decimal.NewFromInt(200)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(1, "resize_channel", resizeParams, userSigner)
		router.HandleResizeChannel(ctx)

		assertErrorResponse(t, ctx, "insufficient unified balance")
	})

	t.Run("ErrorZeroAmounts", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

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
			AllocateAmount:   &decimal.Zero,
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(1, "resize_channel", resizeParams, userSigner)

		router.HandleResizeChannel(ctx)

		res := ctx.Message.Res
		require.NotNil(t, res)

		// Zero allocation should now be rejected as it's a wasteful no-op operation
		assertErrorResponse(t, ctx, "resize operation requires non-zero ResizeAmount or AllocateAmount")
	})

	t.Run("SuccessfulResizeDeposit", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

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
		allocateAmount := decimal.NewFromInt(100)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			ResizeAmount:     &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(1, "resize_channel", resizeParams, userSigner)
		router.HandleResizeChannel(ctx)

		res := assertResponse(t, ctx, "resize_channel")
		resObj, ok := res.Params.(ResizeChannelResponse)
		require.True(t, ok)

		// Should be initial amount (1000) + allocate amount (0) + resize amount (100) = 1100
		expected := initialRawAmount.Add(allocateAmount)
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(expected))
	})

	t.Run("SuccessfulResizeWithdrawal", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

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
		allocateAmount := decimal.NewFromInt(-100)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			ResizeAmount:     &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(1, "resize_channel", resizeParams, userSigner)
		// Call handler
		router.HandleResizeChannel(ctx)

		res := assertResponse(t, ctx, "resize_channel")
		resObj, ok := res.Params.(ResizeChannelResponse)
		require.True(t, ok)

		// Should be initial amount (1000) + allocate amount (0) - resize amount (100) = 900
		expected := initialRawAmount.Add(allocateAmount)
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(expected))
	})

	t.Run("ErrorExcessiveDeallocation", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

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
		allocateAmount := decimal.NewFromInt(-1500) // More than current channel amount
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(7, "resize_channel", resizeParams, userSigner)
		router.HandleResizeChannel(ctx)

		assertErrorResponse(t, ctx, "new channel amount must be positive")
	})

	t.Run("ErrorInvalidSignature", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

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

		allocateAmount := decimal.NewFromInt(100)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(8, "resize_channel", resizeParams, wrongSigner)
		router.HandleResizeChannel(ctx)

		assertErrorResponse(t, ctx, "invalid signature")
	})

	t.Run("BoundaryLargeAllocation", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

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

		allocateAmount := decimal.New(10, 15) // 10^15
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount, // 10^15
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(9, "resize_channel", resizeParams, userSigner)
		router.HandleResizeChannel(ctx)

		res := assertResponse(t, ctx, "resize_channel")
		resObj, ok := res.Params.(ResizeChannelResponse)
		require.True(t, ok)

		// Verify the large allocation was processed correctly
		expectedAmount := decimal.NewFromInt(1000).Add(allocateAmount) // 1000 + 10^15
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(expectedAmount))
	})

	t.Run("SuccessfulAllocationWithResizeDeposit", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		// Seed asset
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
		allocateAmount := decimal.NewFromInt(150) // Allocation: moves funds from user balance to channel
		resizeAmount := decimal.NewFromInt(100)   // Resize: deposits additional funds into channel
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount, // Allocation: moves funds from user balance to channel
			ResizeAmount:     &resizeAmount,   // Resize: deposits additional funds into channel
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(12, "resize_channel", resizeParams, userSigner)
		router.HandleResizeChannel(ctx)

		res := assertResponse(t, ctx, "resize_channel")
		resObj, ok := res.Params.(ResizeChannelResponse)
		require.True(t, ok, "Response should be ResizeChannelResponse")
		require.Equal(t, ch.ChannelID, resObj.ChannelID)
		require.Equal(t, ch.Version+1, resObj.Version)

		// New channel amount should be initial + AllocateAmount + ResizeAmount = 1000 + 150 + 100 = 1250
		expected := initialRawAmount.Add(allocateAmount).Add(resizeAmount)
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(expected), "Combined allocation+resize amount mismatch")
		require.Equal(t, 0, resObj.Allocations[1].RawAmount.Cmp(decimal.Zero), "Broker allocation should be zero")

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
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		// Seed asset
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
		allocateAmount := decimal.NewFromInt(100)
		resizeAmount := decimal.NewFromInt(-100)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount, // Allocation: moves funds from user balance to channel
			ResizeAmount:     &resizeAmount,   // Resize: immediately withdraws allocated funds from channel
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(1, "resize_channel", resizeParams, userSigner)
		router.HandleResizeChannel(ctx)

		res := assertResponse(t, ctx, "resize_channel")
		resObj, ok := res.Params.(ResizeChannelResponse)
		require.True(t, ok, "Response should be ResizeChannelResponse")
		require.Equal(t, ch.ChannelID, resObj.ChannelID)
		require.Equal(t, ch.Version+1, resObj.Version)

		// New channel amount should be initial + AllocateAmount + ResizeAmount = 0 + 100 - 100 = 0
		require.Equal(t, 0, resObj.Allocations[0].RawAmount.Cmp(decimal.Zero), "Combined allocation+resize amount mismatch")
		require.Equal(t, 0, resObj.Allocations[1].RawAmount.Cmp(decimal.Zero), "Broker allocation should be zero")

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
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	userSigner := Signer{privateKey: key}
	userAddress := userSigner.GetAddress()
	userAccountID := NewAccountID(userAddress.Hex())

	t.Run("SuccessfulCloseChannel", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		// Seed asset
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

		closeParams := CloseChannelParams{
			ChannelID:        ch.ChannelID,
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(10, "close_channel", closeParams, userSigner)
		router.HandleCloseChannel(ctx)

		res := assertResponse(t, ctx, "close_channel")
		resObj, ok := res.Params.(CloseChannelResponse)
		require.True(t, ok, "Response should be CloseChannelResponse")
		require.Equal(t, ch.ChannelID, resObj.ChannelID)
		require.Equal(t, ch.Version+1, resObj.Version)

		// Final allocation should send full balance to destination
		require.Equal(t, 0, resObj.FinalAllocations[0].RawAmount.Cmp(initialRawAmount), "Primary allocation mismatch")
		require.Equal(t, 0, resObj.FinalAllocations[1].RawAmount.Cmp(decimal.Zero), "Broker allocation should be zero")
	})
	t.Run("ErrorOtherChannelInChallenge", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		t.Cleanup(cleanup)

		// Seed asset
		asset := Asset{Token: "0xTokenClose", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		initialRawAmount := decimal.NewFromInt(500)

		channels := []Channel{
			{
				ChannelID:   "0xChanChallenged",
				Participant: userAddress.Hex(),
				Wallet:      userAddress.Hex(),
				Status:      ChannelStatusChallenged,
				Token:       asset.Token,
				ChainID:     137,
				RawAmount:   initialRawAmount,
				Version:     2,
			},
			{
				ChannelID:   "0xChanToClose",
				Participant: userAddress.Hex(),
				Wallet:      userAddress.Hex(),
				Status:      ChannelStatusOpen,
				Token:       asset.Token,
				ChainID:     137,
				RawAmount:   initialRawAmount,
				Version:     2,
			},
		}

		require.NoError(t, db.Create(channels).Error)

		// Fund participant ledger so that raw units match channel.Amount
		require.NoError(t, GetWalletLedger(db, userAddress).Record(
			userAccountID,
			"usdc",
			rawToDecimal(initialRawAmount.BigInt(), asset.Decimals),
		))

		closeParams := CloseChannelParams{
			ChannelID:        "0xChanToClose",
			FundsDestination: userAddress.Hex(),
		}

		ctx := createSignedRPCContext(1, "close_channel", closeParams, userSigner)
		router.HandleCloseChannel(ctx)

		assertErrorResponse(t, ctx, "has challenged channels")
	})
}
