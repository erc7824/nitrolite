package sdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/erc7824/nitrolite/pkg/core"
)

// clientAssetStore implements core.AssetStore by fetching data from the Clearnode API.
type clientAssetStore struct {
	client *Client
	cache  map[string]core.Asset // asset symbol -> Asset
}

func newClientAssetStore(client *Client) *clientAssetStore {
	return &clientAssetStore{
		client: client,
		cache:  make(map[string]core.Asset),
	}
}

// GetAssetDecimals returns the decimals for an asset as stored in Clearnode.
func (s *clientAssetStore) GetAssetDecimals(asset string) (uint8, error) {
	// Check cache first
	if cached, ok := s.cache[asset]; ok {
		return cached.Decimals, nil
	}

	// Fetch from node
	assets, err := s.client.GetAssets(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch assets: %w", err)
	}

	// Update cache and find asset
	for _, a := range assets {
		s.cache[a.Symbol] = a
		if strings.EqualFold(a.Symbol, asset) {
			return a.Decimals, nil
		}
	}

	return 0, fmt.Errorf("asset %s not found", asset)
}

// GetTokenDecimals returns the decimals for a specific token on a blockchain.
func (s *clientAssetStore) GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error) {
	// Fetch all assets if cache is empty
	if len(s.cache) == 0 {
		assets, err := s.client.GetAssets(context.Background(), nil)
		if err != nil {
			return 0, fmt.Errorf("failed to fetch assets: %w", err)
		}
		for _, a := range assets {
			s.cache[a.Symbol] = a
		}
	}

	// Search through all assets for matching token
	tokenAddress = strings.ToLower(tokenAddress)
	for _, asset := range s.cache {
		for _, token := range asset.Tokens {
			if token.BlockchainID == blockchainID &&
				strings.EqualFold(token.Address, tokenAddress) {
				return token.Decimals, nil
			}
		}
	}

	return 0, fmt.Errorf("token %s on blockchain %d not found", tokenAddress, blockchainID)
}

// GetTokenAddress returns the token address for a given asset on a specific blockchain.
func (s *clientAssetStore) GetTokenAddress(asset string, blockchainID uint64) (string, error) {
	// Fetch all assets if cache is empty
	if len(s.cache) == 0 {
		assets, err := s.client.GetAssets(context.Background(), nil)
		if err != nil {
			return "", fmt.Errorf("failed to fetch assets: %w", err)
		}
		for _, a := range assets {
			s.cache[a.Symbol] = a
		}
	}

	// Search for the asset and its token on the specified blockchain
	for _, a := range s.cache {
		if strings.EqualFold(a.Symbol, asset) {
			for _, token := range a.Tokens {
				if token.BlockchainID == blockchainID {
					return token.Address, nil
				}
			}
			return "", fmt.Errorf("asset %s not available on blockchain %d", asset, blockchainID)
		}
	}

	// Asset not found in cache, try fetching again
	assets, err := s.client.GetAssets(context.Background(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch assets: %w", err)
	}

	for _, a := range assets {
		s.cache[a.Symbol] = a
		if strings.EqualFold(a.Symbol, asset) {
			for _, token := range a.Tokens {
				if token.BlockchainID == blockchainID {
					return token.Address, nil
				}
			}
			return "", fmt.Errorf("asset %s not available on blockchain %d", asset, blockchainID)
		}
	}

	return "", fmt.Errorf("asset %s not found", asset)
}

// AssetExistsOnBlockchain checks if a specific asset is supported on a specific blockchain.
func (s *clientAssetStore) AssetExistsOnBlockchain(blockchainID uint64, asset string) (bool, error) {
	for _, a := range s.cache {
		if strings.EqualFold(a.Symbol, asset) {
			for _, token := range a.Tokens {
				if token.BlockchainID == blockchainID {
					return true, nil
				}
			}
			// Asset found in cache, but not on this chain
			return false, nil
		}
	}

	assets, err := s.client.GetAssets(context.Background(), nil)
	if err != nil {
		return false, fmt.Errorf("failed to fetch assets: %w", err)
	}

	for _, a := range assets {
		s.cache[a.Symbol] = a
		if strings.EqualFold(a.Symbol, asset) {
			for _, token := range a.Tokens {
				if token.BlockchainID == blockchainID {
					return true, nil
				}
			}
			// Asset found after fetch, but not on this chain
			return false, nil
		}
	}

	// Asset symbol not found at all
	return false, nil
}
