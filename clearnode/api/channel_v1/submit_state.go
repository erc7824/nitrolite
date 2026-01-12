package channel_v1

import (
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/log"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

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
			default:
				return rpc.Errorf("transition '%s' is not supported by this endpoint", incomingTransition.Type.String())
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
