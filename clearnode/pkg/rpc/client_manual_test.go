package rpc_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

const (
	uatWsRpcUrl = "wss://canarynet.yellow.com/ws"
)

func TestManualClient(t *testing.T) {
	walletPK := os.Getenv("TEST_WALLET_PK")
	if walletPK == "" {
		t.Skip("TEST_WALLET_PK not set, skipping manual client test")
	}
	walletSigner, err := sign.NewEthereumSigner(walletPK)
	require.NoError(t, err)
	fmt.Printf("Using wallet address: %s\n", walletSigner.PublicKey().Address().String())

	sessionPK := os.Getenv("TEST_SESSION_PK")
	if sessionPK == "" {
		t.Skip("TEST_SESSION_PK not set, skipping manual client test")
	}
	sessionSigner, err := sign.NewEthereumSigner(sessionPK)
	require.NoError(t, err)
	fmt.Printf("Using session address: %s\n", sessionSigner.PublicKey().Address().String())

	dialer := rpc.NewWebsocketDialer(rpc.DefaultWebsocketDialerConfig)
	client := rpc.NewClient(dialer)

	errCh := make(chan error, 1)
	handleError := func(err error) {
		errCh <- err
	}

	ctx, cancel := context.WithCancel(t.Context())
	err = client.Start(ctx, uatWsRpcUrl, handleError)
	require.NoError(t, err)

	assetSymbol := "ytest.usd"
	testAllowanceAmount := "100"

	var jwtToken string
	t.Run("Authenticate With Signature", func(t *testing.T) {
		authReq := rpc.AuthRequestRequest{
			Address:     walletSigner.PublicKey().Address().String(),
			SessionKey:  sessionSigner.PublicKey().Address().String(),
			Application: "TestClient",
			Allowances: []rpc.Allowance{
				{
					Asset:  assetSymbol,
					Amount: testAllowanceAmount,
				},
			},
			ExpiresAt: uint64(time.Now().Add(1 * time.Hour).Unix()),
			Scope:  "",
		}
		authRes, _, err := client.AuthWithSig(ctx, authReq, walletSigner)
		require.NoError(t, err)

		require.True(t, authRes.Success, "auth_sig_verify should succeed")
		require.NotEmpty(t, authRes.JwtToken, "jwt token should be set")
		jwtToken = authRes.JwtToken
	})

	cancel()
	fmt.Println("Context cancelled, restarting with JWT")

	ctx, cancel = context.WithCancel(t.Context())
	defer cancel()

	currentBalance := decimal.NewFromInt(-1)
	currentBalanceMu := sync.RWMutex{}

	client.HandleBalanceUpdateEvent(func(ctx context.Context, notif rpc.BalanceUpdateNotification, resSig []sign.Signature) {
		for _, ledgerBalance := range notif.BalanceUpdates {
			if ledgerBalance.Asset == assetSymbol {
				currentBalanceMu.Lock()
				defer currentBalanceMu.Unlock()

				currentBalance = ledgerBalance.Amount
				return
			}
		}
	})

	err = client.Start(ctx, uatWsRpcUrl, handleError)
	require.NoError(t, err)

	var appSessionID string
	appAllocationsV0_2 := []rpc.AppAllocation{
		{
			Participant: walletSigner.PublicKey().Address().String(),
			AssetSymbol: assetSymbol,
			Amount:      decimal.NewFromInt(1),
		},
	}
	appAllocationsV0_4_Original := []rpc.AppAllocation{
		{
			Participant: walletSigner.PublicKey().Address().String(),
			AssetSymbol: assetSymbol,
			Amount:      decimal.NewFromInt(1),
		},
	}
	appAllocationsV0_4_Deposited := []rpc.AppAllocation{
		{
			Participant: walletSigner.PublicKey().Address().String(),
			AssetSymbol: assetSymbol,
			Amount:      decimal.NewFromInt(2),
		},
	}

	tcs := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "GetConfig",
			fn: func(t *testing.T) {
				configRes, _, err := client.GetConfig(ctx)
				require.NoError(t, err)
				fmt.Printf("Blockchains: %+v\n", configRes.Blockchains)
			},
		},
		{
			name: "GetAssets",
			fn: func(t *testing.T) {
				assetsReq := rpc.GetAssetsRequest{}
				assetsRes, _, err := client.GetAssets(ctx, assetsReq)
				require.NoError(t, err)
				fmt.Printf("Assets: %+v\n", assetsRes.Assets)
			},
		},
		{
			name: "Authenticate With JWT",
			fn: func(t *testing.T) {
				authVerifyReq := rpc.AuthJWTVerifyRequest{
					JWT: jwtToken,
				}
				verifyRes, _, err := client.AuthJWTVerify(ctx, authVerifyReq)
				require.NoError(t, err)
				require.True(t, verifyRes.Success, "auth_jwt_verify should succeed")
				require.Equal(t, walletSigner.PublicKey().Address().String(), verifyRes.Address, "address should match")
				require.Equal(t, sessionSigner.PublicKey().Address().String(), verifyRes.SessionKey, "session key should match")
			},
		},
		{
			name: "GetUserTag",
			fn: func(t *testing.T) {
				userTagRes, _, err := client.GetUserTag(ctx)
				require.NoError(t, err)
				fmt.Printf("User Tag: %+v\n", userTagRes.Tag)
			},
		},
		{
			name: "GetSessionKeys_VerifyAllowances",
			fn: func(t *testing.T) {
				sessionKeysRes, _, err := client.GetSessionKeys(ctx, rpc.GetSessionKeysRequest{})
				require.NoError(t, err)
				require.NotEmpty(t, sessionKeysRes.SessionKeys, "should have at least one session key")

				// Find our TestClient session key
				var foundSessionKey *rpc.SessionKeyResponse
				for i := range sessionKeysRes.SessionKeys {
					if sessionKeysRes.SessionKeys[i].SessionKey == sessionSigner.PublicKey().Address().String() &&
						sessionKeysRes.SessionKeys[i].Application == "TestClient" {
						foundSessionKey = &sessionKeysRes.SessionKeys[i]
						break
					}
				}

				require.NotNil(t, foundSessionKey, "should find TestClient session key")
				require.Equal(t, "TestClient", foundSessionKey.Application)
				require.NotEmpty(t, foundSessionKey.Allowances, "should have allowances")

				// Verify the allowance for our test asset
				var foundAllowance *rpc.AllowanceUsage
				for i := range foundSessionKey.Allowances {
					if foundSessionKey.Allowances[i].Asset == assetSymbol {
						foundAllowance = &foundSessionKey.Allowances[i]
						break
					}
				}

				require.NotNil(t, foundAllowance, "should have allowance for test asset")
				require.Equal(t, assetSymbol, foundAllowance.Asset)
				expectedAllowance, err := decimal.NewFromString(testAllowanceAmount)
				require.NoError(t, err)
				assert.True(t, expectedAllowance.Equal(foundAllowance.Allowance),
					"allowance should be %s, got %s", testAllowanceAmount, foundAllowance.Allowance)
				assert.True(t, foundAllowance.Used.IsZero(),
					"allowance usage should start at 0, got %s", foundAllowance.Used)

				fmt.Printf("Session Key: %s\n", foundSessionKey.SessionKey)
				fmt.Printf("Application: %s\n", foundSessionKey.Application)
				fmt.Printf("Allowance: %s %s (Used: %s)\n",
					foundAllowance.Allowance, foundAllowance.Asset, foundAllowance.Used)
			},
		},
		{
			name: "CreateAppSession_v0_2",
			fn: func(t *testing.T) {
				currentBalanceMu.RLock()
				balanceBefore := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				createAppReq := rpc.CreateAppSessionRequest{
					Definition: rpc.AppDefinition{
						Protocol: rpc.VersionNitroRPCv0_2,
						ParticipantWallets: []string{
							walletSigner.PublicKey().Address().String(),
							sign.NewEthereumAddress(common.Address{}).Hex(),
						},
						Weights:   []int64{100, 0},
						Quorum:    100,
						Challenge: 86400,
						Nonce:     uint64(uuid.New().ID()),
					},
					Allocations: appAllocationsV0_2,
				}
				createAppPayload, err := client.PreparePayload(rpc.CreateAppSessionMethod, createAppReq)
				require.NoError(t, err)

				createAppHash, err := createAppPayload.Hash()
				require.NoError(t, err)

				createAppResSig, err := sessionSigner.Sign(createAppHash)
				require.NoError(t, err)

				createAppFullReq := rpc.NewRequest(
					createAppPayload,
					createAppResSig,
				)

				createAppRes, _, err := client.CreateAppSession(ctx, &createAppFullReq)
				require.NoError(t, err)
				fmt.Printf("App Session Created: %+v\n", createAppRes.AppSessionID)
				appSessionID = createAppRes.AppSessionID

				currentBalanceMu.RLock()
				balanceAfter := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				balanceDiff := balanceBefore - balanceAfter
				assert.Equal(t, int64(1), balanceDiff, "balance should decrease by 1 unit")
			},
		},
		{
			name: "SubmitAppState_v0_2",
			fn: func(t *testing.T) {
				currentBalanceMu.RLock()
				balanceBefore := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				testSessionData := "{\"test\": true}"
				updateAppReq := rpc.SubmitAppStateRequest{
					AppSessionID: appSessionID,
					Allocations:  appAllocationsV0_2,
					SessionData:  &testSessionData,
				}
				updateAppPayload, err := client.PreparePayload(rpc.SubmitAppStateMethod, updateAppReq)
				require.NoError(t, err)

				updateAppHash, err := updateAppPayload.Hash()
				require.NoError(t, err)

				updateAppResSig, err := sessionSigner.Sign(updateAppHash)
				require.NoError(t, err)

				updateAppFullReq := rpc.NewRequest(
					updateAppPayload,
					updateAppResSig,
				)

				updateAppRes, _, err := client.SubmitAppState(ctx, &updateAppFullReq)
				require.NoError(t, err)
				fmt.Printf("App Session Version Updated: %+v\n", updateAppRes.Version)

				currentBalanceMu.RLock()
				balanceAfter := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				balanceDiff := balanceBefore - balanceAfter
				assert.Equal(t, int64(0), balanceDiff, "balance should not change")
			},
		},
		{
			name: "CloseAppSession_v0_2",
			fn: func(t *testing.T) {
				currentBalanceMu.RLock()
				balanceBefore := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				closeAppReq := rpc.CloseAppSessionRequest{
					AppSessionID: appSessionID,
					Allocations:  appAllocationsV0_2,
				}
				closeAppPayload, err := client.PreparePayload(rpc.CloseAppSessionMethod, closeAppReq)
				require.NoError(t, err)

				closeAppHash, err := closeAppPayload.Hash()
				require.NoError(t, err)

				closeAppResSig, err := sessionSigner.Sign(closeAppHash)
				require.NoError(t, err)

				closeAppFullReq := rpc.NewRequest(
					closeAppPayload,
					closeAppResSig,
				)

				closeAppRes, _, err := client.CloseAppSession(ctx, &closeAppFullReq)
				require.NoError(t, err)
				fmt.Printf("App Session closed with Version : %+v\n", closeAppRes.Version)

				currentBalanceMu.RLock()
				balanceAfter := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				balanceDiff := balanceBefore - balanceAfter
				assert.Equal(t, int64(-1), balanceDiff, "balance should increase by 1 unit")
			},
		},
		{
			name: "CreateAppSession_v0_4",
			fn: func(t *testing.T) {
				currentBalanceMu.RLock()
				balanceBefore := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				createAppReq := rpc.CreateAppSessionRequest{
					Definition: rpc.AppDefinition{
						Protocol: rpc.VersionNitroRPCv0_4,
						ParticipantWallets: []string{
							walletSigner.PublicKey().Address().String(),
							sign.NewEthereumAddress(common.Address{}).Hex(),
						},
						Weights:   []int64{100, 0},
						Quorum:    100,
						Challenge: 86400,
						Nonce:     uint64(uuid.New().ID()),
					},
					Allocations: appAllocationsV0_4_Original,
				}
				createAppPayload, err := client.PreparePayload(rpc.CreateAppSessionMethod, createAppReq)
				require.NoError(t, err)

				createAppHash, err := createAppPayload.Hash()
				require.NoError(t, err)

				createAppResSig, err := sessionSigner.Sign(createAppHash)
				require.NoError(t, err)

				createAppFullReq := rpc.NewRequest(
					createAppPayload,
					createAppResSig,
				)

				createAppRes, _, err := client.CreateAppSession(ctx, &createAppFullReq)
				require.NoError(t, err)
				fmt.Printf("App Session Created: %+v\n", createAppRes.AppSessionID)
				appSessionID = createAppRes.AppSessionID

				currentBalanceMu.RLock()
				balanceAfter := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				balanceDiff := balanceBefore - balanceAfter
				assert.Equal(t, int64(1), balanceDiff, "balance should decrease by 1 unit")

				// Verify allowance was reduced by 1
				sessionKeysRes, _, err := client.GetSessionKeys(ctx, rpc.GetSessionKeysRequest{})
				require.NoError(t, err)
				var foundSessionKey *rpc.SessionKeyResponse
				for i := range sessionKeysRes.SessionKeys {
					if sessionKeysRes.SessionKeys[i].SessionKey == sessionSigner.PublicKey().Address().String() {
						foundSessionKey = &sessionKeysRes.SessionKeys[i]
						break
					}
				}
				require.NotNil(t, foundSessionKey, "should find session key")
				var foundAllowance *rpc.AllowanceUsage
				for i := range foundSessionKey.Allowances {
					if foundSessionKey.Allowances[i].Asset == assetSymbol {
						foundAllowance = &foundSessionKey.Allowances[i]
						break
					}
				}
				require.NotNil(t, foundAllowance, "should have allowance")
				assert.True(t, decimal.NewFromInt(1).Equal(foundAllowance.Used),
					"allowance usage should be 1 after creating app session, got %s", foundAllowance.Used)
				fmt.Printf("Allowance after CreateAppSession: %s / %s used\n",
					foundAllowance.Used, foundAllowance.Allowance)
			},
		},
		{
			name: "SubmitAppState_v0_4_Operate",
			fn: func(t *testing.T) {
				currentBalanceMu.RLock()
				balanceBefore := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				testSessionData := "{\"test\": true}"
				updateAppReq := rpc.SubmitAppStateRequest{
					AppSessionID: appSessionID,
					Intent:       rpc.AppSessionIntentOperate,
					Version:      2,
					Allocations:  appAllocationsV0_4_Original,
					SessionData:  &testSessionData,
				}
				updateAppPayload, err := client.PreparePayload(rpc.SubmitAppStateMethod, updateAppReq)
				require.NoError(t, err)

				updateAppHash, err := updateAppPayload.Hash()
				require.NoError(t, err)

				updateAppResSig, err := sessionSigner.Sign(updateAppHash)
				require.NoError(t, err)

				updateAppFullReq := rpc.NewRequest(
					updateAppPayload,
					updateAppResSig,
				)

				updateAppRes, _, err := client.SubmitAppState(ctx, &updateAppFullReq)
				require.NoError(t, err)
				fmt.Printf("App Session Version Updated: %+v\n", updateAppRes.Version)

				currentBalanceMu.RLock()
				balanceAfter := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				balanceDiff := balanceBefore - balanceAfter
				assert.Equal(t, int64(0), balanceDiff, "balance should not change")
			},
		},
		{
			name: "SubmitAppState_v0_4_Deposit",
			fn: func(t *testing.T) {
				currentBalanceMu.RLock()
				balanceBefore := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				updateAppReq := rpc.SubmitAppStateRequest{
					AppSessionID: appSessionID,
					Intent:       rpc.AppSessionIntentDeposit,
					Version:      3,
					Allocations:  appAllocationsV0_4_Deposited,
				}
				updateAppPayload, err := client.PreparePayload(rpc.SubmitAppStateMethod, updateAppReq)
				require.NoError(t, err)

				updateAppHash, err := updateAppPayload.Hash()
				require.NoError(t, err)

				updateAppResSig, err := sessionSigner.Sign(updateAppHash)
				require.NoError(t, err)

				updateAppFullReq := rpc.NewRequest(
					updateAppPayload,
					updateAppResSig,
				)

				updateAppRes, _, err := client.SubmitAppState(ctx, &updateAppFullReq)
				require.NoError(t, err)
				fmt.Printf("App Session Version Updated: %+v\n", updateAppRes.Version)

				currentBalanceMu.RLock()
				balanceAfter := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				balanceDiff := balanceBefore - balanceAfter
				assert.Equal(t, int64(1), balanceDiff, "balance should decrease by 1 unit")

				// Verify allowance was reduced by another 1 (total 2)
				sessionKeysRes, _, err := client.GetSessionKeys(ctx, rpc.GetSessionKeysRequest{})
				require.NoError(t, err)
				var foundSessionKey *rpc.SessionKeyResponse
				for i := range sessionKeysRes.SessionKeys {
					if sessionKeysRes.SessionKeys[i].SessionKey == sessionSigner.PublicKey().Address().String() {
						foundSessionKey = &sessionKeysRes.SessionKeys[i]
						break
					}
				}
				require.NotNil(t, foundSessionKey, "should find session key")
				var foundAllowance *rpc.AllowanceUsage
				for i := range foundSessionKey.Allowances {
					if foundSessionKey.Allowances[i].Asset == assetSymbol {
						foundAllowance = &foundSessionKey.Allowances[i]
						break
					}
				}
				require.NotNil(t, foundAllowance, "should have allowance")
				assert.True(t, decimal.NewFromInt(2).Equal(foundAllowance.Used),
					"allowance usage should be 2 after deposit, got %s", foundAllowance.Used)
				fmt.Printf("Allowance after Deposit: %s / %s used\n",
					foundAllowance.Used, foundAllowance.Allowance)
			},
		},
		{
			name: "SubmitAppState_v0_4_Withdraw",
			fn: func(t *testing.T) {
				currentBalanceMu.RLock()
				balanceBefore := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				updateAppReq := rpc.SubmitAppStateRequest{
					AppSessionID: appSessionID,
					Intent:       rpc.AppSessionIntentWithdraw,
					Version:      4,
					Allocations:  appAllocationsV0_4_Original,
				}
				updateAppPayload, err := client.PreparePayload(rpc.SubmitAppStateMethod, updateAppReq)
				require.NoError(t, err)

				updateAppHash, err := updateAppPayload.Hash()
				require.NoError(t, err)

				updateAppResSig, err := sessionSigner.Sign(updateAppHash)
				require.NoError(t, err)

				updateAppFullReq := rpc.NewRequest(
					updateAppPayload,
					updateAppResSig,
				)

				updateAppRes, _, err := client.SubmitAppState(ctx, &updateAppFullReq)
				require.NoError(t, err)
				fmt.Printf("App Session Version Updated: %+v\n", updateAppRes.Version)

				currentBalanceMu.RLock()
				balanceAfter := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				balanceDiff := balanceBefore - balanceAfter
				assert.Equal(t, int64(-1), balanceDiff, "balance should increase by 1 unit")
			},
		},
		{
			name: "CloseAppSession_v0_4",
			fn: func(t *testing.T) {
				currentBalanceMu.RLock()
				balanceBefore := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				closeAppReq := rpc.CloseAppSessionRequest{
					AppSessionID: appSessionID,
					Allocations:  appAllocationsV0_4_Original,
				}
				closeAppPayload, err := client.PreparePayload(rpc.CloseAppSessionMethod, closeAppReq)
				require.NoError(t, err)

				closeAppHash, err := closeAppPayload.Hash()
				require.NoError(t, err)

				closeAppResSig, err := sessionSigner.Sign(closeAppHash)
				require.NoError(t, err)

				closeAppFullReq := rpc.NewRequest(
					closeAppPayload,
					closeAppResSig,
				)

				closeAppRes, _, err := client.CloseAppSession(ctx, &closeAppFullReq)
				require.NoError(t, err)
				fmt.Printf("App Session closed with Version : %+v\n", closeAppRes.Version)

				currentBalanceMu.RLock()
				balanceAfter := currentBalance.IntPart()
				currentBalanceMu.RUnlock()

				balanceDiff := balanceBefore - balanceAfter
				assert.Equal(t, int64(-1), balanceDiff, "balance should increase by 1 unit")
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.fn(t)
		})
	}
}
