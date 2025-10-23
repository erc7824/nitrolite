package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
)

const (
	ErrNegativeAllocation = "negative allocation"
	ErrRecordTransaction  = "failed to record transaction"
)

// UpdateResult represents the result of an app session update operation.
type UpdateResult struct {
	ParticipantsAffected map[string]bool
	UpdatedAppSession    *AppSession
}

// AppSessionUpdater defines the interface for handling different app session update intents.
type AppSessionUpdater interface {
	Update(ctx context.Context, tx *gorm.DB) (UpdateResult, error)
}

// DepositUpdater handles deposit intent updates.
type DepositUpdater struct {
	appSession *AppSession
	params     *SubmitAppStateParams
	rpcWallets map[string]struct{}
	rpcSigners map[string]struct{}
}

// NewDepositUpdater creates a new DepositUpdater.
func NewDepositUpdater(params *SubmitAppStateParams, appSession *AppSession, rpcWallets, rpcSigners map[string]struct{}) (*DepositUpdater, error) {
	return &DepositUpdater{
		appSession: appSession,
		params:     params,
		rpcWallets: rpcWallets,
		rpcSigners: rpcSigners,
	}, nil
}

// Update implements the AppSessionUpdater interface for deposit intents.
func (d *DepositUpdater) Update(ctx context.Context, tx *gorm.DB) (UpdateResult, error) {
	participantWeights, err := verifyQuorum(d.appSession, d.rpcWallets, d.params.Intent)
	if err != nil {
		return UpdateResult{}, err
	}

	sessionAccountID := NewAccountID(d.params.AppSessionID)

	participantsWithUpdatedBalance := make(map[string]bool)
	sessionKeysUsed := make(map[string]bool) // Track which session keys were actually used

	currentAllocations, err := getParticipantAllocations(tx, *d.appSession, sessionAccountID)
	if err != nil {
		return UpdateResult{}, RPCErrorf("failed to get current allocations: %w", err)
	}

	for _, alloc := range d.params.Allocations {
		walletAddress := alloc.Participant
		currentAmount := currentAllocations[walletAddress][alloc.AssetSymbol]
		if alloc.Amount.LessThan(currentAmount) {
			return UpdateResult{}, RPCErrorf("incorrect deposit request: decreased allocation for participant %s", walletAddress)
		}
	}

	noDeposits := true

	for _, alloc := range d.params.Allocations {
		walletAddress := alloc.Participant // ParticipantWallet should always be the main wallet
		currentAmount := currentAllocations[walletAddress][alloc.AssetSymbol]

		if alloc.Amount.GreaterThan(currentAmount) {
			if alloc.Amount.IsNegative() {
				return UpdateResult{}, RPCErrorf(ErrNegativeAllocation+": %s for asset %s", alloc.Amount, alloc.AssetSymbol)
			}

			if err := validateAppParticipant(walletAddress, participantWeights); err != nil {
				return UpdateResult{}, err
			}

			depositAmount := alloc.Amount.Sub(currentAmount)

			// Validate deposit amount is positive
			if depositAmount.LessThanOrEqual(decimal.Zero) {
				return UpdateResult{}, RPCErrorf("invalid deposit amount: %s for asset %s", depositAmount, alloc.AssetSymbol)
			}

			noDeposits = false

			// Check if participant has signed directly or via session key
			// Only apply session key limits if wallet didn't sign directly
			var sessionKeyAddress *string
			isDepositorSig := false

			// Check if participant wallet signed directly
			if _, ok := d.rpcSigners[alloc.Participant]; ok {
				isDepositorSig = true
			} else if _, ok := d.rpcWallets[alloc.Participant]; ok {
				// Check if any of the signers is a session key with spending limits for this wallet
				for signer := range d.rpcSigners {
					if GetWalletBySessionKey(signer) == alloc.Participant {
						sessionKeyAddress = &signer
						break
					}
				}
				isDepositorSig = true
			}

			if !isDepositorSig {
				return UpdateResult{}, RPCErrorf("incorrect deposit request: depositor signature is required")
			}

			userAddress := common.HexToAddress(walletAddress)
			userAccountID := NewAccountID(walletAddress)
			ledger := GetWalletLedger(tx, userAddress)
			balance, err := ledger.Balance(userAccountID, alloc.AssetSymbol)
			if err != nil {
				return UpdateResult{}, RPCErrorf(ErrGetAccountBalance+": %w", err)
			}

			if depositAmount.GreaterThan(balance) {
				return UpdateResult{}, RPCErrorf("incorrect deposit request: insufficient unified balance")
			}

			// Validate session key spending only when wallet didn't sign
			if sessionKeyAddress != nil {
				if err := ValidateSessionKeySpending(tx, *sessionKeyAddress, alloc.AssetSymbol, depositAmount); err != nil {
					return UpdateResult{}, RPCErrorf("session key spending validation failed: %w", err)
				}
				// Track that this session key was actually used
				sessionKeysUsed[*sessionKeyAddress] = true
			}

			if err := ledger.Record(userAccountID, alloc.AssetSymbol, depositAmount.Neg(), sessionKeyAddress); err != nil {
				return UpdateResult{}, RPCErrorf(ErrDebitSourceAccount+": %w", err)
			}
			if err := ledger.Record(sessionAccountID, alloc.AssetSymbol, depositAmount, nil); err != nil {
				return UpdateResult{}, RPCErrorf(ErrCreditDestinationAccount+": %w", err)
			}
			_, err = RecordLedgerTransaction(tx, TransactionTypeAppDeposit, userAccountID, sessionAccountID, alloc.AssetSymbol, depositAmount)
			if err != nil {
				return UpdateResult{}, RPCErrorf(ErrRecordTransaction+": %w", err)
			}

			participantsWithUpdatedBalance[walletAddress] = true
		}
	}

	if noDeposits {
		return UpdateResult{}, RPCErrorf("incorrect deposit request: non-positive allocations sum delta")
	}

	d.appSession.Version++

	// Update session key usage for any session keys that were actually used
	for sessionKey := range sessionKeysUsed {
		if err := UpdateSessionKeyUsage(tx, sessionKey); err != nil {
			return UpdateResult{}, fmt.Errorf("failed to update session key usage for %s: %w", sessionKey, err)
		}
	}

	return UpdateResult{
		ParticipantsAffected: participantsWithUpdatedBalance,
		UpdatedAppSession:    d.appSession,
	}, nil
}

