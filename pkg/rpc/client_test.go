package rpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/erc7824/nitrolite/pkg/sign"
)

// Test helpers
var (
	testCtx     = context.Background()
	fixedTime   = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	testWallet  = "0x1234"
	testWallet2 = "0x5678"
	testChainID = uint32(1)
	testToken   = "0xUSDC"
	testSymbol  = "USDC"
)

// setupClient creates a test client with mock dialer
func setupClient() (*rpc.Client, *MockDialer) {
	mockDialer := NewMockDialer()
	client := rpc.NewClient(mockDialer)
	return client, mockDialer
}

// createResponse creates an RPC response with the given data
func createResponse[T any](method rpc.Method, data T) (*rpc.Message, error) {
	params, err := rpc.NewPayload(data)
	if err != nil {
		return nil, err
	}
	res := rpc.NewResponse(0, string(method), params)
	return &res, nil
}

// registerSimpleHandler registers a handler that returns the given response
func registerSimpleHandler[T any](dialer *MockDialer, method rpc.Method, response T) {
	dialer.RegisterHandler(method, func(params rpc.Payload, publish MockNotificationPublisher) (*rpc.Message, error) {
		return createResponse(method, response)
	})
}

// registerErrorHandler registers a handler that returns an error response
func registerErrorHandler(dialer *MockDialer, method rpc.Method, errMsg string) {
	dialer.RegisterHandler(method, func(params rpc.Payload, publish MockNotificationPublisher) (*rpc.Message, error) {
		res := rpc.NewErrorResponse(0, errMsg)
		return &res, nil
	})
}

func TestClient_Ping(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	// Ping returns pong
	dialer.RegisterHandler(rpc.PingMethod, func(params rpc.Payload, publish MockNotificationPublisher) (*rpc.Message, error) {
		res := rpc.NewResponse(0, string(rpc.PongMethod), rpc.Payload{})
		return &res, nil
	})

	sigs, err := client.Ping(testCtx)
	assert.NoError(t, err)
	assert.Empty(t, sigs)
}

func TestClient_GetConfig(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	config := rpc.BrokerConfig{
		BrokerAddress: testWallet,
		Blockchains: []rpc.BlockchainInfo{{
			ID:                 testChainID,
			CustodyAddress:     "0xabc",
			AdjudicatorAddress: "0xdef",
		}},
	}

	registerSimpleHandler(dialer, rpc.GetConfigMethod, config)

	resp, sigs, err := client.GetConfig(testCtx)
	assert.NoError(t, err)
	assert.Empty(t, sigs)
	assert.Equal(t, rpc.GetConfigResponse(config), resp)
}

func TestClient_GetAssets(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	// Test data
	assets := []rpc.Asset{
		{Token: testToken, ChainID: testChainID, Symbol: testSymbol, Decimals: 6},
		{Token: "0xETH", ChainID: testChainID, Symbol: "ETH", Decimals: 18},
		{Token: "0xDAI", ChainID: 2, Symbol: "DAI", Decimals: 18},
	}

	// Handler with filtering logic
	dialer.RegisterHandler(rpc.GetAssetsMethod, func(params rpc.Payload, publish MockNotificationPublisher) (*rpc.Message, error) {
		var req rpc.GetAssetsRequest
		params.Translate(&req)

		filtered := assets
		if req.ChainID != nil {
			var result []rpc.Asset
			for _, a := range assets {
				if a.ChainID == *req.ChainID {
					result = append(result, a)
				}
			}
			filtered = result
		}

		return createResponse(rpc.GetAssetsMethod, rpc.GetAssetsResponse{Assets: filtered})
	})

	t.Run("no filter", func(t *testing.T) {
		resp, _, err := client.GetAssets(testCtx, rpc.GetAssetsRequest{})
		require.NoError(t, err)
		assert.Len(t, resp.Assets, 3)
	})

	t.Run("with chain filter", func(t *testing.T) {
		resp, _, err := client.GetAssets(testCtx, rpc.GetAssetsRequest{ChainID: &testChainID})
		require.NoError(t, err)
		assert.Len(t, resp.Assets, 2)
	})
}

