package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var (
	rawA, _        = crypto.GenerateKey()
	rawB, _        = crypto.GenerateKey()
	rawC, _        = crypto.GenerateKey()
	signerA        = Signer{privateKey: rawA}
	signerB        = Signer{privateKey: rawB}
	signerC        = Signer{privateKey: rawC}
	userAddressA   = signerA.GetAddress()
	userAddressB   = signerB.GetAddress()
	userAddressC   = signerC.GetAddress()
	userAccountIDA = NewAccountID(userAddressA.Hex())
	userAccountIDB = NewAccountID(userAddressB.Hex())
	userAccountIDC = NewAccountID(userAddressC.Hex())
)

func assertNotifications(t *testing.T, capturedNotifications map[string][]Notification, userID string, expectedCount int) {
	assert.Contains(t, capturedNotifications, userID)
	assert.Len(t, capturedNotifications[userID], expectedCount)
}

func setupWallets(t *testing.T, db *gorm.DB, funds map[common.Address]map[string]int) {
	for addr, assets := range funds {
		accountID := NewAccountID(addr.Hex())
		for asset, amount := range assets {
			require.NoError(t, GetWalletLedger(db, addr).Record(accountID, asset, decimal.NewFromInt(int64(amount)), nil))
		}
	}
}

func createTestAppSession(t *testing.T, db *gorm.DB, sessionID string, protocol rpc.Version, participants []string, weights []int64, quorum uint64) *AppSession {
	session := &AppSession{
		SessionID:          sessionID,
		Application:        "TestApp",
		Protocol:           protocol,
		ParticipantWallets: participants,
		Weights:            weights,
		Quorum:             quorum,
		Status:             ChannelStatusOpen,
		Version:            1,
	}
	require.NoError(t, db.Create(session).Error)
	return session
}

func createTestAppSessionService(db *gorm.DB, capturedNotifications map[string][]Notification) *AppSessionService {
	var notifyFunc func(userID string, method string, params RPCDataParams)
	if capturedNotifications != nil {
		notifyFunc = func(userID string, method string, params RPCDataParams) {
			capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}
	} else {
		notifyFunc = func(userID string, method string, params RPCDataParams) {}
	}
	return NewAppSessionService(db, NewWSNotifier(notifyFunc, nil))
}

func setupAppSessionBalances(t *testing.T, db *gorm.DB, sessionAccountID AccountID, balances map[common.Address]map[string]int) {
	for addr, assets := range balances {
		for asset, amount := range assets {
			require.NoError(t, GetWalletLedger(db, addr).Record(sessionAccountID, asset, decimal.NewFromInt(int64(amount)), nil))
		}
	}
}

func rpcSigners(addresses ...common.Address) map[string]struct{} {
	signers := make(map[string]struct{})
	for _, addr := range addresses {
		signers[addr.Hex()] = struct{}{}
	}
	return signers
}

func TestAppSessionService_CreateApplication(t *testing.T) {
	t.Run("SuccessfulCreateApplication", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 200},
		})

		capturedNotifications := make(map[string][]Notification)
		service := createTestAppSessionService(db, capturedNotifications)

		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           rpc.VersionNitroRPCv0_4,
				ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
				Weights:            []int64{1, 1},
				Quorum:             2,
				Challenge:          60,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
		}

		appSession, err := service.CreateAppSession(params, rpcSigners(userAddressA, userAddressB))
		require.NoError(t, err)
		assert.NotEmpty(t, appSession.AppSessionID)
		assert.Equal(t, uint64(1), appSession.Version)
		assert.Equal(t, string(ChannelStatusOpen), appSession.Status)

		assertNotifications(t, capturedNotifications, userAddressA.Hex(), 1)
		assertNotifications(t, capturedNotifications, userAddressB.Hex(), 1)

		sessionAccountID := NewAccountID(appSession.AppSessionID)
		balA, _ := GetWalletLedger(db, userAddressA).Balance(userAccountIDA, "usdc")
		appBalA, _ := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "usdc")
		assert.True(t, balA.IsZero())
		assert.Equal(t, decimal.NewFromInt(100), appBalA)

		var transactions []LedgerTransaction
		db.Where("tx_type = ?", TransactionTypeAppDeposit).Find(&transactions)
		assert.Len(t, transactions, 2)
	})

	t.Run("ErrorInsufficientFunds", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{userAddressA: {"usdc": 50}})
		service := createTestAppSessionService(db, nil)

		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           rpc.VersionNitroRPCv0_4,
				ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
				Weights:            []int64{1, 0},
				Quorum:             1,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		_, err := service.CreateAppSession(params, rpcSigners(userAddressA))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
	})

	t.Run("ErrorNegativeAllocation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{userAddressA: {"usdc": 100}})
		service := createTestAppSessionService(db, nil)

		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           rpc.VersionNitroRPCv0_4,
				ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
				Weights:            []int64{1, 0},
				Quorum:             1,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(-50)},
			},
		}

		_, err := service.CreateAppSession(params, rpcSigners(userAddressA))
		require.Error(t, err)
		assert.Contains(t, err.Error(), ErrNegativeAllocation)
	})

	t.Run("ErrorChallengedChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{userAddressA: {"usdc": 100}})
		db.Create(&Channel{Wallet: userAddressA.Hex(), Status: ChannelStatusChallenged})
		service := createTestAppSessionService(db, nil)

		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           rpc.VersionNitroRPCv0_4,
				ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
				Weights:            []int64{1, 0},
				Quorum:             1,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		_, err := service.CreateAppSession(params, rpcSigners(userAddressA))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has challenged channels")
	})

	t.Run("ErrorQuorumExceedsTotalWeights", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 100},
		})
		service := createTestAppSessionService(db, nil)

		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           rpc.VersionNitroRPCv0_4,
				ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
				Weights:            []int64{3, 5}, // Total weight is 8
				Quorum:             10,            // Quorum exceeds total weight (10 > 8)
				Challenge:          60,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		_, err := service.CreateAppSession(params, rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "target quorum (10) cannot be greater than total sum of weights (8)")
	})

	t.Run("SuccessQuorumEqualsToTotalWeights", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 100},
		})
		service := createTestAppSessionService(db, nil)

		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           rpc.VersionNitroRPCv0_4,
				ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
				Weights:            []int64{3, 5}, // Total weight is 8
				Quorum:             8,             // Quorum equals total weight (8 == 8) - should be valid
				Challenge:          60,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		appSession, err := service.CreateAppSession(params, rpcSigners(userAddressA, userAddressB))
		require.NoError(t, err)
		assert.NotEmpty(t, appSession.AppSessionID)
		assert.Equal(t, uint64(1), appSession.Version)
		assert.Equal(t, string(ChannelStatusOpen), appSession.Status)
	})

	t.Run("ErrorNonZeroChannelAllocation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{userAddressA: {"usdc": 100}})

		// Should error if depositor has non-zero allocation in any open channel
		ch := &Channel{
			ChannelID:   "0xChannel",
			Wallet:      userAddressA.Hex(),
			Participant: userAddressA.Hex(),
			Status:      ChannelStatusOpen,
			Token:       "0xTokenXYZ",
			Nonce:       1,
			RawAmount:   decimal.NewFromInt(1),
		}
		require.NoError(t, db.Create(ch).Error)

		service := createTestAppSessionService(db, nil)

		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Protocol:           rpc.VersionNitroRPCv0_4,
				ParticipantWallets: []string{userAddressA.Hex(), userAddressB.Hex()},
				Weights:            []int64{1, 0},
				Quorum:             1,
				Nonce:              uint64(time.Now().Unix()),
			},
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		_, err := service.CreateAppSession(params, rpcSigners(userAddressA))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation denied: non-zero allocation in 1 channel(s) detected owned by wallet "+userAddressA.Hex())
	})
}

