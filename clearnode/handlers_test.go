package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	container "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestSqlite creates an in-memory SQLite DB for testing
func setupTestSqlite(t testing.TB) *gorm.DB {
	t.Helper()

	uniqueDSN := fmt.Sprintf("file::memory:test%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(uniqueDSN), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&Entry{}, &Channel{}, &AppSession{}, &RPCRecord{}, &Asset{}, &SignerWallet{})
	require.NoError(t, err)

	return db
}

// setupTestPostgres creates a PostgreSQL database using testcontainers
func setupTestPostgres(ctx context.Context, t testing.TB) (*gorm.DB, testcontainers.Container) {
	t.Helper()

	const dbName = "postgres"
	const dbUser = "postgres"
	const dbPassword = "postgres"

	postgresContainer, err := container.Run(ctx,
		"postgres:16-alpine",
		container.WithDatabase(dbName),
		container.WithUsername(dbUser),
		container.WithPassword(dbPassword),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForLog("database system is ready to accept connections"),
				wait.ForListeningPort("5432/tcp"),
			)))
	require.NoError(t, err)
	log.Println("Started container:", postgresContainer.GetContainerID())

	url, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	log.Println("PostgreSQL URL:", url)

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&Entry{}, &Channel{}, &AppSession{}, &RPCRecord{}, &Asset{})
	require.NoError(t, err)

	return db, postgresContainer
}

// setupTestDB chooses SQLite or Postgres based on TEST_DB_DRIVER
func setupTestDB(t testing.TB) (*gorm.DB, func()) {
	t.Helper()

	ctx := context.Background()
	var db *gorm.DB
	var cleanup func()

	switch os.Getenv("TEST_DB_DRIVER") {
	case "postgres":
		log.Println("Using PostgreSQL for testing")
		var container testcontainers.Container
		db, container = setupTestPostgres(ctx, t)
		cleanup = func() {
			if container != nil {
				if err := container.Terminate(ctx); err != nil {
					log.Printf("Failed to terminate PostgreSQL container: %v", err)
				}
			}
		}
	default:
		log.Println("Using SQLite for testing (default)")
		db = setupTestSqlite(t)
		cleanup = func() {}
	}

	return db, cleanup
}

// TestHandlePing tests the ping handler functionality
func TestHandlePing(t *testing.T) {
	rpcRequest1 := &RPCMessage{
		Req: &RPCData{
			RequestID: 1,
			Method:    "ping",
			Params:    []any{nil},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}
	// No need to set ReqRaw here, because HandlePing doesn’t validate signatures.

	response1, err := HandlePing(rpcRequest1)
	require.NoError(t, err)
	assert.NotNil(t, response1)
	require.Equal(t, "pong", response1.Res.Method)
}

// TestHandleGetAppDefinition tests the GetAppDefinition handler
func TestHandleGetAppDefinition_Success(t *testing.T) {
	// Setup
	db, cleanup := setupTestDB(t)
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

	// Build RPC request
	params := map[string]string{"app_session_id": session.SessionID}
	b, _ := json.Marshal(params)
	rpcReq := &RPCMessage{Req: &RPCData{RequestID: 5, Method: "get_app_definition", Params: []any{json.RawMessage(b)}, Timestamp: uint64(time.Now().Unix())}}

	// Call handler
	resp, err := HandleGetAppDefinition(rpcReq, db)
	require.NoError(t, err)
	assert.Equal(t, "get_app_definition", resp.Res.Method)

	// Validate response payload
	def, ok := resp.Res.Params[0].(AppDefinition)
	require.True(t, ok)
	assert.Equal(t, session.Protocol, def.Protocol)
	assert.EqualValues(t, session.ParticipantWallets, def.ParticipantWallets)
	assert.EqualValues(t, session.Weights, def.Weights)
	assert.Equal(t, session.Quorum, def.Quorum)
	assert.Equal(t, session.Challenge, def.Challenge)
	assert.Equal(t, session.Nonce, def.Nonce)
}

func TestHandleGetAppDefinition_MissingID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	rpcReq := &RPCMessage{Req: &RPCData{RequestID: 6, Method: "get_app_definition", Params: []any{}, Timestamp: uint64(time.Now().Unix())}}

	_, err := HandleGetAppDefinition(rpcReq, db)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing account ID")
}

func TestHandleGetAppDefinition_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	params := map[string]string{"app_session_id": "nonexistent"}
	b, _ := json.Marshal(params)
	rpcReq := &RPCMessage{Req: &RPCData{RequestID: 7, Method: "get_app_definition", Params: []any{json.RawMessage(b)}, Timestamp: uint64(time.Now().Unix())}}

	_, err := HandleGetAppDefinition(rpcReq, db)
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to find application")
}

