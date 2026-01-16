package core

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type ChannelType uint8

var (
	ChannelTypeHome   ChannelType = 1
	ChannelTypeEscrow ChannelType = 2
)

type ChannelStatus uint8

var (
	ChannelStatusVoid       ChannelStatus = 0
	ChannelStatusOpen       ChannelStatus = 1
	ChannelStatusChallenged ChannelStatus = 2
	ChannelStatusClosed     ChannelStatus = 3
)

// Channel represents an on-chain channel
type Channel struct {
	ChannelID    string        `json:"channel_id"`    // Unique identifier for the channel
	UserWallet   string        `json:"user_wallet"`   // User wallet address
	NodeWallet   string        `json:"node_wallet"`   // Node wallet address
	Type         ChannelType   `json:"type"`          // Type of the channel (home, escrow)
	BlockchainID uint32        `json:"blockchain_id"` // Unique identifier for the blockchain
	TokenAddress string        `json:"token_address"` // Address of the token used in the channel
	Challenge    uint64        `json:"challenge"`     // Challenge period for the channel in seconds
	Nonce        uint64        `json:"nonce"`         // Nonce for the channel
	Status       ChannelStatus `json:"status"`        // Current status of the channel (void, open, challenged, closed)
	StateVersion uint64        `json:"state_version"` // On-chain state version of the channel
}

func NewChannel(channelID, userWallet, nodeWallet string, ChType ChannelType, blockchainID uint32, tokenAddress string, nonce, challenge uint64) *Channel {
	return &Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		NodeWallet:   nodeWallet,
		Type:         ChType,
		BlockchainID: blockchainID,
		TokenAddress: tokenAddress,
		Nonce:        nonce,
		Challenge:    challenge,
		Status:       ChannelStatusVoid,
		StateVersion: 0,
	}
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

func NewVoidState(asset, userWallet string) *State {
	return &State{
		Asset:      asset,
		UserWallet: userWallet,
		HomeLedger: Ledger{
			UserBalance: decimal.Zero,
			UserNetFlow: decimal.Zero,
			NodeBalance: decimal.Zero,
			NodeNetFlow: decimal.Zero,
		},
	}
}

func (state State) GetLastTransition() *Transition {
	if len(state.Transitions) == 0 {
		return nil
	}

	lastTransition := state.Transitions[len(state.Transitions)-1]
	if lastTransition.Type == TransitionTypeTransferReceive || lastTransition.Type == TransitionTypeRelease {
		return nil
	}

	return &lastTransition
}

