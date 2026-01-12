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

// toCoreChannelDefinition converts RPC channel definition to core type.
func toCoreChannelDefinition(def rpc.ChannelDefinitionV1) (core.ChannelDefinition, error) {
	nonce, err := strconv.ParseUint(def.Nonce, 10, 64)
	if err != nil {
		return core.ChannelDefinition{}, fmt.Errorf("failed to parse nonce: %w", err)
	}

	challenge, err := strconv.ParseUint(def.Challenge, 10, 64)
	if err != nil {
		return core.ChannelDefinition{}, fmt.Errorf("failed to parse challenge: %w", err)
	}

	return core.ChannelDefinition{
		Nonce:     nonce,
		Challenge: challenge,
	}, nil
}

// validateInitialState validates that the state is a valid initial state for channel creation.
func validateInitialState(state core.State) error {
	// Must be version 1, epoch 1
	if state.Version != 1 {
		return fmt.Errorf("initial state must have version 1")
	}
	if state.Epoch != 1 {
		return fmt.Errorf("initial state must have epoch 1")
	}

	// Must have no transitions (clean initial state)
	if len(state.Transitions) > 0 {
		return fmt.Errorf("initial state must have no transitions")
	}

	// Must have user wallet and asset
	if state.UserWallet == "" {
		return fmt.Errorf("user wallet is required")
	}
	if state.Asset == "" {
		return fmt.Errorf("asset is required")
	}

	// Home ledger must be initialized with zero balances
	if !state.HomeLedger.UserBalance.IsZero() || !state.HomeLedger.NodeBalance.IsZero() {
		return fmt.Errorf("initial state must have zero balances")
	}
	if !state.HomeLedger.UserNetFlow.IsZero() || !state.HomeLedger.NodeNetFlow.IsZero() {
		return fmt.Errorf("initial state must have zero net flows")
	}

	// Must not be final
	if state.IsFinal {
		return fmt.Errorf("initial state cannot be final")
	}

	// HomeChannelID should be nil for initial state (will be set by node)
	if state.HomeChannelID != nil && *state.HomeChannelID != "" {
		return fmt.Errorf("initial state should not have home_channel_id set")
	}

	// Must have valid state ID
	expectedID := core.GetStateID(state.UserWallet, state.Asset, state.Epoch, state.Version)
	if state.ID != expectedID {
		return fmt.Errorf("state ID mismatch: expected %s, got %s", expectedID, state.ID)
	}

	return nil
}
