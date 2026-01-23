package database

import (
	"fmt"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type DBStore struct {
	db *gorm.DB
}

func NewDBStore(db *gorm.DB) DatabaseStore {
	return &DBStore{db: db}
}

// GetUserBalances retrieves the balances for a user's wallet.
func (s *DBStore) GetUserBalances(wallet string) ([]core.BalanceEntry, error) {
	type balanceEntry struct {
		Asset           string          `gorm:"column:asset"`
		HomeUserBalance decimal.Decimal `gorm:"column:home_user_balance"`
	}
	var balanceEntries []balanceEntry

	// Get the latest state for each asset (highest epoch and version)
	// For each asset, find the state with max epoch, and for that epoch, max version
	err := s.db.Raw(`
		SELECT cs.asset, cs.home_user_balance
		FROM channel_states cs
		INNER JOIN (
			SELECT asset, MAX(epoch) as max_epoch
			FROM channel_states
			WHERE user_wallet = ?
			GROUP BY asset
		) max_e ON cs.asset = max_e.asset AND cs.epoch = max_e.max_epoch
		INNER JOIN (
			SELECT asset, epoch, MAX(version) as max_version
			FROM channel_states
			WHERE user_wallet = ?
			GROUP BY asset, epoch
		) max_v ON cs.asset = max_v.asset AND cs.epoch = max_v.epoch AND cs.version = max_v.max_version
		WHERE cs.user_wallet = ?
	`, wallet, wallet, wallet).Scan(&balanceEntries).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user states: %w", err)
	}

	result := make([]core.BalanceEntry, 0, len(balanceEntries))
	for _, entry := range balanceEntries {
		result = append(result, core.BalanceEntry{
			Asset:   entry.Asset,
			Balance: entry.HomeUserBalance,
		})
	}

	return result, nil
}

// EnsureNoOngoingStateTransitions validates that no conflicting blockchain operations are pending.
// This method prevents race conditions by ensuring blockchain state versions
// match the user's last signed state version before accepting new transitions.
//
// Validation logic by transition type:
//   - home_deposit: Verify last_state.version == home_channel.state_version
//   - mutual_lock: Verify last_state.version == home_channel.state_version == escrow_channel.state_version
//     AND next transition must be escrow_deposit
//   - escrow_lock: Verify last_state.version == escrow_channel.state_version
//     AND next transition must be escrow_withdraw or migrate
//   - escrow_withdraw: Verify last_state.version == escrow_channel.state_version
//   - migrate: Verify last_state.version == home_channel.state_version
func (s *DBStore) EnsureNoOngoingStateTransitions(wallet, asset string, prevTransitionType core.TransitionType) error {
	type versionCheck struct {
		StateVersion         uint64
		HomeChannelVersion   *uint64
		EscrowChannelVersion *uint64
	}

	var result versionCheck
	err := s.db.Raw(`
		SELECT
			s.version as state_version,
			hc.state_version as home_channel_version,
			ec.state_version as escrow_channel_version
		FROM channel_states s
		LEFT JOIN channels hc ON hc.channel_id = s.home_channel_id
		LEFT JOIN channels ec ON ec.channel_id = s.escrow_channel_id
		WHERE s.user_wallet = ?
			AND s.asset = ?
			AND s.user_sig IS NOT NULL
			AND s.node_sig IS NOT NULL
		ORDER BY s.epoch DESC, s.version DESC
		LIMIT 1
	`, wallet, asset).Scan(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return fmt.Errorf("failed to check state transitions: %w", err)
	}

	// No previous state found (result will have zero values)
	if result.StateVersion == 0 {
		return nil
	}

	// Validation logic by transition type
	switch prevTransitionType {
	case core.TransitionTypeHomeDeposit:
		// Verify last_state.version == home_channel.state_version
		if result.HomeChannelVersion != nil && result.StateVersion != *result.HomeChannelVersion {
			return fmt.Errorf("home deposit is still ongoing")
		}

	case core.TransitionTypeHomeWithdrawal:
		// Verify last_state.version == home_channel.state_version
		if result.HomeChannelVersion != nil && result.StateVersion != *result.HomeChannelVersion {
			return fmt.Errorf("home withdrawal is still ongoing")
		}

	case core.TransitionTypeMutualLock:
		// Verify last_state.version == home_channel.state_version == escrow_channel.state_version
		if result.HomeChannelVersion != nil && result.StateVersion != *result.HomeChannelVersion ||
			result.EscrowChannelVersion != nil && result.StateVersion != *result.EscrowChannelVersion {
			return fmt.Errorf("mutual lock is still ongoing")
		}

	case core.TransitionTypeEscrowLock:
		// Verify last_state.version == escrow_channel.state_version
		if result.EscrowChannelVersion != nil && result.StateVersion != *result.EscrowChannelVersion {
			return fmt.Errorf("escrow lock is still ongoing")
		}

	case core.TransitionTypeEscrowWithdraw:
		// Verify last_state.version == escrow_channel.state_version
		if result.EscrowChannelVersion != nil && result.StateVersion != *result.EscrowChannelVersion {
			return fmt.Errorf("escrow withdrawal is still ongoing")
		}

	case core.TransitionTypeMigrate:
		// Verify last_state.version == home_channel.state_version
		if result.HomeChannelVersion != nil && result.StateVersion != *result.HomeChannelVersion {
			return fmt.Errorf("home chain migration is still ongoing")
		}
	}

	return nil
}
