package app_session_v1

import (
	"fmt"
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
		IsFinal:         state.IsFinal,
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
