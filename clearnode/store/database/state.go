package database

import (
	"fmt"
	"time"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Allocation struct {
	Participant  string          `json:"destination"`
	TokenAddress string          `json:"token"`
	RawAmount    decimal.Decimal `json:"amount"`
}

// State represents an immutable state in the system
// ID is deterministic: Hash(UserWallet, Asset, CycleIndex, Version)
type State struct {
	// ID is a 64-character deterministic hash
	ID          string         `gorm:"column:id;primaryKey;size:64"`
	Transitions datatypes.JSON `gorm:"column:transitions;type:text;not null"`
	Asset       string         `gorm:"column:asset;not null"`
	UserWallet  string         `gorm:"column:user_wallet;not null"`
	Epoch       uint64         `gorm:"column:epoch;not null"`

	Version uint64 `gorm:"column:version;not null"`

	// Optional channel references
	HomeChannelID   *string `gorm:"column:home_channel_id"`
	EscrowChannelID *string `gorm:"column:escrow_channel_id"`

	// Home Channel balances and flows
	// Using decimal.Decimal for int256 values and int64 for flow values
	HomeUserBalance decimal.Decimal `gorm:"column:home_user_balance;type:varchar(78)"`
	HomeUserNetFlow decimal.Decimal `gorm:"column:home_user_net_flow;default:0"`
	HomeNodeBalance decimal.Decimal `gorm:"column:home_node_balance;type:varchar(78)"`
	HomeNodeNetFlow decimal.Decimal `gorm:"column:home_node_net_flow;default:0"`

	// Escrow Channel balances and flows
	EscrowUserBalance decimal.Decimal `gorm:"column:escrow_user_balance;type:varchar(78)"`
	EscrowUserNetFlow decimal.Decimal `gorm:"column:escrow_user_net_flow;default:0"`
	EscrowNodeBalance decimal.Decimal `gorm:"column:escrow_node_balance;type:varchar(78)"`
	EscrowNodeNetFlow decimal.Decimal `gorm:"column:escrow_node_net_flow;default:0"`

	UserSig *string `gorm:"column:user_sig;type:text"`
	NodeSig *string `gorm:"column:node_sig;type:text"`

	// Read-only fields populated from JOINs with channels table
	HomeBlockchainID   *uint64 `gorm:"->;column:home_blockchain_id"`
	HomeTokenAddress   *string `gorm:"->;column:home_token_address"`
	EscrowBlockchainID *uint64 `gorm:"->;column:escrow_blockchain_id"`
	EscrowTokenAddress *string `gorm:"->;column:escrow_token_address"`

	CreatedAt time.Time
}

// TableName specifies the table name for the State model
func (State) TableName() string {
	return "channel_states"
}

// GetLastUserState retrieves the most recent state for a user's asset.
func (s *DBStore) GetLastUserState(wallet, asset string, signed bool) (*core.State, error) {
	var dbState State
	query := s.db.Table("channel_states AS s").
		Select("s.*, hc.blockchain_id AS home_blockchain_id, hc.token AS home_token_address, ec.blockchain_id AS escrow_blockchain_id, ec.token AS escrow_token_address").
		Joins("LEFT JOIN channels AS hc ON s.home_channel_id = hc.channel_id").
		Joins("LEFT JOIN channels AS ec ON s.escrow_channel_id = ec.channel_id").
		Where("s.user_wallet = ? AND s.asset = ?", wallet, asset)

	if signed {
		query = query.Where("s.user_sig IS NOT NULL AND s.node_sig IS NOT NULL")
	}

	err := query.Order("s.epoch DESC, s.version DESC").First(&dbState).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last user state: %w", err)
	}

	return databaseStateToCore(&dbState)
}

// StoreUserState persists a new user state to the database.
func (s *DBStore) StoreUserState(state core.State) error {
	dbState, err := coreStateToDB(&state)
	if err != nil {
		return fmt.Errorf("failed to encode transitions while creating a db state: %w", err)
	}

	if err := s.db.Create(dbState).Error; err != nil {
		return fmt.Errorf("failed to store user state: %w", err)
	}

	return nil
}

// GetLastStateByChannelID retrieves the most recent state for a given channel.
func (s *DBStore) GetLastStateByChannelID(channelID string, signed bool) (*core.State, error) {
	var dbState State
	query := s.db.Table("channel_states AS s").
		Select("s.*, hc.blockchain_id AS home_blockchain_id, hc.token AS home_token_address, ec.blockchain_id AS escrow_blockchain_id, ec.token AS escrow_token_address").
		Joins("LEFT JOIN channels AS hc ON s.home_channel_id = hc.channel_id").
		Joins("LEFT JOIN channels AS ec ON s.escrow_channel_id = ec.channel_id").
		Where("s.home_channel_id = ? OR s.escrow_channel_id = ?", channelID, channelID)

	if signed {
		query = query.Where("s.user_sig IS NOT NULL AND s.node_sig IS NOT NULL")
	}

	err := query.Order("s.epoch DESC, s.version DESC").First(&dbState).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last state by channel ID: %w", err)
	}

	return databaseStateToCore(&dbState)
}

// GetStateByChannelIDAndVersion retrieves a specific state version for a channel.
func (s *DBStore) GetStateByChannelIDAndVersion(channelID string, version uint64) (*core.State, error) {
	var dbState State
	err := s.db.Table("channel_states AS s").
		Select("s.*, hc.blockchain_id AS home_blockchain_id, hc.token AS home_token_address, ec.blockchain_id AS escrow_blockchain_id, ec.token AS escrow_token_address").
		Joins("LEFT JOIN channels AS hc ON s.home_channel_id = hc.channel_id").
		Joins("LEFT JOIN channels AS ec ON s.escrow_channel_id = ec.channel_id").
		Where("(s.home_channel_id = ? OR s.escrow_channel_id = ?) AND s.version = ?", channelID, channelID, version).
		First(&dbState).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get state by channel ID and version: %w", err)
	}

	return databaseStateToCore(&dbState)
}
