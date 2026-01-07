package core

import "github.com/shopspring/decimal"

// Channel represents an on-chain channel
type Channel struct {
	ChannelID    string `json:"channel_id"`    // Unique identifier for the channel
	UserWallet   string `json:"user_wallet"`   // User wallet address
	NodeWallet   string `json:"node_wallet"`   // Node wallet address
	Type         string `json:"type"`          // Type of the channel (home, escrow)
	BlockchainID uint32 `json:"blockchain_id"` // Unique identifier for the blockchain
	TokenAddress string `json:"token_address"` // Address of the token used in the channel
	Challenge    uint64 `json:"challenge"`     // Challenge period for the channel in seconds
	Nonce        uint64 `json:"nonce"`         // Nonce for the channel
	Status       string `json:"status"`        // Current status of the channel (void, open, challenged, closed)
	StateVersion uint64 `json:"state_version"` // On-chain state version of the channel
}

// ChannelDefinition represents configuration for creating a channel
type ChannelDefinition struct {
	Nonce     uint64 `json:"nonce"`     // A unique number to prevent replay attacks
	Challenge uint64 `json:"challenge"` // Challenge period for the channel in seconds
}

// State represents the current state of the user stored on Node
type State struct {
	ID              string       `json:"id"`                          // Deterministic ID (hash) of the state
	Transitions     []Transition `json:"transitions"`                 // List of transitions included in the state
	Asset           string       `json:"asset"`                       // Asset type of the state
	UserWallet      string       `json:"user_wallet"`                 // User wallet address
	Epoch           uint64       `json:"epoch"`                       // User Epoch Index
	Version         uint64       `json:"version"`                     // Version of the state
	HomeChannelID   *string      `json:"home_channel_id,omitempty"`   // Identifier for the home Channel ID
	EscrowChannelID *string      `json:"escrow_channel_id,omitempty"` // Identifier for the escrow Channel ID
	HomeLedger      Ledger       `json:"home_ledger"`                 // User and node balances for the home channel
	EscrowLedger    *Ledger      `json:"escrow_ledger,omitempty"`     // User and node balances for the escrow channel
	IsFinal         bool         `json:"is_final"`                    // Indicates if the state is final
	UserSig         *string      `json:"user_sig,omitempty"`          // User signature for the state
	NodeSig         *string      `json:"node_sig,omitempty"`          // Node signature for the state
}

// Ledger represents ledger balances
type Ledger struct {
	TokenAddress string          `json:"token_address"` // Address of the token used in this channel
	BlockchainID uint64          `json:"blockchain_id"` // Unique identifier for the blockchain
	UserBalance  decimal.Decimal `json:"user_balance"`  // User balance in the channel
	UserNetFlow  decimal.Decimal `json:"user_net_flow"` // User net flow in the channel
	NodeBalance  decimal.Decimal `json:"node_balance"`  // Node balance in the channel
	NodeNetFlow  decimal.Decimal `json:"node_net_flow"` // Node net flow in the channel
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeTransfer       TransactionType = "transfer"
	TransactionTypeRelease        TransactionType = "release"
	TransactionTypeCommit         TransactionType = "commit"
	TransactionTypeHomeDeposit    TransactionType = "home_deposit"
	TransactionTypeHomeWithdrawal TransactionType = "home_withdrawal"
	TransactionTypeMutualLock     TransactionType = "mutual_lock"
	TransactionTypeEscrowDeposit  TransactionType = "escrow_deposit"
	TransactionTypeEscrowLock     TransactionType = "escrow_lock"
	TransactionTypeEscrowWithdraw TransactionType = "escrow_withdraw"
	TransactionTypeMigrate        TransactionType = "migrate"
)

// Transaction represents a transaction record
type Transaction struct {
	ID                 string          `json:"id"`                              // Unique transaction reference
	Asset              string          `json:"asset"`                           // Asset symbol
	TxType             TransactionType `json:"tx_type"`                         // Transaction type
	FromAccount        string          `json:"from_account"`                    // The account that sent the funds
	ToAccount          string          `json:"to_account"`                      // The account that received the funds
	SenderNewStateID   *string         `json:"sender_new_state_id,omitempty"`   // The ID of the new sender's channel state
	ReceiverNewStateID *string         `json:"receiver_new_state_id,omitempty"` // The ID of the new receiver's channel state
	Amount             decimal.Decimal `json:"amount"`                          // Transaction amount
	CreatedAt          string          `json:"created_at"`                      // When the transaction was created
}

// TransitionType represents the type of state transition
type TransitionType string

const (
	TransitionTypeTransferReceive TransitionType = "transfer_receive"
	TransitionTypeTransferSend    TransitionType = "transfer_send"
	TransitionTypeRelease         TransitionType = "release"
	TransitionTypeCommit          TransitionType = "commit"
	TransitionTypeHomeDeposit     TransitionType = "home_deposit"
	TransitionTypeHomeWithdrawal  TransitionType = "home_withdrawal"
	TransitionTypeMutualLock      TransitionType = "mutual_lock"
	TransitionTypeEscrowDeposit   TransitionType = "escrow_deposit"
	TransitionTypeEscrowLock      TransitionType = "escrow_lock"
	TransitionTypeEscrowWithdraw  TransitionType = "escrow_withdraw"
	TransitionTypeMigrate         TransitionType = "migrate"
)

// Transition represents a state transition
type Transition struct {
	Type      TransitionType  `json:"type"`                 // Type of state transition
	TxHash    string          `json:"tx_hash"`              // Transaction hash associated with the transition
	AccountID *string         `json:"account_id,omitempty"` // Account identifier (varies based on transition type)
	Amount    decimal.Decimal `json:"amount"`               // Amount involved in the transition
}

// Asset represents information about a supported asset
type Asset struct {
	Token    string `json:"token"`    // Token contract address
	ChainID  uint64 `json:"chain_id"` // Blockchain network ID
	Symbol   string `json:"symbol"`   // Asset symbol
	Decimals uint64 `json:"decimals"` // Number of decimal places
}

// SessionKey represents a session key with spending allowances
type SessionKey struct {
	ID          uint64           `json:"id"`              // Unique identifier for the session key record
	SessionKey  string           `json:"session_key"`     // The address of the session key
	Application string           `json:"application"`     // Name of the application authorized for this session key
	Allowances  []AssetAllowance `json:"allowances"`      // Asset allowances with usage tracking
	Scope       *string          `json:"scope,omitempty"` // Permission scope for this session key
	ExpiresAt   string           `json:"expires_at"`      // When this session key expires (ISO 8601 format)
	CreatedAt   string           `json:"created_at"`      // When the session key was created (ISO 8601 format)
}

// AssetAllowance represents asset allowance with usage tracking
type AssetAllowance struct {
	Asset     string          `json:"asset"`     // Symbol of the asset
	Allowance decimal.Decimal `json:"allowance"` // Maximum amount the session key can spend
	Used      decimal.Decimal `json:"used"`      // Amount already spent by this session key
}
