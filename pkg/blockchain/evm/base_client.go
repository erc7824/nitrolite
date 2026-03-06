package evm

// BaseClient holds the shared fields for all EVM-based clients.
type BaseClient struct {
	evmClient    EVMClient
	blockchainID uint64
}
