package memory

import (
	"github.com/erc7824/nitrolite/pkg/core"
)

// MemoryStore defines an in-memory data store interface for retrieving
// supported blockchains and assets.
type MemoryStore interface {
	// GetBlockchains retrieves the list of supported blockchains.
	GetBlockchains() ([]core.Blockchain, error)

	// GetAssets retrieves the list of supported assets.
	// If blockchainID is provided, filters assets to only include tokens on that blockchain.
	GetAssets(blockchainID *uint32) ([]core.Asset, error)

	// IsAssetSupported checks if a given asset (token) is supported on the specified blockchain.
	IsAssetSupported(asset, tokenAddress string, blockchainID uint32) (bool, error)

	// GetAssetDecimals checks if an asset exists and returns its decimals in YN
	GetAssetDecimals(asset string) (uint8, error)

	// GetTokenDecimals returns the decimals for a token on a specific blockchain
	GetTokenDecimals(blockchainID uint32, tokenAddress string) (uint8, error)
}
