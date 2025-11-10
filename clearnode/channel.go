package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ChannelStatus represents the current state of a channel (open or closed)
type ChannelStatus string

var (
	ChannelStatusOpen       ChannelStatus = "open"
	ChannelStatusClosed     ChannelStatus = "closed"
	ChannelStatusResizing   ChannelStatus = "resizing"
	ChannelStatusChallenged ChannelStatus = "challenged"
)

// Channel represents a state channel between participants
type Channel struct {
	ChannelID   string `gorm:"column:channel_id;primaryKey;"`
	ChainID     uint32 `gorm:"column:chain_id;not null"`
	Token       string `gorm:"column:token;not null"`
	Wallet      string `gorm:"column:wallet;not null"`
	Participant string `gorm:"column:participant;not null"`
	// RawAmount represents an Integer value of token amount (wei) as represented on the blockchain
	// type:varchar(78) is set for sqlite to address the issue of not supporting big decimals
	RawAmount            decimal.Decimal `gorm:"column:raw_amount;type:varchar(78);not null"`
	Status               ChannelStatus   `gorm:"column:status;not null;"`
	Challenge            uint64          `gorm:"column:challenge;default:0"`
	Nonce                uint64          `gorm:"column:nonce;default:0"`
	Adjudicator          string          `gorm:"column:adjudicator;not null"`
	State                UnsignedState   `gorm:"column:state;type:text;not null"`
	ServerStateSignature *Signature      `gorm:"column:server_state_signature;type:text"`
	UserStateSignature   *Signature      `gorm:"column:user_state_signature;type:text"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// TableName specifies the table name for the Channel model
func (Channel) TableName() string {
	return "channels"
}

// CreateChannel creates a new channel in the database
func CreateChannel(tx *gorm.DB, channelID, wallet, participantSigner string, nonce uint64, challenge uint64, adjudicator string, chainID uint32, tokenAddress string, amount decimal.Decimal, state UnsignedState) (Channel, error) {
	channel := Channel{
		ChannelID:   channelID,
		Wallet:      wallet,
		Participant: participantSigner,
		ChainID:     chainID,
		Status:      ChannelStatusOpen,
		Nonce:       nonce,
		Adjudicator: adjudicator,
		Challenge:   challenge,
		Token:       tokenAddress,
		RawAmount:   amount,
		State:       state,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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
		q = q.Where("wallet = ?", wallet)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Find(&channels).Error; err != nil {
		return nil, fmt.Errorf("error finding channels for a wallet %s: %w", wallet, err)
	}

	return channels, nil
}

// CheckExistingChannels checks if there is an existing open channel on the same network between participant and broker
func CheckExistingChannels(tx *gorm.DB, wallet, token string, chainID uint32) (*Channel, error) {
	var channel Channel
	err := tx.Where("wallet = ? AND token = ? AND chain_id = ? AND status = ?", wallet, token, chainID, ChannelStatusOpen).
		First(&channel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No open channel found
		}
		return nil, fmt.Errorf("error checking for existing open channel: %w", err)
	}

	return &channel, nil
}

type ChannelAmountSum struct {
	Count int             `gorm:"column:count"`
	Sum   decimal.Decimal `gorm:"column:sum"`
}

func GetChannelAmountSumByWallet(tx *gorm.DB, senderWallet string) (ChannelAmountSum, error) {
	var result ChannelAmountSum
	err := tx.Model(&Channel{}).
		Select("COUNT(channel_id) as count, COALESCE(SUM(CAST(raw_amount AS NUMERIC)), 0) as sum").
		Where("wallet = ? AND status IN (?, ?)", senderWallet, ChannelStatusOpen, ChannelStatusResizing).
		Scan(&result).Error
	if err != nil {
		return ChannelAmountSum{}, fmt.Errorf("error calculating channel amount sum: %w", err)
	}

	return result, nil
}
