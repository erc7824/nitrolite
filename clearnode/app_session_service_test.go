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
	addrA := signerA.GetAddress().Hex()
	addrB := signerB.GetAddress().Hex()

	t.Run("SuccessfulCreateApplication", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Setup: Create wallets and fund them
		require.NoError(t, db.Create(&SignerWallet{Signer: addrA, Wallet: addrA}).Error)
		require.NoError(t, db.Create(&SignerWallet{Signer: addrB, Wallet: addrB}).Error)
		require.NoError(t, GetWalletLedger(db, addrA).Record(addrA, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, addrB).Record(addrB, "usdc", decimal.NewFromInt(200)))

		service := NewAppSessionService(db)
		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           "test-proto",
				ParticipantWallets: []string{addrA, addrB},
				Weights:            []int64{1, 1},
				Quorum:             2,
				Challenge:          60,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{ParticipantWallet: addrA, AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: addrB, AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
		}
		rpcSigners := map[string]struct{}{
			addrA: {},
			addrB: {},
		}

		appSession, err := service.CreateApplication(params, rpcSigners)
		require.NoError(t, err)
		assert.NotNil(t, appSession)
		assert.NotEmpty(t, appSession.SessionID)
		assert.Equal(t, uint64(1), appSession.Version)
		assert.Equal(t, ChannelStatusOpen, appSession.Status)

		// Verify balances
		balA, err := GetWalletLedger(db, addrA).Balance(addrA, "usdc")
		require.NoError(t, err)
		assert.True(t, balA.IsZero())

		appBalA, err := GetWalletLedger(db, addrA).Balance(appSession.SessionID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(100), appBalA)
	})

	t.Run("ErrorInsufficientFunds", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		require.NoError(t, db.Create(&SignerWallet{Signer: addrA, Wallet: addrA}).Error)
		require.NoError(t, GetWalletLedger(db, addrA).Record(addrA, "usdc", decimal.NewFromInt(50))) // Not enough

		service := NewAppSessionService(db)
		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           "test-proto",
				ParticipantWallets: []string{addrA, addrB},
				Weights:            []int64{1, 0},
				Quorum:             1,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{ParticipantWallet: addrA, AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}
		rpcSigners := map[string]struct{}{addrA: {}}

		_, err := service.CreateApplication(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
	})

	t.Run("ErrorChallengedChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		require.NoError(t, db.Create(&SignerWallet{Signer: addrA, Wallet: addrA}).Error)
		require.NoError(t, GetWalletLedger(db, addrA).Record(addrA, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, db.Create(&Channel{
			Wallet: addrA,
			Status: ChannelStatusChallenged,
		}).Error)

		service := NewAppSessionService(db)
		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           "test-proto",
				ParticipantWallets: []string{addrA, addrB},
				Weights:            []int64{1, 0},
				Quorum:             1,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{ParticipantWallet: addrA, AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}
		rpcSigners := map[string]struct{}{addrA: {}}

		_, err := service.CreateApplication(params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has challenged channels")
	})
}

func TestAppSessionService_SubmitState(t *testing.T) {
	rawA, _ := crypto.GenerateKey()
	rawB, _ := crypto.GenerateKey()
	signerA := Signer{privateKey: rawA}
	signerB := Signer{privateKey: rawB}
	addrA := signerA.GetAddress().Hex()
	addrB := signerB.GetAddress().Hex()

	t.Run("SuccessfulSubmitState", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := NewAppSessionService(db)
		session := &AppSession{
			SessionID:          "test-session",
			Protocol:           "test-proto",
			ParticipantWallets: []string{addrA, addrB},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)

		ledgerA := GetWalletLedger(db, addrA)
		require.NoError(t, ledgerA.Record(session.SessionID, "usdc", decimal.NewFromInt(100)))

		params := &SubmitStateParams{
			AppSessionID: session.SessionID,
			Allocations: []AppAllocation{
				{ParticipantWallet: addrA, AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{ParticipantWallet: addrB, AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		// Mock ledger balances for the app session
		require.NoError(t, GetWalletLedger(db, addrB).Record(session.SessionID, "usdc", decimal.NewFromInt(0)))

		rpcSigners := map[string]struct{}{
			addrA: {},
			addrB: {},
		}

		newVersion, err := service.SubmitState(params, rpcSigners)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), newVersion)

		// Verify balances
		appBalA, err := ledgerA.Balance(session.SessionID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(50), appBalA)

		appBalB, err := GetWalletLedger(db, addrB).Balance(session.SessionID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(50), appBalB)
	})
}

func TestAppSessionService_CloseApplication(t *testing.T) {
	rawA, _ := crypto.GenerateKey()
	rawB, _ := crypto.GenerateKey()
	signerA := Signer{privateKey: rawA}
	signerB := Signer{privateKey: rawB}
	addrA := signerA.GetAddress().Hex()
	addrB := signerB.GetAddress().Hex()

	t.Run("SuccessfulCloseApplication", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := NewAppSessionService(db)
		session := &AppSession{
			SessionID:          "test-session-close",
			Protocol:           "test-proto",
			ParticipantWallets: []string{addrA, addrB},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Status:             ChannelStatusOpen,
			Version:            1,
		}
		require.NoError(t, db.Create(session).Error)

		ledgerA := GetWalletLedger(db, addrA)
		ledgerB := GetWalletLedger(db, addrB)
		require.NoError(t, ledgerA.Record(session.SessionID, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, ledgerB.Record(session.SessionID, "usdc", decimal.NewFromInt(200)))

		params := &CloseAppSessionParams{
			AppSessionID: session.SessionID,
			Allocations: []AppAllocation{
				{ParticipantWallet: addrA, AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: addrB, AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
		}
		rpcSigners := map[string]struct{}{
			addrA: {},
			addrB: {},
		}

		newVersion, err := service.CloseApplication(params, rpcSigners)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), newVersion)

		var closedSession AppSession
		require.NoError(t, db.First(&closedSession, "session_id = ?", session.SessionID).Error)
		assert.Equal(t, ChannelStatusClosed, closedSession.Status)

		// Verify balances
		appBalA, err := ledgerA.Balance(session.SessionID, "usdc")
		require.NoError(t, err)
		assert.True(t, appBalA.IsZero())

		walletBalA, err := ledgerA.Balance(addrA, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(100), walletBalA)
	})
}
