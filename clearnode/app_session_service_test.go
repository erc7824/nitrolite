package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertNotifications(t *testing.T, capturedNotifications map[string][]Notification, userID string, expectedCount int) {
	assert.Contains(t, capturedNotifications, userID)
	assert.Len(t, capturedNotifications[userID], expectedCount)
}

func TestAppSessionService_CreateApplication(t *testing.T) {
	rawA, _ := crypto.GenerateKey()
	rawB, _ := crypto.GenerateKey()
	signerA := Signer{privateKey: rawA}
	signerB := Signer{privateKey: rawB}
	userAddressA := signerA.GetAddress()
	userAddressB := signerB.GetAddress()
	userAccountIDA := NewAccountID(userAddressA.Hex())
	userAccountIDB := NewAccountID(userAddressB.Hex())

	t.Run("SuccessfulCreateApplication", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Setup: Create wallets and fund them
		require.NoError(t, db.Create(&SignerWallet{Signer: userAddressA.Hex(), Wallet: userAddressA.Hex()}).Error)
		require.NoError(t, db.Create(&SignerWallet{Signer: userAddressB.Hex(), Wallet: userAddressB.Hex()}).Error)
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(userAccountIDA, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(userAccountIDB, "usdc", decimal.NewFromInt(200)))

		capturedNotifications := make(map[string][]Notification)
		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {

			capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}, nil))

		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           rpc.VersionNitroRPCv0_2,
				ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
				Weights:            []int64{1, 1},
				Quorum:             2,
				Challenge:          60,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
		}
		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		appSession, err := service.CreateApplication(params, rpcSigners)
		require.NoError(t, err)
		assert.NotNil(t, appSession)
		assert.NotEmpty(t, appSession.AppSessionID)
		assert.Equal(t, uint64(1), appSession.Version)
		assert.Equal(t, ChannelStatusOpen, ChannelStatus(appSession.Status))

		sessionAccountID := NewAccountID(appSession.AppSessionID)

		assert.Len(t, capturedNotifications, 2)
		assertNotifications(t, capturedNotifications, userAddressA.Hex(), 1)
		assertNotifications(t, capturedNotifications, userAddressB.Hex(), 1)

		// Verify balances
		balA, err := GetWalletLedger(db, userAddressA).Balance(userAccountIDA, "usdc")
		require.NoError(t, err)
		assert.True(t, balA.IsZero())

		appBalA, err := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(100), appBalA)

		// Verify transactions were recorded to the database
		var transactions []LedgerTransaction
		err = db.Where("tx_type = ?", TransactionTypeAppDeposit).Find(&transactions).Error
		require.NoError(t, err)
		assert.Len(t, transactions, 2, "Should have 2 app deposit transactions recorded")

		// Verify transaction details
		expectedTxs := map[string]decimal.Decimal{
			userAddressA.Hex(): decimal.NewFromInt(100),
			userAddressB.Hex(): decimal.NewFromInt(200),
		}

		assert.Equal(t, len(transactions), 2, "Expected 2 transactions to be recorded")
		for _, tx := range transactions {
			expectedAmount, exists := expectedTxs[tx.FromAccount]
			assert.True(t, exists, "Unexpected destination of a transaction: %s", tx.FromAccount)
			assert.Equal(t, TransactionTypeAppDeposit, tx.Type, "Transaction type should be app deposit")
			assert.Equal(t, appSession.AppSessionID, tx.ToAccount, "To account should be app session ID")
			assert.Equal(t, "usdc", tx.AssetSymbol, "Asset symbol should be usdc")
			assert.Equal(t, expectedAmount, tx.Amount, "Amount should match allocation")
			assert.False(t, tx.CreatedAt.IsZero(), "CreatedAt should be set")
		}
	})

	t.Run("ErrorInsufficientFunds", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		require.NoError(t, db.Create(&SignerWallet{Signer: userAddressA.Hex(), Wallet: userAddressA.Hex()}).Error)
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(userAccountIDA, "usdc", decimal.NewFromInt(50))) // Not enough

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))
		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           rpc.VersionNitroRPCv0_2,
				ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
				Weights:            []int64{1, 0},
				Quorum:             1,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}
		rpcSigners := map[string]struct{}{userAddressA.Hex(): {}}

		_, err := service.CreateApplication(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
	})

	t.Run("ErrorChallengedChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		require.NoError(t, db.Create(&SignerWallet{Signer: userAddressA.Hex(), Wallet: userAddressA.Hex()}).Error)
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(userAccountIDA, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, db.Create(&Channel{
			Wallet: userAddressA.Hex(),
			Status: ChannelStatusChallenged,
		}).Error)

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))
		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           rpc.VersionNitroRPCv0_2,
				ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
				Weights:            []int64{1, 0},
				Quorum:             1,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}
		rpcSigners := map[string]struct{}{userAddressA.Hex(): {}}

		_, err := service.CreateApplication(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has challenged channels")
	})

	t.Run("ErrorNegativeAllocation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		require.NoError(t, db.Create(&SignerWallet{Signer: userAddressA.Hex(), Wallet: userAddressA.Hex()}).Error)
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(userAccountIDA, "usdc", decimal.NewFromInt(100)))

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))
		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           rpc.VersionNitroRPCv0_2,
				ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
				Weights:            []int64{1, 0},
				Quorum:             1,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(-50)},
			},
		}
		rpcSigners := map[string]struct{}{userAddressA.Hex(): {}}

		_, err := service.CreateApplication(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "negative allocation: -50 for asset usdc")
	})
}