// WithdrawUpdater handles withdraw intent updates.
type WithdrawUpdater struct {
	appSession *AppSession
	params     *SubmitAppStateParams
	rpcWallets map[string]struct{} // Wallet addresses for quorum
}

// NewWithdrawUpdater creates a new WithdrawUpdater.
func NewWithdrawUpdater(params *SubmitAppStateParams, appSession *AppSession, rpcWallets map[string]struct{}) (*WithdrawUpdater, error) {
	return &WithdrawUpdater{
		appSession: appSession,
		params:     params,
		rpcWallets: rpcWallets,
	}, nil
}

// Update implements the AppSessionUpdater interface for withdraw intents.
func (w *WithdrawUpdater) Update(ctx context.Context, tx *gorm.DB) (UpdateResult, error) {
	_, err := verifyQuorum(w.appSession, w.rpcWallets, w.params.Intent)
	if err != nil {
		return UpdateResult{}, err
	}

	sessionAccountID := NewAccountID(w.params.AppSessionID)
	participantsWithUpdatedBalance := make(map[string]bool)

	currentAllocations, err := getParticipantAllocations(tx, *w.appSession, sessionAccountID)
	if err != nil {
		return UpdateResult{}, RPCErrorf("failed to get current allocations: %w", err)
	}

	for _, alloc := range w.params.Allocations {
		walletAddress := alloc.Participant
		currentAmount := currentAllocations[walletAddress][alloc.AssetSymbol]
		if alloc.Amount.GreaterThan(currentAmount) {
			return UpdateResult{}, RPCErrorf("incorrect withdrawal request: increased allocation for participant %s", walletAddress)
		}
	}

	noWithdrawals := true

	for _, alloc := range w.params.Allocations {
		if alloc.Amount.IsNegative() {
			return UpdateResult{}, RPCErrorf(ErrNegativeAllocation+": %s for asset %s", alloc.Amount, alloc.AssetSymbol)
		}

		currentAmount := currentAllocations[alloc.Participant][alloc.AssetSymbol]

		if alloc.Amount.LessThan(currentAmount) {
			withdrawalAmount := currentAmount.Sub(alloc.Amount)
			noWithdrawals = false

			userAddress := common.HexToAddress(alloc.Participant)
			userAccountID := NewAccountID(alloc.Participant)
			ledger := GetWalletLedger(tx, userAddress)

			if err := ledger.Record(sessionAccountID, alloc.AssetSymbol, withdrawalAmount.Neg(), nil); err != nil {
				return UpdateResult{}, RPCErrorf(ErrDebitSourceAccount+": %w", err)
			}
			if err := ledger.Record(userAccountID, alloc.AssetSymbol, withdrawalAmount, nil); err != nil {
				return UpdateResult{}, RPCErrorf(ErrCreditDestinationAccount+": %w", err)
			}
			_, err = RecordLedgerTransaction(tx, TransactionTypeAppWithdrawal, sessionAccountID, userAccountID, alloc.AssetSymbol, withdrawalAmount)
			if err != nil {
				return UpdateResult{}, RPCErrorf(ErrRecordTransaction+": %w", err)
			}

			participantsWithUpdatedBalance[alloc.Participant] = true
		}
	}

	if noWithdrawals {
		return UpdateResult{}, RPCErrorf("incorrect withdrawal request: non-negative allocations sum delta")
	}

	w.appSession.Version++

	return UpdateResult{
		ParticipantsAffected: participantsWithUpdatedBalance,
		UpdatedAppSession:    w.appSession,
	}, nil
}