func TestAppSessionService_SubmitAppState(t *testing.T) {
	t.Run("NitroRPCv0.2_OperateSuccess", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := createTestAppSessionService(db, nil)
		session := createTestAppSession(t, db, "test-session", rpc.VersionNitroRPCv0_2,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 0},
		})

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Version:      0,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		resp, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		appBalA, _ := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "usdc")
		appBalB, _ := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "usdc")
		assert.Equal(t, decimal.NewFromInt(50), appBalA)
		assert.Equal(t, decimal.NewFromInt(50), appBalB)
	})

	t.Run("NitroRPCv0.2_Operate_ErrorNegativeAllocation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := createTestAppSessionService(db, nil)
		session := createTestAppSession(t, db, "test-session-negative", rpc.VersionNitroRPCv0_2,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
		})

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(-50)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), ErrNegativeAllocation)
	})

	t.Run("NitroRPCv0.4_OperateSuccess", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := createTestAppSessionService(db, nil)
		session := createTestAppSession(t, db, "test-session-v04-operate", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 0},
		})

		// Shouldn't return error if operator has non-zero allocation in any open channel
		ch := &Channel{
			ChannelID:   "0xChannel",
			Wallet:      userAddressA.Hex(),
			Participant: userAddressA.Hex(),
			Status:      ChannelStatusOpen,
			Token:       "0xTokenXYZ",
			Nonce:       1,
			RawAmount:   decimal.NewFromInt(1),
		}
		require.NoError(t, db.Create(ch).Error)

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentOperate,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		resp, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		appBalA, _ := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "usdc")
		appBalB, _ := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "usdc")
		assert.Equal(t, decimal.NewFromInt(50), appBalA)
		assert.Equal(t, decimal.NewFromInt(50), appBalB)
	})

	t.Run("NitroRPCv0.4_OperateInvalidVersion", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := createTestAppSessionService(db, nil)
		session := createTestAppSession(t, db, "test-session-v04-invalid-version", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 0},
		})

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentOperate,
			Version:      3,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.Equal(t, fmt.Sprintf("incorrect app state: incorrect version: expected %d, got %d", 2, params.Version), err.Error())
	})

	t.Run("OperateIntentNonZeroDeltaError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-operate-error", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentOperate,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(80)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect operate request: non-zero allocations sum delta")
	})
	t.Run("UnsupportedIntentError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-unsupported-intent", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       "unknown_intent",
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect app state: unsupported intent: unknown_intent")
	})

	t.Run("NitroRPCv0.4_Operate_ZeroAllocations_Success", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := createTestAppSessionService(db, nil)
		session := createTestAppSession(t, db, "test-session-v04-operate-zero", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 0},
			userAddressB: {"usdc": 0},
		})

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentOperate,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
			},
		}

		resp, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		appBalA, _ := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "usdc")
		appBalB, _ := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "usdc")
		assert.True(t, appBalA.IsZero())
		assert.True(t, appBalB.IsZero())
	})

	t.Run("Operate_WithdrawnAsset_NotRequired", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := createTestAppSessionService(db, nil)
		session := createTestAppSession(t, db, "test-session-withdrawn-asset", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		// Step 1: Simulate that USDC was deposited and then fully withdrawn
		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
		})
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, "usdc", decimal.NewFromInt(-100), nil))

		// Step 2: Deposit ETH into the session
		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"eth": 100},
		})

		// Step 3: Operate on ETH only - should NOT require USDC allocations
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentOperate,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "eth", Amount: decimal.NewFromInt(50)},
				{Participant: userAddressB.Hex(), AssetSymbol: "eth", Amount: decimal.NewFromInt(50)},
			},
		}

		resp, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.NoError(t, err, "Should succeed without mentioning withdrawn USDC")
		assert.Equal(t, uint64(2), resp.Version)

		// Verify ETH was redistributed correctly
		appBalA, _ := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "eth")
		appBalB, _ := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "eth")
		assert.Equal(t, decimal.NewFromInt(50), appBalA)
		assert.Equal(t, decimal.NewFromInt(50), appBalB)

		// Verify USDC balance remains zero and wasn't affected
		usdcBalA, _ := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "usdc")
		usdcBalB, _ := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "usdc")
		assert.True(t, usdcBalA.IsZero())
		assert.True(t, usdcBalB.IsZero())
	})

	t.Run("Operate_AllocateWithdrawnAsset_Error", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := createTestAppSessionService(db, nil)
		session := createTestAppSession(t, db, "test-session-allocate-withdrawn", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		// Step 1: Simulate that USDC was deposited and then fully withdrawn
		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
		})
		require.NoError(t, GetWalletLedger(db, userAddressA).Record(sessionAccountID, "usdc", decimal.NewFromInt(-100), nil))

		// Step 2: Deposit ETH into the session
		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"eth": 100},
		})

		// Step 3: Try to operate on both ETH and USDC, but USDC has zero balance
		// This should fail because we're trying to allocate a withdrawn asset
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentOperate,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "eth", Amount: decimal.NewFromInt(50)},
				{Participant: userAddressB.Hex(), AssetSymbol: "eth", Amount: decimal.NewFromInt(50)},
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "asset usdc is not deposited into the app session")
	})

	t.Run("Operate_AllocateNeverDepositedAsset_Error", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := createTestAppSessionService(db, nil)
		session := createTestAppSession(t, db, "test-session-never-deposited", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		// Only deposit ETH into the session
		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"eth": 100},
		})

		// Try to operate on both ETH and USDC, but USDC was never deposited
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentOperate,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "eth", Amount: decimal.NewFromInt(50)},
				{Participant: userAddressB.Hex(), AssetSymbol: "eth", Amount: decimal.NewFromInt(50)},
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(30)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(70)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		// When asset was never deposited, it's not in appSessionBalance, so caught by second validation loop
		assert.Contains(t, err.Error(), "allocation references unknown asset usdc")
	})
}

