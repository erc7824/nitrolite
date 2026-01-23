package app_session_v1

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/shopspring/decimal"
)

func unmapAppDefinitionV1(def rpc.AppDefinitionV1) app.AppDefinitionV1 {
	participants := make([]app.AppParticipantV1, len(def.Participants))
	for i, p := range def.Participants {
		participants[i] = app.AppParticipantV1{
			WalletAddress:   p.WalletAddress,
			SignatureWeight: p.SignatureWeight,
		}
	}

	return app.AppDefinitionV1{
		Application:  def.Application,
		Participants: participants,
		Quorum:       def.Quorum,
		Nonce:        def.Nonce,
	}
}

// unmapStateV1 converts an RPC StateV1 to a core.State.
func unmapStateV1(state rpc.StateV1) (core.State, error) {
	coreTransitions := make([]core.Transition, len(state.Transitions))
	for i, transition := range state.Transitions {
		decimalTxAmount, err := decimal.NewFromString(transition.Amount)
		if err != nil {
			return core.State{}, fmt.Errorf("failed to parse amount: %w", err)
		}

		coreTransition := core.Transition{
			Type:      transition.Type,
			TxID:      transition.TxID,
			AccountID: transition.AccountID,
			Amount:    decimalTxAmount,
		}
		coreTransitions[i] = coreTransition
	}

	epoch, err := strconv.ParseUint(state.Epoch, 10, 64)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse epoch: %w", err)
	}

	version, err := strconv.ParseUint(state.Version, 10, 64)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse version: %w", err)
	}

	homeLedger, err := unmapLedgerV1(&state.HomeLedger)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse home ledger: %w", err)
	}

	escrowLedger, err := unmapLedgerV1(state.EscrowLedger)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse escrow ledger: %w", err)
	}

	return core.State{
		ID:              state.ID,
		Transitions:     coreTransitions,
		Asset:           state.Asset,
		UserWallet:      state.UserWallet,
		Epoch:           epoch,
		Version:         version,
		HomeChannelID:   state.HomeChannelID,
		EscrowChannelID: state.EscrowChannelID,
		HomeLedger:      *homeLedger,
		EscrowLedger:    escrowLedger,
		UserSig:         state.UserSig,
		NodeSig:         state.NodeSig,
	}, nil
}

func unmapLedgerV1(ledger *rpc.LedgerV1) (*core.Ledger, error) {
	if ledger == nil {
		return nil, nil
	}

	userBalance, err := decimal.NewFromString(ledger.UserBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user balance: %w", err)
	}

	userNetFlow, err := decimal.NewFromString(ledger.UserNetFlow)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user net-flow: %w", err)
	}

	nodeBalance, err := decimal.NewFromString(ledger.NodeBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node balance: %w", err)
	}

	nodeNetFlow, err := decimal.NewFromString(ledger.NodeNetFlow)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node net-flow: %w", err)
	}

	return &core.Ledger{
		BlockchainID: ledger.BlockchainID,
		TokenAddress: ledger.TokenAddress,
		UserBalance:  userBalance,
		UserNetFlow:  userNetFlow,
		NodeBalance:  nodeBalance,
		NodeNetFlow:  nodeNetFlow,
	}, nil
}

func unmapAppStateUpdateV1(upd *rpc.AppStateUpdateV1) (app.AppStateUpdateV1, error) {
	allocations := make([]app.AppAllocationV1, len(upd.Allocations))
	for i, alloc := range upd.Allocations {
		decAmount, err := decimal.NewFromString(alloc.Amount)
		if err != nil {
			return app.AppStateUpdateV1{}, fmt.Errorf("failed to parse amount: %w", err)
		}

		allocations[i] = app.AppAllocationV1{
			Participant: alloc.Participant,
			Asset:       alloc.Asset,
			Amount:      decAmount,
		}
	}

	return app.AppStateUpdateV1{
		AppSessionID: upd.AppSessionID,
		Intent:       upd.Intent,
		Version:      upd.Version,
		Allocations:  allocations,
		SessionData:  upd.SessionData,
	}, nil
}

// getParticipantWeights creates a map of participant wallet addresses to their weights.
func getParticipantWeights(participants []app.AppParticipantV1) map[string]uint8 {
	weights := make(map[string]uint8, len(participants))
	for _, p := range participants {
		weights[p.WalletAddress] = p.SignatureWeight
	}
	return weights
}

func mapAppSessionInfoV1(session app.AppSessionV1, allocations map[string]map[string]decimal.Decimal) rpc.AppSessionInfoV1 {
	participants := make([]rpc.AppParticipantV1, len(session.Participants))
	for i, p := range session.Participants {
		participants[i] = rpc.AppParticipantV1{
			WalletAddress:   p.WalletAddress,
			SignatureWeight: p.SignatureWeight,
		}
	}

	var sessionData *string
	if session.SessionData != "" {
		sessionData = &session.SessionData
	}

	// Convert allocations map to RPC format
	rpcAllocations := []rpc.AppAllocationV1{}
	for participant, assetMap := range allocations {
		for asset, amount := range assetMap {
			rpcAllocations = append(rpcAllocations, rpc.AppAllocationV1{
				Participant: participant,
				Asset:       asset,
				Amount:      amount.String(),
			})
		}
	}
	slices.SortFunc(rpcAllocations, func(a, b rpc.AppAllocationV1) int {
		if a.Asset > b.Asset {
			return 1
		} else if a.Asset < b.Asset {
			return -1
		}

		if a.Participant > b.Participant {
			return 1
		} else if a.Participant < b.Participant {
			return -1
		}

		return 0
	})

	return rpc.AppSessionInfoV1{
		AppSessionID: session.SessionID,
		Status:       session.Status.String(),
		Participants: participants,
		SessionData:  sessionData,
		Quorum:       session.Quorum,
		Version:      session.Version,
		Nonce:        session.Nonce,
		Allocations:  rpcAllocations,
	}
}

func mapPaginationMetadataV1(meta core.PaginationMetadata) rpc.PaginationMetadataV1 {
	return rpc.PaginationMetadataV1{
		Page:       meta.Page,
		PerPage:    meta.PerPage,
		TotalCount: meta.TotalCount,
		PageCount:  meta.PageCount,
	}
}
