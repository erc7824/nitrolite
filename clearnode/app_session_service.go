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

// AppSessionService handles the business logic for app sessions.
type AppSessionService struct {
	db         *gorm.DB
	wsNotifier *WSNotifier
}

// NewAppSessionService creates a new AppSessionService.
func NewAppSessionService(db *gorm.DB, wsNotifier *WSNotifier) *AppSessionService {
	return &AppSessionService{db: db, wsNotifier: wsNotifier}
}

func (s *AppSessionService) CreateApplication(params *CreateAppSessionParams, rpcSigners map[string]struct{}) (AppSessionResponse, error) {
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

	// Generate a unique ID for the virtual application
	appBytes, err := json.Marshal(params.Definition)
	if err != nil {
		return AppSessionResponse{}, RPCErrorf("failed to generate app session ID")
	}
	appSessionID := crypto.Keccak256Hash(appBytes).Hex()
	sessionAccountID := NewAccountID(appSessionID)

	participants := make(map[string]bool)
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, alloc := range params.Allocations {
			if alloc.Amount.IsPositive() {
				if _, ok := rpcSigners[alloc.ParticipantWallet]; !ok {
					return RPCErrorf("missing signature for participant %s", alloc.ParticipantWallet)
				}
			}
			if alloc.Amount.IsNegative() {
				return RPCErrorf("negative allocation: %s for asset %s", alloc.Amount, alloc.AssetSymbol)
			}
			walletAddress := alloc.ParticipantWallet
			if wallet := GetWalletBySigner(alloc.ParticipantWallet); wallet != "" {
				walletAddress = wallet
			}

			if err := checkChallengedChannels(tx, walletAddress); err != nil {
				return err
			}

			userAddress := common.HexToAddress(walletAddress)
			userAccountID := NewAccountID(walletAddress)
			ledger := GetWalletLedger(tx, userAddress)
			balance, err := ledger.Balance(userAccountID, alloc.AssetSymbol)
			if err != nil {
				return RPCErrorf("failed to check participant balance: %w", err)
			}

			if alloc.Amount.GreaterThan(balance) {
				return RPCErrorf("insufficient funds: %s for asset %s", walletAddress, alloc.AssetSymbol)
			}

			if err = ledger.Record(userAccountID, alloc.AssetSymbol, alloc.Amount.Neg()); err != nil {
				return RPCErrorf("failed to debit source account: %w", err)
			}
			if err = ledger.Record(sessionAccountID, alloc.AssetSymbol, alloc.Amount); err != nil {
				return RPCErrorf("failed to credit destination account: %w", err)
			}
			_, err = RecordLedgerTransaction(tx, TransactionTypeAppDeposit, userAccountID, sessionAccountID, alloc.AssetSymbol, alloc.Amount)
			if err != nil {
				return RPCErrorf("failed to record transaction: %w", err)
			}
			participants[walletAddress] = true
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

		return tx.Create(session).Error
	})

	if err != nil {
		return AppSessionResponse{}, err
	}

	for participant := range participants {
		s.wsNotifier.Notify(NewBalanceNotification(participant, s.db))
	}

	return AppSessionResponse{
		AppSessionID: appSessionID,
		Version:      1,
		Status:       string(ChannelStatusOpen),
	}, nil
}

func (s *AppSessionService) SubmitAppState(params *SubmitAppStateParams, rpcSigners map[string]struct{}) (AppSessionResponse, error) {
	var newVersion uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		appSession, participantWeights, err := verifyQuorum(tx, params.AppSessionID, rpcSigners)
		if err != nil {
			return err
		}
		sessionAccountID := NewAccountID(appSession.SessionID)

		newVersion = appSession.Version + 1
		switch appSession.Protocol {
		case rpc.VersionNitroRPCv0_2:
			// nothing additional to check
		case rpc.VersionNitroRPCv0_4:
			if newVersion != params.Version {
				return RPCErrorf("invalid version: expected %d, got %d", newVersion, params.Version)
			}

			switch params.Intent {
			case rpc.AppSessionIntentOperate:
				// no additional actions needed
			default:
				return RPCErrorf("unsupported intent: %s", params.Intent)
			}
		default:
			return RPCErrorf("unsupported protocol: %s", appSession.Protocol)
		}

		appSessionBalance, err := getAppSessionBalances(tx, sessionAccountID)
		if err != nil {
			return err
		}

		allocationSum := map[string]decimal.Decimal{}
		for _, alloc := range params.Allocations {
			if alloc.Amount.IsNegative() {
				return RPCErrorf("negative allocation: %s for asset %s", alloc.Amount, alloc.AssetSymbol)
			}

			walletAddress := GetWalletBySigner(alloc.ParticipantWallet)
			if walletAddress == "" {
				walletAddress = alloc.ParticipantWallet
			}

			if _, ok := participantWeights[walletAddress]; !ok {
				return RPCErrorf("allocation to non-participant %s", walletAddress)
			}

			userAddress := common.HexToAddress(walletAddress)
			ledger := GetWalletLedger(tx, userAddress)
			balance, err := ledger.Balance(sessionAccountID, alloc.AssetSymbol)
			if err != nil {
				return RPCErrorf("failed to get session balance for asset %s", alloc.AssetSymbol)
			}

			// Reset participant allocation in app session to the new amount
			if err := ledger.Record(sessionAccountID, alloc.AssetSymbol, balance.Neg()); err != nil {
				return RPCErrorf("failed to debit session: %w", err)
			}
			if err := ledger.Record(sessionAccountID, alloc.AssetSymbol, alloc.Amount); err != nil {
				return RPCErrorf("failed to credit participant: %w", err)
			}

			if !alloc.Amount.IsZero() {
				allocationSum[alloc.AssetSymbol] = allocationSum[alloc.AssetSymbol].Add(alloc.Amount)
			}
		}

		if err := verifyAllocations(appSessionBalance, allocationSum); err != nil {
			return err
		}

		updates := map[string]any{
			"version": newVersion,
		}
		if params.SessionData != nil {
			updates["session_data"] = *params.SessionData
		}

		return tx.Model(&appSession).Updates(updates).Error
	})

	if err != nil {
		return AppSessionResponse{}, err
	}

	return AppSessionResponse{
		AppSessionID: params.AppSessionID,
		Version:      newVersion,
		Status:       string(ChannelStatusOpen),
	}, nil
}

