package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type StateIntent uint8

const (
	StateIntentOperate    StateIntent = 0 // Operate the state application
	StateIntentInitialize StateIntent = 1 // Initial funding state
	StateIntentResize     StateIntent = 2 // Resize state
	StateIntentFinalize   StateIntent = 3 // Final closing state
)

type UnsignedState struct {
	Intent      StateIntent  `json:"intent"`
	Version     uint64       `json:"version"`
	Data        string       `json:"state_data"`
	Allocations []Allocation `json:"allocations"`
}

// Value implements driver.Valuer interface for database storage
func (u UnsignedState) Value() (driver.Value, error) {
	return json.Marshal(u)
}

// Scan implements sql.Scanner interface for database retrieval
func (u *UnsignedState) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into UnsignedState", value)
	}

	return json.Unmarshal(bytes, u)
}

type Allocation struct {
	Participant  string          `json:"destination"`
	TokenAddress string          `json:"token"`
	RawAmount    decimal.Decimal `json:"amount"`
}

// State represents an immutable state in the system
// ID is deterministic: Hash(UserWallet, Asset, CycleIndex, Version)
type State struct {
	// ID is a 64-character deterministic hash
	ID string `gorm:"column:id;primaryKey;size:64"`

	Data       string `gorm:"column:data;type:text"`
	Asset      string `gorm:"column:asset;not null"`
	UserWallet string `gorm:"column:user_wallet;not null"`
	CycleIndex uint64 `gorm:"column:cycle_index;not null"`

	Version uint64 `gorm:"column:version;not null"`

	// Optional channel references
	HomeChannelID   *string `gorm:"column:home_channel_id"`
	EscrowChannelID *string `gorm:"column:escrow_channel_id"`

	// Home Channel balances and flows
	// Using decimal.Decimal for int256 values and int64 for flow values
	HomeUserBalance decimal.Decimal `gorm:"column:home_user_balance;type:varchar(78)"`
	HomeUserNetFlow int64           `gorm:"column:home_user_net_flow;default:0"`
	HomeNodeBalance decimal.Decimal `gorm:"column:home_node_balance;type:varchar(78)"`
	HomeNodeNetFlow int64           `gorm:"column:home_node_net_flow;default:0"`

	// Escrow Channel balances and flows
	EscrowUserBalance decimal.Decimal `gorm:"column:escrow_user_balance;type:varchar(78)"`
	EscrowUserNetFlow int64           `gorm:"column:escrow_user_net_flow;default:0"`
	EscrowNodeBalance decimal.Decimal `gorm:"column:escrow_node_balance;type:varchar(78)"`
	EscrowNodeNetFlow int64           `gorm:"column:escrow_node_net_flow;default:0"`

	// TODO: Remove in the future if redundant
	IsFinal bool `gorm:"column:is_final;default:false"`

	UserSig string `gorm:"column:user_sig;type:text"`
	NodeSig string `gorm:"column:node_sig;type:text"`

	CreatedAt time.Time
}

// TableName specifies the table name for the State model
func (State) TableName() string {
	return "states"
}

// CreateState creates a new state in the database
func CreateState(tx *gorm.DB, state *State) error {
	if err := tx.Create(state).Error; err != nil {
		return err
	}
	return nil
}

// GetStateByID retrieves a state by its ID
func GetStateByID(tx *gorm.DB, id string) (*State, error) {
	var state State
	if err := tx.Where("id = ?", id).First(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

// GetStateByUserWalletAndAsset retrieves the state for a given user wallet and asset
// Since there is one state per asset per user, this returns a single state
func GetStateByUserWalletAndAsset(tx *gorm.DB, userWallet, asset string) (*State, error) {
	var state State
	if err := tx.Where("user_wallet = ? AND asset = ?", userWallet, asset).First(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}
