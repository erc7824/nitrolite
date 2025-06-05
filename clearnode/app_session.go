package main

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// AppSession represents a virtual payment application session between participants
type AppSession struct {
	ID                 uint           `gorm:"primaryKey"`
	Protocol           string         `gorm:"column:protocol;default:'NitroRPC/0.2';not null"`
	SessionID          string         `gorm:"column:session_id;not null;uniqueIndex"`
	Challenge          uint64         `gorm:"column:challenge;"`
	Nonce              uint64         `gorm:"column:nonce;not null"`
	ParticipantWallets pq.StringArray `gorm:"type:text[];column:participants;not null"`
	Weights            pq.Int64Array  `gorm:"type:integer[];column:weights"`
	Quorum             uint64         `gorm:"column:quorum;default:100"`
	Version            uint64         `gorm:"column:version;default:1"`
	Status             ChannelStatus  `gorm:"column:status;not null"`
}

func (AppSession) TableName() string {
	return "app_sessions"
}

// getAppSessions finds all app sessions
// If participantWallet is specified, it returns only sessions for that participant
// If participantWallet is empty, it returns all sessions
func getAppSessions(tx *gorm.DB, participantWallet string, status string) ([]AppSession, error) {
	var sessions []AppSession
	query := tx

	if participantWallet != "" {
		switch tx.Dialector.Name() {
		case "postgres":
			query = query.Where("? = ANY(participants)", participantWallet)
		case "sqlite":
			query = query.Where("instr(participants, ?) > 0", participantWallet)
		default:
			return nil, fmt.Errorf("unsupported database driver: %s", tx.Dialector.Name())
		}
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&sessions).Error; err != nil {
		return nil, err
	}

	return sessions, nil
}

// getAppSessionByID generates a session ID from the given app definition
func getAppSessionID(appDedinition AppDefinition) common.Hash {
	b, _ := json.Marshal(appDedinition)
	return crypto.Keccak256Hash(b)
}