func TestAppSessionService_SubmitAppStateDeposit(t *testing.T) {
	depositorAddress := userAddressA
	depositorAccountID := userAccountIDA

	t.Run("BasicDepositSuccess", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 200},
			userAddressB:     {"usdc": 100},
		})

		session := createTestAppSession(t, db, "test-session-deposit", rpc.VersionNitroRPCv0_4,
			[]string{depositorAddress.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 100},
			userAddressB:     {"usdc": 100},
		})

		capturedNotifications := make(map[string][]Notification)
		service := createTestAppSessionService(db, capturedNotifications)

		testSessionData := `{"state":"updated","counter":42}`
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			SessionData:  &testSessionData,
			Allocations: []AppAllocation{
				{Participant: depositorAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		resp, err := service.SubmitAppState(context.Background(), params, rpcSigners(depositorAddress, userAddressB), rpcSigners(depositorAddress, userAddressB))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)
		assert.Equal(t, string(ChannelStatusOpen), resp.Status)

		balA, _ := GetWalletLedger(db, depositorAddress).Balance(depositorAccountID, "usdc")
		appBalA, _ := GetWalletLedger(db, depositorAddress).Balance(sessionAccountID, "usdc")
		appBalB, _ := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "usdc")
		assert.Equal(t, decimal.NewFromInt(150), balA)
		assert.Equal(t, decimal.NewFromInt(150), appBalA)
		assert.Equal(t, decimal.NewFromInt(100), appBalB)

		assert.Contains(t, capturedNotifications, depositorAddress.Hex())
		assert.Contains(t, capturedNotifications, userAddressB.Hex())

		// Verify AppSession fields are not empty in the notification
		for _, notifications := range capturedNotifications {
			for _, notification := range notifications {
				if notification.eventType == AppSessionUpdateEventType {
					notificationData, ok := notification.data.(rpc.AppSessionUpdateNotification)
					require.True(t, ok, "notification data should be AppSessionUpdateNotification")

					assert.Equal(t, session.SessionID, notificationData.AppSession.AppSessionID, "AppSessionID should match")
					assert.Equal(t, string(ChannelStatusOpen), notificationData.AppSession.Status, "Status should be open")
					assert.Equal(t, rpc.VersionNitroRPCv0_4, notificationData.AppSession.Protocol, "Protocol should match")
					assert.Equal(t, []string{depositorAddress.Hex(), userAddressB.Hex()}, notificationData.AppSession.ParticipantWallets, "ParticipantWallets should match")
					assert.Equal(t, []int64{1, 1}, notificationData.AppSession.Weights, "Weights should match")
					assert.Equal(t, uint64(2), notificationData.AppSession.Quorum, "Quorum should match")
					assert.Equal(t, uint64(2), notificationData.AppSession.Version, "Version should be 2 after update")
					assert.Equal(t, testSessionData, notificationData.AppSession.SessionData, "SessionData should be updated")
					assert.NotEmpty(t, notificationData.AppSession.CreatedAt, "CreatedAt should not be empty")
					assert.NotEmpty(t, notificationData.AppSession.UpdatedAt, "UpdatedAt should not be empty")

					// Verify timestamps are properly formatted
					createdAt, err := time.Parse(time.RFC3339, notificationData.AppSession.CreatedAt)
					assert.NoError(t, err, "CreatedAt should be valid RFC3339 timestamp")
					updatedAt, err := time.Parse(time.RFC3339, notificationData.AppSession.UpdatedAt)
					assert.NoError(t, err, "UpdatedAt should be valid RFC3339 timestamp")
					assert.False(t, createdAt.IsZero(), "CreatedAt should have a valid time")
					assert.False(t, updatedAt.IsZero(), "UpdatedAt should have a valid time")
					assert.True(t, updatedAt.After(createdAt) || updatedAt.Equal(createdAt), "UpdatedAt should be >= CreatedAt")

					// Verify participant allocations
					require.Len(t, notificationData.ParticipantAllocations, 2, "Should have 2 participant allocations")
					totalAllocations := decimal.Zero
					for _, alloc := range notificationData.ParticipantAllocations {
						assert.NotEmpty(t, alloc.Participant, "ParticipantWallet should not be empty")
						assert.Equal(t, "usdc", alloc.AssetSymbol, "AssetSymbol should be usdc")
						assert.True(t, alloc.Amount.IsPositive(), "Amount should be positive")
						totalAllocations = totalAllocations.Add(alloc.Amount)
					}
					assert.Equal(t, decimal.NewFromInt(250), totalAllocations, "Total allocations should be 250")
				}
			}
		}

		var depositTx []LedgerTransaction
		db.Where("tx_type = ? AND from_account = ? AND asset_symbol = ?",
			TransactionTypeAppDeposit, depositorAddress.Hex(), "usdc").Find(&depositTx)
		assert.Len(t, depositTx, 1)
		assert.Equal(t, decimal.NewFromInt(50), depositTx[0].Amount)
	})

	t.Run("MultipleParticipantsTokens", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 200, "eth": 5},
			userAddressB:     {"usdc": 300},
			userAddressC:     {"eth": 10},
		})

		session := createTestAppSession(t, db, "test-session-multi-deposit", rpc.VersionNitroRPCv0_4,
			[]string{depositorAddress.Hex(), userAddressB.Hex(), userAddressC.Hex()}, []int64{1, 1, 1}, 3)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 100, "eth": 1},
			userAddressB:     {"usdc": 200},
			userAddressC:     {"eth": 3},
		})

		capturedNotifications := make(map[string][]Notification)
		service := createTestAppSessionService(db, capturedNotifications)

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: depositorAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
				{Participant: depositorAddress.Hex(), AssetSymbol: "eth", Amount: decimal.NewFromInt(3)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(250)},
				{Participant: userAddressC.Hex(), AssetSymbol: "eth", Amount: decimal.NewFromInt(5)},
			},
		}

		resp, err := service.SubmitAppState(context.Background(), params, rpcSigners(depositorAddress, userAddressB, userAddressC), rpcSigners(depositorAddress, userAddressB, userAddressC))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		balA_usdc, _ := GetWalletLedger(db, depositorAddress).Balance(depositorAccountID, "usdc")
		balA_eth, _ := GetWalletLedger(db, depositorAddress).Balance(depositorAccountID, "eth")
		balB_usdc, _ := GetWalletLedger(db, userAddressB).Balance(userAccountIDB, "usdc")
		balC_eth, _ := GetWalletLedger(db, userAddressC).Balance(userAccountIDC, "eth")
		assert.Equal(t, decimal.NewFromInt(150), balA_usdc)
		assert.Equal(t, decimal.NewFromInt(3), balA_eth)
		assert.Equal(t, decimal.NewFromInt(250), balB_usdc)
		assert.Equal(t, decimal.NewFromInt(8), balC_eth)

		assert.Len(t, capturedNotifications, 3)
		var depositTxs []LedgerTransaction
		db.Where("tx_type = ?", TransactionTypeAppDeposit).Find(&depositTxs)
		assert.Len(t, depositTxs, 4)
	})

	t.Run("NonIncreasedAllocationError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-no-increase", rpc.VersionNitroRPCv0_4,
			[]string{depositorAddress.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 100},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: depositorAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(depositorAddress, userAddressB), rpcSigners(depositorAddress, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: non-positive allocations sum delta")
	})

	t.Run("InsufficientBalanceError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 30},
		})

		session := createTestAppSession(t, db, "test-session-insufficient", rpc.VersionNitroRPCv0_4,
			[]string{depositorAddress.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 50},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: depositorAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(depositorAddress, userAddressB), rpcSigners(depositorAddress, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: insufficient unified balance")
	})

	t.Run("ProtocolVersionValidationError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-v02", rpc.VersionNitroRPCv0_2,
			[]string{depositorAddress.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Allocations: []AppAllocation{
				{Participant: depositorAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(depositorAddress, userAddressB), rpcSigners(depositorAddress, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect request: specified parameters are not supported in this protocol")
	})

	t.Run("Error_NoQuorumNoDepositorSignature", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-signature", rpc.VersionNitroRPCv0_4,
			[]string{depositorAddress.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 100},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: depositorAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressB), rpcSigners(userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: quorum not reached")
	})

	t.Run("Error_noQuorum", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-quorum", rpc.VersionNitroRPCv0_4,
			[]string{depositorAddress.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 100},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: depositorAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(depositorAddress), rpcSigners(depositorAddress))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: quorum not reached")
	})

	t.Run("Error_quorumReached_noDepositorSignature_noSessionKey", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 500},
			userAddressB:     {"usdc": 300},
			userAddressC:     {"usdc": 200},
		})

		session := createTestAppSession(t, db, "test-session-ac7", rpc.VersionNitroRPCv0_4,
			[]string{depositorAddress.Hex(), userAddressB.Hex(), userAddressC.Hex()}, []int64{1, 1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 100},
			userAddressB:     {"usdc": 50},
			userAddressC:     {"usdc": 50},
		})

		service := createTestAppSessionService(db, nil)
		// UserA wants to deposit 100 more (from 100 to 200)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: depositorAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{Participant: userAddressC.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		// Quorum is satisfied but depositor signature is still required when no session key is used
		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressB, userAddressC), rpcSigners(userAddressB, userAddressC))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: depositor signature is required")
		assert.NotContains(t, err.Error(), "quorum not reached")
	})

	t.Run("Success_quorumReached_withSessionKey", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Load session key cache
		err := loadSessionKeyCache(db)
		require.NoError(t, err)

		walletAddress := depositorAddress.Hex()
		sessionKeyAddr := common.HexToAddress("0x1234567890123456789012345678901234567890") // Session key for depositor
		sessionKeyAddress := sessionKeyAddr.Hex()

		// Create session key with spending allowance
		allowances := []Allowance{
			{Asset: "usdc", Amount: "1000"},
		}
		err = AddSessionKey(db, walletAddress, sessionKeyAddress, "TestApp", "trade", allowances, time.Now().Add(24*time.Hour))
		require.NoError(t, err)

		// Setup depositor wallet balance (no custody signer needed - using session key)
		depositorAccountID := NewAccountID(walletAddress)
		require.NoError(t, GetWalletLedger(db, depositorAddress).Record(depositorAccountID, "usdc", decimal.NewFromInt(500), nil))

		// Setup other participants as custody signers with balances
		setupWallets(t, db, map[common.Address]map[string]int{
			userAddressB: {"usdc": 300},
			userAddressC: {"usdc": 200},
		})

		session := createTestAppSession(t, db, "test-session-sk", rpc.VersionNitroRPCv0_4,
			[]string{depositorAddress.Hex(), userAddressB.Hex(), userAddressC.Hex()}, []int64{1, 1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 100},
			userAddressB:     {"usdc": 50},
			userAddressC:     {"usdc": 50},
		})

		service := createTestAppSessionService(db, nil)
		// UserA wants to deposit 100 more (from 100 to 200)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: depositorAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{Participant: userAddressC.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		// Quorum is satisfied and session key provides depositor authorization
		// rpcWallets contains the actual wallet addresses (session key maps to depositor wallet)
		// rpcSigners contains the actual signer addresses (session key address itself)
		_, err = service.SubmitAppState(context.Background(), params, rpcSigners(depositorAddress, userAddressB, userAddressC), rpcSigners(sessionKeyAddr, userAddressB, userAddressC))
		require.NoError(t, err, "Should succeed when using session key for depositor authorization")

		// Verify session key usage was updated
		spending, err := CalculateSessionKeySpending(db, sessionKeyAddress, "usdc")
		require.NoError(t, err)
		assert.Equal(t, "100", spending.String(), "Should track spending from deposit")
	})

	t.Run("ZeroAllocationIncreaseError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-zero-increase", rpc.VersionNitroRPCv0_4,
			[]string{depositorAddress.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 100},
			userAddressB:     {"usdc": 100},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: depositorAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)}, // no change
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},     // no change
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(depositorAddress, userAddressB), rpcSigners(depositorAddress, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: non-positive allocations sum delta")
	})

	t.Run("DecreasedAllocationError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 200},
			userAddressB:     {"usdc": 100},
		})

		session := createTestAppSession(t, db, "test-session-decreased", rpc.VersionNitroRPCv0_4,
			[]string{depositorAddress.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 100},
			userAddressB:     {"usdc": 50},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: depositorAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(80)}, // decrease from 100 to 80
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},     // no change
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(depositorAddress, userAddressB), rpcSigners(depositorAddress, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: decreased allocation for participant")
	})

	t.Run("MultipleDepositsSuccess", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 500},
			userAddressB:     {"usdc": 300},
			userAddressC:     {"usdc": 200},
		})

		session := createTestAppSession(t, db, "test-session-mixed", rpc.VersionNitroRPCv0_4,
			[]string{depositorAddress.Hex(), userAddressB.Hex(), userAddressC.Hex()}, []int64{1, 1, 1}, 3)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 100},
			userAddressB:     {"usdc": 50},
			userAddressC:     {"usdc": 50},
		})

		service := createTestAppSessionService(db, nil)
		// UserA deposits 50 more, UserB deposits 25 more, UserC no change
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: depositorAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)}, // +50
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(75)},      // +25
				{Participant: userAddressC.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},      // no change
			},
		}

		resp, err := service.SubmitAppState(context.Background(), params, rpcSigners(depositorAddress, userAddressB, userAddressC), rpcSigners(depositorAddress, userAddressB, userAddressC))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		balanceA, _ := GetWalletLedger(db, depositorAddress).Balance(userAccountIDA, "usdc")
		balanceB, _ := GetWalletLedger(db, userAddressB).Balance(userAccountIDB, "usdc")
		balanceC, _ := GetWalletLedger(db, userAddressC).Balance(userAccountIDC, "usdc")
		assert.Equal(t, "450", balanceA.String())
		assert.Equal(t, "275", balanceB.String())
		assert.Equal(t, "200", balanceC.String())
	})

	t.Run("ErrorNonZeroChannelAllocation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 200},
			userAddressB:     {"usdc": 100},
		})

		session := createTestAppSession(t, db, "test-session-deposit", rpc.VersionNitroRPCv0_4,
			[]string{depositorAddress.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			depositorAddress: {"usdc": 100},
			userAddressB:     {"usdc": 100},
		})

		capturedNotifications := make(map[string][]Notification)
		service := createTestAppSessionService(db, capturedNotifications)

		// Should error if depositor has non-zero allocation in any open channel
		ch := &Channel{
			ChannelID:   "0xChannel",
			Wallet:      depositorAddress.Hex(),
			Participant: depositorAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       "0xTokenXYZ",
			Nonce:       1,
			RawAmount:   decimal.NewFromInt(1),
		}
		require.NoError(t, db.Create(ch).Error)

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			SessionData:  nil,
			Allocations: []AppAllocation{
				{Participant: depositorAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(depositorAddress, userAddressB), rpcSigners(depositorAddress, userAddressB))
		require.Error(t, err)

		assert.Contains(t, err.Error(), "operation denied: non-zero allocation in 1 channel(s) detected owned by wallet "+depositorAddress.Hex())
	})
}

