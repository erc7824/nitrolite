package main

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ActionType string
type ActionStatus string

const (
	ActionTypeCheckpoint ActionType = "checkpoint"
)

const (
	StatusPending   ActionStatus = "pending"
	StatusCompleted ActionStatus = "completed"
	StatusFailed    ActionStatus = "failed"
)

type BlockchainAction struct {
	ID        int64          `gorm:"primary_key"`
	Type      ActionType     `gorm:"column:action_type;not null"`
	ChannelID string         `gorm:"column:channel_id;not null"`
	ChainID   uint32         `gorm:"column:chain_id;not null"`
	Data      datatypes.JSON `gorm:"column:action_data;type:text;not null"`
	Status    ActionStatus   `gorm:"column:status;not null"`
	Retries   int            `gorm:"column:retry_count;default:0"`
	Error     string         `gorm:"column:last_error;type:text"`
	TxHash    string         `gorm:"column:transaction_hash"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
}

func (BlockchainAction) TableName() string {
	return "blockchain_actions"
}

type CheckpointData struct {
	State     UnsignedState `json:"state"`
	UserSig   Signature     `json:"user_sig"`
	ServerSig Signature     `json:"server_sig"`
}

func CreateCheckpoint(tx *gorm.DB, channel string, chainID uint32, state UnsignedState, userSig, serverSig Signature) error {
	data := CheckpointData{
		State:     state,
		UserSig:   userSig,
		ServerSig: serverSig,
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal checkpoint data: %w", err)
	}

	action := &BlockchainAction{
		Type:      ActionTypeCheckpoint,
		ChannelID: channel,
		ChainID:   chainID,
		Data:      bytes,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return tx.Create(action).Error
}

func (a *BlockchainAction) Fail(tx *gorm.DB, err string) error {
	a.Status = StatusFailed
	a.Error = err
	a.Retries++
	a.UpdatedAt = time.Now()
	return tx.Save(a).Error
}

func (a *BlockchainAction) RecordAttempt(tx *gorm.DB, attemptErr string) error {
	a.Retries++
	a.Error = attemptErr
	a.UpdatedAt = time.Now()
	return tx.Save(a).Error
}

func (a *BlockchainAction) Complete(tx *gorm.DB, txHash string) error {
	a.Status = StatusCompleted
	a.TxHash = txHash
	a.Error = ""
	a.UpdatedAt = time.Now()
	return tx.Save(a).Error
}

func getActionsForChain(db *gorm.DB, chainID uint32, limit int) ([]BlockchainAction, error) {
	var actions []BlockchainAction
	query := db.Where("status = ? AND chain_id = ?", StatusPending, chainID).Order("created_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&actions).Error; err != nil {
		return nil, fmt.Errorf("query pending actions for chain %d: %w", chainID, err)
	}
	return actions, nil
}
