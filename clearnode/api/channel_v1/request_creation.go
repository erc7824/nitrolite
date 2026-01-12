package channel_v1

import (
	"time"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/log"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// RequestCreation processes channel creation requests from users.
// It validates the channel definition and initial state, checks for existing channels,
// signs the state with the node's key, and persists the new pending state and channel.
func (h *Handler) RequestCreation(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var reqPayload rpc.ChannelsV1RequestCreationRequest
	if err := c.Request.Payload.Translate(&reqPayload); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	incomingState, err := toCoreState(reqPayload.State)
	if err != nil {
		c.Fail(err, "failed to parse state")
		return
	}

	channelDef, err := toCoreChannelDefinition(reqPayload.ChannelDefinition)
	if err != nil {
		c.Fail(err, "failed to parse channel definition")
		return
	}

	logger = logger.
		WithKV("userWallet", incomingState.UserWallet).
		WithKV("asset", incomingState.Asset)

	var nodeSig string
	err = h.useStoreInTx(func(tx Store) error {
		// Check if channel already exists
		currentState, err := tx.GetLastUserState(incomingState.UserWallet, incomingState.Asset, false)
		if err != nil {
			return rpc.Errorf("failed to check existing channel: %v", err)
		}
		// User has no signed previous state
		if currentState == nil {
			logger.Debug("no previous signed state found, issuing a void state")
			currentState = core.NewVoidState(incomingState.Asset, incomingState.UserWallet)
		}

		if channelDef.Nonce == 0 {
			return rpc.Errorf("nonce must be non-zero")
		}
		if channelDef.Challenge < h.minChallenge {
			return rpc.Errorf("challenge period must be non-zero")
		}
		logger.Debug("processing channel creation request", "incomingVersion", incomingState.Version)

		if err := h.stateAdvancer.ValidateTransitions(*currentState, incomingState); err != nil {
			return rpc.Errorf("invalid state: %v", err)
		}

		// Pack and validate user signature
		packedState, err := core.PackState(incomingState)
		if err != nil {
			return rpc.Errorf("failed to pack state: %v", err)
		}

		if incomingState.UserSig == nil {
			return rpc.Errorf("missing user signature")
		}
		userSigBytes, err := hexutil.Decode(*incomingState.UserSig)
		if err != nil {
			return rpc.Errorf("failed to decode user signature: %v", err)
		}

		sigValidator := h.sigValidators[EcdsaSigValidatorType]
		if err := sigValidator.Verify(incomingState.UserWallet, packedState, userSigBytes); err != nil {
			return rpc.Errorf("invalid user signature: %v", err)
		}

		// Calculate home channel ID
		homeChannelID, err := core.GetHomeChannelID(
			h.nodeAddress,
			incomingState.UserWallet,
			incomingState.HomeLedger.TokenAddress,
			channelDef.Nonce,
			channelDef.Challenge,
		)
		if err != nil {
			return rpc.Errorf("failed to calculate channel ID: %v", err)
		}

		// Set the home channel ID in the state
		incomingState.HomeChannelID = &homeChannelID

		newHomeChannel := core.NewHomeChannel(
			homeChannelID,
			incomingState.UserWallet,
			h.nodeAddress,
			incomingState.HomeLedger.BlockchainID,
			incomingState.HomeLedger.TokenAddress,
			channelDef.Nonce,
			channelDef.Challenge,
		)

		// Create the home channel entity
		if err := tx.CreateHomeChannel(*newHomeChannel); err != nil {
			return rpc.Errorf("failed to create channel: %v", err)
		}

		// Provide node's signature
		stateHash := crypto.Keccak256Hash(packedState).Bytes()
		_nodeSig, err := h.signer.Sign(stateHash)
		if err != nil {
			return rpc.Errorf("failed to sign state: %v", err)
		}
		nodeSig = _nodeSig.String()
		incomingState.NodeSig = &nodeSig

		incomingTransition := incomingState.GetLastTransition()
		if incomingTransition != nil {
			switch incomingTransition.Type {
			case core.TransitionTypeHomeDeposit, core.TransitionTypeHomeWithdrawal:
				// We return Node's signature, the user is expected to submit this on blockchain.
				if err := h.recordTransaction(ctx, tx, incomingState, nil, *incomingTransition); err != nil {
					return rpc.Errorf("failed to record transaction: %v", err)
				}
			default:
				return rpc.Errorf("transition '%s' is not supported by this endpoint", incomingTransition.Type.String())
			}
		}

		// Store the pending state
		if err := tx.StoreUserState(incomingState); err != nil {
			return rpc.Errorf("failed to store state: %v", err)
		}

		logger.Info("channel creation request processed",
			"homeChannelID", homeChannelID,
			"nonce", channelDef.Nonce,
			"challengeDuration", time.Duration(channelDef.Challenge)*time.Second,
			"incomingVersion", incomingState.Version)
		return nil
	})
	if err != nil {
		logger.Error("failed to process channel creation request", "error", err)
		c.Fail(err, "failed to process channel creation request")
		return
	}

	resp := rpc.ChannelsV1RequestCreationResponse{
		Signature: nodeSig,
	}
	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
	logger.Debug("channel creation request completed", "userWallet", incomingState.UserWallet, "asset", incomingState.Asset)
}
