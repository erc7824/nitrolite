package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type MockCustody struct {
	checkpointFn func() (common.Hash, error)
	mu           sync.Mutex
	callCount    int
}

var _ CustodyInterface = (*MockCustody)(nil)

func (m *MockCustody) Checkpoint(channelID common.Hash, state UnsignedState, userSig, serverSig Signature, proofs []nitrolite.State) (common.Hash, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()

	if m.checkpointFn != nil {
		return m.checkpointFn()
	}
	return common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"), nil
}

func (m *MockCustody) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func setupWorker(t *testing.T, custodyClients map[uint32]CustodyInterface) (*BlockchainWorker, *gorm.DB, func()) {
	t.Helper()
	db, cleanup := setupTestDB(t)
	logger := NewLoggerIPFS("test")
	worker := NewBlockchainWorker(db, custodyClients, logger)
	return worker, db, cleanup
}

func validCheckpointData(t *testing.T) []byte {
	t.Helper()
	data := CheckpointData{
		State:     UnsignedState{Version: 1},
		UserSig:   Signature{1},
		ServerSig: Signature{2},
	}
	bytes, err := json.Marshal(data)
	require.NoError(t, err)
	return bytes
}

func TestGetPendingActionsForChain(t *testing.T) {
	worker, db, cleanup := setupWorker(t, map[uint32]CustodyInterface{1: &MockCustody{}})
	defer cleanup()

	channelIdA := common.HexToHash("ch1-a")
	channelIdB := common.HexToHash("ch1-b")

	require.NoError(t, db.Create(&BlockchainAction{ChannelID: channelIdB, ChainID: 1, Status: StatusPending, Data: []byte{1}, CreatedAt: time.Now()}).Error)
	require.NoError(t, db.Create(&BlockchainAction{ChannelID: channelIdA, ChainID: 1, Status: StatusPending, Data: []byte{1}, CreatedAt: time.Now().Add(-time.Second)}).Error)

	result, err := getActionsForChain(worker.db, 1, 5)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, channelIdA, result[0].ChannelID)
}

func TestProcessAction(t *testing.T) {
	t.Run("Success case completes the action", func(t *testing.T) {
		mockCustody := &MockCustody{}
		worker, db, cleanup := setupWorker(t, map[uint32]CustodyInterface{1: mockCustody})
		defer cleanup()

		action := &BlockchainAction{Type: ActionTypeCheckpoint, ChainID: 1, Data: validCheckpointData(t), Status: StatusPending}
		require.NoError(t, db.Create(action).Error)

		worker.processAction(context.Background(), *action)

		assert.Equal(t, 1, mockCustody.CallCount())
		var updatedAction BlockchainAction
		require.NoError(t, db.First(&updatedAction, action.ID).Error)
		assert.Equal(t, StatusCompleted, updatedAction.Status)
		assert.Equal(t, 0, updatedAction.Retries)
		expected := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		assert.Equal(t, expected, updatedAction.TxHash)
	})

	t.Run("Permanent failure for missing custody client", func(t *testing.T) {
		worker, db, cleanup := setupWorker(t, map[uint32]CustodyInterface{})
		defer cleanup()

		action := &BlockchainAction{Type: ActionTypeCheckpoint, ChainID: 999, Data: validCheckpointData(t), Status: StatusPending}
		require.NoError(t, db.Create(action).Error)

		worker.processAction(context.Background(), *action)

		var updatedAction BlockchainAction
		require.NoError(t, db.First(&updatedAction, action.ID).Error)
		assert.Equal(t, StatusFailed, updatedAction.Status)
		assert.Contains(t, updatedAction.Error, "no custody client for chain 999")
	})

	t.Run("Transient error on first attempt increments retries and leaves action pending", func(t *testing.T) {
		mockCustody := &MockCustody{
			checkpointFn: func() (common.Hash, error) {
				return common.Hash{}, errors.New("RPC node is down")
			},
		}
		worker, db, cleanup := setupWorker(t, map[uint32]CustodyInterface{1: mockCustody})
		defer cleanup()

		action := &BlockchainAction{Type: ActionTypeCheckpoint, ChainID: 1, Data: validCheckpointData(t), Status: StatusPending, Retries: 0}
		require.NoError(t, db.Create(action).Error)

		worker.processAction(context.Background(), *action)

		assert.Equal(t, 1, mockCustody.CallCount())
		var updatedAction BlockchainAction
		require.NoError(t, db.First(&updatedAction, action.ID).Error)
		assert.Equal(t, StatusPending, updatedAction.Status)
		assert.Equal(t, 1, updatedAction.Retries)
		assert.Equal(t, "RPC node is down", updatedAction.Error)
	})

	t.Run("Permanent failure for invalid action data fails the action", func(t *testing.T) {
		mockCustody := &MockCustody{}
		worker, db, cleanup := setupWorker(t, map[uint32]CustodyInterface{1: mockCustody})
		defer cleanup()

		action := &BlockchainAction{Type: ActionTypeCheckpoint, ChainID: 1, Data: []byte{1, 2, 3}, Status: StatusPending}
		require.NoError(t, db.Create(action).Error)

		worker.processAction(context.Background(), *action)

		assert.Equal(t, 0, mockCustody.CallCount())
		var updatedAction BlockchainAction
		require.NoError(t, db.First(&updatedAction, action.ID).Error)
		assert.Equal(t, StatusFailed, updatedAction.Status)
		assert.Contains(t, updatedAction.Error, "unmarshal checkpoint data")
	})

	t.Run("Action fails after 5 attempts", func(t *testing.T) {
		mockCustody := &MockCustody{
			checkpointFn: func() (common.Hash, error) {
				return common.Hash{}, errors.New("RPC still down")
			},
		}
		worker, db, cleanup := setupWorker(t, map[uint32]CustodyInterface{1: mockCustody})
		defer cleanup()

		action := &BlockchainAction{
			Type:    ActionTypeCheckpoint,
			ChainID: 1,
			Data:    validCheckpointData(t),
			Status:  StatusPending,
			Retries: maxActionRetries,
		}
		require.NoError(t, db.Create(action).Error)

		worker.processAction(context.Background(), *action)

		assert.Equal(t, 1, mockCustody.CallCount())
		var updatedAction BlockchainAction
		require.NoError(t, db.First(&updatedAction, action.ID).Error)

		assert.Equal(t, StatusFailed, updatedAction.Status)
		assert.Equal(t, maxActionRetries, updatedAction.Retries)
		assert.Contains(t, updatedAction.Error, fmt.Sprintf("failed after %d retries: RPC still down", maxActionRetries))
	})
}
