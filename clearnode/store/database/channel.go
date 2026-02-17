package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/pkg/core"
	"gorm.io/gorm"
)

// Channel represents a state channel between participants
type Channel struct {
	ChannelID             string             `gorm:"column:channel_id;primaryKey;"`
	UserWallet            string             `gorm:"column:user_wallet;not null"`
	Type                  core.ChannelType   `gorm:"column:type;not null"`
	BlockchainID          uint64             `gorm:"column:blockchain_id;not null"`
	Token                 string             `gorm:"column:token;not null"`
	ChallengeDuration     uint32             `gorm:"column:challenge_duration;not null"`
	ChallengeExpiresAt    *time.Time         `gorm:"column:challenge_expires_at;default:null"`
	Nonce                 uint64             `gorm:"column:nonce;not null;"`
	ApprovedSigValidators string             `gorm:"column:approved_sig_validators;not null;"`
	Status                core.ChannelStatus `gorm:"column:status;not null;"`
	StateVersion          uint64             `gorm:"column:state_version;not null;"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// TableName specifies the table name for the Channel model
func (Channel) TableName() string {
	return "channels"
}

// CreateChannel creates a new channel entity in the database.
func (s *DBStore) CreateChannel(channel core.Channel) error {
	dbChannel := Channel{
		ChannelID:             strings.ToLower(channel.ChannelID),
		UserWallet:            strings.ToLower(channel.UserWallet),
		Type:                  channel.Type,
		BlockchainID:          channel.BlockchainID,
		Token:                 strings.ToLower(channel.TokenAddress),
		ChallengeDuration:     channel.ChallengeDuration,
		ChallengeExpiresAt:    channel.ChallengeExpiresAt,
		Nonce:                 channel.Nonce,
		ApprovedSigValidators: channel.ApprovedSigValidators,
		Status:                channel.Status,
		StateVersion:          channel.StateVersion,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := s.db.Create(&dbChannel).Error; err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	return nil
}

// GetChannelByID retrieves a channel by its unique identifier.
func (s *DBStore) GetChannelByID(channelID string) (*core.Channel, error) {
	channelID = strings.ToLower(channelID)

	// Ensure channelID has 0x prefix
	if !strings.HasPrefix(channelID, "0x") {
		channelID = "0x" + channelID
	}

	var dbChannel Channel
	if err := s.db.Where("channel_id = ?", channelID).First(&dbChannel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get channel by ID: %w", err)
	}

	return databaseChannelToCore(&dbChannel), nil
}

// GetActiveHomeChannel retrieves the active home channel for a user's wallet and asset.
func (s *DBStore) GetActiveHomeChannel(wallet, asset string) (*core.Channel, error) {
	var dbChannel Channel
	err := s.db.
		Joins("JOIN channel_states ON channel_states.home_channel_id = channels.channel_id").
		Where("channel_states.user_wallet = ? AND channel_states.asset = ?", strings.ToLower(wallet), asset).
		Where("channels.status <= ? AND channels.type = ?", core.ChannelStatusOpen, core.ChannelTypeHome).
		Order("channel_states.epoch DESC, channel_states.version DESC").
		First(&dbChannel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active home channel: %w", err)
	}

	return databaseChannelToCore(&dbChannel), nil
}

// CheckOpenChannel verifies if a user has an active channel for the given asset
// and returns the approved signature validators if such a channel exists.
func (s *DBStore) CheckOpenChannel(wallet, asset string) (string, bool, error) {
	var approvedSigValidators string
	result := s.db.Raw(`
		SELECT c.approved_sig_validators
		FROM channel_states s
		INNER JOIN channels c ON c.channel_id = s.home_channel_id
		WHERE s.user_wallet = ?
			AND s.asset = ?
			AND c.status <= ?
			AND c.type = ?
		LIMIT 1
	`, strings.ToLower(wallet), asset, core.ChannelStatusOpen, core.ChannelTypeHome).Scan(&approvedSigValidators)
	if result.Error != nil {
		return "", false, fmt.Errorf("failed to check open channel: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return "", false, nil
	}

	return approvedSigValidators, true, nil
}

// ChannelCount holds the result of a COUNT() GROUP BY query on channels.
type ChannelCount struct {
	Asset  string             `gorm:"column:asset"`
	Status core.ChannelStatus `gorm:"column:status"`
	Count  uint64             `gorm:"column:count"`
}

// CountChannelsByStatus returns channel counts grouped by (asset, status).
// Joins with channel_states to resolve the asset name for each channel.
func (s *DBStore) CountChannelsByStatus() ([]ChannelCount, error) {
	var results []ChannelCount
	err := s.db.Raw(`
		SELECT cs.asset, c.status, COUNT(DISTINCT c.channel_id) as count
		FROM channels c
		INNER JOIN channel_states cs ON cs.home_channel_id = c.channel_id
		GROUP BY cs.asset, c.status
	`).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count channels: %w", err)
	}
	return results, nil
}

// UpdateChannel persists changes to a channel's metadata (status, version, etc).
func (s *DBStore) UpdateChannel(channel core.Channel) error {
	updates := map[string]interface{}{
		"status":               channel.Status,
		"state_version":        channel.StateVersion,
		"blockchain_id":        channel.BlockchainID,
		"token":                strings.ToLower(channel.TokenAddress),
		"nonce":                channel.Nonce,
		"challenge_expires_at": channel.ChallengeExpiresAt,
		"updated_at":           time.Now(),
	}

	if err := s.db.Model(&Channel{}).Where("channel_id = ?", strings.ToLower(channel.ChannelID)).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	return nil
}
