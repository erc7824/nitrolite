package core

import (
	"fmt"
)

var _ StateAdvancer = &StateAdvancerV1{}

// StateAdvancerV1 provides basic validation for state transitions
type StateAdvancerV1 struct{}

// NewStateAdvancerV1 creates a new simple transition validator
func NewStateAdvancerV1() *StateAdvancerV1 {
	return &StateAdvancerV1{}
}

// ValidateAdvancement validates that the proposed state is a valid advancement of the current state
func (v *StateAdvancerV1) ValidateAdvancement(currentState, proposedState State) error {
	expectedState := currentState.NextState()
	if proposedState.HomeChannelID == nil {
		return fmt.Errorf("home channel ID cannot be nil")
	}

	if expectedState.HomeChannelID == nil {
		expectedState.HomeChannelID = proposedState.HomeChannelID
		expectedState.HomeLedger.BlockchainID = proposedState.HomeLedger.BlockchainID
		expectedState.HomeLedger.TokenAddress = proposedState.HomeLedger.TokenAddress
	}

	if *proposedState.HomeChannelID != *expectedState.HomeChannelID {
		return fmt.Errorf("home channel ID mismatch: expected=%s, proposed=%s", *expectedState.HomeChannelID, *proposedState.HomeChannelID)
	}

	// Version must increment
	if proposedState.Version != expectedState.Version {
		return fmt.Errorf("version mismatch: expected=%d, proposed=%d", expectedState.Version, proposedState.Version)
	}

	// User wallet must match
	if proposedState.UserWallet != expectedState.UserWallet {
		return fmt.Errorf("user wallet mismatch: expected=%s, proposed=%s", expectedState.UserWallet, proposedState.UserWallet)
	}

	// Asset must match
	if proposedState.Asset != expectedState.Asset {
		return fmt.Errorf("asset mismatch: expected=%s, proposed=%s", expectedState.Asset, proposedState.Asset)
	}

	// Epoch must match
	if proposedState.Epoch != expectedState.Epoch {
		return fmt.Errorf("epoch mismatch: expected=%d, proposed=%d", expectedState.Epoch, proposedState.Epoch)
	}

	// State ID must match
	if proposedState.ID != expectedState.ID {
		return fmt.Errorf("state ID mismatch: expected=%s, proposed=%s", expectedState.ID, proposedState.ID)
	}

	if proposedState.UserSig == nil {
		return fmt.Errorf("user signature is required")
	}

	transitionLenDiff := len(proposedState.Transitions) - len(expectedState.Transitions)
	if transitionLenDiff < 0 {
		return fmt.Errorf("proposed state is missing transitions")
	}
	for i := range expectedState.Transitions {
		expectedTransition := expectedState.Transitions[i]
		proposedTransition := proposedState.Transitions[i]

		if err := expectedTransition.Equal(proposedTransition); err != nil {
			return fmt.Errorf("unexpected transition at index %d: %w", i, err)
		}
	}

	if transitionLenDiff > 1 {
		return fmt.Errorf("proposed state contains more than one new transition")
	}

	if transitionLenDiff == 0 {
		if !proposedState.IsFinal {
			return fmt.Errorf("no new transitions in non-final state")
		}

		expectedState.Finalize()
	}

	if transitionLenDiff == 1 {
		if proposedState.IsFinal {
			return fmt.Errorf("cannot add new transitions to a final state")
		}

		newTransition := proposedState.Transitions[len(proposedState.Transitions)-1]
		lastTransition := currentState.GetLastTransition()

		var err error
		switch newTransition.Type {
		case TransitionTypeHomeDeposit:
			_, err = expectedState.ApplyHomeDepositTransition(newTransition.Amount)
		case TransitionTypeHomeWithdrawal:
			_, err = expectedState.ApplyHomeWithdrawalTransition(newTransition.Amount)
		case TransitionTypeTransferSend:
			_, err = expectedState.ApplyTransferSendTransition(newTransition.AccountID, newTransition.Amount)
		case TransitionTypeCommit:
			_, err = expectedState.ApplyCommitTransition(newTransition.AccountID, newTransition.Amount)
		case TransitionTypeMutualLock:
			if proposedState.EscrowLedger == nil {
				return fmt.Errorf("proposed state escrow ledger is nil")
			}
			_, err = expectedState.ApplyMutualLockTransition(
				proposedState.EscrowLedger.BlockchainID,
				proposedState.EscrowLedger.TokenAddress,
				newTransition.Amount)
		case TransitionTypeEscrowDeposit:
			if lastTransition != nil && lastTransition.Type == TransitionTypeMutualLock {
				if !lastTransition.Amount.Equal(newTransition.Amount) {
					return fmt.Errorf("escrow deposit amount must be the same as mutual lock amount")
				}
				_, err = expectedState.ApplyEscrowDepositTransition(newTransition.Amount)
			} else {
				return fmt.Errorf("escrow deposit transition must follow a mutual lock transition")
			}
		case TransitionTypeEscrowLock:
			if proposedState.EscrowLedger == nil {
				return fmt.Errorf("proposed state escrow ledger is nil")
			}
			_, err = expectedState.ApplyEscrowLockTransition(
				proposedState.EscrowLedger.BlockchainID,
				proposedState.EscrowLedger.TokenAddress,
				newTransition.Amount)
		case TransitionTypeEscrowWithdraw:
			if lastTransition != nil && lastTransition.Type == TransitionTypeEscrowLock {
				if !lastTransition.Amount.Equal(newTransition.Amount) {
					return fmt.Errorf("escrow withdraw amount must be the same as escrow lock amount")
				}
				_, err = expectedState.ApplyEscrowWithdrawTransition(newTransition.Amount)
			} else {
				return fmt.Errorf("escrow withdraw transition must follow an escrow lock transition")
			}
		case TransitionTypeMigrate:
			_, err = expectedState.ApplyMigrateTransition(newTransition.Amount)
		default:
			return fmt.Errorf("unsupported type for new transition: %d", newTransition.Type)
		}
		if err != nil {
			return fmt.Errorf("failed to apply new transition: %w", err)
		}

		expectedTransition := expectedState.Transitions[len(expectedState.Transitions)-1]
		if err := expectedTransition.Equal(newTransition); err != nil {
			return fmt.Errorf("new transition does not match expected: %w", err)
		}
	}

	if err := proposedState.HomeLedger.Equal(expectedState.HomeLedger); err != nil {
		return fmt.Errorf("home ledger mismatch: %w", err)
	}
	if err := proposedState.HomeLedger.Validate(); err != nil {
		return fmt.Errorf("invalid home ledger: %w", err)
	}
	if proposedState.IsFinal && !expectedState.IsFinal {
		return fmt.Errorf("expected state is not final but proposed state is final")
	}

	if (expectedState.EscrowChannelID == nil) != (proposedState.EscrowChannelID == nil) {
		return fmt.Errorf("escrow channel ID presence mismatch")
	}

	if expectedState.EscrowChannelID != nil && proposedState.EscrowChannelID != nil {
		if *expectedState.EscrowChannelID != *proposedState.EscrowChannelID {
			return fmt.Errorf("escrow channel ID mismatch: expected=%s, proposed=%s", *expectedState.EscrowChannelID, *proposedState.EscrowChannelID)
		}
	}

	if (expectedState.EscrowLedger == nil) != (proposedState.EscrowLedger == nil) {
		return fmt.Errorf("escrow ledger presence mismatch")
	}

	if expectedState.EscrowLedger != nil && proposedState.EscrowLedger != nil {
		if err := proposedState.EscrowLedger.Equal(*expectedState.EscrowLedger); err != nil {
			return fmt.Errorf("escrow ledger mismatch: %w", err)
		}
		if err := proposedState.EscrowLedger.Validate(); err != nil {
			return fmt.Errorf("invalid escrow ledger: %w", err)
		}

		if proposedState.EscrowLedger.BlockchainID == proposedState.HomeLedger.BlockchainID {
			return fmt.Errorf("escrow ledger blockchain ID cannot match home ledger blockchain ID")
		}
	}

	return nil
}
