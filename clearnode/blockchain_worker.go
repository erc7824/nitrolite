package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

type BlockchainWorker struct {
	db      *gorm.DB
	custody map[uint32]CustodyInterface
	logger  Logger
}

func NewBlockchainWorker(db *gorm.DB, custody map[uint32]CustodyInterface, logger Logger) *BlockchainWorker {
	return &BlockchainWorker{
		db:      db,
		custody: custody,
		logger:  logger.NewSystem("blockchain-worker"),
	}
}

func (w *BlockchainWorker) Start(ctx context.Context) {
	w.logger.Info("starting blockchain worker with dedicated workers for each chain")
	var wg sync.WaitGroup
	for chainID := range w.custody {
		wg.Add(1)
		go w.runChainWorker(ctx, &wg, chainID)
	}
	<-ctx.Done()
	w.logger.Info("shutdown signal received, waiting for chain workers to stop...")
	wg.Wait()
	w.logger.Info("all chain workers have stopped")
}

func (w *BlockchainWorker) runChainWorker(ctx context.Context, wg *sync.WaitGroup, chainID uint32) {
	defer wg.Done()
	chainLogger := w.logger.With("chain", chainID)
	chainLogger.Info("chain worker started")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	w.processActionsForChain(ctx, chainID, chainLogger)

	for {
		select {
		case <-ctx.Done():
			chainLogger.Info("chain worker stopping")
			return
		case <-ticker.C:
			w.processActionsForChain(ctx, chainID, chainLogger)
		}
	}
}

func (w *BlockchainWorker) processActionsForChain(ctx context.Context, chainID uint32, logger Logger) {
	actions, err := getActionsForChain(w.db, chainID, 5)
	if err != nil {
		logger.Error("failed to get pending actions for chain", "error", err)
		return
	}
	if len(actions) == 0 {
		return
	}

	logger.Debug("processing batch of actions", "count", len(actions))
	for _, action := range actions {
		if ctx.Err() != nil {
			logger.Info("context cancelled, stopping batch processing")
			return
		}
		w.processAction(ctx, action)
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
		logger.Error("custody client not found, failing action", "error", err)
		if err := action.Fail(w.db, err.Error()); err != nil {
			logger.Error("failed to mark action as failed", "error", err)
		}
		return
	}

	var txHash common.Hash
	var err error

	switch action.Type {
	case ActionTypeCheckpoint:
		txHash, err = w.processCheckpoint(ctx, action, custody)
	default:
		err = fmt.Errorf("unknown action type: %s", action.Type)
	}

	if err != nil {
		logger.Error("processing attempt failed, will retry later", "error", err)
		return
	}
	if err := action.Complete(w.db, txHash.Hex()); err != nil {
		logger.Error("failed to mark action as completed", "error", err)
		return
	}
	logger.Info("action completed successfully", "txHash", txHash.Hex())
}

func (w *BlockchainWorker) processCheckpoint(_ context.Context, action BlockchainAction, custody CustodyInterface) (common.Hash, error) {
	var data CheckpointData
	if err := json.Unmarshal([]byte(action.Data), &data); err != nil {
		failErr := fmt.Errorf("unmarshal checkpoint data: %w", err)
		if err := action.Fail(w.db, failErr.Error()); err != nil {
			w.logger.Error("failed to mark action as failed", "error", err)
		}
		return common.Hash{}, failErr
	}
	return custody.Checkpoint(action.ChannelID, data.State, data.UserSig, data.ServerSig, []nitrolite.State{})
}
