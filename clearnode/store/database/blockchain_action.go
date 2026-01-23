package database

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BlockchainActionType uint8

const (
	ActionTypeCheckpoint               BlockchainActionType = 1
	ActionTypeInitiateEscrowWithdrawal BlockchainActionType = 21
	ActionTypeFinalizeEscrowDeposit    BlockchainActionType = 12
	ActionTypeFinalizeEscrowWithdrawal BlockchainActionType = 22
)

type BlockchainActionStatus uint8

const (
	BlockchainActionStatusPending BlockchainActionStatus = iota
	BlockchainActionStatusCompleted
	BlockchainActionStatusFailed
)

type BlockchainAction struct {
	ID      int64                `gorm:"primary_key"`
	Type    BlockchainActionType `gorm:"column:action_type;not null"`
	StateID string               `gorm:"column:state_id;not null"`
	// ChainID   uint32                 `gorm:"column:chain_id;not null"`
	Data      datatypes.JSON         `gorm:"column:action_data;type:text;not null"`
	Status    BlockchainActionStatus `gorm:"column:status;not null"`
	Retries   int                    `gorm:"column:retry_count;default:0"`
	Error     string                 `gorm:"column:last_error;type:text"`
	TxHash    common.Hash            `gorm:"column:transaction_hash"`
	CreatedAt time.Time              `gorm:"column:created_at"`
	UpdatedAt time.Time              `gorm:"column:updated_at"`
}

func (BlockchainAction) TableName() string {
	return "blockchain_actions"
}

// ScheduleInitiateEscrowWithdrawal queues a blockchain action to initiate withdrawal.
func (s *DBStore) ScheduleInitiateEscrowWithdrawal(stateID string) error {
	// bytes, err := json.Marshal(data)
	// if err != nil {
	// 	return fmt.Errorf("marshal checkpoint data: %w", err)
	// }

	action := &BlockchainAction{
		Type:    ActionTypeInitiateEscrowWithdrawal,
		StateID: stateID,
		// ChainID: 1,
		// Data:      bytes,
		Status:    BlockchainActionStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.db.Create(action).Error
}

func (s *DBStore) ScheduleCheckpoint(stateID string) error {
	// bytes, err := json.Marshal(data)
	// if err != nil {
	// 	return fmt.Errorf("marshal checkpoint data: %w", err)
	// }

	action := &BlockchainAction{
		Type:    ActionTypeCheckpoint,
		StateID: stateID,
		// Data:      bytes,
		Status:    BlockchainActionStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.db.Create(action).Error
}

// ScheduleFinalizeEscrowDeposit schedules a finalize for an escrow deposit operation.
func (s *DBStore) ScheduleFinalizeEscrowDeposit(stateID string) error {
	// bytes, err := json.Marshal(data)
	// if err != nil {
	// 	return fmt.Errorf("marshal checkpoint data: %w", err)
	// }

	action := &BlockchainAction{
		Type:    ActionTypeFinalizeEscrowDeposit,
		StateID: stateID,
		// Data:      bytes,
		Status:    BlockchainActionStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.db.Create(action).Error
}

// ScheduleFinalizeEscrowWithdrawal schedules a finalize for an escrow withdrawal operation.
func (s *DBStore) ScheduleFinalizeEscrowWithdrawal(stateID string) error {
	// bytes, err := json.Marshal(data)
	// if err != nil {
	// 	return fmt.Errorf("marshal checkpoint data: %w", err)
	// }

	action := &BlockchainAction{
		Type:    ActionTypeFinalizeEscrowWithdrawal,
		StateID: stateID,
		// Data:      bytes,
		Status:    BlockchainActionStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.db.Create(action).Error
}

func (a *BlockchainAction) Fail(tx *gorm.DB, err string) error {
	a.Status = BlockchainActionStatusFailed
	a.Error = err
	a.Retries++
	a.UpdatedAt = time.Now()
	return tx.Save(a).Error
}

func (a *BlockchainAction) FailNoRetry(tx *gorm.DB, err string) error {
	a.Status = BlockchainActionStatusFailed
	a.Error = err
	a.UpdatedAt = time.Now()
	return tx.Save(a).Error
}

func (a *BlockchainAction) RecordAttempt(tx *gorm.DB, attemptErr string) error {
	a.Retries++
	a.Error = attemptErr
	a.UpdatedAt = time.Now()
	return tx.Save(a).Error
}

func (a *BlockchainAction) Complete(tx *gorm.DB, txHash common.Hash) error {
	a.Status = BlockchainActionStatusCompleted
	a.TxHash = txHash
	a.Error = ""
	a.UpdatedAt = time.Now()
	return tx.Save(a).Error
}

func GetActionsForChain(db *gorm.DB, chainID uint32, limit int) ([]BlockchainAction, error) {
	var actions []BlockchainAction
	query := db.Where("status = ? AND chain_id = ?", BlockchainActionStatusPending, chainID).Order("created_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&actions).Error; err != nil {
		return nil, fmt.Errorf("query pending actions for chain %d: %w", chainID, err)
	}
	return actions, nil
}
