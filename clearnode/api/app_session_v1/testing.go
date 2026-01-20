package app_session_v1

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/erc7824/nitrolite/pkg/sign"
)

// MockStore is a mock implementation of the Store interface
type MockStore struct {
	mock.Mock
}

func (m *MockStore) CreateAppSession(session app.AppSessionV1) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockStore) GetAppSession(sessionID string) (*app.AppSessionV1, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*app.AppSessionV1), args.Error(1)
}

func (m *MockStore) GetAppSessions(appSessionID *string, participant *string, status *string, pagination *core.PaginationParams) ([]app.AppSessionV1, core.PaginationMetadata, error) {
	args := m.Called(appSessionID, participant, status, pagination)
	if args.Get(0) == nil {
		return nil, core.PaginationMetadata{}, args.Error(2)
	}
	var metadata core.PaginationMetadata
	if args.Get(1) != nil {
		metadata = args.Get(1).(core.PaginationMetadata)
	}
	return args.Get(0).([]app.AppSessionV1), metadata, args.Error(2)
}

func (m *MockStore) UpdateAppSession(session app.AppSessionV1) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockStore) GetAppSessionBalances(sessionID string) (map[string]decimal.Decimal, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]decimal.Decimal), args.Error(1)
}

func (m *MockStore) GetParticipantAllocations(sessionID string) (map[string]map[string]decimal.Decimal, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]map[string]decimal.Decimal), args.Error(1)
}

func (m *MockStore) RecordLedgerEntry(accountID, asset string, amount decimal.Decimal, sessionKey *string) error {
	args := m.Called(accountID, asset, amount, sessionKey)
	return args.Error(0)
}

func (m *MockStore) GetAccountBalance(accountID, asset string) (decimal.Decimal, error) {
	args := m.Called(accountID, asset)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockStore) RecordChannelTransaction(tx core.Transaction) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockStore) RecordTransaction(tx core.Transaction) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockStore) CheckOpenChannel(wallet, asset string) (bool, error) {
	args := m.Called(wallet, asset)
	return args.Bool(0), args.Error(1)
}

func (m *MockStore) GetLastUserState(wallet, asset string, signed bool) (*core.State, error) {
	args := m.Called(wallet, asset, signed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	state := args.Get(0).(core.State)
	return &state, args.Error(1)
}

func (m *MockStore) StoreUserState(state core.State) error {
	args := m.Called(state)
	return args.Error(0)
}

func (m *MockStore) EnsureNoOngoingStateTransitions(wallet, asset string) error {
	args := m.Called(wallet, asset)
	return args.Error(0)
}

func (m *MockStore) CheckChallengedChannels(wallet string) error {
	args := m.Called(wallet)
	return args.Error(0)
}

func (m *MockStore) EnsureWalletHasAllAllocationsEmpty(wallet string) error {
	args := m.Called(wallet)
	return args.Error(0)
}

// MockSigValidator is a mock implementation of the SigValidator interface
type MockSigValidator struct {
	mock.Mock
}

func (m *MockSigValidator) Recover(data, sig []byte) (string, error) {
	args := m.Called(data, sig)
	return args.String(0), args.Error(1)
}

func (m *MockSigValidator) Verify(wallet string, data, sig []byte) error {
	args := m.Called(wallet, data, sig)
	return args.Error(0)
}

// NewMockSigner creates a mock signer for testing
func NewMockSigner() sign.Signer {
	key, _ := crypto.GenerateKey()
	signer, _ := sign.NewEthereumSigner(hexutil.Encode(crypto.FromECDSA(key)))
	return signer
}

// toRPCState converts a core.State to rpc.StateV1 for testing
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
