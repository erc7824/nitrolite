package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionKey(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	walletAddress := "0x1234567890123456789012345678901234567890"
	sessionSignerAddress := "0xabcdef1234567890abcdef1234567890abcdef12"
	app := "TestApp"
	scope := "trade"
	allowances := []Allowance{
		{Asset: "usdc", Amount: "1000"},
		{Asset: "eth", Amount: "5"},
	}
	expirationTime := time.Now().Add(24 * time.Hour)

	err := AddSessionKey(db, walletAddress, sessionSignerAddress, app, scope, allowances, expirationTime)
	require.NoError(t, err)

	retrievedWallet := GetWalletBySessionKey(sessionSignerAddress)
	assert.Equal(t, walletAddress, retrievedWallet)

	retrievedKeys, err := GetSessionKeysByWallet(db, walletAddress)
	require.NoError(t, err)
	require.Len(t, retrievedKeys, 1)

	sk := retrievedKeys[0]
	assert.Equal(t, walletAddress, sk.WalletAddress)
	assert.Equal(t, sessionSignerAddress, sk.Address)
	assert.Equal(t, app, sk.Application)
	assert.Equal(t, scope, sk.Scope)
	assert.WithinDuration(t, expirationTime, sk.ExpiresAt, time.Second)

	var retrievedSpendingCap []Allowance
	err = json.Unmarshal([]byte(*sk.Allowance), &retrievedSpendingCap)
	require.NoError(t, err)
	assert.Equal(t, allowances, retrievedSpendingCap)
	assert.Equal(t, walletAddress, GetWalletBySessionKey(sessionSignerAddress))

	sessionKeys, err := GetSessionKeysByWallet(db, walletAddress)
	require.NoError(t, err)
	assert.Len(t, sessionKeys, 1)
	assert.Equal(t, sessionSignerAddress, sessionKeys[0].Address)
	assert.Equal(t, walletAddress, sessionKeys[0].WalletAddress)
}

func TestSessionKeyMultipleKeys(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	walletAddress := "0x1234567890123456789012345678901234567890"

	sessionKeys := []struct {
		signerAddress string
		application   string
		scope         string
	}{
		{"0xkey1", "App1", "trade"},
		{"0xkey2", "App2", "view"},
		{"0xkey3", "App3", "admin"},
	}

	for _, sk := range sessionKeys {
		err := AddSessionKey(db, walletAddress, sk.signerAddress, sk.application, sk.scope, []Allowance{}, time.Now().Add(time.Hour))
		require.NoError(t, err)
	}

	retrievedKeys, err := GetSessionKeysByWallet(db, walletAddress)
	require.NoError(t, err)
	require.Len(t, retrievedKeys, 3)

	// Verify they are ordered by created_at DESC
	for i := 1; i < len(retrievedKeys); i++ {
		assert.True(t, retrievedKeys[i-1].CreatedAt.After(retrievedKeys[i].CreatedAt) ||
			retrievedKeys[i-1].CreatedAt.Equal(retrievedKeys[i].CreatedAt))
	}
}

