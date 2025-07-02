package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCRouterHandlePing(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	c := &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 1,
				Method:    "ping",
				Params:    []any{nil},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	router.HandlePing(c)
	res := c.Message.Res
	require.NotNil(t, res)

	assert.Equal(t, "pong", res.Method)
}

func TestRPCRouterHandleGetConfig(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	mockConfig := &Config{
		networks: map[string]*NetworkConfig{
			"polygon": {
				Name:           "polygon",
				ChainID:        137,
				InfuraURL:      "https://polygon-mainnet.infura.io/v3/test",
				CustodyAddress: "0xCustodyAddress1",
			},
			"celo": {
				Name:           "celo",
				ChainID:        42220,
				InfuraURL:      "https://celo-mainnet.infura.io/v3/test",
				CustodyAddress: "0xCustodyAddress2",
			},
			"base": {
				Name:           "base",
				ChainID:        8453,
				InfuraURL:      "https://base-mainnet.infura.io/v3/test",
				CustodyAddress: "0xCustodyAddress3",
			},
		},
	}
	router.Config = mockConfig

	c := &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 1,
				Method:    "get_config",
				Params:    []any{},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	router.HandleGetConfig(c)
	res := c.Message.Res
	require.NotNil(t, res)

	require.NotEmpty(t, res.Params)
	configMap, ok := res.Params[0].(BrokerConfig)
	require.True(t, ok, "Response should contain a BrokerConfig")
	assert.Equal(t, router.Signer.GetAddress().Hex(), configMap.BrokerAddress)
	require.Len(t, configMap.Networks, 3, "Should have 3 supported networks")

	expectedNetworks := map[string]uint32{
		"polygon": 137,
		"celo":    42220,
		"base":    8453,
	}
	for _, network := range configMap.Networks {
		expectedChainID, exists := expectedNetworks[network.Name]
		assert.True(t, exists, "Network %s should be in expected networks", network.Name)
		assert.Equal(t, expectedChainID, network.ChainID, "Chain ID should match for %s", network.Name)
		assert.Contains(t, network.CustodyAddress, "0xCustodyAddress", "Custody address should be present")
		delete(expectedNetworks, network.Name)
	}
	assert.Empty(t, expectedNetworks, "All expected networks should be found")
}

func TestRPCRouterHandleGetAssets(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	testAssets := []Asset{
		{Token: "0xToken1", ChainID: 137, Symbol: "usdc", Decimals: 6},
		{Token: "0xToken2", ChainID: 137, Symbol: "weth", Decimals: 18},
		{Token: "0xToken3", ChainID: 42220, Symbol: "celo", Decimals: 18},
		{Token: "0xToken4", ChainID: 8453, Symbol: "usdbc", Decimals: 6},
	}

	for _, asset := range testAssets {
		require.NoError(t, router.DB.Create(&asset).Error)
	}

	tcs := []struct {
		name               string
		params             map[string]interface{}
		expectedTokenNames []string
	}{
		{
			name:               "Get all with no sort (default asc, by chain_id and symbol)",
			params:             map[string]interface{}{},
			expectedTokenNames: []string{"0xToken3", "0xToken4", "0xToken1", "0xToken2"},
		},
		{
			name:               "Filter by chain_id=137",
			params:             map[string]interface{}{"chain_id": float64(137)},
			expectedTokenNames: []string{"0xToken1", "0xToken2"},
		},
		{
			name:               "Filter by chain_id=42220",
			params:             map[string]interface{}{"chain_id": float64(42220)},
			expectedTokenNames: []string{"0xToken3"},
		},
		{
			name:               "Filter by non-existent chain_id=1",
			params:             map[string]interface{}{"chain_id": float64(1)},
			expectedTokenNames: []string{},
		},
	}

	for idx, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			paramsJSON, err := json.Marshal(tc.params)
			require.NoError(t, err, "Failed to marshal params")

			c := &RPCContext{
				Context: context.TODO(),
				Message: RPCMessage{
					Req: &RPCData{
						RequestID: uint64(idx),
						Method:    "get_assets",
						Params:    []any{json.RawMessage(paramsJSON)},
						Timestamp: uint64(time.Now().Unix()),
					},
					Sig: []string{"dummy-signature"},
				},
			}

			router.HandleGetAssets(c)
			res := c.Message.Res
			require.NotNil(t, res)

			assert.Equal(t, "get_assets", res.Method)
			assert.Equal(t, uint64(idx), res.RequestID)
			require.Len(t, res.Params, 1, "Response should contain an array of AssetResponse")

			responseAssets, ok := res.Params[0].([]GetAssetsResponse)
			require.True(t, ok, "Response parameter should be a slice of AssetResponse")
			assert.Len(t, responseAssets, len(tc.expectedTokenNames), "Should return expected number of assets")

			for idx, asset := range responseAssets {
				assert.True(t, asset.Token == tc.expectedTokenNames[idx], "Should include token %s", tc.expectedTokenNames[idx])
			}
		})
	}
}

