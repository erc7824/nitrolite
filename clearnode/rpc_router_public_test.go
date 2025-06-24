package main

import (
	"context"
	"encoding/json"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestRPCRouterHandleGetChannels(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: rawKey}
	participantSigner := signer.GetAddress().Hex()
	participantWallet := "wallet_address"
	tokenAddress := "0xToken123"
	chainID := uint32(137)

	channels := []Channel{
		{
			ChannelID:   "0xChannel1",
			Wallet:      participantWallet,
			Participant: participantSigner,
			Status:      ChannelStatusOpen,
			Token:       tokenAddress + "1",
			ChainID:     chainID,
			Amount:      1000,
			Nonce:       1,
			Version:     10,
			Challenge:   86400,
			Adjudicator: "0xAdj1",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ChannelID:   "0xChannel2",
			Wallet:      participantWallet,
			Participant: participantSigner,
			Status:      ChannelStatusClosed,
			Token:       tokenAddress + "2",
			ChainID:     chainID,
			Amount:      2000,
			Nonce:       2,
			Version:     20,
			Challenge:   86400,
			Adjudicator: "0xAdj2",
			CreatedAt:   time.Now().Add(-12 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ChannelID:   "0xChannel3",
			Wallet:      participantWallet,
			Participant: participantSigner,
			Status:      ChannelStatusJoining,
			Token:       tokenAddress + "3",
			ChainID:     chainID,
			Amount:      3000,
			Nonce:       3,
			Version:     30,
			Challenge:   86400,
			Adjudicator: "0xAdj3",
			CreatedAt:   time.Now().Add(-6 * time.Hour),
			UpdatedAt:   time.Now(),
		},
	}

	for _, channel := range channels {
		require.NoError(t, router.DB.Create(&channel).Error)
	}

	otherChannel := Channel{
		ChannelID:   "0xOtherChannel",
		Participant: "0xOtherParticipant",
		Status:      ChannelStatusOpen,
		Token:       tokenAddress + "4",
		ChainID:     chainID,
		Amount:      5000,
		Nonce:       4,
		Version:     40,
		Challenge:   86400,
		Adjudicator: "0xAdj4",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, router.DB.Create(&otherChannel).Error)

	params := map[string]string{
		"participant": participantWallet,
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	c := &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 123,
				Method:    "get_channels",
				Params:    []any{json.RawMessage(paramsJSON)},
				Timestamp: uint64(time.Now().Unix()),
			},
		},
	}

	// Call handler
	router.HandleGetChannels(c)
	res := c.Message.Res
	require.NotNil(t, res)

	assert.Equal(t, "get_channels", res.Method)
	assert.Equal(t, uint64(123), res.RequestID)

	require.Len(t, res.Params, 1, "Response should contain a slice of ChannelResponse")
	channelsSlice, ok := res.Params[0].([]ChannelResponse)
	require.True(t, ok, "Response parameter should be a slice of ChannelResponse")

	// Expect 3 channels for this participant, ordered newest first
	assert.Len(t, channelsSlice, 3, "Should return all 3 channels for the participant")
	assert.Equal(t, "0xChannel3", channelsSlice[0].ChannelID, "First channel should be the newest")
	assert.Equal(t, "0xChannel2", channelsSlice[1].ChannelID, "Second channel should be the middle one")
	assert.Equal(t, "0xChannel1", channelsSlice[2].ChannelID, "Third channel should be the oldest")

	for _, ch := range channelsSlice {
		assert.Equal(t, participantSigner, ch.Participant, "ParticipantA should match")
		assert.True(t, strings.HasPrefix(ch.Token, tokenAddress), "Token should start with the base token address")
		assert.Equal(t, chainID, ch.ChainID, "NetworkID should match")

		var originalChannel Channel
		for _, c := range channels {
			if c.ChannelID == ch.ChannelID {
				originalChannel = c
				break
			}
		}

		assert.Equal(t, originalChannel.Status, ch.Status, "Status should match")
		assert.Equal(t, big.NewInt(int64(originalChannel.Amount)), ch.Amount, "Amount should match")
		assert.Equal(t, originalChannel.Nonce, ch.Nonce, "Nonce should match")
		assert.Equal(t, originalChannel.Version, ch.Version, "Version should match")
		assert.Equal(t, originalChannel.Challenge, ch.Challenge, "Challenge should match")
		assert.Equal(t, originalChannel.Adjudicator, ch.Adjudicator, "Adjudicator should match")
		assert.NotEmpty(t, ch.CreatedAt, "CreatedAt should not be empty")
		assert.NotEmpty(t, ch.UpdatedAt, "UpdatedAt should not be empty")
	}

	// Filter by status="open"
	openStatusParams := map[string]string{
		"participant": participantWallet,
		"status":      string(ChannelStatusOpen),
	}
	openStatusParamsJSON, err := json.Marshal(openStatusParams)
	require.NoError(t, err)

	c = &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 456,
				Method:    "get_channels",
				Params:    []any{json.RawMessage(openStatusParamsJSON)},
				Timestamp: uint64(time.Now().Unix()),
			},
		},
	}

	// Call handler
	router.HandleGetChannels(c)
	res = c.Message.Res
	require.NotNil(t, res)

	openChannels, ok := res.Params[0].([]ChannelResponse)
	require.True(t, ok, "Response parameter should be a slice of ChannelResponse")
	assert.Len(t, openChannels, 1, "Should return only 1 open channel")
	assert.Equal(t, "0xChannel1", openChannels[0].ChannelID, "Should return the open channel")
	assert.Equal(t, ChannelStatusOpen, openChannels[0].Status, "Status should be open")

	// Filter by status="closed"
	closedStatusParams := map[string]string{
		"participant": participantWallet,
		"status":      string(ChannelStatusClosed),
	}
	closedStatusParamsJSON, err := json.Marshal(closedStatusParams)
	require.NoError(t, err)

	c = &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 457,
				Method:    "get_channels",
				Params:    []any{json.RawMessage(closedStatusParamsJSON)},
				Timestamp: uint64(time.Now().Unix()),
			},
		},
	}

	// Call handler
	router.HandleGetChannels(c)
	res = c.Message.Res
	require.NotNil(t, res)

	closedChannels, ok := res.Params[0].([]ChannelResponse)
	require.True(t, ok, "Response parameter should be a slice of ChannelResponse")
	assert.Len(t, closedChannels, 1, "Should return only 1 closed channel")
	assert.Equal(t, "0xChannel2", closedChannels[0].ChannelID, "Should return the closed channel")
	assert.Equal(t, ChannelStatusClosed, closedChannels[0].Status, "Status should be closed")

	// Filter by status="joining"
	joiningStatusParams := map[string]string{
		"participant": participantWallet,
		"status":      string(ChannelStatusJoining),
	}
	joiningStatusParamsJSON, err := json.Marshal(joiningStatusParams)
	require.NoError(t, err)

	c = &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 458,
				Method:    "get_channels",
				Params:    []any{json.RawMessage(joiningStatusParamsJSON)},
				Timestamp: uint64(time.Now().Unix()),
			},
		},
	}

	// Call handler
	router.HandleGetChannels(c)
	res = c.Message.Res
	require.NotNil(t, res)

	joiningChannels, ok := res.Params[0].([]ChannelResponse)
	require.True(t, ok, "Response parameter should be a slice of ChannelResponse")
	assert.Len(t, joiningChannels, 1, "Should return only 1 joining channel")
	assert.Equal(t, "0xChannel3", joiningChannels[0].ChannelID, "Should return the joining channel")
	assert.Equal(t, ChannelStatusJoining, joiningChannels[0].Status, "Status should be joining")

	// No participant parameter: return all 4 channels
	c = &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 789,
				Method:    "get_channels",
				Params:    []any{map[string]string{}},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{},
		},
	}

	// Call handler
	router.HandleGetChannels(c)
	res = c.Message.Res
	require.NotNil(t, res)

	allChannels, ok := res.Params[0].([]ChannelResponse)
	require.True(t, ok, "Response parameter should be a slice of ChannelResponse")
	assert.Len(t, allChannels, 4, "Should return all 4 channels")

	foundChannelIDs := make(map[string]bool)
	for _, channel := range allChannels {
		foundChannelIDs[channel.ChannelID] = true
	}
	assert.True(t, foundChannelIDs["0xChannel1"], "Should include Channel1")
	assert.True(t, foundChannelIDs["0xChannel2"], "Should include Channel2")
	assert.True(t, foundChannelIDs["0xChannel3"], "Should include Channel3")
	assert.True(t, foundChannelIDs["0xOtherChannel"], "Should include OtherChannel")

	// No participant but status="open": return 2 open channels
	openStatusOnlyParams := map[string]string{
		"status": string(ChannelStatusOpen),
	}
	openStatusOnlyParamsJSON, err := json.Marshal(openStatusOnlyParams)
	require.NoError(t, err)

	c = &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 790,
				Method:    "get_channels",
				Params:    []any{json.RawMessage(openStatusOnlyParamsJSON)},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{},
		},
	}

	// Call handler
	router.HandleGetChannels(c)
	res = c.Message.Res
	require.NotNil(t, res)

	openChannelsOnly, ok := res.Params[0].([]ChannelResponse)
	require.True(t, ok, "Response parameter should be a slice of ChannelResponse")
	assert.Len(t, openChannelsOnly, 2, "Should return 2 open channels")

	openChannelIDs := make(map[string]bool)
	for _, channel := range openChannelsOnly {
		openChannelIDs[channel.ChannelID] = true
		assert.Equal(t, ChannelStatusOpen, channel.Status, "All channels should have open status")
	}

	assert.True(t, openChannelIDs["0xChannel1"], "Should include open Channel1")
	assert.True(t, openChannelIDs["0xOtherChannel"], "Should include open OtherChannel")
	assert.False(t, openChannelIDs["0xChannel2"], "Should not include closed Channel2")
	assert.False(t, openChannelIDs["0xChannel3"], "Should not include joining Channel3")
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
			name:               "Get all with no sort (default asc)",
			params:             map[string]interface{}{},
			expectedTokenNames: []string{"0xToken1", "0xToken2", "0xToken3", "0xToken4"},
		},
		{
			name:               "Get all with ascending sort",
			params:             map[string]interface{}{"sort": "asc"},
			expectedTokenNames: []string{"0xToken1", "0xToken2", "0xToken3", "0xToken4"},
		},
		{
			name:               "Get all with descending sort",
			params:             map[string]interface{}{"sort": "desc"},
			expectedTokenNames: []string{"0xToken4", "0xToken3", "0xToken2", "0xToken1"},
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

			assets1, ok := res.Params[0].([]GetAssetsResponse)
			require.True(t, ok, "Response parameter should be a slice of AssetResponse")
			assert.Len(t, assets1, len(tc.expectedTokenNames), "Should return expected number of assets")

			foundTokens := make(map[string]bool)
			for _, asset := range assets1 {
				foundTokens[asset.Token] = true
			}

			for _, expectedTokenName := range tc.expectedTokenNames {
				assert.True(t, foundTokens[expectedTokenName], "Should include token %s", expectedTokenName)
			}
		})
	}
}

