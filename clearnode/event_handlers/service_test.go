package event_handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/log"
)

func TestHandleHomeChannelCreated_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	logger := log.NewNoopLogger()

	service := &EventHandlerService{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		logger: logger,
	}

	// Test data
	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusVoid,
		StateVersion: 0,
	}

	event := &core.HomeChannelCreatedEvent{
		ChannelID:    channelID,
		StateVersion: 1,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusOpen &&
			ch.StateVersion == 1
	})).Return(nil)

	// Execute
	err := service.HandleHomeChannelCreated(event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleHomeChannelCheckpointed_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	logger := log.NewNoopLogger()

	service := &EventHandlerService{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		logger: logger,
	}

	// Test data
	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	expiryTime := time.Now().Add(time.Hour)

	channel := &core.Channel{
		ChannelID:          channelID,
		UserWallet:         userWallet,
		Type:               core.ChannelTypeHome,
		Status:             core.ChannelStatusChallenged,
		StateVersion:       3,
		ChallengeExpiresAt: &expiryTime,
	}

	event := &core.HomeChannelCheckpointedEvent{
		ChannelID:    channelID,
		StateVersion: 5,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusOpen &&
			ch.StateVersion == 5
	})).Return(nil)

	// Execute
	err := service.HandleHomeChannelCheckpointed(event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleHomeChannelChallenged_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	logger := log.NewNoopLogger()

	service := &EventHandlerService{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		logger: logger,
	}

	// Test data
	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusOpen,
		StateVersion: 3,
	}

	state := &core.State{
		ID:      "state123",
		Version: 6,
	}

	event := &core.HomeChannelChallengedEvent{
		ChannelID:       channelID,
		StateVersion:    4,
		ChallengeExpiry: challengeExpiry,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusChallenged &&
			ch.StateVersion == 4 &&
			ch.ChallengeExpiresAt != nil
	})).Return(nil)
	mockStore.On("GetLastStateByChannelID", channelID, true).Return(state, nil)
	mockStore.On("ScheduleCheckpoint", "state123").Return(nil)

	// Execute
	err := service.HandleHomeChannelChallenged(event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleHomeChannelClosed_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	logger := log.NewNoopLogger()

	service := &EventHandlerService{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		logger: logger,
	}

	// Test data
	channelID := "0xHomeChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Type:         core.ChannelTypeHome,
		Status:       core.ChannelStatusOpen,
		StateVersion: 5,
	}

	event := &core.HomeChannelClosedEvent{
		ChannelID:    channelID,
		StateVersion: 10,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusClosed &&
			ch.StateVersion == 10
	})).Return(nil)

	// Execute
	err := service.HandleHomeChannelClosed(event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowDepositInitiated_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	logger := log.NewNoopLogger()

	service := &EventHandlerService{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		logger: logger,
	}

	// Test data
	channelID := "0xEscrowChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusVoid,
		StateVersion: 0,
	}

	state := &core.State{
		ID:      "state123",
		Version: 1,
	}

	event := &core.EscrowDepositInitiatedEvent{
		ChannelID:    channelID,
		StateVersion: 1,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusOpen &&
			ch.StateVersion == 1
	})).Return(nil)
	mockStore.On("GetStateByChannelIDAndVersion", channelID, uint64(1)).Return(state, nil)
	mockStore.On("ScheduleCheckpoint", "state123").Return(nil)

	// Execute
	err := service.HandleEscrowDepositInitiated(event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowDepositChallenged_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	logger := log.NewNoopLogger()

	service := &EventHandlerService{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		logger: logger,
	}

	// Test data
	channelID := "0xEscrowChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 1,
	}

	state := &core.State{
		ID:      "state123",
		Version: 5,
	}

	event := &core.EscrowDepositChallengedEvent{
		ChannelID:       channelID,
		StateVersion:    3,
		ChallengeExpiry: challengeExpiry,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusChallenged &&
			ch.StateVersion == 3 &&
			ch.ChallengeExpiresAt != nil
	})).Return(nil)
	mockStore.On("GetLastStateByChannelID", channelID, true).Return(state, nil)
	mockStore.On("ScheduleFinalizeEscrowDeposit", "state123").Return(nil)

	// Execute
	err := service.HandleEscrowDepositChallenged(event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowDepositFinalized_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	logger := log.NewNoopLogger()

	service := &EventHandlerService{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		logger: logger,
	}

	// Test data
	channelID := "0xEscrowChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 3,
	}

	event := &core.EscrowDepositFinalizedEvent{
		ChannelID:    channelID,
		StateVersion: 5,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusClosed &&
			ch.StateVersion == 5
	})).Return(nil)

	// Execute
	err := service.HandleEscrowDepositFinalized(event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowWithdrawalInitiated_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	logger := log.NewNoopLogger()

	service := &EventHandlerService{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		logger: logger,
	}

	// Test data
	channelID := "0xEscrowChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusVoid,
		StateVersion: 0,
	}

	event := &core.EscrowWithdrawalInitiatedEvent{
		ChannelID:    channelID,
		StateVersion: 1,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusOpen &&
			ch.StateVersion == 1
	})).Return(nil)

	// Execute
	err := service.HandleEscrowWithdrawalInitiated(event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowWithdrawalChallenged_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	logger := log.NewNoopLogger()

	service := &EventHandlerService{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		logger: logger,
	}

	// Test data
	channelID := "0xEscrowChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"
	challengeExpiry := uint64(time.Now().Add(time.Hour).Unix())

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 1,
	}

	state := &core.State{
		ID:      "state123",
		Version: 5,
	}

	event := &core.EscrowWithdrawalChallengedEvent{
		ChannelID:       channelID,
		StateVersion:    3,
		ChallengeExpiry: challengeExpiry,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusChallenged &&
			ch.StateVersion == 3 &&
			ch.ChallengeExpiresAt != nil
	})).Return(nil)
	mockStore.On("GetLastStateByChannelID", channelID, true).Return(state, nil)
	mockStore.On("ScheduleFinalizeEscrowWithdrawal", "state123").Return(nil)

	// Execute
	err := service.HandleEscrowWithdrawalChallenged(event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestHandleEscrowWithdrawalFinalized_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	logger := log.NewNoopLogger()

	service := &EventHandlerService{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		logger: logger,
	}

	// Test data
	channelID := "0xEscrowChannel123"
	userWallet := "0x1234567890123456789012345678901234567890"

	channel := &core.Channel{
		ChannelID:    channelID,
		UserWallet:   userWallet,
		Type:         core.ChannelTypeEscrow,
		Status:       core.ChannelStatusOpen,
		StateVersion: 3,
	}

	event := &core.EscrowWithdrawalFinalizedEvent{
		ChannelID:    channelID,
		StateVersion: 5,
	}

	// Mock expectations
	mockStore.On("GetChannelByID", channelID).Return(channel, nil)
	mockStore.On("UpdateChannel", mock.MatchedBy(func(ch core.Channel) bool {
		return ch.ChannelID == channelID &&
			ch.Status == core.ChannelStatusClosed &&
			ch.StateVersion == 5
	})).Return(nil)

	// Execute
	err := service.HandleEscrowWithdrawalFinalized(event)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}