// TestHandleGetLedgerBalances tests the get ledger balances handler functionality
func TestHandleGetLedgerBalances(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ledger := GetWalletLedger(db, "0xParticipant1")
	err := ledger.Record("0xParticipant1", "usdc", decimal.NewFromInt(1000))
	require.NoError(t, err)

	params := map[string]string{"account_id": "0xParticipant1"}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	rpcRequest := &RPCMessage{
		Req: &RPCData{
			RequestID: 1,
			Method:    "get_ledger_balances",
			Params:    []any{json.RawMessage(paramsJSON)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}
	// Handler does not validate signature here, so we don’t need ReqRaw.

	msg, err := HandleGetLedgerBalances(rpcRequest, "0xParticipant1", db)
	require.NoError(t, err)
	assert.NotNil(t, msg)

	responseParams := msg.Res.Params
	require.NotEmpty(t, responseParams)

	balancesArray, ok := responseParams[0].([]Balance)
	require.True(t, ok, "Response should contain an array of Balance")
	assert.Equal(t, 1, len(balancesArray), "Should have 1 balance entry")

	expectedAssets := map[string]decimal.Decimal{"usdc": decimal.NewFromInt(1000)}
	for _, balance := range balancesArray {
		expectedBalance, exists := expectedAssets[balance.Asset]
		assert.True(t, exists, "Unexpected asset in response: %s", balance.Asset)
		assert.Equal(t, expectedBalance, balance.Amount, "Incorrect balance for asset %s", balance.Asset)
		delete(expectedAssets, balance.Asset)
	}
	assert.Empty(t, expectedAssets, "Not all expected assets were found")
}

// TestHandleGetConfig tests the get config handler functionality
func TestHandleGetConfig(t *testing.T) {
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

	rpcRequest := &RPCMessage{
		Req: &RPCData{
			RequestID: 1,
			Method:    "get_config",
			Params:    []any{},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}
	// HandleGetConfig does not validate signature, so no ReqRaw needed.

	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: rawKey}

	response, err := HandleGetConfig(rpcRequest, mockConfig, &signer)
	require.NoError(t, err)
	assert.NotNil(t, response)

	responseParams := response.Res.Params
	require.NotEmpty(t, responseParams)

	configMap, ok := responseParams[0].(BrokerConfig)
	require.True(t, ok, "Response should contain a BrokerConfig")
	assert.Equal(t, signer.GetAddress().Hex(), configMap.BrokerAddress)
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

// TestHandleGetChannels tests the get channels functionality
func TestHandleGetChannels(t *testing.T) {
	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: rawKey}
	participantSigner := signer.GetAddress().Hex()
	participantWallet := "wallet_address"

	db, cleanup := setupTestDB(t)
	defer cleanup()

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
		require.NoError(t, db.Create(&channel).Error)
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
	require.NoError(t, db.Create(&otherChannel).Error)

	params := map[string]string{
		"participant": participantWallet,
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	rpcRequest := &RPCMessage{
		Req: &RPCData{
			RequestID: 123,
			Method:    "get_channels",
			Params:    []any{json.RawMessage(paramsJSON)},
			Timestamp: uint64(time.Now().Unix()),
		},
	}

	response, err := HandleGetChannels(rpcRequest, db)
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.Equal(t, "get_channels", response.Res.Method)
	assert.Equal(t, uint64(123), response.Res.RequestID)

	require.Len(t, response.Res.Params, 1, "Response should contain a slice of ChannelResponse")
	channelsSlice, ok := response.Res.Params[0].([]ChannelResponse)
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

	openStatusRequest := &RPCMessage{
		Req: &RPCData{
			RequestID: 456,
			Method:    "get_channels",
			Params:    []any{json.RawMessage(openStatusParamsJSON)},
			Timestamp: uint64(time.Now().Unix()),
		},
	}

	openStatusResponse, err := HandleGetChannels(openStatusRequest, db)
	require.NoError(t, err)
	require.NotNil(t, openStatusResponse)

	openChannels, ok := openStatusResponse.Res.Params[0].([]ChannelResponse)
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

	closedStatusRequest := &RPCMessage{
		Req: &RPCData{
			RequestID: 457,
			Method:    "get_channels",
			Params:    []any{json.RawMessage(closedStatusParamsJSON)},
			Timestamp: uint64(time.Now().Unix()),
		},
	}

	closedStatusResponse, err := HandleGetChannels(closedStatusRequest, db)
	require.NoError(t, err)
	require.NotNil(t, closedStatusResponse)

	closedChannels, ok := closedStatusResponse.Res.Params[0].([]ChannelResponse)
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

	joiningStatusRequest := &RPCMessage{
		Req: &RPCData{
			RequestID: 458,
			Method:    "get_channels",
			Params:    []any{json.RawMessage(joiningStatusParamsJSON)},
			Timestamp: uint64(time.Now().Unix()),
		},
	}

	joiningStatusResponse, err := HandleGetChannels(joiningStatusRequest, db)
	require.NoError(t, err)
	require.NotNil(t, joiningStatusResponse)

	joiningChannels, ok := joiningStatusResponse.Res.Params[0].([]ChannelResponse)
	require.True(t, ok, "Response parameter should be a slice of ChannelResponse")
	assert.Len(t, joiningChannels, 1, "Should return only 1 joining channel")
	assert.Equal(t, "0xChannel3", joiningChannels[0].ChannelID, "Should return the joining channel")
	assert.Equal(t, ChannelStatusJoining, joiningChannels[0].Status, "Status should be joining")

	// No participant parameter: return all 4 channels
	noParamReq := &RPCMessage{
		Req: &RPCData{
			RequestID: 789,
			Method:    "get_channels",
			Params:    []any{map[string]string{}},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{},
	}
	allChannelsResp, err := HandleGetChannels(noParamReq, db)
	require.NoError(t, err, "Should not return error when participant is not specified")
	require.NotNil(t, allChannelsResp)

	allChannels, ok := allChannelsResp.Res.Params[0].([]ChannelResponse)
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

	openStatusOnlyReq := &RPCMessage{
		Req: &RPCData{
			RequestID: 790,
			Method:    "get_channels",
			Params:    []any{json.RawMessage(openStatusOnlyParamsJSON)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{},
	}

	openChannelsResp, err := HandleGetChannels(openStatusOnlyReq, db)
	require.NoError(t, err)
	require.NotNil(t, openChannelsResp)

	openChannelsOnly, ok := openChannelsResp.Res.Params[0].([]ChannelResponse)
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

// TestHandleGetAssets tests the get assets handler functionality
func TestHandleGetAssets(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	testAssets := []Asset{
		{Token: "0xToken1", ChainID: 137, Symbol: "usdc", Decimals: 6},
		{Token: "0xToken2", ChainID: 137, Symbol: "weth", Decimals: 18},
		{Token: "0xToken3", ChainID: 42220, Symbol: "celo", Decimals: 18},
		{Token: "0xToken4", ChainID: 8453, Symbol: "usdbc", Decimals: 6},
	}

	for _, asset := range testAssets {
		require.NoError(t, db.Create(&asset).Error)
	}

	// Case 1: Get all
	rpcRequest1 := &RPCMessage{
		Req: &RPCData{
			RequestID: 1,
			Method:    "get_assets",
			Params:    []any{},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp1, err := HandleGetAssets(rpcRequest1, db)
	require.NoError(t, err)
	assert.NotNil(t, resp1)

	assert.Equal(t, "get_assets", resp1.Res.Method)
	assert.Equal(t, uint64(1), resp1.Res.RequestID)
	require.Len(t, resp1.Res.Params, 1, "Response should contain an array of AssetResponse")

	assets1, ok := resp1.Res.Params[0].([]AssetResponse)
	require.True(t, ok, "Response parameter should be a slice of AssetResponse")
	assert.Len(t, assets1, 4, "Should return all 4 assets")

	foundSymbols := make(map[string]bool)
	for _, asset := range assets1 {
		foundSymbols[asset.Symbol] = true
		var orig Asset
		for _, a := range testAssets {
			if a.Symbol == asset.Symbol && a.ChainID == asset.ChainID {
				orig = a
				break
			}
		}
		assert.Equal(t, orig.Token, asset.Token, "Token should match")
		assert.Equal(t, orig.ChainID, asset.ChainID, "ChainID should match")
		assert.Equal(t, orig.Decimals, asset.Decimals, "Decimals should match")
	}
	assert.Len(t, foundSymbols, 4)
	assert.True(t, foundSymbols["usdc"])
	assert.True(t, foundSymbols["weth"])
	assert.True(t, foundSymbols["celo"])
	assert.True(t, foundSymbols["usdbc"])

	// Case 2: Filter by chain_id=137
	params2 := map[string]interface{}{"chain_id": float64(137)}
	paramsJSON2, err := json.Marshal(params2)
	require.NoError(t, err)

	rpcRequest2 := &RPCMessage{
		Req: &RPCData{
			RequestID: 2,
			Method:    "get_assets",
			Params:    []any{json.RawMessage(paramsJSON2)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp2, err := HandleGetAssets(rpcRequest2, db)
	require.NoError(t, err)
	assert.NotNil(t, resp2)

	assert.Equal(t, "get_assets", resp2.Res.Method)
	assert.Equal(t, uint64(2), resp2.Res.RequestID)

	assets2, ok := resp2.Res.Params[0].([]AssetResponse)
	require.True(t, ok, "Response parameter should be a slice of AssetResponse")
	assert.Len(t, assets2, 2, "Should return 2 Polygon assets")

	symbols2 := make(map[string]bool)
	for _, asset := range assets2 {
		assert.Equal(t, uint32(137), asset.ChainID, "ChainID should be Polygon")
		symbols2[asset.Symbol] = true
	}
	assert.Len(t, symbols2, 2)
	assert.True(t, symbols2["usdc"])
	assert.True(t, symbols2["weth"])

	// Case 3: Filter by chain_id=42220
	params3 := map[string]interface{}{"chain_id": float64(42220)}
	paramsJSON3, err := json.Marshal(params3)
	require.NoError(t, err)

	rpcRequest3 := &RPCMessage{
		Req: &RPCData{
			RequestID: 3,
			Method:    "get_assets",
			Params:    []any{json.RawMessage(paramsJSON3)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp3, err := HandleGetAssets(rpcRequest3, db)
	require.NoError(t, err)
	assert.NotNil(t, resp3)

	assets3, ok := resp3.Res.Params[0].([]AssetResponse)
	require.True(t, ok, "Response parameter should be a slice of AssetResponse")
	assert.Len(t, assets3, 1, "Should return 1 Celo asset")
	assert.Equal(t, "celo", assets3[0].Symbol)
	assert.Equal(t, uint32(42220), assets3[0].ChainID)

	// Case 4: Filter by non-existent chain_id=1
	params4 := map[string]interface{}{"chain_id": float64(1)}
	paramsJSON4, err := json.Marshal(params4)
	require.NoError(t, err)

	rpcRequest4 := &RPCMessage{
		Req: &RPCData{
			RequestID: 4,
			Method:    "get_assets",
			Params:    []any{json.RawMessage(paramsJSON4)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp4, err := HandleGetAssets(rpcRequest4, db)
	require.NoError(t, err)
	assert.NotNil(t, resp4)

	assets4, ok := resp4.Res.Params[0].([]AssetResponse)
	require.True(t, ok, "Response parameter should be a slice of AssetResponse")
	assert.Len(t, assets4, 0, "Should return 0 assets for chain_id=1")
}

// TestHandleGetAppSessions tests the get app sessions handler functionality
func TestHandleGetAppSessions(t *testing.T) {
	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: rawKey}
	participantAddr := signer.GetAddress().Hex()

	db, cleanup := setupTestDB(t)
	defer cleanup()

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
		require.NoError(t, db.Create(&session).Error)
	}

	// Case 1: Get all for participant
	params1 := map[string]string{"participant": participantAddr}
	paramsJSON1, err := json.Marshal(params1)
	require.NoError(t, err)

	rpcRequest1 := &RPCMessage{
		Req: &RPCData{
			RequestID: 1,
			Method:    "get_app_sessions",
			Params:    []any{json.RawMessage(paramsJSON1)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp1, err := HandleGetAppSessions(rpcRequest1, NewAppSessionService(db))
	require.NoError(t, err)
	assert.NotNil(t, resp1)

	assert.Equal(t, "get_app_sessions", resp1.Res.Method)
	assert.Equal(t, uint64(1), resp1.Res.RequestID)
	require.Len(t, resp1.Res.Params, 1, "Response should contain an array of AppSessionResponse")

	sessionResponses, ok := resp1.Res.Params[0].([]AppSessionResponse)
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

	rpcRequest2 := &RPCMessage{
		Req: &RPCData{
			RequestID: 2,
			Method:    "get_app_sessions",
			Params:    []any{json.RawMessage(paramsJSON2)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp2, err := HandleGetAppSessions(rpcRequest2, NewAppSessionService(db))
	require.NoError(t, err)
	assert.NotNil(t, resp2)

	sessionResponses2, ok := resp2.Res.Params[0].([]AppSessionResponse)
	require.True(t, ok, "Response parameter should be a slice of AppSessionResponse")
	assert.Len(t, sessionResponses2, 1, "Should return 1 open app session")
	assert.Equal(t, "0xSession1", sessionResponses2[0].AppSessionID, "Should be Session1")
	assert.Equal(t, string(ChannelStatusOpen), sessionResponses2[0].Status)

	// Case 3: No participant (all sessions)
	rpcRequest3 := &RPCMessage{
		Req: &RPCData{
			RequestID: 3,
			Method:    "get_app_sessions",
			Params:    []any{json.RawMessage(`{}`)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp3, err := HandleGetAppSessions(rpcRequest3, NewAppSessionService(db))
	require.NoError(t, err)
	require.NotNil(t, resp3)

	allSessions, ok := resp3.Res.Params[0].([]AppSessionResponse)
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

	openStatusRequest := &RPCMessage{
		Req: &RPCData{
			RequestID: 4,
			Method:    "get_app_sessions",
			Params:    []any{json.RawMessage(openStatusParamsJSON)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	openStatusResponse, err := HandleGetAppSessions(openStatusRequest, NewAppSessionService(db))
	require.NoError(t, err)
	require.NotNil(t, openStatusResponse)

	openSessions, ok := openStatusResponse.Res.Params[0].([]AppSessionResponse)
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

func TestHandleGetRPCHistory(t *testing.T) {
	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: rawKey}
	participantAddr := signer.GetAddress().Hex()

	db, cleanup := setupTestDB(t)
	defer cleanup()

	rpcStore := NewRPCStore(db)
	timestamp := uint64(time.Now().Unix())

	records := []RPCRecord{
		{
			Sender:    participantAddr,
			ReqID:     1,
			Method:    "ping",
			Params:    []byte(`[null]`),
			Timestamp: timestamp - 3600,
			ReqSig:    []string{"sig1"},
			Response:  []byte(`{"res":[1,"pong",[],1621234567890]}`),
			ResSig:    []string{},
		},
		{
			Sender:    participantAddr,
			ReqID:     2,
			Method:    "get_config",
			Params:    []byte(`[]`),
			Timestamp: timestamp - 1800,
			ReqSig:    []string{"sig2"},
			Response:  []byte(`{"res":[2,"get_config",[{"broker_address":"0xBroker"}],1621234597890]}`),
			ResSig:    []string{},
		},
		{
			Sender:    participantAddr,
			ReqID:     3,
			Method:    "get_channels",
			Params:    []byte(fmt.Sprintf(`[{"participant":"%s"}]`, participantAddr)),
			Timestamp: timestamp - 900,
			ReqSig:    []string{"sig3"},
			Response:  []byte(`{"res":[3,"get_channels",[[]],1621234627890]}`),
			ResSig:    []string{},
		},
	}

	for _, record := range records {
		require.NoError(t, db.Create(&record).Error)
	}

	otherRecord := RPCRecord{
		Sender:    "0xOtherParticipant",
		ReqID:     4,
		Method:    "ping",
		Params:    []byte(`[null]`),
		Timestamp: timestamp,
		ReqSig:    []string{"sig4"},
		Response:  []byte(`{"res":[4,"pong",[],1621234657890]}`),
		ResSig:    []string{},
	}
	require.NoError(t, db.Create(&otherRecord).Error)

	rpcRequest := &RPCMessage{
		Req: &RPCData{
			RequestID: 100,
			Method:    "get_rpc_history",
			Params:    []any{},
			Timestamp: timestamp,
		},
	}

	// Set ReqRaw so it’s available—though this handler doesn’t perform signature validation
	rawReq, err := json.Marshal(rpcRequest.Req)
	require.NoError(t, err)
	rpcRequest.Req.rawBytes = rawReq

	signed, err := signer.Sign(rawReq)
	require.NoError(t, err)
	rpcRequest.Sig = []string{hexutil.Encode(signed)}

	response, err := HandleGetRPCHistory(&Policy{Wallet: participantAddr}, rpcRequest, rpcStore)
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.Equal(t, "get_rpc_history", response.Res.Method)
	assert.Equal(t, uint64(100), response.Res.RequestID)

	require.Len(t, response.Res.Params, 1, "Response should contain RPCEntry entries")
	rpcHistory, ok := response.Res.Params[0].([]RPCEntry)
	require.True(t, ok, "Response parameter should be a slice of RPCEntry")
	assert.Len(t, rpcHistory, 3, "Should return 3 records for the participant")

	assert.Equal(t, uint64(3), rpcHistory[0].ReqID, "First record should be the newest")
	assert.Equal(t, uint64(2), rpcHistory[1].ReqID, "Second record should be the middle one")
	assert.Equal(t, uint64(1), rpcHistory[2].ReqID, "Third record should be the oldest")

	missingParamReq := &RPCMessage{
		Req: &RPCData{
			RequestID: 789,
			Method:    "get_rpc_history",
			Params:    []any{},
			Timestamp: uint64(time.Now().Unix()),
		},
	}
	// Also set ReqRaw to avoid nil pointer, though policy is empty
	rawReq2, err := json.Marshal(missingParamReq.Req)
	require.NoError(t, err)
	missingParamReq.Req.rawBytes = rawReq2
	missingParamReq.Sig = []string{hexutil.Encode(signed)}

	_, err = HandleGetRPCHistory(&Policy{}, missingParamReq, rpcStore)
	assert.Error(t, err, "Should return error with missing participant")
	assert.Contains(t, err.Error(), "missing participant", "Error should mention missing participant")
}

// TestHandleGetLedgerEntries tests the get ledger entries handler functionality
func TestHandleGetLedgerEntries(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	participant1 := "0xParticipant1"
	participant2 := "0xParticipant2"

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
		err := ledger1.Record(participant1, data.asset, data.amount)
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
		err := ledger2.Record(participant2, data.asset, data.amount)
		require.NoError(t, err)
	}

	// Case 1: Filter by account_id only
	params1 := map[string]string{"account_id": participant1}
	paramsJSON1, err := json.Marshal(params1)
	require.NoError(t, err)

	rpcRequest1 := &RPCMessage{
		Req: &RPCData{
			RequestID: 1,
			Method:    "get_ledger_entries",
			Params:    []any{json.RawMessage(paramsJSON1)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp1, err := HandleGetLedgerEntries(rpcRequest1, "", db)
	require.NoError(t, err)
	assert.NotNil(t, resp1)

	assert.Equal(t, "get_ledger_entries", resp1.Res.Method)
	assert.Equal(t, uint64(1), resp1.Res.RequestID)
	require.Len(t, resp1.Res.Params, 1, "Response should contain an array of Entry objects")

	entries1, ok := resp1.Res.Params[0].([]LedgerEntryResponse)
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

	rpcRequest2 := &RPCMessage{
		Req: &RPCData{
			RequestID: 2,
			Method:    "get_ledger_entries",
			Params:    []any{json.RawMessage(paramsJSON2)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp2, err := HandleGetLedgerEntries(rpcRequest2, "", db)
	require.NoError(t, err)
	assert.NotNil(t, resp2)

	entries2, ok := resp2.Res.Params[0].([]LedgerEntryResponse)
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

	rpcRequest3 := &RPCMessage{
		Req: &RPCData{
			RequestID: 3,
			Method:    "get_ledger_entries",
			Params:    []any{json.RawMessage(paramsJSON3)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp3, err := HandleGetLedgerEntries(rpcRequest3, "", db)
	require.NoError(t, err)
	assert.NotNil(t, resp3)

	entries3, ok := resp3.Res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")
	assert.Len(t, entries3, 2, "Should return all 2 entries for participant2")

	for _, entry := range entries3 {
		assert.Equal(t, participant2, entry.Participant)
	}

	// Case 4: Filter by wallet and asset
	params4 := map[string]string{"wallet": participant2, "asset": "usdc"}
	paramsJSON4, err := json.Marshal(params4)
	require.NoError(t, err)

	rpcRequest4 := &RPCMessage{
		Req: &RPCData{
			RequestID: 4,
			Method:    "get_ledger_entries",
			Params:    []any{json.RawMessage(paramsJSON4)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp4, err := HandleGetLedgerEntries(rpcRequest4, "", db)
	require.NoError(t, err)
	assert.NotNil(t, resp4)

	entries4, ok := resp4.Res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")
	assert.Len(t, entries4, 1, "Should return 1 entry for participant2 with usdc")
	assert.Equal(t, "usdc", entries4[0].Asset)
	assert.Equal(t, participant2, entries4[0].Participant)

	// Case 5: Filter by account_id and wallet (no overlap)
	params5 := map[string]string{"account_id": participant1, "wallet": participant2}
	paramsJSON5, err := json.Marshal(params5)
	require.NoError(t, err)

	rpcRequest5 := &RPCMessage{
		Req: &RPCData{
			RequestID: 5,
			Method:    "get_ledger_entries",
			Params:    []any{json.RawMessage(paramsJSON5)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp5, err := HandleGetLedgerEntries(rpcRequest5, "", db)
	require.NoError(t, err)
	assert.NotNil(t, resp5)

	entries5, ok := resp5.Res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")
	assert.Len(t, entries5, 0, "Should return 0 entries when account_id and wallet don't match")

	// Case 6: No filters (all entries)
	rpcRequest6 := &RPCMessage{
		Req: &RPCData{
			RequestID: 6,
			Method:    "get_ledger_entries",
			Params:    []any{map[string]string{}}, // Empty map
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp6, err := HandleGetLedgerEntries(rpcRequest6, "", db)
	require.NoError(t, err)
	assert.NotNil(t, resp6)

	entries6, ok := resp6.Res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")
	assert.Len(t, entries6, 7, "Should return all 7 entries")

	foundParticipants := make(map[string]bool)
	for _, entry := range entries6 {
		foundParticipants[entry.Participant] = true
	}
	assert.True(t, foundParticipants[participant1], "Should include entries for participant1")
	assert.True(t, foundParticipants[participant2], "Should include entries for participant2")

	// Case 7: Default wallet provided
	rpcRequest7 := &RPCMessage{
		Req: &RPCData{
			RequestID: 7,
			Method:    "get_ledger_entries",
			Params:    []any{map[string]string{}},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	resp7, err := HandleGetLedgerEntries(rpcRequest7, participant1, db)
	require.NoError(t, err)
	assert.NotNil(t, resp7)

	entries7, ok := resp7.Res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")
	assert.Len(t, entries7, 5, "Should return 5 entries for default wallet participant1")

	for _, entry := range entries7 {
		assert.Equal(t, participant1, entry.Participant)
	}
}

// TestAssetsForWebSocketConnection tests that assets can be fetched for WebSocket connection
func TestAssetsForWebSocketConnection(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	testAssets := []Asset{
		{Token: "0xToken1", ChainID: 137, Symbol: "usdc", Decimals: 6},
		{Token: "0xToken2", ChainID: 42220, Symbol: "celo", Decimals: 18},
	}

	for _, a := range testAssets {
		require.NoError(t, db.Create(&a).Error)
	}

	assets, err := GetAllAssets(db, nil)
	require.NoError(t, err)
	assert.Len(t, assets, 2, "Should have 2 assets in database")

	foundSymbols := make(map[string]bool)
	for _, asset := range assets {
		foundSymbols[asset.Symbol] = true
		assert.NotEmpty(t, asset.Token, "Token should not be empty")
		assert.NotZero(t, asset.ChainID, "ChainID should not be zero")
		assert.NotEmpty(t, asset.Symbol, "Symbol should not be empty")
		assert.NotZero(t, asset.Decimals, "Decimals should not be zero")
	}
	assert.True(t, foundSymbols["usdc"], "Should include USDC")
	assert.True(t, foundSymbols["celo"], "Should include CELO")
}

func TestHandleCreateAppSession(t *testing.T) {
	rawA, _ := crypto.GenerateKey()
	rawB, _ := crypto.GenerateKey()
	signerA := Signer{privateKey: rawA}
	signerB := Signer{privateKey: rawB}
	addrA := signerA.GetAddress().Hex()
	addrB := signerB.GetAddress().Hex()
	t.Run("SuccessfulCreateAppSession", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		token := "0xTokenXYZ"
		for i, p := range []string{addrA, addrB} {
			ch := &Channel{
				ChannelID:   fmt.Sprintf("0xChannel%c", 'A'+i),
				Wallet:      p,
				Participant: p,
				Status:      ChannelStatusOpen,
				Token:       token,
				Nonce:       1,
			}
			require.NoError(t, db.Create(ch).Error)
			require.NoError(t, db.Create(&SignerWallet{
				Signer: p, Wallet: p,
			}).Error)
		}

		require.NoError(t, GetWalletLedger(db, addrA).Record(addrA, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, addrB).Record(addrB, "usdc", decimal.NewFromInt(200)))

		ts := uint64(time.Now().Unix())
		def := AppDefinition{
			Protocol:           "test-proto",
			ParticipantWallets: []string{addrA, addrB},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Challenge:          60,
			Nonce:              ts,
		}
		createParams := CreateAppSessionParams{
			Definition: def,
			Allocations: []AppAllocation{
				{ParticipantWallet: addrA, AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: addrB, AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
		}

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 42,
				Method:    "create_app_session",
				Params:    []any{createParams},
				Timestamp: ts,
			},
		}

		// 1) Marshal rpcReq.Req exactly as a JSON array
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq

		// 2) Sign rawReq with both participants
		sigA, err := signerA.Sign(rawReq)
		require.NoError(t, err)
		sigB, err := signerB.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sigA), hexutil.Encode(sigB)}

		appSesionService := NewAppSessionService(db)
		// Call handler
		resp, err := HandleCreateApplication(nil, rpcReq, appSesionService)
		require.NoError(t, err)

		assert.Equal(t, "create_app_session", resp.Res.Method)
		appResp, ok := resp.Res.Params[0].(*AppSessionResponse)
		require.True(t, ok)
		assert.Equal(t, string(ChannelStatusOpen), appResp.Status)
		assert.Equal(t, uint64(1), appResp.Version)

		var vApp AppSession
		require.NoError(t, db.Where("session_id = ?", appResp.AppSessionID).First(&vApp).Error)
		assert.ElementsMatch(t, []string{addrA, addrB}, vApp.ParticipantWallets)
		assert.Equal(t, uint64(1), vApp.Version)

		// Participant accounts drained
		partBalA, _ := GetWalletLedger(db, addrA).Balance(addrA, "usdc")
		partBalB, _ := GetWalletLedger(db, addrB).Balance(addrB, "usdc")
		assert.True(t, partBalA.IsZero(), "Participant A balance should be zero")
		assert.True(t, partBalB.IsZero(), "Participant B balance should be zero")

		// Virtual-app funded
		vBalA, _ := GetWalletLedger(db, addrA).Balance(appResp.AppSessionID, "usdc")
		vBalB, _ := GetWalletLedger(db, addrB).Balance(appResp.AppSessionID, "usdc")
		assert.Equal(t, decimal.NewFromInt(100).String(), vBalA.String())
		assert.Equal(t, decimal.NewFromInt(200).String(), vBalB.String())
	})
	t.Run("ErrorChallengedChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		token := "0xTokenXYZ"
		for i, p := range []string{addrA, addrB} {
			ch := &Channel{
				ChannelID:   fmt.Sprintf("0xChannel%c", 'A'+i),
				Wallet:      p,
				Participant: p,
				Status:      ChannelStatusChallenged,
				Token:       token,
				Nonce:       1,
			}
			require.NoError(t, db.Create(ch).Error)
			require.NoError(t, db.Create(&SignerWallet{
				Signer: p, Wallet: p,
			}).Error)
		}

		require.NoError(t, GetWalletLedger(db, addrA).Record(addrA, "usdc", decimal.NewFromInt(100)))
		require.NoError(t, GetWalletLedger(db, addrB).Record(addrB, "usdc", decimal.NewFromInt(200)))

		ts := uint64(time.Now().Unix())
		def := AppDefinition{
			Protocol:           "test-proto",
			ParticipantWallets: []string{addrA, addrB},
			Weights:            []int64{1, 1},
			Quorum:             2,
			Challenge:          60,
			Nonce:              ts,
		}
		createParams := CreateAppSessionParams{
			Definition: def,
			Allocations: []AppAllocation{
				{ParticipantWallet: addrA, AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
				{ParticipantWallet: addrB, AssetSymbol: "usdc", Amount: decimal.NewFromInt(200)},
			},
		}

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 42,
				Method:    "create_app_session",
				Params:    []any{createParams},
				Timestamp: ts,
			},
		}

		// 1) Marshal rpcReq.Req exactly as a JSON array
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq

		// 2) Sign rawReq with both participants
		sigA, err := signerA.Sign(rawReq)
		require.NoError(t, err)
		sigB, err := signerB.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sigA), hexutil.Encode(sigB)}

		// Call handler
		_, err = HandleCreateApplication(nil, rpcReq, NewAppSessionService(db))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "has challenged channels")
	})
}

// TestHandleSubmitState tests the submit state into a virtual app handler functionality
func TestHandleSubmitState(t *testing.T) {
	t.Run("SuccessfulSubmitState", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		participantA := signer.GetAddress().Hex()
		participantB := "0xParticipantB"

		db, cleanup := setupTestDB(t)
		defer cleanup()

		tokenAddress := "0xToken123"
		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChannelA",
			Participant: participantA,
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			Nonce:       1,
		}).Error)
		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChannelB",
			Participant: participantB,
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			Nonce:       1,
		}).Error)

		vAppID := "0xVApp123"
		require.NoError(t, db.Create(&AppSession{
			SessionID:          vAppID,
			ParticipantWallets: []string{participantA, participantB},
			Status:             ChannelStatusOpen,
			Challenge:          60,
			Weights:            []int64{100, 0},
			Quorum:             100,
			Version:            1,
		}).Error)

		assetSymbol := "usdc"
		require.NoError(t, GetWalletLedger(db, participantA).Record(vAppID, assetSymbol, decimal.NewFromInt(200)))
		require.NoError(t, GetWalletLedger(db, participantB).Record(vAppID, assetSymbol, decimal.NewFromInt(300)))

		submitStateParams := SubmitStateParams{
			AppSessionID: vAppID,
			Allocations: []AppAllocation{
				{ParticipantWallet: participantA, AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
				{ParticipantWallet: participantB, AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
			},
		}

		// Create RPC request
		paramsJSON, _ := json.Marshal(submitStateParams)
		req := &RPCMessage{
			Req: &RPCData{
				RequestID: 1,
				Method:    "submit_state",
				Params:    []any{json.RawMessage(paramsJSON)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		// 1) Marshal rpc.Req to get the exact raw bytes of [request_id, method, params, timestamp]
		rawReq, err := json.Marshal(req.Req)
		require.NoError(t, err)
		req.Req.rawBytes = rawReq

		// 2) Sign rawReq directly
		sigBytes, err := signer.Sign(rawReq)
		require.NoError(t, err)
		req.Sig = []string{hexutil.Encode(sigBytes)}

		// Call handler
		resp, err := HandleSubmitState(nil, req, NewAppSessionService(db))
		require.NoError(t, err)
		assert.Equal(t, "submit_state", resp.Res.Method)
		appResp, ok := resp.Res.Params[0].(*AppSessionResponse)
		require.True(t, ok)
		assert.Equal(t, string(ChannelStatusOpen), appResp.Status)
		assert.Equal(t, uint64(2), appResp.Version)

		var updated AppSession
		require.NoError(t, db.Where("session_id = ?", vAppID).First(&updated).Error)
		assert.Equal(t, ChannelStatusOpen, updated.Status)
		assert.Equal(t, uint64(2), updated.Version)

		// Check balances redistributed
		balA, _ := GetWalletLedger(db, participantA).Balance(vAppID, "usdc")
		balB, _ := GetWalletLedger(db, participantB).Balance(vAppID, "usdc")
		assert.Equal(t, decimal.NewFromInt(250), balA)
		assert.Equal(t, decimal.NewFromInt(250), balB)
	})
}

// TestHandleCloseVirtualApp tests the close virtual app handler functionality
func TestHandleCloseVirtualApp(t *testing.T) {
	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: rawKey}
	participantA := signer.GetAddress().Hex()
	participantB := "0xParticipantB"

	db, cleanup := setupTestDB(t)
	defer cleanup()

	tokenAddress := "0xToken123"
	require.NoError(t, db.Create(&Channel{
		ChannelID:   "0xChannelA",
		Participant: participantA,
		Status:      ChannelStatusOpen,
		Token:       tokenAddress,
		Nonce:       1,
	}).Error)
	require.NoError(t, db.Create(&Channel{
		ChannelID:   "0xChannelB",
		Participant: participantB,
		Status:      ChannelStatusOpen,
		Token:       tokenAddress,
		Nonce:       1,
	}).Error)

	vAppID := "0xVApp123"
	require.NoError(t, db.Create(&AppSession{
		SessionID:          vAppID,
		ParticipantWallets: []string{participantA, participantB},
		Status:             ChannelStatusOpen,
		Challenge:          60,
		Weights:            []int64{100, 0},
		Quorum:             100,
		Version:            2,
	}).Error)

	assetSymbol := "usdc"
	require.NoError(t, GetWalletLedger(db, participantA).Record(vAppID, assetSymbol, decimal.NewFromInt(200)))
	require.NoError(t, GetWalletLedger(db, participantB).Record(vAppID, assetSymbol, decimal.NewFromInt(300)))

	closeParams := CloseAppSessionParams{
		AppSessionID: vAppID,
		Allocations: []AppAllocation{
			{ParticipantWallet: participantA, AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
			{ParticipantWallet: participantB, AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
		},
	}

	// Create RPC request
	paramsJSON, _ := json.Marshal(closeParams)
	req := &RPCMessage{
		Req: &RPCData{
			RequestID: 1,
			Method:    "close_app_session",
			Params:    []any{json.RawMessage(paramsJSON)},
			Timestamp: uint64(time.Now().Unix()),
		},
	}

	// 1) Marshal rpc.Req to get the exact raw bytes of [request_id, method, params, timestamp]
	rawReq, err := json.Marshal(req.Req)
	require.NoError(t, err)
	req.Req.rawBytes = rawReq

	// 2) Sign rawReq directly
	sigBytes, err := signer.Sign(rawReq)
	require.NoError(t, err)
	req.Sig = []string{hexutil.Encode(sigBytes)}

	// Call handler
	appSesionService := NewAppSessionService(db)
	resp, err := HandleCloseApplication(nil, req, appSesionService)
	require.NoError(t, err)
	assert.Equal(t, "close_app_session", resp.Res.Method)
	appResp, ok := resp.Res.Params[0].(*AppSessionResponse)
	require.True(t, ok)
	assert.Equal(t, string(ChannelStatusClosed), appResp.Status)
	assert.Equal(t, uint64(3), appResp.Version)

	var updated AppSession
	require.NoError(t, db.Where("session_id = ?", vAppID).First(&updated).Error)
	assert.Equal(t, ChannelStatusClosed, updated.Status)
	assert.Equal(t, uint64(3), updated.Version)

	// Check balances redistributed
	balA, _ := GetWalletLedger(db, participantA).Balance(participantA, "usdc")
	balB, _ := GetWalletLedger(db, participantB).Balance(participantB, "usdc")
	assert.Equal(t, decimal.NewFromInt(250), balA)
	assert.Equal(t, decimal.NewFromInt(250), balB)

	// v-app accounts drained
	vBalA, _ := GetWalletLedger(db, participantA).Balance(vAppID, "usdc")
	vBalB, _ := GetWalletLedger(db, participantB).Balance(vAppID, "usdc")
	assert.True(t, vBalA.IsZero(), "Participant A vApp balance should be zero")
	assert.True(t, vBalB.IsZero(), "Participant B vApp balance should be zero")
}

// TestHandleResizeChannel tests the resize channel handler functionality
func TestHandleTransfer(t *testing.T) {
	// Create signers
	senderKey, _ := crypto.GenerateKey()
	senderSigner := Signer{privateKey: senderKey}
	senderAddr := senderSigner.GetAddress().Hex()
	recipientAddr := "0x" + strings.Repeat("1", 40) // Valid ethereum address with 1s

	t.Run("SuccessfulTransfer", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(1000)))
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "eth", decimal.NewFromInt(5)))

		// Create transfer parameters
		transferParams := Transfer{
			Destination: recipientAddr,
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
				{AssetSymbol: "eth", Amount: decimal.NewFromInt(2)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 42,
				Method:    "transfer",
				Params:    []any{transferParams},
				Timestamp: ts,
			},
		}

		// Marshal and sign the request
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq

		signed, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(signed)}

		// Call handler
		policy := &Policy{Wallet: senderAddr}
		resp, err := HandleTransfer(policy, rpcReq, db)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify response
		assert.Equal(t, "transfer", resp.Res.Method)
		assert.Equal(t, uint64(42), resp.Res.RequestID)
		// Verify response structure
		transferResp, ok := resp.Res.Params[0].(*TransferResponse)
		require.True(t, ok, "Response should be a TransferResponse")
		assert.Equal(t, senderAddr, transferResp.From)
		assert.Equal(t, recipientAddr, transferResp.To)
		assert.False(t, transferResp.CreatedAt.IsZero(), "CreatedAt should be set")

		// Check balances were updated correctly
		// Sender should have 500 USDC and 3 ETH left
		senderUSDC, err := GetWalletLedger(db, senderAddr).Balance(senderAddr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(500).String(), senderUSDC.String())

		senderETH, err := GetWalletLedger(db, senderAddr).Balance(senderAddr, "eth")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(3).String(), senderETH.String())

		// Recipient should have 500 USDC and 2 ETH
		recipientUSDC, err := GetWalletLedger(db, recipientAddr).Balance(recipientAddr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(500).String(), recipientUSDC.String())

		recipientETH, err := GetWalletLedger(db, recipientAddr).Balance(recipientAddr, "eth")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(2).String(), recipientETH.String())
	})

	t.Run("ErrorInvalidDestinationAddress", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(1000)))

		// Create transfer with invalid destination
		transferParams := Transfer{
			Destination: "not-a-valid-address",
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 43,
				Method:    "transfer",
				Params:    []any{transferParams},
				Timestamp: ts,
			},
		}

		// Marshal and sign the request
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq

		signed, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(signed)}

		// Call handler
		policy := &Policy{Wallet: senderAddr}
		_, err = HandleTransfer(policy, rpcReq, db)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid destination account")
	})

	t.Run("ErrorTransferToSelf", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(1000)))

		// Create transfer to self
		transferParams := Transfer{
			Destination: senderAddr,
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 44,
				Method:    "transfer",
				Params:    []any{transferParams},
				Timestamp: ts,
			},
		}

		// Marshal and sign the request
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq

		signed, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(signed)}

		// Call handler
		policy := &Policy{Wallet: senderAddr}
		_, err = HandleTransfer(policy, rpcReq, db)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid destination")
	})

	t.Run("ErrorInsufficientFunds", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account with a small amount
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(100)))

		// Create transfer for more than available
		transferParams := Transfer{
			Destination: recipientAddr,
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 45,
				Method:    "transfer",
				Params:    []any{transferParams},
				Timestamp: ts,
			},
		}

		// Marshal and sign the request
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq

		signed, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(signed)}

		// Call handler
		policy := &Policy{Wallet: senderAddr}
		_, err = HandleTransfer(policy, rpcReq, db)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
	})

	t.Run("ErrorEmptyAllocations", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Create transfer with empty allocations
		transferParams := Transfer{
			Destination: recipientAddr,
			Allocations: []TransferAllocation{},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 46,
				Method:    "transfer",
				Params:    []any{transferParams},
				Timestamp: ts,
			},
		}

		// Marshal and sign the request
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq

		signed, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(signed)}

		// Call handler
		policy := &Policy{Wallet: senderAddr}
		_, err = HandleTransfer(policy, rpcReq, db)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty allocations")
	})

	t.Run("ErrorZeroAmount", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(1000)))

		// Create transfer with zero amount
		transferParams := Transfer{
			Destination: recipientAddr,
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(0)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 49,
				Method:    "transfer",
				Params:    []any{transferParams},
				Timestamp: ts,
			},
		}

		// Marshal and sign the request
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq

		signed, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(signed)}

		// Call handler
		policy := &Policy{Wallet: senderAddr}
		_, err = HandleTransfer(policy, rpcReq, db)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid allocation")
	})

	t.Run("ErrorNegativeAmount", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(1000)))

		// Create transfer with negative amount
		transferParams := Transfer{
			Destination: recipientAddr,
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(-500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 47,
				Method:    "transfer",
				Params:    []any{transferParams},
				Timestamp: ts,
			},
		}

		// Marshal and sign the request
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq

		signed, err := senderSigner.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(signed)}

		// Call handler
		policy := &Policy{Wallet: senderAddr}
		_, err = HandleTransfer(policy, rpcReq, db)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid allocation")
	})

	t.Run("ErrorInvalidSignature", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Setup signer wallet relation
		require.NoError(t, db.Create(&SignerWallet{
			Signer: senderAddr, Wallet: senderAddr,
		}).Error)

		// Fund sender's account
		require.NoError(t, GetWalletLedger(db, senderAddr).Record(senderAddr, "usdc", decimal.NewFromInt(1000)))

		// Create transfer parameters
		transferParams := Transfer{
			Destination: recipientAddr,
			Allocations: []TransferAllocation{
				{AssetSymbol: "usdc", Amount: decimal.NewFromInt(500)},
			},
		}

		// Create RPC request
		ts := uint64(time.Now().Unix())
		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 48,
				Method:    "transfer",
				Params:    []any{transferParams},
				Timestamp: ts,
			},
		}

		// Marshal the request
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq

		// Sign with a different key
		wrongKey, _ := crypto.GenerateKey()
		wrongSigner := Signer{privateKey: wrongKey}
		signed, err := wrongSigner.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(signed)}

		// Call handler
		policy := &Policy{Wallet: senderAddr}
		_, err = HandleTransfer(policy, rpcReq, db)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature")
	})
}

