package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/pkg/core"
	"gorm.io/gorm"
)

// ChannelSessionKeyStateV1 represents a channel session key state in the database.
type ChannelSessionKeyStateV1 struct {
	ID           string                     `gorm:"column:id;primaryKey"`
	UserAddress  string                     `gorm:"column:user_address;not null;uniqueIndex:idx_channel_session_key_states_v1_user_key_ver,priority:1"`
	SessionKey   string                     `gorm:"column:session_key;not null;uniqueIndex:idx_channel_session_key_states_v1_user_key_ver,priority:2"`
	Version      uint64                     `gorm:"column:version;not null;uniqueIndex:idx_channel_session_key_states_v1_user_key_ver,priority:3"`
	Assets       []ChannelSessionKeyAssetV1 `gorm:"foreignKey:SessionKeyStateID;references:ID"`
	MetadataHash string                     `gorm:"column:metadata_hash;type:char(66);not null"`
	ExpiresAt    time.Time                  `gorm:"column:expires_at;not null"`
	UserSig      string                     `gorm:"column:user_sig;not null"`
	CreatedAt    time.Time
}

func (ChannelSessionKeyStateV1) TableName() string {
	return "channel_session_key_states_v1"
}

// ChannelSessionKeyHeadV1 represents the current head (latest version) for a (user_address, session_key) pair
// This table provides O(1) reads and proper row-level locking for session key updates
type ChannelSessionKeyHeadV1 struct {
	UserAddress  string    `gorm:"column:user_address;primaryKey"`
	SessionKey   string    `gorm:"column:session_key;primaryKey"`
	Version      uint64    `gorm:"column:version;not null"`
	MetadataHash string    `gorm:"column:metadata_hash;type:char(66);not null"`
	ExpiresAt    time.Time `gorm:"column:expires_at;not null"`
	UserSig      string    `gorm:"column:user_sig;not null"`

	// Reference to history
	HistoryID *string `gorm:"column:history_id"` // References current state in channel_session_key_states_v1

	UpdatedAt time.Time
}

func (ChannelSessionKeyHeadV1) TableName() string {
	return "channel_session_key_heads_v1"
}

// ChannelSessionKeyAssetV1 links a channel session key state to an asset.
type ChannelSessionKeyAssetV1 struct {
	SessionKeyStateID string `gorm:"column:session_key_state_id;not null;primaryKey;priority:1"`
	Asset             string `gorm:"column:asset;not null;primaryKey;priority:2;index"`
}

func (ChannelSessionKeyAssetV1) TableName() string {
	return "channel_session_key_assets_v1"
}

// StoreChannelSessionKeyState stores a new channel session key state version.
// IMPORTANT: This method MUST be called within a transaction (ExecuteInTransaction).
// It assumes the head row was already locked via a previous GetLast* call in the transaction.
// It inserts the new state into history and updates/creates the head row.
func (s *DBStore) StoreChannelSessionKeyState(state core.ChannelSessionKeyStateV1) error {
	// Safety guard: ensure we're in a transaction
	if !s.inTx {
		return fmt.Errorf("StoreChannelSessionKeyState must be called within a transaction (use ExecuteInTransaction)")
	}

	userAddress := strings.ToLower(state.UserAddress)
	sessionKey := strings.ToLower(state.SessionKey)

	id, err := core.GenerateSessionKeyStateIDV1(userAddress, sessionKey, state.Version)
	if err != nil {
		return fmt.Errorf("failed to generate session key state ID: %w", err)
	}

	metadataHash, err := core.GetChannelSessionKeyAuthMetadataHashV1(state.Version, state.Assets, state.ExpiresAt.Unix())
	if err != nil {
		return fmt.Errorf("failed to compute metadata hash: %w", err)
	}

	// Step 1: Insert into history (channel_session_key_states_v1)
	dbState := ChannelSessionKeyStateV1{
		ID:           id,
		UserAddress:  userAddress,
		SessionKey:   sessionKey,
		Version:      state.Version,
		MetadataHash: strings.ToLower(metadataHash.Hex()),
		ExpiresAt:    state.ExpiresAt.UTC(),
		UserSig:      state.UserSig,
	}

	if err := s.db.Create(&dbState).Error; err != nil {
		return fmt.Errorf("failed to store channel session key state: %w", err)
	}

	// Store related Assets
	if len(state.Assets) > 0 {
		assets := make([]ChannelSessionKeyAssetV1, len(state.Assets))
		for i, asset := range state.Assets {
			assets[i] = ChannelSessionKeyAssetV1{
				SessionKeyStateID: id,
				Asset:             strings.ToLower(asset),
			}
		}
		if err := s.db.Create(&assets).Error; err != nil {
			return fmt.Errorf("failed to store channel session key assets: %w", err)
		}
	}

	// Step 2: Upsert head (INSERT ... ON CONFLICT UPDATE)
	now := time.Now().UTC()
	err = s.db.Exec(`
		INSERT INTO channel_session_key_heads_v1
		(user_address, session_key, version, metadata_hash, expires_at, user_sig, history_id, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_address, session_key)
		DO UPDATE SET
			version = EXCLUDED.version,
			metadata_hash = EXCLUDED.metadata_hash,
			expires_at = EXCLUDED.expires_at,
			user_sig = EXCLUDED.user_sig,
			history_id = EXCLUDED.history_id,
			updated_at = EXCLUDED.updated_at
		WHERE channel_session_key_heads_v1.version < EXCLUDED.version
	`, userAddress, sessionKey, state.Version, strings.ToLower(metadataHash.Hex()), state.ExpiresAt.UTC(), state.UserSig, id, now).Error

	if err != nil {
		return fmt.Errorf("failed to upsert channel session key head: %w", err)
	}

	return nil
}

