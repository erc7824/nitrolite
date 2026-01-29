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
	ChannelID          string             `gorm:"column:channel_id;primaryKey;"`
	UserWallet         string             `gorm:"column:user_wallet;not null"`
	Type               core.ChannelType   `gorm:"column:type;not null"`
	BlockchainID       uint64             `gorm:"column:blockchain_id;not null"`
	Token              string             `gorm:"column:token;not null"`
	ChallengeDuration  uint32             `gorm:"column:challenge_duration;not null"`
	ChallengeExpiresAt *time.Time         `gorm:"column:challenge_expires_at;default:null"`
	Nonce              uint64             `gorm:"column:nonce;not null;"`
	Status             core.ChannelStatus `gorm:"column:status;not null;"`
	StateVersion       uint64             `gorm:"column:state_version;not null;"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// TableName specifies the table name for the Channel model
func (Channel) TableName() string {
	return "channels"
}

// CreateChannel creates a new channel entity in the database.
func (s *DBStore) CreateChannel(channel core.Channel) error {
	dbChannel := Channel{
		ChannelID:          channel.ChannelID,
		UserWallet:         channel.UserWallet,
		Type:               channel.Type,
		BlockchainID:       channel.BlockchainID,
		Token:              channel.TokenAddress,
		ChallengeDuration:  channel.ChallengeDuration,
		ChallengeExpiresAt: channel.ChallengeExpiresAt,
		Nonce:              channel.Nonce,
		Status:             channel.Status,
		StateVersion:       channel.StateVersion,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.db.Create(&dbChannel).Error; err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	return nil
}

// GetChannelByID retrieves a channel by its unique identifier.
func (s *DBStore) GetChannelByID(channelID string) (*core.Channel, error) {
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
	var state State
	err := s.db.Where("user_wallet = ? AND asset = ?", wallet, asset).
		Order("epoch DESC, version DESC").
		First(&state).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user state: %w", err)
	}

	if state.HomeChannelID == nil {
		return nil, nil
	}

	// Get the channel by its ID
	var dbChannel Channel
	err = s.db.Where("channel_id = ?", *state.HomeChannelID).First(&dbChannel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get home channel: %w", err)
	}

	if dbChannel.Status != core.ChannelStatusOpen || dbChannel.Type != core.ChannelTypeHome {
		return nil, nil
	}

	return databaseChannelToCore(&dbChannel), nil
}

// CheckOpenChannel verifies if a user has an active channel for the given asset.
func (s *DBStore) CheckOpenChannel(wallet, asset string) (bool, error) {
	var count int64
	err := s.db.Raw(`
		SELECT COUNT(s.id)
		FROM channel_states s
		INNER JOIN channels c ON c.channel_id = s.home_channel_id
		WHERE s.user_wallet = ?
			AND s.asset = ?
			AND c.status = ?
			AND c.type = ?
		LIMIT 1
	`, wallet, asset, core.ChannelStatusOpen, core.ChannelTypeHome).Scan(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check open channel: %w", err)
	}

	return count > 0, nil
}

// UpdateChannel persists changes to a channel's metadata (status, version, etc).
func (s *DBStore) UpdateChannel(channel core.Channel) error {
	updates := map[string]interface{}{
		"status":               channel.Status,
		"state_version":        channel.StateVersion,
		"blockchain_id":        channel.BlockchainID,
		"token":                channel.TokenAddress,
		"nonce":                channel.Nonce,
		"challenge_expires_at": channel.ChallengeExpiresAt,
		"updated_at":           time.Now(),
	}

	if err := s.db.Model(&Channel{}).Where("channel_id = ?", channel.ChannelID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	return nil
}
