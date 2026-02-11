package channel_v1

import (
	"github.com/erc7824/nitrolite/pkg/core"
)

// StoreTxHandler is a function that executes Store operations within a transaction.
// If the handler returns an error, the transaction is rolled back; otherwise it's committed.
type StoreTxHandler func(Store) error

// StoreTxProvider wraps Store operations in a database transaction.
// It accepts a StoreTxHandler and manages transaction lifecycle (begin, commit, rollback).
// Returns an error if the handler fails or the transaction cannot be committed.
type StoreTxProvider func(StoreTxHandler) error

// Store defines the persistence layer interface for channel state management.
// All methods should be implemented to work within database transactions.
type Store interface {
	// GetLastUserState retrieves the most recent state for a user's asset.
	// If signed is true, only returns states with both user and node signatures.
	// Returns nil state if no matching state exists.
	GetLastUserState(wallet, asset string, signed bool) (*core.State, error)

	// CheckOpenChannel verifies if a user has an active channel for the given asset.
	CheckOpenChannel(wallet, asset string) (bool, error)

	// StoreUserState persists a new user state to the database.
	StoreUserState(state core.State) error

	// EnsureNoOngoingStateTransitions validates that no blockchain operations are pending
	// that would conflict with submitting a new state transition.
	EnsureNoOngoingStateTransitions(wallet, asset string) error

	// ScheduleInitiateEscrowWithdrawal queues a blockchain action to initiate
	// withdrawal from an escrow channel (triggered by escrow_lock transition).
	ScheduleInitiateEscrowWithdrawal(stateID string, chainID uint64) error

	// RecordTransaction creates a transaction record linking state transitions
	// to track the history of operations (deposits, withdrawals, transfers, etc.).
	RecordTransaction(tx core.Transaction) error

	// CreateChannel creates a new channel entity in the database.
	// This is called during channel creation before the channel exists on-chain.
	// The channel starts with OnChainStateVersion=0 to indicate it's pending blockchain confirmation.
	CreateChannel(channel core.Channel) error

	// GetChannelByID retrieves a channel by its unique identifier.
	// Returns nil if the channel doesn't exist.
	GetChannelByID(channelID string) (*core.Channel, error)

	// GetActiveHomeChannel retrieves the active home channel for a user's wallet and asset.
	// Returns nil if no home channel exists for the given wallet and asset.
	GetActiveHomeChannel(wallet, asset string) (*core.Channel, error)
}

// SigValidator validates cryptographic signatures on state transitions.
type SigValidator interface {
	// Verify checks that the signature is valid for the given data and wallet address.
	// Returns an error if the signature is invalid or cannot be verified.
	Verify(wallet string, data, sig []byte) error
}

// SigValidatorType identifies the signature validation algorithm to use.
type SigValidatorType string

// EcdsaSigValidatorType represents the ECDSA (Elliptic Curve Digital Signature Algorithm)
// validator, used for Ethereum-style signature verification.
const EcdsaSigValidatorType SigValidatorType = "ecdsa"

type MemoryStore interface {
	// IsAssetSupported checks if a given asset (token) is supported on the specified blockchain.
	IsAssetSupported(asset, tokenAddress string, blockchainID uint64) (bool, error)

	// GetAssetDecimals checks if an asset exists and returns its decimals in YN
	GetAssetDecimals(asset string) (uint8, error)

	// GetTokenDecimals returns the decimals for a token on a specific blockchain
	GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error)
}