func TestRPCRouterHandleGetChannels(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
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
			Status:      ChannelStatusJoining,
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

	for _, channel := range channels {
		require.NoError(t, router.DB.Create(&channel).Error)
	}

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
			expectedChannelIDs: []string{"0xOtherChannel", "0xChannel1"},
		},
		{
			name:               "Filter by participant and status open",
			params:             map[string]interface{}{"participant": participantWallet, "status": string(ChannelStatusOpen)},
			expectedChannelIDs: []string{"0xChannel1"},
		},
		{
			name:               "Filter by participant and status closed",
			params:             map[string]interface{}{"participant": participantWallet, "status": string(ChannelStatusClosed)},
			expectedChannelIDs: []string{"0xChannel2"},
		},
		{
			name:               "Filter by participant and status joining",
			params:             map[string]interface{}{"participant": participantWallet, "status": string(ChannelStatusJoining)},
			expectedChannelIDs: []string{"0xChannel3"},
		},
		{
			name:               "Filter by status closed only",
			params:             map[string]interface{}{"status": string(ChannelStatusClosed)},
			expectedChannelIDs: []string{"0xChannel2"},
		},
	}

	for idx, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			paramsJSON, err := json.Marshal(tc.params)
			require.NoError(t, err, "Failed to marshal params")

			c := &RPCContext{
				Context: context.TODO(),
				Message: RPCMessage{
					Req: &RPCData{
						RequestID: uint64(idx),
						Method:    "get_channels",
						Params:    []any{json.RawMessage(paramsJSON)},
						Timestamp: uint64(time.Now().Unix()),
					},
					Sig: []string{"dummy-signature"},
				},
			}

			router.HandleGetChannels(c)
			res := c.Message.Res
			require.NotNil(t, res)

			assert.Equal(t, "get_channels", res.Method)
			assert.Equal(t, uint64(idx), res.RequestID)
			require.Len(t, res.Params, 1, "Response should contain a slice of ChannelResponse")

			responseChannels, ok := res.Params[0].([]ChannelResponse)
			require.True(t, ok, "Response parameter should be a slice of ChannelResponse")
			assert.Len(t, responseChannels, len(tc.expectedChannelIDs), "Should return expected number of channels")

			for idx, channel := range responseChannels {
				assert.True(t, channel.ChannelID == tc.expectedChannelIDs[idx], "%d-th result (%s) should equal %s", idx, channel.ChannelID, tc.expectedChannelIDs[idx])
			}
		})
	}
}

