package sdk

import (
	"fmt"
	"strconv"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/shopspring/decimal"
)

// ============================================================================
// NodeConfig and Blockchain Transformations
// ============================================================================

// transformNodeConfig converts an RPC NodeV1GetConfigResponse to SDK NodeConfig type.
func transformNodeConfig(resp rpc.NodeV1GetConfigResponse) *core.NodeConfig {
	blockchains := make([]core.Blockchain, 0, len(resp.Blockchains))
	for _, info := range resp.Blockchains {
		blockchains = append(blockchains, core.Blockchain{
			Name:            info.Name,
			ID:              info.BlockchainID,
			ContractAddress: info.ContractAddress,
			BlockStep:       0, // Not provided in RPC response
		})
	}

	return &core.NodeConfig{
		NodeAddress: resp.NodeAddress,
		NodeVersion: resp.NodeVersion,
		Blockchains: blockchains,
	}
}

// ============================================================================
// Asset and Token Transformations
// ============================================================================

// transformAssets converts RPC AssetV1 slice to core.Asset slice.
func transformAssets(assets []rpc.AssetV1) []core.Asset {
	result := make([]core.Asset, 0, len(assets))
	for _, asset := range assets {
		tokens := make([]core.Token, 0, len(asset.Tokens))
		for _, token := range asset.Tokens {
			tokens = append(tokens, core.Token{
				Name:         token.Name,
				Symbol:       token.Symbol,
				Address:      token.Address,
				BlockchainID: token.BlockchainID,
				Decimals:     token.Decimals,
			})
		}
		result = append(result, core.Asset{
			Name:     asset.Name,
			Symbol:   asset.Symbol,
			Decimals: asset.Decimals,
			Tokens:   tokens,
		})
	}
	return result
}

// ============================================================================
// Balance Transformations
// ============================================================================

// transformBalances converts RPC BalanceEntryV1 slice to core.BalanceEntry slice.
func transformBalances(balances []rpc.BalanceEntryV1) []core.BalanceEntry {
	result := make([]core.BalanceEntry, 0, len(balances))
	for _, balance := range balances {
		amount, _ := decimal.NewFromString(balance.Amount)
		result = append(result, core.BalanceEntry{
			Asset:   balance.Asset,
			Balance: amount,
		})
	}
	return result
}

// ============================================================================
// Channel Transformations
// ============================================================================

// transformChannels converts RPC ChannelV1 slice to core.Channel slice.
func transformChannels(channels []rpc.ChannelV1) []core.Channel {
	result := make([]core.Channel, 0, len(channels))
	for _, channel := range channels {
		// Parse channel type
		var channelType core.ChannelType
		switch channel.Type {
		case "home":
			channelType = core.ChannelTypeHome
		case "escrow":
			channelType = core.ChannelTypeEscrow
		}

		// Parse channel status
		var channelStatus core.ChannelStatus
		switch channel.Status {
		case "void":
			channelStatus = core.ChannelStatusVoid
		case "open":
			channelStatus = core.ChannelStatusOpen
		case "challenged":
			channelStatus = core.ChannelStatusChallenged
		case "closed":
			channelStatus = core.ChannelStatusClosed
		}

		result = append(result, core.Channel{
			ChannelID:         channel.ChannelID,
			UserWallet:        channel.UserWallet,
			Type:              channelType,
			BlockchainID:      channel.BlockchainID,
			TokenAddress:      channel.TokenAddress,
			ChallengeDuration: channel.ChallengeDuration,
			Nonce:             0, // Not in RPC
			Status:            channelStatus,
			StateVersion:      0, // Not in RPC
		})
	}
	return result
}

// ============================================================================
// Transaction Transformations
// ============================================================================

// transformTransactions converts RPC TransactionV1 slice to core.Transaction slice.
func transformTransactions(transactions []rpc.TransactionV1) []core.Transaction {
	result := make([]core.Transaction, 0, len(transactions))
	for _, tx := range transactions {
		amount, _ := decimal.NewFromString(tx.Amount)

		// Parse timestamp
		createdAt, _ := time.Parse(time.RFC3339, tx.CreatedAt)

		result = append(result, core.Transaction{
			ID:                 tx.ID,
			Asset:              tx.Asset,
			TxType:             tx.TxType,
			FromAccount:        tx.FromAccount,
			ToAccount:          tx.ToAccount,
			SenderNewStateID:   nil, // Not in RPC
			ReceiverNewStateID: nil, // Not in RPC
			Amount:             amount,
			CreatedAt:          createdAt,
		})
	}
	return result
}

// ============================================================================
// Pagination Transformations
// ============================================================================

// transformPaginationMetadata converts RPC PaginationMetadataV1 to core.PaginationMetadata.
func transformPaginationMetadata(meta rpc.PaginationMetadataV1) core.PaginationMetadata {
	return core.PaginationMetadata{
		Page:       meta.Page,
		PerPage:    meta.PerPage,
		TotalCount: meta.TotalCount,
		PageCount:  meta.PageCount,
	}
}

