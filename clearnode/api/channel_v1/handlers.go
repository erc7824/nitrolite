package channel_v1

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/log"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/erc7824/nitrolite/pkg/sign"
)

type Handler struct {
	transitionApplier   core.TransitionApplier
	transitionValidator core.TransitionValidator
	store               Store
	signer              sign.Signer
	sigValidators       map[SigValidatorType]SigValidator
}

func NewHandler(
	store Store,
	signer sign.Signer,
	sigValidators map[SigValidatorType]SigValidator,
) *Handler {
	return &Handler{
		transitionApplier:   core.NewTransitionApplier(),
		transitionValidator: core.NewSimpleTransitionValidator(),
		store:               store,
		signer:              signer,
		sigValidators:       sigValidators,
	}
}

// HandleTransfer unified balance funds to the specified account
func (h *Handler) SubmitState(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var reqPayload rpc.ChannelsV1SubmitStateRequest
	if err := c.Request.Payload.Translate(&reqPayload); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	incomingState, err := toCoreState(reqPayload.State)
	if err != nil {
		c.Fail(err, "failed to parse state")
		return
	}

	tx, commitTx, _ := h.store.BeginTx()

	currentState, err := tx.GetLastUserState(incomingState.UserWallet, incomingState.Asset, false)
	if err != nil {
		c.Fail(err, "failed to get last user state")
		return
	}

	if err := tx.EnsureNoOngoingStateTransitions(incomingState.UserWallet, incomingState.Asset); err != nil {
		c.Fail(err, "failed to check for ongoing state transitions")
		return
	}

	if err := h.transitionValidator.ValidateTransition(currentState, incomingState); err != nil {
		c.Fail(err, "invalid state transition")
		return
	}

	packedState, err := core.PackState(incomingState)
	if err != nil {
		c.Fail(err, "failed to pack state")
		return
	}

	// Validate user's signature
	if incomingState.UserSig == nil {
		c.Fail(nil, "missing incoming state user signature")
		return
	}
	userSigBytes, err := hexutil.Decode(*incomingState.UserSig)
	if err != nil {
		c.Fail(nil, "incorrect incoming state user signature")
		return
	}

	sigValidator := h.sigValidators[EcdsaSigValidatorType]
	if err := sigValidator.Verify(incomingState.UserWallet, packedState, userSigBytes); err != nil {
		c.Fail(err, "invalid incoming state user signature")
		return
	}

	// Provide node's signature
	stateHash := crypto.Keccak256Hash(packedState).Bytes()
	_nodeSig, err := h.signer.Sign(stateHash)
	if err != nil {
		c.Fail(err, "failed to sign incoming state")
		return
	}
	nodeSig := _nodeSig.String()
	incomingState.NodeSig = &nodeSig

	lastTransition := incomingState.GetLastTransition()
	if lastTransition != nil {
		switch lastTransition.Type {
		case core.TransitionTypeHomeDeposit, core.TransitionTypeHomeWithdrawal, core.TransitionTypeEscrowWithdraw, core.TransitionTypeMutualLock:
			// We return Node's signature, the user is expected to submit this on blockchain.

			if err := h.recordTransaction(tx, incomingState, nil, *lastTransition); err != nil {
				c.Fail(err, "failed to record transaction")
				return
			}
		case core.TransitionTypeTransferSend:
			newReceiverState, err := h.issueTransferReceiverState(tx, incomingState)
			if err != nil {
				c.Fail(err, "failed to issue receiver states")
				return
			}

			if err := h.recordTransaction(tx, incomingState, &newReceiverState, *lastTransition); err != nil {
				c.Fail(err, "failed to record transaction")
				return
			}
		case core.TransitionTypeEscrowDeposit:
			// We return Node's signature, the user is expected to submit this on blockchain.
			// Optionally schedule blockchain action (finalizeEscrowDeposit) on escrow chain
			if err := h.recordTransaction(tx, incomingState, nil, *lastTransition); err != nil {
				c.Fail(err, "failed to record transaction")
				return
			}

		case core.TransitionTypeEscrowLock: // First step in withdrawal through escrow
			if err := tx.ScheduleInitiateEscrowWithdrawal(incomingState); err != nil {
				c.Fail(err, "failed to schedule blockchain action")
				return
			}
			if err := h.recordTransaction(tx, incomingState, nil, *lastTransition); err != nil {
				c.Fail(err, "failed to record transaction")
				return
			}
		case core.TransitionTypeMigrate:
			c.Fail(nil, "transition is not suppoted yet")
		}
	}

	if err := tx.StoreUserState(incomingState); err != nil {
		c.Fail(err, "failed to store state")
		return
	}

	if err := commitTx(); err != nil {
		c.Fail(err, "failed to commit transaction")
		return
	}

	resp := rpc.ChannelsV1SubmitStateResponse{
		Signature: nodeSig,
	}
	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
	logger.Info("state submitted", "userID", incomingState.UserWallet, "asset", incomingState.Asset, "version", incomingState.Version)
}

