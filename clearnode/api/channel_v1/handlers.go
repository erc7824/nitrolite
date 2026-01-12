package channel_v1

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
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
}

// NewHandler creates a new Handler instance with the provided dependencies.
func NewHandler(
	useStoreInTx StoreTxProvider,
	signer sign.Signer,
	sigValidators map[SigValidatorType]SigValidator,
) *Handler {
	return &Handler{
		stateAdvancer: core.NewStateAdvancerV1(),
		useStoreInTx:  useStoreInTx,
		signer:        signer,
		sigValidators: sigValidators,
	}
}

// SubmitState processes user-submitted state transitions, validates them against the current state,
// verifies user signatures, signs the new state with the node's key, and persists changes.
// For transfer transitions, it automatically creates corresponding receiver states.
// For certain transitions (escrow lock, etc.), it schedules blockchain actions.
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

	logger = logger.
		WithKV("userWallet", incomingState.UserWallet).
		WithKV("asset", incomingState.Asset)

	var nodeSig string
	incomingTransition := incomingState.GetLastTransition()
	err = h.useStoreInTx(func(tx Store) error {
		signedState := false
		if incomingTransition != nil {
			switch incomingTransition.Type {
			case core.TransitionTypeEscrowDeposit, core.TransitionTypeEscrowWithdraw, core.TransitionTypeMigrate:
				signedState = true
			}
			if incomingTransition.Type.RequiresOpenChannel() {
				userHasOpenChannel, err := tx.CheckOpenChannel(incomingState.UserWallet, incomingState.Asset)
				if err != nil {
					return rpc.Errorf("failed to check open channel: %v", err)
				}
				if !userHasOpenChannel {
					return rpc.Errorf("user has no open channel")
				}
			}
		}
		logger.Debug("processing incoming state", "incomingTransition", incomingTransition.Type.String())

		currentState, err := tx.GetLastUserState(incomingState.UserWallet, incomingState.Asset, signedState)
		if err != nil {
			return rpc.Errorf("failed to get last user state: %v", err)
		}
		// User has no signed previous state
		if currentState == nil {
			logger.Info("no previous signed state found, issuing a void state")
			currentState = core.NewVoidState(incomingState.Asset, incomingState.UserWallet)
		}
		if err := tx.EnsureNoOngoingStateTransitions(incomingState.UserWallet, incomingState.Asset); err != nil {
			return rpc.Errorf("failed to check for ongoing state transitions: %v", err)
		}

		if err := h.stateAdvancer.ValidateTransitions(*currentState, incomingState); err != nil {
			return rpc.Errorf("invalid state transition: %v", err)
		}

		packedState, err := core.PackState(incomingState)
		if err != nil {
			return rpc.Errorf("failed to pack state: %v", err)
		}

		// Validate user's signature
		if incomingState.UserSig == nil {
			return rpc.Errorf("missing incoming state user signature: %v", err)
		}
		userSigBytes, err := hexutil.Decode(*incomingState.UserSig)
		if err != nil {
			return rpc.Errorf("failed to decode incoming state user signature: %v", err)
		}

		sigValidator := h.sigValidators[EcdsaSigValidatorType]
		if err := sigValidator.Verify(incomingState.UserWallet, packedState, userSigBytes); err != nil {
			return rpc.Errorf("invalid incoming state user signature: %v", err)
		}

		// Provide node's signature
		stateHash := crypto.Keccak256Hash(packedState).Bytes()
		_nodeSig, err := h.signer.Sign(stateHash)
		if err != nil {
			return rpc.Errorf("failed to sign incoming state: %v", err)
		}
		nodeSig = _nodeSig.String()
		incomingState.NodeSig = &nodeSig

		if incomingTransition != nil {
			switch incomingTransition.Type {
			case core.TransitionTypeHomeDeposit, core.TransitionTypeHomeWithdrawal, core.TransitionTypeMutualLock:
				// We return Node's signature, the user is expected to submit this on blockchain.
				if err := h.recordTransaction(ctx, tx, incomingState, nil, *incomingTransition); err != nil {
					return rpc.Errorf("failed to record transaction: %v", err)
				}
			case core.TransitionTypeTransferSend:
				newReceiverState, err := h.issueTransferReceiverState(ctx, tx, incomingState)
				if err != nil {
					return rpc.Errorf("failed to issue receiver states: %v", err)
				}

				if err := h.recordTransaction(ctx, tx, incomingState, &newReceiverState, *incomingTransition); err != nil {
					return rpc.Errorf("failed to record transaction: %v", err)
				}
			case core.TransitionTypeEscrowLock: // First step in withdrawal through escrow
				if err := tx.ScheduleInitiateEscrowWithdrawal(incomingState); err != nil {
					return rpc.Errorf("failed to schedule blockchain action: %v", err)
				}
				if err := h.recordTransaction(ctx, tx, incomingState, nil, *incomingTransition); err != nil {
					return rpc.Errorf("failed to record transaction %v", err)
				}
			case core.TransitionTypeEscrowDeposit:
				// We return Node's signature, the user is expected to submit this on blockchain.
				// Optionally schedule blockchain action (finalizeEscrowDeposit) on escrow chain
				if err := h.recordTransaction(ctx, tx, incomingState, nil, *incomingTransition); err != nil {
					return rpc.Errorf("failed to record transaction: %v", err)
				}
				extraState, err := h.issueExtraState(ctx, tx, incomingState)
				if err != nil {
					return rpc.Errorf("failed to issue an extra state: %v", err)
				}
				logger.Info("extra state issued", "userID", extraState.UserWallet, "asset", extraState.Asset, "version", extraState.Version)

			case core.TransitionTypeEscrowWithdraw:
				if err := h.recordTransaction(ctx, tx, incomingState, nil, *incomingTransition); err != nil {
					return rpc.Errorf("failed to record transaction: %v", err)
				}
				extraState, err := h.issueExtraState(ctx, tx, incomingState)
				if err != nil {
					return rpc.Errorf("failed to issue an extra state: %v", err)
				}
				logger.Info("extra state issued", "userID", extraState.UserWallet, "asset", extraState.Asset, "version", extraState.Version)
			case core.TransitionTypeMigrate:
				return rpc.Errorf("transition is not supported yet")
				// extraState, err := h.issueExtraState(ctx, tx, incomingState)
				// if err != nil {
				// 	return rpc.Errorf("failed to issue extra state: %v", err)
				// }
			}
		}

		if err := tx.StoreUserState(incomingState); err != nil {
			return rpc.Errorf("failed to store user state: %v", err)
		}

		return nil
	})
	if err != nil {
		logger.Error("failed to process incoming state", "error", err)
		c.Fail(err, "failed to process incoming state")
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
	logger.Debug("processed incoming state", "incomingVersion", incomingState.Version, "incomingTransition", incomingTransition.Type.String())
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
