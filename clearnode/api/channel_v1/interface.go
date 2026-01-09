package channel_v1

import (
	"github.com/erc7824/nitrolite/pkg/core"
)

type Store interface {
	BeginTx() (tx Store, commit, revert func() error)
	GetLastUserState(wallet, asset string, signed bool) (core.State, error)
	StoreUserState(state core.State) error
	EnsureNoOngoingStateTransitions(wallet, asset string) error
	ScheduleInitiateEscrowWithdrawal(core.State) error
	RecordTransaction(core.Transaction) error
}

// EnsureNoOngoingStateTransitions pseudocode
// ------------------
// Get user's last signed state
// Check last transition
// Check home and escrow channel depending on action

// If switch transition.type:
// case "home_deposit" (last_state.version == home_channel.state_version)
// case "mutual_lock" (last_state.version == home_channel.state_version == escrow_channel.state_version) and new_state.LastTransition().Type == "escrow_deposit"
// case "escrow_lock" (last_state.version == escrow_channel.state_version) and (new_state.LastTransition().Type IN ("escrow_withdraw", "migrate"))
// case "escrow_withdraw" (last_state.version == escrow_channel.state_version)
// case "migrate" (last_state.version == home_channel.state_version)

// channel creation (check home_channel.state_version != 0)

// in between:
// case "mutual_lock" -> "escrow_deposit"
// case "escrow_lock" -> "escrow_withdraw"
// case "escrow_lock" -> "migrate"
// we take as base last signed state, and if there is an unsigned state after signed,
// then its transitions are applied into incoming state,
// what results in new signed state and new unsigned state

type SigValidator interface {
	Verify(wallet string, data, sig []byte) error
}

type SigValidatorType string

const EcdsaSigValidatorType = "ecdsa"
