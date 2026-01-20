package app_session_v1

import (
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/log"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

// SubmitDepositState processes app session deposit state submissions.
func (h *Handler) SubmitDepositState(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var reqPayload rpc.AppSessionsV1SubmitDepositStateRequest
	if err := c.Request.Payload.Translate(&reqPayload); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	logger.Debug("processing app session deposit request",
		"appSessionID", reqPayload.AppStateUpdate.AppSessionID,
		"version", reqPayload.AppStateUpdate.Version)

	appStateUpd, err := unmapAppStateUpdateV1(&reqPayload.AppStateUpdate)
	if err != nil {
		c.Fail(err, "failed to parse app state update")
		return
	}
	userState, err := unmapStateV1(reqPayload.UserState)
	if err != nil {
		c.Fail(err, "failed to parse user state")
		return
	}

	var nodeSig string
	err = h.useStoreInTx(func(tx Store) error {
		lastTransition := userState.GetLastTransition()
		if lastTransition == nil {
			return rpc.Errorf("user state has no transitions")
		}
		if lastTransition.Type != core.TransitionTypeCommit {
			return rpc.Errorf("user state transition must have 'commit' type, got '%s'", lastTransition.Type.String())
		}

		if lastTransition.Type.RequiresOpenChannel() {
			userHasOpenChannel, err := tx.CheckOpenChannel(userState.UserWallet, userState.Asset)
			if err != nil {
				return rpc.Errorf("failed to check open channel: %v", err)
			}
			if !userHasOpenChannel {
				return rpc.Errorf("user has no open channel")
			}
		}

		if lastTransition.AccountID != appStateUpd.AppSessionID {
			return rpc.Errorf("user state transition account ID '%s' does not match app session ID '%s'",
				lastTransition.AccountID, appStateUpd.AppSessionID)
		}

		// Validate user signature on user state
		if userState.UserSig == nil {
			return rpc.Errorf("missing user signature on user state")
		}

		packedUserState, err := core.PackState(userState)
		if err != nil {
			return rpc.Errorf("failed to pack user state: %v", err)
		}

		userSigBytes, err := hexutil.Decode(*userState.UserSig)
		if err != nil {
			return rpc.Errorf("failed to decode user signature: %v", err)
		}

		sigValidator := h.sigValidator[EcdsaSigType]
		err = sigValidator.Verify(userState.UserWallet, packedUserState, userSigBytes)
		if err != nil {
			return rpc.Errorf("failed to validate signature: %v", err)
		}

		currentState, err := tx.GetLastUserState(userState.UserWallet, userState.Asset, false)
		if err != nil {
			return rpc.Errorf("failed to get last user state: %v", err)
		}
		if currentState == nil {
			currentState = core.NewVoidState(userState.Asset, userState.UserWallet)
		}
		if err := tx.EnsureNoOngoingStateTransitions(userState.UserWallet, userState.Asset); err != nil {
			return rpc.Errorf("ongoing state transitions check failed: %v", err)
		}

		if err := h.stateAdvancer.ValidateAdvancement(*currentState, userState); err != nil {
			return rpc.Errorf("invalid state transitions: %v", err)
		}

		appSession, err := tx.GetAppSession(appStateUpd.AppSessionID)
		if err != nil {
			return rpc.Errorf("app session not found: %v", err)
		}
		if appSession == nil {
			return rpc.Errorf("app session not found")
		}
		if appSession.Status == app.AppSessionStatusClosed {
			return rpc.Errorf("app session is already closed")
		}
		if appStateUpd.Version != appSession.Version+1 {
			return rpc.Errorf("invalid app session version: expected %d, got %d", appSession.Version+1, appStateUpd.Version)
		}

		if appStateUpd.Intent != app.AppStateUpdateIntentDeposit {
			return rpc.Errorf("invalid intent: expected 'deposit', got '%s'", appStateUpd.Intent)
		}

		participantWeights := getParticipantWeights(appSession.Participants)

		if len(reqPayload.AppStateSignatures) == 0 {
			return rpc.Errorf("no signatures provided")
		}

		// Pack the app state update for signature verification
		packedStateUpdate, err := app.PackAppStateUpdateV1(appStateUpd)
		if err != nil {
			return rpc.Errorf("failed to pack app state update: %v", err)
		}

		if err := h.verifyQuorum(participantWeights, appSession.Quorum, packedStateUpdate, reqPayload.AppStateSignatures); err != nil {
			return err
		}

		currentAllocations, err := tx.GetParticipantAllocations(appSession.SessionID)
		if err != nil {
			return rpc.Errorf("failed to get current allocations: %v", err)
		}

		// Track total deposit amount to validate against transition amount
		totalDepositAmount := decimal.Zero

		incomingAllocations := make(map[string]map[string]decimal.Decimal)
		for _, alloc := range appStateUpd.Allocations {
			if alloc.Amount.IsNegative() {
				return rpc.Errorf("negative allocation: %s for asset %s", alloc.Amount, alloc.Asset)
			}

			participantAllocs := currentAllocations[alloc.Participant]
			if participantAllocs == nil {
				participantAllocs = make(map[string]decimal.Decimal, 0)
			}
			currentAmount := participantAllocs[alloc.Asset]

			if alloc.Amount.LessThan(currentAmount) {
				return rpc.Errorf("decreased allocation for %s for participant %s", alloc.Asset, alloc.Participant)
			}

			if alloc.Amount.GreaterThan(currentAmount) {
				// Validate participant
				if _, ok := participantWeights[alloc.Participant]; !ok {
					return rpc.Errorf("allocation to non-participant %s", alloc.Participant)
				}

				// Validate that allocation asset matches user state asset
				if alloc.Asset != userState.Asset {
					return rpc.Errorf("app session deposit allocation for asset '%s' does not match user channel state asset '%s'", alloc.Asset, userState.Asset)
				}

				depositAmount := alloc.Amount.Sub(currentAmount)

				// Accumulate total deposit amount
				totalDepositAmount = totalDepositAmount.Add(depositAmount)

				if err := tx.RecordLedgerEntry(appSession.SessionID, alloc.Asset, depositAmount, nil); err != nil {
					return rpc.Errorf("failed to record ledger entry: %v", err)
				}
			}

			// Store in incoming allocations map
			if incomingAllocations[alloc.Participant] == nil {
				incomingAllocations[alloc.Participant] = make(map[string]decimal.Decimal)
			}
			incomingAllocations[alloc.Participant][alloc.Asset] = alloc.Amount
		}

		// Verify all session balances are accounted for
		for participant, assets := range currentAllocations {
			for asset, currentAmount := range assets {
				if currentAmount.IsZero() {
					continue
				}
				if asset == userState.Asset {
					// Skip asset being deposited to avoid double-checking
					continue
				}

				// Check if this participant+asset is included in the incoming request
				incomingAmount, found := incomingAllocations[participant][asset]
				if !found {
					return rpc.Errorf("deposit intent missing allocation for participant %s, asset %s with current amount %s",
						participant, asset, currentAmount.String())
				}

				// Verify amounts match exactly
				if !incomingAmount.Equal(currentAmount) {
					return rpc.Errorf("deposit intent requires non-deposited asset allocations to match current state: participant %s, asset %s, current %s, provided %s",
						participant, asset, currentAmount.String(), incomingAmount.String())
				}
			}
		}

		// Validate that total deposit amount matches the transition amount
		if !totalDepositAmount.Equal(lastTransition.Amount) {
			return rpc.Errorf("total deposit amount %s does not match transition amount %s", totalDepositAmount.String(), lastTransition.Amount.String())
		}

		// Update app session version
		appSession.Version++
		// Overwrite session data if provided
		if reqPayload.AppStateUpdate.SessionData != "" {
			appSession.SessionData = reqPayload.AppStateUpdate.SessionData
		}
		appSession.UpdatedAt = time.Now()

		if err := tx.UpdateAppSession(*appSession); err != nil {
			return rpc.Errorf("failed to update app session: %v", err)
		}

		// Sign the user state with node's signature
		// TODO:create a function to handle state signing
		userStateHash := crypto.Keccak256Hash(packedUserState).Bytes()
		_nodeSig, err := h.signer.Sign(userStateHash)
		if err != nil {
			return rpc.Errorf("failed to sign user state: %v", err)
		}
		nodeSig = _nodeSig.String()
		userState.NodeSig = &nodeSig

		if err := tx.StoreUserState(userState); err != nil {
			return rpc.Errorf("failed to store user state: %v", err)
		}

		transaction, err := core.NewTransactionFromTransition(&userState, nil, *lastTransition)
		if err != nil {
			return rpc.Errorf("failed to create transaction: %v", err)
		}

		if err := tx.RecordTransaction(*transaction); err != nil {
			return rpc.Errorf("failed to record transaction: %v", err)
		}

		logger.Info("processed deposit state",
			"appSessionID", appSession.SessionID,
			"appSessionVersion", appSession.Version,
			"userWallet", userState.UserWallet,
			"userStateVersion", userState.Version,
			"channelTransition", lastTransition.Type.String())

		return nil
	})

	if err != nil {
		logger.Error("failed to process deposit state", "error", err)
		c.Fail(err, "failed to process deposit state")
		return
	}

	resp := rpc.AppSessionsV1SubmitDepositStateResponse{
		StateNodeSig: nodeSig,
	}

	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
	logger.Info("successfully processed deposit state",
		"appSessionID", reqPayload.AppStateUpdate.AppSessionID,
		"userWallet", userState.UserWallet,
		"userStateVersion", userState.Version,
		"asset", userState.Asset,
		"amount", userState.GetLastTransition().Amount)
}
