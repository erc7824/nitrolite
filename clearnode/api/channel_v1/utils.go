package channel_v1

import (
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

func toCoreState(state rpc.StateV1) (core.State, error) {
	coreTransitions := make([]core.Transition, len(state.Transitions))
	for i, transition := range state.Transitions {
		decimalTxAmount, err := decimal.NewFromString(transition.Amount)
		if err != nil {
			return core.State{}, fmt.Errorf("failed to parse amount: %w", err)
		}

		coreTransition := core.Transition{
			Type:      transition.Type,
			TxHash:    transition.TxHash,
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

	homeLedger, err := toCoreLedger(&state.HomeLedger)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse home ledger: %w", err)
	}

	escrowLedger, err := toCoreLedger(state.EscrowLedger)
	if err != nil {
		return core.State{}, fmt.Errorf("failed to parse home ledger: %w", err)
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

func toCoreLedger(ledger *rpc.LedgerV1) (*core.Ledger, error) {
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
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}

	nodeNetFlow, err := decimal.NewFromString(ledger.NodeNetFlow)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount: %w", err)
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