func TestRPCRouterHandleGetAssets_Pagination(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	tokenNames := []string{
		"0xToken01", "0xToken02", "0xToken03", "0xToken04",
		"0xToken05", "0xToken06", "0xToken07", "0xToken08",
		"0xToken09", "0xToken10", "0xToken11"}

	testAssets := []Asset{
		{ChainID: 1, Symbol: "eth", Decimals: 18},
		{ChainID: 1, Symbol: "weth", Decimals: 18},
		{ChainID: 1, Symbol: "wbtc", Decimals: 18},
		{ChainID: 1, Symbol: "usdc", Decimals: 6},
		{ChainID: 137, Symbol: "pol", Decimals: 18},
		{ChainID: 137, Symbol: "weth", Decimals: 18},
		{ChainID: 137, Symbol: "wbtc", Decimals: 18},
		{ChainID: 137, Symbol: "usdc", Decimals: 6},
		{ChainID: 42220, Symbol: "usdc", Decimals: 6},
		{ChainID: 42220, Symbol: "celo", Decimals: 18},
		{ChainID: 8453, Symbol: "usdbc", Decimals: 6},
	}

	for i := range testAssets {
		testAssets[i].Token = tokenNames[i]
	}

	for _, asset := range testAssets {
		require.NoError(t, router.DB.Create(&asset).Error)
	}

	tcs := []struct {
		name               string
		params             map[string]interface{}
		expectedTokenNames []string
	}{
		{name: "No params",
			params:             map[string]interface{}{},
			expectedTokenNames: tokenNames[:10], // Default pagination should return first 10 tokens
		},
		{name: "Offset only",
			params:             map[string]interface{}{"offset": float64(2)},
			expectedTokenNames: tokenNames[2:11], // Default page_size is 10, total 11, so offset 2 returns Tokens 3 to 11
		},
		{name: "Page size only",
			params:             map[string]interface{}{"page_size": float64(5)},
			expectedTokenNames: tokenNames[:5], // Default offset is 0, so page_size 5 returns First 5 tokens
		},
		{name: "Offset and page size",
			params:             map[string]interface{}{"offset": float64(2), "page_size": float64(3)},
			expectedTokenNames: tokenNames[2:5], // Offset 2 with page_size 3 returns Tokens 3 to 5
		},
		{name: "Pagination with sort",
			params:             map[string]interface{}{"offset": float64(2), "page_size": float64(3), "sort": "desc"},
			expectedTokenNames: []string{"0xToken09", "0xToken08", "0xToken07"}, // Offset 2 with page_size 3 returns Tokens 9 to 7
		},
		{name: "Pagination with chain_id",
			params:             map[string]interface{}{"chain_id": float64(137), "offset": float64(1), "page_size": float64(2)},
			expectedTokenNames: tokenNames[5:7], // Chain ID 137 with offset 1 and page_size 2 returns Tokens 6 to 7
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
						Method:    "get_assets",
						Params:    []any{json.RawMessage(paramsJSON)},
						Timestamp: uint64(time.Now().Unix()),
					},
					Sig: []string{"dummy-signature"},
				},
			}

			// Call handler
			router.HandleGetAssets(c)
			res := c.Message.Res
			require.NotNil(t, res)

			require.Len(t, res.Params, 1, "Response should contain an array of AssetResponse")
			assets, ok := res.Params[0].([]GetAssetsResponse)
			require.True(t, ok, "Response parameter should be a slice of AssetResponse")
			assert.Len(t, assets, len(tc.expectedTokenNames), "Should return expected number of assets")

			// Check token names are included
			foundTokens := make(map[string]bool)
			for _, asset := range assets {
				foundTokens[asset.Token] = true
			}

			for _, tokenName := range tc.expectedTokenNames {
				assert.True(t, foundTokens[tokenName], "Should include token %s", tokenName)
			}
		})
	}
}

