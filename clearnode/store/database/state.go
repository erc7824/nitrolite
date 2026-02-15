package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// State represents an immutable state in the system
// ID is deterministic: Hash(UserWallet, Asset, CycleIndex, Version)
type State struct {
	// ID is a 66-character deterministic hash
	ID         string `gorm:"column:id;primaryKey;size:66"`
	Asset      string `gorm:"column:asset;not null"`
	UserWallet string `gorm:"column:user_wallet;not null"`
	Epoch      uint64 `gorm:"column:epoch;not null"`
	Version    uint64 `gorm:"column:version;not null"`

	// Transition
	TransitionType      uint8           `gorm:"column:transition_type;not null"`
	TransitionTxID      string          `gorm:"column:transition_tx_id;size:66;not null"`
	TransitionAccountID string          `gorm:"column:transition_account_id;size:66;not null"`
	TransitionAmount    decimal.Decimal `gorm:"column:transition_amount;type:varchar(78);not null"`

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

// StateHead represents the current head (latest state) for a (user_wallet, asset) pair
// This table provides O(1) reads and proper row-level locking for state transitions
type StateHead struct {
	UserWallet string `gorm:"column:user_wallet;primaryKey"`
	Asset      string `gorm:"column:asset;primaryKey"`

	// All fields mirroring the State table - with defaults matching migration
	Epoch   uint64 `gorm:"column:epoch;not null;default:0"`
	Version uint64 `gorm:"column:version;not null;default:0"`

	TransitionType      uint8           `gorm:"column:transition_type;not null;default:0"`
	TransitionTxID      string          `gorm:"column:transition_tx_id;size:66"`
	TransitionAccountID string          `gorm:"column:transition_account_id;size:66"`
	TransitionAmount    decimal.Decimal `gorm:"column:transition_amount;type:varchar(78);not null;default:0"`

	HomeChannelID   *string `gorm:"column:home_channel_id"`
	EscrowChannelID *string `gorm:"column:escrow_channel_id"`

	HomeUserBalance decimal.Decimal `gorm:"column:home_user_balance;type:varchar(78);not null;default:0"`
	HomeUserNetFlow decimal.Decimal `gorm:"column:home_user_net_flow;not null;default:0"`
	HomeNodeBalance decimal.Decimal `gorm:"column:home_node_balance;type:varchar(78);not null;default:0"`
	HomeNodeNetFlow decimal.Decimal `gorm:"column:home_node_net_flow;not null;default:0"`

	EscrowUserBalance decimal.Decimal `gorm:"column:escrow_user_balance;type:varchar(78);not null;default:0"`
	EscrowUserNetFlow decimal.Decimal `gorm:"column:escrow_user_net_flow;not null;default:0"`
	EscrowNodeBalance decimal.Decimal `gorm:"column:escrow_node_balance;type:varchar(78);not null;default:0"`
	EscrowNodeNetFlow decimal.Decimal `gorm:"column:escrow_node_net_flow;not null;default:0"`

	UserSig *string `gorm:"column:user_sig;type:text"`
	NodeSig *string `gorm:"column:node_sig;type:text"`

	// References to history
	HistoryID         *string `gorm:"column:history_id"`           // References current state in channel_states
	LastSignedStateID *string `gorm:"column:last_signed_state_id"` // References most recent fully signed state

	// Read-only fields from JOINs
	HomeBlockchainID   *uint64 `gorm:"->;column:home_blockchain_id"`
	HomeTokenAddress   *string `gorm:"->;column:home_token_address"`
	EscrowBlockchainID *uint64 `gorm:"->;column:escrow_blockchain_id"`
	EscrowTokenAddress *string `gorm:"->;column:escrow_token_address"`

	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName specifies the table name for the StateHead model
func (StateHead) TableName() string {
	return "channel_state_heads"
}

// StateRow is a unified type for scanning state data from queries that may return
// either head rows or history rows (or a union of both). This eliminates duplication
// in conversion logic and enables single-query optimizations.
type StateRow struct {
	// Core state fields (present in both head and history)
	ID         string `gorm:"column:id"` // history_id for heads, id for history rows
	Asset      string `gorm:"column:asset"`
	UserWallet string `gorm:"column:user_wallet"`
	Epoch      uint64 `gorm:"column:epoch"`
	Version    uint64 `gorm:"column:version"`

	TransitionType      uint8           `gorm:"column:transition_type"`
	TransitionTxID      string          `gorm:"column:transition_tx_id"`
	TransitionAccountID string          `gorm:"column:transition_account_id"`
	TransitionAmount    decimal.Decimal `gorm:"column:transition_amount"`

	HomeChannelID   *string `gorm:"column:home_channel_id"`
	EscrowChannelID *string `gorm:"column:escrow_channel_id"`

	HomeUserBalance decimal.Decimal `gorm:"column:home_user_balance"`
	HomeUserNetFlow decimal.Decimal `gorm:"column:home_user_net_flow"`
	HomeNodeBalance decimal.Decimal `gorm:"column:home_node_balance"`
	HomeNodeNetFlow decimal.Decimal `gorm:"column:home_node_net_flow"`

	EscrowUserBalance decimal.Decimal `gorm:"column:escrow_user_balance"`
	EscrowUserNetFlow decimal.Decimal `gorm:"column:escrow_user_net_flow"`
	EscrowNodeBalance decimal.Decimal `gorm:"column:escrow_node_balance"`
	EscrowNodeNetFlow decimal.Decimal `gorm:"column:escrow_node_net_flow"`

	UserSig *string `gorm:"column:user_sig"`
	NodeSig *string `gorm:"column:node_sig"`

	// Joined fields from channels table
	HomeBlockchainID   *uint64 `gorm:"column:home_blockchain_id"`
	HomeTokenAddress   *string `gorm:"column:home_token_address"`
	EscrowBlockchainID *uint64 `gorm:"column:escrow_blockchain_id"`
	EscrowTokenAddress *string `gorm:"column:escrow_token_address"`
}

// GetLastUserState retrieves the most recent state for a user's asset.
// When called within a transaction (s.inTx == true), this method acquires a FOR UPDATE lock
// on the head row, ensuring no concurrent modifications can occur during the transaction.
// If the head row doesn't exist in a transaction, an initial row is created and locked.
func (s *DBStore) GetLastUserState(wallet, asset string, signed bool) (*core.State, error) {
	wallet = strings.ToLower(wallet)

	if s.inTx {
		// Transaction mode: read from head table with locking
		return s.getLastUserStateWithLock(wallet, asset, signed)
	}

	// Non-transaction mode: read from head table without locking (read-only operations)
	return s.getLastUserStateNoLock(wallet, asset, signed)
}

// getLastUserStateWithLock reads the head with FOR UPDATE lock (transaction mode)
// Uses INSERT ON CONFLICT followed by a single query with CTE and UNION ALL for optimal performance
func (s *DBStore) getLastUserStateWithLock(wallet, asset string, signed bool) (*core.State, error) {
	// First, ensure head row exists (INSERT ON CONFLICT DO NOTHING)
	err := s.db.Exec(`
		INSERT INTO channel_state_heads (user_wallet, asset)
		VALUES (?, ?)
		ON CONFLICT (user_wallet, asset) DO NOTHING
	`, wallet, asset).Error
	if err != nil {
		return nil, fmt.Errorf("failed to ensure head exists: %w", err)
	}

	var row StateRow

	if !signed {
		// Simple case: just return head with FOR UPDATE lock
		query := fmt.Sprintf(`
			SELECT
				%s,
				%s
			FROM channel_state_heads h
			%s
			WHERE h.user_wallet = ? AND h.asset = ?
			FOR UPDATE
			LIMIT 1
		`, stateRowSelectColumns("h", true), channelSelectColumns(), channelJoinsFragment("h"))

		err = s.db.Raw(query, wallet, asset).Scan(&row).Error
	} else {
		// Signed requested: return signed head OR last_signed_state_id history
		query := fmt.Sprintf(`
			WITH h AS (
				SELECT * FROM channel_state_heads
				WHERE user_wallet = $1 AND asset = $2
				FOR UPDATE
			),
			picked AS (
				-- Return head if signed
				SELECT %s
				FROM h
				WHERE h.user_sig IS NOT NULL AND h.node_sig IS NOT NULL

				UNION ALL

				-- Fallback to last signed history
				SELECT %s
				FROM h
				JOIN channel_states s ON s.id = h.last_signed_state_id
				WHERE (h.user_sig IS NULL OR h.node_sig IS NULL) AND h.last_signed_state_id IS NOT NULL
			)
			SELECT
				picked.*,
				%s
			FROM picked
			%s
			LIMIT 1
		`, stateRowSelectColumns("h", true), stateRowSelectColumns("s", false), channelSelectColumns(), channelJoinsFragment("picked"))

		err = s.db.Raw(query, wallet, asset).Scan(&row).Error
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound || row.UserWallet == "" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	// Check if we got a result
	if row.UserWallet == "" {
		return nil, nil
	}

	return stateRowToCore(&row)
}

// getLastUserStateNoLock reads the head without locking (non-transaction mode)
// Uses a single SQL query with UNION ALL for optimal performance
func (s *DBStore) getLastUserStateNoLock(wallet, asset string, signed bool) (*core.State, error) {
	var row StateRow

	var err error
	if !signed {
		// Simple case: just return head (no UNION needed)
		query := fmt.Sprintf(`
			SELECT
				%s,
				%s
			FROM channel_state_heads h
			%s
			WHERE h.user_wallet = ? AND h.asset = ?
			LIMIT 1
		`, stateRowSelectColumns("h", true), channelSelectColumns(), channelJoinsFragment("h"))

		err = s.db.Raw(query, wallet, asset).Scan(&row).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get head: %w", err)
		}
	} else {
		// Signed requested: return signed head OR last_signed_state_id history
		query := fmt.Sprintf(`
			WITH h AS (
				SELECT * FROM channel_state_heads
				WHERE user_wallet = $1 AND asset = $2
			),
			picked AS (
				-- Return head if signed
				SELECT %s
				FROM h
				WHERE h.user_sig IS NOT NULL AND h.node_sig IS NOT NULL

				UNION ALL

				-- Fallback to last signed history
				SELECT %s
				FROM h
				JOIN channel_states s ON s.id = h.last_signed_state_id
				WHERE (h.user_sig IS NULL OR h.node_sig IS NULL) AND h.last_signed_state_id IS NOT NULL
			)
			SELECT
				picked.*,
				%s
			FROM picked
			%s
			LIMIT 1
		`, stateRowSelectColumns("h", true), stateRowSelectColumns("s", false), channelSelectColumns(), channelJoinsFragment("picked"))

		err = s.db.Raw(query, wallet, asset).Scan(&row).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get state: %w", err)
		}
	}

	// Check if we got a result
	if row.UserWallet == "" {
		return nil, nil
	}

	return stateRowToCore(&row)
}

// StoreUserState persists a new user state to the database.
// IMPORTANT: This method MUST be called within a transaction (ExecuteInTransaction).
// It assumes the head row was already locked via GetLastUserState.
// It inserts the new state into channel_states history and updates the head row
// with optimistic locking (version check) to prevent concurrent modifications.
func (s *DBStore) StoreUserState(state core.State) error {
	// Safety guard: ensure we're in a transaction
	if !s.inTx {
		return fmt.Errorf("StoreUserState must be called within a transaction (use ExecuteInTransaction)")
	}

	wallet := strings.ToLower(state.UserWallet)

	// Step 1: Insert into history (channel_states)
	dbState, err := coreStateToDB(&state)
	if err != nil {
		return fmt.Errorf("failed to convert state to DB model: %w", err)
	}

	if err := s.db.Create(dbState).Error; err != nil {
		return fmt.Errorf("failed to insert state into history: %w", err)
	}

	// Step 2: Determine if new state is fully signed
	newStateIsSigned := state.UserSig != nil && state.NodeSig != nil

	// Step 3: Build the head update
	// We need to update all fields to mirror the new state
	updates := map[string]interface{}{
		"epoch":                 state.Epoch,
		"version":               state.Version,
		"transition_type":       uint8(state.Transition.Type),
		"transition_tx_id":      strings.ToLower(state.Transition.TxID),
		"transition_account_id": strings.ToLower(state.Transition.AccountID),
		"transition_amount":     state.Transition.Amount,
		"home_user_balance":     state.HomeLedger.UserBalance,
		"home_user_net_flow":    state.HomeLedger.UserNetFlow,
		"home_node_balance":     state.HomeLedger.NodeBalance,
		"home_node_net_flow":    state.HomeLedger.NodeNetFlow,
		"history_id":            strings.ToLower(state.ID),
		"updated_at":            time.Now(),
	}

	// Handle optional channel IDs
	if state.HomeChannelID != nil {
		updates["home_channel_id"] = strings.ToLower(*state.HomeChannelID)
	} else {
		updates["home_channel_id"] = nil
	}

	if state.EscrowChannelID != nil {
		updates["escrow_channel_id"] = strings.ToLower(*state.EscrowChannelID)
	} else {
		updates["escrow_channel_id"] = nil
	}

	// Handle escrow ledger
	if state.EscrowLedger != nil {
		updates["escrow_user_balance"] = state.EscrowLedger.UserBalance
		updates["escrow_user_net_flow"] = state.EscrowLedger.UserNetFlow
		updates["escrow_node_balance"] = state.EscrowLedger.NodeBalance
		updates["escrow_node_net_flow"] = state.EscrowLedger.NodeNetFlow
	} else {
		updates["escrow_user_balance"] = decimal.Zero
		updates["escrow_user_net_flow"] = decimal.Zero
		updates["escrow_node_balance"] = decimal.Zero
		updates["escrow_node_net_flow"] = decimal.Zero
	}

	// Handle signatures
	if state.UserSig != nil {
		updates["user_sig"] = *state.UserSig
	} else {
		updates["user_sig"] = nil
	}

	if state.NodeSig != nil {
		updates["node_sig"] = *state.NodeSig
	} else {
		updates["node_sig"] = nil
	}

	// Update last_signed_state_id if new state is fully signed
	if newStateIsSigned {
		updates["last_signed_state_id"] = strings.ToLower(state.ID)
	}
	// Otherwise, keep last_signed_state_id unchanged (don't add it to updates)

	// Step 4: Update head with optimistic locking (version check)
	// This ensures the head wasn't modified by another transaction
	result := s.db.Model(&StateHead{}).
		Where("user_wallet = ? AND asset = ? AND version = ?", wallet, state.Asset, state.Version-1).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update head: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("concurrent modification detected: head version mismatch (expected version %d)", state.Version-1)
	}

	return nil
}