func TestAppSessionService_SubmitAppStateWithdraw(t *testing.T) {
	withdrawerAddress := userAddressA
	withdrawerAccountID := userAccountIDA
	t.Run("BasicWithdrawalSuccess", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			withdrawerAddress: {"usdc": 100},
			userAddressB:      {"usdc": 50},
		})

		session := createTestAppSession(t, db, "test-session-withdraw", rpc.VersionNitroRPCv0_4,
			[]string{withdrawerAddress.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			withdrawerAddress: {"usdc": 150},
			userAddressB:      {"usdc": 100},
		})

		capturedNotifications := make(map[string][]Notification)
		service := createTestAppSessionService(db, capturedNotifications)

		// Shouldn't return error if withdrawer has non-zero allocation in any open channel
		ch := &Channel{
			ChannelID:   "0xChannel",
			Wallet:      withdrawerAddress.Hex(),
			Participant: withdrawerAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       "0xTokenXYZ",
			Nonce:       1,
			RawAmount:   decimal.NewFromInt(1),
		}
		require.NoError(t, db.Create(ch).Error)

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentWithdraw,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: withdrawerAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		resp, err := service.SubmitAppState(context.Background(), params, rpcSigners(withdrawerAddress, userAddressB), rpcSigners(withdrawerAddress, userAddressB))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)
		assert.Equal(t, string(ChannelStatusOpen), resp.Status)

		withdrawerBalance, _ := GetWalletLedger(db, withdrawerAddress).Balance(withdrawerAccountID, "usdc")
		withdrawerAppBalance, _ := GetWalletLedger(db, withdrawerAddress).Balance(sessionAccountID, "usdc")
		participantAppBalance, _ := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "usdc")
		assert.Equal(t, decimal.NewFromInt(150), withdrawerBalance)    // 100 initial + 50 withdrawn
		assert.Equal(t, decimal.NewFromInt(100), withdrawerAppBalance) // 150 - 50
		assert.Equal(t, decimal.NewFromInt(100), participantAppBalance)

		assert.Contains(t, capturedNotifications, withdrawerAddress.Hex())
		assert.Contains(t, capturedNotifications, userAddressB.Hex())

		var withdrawTx []LedgerTransaction
		db.Where("tx_type = ? AND to_account = ? AND asset_symbol = ?",
			TransactionTypeAppWithdrawal, withdrawerAddress.Hex(), "usdc").Find(&withdrawTx)
		assert.Len(t, withdrawTx, 1)
		assert.Equal(t, decimal.NewFromInt(50), withdrawTx[0].Amount)
	})

	t.Run("MultipleParticipantsMultipleTokens", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			withdrawerAddress: {"usdc": 50, "eth": 2},
			userAddressB:      {"usdc": 100},
			userAddressC:      {"eth": 5},
		})

		session := createTestAppSession(t, db, "test-session-multi-withdraw", rpc.VersionNitroRPCv0_4,
			[]string{withdrawerAddress.Hex(), userAddressB.Hex(), userAddressC.Hex()}, []int64{1, 1, 1}, 3)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			withdrawerAddress: {"usdc": 150, "eth": 5},
			userAddressB:      {"usdc": 250},
			userAddressC:      {"eth": 8},
		})

		capturedNotifications := make(map[string][]Notification)
		service := createTestAppSessionService(db, capturedNotifications)

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentWithdraw,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: withdrawerAddress.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)}, // withdraw 50
				{Participant: withdrawerAddress.Hex(), AssetSymbol: "eth", Amount: decimal.NewFromInt(3)},    // withdraw 2
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},      // withdraw 50
				{Participant: userAddressC.Hex(), AssetSymbol: "eth", Amount: decimal.NewFromInt(5)},         // withdraw 3
			},
		}

		resp, err := service.SubmitAppState(context.Background(), params, rpcSigners(withdrawerAddress, userAddressB, userAddressC), rpcSigners(withdrawerAddress, userAddressB, userAddressC))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		withdrawerUsdcBalance, _ := GetWalletLedger(db, withdrawerAddress).Balance(withdrawerAccountID, "usdc")
		withdrawerEthBalance, _ := GetWalletLedger(db, withdrawerAddress).Balance(withdrawerAccountID, "eth")
		participant1UsdcBalance, _ := GetWalletLedger(db, userAddressB).Balance(userAccountIDB, "usdc")
		participant2EthBalance, _ := GetWalletLedger(db, userAddressC).Balance(userAccountIDC, "eth")
		assert.Equal(t, decimal.NewFromInt(100), withdrawerUsdcBalance)   // 50 + 50 withdrawn
		assert.Equal(t, decimal.NewFromInt(4), withdrawerEthBalance)      // 2 + 2 withdrawn
		assert.Equal(t, decimal.NewFromInt(150), participant1UsdcBalance) // 100 + 50 withdrawn
		assert.Equal(t, decimal.NewFromInt(8), participant2EthBalance)    // 5 + 3 withdrawn

		assert.Len(t, capturedNotifications, 3)
		var withdrawTxs []LedgerTransaction
		db.Where("tx_type = ?", TransactionTypeAppWithdrawal).Find(&withdrawTxs)
		assert.Len(t, withdrawTxs, 4)
	})

	t.Run("NonDecreasedAllocationError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-no-decrease", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentWithdraw,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect withdrawal request: non-negative allocations sum delta")
	})

	t.Run("NitroRPCv0_2ProtocolError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-v02", rpc.VersionNitroRPCv0_2,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentWithdraw,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect request: specified parameters are not supported in this protocol")
	})

	t.Run("IncreasedAllocationError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 50},
		})

		session := createTestAppSession(t, db, "test-session-increased", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 50},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentWithdraw,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(120)}, // increase from 100 to 120
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},  // no change
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect withdrawal request: increased allocation for participant")
	})

	t.Run("QuorumNotReachedError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-quorum", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex(), userAddressC.Hex()}, []int64{1, 1, 1}, 3)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentWithdraw,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB)) // Only 2 signers, need 3
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect withdrawal request: quorum not reached")
	})
}