// transformPaginationParams converts core.PaginationParams to RPC PaginationParamsV1.
func transformPaginationParams(params *core.PaginationParams) *rpc.PaginationParamsV1 {
	if params == nil {
		return nil
	}
	return &rpc.PaginationParamsV1{
		Offset: params.Offset,
		Limit:  params.Limit,
		Sort:   params.Sort,
	}
}

// ============================================================================
// State Management Transformations
// ============================================================================

// transformState converts RPC StateV1 to core.State.
func transformState(state rpc.StateV1) (core.State, error) {
	// Parse numeric strings
	epoch, err := strconv.ParseUint(state.Epoch, 10, 64)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse epoch: %w", err)
	}

	version, err := strconv.ParseUint(state.Version, 10, 64)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse version: %w", err)
	}

	// Transform transitions
	transitions := make([]core.Transition, 0, len(state.Transitions))
	for _, t := range state.Transitions {
		amount, err := decimal.NewFromString(t.Amount)
		if err != nil {
			return core.State{}, fmt.Errorf("failed to parse transition amount: %w", err)
		}
		transitions = append(transitions, core.Transition{
			Type:      t.Type,
			TxID:      t.TxID,
			AccountID: t.AccountID,
			Amount:    amount,
		})
	}

	// Transform ledgers
	homeLedger, err := transformLedger(state.HomeLedger)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to transform home ledger: %w", err)
	}

	var escrowLedger *core.Ledger
	if state.EscrowLedger != nil {
		el, err := transformLedger(*state.EscrowLedger)
		if err != nil {
			return core.State{}, fmt.Errorf("failed to transform escrow ledger: %w", err)
		}
		escrowLedger = &el
	}

	result := core.State{
		ID:              state.ID,
		Transitions:     transitions,
		Asset:           state.Asset,
		UserWallet:      state.UserWallet,
		Epoch:           epoch,
		Version:         version,
		HomeChannelID:   state.HomeChannelID,
		EscrowChannelID: state.EscrowChannelID,
		HomeLedger:      homeLedger,
		EscrowLedger:    escrowLedger,
		UserSig:         state.UserSig,
		NodeSig:         state.NodeSig,
		// Note: IsFinal is computed from transitions, not stored
	}

	return result, nil
}

