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

// SubmitAppState processes app session state updates for operate, withdraw, and close intents.
// Deposit intents should use the SubmitDepositState endpoint instead.
func (h *Handler) SubmitAppState(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var reqPayload rpc.AppSessionsV1SubmitAppStateRequest
	if err := c.Request.Payload.Translate(&reqPayload); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	logger.Debug("processing app session state update request",
		"appSessionID", reqPayload.AppStateUpdate.AppSessionID,
		"version", reqPayload.AppStateUpdate.Version,
		"intent", reqPayload.AppStateUpdate.Intent)

	appStateUpd, err := unmapAppStateUpdateV1(&reqPayload.AppStateUpdate)
	if err != nil {
		c.Fail(err, "failed to parse app state update")
		return
	}

	// Ensure this is not a deposit intent (should use submit_deposit_state)
	if appStateUpd.Intent == app.AppStateUpdateIntentDeposit {
		c.Fail(rpc.Errorf("deposit intent must use submit_deposit_state endpoint"), "")
		return
	}

	// Validate intent is valid
	if appStateUpd.Intent != app.AppStateUpdateIntentOperate &&
		appStateUpd.Intent != app.AppStateUpdateIntentWithdraw &&
		appStateUpd.Intent != app.AppStateUpdateIntentClose {
		c.Fail(rpc.Errorf("invalid intent: %s", appStateUpd.Intent.String()), "")
		return
	}

	err = h.useStoreInTx(func(tx Store) error {
		appSession, err := tx.GetAppSession(appStateUpd.AppSessionID)
		if err != nil {
			return rpc.Errorf("app session not found: %v", err)
		}
		if appSession == nil {
			return rpc.Errorf("app session not found")
		}

		if appSession.IsClosed {
			return rpc.Errorf("app session is already closed")
		}

		if appStateUpd.Version != appSession.Version+1 {
			return rpc.Errorf("invalid app session version: expected %d, got %d", appSession.Version+1, appStateUpd.Version)
		}

		participantWeights := getParticipantWeights(appSession.Participants)

		// Validate signatures and quorum
		if len(reqPayload.Signatures) == 0 {
			return rpc.Errorf("no signatures provided")
		}

		// Pack the app state update for signature verification
		packedStateUpdate, err := app.PackAppStateUpdateV1(appStateUpd)
		if err != nil {
			return rpc.Errorf("failed to pack app state update: %v", err)
		}

		// Verify signatures and calculate quorum
		sigRecoverer := h.sigValidator[EcdsaSigType]
		signedWeights := make(map[string]bool)
		var achievedQuorum uint8

		for _, sigHex := range reqPayload.Signatures {
			sigBytes, err := hexutil.Decode(sigHex)
			if err != nil {
				return rpc.Errorf("failed to decode signature: %v", err)
			}

			// Recover the signer address from the signature
			signerAddress, err := sigRecoverer.Recover(packedStateUpdate, sigBytes)
			if err != nil {
				return rpc.Errorf("failed to recover signer address: %v", err)
			}

			// Check if signer is a participant
			weight, isParticipant := participantWeights[signerAddress]
			if !isParticipant {
				return rpc.Errorf("signature from non-participant: %s", signerAddress)
			}

			// Add weight if not already counted
			if !signedWeights[signerAddress] {
				signedWeights[signerAddress] = true
				achievedQuorum += weight
			}
		}

		// Check if quorum is met
		if achievedQuorum < appSession.Quorum {
			return rpc.Errorf("quorum not met: achieved %d, required %d", achievedQuorum, appSession.Quorum)
		}

		currentAllocations, err := tx.GetParticipantAllocations(appSession.SessionID)
		if err != nil {
			return rpc.Errorf("failed to get current allocations: %v", err)
		}

		// Handle different intents
		switch appStateUpd.Intent {
		case app.AppStateUpdateIntentOperate:
			// For operate intent, total allocations per asset must match session balance (redistribution allowed)
			if err := h.handleOperateIntent(tx, appStateUpd, currentAllocations, participantWeights); err != nil {
				return err
			}

		case app.AppStateUpdateIntentWithdraw:
			// For withdraw intent, validate and record ledger changes
			if err := h.handleWithdrawIntent(tx, appStateUpd, currentAllocations, participantWeights); err != nil {
				return err
			}

		case app.AppStateUpdateIntentClose:
			// For close intent, validate final allocations and mark session as closed
			if err := h.handleCloseIntent(tx, appStateUpd, currentAllocations, participantWeights); err != nil {
				return err
			}
			appSession.IsClosed = true
		}

		// Update app session version and data
		appSession.Version++
		if reqPayload.AppStateUpdate.SessionData != "" {
			appSession.SessionData = reqPayload.AppStateUpdate.SessionData
		}
		appSession.UpdatedAt = time.Now()

		if err := tx.UpdateAppSession(*appSession); err != nil {
			return rpc.Errorf("failed to update app session: %v", err)
		}

		logger.Info("processed app state update",
			"appSessionID", appSession.SessionID,
			"appSessionVersion", appSession.Version,
			"intent", appStateUpd.Intent.String(),
			"isClosed", appSession.IsClosed)

		return nil
	})

	if err != nil {
		logger.Error("failed to process app state update", "error", err)
		c.Fail(err, "failed to process app state update")
		return
	}

	resp := rpc.AppSessionsV1SubmitAppStateResponse{
		Signature: "", // No user state signature needed for these intents
	}

	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
	logger.Info("successfully processed app state update",
		"appSessionID", reqPayload.AppStateUpdate.AppSessionID,
		"intent", appStateUpd.Intent.String())
}

