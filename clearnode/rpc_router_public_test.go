package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRPCContext(id int, method string, params RPCDataParams) *RPCContext {
	if params == nil {
		params = struct{}{}
	}

	return &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: uint64(id),
				Method:    method,
				Params:    params,
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []Signature{Signature([]byte("dummy-signature"))},
		},
	}
}

func TestRPCRouterHandlePing(t *testing.T) {
	t.Parallel()

	router, _, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	c := createRPCContext(1, "ping", nil)
	router.HandlePing(c)

	assertResponse(t, c, "pong")
}

func TestRPCRouterHandleGetConfig(t *testing.T) {
	t.Parallel()

	router, _, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	router.Config = &Config{
		blockchains: map[uint32]BlockchainConfig{
			137:   {ID: 137, BlockchainRPC: "https://polygon-mainnet.infura.io/v3/test", ContractAddresses: ContractAddressesConfig{Custody: "0xCustodyAddress1", Adjudicator: "0xAdjudicatorAddress1"}},
			42220: {ID: 42220, BlockchainRPC: "https://celo-mainnet.infura.io/v3/test", ContractAddresses: ContractAddressesConfig{Custody: "0xCustodyAddress2", Adjudicator: "0xAdjudicatorAddress2"}},
			8453:  {ID: 8453, BlockchainRPC: "https://base-mainnet.infura.io/v3/test", ContractAddresses: ContractAddressesConfig{Custody: "0xCustodyAddress3", Adjudicator: "0xAdjudicatorAddress3"}},
		},
	}

	ctx := createRPCContext(1, "get_config", map[string]interface{}{})
	router.HandleGetConfig(ctx)

	res := assertResponse(t, ctx, "get_config")
	configMap, ok := res.Params.(rpc.BrokerConfig)
	require.True(t, ok, "Response should contain a BrokerConfig")
	assert.Equal(t, router.Signer.GetAddress().Hex(), configMap.BrokerAddress)
	require.Len(t, configMap.Blockchains, 3, "Should have 3 supported blockchains")

	expectedBlockchains := map[uint32]struct{}{
		137:   {},
		42220: {},
		8453:  {},
	}
	for _, blockchain := range configMap.Blockchains {
		_, exists := expectedBlockchains[blockchain.ID]
		assert.True(t, exists, "Blockchain %d should be in expected blockchains", blockchain.ID)
		assert.Contains(t, blockchain.CustodyAddress, "0xCustodyAddress", "Custody address should be present")
		delete(expectedBlockchains, blockchain.ID)
	}
	assert.Empty(t, expectedBlockchains, "All expected blockchains should be found")
}

func TestRPCRouterHandleGetAssets(t *testing.T) {
	t.Parallel()

	router, _, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	testTokens := []TokenConfig{
		{
			Address:      "0xToken1",
			BlockchainID: 137,
			Decimals:     6,
			Symbol:       "usdc",
		},
		{
			Address:      "0xToken2",
			BlockchainID: 137,
			Decimals:     18,
			Symbol:       "weth",
		},
		{
			Address:      "0xToken3",
			BlockchainID: 42220,
			Decimals:     18,
			Symbol:       "celo",
		},
		{
			Address:      "0xToken4",
			BlockchainID: 8453,
			Decimals:     6,
			Symbol:       "usdbc",
		},
	}

	for _, token := range testTokens {
		seedAsset(t, &router.Config.assets,
			token.Address,
			token.BlockchainID,
			token.Symbol,
			token.Decimals)
	}

	tcs := []struct {
		name               string
		params             map[string]interface{}
		expectedTokenNames []string
	}{
		{"Get all", map[string]interface{}{}, []string{"0xToken1", "0xToken2", "0xToken3", "0xToken4"}},
		{"Filter by chain_id=137", map[string]interface{}{"chain_id": float64(137)}, []string{"0xToken1", "0xToken2"}},
		{"Filter by chain_id=42220", map[string]interface{}{"chain_id": float64(42220)}, []string{"0xToken3"}},
		{"Filter by non-existent chain_id=1", map[string]interface{}{"chain_id": float64(1)}, []string{}},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx := createRPCContext(1, "get_assets", tc.params)
			router.HandleGetAssets(ctx)

			res := assertResponse(t, ctx, "get_assets")
			responseAssets, ok := res.Params.(GetAssetsResponse)
			require.True(t, ok, "Response parameter should be a GetAssetsResponse")
			assert.Len(t, responseAssets.Assets, len(tc.expectedTokenNames), "Should return expected number of assets")

			for idx, asset := range responseAssets.Assets {
				assert.True(t, asset.Token == tc.expectedTokenNames[idx], "Should include token %s", tc.expectedTokenNames[idx])
			}
		})
	}
}