func TestAppSessionService_CloseApplication(t *testing.T) {
	t.Run("SuccessfulCloseApplication", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		capturedNotifications := make(map[string][]Notification)
		service := createTestAppSessionService(db, capturedNotifications)

		session := createTestAppSession(t, db, "test-session-close", rpc.VersionNitroRPCv0_2,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 200},
		})

		// Shouldn't return error if withdrawer has non-zero allocation in any open channel
		ch := &Channel{
			ChannelID:   "0xChannel",
			Wallet:      userAddressA.Hex(),
			Participant: userAddressA.Hex(),
			Status:      ChannelStatusOpen,
			Token:       "0xTokenXYZ",
			Nonce:       1,
			RawAmount:   decimal.NewFromInt(1),
		}
		require.NoError(t, db.Create(ch).Error)

		params := &CloseAppSessionParams{
			AppSessionID: session.SessionID,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
		}

		resp, err := service.CloseApplication(params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		assertNotifications(t, capturedNotifications, userAddressA.Hex(), 1)
		assertNotifications(t, capturedNotifications, userAddressB.Hex(), 1)

		var closedSession AppSession
		db.First(&closedSession, "session_id = ?", session.SessionID)
		assert.Equal(t, ChannelStatusClosed, closedSession.Status)

		appBalA, _ := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "usdc")
		walletBalA, _ := GetWalletLedger(db, userAddressA).Balance(userAccountIDA, "usdc")
		assert.True(t, appBalA.IsZero())
		assert.Equal(t, decimal.NewFromInt(100), walletBalA)

		var transactions []LedgerTransaction
		db.Where("tx_type = ?", TransactionTypeAppWithdrawal).Find(&transactions)
		assert.Len(t, transactions, 2)
	})

	t.Run("SuccessfulCloseApplicationWithZeroAllocation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		capturedNotifications := make(map[string][]Notification)
		service := createTestAppSessionService(db, capturedNotifications)

		session := createTestAppSession(t, db, "test-session-close-zero", rpc.VersionNitroRPCv0_2,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 0},
			userAddressB: {"usdc": 0},
		})

		params := &CloseAppSessionParams{
			AppSessionID: session.SessionID,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
			},
		}

		resp, err := service.CloseApplication(params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		var closedSession AppSession
		db.First(&closedSession, "session_id = ?", session.SessionID)
		assert.Equal(t, ChannelStatusClosed, closedSession.Status)

		appBalA, _ := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "usdc")
		walletBalA, _ := GetWalletLedger(db, userAddressA).Balance(userAccountIDA, "usdc")
		assert.True(t, appBalA.IsZero())
		assert.True(t, walletBalA.IsZero())

		assert.Len(t, capturedNotifications, 0)
	})

	t.Run("ErrorNegativeAllocation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		service := createTestAppSessionService(db, nil)
		session := createTestAppSession(t, db, "test-session-close-negative", rpc.VersionNitroRPCv0_2,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 200},
		})

		params := &CloseAppSessionParams{
			AppSessionID: session.SessionID,
			Allocations: []AppAllocation{
				{Participant: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(-100)},
				{Participant: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(400)},
			},
		}

		_, err := service.CloseApplication(params, rpcSigners(userAddressA, userAddressB), rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), ErrNegativeAllocation)
	})
}

