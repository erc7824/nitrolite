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

	// Generate a unique DSN for the in-memory DB to avoid sharing data between tests
	uniqueDSN := fmt.Sprintf("file::memory:test%s?mode=memory&cache=shared", uuid.NewString())

	db, err := gorm.Open(sqlite.Open(uniqueDSN), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate all required models
	err = db.AutoMigrate(&Entry{}, &Channel{}, &AppSession{}, &RPCRecord{}, &Asset{})
	require.NoError(t, err)

	return db
}

// setupTestPostgres creates a PostgreSQL database using testcontainers
func setupTestPostgres(ctx context.Context, t testing.TB) (*gorm.DB, testcontainers.Container) {
	t.Helper()

	const dbName = "postgres"
	const dbUser = "postgres"
	const dbPassword = "postgres"

	// Start the PostgreSQL container
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

	// Get connection string
	url, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	log.Println("PostgreSQL URL:", url)

	// Connect to database
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate all required models
	err = db.AutoMigrate(&Entry{}, &Channel{}, &AppSession{}, &RPCRecord{}, &Asset{})
	require.NoError(t, err)

	return db, postgresContainer
}

// setupTestDB creates a test database based on the TEST_DB_DRIVER environment variable
func setupTestDB(t testing.TB) (*gorm.DB, func()) {
	t.Helper()

	// Create a context with the test timeout
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
		cleanup = func() {} // No cleanup needed for SQLite in-memory database
	}

	return db, cleanup
}