func TestRPCRouterHandleGetChannels(t *testing.T) {
	t.Run("RPCRouterHandleGetChannels", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		participantSigner := signer.GetAddress().Hex()
		participantWallet := "wallet_address"

		// Create channels with specific creation times to test sorting
		baseTime := time.Now().Add(-24 * time.Hour)
		channels := []Channel{
			{
				ChannelID:   "0xChannel1",
				Wallet:      participantWallet,
				Participant: participantSigner,
				Status:      ChannelStatusOpen,
				Nonce:       1,
				CreatedAt:   baseTime,
			},
			{
				ChannelID:   "0xChannel2",
				Wallet:      participantWallet,
				Participant: participantSigner,
				Status:      ChannelStatusClosed,
				Nonce:       2,
				CreatedAt:   baseTime.Add(1 * time.Hour),
			},
			{
				ChannelID:   "0xChannel3",
				Wallet:      participantWallet,
				Participant: participantSigner,
				Status:      ChannelStatusOpen,
				Nonce:       3,
				CreatedAt:   baseTime.Add(2 * time.Hour),
			},
			{
				ChannelID:   "0xOtherChannel",
				Wallet:      "other_wallet",
				Participant: "0xOtherParticipant",
				Status:      ChannelStatusOpen,
				Nonce:       4,
				CreatedAt:   baseTime.Add(3 * time.Hour),
			},
		}
		require.NoError(t, db.Create(channels).Error)
		tcs := []struct {
			name               string
			params             map[string]interface{}
			expectedChannelIDs []string
		}{
			{
				name:               "Get all with no sort (default desc by created_at)",
				params:             map[string]interface{}{},
				expectedChannelIDs: []string{"0xOtherChannel", "0xChannel3", "0xChannel2", "0xChannel1"},
			},
			{
				name:               "Get all with ascending sort",
				params:             map[string]interface{}{"sort": "asc"},
				expectedChannelIDs: []string{"0xChannel1", "0xChannel2", "0xChannel3", "0xOtherChannel"},
			},
			{
				name:               "Get all with descending sort",
				params:             map[string]interface{}{"sort": "desc"},
				expectedChannelIDs: []string{"0xOtherChannel", "0xChannel3", "0xChannel2", "0xChannel1"},
			},
			{
				name:               "Filter by participant",
				params:             map[string]interface{}{"participant": participantWallet},
				expectedChannelIDs: []string{"0xChannel3", "0xChannel2", "0xChannel1"},
			},
			{
				name:               "Filter by participant with ascending sort",
				params:             map[string]interface{}{"participant": participantWallet, "sort": "asc"},
				expectedChannelIDs: []string{"0xChannel1", "0xChannel2", "0xChannel3"},
			},
			{
				name:               "Filter by status open",
				params:             map[string]interface{}{"status": string(ChannelStatusOpen)},
				expectedChannelIDs: []string{"0xOtherChannel", "0xChannel3", "0xChannel1"},
			},
			{
				name:               "Filter by participant and status open",
				params:             map[string]interface{}{"participant": participantWallet, "status": string(ChannelStatusOpen)},
				expectedChannelIDs: []string{"0xChannel3", "0xChannel1"},
			},
			{
				name:               "Filter by participant and status closed",
				params:             map[string]interface{}{"participant": participantWallet, "status": string(ChannelStatusClosed)},
				expectedChannelIDs: []string{"0xChannel2"},
			},
			{
				name:               "Filter by status closed only",
				params:             map[string]interface{}{"status": string(ChannelStatusClosed)},
				expectedChannelIDs: []string{"0xChannel2"},
			},
		}

		for id, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				ctx := createRPCContext(id, "get_channels", tc.params)
				router.HandleGetChannels(ctx)

				res := assertResponse(t, ctx, "get_channels")
				responseChannels, ok := res.Params.(GetChannelsResponse)
				require.True(t, ok, "Response parameter should be a GetChannelsResponse")
				assert.Len(t, responseChannels.Channels, len(tc.expectedChannelIDs), "Should return expected number of channels")

				for idx, channel := range responseChannels.Channels {
					assert.True(t, channel.ChannelID == tc.expectedChannelIDs[idx], "%d-th result (%s) should equal %s", idx, channel.ChannelID, tc.expectedChannelIDs[idx])
				}
			})
		}
	})
	t.Run("RPCRouterHandleGetChannels_Pagination", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		defer cleanup()

		channelIDs := []string{
			"0xChannel01", "0xChannel02", "0xChannel03", "0xChannel04",
			"0xChannel05", "0xChannel06", "0xChannel07", "0xChannel08",
			"0xChannel09", "0xChannel10", "0xChannel11"}

		testChannels := []Channel{
			{Wallet: "0xWallet1", Participant: "0xParticipant1", Status: ChannelStatusOpen, Nonce: 1},
			{Wallet: "0xWallet2", Participant: "0xParticipant2", Status: ChannelStatusClosed, Nonce: 2},
			{Wallet: "0xWallet3", Participant: "0xParticipant3", Status: ChannelStatusOpen, Nonce: 3},
			{Wallet: "0xWallet5", Participant: "0xParticipant5", Status: ChannelStatusOpen, Nonce: 4},
			{Wallet: "0xWallet6", Participant: "0xParticipant6", Status: ChannelStatusChallenged, Nonce: 5},
			{Wallet: "0xWallet7", Participant: "0xParticipant7", Status: ChannelStatusOpen, Nonce: 6},
			{Wallet: "0xWallet8", Participant: "0xParticipant8", Status: ChannelStatusClosed, Nonce: 7},
			{Wallet: "0xWallet9", Participant: "0xParticipant9", Status: ChannelStatusOpen, Nonce: 8},
			{Wallet: "0xWallet10", Participant: "0xParticipant10", Status: ChannelStatusOpen, Nonce: 9},
			{Wallet: "0xWallet11", Participant: "0xParticipant11", Status: ChannelStatusOpen, Nonce: 10},
			{Wallet: "0xWallet12", Participant: "0xParticipant12", Status: ChannelStatusOpen, Nonce: 11},
		}

		for i := range testChannels {
			testChannels[i].ChannelID = channelIDs[i]
			// Stagger creation times in descending order, so that default sort returns them in `channelIDs` order
			testChannels[i].CreatedAt = time.Now().Add(time.Duration(1)*time.Hour - time.Duration(i)*time.Minute)
		}

		for _, channel := range testChannels {
			require.NoError(t, db.Create(&channel).Error)
		}

		tcs := []struct {
			name               string
			params             map[string]interface{}
			expectedChannelIDs []string
		}{
			{name: "No params",
				params:             map[string]interface{}{},
				expectedChannelIDs: channelIDs[:10], // Default pagination with desc sort
			},
			{name: "Offset only",
				params:             map[string]interface{}{"offset": float64(2)},
				expectedChannelIDs: channelIDs[2:], // Skip first 2
			},
			{name: "Limit only",
				params:             map[string]interface{}{"limit": float64(5)},
				expectedChannelIDs: channelIDs[:5], // First 5 channels
			},
			{name: "Offset and limit",
				params:             map[string]interface{}{"offset": float64(2), "limit": float64(3)},
				expectedChannelIDs: channelIDs[2:5], // Skip 2, take 3
			},
			{name: "Pagination with sort asc",
				params:             map[string]interface{}{"offset": float64(1), "limit": float64(3), "sort": "asc"},
				expectedChannelIDs: []string{"0xChannel10", "0xChannel09", "0xChannel08"}, // Ascending order, skip 1, take 3
			},
			{name: "Pagination with status filter",
				params:             map[string]interface{}{"status": "open", "limit": float64(3)},
				expectedChannelIDs: []string{"0xChannel01", "0xChannel03", "0xChannel04"}, // Only open channels, first 3
			},
		}

		for idx, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				ctx := createRPCContext(idx, "get_channels", tc.params)
				router.HandleGetChannels(ctx)

				res := assertResponse(t, ctx, "get_channels")
				responseChannels, ok := res.Params.(GetChannelsResponse)
				require.True(t, ok, "Response parameter should be a GetChannelsResponse")
				assert.Len(t, responseChannels.Channels, len(tc.expectedChannelIDs), "Should return expected number of channels")

				// Check channel IDs are included in expected order
				for idx, channel := range responseChannels.Channels {
					assert.Equal(t, tc.expectedChannelIDs[idx], channel.ChannelID, "Should include channel %s at position %d", tc.expectedChannelIDs[idx], idx)
				}
			})
		}
	})
}

