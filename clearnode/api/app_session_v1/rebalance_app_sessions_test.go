package app_session_v1

import (
	"context"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

var (
	// Valid ECDSA signatures (65 bytes = 130 hex chars + 0x prefix)
	validSig1 = "0x" + strings.Repeat("a", 130)
	validSig2 = "0x" + strings.Repeat("b", 130)
)

// assertSuccess checks if the RPC context has a successful response
func assertSuccess(t *testing.T, ctx *rpc.Context) {
	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)
}

// assertError checks if the RPC context has an error response with the expected message
func assertError(t *testing.T, ctx *rpc.Context, expectedMessage string) {
	require.NotNil(t, ctx.Response)
	err := ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedMessage)
}

func TestRebalanceAppSessions_Success_TwoSessions(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigValidator := new(MockSigValidator)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		nil,
		map[SigType]SigValidator{
			EcdsaSigType: mockSigValidator,
		},
		"0xNode",
	)

	sessionID1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID2 := "0x2222222222222222222222222222222222222222222222222222222222222222"
	participant1 := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	participant2 := "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	session1 := &app.AppSessionV1{
		SessionID:   sessionID1,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     5,
		SessionData: `{"data":"session1"}`,
	}

	session2 := &app.AppSessionV1{
		SessionID:   sessionID2,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant2, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     3,
		SessionData: `{"data":"session2"}`,
	}

	// Session 1: currently has 200 USDC, will have 100 USDC (loses 100)
	currentAllocations1 := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(200),
		},
	}

	// Session 2: currently has 50 USDC, will have 150 USDC (gains 100)
	currentAllocations2 := map[string]map[string]decimal.Decimal{
		participant2: {
			"USDC": decimal.NewFromInt(50),
		},
	}

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID1,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      6,
					Allocations: []rpc.AppAllocationV1{
						{Participant: participant1, Asset: "USDC", Amount: "100"},
					},
					SessionData: `{"data":"session1_updated"}`,
				},
				QuorumSigs: []string{validSig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID2,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      4,
					Allocations: []rpc.AppAllocationV1{
						{Participant: participant2, Asset: "USDC", Amount: "150"},
					},
					SessionData: `{"data":"session2_updated"}`,
				},
				QuorumSigs: []string{validSig2},
			},
		},
	}

	// Mock expectations for session 1
	mockStore.On("GetAppSession", sessionID1).Return(session1, nil)
	mockStore.On("GetParticipantAllocations", sessionID1).Return(currentAllocations1, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil).Once()
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.SessionID == sessionID1 &&
			session.Version == 6 &&
			session.SessionData == `{"data":"session1_updated"}`
	})).Return(nil).Once()

	// Mock expectations for session 2
	mockStore.On("GetAppSession", sessionID2).Return(session2, nil)
	mockStore.On("GetParticipantAllocations", sessionID2).Return(currentAllocations2, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant2, nil).Once()
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.SessionID == sessionID2 &&
			session.Version == 4 &&
			session.SessionData == `{"data":"session2_updated"}`
	})).Return(nil).Once()

	// Mock ledger entry and transaction recording
	mockStore.On("RecordLedgerEntry", sessionID1, "USDC", decimal.NewFromInt(-100), (*string)(nil)).Return(nil)
	mockStore.On("RecordLedgerEntry", sessionID2, "USDC", decimal.NewFromInt(100), (*string)(nil)).Return(nil)
	mockStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeRebalance && tx.Asset == "USDC"
	})).Return(nil).Twice()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	assertSuccess(t, ctx)
	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)

	// Verify response contains batch_id
	var response rpc.AppSessionsV1RebalanceAppSessionsResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.BatchID)
}

