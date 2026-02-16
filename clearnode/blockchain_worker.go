package main

import (
	"context"

	"github.com/erc7824/nitrolite/clearnode/store/database"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/log"
)

// const (
// 	// actionBatchSize determines how many blockchain actions to process at once
// 	actionBatchSize = 20

// 	// maxActionRetries is the maximum number of times to retry a failed action
// 	maxActionRetries = 5

// 	// chainWorkerTickInterval is how frequently each chain worker checks for new actions
// 	chainWorkerTickInterval = 30 * time.Second

// 	unmarshalCheckpointDataError = "unmarshal checkpoint data"
// )

type BlockchainWorker struct {
	client core.Client
	db     database.DatabaseStore
	logger log.Logger
}

func NewBlockchainWorker(client core.Client, db database.DatabaseStore, logger log.Logger) *BlockchainWorker {
	return &BlockchainWorker{
		client: client,
		db:     db,
		logger: logger.WithName("bw"),
	}
}

func (w *BlockchainWorker) Start(ctx context.Context) {
	w.logger.Info("starting blockchain worker")

	// w.logger.Debug("starting blockchain worker with dedicated workers for each chain")
	// var wg sync.WaitGroup
	// for chainID := range w.custody {
	// 	wg.Add(1)
	// 	go w.runChainWorker(ctx, &wg, chainID)
	// }
	// w.logger.Info("blockchain workers started")
	// <-ctx.Done()
	// w.logger.Debug("shutdown signal received, waiting for chain workers to stop...")
	// wg.Wait()
	// w.logger.Info("all chain workers have stopped")
}

// func (w *BlockchainWorker) runChainWorker(ctx context.Context, wg *sync.WaitGroup, chainID uint32) {
// 	defer wg.Done()
// 	chainLogger := w.logger.WithKV("chain", chainID)
// 	chainLogger.Info("chain worker started", "chainId", chainID)

// 	ticker := time.NewTicker(chainWorkerTickInterval)
// 	defer ticker.Stop()

// 	w.processActionsForChain(ctx, chainID, chainLogger)

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			chainLogger.Debug("chain worker stopping")
// 			defer chainLogger.Info("chain worker stopped")
// 			return
// 		case <-ticker.C:
// 			w.processActionsForChain(ctx, chainID, chainLogger)
// 		}
// 	}
// }

// func (w *BlockchainWorker) processActionsForChain(ctx context.Context, chainID uint32, logger log.Logger) {
// 	actions, err := db.GetActionsForChain(w.db, chainID, actionBatchSize)
// 	if err != nil {
// 		logger.Error("failed to get pending actions for chain", "error", err)
// 		return
// 	}
// 	if len(actions) == 0 {
// 		return
// 	}

// 	logger.Debug("processing batch of actions", "count", len(actions))
// 	for _, action := range actions {
// 		if ctx.Err() != nil {
// 			logger.Info("context cancelled, stopping batch processing")
// 			return
// 		}
// 		w.processAction(ctx, action)
// 	}
// }

// func (w *BlockchainWorker) processAction(ctx context.Context, action db.BlockchainAction) {
// 	logger := w.logger.
// 		WithKV("id", action.ID).
// 		WithKV("type", action.Type).
// 		WithKV("channel", action.ChannelID).
// 		WithKV("chain", action.ChainID).
// 		WithKV("attempt", action.Retries)

// 	custody, exists := w.custody[action.ChainID]
// 	if !exists {
// 		err := fmt.Errorf("no custody client for chain %d", action.ChainID)
// 		logger.Error("custody client not found, failing action", "error", err)
// 		if err := action.Fail(w.db, err.Error()); err != nil {
// 			logger.Error("failed to mark action as failed", "error", err)
// 		}
// 		return
// 	}

// 	var txHash common.Hash
// 	var err error

// 	switch action.Type {
// 	case db.ActionTypeCheckpoint:
// 		txHash, err = w.processCheckpoint(ctx, action, custody)
// 	default:
// 		err = fmt.Errorf("unknown action type: %s", action.Type)
// 	}

// 	if err != nil {
// 		isFatalError := strings.Contains(err.Error(), unmarshalCheckpointDataError)

// 		if isFatalError {
// 			logger.Error("action failed due to fatal data error", "error", err)
// 			if failErr := action.Fail(w.db, err.Error()); failErr != nil {
// 				logger.Error("failed to mark action as permanently failed", "error", failErr)
// 			}
// 		} else {
// 			if action.Retries >= maxActionRetries {
// 				logger.Warn("action failed after reaching max retries", "error", err)
// 				finalErr := fmt.Errorf("failed after %d retries: %w", action.Retries, err)

// 				if saveErr := action.FailNoRetry(w.db, finalErr.Error()); saveErr != nil {
// 					logger.Error("failed to mark action as permanently failed", "error", saveErr)
// 				}
// 			} else {
// 				logger.Error("processing attempt failed, will retry later", "error", err)
// 				if recordErr := action.RecordAttempt(w.db, err.Error()); recordErr != nil {
// 					logger.Error("failed to record failed attempt", "error", recordErr)
// 				}
// 			}
// 		}
// 		return
// 	}

// 	// Success case
// 	if err := action.Complete(w.db, txHash); err != nil {
// 		logger.Error("failed to mark action as completed", "error", err)
// 		return
// 	}
// 	logger.Info("action completed successfully", "txHash", txHash.Hex())
// }

// func (w *BlockchainWorker) processCheckpoint(ctx context.Context, action db.BlockchainAction, custody custody.CustodyInterface) (common.Hash, error) {
// 	var data db.CheckpointData
// 	if err := json.Unmarshal([]byte(action.Data), &data); err != nil {
// 		return common.Hash{}, fmt.Errorf("%s: %w", unmarshalCheckpointDataError, err)
// 	}

// 	return custody.Checkpoint(action.ChannelID, data.State, data.UserSig, data.ServerSig, []nitrolite.State{})
// }