func TestAppSessionService_SubmitAppState(t *testing.T) {
	rawA, _ := crypto.GenerateKey()
	rawB, _ := crypto.GenerateKey()
	signerA := Signer{privateKey: rawA}
	signerB := Signer{privateKey: rawB}
	userAddressA := signerA.GetAddress()
	userAddressB := signerB.GetAddress()

	t.Run("SuccessfulSubmitAppState", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))
		session := &AppSession{
			SessionID:          "test-session",
			Protocol:           rpc.VersionNitroRPCv0_2,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		ledgerA := GetWalletLedger(db, userAddressA)
		require.NoError(t, ledgerA.Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Version:      0, // no version for NitroRPCv0_2
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		// Mock ledger balances for the app session
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(sessionAccountID, "usdc", decimal.NewFromInt(0)))

		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		resp, err := service.SubmitAppState(params, rpcSigners)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		// Verify balances
		appBalA, err := ledgerA.Balance(sessionAccountID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(50), appBalA)

		appBalB, err := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(50), appBalB)
	})

	t.Run("ErrorNegativeAllocation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))
		session := &AppSession{
			SessionID:          "test-session-negative",
			Protocol:           rpc.VersionNitroRPCv0_2,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		ledgerA := GetWalletLedger(db, userAddressA)
		require.NoError(t, ledgerA.Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(-50)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
			},
		}

		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		_, err := service.SubmitAppState(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "negative allocation: -50 for asset usdc")
	})

	t.Run("NitroRPCv0.4_OperateSuccess", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))
		session := &AppSession{
			SessionID:          "test-session-v04-operate",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		ledgerA := GetWalletLedger(db, userAddressA)
		require.NoError(t, ledgerA.Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentOperate,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		// Mock ledger balances for the app session
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(sessionAccountID, "usdc", decimal.NewFromInt(0)))

		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		resp, err := service.SubmitAppState(params, rpcSigners)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		// Verify balances
		appBalA, err := ledgerA.Balance(sessionAccountID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(50), appBalA)

		appBalB, err := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(50), appBalB)
	})

	t.Run("NitroRPCv0.4_OperateInvalidVersion", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))
		session := &AppSession{
			SessionID:          "test-session-v04-invalid-version",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		ledgerA := GetWalletLedger(db, userAddressA)
		require.NoError(t, ledgerA.Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentOperate,
			Version:      3,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		// Mock ledger balances for the app session
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(sessionAccountID, "usdc", decimal.NewFromInt(0)))

		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		_, err := service.SubmitAppState(params, rpcSigners)
		require.Equal(t, fmt.Sprintf("incorrect app state: incorrect version: expected %d, got %d", 2, params.Version), err.Error())
	})
}