func TestRebalanceAppSessions_Success_MultiAsset(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigValidator := new(MockSigValidator)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		nil,
		map[SigType]SigValidator{
			EcdsaSigType: mockSigValidator,
		},
		"0xNode",
	)

	sessionID1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID2 := "0x2222222222222222222222222222222222222222222222222222222222222222"
	participant1 := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	participant2 := "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	session1 := &app.AppSessionV1{
		SessionID:   sessionID1,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:  10,
		Status:  app.AppSessionStatusOpen,
		Version: 1,
	}

	session2 := &app.AppSessionV1{
		SessionID:   sessionID2,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant2, SignatureWeight: 10},
		},
		Quorum:  10,
		Status:  app.AppSessionStatusOpen,
		Version: 1,
	}

	// Session 1: 200 USDC, 1 ETH -> 100 USDC, 1.5 ETH (loses 100 USDC, gains 0.5 ETH)
	currentAllocations1 := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(200),
			"ETH":  decimal.NewFromInt(1),
		},
	}

	// Session 2: 50 USDC, 2 ETH -> 150 USDC, 1.5 ETH (gains 100 USDC, loses 0.5 ETH)
	currentAllocations2 := map[string]map[string]decimal.Decimal{
		participant2: {
			"USDC": decimal.NewFromInt(50),
			"ETH":  decimal.NewFromInt(2),
		},
	}

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID1,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      2,
					Allocations: []rpc.AppAllocationV1{
						{Participant: participant1, Asset: "USDC", Amount: "100"},
						{Participant: participant1, Asset: "ETH", Amount: "1.5"},
					},
				},
				QuorumSigs: []string{validSig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID2,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      2,
					Allocations: []rpc.AppAllocationV1{
						{Participant: participant2, Asset: "USDC", Amount: "150"},
						{Participant: participant2, Asset: "ETH", Amount: "1.5"},
					},
				},
				QuorumSigs: []string{validSig2},
			},
		},
	}

	// Mock expectations
	mockStore.On("GetAppSession", sessionID1).Return(session1, nil)
	mockStore.On("GetParticipantAllocations", sessionID1).Return(currentAllocations1, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil).Once()
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(s app.AppSessionV1) bool {
		return s.SessionID == sessionID1 && s.Version == 2
	})).Return(nil).Once()

	mockStore.On("GetAppSession", sessionID2).Return(session2, nil)
	mockStore.On("GetParticipantAllocations", sessionID2).Return(currentAllocations2, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant2, nil).Once()
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(s app.AppSessionV1) bool {
		return s.SessionID == sessionID2 && s.Version == 2
	})).Return(nil).Once()

	// Ledger entries
	mockStore.On("RecordLedgerEntry", sessionID1, "USDC", decimal.NewFromInt(-100), (*string)(nil)).Return(nil)
	mockStore.On("RecordLedgerEntry", sessionID1, "ETH", decimal.RequireFromString("0.5"), (*string)(nil)).Return(nil)
	mockStore.On("RecordLedgerEntry", sessionID2, "USDC", decimal.NewFromInt(100), (*string)(nil)).Return(nil)
	mockStore.On("RecordLedgerEntry", sessionID2, "ETH", decimal.RequireFromString("-0.5"), (*string)(nil)).Return(nil)
	mockStore.On("RecordTransaction", mock.Anything).Return(nil).Times(4) // 2 assets × 2 sessions

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	assertSuccess(t, ctx)
	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestRebalanceAppSessions_Error_InsufficientSessions(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		nil,
		nil,
		"0xNode",
	)

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: "0x1111111111111111111111111111111111111111111111111111111111111111",
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      2,
				},
				QuorumSigs: []string{validSig1},
			},
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "rebalancing requires at least 2 sessions")
}

func TestRebalanceAppSessions_Error_InvalidIntent(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		nil,
		nil,
		"0xNode",
	)

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: "0x1111111111111111111111111111111111111111111111111111111111111111",
					Intent:       app.AppStateUpdateIntentOperate, // Wrong intent
					Version:      2,
				},
				QuorumSigs: []string{validSig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: "0x2222222222222222222222222222222222222222222222222222222222222222",
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      2,
				},
				QuorumSigs: []string{validSig2},
			},
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "all updates must have 'rebalance' intent")
}

func TestRebalanceAppSessions_Error_DuplicateSession(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		nil,
		nil,
		"0xNode",
	)

	sessionID := "0x1111111111111111111111111111111111111111111111111111111111111111"

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      2,
				},
				QuorumSigs: []string{validSig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID, // Duplicate
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      3,
				},
				QuorumSigs: []string{validSig2},
			},
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "duplicate session in rebalance")
}