func TestRPCRouterHandleGetAppSessions(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: rawKey}
	participantAddr := signer.GetAddress().Hex()

	sessions := []AppSession{
		{
			SessionID:          "0xSession1",
			ParticipantWallets: []string{participantAddr, "0xParticipant2"},
			Status:             ChannelStatusOpen,
			Protocol:           "test-app-1",
			Challenge:          60,
			Weights:            []int64{50, 50},
			Quorum:             75,
			Nonce:              1,
			Version:            1,
		},
		{
			SessionID:          "0xSession2",
			ParticipantWallets: []string{participantAddr, "0xParticipant3"},
			Status:             ChannelStatusClosed,
			Protocol:           "test-app-2",
			Challenge:          120,
			Weights:            []int64{30, 70},
			Quorum:             80,
			Nonce:              2,
			Version:            2,
		},
		{
			SessionID:          "0xSession3",
			ParticipantWallets: []string{"0xParticipant4", "0xParticipant5"},
			Status:             ChannelStatusOpen,
			Protocol:           "test-app-3",
			Challenge:          90,
			Weights:            []int64{40, 60},
			Quorum:             60,
			Nonce:              3,
			Version:            3,
		},
	}

	for _, session := range sessions {
		require.NoError(t, router.DB.Create(&session).Error)
	}

	// Case 1: Get all for participant
	params1 := map[string]string{"participant": participantAddr}
	paramsJSON1, err := json.Marshal(params1)
	require.NoError(t, err)

	c := &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 1,
				Method:    "get_app_sessions",
				Params:    []any{json.RawMessage(paramsJSON1)},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	// Call handler
	router.HandleGetAppSessions(c)
	res := c.Message.Res
	require.NotNil(t, res)

	assert.Equal(t, "get_app_sessions", res.Method)
	assert.Equal(t, uint64(1), res.RequestID)
	require.Len(t, res.Params, 1, "Response should contain an array of AppSessionResponse")

	sessionResponses, ok := res.Params[0].([]AppSessionResponse)
	require.True(t, ok, "Response parameter should be a slice of AppSessionResponse")
	assert.Len(t, sessionResponses, 2, "Should return 2 app sessions for the participant")

	foundSessions := make(map[string]bool)
	for _, session := range sessionResponses {
		foundSessions[session.AppSessionID] = true
		var orig AppSession
		for _, s := range sessions {
			if s.SessionID == session.AppSessionID {
				orig = s
				break
			}
		}
		assert.Equal(t, string(orig.Status), session.Status, "Status should match")
	}
	assert.True(t, foundSessions["0xSession1"], "Should include Session1")
	assert.True(t, foundSessions["0xSession2"], "Should include Session2")
	assert.False(t, foundSessions["0xSession3"], "Should not include Session3")

	// Case 2: Filter by status="open"
	params2 := map[string]string{"participant": participantAddr, "status": string(ChannelStatusOpen)}
	paramsJSON2, err := json.Marshal(params2)
	require.NoError(t, err)

	c = &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 2,
				Method:    "get_app_sessions",
				Params:    []any{json.RawMessage(paramsJSON2)},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	// Call handler
	router.HandleGetAppSessions(c)
	res = c.Message.Res
	require.NotNil(t, res)

	sessionResponses2, ok := res.Params[0].([]AppSessionResponse)
	require.True(t, ok, "Response parameter should be a slice of AppSessionResponse")
	assert.Len(t, sessionResponses2, 1, "Should return 1 open app session")
	assert.Equal(t, "0xSession1", sessionResponses2[0].AppSessionID, "Should be Session1")
	assert.Equal(t, string(ChannelStatusOpen), sessionResponses2[0].Status)

	// Case 3: No participant (all sessions)
	c = &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 3,
				Method:    "get_app_sessions",
				Params:    []any{json.RawMessage(`{}`)},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	// Call handler
	router.HandleGetAppSessions(c)
	res = c.Message.Res
	require.NotNil(t, res)

	allSessions, ok := res.Params[0].([]AppSessionResponse)
	require.True(t, ok, "Response parameter should be a slice of AppSessionResponse")
	assert.Len(t, allSessions, 3, "Should return all 3 app sessions")

	foundSessionIDs := make(map[string]bool)
	for _, session := range allSessions {
		foundSessionIDs[session.AppSessionID] = true
	}
	assert.True(t, foundSessionIDs["0xSession1"], "Should include Session1")
	assert.True(t, foundSessionIDs["0xSession2"], "Should include Session2")
	assert.True(t, foundSessionIDs["0xSession3"], "Should include Session3")

	// Case 4: No participant, status="open"
	openStatusParams := map[string]string{"status": string(ChannelStatusOpen)}
	openStatusParamsJSON, err := json.Marshal(openStatusParams)
	require.NoError(t, err)

	c = &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 4,
				Method:    "get_app_sessions",
				Params:    []any{json.RawMessage(openStatusParamsJSON)},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	// Call handler
	router.HandleGetAppSessions(c)
	res = c.Message.Res
	require.NotNil(t, res)

	openSessions, ok := res.Params[0].([]AppSessionResponse)
	require.True(t, ok, "Response parameter should be a slice of AppSessionResponse")
	assert.Len(t, openSessions, 2, "Should return 2 open sessions")

	openSessionIDs := make(map[string]bool)
	for _, session := range openSessions {
		openSessionIDs[session.AppSessionID] = true
		assert.Equal(t, string(ChannelStatusOpen), session.Status, "All sessions should be open")
	}
	assert.True(t, openSessionIDs["0xSession1"], "Should include Session1")
	assert.True(t, openSessionIDs["0xSession3"], "Should include Session3")
	assert.False(t, openSessionIDs["0xSession2"], "Should not include Session2")
}

