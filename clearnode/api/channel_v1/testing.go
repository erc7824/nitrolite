package channel_v1

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/mock"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/sign"
)

// MockStore is a mock implementation of the Store interface
type MockStore struct {
	mock.Mock
}

func (m *MockStore) BeginTx() (Store, func() error, func() error) {
	args := m.Called()
	return args.Get(0).(Store), args.Get(1).(func() error), args.Get(2).(func() error)
}

func (m *MockStore) GetLastUserState(wallet, asset string, signed bool) (*core.State, error) {
	args := m.Called(wallet, asset, signed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	state := args.Get(0).(core.State)
	return &state, args.Error(1)
}

func (m *MockStore) CheckOpenChannel(wallet, asset string) (bool, error) {
	args := m.Called(wallet, asset)
	return args.Bool(0), args.Error(1)
}

func (m *MockStore) StoreUserState(state core.State) error {
	args := m.Called(state)
	return args.Error(0)
}

func (m *MockStore) EnsureNoOngoingStateTransitions(wallet, asset string) error {
	args := m.Called(wallet, asset)
	return args.Error(0)
}

func (m *MockStore) ScheduleInitiateEscrowWithdrawal(stateID string) error {
	args := m.Called(stateID)
	return args.Error(0)
}

func (m *MockStore) RecordTransaction(tx core.Transaction) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockStore) CreateChannel(channel core.Channel) error {
	args := m.Called(channel)
	return args.Error(0)
}

func (m *MockStore) GetChannelByID(channelID string) (*core.Channel, error) {
	args := m.Called(channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Channel), args.Error(1)
}

func (m *MockStore) GetActiveHomeChannel(wallet, asset string) (*core.Channel, error) {
	args := m.Called(wallet, asset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Channel), args.Error(1)
}

func NewMockSigner() sign.Signer {
	key, _ := crypto.GenerateKey()

	signer, _ := sign.NewEthereumSigner(hexutil.Encode(crypto.FromECDSA(key)))
	return signer
}

// MockSigValidator is a mock implementation of the SigValidator interface
type MockSigValidator struct {
	mock.Mock
}

func (m *MockSigValidator) Verify(wallet string, data, sig []byte) error {
	args := m.Called(wallet, data, sig)
	return args.Error(0)
}