func TestRebalanceAppSessions_Error_ConservationViolation(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigValidator := new(MockSigValidator)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		nil,
		map[SigType]SigValidator{
			EcdsaSigType: mockSigValidator,
		},
		"0xNode",
	)

	sessionID1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID2 := "0x2222222222222222222222222222222222222222222222222222222222222222"
	participant1 := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	participant2 := "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	session1 := &app.AppSessionV1{
		SessionID:   sessionID1,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:  10,
		Status:  app.AppSessionStatusOpen,
		Version: 1,
	}

	session2 := &app.AppSessionV1{
		SessionID:   sessionID2,
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant2, SignatureWeight: 10},
		},
		Quorum:  10,
		Status:  app.AppSessionStatusOpen,
		Version: 1,
	}

	currentAllocations1 := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(200),
		},
	}

	currentAllocations2 := map[string]map[string]decimal.Decimal{
		participant2: {
			"USDC": decimal.NewFromInt(50),
		},
	}

	// Session 1 loses 100 USDC, Session 2 gains 200 USDC (not conserved!)
	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID1,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      2,
					Allocations: []rpc.AppAllocationV1{
						{Participant: participant1, Asset: "USDC", Amount: "100"},
					},
				},
				QuorumSigs: []string{validSig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID2,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      2,
					Allocations: []rpc.AppAllocationV1{
						{Participant: participant2, Asset: "USDC", Amount: "250"}, // Conservation violation
					},
				},
				QuorumSigs: []string{validSig2},
			},
		},
	}

	// Mock expectations
	mockStore.On("GetAppSession", sessionID1).Return(session1, nil)
	mockStore.On("GetParticipantAllocations", sessionID1).Return(currentAllocations1, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil).Once()
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(s app.AppSessionV1) bool {
		return s.SessionID == sessionID1 && s.Version == 2
	})).Return(nil).Once()

	mockStore.On("GetAppSession", sessionID2).Return(session2, nil)
	mockStore.On("GetParticipantAllocations", sessionID2).Return(currentAllocations2, nil)
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant2, nil).Once()
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(s app.AppSessionV1) bool {
		return s.SessionID == sessionID2 && s.Version == 2
	})).Return(nil).Once()

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "conservation violation")
	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestRebalanceAppSessions_Error_SessionNotFound(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		nil,
		map[SigType]SigValidator{
			EcdsaSigType: new(MockSigValidator),
		},
		"0xNode",
	)

	sessionID1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID2 := "0x2222222222222222222222222222222222222222222222222222222222222222"

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID1,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      2,
				},
				QuorumSigs: []string{validSig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID2,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      2,
				},
				QuorumSigs: []string{validSig2},
			},
		},
	}

	// Mock first session returns nil (not found)
	mockStore.On("GetAppSession", sessionID1).Return(nil, nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "app session not found")
	mockStore.AssertExpectations(t)
}

func TestRebalanceAppSessions_Error_ClosedSession(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		nil,
		map[SigType]SigValidator{
			EcdsaSigType: new(MockSigValidator),
		},
		"0xNode",
	)

	sessionID1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID2 := "0x2222222222222222222222222222222222222222222222222222222222222222"

	session1 := &app.AppSessionV1{
		SessionID: sessionID1,
		Status:    app.AppSessionStatusClosed, // Closed
		Version:   1,
	}

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID1,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      2,
				},
				QuorumSigs: []string{validSig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID2,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      2,
				},
				QuorumSigs: []string{validSig2},
			},
		},
	}

	mockStore.On("GetAppSession", sessionID1).Return(session1, nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "already closed")
	mockStore.AssertExpectations(t)
}

func TestRebalanceAppSessions_Error_InvalidVersion(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		nil,
		map[SigType]SigValidator{
			EcdsaSigType: new(MockSigValidator),
		},
		"0xNode",
	)

	sessionID1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID2 := "0x2222222222222222222222222222222222222222222222222222222222222222"

	session1 := &app.AppSessionV1{
		SessionID: sessionID1,
		Status:    app.AppSessionStatusOpen,
		Version:   5, // Current version is 5
	}

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID1,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      10, // Wrong version (should be 6)
				},
				QuorumSigs: []string{validSig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID2,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      2,
				},
				QuorumSigs: []string{validSig2},
			},
		},
	}

	mockStore.On("GetAppSession", sessionID1).Return(session1, nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "invalid version")
	mockStore.AssertExpectations(t)
}
