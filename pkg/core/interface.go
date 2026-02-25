package core

import (
	"context"

	"github.com/shopspring/decimal"
)

// ========= Client Interface =========

// Client defines the interface for interacting with the ChannelsHub smart contract
// TODO: add context to all methods
type Client interface {
	// Getters - IVault
	GetAccountsBalances(accounts []string, tokens []string) ([][]decimal.Decimal, error)

	// Getters - ERC20
	GetTokenBalance(token string, account string) (decimal.Decimal, error)
	Approve(asset string, amount decimal.Decimal) (string, error)

	// Getters - ChannelsHub
	GetNodeBalance(token string) (decimal.Decimal, error)
	GetOpenChannels(user string) ([]string, error)
	GetHomeChannelData(homeChannelID string) (HomeChannelDataResponse, error)
	GetEscrowDepositData(escrowChannelID string) (EscrowDepositDataResponse, error)
	GetEscrowWithdrawalData(escrowChannelID string) (EscrowWithdrawalDataResponse, error)

	// IVault functions
	Deposit(node, token string, amount decimal.Decimal) (string, error)
	Withdraw(node, token string, amount decimal.Decimal) (string, error)

	// Channel lifecycle
	Create(def ChannelDefinition, initCCS State) (string, error)
	MigrateChannelHere(def ChannelDefinition, candidate State) (string, error)
	Checkpoint(candidate State) (string, error)
	Challenge(candidate State, challengerSig []byte, challengerIdx ChannelParticipant) (string, error)
	Close(candidate State) (string, error)

	// Escrow deposit
	InitiateEscrowDeposit(def ChannelDefinition, initCCS State) (string, error)
	ChallengeEscrowDeposit(candidate State, challengerSig []byte, challengerIdx ChannelParticipant) (string, error)
	FinalizeEscrowDeposit(candidate State) (string, error)

	// Escrow withdrawal
	InitiateEscrowWithdrawal(def ChannelDefinition, initCCS State) (string, error)
	ChallengeEscrowWithdrawal(candidate State, challengerSig []byte, challengerIdx ChannelParticipant) (string, error)
	FinalizeEscrowWithdrawal(candidate State) (string, error)
}

// ========= TransitionValidator Interface =========

// StateAdvancer applies state transitions
type StateAdvancer interface {
	ValidateAdvancement(currentState, proposedState State) error
}

// ========= StatePacker Interface =========

// StatePacker serializes channel states
type StatePacker interface {
	PackState(state State) ([]byte, error)
}

// ========= AssetStore Interface =========

type AssetStore interface {
	// GetAssetDecimals checks if an asset exists and returns its decimals in YN
	GetAssetDecimals(asset string) (uint8, error)

	// GetTokenDecimals returns the decimals for a token on a specific blockchain
	GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error)
}

// Channel lifecycle event handlers
type BlockchainEventHandler interface {
	HandleHomeChannelCreated(context.Context, *HomeChannelCreatedEvent) error
	HandleHomeChannelMigrated(context.Context, *HomeChannelMigratedEvent) error
	HandleHomeChannelCheckpointed(context.Context, *HomeChannelCheckpointedEvent) error
	HandleHomeChannelChallenged(context.Context, *HomeChannelChallengedEvent) error
	HandleHomeChannelClosed(context.Context, *HomeChannelClosedEvent) error
	HandleEscrowDepositInitiated(context.Context, *EscrowDepositInitiatedEvent) error
	HandleEscrowDepositChallenged(context.Context, *EscrowDepositChallengedEvent) error
	HandleEscrowDepositFinalized(context.Context, *EscrowDepositFinalizedEvent) error
	HandleEscrowWithdrawalInitiated(context.Context, *EscrowWithdrawalInitiatedEvent) error
	HandleEscrowWithdrawalChallenged(context.Context, *EscrowWithdrawalChallengedEvent) error
	HandleEscrowWithdrawalFinalized(context.Context, *EscrowWithdrawalFinalizedEvent) error
}