func TestRPCRouterHandleGetAppDefinition(t *testing.T) {
	t.Run("RPCRouterHandleGetAppDefinition", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		defer cleanup()

		// Seed an AppSession
		session := AppSession{
			SessionID:          "0xSess123",
			ParticipantWallets: []string{"0xA", "0xB"},
			Protocol:           "proto",
			Weights:            []int64{10, 20},
			Quorum:             15,
			Challenge:          30,
			Nonce:              99,
		}
		require.NoError(t, db.Create(&session).Error)

		ctx := createRPCContext(5, "get_app_definition", map[string]string{"app_session_id": session.SessionID})
		router.HandleGetAppDefinition(ctx)

		res := assertResponse(t, ctx, "get_app_definition")
		def, ok := res.Params.(AppDefinition)
		require.True(t, ok)
		assert.Equal(t, session.Protocol, def.Protocol)
		assert.EqualValues(t, session.ParticipantWallets, def.ParticipantWallets)
		assert.EqualValues(t, session.Weights, def.Weights)
		assert.Equal(t, session.Quorum, def.Quorum)
		assert.Equal(t, session.Challenge, def.Challenge)
		assert.Equal(t, session.Nonce, def.Nonce)
	})
	t.Run("RPCRouterHandleGetAppDefinition_MissingID", func(t *testing.T) {
		router, _, cleanup := setupTestRPCRouter(t)
		defer cleanup()

		ctx := createRPCContext(6, "get_app_definition", nil)
		router.HandleGetAppDefinition(ctx)

		assertErrorResponse(t, ctx, "missing account ID")
	})

	t.Run("TestRPCRouterHandleGetAppDefinition_NotFound", func(t *testing.T) {
		router, _, cleanup := setupTestRPCRouter(t)
		defer cleanup()

		ctx := createRPCContext(6, "get_app_definition", map[string]string{"app_session_id": "nonexistent"})

		router.HandleGetAppDefinition(ctx)
		assertErrorResponse(t, ctx, "failed to get application session")
	})
}

