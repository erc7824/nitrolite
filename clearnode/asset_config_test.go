package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAssetsConfig_verifyVariables tests the validation logic for asset configuration
func TestAssetsConfig_verifyVariables(t *testing.T) {
	// Test missing asset symbol
	t.Run("missing asset symbol", func(t *testing.T) {
		cfg := AssetsConfig{
			Assets: []AssetConfig{
				{
					Symbol: "", // Missing symbol
				},
			},
		}
		err := cfg.verifyVariables()
		require.Error(t, err)
		assert.Equal(t, "missing asset symbol for asset[0]", err.Error())
	})

	// Test missing token address
	t.Run("missing token address", func(t *testing.T) {
		cfg := AssetsConfig{
			Assets: []AssetConfig{
				{
					Name:   "USD Coin",
					Symbol: "USDC",
					Tokens: []TokenConfig{
						{
							Name:         "USD Coin",
							Symbol:       "USDC",
							BlockchainID: 1,
							Address:      "", // Missing address
						},
					},
				},
			},
		}
		err := cfg.verifyVariables()
		require.Error(t, err)
		assert.Equal(t, "missing USD Coin token address for blockchain with id 1", err.Error())
	})

	// Test invalid token address
	t.Run("invalid token address", func(t *testing.T) {
		cfg := AssetsConfig{
			Assets: []AssetConfig{
				{
					Name:   "USD Coin",
					Symbol: "USDC",
					Tokens: []TokenConfig{
						{
							Name:         "USD Coin",
							Symbol:       "USDC",
							BlockchainID: 1,
							Address:      "0xinvalid", // Invalid address
						},
					},
				},
			},
		}
		err := cfg.verifyVariables()
		require.Error(t, err)
		assert.Equal(t, "invalid USD Coin token address '0xinvalid' for blockchain with id 1", err.Error())
	})

	// Test custom symbol for token (inherits from asset when empty)
	t.Run("custom symbol for token", func(t *testing.T) {
		cfg := AssetsConfig{
			Assets: []AssetConfig{
				{
					Name:   "USD Coin",
					Symbol: "USDC",
					Tokens: []TokenConfig{
						{
							Name:         "Bridged USDC",
							Symbol:       "", // Should inherit "USDC" from asset
							BlockchainID: 137,
							Address:      "0x2791bca1f2de4661ed88a30c99a7a9449aa84174",
						},
					},
				},
			},
		}
		err := cfg.verifyVariables()
		require.NoError(t, err)
		assert.Equal(t, "USDC", cfg.Assets[0].Tokens[0].Symbol)
	})
}

// TestAssetsConfig_GetAssetTokenByAddressAndChainID tests the asset token lookup
func TestAssetsConfig_GetAssetTokenByAddressAndChainID(t *testing.T) {
	// Setup test configuration
	cfg := AssetsConfig{
		Assets: []AssetConfig{
			{
				Name:     "USD Coin",
				Symbol:   "USDC",
				Disabled: false,
				Tokens: []TokenConfig{
					{
						Name:         "USD Coin",
						Symbol:       "USDC",
						BlockchainID: 1,
						Disabled:     false,
						Address:      "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
						Decimals:     6,
					},
					{
						Name:         "USD Coin",
						Symbol:       "USDC",
						BlockchainID: 137,
						Disabled:     true, // Disabled token
						Address:      "0x2791bca1f2de4661ed88a30c99a7a9449aa84174",
						Decimals:     6,
					},
				},
			},
			{
				Name:     "Tether",
				Symbol:   "USDT",
				Disabled: true, // Disabled asset
				Tokens: []TokenConfig{
					{
						Name:         "Tether",
						Symbol:       "USDT",
						BlockchainID: 1,
						Disabled:     false,
						Address:      "0xdac17f958d2ee523a2206206994597c13d831ec7",
						Decimals:     6,
					},
				},
			},
		},
	}

	// Test not found
	t.Run("not found", func(t *testing.T) {
		result, found := cfg.GetAssetTokenByAddressAndChainID("0x0000000000000000000000000000000000000000", 1)
		assert.False(t, found)
		assert.Equal(t, AssetTokenConfig{}, result)
	})

	// Test token disabled
	t.Run("token disabled", func(t *testing.T) {
		result, found := cfg.GetAssetTokenByAddressAndChainID("0x2791bca1f2de4661ed88a30c99a7a9449aa84174", 137)
		assert.False(t, found)
		assert.Equal(t, AssetTokenConfig{}, result)
	})

	// Test asset disabled
	t.Run("asset disabled", func(t *testing.T) {
		result, found := cfg.GetAssetTokenByAddressAndChainID("0xdac17f958d2ee523a2206206994597c13d831ec7", 1)
		assert.False(t, found)
		assert.Equal(t, AssetTokenConfig{}, result)
	})

	// Test all good - verify returned data
	t.Run("all good", func(t *testing.T) {
		result, found := cfg.GetAssetTokenByAddressAndChainID("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", 1)
		assert.True(t, found)

		// Verify asset-level data
		assert.Equal(t, "USD Coin", result.Name)
		assert.Equal(t, "USDC", result.Symbol)
		assert.False(t, result.Disabled)

		// Verify token-level data
		assert.Equal(t, "USD Coin", result.Token.Name)
		assert.Equal(t, "USDC", result.Token.Symbol)
		assert.Equal(t, uint32(1), result.Token.BlockchainID)
		assert.False(t, result.Token.Disabled)
		assert.Equal(t, "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", result.Token.Address)
		assert.Equal(t, uint8(6), result.Token.Decimals)
	})
}