// OperateUpdater handles operate intent updates.
type OperateUpdater struct {
	appSession *AppSession
	params     *SubmitAppStateParams
	rpcWallets map[string]struct{}
}

// NewOperateUpdater creates a new OperateUpdater.
func NewOperateUpdater(params *SubmitAppStateParams, appSession *AppSession, rpcWallets map[string]struct{}) (*OperateUpdater, error) {
	return &OperateUpdater{
		appSession: appSession,
		params:     params,
		rpcWallets: rpcWallets,
	}, nil
}

// Update implements the AppSessionUpdater interface for operate intents.
func (o *OperateUpdater) Update(ctx context.Context, tx *gorm.DB) (UpdateResult, error) {
	participantWeights, err := verifyQuorum(o.appSession, o.rpcWallets, o.params.Intent)
	if err != nil {
		return UpdateResult{}, err
	}

	sessionAccountID := NewAccountID(o.params.AppSessionID)
	appSessionBalance, err := getAppSessionBalances(tx, sessionAccountID)
	if err != nil {
		return UpdateResult{}, err
	}

	allocationSum := map[string]decimal.Decimal{}
	for _, alloc := range o.params.Allocations {
		if alloc.Amount.IsNegative() {
			return UpdateResult{}, RPCErrorf(ErrNegativeAllocation+": %s for asset %s", alloc.Amount, alloc.AssetSymbol)
		}

		if err := validateAppParticipant(alloc.Participant, participantWeights); err != nil {
			return UpdateResult{}, err
		}

		userAddress := common.HexToAddress(alloc.Participant)
		ledger := GetWalletLedger(tx, userAddress)
		balance, err := ledger.Balance(sessionAccountID, alloc.AssetSymbol)
		if err != nil {
			return UpdateResult{}, RPCErrorf(ErrGetAccountBalance+": %w", err)
		}

		diff := alloc.Amount.Sub(balance)
		if !diff.IsZero() {
			if err := ledger.Record(sessionAccountID, alloc.AssetSymbol, diff, nil); err != nil {
				return UpdateResult{}, RPCErrorf("failed to update session balance: %w", err)
			}
		}

		allocationSum[alloc.AssetSymbol] = allocationSum[alloc.AssetSymbol].Add(alloc.Amount)
	}

	if err := verifyAllocations(appSessionBalance, allocationSum); err != nil {
		return UpdateResult{}, RPCErrorf("incorrect operate request: non-zero allocations sum delta: %w", err)
	}

	o.appSession.Version++

	// Operate intent doesn't affect participant balances for notifications
	return UpdateResult{
		ParticipantsAffected: make(map[string]bool),
		UpdatedAppSession:    o.appSession,
	}, nil
}