func TestRPCRouterHandleGetChannels_Pagination(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	channelIDs := []string{
		"0xChannel01", "0xChannel02", "0xChannel03", "0xChannel04",
		"0xChannel05", "0xChannel06", "0xChannel07", "0xChannel08",
		"0xChannel09", "0xChannel10", "0xChannel11"}

	testChannels := []Channel{
		{Wallet: "0xWallet1", Participant: "0xParticipant1", Status: ChannelStatusOpen, Nonce: 1},
		{Wallet: "0xWallet2", Participant: "0xParticipant2", Status: ChannelStatusClosed, Nonce: 2},
		{Wallet: "0xWallet3", Participant: "0xParticipant3", Status: ChannelStatusOpen, Nonce: 3},
		{Wallet: "0xWallet4", Participant: "0xParticipant4", Status: ChannelStatusJoining, Nonce: 4},
		{Wallet: "0xWallet5", Participant: "0xParticipant5", Status: ChannelStatusOpen, Nonce: 5},
		{Wallet: "0xWallet6", Participant: "0xParticipant6", Status: ChannelStatusChallenged, Nonce: 6},
		{Wallet: "0xWallet7", Participant: "0xParticipant7", Status: ChannelStatusOpen, Nonce: 7},
		{Wallet: "0xWallet8", Participant: "0xParticipant8", Status: ChannelStatusClosed, Nonce: 8},
		{Wallet: "0xWallet9", Participant: "0xParticipant9", Status: ChannelStatusOpen, Nonce: 9},
		{Wallet: "0xWallet10", Participant: "0xParticipant10", Status: ChannelStatusJoining, Nonce: 10},
		{Wallet: "0xWallet11", Participant: "0xParticipant11", Status: ChannelStatusOpen, Nonce: 11},
	}

	for i := range testChannels {
		testChannels[i].ChannelID = channelIDs[i]
		// Stagger creation times in descending order, so that default sort returns them in `channelIDs` order
		testChannels[i].CreatedAt = time.Now().Add(time.Duration(1)*time.Hour - time.Duration(i)*time.Minute)
	}

	for _, channel := range testChannels {
		require.NoError(t, router.DB.Create(&channel).Error)
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
			expectedChannelIDs: []string{"0xChannel01", "0xChannel03", "0xChannel05"}, // Only open channels, first 3
		},
	}

	for idx, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			paramsJSON, err := json.Marshal(tc.params)
			require.NoError(t, err)

			c := &RPCContext{
				Context: context.TODO(),
				Message: RPCMessage{
					Req: &RPCData{
						RequestID: uint64(idx),
						Method:    "get_channels",
						Params:    []any{json.RawMessage(paramsJSON)},
						Timestamp: uint64(time.Now().Unix()),
					},
					Sig: []string{"dummy-signature"},
				},
			}

			// Call handler
			router.HandleGetChannels(c)
			res := c.Message.Res
			require.NotNil(t, res)

			require.Len(t, res.Params, 1, "Response should contain an array of ChannelResponse")
			responseChannels, ok := res.Params[0].([]ChannelResponse)
			require.True(t, ok, "Response parameter should be a slice of ChannelResponse")
			assert.Len(t, responseChannels, len(tc.expectedChannelIDs), "Should return expected number of channels")

			// Check channel IDs are included in expected order
			for idx, channel := range responseChannels {
				assert.Equal(t, tc.expectedChannelIDs[idx], channel.ChannelID, "Should include channel %s at position %d", tc.expectedChannelIDs[idx], idx)
			}
		})
	}
}

func TestRPCRouterHandleGetAppDefinition_Success(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
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
	require.NoError(t, router.DB.Create(&session).Error)

	// Build RPC request
	params := map[string]string{"app_session_id": session.SessionID}
	b, _ := json.Marshal(params)
	c := &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 5,
				Method:    "get_app_definition",
				Params:    []any{json.RawMessage(b)},
				Timestamp: uint64(time.Now().Unix()),
			},
		},
	}

	// Call handler
	router.HandleGetAppDefinition(c)
	res := c.Message.Res
	require.NotNil(t, res)

	assert.Equal(t, "get_app_definition", res.Method)

	// Validate response payload
	def, ok := res.Params[0].(AppDefinition)
	require.True(t, ok)
	assert.Equal(t, session.Protocol, def.Protocol)
	assert.EqualValues(t, session.ParticipantWallets, def.ParticipantWallets)
	assert.EqualValues(t, session.Weights, def.Weights)
	assert.Equal(t, session.Quorum, def.Quorum)
	assert.Equal(t, session.Challenge, def.Challenge)
	assert.Equal(t, session.Nonce, def.Nonce)
}

func TestRPCRouterHandleGetAppDefinition_MissingID(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	c := &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 6,
				Method:    "get_app_definition",
				Params:    []any{},
				Timestamp: uint64(time.Now().Unix()),
			},
		},
	}

	// Call handler
	router.HandleGetAppDefinition(c)
	res := c.Message.Res
	require.NotNil(t, res)

	assert.Equal(t, "error", res.Method)
	require.Len(t, res.Params, 1)
	assert.Contains(t, res.Params[0], "missing account ID")
}