func TestHandleResizeChannel(t *testing.T) {
	t.Run("SuccessfulAllocation", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		// Setup test DB
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Create asset
		asset := Asset{Token: "0xTokenResize", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with initial amount 1000
		initialAmount := uint64(1000)
		ch := Channel{
			ChannelID:   "0xChanResize",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger with 1500 USDC (enough for resize)
		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(1500)))

		// Verify initial balance
		initialBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(1500), initialBalance)

		// Prepare allocation params: allocate 200 to channel (does not change user's total balance)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(200),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 1,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		// Sign request
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		// Call handler
		resp, err := HandleResizeChannel(nil, rpcReq, db, &signer)
		require.NoError(t, err)

		// Validate response
		assert.Equal(t, "resize_channel", resp.Res.Method)
		resObj, ok := resp.Res.Params[0].(ResizeChannelResponse)
		require.True(t, ok, "Response should be ResizeChannelResponse")
		assert.Equal(t, ch.ChannelID, resObj.ChannelID)
		assert.Equal(t, ch.Version+1, resObj.Version)

		// New channel amount should be initial + 200
		expected := new(big.Int).Add(new(big.Int).SetUint64(initialAmount), big.NewInt(200))
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(expected), "Allocated amount mismatch")
		assert.Equal(t, 0, resObj.Allocations[1].Amount.Cmp(big.NewInt(0)), "Broker allocation should be zero")

		// Verify channel state in database remains unchanged (no update until blockchain confirmation)
		var unchangedChannel Channel
		require.NoError(t, db.Where("channel_id = ?", ch.ChannelID).First(&unchangedChannel).Error)
		assert.Equal(t, initialAmount, unchangedChannel.Amount) // Should remain unchanged
		assert.Equal(t, ch.Version, unchangedChannel.Version)   // Should remain unchanged
		assert.Equal(t, ChannelStatusOpen, unchangedChannel.Status)

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(1500), finalBalance) // Should remain unchanged
	})

	t.Run("SuccessfulDeallocation", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		db, cleanup := setupTestDB(t)
		defer cleanup()

		asset := Asset{Token: "0xTokenResize2", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		initialAmount := uint64(1000)
		ch := Channel{
			ChannelID:   "0xChanResize2",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(500)))

		// Prepare resize params: decrease by 300
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(-300),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 2,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		resp, err := HandleResizeChannel(nil, rpcReq, db, &signer)
		require.NoError(t, err)

		resObj, ok := resp.Res.Params[0].(ResizeChannelResponse)
		require.True(t, ok)

		// Channel amount should decrease
		expected := new(big.Int).Sub(new(big.Int).SetUint64(initialAmount), big.NewInt(300))
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(expected), "Decreased amount mismatch")

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(500), finalBalance) // Should remain unchanged
	})

	t.Run("ErrorInvalidChannelID", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		db, cleanup := setupTestDB(t)
		defer cleanup()

		resizeParams := ResizeChannelParams{
			ChannelID:        "0xNonExistentChannel",
			AllocateAmount:   big.NewInt(100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 3,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		_, err = HandleResizeChannel(nil, rpcReq, db, &signer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "channel with id 0xNonExistentChannel not found")
	})

	t.Run("ErrorChannelClosed", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		db, cleanup := setupTestDB(t)
		defer cleanup()

		asset := Asset{Token: "0xTokenClosed", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanClosed",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusClosed,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 4,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		_, err = HandleResizeChannel(nil, rpcReq, db, &signer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "channel 0xChanClosed must be open to resize, current status: closed")
	})

	t.Run("ErrorChannelJoining", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		db, cleanup := setupTestDB(t)
		defer cleanup()

		asset := Asset{Token: "0xTokenJoining", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanJoining",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusJoining,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 10,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		_, err = HandleResizeChannel(nil, rpcReq, db, &signer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "channel 0xChanJoining must be open to resize, current status: joining")
		assert.Contains(t, err.Error(), "joining")
	})

	t.Run("ErrorOtherChallengedChannel", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		db, cleanup := setupTestDB(t)
		defer cleanup()

		asset := Asset{Token: "0xToken", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChanChallenged",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusChallenged,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}).Error)

		ch := Channel{
			ChannelID:   "0xChan",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 10,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		_, err = HandleResizeChannel(nil, rpcReq, db, &signer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has challenged channels")
	})

	t.Run("ErrorInsufficientFunds", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		db, cleanup := setupTestDB(t)
		defer cleanup()

		asset := Asset{Token: "0xTokenInsufficient", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanInsufficient",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund with very small amount (0.000001 USDC), but try to allocate 200 raw units
		// This will create insufficient balance when converted to raw units
		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromFloat(0.000001)))

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(200),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 5,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		_, err = HandleResizeChannel(nil, rpcReq, db, &signer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient unified balance")
	})

	t.Run("ErrorZeroAmounts", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		db, cleanup := setupTestDB(t)
		defer cleanup()

		asset := Asset{Token: "0xTokenZero", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanZero",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(1500)))

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(0),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 6,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		// Zero allocation should now be rejected as it's a wasteful no-op operation
		_, err = HandleResizeChannel(nil, rpcReq, db, &signer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resize operation requires non-zero ResizeAmount or AllocateAmount")
	})

	t.Run("SuccessfulResizeDeposit", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		db, cleanup := setupTestDB(t)
		defer cleanup()

		asset := Asset{Token: "0xTokenResizeOnly", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		initialAmount := uint64(1000)
		ch := Channel{
			ChannelID:   "0xChanResizeOnly",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund the ledger to pass balance validation
		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(1500)))

		// Resize operation: deposit 100 into channel (changes user's total balance)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			ResizeAmount:     big.NewInt(100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 11,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		resp, err := HandleResizeChannel(nil, rpcReq, db, &signer)
		require.NoError(t, err)

		resObj, ok := resp.Res.Params[0].(ResizeChannelResponse)
		require.True(t, ok)

		// Should be initial amount (1000) + allocate amount (0) + resize amount (100) = 1100
		expected := new(big.Int).Add(new(big.Int).SetUint64(initialAmount), big.NewInt(100))
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(expected))
	})

	t.Run("SuccessfulResizeWithdrawal", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		db, cleanup := setupTestDB(t)
		defer cleanup()

		asset := Asset{Token: "0xTokenResizeOnly", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		initialAmount := uint64(1000)
		ch := Channel{
			ChannelID:   "0xChanResizeOnly",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund the ledger to pass balance validation
		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(1500)))

		// Resize operation: withdraw 100 from channel (changes user's total balance)
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			ResizeAmount:     big.NewInt(-100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 11,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		resp, err := HandleResizeChannel(nil, rpcReq, db, &signer)
		require.NoError(t, err)

		resObj, ok := resp.Res.Params[0].(ResizeChannelResponse)
		require.True(t, ok)

		// Should be initial amount (1000) + allocate amount (0) - resize amount (100) = 900
		expected := new(big.Int).Add(new(big.Int).SetUint64(initialAmount), big.NewInt(-100))
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(expected))
	})

	t.Run("ErrorExcessiveDeallocation", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		db, cleanup := setupTestDB(t)
		defer cleanup()

		asset := Asset{Token: "0xTokenExcessive", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanExcessive",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Try to decrease by more than channel amount
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(-1500),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 7,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		_, err = HandleResizeChannel(nil, rpcReq, db, &signer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "new channel amount must be positive")
	})

	t.Run("ErrorInvalidSignature", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		// Create a different signer for invalid signature
		wrongKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		wrongSigner := Signer{privateKey: wrongKey}

		db, cleanup := setupTestDB(t)
		defer cleanup()

		asset := Asset{Token: "0xTokenSig", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanSig",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 8,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq

		// Sign with wrong signer
		sig, err := wrongSigner.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		_, err = HandleResizeChannel(nil, rpcReq, db, &signer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature")
	})

	t.Run("BoundaryLargeAllocation", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		db, cleanup := setupTestDB(t)
		defer cleanup()

		asset := Asset{Token: "0xTokenLarge", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanLarge",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      1000,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund with a very large amount
		ledger := GetWalletLedger(db, addr)
		largeAmount := decimal.NewFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil), 0) // 10^18
		require.NoError(t, ledger.Record(addr, "usdc", largeAmount))

		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil), // 10^15
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 9,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		resp, err := HandleResizeChannel(nil, rpcReq, db, &signer)
		require.NoError(t, err)

		resObj, ok := resp.Res.Params[0].(ResizeChannelResponse)
		require.True(t, ok)

		// Verify the large allocation was processed correctly
		expectedAmount := new(big.Int).Add(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil))
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(expectedAmount))
	})

	t.Run("SuccessfulAllocationWithResizeDeposit", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		// Setup test DB
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Create asset
		asset := Asset{Token: "0xTokenMixed", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with initial amount 1000
		initialAmount := uint64(1000)
		ch := Channel{
			ChannelID:   "0xChanMixed",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger with 2000 USDC (enough for both operations)
		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(2000)))

		// Verify initial balance
		initialBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(2000), initialBalance)

		// Combined operation: allocate 150 to channel + resize (deposit) 100 more
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(150), // Allocation: moves funds from user balance to channel
			ResizeAmount:     big.NewInt(100), // Resize: deposits additional funds into channel
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 12,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		// Sign request
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		// Call handler
		resp, err := HandleResizeChannel(nil, rpcReq, db, &signer)
		require.NoError(t, err)

		// Validate response
		assert.Equal(t, "resize_channel", resp.Res.Method)
		resObj, ok := resp.Res.Params[0].(ResizeChannelResponse)
		require.True(t, ok, "Response should be ResizeChannelResponse")
		assert.Equal(t, ch.ChannelID, resObj.ChannelID)
		assert.Equal(t, ch.Version+1, resObj.Version)

		// New channel amount should be initial + AllocateAmount + ResizeAmount = 1000 + 150 + 100 = 1250
		expected := new(big.Int).Add(new(big.Int).SetUint64(initialAmount), big.NewInt(250))
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(expected), "Combined allocation+resize amount mismatch")
		assert.Equal(t, 0, resObj.Allocations[1].Amount.Cmp(big.NewInt(0)), "Broker allocation should be zero")

		// Verify channel state in database remains unchanged (no update until blockchain confirmation)
		var unchangedChannel Channel
		require.NoError(t, db.Where("channel_id = ?", ch.ChannelID).First(&unchangedChannel).Error)
		assert.Equal(t, initialAmount, unchangedChannel.Amount) // Should remain unchanged
		assert.Equal(t, ch.Version, unchangedChannel.Version)   // Should remain unchanged
		assert.Equal(t, ChannelStatusOpen, unchangedChannel.Status)

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(2000), finalBalance) // Should remain unchanged
	})

	t.Run("SuccessfulAllocationWithResizeWithdrawal", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		// Setup test DB
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Create asset
		asset := Asset{Token: "0xTokenMixed", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with initial amount 0
		initialAmount := uint64(0)
		ch := Channel{
			ChannelID:   "0xChanMixed",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger with 2000 USDC (enough for both operations)
		ledger := GetWalletLedger(db, addr)
		require.NoError(t, ledger.Record(addr, "usdc", decimal.NewFromInt(2000)))

		// Verify initial balance
		initialBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(2000), initialBalance)

		// Combined operation: allocate 150 to channel + resize (deposit) 100 more
		resizeParams := ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),  // Allocation: moves funds from user balance to channel
			ResizeAmount:     big.NewInt(-100), // Resize: immediately withdraws allocated funds from channel
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(resizeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 12,
				Method:    "resize_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		// Sign request
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		// Call handler
		resp, err := HandleResizeChannel(nil, rpcReq, db, &signer)
		require.NoError(t, err)

		// Validate response
		assert.Equal(t, "resize_channel", resp.Res.Method)
		resObj, ok := resp.Res.Params[0].(ResizeChannelResponse)
		require.True(t, ok, "Response should be ResizeChannelResponse")
		assert.Equal(t, ch.ChannelID, resObj.ChannelID)
		assert.Equal(t, ch.Version+1, resObj.Version)

		// New channel amount should be initial + AllocateAmount + ResizeAmount = 0 + 100 - 100 = 0
		assert.Equal(t, 0, resObj.Allocations[0].Amount.Cmp(big.NewInt(0)), "Combined allocation+resize amount mismatch")
		assert.Equal(t, 0, resObj.Allocations[1].Amount.Cmp(big.NewInt(0)), "Broker allocation should be zero")

		// Verify channel state in database remains unchanged (no update until blockchain confirmation)
		var unchangedChannel Channel
		require.NoError(t, db.Where("channel_id = ?", ch.ChannelID).First(&unchangedChannel).Error)
		assert.Equal(t, initialAmount, unchangedChannel.Amount) // Should remain unchanged
		assert.Equal(t, ch.Version, unchangedChannel.Version)   // Should remain unchanged
		assert.Equal(t, ChannelStatusOpen, unchangedChannel.Status)

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(addr, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(2000), finalBalance) // Should remain unchanged
	})
}

