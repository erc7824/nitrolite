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
	mockStore := new(MockStore)
	mockTxStore := new(MockStore)
	mockSigner := NewMockSigner()
	mockSigValidator := new(MockSigValidator)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint64(3600)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(),
		useStoreInTx: func(handler StoreTxHandler) error {
			commit := func() error { return nil }
			revert := func() error { return nil }
			mockStore.On("BeginTx").Return(mockTxStore, commit, revert).Once()

			_, _, _ = mockStore.BeginTx()
			err := handler(mockTxStore)
			if err != nil {
				_ = revert()
				return err
			}
			return commit()
		},
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
	txHash := "0xTransferTxHash"

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
			UserNetFlow:  decimal.NewFromInt(0),
			NodeBalance:  decimal.NewFromInt(500),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: nil,
		IsFinal:      false,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Create incoming sender state (with transfer send transition)
	incomingSenderState := currentSenderState.NextState()
	transferSendTransition := core.Transition{
		Type:      core.TransitionTypeTransferSend,
		TxHash:    txHash,
		AccountID: receiverWallet,
		Amount:    transferAmount,
	}

	// Apply the transfer send transition to update balances
	var err error
	incomingSenderState, err = handler.stateAdvancer.ApplyTransition(incomingSenderState, transferSendTransition)
	require.NoError(t, err)

	// Sign the incoming sender state with user's signature
	userKey, _ := crypto.GenerateKey()
	packedSenderState, _ := core.PackState(incomingSenderState)
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
			UserNetFlow:  decimal.NewFromInt(0),
			NodeBalance:  decimal.NewFromInt(800),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: nil,
		IsFinal:      false,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Expected receiver state after transfer receive
	expectedReceiverState := currentReceiverState.NextState()
	transferReceiveTransition := core.Transition{
		Type:      core.TransitionTypeTransferReceive,
		TxHash:    txHash,
		AccountID: senderWallet,
		Amount:    transferAmount,
	}
	expectedReceiverState, err = handler.stateAdvancer.ApplyTransition(expectedReceiverState, transferReceiveTransition)
	require.NoError(t, err)

	// Mock expectations
	mockTxStore.On("GetLastUserState", senderWallet, asset, false).Return(currentSenderState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", senderWallet, asset).Return(nil)
	mockSigValidator.On("Verify", senderWallet, packedSenderState, userSigBytes).Return(nil)

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
	rpcState := toRPCState(incomingSenderState)
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
	mockStore.AssertExpectations(t)
	mockTxStore.AssertExpectations(t)
	mockSigValidator.AssertExpectations(t)
}

// Helper function to convert core.State to rpc.StateV1
func toRPCState(state core.State) rpc.StateV1 {
	transitions := make([]rpc.TransitionV1, len(state.Transitions))
	for i, t := range state.Transitions {
		transitions[i] = rpc.TransitionV1{
			Type:      t.Type,
			TxHash:    t.TxHash,
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
		IsFinal: state.IsFinal,
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
