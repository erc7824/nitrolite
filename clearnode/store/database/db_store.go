package database

import (
	"fmt"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/ethereum/go-ethereum/common"
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
	err := s.db.Raw(`
		SELECT DISTINCT ON (asset) asset, home_user_balance
		FROM channel_states
		WHERE user_wallet = ?
		ORDER BY asset, epoch DESC, version DESC
	`, wallet).Scan(&balanceEntries).Error

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
func (s *DBStore) EnsureNoOngoingStateTransitions(wallet, asset string) error {
	// Get the user's state to find their channels
	var state State
	err := s.db.Where("user_wallet = ? AND asset = ?", wallet, asset).
		Order("epoch DESC, version DESC").
		First(&state).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return fmt.Errorf("failed to get user state: %w", err)
	}

	// Check for pending blockchain actions on the home channel
	if state.HomeChannelID != nil {
		channelHash := common.HexToHash(*state.HomeChannelID)
		var count int64
		err := s.db.Model(&BlockchainAction{}).
			Where("channel_id = ? AND status = ?", channelHash, StatusPending).
			Count(&count).Error
		if err != nil {
			return fmt.Errorf("failed to check pending actions: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("there are %d pending blockchain operations for this channel", count)
		}
	}

	// Check for pending blockchain actions on the escrow channel
	if state.EscrowChannelID != nil {
		channelHash := common.HexToHash(*state.EscrowChannelID)
		var count int64
		err := s.db.Model(&BlockchainAction{}).
			Where("channel_id = ? AND status = ?", channelHash, StatusPending).
			Count(&count).Error
		if err != nil {
			return fmt.Errorf("failed to check pending escrow actions: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("there are %d pending blockchain operations for the escrow channel", count)
		}
	}

	return nil
}
