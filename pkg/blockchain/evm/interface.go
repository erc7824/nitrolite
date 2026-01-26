package evm

import (
	"context"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/ethereum/go-ethereum/core/types"
)

// StoreTxHandler is a function that executes Store operations within a transaction.
// If the handler returns an error, the transaction is rolled back; otherwise it's committed.
type StoreTxHandler func(Store) error

// StoreTxProvider wraps Store operations in a database transaction.
// It accepts a StoreTxHandler and manages transaction lifecycle (begin, commit, rollback).
// Returns an error if the handler fails or the transaction cannot be committed.
type StoreTxProvider func(StoreTxHandler) error

type Store interface {
	StoreContractEvent(ev core.BlockchainEvent, data any)
	// GetLatestContractEvent(contractAddress string, networkID uint32) (*ContractEvent, error)
}

type HandleEvent func(ctx context.Context, eventLog types.Log)

type AssetStore interface {
	// GetAssetDecimals checks if an asset exists and returns its decimals in YN
	GetAssetDecimals(asset string) (uint8, error)

	// GetTokenDecimals returns the decimals for a token on a specific blockchain
	GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error)
}
