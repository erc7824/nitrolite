package database

import (
	"fmt"
	"strings"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type DBStore struct {
	inTx bool
	db   *gorm.DB
}

func NewDBStore(db *gorm.DB) DatabaseStore {
	return &DBStore{db: db}
}

func (s *DBStore) ExecuteInTransaction(txFunc StoreTxHandler) error {
	if s.inTx {
		return txFunc(s)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		txStore := &DBStore{
			inTx: true,
			db:   tx,
		}
		return txFunc(txStore)
	})
}

// GetUserBalances retrieves the balances for a user's wallet.
// This method reads directly from channel_state_heads for O(1) performance.
func (s *DBStore) GetUserBalances(wallet string) ([]core.BalanceEntry, error) {
	wallet = strings.ToLower(wallet)

	type balanceEntry struct {
		Asset           string          `gorm:"column:asset"`
		HomeUserBalance decimal.Decimal `gorm:"column:home_user_balance"`
	}
	var balanceEntries []balanceEntry

	// Simple query on head table - one row per asset
	err := s.db.Table("channel_state_heads").
		Select("asset, home_user_balance").
		Where("user_wallet = ?", wallet).
		Scan(&balanceEntries).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user balances: %w", err)
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
func (s *DBStore) EnsureNoOngoingStateTransitions(wallet, asset string) error {
	wallet = strings.ToLower(wallet)

	type versionCheck struct {
		TransitionType       core.TransitionType
		StateVersion         uint64
		LastSignedStateID    *string
		HomeChannelVersion   *uint64
		EscrowChannelVersion *uint64
	}

	var result versionCheck

	// Read from head table with left join to channels
	// If head is unsigned, use last_signed_state_id to get the version from history
	err := s.db.Raw(`
		SELECT
			CASE
				WHEN h.user_sig IS NOT NULL AND h.node_sig IS NOT NULL
				THEN h.transition_type
				ELSE (SELECT transition_type FROM channel_states WHERE id = h.last_signed_state_id)
			END as transition_type,
			CASE
				WHEN h.user_sig IS NOT NULL AND h.node_sig IS NOT NULL
				THEN h.version
				ELSE (SELECT version FROM channel_states WHERE id = h.last_signed_state_id)
			END as state_version,
			h.last_signed_state_id,
			hc.state_version as home_channel_version,
			ec.state_version as escrow_channel_version
		FROM channel_state_heads h
		LEFT JOIN channels hc ON hc.channel_id = h.home_channel_id
		LEFT JOIN channels ec ON ec.channel_id = h.escrow_channel_id
		WHERE h.user_wallet = ? AND h.asset = ?
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
	switch result.TransitionType {
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