func TestClient_Channels(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()
	amount := decimal.NewFromInt(0) // Zero amount for channel creation

	t.Run("create", func(t *testing.T) {
		expected := rpc.CreateChannelResponse{
			ChannelID: "ch123",
			State: rpc.UnsignedState{
				Intent: rpc.StateIntentInitialize, Version: 0,
				Allocations: []rpc.StateAllocation{{
					Participant: testWallet, TokenAddress: testToken, RawAmount: amount,
				}},
			},
			StateSignature: sign.Signature{},
		}
		registerSimpleHandler(dialer, rpc.CreateChannelMethod, expected)

		req := rpc.CreateChannelRequest{ChainID: testChainID, Token: testToken}
		resp, _, err := client.CreateChannel(testCtx, req)
		require.NoError(t, err)
		assert.Equal(t, expected, resp)
	})

	t.Run("list", func(t *testing.T) {
		channels := rpc.GetChannelsResponse{
			Channels: []rpc.Channel{{
				ChannelID: "ch123", Participant: testWallet, Status: rpc.ChannelStatusOpen,
				Token: testToken, ChainID: testChainID, RawAmount: amount,
			}},
		}
		registerSimpleHandler(dialer, rpc.GetChannelsMethod, channels)

		resp, _, err := client.GetChannels(testCtx, rpc.GetChannelsRequest{})
		require.NoError(t, err)
		assert.Len(t, resp.Channels, 1)
	})
}

func TestClient_Ledger(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	t.Run("balances", func(t *testing.T) {
		balances := rpc.GetLedgerBalancesResponse{
			LedgerBalances: []rpc.LedgerBalance{
				{Asset: testSymbol, Amount: decimal.NewFromInt(1000)},
				{Asset: "eth", Amount: decimal.NewFromInt(5)},
			},
		}
		registerSimpleHandler(dialer, rpc.GetLedgerBalancesMethod, balances)

		resp, _, err := client.GetLedgerBalances(testCtx, rpc.GetLedgerBalancesRequest{})
		require.NoError(t, err)
		assert.Len(t, resp.LedgerBalances, 2)
	})

	t.Run("transactions", func(t *testing.T) {
		dialer.RegisterHandler(rpc.GetLedgerTransactionsMethod, func(params rpc.Payload, publish MockNotificationPublisher) (*rpc.Message, error) {
			txns := rpc.GetLedgerTransactionsResponse{
				LedgerTransactions: []rpc.LedgerTransaction{{
					Id: 1, TxType: "transfer",
					FromAccount: "acc1", ToAccount: "acc2",
					Asset: testSymbol, Amount: decimal.NewFromInt(50),
					CreatedAt: fixedTime,
				}},
			}
			return createResponse(rpc.GetLedgerTransactionsMethod, txns)
		})

		resp, _, err := client.GetLedgerTransactions(testCtx, rpc.GetLedgerTransactionsRequest{})
		require.NoError(t, err)
		require.Len(t, resp.LedgerTransactions, 1)

		txn := resp.LedgerTransactions[0]
		assert.True(t, txn.Amount.Equal(decimal.NewFromInt(50)))
		assert.False(t, txn.CreatedAt.IsZero())
	})
}

func TestClient_Transfer(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	dialer.RegisterHandler(rpc.TransferMethod, func(params rpc.Payload, publish MockNotificationPublisher) (*rpc.Message, error) {
		// Async notification
		go func() {
			time.Sleep(10 * time.Millisecond)
			balanceUpdate := rpc.BalanceUpdateNotification{
				BalanceUpdates: []rpc.LedgerBalance{{Asset: testSymbol, Amount: decimal.NewFromInt(900)}},
			}
			notifParams, _ := rpc.NewPayload(balanceUpdate)
			publish(rpc.BalanceUpdateEvent, notifParams)
		}()

		txns := rpc.TransferResponse{
			Transactions: []rpc.LedgerTransaction{{
				Id: 1, TxType: "transfer",
				FromAccount: "acc1", ToAccount: "acc2",
				Asset: testSymbol, Amount: decimal.NewFromInt(100),
				CreatedAt: fixedTime,
			}},
		}
		return createResponse(rpc.TransferMethod, txns)
	})

	req := rpc.TransferRequest{
		Destination: testWallet2,
		Allocations: []rpc.TransferAllocation{{AssetSymbol: testSymbol, Amount: decimal.NewFromInt(100)}},
	}

	resp, _, err := client.Transfer(testCtx, req)
	require.NoError(t, err)
	assert.Len(t, resp.Transactions, 1)
}