func (state State) NextState() *State {
	var nextState *State
	if state.IsFinal {
		nextState = &State{
			Transitions:     []Transition{},
			Asset:           state.Asset,
			UserWallet:      state.UserWallet,
			Epoch:           state.Epoch + 1,
			Version:         0,
			HomeChannelID:   nil,
			EscrowChannelID: nil,
			HomeLedger: Ledger{
				UserBalance: decimal.Zero,
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			EscrowLedger: nil,
			IsFinal:      false,
		}
	} else {
		nextState = &State{
			Transitions:     []Transition{},
			Asset:           state.Asset,
			UserWallet:      state.UserWallet,
			Epoch:           state.Epoch,
			Version:         state.Version + 1,
			HomeChannelID:   state.HomeChannelID,
			EscrowChannelID: state.EscrowChannelID,
			HomeLedger:      state.HomeLedger,
			EscrowLedger:    nil,
			IsFinal:         false,
		}
		if state.EscrowLedger != nil {
			nextState.EscrowLedger = &Ledger{
				TokenAddress: state.EscrowLedger.TokenAddress,
				BlockchainID: state.EscrowLedger.BlockchainID,
				UserBalance:  state.EscrowLedger.UserBalance,
				UserNetFlow:  state.EscrowLedger.UserNetFlow,
				NodeBalance:  state.EscrowLedger.NodeBalance,
				NodeNetFlow:  state.EscrowLedger.NodeNetFlow,
			}
		}

		if state.UserSig == nil {
			nextState.Transitions = state.Transitions
		} else if t := state.GetLastTransition(); t != nil && (t.Type == TransitionTypeEscrowDeposit || t.Type == TransitionTypeEscrowWithdraw) {
			// escrowChannelID, escrowLedger: not-nil -> nil
			nextState.EscrowChannelID = nil
			nextState.EscrowLedger = nil
		}
	}
	nextState.ID = GetStateID(nextState.UserWallet, nextState.Asset, nextState.Epoch, nextState.Version)

	return nextState
}

func (state *State) ApplyReceiverTransitions(transitions ...Transition) error {
	for _, transition := range transitions {
		var err error
		switch transition.Type {
		case TransitionTypeTransferReceive:
			_, err = state.ApplyTransferReceiveTransition(transition.AccountID, transition.Amount, transition.TxID)
		case TransitionTypeRelease:
			_, err = state.ApplyReleaseTransition(transition.AccountID, transition.Amount)
		default:
			return fmt.Errorf("transition '%s' cannot be applied by receiver", transition.Type.String())
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (state *State) ApplyHomeDepositTransition(amount decimal.Decimal) (Transition, error) {
	if state.HomeChannelID == nil {
		return Transition{}, fmt.Errorf("missing home channel ID")
	}

	accountID := *state.HomeChannelID
	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeHomeDeposit, txID, accountID, amount)
	state.Transitions = append(state.Transitions, *newTransition)
	state.HomeLedger.UserNetFlow = state.HomeLedger.UserNetFlow.Add(newTransition.Amount)
	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Add(newTransition.Amount)

	return *newTransition, nil
}

func (state *State) ApplyHomeWithdrawalTransition(amount decimal.Decimal) (Transition, error) {
	if state.HomeChannelID == nil {
		return Transition{}, fmt.Errorf("missing home channel ID")
	}

	accountID := *state.HomeChannelID
	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeHomeWithdrawal, txID, accountID, amount)
	state.Transitions = append(state.Transitions, *newTransition)
	state.HomeLedger.UserNetFlow = state.HomeLedger.UserNetFlow.Sub(newTransition.Amount)
	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Sub(newTransition.Amount)

	return *newTransition, nil
}

func (state *State) ApplyTransferSendTransition(recipient string, amount decimal.Decimal) (Transition, error) {
	// TODO: maybe validate that recipient is a correct UserWallet format
	accountID := recipient
	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeTransferSend, txID, accountID, amount)
	state.Transitions = append(state.Transitions, *newTransition)
	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Sub(newTransition.Amount)
	state.HomeLedger.NodeNetFlow = state.HomeLedger.NodeNetFlow.Sub(newTransition.Amount)

	return *newTransition, nil
}

func (state *State) ApplyTransferReceiveTransition(sender string, amount decimal.Decimal, txID string) (Transition, error) {
	// TODO: maybe validate that recipient is a correct UserWallet format
	accountID := sender

	newTransition := NewTransition(TransitionTypeTransferReceive, txID, accountID, amount)
	state.Transitions = append(state.Transitions, *newTransition)
	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Add(newTransition.Amount)
	state.HomeLedger.NodeNetFlow = state.HomeLedger.NodeNetFlow.Add(newTransition.Amount)
	return *newTransition, nil
}

func (state *State) ApplyCommitTransition(accountID string, amount decimal.Decimal) (Transition, error) {
	// TODO: maybe validate that AccountID has correct AppSessionID format
	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeCommit, txID, accountID, amount)
	state.Transitions = append(state.Transitions, *newTransition)
	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Sub(newTransition.Amount)
	state.HomeLedger.NodeNetFlow = state.HomeLedger.NodeNetFlow.Sub(newTransition.Amount)

	return *newTransition, nil
}

func (state *State) ApplyReleaseTransition(accountID string, amount decimal.Decimal) (Transition, error) {
	// TODO: maybe validate that recipient is a correct UserWallet format
	txID, err := GetReceiverTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeRelease, txID, accountID, amount)
	state.Transitions = append(state.Transitions, *newTransition)
	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Add(newTransition.Amount)
	state.HomeLedger.NodeNetFlow = state.HomeLedger.NodeNetFlow.Add(newTransition.Amount)
	return *newTransition, nil
}

func (state *State) ApplyMutualLockTransition(blockchainID uint32, tokenAddress string, amount decimal.Decimal) (Transition, error) {
	if state.HomeChannelID == nil {
		return Transition{}, fmt.Errorf("missing home channel ID")
	}
	if blockchainID == 0 {
		return Transition{}, fmt.Errorf("invalid blockchain ID")
	}
	if tokenAddress == "" {
		return Transition{}, fmt.Errorf("invalid token address")
	}

	escrowChannelID, err := GetEscrowChannelID(*state.HomeChannelID, state.Version)
	if err != nil {
		return Transition{}, err
	}
	state.EscrowChannelID = &escrowChannelID
	accountID := escrowChannelID

	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeMutualLock, txID, accountID, amount)
	state.Transitions = append(state.Transitions, *newTransition)

	state.HomeLedger.NodeBalance = state.HomeLedger.NodeBalance.Add(newTransition.Amount)
	state.HomeLedger.NodeNetFlow = state.HomeLedger.NodeNetFlow.Add(newTransition.Amount)

	state.EscrowLedger = &Ledger{
		BlockchainID: blockchainID,
		TokenAddress: tokenAddress,
		UserBalance:  decimal.Zero.Add(newTransition.Amount),
		UserNetFlow:  decimal.Zero.Add(newTransition.Amount),
		NodeBalance:  decimal.Zero,
		NodeNetFlow:  decimal.Zero,
	}

	return *newTransition, nil
}

