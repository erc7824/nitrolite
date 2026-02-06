package app_session_v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

func TestCreateAppSession_Success(t *testing.T) {
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
		"0xnode",
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"
	participant2 := "0x2222222222222222222222222222222222222222"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
				{
					WalletAddress:   participant2,
					SignatureWeight: 1,
				},
			},
			Quorum: 1, // Only need 1 signature
			Nonce:  "12345",
		},
		QuorumSigs: []string{
			"0x1234567890abcdef", // Mock signature from participant1
		},
		SessionData: `{"test": "data"}`,
	}

	// Mock expectations
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil).Once()
	mockStore.On("CreateAppSession", mock.MatchedBy(func(session any) bool {
		return true // Accept any app session for now
	})).Return(nil).Once()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	// Execute
	handler.CreateAppSession(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Check for errors first
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}

	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	// Parse the response
	var resp rpc.AppSessionsV1CreateAppSessionResponse
	err = ctx.Response.Payload.Translate(&resp)
	require.NoError(t, err)

	// Verify response fields
	assert.NotEmpty(t, resp.AppSessionID)
	assert.Equal(t, "1", resp.Version)
	assert.Equal(t, app.AppSessionStatusOpen.String(), resp.Status)

	// Verify all mocks were called
	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestCreateAppSession_QuorumWithMultipleSignatures(t *testing.T) {
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
		"0xnode",
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"
	participant2 := "0x2222222222222222222222222222222222222222"
	participant3 := "0x3333333333333333333333333333333333333333"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 2, // Weight of 2
				},
				{
					WalletAddress:   participant2,
					SignatureWeight: 1, // Weight of 1
				},
				{
					WalletAddress:   participant3,
					SignatureWeight: 1, // Weight of 1
				},
			},
			Quorum: 3, // Need total weight of 3
			Nonce:  "12345",
		},
		QuorumSigs: []string{
			"0x1234", // participant1 (weight 2)
			"0x5678", // participant2 (weight 1)
		},
		SessionData: "",
	}

	// Mock expectations - signatures will be recovered in order
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil).Once()
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant2, nil).Once()
	mockStore.On("CreateAppSession", mock.Anything).Return(nil).Once()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	// Execute
	handler.CreateAppSession(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Check for errors first
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}

	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	// Verify all mocks were called
	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestCreateAppSession_ZeroNonce(t *testing.T) {
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
		"0xnode",
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
			},
			Quorum: 1,
			Nonce:  "0", // Zero nonce - invalid
		},
		QuorumSigs: []string{"0x1234567890abcdef"},
	}

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	// Execute
	handler.CreateAppSession(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Verify response contains error about nonce
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonce")

	// Verify no mocks were called since we fail early
	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestCreateAppSession_QuorumExceedsTotalWeights(t *testing.T) {
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
		"0xnode",
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"
	participant2 := "0x2222222222222222222222222222222222222222"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
				{
					WalletAddress:   participant2,
					SignatureWeight: 1,
				},
			},
			Quorum: 5, // Total weights = 2, but quorum = 5
			Nonce:  "12345",
		},
		QuorumSigs: []string{"0x1234567890abcdef"},
	}

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	// Execute
	handler.CreateAppSession(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Verify response contains error about quorum
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quorum")
	assert.Contains(t, err.Error(), "weights")

	// Verify no mocks were called since we fail early
	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestCreateAppSession_NoSignatures(t *testing.T) {
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
		"0xnode",
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
			},
			Quorum: 1,
			Nonce:  "12345",
		},
		QuorumSigs: []string{}, // Empty signatures
	}

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	// Execute
	handler.CreateAppSession(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Verify response contains error about signatures
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no signatures")

	// Verify no mocks were called since we fail early
	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestCreateAppSession_SignatureFromNonParticipant(t *testing.T) {
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
		"0xnode",
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"
	nonParticipant := "0x9999999999999999999999999999999999999999"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
			},
			Quorum: 1,
			Nonce:  "12345",
		},
		QuorumSigs: []string{"0x1234567890abcdef"},
	}

	// Mock expectations - signature recovered from non-participant
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(nonParticipant, nil).Once()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	// Execute
	handler.CreateAppSession(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Verify response contains error about non-participant
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-participant")

	// Verify mocks were called
	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestCreateAppSession_QuorumNotMet(t *testing.T) {
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
		"0xnode",
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"
	participant2 := "0x2222222222222222222222222222222222222222"
	participant3 := "0x3333333333333333333333333333333333333333"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
				{
					WalletAddress:   participant2,
					SignatureWeight: 1,
				},
				{
					WalletAddress:   participant3,
					SignatureWeight: 1,
				},
			},
			Quorum: 3, // Need all 3
			Nonce:  "12345",
		},
		QuorumSigs: []string{
			"0x1234", // Only one signature, need 3 total weight
		},
	}

	// Mock expectations - only one signature provided
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil).Once()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	// Execute
	handler.CreateAppSession(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Verify response contains error about quorum not met
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quorum not met")

	// Verify mocks were called
	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestCreateAppSession_DuplicateSignatures(t *testing.T) {
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
		"0xnode",
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"
	participant2 := "0x2222222222222222222222222222222222222222"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
				{
					WalletAddress:   participant2,
					SignatureWeight: 1,
				},
			},
			Quorum: 2, // Need both participants
			Nonce:  "12345",
		},
		QuorumSigs: []string{
			"0x1234", // participant1
			"0x5678", // participant1 again (duplicate)
		},
	}

	// Mock expectations - both signatures from same participant
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil).Once()
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return(participant1, nil).Once()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	// Execute
	handler.CreateAppSession(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Verify response contains error - duplicate signatures shouldn't count twice
	// Should fail with "quorum not met" since only 1 weight achieved, need 2
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quorum not met")

	// Verify mocks were called
	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestCreateAppSession_InvalidSignatureHex(t *testing.T) {
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
		"0xnode",
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
			},
			Quorum: 1,
			Nonce:  "12345",
		},
		QuorumSigs: []string{"not-valid-hex"}, // Invalid hex string
	}

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	// Execute
	handler.CreateAppSession(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Verify response contains error about signature decoding
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode signature")

	// Verify no mocks were called since we fail at signature decoding
	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestCreateAppSession_SignatureRecoveryFailure(t *testing.T) {
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
		"0xnode",
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
			},
			Quorum: 1,
			Nonce:  "12345",
		},
		QuorumSigs: []string{"0x1234567890abcdef"},
	}

	// Mock expectations - signature recovery fails
	mockSigValidator.On("Recover", mock.Anything, mock.Anything).Return("", assert.AnError).Once()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	// Execute
	handler.CreateAppSession(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Verify response contains error about signature recovery
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "recover signer address")

	// Verify mocks were called
	mockStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}
