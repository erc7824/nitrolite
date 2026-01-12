package channel_v1

import (
	"context"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/log"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/erc7824/nitrolite/pkg/sign"
)

// Handler manages channel state transitions and provides RPC endpoints for state submission.
type Handler struct {
	stateAdvancer core.StateAdvancer
	useStoreInTx  StoreTxProvider
	signer        sign.Signer
	sigValidators map[SigValidatorType]SigValidator
	nodeAddress   string // Node's wallet address for channel ID calculation
	minChallenge  uint64
}

// NewHandler creates a new Handler instance with the provided dependencies.
func NewHandler(
	useStoreInTx StoreTxProvider,
	signer sign.Signer,
	sigValidators map[SigValidatorType]SigValidator,
	nodeAddress string,
	minChallenge uint64,
) *Handler {
	return &Handler{
		stateAdvancer: core.NewStateAdvancerV1(),
		useStoreInTx:  useStoreInTx,
		signer:        signer,
		sigValidators: sigValidators,
		nodeAddress:   nodeAddress,
		minChallenge:  minChallenge,
	}
}

// issueTransferReceiverState creates and stores a new state for the receiver of a transfer.
// It reads the receiver's current state, applies a transfer_receive transition with the same
// amount and tx hash, signs it with the node's key, and persists it.
func (h *Handler) issueTransferReceiverState(ctx context.Context, tx Store, senderState core.State) (core.State, error) {
	logger := log.FromContext(ctx)

	incomingTransition := senderState.GetLastTransition()
	if incomingTransition == nil {
		return core.State{}, rpc.Errorf("incoming state has no transitions")
	}
	if incomingTransition.Type != core.TransitionTypeTransferSend {
		return core.State{}, rpc.Errorf("incoming state doesn't have 'transfer_send' transition")
	}
	receiverWallet := incomingTransition.AccountID
	logger = logger.
		WithKV("sender", senderState.UserWallet).
		WithKV("receiver", receiverWallet).
		WithKV("asset", senderState.Asset)

	logger.Debug("issuing transfer receiver state")

	currentState, err := tx.GetLastUserState(receiverWallet, senderState.Asset, false)
	if err != nil {
		return core.State{}, rpc.Errorf("failed to get last %s user state for transfer receiver with address %s", senderState.Asset, incomingTransition.AccountID)
	}
	newState := currentState.NextState()

	receiveTransition := core.Transition{
		Type:      core.TransitionTypeTransferReceive,
		TxHash:    incomingTransition.TxHash,
		AccountID: senderState.UserWallet,
		Amount:    incomingTransition.Amount,
	}
	newState, err = h.stateAdvancer.ApplyTransition(newState, receiveTransition)
	if err != nil {
		return core.State{}, err
	}

	lastSignedState, err := tx.GetLastUserState(receiverWallet, senderState.Asset, true)
	if err != nil {
		return core.State{}, rpc.Errorf("failed to get last %s user state for transfer receiver with address %s", senderState.Asset, incomingTransition.AccountID)
	}
	var lastStateTransition *core.Transition
	if lastSignedState != nil {
		lastStateTransition = lastSignedState.GetLastTransition()
	}

	if !(lastStateTransition != nil && (lastStateTransition.Type == core.TransactionTypeMutualLock || lastStateTransition.Type == core.TransactionTypeEscrowLock)) {
		packedState, err := core.PackState(newState)
		if err != nil {
			return core.State{}, rpc.Errorf("failed to pack receiver state")
		}

		stateHash := crypto.Keccak256Hash(packedState).Bytes()
		_nodeSig, err := h.signer.Sign(stateHash)
		if err != nil {
			return core.State{}, rpc.Errorf("failed to sign receiver state")
		}
		nodeSig := _nodeSig.String()
		newState.NodeSig = &nodeSig
	}
	if err := tx.StoreUserState(newState); err != nil {
		return core.State{}, rpc.Errorf("failed to store receiver state")
	}

	logger.Info("issued transfer receiver state", "receiverStateVersion", newState.Version)
	return newState, nil
}