// GetLastChannelSessionKeyStates retrieves the latest channel session key states for a user with optional filtering.
// Returns only non-expired session keys. Optimized to use 2 queries total regardless of N.
func (s *DBStore) GetLastChannelSessionKeyStates(wallet string, sessionKey *string) ([]core.ChannelSessionKeyStateV1, error) {
	wallet = strings.ToLower(wallet)

	// Query 1: Fetch all heads
	query := s.db.Model(&ChannelSessionKeyHeadV1{}).
		Where("user_address = ? AND expires_at > ?", wallet, time.Now().UTC()).
		Order("updated_at DESC")

	if sessionKey != nil && *sessionKey != "" {
		query = query.Where("session_key = ?", strings.ToLower(*sessionKey))
	}

	var heads []ChannelSessionKeyHeadV1
	if err := query.Find(&heads).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []core.ChannelSessionKeyStateV1{}, nil
		}
		return nil, fmt.Errorf("failed to get channel session key heads: %w", err)
	}

	if len(heads) == 0 {
		return []core.ChannelSessionKeyStateV1{}, nil
	}

	// Collect all history IDs for batch fetching
	historyIDs := make([]string, 0, len(heads))
	for _, head := range heads {
		if head.HistoryID != nil {
			historyIDs = append(historyIDs, *head.HistoryID)
		}
	}

	// Query 2: Batch fetch all Assets
	var assetLinks []ChannelSessionKeyAssetV1
	if len(historyIDs) > 0 {
		err := s.db.Model(&ChannelSessionKeyAssetV1{}).
			Where("session_key_state_id IN ?", historyIDs).
			Find(&assetLinks).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to fetch assets: %w", err)
		}
	}

	// Build map for fast lookup
	assetsByHistoryID := make(map[string][]string)
	for _, link := range assetLinks {
		assetsByHistoryID[link.SessionKeyStateID] = append(assetsByHistoryID[link.SessionKeyStateID], link.Asset)
	}

	// Assemble results
	states := make([]core.ChannelSessionKeyStateV1, 0, len(heads))
	for _, head := range heads {
		state := core.ChannelSessionKeyStateV1{
			UserAddress: head.UserAddress,
			SessionKey:  head.SessionKey,
			Version:     head.Version,
			ExpiresAt:   head.ExpiresAt,
			UserSig:     head.UserSig,
		}

		// Add assets if history exists
		if head.HistoryID != nil {
			state.Assets = assetsByHistoryID[*head.HistoryID]
		}

		// Ensure non-nil slice
		if state.Assets == nil {
			state.Assets = []string{}
		}

		states = append(states, state)
	}

	return states, nil
}

// GetLastChannelSessionKeyVersion returns the latest version of a non-expired channel session key state.
// Returns 0 if no state exists. Uses head table for O(1) lookup.
func (s *DBStore) GetLastChannelSessionKeyVersion(wallet, sessionKey string) (uint64, error) {
	wallet = strings.ToLower(wallet)
	sessionKey = strings.ToLower(sessionKey)

	var result struct {
		Version uint64
	}
	err := s.db.Model(&ChannelSessionKeyHeadV1{}).
		Select("version").
		Where("user_address = ? AND session_key = ? AND expires_at > ?", wallet, sessionKey, time.Now().UTC()).
		Take(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to check channel session key state: %w", err)
	}

	return result.Version, nil
}

// ValidateChannelSessionKeyForAsset checks in a single query that:
// - a session key state exists for the (wallet, sessionKey) pair,
// - it is the latest version (using head table),
// - it is not expired,
// - the asset is in the allowed list,
// - the metadata hash matches.
// Uses head table for O(1) lookup.
func (s *DBStore) ValidateChannelSessionKeyForAsset(wallet, sessionKey, asset, metadataHash string) (bool, error) {
	wallet = strings.ToLower(wallet)
	sessionKey = strings.ToLower(sessionKey)
	asset = strings.ToLower(asset)
	metadataHash = strings.ToLower(metadataHash)

	now := time.Now().UTC()

	// Check head table and join with assets table
	var count int64
	err := s.db.Model(&ChannelSessionKeyHeadV1{}).
		Where("user_address = ? AND session_key = ? AND expires_at > ? AND metadata_hash = ?",
			wallet, sessionKey, now, metadataHash).
		Joins("JOIN channel_session_key_assets_v1 ON channel_session_key_assets_v1.session_key_state_id = channel_session_key_heads_v1.history_id AND channel_session_key_assets_v1.asset = ?", asset).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to validate session key for asset: %w", err)
	}

	return count > 0, nil
}

func dbChannelSessionKeyStateToCore(dbState *ChannelSessionKeyStateV1) core.ChannelSessionKeyStateV1 {
	assets := make([]string, len(dbState.Assets))
	for i, a := range dbState.Assets {
		assets[i] = a.Asset
	}

	return core.ChannelSessionKeyStateV1{
		UserAddress: dbState.UserAddress,
		SessionKey:  dbState.SessionKey,
		Version:     dbState.Version,
		Assets:      assets,
		ExpiresAt:   dbState.ExpiresAt,
		UserSig:     dbState.UserSig,
	}
}