func TestAppSessionSessionKeySpendingValidation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	err := loadSessionKeyCache(db)
	require.NoError(t, err)

	walletAddress := userAddressA.Hex()
	sessionKeyAddr := common.HexToAddress("0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC") // Dedicated session key address
	sessionKeyAddress := sessionKeyAddr.Hex()

	allowances := []Allowance{
		{Asset: "usdc", Amount: "500"},
		{Asset: "eth", Amount: "2"},
	}
	err = AddSessionKey(db, walletAddress, sessionKeyAddress, "TestApp", "trade", allowances, time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	accountID := NewAccountID(walletAddress)
	require.NoError(t, GetWalletLedger(db, userAddressA).Record(accountID, "usdc", decimal.NewFromInt(1000), nil))
	require.NoError(t, GetWalletLedger(db, userAddressA).Record(accountID, "eth", decimal.NewFromInt(5), nil))

	accountIDB := NewAccountID(userAddressB.Hex())
	require.NoError(t, GetWalletLedger(db, userAddressB).Record(accountIDB, "usdc", decimal.NewFromInt(1000), nil))

	capturedNotifications := make(map[string][]Notification)
	service := createTestAppSessionService(db, capturedNotifications)

	t.Run("CreateAppSession_WithinSpendingCap_Success", func(t *testing.T) {
		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Application:        "TestApp",
				Protocol:           rpc.VersionNitroRPCv0_4,
				ParticipantWallets: []string{walletAddress, userAddressB.Hex()},
				Weights:            []int64{2, 0},
				Quorum:             2,
				Challenge:          60,
				Nonce:              1,
			},
			Allocations: []AppAllocation{
				{Participant: walletAddress, AssetSymbol: "usdc", Amount: decimal.NewFromInt(300)},
			},
		}

		_, err := service.CreateAppSession(params, rpcSigners(sessionKeyAddr)) // Sign with session key
		assert.NoError(t, err, "Should allow app session creation within spending cap")

		// Verify session key usage was updated
		spending, err := CalculateSessionKeySpending(db, sessionKeyAddress, "usdc")
		require.NoError(t, err)
		assert.Equal(t, "300", spending.String(), "Should track spending from app session creation")
	})

	t.Run("CreateAppSession_ExceedsSpendingCap_Fails", func(t *testing.T) {
		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Application:        "TestApp",
				Protocol:           rpc.VersionNitroRPCv0_4,
				ParticipantWallets: []string{walletAddress, userAddressB.Hex()},
				Weights:            []int64{2, 1},
				Quorum:             2,
				Challenge:          60,
				Nonce:              2,
			},
			Allocations: []AppAllocation{
				{Participant: walletAddress, AssetSymbol: "usdc", Amount: decimal.NewFromInt(300)}, // This would make total 600, exceeding 500 limit
			},
		}

		_, err := service.CreateAppSession(params, rpcSigners(sessionKeyAddr))
		assert.Error(t, err, "Should reject app session creation that exceeds spending cap")
		assert.Contains(t, err.Error(), "session key spending validation failed")
	})

	t.Run("CreateAppSession_UnsupportedAsset_Fails", func(t *testing.T) {
		params := &CreateAppSessionParams{
			Definition: AppDefinition{
				Application:        "TestApp",
				Protocol:           rpc.VersionNitroRPCv0_4,
				ParticipantWallets: []string{walletAddress, userAddressB.Hex()},
				Weights:            []int64{2, 0},
				Quorum:             2,
				Challenge:          60,
				Nonce:              3,
			},
			Allocations: []AppAllocation{
				{Participant: walletAddress, AssetSymbol: "btc", Amount: decimal.NewFromInt(1)}, // BTC not in allowances
			},
		}

		_, err := service.CreateAppSession(params, rpcSigners(sessionKeyAddr))
		assert.Error(t, err, "Should reject unsupported asset")
		assert.Contains(t, err.Error(), "not allowed in session key spending cap")
	})

	t.Run("AppSessionDeposit_WithinSpendingCap_Success", func(t *testing.T) {
		session := createTestAppSession(t, db, "test-session-deposit", rpc.VersionNitroRPCv0_4, []string{walletAddress}, []int64{2, 0}, 2)
		_ = NewAccountID(session.SessionID)

		// Deposit 100
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{Participant: walletAddress, AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		// rpcWallets contains wallet address, rpcSigners contains session key address
		_, err := service.SubmitAppState(context.Background(), params, rpcSigners(userAddressA), rpcSigners(sessionKeyAddr))
		assert.NoError(t, err, "Should allow deposit within spending cap")

		// Verify the total spending is now 400
		totalSpending, err := CalculateSessionKeySpending(db, sessionKeyAddress, "usdc")
		require.NoError(t, err)
		assert.Equal(t, "400", totalSpending.String(), "Should track total spending including deposits")
	})
}