// handleOperateIntent processes operate intent by validating total allocations and recording ledger changes.
// Operate intent allows redistribution of funds between participants as long as the total per asset stays the same.
func (h *Handler) handleOperateIntent(
	tx Store,
	appStateUpd app.AppStateUpdateV1,
	currentAllocations map[string]map[string]decimal.Decimal,
	participantWeights map[string]uint8,
) error {
	// Get session balances to verify total allocations
	sessionBalances, err := tx.GetAppSessionBalances(appStateUpd.AppSessionID)
	if err != nil {
		return rpc.Errorf("failed to get app session balances: %v", err)
	}

	// Track total allocations per asset
	allocationSum := make(map[string]decimal.Decimal)

	for _, alloc := range appStateUpd.Allocations {
		// Validate participant exists
		if _, ok := participantWeights[alloc.Participant]; !ok {
			return rpc.Errorf("allocation to non-participant %s", alloc.Participant)
		}

		if alloc.Amount.IsNegative() {
			return rpc.Errorf("negative allocation: %s for asset %s", alloc.Amount, alloc.Asset)
		}

		// Sum up allocations per asset
		if existing, ok := allocationSum[alloc.Asset]; ok {
			allocationSum[alloc.Asset] = existing.Add(alloc.Amount)
		} else {
			allocationSum[alloc.Asset] = alloc.Amount
		}

		// Get current allocation for this participant and asset
		participantAllocs := currentAllocations[alloc.Participant]
		if participantAllocs == nil {
			participantAllocs = make(map[string]decimal.Decimal)
		}
		currentAmount := participantAllocs[alloc.Asset]

		// Calculate the difference and record ledger entry if changed
		diff := alloc.Amount.Sub(currentAmount)
		if !diff.IsZero() {
			if err := tx.RecordLedgerEntry(appStateUpd.AppSessionID, alloc.Asset, diff, nil); err != nil {
				return rpc.Errorf("failed to record operate ledger entry: %v", err)
			}
		}
	}

	// Verify that total allocations per asset match session balances
	for asset, totalAlloc := range allocationSum {
		sessionBalance, ok := sessionBalances[asset]
		if !ok {
			sessionBalance = decimal.Zero
		}

		if !totalAlloc.Equal(sessionBalance) {
			return rpc.Errorf("operate intent allocation mismatch for asset %s: total allocations %s, session balance %s",
				asset, totalAlloc.String(), sessionBalance.String())
		}
	}

	// Verify all session balances are accounted for
	for asset, sessionBalance := range sessionBalances {
		if sessionBalance.IsZero() {
			continue
		}

		totalAlloc, ok := allocationSum[asset]
		if !ok {
			return rpc.Errorf("operate intent missing allocations for asset %s with balance %s",
				asset, sessionBalance.String())
		}

		if !totalAlloc.Equal(sessionBalance) {
			return rpc.Errorf("operate intent allocation mismatch for asset %s: total allocations %s, session balance %s",
				asset, totalAlloc.String(), sessionBalance.String())
		}
	}

	return nil
}

