package database

import (
	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
)

// StoreTxHandler is a function that executes Store operations within a transaction.
type StoreTxHandler func(DatabaseStore) error

// DatabaseStore defines the unified persistence layer.
type DatabaseStore interface {
	// ExecuteInTransaction runs the provided handler within a database transaction.
	// If the handler returns an error, the transaction is rolled back.
	// If the handler completes successfully, the transaction is committed.
	ExecuteInTransaction(handler StoreTxHandler) error

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
	EnsureNoOngoingStateTransitions(wallet, asset string) error

	// --- Blockchain Action Operations ---

	// ScheduleInitiateEscrowWithdrawal queues a blockchain action to initiate withdrawal.
	// This queues the state to be submitted on-chain to initiate an escrow withdrawal.
	ScheduleInitiateEscrowWithdrawal(stateID string, chainID uint64) error

	// ScheduleCheckpoint schedules a checkpoint operation for a home channel state.
	// This queues the state to be submitted on-chain to update the channel's on-chain state.
	ScheduleCheckpoint(stateID string, chainID uint64) error

	// ScheduleFinalizeEscrowDeposit schedules a checkpoint for an escrow deposit operation.
	// This queues the state to be submitted on-chain to finalize an escrow deposit.
	ScheduleFinalizeEscrowDeposit(stateID string, chainID uint64) error

	// ScheduleFinalizeEscrowWithdrawal schedules a checkpoint for an escrow withdrawal operation.
	// This queues the state to be submitted on-chain to finalize an escrow withdrawal.
	ScheduleFinalizeEscrowWithdrawal(stateID string, chainID uint64) error

	// ScheduleInitiateEscrowDeposit schedules a checkpoint for an escrow deposit operation.
	// This queues the state to be submitted on-chain for an escrow deposit on home chain.
	ScheduleInitiateEscrowDeposit(stateID string, chainID uint64) error

	// Fail marks a blockchain action as failed and increments the retry counter.
	Fail(actionID int64, err string) error

	// FailNoRetry marks a blockchain action as failed without incrementing the retry counter.
	FailNoRetry(actionID int64, err string) error

	// RecordAttempt records a failed attempt for a blockchain action and increments the retry counter.
	// The action remains in pending status.
	RecordAttempt(actionID int64, err string) error

	// Complete marks a blockchain action as completed with the given transaction hash.
	Complete(actionID int64, txHash string) error

	// GetActions retrieves pending blockchain actions, optionally limited by count.
	GetActions(limit uint8, chainID uint64) ([]BlockchainAction, error)

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

	// --- Contract Event Operations ---

	// StoreContractEvent stores a blockchain event to prevent duplicate processing.
	StoreContractEvent(ev core.BlockchainEvent) error

	// GetLatestEvent returns the latest block number and log index for a given contract.
	GetLatestEvent(contractAddress string, blockchainID uint64) (core.BlockchainEvent, error)
}