func TestRPCRouterHandleGetAppSessions(t *testing.T) {
	t.Run("RPCRouterHandleGetAppSessions", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		defer cleanup()

		key, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: key}
		userAddress := signer.GetAddress().Hex()

		// Create sessions with specific creation times to test sorting
		baseTime := time.Now().Add(-24 * time.Hour)
		sessions := []AppSession{
			{
				SessionID:          "0xSession1",
				ParticipantWallets: []string{userAddress, "0xParticipant2"},
				SessionData:        `{"key":"value"}`,
				Status:             ChannelStatusOpen,
				Protocol:           "test-app-1",
				Challenge:          60,
				Weights:            []int64{50, 50},
				Quorum:             75,
				Nonce:              1,
				Version:            1,
				CreatedAt:          baseTime,
				UpdatedAt:          baseTime,
			},
			{
				SessionID:          "0xSession2",
				ParticipantWallets: []string{userAddress, "0xParticipant3"},
				SessionData:        `{"key":"value"}`,
				Status:             ChannelStatusClosed,
				Protocol:           "test-app-2",
				Challenge:          120,
				Weights:            []int64{30, 70},
				Quorum:             80,
				Nonce:              2,
				Version:            2,
				CreatedAt:          baseTime.Add(1 * time.Hour),
				UpdatedAt:          baseTime.Add(1 * time.Hour),
			},
			{
				SessionID:          "0xSession3",
				ParticipantWallets: []string{"0xParticipant4", "0xParticipant5"},
				SessionData:        `{"key":"value"}`,
				Status:             ChannelStatusOpen,
				Protocol:           "test-app-3",
				Challenge:          90,
				Weights:            []int64{40, 60},
				Quorum:             60,
				Nonce:              3,
				Version:            3,
				CreatedAt:          baseTime.Add(2 * time.Hour),
				UpdatedAt:          baseTime.Add(2 * time.Hour),
			},
		}

		for _, session := range sessions {
			require.NoError(t, db.Create(&session).Error)
		}

		tcs := []struct {
			name               string
			params             map[string]interface{}
			expectedSessionIDs []string
		}{
			{
				name:               "Get all with no sort (default desc by created_at)",
				params:             map[string]interface{}{},
				expectedSessionIDs: []string{"0xSession3", "0xSession2", "0xSession1"},
			},
			{
				name:               "Get all with ascending sort",
				params:             map[string]interface{}{"sort": "asc"},
				expectedSessionIDs: []string{"0xSession1", "0xSession2", "0xSession3"},
			},
			{
				name:               "Get all with descending sort",
				params:             map[string]interface{}{"sort": "desc"},
				expectedSessionIDs: []string{"0xSession3", "0xSession2", "0xSession1"},
			},
			{
				name:               "Filter by participant",
				params:             map[string]interface{}{"participant": userAddress},
				expectedSessionIDs: []string{"0xSession2", "0xSession1"},
			},
			{
				name:               "Filter by participant with ascending sort",
				params:             map[string]interface{}{"participant": userAddress, "sort": "asc"},
				expectedSessionIDs: []string{"0xSession1", "0xSession2"},
			},
			{
				name:               "Filter by status open",
				params:             map[string]interface{}{"status": string(ChannelStatusOpen)},
				expectedSessionIDs: []string{"0xSession3", "0xSession1"},
			},
			{
				name:               "Filter by participant and status open",
				params:             map[string]interface{}{"participant": userAddress, "status": string(ChannelStatusOpen)},
				expectedSessionIDs: []string{"0xSession1"},
			},
			{
				name:               "Filter by status closed",
				params:             map[string]interface{}{"status": string(ChannelStatusClosed)},
				expectedSessionIDs: []string{"0xSession2"},
			},
		}

		for idx, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				ctx := createRPCContext(idx, "get_app_sessions", tc.params)
				router.HandleGetAppSessions(ctx)

				res := assertResponse(t, ctx, "get_app_sessions")
				assert.Equal(t, uint64(idx), res.RequestID)

				sessionResponses, ok := res.Params.(GetAppSessionsResponse)
				require.True(t, ok, "Response parameter should be a GetAppSessionsResponse")
				assert.Len(t, sessionResponses.AppSessions, len(tc.expectedSessionIDs), "Should return expected number of app sessions")

				for idx, sessionResponse := range sessionResponses.AppSessions {
					assert.True(t, sessionResponse.AppSessionID == tc.expectedSessionIDs[idx], "Should include session %s", tc.expectedSessionIDs[idx])
				}
			})
		}
	})
	t.Run("RPCRouterHandleGetAppSessions_Pagination", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		defer cleanup()

		baseTime := time.Now()

		sessionIDs := []string{
			"0xSession11", "0xSession10", "0xSession09",
			"0xSession08", "0xSession07", "0xSession06",
			"0xSession05", "0xSession04", "0xSession03",
			"0xSession02", "0xSession01",
		}

		testSessions := []AppSession{
			{Nonce: 11, ParticipantWallets: []string{"0xParticipant11"}, Status: ChannelStatusOpen},
			{Nonce: 10, ParticipantWallets: []string{"0xParticipant10"}, Status: ChannelStatusOpen},
			{Nonce: 9, ParticipantWallets: []string{"0xParticipant9"}, Status: ChannelStatusOpen},
			{Nonce: 8, ParticipantWallets: []string{"0xParticipant8"}, Status: ChannelStatusOpen},
			{Nonce: 7, ParticipantWallets: []string{"0xParticipant7"}, Status: ChannelStatusOpen},
			{Nonce: 6, ParticipantWallets: []string{"0xParticipant6"}, Status: ChannelStatusOpen},
			{Nonce: 5, ParticipantWallets: []string{"0xParticipant5"}, Status: ChannelStatusOpen},
			{Nonce: 4, ParticipantWallets: []string{"0xParticipant4"}, Status: ChannelStatusOpen},
			{Nonce: 3, ParticipantWallets: []string{"0xParticipant3"}, Status: ChannelStatusOpen},
			{Nonce: 2, ParticipantWallets: []string{"0xParticipant2"}, Status: ChannelStatusOpen},
			{Nonce: 1, ParticipantWallets: []string{"0xParticipant1"}, Status: ChannelStatusOpen},
		}

		for i := range testSessions {
			testSessions[i].SessionID = sessionIDs[i]
			testSessions[i].UpdatedAt = baseTime.Add(-time.Duration(i) * time.Hour)
			testSessions[i].CreatedAt = testSessions[i].UpdatedAt
		}

		for _, session := range testSessions {
			require.NoError(t, db.Create(&session).Error)
		}

		tcs := []struct {
			name               string
			params             map[string]interface{}
			expectedSessionIDs []string
		}{
			{name: "No params",
				params:             map[string]interface{}{},
				expectedSessionIDs: sessionIDs[:10], // Default pagination should return first 10 sessions (desc order)
			},
			{name: "Offset only",
				params:             map[string]interface{}{"offset": float64(2)},
				expectedSessionIDs: sessionIDs[2:11], // Default limit is 10, total 11, so offset 2 returns 9 sessions
			},
			{name: "Limit only",
				params:             map[string]interface{}{"limit": float64(5)},
				expectedSessionIDs: sessionIDs[:5], // Default offset is 0, so limit 5 returns first 5 sessions
			},
			{name: "Offset and limit",
				params:             map[string]interface{}{"offset": float64(2), "limit": float64(3)},
				expectedSessionIDs: sessionIDs[2:5], // Offset 2 with limit 3 returns 3 sessions
			},
			{name: "Pagination with sort",
				params:             map[string]interface{}{"offset": float64(2), "limit": float64(3), "sort": "asc"},
				expectedSessionIDs: []string{"0xSession03", "0xSession04", "0xSession05"}, // Offset 2 with limit 3 returns Sessions 3 to 5 (asc order)
			},
			{name: "Pagination with participant",
				params:             map[string]interface{}{"participant": "0xNonExistentParticipant", "offset": float64(1), "limit": float64(2)},
				expectedSessionIDs: []string{}, // No sessions for non-existent participant
			},
		}

		for idx, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				ctx := createRPCContext(idx, "get_app_sessions", tc.params)
				router.HandleGetAppSessions(ctx)

				res := assertResponse(t, ctx, "get_app_sessions")
				responseSessions, ok := res.Params.(GetAppSessionsResponse)
				require.True(t, ok, "Response parameter should be a GetAppSessionsResponse")
				assert.Len(t, responseSessions.AppSessions, len(tc.expectedSessionIDs), "Should return expected number of sessions")

				// Check session IDs are in expected order
				for idx, session := range responseSessions.AppSessions {
					assert.True(t, session.AppSessionID == tc.expectedSessionIDs[idx], "Retrieved %d-th session ID should be equal %s", idx, tc.expectedSessionIDs[idx])
				}
			})
		}
	})
}

