package app_session_v1

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
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

func (m *MockStore) GetAppSessions(appSessionID *string, participant *string, status app.AppSessionStatus, pagination *core.PaginationParams) ([]app.AppSessionV1, core.PaginationMetadata, error) {
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

func (m *MockStore) RecordLedgerEntry(userWallet, accountID, asset string, amount decimal.Decimal) error {
	args := m.Called(userWallet, accountID, asset, amount)
	return args.Error(0)
}

func (m *MockStore) GetAccountBalance(accountID, asset string) (decimal.Decimal, error) {
	args := m.Called(accountID, asset)
	return args.Get(0).(decimal.Decimal), args.Error(1)
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

func (m *MockStore) EnsureNoOngoingStateTransitions(wallet, asset string, prevTransitionType core.TransitionType) error {
	args := m.Called(wallet, asset, prevTransitionType)
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

// MockAssetStore is a mock implementation of the core.AssetStore interface
type MockAssetStore struct {
	mock.Mock
}

func (m *MockAssetStore) GetAssetDecimals(asset string) (uint8, error) {
	args := m.Called(asset)
	return args.Get(0).(uint8), args.Error(1)
}

func (m *MockAssetStore) GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error) {
	args := m.Called(blockchainID, tokenAddress)
	return args.Get(0).(uint8), args.Error(1)
}

type MockStatePacker struct {
	mock.Mock
}

func (m *MockStatePacker) PackState(state core.State) ([]byte, error) {
	args := m.Called(state)
	return args.Get(0).([]byte), args.Error(1)
}

// NewMockSigner creates a mock signer for testing
func NewMockSigner() sign.Signer {
	key, _ := crypto.GenerateKey()
	signer, _ := sign.NewEthereumMsgSigner(hexutil.Encode(crypto.FromECDSA(key)))
	return signer
}