// GetLastStateByChannelID retrieves the most recent state for a given channel.
// Uses head table with UNION ALL to avoid slow OR condition, and CTE for single-query optimization.
func (s *DBStore) GetLastStateByChannelID(channelID string, signed bool) (*core.State, error) {
	channelID = strings.ToLower(channelID)

	var row StateRow

	var err error
	if !signed {
		// Simple case: return any head matching the channel ID
		query := fmt.Sprintf(`
			WITH matched_heads AS (
				SELECT * FROM channel_state_heads WHERE home_channel_id = $1
				UNION ALL
				SELECT * FROM channel_state_heads WHERE escrow_channel_id = $1
			)
			SELECT
				%s,
				%s
			FROM matched_heads h
			%s
			ORDER BY h.epoch DESC, h.version DESC
			LIMIT 1
		`, stateRowSelectColumns("h", true), channelSelectColumns(), channelJoinsFragment("h"))

		err = s.db.Raw(query, channelID).Scan(&row).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get last state by channel ID: %w", err)
		}
	} else {
		// Signed requested: return signed head OR last_signed_state_id history
		query := fmt.Sprintf(`
			WITH matched_heads AS (
				SELECT * FROM channel_state_heads WHERE home_channel_id = $1
				UNION ALL
				SELECT * FROM channel_state_heads WHERE escrow_channel_id = $1
			),
			latest_head AS (
				SELECT * FROM matched_heads
				ORDER BY epoch DESC, version DESC
				LIMIT 1
			),
			picked AS (
				-- Return head if signed
				SELECT %s
				FROM latest_head h
				WHERE h.user_sig IS NOT NULL AND h.node_sig IS NOT NULL

				UNION ALL

				-- Fallback to last signed history
				SELECT %s
				FROM latest_head h
				JOIN channel_states s ON s.id = h.last_signed_state_id
				WHERE h.user_sig IS NULL OR h.node_sig IS NULL
			)
			SELECT
				picked.*,
				%s
			FROM picked
			%s
			LIMIT 1
		`, stateRowSelectColumns("h", true), stateRowSelectColumns("s", false), channelSelectColumns(), channelJoinsFragment("picked"))

		err = s.db.Raw(query, channelID).Scan(&row).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get last state by channel ID: %w", err)
		}
	}

	// Check if we got a result
	if row.UserWallet == "" {
		return nil, nil
	}

	return stateRowToCore(&row)
}

// GetStateByChannelIDAndVersion retrieves a specific state version for a channel.
// Uses UNION ALL to avoid slow OR condition and leverage indexes.
func (s *DBStore) GetStateByChannelIDAndVersion(channelID string, version uint64) (*core.State, error) {
	channelID = strings.ToLower(channelID)

	var row StateRow

	// Use UNION ALL to query home and escrow separately (indexed lookups)
	query := fmt.Sprintf(`
		SELECT
			%s,
			%s
		FROM (
			SELECT * FROM channel_states WHERE home_channel_id = ? AND version = ?
			UNION ALL
			SELECT * FROM channel_states WHERE escrow_channel_id = ? AND version = ?
		) s
		%s
		LIMIT 1
	`, stateRowSelectColumns("s", false), channelSelectColumns(), channelJoinsFragment("s"))

	err := s.db.Raw(query, channelID, version, channelID, version).Scan(&row).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get state by channel ID and version: %w", err)
	}

	// Check if we found a result
	if row.ID == "" {
		return nil, nil
	}

	return stateRowToCore(&row)
}