func (state *State) ApplyEscrowDepositTransition(amount decimal.Decimal) (Transition, error) {
	if state.EscrowChannelID == nil {
		return Transition{}, fmt.Errorf("internal error: escrow channel ID is nil")
	}
	if state.EscrowLedger == nil {
		return Transition{}, fmt.Errorf("escrow ledger is nil")
	}
	accountID := *state.EscrowChannelID

	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeEscrowDeposit, txID, accountID, amount)
	state.Transitions = append(state.Transitions, *newTransition)

	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Add(newTransition.Amount)
	state.HomeLedger.NodeNetFlow = state.HomeLedger.NodeNetFlow.Add(newTransition.Amount)

	state.EscrowLedger.UserBalance = state.EscrowLedger.UserBalance.Sub(newTransition.Amount)
	state.EscrowLedger.NodeNetFlow = state.EscrowLedger.NodeNetFlow.Sub(newTransition.Amount)

	return *newTransition, nil
}

func (state *State) ApplyEscrowLockTransition(blockchainID uint32, tokenAddress string, amount decimal.Decimal) (Transition, error) {
	if state.HomeChannelID == nil {
		return Transition{}, fmt.Errorf("missing home channel ID")
	}
	if blockchainID == 0 {
		return Transition{}, fmt.Errorf("invalid blockchain ID")
	}
	if tokenAddress == "" {
		return Transition{}, fmt.Errorf("invalid token address")
	}

	escrowChannelID, err := GetEscrowChannelID(*state.HomeChannelID, state.Version)
	if err != nil {
		return Transition{}, err
	}
	state.EscrowChannelID = &escrowChannelID
	accountID := escrowChannelID

	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeEscrowLock, txID, accountID, amount)
	state.Transitions = append(state.Transitions, *newTransition)

	state.EscrowLedger = &Ledger{
		BlockchainID: blockchainID,
		TokenAddress: tokenAddress,
		UserBalance:  decimal.Zero,
		UserNetFlow:  decimal.Zero,
		NodeBalance:  decimal.Zero.Add(newTransition.Amount),
		NodeNetFlow:  decimal.Zero.Add(newTransition.Amount),
	}

	return *newTransition, nil
}

func (state *State) ApplyEscrowWithdrawTransition(amount decimal.Decimal) (Transition, error) {
	if state.EscrowChannelID == nil {
		return Transition{}, fmt.Errorf("internal error: escrow channel ID is nil")
	}
	if state.EscrowLedger == nil {
		return Transition{}, fmt.Errorf("escrow ledger is nil")
	}
	accountID := *state.EscrowChannelID

	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeEscrowWithdraw, txID, accountID, amount)
	state.Transitions = append(state.Transitions, *newTransition)

	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Sub(newTransition.Amount)
	state.HomeLedger.NodeNetFlow = state.HomeLedger.NodeNetFlow.Sub(newTransition.Amount)

	state.EscrowLedger.UserNetFlow = state.EscrowLedger.UserNetFlow.Sub(newTransition.Amount)
	state.EscrowLedger.NodeBalance = state.EscrowLedger.NodeBalance.Sub(newTransition.Amount)

	return *newTransition, nil
}

