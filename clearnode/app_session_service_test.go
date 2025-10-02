package main

import (
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
		require.NoError(t, db.Create(&SignerWallet{Signer: addr.Hex(), Wallet: addr.Hex()}).Error)
		accountID := NewAccountID(addr.Hex())
		for asset, amount := range assets {
			require.NoError(t, GetWalletLedger(db, addr).Record(accountID, asset, decimal.NewFromInt(int64(amount))))
		}
	}
}

func createTestAppSession(t *testing.T, db *gorm.DB, sessionID string, protocol rpc.Version, participants []string, weights []int64, quorum uint64) *AppSession {
	session := &AppSession{
		SessionID:          sessionID,
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
			require.NoError(t, GetWalletLedger(db, addr).Record(sessionAccountID, asset, decimal.NewFromInt(int64(amount))))
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

		appSession, err := service.CreateApplication(params, rpcSigners(userAddressA, userAddressB))
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

		_, err := service.CreateApplication(params, rpcSigners(userAddressA))
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

		_, err := service.CreateApplication(params, rpcSigners(userAddressA))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "negative allocation")
	})

	t.Run("ErrorChallengedChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{userAddressA: {"usdc": 100}})
		db.Create(&Channel{Wallet: userAddressA.Hex(), Status: ChannelStatusChallenged})
		service := createTestAppSessionService(db, nil)

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

		_, err := service.CreateApplication(params, rpcSigners(userAddressA))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has challenged channels")
	})
}

func TestAppSessionService_SubmitAppState(t *testing.T) {
	t.Run("SuccessfulSubmitAppState", func(t *testing.T) {
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
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		resp, err := service.SubmitAppState(params, rpcSigners(userAddressA, userAddressB))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		appBalA, _ := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "usdc")
		appBalB, _ := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "usdc")
		assert.Equal(t, decimal.NewFromInt(50), appBalA)
		assert.Equal(t, decimal.NewFromInt(50), appBalB)
	})

	t.Run("ErrorNegativeAllocation", func(t *testing.T) {
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
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(-50)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
			},
		}

		_, err := service.SubmitAppState(params, rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "negative allocation: -50 for asset usdc")
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

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentOperate,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		resp, err := service.SubmitAppState(params, rpcSigners(userAddressA, userAddressB))
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
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		_, err := service.SubmitAppState(params, rpcSigners(userAddressA, userAddressB))
		require.Equal(t, fmt.Sprintf("incorrect app state: incorrect version: expected %d, got %d", 2, params.Version), err.Error())
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

		params := &CloseAppSessionParams{
			AppSessionID: session.SessionID,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
		}

		resp, err := service.CloseApplication(params, rpcSigners(userAddressA, userAddressB))
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
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
			},
		}

		resp, err := service.CloseApplication(params, rpcSigners(userAddressA, userAddressB))
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
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(-100)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(400)},
			},
		}

		_, err := service.CloseApplication(params, rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "negative allocation: -100 for asset usdc")
	})
}

