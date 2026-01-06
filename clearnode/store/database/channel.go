package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ChannelStatus represents the current state of a channel
type ChannelStatus string

var (
	ChannelStatusOpen       ChannelStatus = "open"
	ChannelStatusClosed     ChannelStatus = "closed"
	ChannelStatusChallenged ChannelStatus = "challenged"
)

// ChannelType represents the type of channel (escrow or home)
type ChannelType string

var (
	ChannelTypeEscrow ChannelType = "escrow"
	ChannelTypeHome   ChannelType = "home"
)

// Channel represents a state channel between participants
type Channel struct {
	ChannelID           string        `gorm:"column:channel_id;primaryKey;"`
	UserWallet          string        `gorm:"column:user_wallet;not null"`
	Type                ChannelType   `gorm:"column:type;not null"`
	BlockchainID        uint32        `gorm:"column:blockchain_id;not null"`
	Token               string        `gorm:"column:token;not null"`
	Challenge           uint64        `gorm:"column:challenge;default:0"`
	Nonce               uint64        `gorm:"column:nonce;default:0"`
	Status              ChannelStatus `gorm:"column:status;not null;"`
	OnChainStateVersion uint64        `gorm:"column:on_chain_state_version;default:0"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// TableName specifies the table name for the Channel model
func (Channel) TableName() string {
	return "channels"
}

// CreateChannel creates a new channel in the database (legacy version with all fields)
func CreateChannel(tx *gorm.DB, channelID, wallet, participantSigner string, nonce uint64, challenge uint64, adjudicator string, chainID uint32, tokenAddress string, amount decimal.Decimal, state UnsignedState) (Channel, error) {
	channel := Channel{
		ChannelID:           channelID,
		UserWallet:          wallet,
		BlockchainID:        chainID,
		Status:              ChannelStatusOpen,
		Nonce:               nonce,
		Challenge:           challenge,
		Token:               tokenAddress,
		OnChainStateVersion: 0,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := tx.Create(&channel).Error; err != nil {
		return Channel{}, fmt.Errorf("failed to create channel: %w", err)
	}

	return channel, nil
}

// CreateChannelNew creates a new channel in the database (new version without legacy fields)
func CreateChannelNew(tx *gorm.DB, channelID, userWallet string, channelType ChannelType, blockchainID uint32, tokenAddress string, nonce uint64, challenge uint64) (Channel, error) {
	channel := Channel{
		ChannelID:           channelID,
		UserWallet:          userWallet,
		Type:                channelType,
		BlockchainID:        blockchainID,
		Token:               tokenAddress,
		Nonce:               nonce,
		Challenge:           challenge,
		Status:              ChannelStatusOpen,
		OnChainStateVersion: 0,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := tx.Create(&channel).Error; err != nil {
		return Channel{}, fmt.Errorf("failed to create channel: %w", err)
	}

	return channel, nil
}

// GetChannelByID retrieves a channel by its ID
func GetChannelByID(tx *gorm.DB, channelID string) (*Channel, error) {
	var channel Channel
	if err := tx.Where("channel_id = ?", channelID).First(&channel).Error; err != nil {
		return nil, err
	}

	return &channel, nil
}

// getChannelsByWallet finds all channels for a wallet
func getChannelsByWallet(tx *gorm.DB, wallet string, status string) ([]Channel, error) {
	var channels []Channel
	q := tx
	if wallet != "" {
		q = q.Where("user_wallet = ?", wallet)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Find(&channels).Error; err != nil {
		return nil, fmt.Errorf("error finding channels for a wallet %s: %w", wallet, err)
	}

	return channels, nil
}

// CheckExistingChannels checks if there is an existing open channel on the same network between user and node
func CheckExistingChannels(tx *gorm.DB, wallet, token string, blockchainID uint32) (*Channel, error) {
	var channel Channel
	err := tx.Where("user_wallet = ? AND token = ? AND blockchain_id = ? AND status = ?", wallet, token, blockchainID, ChannelStatusOpen).
		First(&channel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No open channel found
		}
		return nil, fmt.Errorf("error checking for existing open channel: %w", err)
	}

	return &channel, nil
}

// Note: Channel balances are now tracked in the State model, not in the Channel model
