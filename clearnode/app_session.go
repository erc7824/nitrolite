package main

import (
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/lib/pq"
)

// AppSession represents a virtual payment application session between participants
type AppSession struct {
	ID                 uint           `gorm:"primaryKey"`
	Protocol           rpc.Version    `gorm:"column:protocol;default:'NitroRPC/0.2';not null"`
	SessionID          string         `gorm:"column:session_id;not null;uniqueIndex"`
	Challenge          uint64         `gorm:"column:challenge;"`
	Nonce              uint64         `gorm:"column:nonce;not null"`
	ParticipantWallets pq.StringArray `gorm:"type:text[];column:participants;not null"`
	Weights            pq.Int64Array  `gorm:"type:integer[];column:weights"`
	SessionData        string         `gorm:"column:session_data;type:text;not null"`
	Quorum             uint64         `gorm:"column:quorum;default:100"`
	Version            uint64         `gorm:"column:version;default:1"`
	Status             ChannelStatus  `gorm:"column:status;not null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (AppSession) TableName() string {
	return "app_sessions"
}