func (state *State) ApplyMigrateTransition(amount decimal.Decimal) (Transition, error) {
	return Transition{}, fmt.Errorf("migrate transition not implemented yet")
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

func (l1 Ledger) Equal(l2 Ledger) error {
	if l1.TokenAddress != l2.TokenAddress {
		return fmt.Errorf("token address mismatch: expected=%s, proposed=%s", l1.TokenAddress, l2.TokenAddress)
	}
	if l1.BlockchainID != l2.BlockchainID {
		return fmt.Errorf("blockchain ID mismatch: expected=%d, proposed=%d", l1.BlockchainID, l2.BlockchainID)
	}
	if !l1.UserBalance.Equal(l2.UserBalance) {
		return fmt.Errorf("user balance mismatch: expected=%s, proposed=%s", l1.UserBalance.String(), l2.UserBalance.String())
	}
	if !l1.UserNetFlow.Equal(l2.UserNetFlow) {
		return fmt.Errorf("user net flow mismatch: expected=%s, proposed=%s", l1.UserNetFlow.String(), l2.UserNetFlow.String())
	}
	if !l1.NodeBalance.Equal(l2.NodeBalance) {
		return fmt.Errorf("node balance mismatch: expected=%s, proposed=%s", l1.NodeBalance.String(), l2.NodeBalance.String())
	}
	if !l1.NodeNetFlow.Equal(l2.NodeNetFlow) {
		return fmt.Errorf("node net flow mismatch: expected=%s, proposed=%s", l1.NodeNetFlow.String(), l2.NodeNetFlow.String())
	}
	return nil
}

func (l Ledger) Validate() error {
	if l.TokenAddress == "" {
		return fmt.Errorf("invalid token address")
	}
	if l.BlockchainID == 0 {
		return fmt.Errorf("invalid blockchain ID")
	}
	if l.UserBalance.IsNegative() {
		return fmt.Errorf("user balance cannot be negative")
	}
	if l.NodeBalance.IsNegative() {
		return fmt.Errorf("node balance cannot be negative")
	}
	sumBalances := l.UserBalance.Add(l.NodeBalance)
	sumNetFlows := l.UserNetFlow.Add(l.NodeNetFlow)
	if !sumBalances.Equal(sumNetFlows) {
		return fmt.Errorf("ledger balances do not match net flows: balances=%s, net_flows=%s", sumBalances.String(), sumNetFlows.String())
	}

	return nil
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
func NewTransaction(id, asset string, txType TransactionType, fromAccount, toAccount string, senderNewStateID, receiverNewStateID *string, amount decimal.Decimal) *Transaction {
	return &Transaction{
		ID:                 id,
		Asset:              asset,
		TxType:             txType,
		FromAccount:        fromAccount,
		ToAccount:          toAccount,
		SenderNewStateID:   senderNewStateID,
		ReceiverNewStateID: receiverNewStateID,
		Amount:             amount,
		CreatedAt:          time.Now().UTC(),
	}
}

// NewTransactionFromTransition maps the transition type to the appropriate transaction type and returns a pointer to a Transaction.
func NewTransactionFromTransition(senderState State, receiverState *State, transition Transition) (*Transaction, error) {
	var txType TransactionType
	var toAccount, fromAccount string
	// Transition validator is expected to make sure that all the fields are present and valid.

	switch transition.Type {
	case TransitionTypeHomeDeposit:
		if senderState.HomeChannelID == nil {
			return nil, fmt.Errorf("sender state has no home channel ID")
		}

		txType = TransactionTypeHomeDeposit
		fromAccount = *senderState.HomeChannelID
		toAccount = senderState.UserWallet

	case TransitionTypeHomeWithdrawal:
		if senderState.HomeChannelID == nil {
			return nil, fmt.Errorf("sender state has no home channel ID")
		}

		txType = TransactionTypeHomeWithdrawal
		fromAccount = senderState.UserWallet
		toAccount = *senderState.HomeChannelID

	case TransitionTypeEscrowDeposit:
		if senderState.EscrowChannelID == nil {
			return nil, fmt.Errorf("sender state has no escrow channel ID")
		}

		txType = TransactionTypeEscrowDeposit
		fromAccount = *senderState.EscrowChannelID
		toAccount = senderState.UserWallet

	case TransitionTypeEscrowWithdraw:
		if senderState.EscrowChannelID == nil || senderState.HomeChannelID == nil {
			return nil, fmt.Errorf("sender state has no escrow or home channel ID")
		}

		txType = TransactionTypeEscrowWithdraw
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	case TransitionTypeTransferSend:
		if receiverState == nil {
			return nil, fmt.Errorf("receiver state must not be nil for 'transfer_send' transition")
		}

		txType = TransactionTypeTransfer
		fromAccount = senderState.UserWallet
		toAccount = transition.AccountID

	case TransactionTypeCommit:
		txType = TransactionTypeCommit
		fromAccount = senderState.UserWallet
		toAccount = transition.AccountID

	case TransactionTypeRelease:
		txType = TransactionTypeRelease
		fromAccount = transition.AccountID
		toAccount = senderState.UserWallet
		if receiverState != nil {
			return nil, fmt.Errorf("receiver state must not be nil for 'release' transition")
		}

	case TransitionTypeMutualLock:
		if senderState.EscrowChannelID == nil || senderState.HomeChannelID == nil {
			return nil, fmt.Errorf("sender state has no escrow or home channel ID")
		}

		txType = TransactionTypeMutualLock
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	case TransitionTypeEscrowLock:
		if senderState.EscrowChannelID == nil || senderState.HomeChannelID == nil {
			return nil, fmt.Errorf("sender state has no escrow or home channel ID")
		}

		txType = TransactionTypeEscrowLock
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	case TransitionTypeMigrate:
		if senderState.EscrowChannelID == nil || senderState.HomeChannelID == nil {
			return nil, fmt.Errorf("sender state has no escrow or home channel ID")
		}

		txType = TransactionTypeMigrate
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	default:
		return nil, fmt.Errorf("invalid transition type")
	}

	var receiverStateID *string
	var txID string
	var err error
	if receiverState != nil {
		receiverStateID = &receiverState.ID
		txID, err = GetReceiverTransactionID(fromAccount, receiverState.ID)
	} else {
		txID, err = GetSenderTransactionID(toAccount, senderState.ID)
	}
	if err != nil {
		return nil, err
	}

	return NewTransaction(
		txID,
		senderState.Asset,
		txType,
		fromAccount,
		toAccount,
		&senderState.ID,
		receiverStateID,
		transition.Amount,
	), nil
}

// TransitionType represents the type of state transition
type TransitionType uint8

const (
	TransitionTypeHomeDeposit    = 10 // AccountID: HomeChannelID
	TransitionTypeHomeWithdrawal = 11 // AccountID: HomeChannelID

	TransitionTypeEscrowDeposit  = 20 // AccountID: EscrowChannelID
	TransitionTypeEscrowWithdraw = 21 // AccountID: EscrowChannelID

	TransitionTypeTransferSend    TransitionType = 30 // AccountID: Receiver's UserWallet
	TransitionTypeTransferReceive TransitionType = 31 // AccountID: Sender's UserWallet

	TransitionTypeCommit  = 40 // AccountID: AppSessionID
	TransitionTypeRelease = 41 // AccountID: AppSessionID

	TransitionTypeMigrate    = 100 // AccountID: EscrowChannelID
	TransitionTypeEscrowLock = 110 // AccountID: EscrowChannelID
	TransitionTypeMutualLock = 120 // AccountID: EscrowChannelID
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

func (t TransitionType) RequiresOpenChannel() bool {
	switch t {
	case TransitionTypeTransferReceive,
		TransitionTypeRelease:
		return false
	case TransitionTypeTransferSend,
		TransitionTypeCommit,
		TransitionTypeHomeDeposit,
		TransitionTypeHomeWithdrawal,
		TransitionTypeMutualLock,
		TransitionTypeEscrowDeposit,
		TransitionTypeEscrowLock,
		TransitionTypeEscrowWithdraw,
		TransitionTypeMigrate:
		return true
	default:
		return true
	}
}

// Transition represents a state transition
type Transition struct {
	Type      TransitionType  `json:"type"`       // Type of state transition
	TxID      string          `json:"tx_id"`      // Transaction ID associated with the transition
	AccountID string          `json:"account_id"` // Account identifier (varies based on transition type)
	Amount    decimal.Decimal `json:"amount"`     // Amount involved in the transition
}

// NewTransition creates a new state transition
func NewTransition(transitionType TransitionType, txID, accountID string, amount decimal.Decimal) *Transition {
	return &Transition{
		Type:      transitionType,
		TxID:      txID,
		AccountID: accountID,
		Amount:    amount,
	}
}

// Equal checks if two transitions are equal
func (t1 Transition) Equal(t2 Transition) error {
	if t1.Type != t2.Type {
		return fmt.Errorf("transition type mismatch: expected=%s, proposed=%s", t1.Type.String(), t2.Type.String())
	}
	if t1.TxID != t2.TxID {
		return fmt.Errorf("transaction ID mismatch: expected=%s, proposed=%s", t1.TxID, t2.TxID)
	}
	if t1.AccountID != t2.AccountID {
		return fmt.Errorf("account ID mismatch: expected=%s, proposed=%s", t1.AccountID, t2.AccountID)
	}
	if !t1.Amount.Equal(t2.Amount) {
		return fmt.Errorf("amount mismatch: expected=%s, proposed=%s", t1.Amount.String(), t2.Amount.String())
	}
	return nil
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