func TestClient_AppSessions(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	appDef := rpc.AppDefinition{
		Protocol:           "game",
		ParticipantWallets: []string{testWallet, testWallet2},
		Weights:            []int64{1, 1}, Quorum: 2, Challenge: 3600, Nonce: 1,
	}

	t.Run("create", func(t *testing.T) {
		dialer.RegisterHandler(rpc.CreateAppSessionMethod, func(params rpc.Payload, publish MockNotificationPublisher) (*rpc.Message, error) {
			var req rpc.CreateAppSessionRequest
			params.Translate(&req)

			if len(req.Definition.ParticipantWallets) < 2 {
				return nil, errors.New("need at least 2 participants")
			}

			session := rpc.CreateAppSessionResponse{
				AppSessionID: "app123", Status: "open",
				ParticipantWallets: req.Definition.ParticipantWallets,
			}
			return createResponse(rpc.CreateAppSessionMethod, session)
		})

		req := rpc.CreateAppSessionRequest{
			Definition: appDef,
			Allocations: []rpc.AppAllocation{
				{Participant: testWallet, AssetSymbol: testSymbol, Amount: decimal.NewFromInt(100)},
				{Participant: testWallet2, AssetSymbol: testSymbol, Amount: decimal.NewFromInt(100)},
			},
		}

		resp, _, err := client.CreateAppSession(testCtx, req)

		require.NoError(t, err)
		assert.Equal(t, "app123", resp.AppSessionID)
		assert.Len(t, resp.ParticipantWallets, 2)
	})

	t.Run("list", func(t *testing.T) {
		sessions := rpc.GetAppSessionsResponse{
			AppSessions: []rpc.AppSession{{
				AppSessionID: "app123", Status: "open",
				ParticipantWallets: []string{testWallet, testWallet2},
			}},
		}
		registerSimpleHandler(dialer, rpc.GetAppSessionsMethod, sessions)

		resp, _, err := client.GetAppSessions(testCtx, rpc.GetAppSessionsRequest{})
		require.NoError(t, err)
		assert.Len(t, resp.AppSessions, 1)
	})
}

func TestClient_ErrorHandling(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	// No handler registered
	_, _, err := client.GetConfig(testCtx)
	assert.Contains(t, err.Error(), "method not found")

	// Handler returns error response
	registerErrorHandler(dialer, rpc.GetAssetsMethod, "internal server error")
	_, _, err = client.GetAssets(testCtx, rpc.GetAssetsRequest{})
	assert.Contains(t, err.Error(), "internal server error")
}

func TestClient_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	// Handler with delay
	dialer.RegisterHandler(rpc.GetAssetsMethod, func(params rpc.Payload, publish MockNotificationPublisher) (*rpc.Message, error) {
		time.Sleep(10 * time.Millisecond)
		assets := rpc.GetAssetsResponse{Assets: []rpc.Asset{{Token: testToken}}}
		return createResponse(rpc.GetAssetsMethod, assets)
	})

	// Run concurrent requests
	const numRequests = 10
	errs := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			_, _, err := client.GetAssets(testCtx, rpc.GetAssetsRequest{})
			errs <- err
		}()
	}

	// Verify all succeeded
	for i := 0; i < numRequests; i++ {
		assert.NoError(t, <-errs)
	}
}

// Additional test coverage for remaining methods
func TestClient_AdditionalMethods(t *testing.T) {
	t.Parallel()

	client, dialer := setupClient()

	t.Run("GetAppDefinition", func(t *testing.T) {
		def := rpc.GetAppDefinitionResponse{
			Protocol:           "game",
			ParticipantWallets: []string{testWallet, testWallet2},
		}
		registerSimpleHandler(dialer, rpc.GetAppDefinitionMethod, def)

		resp, _, err := client.GetAppDefinition(testCtx, rpc.GetAppDefinitionRequest{
			AppSessionID: "app123",
		})
		require.NoError(t, err)
		assert.Equal(t, def.Protocol, resp.Protocol)
	})

	t.Run("CloseChannel", func(t *testing.T) {
		closeResp := rpc.CloseChannelResponse{
			ChannelID: "ch123",
			State:     rpc.UnsignedState{Intent: rpc.StateIntentFinalize},
		}
		registerSimpleHandler(dialer, rpc.CloseChannelMethod, closeResp)

		req := rpc.CloseChannelRequest{
			ChannelID: "ch123", FundsDestination: testWallet,
		}

		resp, _, err := client.CloseChannel(testCtx, req)
		require.NoError(t, err)
		assert.Equal(t, rpc.StateIntentFinalize, resp.State.Intent)
	})

	t.Run("SubmitAppState", func(t *testing.T) {
		submit := rpc.SubmitAppStateResponse{
			AppSessionID: "app123", Version: 2,
		}
		registerSimpleHandler(dialer, rpc.SubmitAppStateMethod, submit)

		req := rpc.SubmitAppStateRequest{
			AppSessionID: "app123",
		}
		resp, _, err := client.SubmitAppState(testCtx, req)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), resp.Version)
	})

	t.Run("CloseAppSession", func(t *testing.T) {
		closeApp := rpc.CloseAppSessionResponse{
			AppSessionID: "app123", Status: "closed",
		}
		registerSimpleHandler(dialer, rpc.CloseAppSessionMethod, closeApp)

		req := rpc.CloseAppSessionRequest{
			AppSessionID: "app123",
		}

		resp, _, err := client.CloseAppSession(testCtx, req)
		require.NoError(t, err)
		assert.Equal(t, "closed", resp.Status)
	})
}