func TestSessionKeyActiveKeys(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	walletAddress := "0x1234567890123456789012345678901234567890"

	err := AddSessionKey(db, walletAddress, "0xactive123", "ActiveApp", "trade", []Allowance{}, time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	// Create an expired session key by directly inserting into DB (bypassing validation)
	expiredKey := SessionKey{
		Address:       "0xexpired123",
		WalletAddress: walletAddress,
		Application:   "ExpiredApp",
		Allowance:     strPtr("[]"),
		Scope:         "view",
		ExpiresAt:     time.Now().Add(-1 * time.Hour).UTC(),
	}
	err = db.Create(&expiredKey).Error
	require.NoError(t, err)

	// Get all session keys (should be 2)
	allKeys, err := GetSessionKeysByWallet(db, walletAddress)
	require.NoError(t, err)
	assert.Len(t, allKeys, 2)

	// Get only active session keys (should be 1)
	activeKeys, err := GetActiveSessionKeysByWallet(db, walletAddress, nil)
	require.NoError(t, err)
	assert.Len(t, activeKeys, 1)
	assert.Equal(t, "0xactive123", activeKeys[0].Address)
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}

func TestSessionKeySpendingValidation(t *testing.T) {
	db := setupTestSqlite(t)
	err := loadSessionKeyCache(db)
	require.NoError(t, err)

	walletAddress := "0x1234567890123456789012345678901234567890"
	sessionKeyAddress := "0xsessionkey1234567890abcdef1234567890abcdef"

	allowances := []Allowance{
		{Asset: "usdc", Amount: "1000"},
		{Asset: "eth", Amount: "5"},
	}
	err = AddSessionKey(db, walletAddress, sessionKeyAddress, "TestApp", "trade", allowances, time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	sessionKey, err := GetSessionKeyIfActive(db, sessionKeyAddress)
	require.NoError(t, err, "Session key should be active")

	// Test 1: Valid spending within limits
	err = ValidateSessionKeySpending(db, sessionKey, "usdc", decimal.NewFromInt(100))
	assert.NoError(t, err, "Should allow spending within limits")

	err = ValidateSessionKeySpending(db, sessionKey, "eth", decimal.NewFromInt(2))
	assert.NoError(t, err, "Should allow ETH spending within limits")

	// Test 2: Spending exactly at the limit should be allowed
	err = ValidateSessionKeySpending(db, sessionKey, "usdc", decimal.NewFromInt(1000))
	assert.NoError(t, err, "Should allow spending exactly at limit")

	// Test 3: Spending over limit should fail
	err = ValidateSessionKeySpending(db, sessionKey, "usdc", decimal.NewFromInt(1001))
	assert.Error(t, err, "Should reject spending over limit")
	assert.Contains(t, err.Error(), "operation denied: insufficient session key allowance")

	// Test 4: Unauthorized asset should fail
	err = ValidateSessionKeySpending(db, sessionKey, "BTC", decimal.NewFromInt(1))
	assert.Error(t, err, "Should reject unauthorized asset")
	assert.Contains(t, err.Error(), "not allowed in session key spending cap")

	// Test 5: Simulate some spending through ledger entries
	walletAddr := common.HexToAddress(walletAddress)
	fromAccountID := NewAccountID(walletAddress)
	ledger := GetWalletLedger(db, walletAddr)

	// Simulate a spending transaction
	err = ledger.Record(fromAccountID, "usdc", decimal.NewFromInt(-200), &sessionKeyAddress)
	require.NoError(t, err)

	// Test 6: Check spending calculation
	currentSpending, err := CalculateSessionKeySpending(db, sessionKeyAddress, "usdc")
	require.NoError(t, err)
	assert.Equal(t, "200", currentSpending.String(), "Should correctly calculate current spending")

	// Test 7: Validate remaining allowance (refresh session key after spending)
	sessionKey, err = GetSessionKeyIfActive(db, sessionKeyAddress)
	require.NoError(t, err, "Session key should still be active")

	err = ValidateSessionKeySpending(db, sessionKey, "usdc", decimal.NewFromInt(800))
	assert.NoError(t, err, "Should allow spending remaining allowance")

	err = ValidateSessionKeySpending(db, sessionKey, "usdc", decimal.NewFromInt(801))
	assert.Error(t, err, "Should reject spending beyond remaining allowance")
	assert.Contains(t, err.Error(), "operation denied: insufficient session key allowance")

	// Test 8: Verify usage is calculated correctly on the fly
	usdcUsage, err := CalculateSessionKeySpending(db, sessionKeyAddress, "usdc")
	require.NoError(t, err)
	assert.Equal(t, "200", usdcUsage.String(), "Used allowance should reflect actual spending")
}

func TestSessionKeySpendingEdgeCases(t *testing.T) {
	db := setupTestSqlite(t)
	err := loadSessionKeyCache(db)
	require.NoError(t, err)

	walletAddress := "0x1234567890123456789012345678901234567890"
	sessionKeyAddress := "0xsessionkey1234567890abcdef1234567890abcdef"

	// Test 1: Session key with zero allowance
	zeroAllowances := []Allowance{
		{Asset: "usdc", Amount: "0"},
	}
	err = AddSessionKey(db, walletAddress, sessionKeyAddress, "ZeroApp", "trade", zeroAllowances, time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	sessionKey, err := GetSessionKeyIfActive(db, sessionKeyAddress)
	require.NoError(t, err)

	err = ValidateSessionKeySpending(db, sessionKey, "usdc", decimal.NewFromInt(1))
	assert.Error(t, err, "Should reject any spending with zero allowance")

	// Test 2: Non-existent session key
	_, err = GetSessionKeyIfActive(db, "0xnonexistent")
	assert.Error(t, err, "Should fail for non-existent session key")

	// Test 3: Negative spending amount (should not happen but let's test)
	err = ValidateSessionKeySpending(db, sessionKey, "usdc", decimal.NewFromInt(-10))
	assert.NoError(t, err, "Negative amounts should not trigger spending cap validation")
}

func TestSessionKeyTransferIntegration(t *testing.T) {
	router, db, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	err := loadSessionKeyCache(db)
	require.NoError(t, err)

	walletAddress := "0x1234567890123456789012345678901234567890"
	recipientAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	sessionKey, _ := crypto.GenerateKey()
	sessionSigner := Signer{privateKey: sessionKey}
	sessionKeyAddress := sessionSigner.GetAddress().Hex()

	allowances := []Allowance{
		{Asset: "usdc", Amount: "500"},
		{Asset: "eth", Amount: "2"},
	}
	err = AddSessionKey(db, walletAddress, sessionKeyAddress, "TestApp", "trade", allowances, time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	err = loadSessionKeyCache(db)
	require.NoError(t, err)

	walletAddr := common.HexToAddress(walletAddress)
	fromAccountID := NewAccountID(walletAddress)
	ledger := GetWalletLedger(db, walletAddr)
	err = ledger.Record(fromAccountID, "usdc", decimal.NewFromInt(1000), nil)
	require.NoError(t, err)
	err = ledger.Record(fromAccountID, "eth", decimal.NewFromInt(5), nil)
	require.NoError(t, err)

	// Test 1: Transfer within spending cap should succeed
	transferParams := TransferParams{
		Destination: recipientAddress,
		Allocations: []TransferAllocation{
			{AssetSymbol: "usdc", Amount: decimal.NewFromInt(300)},
		},
	}

	ctx := createSignedRPCContext(1, "transfer", transferParams, sessionSigner)
	// Set UserID to the main wallet address, not the session key address
	ctx.UserID = walletAddress
	router.HandleTransfer(ctx)

	res := assertResponse(t, ctx, "transfer")
	transferResp, ok := res.Params.(TransferResponse)
	require.True(t, ok)
	require.Len(t, transferResp.Transactions, 1)

	// Test 2: Verify spending was recorded
	spending, err := CalculateSessionKeySpending(db, sessionKeyAddress, "usdc")
	require.NoError(t, err)
	assert.Equal(t, "300", spending.String())

	// Test 3: Another transfer that would exceed cap should fail
	transferParams2 := TransferParams{
		Destination: recipientAddress,
		Allocations: []TransferAllocation{
			{AssetSymbol: "usdc", Amount: decimal.NewFromInt(300)}, // This would make total 600, exceeding 500 cap
		},
	}

	ctx2 := createSignedRPCContext(2, "transfer", transferParams2, sessionSigner)

	ctx2.UserID = walletAddress
	router.HandleTransfer(ctx2)

	assertErrorResponse(t, ctx2, "operation denied: insufficient session key allowance")

	// Test 4: Transfer with unauthorized asset should fail
	transferParams3 := TransferParams{
		Destination: recipientAddress,
		Allocations: []TransferAllocation{
			{AssetSymbol: "btc", Amount: decimal.NewFromInt(1)}, // BTC not in allowances
		},
	}

	ctx3 := createSignedRPCContext(3, "transfer", transferParams3, sessionSigner)
	// Set UserID to the main wallet address, not the session key address
	ctx3.UserID = walletAddress
	router.HandleTransfer(ctx3)

	assertErrorResponse(t, ctx3, "not allowed in session key spending cap")
}

func TestUnsupportedAssetValidation(t *testing.T) {
	assetsCfg := &AssetsConfig{
		Assets: []AssetConfig{
			{
				Symbol: "usdc",
				Name:   "USD Coin",
				Tokens: []TokenConfig{
					{BlockchainID: 1, Address: "0xA0b86991c431e803859e9c5092D6B0a2a22B6e", Decimals: 6},
				},
			},
			{
				Symbol: "eth",
				Name:   "Ethereum",
				Tokens: []TokenConfig{
					{BlockchainID: 1, Address: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", Decimals: 18},
				},
			},
		},
	}

	// Test 1: Supported assets should pass validation
	supportedAllowances := []Allowance{
		{Asset: "usdc", Amount: "1000"},
		{Asset: "eth", Amount: "5"},
	}
	err := validateAllowances(assetsCfg, supportedAllowances)
	assert.NoError(t, err, "Should accept supported assets")

	// Test 2: Empty allowances should pass validation
	err = validateAllowances(assetsCfg, []Allowance{})
	assert.NoError(t, err, "Should accept empty allowances")

	// Test 3: Unsupported asset should fail validation
	unsupportedAllowances := []Allowance{
		{Asset: "usdc", Amount: "1000"}, // supported
		{Asset: "btc", Amount: "1"},     // unsupported
	}
	err = validateAllowances(assetsCfg, unsupportedAllowances)
	assert.Error(t, err, "Should reject unsupported assets")
	assert.Contains(t, err.Error(), "asset 'btc' is not supported")

	// Test 4: Zero amount should pass validation (0 is allowed)
	zeroAllowances := []Allowance{
		{Asset: "usdc", Amount: "0"},
	}
	err = validateAllowances(assetsCfg, zeroAllowances)
	assert.NoError(t, err, "Should accept zero amounts")

	// Test 5: Negative amount should fail validation
	negativeAllowances := []Allowance{
		{Asset: "usdc", Amount: "-100"},
	}
	err = validateAllowances(assetsCfg, negativeAllowances)
	assert.Error(t, err, "Should reject negative amounts")
	assert.Contains(t, err.Error(), "allowance amount cannot be negative")

	// Test 6: Invalid decimal format should fail validation
	invalidAllowances := []Allowance{
		{Asset: "usdc", Amount: "not-a-number"},
	}
	err = validateAllowances(assetsCfg, invalidAllowances)
	assert.Error(t, err, "Should reject invalid decimal format")
	assert.Contains(t, err.Error(), "invalid amount")
}

func TestOneSessionKeyPerApp(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	walletAddress := "0x742d35Cc6435C0532fd5c5fdb1d1d2B4E5b6a6Ad"
	sessionKey1 := "0x8ba1f109551bD432803012645Hac136c9SessionKey1"
	sessionKey2 := "0x8ba1f109551bD432803012645Hac136c9SessionKey2"
	app := "TestApp"

	allowances := []Allowance{
		{Asset: "usdc", Amount: "500"},
	}
	expiration := time.Now().Add(24 * time.Hour)

	err := loadSessionKeyCache(db)
	require.NoError(t, err)

	err = AddSessionKey(db, walletAddress, sessionKey1, app, "trade", allowances, expiration)
	require.NoError(t, err)

	assert.Equal(t, walletAddress, GetWalletBySessionKey(sessionKey1))

	sessionKeys, err := GetSessionKeysByWallet(db, walletAddress)
	require.NoError(t, err)
	assert.Len(t, sessionKeys, 1)
	assert.Equal(t, sessionKey1, sessionKeys[0].Address)
	assert.Equal(t, app, sessionKeys[0].Application)

	// Add second session key for the SAME app - should invalidate the first one
	err = AddSessionKey(db, walletAddress, sessionKey2, app, "trade", allowances, expiration)
	require.NoError(t, err)

	// Verify second session key exists and is cached
	assert.Equal(t, walletAddress, GetWalletBySessionKey(sessionKey2))

	// Verify first session key is no longer cached
	assert.Equal(t, "", GetWalletBySessionKey(sessionKey1))

	// Verify only second session key exists in database
	sessionKeys, err = GetSessionKeysByWallet(db, walletAddress)
	require.NoError(t, err)
	assert.Len(t, sessionKeys, 1, "Should only have one session key per app")
	assert.Equal(t, sessionKey2, sessionKeys[0].Address)
	assert.Equal(t, app, sessionKeys[0].Application)

	// Test that different apps can have different session keys
	sessionKey3 := "0x8ba1f109551bD432803012645Hac136c9SessionKey3"
	differentApp := "DifferentApp"

	err = AddSessionKey(db, walletAddress, sessionKey3, differentApp, "trade", allowances, expiration)
	require.NoError(t, err)

	// Verify both session keys exist (different apps)
	sessionKeys, err = GetSessionKeysByWallet(db, walletAddress)
	require.NoError(t, err)
	assert.Len(t, sessionKeys, 2, "Different apps should allow separate session keys")

	assert.Equal(t, walletAddress, GetWalletBySessionKey(sessionKey2))
	assert.Equal(t, walletAddress, GetWalletBySessionKey(sessionKey3))
}