// AppSessionService handles the business logic for app sessions.
type AppSessionService struct {
	db         *gorm.DB
	wsNotifier *WSNotifier
}

// NewAppSessionService creates a new AppSessionService.
func NewAppSessionService(db *gorm.DB, wsNotifier *WSNotifier) *AppSessionService {
	return &AppSessionService{db: db, wsNotifier: wsNotifier}
}

func (s *AppSessionService) CreateAppSession(params *CreateAppSessionParams, rpcSigners map[string]struct{}) (AppSessionResponse, error) {
	if !rpc.IsSupportedVersion(params.Definition.Protocol) {
		return AppSessionResponse{}, RPCErrorf("unsupported protocol: %s", params.Definition.Protocol)
	}
	if len(params.Definition.ParticipantWallets) < 2 {
		return AppSessionResponse{}, RPCErrorf("invalid number of participants")
	}
	if len(params.Definition.Weights) != len(params.Definition.ParticipantWallets) {
		return AppSessionResponse{}, RPCErrorf("number of weights must be equal to participants")
	}
	if params.Definition.Nonce == 0 {
		return AppSessionResponse{}, RPCErrorf("nonce is zero or not provided")
	}
	for i, weight := range params.Definition.Weights {
		if weight < 0 {
			return AppSessionResponse{}, RPCErrorf("participant %s weight cannot be negative", params.Definition.ParticipantWallets[i])
		}
	}

	// Validate the weights against quorum
	var totalWeights int64
	for _, weight := range params.Definition.Weights {
		totalWeights += weight
	}
	if params.Definition.Quorum > uint64(totalWeights) {
		return AppSessionResponse{}, RPCErrorf("target quorum (%d) cannot be greater than total sum of weights (%d)", params.Definition.Quorum, totalWeights)
	}

	// Generate a unique ID for the application session
	appBytes, err := json.Marshal(params.Definition)
	if err != nil {
		return AppSessionResponse{}, RPCErrorf("failed to generate app session ID")
	}
	appSessionID := crypto.Keccak256Hash(appBytes).Hex()
	sessionAccountID := NewAccountID(appSessionID)

	participantsWithUpdatedBalance := make(map[string]bool)
	sessionKeysUsed := make(map[string]bool) // Track which session keys were actually used
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, alloc := range params.Allocations {
			if alloc.Amount.IsZero() {
				continue
			}
			if alloc.Amount.IsNegative() {
				return RPCErrorf(ErrNegativeAllocation+": %s for asset %s", alloc.Amount, alloc.AssetSymbol)
			}

			// Check if participant has signed directly or via session key
			// Only apply session key limits if wallet didn't sign directly
			var sessionKeyAddress *string
			signatureProvided := false

			// Check direct signature from participant
			if _, ok := rpcSigners[alloc.Participant]; ok {
				signatureProvided = true
			} else {
				// Check if any signer is a session key with spending limits for this participant
				for signer := range rpcSigners {
					if GetWalletBySessionKey(signer) == alloc.Participant {
						signatureProvided = true
						sessionKeyAddress = &signer
						break
					}
				}
			}

			if alloc.Amount.IsPositive() && !signatureProvided {
				return RPCErrorf("missing signature for participant %s", alloc.Participant)
			}

			// Validate session key spending only when wallet didn't sign
			if sessionKeyAddress != nil {
				if err := ValidateSessionKeySpending(tx, *sessionKeyAddress, alloc.AssetSymbol, alloc.Amount); err != nil {
					return RPCErrorf("session key spending validation failed: %w", err)
				}
				// Track that this session key was actually used
				sessionKeysUsed[*sessionKeyAddress] = true
			}

			walletAddress := alloc.Participant

			if err := checkChallengedChannels(tx, walletAddress); err != nil {
				return err
			}

			userAddress := common.HexToAddress(walletAddress)
			userAccountID := NewAccountID(walletAddress)
			ledger := GetWalletLedger(tx, userAddress)
			balance, err := ledger.Balance(userAccountID, alloc.AssetSymbol)
			if err != nil {
				return RPCErrorf(ErrGetAccountBalance+": %w", err)
			}

			if alloc.Amount.GreaterThan(balance) {
				return RPCErrorf("insufficient funds: %s for asset %s", walletAddress, alloc.AssetSymbol)
			}

			if err = ledger.Record(userAccountID, alloc.AssetSymbol, alloc.Amount.Neg(), sessionKeyAddress); err != nil {
				return RPCErrorf(ErrDebitSourceAccount+": %w", err)
			}
			if err = ledger.Record(sessionAccountID, alloc.AssetSymbol, alloc.Amount, nil); err != nil {
				return RPCErrorf(ErrCreditDestinationAccount+": %w", err)
			}
			_, err = RecordLedgerTransaction(tx, TransactionTypeAppDeposit, userAccountID, sessionAccountID, alloc.AssetSymbol, alloc.Amount)
			if err != nil {
				return RPCErrorf(ErrRecordTransaction+": %w", err)
			}
			participantsWithUpdatedBalance[walletAddress] = true
		}

		session := &AppSession{
			Protocol:           params.Definition.Protocol,
			SessionID:          appSessionID,
			ParticipantWallets: params.Definition.ParticipantWallets,
			Status:             ChannelStatusOpen,
			Challenge:          params.Definition.Challenge,
			Weights:            params.Definition.Weights,
			Quorum:             params.Definition.Quorum,
			Nonce:              params.Definition.Nonce,
			Version:            1,
		}
		if params.SessionData != nil {
			session.SessionData = *params.SessionData
		}

		// Update session key usage for any session keys that were actually used
		for sessionKey := range sessionKeysUsed {
			if err := UpdateSessionKeyUsage(tx, sessionKey); err != nil {
				return fmt.Errorf("failed to update session key usage for %s: %w", sessionKey, err)
			}
		}

		return tx.Create(session).Error
	})

	if err != nil {
		return AppSessionResponse{}, err
	}

	for participant := range participantsWithUpdatedBalance {
		s.wsNotifier.Notify(NewBalanceNotification(participant, s.db))
	}

	return AppSessionResponse{
		AppSessionID: appSessionID,
		Version:      1,
		Status:       string(ChannelStatusOpen),
	}, nil
}

