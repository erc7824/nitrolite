package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type MockCustody struct {
	checkpointFn func(channelID string, state UnsignedState, userSig, serverSig Signature, proofs []any) (common.Hash, error)
	callCount    int
}

func (m *MockCustody) Checkpoint(channelID string, state UnsignedState, userSig, serverSig Signature, proofs []any) (common.Hash, error) {
	m.callCount++
	if m.checkpointFn != nil {
		return m.checkpointFn(channelID, state, userSig, serverSig, proofs)
	}
	return common.HexToHash("0x1234567890abcdef"), nil
}

func setupWorker(t *testing.T, custodyClients map[uint32]*MockCustody) (*BlockchainWorker, *gorm.DB, func()) {
	t.Helper()
	db, cleanup := setupTestDB(t)
	logger := NewLoggerIPFS("test")

	custodyMap := make(map[uint32]*Custody)
	for chainID := range custodyClients {
		custodyMap[chainID] = &Custody{chainID: chainID}
	}

	worker := NewBlockchainWorker(db, custodyMap, logger)
	return worker, db, cleanup
}

func TestNewBlockchainWorker(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	logger := NewLoggerIPFS("test")
	custody := map[uint32]*Custody{1: {chainID: 1}}

	worker := NewBlockchainWorker(db, custody, logger)

	assert.NotNil(t, worker)
	assert.Equal(t, db, worker.db)
	assert.Equal(t, custody, worker.custody)
	assert.NotNil(t, worker.logger)
}