func TestRPCRouterHandleGetLedgerEntries(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	participant1 := "0xParticipant1"
	participant2 := "0xParticipant2"

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
		err := ledger1.Record(participant1, data.asset, data.amount)
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
		err := ledger2.Record(participant2, data.asset, data.amount)
		require.NoError(t, err)
	}

	// Case 1: Filter by account_id only
	params1 := map[string]string{"account_id": participant1}
	paramsJSON1, err := json.Marshal(params1)
	require.NoError(t, err)

	c := &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 1,
				Method:    "get_ledger_entries",
				Params:    []any{json.RawMessage(paramsJSON1)},
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
	assert.Equal(t, uint64(1), res.RequestID)
	require.Len(t, res.Params, 1, "Response should contain an array of Entry objects")

	entries1, ok := res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")
	assert.Len(t, entries1, 5, "Should return all 5 entries for participant1")

	assetCounts := map[string]int{}
	for _, entry := range entries1 {
		assetCounts[entry.Asset]++
		assert.Equal(t, participant1, entry.AccountID)
		assert.Equal(t, participant1, entry.Participant)
	}
	assert.Equal(t, 3, assetCounts["usdc"], "Should have 3 USDC entries")
	assert.Equal(t, 2, assetCounts["eth"], "Should have 2 ETH entries")

	// Case 2: Filter by account_id and asset
	params2 := map[string]string{"account_id": participant1, "asset": "usdc"}
	paramsJSON2, err := json.Marshal(params2)
	require.NoError(t, err)

	c = &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 2,
				Method:    "get_ledger_entries",
				Params:    []any{json.RawMessage(paramsJSON2)},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	// Call handler
	router.HandleGetLedgerEntries(c)
	res = c.Message.Res
	require.NotNil(t, res)

	entries2, ok := res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")
	assert.Len(t, entries2, 3, "Should return 3 USDC entries for participant1")

	for _, entry := range entries2 {
		assert.Equal(t, "usdc", entry.Asset)
		assert.Equal(t, participant1, entry.AccountID)
		assert.Equal(t, participant1, entry.Participant)
	}

	// Case 3: Filter by wallet only
	params3 := map[string]string{"wallet": participant2}
	paramsJSON3, err := json.Marshal(params3)
	require.NoError(t, err)

	c = &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 3,
				Method:    "get_ledger_entries",
				Params:    []any{json.RawMessage(paramsJSON3)},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	// Call handler
	router.HandleGetLedgerEntries(c)
	res = c.Message.Res
	require.NotNil(t, res)

	entries3, ok := res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")
	assert.Len(t, entries3, 2, "Should return all 2 entries for participant2")

	for _, entry := range entries3 {
		assert.Equal(t, participant2, entry.Participant)
	}

	// Case 4: Filter by wallet and asset
	params4 := map[string]string{"wallet": participant2, "asset": "usdc"}
	paramsJSON4, err := json.Marshal(params4)
	require.NoError(t, err)

	c = &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 4,
				Method:    "get_ledger_entries",
				Params:    []any{json.RawMessage(paramsJSON4)},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	// Call handler
	router.HandleGetLedgerEntries(c)
	res = c.Message.Res
	require.NotNil(t, res)

	entries4, ok := res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")
	assert.Len(t, entries4, 1, "Should return 1 entry for participant2 with usdc")
	assert.Equal(t, "usdc", entries4[0].Asset)
	assert.Equal(t, participant2, entries4[0].Participant)

	// Case 5: Filter by account_id and wallet (no overlap)
	params5 := map[string]string{"account_id": participant1, "wallet": participant2}
	paramsJSON5, err := json.Marshal(params5)
	require.NoError(t, err)

	c = &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 5,
				Method:    "get_ledger_entries",
				Params:    []any{json.RawMessage(paramsJSON5)},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	// Call handler
	router.HandleGetLedgerEntries(c)
	res = c.Message.Res
	require.NotNil(t, res)

	entries5, ok := res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")
	assert.Len(t, entries5, 0, "Should return 0 entries when account_id and wallet don't match")

	// Case 6: No filters (all entries)
	c = &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 6,
				Method:    "get_ledger_entries",
				Params:    []any{map[string]string{}}, // Empty map
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	// Call handler
	router.HandleGetLedgerEntries(c)
	res = c.Message.Res
	require.NotNil(t, res)

	entries6, ok := res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")
	assert.Len(t, entries6, 7, "Should return all 7 entries")

	foundParticipants := make(map[string]bool)
	for _, entry := range entries6 {
		foundParticipants[entry.Participant] = true
	}
	assert.True(t, foundParticipants[participant1], "Should include entries for participant1")
	assert.True(t, foundParticipants[participant2], "Should include entries for participant2")

	// Case 7: Default wallet provided
	c = &RPCContext{
		Context: context.TODO(),
		UserID:  participant1,
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 7,
				Method:    "get_ledger_entries",
				Params:    []any{map[string]string{}},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	// Call handler
	router.HandleGetLedgerEntries(c)
	res = c.Message.Res
	require.NotNil(t, res)

	entries7, ok := res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")
	assert.Len(t, entries7, 5, "Should return 5 entries for default wallet participant1")

	for _, entry := range entries7 {
		assert.Equal(t, participant1, entry.Participant)
	}
}