func (s *AppSessionService) SubmitAppState(ctx context.Context, params *SubmitAppStateParams, rpcWallets, rpcSigners map[string]struct{}) (AppSessionResponse, error) {
	participants := make(map[string]bool)
	var updatedAppSession AppSession

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var updater AppSessionUpdater
		var err error

		appSession, err := getAppSession(tx, params.AppSessionID, "open")
		if err != nil {
			return RPCErrorf("an open app session not found: %w", err)
		}

		// Protocol-specific validation
		switch appSession.Protocol {
		case rpc.VersionNitroRPCv0_4:
			if appSession.Version+1 != params.Version {
				return RPCErrorf("incorrect app state: incorrect version: expected %d, got %d", appSession.Version+1, params.Version)
			}
		case rpc.VersionNitroRPCv0_2:
			if params.Intent != "" || params.Version != 0 {
				return RPCErrorf("incorrect request: specified parameters are not supported in this protocol")
			}
		default:
			return RPCErrorf("incorrect app state: unsupported protocol: %s", appSession.Protocol)
		}

		switch params.Intent {
		case rpc.AppSessionIntentDeposit:
			updater, err = NewDepositUpdater(params, appSession, rpcWallets, rpcSigners)
		case rpc.AppSessionIntentWithdraw:
			updater, err = NewWithdrawUpdater(params, appSession, rpcWallets)
		case rpc.AppSessionIntentOperate:
			updater, err = NewOperateUpdater(params, appSession, rpcWallets)
		case "":
			updater, err = NewOperateUpdater(params, appSession, rpcWallets)
		default:
			return RPCErrorf("incorrect app state: unsupported intent: %s", params.Intent)
		}
		if err != nil {
			return err
		}

		result, err := updater.Update(ctx, tx)
		if err != nil {
			return err
		}

		participants = result.ParticipantsAffected
		updatedAppSession = *result.UpdatedAppSession

		if params.SessionData != nil {
			updatedAppSession.SessionData = *params.SessionData
		}

		err = tx.Save(&updatedAppSession).Error
		if err != nil {
			return err
		}

		reloadedSession, err := getAppSession(tx, params.AppSessionID, "")
		if err != nil {
			return RPCErrorf("failed to reload app session after update: %w", err)
		}
		updatedAppSession = *reloadedSession

		return nil
	})

	if err != nil {
		return AppSessionResponse{}, err
	}

	// Notify only participants whose balances were affected by deposit operations
	for participant := range participants {
		s.wsNotifier.Notify(NewBalanceNotification(participant, s.db))
	}

	participantAllocations, err := getParticipantAllocations(s.db, updatedAppSession, NewAccountID(params.AppSessionID))
	if err != nil {
		logger := LoggerFromContext(ctx)
		logger.Error("failed to get participant allocations for app session, notifications will not be sent", "sessionID", updatedAppSession.SessionID, "error", err)
	} else {
		for _, participant := range updatedAppSession.ParticipantWallets {
			s.wsNotifier.Notify(NewAppSessionNotification(participant, updatedAppSession, prepareAppAllocations(participantAllocations)))
		}
	}

	return AppSessionResponse{
		AppSessionID: params.AppSessionID,
		Version:      updatedAppSession.Version,
		Status:       string(ChannelStatusOpen),
	}, nil
}

