package main

import (
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/lib/pq"
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
	SessionData  string         `gorm:"column:session_data;type:text;not null"`
	Quorum       uint64         `gorm:"column:quorum;default:100"`
	Version      uint64         `gorm:"column:version;default:1"`
	Status       ChannelStatus  `gorm:"column:status;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (AppSession) TableName() string {
	return "app_sessions"
}

// isAppSessionID checks if the given string is a valid AppSessionID
func isAppSessionID(appSessionID string) bool {
	// AppSessionID is a hex string of 64 characters prefixed with "0x"
	if len(appSessionID) != 66 {
		return false
	}
	_, err := hexutil.Decode(appSessionID)
	return err == nil
}