func TestAppSessionService_CloseApplication(t *testing.T) {
	rawA, _ := crypto.GenerateKey()
	rawB, _ := crypto.GenerateKey()
	signerA := Signer{privateKey: rawA}
	signerB := Signer{privateKey: rawB}
	userAddressA := signerA.GetAddress()
	userAddressB := signerB.GetAddress()
	userAccountIDA := NewAccountID(userAddressA.Hex())

	t.Run("SuccessfulCloseApplication", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		capturedNotifications := make(map[string][]Notification)
		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {
			capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}, nil))

		session := &AppSession{
			SessionID:          "test-session-close",
			Protocol:           rpc.VersionNitroRPCv0_2,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		ledgerA := GetWalletLedger(db, userAddressA)
		ledgerB := GetWalletLedger(db, userAddressB)
		require.NoError(t, ledgerA.Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, ledgerB.Record(sessionAccountID, "usdc", decimal.NewFromInt(200)))

		params := &CloseAppSessionParams{
			AppSessionID: session.SessionID,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
		}
		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		resp, err := service.CloseApplication(params, rpcSigners)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		assert.Len(t, capturedNotifications, 2)
		assertNotifications(t, capturedNotifications, userAddressA.Hex(), 1)
		assertNotifications(t, capturedNotifications, userAddressB.Hex(), 1)

		var closedSession AppSession
		require.NoError(t, db.First(&closedSession, "session_id = ?", session.SessionID).Error)
		assert.Equal(t, ChannelStatusClosed, closedSession.Status)

		// Verify balances
		appBalA, err := ledgerA.Balance(sessionAccountID, "usdc")
		require.NoError(t, err)
		assert.True(t, appBalA.IsZero())

		walletBalA, err := ledgerA.Balance(userAccountIDA, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(100), walletBalA)

		// Verify transactions were recorded to the database
		var transactions []LedgerTransaction
		err = db.Where("tx_type = ?", TransactionTypeAppWithdrawal).Find(&transactions).Error
		require.NoError(t, err)
		assert.Len(t, transactions, 2, "Should have 2 app withdrawal transactions recorded")

		// Verify transaction details
		expectedTxs := map[string]decimal.Decimal{
			userAddressA.Hex(): decimal.NewFromInt(100),
			userAddressB.Hex(): decimal.NewFromInt(200),
		}

		assert.Equal(t, len(transactions), 2, "Expected 2 transactions to be recorded")
		for _, tx := range transactions {
			expectedAmount, exists := expectedTxs[tx.ToAccount]
			assert.True(t, exists, "Unexpected destination of a transaction: %s", tx.ToAccount)
			assert.Equal(t, TransactionTypeAppWithdrawal, tx.Type, "Transaction type should be app withdrawal")
			assert.Equal(t, session.SessionID, tx.FromAccount, "From account should be app session ID")
			assert.Equal(t, "usdc", tx.AssetSymbol, "Asset symbol should be usdc")
			assert.Equal(t, expectedAmount, tx.Amount, "Amount should match allocation")
			assert.False(t, tx.CreatedAt.IsZero(), "CreatedAt should be set")
		}
	})

	t.Run("SuccessfulCloseApplicationWithZeroAllocation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		capturedNotifications := make(map[string][]Notification)
		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {

			capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}, nil))

		session := &AppSession{
			SessionID:          "test-session-close",
			Protocol:           rpc.VersionNitroRPCv0_2,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		ledgerA := GetWalletLedger(db, userAddressA)
		ledgerB := GetWalletLedger(db, userAddressB)
		require.NoError(t, ledgerA.Record(sessionAccountID, "usdc", decimal.NewFromInt(0)))
		require.NoError(t, ledgerB.Record(sessionAccountID, "usdc", decimal.NewFromInt(0)))

		params := &CloseAppSessionParams{
			AppSessionID: session.SessionID,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
			},
		}
		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		resp, err := service.CloseApplication(params, rpcSigners)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		var closedSession AppSession
		require.NoError(t, db.First(&closedSession, "session_id = ?", session.SessionID).Error)
		assert.Equal(t, ChannelStatusClosed, closedSession.Status)

		// Verify balances
		appBalA, err := ledgerA.Balance(sessionAccountID, "usdc")
		require.NoError(t, err)
		assert.True(t, appBalA.IsZero())

		walletBalA, err := ledgerA.Balance(userAccountIDA, "usdc")
		require.NoError(t, err)
		assert.True(t, walletBalA.IsZero())

		assert.Len(t, capturedNotifications, 0)
	})

	t.Run("ErrorNegativeAllocation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))
		session := &AppSession{
			SessionID:          "test-session-close-negative",
			Protocol:           rpc.VersionNitroRPCv0_2,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		ledgerA := GetWalletLedger(db, userAddressA)
		ledgerB := GetWalletLedger(db, userAddressB)
		require.NoError(t, ledgerA.Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, ledgerB.Record(sessionAccountID, "usdc", decimal.NewFromInt(200)))

		params := &CloseAppSessionParams{
			AppSessionID: session.SessionID,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(-100)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(400)},
			},
		}
		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		_, err := service.CloseApplication(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "negative allocation: -100 for asset usdc")
	})
}