// transformStates converts RPC StateV1 slice to core.State slice.
func transformStates(states []rpc.StateV1) ([]core.State, error) {
	result := make([]core.State, 0, len(states))
	for _, state := range states {
		s, err := transformState(state)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

// transformStateToRPC converts core.State to RPC StateV1.
func transformStateToRPC(state core.State) rpc.StateV1 {
	// Transform transitions
	transitions := make([]rpc.TransitionV1, 0, len(state.Transitions))
	for _, t := range state.Transitions {
		transitions = append(transitions, rpc.TransitionV1{
			Type:      t.Type,
			TxID:      t.TxID,
			AccountID: t.AccountID,
			Amount:    t.Amount.String(),
		})
	}

	// Transform ledgers
	homeLedger := transformLedgerToRPC(state.HomeLedger)

	var escrowLedger *rpc.LedgerV1
	if state.EscrowLedger != nil {
		el := transformLedgerToRPC(*state.EscrowLedger)
		escrowLedger = &el
	}

	result := rpc.StateV1{
		ID:              state.ID,
		Transitions:     transitions,
		Asset:           state.Asset,
		UserWallet:      state.UserWallet,
		Epoch:           strconv.FormatUint(state.Epoch, 10),
		Version:         strconv.FormatUint(state.Version, 10),
		HomeChannelID:   state.HomeChannelID,
		EscrowChannelID: state.EscrowChannelID,
		HomeLedger:      homeLedger,
		EscrowLedger:    escrowLedger,
		IsFinal:         state.IsFinal(), // Computed method, not field
		UserSig:         state.UserSig,
		NodeSig:         state.NodeSig,
	}

	return result
}

// transformLedger converts RPC LedgerV1 to core.Ledger.
func transformLedger(ledger rpc.LedgerV1) (core.Ledger, error) {
	userBalance, err := decimal.NewFromString(ledger.UserBalance)
	if err != nil {
		return core.Ledger{}, fmt.Errorf("failed to parse user balance: %w", err)
	}

	userNetFlow, err := decimal.NewFromString(ledger.UserNetFlow)
	if err != nil {
		return core.Ledger{}, fmt.Errorf("failed to parse user net flow: %w", err)
	}

	nodeBalance, err := decimal.NewFromString(ledger.NodeBalance)
	if err != nil {
		return core.Ledger{}, fmt.Errorf("failed to parse node balance: %w", err)
	}

	nodeNetFlow, err := decimal.NewFromString(ledger.NodeNetFlow)
	if err != nil {
		return core.Ledger{}, fmt.Errorf("failed to parse node net flow: %w", err)
	}

	return core.Ledger{
		TokenAddress: ledger.TokenAddress,
		BlockchainID: ledger.BlockchainID,
		UserBalance:  userBalance,
		UserNetFlow:  userNetFlow,
		NodeBalance:  nodeBalance,
		NodeNetFlow:  nodeNetFlow,
	}, nil
}

// transformLedgerToRPC converts core.Ledger to RPC LedgerV1.
func transformLedgerToRPC(ledger core.Ledger) rpc.LedgerV1 {
	return rpc.LedgerV1{
		TokenAddress: ledger.TokenAddress,
		BlockchainID: ledger.BlockchainID,
		UserBalance:  ledger.UserBalance.String(),
		UserNetFlow:  ledger.UserNetFlow.String(),
		NodeBalance:  ledger.NodeBalance.String(),
		NodeNetFlow:  ledger.NodeNetFlow.String(),
	}
}

// transformChannelDefinitionToRPC converts core.ChannelDefinition to RPC ChannelDefinitionV1.
func transformChannelDefinitionToRPC(def core.ChannelDefinition) rpc.ChannelDefinitionV1 {
	return rpc.ChannelDefinitionV1{
		Nonce:     def.Nonce,
		Challenge: def.Challenge,
	}
}

// ============================================================================
// App Session Transformations
// ============================================================================

// transformAppSessions converts RPC AppSessionInfoV1 slice to app.AppSessionInfoV1 slice.
func transformAppSessions(sessions []rpc.AppSessionInfoV1) []app.AppSessionInfoV1 {
	result := make([]app.AppSessionInfoV1, 0, len(sessions))
	for _, s := range sessions {
		// Transform participants
		participants := make([]app.AppParticipantV1, 0, len(s.Participants))
		for _, p := range s.Participants {
			participants = append(participants, app.AppParticipantV1{
				WalletAddress:   p.WalletAddress,
				SignatureWeight: p.SignatureWeight,
			})
		}

		// Transform allocations
		allocations := make([]app.AppAllocationV1, 0, len(s.Allocations))
		for _, a := range s.Allocations {
			amount, _ := decimal.NewFromString(a.Amount)
			allocations = append(allocations, app.AppAllocationV1{
				Participant: a.Participant,
				Asset:       a.Asset,
				Amount:      amount,
			})
		}

		// Parse status - RPC uses string, app uses IsClosed bool
		isClosed := (s.Status == "closed")

		// Handle session data - RPC uses *string, app uses string
		sessionData := ""
		if s.SessionData != nil {
			sessionData = *s.SessionData
		}

		result = append(result, app.AppSessionInfoV1{
			AppSessionID: s.AppSessionID,
			IsClosed:     isClosed,
			Participants: participants,
			SessionData:  sessionData,
			Quorum:       s.Quorum,
			Version:      s.Version,
			Nonce:        s.Nonce,
			Allocations:  allocations,
		})
	}
	return result
}

// transformAppDefinition converts RPC AppDefinitionV1 to app.AppDefinitionV1.
func transformAppDefinition(def rpc.AppDefinitionV1) app.AppDefinitionV1 {
	participants := make([]app.AppParticipantV1, 0, len(def.Participants))
	for _, p := range def.Participants {
		participants = append(participants, app.AppParticipantV1{
			WalletAddress:   p.WalletAddress,
			SignatureWeight: p.SignatureWeight,
		})
	}

	return app.AppDefinitionV1{
		Application:  def.Application,
		Participants: participants,
		Quorum:       def.Quorum,
		Nonce:        def.Nonce,
	}
}

// transformAppDefinitionToRPC converts app.AppDefinitionV1 to RPC AppDefinitionV1.
func transformAppDefinitionToRPC(def app.AppDefinitionV1) rpc.AppDefinitionV1 {
	participants := make([]rpc.AppParticipantV1, 0, len(def.Participants))
	for _, p := range def.Participants {
		participants = append(participants, rpc.AppParticipantV1{
			WalletAddress:   p.WalletAddress,
			SignatureWeight: p.SignatureWeight,
		})
	}

	return rpc.AppDefinitionV1{
		Application:  def.Application,
		Participants: participants,
		Quorum:       def.Quorum,
		Nonce:        def.Nonce,
	}
}

// transformAppStateUpdateToRPC converts app.AppStateUpdateV1 to RPC AppStateUpdateV1.
func transformAppStateUpdateToRPC(update app.AppStateUpdateV1) rpc.AppStateUpdateV1 {
	allocations := make([]rpc.AppAllocationV1, 0, len(update.Allocations))
	for _, a := range update.Allocations {
		allocations = append(allocations, rpc.AppAllocationV1{
			Participant: a.Participant,
			Asset:       a.Asset,
			Amount:      a.Amount.String(),
		})
	}

	return rpc.AppStateUpdateV1{
		AppSessionID: update.AppSessionID,
		Intent:       update.Intent,
		Version:      update.Version,
		Allocations:  allocations,
		SessionData:  update.SessionData,
	}
}

// transformSignedAppStateUpdateToRPC converts app.SignedAppStateUpdateV1 to RPC SignedAppStateUpdateV1.
func transformSignedAppStateUpdateToRPC(signed app.SignedAppStateUpdateV1) rpc.SignedAppStateUpdateV1 {
	return rpc.SignedAppStateUpdateV1{
		AppStateUpdate: transformAppStateUpdateToRPC(signed.AppStateUpdate),
		QuorumSigs:     signed.QuorumSigs,
	}
}
