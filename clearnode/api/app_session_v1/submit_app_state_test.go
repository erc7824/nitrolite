package app_session_v1

import (
	"context"
	"errors"
	"testing"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSubmitAppState_OperateIntent_NoRedistribution_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigValidator := new(MockSigValidator)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		map[SigType]SigValidator{
			EcdsaSigType: mockSigValidator,
		},
		"0xNode",
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	participant1 := "0x1111111111111111111111111111111111111111"
	participant2 := "0x2222222222222222222222222222222222222222"

	existingSession := &app.AppSessionV1{
		SessionID:   appSessionID,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 5},
			{WalletAddress: participant2, SignatureWeight: 5},
		},
		Quorum:      5,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: `{"state":"initial"}`,
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
		participant2: {
			"USDC": decimal.NewFromInt(50),
		},
	}

	sessionBalances := map[string]decimal.Decimal{
		"USDC": decimal.NewFromInt(150),
	}

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      2,
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "100"},
				{Participant: participant2, Asset: "USDC", Amount: "50"},
			},
			SessionData: `{"state":"updated"}`,
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockStore.On("GetAppSessionBalances", appSessionID).Return(sessionBalances, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.Version == 2 && session.SessionData == `{"state":"updated"}` && session.Status == app.AppSessionStatusOpen
	})).Return(nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitAppState_OperateIntent_WithRedistribution_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigValidator := new(MockSigValidator)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		map[SigType]SigValidator{
			EcdsaSigType: mockSigValidator,
		},
		"0xNode",
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	participant1 := "0x1111111111111111111111111111111111111111"
	participant2 := "0x2222222222222222222222222222222222222222"

	existingSession := &app.AppSessionV1{
		SessionID:   appSessionID,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 5},
			{WalletAddress: participant2, SignatureWeight: 5},
		},
		Quorum:      5,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	// Current allocations: p1=100, p2=50 (total=150)
	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
		participant2: {
			"USDC": decimal.NewFromInt(50),
		},
	}

	sessionBalances := map[string]decimal.Decimal{
		"USDC": decimal.NewFromInt(150),
	}

	// New allocations: p1=75, p2=75 (total=150) - redistribution!
	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      2,
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "75"}, // -25
				{Participant: participant2, Asset: "USDC", Amount: "75"}, // +25
			},
			SessionData: "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockStore.On("GetAppSessionBalances", appSessionID).Return(sessionBalances, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil)
	// Expect ledger entries for the redistribution
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "USDC", decimal.NewFromInt(-25)).Return(nil).Once()
	mockStore.On("RecordLedgerEntry", participant2, appSessionID, "USDC", decimal.NewFromInt(25)).Return(nil).Once()
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.Version == 2 && session.Status == app.AppSessionStatusOpen
	})).Return(nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitAppState_WithdrawIntent_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigValidator := new(MockSigValidator)
	mockSigner := NewMockSigner()

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		map[SigType]SigValidator{
			EcdsaSigType: mockSigValidator,
		},
		"0xNode",
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	participant1 := "0x1111111111111111111111111111111111111111"

	existingSession := &app.AppSessionV1{
		SessionID:   appSessionID,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
	}

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentWithdraw,
			Version:      2,
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "60"}, // Withdraw 40
			},
			SessionData: "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "USDC", decimal.NewFromInt(-40)).Return(nil)

	// Mock expectations for channel state issuance (issueReleaseReceiverState)
	mockStore.On("GetLastUserState", participant1, "USDC", false).Return(nil, nil)
	mockStore.On("GetLastUserState", participant1, "USDC", true).Return(nil, nil)
	mockStatePacker.On("PackState", mock.Anything).Return([]byte("packed"), nil)
	mockStore.On("RecordTransaction", mock.Anything).Return(nil)
	mockStore.On("StoreUserState", mock.Anything).Return(nil)

	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.Version == 2 && session.Status == app.AppSessionStatusOpen
	})).Return(nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitAppState_CloseIntent_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigValidator := new(MockSigValidator)
	mockSigner := NewMockSigner()

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		map[SigType]SigValidator{
			EcdsaSigType: mockSigValidator,
		},
		"0xNode",
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	participant1 := "0x1111111111111111111111111111111111111111"
	participant2 := "0x2222222222222222222222222222222222222222"

	existingSession := &app.AppSessionV1{
		SessionID:   appSessionID,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 5},
			{WalletAddress: participant2, SignatureWeight: 5},
		},
		Quorum:      5,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
		participant2: {
			"USDC": decimal.NewFromInt(50),
		},
	}

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentClose,
			Version:      2,
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "100"},
				{Participant: participant2, Asset: "USDC", Amount: "50"},
			},
			SessionData: "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)

	// Mock expectations for fund release and channel state issuance on close
	// Participant 1: 100 USDC
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "USDC", decimal.NewFromInt(-100)).Return(nil)
	mockStore.On("GetLastUserState", participant1, "USDC", false).Return(nil, nil)
	mockStore.On("GetLastUserState", participant1, "USDC", true).Return(nil, nil)
	mockStatePacker.On("PackState", mock.Anything).Return([]byte("packed"), nil)
	mockStore.On("RecordTransaction", mock.Anything).Return(nil)
	mockStore.On("StoreUserState", mock.Anything).Return(nil).Once()

	// Participant 2: 50 USDC
	mockStore.On("RecordLedgerEntry", participant2, appSessionID, "USDC", decimal.NewFromInt(-50)).Return(nil)
	mockStore.On("GetLastUserState", participant2, "USDC", false).Return(nil, nil)
	mockStore.On("GetLastUserState", participant2, "USDC", true).Return(nil, nil)
	mockStore.On("StoreUserState", mock.Anything).Return(nil).Once()

	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.Version == 2 && session.Status == app.AppSessionStatusClosed
	})).Return(nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitAppState_CloseIntent_AllocationMismatch_Rejected(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigValidator := new(MockSigValidator)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		map[SigType]SigValidator{
			EcdsaSigType: mockSigValidator,
		},
		"0xNode",
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	participant1 := "0x1111111111111111111111111111111111111111"

	existingSession := &app.AppSessionV1{
		SessionID:   appSessionID,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
	}

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentClose,
			Version:      2,
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "50"}, // Mismatch: trying to close with different amount
			},
			SessionData: "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil).Maybe()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert - should fail because allocations don't match current state
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error for close with allocation mismatch")
	assert.Contains(t, respErr.Error(), "close intent requires allocations to match current state")

	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitAppState_OperateIntent_MissingAllocation_Rejected(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigValidator := new(MockSigValidator)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		map[SigType]SigValidator{
			EcdsaSigType: mockSigValidator,
		},
		"0xNode",
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	participant1 := "0x1111111111111111111111111111111111111111"
	participant2 := "0x2222222222222222222222222222222222222222"

	existingSession := &app.AppSessionV1{
		SessionID:   appSessionID,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 5},
			{WalletAddress: participant2, SignatureWeight: 5},
		},
		Quorum:      5,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
		participant2: {
			"USDC": decimal.NewFromInt(50),
		},
	}

	sessionBalances := map[string]decimal.Decimal{
		"USDC": decimal.NewFromInt(150),
	}

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      2,
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "150"}, // Only one participant - missing participant2
			},
			SessionData: "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockStore.On("GetAppSessionBalances", appSessionID).Return(sessionBalances, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil).Maybe()

	// Map iteration order is non-deterministic, so participant1 might be processed before the participant2 missing error
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "USDC", decimal.NewFromInt(50)).Return(nil).Maybe()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert - should fail because participant2 allocation is missing
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error for operate with missing allocation")
	assert.Contains(t, respErr.Error(), "operate intent missing allocation")

	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitAppState_WithdrawIntent_MissingAllocation_Rejected(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigValidator := new(MockSigValidator)
	mockSigner := NewMockSigner()

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		map[SigType]SigValidator{
			EcdsaSigType: mockSigValidator,
		},
		"0xNode",
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	participant1 := "0x1111111111111111111111111111111111111111"

	existingSession := &app.AppSessionV1{
		SessionID:   appSessionID,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
			"DAI":  decimal.NewFromInt(50),
		},
	}

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentWithdraw,
			Version:      2,
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "60"}, // Missing DAI allocation
			},
			SessionData: "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetAssetDecimals", "DAI").Return(uint8(18), nil).Maybe()
	mockStatePacker.On("PackState", mock.Anything).Return([]byte("packed"), nil).Maybe()

	// Map iteration order is non-deterministic, so USDC might be processed before the DAI missing error
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "USDC", decimal.NewFromInt(-40)).Return(nil).Maybe()
	mockStore.On("GetLastUserState", participant1, "USDC", false).Return(nil, nil).Maybe()
	mockStore.On("GetLastUserState", participant1, "USDC", true).Return(nil, nil).Maybe()
	mockStore.On("StoreUserState", mock.Anything).Return(nil).Maybe()
	mockStore.On("RecordTransaction", mock.Anything).Return(nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert - should fail because DAI allocation is missing
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error for withdraw with missing allocation")
	assert.Contains(t, respErr.Error(), "withdraw intent missing allocation")

	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitAppState_DepositIntent_Rejected(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		map[SigType]SigValidator{},
		"0xNode",
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentDeposit,
			Version:      2,
			Allocations:  []rpc.AppAllocationV1{},
			SessionData:  "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "deposit intent must use submit_deposit_state endpoint")
}

func TestSubmitAppState_ClosedSession_Rejected(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigValidator := new(MockSigValidator)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		map[SigType]SigValidator{
			EcdsaSigType: mockSigValidator,
		},
		"0xNode",
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	existingSession := &app.AppSessionV1{
		SessionID: appSessionID,
		Status:    app.AppSessionStatusClosed, // Already closed
		Version:   1,
	}

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      2,
			Allocations:  []rpc.AppAllocationV1{},
			SessionData:  "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "app session is already closed")

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_InvalidVersion_Rejected(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		map[SigType]SigValidator{},
		"0xNode",
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	existingSession := &app.AppSessionV1{
		SessionID: appSessionID,
		Status:    app.AppSessionStatusOpen,
		Version:   5, // Current version is 5
	}

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      10, // Wrong version
			Allocations:  []rpc.AppAllocationV1{},
			SessionData:  "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "invalid app session version")

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_SessionNotFound_Rejected(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		map[SigType]SigValidator{},
		"0xNode",
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      2,
			Allocations:  []rpc.AppAllocationV1{},
			SessionData:  "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations - session not found
	mockStore.On("GetAppSession", appSessionID).Return(nil, errors.New("not found"))

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "app session not found")

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_OperateIntent_InvalidDecimalPrecision_Rejected(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigValidator := new(MockSigValidator)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		map[SigType]SigValidator{
			EcdsaSigType: mockSigValidator,
		},
		"0xNode",
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	participant1 := "0x1111111111111111111111111111111111111111"

	existingSession := &app.AppSessionV1{
		SessionID:   appSessionID,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
	}

	sessionBalances := map[string]decimal.Decimal{
		"USDC": decimal.NewFromInt(100),
	}

	// Create amount with too many decimal places (7 decimals for USDC which has 6)
	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      2,
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "100.1234567"}, // 7 decimal places
			},
			SessionData: "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockStore.On("GetAppSessionBalances", appSessionID).Return(sessionBalances, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert - should fail because of invalid decimal precision
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error for invalid decimal precision")
	assert.Contains(t, respErr.Error(), "invalid amount for allocation with asset USDC")
	assert.Contains(t, respErr.Error(), "amount exceeds maximum decimal precision")

	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitAppState_WithdrawIntent_InvalidDecimalPrecision_Rejected(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigValidator := new(MockSigValidator)
	mockSigner := NewMockSigner()

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		map[SigType]SigValidator{
			EcdsaSigType: mockSigValidator,
		},
		"0xNode",
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	participant1 := "0x1111111111111111111111111111111111111111"

	existingSession := &app.AppSessionV1{
		SessionID:   appSessionID,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
	}

	// Create amount with too many decimal places (7 decimals for USDC which has 6)
	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentWithdraw,
			Version:      2,
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "60.1234567"}, // 7 decimal places, withdrawing 40
			},
			SessionData: "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)
	// RecordLedgerEntry will be called before validation, but then validation will fail
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "USDC", mock.Anything).Return(nil).Maybe()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert - should fail because of invalid decimal precision
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error for invalid decimal precision")
	assert.Contains(t, respErr.Error(), "invalid withdraw amount for allocation with asset USDC")
	assert.Contains(t, respErr.Error(), "amount exceeds maximum decimal precision")

	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}
