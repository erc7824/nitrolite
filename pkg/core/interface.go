package core

import (
	"github.com/shopspring/decimal"
)

// ========= Listener Interface =========

// Listener defines the interface for listening to channel events
type Listener interface {
	// Listen starts listening for events
	Listen() error

	// Channel lifecycle event handlers
	RegisterChannelCreated(handler func(HomeChannelCreatedEvent) error)
	RegisterChannelMigrated(handler func(HomeChannelMigratedEvent) error)
	RegisterChannelCheckpointed(handler func(HomeChannelCheckpointedEvent) error)
	RegisterChannelChallenged(handler func(HomeChannelChallengedEvent) error)
	RegisterChannelClosed(handler func(HomeChannelClosedEvent) error)

	// Escrow deposit event handlers
	RegisterEscrowDepositInitiated(handler func(EscrowDepositInitiatedEvent) error)
	RegisterEscrowDepositChallenged(handler func(EscrowDepositChallengedEvent) error)
	RegisterEscrowDepositFinalized(handler func(EscrowDepositFinalizedEvent) error)

	// Escrow withdrawal event handlers
	RegisterEscrowWithdrawalInitiated(handler func(EscrowWithdrawalInitiatedEvent) error)
	RegisterEscrowWithdrawalChallenged(handler func(EscrowWithdrawalChallengedEvent) error)
	RegisterEscrowWithdrawalFinalized(handler func(EscrowWithdrawalFinalizedEvent) error)
}

// ========= Client Interface =========

// Client defines the interface for interacting with the ChannelsHub smart contract
// TODO: add context to all methods
type Client interface {
	// Getters - IVault
	GetAccountsBalances(accounts []string, tokens []string) ([][]decimal.Decimal, error)

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
	MigrateChannelHere(def ChannelDefinition, candidate State, proof []State) (string, error)
	Checkpoint(candidate State, proofs []State) (string, error)
	Challenge(candidate State, proofs []State, challengerSig []byte) (string, error)
	Close(candidate State, proofs []State) (string, error)

	// Escrow deposit
	InitiateEscrowDeposit(def ChannelDefinition, initCCS State) (string, error)
	ChallengeEscrowDeposit(candidate State, proof []State) (string, error)
	FinalizeEscrowDeposit(candidate State, proof [2]State) (string, error)

	// Escrow withdrawal
	InitiateEscrowWithdrawal(def ChannelDefinition, initCCS State) (string, error)
	ChallengeEscrowWithdrawal(candidate State, proof []State) (string, error)
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