func TestRPCRouterHandleGetAppDefinition_NotFound(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	params := map[string]string{"app_session_id": "nonexistent"}
	b, _ := json.Marshal(params)
	c := &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 6,
				Method:    "get_app_definition",
				Params:    []any{json.RawMessage(b)},
				Timestamp: uint64(time.Now().Unix()),
			},
		},
	}

	// Call handler
	router.HandleGetAppDefinition(c)
	res := c.Message.Res
	require.NotNil(t, res)

	assert.Equal(t, "error", res.Method)
	require.Len(t, res.Params, 1)
	assert.Contains(t, res.Params[0], "failed to get application session")
}

func TestRPCRouterHandleGetAppSessions(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: rawKey}
	participantAddr := signer.GetAddress().Hex()

	// Create sessions with specific creation times to test sorting
	baseTime := time.Now().Add(-24 * time.Hour)
	sessions := []AppSession{
		{
			SessionID:          "0xSession1",
			ParticipantWallets: []string{participantAddr, "0xParticipant2"},
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
			ParticipantWallets: []string{participantAddr, "0xParticipant3"},
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
		require.NoError(t, router.DB.Create(&session).Error)
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
			params:             map[string]interface{}{"participant": participantAddr},
			expectedSessionIDs: []string{"0xSession2", "0xSession1"},
		},
		{
			name:               "Filter by participant with ascending sort",
			params:             map[string]interface{}{"participant": participantAddr, "sort": "asc"},
			expectedSessionIDs: []string{"0xSession1", "0xSession2"},
		},
		{
			name:               "Filter by status open",
			params:             map[string]interface{}{"status": string(ChannelStatusOpen)},
			expectedSessionIDs: []string{"0xSession3", "0xSession1"},
		},
		{
			name:               "Filter by participant and status open",
			params:             map[string]interface{}{"participant": participantAddr, "status": string(ChannelStatusOpen)},
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
			paramsJSON, err := json.Marshal(tc.params)
			require.NoError(t, err, "Failed to marshal params")

			c := &RPCContext{
				Context: context.TODO(),
				Message: RPCMessage{
					Req: &RPCData{
						RequestID: uint64(idx),
						Method:    "get_app_sessions",
						Params:    []any{json.RawMessage(paramsJSON)},
						Timestamp: uint64(time.Now().Unix()),
					},
					Sig: []string{"dummy-signature"},
				},
			}

			router.HandleGetAppSessions(c)
			res := c.Message.Res
			require.NotNil(t, res)

			assert.Equal(t, "get_app_sessions", res.Method)
			assert.Equal(t, uint64(idx), res.RequestID)
			require.Len(t, res.Params, 1, "Response should contain an array of AppSessionResponse")

			sessionResponses, ok := res.Params[0].([]AppSessionResponse)
			require.True(t, ok, "Response parameter should be a slice of AppSessionResponse")
			assert.Len(t, sessionResponses, len(tc.expectedSessionIDs), "Should return expected number of app sessions")

			for idx, sessionResponse := range sessionResponses {
				assert.True(t, sessionResponse.AppSessionID == tc.expectedSessionIDs[idx], "Should include session %s", tc.expectedSessionIDs[idx])
			}
		})
	}
}

func TestRPCRouterHandleGetAppSessions_Pagination(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
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
		require.NoError(t, router.DB.Create(&session).Error)
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
			paramsJSON, err := json.Marshal(tc.params)
			require.NoError(t, err)

			c := &RPCContext{
				Context: context.TODO(),
				Message: RPCMessage{
					Req: &RPCData{
						RequestID: uint64(idx),
						Method:    "get_app_sessions",
						Params:    []any{json.RawMessage(paramsJSON)},
						Timestamp: uint64(time.Now().Unix()),
					},
					Sig: []string{"dummy-signature"},
				},
			}

			// Call handler
			router.HandleGetAppSessions(c)
			res := c.Message.Res
			require.NotNil(t, res)

			require.Len(t, res.Params, 1, "Response should contain an array of AppSessionResponse")
			responseSessions, ok := res.Params[0].([]AppSessionResponse)
			require.True(t, ok, "Response parameter should be a slice of AppSessionResponse")
			assert.Len(t, responseSessions, len(tc.expectedSessionIDs), "Should return expected number of sessions")

			fmt.Println("Response Sessions:", responseSessions)

			// Check session IDs are in expected order
			for idx, session := range responseSessions {
				assert.True(t, session.AppSessionID == tc.expectedSessionIDs[idx], "Retrieved %d-th session ID should be equal %s", idx, tc.expectedSessionIDs[idx])
			}
		})
	}
}

