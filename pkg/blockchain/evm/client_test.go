package evm

import (
	"math/big"
	"strings"
	"testing"

	"github.com/erc7824/nitrolite/pkg/sign"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSigner implements sign.Signer
type MockSigner struct {
	mock.Mock
}

func (m *MockSigner) Sign(data []byte) (sign.Signature, error) {
	args := m.Called(data)
	return args.Get(0).(sign.Signature), args.Error(1)
}

func (m *MockSigner) PublicKey() sign.PublicKey {
	args := m.Called()
	return args.Get(0).(sign.PublicKey)
}

// MockPublicKey implements sign.PublicKey
type MockPublicKey struct {
	addr common.Address
}

func (m *MockPublicKey) Address() sign.Address {
	return sign.NewEthereumAddress(m.addr)
}

func (m *MockPublicKey) Bytes() []byte {
	return m.addr.Bytes()
}

func TestNewClient(t *testing.T) {
	t.Parallel()
	mockEVMClient := new(MockEVMClient)
	mockAssetStore := new(MockAssetStore)
	mockSigner := new(MockSigner)

	// Setup mock signer
	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	mockSigner.On("PublicKey").Return(&MockPublicKey{addr: addr})

	contractAddress := common.HexToAddress("0x123")
	nodeAddress := "0x456"
	blockchainID := uint64(1337)

	// NewClient calls NewChannelHub which doesn't make external calls in standard abigen,
	// but let's see. If it checks code, we need CodeAt.
	// Assuming standard abigen, it just returns the struct.

	client, err := NewClient(
		contractAddress,
		mockEVMClient,
		mockSigner,
		blockchainID,
		nodeAddress,
		mockAssetStore,
	)

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestClient_GetAccountsBalances(t *testing.T) {
	t.Parallel()
	mockEVMClient := new(MockEVMClient)
	mockAssetStore := new(MockAssetStore)
	mockSigner := new(MockSigner)

	// Setup mock signer
	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	mockSigner.On("PublicKey").Return(&MockPublicKey{addr: addr})

	contractAddress := common.HexToAddress("0xContract")
	client, err := NewClient(contractAddress, mockEVMClient, mockSigner, 1, "0xNode", mockAssetStore)
	require.NoError(t, err)

	accounts := []string{"0xUser1", "0xUser2"}
	tokens := []string{"0xToken1"}

	// Mock CallContract for GetAccountBalance
	// We need to match the call data. Since we can't easily reproduce abi packing here without the abi,
	// we will mock CallContract to return a generic success for any call.
	// In reality, we should check the method ID.
	// GetAccountBalance(bytes32 channelID, address token) returns (uint256)
	// Wait, the client calls c.contract.GetAccountBalance(nil, accountAddr, tokenAddr).

	// Mock successful return (uint256 = 100)
	ret := common.LeftPadBytes(big.NewInt(100).Bytes(), 32)
	mockEVMClient.On("CallContract", mock.Anything, mock.Anything, mock.Anything).Return(ret, nil)

	balances, err := client.GetAccountsBalances(accounts, tokens)
	require.NoError(t, err)
	assert.Len(t, balances, 2)
	assert.Len(t, balances[0], 1)
	assert.Equal(t, "100", balances[0][0].String())
	assert.Equal(t, "100", balances[1][0].String())
}

func TestClient_GetNodeBalance(t *testing.T) {
	t.Parallel()
	mockEVMClient := new(MockEVMClient)
	mockAssetStore := new(MockAssetStore)
	mockSigner := new(MockSigner)

	// Setup mock signer
	privKey, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	mockSigner.On("PublicKey").Return(&MockPublicKey{addr: addr})

	client, _ := NewClient(common.Address{}, mockEVMClient, mockSigner, 1, "0xNode", mockAssetStore)

	token := "0xToken"
	mockAssetStore.On("GetTokenDecimals", uint64(1), token).Return(uint8(18), nil)

	// Mock GetAccountBalance call
	ret := common.LeftPadBytes(big.NewInt(1000000000000000000).Bytes(), 32) // 1 ETH
	mockEVMClient.On("CallContract", mock.Anything, mock.Anything, mock.Anything).Return(ret, nil)

	balance, err := client.GetNodeBalance(token)
	require.NoError(t, err)
	assert.Equal(t, "1", balance.String())
}

func TestClient_GetOpenChannels(t *testing.T) {
	t.Parallel()
	mockEVMClient := new(MockEVMClient)
	mockAssetStore := new(MockAssetStore)
	mockSigner := new(MockSigner)

	// Setup mock signer
	privKey, err1 := crypto.GenerateKey()
	require.NoError(t, err1)
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	mockSigner.On("PublicKey").Return(&MockPublicKey{addr: addr})

	client, err2 := NewClient(common.Address{}, mockEVMClient, mockSigner, 1, "0xNode", mockAssetStore)
	require.NoError(t, err2)

	// Mock GetOpenChannels return: bytes32[]
	// Let's return 1 channel ID
	chanID := common.HexToHash("0x1234")
	// ABI encoding for dynamic array: offset, length, data
	// offset to data (32 bytes)
	offset := common.LeftPadBytes(big.NewInt(32).Bytes(), 32)
	// length (1)
	length := common.LeftPadBytes(big.NewInt(1).Bytes(), 32)
	// data (chanID)
	data := chanID.Bytes()

	ret := append(offset, length...)
	ret = append(ret, data...)

	mockEVMClient.On("CallContract", mock.Anything, mock.Anything, mock.Anything).Return(ret, nil)

	channels, err3 := client.GetOpenChannels("0xUser")
	require.NoError(t, err3)
	assert.Len(t, channels, 1)
	assert.Equal(t, strings.ToLower(hexutil.Encode(chanID[:])), strings.ToLower(channels[0]))
}
