package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

type BlockchainWorker struct {
	db      *gorm.DB
	custody map[uint32]*Custody
	logger  Logger
}

func NewBlockchainWorker(db *gorm.DB, custody map[uint32]*Custody, logger Logger) *BlockchainWorker {
	return &BlockchainWorker{
		db:      db,
		custody: custody,
		logger:  logger.NewSystem("blockchain-worker"),
	}
}

func (w *BlockchainWorker) Start(ctx context.Context) {
	w.logger.Info("starting blockchain worker")

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.logger.Info("stopping blockchain worker")
				return
			case <-ticker.C:
				w.processPending(ctx)
			}
		}
	}()
}

func (w *BlockchainWorker) processPending(ctx context.Context) {
	for {
		actions, err := w.getPendingActions(5)
		if err != nil {
			w.logger.Error("failed to get pending actions", "error", err)
			return
		}
		if len(actions) == 0 {
			return
		}

		w.logger.Debug("processing batch", "count", len(actions))
		w.processBatch(ctx, actions)
	}
}

func (w *BlockchainWorker) getPendingActions(limit int) ([]BlockchainAction, error) {
	var actions []BlockchainAction
	query := w.db.Where("status = ?", StatusPending).Order("created_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&actions).Error; err != nil {
		return nil, fmt.Errorf("query pending actions: %w", err)
	}
	return actions, nil
}

func (w *BlockchainWorker) processBatch(ctx context.Context, actions []BlockchainAction) {
	done := make(chan struct{}, len(actions))

	for _, action := range actions {
		go func(action BlockchainAction) {
			w.processAction(ctx, action)
			done <- struct{}{}
		}(action)
	}

	for range len(actions) {
		<-done
	}
}

func (w *BlockchainWorker) processAction(ctx context.Context, action BlockchainAction) {
	logger := w.logger.
		With("id", action.ID).
		With("type", action.Type).
		With("channel", action.ChannelID).
		With("chain", action.ChainID)

	custody, exists := w.custody[action.ChainID]
	if !exists {
		err := fmt.Errorf("no custody client for chain %d", action.ChainID)
		logger.Error("custody not found", "error", err)
		action.Fail(w.db, err.Error())
		return
	}

	const maxRetries = 5
	for attempt := 0; attempt <= maxRetries; attempt++ {
		action.Retries = attempt

		var txHash common.Hash
		var err error

		switch action.Type {
		case ActionTypeCheckpoint:
			txHash, err = w.processCheckpoint(ctx, action, custody)
		default:
			err = fmt.Errorf("unknown action type: %s", action.Type)
		}

		if err == nil {
			if err := action.Complete(w.db, txHash.Hex()); err != nil {
				logger.Error("failed to mark completed", "error", err)
				return
			}
			logger.Info("completed", "txHash", txHash.Hex(), "attempts", attempt+1)
			return
		}

		logger.Error("attempt failed", "error", err, "attempt", attempt+1)

		if attempt >= maxRetries {
			action.Fail(w.db, err.Error())
			logger.Error("max retries exceeded")
			return
		}

		backoff := int(math.Min(math.Pow(2, float64(attempt+1)), 30))
		delay := time.Duration(backoff) * time.Minute
		logger.Info("retrying", "delay", delay)

		select {
		case <-ctx.Done():
			logger.Info("context cancelled")
			return
		case <-time.After(delay):
		}
	}
}

func (w *BlockchainWorker) processCheckpoint(_ context.Context, action BlockchainAction, custody *Custody) (common.Hash, error) {
	var data CheckpointData
	if err := json.Unmarshal([]byte(action.Data), &data); err != nil {
		return common.Hash{}, fmt.Errorf("unmarshal checkpoint data: %w", err)
	}

	return custody.Checkpoint(action.ChannelID, data.State, data.UserSig, data.ServerSig, []nitrolite.State{})
}
