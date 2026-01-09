package core

import (
	"time"

	"github.com/shopspring/decimal"
)

type ChannelType uint8

var (
	ChannelTypeHome   = 1
	ChannelTypeEscrow = 2
)

// Channel represents an on-chain channel
type Channel struct {
	ChannelID    string      `json:"channel_id"`    // Unique identifier for the channel
	UserWallet   string      `json:"user_wallet"`   // User wallet address
	NodeWallet   string      `json:"node_wallet"`   // Node wallet address
	Type         ChannelType `json:"type"`          // Type of the channel (home, escrow)
	BlockchainID uint32      `json:"blockchain_id"` // Unique identifier for the blockchain
	TokenAddress string      `json:"token_address"` // Address of the token used in the channel
	Challenge    uint64      `json:"challenge"`     // Challenge period for the channel in seconds
	Nonce        uint64      `json:"nonce"`         // Nonce for the channel
	Status       string      `json:"status"`        // Current status of the channel (void, open, challenged, closed)
	StateVersion uint64      `json:"state_version"` // On-chain state version of the channel
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

func (state *State) GetLastTransition() *Transition {
	if len(state.Transitions) == 0 {
		return nil
	}

	lastTransition := state.Transitions[len(state.Transitions)-1]
	if lastTransition.Type == TransitionTypeTransferReceive || lastTransition.Type == TransitionTypeRelease {
		return nil
	}

	return &lastTransition
}

func (state *State) NextState() State {
	var nextState State
	if state.IsFinal {
		nextState = State{
			Transitions:     []Transition{},
			Asset:           state.Asset,
			UserWallet:      state.UserWallet,
			Epoch:           state.Epoch + 1,
			Version:         0,
			HomeChannelID:   nil,
			EscrowChannelID: nil,
			HomeLedger:      Ledger{},
			EscrowLedger:    nil,
			IsFinal:         false,
		}
	} else {
		nextState = State{
			Transitions:     []Transition{},
			Asset:           state.Asset,
			UserWallet:      state.UserWallet,
			Epoch:           state.Epoch,
			Version:         state.Version + 1,
			HomeChannelID:   state.HomeChannelID,
			EscrowChannelID: state.EscrowChannelID,
			HomeLedger:      state.HomeLedger,
			EscrowLedger:    state.EscrowLedger,
			IsFinal:         false,
		}

		if state.UserSig == nil {
			nextState.Transitions = append(nextState.Transitions, state.Transitions...)
		} else if t := state.GetLastTransition(); t.Type == TransitionTypeEscrowDeposit || t.Type == TransitionTypeEscrowWithdraw {
			// escrowChannelID, escrowLedger: not-nil -> nil
			nextState.EscrowChannelID = nil
			nextState.EscrowLedger = nil
		}
	}
	nextState.ID = GetStateID(nextState.UserWallet, nextState.Asset, nextState.Epoch, nextState.Version)

	return nextState
}

// Ledger represents ledger balances
type Ledger struct {
	TokenAddress string          `json:"token_address"` // Address of the token used in this channel
	BlockchainID uint32          `json:"blockchain_id"` // Unique identifier for the blockchain
	UserBalance  decimal.Decimal `json:"user_balance"`  // User balance in the channel
	UserNetFlow  decimal.Decimal `json:"user_net_flow"` // User net flow in the channel
	NodeBalance  decimal.Decimal `json:"node_balance"`  // Node balance in the channel
	NodeNetFlow  decimal.Decimal `json:"node_net_flow"` // Node net flow in the channel
}

// TransactionType represents the type of transaction
type TransactionType uint8

const (
	TransactionTypeHomeDeposit    = 10
	TransactionTypeHomeWithdrawal = 11

	TransactionTypeEscrowDeposit  = 20
	TransactionTypeEscrowWithdraw = 21

	TransactionTypeTransfer TransactionType = 30

	TransactionTypeCommit  = 40
	TransactionTypeRelease = 41

	TransactionTypeMigrate    = 100
	TransactionTypeEscrowLock = 110
	TransactionTypeMutualLock = 120
)

// String returns the human-readable name of the transaction type
func (t TransactionType) String() string {
	switch t {
	case TransactionTypeTransfer:
		return "transfer"
	case TransactionTypeRelease:
		return "release"
	case TransactionTypeCommit:
		return "commit"
	case TransactionTypeHomeDeposit:
		return "home_deposit"
	case TransactionTypeHomeWithdrawal:
		return "home_withdrawal"
	case TransactionTypeMutualLock:
		return "mutual_lock"
	case TransactionTypeEscrowDeposit:
		return "escrow_deposit"
	case TransactionTypeEscrowLock:
		return "escrow_lock"
	case TransactionTypeEscrowWithdraw:
		return "escrow_withdraw"
	case TransactionTypeMigrate:
		return "migrate"
	default:
		return "unknown"
	}
}

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
	CreatedAt          time.Time       `json:"created_at"`                      // When the transaction was created
}