func TestRPCRouterHandleGetLedgerEntries(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	participant1 := newTestCommonAddress("0xParticipant1")
	participant1AccountID := NewAccountID(participant1.Hex())
	participant2 := newTestCommonAddress("0xParticipant2")
	participant2AccountID := NewAccountID(participant2.Hex())

	// Setup test data
	ledger1 := GetWalletLedger(router.DB, participant1)
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
		err := ledger1.Record(participant1AccountID, data.asset, data.amount)
		require.NoError(t, err)
	}

	ledger2 := GetWalletLedger(router.DB, participant2)
	testData2 := []struct {
		asset  string
		amount decimal.Decimal
	}{
		{"usdc", decimal.NewFromInt(300)},
		{"btc", decimal.NewFromFloat(0.05)},
	}
	for _, data := range testData2 {
		err := ledger2.Record(participant2AccountID, data.asset, data.amount)
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
			paramsJSON, err := json.Marshal(tc.params)
			require.NoError(t, err)

			c := &RPCContext{
				Context: context.TODO(),
				UserID:  tc.userID,
				Message: RPCMessage{
					Req: &RPCData{
						RequestID: uint64(idx + 1),
						Method:    "get_ledger_entries",
						Params:    []any{json.RawMessage(paramsJSON)},
						Timestamp: uint64(time.Now().Unix()),
					},
					Sig: []string{"dummy-signature"},
				},
			}

			// Call handler
			router.HandleGetLedgerEntries(c)
			res := c.Message.Res
			require.NotNil(t, res)

			assert.Equal(t, "get_ledger_entries", res.Method)
			assert.Equal(t, uint64(idx+1), res.RequestID)
			require.Len(t, res.Params, 1, "Response should contain an array of Entry objects")

			entries, ok := res.Params[0].([]LedgerEntryResponse)
			require.True(t, ok, "Response parameter should be a slice of Entry")
			assert.Len(t, entries, tc.expectedCount, "Should return expected number of entries")

			tc.validateFunc(t, entries)
		})
	}
}

func TestRPCRouterHandleGetLedgerEntries_Pagination(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	userAddress := newTestCommonAddress("0xParticipant1")
	userAccountID := NewAccountID(userAddress.Hex())

	tokenNames := []string{
		"eth1", "eth2", "eth3", "eth4", "eth5", "eth6", "eth7", "eth8", "eth9", "eth10", "eth11"}

	// Create 11 ledger entries for pagination testing
	ledger := GetWalletLedger(router.DB, userAddress)
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
		err := ledger.Record(userAccountID, data.asset, data.amount)
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
			paramsJSON, err := json.Marshal(tc.params)
			require.NoError(t, err)

			c := &RPCContext{
				Context: context.TODO(),
				Message: RPCMessage{
					Req: &RPCData{
						RequestID: uint64(idx),
						Method:    "get_ledger_entries",
						Params:    []any{json.RawMessage(paramsJSON)},
						Timestamp: uint64(time.Now().Unix()),
					},
					Sig: []string{"dummy-signature"},
				},
			}

			// Call handler
			router.HandleGetLedgerEntries(c)
			res := c.Message.Res
			require.NotNil(t, res)

			require.Len(t, res.Params, 1, "Response should contain an array of LedgerEntryResponse")
			responseEntries, ok := res.Params[0].([]LedgerEntryResponse)
			require.True(t, ok, "Response parameter should be a slice of LedgerEntryResponse")
			assert.Len(t, responseEntries, len(tc.expectedToken), "Should return expected number of entries")

			// Check token names are included in expected order
			for idx, entry := range responseEntries {
				assert.Equal(t, tc.expectedToken[idx], entry.Asset, "Should include token %s at position %d", tc.expectedToken[idx], idx)
			}
		})
	}
}

