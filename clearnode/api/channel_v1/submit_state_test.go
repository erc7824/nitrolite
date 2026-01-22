package channel_v1

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

func TestSubmitState_TransferSend_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	mockSigValidator := new(MockSigValidator)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint64(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:  mockMemoryStore,
		signer:       mockSigner,
		nodeAddress:  nodeAddress,
		minChallenge: minChallenge,
		sigValidators: map[SigValidatorType]SigValidator{
			EcdsaSigValidatorType: mockSigValidator,
		},
	}

	// Test data
	senderWallet := "0x1234567890123456789012345678901234567890"
	receiverWallet := "0x0987654321098765432109876543210987654321"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	transferAmount := decimal.NewFromInt(100)

	// Create sender's current state (before transfer)
	currentSenderState := core.State{
		ID:            core.GetStateID(senderWallet, asset, 1, 1),
		Transitions:   []core.Transition{},
		Asset:         asset,
		UserWallet:    senderWallet,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(500),
			UserNetFlow:  decimal.NewFromInt(500),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: nil,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Create incoming sender state (with transfer send transition)
	incomingSenderState := currentSenderState.NextState()

	// Apply the transfer send transition to update balances
	transferSendTransition, err := incomingSenderState.ApplyTransferSendTransition(receiverWallet, transferAmount)
	require.NoError(t, err)

	// Sign the incoming sender state with user's signature
	userKey, _ := crypto.GenerateKey()
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil).Once()
	packedSenderState, _ := core.PackState(*incomingSenderState, mockAssetStore)
	userSigBytes, _ := crypto.Sign(crypto.Keccak256Hash(packedSenderState).Bytes(), userKey)
	userSigHex := hexutil.Encode(userSigBytes)
	incomingSenderState.UserSig = &userSigHex

	// Create receiver's current state
	currentReceiverState := core.State{
		ID:            core.GetStateID(receiverWallet, asset, 1, 1),
		Transitions:   []core.Transition{},
		Asset:         asset,
		UserWallet:    receiverWallet,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(200),
			UserNetFlow:  decimal.NewFromInt(200),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: nil,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Expected receiver state after transfer receive
	expectedReceiverState := currentReceiverState.NextState()
	_, err = expectedReceiverState.ApplyTransferReceiveTransition(senderWallet, transferAmount, transferSendTransition.TxID)
	require.NoError(t, err)

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("CheckOpenChannel", senderWallet, asset).Return(true, nil)
	mockTxStore.On("GetLastUserState", senderWallet, asset, false).Return(currentSenderState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", senderWallet, asset).Return(nil)
	mockSigValidator.On("Verify", senderWallet, mock.Anything, userSigBytes).Return(nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedSenderState, nil).Maybe()

	// For issueTransferReceiverState
	mockTxStore.On("GetLastUserState", receiverWallet, asset, false).Return(currentReceiverState, nil)
	mockTxStore.On("GetLastUserState", receiverWallet, asset, true).Return(nil, nil)
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		// Verify receiver state
		return state.UserWallet == receiverWallet &&
			state.Version == expectedReceiverState.Version &&
			len(state.Transitions) == 1 &&
			state.Transitions[0].Type == core.TransitionTypeTransferReceive &&
			state.NodeSig != nil
	})).Return(nil)

	// For recordTransaction
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeTransfer &&
			tx.Amount.Equal(transferAmount) &&
			tx.FromAccount == senderWallet &&
			tx.ToAccount == receiverWallet
	})).Return(nil)

	// For storing sender state
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		// Verify sender state
		return state.UserWallet == senderWallet &&
			state.Version == incomingSenderState.Version &&
			len(state.Transitions) == 1 &&
			state.Transitions[0].Type == core.TransitionTypeTransferSend &&
			state.NodeSig != nil
	})).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingSenderState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitState_EscrowLock_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	mockSigValidator := new(MockSigValidator)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint64(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:  mockMemoryStore,
		signer:       mockSigner,
		nodeAddress:  nodeAddress,
		minChallenge: minChallenge,
		sigValidators: map[SigValidatorType]SigValidator{
			EcdsaSigValidatorType: mockSigValidator,
		},
	}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	lockAmount := decimal.NewFromInt(100)
	nonce := uint64(12345)
	challenge := uint64(86400)

	// Create user's current state (with existing home channel)
	currentState := core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 1),
		Transitions:   []core.Transition{},
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(500),
			UserNetFlow:  decimal.NewFromInt(500),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: nil,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Create incoming state with escrow lock transition
	incomingState := currentState.NextState()

	// Apply the escrow lock transition to update balances
	_, err := incomingState.ApplyEscrowLockTransition(2, "0xTokenAddress", lockAmount)
	require.NoError(t, err)
	escrowChannelID := *incomingState.EscrowChannelID

	// Sign the incoming state with user's signature
	userKey, _ := crypto.GenerateKey()
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSigBytes, _ := crypto.Sign(crypto.Keccak256Hash(packedState).Bytes(), userKey)
	userSigHex := hexutil.Encode(userSigBytes)
	incomingState.UserSig = &userSigHex

	// Create home channel for mocking
	homeChannel := core.Channel{
		ChannelID:         homeChannelID,
		UserWallet:        userWallet,
		Type:              core.ChannelTypeHome,
		BlockchainID:      1,
		TokenAddress:      "0xTokenAddress",
		ChallengeDuration: challenge,
		Nonce:             nonce,
		Status:            core.ChannelStatusOpen,
		StateVersion:      1,
	}

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return(true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", userWallet, asset).Return(nil)
	mockSigValidator.On("Verify", userWallet, packedState, userSigBytes).Return(nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("GetChannelByID", homeChannelID).Return(&homeChannel, nil)
	mockMemoryStore.On("IsAssetSupported", asset, "0xTokenAddress", uint32(2)).Return(true, nil)
	mockTxStore.On("CreateChannel", mock.MatchedBy(func(channel core.Channel) bool {
		return channel.ChannelID == escrowChannelID &&
			channel.Type == core.ChannelTypeEscrow &&
			channel.UserWallet == userWallet
	})).Return(nil)
	mockTxStore.On("ScheduleInitiateEscrowWithdrawal", mock.MatchedBy(func(stateID string) bool {
		return stateID == incomingState.ID
	})).Return(nil)
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeEscrowLock &&
			tx.FromAccount == homeChannelID &&
			tx.ToAccount == escrowChannelID &&
			tx.Amount.Equal(lockAmount)
	})).Return(nil)
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&
			len(state.Transitions) == 1 &&
			state.Transitions[0].Type == core.TransitionTypeEscrowLock &&
			state.NodeSig != nil
	})).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitState_EscrowWithdraw_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	mockSigValidator := new(MockSigValidator)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint64(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:  mockMemoryStore,
		signer:       mockSigner,
		nodeAddress:  nodeAddress,
		minChallenge: minChallenge,
		sigValidators: map[SigValidatorType]SigValidator{
			EcdsaSigValidatorType: mockSigValidator,
		},
	}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	escrowChannelID := "0xEscrowChannel456"
	withdrawAmount := decimal.NewFromInt(100)

	// Create user's current state (signed, with escrow ledger)
	// The last transition must be an EscrowLock for the EscrowWithdraw to be valid
	currentSignedState := core.State{
		ID: core.GetStateID(userWallet, asset, 1, 2),
		Transitions: []core.Transition{
			{
				Type:      core.TransitionTypeEscrowLock,
				TxID:      "0xPreviousEscrowLockTx",
				AccountID: "",
				Amount:    withdrawAmount,
			},
		},
		Asset:           asset,
		UserWallet:      userWallet,
		Epoch:           1,
		Version:         2,
		HomeChannelID:   &homeChannelID,
		EscrowChannelID: &escrowChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(400),
			UserNetFlow:  decimal.NewFromInt(400),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: &core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 2,
			UserBalance:  decimal.NewFromInt(0),
			UserNetFlow:  decimal.NewFromInt(0),
			NodeBalance:  decimal.NewFromInt(100),
			NodeNetFlow:  decimal.NewFromInt(100),
		},
		UserSig: stringPtr("0xPreviousUserSig"),
		NodeSig: stringPtr("0xPreviousNodeSig"),
	}

	// Create incoming state with escrow withdraw transition
	incomingState := currentSignedState.NextState()

	// Apply the escrow withdraw transition to update balances
	_, err := incomingState.ApplyEscrowWithdrawTransition(withdrawAmount)
	require.NoError(t, err)

	// Sign the incoming state with user's signature
	userKey, _ := crypto.GenerateKey()
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSigBytes, _ := crypto.Sign(crypto.Keccak256Hash(packedState).Bytes(), userKey)
	userSigHex := hexutil.Encode(userSigBytes)
	incomingState.UserSig = &userSigHex

	// Create a copy for the unsigned state mock
	currentUnsignedState := currentSignedState

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return(true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentUnsignedState, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, true).Return(currentSignedState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", userWallet, asset).Return(nil)
	mockSigValidator.On("Verify", userWallet, packedState, userSigBytes).Return(nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)

	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeEscrowWithdraw &&
			tx.FromAccount == homeChannelID &&
			tx.ToAccount == escrowChannelID &&
			tx.Amount.Equal(withdrawAmount)
	})).Return(nil)

	// Store incoming state with node signature
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&
			len(state.Transitions) == 1 &&
			state.Transitions[0].Type == core.TransitionTypeEscrowWithdraw &&
			state.NodeSig != nil
	})).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitState_HomeDeposit_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	mockSigValidator := new(MockSigValidator)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint64(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:  mockMemoryStore,
		signer:       mockSigner,
		nodeAddress:  nodeAddress,
		minChallenge: minChallenge,
		sigValidators: map[SigValidatorType]SigValidator{
			EcdsaSigValidatorType: mockSigValidator,
		},
	}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	depositAmount := decimal.NewFromInt(100)

	// Create user's current state (with existing home channel)
	currentState := core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 1),
		Transitions:   []core.Transition{},
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(500),
			UserNetFlow:  decimal.NewFromInt(0),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(500),
		},
		EscrowLedger: nil,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Create incoming state with home deposit transition
	incomingState := currentState.NextState()

	// Apply the home deposit transition to update balances
	_, err := incomingState.ApplyHomeDepositTransition(depositAmount)
	require.NoError(t, err)

	// Sign the incoming state with user's signature
	userKey, _ := crypto.GenerateKey()
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSigBytes, _ := crypto.Sign(crypto.Keccak256Hash(packedState).Bytes(), userKey)
	userSigHex := hexutil.Encode(userSigBytes)
	incomingState.UserSig = &userSigHex

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return(true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", userWallet, asset).Return(nil)
	mockSigValidator.On("Verify", userWallet, packedState, userSigBytes).Return(nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeHomeDeposit &&
			tx.FromAccount == homeChannelID &&
			tx.ToAccount == userWallet &&
			tx.Amount.Equal(depositAmount)
	})).Return(nil)
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&
			len(state.Transitions) == 1 &&
			state.Transitions[0].Type == core.TransitionTypeHomeDeposit &&
			state.NodeSig != nil
	})).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitState_HomeWithdrawal_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	mockSigValidator := new(MockSigValidator)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint64(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:  mockMemoryStore,
		signer:       mockSigner,
		nodeAddress:  nodeAddress,
		minChallenge: minChallenge,
		sigValidators: map[SigValidatorType]SigValidator{
			EcdsaSigValidatorType: mockSigValidator,
		},
	}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	withdrawalAmount := decimal.NewFromInt(100)

	// Create user's current state (with existing home channel)
	currentState := core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 1),
		Transitions:   []core.Transition{},
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(300),
			UserNetFlow:  decimal.NewFromInt(400),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(-100),
		},
		EscrowLedger: nil,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Create incoming state with home withdrawal transition
	incomingState := currentState.NextState()

	// Apply the home withdrawal transition to update balances
	_, err := incomingState.ApplyHomeWithdrawalTransition(withdrawalAmount)
	require.NoError(t, err)

	// Sign the incoming state with user's signature
	userKey, _ := crypto.GenerateKey()
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSigBytes, _ := crypto.Sign(crypto.Keccak256Hash(packedState).Bytes(), userKey)
	userSigHex := hexutil.Encode(userSigBytes)
	incomingState.UserSig = &userSigHex

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return(true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", userWallet, asset).Return(nil)
	mockSigValidator.On("Verify", userWallet, packedState, userSigBytes).Return(nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeHomeWithdrawal &&
			tx.FromAccount == userWallet &&
			tx.ToAccount == homeChannelID &&
			tx.Amount.Equal(withdrawalAmount)
	})).Return(nil)
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&
			len(state.Transitions) == 1 &&
			state.Transitions[0].Type == core.TransitionTypeHomeWithdrawal &&
			state.NodeSig != nil
	})).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitState_MutualLock_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	mockSigValidator := new(MockSigValidator)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint64(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:  mockMemoryStore,
		signer:       mockSigner,
		nodeAddress:  nodeAddress,
		minChallenge: minChallenge,
		sigValidators: map[SigValidatorType]SigValidator{
			EcdsaSigValidatorType: mockSigValidator,
		},
	}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	lockAmount := decimal.NewFromInt(100)
	nonce := uint64(12345)
	challenge := uint64(86400)

	// Create user's current state (with existing home channel)
	currentState := core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 1),
		Transitions:   []core.Transition{},
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(500),
			UserNetFlow:  decimal.NewFromInt(500),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: nil,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Create incoming state with mutual lock transition
	incomingState := currentState.NextState()

	// Apply the mutual lock transition to update balances
	_, err := incomingState.ApplyMutualLockTransition(2, "0xTokenAddress", lockAmount)
	require.NoError(t, err)
	escrowChannelID := *incomingState.EscrowChannelID

	// Sign the incoming state with user's signature
	userKey, _ := crypto.GenerateKey()
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSigBytes, _ := crypto.Sign(crypto.Keccak256Hash(packedState).Bytes(), userKey)
	userSigHex := hexutil.Encode(userSigBytes)
	incomingState.UserSig = &userSigHex

	// Create home channel for mocking
	homeChannel := core.Channel{
		ChannelID:         homeChannelID,
		UserWallet:        userWallet,
		Type:              core.ChannelTypeHome,
		BlockchainID:      1,
		TokenAddress:      "0xTokenAddress",
		ChallengeDuration: challenge,
		Nonce:             nonce,
		Status:            core.ChannelStatusOpen,
		StateVersion:      1,
	}

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return(true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", userWallet, asset).Return(nil)
	mockSigValidator.On("Verify", userWallet, packedState, userSigBytes).Return(nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("GetChannelByID", homeChannelID).Return(&homeChannel, nil)
	mockMemoryStore.On("IsAssetSupported", asset, "0xTokenAddress", uint32(2)).Return(true, nil)
	mockTxStore.On("CreateChannel", mock.MatchedBy(func(channel core.Channel) bool {
		return channel.ChannelID == escrowChannelID &&
			channel.Type == core.ChannelTypeEscrow &&
			channel.UserWallet == userWallet
	})).Return(nil)
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeMutualLock &&
			tx.FromAccount == homeChannelID &&
			tx.ToAccount == escrowChannelID &&
			tx.Amount.Equal(lockAmount)
	})).Return(nil)
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&
			len(state.Transitions) == 1 &&
			state.Transitions[0].Type == core.TransitionTypeMutualLock &&
			state.NodeSig != nil
	})).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitState_EscrowDeposit_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	mockSigValidator := new(MockSigValidator)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint64(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:  mockMemoryStore,
		signer:       mockSigner,
		nodeAddress:  nodeAddress,
		minChallenge: minChallenge,
		sigValidators: map[SigValidatorType]SigValidator{
			EcdsaSigValidatorType: mockSigValidator,
		},
	}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	escrowChannelID := "0xEscrowChannel456"
	depositAmount := decimal.NewFromInt(100)

	// Create user's current state (signed, with escrow ledger)
	// The last transition must be a MutualLock for the EscrowDeposit to be valid
	currentSignedState := core.State{
		ID: core.GetStateID(userWallet, asset, 1, 2),
		Transitions: []core.Transition{
			{
				Type:      core.TransitionTypeMutualLock,
				TxID:      "0xPreviousMutualLockTx",
				AccountID: "",
				Amount:    depositAmount,
			},
		},
		Asset:           asset,
		UserWallet:      userWallet,
		Epoch:           1,
		Version:         2,
		HomeChannelID:   &homeChannelID,
		EscrowChannelID: &escrowChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(400),
			UserNetFlow:  decimal.NewFromInt(400),
			NodeBalance:  decimal.NewFromInt(100),
			NodeNetFlow:  decimal.NewFromInt(100),
		},
		EscrowLedger: &core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 2,
			UserBalance:  decimal.NewFromInt(100),
			UserNetFlow:  decimal.NewFromInt(100),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		UserSig: stringPtr("0xPreviousUserSig"),
		NodeSig: stringPtr("0xPreviousNodeSig"),
	}

	// Create incoming state with escrow deposit transition
	incomingState := currentSignedState.NextState()

	// Apply the escrow deposit transition to update balances
	_, err := incomingState.ApplyEscrowDepositTransition(depositAmount)
	require.NoError(t, err)

	// Sign the incoming state with user's signature
	userKey, _ := crypto.GenerateKey()
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSigBytes, _ := crypto.Sign(crypto.Keccak256Hash(packedState).Bytes(), userKey)
	userSigHex := hexutil.Encode(userSigBytes)
	incomingState.UserSig = &userSigHex

	// Create a copy for the unsigned state mock
	currentUnsignedState := currentSignedState

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return(true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentUnsignedState, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, true).Return(currentSignedState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", userWallet, asset).Return(nil)
	mockSigValidator.On("Verify", userWallet, packedState, userSigBytes).Return(nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)

	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeEscrowDeposit &&
			tx.FromAccount == escrowChannelID &&
			tx.ToAccount == userWallet &&
			tx.Amount.Equal(depositAmount)
	})).Return(nil)

	// Store incoming state with node signature
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&
			len(state.Transitions) == 1 &&
			state.Transitions[0].Type == core.TransitionTypeEscrowDeposit &&
			state.NodeSig != nil
	})).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestSubmitState_Finalize_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	mockSigValidator := new(MockSigValidator)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint64(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:  mockMemoryStore,
		signer:       mockSigner,
		nodeAddress:  nodeAddress,
		minChallenge: minChallenge,
		sigValidators: map[SigValidatorType]SigValidator{
			EcdsaSigValidatorType: mockSigValidator,
		},
	}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	userBalance := decimal.NewFromInt(300)

	// Create user's current state (with existing home channel and balance)
	currentState := core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 1),
		Transitions:   []core.Transition{},
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  userBalance,
			UserNetFlow:  userBalance,
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: nil,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Create incoming state with finalize transition
	incomingState := currentState.NextState()

	// Apply the finalize transition to update balances
	finalizeTransition, err := incomingState.ApplyFinalizeTransition()
	require.NoError(t, err)

	// Sign the incoming state with user's signature
	userKey, _ := crypto.GenerateKey()
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint32(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSigBytes, _ := crypto.Sign(crypto.Keccak256Hash(packedState).Bytes(), userKey)
	userSigHex := hexutil.Encode(userSigBytes)
	incomingState.UserSig = &userSigHex

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint32(1), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return(true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", userWallet, asset).Return(nil)
	mockSigValidator.On("Verify", userWallet, packedState, userSigBytes).Return(nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeFinalize &&
			tx.FromAccount == userWallet &&
			tx.ToAccount == homeChannelID &&
			tx.Amount.Equal(userBalance)
	})).Return(nil)
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&
			len(state.Transitions) == 1 &&
			state.Transitions[0].Type == core.TransitionTypeFinalize &&
			state.Transitions[0].Amount.Equal(userBalance) &&
			state.HomeLedger.UserBalance.IsZero() &&
			state.NodeSig != nil
	})).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)

	// Additional assertions specific to finalize transition
	assert.True(t, incomingState.IsFinal(), "State should be marked as final")
	assert.True(t, incomingState.HomeLedger.UserBalance.IsZero(), "User balance should be zero after finalize")
	assert.Equal(t, userBalance, finalizeTransition.Amount, "Finalize amount should equal the original user balance")
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}