// TestHandlePing tests the ping handler functionality
func TestHandlePing(t *testing.T) {
	// Test case 1: Simple ping with no parameters
	rpcRequest1 := &RPCMessage{
		Req: &RPCData{
			RequestID: 1,
			Method:    "ping",
			Params:    []any{nil},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	response1, err := HandlePing(rpcRequest1)
	require.NoError(t, err)
	assert.NotNil(t, response1)

	require.Equal(t, "pong", response1.Res.Method)
}

// TestHandleCloseVirtualApp tests the close virtual app handler functionality
func TestHandleCloseVirtualApp(t *testing.T) {
	raw, err := crypto.GenerateKey()
	require.NoError(t, err)

	signer := Signer{privateKey: raw}
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

	// Create a virtual app
	vAppID := "0xVApp123"
	require.NoError(t, db.Create(&AppSession{
		SessionID:    vAppID,
		Participants: []string{participantA, participantB},
		Status:       ChannelStatusOpen,
		Challenge:    60,
		Weights:      []int64{100, 0},
		Quorum:       100,
	}).Error)

	assetSymbol := "usdc"

	require.NoError(t, GetParticipantLedger(db, participantA).Record(vAppID, assetSymbol, decimal.NewFromInt(200)))
	require.NoError(t, GetParticipantLedger(db, participantB).Record(vAppID, assetSymbol, decimal.NewFromInt(300)))

	closeParams := CloseAppSessionParams{
		AppSessionID: vAppID,
		Allocations: []AppAllocation{
			{Participant: participantA, AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
			{Participant: participantB, AssetSymbol: assetSymbol, Amount: decimal.NewFromInt(250)},
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

	signData := CloseAppSignData{
		RequestID: req.Req.RequestID,
		Method:    req.Req.Method,
		Params:    []CloseAppSessionParams{closeParams},
		Timestamp: req.Req.Timestamp,
	}
	signBytes, _ := json.Marshal(signData)
	sig, _ := signer.Sign(signBytes)
	req.Sig = []string{hexutil.Encode(sig)}

	resp, err := HandleCloseApplication(req, db)
	require.NoError(t, err)
	assert.Equal(t, "close_app_session", resp.Res.Method)
	var updated AppSession
	require.NoError(t, db.Where("session_id = ?", vAppID).First(&updated).Error)
	assert.Equal(t, ChannelStatusClosed, updated.Status)

	// Check that funds were transferred back to channels according to allocations
	balA, _ := GetParticipantLedger(db, participantA).Balance(participantA, "usdc")
	balB, _ := GetParticipantLedger(db, participantB).Balance(participantB, "usdc")
	assert.Equal(t, decimal.NewFromInt(250), balA)
	assert.Equal(t, decimal.NewFromInt(250), balB)

	// ► v-app accounts drained
	vBalA, _ := GetParticipantLedger(db, participantA).Balance(vAppID, "usdc")
	vBalB, _ := GetParticipantLedger(db, participantB).Balance(vAppID, "usdc")

	assert.True(t, vBalA.IsZero(), "Participant A vApp balance should be zero")
	assert.True(t, vBalB.IsZero(), "Participant B vApp balance should be zero")
}

func TestHandleCreateVirtualApp(t *testing.T) {
	// Generate private keys for both participants
	rawA, _ := crypto.GenerateKey()
	rawB, _ := crypto.GenerateKey()
	signerA := Signer{privateKey: rawA}
	signerB := Signer{privateKey: rawB}
	addrA := signerA.GetAddress().Hex()
	addrB := signerB.GetAddress().Hex()

	db, cleanup := setupTestDB(t)
	defer cleanup()

	// open direct channels (still required elsewhere in code-base)
	token := "0xTokenXYZ"
	for i, p := range []string{addrA, addrB} {
		ch := &Channel{
			ChannelID:   fmt.Sprintf("0xChannel%c", 'A'+i),
			Participant: p,
			Status:      ChannelStatusOpen,
			Token:       token,
			Nonce:       1,
		}
		require.NoError(t, db.Create(ch).Error)
	}

	require.NoError(t, GetParticipantLedger(db, addrA).Record(addrA, "usdc", decimal.NewFromInt(100)))
	require.NoError(t, GetParticipantLedger(db, addrB).Record(addrB, "usdc", decimal.NewFromInt(200)))

	ts := uint64(time.Now().Unix())
	def := AppDefinition{
		Protocol:     "test-proto",
		Participants: []string{addrA, addrB},
		Weights:      []uint64{1, 1},
		Quorum:       2,
		Challenge:    60,
		Nonce:        ts, // if omitted, handler would use ts anyway
	}
	asset := "usdc"
	createParams := CreateAppSessionParams{
		Definition: def,
		Allocations: []AppAllocation{
			{Participant: addrA, AssetSymbol: asset, Amount: decimal.NewFromInt(100)},
			{Participant: addrB, AssetSymbol: asset, Amount: decimal.NewFromInt(200)},
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

	// sign exactly like the handler
	signData := CreateAppSignData{
		RequestID: rpcReq.Req.RequestID,
		Method:    rpcReq.Req.Method,
		Params:    []CreateAppSessionParams{createParams},
		Timestamp: rpcReq.Req.Timestamp,
	}
	signBytes, _ := signData.MarshalJSON()
	sigA, _ := signerA.Sign(signBytes)
	sigB, _ := signerB.Sign(signBytes)
	rpcReq.Sig = []string{hexutil.Encode(sigA), hexutil.Encode(sigB)}

	resp, err := HandleCreateApplication(rpcReq, db)
	require.NoError(t, err)

	// ► response sanity
	assert.Equal(t, "create_app_session", resp.Res.Method)
	appResp, ok := resp.Res.Params[0].(*AppSessionResponse)
	require.True(t, ok)
	assert.Equal(t, string(ChannelStatusOpen), appResp.Status)

	// ► v-app row exists
	var vApp AppSession
	require.NoError(t, db.Where("session_id = ?", appResp.AppSessionID).First(&vApp).Error)
	assert.ElementsMatch(t, []string{addrA, addrB}, vApp.Participants)

	// ► participant accounts drained
	partBalA, _ := GetParticipantLedger(db, addrA).Balance(addrA, "usdc")
	partBalB, _ := GetParticipantLedger(db, addrB).Balance(addrB, "usdc")
	assert.True(t, partBalA.IsZero(), "Participant A balance should be zero")
	assert.True(t, partBalB.IsZero(), "Participant B balance should be zero")

	// ► virtual-app funded - each participant can see the total app session balance (300)
	vBalA, _ := GetParticipantLedger(db, addrA).Balance(appResp.AppSessionID, "usdc")
	vBalB, _ := GetParticipantLedger(db, addrB).Balance(appResp.AppSessionID, "usdc")
	assert.Equal(t, decimal.NewFromInt(100).String(), vBalA.String())
	assert.Equal(t, decimal.NewFromInt(200).String(), vBalB.String())
}

// TestHandleGetLedgerBalances tests the get ledger balances handler functionality
func TestHandleGetLedgerBalances(t *testing.T) {
	// Set up test database with cleanup
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ledger := GetParticipantLedger(db, "0xParticipant1")
	err := ledger.Record("0xParticipant1", "usdc", decimal.NewFromInt(1000))
	require.NoError(t, err)

	// Create RPC request with token address parameter
	params := map[string]string{
		"participant": "0xParticipant1",
	}
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

	// Use the test-specific handler instead of the actual one
	msg, err := HandleGetLedgerBalances(rpcRequest, "0xParticipant1", db)
	require.NoError(t, err)
	assert.NotNil(t, msg)

	// Extract the response data
	var responseParams []any
	responseParams = msg.Res.Params
	require.NotEmpty(t, responseParams)

	// First parameter should be an array of Balance
	balancesArray, ok := responseParams[0].([]Balance)
	require.True(t, ok, "Response should contain an array of Balance")

	// We should have 1 balance entry
	assert.Equal(t, 1, len(balancesArray), "Should have 1 balance entry")

	// Check the contents of each balance
	expectedAssets := map[string]decimal.Decimal{
		"usdc": decimal.NewFromInt(1000),
	}

	for _, balance := range balancesArray {
		expectedBalance, exists := expectedAssets[balance.Asset]
		assert.True(t, exists, "Unexpected asset in response: %s", balance.Asset)
		assert.Equal(t, expectedBalance, balance.Amount, "Incorrect balance for asset %s", balance.Asset)

		// Remove from map to ensure each asset appears only once
		delete(expectedAssets, balance.Asset)
	}

	assert.Empty(t, expectedAssets, "Not all expected assets were found in the response")
}

// TestHandleGetConfig tests the get config handler functionality
func TestHandleGetConfig(t *testing.T) {
	// Create a mock config with supported networks
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

	raw, err := crypto.GenerateKey()
	require.NoError(t, err)

	signer := Signer{privateKey: raw}

	response, err := HandleGetConfig(rpcRequest, mockConfig, &signer)
	require.NoError(t, err)
	assert.NotNil(t, response)

	// Extract the response data
	var responseParams []any
	responseParams = response.Res.Params
	require.NotEmpty(t, responseParams)

	// First parameter should be a BrokerConfig
	configMap, ok := responseParams[0].(BrokerConfig)
	require.True(t, ok, "Response should contain a BrokerConfig")

	// Verify broker address
	assert.Equal(t, signer.GetAddress().Hex(), configMap.BrokerAddress)

	// Verify supported networks
	require.Len(t, configMap.Networks, 3, "Should have 3 supported networks")

	// Map to check all networks are present
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

	assert.Empty(t, expectedNetworks, "All expected networks should be found in the response")
}

// TestHandleGetChannels tests the get channels functionality
func TestHandleGetChannels(t *testing.T) {
	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: rawKey}
	participantAddr := signer.GetAddress().Hex()

	db, cleanup := setupTestDB(t)
	defer cleanup()

	tokenAddress := "0xToken123"
	chainID := uint32(137)

	channels := []Channel{
		{
			ChannelID:   "0xChannel1",
			Participant: participantAddr,
			Status:      ChannelStatusOpen,
			Token:       tokenAddress + "1",
			ChainID:     chainID,
			Amount:      1000,
			Nonce:       1,
			Version:     10,
			Challenge:   86400,
			Adjudicator: "0xAdj1",
			CreatedAt:   time.Now().Add(-24 * time.Hour), // 1 day ago
			UpdatedAt:   time.Now(),
		},
		{
			ChannelID:   "0xChannel2",
			Participant: participantAddr,
			Status:      ChannelStatusClosed,
			Token:       tokenAddress + "2",
			ChainID:     chainID,
			Amount:      2000,
			Nonce:       2,
			Version:     20,
			Challenge:   86400,
			Adjudicator: "0xAdj2",
			CreatedAt:   time.Now().Add(-12 * time.Hour), // 12 hours ago
			UpdatedAt:   time.Now(),
		},
		{
			ChannelID:   "0xChannel3",
			Participant: participantAddr,
			Status:      ChannelStatusJoining,
			Token:       tokenAddress + "3",
			ChainID:     chainID,
			Amount:      3000,
			Nonce:       3,
			Version:     30,
			Challenge:   86400,
			Adjudicator: "0xAdj3",
			CreatedAt:   time.Now().Add(-6 * time.Hour), // 6 hours ago
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
		"participant": participantAddr,
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

	reqBytes, err := json.Marshal(rpcRequest.Req)
	require.NoError(t, err)
	signed, err := signer.Sign(reqBytes)
	require.NoError(t, err)
	rpcRequest.Sig = []string{hexutil.Encode(signed)}

	response, err := HandleGetChannels(rpcRequest, db)
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.Equal(t, "get_channels", response.Res.Method)
	assert.Equal(t, uint64(123), response.Res.RequestID)

	require.Len(t, response.Res.Params, 1, "Response should contain a slice of ChannelResponse")
	channelsSlice, ok := response.Res.Params[0].([]ChannelResponse)
	require.True(t, ok, "Response parameter should be a slice of ChannelResponse")

	// Should return all 3 channels for the participant
	assert.Len(t, channelsSlice, 3, "Should return all 3 channels for the participant")

	// Verify the channels are ordered by creation date (newest first)
	assert.Equal(t, "0xChannel3", channelsSlice[0].ChannelID, "First channel should be the newest")
	assert.Equal(t, "0xChannel2", channelsSlice[1].ChannelID, "Second channel should be the middle one")
	assert.Equal(t, "0xChannel1", channelsSlice[2].ChannelID, "Third channel should be the oldest")

	// Verify channel data is correct
	for _, ch := range channelsSlice {
		assert.Equal(t, participantAddr, ch.Participant, "ParticipantA should match")
		// Token now has a suffix, so we check it starts with the base token address
		assert.True(t, strings.HasPrefix(ch.Token, tokenAddress), "Token should start with the base token address")
		assert.Equal(t, chainID, ch.ChainID, "NetworkID should match")

		// Find the corresponding original channel to compare with
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

	// Test with status filter for "open" channels
	openStatusParams := map[string]string{
		"participant": participantAddr,
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

	reqBytes, err = json.Marshal(openStatusRequest.Req)
	require.NoError(t, err)
	signed, err = signer.Sign(reqBytes)
	require.NoError(t, err)
	openStatusRequest.Sig = []string{hexutil.Encode(signed)}

	openStatusResponse, err := HandleGetChannels(openStatusRequest, db)
	require.NoError(t, err)
	require.NotNil(t, openStatusResponse)

	// Extract and verify filtered channels
	openChannels, ok := openStatusResponse.Res.Params[0].([]ChannelResponse)
	require.True(t, ok, "Response parameter should be a slice of ChannelResponse")
	assert.Len(t, openChannels, 1, "Should return only 1 open channel")
	assert.Equal(t, "0xChannel1", openChannels[0].ChannelID, "Should return the open channel")
	assert.Equal(t, ChannelStatusOpen, openChannels[0].Status, "Status should be open")

	// Test with status filter for "closed" channels
	closedStatusParams := map[string]string{
		"participant": participantAddr,
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

	reqBytes, err = json.Marshal(closedStatusRequest.Req)
	require.NoError(t, err)
	signed, err = signer.Sign(reqBytes)
	require.NoError(t, err)
	closedStatusRequest.Sig = []string{hexutil.Encode(signed)}

	closedStatusResponse, err := HandleGetChannels(closedStatusRequest, db)
	require.NoError(t, err)
	require.NotNil(t, closedStatusResponse)

	// Extract and verify filtered channels
	closedChannels, ok := closedStatusResponse.Res.Params[0].([]ChannelResponse)
	require.True(t, ok, "Response parameter should be a slice of ChannelResponse")
	assert.Len(t, closedChannels, 1, "Should return only 1 closed channel")
	assert.Equal(t, "0xChannel2", closedChannels[0].ChannelID, "Should return the closed channel")
	assert.Equal(t, ChannelStatusClosed, closedChannels[0].Status, "Status should be closed")

	// Test with status filter for "joining" channels
	joiningStatusParams := map[string]string{
		"participant": participantAddr,
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

	reqBytes, err = json.Marshal(joiningStatusRequest.Req)
	require.NoError(t, err)
	signed, err = signer.Sign(reqBytes)
	require.NoError(t, err)
	joiningStatusRequest.Sig = []string{hexutil.Encode(signed)}

	joiningStatusResponse, err := HandleGetChannels(joiningStatusRequest, db)
	require.NoError(t, err)
	require.NotNil(t, joiningStatusResponse)

	// Extract and verify filtered channels
	joiningChannels, ok := joiningStatusResponse.Res.Params[0].([]ChannelResponse)
	require.True(t, ok, "Response parameter should be a slice of ChannelResponse")
	assert.Len(t, joiningChannels, 1, "Should return only 1 joining channel")
	assert.Equal(t, "0xChannel3", joiningChannels[0].ChannelID, "Should return the joining channel")
	assert.Equal(t, ChannelStatusJoining, joiningChannels[0].Status, "Status should be joining")

	// Test with missing participant parameter
	missingParamReq := &RPCMessage{
		Req: &RPCData{
			RequestID: 789,
			Method:    "get_channels",
			Params:    []any{map[string]string{}}, // Empty map
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{hexutil.Encode(signed)},
	}

	_, err = HandleGetChannels(missingParamReq, db)
	assert.Error(t, err, "Should return error with missing participant")
	assert.Contains(t, err.Error(), "missing participant", "Error should mention missing participant")
}

// TestHandleGetAssets tests the get assets handler functionality
func TestHandleGetAssets(t *testing.T) {
	// Set up test database with cleanup
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create some test assets
	testAssets := []Asset{
		{
			Token:    "0xToken1",
			ChainID:  137, // Polygon
			Symbol:   "usdc",
			Decimals: 6,
		},
		{
			Token:    "0xToken2",
			ChainID:  137, // Polygon
			Symbol:   "weth",
			Decimals: 18,
		},
		{
			Token:    "0xToken3",
			ChainID:  42220, // Celo
			Symbol:   "celo",
			Decimals: 18,
		},
		{
			Token:    "0xToken4",
			ChainID:  8453, // Base
			Symbol:   "usdbc",
			Decimals: 6,
		},
	}

	for _, asset := range testAssets {
		require.NoError(t, db.Create(&asset).Error)
	}

	// Test Case 1: Get all assets
	rpcRequest1 := &RPCMessage{
		Req: &RPCData{
			RequestID: 1,
			Method:    "get_assets",
			Params:    []any{},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	// Call the handler
	resp1, err := HandleGetAssets(rpcRequest1, db)
	require.NoError(t, err)
	assert.NotNil(t, resp1)

	// Verify response format
	assert.Equal(t, "get_assets", resp1.Res.Method)
	assert.Equal(t, uint64(1), resp1.Res.RequestID)
	require.Len(t, resp1.Res.Params, 1, "Response should contain an array of AssetResponse objects")

	// Extract and verify assets
	assets1, ok := resp1.Res.Params[0].([]AssetResponse)
	require.True(t, ok, "Response parameter should be a slice of AssetResponse")
	assert.Len(t, assets1, 4, "Should return all 4 assets")

	// Verify specific asset details
	foundSymbols := make(map[string]bool)
	for _, asset := range assets1 {
		foundSymbols[asset.Symbol] = true

		// Find original asset to compare with
		var originalAsset Asset
		for _, a := range testAssets {
			if a.Symbol == asset.Symbol && a.ChainID == asset.ChainID {
				originalAsset = a
				break
			}
		}

		assert.Equal(t, originalAsset.Token, asset.Token, "Token should match")
		assert.Equal(t, originalAsset.ChainID, asset.ChainID, "ChainID should match")
		assert.Equal(t, originalAsset.Decimals, asset.Decimals, "Decimals should match")
	}

	assert.Len(t, foundSymbols, 4, "Should have found 4 unique symbols")
	assert.True(t, foundSymbols["usdc"], "Should include USDC")
	assert.True(t, foundSymbols["weth"], "Should include WETH")
	assert.True(t, foundSymbols["celo"], "Should include CELO")
	assert.True(t, foundSymbols["usdbc"], "Should include USDBC")

	// Test Case 2: Filter by chain_id (Polygon)
	chainID := float64(137) // Polygon
	params2 := map[string]interface{}{
		"chain_id": chainID,
	}
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

	// Call the handler with chain_id filter
	resp2, err := HandleGetAssets(rpcRequest2, db)
	require.NoError(t, err)
	assert.NotNil(t, resp2)

	// Verify response format for filtered assets
	assert.Equal(t, "get_assets", resp2.Res.Method)
	assert.Equal(t, uint64(2), resp2.Res.RequestID)

	// Extract and verify filtered assets
	assets2, ok := resp2.Res.Params[0].([]AssetResponse)
	require.True(t, ok, "Response parameter should be a slice of AssetResponse")
	assert.Len(t, assets2, 2, "Should return 2 Polygon assets")

	// Ensure all returned assets are for Polygon
	for _, asset := range assets2 {
		assert.Equal(t, uint32(137), asset.ChainID, "ChainID should be Polygon (137)")
	}

	// Verify the symbols of the returned assets
	symbols := make(map[string]bool)
	for _, asset := range assets2 {
		symbols[asset.Symbol] = true
	}
	assert.Len(t, symbols, 2, "Should have 2 unique symbols")
	assert.True(t, symbols["usdc"], "Should include USDC on Polygon")
	assert.True(t, symbols["weth"], "Should include WETH on Polygon")

	// Test Case 3: Filter by chain_id (Celo)
	chainID = float64(42220) // Celo
	params3 := map[string]interface{}{
		"chain_id": chainID,
	}
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

	// Call the handler with chain_id filter
	resp3, err := HandleGetAssets(rpcRequest3, db)
	require.NoError(t, err)
	assert.NotNil(t, resp3)

	// Extract and verify filtered assets
	assets3, ok := resp3.Res.Params[0].([]AssetResponse)
	require.True(t, ok, "Response parameter should be a slice of AssetResponse")
	assert.Len(t, assets3, 1, "Should return 1 Celo asset")
	assert.Equal(t, "celo", assets3[0].Symbol, "Should be CELO")
	assert.Equal(t, uint32(42220), assets3[0].ChainID, "ChainID should be Celo (42220)")

	// Test Case 4: Filter by non-existent chain_id
	chainID = float64(1) // Ethereum Mainnet (not in our test data)
	params4 := map[string]interface{}{
		"chain_id": chainID,
	}
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

	// Call the handler with non-existent chain_id filter
	resp4, err := HandleGetAssets(rpcRequest4, db)
	require.NoError(t, err)
	assert.NotNil(t, resp4)

	// Extract and verify filtered assets (should be empty)
	assets4, ok := resp4.Res.Params[0].([]AssetResponse)
	require.True(t, ok, "Response parameter should be a slice of AssetResponse")
	assert.Len(t, assets4, 0, "Should return 0 assets for non-existent chain_id")
}

// TestHandleGetAppSessions tests the get app sessions handler functionality
func TestHandleGetAppSessions(t *testing.T) {
	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: rawKey}
	participantAddr := signer.GetAddress().Hex()

	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create some test app sessions
	sessions := []AppSession{
		{
			SessionID:    "0xSession1",
			Participants: []string{participantAddr, "0xParticipant2"},
			Status:       ChannelStatusOpen,
			Protocol:     "test-app-1",
			Challenge:    60,
			Weights:      []int64{50, 50},
			Quorum:       75,
			Nonce:        1,
			Version:      1,
		},
		{
			SessionID:    "0xSession2",
			Participants: []string{participantAddr, "0xParticipant3"},
			Status:       ChannelStatusClosed,
			Protocol:     "test-app-2",
			Challenge:    120,
			Weights:      []int64{30, 70},
			Quorum:       80,
			Nonce:        2,
			Version:      2,
		},
		{
			SessionID:    "0xSession3",
			Participants: []string{"0xParticipant4", "0xParticipant5"},
			Status:       ChannelStatusOpen,
			Protocol:     "test-app-3",
			Challenge:    90,
			Weights:      []int64{40, 60},
			Quorum:       60,
			Nonce:        3,
			Version:      3,
		},
	}

	for _, session := range sessions {
		require.NoError(t, db.Create(&session).Error)
	}

	// Test Case 1: Get all app sessions for the participant
	params1 := map[string]string{
		"participant": participantAddr,
	}
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

	// Call the handler
	resp1, err := HandleGetAppSessions(rpcRequest1, db)
	require.NoError(t, err)
	assert.NotNil(t, resp1)

	// Verify response format
	assert.Equal(t, "get_app_sessions", resp1.Res.Method)
	assert.Equal(t, uint64(1), resp1.Res.RequestID)
	require.Len(t, resp1.Res.Params, 1, "Response should contain an array of AppSessionResponse objects")

	// Extract and verify app sessions
	sessionResponses, ok := resp1.Res.Params[0].([]AppSessionResponse)
	require.True(t, ok, "Response parameter should be a slice of AppSessionResponse")
	assert.Len(t, sessionResponses, 2, "Should return 2 app sessions for the participant")

	// Verify the response contains the expected app sessions
	foundSessions := make(map[string]bool)
	for _, session := range sessionResponses {
		foundSessions[session.AppSessionID] = true

		// Find the original session to compare with
		var originalSession AppSession
		for _, s := range sessions {
			if s.SessionID == session.AppSessionID {
				originalSession = s
				break
			}
		}

		assert.Equal(t, string(originalSession.Status), session.Status, "Status should match")
	}

	assert.True(t, foundSessions["0xSession1"], "Should include Session1")
	assert.True(t, foundSessions["0xSession2"], "Should include Session2")
	assert.False(t, foundSessions["0xSession3"], "Should not include Session3")

	// Test Case 2: Get open app sessions for the participant
	params2 := map[string]string{
		"participant": participantAddr,
		"status":      string(ChannelStatusOpen),
	}
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

	// Call the handler
	resp2, err := HandleGetAppSessions(rpcRequest2, db)
	require.NoError(t, err)
	assert.NotNil(t, resp2)

	// Extract and verify filtered app sessions
	sessionResponses2, ok := resp2.Res.Params[0].([]AppSessionResponse)
	require.True(t, ok, "Response parameter should be a slice of AppSessionResponse")
	assert.Len(t, sessionResponses2, 1, "Should return 1 open app session for the participant")
	assert.Equal(t, "0xSession1", sessionResponses2[0].AppSessionID, "Should be Session1")
	assert.Equal(t, string(ChannelStatusOpen), sessionResponses2[0].Status, "Status should be open")

	// Test Case 3: Error case - missing participant
	rpcRequest3 := &RPCMessage{
		Req: &RPCData{
			RequestID: 3,
			Method:    "get_app_sessions",
			Params:    []any{json.RawMessage(`{}`)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	// Call with missing participant
	resp3, err := HandleGetAppSessions(rpcRequest3, db)
	assert.Error(t, err, "Should return error with missing participant")
	assert.Nil(t, resp3)
	assert.Contains(t, err.Error(), "missing participant", "Error should mention missing participant")
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
			Timestamp: timestamp - 3600, // 1 hour ago
			ReqSig:    []string{"sig1"},
			Response:  []byte(`{"res":[1,"pong",[],1621234567890]}`),
			ResSig:    []string{},
		},
		{
			Sender:    participantAddr,
			ReqID:     2,
			Method:    "get_config",
			Params:    []byte(`[]`),
			Timestamp: timestamp - 1800, // 30 minutes ago
			ReqSig:    []string{"sig2"},
			Response:  []byte(`{"res":[2,"get_config",[{"broker_address":"0xBroker"}],1621234597890]}`),
			ResSig:    []string{},
		},
		{
			Sender:    participantAddr,
			ReqID:     3,
			Method:    "get_channels",
			Params:    []byte(`[{"participant":"` + participantAddr + `"}]`),
			Timestamp: timestamp - 900, // 15 minutes ago
			ReqSig:    []string{"sig3"},
			Response:  []byte(`{"res":[3,"get_channels",[[]]],1621234627890]}`),
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

	reqBytes, err := json.Marshal(rpcRequest.Req)
	require.NoError(t, err)
	signed, err := signer.Sign(reqBytes)
	require.NoError(t, err)
	rpcRequest.Sig = []string{hexutil.Encode(signed)}

	response, err := HandleGetRPCHistory(participantAddr, rpcRequest, rpcStore)
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
		Sig: []string{hexutil.Encode(signed)},
	}

	_, err = HandleGetRPCHistory("", missingParamReq, rpcStore)
	assert.Error(t, err, "Should return error with missing participant")
	assert.Contains(t, err.Error(), "missing participant", "Error should mention missing participant")
}

// TestHandleGetLedgerEntries tests the get ledger entries handler functionality
func TestHandleGetLedgerEntries(t *testing.T) {
	// Set up test database with cleanup
	db, cleanup := setupTestDB(t)
	defer cleanup()

	participant := "0xParticipant1"
	ledger := GetParticipantLedger(db, participant)

	// Create test entries with different assets
	testData := []struct {
		asset  string
		amount decimal.Decimal
	}{
		{"usdc", decimal.NewFromInt(100)},
		{"usdc", decimal.NewFromInt(200)},
		{"usdc", decimal.NewFromInt(-50)},
		{"eth", decimal.NewFromFloat(1.5)},
		{"eth", decimal.NewFromFloat(-0.5)},
	}

	// Insert test entries
	for _, data := range testData {
		err := ledger.Record(participant, data.asset, data.amount)
		require.NoError(t, err)
	}

	// Test Case 1: Get all entries for the participant
	params1 := map[string]string{
		"account_id": participant,
	}
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

	// Call the handler
	resp1, err := HandleGetLedgerEntries(rpcRequest1, participant, db)
	require.NoError(t, err)
	assert.NotNil(t, resp1)

	// Verify response format
	assert.Equal(t, "get_ledger_entries", resp1.Res.Method)
	assert.Equal(t, uint64(1), resp1.Res.RequestID)
	require.Len(t, resp1.Res.Params, 1, "Response should contain an array of Entry objects")

	// Extract and verify entries
	entries1, ok := resp1.Res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")

	assert.Len(t, entries1, 5, "Should return all 5 entries")

	// Count entries by asset
	assetCounts := map[string]int{}
	for _, entry := range entries1 {
		assetCounts[entry.Asset]++
		assert.Equal(t, participant, entry.AccountID)
		assert.Equal(t, participant, entry.Participant)
	}
	assert.Equal(t, 3, assetCounts["usdc"], "Should have 3 USDC entries")
	assert.Equal(t, 2, assetCounts["eth"], "Should have 2 ETH entries")

	// Test Case 2: Get entries for a specific asset
	params2 := map[string]string{
		"account_id": participant,
		"asset":      "usdc",
	}
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

	// Call the handler
	resp2, err := HandleGetLedgerEntries(rpcRequest2, participant, db)
	require.NoError(t, err)
	assert.NotNil(t, resp2)

	// Verify response format for specific asset
	assert.Equal(t, "get_ledger_entries", resp2.Res.Method)
	assert.Equal(t, uint64(2), resp2.Res.RequestID)

	// Extract and verify entries for specific asset
	entries2, ok := resp2.Res.Params[0].([]LedgerEntryResponse)
	require.True(t, ok, "Response parameter should be a slice of Entry")
	assert.Len(t, entries2, 3, "Should return 3 USDC entries")

	// Ensure all entries are for USDC
	for _, entry := range entries2 {
		assert.Equal(t, "usdc", entry.Asset)
		assert.Equal(t, participant, entry.AccountID)
		assert.Equal(t, participant, entry.Participant)
	}

	// Test Case 3: Error case - invalid participant
	rpcRequest3 := &RPCMessage{
		Req: &RPCData{
			RequestID: 3,
			Method:    "get_ledger_entries",
			Params:    []any{json.RawMessage(`{"participant": ""}`)},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"dummy-signature"},
	}

	// Call with empty participant
	resp3, err := HandleGetLedgerEntries(rpcRequest3, "", db)
	assert.Error(t, err, "Should return error with empty participant")
	assert.Nil(t, resp3)
}