// TestHandleCloseChannel tests the close channel handler functionality
func TestHandleCloseChannel(t *testing.T) {
	t.Run("SuccessfulCloseChannel", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		// Setup test DB
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Create asset
		asset := Asset{Token: "0xTokenClose", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with amount 500
		initialAmount := uint64(500)
		ch := Channel{
			ChannelID:   "0xChanClose",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     2,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger so that raw units match channel.Amount
		require.NoError(t, GetWalletLedger(db, addr).Record(
			addr,
			"usdc",
			decimal.NewFromBigInt(big.NewInt(int64(initialAmount)), -int32(asset.Decimals)),
		))

		// Prepare close params
		closeParams := CloseChannelParams{
			ChannelID:        ch.ChannelID,
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(closeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 10,
				Method:    "close_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		// Sign request
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		// Call handler
		resp, err := HandleCloseChannel(nil, rpcReq, db, &signer)
		require.NoError(t, err)

		// Validate response
		assert.Equal(t, "close_channel", resp.Res.Method)
		resObj, ok := resp.Res.Params[0].(CloseChannelResponse)
		require.True(t, ok, "Response should be CloseChannelResponse")
		assert.Equal(t, ch.ChannelID, resObj.ChannelID)
		assert.Equal(t, ch.Version+1, resObj.Version)

		// Final allocation should send full balance to destination
		assert.Equal(t, 0, resObj.FinalAllocations[0].Amount.Cmp(new(big.Int).SetUint64(initialAmount)), "Primary allocation mismatch")
		assert.Equal(t, 0, resObj.FinalAllocations[1].Amount.Cmp(big.NewInt(0)), "Broker allocation should be zero")
	})
	t.Run("ErrorOtherChallengedChannel", func(t *testing.T) {
		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		addr := signer.GetAddress().Hex()

		// Setup test DB
		db, cleanup := setupTestDB(t)
		defer cleanup()

		// Create asset
		asset := Asset{Token: "0xTokenClose", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with amount 500
		initialAmount := uint64(500)

		// Seed other challenged channel
		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChanChallenged",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusChallenged,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     2,
		}).Error)

		ch := Channel{
			ChannelID:   "0xChanClose",
			Participant: addr,
			Wallet:      addr,
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			Amount:      initialAmount,
			Version:     2,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger so that raw units match channel.Amount
		require.NoError(t, GetWalletLedger(db, addr).Record(
			addr,
			"usdc",
			decimal.NewFromBigInt(big.NewInt(int64(initialAmount)), -int32(asset.Decimals)),
		))

		// Prepare close params
		closeParams := CloseChannelParams{
			ChannelID:        ch.ChannelID,
			FundsDestination: addr,
		}
		paramsBytes, _ := json.Marshal(closeParams)

		rpcReq := &RPCMessage{
			Req: &RPCData{
				RequestID: 10,
				Method:    "close_channel",
				Params:    []any{json.RawMessage(paramsBytes)},
				Timestamp: uint64(time.Now().Unix()),
			},
		}

		// Sign request
		rawReq, err := json.Marshal(rpcReq.Req)
		require.NoError(t, err)
		rpcReq.Req.rawBytes = rawReq
		sig, err := signer.Sign(rawReq)
		require.NoError(t, err)
		rpcReq.Sig = []string{hexutil.Encode(sig)}

		// Call handler
		_, err = HandleCloseChannel(nil, rpcReq, db, &signer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has challenged channels")
	})
}