// TestAssetsConfig_GetAssetTokensByChainID tests retrieving all tokens for a blockchain
func TestAssetsConfig_GetAssetTokensByChainID(t *testing.T) {
	// Setup test configuration
	cfg := AssetsConfig{
		Assets: []AssetConfig{
			{
				Name:     "USD Coin",
				Symbol:   "USDC",
				Disabled: false,
				Tokens: []TokenConfig{
					{
						Name:         "USD Coin",
						Symbol:       "USDC",
						BlockchainID: 1,
						Disabled:     false,
						Address:      "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
						Decimals:     6,
					},
					{
						Name:         "USD Coin",
						Symbol:       "USDC",
						BlockchainID: 1,
						Disabled:     true, // Disabled token on same chain
						Address:      "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb49",
						Decimals:     6,
					},
				},
			},
			{
				Name:     "Tether",
				Symbol:   "USDT",
				Disabled: true, // Disabled asset
				Tokens: []TokenConfig{
					{
						Name:         "Tether",
						Symbol:       "USDT",
						BlockchainID: 1,
						Disabled:     false,
						Address:      "0xdac17f958d2ee523a2206206994597c13d831ec7",
						Decimals:     6,
					},
				},
			},
			{
				Name:     "WETH",
				Symbol:   "WETH",
				Disabled: false,
				Tokens: []TokenConfig{
					{
						Name:         "Wrapped Ether",
						Symbol:       "WETH",
						BlockchainID: 1,
						Disabled:     false,
						Address:      "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
						Decimals:     18,
					},
				},
			},
		},
	}

	// Test not found (empty result for non-existent chain)
	t.Run("not found", func(t *testing.T) {
		tokens := cfg.GetAssetTokensByChainID(999)
		assert.Empty(t, tokens)
	})

	// Test with disabled tokens and assets filtered out
	t.Run("filters disabled", func(t *testing.T) {
		tokens := cfg.GetAssetTokensByChainID(1)

		// Should only get 2 tokens: USDC and WETH (not disabled USDC variant or USDT from disabled asset)
		assert.Len(t, tokens, 2)

		// Verify first token (USDC)
		assert.Equal(t, "USD Coin", tokens[0].Name)
		assert.Equal(t, "USDC", tokens[0].Symbol)
		assert.False(t, tokens[0].Disabled)
		assert.Equal(t, "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", tokens[0].Token.Address)

		// Verify second token (WETH)
		assert.Equal(t, "WETH", tokens[1].Name)
		assert.Equal(t, "WETH", tokens[1].Symbol)
		assert.False(t, tokens[1].Disabled)
		assert.Equal(t, "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", tokens[1].Token.Address)
	})
}
