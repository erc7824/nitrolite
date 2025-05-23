package main

import (
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// AppSession represents a virtual payment application session between participants
type AppSession struct {
	ID           uint           `gorm:"primaryKey"`
	Protocol     string         `gorm:"column:protocol;default:'NitroRPC/0.2';not null"`
	SessionID    string         `gorm:"column:session_id;not null;uniqueIndex"`
	Challenge    uint64         `gorm:"column:challenge;"`
	Nonce        uint64         `gorm:"column:nonce;not null"`
	Participants pq.StringArray `gorm:"type:text[];column:participants;not null"`
	Weights      pq.Int64Array  `gorm:"type:integer[];column:weights"`
	Quorum       uint64         `gorm:"column:quorum;default:100"`
	Version      uint64         `gorm:"column:version;default:1"`
	Status       ChannelStatus  `gorm:"column:status;not null"`
}

func (AppSession) TableName() string {
	return "app_sessions"
}

// getAppSessionsForParticipant finds all channels for a participant
func getAppSessionsForParticipant(tx *gorm.DB, participant string, status string) ([]AppSession, error) {
	var sessions []AppSession
	switch tx.Dialector.Name() {
	case "postgres":
		tx = tx.Where("? = ANY(participants)", participant)
	case "sqlite":
		tx = tx.Where("instr(participants, ?) > 0", participant)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", tx.Dialector.Name())
	}
	if status != "" {
		tx = tx.Where("status = ?", status)
	}

	if err := tx.Find(&sessions).Error; err != nil {
		return nil, err
	}

	return sessions, nil
}
