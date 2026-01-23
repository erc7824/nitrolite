package database

import (
	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
)

// StoreTxHandler is a function that executes Store operations within a transaction.
type StoreTxHandler func(DatabaseStore) error

// StoreTxProvider wraps Store operations in a database transaction.
type StoreTxProvider func(StoreTxHandler) error

// DatabaseStore defines the unified persistence layer.
type DatabaseStore interface {
	// --- User & Balance Operations ---

	// GetUserBalances retrieves the balances for a user's wallet.
	GetUserBalances(wallet string) ([]core.BalanceEntry, error)

	// GetUserTransactions retrieves transaction history for a user with optional filters.
	GetUserTransactions(wallet string, asset *string, txType *core.TransactionType, fromTime *uint64, toTime *uint64, paginate *core.PaginationParams) ([]core.Transaction, core.PaginationMetadata, error)

	// RecordTransaction creates a transaction record linking state transitions.
	RecordTransaction(tx core.Transaction) error

	// --- Channel Operations ---

	// CreateChannel creates a new channel entity in the database.
	CreateChannel(channel core.Channel) error

	// GetChannelByID retrieves a channel by its unique identifier.
	GetChannelByID(channelID string) (*core.Channel, error)

	// GetActiveHomeChannel retrieves the active home channel for a user's wallet and asset.
	GetActiveHomeChannel(wallet, asset string) (*core.Channel, error)

	// CheckOpenChannel verifies if a user has an active channel for the given asset.
	CheckOpenChannel(wallet, asset string) (bool, error)

	// UpdateChannel persists changes to a channel's metadata (status, version, etc).
	UpdateChannel(channel core.Channel) error

	// --- State Management ---

	// GetLastStateByChannelID retrieves the most recent state for a given channel.
	// If signed is true, only returns states with both user and node signatures.
	GetLastStateByChannelID(channelID string, signed bool) (*core.State, error)

	// GetStateByChannelIDAndVersion retrieves a specific state version for a channel.
	// Returns nil if the state with the specified version does not exist.
	GetStateByChannelIDAndVersion(channelID string, version uint64) (*core.State, error)

	// GetLastUserState retrieves the most recent state for a user's asset.
	GetLastUserState(wallet, asset string, signed bool) (*core.State, error)

	// StoreUserState persists a new user state to the database.
	StoreUserState(state core.State) error

	// EnsureNoOngoingStateTransitions validates that no conflicting blockchain operations are pending.
	EnsureNoOngoingStateTransitions(wallet, asset string, prevTransitionType core.TransitionType) error

	// ScheduleInitiateEscrowWithdrawal queues a blockchain action to initiate withdrawal.
	ScheduleInitiateEscrowWithdrawal(stateID string) error

	ScheduleCheckpoint(stateID string) error

	// ScheduleCheckpointEscrowDeposit schedules a checkpoint for an escrow deposit operation.
	// This queues the state to be submitted on-chain to finalize an escrow deposit.
	ScheduleCheckpointEscrowDeposit(stateID string) error

	// ScheduleCheckpointEscrowWithdrawal schedules a checkpoint for an escrow withdrawal operation.
	// This queues the state to be submitted on-chain to finalize an escrow withdrawal.
	ScheduleCheckpointEscrowWithdrawal(stateID string) error

	// --- App Session Operations ---

	// CreateAppSession initializes a new application session.
	CreateAppSession(session app.AppSessionV1) error

	// GetAppSession retrieves a specific session by ID.
	GetAppSession(sessionID string) (*app.AppSessionV1, error)

	// GetAppSessions retrieves filtered sessions with pagination.
	GetAppSessions(appSessionID *string, participant *string, status app.AppSessionStatus, pagination *core.PaginationParams) ([]app.AppSessionV1, core.PaginationMetadata, error)

	// UpdateAppSession updates existing session data.
	UpdateAppSession(session app.AppSessionV1) error

	// --- App Ledger Operations ---

	// GetAppSessionBalances retrieves the total balances associated with a session.
	GetAppSessionBalances(sessionID string) (map[string]decimal.Decimal, error)

	// GetParticipantAllocations retrieves specific asset allocations per participant.
	GetParticipantAllocations(sessionID string) (map[string]map[string]decimal.Decimal, error)

	// RecordLedgerEntry logs a movement of funds within the internal ledger.
	RecordLedgerEntry(userWallet, accountID, asset string, amount decimal.Decimal) error
}