// CloseApplication closes a virtual app session and redistributes funds to participants
func (s *AppSessionService) CloseApplication(params *CloseAppSessionParams, rpcWallets map[string]struct{}) (AppSessionResponse, error) {
	if params.AppSessionID == "" || len(params.Allocations) == 0 {
		return AppSessionResponse{}, RPCErrorf("missing required parameters: app_id or allocations")
	}

	participantsWithUpdatedBalance := make(map[string]bool)
	var newVersion uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		session, err := getAppSession(tx, params.AppSessionID, "open")
		if err != nil {
			return RPCErrorf("an open app session not found: %w", err)
		}

		participantWeights, err := verifyQuorum(session, rpcWallets, "")
		if err != nil {
			return err
		}
		sessionAccountID := NewAccountID(session.SessionID)

		appSessionBalance, err := getAppSessionBalances(tx, sessionAccountID)
		if err != nil {
			return RPCErrorf("failed to get app session balances: %w", err)
		}

		allocationSum := map[string]decimal.Decimal{}
		for _, alloc := range params.Allocations {
			if alloc.Amount.IsNegative() {
				return RPCErrorf(ErrNegativeAllocation+": %s for asset %s", alloc.Amount, alloc.AssetSymbol)
			}

			walletAddress := alloc.Participant

			if err := validateAppParticipant(walletAddress, participantWeights); err != nil {
				return err
			}

			userAddress := common.HexToAddress(walletAddress)
			userAccountID := NewAccountID(walletAddress)
			ledger := GetWalletLedger(tx, userAddress)
			balance, err := ledger.Balance(sessionAccountID, alloc.AssetSymbol)
			if err != nil {
				return RPCErrorf("failed to get session balance for asset %s", alloc.AssetSymbol)
			}

			// Debit session, credit participant
			if err := ledger.Record(sessionAccountID, alloc.AssetSymbol, balance.Neg(), nil); err != nil {
				return RPCErrorf(ErrDebitSourceAccount+": %w", err)
			}
			if err := ledger.Record(userAccountID, alloc.AssetSymbol, alloc.Amount, nil); err != nil {
				return RPCErrorf(ErrCreditDestinationAccount+": %w", err)
			}
			_, err = RecordLedgerTransaction(tx, TransactionTypeAppWithdrawal, sessionAccountID, userAccountID, alloc.AssetSymbol, alloc.Amount)
			if err != nil {
				return RPCErrorf(ErrRecordTransaction+": %w", err)
			}

			if !alloc.Amount.IsZero() {
				allocationSum[alloc.AssetSymbol] = allocationSum[alloc.AssetSymbol].Add(alloc.Amount)
				participantsWithUpdatedBalance[walletAddress] = true
			}
		}

		if err := verifyAllocations(appSessionBalance, allocationSum); err != nil {
			return err
		}

		newVersion = session.Version + 1
		updates := map[string]any{
			"status":  ChannelStatusClosed,
			"version": newVersion,
		}
		if params.SessionData != nil {
			updates["session_data"] = *params.SessionData
		}

		return tx.Model(&session).Updates(updates).Error
	})

	if err != nil {
		return AppSessionResponse{}, err
	}

	// Notify only participants who received non-zero allocations during session closure
	for participant := range participantsWithUpdatedBalance {
		s.wsNotifier.Notify(NewBalanceNotification(participant, s.db))
	}

	return AppSessionResponse{
		AppSessionID: params.AppSessionID,
		Version:      newVersion,
		Status:       string(ChannelStatusClosed),
	}, nil
}