func TestGetPendingActions(t *testing.T) {
	t.Run("Orders by created time", func(t *testing.T) {
		worker, db, cleanup := setupWorker(t, map[uint32]*MockCustody{1: {}})
		defer cleanup()

		now := time.Now()
		actions := []BlockchainAction{
			{Type: ActionTypeCheckpoint, ChannelID: "channel3", ChainID: 1, Data: "{}", Status: StatusPending, Created: now.Add(2 * time.Second)},
			{Type: ActionTypeCheckpoint, ChannelID: "channel1", ChainID: 1, Data: "{}", Status: StatusPending, Created: now},
			{Type: ActionTypeCheckpoint, ChannelID: "channel2", ChainID: 1, Data: "{}", Status: StatusCompleted, Created: now.Add(1 * time.Second)},
		}

		for _, action := range actions {
			require.NoError(t, db.Create(&action).Error)
		}

		result, err := worker.getPendingActions(10)
		require.NoError(t, err)

		assert.Len(t, result, 2)
		assert.Equal(t, "channel1", result[0].ChannelID)
		assert.Equal(t, "channel3", result[1].ChannelID)
	})

	t.Run("Respects limit", func(t *testing.T) {
		worker, db, cleanup := setupWorker(t, map[uint32]*MockCustody{1: {}})
		defer cleanup()

		for i := range 5 {
			action := BlockchainAction{
				Type: ActionTypeCheckpoint, ChannelID: fmt.Sprintf("channel%d", i),
				ChainID: 1, Data: "{}", Status: StatusPending, Created: time.Now(),
			}
			require.NoError(t, db.Create(&action).Error)
		}

		result, err := worker.getPendingActions(3)
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("No limit", func(t *testing.T) {
		worker, db, cleanup := setupWorker(t, map[uint32]*MockCustody{1: {}})
		defer cleanup()

		for i := range 3 {
			action := BlockchainAction{
				Type: ActionTypeCheckpoint, ChannelID: fmt.Sprintf("channel%d", i),
				ChainID: 1, Data: "{}", Status: StatusPending, Created: time.Now(),
			}
			require.NoError(t, db.Create(&action).Error)
		}

		result, err := worker.getPendingActions(0)
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("Database error", func(t *testing.T) {
		worker, db, cleanup := setupWorker(t, map[uint32]*MockCustody{1: {}})
		defer cleanup()

		sqlDB, err := db.DB()
		require.NoError(t, err)
		sqlDB.Close()

		result, err := worker.getPendingActions(10)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "query pending actions")
	})
}

func TestProcessBatch(t *testing.T) {
	worker, db, cleanup := setupWorker(t, map[uint32]*MockCustody{1: {}})
	defer cleanup()

	actions := []BlockchainAction{
		{ID: 1, Type: ActionTypeCheckpoint, ChannelID: "channel1", ChainID: 999, Data: validCheckpointData(t), Status: StatusPending},
		{ID: 2, Type: ActionTypeCheckpoint, ChannelID: "channel2", ChainID: 999, Data: validCheckpointData(t), Status: StatusPending},
	}

	for _, action := range actions {
		require.NoError(t, db.Create(&action).Error)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	worker.processBatch(ctx, actions)
}

func TestProcessCheckpoint(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &MockCustody{
			checkpointFn: func(channelID string, state UnsignedState, userSig, serverSig Signature, proofs []any) (common.Hash, error) {
				assert.Equal(t, "test-channel", channelID)
				assert.Equal(t, uint64(5), state.Version)
				return common.HexToHash("0xabcdef1234567890"), nil
			},
		}

		_, _, cleanup := setupWorker(t, map[uint32]*MockCustody{1: mock})
		defer cleanup()

		action := BlockchainAction{ChannelID: "test-channel", ChainID: 1, Data: validCheckpointData(t)}

		var data CheckpointData
		err := json.Unmarshal([]byte(action.Data), &data)
		require.NoError(t, err)
		assert.Equal(t, uint64(5), data.State.Version)
	})

	t.Run("Invalid data", func(t *testing.T) {
		_, _, cleanup := setupWorker(t, map[uint32]*MockCustody{1: {}})
		defer cleanup()

		action := BlockchainAction{ChannelID: "test-channel", ChainID: 1, Data: "invalid-json"}

		var data CheckpointData
		err := json.Unmarshal([]byte(action.Data), &data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character")
	})
}

func TestStart(t *testing.T) {
	worker, _, cleanup := setupWorker(t, map[uint32]*MockCustody{1: {}})
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	worker.Start(ctx)
	<-ctx.Done()
}

func TestProcessPending(t *testing.T) {
	t.Run("Multiple batches", func(t *testing.T) {
		worker, db, cleanup := setupWorker(t, map[uint32]*MockCustody{1: {}})
		defer cleanup()

		for i := range 12 {
			action := BlockchainAction{
				Type: ActionTypeCheckpoint, ChannelID: fmt.Sprintf("channel%d", i), ChainID: 888,
				Data: validCheckpointData(t), Status: StatusPending, Created: time.Now(),
			}
			require.NoError(t, db.Create(&action).Error)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		worker.processPending(ctx)
	})

	t.Run("Empty result", func(t *testing.T) {
		worker, _, cleanup := setupWorker(t, map[uint32]*MockCustody{1: {}})
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		worker.processPending(ctx)
	})
}

func TestCustodyRouting(t *testing.T) {
	t.Run("Correct chain routing", func(t *testing.T) {
		worker, db, cleanup := setupWorker(t, map[uint32]*MockCustody{1: {}, 2: {}})
		defer cleanup()

		actions := []BlockchainAction{
			{Type: ActionTypeCheckpoint, ChannelID: "channel1", ChainID: 1, Data: validCheckpointData(t), Status: StatusPending},
			{Type: ActionTypeCheckpoint, ChannelID: "channel2", ChainID: 2, Data: validCheckpointData(t), Status: StatusPending},
		}

		for _, action := range actions {
			require.NoError(t, db.Create(&action).Error)
		}

		assert.Contains(t, worker.custody, uint32(1))
		assert.Contains(t, worker.custody, uint32(2))
		assert.Equal(t, uint32(1), worker.custody[1].chainID)
		assert.Equal(t, uint32(2), worker.custody[2].chainID)
	})

	t.Run("Missing custody client", func(t *testing.T) {
		worker, db, cleanup := setupWorker(t, map[uint32]*MockCustody{1: {}})
		defer cleanup()

		action := BlockchainAction{Type: ActionTypeCheckpoint, ChannelID: "channel1", ChainID: 999, Data: validCheckpointData(t), Status: StatusPending}
		require.NoError(t, db.Create(&action).Error)

		assert.NotContains(t, worker.custody, uint32(999))
	})
}

func TestErrorHandling(t *testing.T) {
	worker, db, cleanup := setupWorker(t, map[uint32]*MockCustody{1: {}})
	defer cleanup()

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	worker.processPending(ctx)
}

func validCheckpointData(t *testing.T) string {
	t.Helper()
	data := CheckpointData{
		State: UnsignedState{
			Intent: StateIntent(1), Version: 5, Data: "test-data",
			Allocations: []Allocation{{Participant: "0xUser123", TokenAddress: "0xToken456", RawAmount: decimal.NewFromInt(1000)}},
		},
		UserSig: Signature{1, 2, 3}, ServerSig: Signature{4, 5, 6},
	}
	bytes, err := json.Marshal(data)
	require.NoError(t, err)
	return string(bytes)
}
