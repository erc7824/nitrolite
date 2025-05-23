package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// ChannelStatus represents the current state of a channel (open or closed)
type ChannelStatus string

var (
	ChannelStatusJoining ChannelStatus = "joining"
	ChannelStatusOpen    ChannelStatus = "open"
	ChannelStatusClosed  ChannelStatus = "closed"
)

// Channel represents a state channel between participants
type Channel struct {
	ChannelID   string        `gorm:"column:channel_id;primaryKey;"`
	ChainID     uint32        `gorm:"column:chain_id;not null"`
	Token       string        `gorm:"column:token;not null"`
	Participant string        `gorm:"column:participant;not null"`
	Amount      uint64        `gorm:"column:amount;not null"`
	Status      ChannelStatus `gorm:"column:status;not null;"`
	Challenge   uint64        `gorm:"column:challenge;default:0"`
	Nonce       uint64        `gorm:"column:nonce;default:0"`
	Version     uint64        `gorm:"column:version;default:0"`
	Adjudicator string        `gorm:"column:adjudicator;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName specifies the table name for the Channel model
func (Channel) TableName() string {
	return "channels"
}

// CreateChannel creates a new channel in the database
// For real channels, participantB is always the broker application
func CreateChannel(tx *gorm.DB, channelID, participantA string, nonce uint64, adjudicator string, chainID uint32, tokenAddress string, amount uint64) (Channel, error) {
	channel := Channel{
		ChannelID:   channelID,
		Participant: participantA,
		ChainID:     chainID, // Set the network ID for channels
		Status:      ChannelStatusJoining,
		Nonce:       nonce,
		Adjudicator: adjudicator,
		Token:       tokenAddress,
		Amount:      amount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := tx.Create(&channel).Error; err != nil {
		return Channel{}, fmt.Errorf("failed to create channel: %w", err)
	}

	log.Printf("Created new channel with ID: %s, chainID: %d", channelID, chainID)
	return channel, nil
}

// GetChannelByID retrieves a channel by its ID
func GetChannelByID(tx *gorm.DB, channelID string) (*Channel, error) {
	var channel Channel
	if err := tx.Where("channel_id = ?", channelID).First(&channel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Channel not found
		}
		return nil, fmt.Errorf("error finding channel: %w", err)
	}

	return &channel, nil
}

// getChannelsByParticipant finds all channels for a participant
func getChannelsByParticipant(tx *gorm.DB, participant string, status string) ([]Channel, error) {
	var channels []Channel
	q := tx.Where("participant = ?", participant)
	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Order("created_at DESC").Find(&channels).Error; err != nil {
		return nil, fmt.Errorf("error finding channels for participant %s: %w", participant, err)
	}

	return channels, nil
}

// CheckExistingChannels checks if there is an existing open channel on the same network between participant and broker
func CheckExistingChannels(tx *gorm.DB, participantA, token string, chainID uint32) (*Channel, error) {
	var channel Channel
	err := tx.Where("participant = ? AND token = ? AND chain_id = ? AND status = ?", participantA, token, chainID, ChannelStatusOpen).
		First(&channel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No open channel found
		}
		return nil, fmt.Errorf("error checking for existing open channel: %w", err)
	}

	return &channel, nil
}
