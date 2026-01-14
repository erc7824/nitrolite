package app_session_v1

import (
	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
)

// AppStoreV1 defines the persistence layer interface for app session management.
type AppStoreV1 interface {
	// App session operations
	CreateAppSession(session app.AppSessionV1) error
	GetAppSession(sessionID string, isClosed bool) (*app.AppSessionV1, error)
	UpdateAppSession(session app.AppSessionV1) error
	GetAppSessionBalances(sessionID string) (map[string]decimal.Decimal, error)
	GetParticipantAllocations(sessionID string) (map[string]map[string]decimal.Decimal, error)

	// Ledger operations
	RecordLedgerEntry(accountID, asset string, amount decimal.Decimal, sessionKey *string) error
	GetAccountBalance(accountID, asset string) (decimal.Decimal, error)

	RecordTransaction(tx core.Transaction) error

	// Channel state operations

	// CheckOpenChannel verifies if a user has an active channel for the given asset.
	CheckOpenChannel(wallet, asset string) (bool, error)
	GetLastUserState(wallet, asset string, signed bool) (*core.State, error)
	StoreUserState(state core.State) error
	EnsureNoOngoingStateTransitions(wallet, asset string) error

	// TODO: add session keys support
	// Session key operations
	// GetActiveSessionKey(sessionKeyAddress string) (*app.SessionKeyV1, error)
	// ValidateSessionKeyApplication(sessionKey *app.SessionKeyV1, application string) error
	// ValidateSessionKeySpending(sessionKey *app.SessionKeyV1, asset string, amount decimal.Decimal) error
}

// StoreTxHandler is a function that executes Store operations within a transaction.
// If the handler returns an error, the transaction is rolled back; otherwise it's committed.
type StoreTxHandler func(AppStoreV1) error

// StoreTxProvider wraps Store operations in a database transaction.
// It accepts a StoreTxHandler and manages transaction lifecycle (begin, commit, rollback).
// Returns an error if the handler fails or the transaction cannot be committed.
type StoreTxProvider func(StoreTxHandler) error

// SigValidator validates cryptographic signatures on state transitions.
type SigValidator interface {
	// Recover recovers the wallet address from the signature and data.
	// Returns the recovered address or an error if the signature is invalid.
	Recover(data, sig []byte) (string, error)
	Verify(wallet string, data, sig []byte) error
}

// SigType identifies the signature validation algorithm to use.
type SigType string

// EcdsaSigType represents the ECDSA (Elliptic Curve Digital Signature Algorithm)
// validator, used for Ethereum-style signature verification.
const EcdsaSigType SigType = "ecdsa"