func TestRPCRouterHandleGetLedgerEntries(t *testing.T) {
	t.Run("RPCRouterHandleGetLedgerEntries", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		defer cleanup()

		participant1 := newTestCommonAddress("0xParticipant1")
		participant1AccountID := NewAccountID(participant1.Hex())
		participant2 := newTestCommonAddress("0xParticipant2")
		participant2AccountID := NewAccountID(participant2.Hex())

		// Setup test data
		ledger1 := GetWalletLedger(db, participant1)
		testData1 := []struct {
			asset  string
			amount decimal.Decimal
		}{
			{"usdc", decimal.NewFromInt(100)},
			{"usdc", decimal.NewFromInt(200)},
			{"usdc", decimal.NewFromInt(-50)},
			{"eth", decimal.NewFromFloat(1.5)},
			{"eth", decimal.NewFromFloat(-0.5)},
		}
		for _, data := range testData1 {
			err := ledger1.Record(participant1AccountID, data.asset, data.amount, nil)
			require.NoError(t, err)
		}

		ledger2 := GetWalletLedger(db, participant2)
		testData2 := []struct {
			asset  string
			amount decimal.Decimal
		}{
			{"usdc", decimal.NewFromInt(300)},
			{"btc", decimal.NewFromFloat(0.05)},
		}
		for _, data := range testData2 {
			err := ledger2.Record(participant2AccountID, data.asset, data.amount, nil)
			require.NoError(t, err)
		}

		tcs := []struct {
			name          string
			userID        string
			params        map[string]interface{}
			expectedCount int
			validateFunc  func(t *testing.T, entries []LedgerEntryResponse)
		}{
			{
				name:          "Filter by account_id only",
				params:        map[string]interface{}{"account_id": participant1},
				expectedCount: 5,
				validateFunc: func(t *testing.T, entries []LedgerEntryResponse) {
					assetCounts := map[string]int{}
					for _, entry := range entries {
						assetCounts[entry.Asset]++
						assert.Equal(t, participant1.Hex(), entry.AccountID, "Should return correct account_id")
						assert.Equal(t, participant1.Hex(), entry.Participant, "Should return entries for participant1")
					}
					assert.Equal(t, 3, assetCounts["usdc"], "Should have 3 USDC entries")
					assert.Equal(t, 2, assetCounts["eth"], "Should have 2 ETH entries")
				},
			},
			{
				name:          "Filter by account_id and asset",
				params:        map[string]interface{}{"account_id": participant1, "asset": "usdc"},
				expectedCount: 3,
				validateFunc: func(t *testing.T, entries []LedgerEntryResponse) {
					for _, entry := range entries {
						assert.Equal(t, "usdc", entry.Asset)
						assert.Equal(t, participant1.Hex(), entry.AccountID, "Should return correct account_id")
						assert.Equal(t, participant1.Hex(), entry.Participant, "Should return entries for participant1")
					}
				},
			},
			{
				name:          "Filter by wallet only",
				params:        map[string]interface{}{"wallet": participant2},
				expectedCount: 2,
				validateFunc: func(t *testing.T, entries []LedgerEntryResponse) {
					for _, entry := range entries {
						assert.Equal(t, participant2.Hex(), entry.Participant, "Should return entries for participant2")
					}
				},
			},
			{
				name:          "Filter by wallet and asset",
				params:        map[string]interface{}{"wallet": participant2, "asset": "usdc"},
				expectedCount: 1,
				validateFunc: func(t *testing.T, entries []LedgerEntryResponse) {
					assert.Equal(t, "usdc", entries[0].Asset)
					assert.Equal(t, participant2.Hex(), entries[0].Participant)
				},
			},
			{
				name:          "Filter by account_id and wallet (no overlap)",
				params:        map[string]interface{}{"account_id": participant1, "wallet": participant2},
				expectedCount: 0,
				validateFunc:  func(t *testing.T, entries []LedgerEntryResponse) {},
			},
			{
				name:          "No filters (all entries)",
				params:        map[string]interface{}{},
				expectedCount: 7,
				validateFunc: func(t *testing.T, entries []LedgerEntryResponse) {
					foundParticipants := make(map[string]bool)
					for _, entry := range entries {
						foundParticipants[entry.Participant] = true
					}
					assert.True(t, foundParticipants[participant1.Hex()], "Should include entries for participant1")
					assert.True(t, foundParticipants[participant2.Hex()], "Should include entries for participant2")
				},
			},
			{
				name:          "Default wallet provided",
				userID:        participant1.Hex(),
				params:        map[string]interface{}{},
				expectedCount: 5,
				validateFunc: func(t *testing.T, entries []LedgerEntryResponse) {
					for _, entry := range entries {
						assert.Equal(t, participant1.Hex(), entry.Participant, "Should return entries for default wallet participant1")
					}
				},
			},
		}

		for idx, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				ctx := createRPCContext(idx+1, "get_ledger_entries", tc.params)
				ctx.UserID = tc.userID

				router.HandleGetLedgerEntries(ctx)

				res := assertResponse(t, ctx, "get_ledger_entries")
				assert.Equal(t, uint64(idx+1), res.RequestID)

				entries, ok := res.Params.(GetLedgerEntriesResponse)
				require.True(t, ok, "Response parameter should be a GetLedgerEntriesResponse")
				assert.Len(t, entries.LedgerEntries, tc.expectedCount, "Should return expected number of entries")

				tc.validateFunc(t, entries.LedgerEntries)
			})
		}
	})
	t.Run("TestRPCRouterHandleGetLedgerEntries_Pagination", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		defer cleanup()

		userAddress := newTestCommonAddress("0xParticipant1")
		userAccountID := NewAccountID(userAddress.Hex())

		tokenNames := []string{
			"eth1", "eth2", "eth3", "eth4", "eth5", "eth6", "eth7", "eth8", "eth9", "eth10", "eth11"}

		// Create 11 ledger entries for pagination testing
		ledger := GetWalletLedger(db, userAddress)
		testData := []struct {
			asset  string
			amount decimal.Decimal
		}{
			{"eth11", decimal.NewFromInt(100)},
			{"eth10", decimal.NewFromFloat(1.0)},
			{"eth9", decimal.NewFromInt(200)},
			{"eth8", decimal.NewFromFloat(0.1)},
			{"eth7", decimal.NewFromInt(300)},
			{"eth6", decimal.NewFromFloat(2.0)},
			{"eth5", decimal.NewFromInt(400)},
			{"eth4", decimal.NewFromFloat(0.2)},
			{"eth3", decimal.NewFromInt(500)},
			{"eth2", decimal.NewFromFloat(3.0)},
			{"eth1", decimal.NewFromInt(600)},
		}

		// Create all entries
		for _, data := range testData {
			err := ledger.Record(userAccountID, data.asset, data.amount, nil)
			require.NoError(t, err)
		}

		tcs := []struct {
			name          string
			params        map[string]interface{}
			expectedToken []string
		}{
			{name: "No params",
				params:        map[string]interface{}{},
				expectedToken: tokenNames[:10], // Default pagination should return first 10 tokens
			},
			{name: "Offset only",
				params:        map[string]interface{}{"offset": float64(2)},
				expectedToken: tokenNames[2:11], // Skip first 2, return rest
			},
			{name: "Limit only",
				params:        map[string]interface{}{"limit": float64(5)},
				expectedToken: tokenNames[:5], // Return first 5 tokens
			},
			{name: "Offset and limit",
				params:        map[string]interface{}{"offset": float64(2), "limit": float64(3)},
				expectedToken: tokenNames[2:5], // Skip 2, take 3
			},
			{name: "Pagination with sort",
				params:        map[string]interface{}{"offset": float64(2), "limit": float64(3), "sort": "asc"},
				expectedToken: []string{"eth9", "eth8", "eth7"}, // Ascending order by creation time, skip 2, take 3
			},
			{name: "Pagination with asset filter",
				params:        map[string]interface{}{"asset": "eth1", "limit": float64(1)},
				expectedToken: []string{"eth1"}, // Only eth1 asset
			},
		}

		for idx, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				c := createRPCContext(idx, "get_ledger_entries", tc.params)
				router.HandleGetLedgerEntries(c)

				res := assertResponse(t, c, "get_ledger_entries")
				responseEntries, ok := res.Params.(GetLedgerEntriesResponse)
				require.True(t, ok, "Response parameter should be a GetLedgerEntriesResponse")
				assert.Len(t, responseEntries.LedgerEntries, len(tc.expectedToken), "Should return expected number of entries")

				// Check token names are included in expected order
				for idx, entry := range responseEntries.LedgerEntries {
					assert.Equal(t, tc.expectedToken[idx], entry.Asset, "Should include token %s at position %d", tc.expectedToken[idx], idx)
				}
			})
		}
	})
}

