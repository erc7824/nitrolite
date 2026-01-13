package channel_v1

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

func TestRequestCreation_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockSigner := NewMockSigner()
	mockSigValidator := new(MockSigValidator)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint64(3600) // 1 hour

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(),
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
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
	tokenAddress := "0xTokenAddress"
	blockchainID := uint32(1)
	nonce := uint64(12345)
	challenge := uint64(86400)
	depositAmount := decimal.NewFromInt(1000)

	// Create void state (starting point)
	voidState := core.NewVoidState(asset, userWallet)
	voidState.Epoch = 1
	voidState.Version = 0
	voidState.HomeLedger.TokenAddress = tokenAddress
	voidState.HomeLedger.BlockchainID = blockchainID

	// Create next state from void
	initialState := voidState.NextState()

	depositTxID, err := core.GetSenderTransactionID(userWallet, initialState.ID)
	require.NoError(t, err)

	// Create and apply home deposit transition
	homeDepositTransition := core.Transition{
		Type:      core.TransitionTypeHomeDeposit,
		TxID:      depositTxID,
		AccountID: userWallet,
		Amount:    depositAmount,
	}

	// Apply the home deposit transition to update balances
	initialState, err = handler.stateAdvancer.ApplyTransition(initialState, homeDepositTransition)
	require.NoError(t, err)

	// Calculate and set the home channel ID
	homeChannelID, err := core.GetHomeChannelID(
		nodeAddress,
		userWallet,
		tokenAddress,
		nonce,
		challenge,
	)
	require.NoError(t, err)
	initialState.HomeChannelID = &homeChannelID

	// Sign the initial state
	packedState, err := core.PackState(initialState)
	require.NoError(t, err)
	stateHash := crypto.Keccak256Hash(packedState).Bytes()
	userSig, err := mockSigner.Sign(stateHash)
	require.NoError(t, err)
	userSigStr := userSig.String()
	initialState.UserSig = &userSigStr

	// Mock expectations
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(nil, nil).Once()
	mockSigValidator.On("Verify", userWallet, packedState, mock.Anything).Return(nil).Once()
	mockTxStore.On("CreateChannel", mock.MatchedBy(func(channel core.Channel) bool {
		return channel.UserWallet == userWallet &&
			channel.NodeWallet == nodeAddress &&
			channel.Type == core.ChannelTypeHome &&
			channel.BlockchainID == blockchainID &&
			channel.TokenAddress == tokenAddress &&
			channel.Nonce == nonce &&
			channel.Challenge == challenge &&
			channel.Status == core.ChannelStatusVoid &&
			channel.StateVersion == 0
	})).Return(nil).Once()
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		// For home_deposit: fromAccount is homeChannelID, toAccount is userWallet
		return tx.TxType == core.TransactionTypeHomeDeposit &&
			tx.ToAccount == userWallet &&
			tx.FromAccount != "" // homeChannelID will be set by handler
	})).Return(nil).Once()
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Asset == asset &&
			state.Version == 1 &&
			state.Epoch == 1 &&
			state.NodeSig != nil &&
			state.HomeChannelID != nil
	})).Return(nil).Once()

	// Create RPC request
	rpcState := toRPCState(initialState)
	reqPayload := rpc.ChannelsV1RequestCreationRequest{
		State: rpcState,
		ChannelDefinition: rpc.ChannelDefinitionV1{
			Nonce:     decimal.NewFromInt(int64(nonce)).String(),
			Challenge: decimal.NewFromInt(int64(challenge)).String(),
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{
			RequestID: 1,
			Method:    rpc.ChannelsV1RequestCreationMethod.String(),
			Payload:   payload,
		},
	}

	// Execute
	handler.RequestCreation(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Check for errors first
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}

	assert.Equal(t, rpc.ChannelsV1RequestCreationMethod.String(), ctx.Response.Method)
	assert.NotNil(t, ctx.Response.Payload)

	// Verify response contains signature
	var response rpc.ChannelsV1RequestCreationResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.Signature)

	// Verify all mocks were called
	mockTxStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestRequestCreation_InvalidChallenge(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockSigner := NewMockSigner()
	mockSigValidator := new(MockSigValidator)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint64(3600) // 1 hour

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(),
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockTxStore)
		},
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
	lowChallenge := uint64(1800) // 30 minutes - below minimum

	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(nil, nil).Once()

	// Create RPC request with challenge below minimum
	reqPayload := rpc.ChannelsV1RequestCreationRequest{
		State: rpc.StateV1{
			ID:         core.GetStateID(userWallet, asset, 1, 1),
			UserWallet: userWallet,
			Asset:      asset,
			Epoch:      "1",
			Version:    "1",
			HomeLedger: rpc.LedgerV1{
				TokenAddress: "0xToken",
				BlockchainID: 1,
				UserBalance:  "0",
				UserNetFlow:  "0",
				NodeBalance:  "0",
				NodeNetFlow:  "0",
			},
			IsFinal: false,
		},
		ChannelDefinition: rpc.ChannelDefinitionV1{
			Nonce:     "12345",
			Challenge: decimal.NewFromInt(int64(lowChallenge)).String(),
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{
			RequestID: 1,
			Method:    rpc.ChannelsV1RequestCreationMethod.String(),
			Payload:   payload,
		},
	}

	// Execute
	handler.RequestCreation(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Verify response contains error
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "challenge")

	// Verify all mocks were called
	mockTxStore.AssertExpectations(t)
}
