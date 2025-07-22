package main

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

		var capturedNotifications []Notification
		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params ...any) {
			capturedNotifications = append(capturedNotifications, Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}, nil))

		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           "test-proto",
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
		assert.NotEmpty(t, appSession.SessionID)
		assert.Equal(t, uint64(1), appSession.Version)
		assert.Equal(t, ChannelStatusOpen, appSession.Status)

		sessionAccountID := NewAccountID(appSession.SessionID)

		assert.Len(t, capturedNotifications, 2)
		assert.Equal(t, capturedNotifications[0].userID, userAddressA.Hex())
		assert.Equal(t, capturedNotifications[1].userID, userAddressB.Hex())

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
			assert.Equal(t, appSession.SessionID, tx.ToAccount, "To account should be app session ID")
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

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params ...any) {}, nil))
		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           "test-proto",
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

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params ...any) {}, nil))
		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           "test-proto",
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

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params ...any) {}, nil))
		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           "test-proto",
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

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params ...any) {}, nil))
		session := &AppSession{
			SessionID:          "test-session",
			Protocol:           "test-proto",
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

		newVersion, err := service.SubmitAppState(params, rpcSigners)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), newVersion)

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

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params ...any) {}, nil))
		session := &AppSession{
			SessionID:          "test-session-negative",
			Protocol:           "test-proto",
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

		var capturedNotifications []Notification
		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params ...any) {
			capturedNotifications = append(capturedNotifications, Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}, nil))

		session := &AppSession{
			SessionID:          "test-session-close",
			Protocol:           "test-proto",
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

		newVersion, err := service.CloseApplication(params, rpcSigners)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), newVersion)

		assert.Len(t, capturedNotifications, 2)
		assert.Equal(t, capturedNotifications[0].userID, userAddressA.Hex())
		assert.Equal(t, capturedNotifications[1].userID, userAddressB.Hex())

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

		var capturedNotifications []Notification
		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params ...any) {
			capturedNotifications = append(capturedNotifications, Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}, nil))

		session := &AppSession{
			SessionID:          "test-session-close",
			Protocol:           "test-proto",
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

		newVersion, err := service.CloseApplication(params, rpcSigners)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), newVersion)

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

		service := NewAppSessionService(db, NewWSNotifier(func(userID string, method string, params ...any) {}, nil))
		session := &AppSession{
			SessionID:          "test-session-close-negative",
			Protocol:           "test-proto",
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