func TestAppSessionService_SubmitAppStateDeposit(t *testing.T) {
	rawA, _ := crypto.GenerateKey()
	rawB, _ := crypto.GenerateKey()
	rawC, _ := crypto.GenerateKey()
	signerA := Signer{privateKey: rawA}
	signerB := Signer{privateKey: rawB}
	signerC := Signer{privateKey: rawC}
	userAddressA := signerA.GetAddress()
	userAddressB := signerB.GetAddress()
	userAddressC := signerC.GetAddress()
	userAccountIDA := NewAccountID(userAddressA.Hex())
	userAccountIDB := NewAccountID(userAddressB.Hex())
	userAccountIDC := NewAccountID(userAddressC.Hex())

	t.Run("BasicDepositSuccess", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		require.NoError(t, db.Create(&SignerWallet{Signer: userAddressA.Hex(), Wallet: userAddressA.Hex()}).Error)
		require.NoError(t, db.Create(&SignerWallet{Signer: userAddressB.Hex(), Wallet: userAddressB.Hex()}).Error)
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(userAccountIDA, "usdc", decimal.NewFromInt(200)))
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(userAccountIDB, "usdc", decimal.NewFromInt(100)))

		session := &AppSession{
			SessionID:          "test-session-deposit",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		ledgerA := GetWalletLedger(db, userAddressA)
		ledgerB := GetWalletLedger(db, userAddressB)
		require.NoError(t, ledgerA.Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, ledgerB.Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))

		capturedNotifications := make(map[string][]Notification)
		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {
			capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}, nil))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}
		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		resp, err := service.SubmitAppState(params, rpcSigners)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)
		assert.Equal(t, string(ChannelStatusOpen), resp.Status)

		balA, err := ledgerA.Balance(userAccountIDA, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(150), balA)

		appBalA, err := ledgerA.Balance(sessionAccountID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(150), appBalA)

		appBalB, err := ledgerB.Balance(sessionAccountID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(100), appBalB)

		assert.Contains(t, capturedNotifications, userAddressA.Hex())
		assert.Contains(t, capturedNotifications, userAddressB.Hex())

		assert.GreaterOrEqual(t, len(capturedNotifications[userAddressA.Hex()]), 1)
		assert.GreaterOrEqual(t, len(capturedNotifications[userAddressB.Hex()]), 1)

		var appSessionNotificationFound bool
		for _, notification := range capturedNotifications[userAddressB.Hex()] {
			if notification.eventType == AppSessionUpdateEventType {
				appSessionNotificationFound = true

				notificationData := notification.data
				assert.NotNil(t, notificationData, "App session notification should have data")

				break
			}
		}
		assert.True(t, appSessionNotificationFound, "Should have received app session update notification")

		var depositTx []LedgerTransaction
		err = db.Where("tx_type = ? AND from_account = ? AND asset_symbol = ?",
			TransactionTypeAppDeposit, userAddressA.Hex(), "usdc").Find(&depositTx).Error
		require.NoError(t, err)
		assert.Len(t, depositTx, 1)
		assert.Equal(t, decimal.NewFromInt(50), depositTx[0].Amount)
	})

	t.Run("MultipleParticipantsTokens", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		require.NoError(t, db.Create(&SignerWallet{Signer: userAddressA.Hex(), Wallet: userAddressA.Hex()}).Error)
		require.NoError(t, db.Create(&SignerWallet{Signer: userAddressB.Hex(), Wallet: userAddressB.Hex()}).Error)
		require.NoError(t, db.Create(&SignerWallet{Signer: userAddressC.Hex(), Wallet: userAddressC.Hex()}).Error)

		require.NoError(t, GetWalletLedger(db, userAddressA).Record(userAccountIDA, "usdc", decimal.NewFromInt(200)))
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(userAccountIDA, "eth", decimal.NewFromInt(5)))
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(userAccountIDB, "usdc", decimal.NewFromInt(300)))
		require.NoError(t, GetWalletLedger(db, userAddressC).Record(userAccountIDC, "eth", decimal.NewFromInt(10)))

		session := &AppSession{
			SessionID:          "test-session-multi-deposit",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex(), userAddressC.Hex()},
			Weights:            []int64{1, 1, 1},
			Quorum:             3,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, "eth", decimal.NewFromInt(1)))
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(sessionAccountID, "usdc", decimal.NewFromInt(200)))
		require.NoError(t, GetWalletLedger(db, userAddressC).Record(sessionAccountID, "eth", decimal.NewFromInt(3)))

		capturedNotifications := make(map[string][]Notification)
		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {
			capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}, nil))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "eth", Amount: decimal.NewFromInt(3)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(250)},
				{ParticipantWallet: userAddressC.Hex(), AssetSymbol: "eth", Amount: decimal.NewFromInt(5)},
			},
		}
		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
			userAddressC.Hex(): {},
		}

		resp, err := service.SubmitAppState(params, rpcSigners)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		balA_usdc, err := GetWalletLedger(db, userAddressA).Balance(userAccountIDA, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(150), balA_usdc)

		balA_eth, err := GetWalletLedger(db, userAddressA).Balance(userAccountIDA, "eth")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(3), balA_eth)

		balB_usdc, err := GetWalletLedger(db, userAddressB).Balance(userAccountIDB, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(250), balB_usdc)

		balC_eth, err := GetWalletLedger(db, userAddressC).Balance(userAccountIDC, "eth")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(8), balC_eth)

		assert.Len(t, capturedNotifications, 3)
		for _, participant := range []string{userAddressA.Hex(), userAddressB.Hex(), userAddressC.Hex()} {
			assert.Contains(t, capturedNotifications, participant)
			assert.GreaterOrEqual(t, len(capturedNotifications[participant]), 1)
		}

		var depositTxs []LedgerTransaction
		err = db.Where("tx_type = ?", TransactionTypeAppDeposit).Find(&depositTxs).Error
		require.NoError(t, err)
		assert.Len(t, depositTxs, 4)
	})

	t.Run("NonIncreasedAllocationError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := &AppSession{
			SessionID:          "test-session-no-increase",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
			},
		}
		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		_, err := service.SubmitAppState(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: non-positive allocation sum delta")
	})

	t.Run("InsufficientBalanceError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		require.NoError(t, db.Create(&SignerWallet{Signer: userAddressA.Hex(), Wallet: userAddressA.Hex()}).Error)
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(userAccountIDA, "usdc", decimal.NewFromInt(30)))

		session := &AppSession{
			SessionID:          "test-session-insufficient",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, "usdc", decimal.NewFromInt(50)))

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}
		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		_, err := service.SubmitAppState(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: insufficient unified balance")
	})

	t.Run("OperateIntentNonZeroDeltaError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := &AppSession{
			SessionID:          "test-session-operate-error",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentOperate,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(80)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}
		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		_, err := service.SubmitAppState(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect operate request: non-zero allocation sum delta")
	})

	t.Run("ProtocolVersionValidationError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := &AppSession{
			SessionID:          "test-session-v02",
			Protocol:           rpc.VersionNitroRPCv0_2,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}
		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		_, err := service.SubmitAppState(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect request: specified parameters are not supported in this protocol")
	})

	t.Run("DepositorSignatureValidationError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := &AppSession{
			SessionID:          "test-session-signature",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
			},
		}
		rpcSigners := map[string]struct{}{

			userAddressB.Hex(): {},
		}

		_, err := service.SubmitAppState(params, rpcSigners)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "incorrect deposit request: quorum not reached")
	})

	t.Run("QuorumValidationError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := &AppSession{
			SessionID:          "test-session-quorum",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
			},
		}
		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
		}

		_, err := service.SubmitAppState(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: quorum not reached")
	})

	t.Run("UnsupportedIntentError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := &AppSession{
			SessionID:          "test-session-unsupported-intent",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       "unknown_intent",
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}
		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		_, err := service.SubmitAppState(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported intent: unknown_intent")
	})

	t.Run("DepositorSignatureRequired", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		require.NoError(t, db.Create(&SignerWallet{Signer: userAddressA.Hex(), Wallet: userAddressA.Hex()}).Error)
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(userAccountIDA, "usdc", decimal.NewFromInt(200)))

		session := &AppSession{
			SessionID:          "test-session-depositor-sig",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex(), userAddressC.Hex()},
			Weights:            []int64{1, 1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
			},
		}
		rpcSigners := map[string]struct{}{

			userAddressB.Hex(): {},
			userAddressC.Hex(): {},
		}

		_, err := service.SubmitAppState(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: depositor signature is required")
	})

	t.Run("QuorumMetButDepositorSignatureMissing", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := &AppSession{
			SessionID:          "test-session-ac7",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex(), userAddressC.Hex()},
			Weights:            []int64{1, 1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		userAccountIDA := NewAccountID(userAddressA.Hex())
		userAccountIDB := NewAccountID(userAddressB.Hex())
		userAccountIDC := NewAccountID(userAddressC.Hex())
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(userAccountIDA, "usdc", decimal.NewFromInt(500)))
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(userAccountIDB, "usdc", decimal.NewFromInt(300)))
		require.NoError(t, GetWalletLedger(db, userAddressC).Record(userAccountIDC, "usdc", decimal.NewFromInt(200)))

		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(sessionAccountID, "usdc", decimal.NewFromInt(50)))
		require.NoError(t, GetWalletLedger(db, userAddressC).Record(sessionAccountID, "usdc", decimal.NewFromInt(50)))

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))

		// UserA wants to deposit 100 more (from 100 to 200)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{ParticipantWallet: userAddressC.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		// Quorum is satisfied but depositor (userA) signature is missing
		rpcSigners := map[string]struct{}{
			userAddressB.Hex(): {},
			userAddressC.Hex(): {},
		}

		_, err := service.SubmitAppState(params, rpcSigners)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "incorrect deposit request: depositor signature is required")
		assert.NotContains(t, err.Error(), "quorum not reached")
	})

	t.Run("ZeroAllocationIncreaseError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := &AppSession{
			SessionID:          "test-session-zero-increase",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)}, // no change
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)}, // no change
			},
		}

		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {},
			userAddressB.Hex(): {},
		}

		_, err := service.SubmitAppState(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: non-positive allocation sum delta")
	})

	t.Run("MultipleDepositsSuccess", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := &AppSession{
			SessionID:          "test-session-mixed",
			Protocol:           rpc.VersionNitroRPCv0_4,
			ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex(), userAddressC.Hex()},
			Weights:            []int64{1, 1, 1},
			Quorum:             3,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)
		sessionAccountID := NewAccountID(session.SessionID)

		userAccountIDA := NewAccountID(userAddressA.Hex())
		userAccountIDB := NewAccountID(userAddressB.Hex())
		userAccountIDC := NewAccountID(userAddressC.Hex())
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(userAccountIDA, "usdc", decimal.NewFromInt(500)))
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(userAccountIDB, "usdc", decimal.NewFromInt(300)))
		require.NoError(t, GetWalletLedger(db, userAddressC).Record(userAccountIDC, "usdc", decimal.NewFromInt(200)))

		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, userAddressB).Record(sessionAccountID, "usdc", decimal.NewFromInt(50)))
		require.NoError(t, GetWalletLedger(db, userAddressC).Record(sessionAccountID, "usdc", decimal.NewFromInt(50)))

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, nil))

		// UserA deposits 50 more, UserB deposits 25 more, UserC no change
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)}, // +50
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(75)},  // +25
				{ParticipantWallet: userAddressC.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},  // no change
			},
		}

		rpcSigners := map[string]struct{}{
			userAddressA.Hex(): {}, // depositor
			userAddressB.Hex(): {}, // depositor
			userAddressC.Hex(): {}, // non-depositor but needed for quorum
		}

		resp, err := service.SubmitAppState(params, rpcSigners)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		ledgerA := GetWalletLedger(db, userAddressA)
		balanceA, _ := ledgerA.Balance(userAccountIDA, "usdc")
		assert.Equal(t, "450", balanceA.String())

		ledgerB := GetWalletLedger(db, userAddressB)
		balanceB, _ := ledgerB.Balance(userAccountIDB, "usdc")
		assert.Equal(t, "275", balanceB.String())

		ledgerC := GetWalletLedger(db, userAddressC)
		balanceC, _ := ledgerC.Balance(userAccountIDC, "usdc")
		assert.Equal(t, "200", balanceC.String())
	})
}
