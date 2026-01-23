package memory

import (
	"fmt"
	"slices"

	"github.com/erc7824/nitrolite/pkg/core"
)

type MemoryStoreV1 struct {
	blockchains     []core.Blockchain
	assets          []core.Asset
	supportedAssets map[string]map[uint32]map[string]struct{} // map[asset]map[blockchain_id]map[token_address]struct{}
	tokenDecimals   map[uint32]map[string]uint8               // map[blockchain_id]map[token_address]decimals
	assetDecimals   map[string]uint8                          // map[asset]decimals
}

func NewMemoryStoreV1(assetsConfig AssetsConfig, blockchainsConfig map[uint32]BlockchainConfig) MemoryStore {
	supportedBlockchainIDs := make(map[uint32]struct{})
	blockchains := make([]core.Blockchain, 0, len(blockchainsConfig))
	for _, bc := range blockchainsConfig {
		if bc.Disabled {
			continue
		}
		supportedBlockchainIDs[bc.ID] = struct{}{}

		blockchains = append(blockchains, core.Blockchain{
			ID:              bc.ID,
			Name:            bc.Name,
			ContractAddress: bc.ContractAddress,
		})
	}
	slices.SortFunc(blockchains, func(a, b core.Blockchain) int {
		if a.ID < b.ID {
			return -1
		} else if a.ID > b.ID {
			return 1
		}
		return 0
	})

	supportedAssets := make(map[string]map[uint32]map[string]struct{})
	tokenDecimals := make(map[uint32]map[string]uint8)
	assetDecimals := make(map[string]uint8)
	assets := make([]core.Asset, 0, len(assetsConfig.Assets))
	for _, asset := range assetsConfig.Assets {
		if asset.Disabled {
			continue
		}

		tokens := make([]core.Token, 0, len(asset.Tokens))
		for _, token := range asset.Tokens {
			if _, ok := supportedBlockchainIDs[token.BlockchainID]; !ok {
				continue
			}
			if token.Disabled {
				continue
			}

			tokens = append(tokens, core.Token{
				Name:         token.Name,
				Symbol:       token.Symbol,
				Address:      token.Address,
				BlockchainID: token.BlockchainID,
				Decimals:     token.Decimals,
			})

			if _, ok := supportedAssets[asset.Symbol]; !ok {
				supportedAssets[asset.Symbol] = make(map[uint32]map[string]struct{})
			}
			if _, ok := supportedAssets[asset.Symbol][token.BlockchainID]; !ok {
				supportedAssets[asset.Symbol][token.BlockchainID] = make(map[string]struct{})
			}
			supportedAssets[asset.Symbol][token.BlockchainID][token.Address] = struct{}{}

			if _, ok := tokenDecimals[token.BlockchainID]; !ok {
				tokenDecimals[token.BlockchainID] = make(map[string]uint8)
			}
			tokenDecimals[token.BlockchainID][token.Address] = token.Decimals
		}
		if len(tokens) == 0 {
			continue
		}

		slices.SortFunc(tokens, func(a, b core.Token) int {
			if a.BlockchainID < b.BlockchainID {
				return -1
			} else if a.BlockchainID > b.BlockchainID {
				return 1
			}
			return 0
		})

		assets = append(assets, core.Asset{
			Symbol:   asset.Symbol,
			Name:     asset.Name,
			Decimals: asset.Decimals,
			Tokens:   tokens,
		})

		assetDecimals[asset.Symbol] = asset.Decimals
	}

	slices.SortFunc(assets, func(a, b core.Asset) int {
		if a.Symbol < b.Symbol {
			return -1
		} else if a.Symbol > b.Symbol {
			return 1
		}
		return 0
	})

	return &MemoryStoreV1{
		blockchains:     blockchains,
		assets:          assets,
		supportedAssets: supportedAssets,
		tokenDecimals:   tokenDecimals,
		assetDecimals:   assetDecimals,
	}
}

func NewMemoryStoreV1FromConfig(configDirPath string) (MemoryStore, error) {
	blockchainConfig, err := LoadEnabledBlockchains(configDirPath)
	if err != nil {
		return nil, err
	}
	assetsConfig, err := LoadAssets(configDirPath)
	if err != nil {
		return nil, err
	}
	return NewMemoryStoreV1(assetsConfig, blockchainConfig), nil
}

// GetBlockchains retrieves the list of supported blockchains.
func (ms *MemoryStoreV1) GetBlockchains() ([]core.Blockchain, error) {
	return ms.blockchains, nil
}

// GetAssets retrieves the list of supported assets.
// If blockchainID is provided, filters assets to only include tokens on that blockchain.
func (ms *MemoryStoreV1) GetAssets(blockchainID *uint32) ([]core.Asset, error) {
	if blockchainID == nil {
		return ms.assets, nil
	}

	filteredAssets := make([]core.Asset, 0)
	for _, asset := range ms.assets {
		filteredTokens := make([]core.Token, 0)
		for _, token := range asset.Tokens {
			if token.BlockchainID == *blockchainID {
				filteredTokens = append(filteredTokens, token)
			}
		}

		if len(filteredTokens) > 0 {
			filteredAsset := asset
			filteredAsset.Tokens = filteredTokens
			filteredAssets = append(filteredAssets, filteredAsset)
		}
	}

	return filteredAssets, nil
}

// IsAssetSupported checks if a given asset (token) is supported on the specified blockchain.
func (ms *MemoryStoreV1) IsAssetSupported(asset, tokenAddress string, blockchainID uint32) (bool, error) {
	assetsOnChain, ok := ms.supportedAssets[asset]
	if !ok {
		return false, nil
	}
	tokensOnChain, ok := assetsOnChain[blockchainID]
	if !ok {
		return false, nil
	}
	_, supported := tokensOnChain[tokenAddress]
	return supported, nil
}

// GetAssetDecimals checks if an asset exists and returns its decimals in YN
func (ms *MemoryStoreV1) GetAssetDecimals(asset string) (uint8, error) {
	decimals, ok := ms.assetDecimals[asset]
	if !ok {
		return 0, fmt.Errorf("asset '%s' is not supported", asset)
	}
	return decimals, nil
}

// GetTokenDecimals returns the decimals for a token on a specific blockchain
func (ms *MemoryStoreV1) GetTokenDecimals(blockchainID uint32, tokenAddress string) (uint8, error) {
	decimalsOnChain, ok := ms.tokenDecimals[blockchainID]
	if !ok {
		return 0, fmt.Errorf("blockchain with ID '%d' is not supported", blockchainID)
	}
	decimals, ok := decimalsOnChain[tokenAddress]
	if !ok {
		return 0, fmt.Errorf("token %s is not supported on blockchain with ID '%d'", tokenAddress, blockchainID)
	}
	return decimals, nil
}
