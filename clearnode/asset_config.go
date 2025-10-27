package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	assetsFileName = "assets.yaml"
)

// AssetTokenConfig represents a combination of asset metadata and a specific token instance.
// This is returned by lookup functions to provide both asset-level and token-level information.
type AssetTokenConfig struct {
	// Name is the human-readable name of the asset (e.g., "USD Coin")
	Name string
	// Symbol is the ticker symbol for the asset (e.g., "USDC")
	Symbol string
	// Disabled determines if this asset should be processed
	Disabled bool
	// Token contains the blockchain-specific token implementation
	Token TokenConfig
}

// AssetsConfig represents the root configuration structure for all asset settings.
// It contains a list of assets, each of which can have multiple token representations
// across different blockchains.
type AssetsConfig struct {
	Assets []AssetConfig `yaml:"assets"`
}

// AssetConfig represents configuration for a single asset (e.g., USDC, USDT).
// An asset can have multiple token representations across different blockchains.
type AssetConfig struct {
	// Name is the human-readable name of the asset (e.g., "USD Coin")
	// If empty, it will inherit the Symbol value during validation
	Name string `yaml:"name"`
	// Symbol is the ticker symbol for the asset (e.g., "USDC")
	// This field is required for enabled assets
	Symbol string `yaml:"symbol"`
	// Disabled determines if this asset should be processed
	Disabled bool `yaml:"disabled"`
	// Tokens contains the blockchain-specific token implementations
	Tokens []TokenConfig `yaml:"tokens"`
}

// TokenConfig represents a specific token implementation on a blockchain.
// Each token is associated with a parent asset and deployed on a specific blockchain.
type TokenConfig struct {
	// Name is the token name on this blockchain (e.g., "Bridged USDC")
	// If empty, it will inherit from the parent asset's Name
	Name string `yaml:"name"`
	// Symbol is the token symbol on this blockchain
	// If empty, it will inherit from the parent asset's Symbol
	Symbol string `yaml:"symbol"`
	// BlockchainID is the chain ID where this token is deployed
	BlockchainID uint32 `yaml:"blockchain_id"`
	// Disabled determines if this token should be processed
	Disabled bool `yaml:"disabled"`
	// Address is the token's contract address on the blockchain
	// Must be a valid Ethereum address (0x followed by 40 hex characters)
	Address string `yaml:"address"`
	// Decimals is the number of decimal places for the token
	Decimals uint8 `yaml:"decimals"`
}

// LoadAssets loads and validates asset configurations from a YAML file.
// It reads from <configDirPath>/assets.yaml, validates all settings,
// and returns the parsed configuration.
//
// The function performs the following validations:
// - Asset symbols are required for enabled assets
// - Token addresses must be valid Ethereum addresses
// - Inheritance of names and symbols from asset to token level
func LoadAssets(configDirPath string) (AssetsConfig, error) {
	assetsPath := filepath.Join(configDirPath, assetsFileName)
	f, err := os.Open(assetsPath)
	if err != nil {
		return AssetsConfig{}, err
	}
	defer f.Close()

	var cfg AssetsConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return AssetsConfig{}, err
	}

	if err := cfg.verifyVariables(); err != nil {
		return AssetsConfig{}, err
	}

	return cfg, nil
}

// verifyVariables validates the configuration structure and applies defaults.
// It ensures all required fields are present and valid:
// - Asset symbols are required for enabled assets
// - Asset names default to symbols if not specified
// - Token names and symbols inherit from parent asset if not specified
// - Token addresses must be valid Ethereum addresses for enabled tokens
// Note: This method modifies token fields during validation but changes are not persisted
// back to the original slice due to Go's value semantics in range loops.
func (cfg *AssetsConfig) verifyVariables() error {
	for i, asset := range cfg.Assets {
		if asset.Disabled {
			continue
		}

		if asset.Symbol == "" {
			return fmt.Errorf("missing asset symbol for asset[%d]", i)
		}
		if asset.Name == "" {
			cfg.Assets[i].Name = asset.Symbol
		}

		asset = cfg.Assets[i]
		for j, token := range asset.Tokens {
			if token.Disabled {
				continue
			}

			if token.Symbol == "" {
				cfg.Assets[i].Tokens[j].Symbol = asset.Symbol
			}
			if token.Name == "" {
				cfg.Assets[i].Tokens[j].Name = asset.Name
			}

			token = cfg.Assets[i].Tokens[j]
			if token.Address == "" {
				return fmt.Errorf("missing %s token address for blockchain with id %d", token.Name, token.BlockchainID)
			} else if !contractAddressRegex.MatchString(token.Address) {
				return fmt.Errorf("invalid %s token address '%s' for blockchain with id %d", token.Name, token.Address, token.BlockchainID)
			}
		}
	}

	return nil
}

// GetAssetTokenByAddressAndChainID looks up a token by its contract address and blockchain ID.
// It returns an AssetTokenConfig containing both asset and token information if found.
// The second return value indicates whether the token was found.
// Only enabled assets and tokens are considered in the search.
func (cfg AssetsConfig) GetAssetTokenByAddressAndChainID(tokenAddress string, chainID uint32) (AssetTokenConfig, bool) {
	for _, asset := range cfg.Assets {
		if asset.Disabled {
			continue
		}

		for _, token := range asset.Tokens {
			if token.Disabled {
				continue
			}

			if token.BlockchainID == chainID && strings.EqualFold(token.Address, tokenAddress) {
				return AssetTokenConfig{
					Name:   asset.Name,
					Symbol: asset.Symbol,
					Token:  token,
				}, true
			}
		}
	}

	return AssetTokenConfig{}, false
}

// GetAssetTokensByChainID returns all enabled tokens deployed on the specified blockchain.
// The result includes both asset-level and token-level information for each token.
// Only tokens from enabled assets with enabled token entries are included.
func (cfg AssetsConfig) GetAssetTokensByChainID(chainID uint32) []AssetTokenConfig {
	var tokens []AssetTokenConfig
	for _, asset := range cfg.Assets {
		if asset.Disabled {
			continue
		}

		for _, token := range asset.Tokens {
			if !token.Disabled && token.BlockchainID == chainID {
				tokens = append(tokens, AssetTokenConfig{
					Name:   asset.Name,
					Symbol: asset.Symbol,
					Token:  token,
				})
			}
		}
	}
	return tokens
}
