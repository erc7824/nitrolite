package channel_v1

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

func Test_parseRequestCreation(t *testing.T) {
	var payload rpc.Payload
	jsonStr := `
    {
        "state": {
            "id": "0x59c28c20c7676b8d42ab220a8caeef2a9c795e05dec42cb1e05f14f32b7ff836",
            "transitions": [
                {
                    "type": 30,
                    "tx_id": "0xd245bd97aab68a53980565db77f0abf893378d3358c8d57a34289364a293e8cb",
                    "account_id": "0x053aEAD7d3eebE4359300fDE849bCD9E77384989",
                    "amount": "0.1"
                }
            ],
            "asset": "usdc",
            "user_wallet": "0x8a395641469fab5ebf10feb5b33c493c99e251c4",
            "epoch": "0",
            "version": "2",
            "home_channel_id": "0x220337be1fb67e3e5935d548389dbd535bb039e8e0bfc2e30d51db2bbefb47e0",
            "home_ledger": {
                "token_address": "0x6E2C4707DA119425dF2c722E2695300154652f56",
                "blockchain_id": "11155111",
                "user_balance": "0.1",
                "user_net_flow": "0",
                "node_balance": "0",
                "node_net_flow": "0.1"
            },
            "user_sig": "0xe04d98afcf53eb6b249298e827122f251ee01547092b3a23280a1ba266daa9a50309bf782e0f1afcbb12cf6c0db02213631b38a31edfe3c033b07af94d1109e41b"
        },
        "channel_definition": {
            "nonce": "1770377043089581224",
            "challenge": 86400
        }
    }`
	err := json.Unmarshal([]byte(jsonStr), &payload)
	require.NoError(t, err)

	var reqPayload rpc.ChannelsV1RequestCreationRequest
	err = payload.Translate(&reqPayload)
	require.NoError(t, err)

	fmt.Printf("Parsed Request Creation Payload: %+v\n", reqPayload)
	_, err = toCoreChannelDefinition(reqPayload.ChannelDefinition)
	require.NoError(t, err)
}

func TestRequestCreation_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	mockSigValidator := new(MockSigValidator)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600) // 1 hour
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
	tokenAddress := "0xTokenAddress"
	blockchainID := uint64(1)
	nonce := uint64(12345)
	challenge := uint32(86400)
	depositAmount := decimal.NewFromInt(1000)

	// Create void state (starting point)
	voidState := core.NewVoidState(asset, userWallet)

	// Create next state from void
	initialState := voidState.NextState()

	channelDef := core.ChannelDefinition{
		Nonce:     nonce,
		Challenge: challenge,
	}
	_, err := initialState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
	require.NoError(t, err)

	// Apply the home deposit transition to update balances
	_, err = initialState.ApplyHomeDepositTransition(depositAmount)
	require.NoError(t, err)

	// Set up mock for PackState (called during signing)
	mockAssetStore.On("GetTokenDecimals", blockchainID, tokenAddress).Return(uint8(6), nil)

	// Sign the initial state
	packedState, err := core.PackState(*initialState, mockAssetStore)
	require.NoError(t, err)
	stateHash := crypto.Keccak256Hash(packedState).Bytes()
	userSig, err := mockSigner.Sign(stateHash)
	require.NoError(t, err)
	userSigStr := userSig.String()
	initialState.UserSig = &userSigStr

	// Mock expectations for handler
	mockMemoryStore.On("IsAssetSupported", asset, tokenAddress, blockchainID).Return(true, nil).Once()
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil).Once()
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(nil, nil).Once()
	mockSigValidator.On("Verify", userWallet, packedState, mock.Anything).Return(nil).Once()
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("CreateChannel", mock.MatchedBy(func(channel core.Channel) bool {
		return channel.UserWallet == userWallet &&
			channel.Type == core.ChannelTypeHome &&
			channel.BlockchainID == blockchainID &&
			channel.TokenAddress == tokenAddress &&
			channel.Nonce == nonce &&
			channel.ChallengeDuration == challenge &&
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
			state.Epoch == 0 &&
			state.NodeSig != nil &&
			state.HomeChannelID != nil
	})).Return(nil).Once()

	// Create RPC request
	rpcState := toRPCState(*initialState)
	reqPayload := rpc.ChannelsV1RequestCreationRequest{
		State: rpcState,
		ChannelDefinition: rpc.ChannelDefinitionV1{
			Nonce:     strconv.FormatUint(nonce, 10),
			Challenge: challenge,
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
	mockMemoryStore.AssertExpectations(t)
	mockAssetStore.AssertExpectations(t)
	mockTxStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

func TestRequestCreation_InvalidChallenge(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	mockSigValidator := new(MockSigValidator)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600) // 1 hour
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockTxStore)
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
	tokenAddress := "0xToken"
	nonce := uint64(12345)
	lowChallenge := uint32(1800) // 30 minutes - below minimum

	// Calculate home channel ID
	homeChannelID, err := core.GetHomeChannelID(
		nodeAddress,
		userWallet,
		asset,
		nonce,
		lowChallenge,
	)
	require.NoError(t, err)

	mockMemoryStore.On("IsAssetSupported", asset, tokenAddress, uint64(1)).Return(true, nil).Once()
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(nil, nil).Once()

	// Create RPC request with challenge below minimum
	reqPayload := rpc.ChannelsV1RequestCreationRequest{
		State: rpc.StateV1{
			ID:            core.GetStateID(userWallet, asset, 1, 1),
			UserWallet:    userWallet,
			Asset:         asset,
			Epoch:         "1",
			Version:       "1",
			HomeChannelID: &homeChannelID,
			HomeLedger: rpc.LedgerV1{
				TokenAddress: tokenAddress,
				BlockchainID: "1",
				UserBalance:  "0",
				UserNetFlow:  "0",
				NodeBalance:  "0",
				NodeNetFlow:  "0",
			},
		},
		ChannelDefinition: rpc.ChannelDefinitionV1{
			Nonce:     strconv.FormatUint(nonce, 10),
			Challenge: lowChallenge,
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