// NewTransaction creates a new instance of Transaction
// returns error if ID generation failed
func NewTransaction(asset string, txType TransactionType, fromAccount, toAccount string, senderNewStateID, receiverNewStateID *string, amount decimal.Decimal) (Transaction, error) {
	id, err := GetTransactionID(toAccount, fromAccount, senderNewStateID, receiverNewStateID)
	if err != nil {
		return Transaction{}, err
	}

	return Transaction{
		ID:                 id,
		Asset:              asset,
		TxType:             txType,
		FromAccount:        fromAccount,
		ToAccount:          toAccount,
		SenderNewStateID:   senderNewStateID,
		ReceiverNewStateID: receiverNewStateID,
		Amount:             amount,
		CreatedAt:          time.Now().UTC(),
	}, nil
}

// TransitionType represents the type of state transition
type TransitionType uint8

const (
	TransitionTypeHomeDeposit    = 10
	TransitionTypeHomeWithdrawal = 11

	TransitionTypeEscrowDeposit  = 20
	TransitionTypeEscrowWithdraw = 21

	TransitionTypeTransferSend    TransitionType = 30
	TransitionTypeTransferReceive TransitionType = 31

	TransitionTypeCommit  = 40
	TransitionTypeRelease = 41

	TransitionTypeMigrate    = 100
	TransitionTypeEscrowLock = 110
	TransitionTypeMutualLock = 120
)

// String returns the human-readable name of the transition type
func (t TransitionType) String() string {
	switch t {
	case TransitionTypeTransferReceive:
		return "transfer_receive"
	case TransitionTypeTransferSend:
		return "transfer_send"
	case TransitionTypeRelease:
		return "release"
	case TransitionTypeCommit:
		return "commit"
	case TransitionTypeHomeDeposit:
		return "home_deposit"
	case TransitionTypeHomeWithdrawal:
		return "home_withdrawal"
	case TransitionTypeMutualLock:
		return "mutual_lock"
	case TransitionTypeEscrowDeposit:
		return "escrow_deposit"
	case TransitionTypeEscrowLock:
		return "escrow_lock"
	case TransitionTypeEscrowWithdraw:
		return "escrow_withdraw"
	case TransitionTypeMigrate:
		return "migrate"
	default:
		return "unknown"
	}
}

// Transition represents a state transition
type Transition struct {
	Type      TransitionType  `json:"type"`       // Type of state transition
	TxHash    string          `json:"tx_hash"`    // Transaction hash associated with the transition
	AccountID string          `json:"account_id"` // Account identifier (varies based on transition type)
	Amount    decimal.Decimal `json:"amount"`     // Amount involved in the transition
}

// NewTransition creates a new state transition
func NewTransition(transitionType TransitionType, txHash string, accountID string, amount decimal.Decimal) *Transition {
	return &Transition{
		Type:      transitionType,
		TxHash:    txHash,
		AccountID: accountID,
		Amount:    amount,
	}
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

// ========= Blockchain CLient Response Types =========

// HomeChannelDataResponse represents the response from getHomeChannelData
type HomeChannelDataResponse struct {
	Definition      ChannelDefinition `json:"definition"`
	Node            string            `json:"node"`
	LastState       State             `json:"last_state"`
	ChallengeExpiry uint64            `json:"challenge_expiry"`
}

// EscrowDepositDataResponse represents the response from getEscrowDepositData
type EscrowDepositDataResponse struct {
	Definition      ChannelDefinition `json:"definition"`
	Node            string            `json:"node"`
	LastState       State             `json:"last_state"`
	UnlockExpiry    uint64            `json:"unlock_expiry"`
	ChallengeExpiry uint64            `json:"challenge_expiry"`
}

// EscrowWithdrawalDataResponse represents the response from getEscrowWithdrawalData
type EscrowWithdrawalDataResponse struct {
	Definition ChannelDefinition `json:"definition"`
	Node       string            `json:"node"`
	LastState  State             `json:"last_state"`
}
