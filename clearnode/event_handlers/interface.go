package event_handlers

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

	// RecordTransaction creates a transaction record linking state transitions
	// to track the history of operations (deposits, withdrawals, transfers, etc.).
	RecordTransaction(tx core.Transaction) error

	// UpdateChannel updates an existing channel entity in the database.
	UpdateChannel(channel core.Channel) error

	// GetChannelByID retrieves a channel by its unique identifier.
	// Returns nil if the channel doesn't exist.
	GetChannelByID(channelID string) (*core.Channel, error)

	// GetActiveHomeChannel retrieves the active home channel for a user's wallet and asset.
	// Returns nil if no home channel exists for the given wallet and asset.
	GetActiveHomeChannel(wallet, asset string) (*core.Channel, error)

	ScheduleCheckpoint(ChannelID string, StateVersion uint64) error
}
