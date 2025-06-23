package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAssets(t *testing.T) {
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
