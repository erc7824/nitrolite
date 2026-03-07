package evm

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/core"
)

func TestAppRegistryReactor_HandleLocked(t *testing.T) {
	userAddr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	deposited := big.NewInt(500_000_000)    // 500 USDC (6 decimals)
	newBalance := big.NewInt(1_000_000_000) // 1000 USDC
	var tokenDecimals int32 = 6
	blockchainID := uint64(1)

	// ABI-encode the non-indexed parameters: deposited, newBalance
	depositedPadded := common.LeftPadBytes(deposited.Bytes(), 32)
	newBalancePadded := common.LeftPadBytes(newBalance.Bytes(), 32)
	data := append(depositedPadded, newBalancePadded...)

	lockedEventID := lockingContractAbi.Events["Locked"].ID
	logEntry := types.Log{
		Topics: []common.Hash{
			lockedEventID,
			common.BytesToHash(userAddr.Bytes()), // indexed user
		},
		Data:        data,
		BlockNumber: 100,
		TxHash:      common.HexToHash("0xdeadbeef"),
		Index:       0,
	}

	t.Run("success", func(t *testing.T) {
		var capturedEvent *core.UserLockedBalanceUpdatedEvent
		handler := &mockAppRegistryEventHandler{
			handleFn: func(_ context.Context, ev *core.UserLockedBalanceUpdatedEvent) error {
				capturedEvent = ev
				return nil
			},
		}

		var storedEvent core.BlockchainEvent
		storeContractEvent := func(ev core.BlockchainEvent) error {
			storedEvent = ev
			return nil
		}

		reactor, err := NewLockingContractReactor(blockchainID, handler, func() (uint8, error) {
			return uint8(tokenDecimals), nil
		}, storeContractEvent)
		require.NoError(t, err)

		var processedSuccess bool
		reactor.SetOnEventProcessed(func(_ uint64, success bool) {
			processedSuccess = success
		})

		reactor.HandleEvent(context.Background(), logEntry)

		require.NotNil(t, capturedEvent)
		assert.Equal(t, userAddr.String(), common.HexToAddress(capturedEvent.UserAddress).String())
		assert.Equal(t, blockchainID, capturedEvent.BlockchainID)
		expectedBalance := decimal.NewFromBigInt(newBalance, -tokenDecimals)
		assert.True(t, expectedBalance.Equal(capturedEvent.Balance), "expected %s, got %s", expectedBalance, capturedEvent.Balance)

		assert.True(t, processedSuccess)
		assert.Equal(t, "Locked", storedEvent.Name)
		assert.Equal(t, uint64(100), storedEvent.BlockNumber)
		assert.Equal(t, blockchainID, storedEvent.BlockchainID)
	})

	t.Run("getTokenDecimals error", func(t *testing.T) {
		handler := &mockAppRegistryEventHandler{
			handleFn: func(_ context.Context, _ *core.UserLockedBalanceUpdatedEvent) error {
				t.Fatal("handler should not be called")
				return nil
			},
		}

		storeContractEvent := func(_ core.BlockchainEvent) error { return nil }

		_, err := NewLockingContractReactor(blockchainID, handler, func() (uint8, error) {
			return 0, assert.AnError
		}, storeContractEvent)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get token decimals")
	})

	t.Run("handler error", func(t *testing.T) {
		handler := &mockAppRegistryEventHandler{
			handleFn: func(_ context.Context, _ *core.UserLockedBalanceUpdatedEvent) error {
				return assert.AnError
			},
		}

		storeContractEvent := func(_ core.BlockchainEvent) error { return nil }

		reactor, err := NewLockingContractReactor(blockchainID, handler, func() (uint8, error) {
			return uint8(tokenDecimals), nil
		}, storeContractEvent)
		require.NoError(t, err)

		var processedSuccess bool
		reactor.SetOnEventProcessed(func(_ uint64, success bool) {
			processedSuccess = success
		})

		reactor.HandleEvent(context.Background(), logEntry)
		assert.False(t, processedSuccess)
	})
}

type mockAppRegistryEventHandler struct {
	handleFn func(context.Context, *core.UserLockedBalanceUpdatedEvent) error
}

func (m *mockAppRegistryEventHandler) HandleUserLockedBalanceUpdated(ctx context.Context, ev *core.UserLockedBalanceUpdatedEvent) error {
	return m.handleFn(ctx, ev)
}