// CloseApplication closes a virtual app session and redistributes funds to participants
func (s *AppSessionService) CloseApplication(params *CloseAppSessionParams, rpcSigners map[string]struct{}) (AppSessionResponse, error) {
	if params.AppSessionID == "" || len(params.Allocations) == 0 {
		return AppSessionResponse{}, RPCErrorf("missing required parameters: app_id or allocations")
	}

	participants := make(map[string]bool)
	var newVersion uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		appSession, participantWeights, err := verifyQuorum(tx, params.AppSessionID, rpcSigners)
		if err != nil {
			return err
		}
		sessionAccountID := NewAccountID(appSession.SessionID)

		appSessionBalance, err := getAppSessionBalances(tx, sessionAccountID)
		if err != nil {
			return err
		}

		allocationSum := map[string]decimal.Decimal{}
		for _, alloc := range params.Allocations {
			if alloc.Amount.IsNegative() {
				return RPCErrorf("negative allocation: %s for asset %s", alloc.Amount, alloc.AssetSymbol)
			}

			walletAddress := GetWalletBySigner(alloc.ParticipantWallet)
			if walletAddress == "" {
				walletAddress = alloc.ParticipantWallet
			}

			if _, ok := participantWeights[walletAddress]; !ok {
				return RPCErrorf("allocation to non-participant %s", walletAddress)
			}

			userAddress := common.HexToAddress(walletAddress)
			userAccountID := NewAccountID(walletAddress)
			ledger := GetWalletLedger(tx, userAddress)
			balance, err := ledger.Balance(sessionAccountID, alloc.AssetSymbol)
			if err != nil {
				return RPCErrorf("failed to get session balance for asset %s", alloc.AssetSymbol)
			}

			// Debit session, credit participant
			if err := ledger.Record(sessionAccountID, alloc.AssetSymbol, balance.Neg()); err != nil {
				return RPCErrorf("failed to debit session: %w", err)
			}
			if err := ledger.Record(userAccountID, alloc.AssetSymbol, alloc.Amount); err != nil {
				return RPCErrorf("failed to credit participant: %w", err)
			}
			_, err = RecordLedgerTransaction(tx, TransactionTypeAppWithdrawal, sessionAccountID, userAccountID, alloc.AssetSymbol, alloc.Amount)
			if err != nil {
				return RPCErrorf("failed to record transaction: %w", err)
			}

			if !alloc.Amount.IsZero() {
				allocationSum[alloc.AssetSymbol] = allocationSum[alloc.AssetSymbol].Add(alloc.Amount)
				participants[walletAddress] = true
			}
		}

		if err := verifyAllocations(appSessionBalance, allocationSum); err != nil {
			return err
		}

		newVersion = appSession.Version + 1
		updates := map[string]any{
			"status":  ChannelStatusClosed,
			"version": newVersion,
		}
		if params.SessionData != nil {
			updates["session_data"] = *params.SessionData
		}

		return tx.Model(&appSession).Updates(updates).Error
	})

	if err != nil {
		return AppSessionResponse{}, err
	}

	for participant := range participants {
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

// verifyQuorum loads an open AppSession, verifies signatures meet quorum
func verifyQuorum(tx *gorm.DB, appSessionID string, rpcSigners map[string]struct{}) (AppSession, map[string]int64, error) {
	var session AppSession
	if err := tx.Where("session_id = ? AND status = ?", appSessionID, ChannelStatusOpen).
		Order("nonce DESC").First(&session).Error; err != nil {
		return AppSession{}, nil, RPCErrorf("virtual app %s is not opened", appSessionID)
	}

	participantWeights := make(map[string]int64, len(session.ParticipantWallets))
	for i, addr := range session.ParticipantWallets {
		participantWeights[addr] = session.Weights[i]
	}

	var totalWeight int64
	for wallet := range rpcSigners {
		weight, ok := participantWeights[wallet]
		if !ok {
			return AppSession{}, nil, RPCErrorf("signature from unknown participant wallet %s", wallet)
		}
		if weight <= 0 {
			return AppSession{}, nil, RPCErrorf("zero weight for signer %s", wallet)
		}
		totalWeight += weight
	}

	if totalWeight < int64(session.Quorum) {
		return AppSession{}, nil, RPCErrorf("quorum not met: %d / %d", totalWeight, session.Quorum)
	}

	return session, participantWeights, nil
}
