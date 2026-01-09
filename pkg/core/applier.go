package core

import "fmt"

type SimpleTransitionApplier struct {
}

func NewTransitionApplier() *SimpleTransitionApplier {
	return &SimpleTransitionApplier{}
}

func (a *SimpleTransitionApplier) Apply(state State, transition Transition) (State, error) {
	switch transition.Type {
	case TransitionTypeTransferSend:
		transferAmount := transition.Amount
		return State{
			ID:              state.ID,
			Transitions:     append(state.Transitions, transition),
			Asset:           state.Asset,
			UserWallet:      state.UserWallet,
			Epoch:           state.Epoch,
			Version:         state.Version,
			HomeChannelID:   state.HomeChannelID,
			EscrowChannelID: state.EscrowChannelID,
			HomeLedger: Ledger{
				BlockchainID: state.HomeLedger.BlockchainID,
				TokenAddress: state.HomeLedger.TokenAddress,
				UserBalance:  state.HomeLedger.UserBalance.Sub(transferAmount),
				UserNetFlow:  state.HomeLedger.UserNetFlow,
				NodeBalance:  state.HomeLedger.NodeBalance,
				NodeNetFlow:  state.HomeLedger.NodeNetFlow.Sub(transferAmount),
			},
			EscrowLedger: state.EscrowLedger,
			IsFinal:      state.IsFinal,
		}, nil
	case TransitionTypeTransferReceive:
		transferAmount := transition.Amount
		return State{
			ID:              state.ID,
			Transitions:     append(state.Transitions, transition),
			Asset:           state.Asset,
			UserWallet:      state.UserWallet,
			Epoch:           state.Epoch,
			Version:         state.Version,
			HomeChannelID:   state.HomeChannelID,
			EscrowChannelID: state.EscrowChannelID,
			HomeLedger: Ledger{
				BlockchainID: state.HomeLedger.BlockchainID,
				TokenAddress: state.HomeLedger.TokenAddress,
				UserBalance:  state.HomeLedger.UserBalance.Add(transferAmount),
				UserNetFlow:  state.HomeLedger.UserNetFlow,
				NodeBalance:  state.HomeLedger.NodeBalance,
				NodeNetFlow:  state.HomeLedger.NodeNetFlow.Add(transferAmount),
			},
			EscrowLedger: state.EscrowLedger,
			IsFinal:      state.IsFinal,
		}, nil
	default:
		return State{}, fmt.Errorf("transition type is not supported: %d", transition.Type)
	}
}
