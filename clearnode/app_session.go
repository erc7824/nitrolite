package main

import (
	"fmt"
	"time"

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
	SessionData	   string
	Quorum             uint64         `gorm:"column:quorum;default:100"`
	Version            uint64         `gorm:"column:version;default:1"`
	Status             ChannelStatus  `gorm:"column:status;not null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
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

	query = query.Order("updated_at DESC")
	if err := query.Find(&sessions).Error; err != nil {
		return nil, err
	}

	return sessions, nil
}

// verifyQuorum loads an open AppSession, verifies signatures meet quorum
func verifyQuorum(tx *gorm.DB, appSessionID string, rpcSigners map[string]struct{}) (AppSession, map[string]int64, error) {
	var session AppSession
	if err := tx.Where("session_id = ? AND status = ?", appSessionID, ChannelStatusOpen).
		Order("nonce DESC").First(&session).Error; err != nil {
		return AppSession{}, nil, fmt.Errorf("virtual app not found or not open: %w", err)
	}

	participantWeights := make(map[string]int64, len(session.ParticipantWallets))
	for i, addr := range session.ParticipantWallets {
		participantWeights[addr] = session.Weights[i]
	}

	var totalWeight int64
	for wallet := range rpcSigners {
		weight, ok := participantWeights[wallet]
		if !ok {
			return AppSession{}, nil, fmt.Errorf("signature from unknown participant wallet %s", wallet)
		}
		if weight <= 0 {
			return AppSession{}, nil, fmt.Errorf("zero weight for signer %s", wallet)
		}
		totalWeight += weight
	}

	if totalWeight < int64(session.Quorum) {
		return AppSession{}, nil, fmt.Errorf("quorum not met: %d / %d", totalWeight, session.Quorum)
	}

	return session, participantWeights, nil
}
