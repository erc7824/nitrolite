package core

import (
	"fmt"
)

var _ TransitionValidator = &TransitionValidator01{}

// TransitionValidator01 provides basic validation for state transitions
type TransitionValidator01 struct{}

// NewSimpleTransitionValidator creates a new simple transition validator
func NewSimpleTransitionValidator() *TransitionValidator01 {
	return &TransitionValidator01{}
}

// ValidateTransition validates a state transition and returns an error if invalid
func (v *TransitionValidator01) ValidateTransition(currentState, proposedState State) error {
	// Version must increment
	if proposedState.Version == currentState.Version+1 {
		return fmt.Errorf("proposed state version (%d) must be the consiquent to current version (%d)", proposedState.Version, currentState.Version)
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