func TestRPCRouterHandleGetTransactions(t *testing.T) {
	// --- 1. SETUP (Runs once for all test cases) ---
	router, cleanup := setupTestRPCRouter(t)
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
		require.NoError(t, router.DB.Create(&tempTx).Error)
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
					assert.NotEmpty(t, tx.TxHash)
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
			// Marshal params for the current test case
			paramsJSON, err := json.Marshal(tc.params)
			require.NoError(t, err)

			// Create a new context for each request
			c := &RPCContext{
				Context: context.TODO(),
				Message: RPCMessage{
					Req: &RPCData{
						RequestID: uint64(i + 1), // Unique request ID per test case
						Method:    "get_ledger_transactions",
						Params:    []any{json.RawMessage(paramsJSON)},
						Timestamp: uint64(time.Now().Unix()),
					},
				},
			}

			// Call the handler
			router.HandleGetLedgerTransactions(c)

			// General assertions for all cases
			res := c.Message.Res
			require.NotNil(t, res, "Response should not be nil")
			assert.Equal(t, "get_ledger_transactions", res.Method)
			require.Len(t, res.Params, 1, "Response should contain one parameter")

			// Unmarshal the actual transaction data
			var transactions []TransactionResponse
			// We need to marshal the interface{} back to JSON, then unmarshal into our concrete type.
			respBytes, err := json.Marshal(res.Params[0])
			require.NoError(t, err)
			err = json.Unmarshal(respBytes, &transactions)
			require.NoError(t, err, "Response parameter should be a slice of TransactionResponse")

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
		params := map[string]any{"tx_type": "nonexistent"}
		paramsJSON, err := json.Marshal(params)
		require.NoError(t, err)

		c := &RPCContext{
			Context: context.TODO(),
			Message: RPCMessage{
				Req: &RPCData{
					RequestID: 999,
					Method:    "get_ledger_transactions",
					Params:    []any{json.RawMessage(paramsJSON)},
					Timestamp: uint64(time.Now().Unix()),
				},
			},
		}

		// Call the handler
		router.HandleGetLedgerTransactions(c)

		// Should return an error response
		res := c.Message.Res
		require.NotNil(t, res, "Response should not be nil")
		assert.Equal(t, "error", res.Method, "Should return error method")
		require.Len(t, res.Params, 1, "Error response should contain one parameter")

		// Verify error message
		errorMsg, ok := res.Params[0].(string)
		require.True(t, ok, "Error parameter should be a string")
		assert.Equal(t, ErrInvalidTransactionType.Error(), errorMsg, "Should return correct error message")
	})
}

func TestRPCRouterHandleGetLedgerTransactions_Pagination(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
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

	// Insert all test transactions
	for _, tx := range testTransactions {
		tempTx := tx
		require.NoError(t, router.DB.Create(&tempTx).Error)
	}

	// Expected order: most recent first (descending by created_at)
	expectedHashes := make([]string, 11)
	for i := 0; i < 11; i++ {
		var tx LedgerTransaction
		require.NoError(t, router.DB.Where("created_at = ?", testTransactions[10-i].CreatedAt).First(&tx).Error)
		expectedHashes[i] = tx.Hash
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
			expectedCount: 3, // Ascending order, skip 1, take 3
			expectedFirst: expectedHashes[9], // 2nd oldest
			expectedLast:  expectedHashes[7],  // 4th oldest
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
			paramsJSON, err := json.Marshal(tc.params)
			require.NoError(t, err)

			c := &RPCContext{
				Context: context.TODO(),
				Message: RPCMessage{
					Req: &RPCData{
						RequestID: uint64(idx + 100),
						Method:    "get_ledger_transactions",
						Params:    []any{json.RawMessage(paramsJSON)},
						Timestamp: uint64(time.Now().Unix()),
					},
				},
			}

			// Call handler
			router.HandleGetLedgerTransactions(c)
			res := c.Message.Res
			require.NotNil(t, res)

			assert.Equal(t, "get_ledger_transactions", res.Method)
			require.Len(t, res.Params, 1, "Response should contain an array of TransactionResponse")

			var transactions []TransactionResponse
			respBytes, err := json.Marshal(res.Params[0])
			require.NoError(t, err)
			err = json.Unmarshal(respBytes, &transactions)
			require.NoError(t, err)

			assert.Len(t, transactions, tc.expectedCount, "Should return expected number of transactions")

			// For non-filter tests, verify order
			if tc.expectedFirst != "" && tc.expectedLast != "" && len(transactions) > 0 {
				assert.Equal(t, tc.expectedFirst, transactions[0].TxHash, "First transaction hash should match")
				if len(transactions) > 1 {
					assert.Equal(t, tc.expectedLast, transactions[len(transactions)-1].TxHash, "Last transaction hash should match")
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
}