func TestAppSessionService_SubmitAppStateDeposit(t *testing.T) {

	t.Run("BasicDepositSuccess", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			userAddressA: {"usdc": 200},
			userAddressB: {"usdc": 100},
		})

		session := createTestAppSession(t, db, "test-session-deposit", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 100},
		})

		capturedNotifications := make(map[string][]Notification)
		service := createTestAppSessionService(db, capturedNotifications)

		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		resp, err := service.SubmitAppState(params, rpcSigners(userAddressA, userAddressB))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)
		assert.Equal(t, string(ChannelStatusOpen), resp.Status)

		balA, _ := GetWalletLedger(db, userAddressA).Balance(userAccountIDA, "usdc")
		appBalA, _ := GetWalletLedger(db, userAddressA).Balance(sessionAccountID, "usdc")
		appBalB, _ := GetWalletLedger(db, userAddressB).Balance(sessionAccountID, "usdc")
		assert.Equal(t, decimal.NewFromInt(150), balA)
		assert.Equal(t, decimal.NewFromInt(150), appBalA)
		assert.Equal(t, decimal.NewFromInt(100), appBalB)

		assert.Contains(t, capturedNotifications, userAddressA.Hex())
		assert.Contains(t, capturedNotifications, userAddressB.Hex())

		var depositTx []LedgerTransaction
		db.Where("tx_type = ? AND from_account = ? AND asset_symbol = ?",
			TransactionTypeAppDeposit, userAddressA.Hex(), "usdc").Find(&depositTx)
		assert.Len(t, depositTx, 1)
		assert.Equal(t, decimal.NewFromInt(50), depositTx[0].Amount)
	})

	t.Run("MultipleParticipantsTokens", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			userAddressA: {"usdc": 200, "eth": 5},
			userAddressB: {"usdc": 300},
			userAddressC: {"eth": 10},
		})

		session := createTestAppSession(t, db, "test-session-multi-deposit", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex(), userAddressC.Hex()}, []int64{1, 1, 1}, 3)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100, "eth": 1},
			userAddressB: {"usdc": 200},
			userAddressC: {"eth": 3},
		})

		capturedNotifications := make(map[string][]Notification)
		service := createTestAppSessionService(db, capturedNotifications)

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

		resp, err := service.SubmitAppState(params, rpcSigners(userAddressA, userAddressB, userAddressC))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		balA_usdc, _ := GetWalletLedger(db, userAddressA).Balance(userAccountIDA, "usdc")
		balA_eth, _ := GetWalletLedger(db, userAddressA).Balance(userAccountIDA, "eth")
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
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
			},
		}

		_, err := service.SubmitAppState(params, rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: non-positive allocation sum delta")
	})

	t.Run("InsufficientBalanceError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			userAddressA: {"usdc": 30},
		})

		session := createTestAppSession(t, db, "test-session-insufficient", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 50},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		_, err := service.SubmitAppState(params, rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: insufficient unified balance")
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
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(80)},
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(50)},
			},
		}

		_, err := service.SubmitAppState(params, rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect operate request: non-zero allocation sum delta")
	})

	t.Run("ProtocolVersionValidationError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-v02", rpc.VersionNitroRPCv0_2,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		_, err := service.SubmitAppState(params, rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect request: specified parameters are not supported in this protocol")
	})

	t.Run("DepositorSignatureValidationError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-signature", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
			},
		}

		_, err := service.SubmitAppState(params, rpcSigners(userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: quorum not reached")
	})

	t.Run("QuorumValidationError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-quorum", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
			},
		}

		_, err := service.SubmitAppState(params, rpcSigners(userAddressA))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: quorum not reached")
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
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
			},
		}

		_, err := service.SubmitAppState(params, rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported intent: unknown_intent")
	})

	t.Run("DepositorSignatureRequired", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			userAddressA: {"usdc": 200},
		})

		session := createTestAppSession(t, db, "test-session-depositor-sig", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex(), userAddressC.Hex()}, []int64{1, 1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(150)},
			},
		}

		_, err := service.SubmitAppState(params, rpcSigners(userAddressB, userAddressC))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: depositor signature is required")
	})

	t.Run("QuorumMetButDepositorSignatureMissing", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			userAddressA: {"usdc": 500},
			userAddressB: {"usdc": 300},
			userAddressC: {"usdc": 200},
		})

		session := createTestAppSession(t, db, "test-session-ac7", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex(), userAddressC.Hex()}, []int64{1, 1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 50},
			userAddressC: {"usdc": 50},
		})

		service := createTestAppSessionService(db, nil)
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
		_, err := service.SubmitAppState(params, rpcSigners(userAddressB, userAddressC))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: depositor signature is required")
		assert.NotContains(t, err.Error(), "quorum not reached")
	})

	t.Run("ZeroAllocationIncreaseError", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		session := createTestAppSession(t, db, "test-session-zero-increase", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex()}, []int64{1, 1}, 2)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 100},
		})

		service := createTestAppSessionService(db, nil)
		params := &SubmitAppStateParams{
			AppSessionID: session.SessionID,
			Intent:       rpc.AppSessionIntentDeposit,
			Version:      2,
			Allocations: []AppAllocation{
				{ParticipantWallet: userAddressA.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)}, // no change
				{ParticipantWallet: userAddressB.Hex(), AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)}, // no change
			},
		}

		_, err := service.SubmitAppState(params, rpcSigners(userAddressA, userAddressB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect deposit request: non-positive allocation sum delta")
	})

	t.Run("MultipleDepositsSuccess", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		setupWallets(t, db, map[common.Address]map[string]int{
			userAddressA: {"usdc": 500},
			userAddressB: {"usdc": 300},
			userAddressC: {"usdc": 200},
		})

		session := createTestAppSession(t, db, "test-session-mixed", rpc.VersionNitroRPCv0_4,
			[]string{userAddressA.Hex(), userAddressB.Hex(), userAddressC.Hex()}, []int64{1, 1, 1}, 3)
		sessionAccountID := NewAccountID(session.SessionID)

		setupAppSessionBalances(t, db, sessionAccountID, map[common.Address]map[string]int{
			userAddressA: {"usdc": 100},
			userAddressB: {"usdc": 50},
			userAddressC: {"usdc": 50},
		})

		service := createTestAppSessionService(db, nil)
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

		resp, err := service.SubmitAppState(params, rpcSigners(userAddressA, userAddressB, userAddressC))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)

		balanceA, _ := GetWalletLedger(db, userAddressA).Balance(userAccountIDA, "usdc")
		balanceB, _ := GetWalletLedger(db, userAddressB).Balance(userAccountIDB, "usdc")
		balanceC, _ := GetWalletLedger(db, userAddressC).Balance(userAccountIDC, "usdc")
		assert.Equal(t, "450", balanceA.String())
		assert.Equal(t, "275", balanceB.String())
		assert.Equal(t, "200", balanceC.String())
	})
}