// Helper function to convert core.State to rpc.StateV1
func toRPCState(state core.State) rpc.StateV1 {
	transitions := make([]rpc.TransitionV1, len(state.Transitions))
	for i, t := range state.Transitions {
		transitions[i] = rpc.TransitionV1{
			Type:      t.Type,
			TxID:      t.TxID,
			AccountID: t.AccountID,
			Amount:    t.Amount.String(),
		}
	}

	rpcState := rpc.StateV1{
		ID:              state.ID,
		Transitions:     transitions,
		Asset:           state.Asset,
		UserWallet:      state.UserWallet,
		Epoch:           decimal.NewFromInt(int64(state.Epoch)).String(),
		Version:         decimal.NewFromInt(int64(state.Version)).String(),
		HomeChannelID:   state.HomeChannelID,
		EscrowChannelID: state.EscrowChannelID,
		HomeLedger: rpc.LedgerV1{
			TokenAddress: state.HomeLedger.TokenAddress,
			BlockchainID: state.HomeLedger.BlockchainID,
			UserBalance:  state.HomeLedger.UserBalance.String(),
			UserNetFlow:  state.HomeLedger.UserNetFlow.String(),
			NodeBalance:  state.HomeLedger.NodeBalance.String(),
			NodeNetFlow:  state.HomeLedger.NodeNetFlow.String(),
		},
		UserSig: state.UserSig,
		NodeSig: state.NodeSig,
	}

	if state.EscrowLedger != nil {
		rpcState.EscrowLedger = &rpc.LedgerV1{
			TokenAddress: state.EscrowLedger.TokenAddress,
			BlockchainID: state.EscrowLedger.BlockchainID,
			UserBalance:  state.EscrowLedger.UserBalance.String(),
			UserNetFlow:  state.EscrowLedger.UserNetFlow.String(),
			NodeBalance:  state.EscrowLedger.NodeBalance.String(),
			NodeNetFlow:  state.EscrowLedger.NodeNetFlow.String(),
		}
	}

	return rpcState
}
