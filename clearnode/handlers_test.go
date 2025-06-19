package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
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

		// Call handler
		resp, err := HandleCreateApplication(nil, rpcReq, db)
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
		_, err = HandleCreateApplication(nil, rpcReq, db)
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
		resp, err := HandleSubmitState(nil, req, db)
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
	resp, err := HandleCloseApplication(nil, req, db)
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