func (h *Handler) issueTransferReceiverState(tx Store, senderState core.State) (core.State, error) {
	lastTransition := senderState.GetLastTransition()
	if lastTransition == nil {
		return core.State{}, rpc.Errorf("")
	}
	if lastTransition.Type != core.TransitionTypeTransferSend {
		return core.State{}, rpc.Errorf("")
	}

	currentState, err := tx.GetLastUserState(lastTransition.AccountID, senderState.Asset, false)
	if err != nil {
		return core.State{}, rpc.Errorf("failed to get last %s user state for transfer receiver with address %s", senderState.Asset, lastTransition.AccountID)
	}
	newState := currentState.NextState()

	receiveTransition := core.Transition{
		Type:      core.TransitionTypeTransferReceive,
		TxHash:    lastTransition.TxHash,
		AccountID: senderState.UserWallet,
		Amount:    lastTransition.Amount,
	}
	newState, err = h.transitionApplier.Apply(newState, receiveTransition)
	if err != nil {
		return core.State{}, err
	}

	packedState, err := core.PackState(newState)
	if err != nil {
		return core.State{}, rpc.Errorf("failed to pack receiver state")
	}

	stateHash := crypto.Keccak256Hash(packedState).Bytes()
	_nodeSig, err := h.signer.Sign(stateHash)
	if err != nil {
		return core.State{}, rpc.Errorf("failed to sign reciver state")
	}
	nodeSig := _nodeSig.String()
	newState.NodeSig = &nodeSig

	if err := tx.StoreUserState(newState); err != nil {
		return core.State{}, rpc.Errorf("failed to store receiver state")
	}

	return newState, nil
}

func (h *Handler) recordTransaction(tx Store, senderState core.State, receiverState *core.State, transition core.Transition) error {
	var txType core.TransactionType
	var toAccount, fromAccount string
	// Transition validator is expected to make sure that all the fields are present and valid.

	switch transition.Type {
	case core.TransitionTypeHomeDeposit:
		txType = core.TransactionTypeHomeDeposit
		fromAccount = *senderState.HomeChannelID
		toAccount = senderState.UserWallet

	case core.TransitionTypeHomeWithdrawal:
		txType = core.TransactionTypeHomeWithdrawal
		fromAccount = senderState.UserWallet
		toAccount = *senderState.HomeChannelID

	case core.TransitionTypeEscrowDeposit:
		txType = core.TransactionTypeEscrowDeposit
		fromAccount = *senderState.EscrowChannelID
		toAccount = senderState.UserWallet

	case core.TransitionTypeEscrowWithdraw:
		txType = core.TransactionTypeEscrowWithdraw
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	case core.TransitionTypeMutualLock:
		txType = core.TransactionTypeMutualLock
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	case core.TransitionTypeTransferSend:
		txType = core.TransactionTypeTransfer
		fromAccount = senderState.UserWallet
		toAccount = receiverState.UserWallet

	case core.TransitionTypeEscrowLock:
		txType = core.TransactionTypeEscrowLock
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	default:
		txType = core.TransactionTypeMigrate
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID
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

	return nil
}