// GetAppSessions finds all app sessions
// If participantWallet is specified, it returns only sessions for that participant
// If participantWallet is empty, it returns all sessions
func (s *AppSessionService) GetAppSessions(participantWallet string, status string, options *ListOptions) ([]AppSession, error) {
	var sessions []AppSession
	query := s.db.WithContext(context.TODO())
	query = applyListOptions(query, "updated_at", SortTypeDescending, options)

	if participantWallet != "" {
		switch s.db.Dialector.Name() {
		case "postgres":
			query = query.Where("? = ANY(participants)", participantWallet)
		case "sqlite":
			query = query.Where("instr(participants, ?) > 0", participantWallet)
		default:
			return nil, fmt.Errorf("unsupported database driver: %s", s.db.Dialector.Name())
		}
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&sessions).Error; err != nil {
		return nil, err
	}

	return sessions, nil
}

func getParticipantAllocations(db *gorm.DB, appSession AppSession, sessionAccountID AccountID) (map[string]map[string]decimal.Decimal, error) {
	participantAllocations := make(map[string]map[string]decimal.Decimal)

	for _, participant := range appSession.ParticipantWallets {
		participantAllocations[participant] = make(map[string]decimal.Decimal)

		ledger := GetWalletLedger(db, common.HexToAddress(participant))
		balances, err := ledger.GetBalances(sessionAccountID)
		if err != nil {
			return nil, err
		}

		for _, balance := range balances {
			if !balance.Amount.IsZero() {
				participantAllocations[participant][balance.Asset] = balance.Amount
			}
		}
	}

	return participantAllocations, nil
}

// validateAppParticipant checks if wallet exists in participant weights
func validateAppParticipant(walletAddress string, participantWeights map[string]int64) error {
	if _, ok := participantWeights[walletAddress]; !ok {
		return RPCErrorf("allocation to non-participant %s", walletAddress)
	}
	return nil
}

func verifyQuorum(session *AppSession, rpcWallets map[string]struct{}, intent rpc.AppSessionIntent) (map[string]int64, error) {
	participantWeights := make(map[string]int64, len(session.ParticipantWallets))
	for i, addr := range session.ParticipantWallets {
		participantWeights[addr] = session.Weights[i]
	}

	// Track which participants have signed to avoid double-counting
	participantsSigned := make(map[string]struct{})
	var totalWeight int64

	for wallet := range rpcWallets {
		// Check if this wallet is a valid participant
		weight, ok := participantWeights[wallet]
		if !ok {
			return nil, RPCErrorf("signature from unknown participant wallet %s", wallet)
		}

		// Only count each participant once
		if _, alreadyCounted := participantsSigned[wallet]; !alreadyCounted {
			participantsSigned[wallet] = struct{}{}
			totalWeight += weight
		}
	}

	if totalWeight < int64(session.Quorum) {
		err := fmt.Sprintf("quorum not reached: %d / %d", totalWeight, session.Quorum)
		switch intent {
		case rpc.AppSessionIntentDeposit:
			return nil, RPCErrorf("incorrect deposit request: %s", err)
		case rpc.AppSessionIntentWithdraw:
			return nil, RPCErrorf("incorrect withdrawal request: %s", err)
		default:
			return nil, RPCErrorf("incorrect submit_state request: %s", err)
		}

	}

	return participantWeights, nil
}

// prepareAppAllocations converts map format to AppAllocation slice for notifications
func prepareAppAllocations(participantAllocations map[string]map[string]decimal.Decimal) []AppAllocation {
	var allocations []AppAllocation
	for participant, assetMap := range participantAllocations {
		for asset, amount := range assetMap {
			if !amount.IsZero() {
				allocations = append(allocations, AppAllocation{
					Participant: participant,
					AssetSymbol: asset,
					Amount:      amount,
				})
			}
		}
	}
	return allocations
}