func TestRPCRouterHandleGetTransactions(t *testing.T) {
	t.Run("RPCRouterHandleGetTransactions", func(t *testing.T) {
		t.Parallel()

		// --- 1. SETUP (Runs once for all test cases) ---
		router, db, cleanup := setupTestRPCRouter(t)
		defer cleanup()

		account1 := "0xAccount1"
		account2 := "0xAccount2"
		account3 := "0xAccount3"

		// Create and seed test transactions
		testTransactions := []LedgerTransaction{
			{Type: TransactionTypeTransfer, FromAccount: account1, ToAccount: account2, AssetSymbol: "usdc", Amount: decimal.NewFromInt(100), CreatedAt: time.Now().Add(-3 * time.Hour)},
			{Type: TransactionTypeDeposit, FromAccount: account2, ToAccount: account1, AssetSymbol: "usdc", Amount: decimal.NewFromInt(50), CreatedAt: time.Now().Add(-2 * time.Hour)},
			{Type: TransactionTypeTransfer, FromAccount: account1, ToAccount: account3, AssetSymbol: "eth", Amount: decimal.NewFromFloat(1.5), CreatedAt: time.Now().Add(-1 * time.Hour)},
			{Type: TransactionTypeWithdrawal, FromAccount: account3, ToAccount: account2, AssetSymbol: "usdc", Amount: decimal.NewFromInt(25), CreatedAt: time.Now()},
		}
		for _, tx := range testTransactions {
			// We use a temporary variable to avoid taking the address of a loop variable.
			tempTx := tx
			require.NoError(t, db.Create(&tempTx).Error)
		}

		// --- 2. DEFINE TEST CASES ---
		testCases := []struct {
			name        string
			params      map[string]any
			expectedLen int
			assertions  func(t *testing.T, transactions []TransactionResponse) // Optional custom assertions
		}{
			{
				name:        "Get all transactions for a specific account",
				params:      map[string]any{"account_id": account1},
				expectedLen: 3,
				assertions: func(t *testing.T, txs []TransactionResponse) {
					// Verify ordering (newest first)
					require.True(t, txs[0].CreatedAt.After(txs[1].CreatedAt))
					require.True(t, txs[1].CreatedAt.After(txs[2].CreatedAt))
					// Verify account1 is always involved
					for _, tx := range txs {
						assert.True(t, tx.FromAccount == account1 || tx.ToAccount == account1)
					}
				},
			},
			{
				name:        "Filter by account and asset",
				params:      map[string]any{"account_id": account1, "asset": "usdc"},
				expectedLen: 2,
				assertions: func(t *testing.T, txs []TransactionResponse) {
					for _, tx := range txs {
						assert.Equal(t, "usdc", tx.Asset)
						assert.True(t, tx.FromAccount == account1 || tx.ToAccount == account1)
					}
				},
			},
			{
				name:        "Filter by asset only",
				params:      map[string]any{"asset": "eth"},
				expectedLen: 1,
				assertions: func(t *testing.T, txs []TransactionResponse) {
					assert.Equal(t, "eth", txs[0].Asset)
					assert.Equal(t, account1, txs[0].FromAccount)
					assert.Equal(t, account3, txs[0].ToAccount)
				},
			},
			{
				name:        "No filters should return all transactions",
				params:      map[string]any{},
				expectedLen: 4,
				assertions: func(t *testing.T, txs []TransactionResponse) {
					foundAccounts := make(map[string]bool)
					for _, tx := range txs {
						foundAccounts[tx.FromAccount] = true
						foundAccounts[tx.ToAccount] = true
					}
					assert.True(t, foundAccounts[account1])
					assert.True(t, foundAccounts[account2])
					assert.True(t, foundAccounts[account3])
				},
			},
			{
				name:        "Account with no transactions",
				params:      map[string]any{"account_id": "0xNonExistentAccount"},
				expectedLen: 0,
				assertions:  nil, // No extra assertions needed beyond length check
			},
			{
				name:        "Filter by non-existent asset",
				params:      map[string]any{"asset": "nonexistent"},
				expectedLen: 0,
				assertions:  nil,
			},
			{
				name:        "Filter by transaction type - transfer",
				params:      map[string]any{"tx_type": "transfer"},
				expectedLen: 2,
				assertions: func(t *testing.T, txs []TransactionResponse) {
					for _, tx := range txs {
						assert.Equal(t, "transfer", tx.TxType)
					}
				},
			},
			{
				name:        "Filter by transaction type - deposit",
				params:      map[string]any{"tx_type": "deposit"},
				expectedLen: 1,
				assertions: func(t *testing.T, txs []TransactionResponse) {
					assert.Equal(t, "deposit", txs[0].TxType)
					assert.Equal(t, account2, txs[0].FromAccount)
					assert.Equal(t, account1, txs[0].ToAccount)
				},
			},
			{
				name:        "Filter by transaction type - withdrawal",
				params:      map[string]any{"tx_type": "withdrawal"},
				expectedLen: 1,
				assertions: func(t *testing.T, txs []TransactionResponse) {
					assert.Equal(t, "withdrawal", txs[0].TxType)
					assert.Equal(t, account3, txs[0].FromAccount)
					assert.Equal(t, account2, txs[0].ToAccount)
				},
			},
			{
				name:        "Filter by account and transaction type",
				params:      map[string]any{"account_id": account1, "tx_type": "transfer"},
				expectedLen: 2,
				assertions: func(t *testing.T, txs []TransactionResponse) {
					for _, tx := range txs {
						assert.Equal(t, "transfer", tx.TxType)
						assert.True(t, tx.FromAccount == account1 || tx.ToAccount == account1)
					}
				},
			},
			{
				name:        "Filter by asset and transaction type",
				params:      map[string]any{"asset": "usdc", "tx_type": "deposit"},
				expectedLen: 1,
				assertions: func(t *testing.T, txs []TransactionResponse) {
					assert.Equal(t, "usdc", txs[0].Asset)
					assert.Equal(t, "deposit", txs[0].TxType)
				},
			},
		}

		// --- 3. RUN TEST CASES ---
		for i, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				c := createRPCContext(i+1, "get_ledger_transactions", tc.params)
				router.HandleGetLedgerTransactions(c)

				res := assertResponse(t, c, "get_ledger_transactions")

				// Unmarshal the actual transaction data
				var resp GetLedgerTransactionsResponse
				require.NotNil(t, res.Params, "Response parameter should not be nil")
				// We need to marshal the interface{} back to JSON, then unmarshal into our concrete type.
				respBytes, err := json.Marshal(res.Params)
				require.NoError(t, err)
				err = json.Unmarshal(respBytes, &resp)
				require.NoError(t, err, "Response parameter should be a GetLedgerTransactionsResponse")

				transactions := resp.LedgerTransactions
				// Assert the expected number of transactions were returned
				assert.Len(t, transactions, tc.expectedLen)

				// Run specific assertions for this test case, if any
				if tc.assertions != nil {
					tc.assertions(t, transactions)
				}
			})
		}

		// --- 4. ERROR TEST CASES ---
		t.Run("Filter by non-existent transaction type should return error", func(t *testing.T) {
			c := createRPCContext(999, "get_ledger_transactions", map[string]any{"tx_type": "nonexistent"})
			router.HandleGetLedgerTransactions(c)

			assertErrorResponse(t, c, ErrInvalidLedgerTransactionType.Error())
		})
	})
	t.Run("RPCRouterHandleGetLedgerTransactions_Pagination", func(t *testing.T) {
		t.Parallel()

		router, db, cleanup := setupTestRPCRouter(t)
		defer cleanup()

		account1 := "0xAccount1"
		account2 := "0xAccount2"
		account3 := "0xAccount3"

		// Create 11 test transactions for pagination testing
		testTransactions := []LedgerTransaction{
			{Type: TransactionTypeTransfer, FromAccount: account1, ToAccount: account2, AssetSymbol: "usdc", Amount: decimal.NewFromInt(100), CreatedAt: time.Now().Add(-10 * time.Hour)},
			{Type: TransactionTypeDeposit, FromAccount: account2, ToAccount: account1, AssetSymbol: "usdc", Amount: decimal.NewFromInt(50), CreatedAt: time.Now().Add(-9 * time.Hour)},
			{Type: TransactionTypeTransfer, FromAccount: account1, ToAccount: account3, AssetSymbol: "eth", Amount: decimal.NewFromFloat(1.5), CreatedAt: time.Now().Add(-8 * time.Hour)},
			{Type: TransactionTypeWithdrawal, FromAccount: account3, ToAccount: account2, AssetSymbol: "usdc", Amount: decimal.NewFromInt(25), CreatedAt: time.Now().Add(-7 * time.Hour)},
			{Type: TransactionTypeTransfer, FromAccount: account2, ToAccount: account1, AssetSymbol: "usdc", Amount: decimal.NewFromInt(75), CreatedAt: time.Now().Add(-6 * time.Hour)},
			{Type: TransactionTypeDeposit, FromAccount: account1, ToAccount: account3, AssetSymbol: "eth", Amount: decimal.NewFromFloat(0.5), CreatedAt: time.Now().Add(-5 * time.Hour)},
			{Type: TransactionTypeTransfer, FromAccount: account3, ToAccount: account2, AssetSymbol: "usdc", Amount: decimal.NewFromInt(30), CreatedAt: time.Now().Add(-4 * time.Hour)},
			{Type: TransactionTypeWithdrawal, FromAccount: account2, ToAccount: account1, AssetSymbol: "eth", Amount: decimal.NewFromFloat(0.2), CreatedAt: time.Now().Add(-3 * time.Hour)},
			{Type: TransactionTypeTransfer, FromAccount: account1, ToAccount: account2, AssetSymbol: "usdc", Amount: decimal.NewFromInt(60), CreatedAt: time.Now().Add(-2 * time.Hour)},
			{Type: TransactionTypeDeposit, FromAccount: account2, ToAccount: account3, AssetSymbol: "usdc", Amount: decimal.NewFromInt(40), CreatedAt: time.Now().Add(-1 * time.Hour)},
			{Type: TransactionTypeTransfer, FromAccount: account3, ToAccount: account1, AssetSymbol: "eth", Amount: decimal.NewFromFloat(0.1), CreatedAt: time.Now()},
		}

		require.NoError(t, db.Create(testTransactions).Error)

		// Expected order: most recent first (descending by created_at)
		expectedHashes := make([]string, 11)
		for i := 0; i < 11; i++ {
			var tx LedgerTransaction
			require.NoError(t, db.Where("created_at = ?", testTransactions[10-i].CreatedAt).First(&tx).Error)
		}

		tcs := []struct {
			name          string
			params        map[string]interface{}
			expectedCount int
			expectedFirst string
			expectedLast  string
		}{
			{
				name:          "No params (default pagination)",
				params:        map[string]interface{}{},
				expectedCount: 10, // Default limit should be 10
				expectedFirst: expectedHashes[0],
				expectedLast:  expectedHashes[9],
			},
			{
				name:          "Offset only",
				params:        map[string]interface{}{"offset": float64(2)},
				expectedCount: 9, // Skip first 2, get remaining 9
				expectedFirst: expectedHashes[2],
				expectedLast:  expectedHashes[10],
			},
			{
				name:          "Limit only",
				params:        map[string]interface{}{"limit": float64(5)},
				expectedCount: 5, // Get first 5
				expectedFirst: expectedHashes[0],
				expectedLast:  expectedHashes[4],
			},
			{
				name:          "Offset and limit",
				params:        map[string]interface{}{"offset": float64(3), "limit": float64(4)},
				expectedCount: 4, // Skip 3, take 4
				expectedFirst: expectedHashes[3],
				expectedLast:  expectedHashes[6],
			},
			{
				name:          "Pagination with sort asc",
				params:        map[string]interface{}{"offset": float64(1), "limit": float64(3), "sort": "asc"},
				expectedCount: 3,                 // Ascending order, skip 1, take 3
				expectedFirst: expectedHashes[9], // 2nd oldest
				expectedLast:  expectedHashes[7], // 4th oldest
			},
			{
				name:          "Pagination with asset filter",
				params:        map[string]interface{}{"asset": "usdc", "limit": float64(3)},
				expectedCount: 3, // Only USDC transactions, first 3
			},
			{
				name:          "Pagination with account filter",
				params:        map[string]interface{}{"account_id": account1, "limit": float64(4)},
				expectedCount: 4, // Only transactions involving account1, first 4
			},
		}

		for idx, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				ctx := createRPCContext(idx+100, "get_ledger_transactions", tc.params)
				router.HandleGetLedgerTransactions(ctx)

				res := assertResponse(t, ctx, "get_ledger_transactions")

				var resp GetLedgerTransactionsResponse
				respBytes, err := json.Marshal(res.Params)
				require.NoError(t, err)
				err = json.Unmarshal(respBytes, &resp)
				require.NoError(t, err)

				transactions := resp.LedgerTransactions
				assert.Len(t, transactions, tc.expectedCount, "Should return expected number of transactions")

				// For non-filter tests, verify order
				if tc.expectedFirst != "" && tc.expectedLast != "" && len(transactions) > 0 {
					assert.Equal(t, tc.expectedFirst, transactions[0].Id, "First transaction hash should match")
					if len(transactions) > 1 {
						assert.Equal(t, tc.expectedLast, transactions[len(transactions)-1].Id, "Last transaction hash should match")
					}
				}

				// Verify transactions are properly sorted by created_at
				if len(transactions) > 1 {
					sortOrder := tc.params["sort"]
					isAsc := sortOrder == "asc"
					for i := 0; i < len(transactions)-1; i++ {
						curr := transactions[i].CreatedAt
						next := transactions[i+1].CreatedAt
						if isAsc {
							assert.True(t, curr.Before(next) || curr.Equal(next), "Transactions should be sorted ascending by created_at")
						} else {
							assert.True(t, curr.After(next) || curr.Equal(next), "Transactions should be sorted descending by created_at")
						}
					}
				}
			})
		}
	})
}
