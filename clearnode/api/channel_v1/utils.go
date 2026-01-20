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

// channelTypeToString converts core.ChannelType to its string representation
func channelTypeToString(t core.ChannelType) string {
	switch t {
	case core.ChannelTypeHome:
		return "home"
	case core.ChannelTypeEscrow:
		return "escrow"
	default:
		return "unknown"
	}
}

// channelStatusToString converts core.ChannelStatus to its string representation
func channelStatusToString(s core.ChannelStatus) string {
	switch s {
	case core.ChannelStatusVoid:
		return "void"
	case core.ChannelStatusOpen:
		return "open"
	case core.ChannelStatusChallenged:
		return "challenged"
	case core.ChannelStatusClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// coreChannelToRPC converts a core.Channel to rpc.ChannelV1
func coreChannelToRPC(channel core.Channel) rpc.ChannelV1 {
	return rpc.ChannelV1{
		ChannelID:    channel.ChannelID,
		UserWallet:   channel.UserWallet,
		NodeWallet:   channel.NodeWallet,
		Type:         channelTypeToString(channel.Type),
		BlockchainID: channel.BlockchainID,
		TokenAddress: channel.TokenAddress,
		Challenge:    strconv.FormatUint(channel.Challenge, 10),
		Nonce:        strconv.FormatUint(channel.Nonce, 10),
		Status:       channelStatusToString(channel.Status),
		StateVersion: strconv.FormatUint(channel.StateVersion, 10),
	}
}

// coreStateToRPC converts a core.State to rpc.StateV1
func coreStateToRPC(state core.State) rpc.StateV1 {
	transitions := make([]rpc.TransitionV1, len(state.Transitions))
	for i, t := range state.Transitions {
		transitions[i] = rpc.TransitionV1{
			Type:      t.Type,
			TxID:      t.TxID,
			AccountID: t.AccountID,
			Amount:    t.Amount.String(),
		}
	}

	homeLedger := rpc.LedgerV1{
		TokenAddress: state.HomeLedger.TokenAddress,
		BlockchainID: state.HomeLedger.BlockchainID,
		UserBalance:  state.HomeLedger.UserBalance.String(),
		UserNetFlow:  state.HomeLedger.UserNetFlow.String(),
		NodeBalance:  state.HomeLedger.NodeBalance.String(),
		NodeNetFlow:  state.HomeLedger.NodeNetFlow.String(),
	}

	var escrowLedger *rpc.LedgerV1
	if state.EscrowLedger != nil {
		escrowLedger = &rpc.LedgerV1{
			TokenAddress: state.EscrowLedger.TokenAddress,
			BlockchainID: state.EscrowLedger.BlockchainID,
			UserBalance:  state.EscrowLedger.UserBalance.String(),
			UserNetFlow:  state.EscrowLedger.UserNetFlow.String(),
			NodeBalance:  state.EscrowLedger.NodeBalance.String(),
			NodeNetFlow:  state.EscrowLedger.NodeNetFlow.String(),
		}
	}

	return rpc.StateV1{
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
		UserSig:         state.UserSig,
		NodeSig:         state.NodeSig,
	}
}