// handleWithdrawIntent processes withdraw intent by validating and recording ledger changes.
// It also issues new channel states for participants receiving withdrawn funds.
func (h *Handler) handleWithdrawIntent(
	tx Store,
	appStateUpd app.AppStateUpdateV1,
	currentAllocations map[string]map[string]decimal.Decimal,
	participantWeights map[string]uint8,
) error {
	for _, alloc := range appStateUpd.Allocations {
		// Validate participant exists
		if _, ok := participantWeights[alloc.Participant]; !ok {
			return rpc.Errorf("allocation to non-participant %s", alloc.Participant)
		}

		if alloc.Amount.IsNegative() {
			return rpc.Errorf("negative allocation: %s for asset %s", alloc.Amount, alloc.Asset)
		}

		participantAllocs := currentAllocations[alloc.Participant]
		if participantAllocs == nil {
			participantAllocs = make(map[string]decimal.Decimal)
		}
		currentAmount := participantAllocs[alloc.Asset]

		// For withdraw, amounts can only decrease or stay the same
		if alloc.Amount.GreaterThan(currentAmount) {
			return rpc.Errorf("withdraw intent cannot increase allocations: participant %s, asset %s",
				alloc.Participant, alloc.Asset)
		}

		if alloc.Amount.LessThan(currentAmount) {
			// Record the withdrawal (negative ledger entry for the session)
			withdrawAmount := currentAmount.Sub(alloc.Amount)
			if err := tx.RecordLedgerEntry(appStateUpd.AppSessionID, alloc.Asset, withdrawAmount.Neg(), nil); err != nil {
				return rpc.Errorf("failed to record withdrawal ledger entry: %v", err)
			}

			// Issue new channel state for participant receiving withdrawn funds
			if err := h.issueReleaseReceiverState(tx, alloc.Participant, alloc.Asset, appStateUpd.AppSessionID, withdrawAmount); err != nil {
				return rpc.Errorf("failed to issue release state for participant %s: %v", alloc.Participant, err)
			}
		}
	}

	return nil
}

