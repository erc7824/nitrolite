package core

import (
	"fmt"

	"github.com/shopspring/decimal"
)

var _ StateAdvancer = &StateAdvancerV1{}

// StateAdvancerV1 provides basic validation for state transitions
type StateAdvancerV1 struct{}

// NewStateAdvancerV1 creates a new simple transition validator
func NewStateAdvancerV1() *StateAdvancerV1 {
	return &StateAdvancerV1{}
}

func (a *StateAdvancerV1) ApplyTransition(state State, transition Transition) (State, error) {
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
	case TransitionTypeHomeDeposit:
		depositAmount := transition.Amount
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
				UserBalance:  state.HomeLedger.UserBalance.Add(depositAmount),
				UserNetFlow:  state.HomeLedger.UserNetFlow.Add(depositAmount),
				NodeBalance:  state.HomeLedger.NodeBalance,
				NodeNetFlow:  state.HomeLedger.NodeNetFlow,
			},
			EscrowLedger: state.EscrowLedger,
			IsFinal:      state.IsFinal,
		}, nil
	case TransitionTypeHomeWithdrawal:
		withdrawalAmount := transition.Amount
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
				UserBalance:  state.HomeLedger.UserBalance.Sub(withdrawalAmount),
				UserNetFlow:  state.HomeLedger.UserNetFlow.Sub(withdrawalAmount),
				NodeBalance:  state.HomeLedger.NodeBalance,
				NodeNetFlow:  state.HomeLedger.NodeNetFlow,
			},
			EscrowLedger: state.EscrowLedger,
			IsFinal:      state.IsFinal,
		}, nil
	case TransitionTypeMutualLock:
		lockAmount := transition.Amount
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
				UserBalance:  state.HomeLedger.UserBalance,
				UserNetFlow:  state.HomeLedger.UserNetFlow,
				NodeBalance:  state.HomeLedger.NodeBalance.Add(lockAmount),
				NodeNetFlow:  state.HomeLedger.NodeNetFlow.Add(lockAmount),
			},
			EscrowLedger: &Ledger{
				UserBalance: decimal.Zero.Add(lockAmount),
				UserNetFlow: decimal.Zero.Add(lockAmount),
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			IsFinal: state.IsFinal,
		}, nil
	case TransitionTypeEscrowDeposit:
		depositAmount := transition.Amount
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
				UserBalance:  state.HomeLedger.UserBalance.Add(depositAmount),
				UserNetFlow:  state.HomeLedger.UserNetFlow,
				NodeBalance:  state.HomeLedger.NodeBalance,
				NodeNetFlow:  state.HomeLedger.NodeNetFlow.Add(depositAmount),
			},
			EscrowLedger: &Ledger{
				BlockchainID: state.EscrowLedger.BlockchainID,
				TokenAddress: state.EscrowLedger.TokenAddress,
				UserBalance:  state.EscrowLedger.UserBalance.Sub(depositAmount),
				UserNetFlow:  state.EscrowLedger.UserNetFlow,
				NodeBalance:  state.EscrowLedger.NodeBalance,
				NodeNetFlow:  state.EscrowLedger.NodeNetFlow.Sub(depositAmount),
			},
			IsFinal: state.IsFinal,
		}, nil
	case TransitionTypeEscrowLock:
		lockAmount := transition.Amount
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
				UserBalance:  state.HomeLedger.UserBalance,
				UserNetFlow:  state.HomeLedger.UserNetFlow,
				NodeBalance:  state.HomeLedger.NodeBalance,
				NodeNetFlow:  state.HomeLedger.NodeNetFlow,
			},
			EscrowLedger: &Ledger{
				UserBalance: decimal.Zero,
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero.Add(lockAmount),
				NodeNetFlow: decimal.Zero.Add(lockAmount),
			},
			IsFinal: state.IsFinal,
		}, nil
	case TransitionTypeEscrowWithdraw:
		withdrawAmount := transition.Amount
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
				UserBalance:  state.HomeLedger.UserBalance.Sub(withdrawAmount),
				UserNetFlow:  state.HomeLedger.UserNetFlow,
				NodeBalance:  state.HomeLedger.NodeBalance,
				NodeNetFlow:  state.HomeLedger.NodeNetFlow.Sub(withdrawAmount),
			},
			EscrowLedger: &Ledger{
				BlockchainID: state.EscrowLedger.BlockchainID,
				TokenAddress: state.EscrowLedger.TokenAddress,
				UserBalance:  state.EscrowLedger.UserBalance,
				UserNetFlow:  state.EscrowLedger.UserNetFlow.Sub(withdrawAmount),
				NodeBalance:  state.EscrowLedger.NodeBalance.Sub(withdrawAmount),
				NodeNetFlow:  state.EscrowLedger.NodeNetFlow,
			},
			IsFinal: state.IsFinal,
		}, nil
	default:
		return State{}, fmt.Errorf("transition type is not supported: %d", transition.Type)
	}
}

func (a *StateAdvancerV1) ReapplyTransitions(base, new State) (State, error) {
	for _, t := range base.Transitions {
		updatedState, err := a.ApplyTransition(new, t)
		if err != nil {
			return State{}, err
		}
		new = updatedState
	}

	return new, nil
}

// ValidateTransitions validates a state transition and returns an error if invalid
func (v *StateAdvancerV1) ValidateTransitions(currentState, proposedState State) error {
	// Version must increment
	if proposedState.Version != currentState.Version+1 {
		return fmt.Errorf("proposed state version (%d) must be consecutive to current version (%d)", proposedState.Version, currentState.Version)
	}

	// Proposed state must have at least one transition
	if len(proposedState.Transitions) == 0 {
		return fmt.Errorf("proposed state must contain at least one transition")
	}

	// User wallet must match
	if proposedState.UserWallet != currentState.UserWallet {
		return fmt.Errorf("user wallet mismatch: current=%s, proposed=%s", currentState.UserWallet, proposedState.UserWallet)
	}

	// Asset must match
	if proposedState.Asset != currentState.Asset {
		return fmt.Errorf("asset mismatch: current=%s, proposed=%s", currentState.Asset, proposedState.Asset)
	}

	if proposedState.UserSig == nil {
		return fmt.Errorf("user signature is required")
	}

	// TODO: add additional checks
	// Epoch must not decrease
	if proposedState.Epoch < currentState.Epoch {
		return fmt.Errorf("proposed epoch (%d) cannot be less than current epoch (%d)", proposedState.Epoch, currentState.Epoch)
	}

	return nil
}