// issueExtraState creates an additional state by reapplying unsigned transitions to a newly signed state.
// When a user submits a signed state (e.g., after escrow_deposit or escrow_withdraw), any pending
// unsigned transitions from the previous state are reapplied to create a new unsigned state.
// This ensures that pending operations are preserved across state updates that require user signatures.
func (h *Handler) issueExtraState(ctx context.Context, tx Store, incomingState core.State) (core.State, error) {
	logger := log.FromContext(ctx)

	lastTransition := incomingState.GetLastTransition()
	if lastTransition == nil {
		return core.State{}, rpc.Errorf("incoming state has no transitions")
	}

	lastUnsignedState, err := tx.GetLastUserState(incomingState.UserWallet, incomingState.Asset, false)
	if err != nil {
		return core.State{}, rpc.Errorf("failed to get last unsigned user state")
	}

	if lastUnsignedState == nil || lastUnsignedState.UserSig != nil {
		return incomingState, err
	}

	logger = logger.
		WithKV("userWallet", incomingState.UserWallet).
		WithKV("asset", incomingState.Asset)

	extraState := incomingState.NextState()
	logger.Debug("issuing extra state", "extraStateVersion", extraState.Version)

	extraState, err = h.stateAdvancer.ReapplyTransitions(*lastUnsignedState, extraState)
	if err != nil {
		return core.State{}, err
	}

	packedState, err := core.PackState(extraState)
	if err != nil {
		return core.State{}, rpc.Errorf("failed to pack extra state")
	}

	stateHash := crypto.Keccak256Hash(packedState).Bytes()
	_nodeSig, err := h.signer.Sign(stateHash)
	if err != nil {
		return core.State{}, rpc.Errorf("failed to sign extra state")
	}
	nodeSig := _nodeSig.String()
	extraState.NodeSig = &nodeSig

	if err := tx.StoreUserState(extraState); err != nil {
		return core.State{}, rpc.Errorf("failed to store extra state")
	}

	logger.Info("issued extra state", "extraStateVersion", extraState.Version)
	return extraState, nil
}

// recordTransaction creates and persists a transaction record for the given state transition.
// It maps the transition type to the appropriate transaction type and extracts the from/to accounts.
func (h *Handler) recordTransaction(ctx context.Context, tx Store, senderState core.State, receiverState *core.State, transition core.Transition) error {
	logger := log.FromContext(ctx)

	var txType core.TransactionType
	var toAccount, fromAccount string
	// Transition validator is expected to make sure that all the fields are present and valid.

	switch transition.Type {
	case core.TransitionTypeHomeDeposit:
		if senderState.HomeChannelID == nil {
			return rpc.Errorf("sender state has no home channel ID")
		}

		txType = core.TransactionTypeHomeDeposit
		fromAccount = *senderState.HomeChannelID
		toAccount = senderState.UserWallet

	case core.TransitionTypeHomeWithdrawal:
		if senderState.HomeChannelID == nil {
			return rpc.Errorf("sender state has no home channel ID")
		}

		txType = core.TransactionTypeHomeWithdrawal
		fromAccount = senderState.UserWallet
		toAccount = *senderState.HomeChannelID

	case core.TransitionTypeEscrowDeposit:
		if senderState.EscrowChannelID == nil {
			return rpc.Errorf("sender state has no escrow channel ID")
		}

		txType = core.TransactionTypeEscrowDeposit
		fromAccount = *senderState.EscrowChannelID
		toAccount = senderState.UserWallet

	case core.TransitionTypeEscrowWithdraw:
		if senderState.EscrowChannelID == nil || senderState.HomeChannelID == nil {
			return rpc.Errorf("sender state has no escrow or home channel ID")
		}

		txType = core.TransactionTypeEscrowWithdraw
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	case core.TransitionTypeMutualLock:
		if senderState.EscrowChannelID == nil || senderState.HomeChannelID == nil {
			return rpc.Errorf("sender state has no escrow or home channel ID")
		}

		txType = core.TransactionTypeMutualLock
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	case core.TransitionTypeTransferSend:
		txType = core.TransactionTypeTransfer
		fromAccount = senderState.UserWallet
		if receiverState == nil {
			return rpc.Errorf("receiver state has not been issued")
		}
		toAccount = receiverState.UserWallet

	case core.TransitionTypeEscrowLock:
		if senderState.EscrowChannelID == nil || senderState.HomeChannelID == nil {
			return rpc.Errorf("sender state has no escrow or home channel ID")
		}

		txType = core.TransactionTypeEscrowLock
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	case core.TransitionTypeMigrate:
		if senderState.EscrowChannelID == nil || senderState.HomeChannelID == nil {
			return rpc.Errorf("sender state has no escrow or home channel ID")
		}

		txType = core.TransactionTypeMigrate
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	default:
		return rpc.Errorf("invalid transition type")
	}

	var receiverStateID *string
	if receiverState != nil {
		receiverStateID = &receiverState.ID
	}

	transaction, err := core.NewTransaction(
		senderState.Asset,
		txType,
		fromAccount,
		toAccount,
		&senderState.ID,
		receiverStateID,
		transition.Amount,
	)
	if err != nil {
		return rpc.Errorf("failed to create transaction")
	}
	if err := tx.RecordTransaction(transaction); err != nil {
		return rpc.Errorf("failed to record transaction")
	}

	logger.Info("transaction recorder",
		"id", transaction.ID,
		"type", transaction.TxType.String(),
		"from", transaction.FromAccount,
		"to", transaction.ToAccount,
		"senderStateID", transaction.SenderNewStateID,
		"asset", transaction.Asset,
		"amount", transaction.Amount.String(),
	)

	return nil
}