// handleCloseIntent processes close intent by validating final allocations and recording ledger changes.
func (h *Handler) handleCloseIntent(
	tx Store,
	appStateUpd app.AppStateUpdateV1,
	currentAllocations map[string]map[string]decimal.Decimal,
	participantWeights map[string]uint8,
) error {
	// Get total balances in the session
	sessionBalances, err := tx.GetAppSessionBalances(appStateUpd.AppSessionID)
	if err != nil {
		return rpc.Errorf("failed to get app session balances: %v", err)
	}

	// Track total allocations per asset
	totalAllocations := make(map[string]decimal.Decimal)

	for _, alloc := range appStateUpd.Allocations {
		// Validate participant exists
		if _, ok := participantWeights[alloc.Participant]; !ok {
			return rpc.Errorf("allocation to non-participant %s", alloc.Participant)
		}

		if alloc.Amount.IsNegative() {
			return rpc.Errorf("negative allocation: %s for asset %s", alloc.Amount, alloc.Asset)
		}

		// Accumulate total allocations per asset
		if existing, ok := totalAllocations[alloc.Asset]; ok {
			totalAllocations[alloc.Asset] = existing.Add(alloc.Amount)
		} else {
			totalAllocations[alloc.Asset] = alloc.Amount
		}
	}

	// Validate that total allocations equal total session balances for each asset
	for asset, totalAlloc := range totalAllocations {
		sessionBalance, ok := sessionBalances[asset]
		if !ok {
			sessionBalance = decimal.Zero
		}

		if !totalAlloc.Equal(sessionBalance) {
			return rpc.Errorf("close intent allocation mismatch for asset %s: total allocations %s, session balance %s",
				asset, totalAlloc.String(), sessionBalance.String())
		}
	}

	// Verify all session balances are accounted for
	for asset, sessionBalance := range sessionBalances {
		if sessionBalance.IsZero() {
			continue
		}

		totalAlloc, ok := totalAllocations[asset]
		if !ok {
			return rpc.Errorf("close intent missing allocations for asset %s with balance %s",
				asset, sessionBalance.String())
		}

		if !totalAlloc.Equal(sessionBalance) {
			return rpc.Errorf("close intent allocation mismatch for asset %s: total allocations %s, session balance %s",
				asset, totalAlloc.String(), sessionBalance.String())
		}
	}

	// Record final withdrawals for each participant and issue channel states
	for _, alloc := range appStateUpd.Allocations {
		participantAllocs := currentAllocations[alloc.Participant]
		if participantAllocs == nil {
			participantAllocs = make(map[string]decimal.Decimal)
		}
		currentAmount := participantAllocs[alloc.Asset]

		// Calculate the difference (should be zero or negative since we're closing)
		if alloc.Amount.LessThan(currentAmount) {
			withdrawAmount := currentAmount.Sub(alloc.Amount)
			if err := tx.RecordLedgerEntry(appStateUpd.AppSessionID, alloc.Asset, withdrawAmount.Neg(), nil); err != nil {
				return rpc.Errorf("failed to record close ledger entry: %v", err)
			}

			// Issue new channel state for participant receiving funds on close
			if err := h.issueReleaseReceiverState(tx, alloc.Participant, alloc.Asset, appStateUpd.AppSessionID, withdrawAmount); err != nil {
				return rpc.Errorf("failed to issue release state for participant %s: %v", alloc.Participant, err)
			}
		} else if alloc.Amount.GreaterThan(currentAmount) {
			// This shouldn't happen in close, but let's be explicit
			return rpc.Errorf("close intent cannot increase allocations: participant %s, asset %s",
				alloc.Participant, alloc.Asset)
		}
	}

	return nil
}

// issueReleaseReceiverState creates a new channel state for a participant receiving funds from app session.
// This follows the same pattern as issueTransferReceiverState in channel_v1 for transfer_receive transitions.
func (h *Handler) issueReleaseReceiverState(tx Store, receiverWallet, asset, appSessionID string, amount decimal.Decimal) error {
	// Get the receiver's current state (or create void state if none exists)
	currentState, err := tx.GetLastUserState(receiverWallet, asset, false)
	if err != nil {
		return rpc.Errorf("failed to get receiver state: %v", err)
	}
	if currentState == nil {
		currentState = core.NewVoidState(asset, receiverWallet)
	}

	// Create next state and apply release transition
	newState := currentState.NextState()
	_, err = newState.ApplyReleaseTransition(appSessionID, amount)
	if err != nil {
		return rpc.Errorf("failed to apply release transition: %v", err)
	}

	// Check if we need to sign the state (skip signing if last signed state was a lock)
	lastSignedState, err := tx.GetLastUserState(receiverWallet, asset, true)
	if err != nil {
		return rpc.Errorf("failed to get last signed state: %v", err)
	}

	shouldSign := true
	if lastSignedState != nil {
		lastStateTransition := lastSignedState.GetLastTransition()
		if lastStateTransition != nil {
			if lastStateTransition.Type == core.TransitionTypeMutualLock ||
				lastStateTransition.Type == core.TransitionTypeEscrowLock {
				shouldSign = false
			}
		}
	}

	if shouldSign {
		// Pack and sign the state
		packedState, err := core.PackState(*newState)
		if err != nil {
			return rpc.Errorf("failed to pack receiver state: %v", err)
		}

		stateHash := crypto.Keccak256Hash(packedState).Bytes()
		nodeSig, err := h.signer.Sign(stateHash)
		if err != nil {
			return rpc.Errorf("failed to sign receiver state: %v", err)
		}

		nodeSigStr := nodeSig.String()
		newState.NodeSig = &nodeSigStr
	}

	// Store the new state
	if err := tx.StoreUserState(*newState); err != nil {
		return rpc.Errorf("failed to store receiver state: %v", err)
	}

	return nil
}
