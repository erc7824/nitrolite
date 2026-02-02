package channel_v1

import (
	"context"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/log"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/erc7824/nitrolite/pkg/sign"
)

// Handler manages channel state transitions and provides RPC endpoints for state submission.
type Handler struct {
	useStoreInTx  StoreTxProvider
	memoryStore   MemoryStore
	signer        sign.Signer
	stateAdvancer core.StateAdvancer
	statePacker   core.StatePacker
	sigValidators map[SigValidatorType]SigValidator
	nodeAddress   string // Node's wallet address for channel ID calculation
	minChallenge  uint32
}

// NewHandler creates a new Handler instance with the provided dependencies.
func NewHandler(
	useStoreInTx StoreTxProvider,
	memoryStore MemoryStore,
	signer sign.Signer,
	stateAdvancer core.StateAdvancer,
	statePacker core.StatePacker,
	sigValidators map[SigValidatorType]SigValidator,
	nodeAddress string,
	minChallenge uint32,
) *Handler {
	return &Handler{
		stateAdvancer: stateAdvancer,
		statePacker:   statePacker,
		useStoreInTx:  useStoreInTx,
		memoryStore:   memoryStore,
		signer:        signer,
		sigValidators: sigValidators,
		nodeAddress:   nodeAddress,
		minChallenge:  minChallenge,
	}
}

// issueTransferReceiverState creates and stores a new state for the receiver of a transfer.
// It reads the receiver's current state, applies a transfer_receive transition with the same
// amount and tx hash, signs it with the node's key, and persists it.
func (h *Handler) issueTransferReceiverState(ctx context.Context, tx Store, senderState core.State) (*core.State, error) {
	logger := log.FromContext(ctx)

	incomingTransition := senderState.GetLastTransition()
	if incomingTransition == nil {
		return nil, rpc.Errorf("incoming state has no transitions")
	}
	if incomingTransition.Type != core.TransitionTypeTransferSend {
		return nil, rpc.Errorf("incoming state doesn't have 'transfer_send' transition")
	}
	receiverWallet := incomingTransition.AccountID
	logger = logger.
		WithKV("sender", senderState.UserWallet).
		WithKV("receiver", receiverWallet).
		WithKV("asset", senderState.Asset)

	logger.Debug("issuing transfer receiver state")

	currentState, err := tx.GetLastUserState(receiverWallet, senderState.Asset, false)
	if err != nil {
		return nil, rpc.Errorf("failed to get last %s user state for transfer receiver with address %s", senderState.Asset, incomingTransition.AccountID)
	}
	if currentState == nil {
		currentState = core.NewVoidState(senderState.Asset, receiverWallet)
	}
	newState := currentState.NextState()

	_, err = newState.ApplyTransferReceiveTransition(
		senderState.UserWallet,
		incomingTransition.Amount,
		incomingTransition.TxID)
	if err != nil {
		return nil, err
	}

	lastSignedState, err := tx.GetLastUserState(receiverWallet, senderState.Asset, true)
	if err != nil {
		return nil, rpc.Errorf("failed to get last %s user state for transfer receiver with address %s", senderState.Asset, incomingTransition.AccountID)
	}

	// TODO: move to DB query
	shouldSign := true
	if lastSignedState != nil {
		if lastSignedState.HomeChannelID != nil {
			shouldSign = false
		}

		lastStateTransition := lastSignedState.GetLastTransition()
		if lastStateTransition != nil {
			if lastStateTransition.Type == core.TransitionTypeMutualLock ||
				lastStateTransition.Type == core.TransitionTypeEscrowLock {
				shouldSign = false
			}
		}
	}

	if shouldSign {
		packedState, err := h.statePacker.PackState(*newState)
		if err != nil {
			return nil, rpc.Errorf("failed to pack receiver state: %v", err)
		}

		_nodeSig, err := h.signer.Sign(packedState)
		if err != nil {
			return nil, rpc.Errorf("failed to sign receiver state")
		}
		nodeSig := _nodeSig.String()
		newState.NodeSig = &nodeSig
	}
	if err := tx.StoreUserState(*newState); err != nil {
		return nil, rpc.Errorf("failed to store receiver state")
	}

	logger.Info("issued transfer receiver state", "receiverStateVersion", newState.Version)
	return newState, nil
}

// issueExtraState creates an additional state by reapplying unsigned transitions to a newly signed state.
// When a user submits a signed state (e.g., after escrow_deposit or escrow_withdraw), any pending
// unsigned transitions from the previous state are reapplied to create a new unsigned state.
// This ensures that pending operations are preserved across state updates that require user signatures.
func (h *Handler) issueExtraState(ctx context.Context, tx Store, incomingState core.State, extraTransitions []core.Transition) (*core.State, error) {
	logger := log.FromContext(ctx)

	lastTransition := incomingState.GetLastTransition()
	if lastTransition == nil {
		return nil, rpc.Errorf("incoming state has no transitions")
	}

	if len(extraTransitions) == 0 {
		return &incomingState, nil
	}

	logger = logger.
		WithKV("userWallet", incomingState.UserWallet).
		WithKV("asset", incomingState.Asset)

	extraState := incomingState.NextState()
	logger.Debug("issuing extra state", "extraStateVersion", extraState.Version)

	if err := extraState.ApplyReceiverTransitions(extraTransitions...); err != nil {
		return nil, err
	}

	packedState, err := h.statePacker.PackState(*extraState)
	if err != nil {
		return nil, rpc.Errorf("failed to pack extra state")
	}

	_nodeSig, err := h.signer.Sign(packedState)
	if err != nil {
		return nil, rpc.Errorf("failed to sign extra state")
	}
	nodeSig := _nodeSig.String()
	extraState.NodeSig = &nodeSig

	if err := tx.StoreUserState(*extraState); err != nil {
		return nil, rpc.Errorf("failed to store extra state")
	}

	logger.Info("issued extra state", "extraStateVersion", extraState.Version)
	return extraState, nil
}
