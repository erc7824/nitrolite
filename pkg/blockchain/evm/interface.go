package evm

import (
	"context"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/ethereum/go-ethereum/core/types"
)

type HandleEvent func(ctx context.Context, eventLog types.Log)
type StoreContractEvent func(ev core.BlockchainEvent) error
type LatestEventGetter func(contractAddress string, blockchainID uint64) (ev core.BlockchainEvent, err error)

type AssetStore interface {
	// GetAssetDecimals checks if an asset exists and returns its decimals in YN
	GetAssetDecimals(asset string) (uint8, error)

	// GetTokenDecimals returns the decimals for a token on a specific blockchain
	GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error)
}
