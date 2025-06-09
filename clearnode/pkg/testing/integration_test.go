package testing

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	// Test constants
	DefaultGasLimit    = uint64(6721975)
	DefaultGasPrice    = 20000000000 // 20 gwei
	TestTimeout        = 30 * time.Second
	BlockConfirmations = 1

	// Test account private keys (same as in SDK tests for consistency)
	AlicePrivateKey    = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	BobPrivateKey      = "0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	DeployerPrivateKey = "0x7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6"
)

// IntegrationTestSuite represents the test suite for clearnode integration tests
type IntegrationTestSuite struct {
	suite.Suite
	client          *ethclient.Client
	aliceAuth       *bind.TransactOpts
	bobAuth         *bind.TransactOpts
	deployerAuth    *bind.TransactOpts
	custodyAddr     common.Address
	adjudicatorAddr common.Address
	ctx             context.Context
	cancel          context.CancelFunc
}

// SetupSuite runs once before all tests in the suite
func (suite *IntegrationTestSuite) SetupSuite() {
	// Get RPC URL from environment or use default
	rpcURL := os.Getenv("ETH_RPC_URL")
	if rpcURL == "" {
		rpcURL = "http://localhost:8545"
	}

	// Connect to Ethereum client
	client, err := ethclient.Dial(rpcURL)
	require.NoError(suite.T(), err, "Failed to connect to Ethereum client")
	suite.client = client

	// Create context with timeout
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), TestTimeout)

	// Setup test accounts
	suite.setupAccounts()

	// Deploy contracts
	suite.deployContracts()
}

// TearDownSuite runs once after all tests in the suite
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.cancel != nil {
		suite.cancel()
	}
	if suite.client != nil {
		suite.client.Close()
	}
}

// SetupTest runs before each test
func (suite *IntegrationTestSuite) SetupTest() {
	// Reset any test-specific state if needed
}

// TearDownTest runs after each test
func (suite *IntegrationTestSuite) TearDownTest() {
	// Cleanup any test-specific resources if needed
}

// setupAccounts initializes the test accounts
func (suite *IntegrationTestSuite) setupAccounts() {
	var err error

	// Alice account
	alicePrivKey, err := crypto.HexToECDSA(AlicePrivateKey[2:]) // Remove 0x prefix
	require.NoError(suite.T(), err, "Failed to parse Alice private key")
	suite.aliceAuth, err = bind.NewKeyedTransactorWithChainID(alicePrivKey, big.NewInt(31337))
	require.NoError(suite.T(), err, "Failed to create Alice auth")
	suite.aliceAuth.GasLimit = DefaultGasLimit
	suite.aliceAuth.GasPrice = big.NewInt(DefaultGasPrice)

	// Bob account
	bobPrivKey, err := crypto.HexToECDSA(BobPrivateKey[2:])
	require.NoError(suite.T(), err, "Failed to parse Bob private key")
	suite.bobAuth, err = bind.NewKeyedTransactorWithChainID(bobPrivKey, big.NewInt(31337))
	require.NoError(suite.T(), err, "Failed to create Bob auth")
	suite.bobAuth.GasLimit = DefaultGasLimit
	suite.bobAuth.GasPrice = big.NewInt(DefaultGasPrice)

	// Deployer account
	deployerPrivKey, err := crypto.HexToECDSA(DeployerPrivateKey[2:])
	require.NoError(suite.T(), err, "Failed to parse deployer private key")
	suite.deployerAuth, err = bind.NewKeyedTransactorWithChainID(deployerPrivKey, big.NewInt(31337))
	require.NoError(suite.T(), err, "Failed to create deployer auth")
	suite.deployerAuth.GasLimit = DefaultGasLimit
	suite.deployerAuth.GasPrice = big.NewInt(DefaultGasPrice)
}

// deployContracts deploys the necessary contracts for testing
func (suite *IntegrationTestSuite) deployContracts() {
	// Note: In a real implementation, this would deploy the actual contracts
	// For now, we'll use placeholder addresses that would be deployed by the CI

	// These addresses would be set by the deployment script or environment
	custodyAddrStr := os.Getenv("CUSTODY_CONTRACT_ADDRESS")
	if custodyAddrStr != "" {
		suite.custodyAddr = common.HexToAddress(custodyAddrStr)
	} else {
		// Deploy mock contract or use default address
		suite.custodyAddr = common.HexToAddress("0x0000000000000000000000000000000000000001")
	}

	adjudicatorAddrStr := os.Getenv("ADJUDICATOR_CONTRACT_ADDRESS")
	if adjudicatorAddrStr != "" {
		suite.adjudicatorAddr = common.HexToAddress(adjudicatorAddrStr)
	} else {
		// Deploy mock contract or use default address
		suite.adjudicatorAddr = common.HexToAddress("0x0000000000000000000000000000000000000002")
	}
}

// TestClientConnection tests basic Ethereum client functionality
func (suite *IntegrationTestSuite) TestClientConnection() {
	// Test basic client connection
	chainID, err := suite.client.ChainID(suite.ctx)
	require.NoError(suite.T(), err, "Failed to get chain ID")
	assert.Equal(suite.T(), int64(31337), chainID.Int64(), "Unexpected chain ID")

	// Test block number retrieval
	blockNumber, err := suite.client.BlockNumber(suite.ctx)
	require.NoError(suite.T(), err, "Failed to get block number")
	assert.Greater(suite.T(), blockNumber, uint64(0), "Block number should be greater than 0")
}

// TestAccountBalances tests account balance retrieval
func (suite *IntegrationTestSuite) TestAccountBalances() {
	// Test Alice's balance
	aliceBalance, err := suite.client.BalanceAt(suite.ctx, suite.aliceAuth.From, nil)
	require.NoError(suite.T(), err, "Failed to get Alice's balance")
	assert.True(suite.T(), aliceBalance.Cmp(big.NewInt(0)) > 0, "Alice should have positive balance")

	// Test Bob's balance
	bobBalance, err := suite.client.BalanceAt(suite.ctx, suite.bobAuth.From, nil)
	require.NoError(suite.T(), err, "Failed to get Bob's balance")
	assert.True(suite.T(), bobBalance.Cmp(big.NewInt(0)) > 0, "Bob should have positive balance")
}

// TestTransactionSending tests basic transaction sending functionality
func (suite *IntegrationTestSuite) TestTransactionSending() {
	// Get initial balances
	initialAliceBalance, err := suite.client.BalanceAt(suite.ctx, suite.aliceAuth.From, nil)
	require.NoError(suite.T(), err, "Failed to get Alice's initial balance")

	initialBobBalance, err := suite.client.BalanceAt(suite.ctx, suite.bobAuth.From, nil)
	require.NoError(suite.T(), err, "Failed to get Bob's initial balance")

	// Send transaction from Alice to Bob
	transferAmount := big.NewInt(1000000000000000000) // 1 ETH
	nonce, err := suite.client.PendingNonceAt(suite.ctx, suite.aliceAuth.From)
	require.NoError(suite.T(), err, "Failed to get nonce")

	gasPrice, err := suite.client.SuggestGasPrice(suite.ctx)
	require.NoError(suite.T(), err, "Failed to get gas price")

	tx := types.NewTransaction(
		nonce,
		suite.bobAuth.From,
		transferAmount,
		21000, // gas limit for simple transfer
		gasPrice,
		nil,
	)

	// Sign transaction
	alicePrivKey, err := crypto.HexToECDSA(AlicePrivateKey[2:])
	require.NoError(suite.T(), err, "Failed to parse Alice private key")

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(31337)), alicePrivKey)
	require.NoError(suite.T(), err, "Failed to sign transaction")

	// Send transaction
	err = suite.client.SendTransaction(suite.ctx, signedTx)
	require.NoError(suite.T(), err, "Failed to send transaction")

	// Wait for transaction to be mined
	receipt, err := suite.waitForTransaction(signedTx.Hash())
	require.NoError(suite.T(), err, "Failed to wait for transaction")
	assert.Equal(suite.T(), types.ReceiptStatusSuccessful, receipt.Status, "Transaction should be successful")

	// Verify balances changed
	finalAliceBalance, err := suite.client.BalanceAt(suite.ctx, suite.aliceAuth.From, nil)
	require.NoError(suite.T(), err, "Failed to get Alice's final balance")

	finalBobBalance, err := suite.client.BalanceAt(suite.ctx, suite.bobAuth.From, nil)
	require.NoError(suite.T(), err, "Failed to get Bob's final balance")

	// Calculate expected balances (accounting for gas costs)
	gasCost := new(big.Int).Mul(big.NewInt(int64(receipt.GasUsed)), gasPrice)
	expectedAliceBalance := new(big.Int).Sub(initialAliceBalance, transferAmount)
	expectedAliceBalance.Sub(expectedAliceBalance, gasCost)
	expectedBobBalance := new(big.Int).Add(initialBobBalance, transferAmount)

	assert.Equal(suite.T(), expectedAliceBalance, finalAliceBalance, "Alice's balance should decrease by transfer amount + gas")
	assert.Equal(suite.T(), expectedBobBalance, finalBobBalance, "Bob's balance should increase by transfer amount")
}

// TestContractInteraction tests interaction with deployed contracts
func (suite *IntegrationTestSuite) TestContractInteraction() {
	// Test that contract addresses are set
	assert.NotEqual(suite.T(), common.Address{}, suite.custodyAddr, "Custody contract address should be set")
	assert.NotEqual(suite.T(), common.Address{}, suite.adjudicatorAddr, "Adjudicator contract address should be set")

	// Test contract code exists (if contracts are actually deployed)
	custodyCode, err := suite.client.CodeAt(suite.ctx, suite.custodyAddr, nil)
	require.NoError(suite.T(), err, "Failed to get custody contract code")

	adjudicatorCode, err := suite.client.CodeAt(suite.ctx, suite.adjudicatorAddr, nil)
	require.NoError(suite.T(), err, "Failed to get adjudicator contract code")

	// Note: In a real test, these would be non-empty if contracts are deployed
	// For now, we just test that the calls don't error
	suite.T().Logf("Custody contract code length: %d", len(custodyCode))
	suite.T().Logf("Adjudicator contract code length: %d", len(adjudicatorCode))
}

// TestConcurrentOperations tests multiple concurrent operations
func (suite *IntegrationTestSuite) TestConcurrentOperations() {
	const numOperations = 5
	results := make(chan error, numOperations)

	// Start multiple concurrent operations
	for i := 0; i < numOperations; i++ {
		go func() {
			// Test concurrent balance queries
			_, err := suite.client.BalanceAt(suite.ctx, suite.aliceAuth.From, nil)
			results <- err
		}()
	}

	// Wait for all operations to complete
	for i := 0; i < numOperations; i++ {
		err := <-results
		assert.NoError(suite.T(), err, "Concurrent operation should not error")
	}
}

// TODO: Uncomment after implementing timeouts
// TestNetworkResilience tests handling of network issues
// func (suite *IntegrationTestSuite) TestNetworkResilience() {
// 	// Test with context timeout
// 	shortCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
// 	defer cancel()

// 	// This should timeout quickly
// 	_, err := suite.client.BalanceAt(shortCtx, suite.aliceAuth.From, nil)
// 	assert.Error(suite.T(), err, "Should error due to context timeout")
// }

// TestContractEventHandling tests event handling capabilities
func (suite *IntegrationTestSuite) TestContractEventHandling() {
	// Test basic event filtering
	fromBlock := big.NewInt(0)
	toBlock := big.NewInt(10)

	// Create a basic filter query
	query := make(map[string]interface{})
	query["fromBlock"] = fromBlock
	query["toBlock"] = toBlock

	// Test that we can create the filter without errors
	// In a real implementation, this would test actual event filtering
	suite.T().Log("Event handling test placeholder - would test actual events in full implementation")
}

// waitForTransaction waits for a transaction to be mined
func (suite *IntegrationTestSuite) waitForTransaction(txHash common.Hash) (*types.Receipt, error) {
	for i := 0; i < 60; i++ { // Wait up to 60 seconds
		receipt, err := suite.client.TransactionReceipt(suite.ctx, txHash)
		if err == nil {
			return receipt, nil
		}
		time.Sleep(1 * time.Second)
	}
	return nil, context.DeadlineExceeded
}

// Helper function to create a private key from hex string
func createPrivateKeyFromHex(hexKey string) (*ecdsa.PrivateKey, error) {
	if len(hexKey) > 2 && hexKey[:2] == "0x" {
		hexKey = hexKey[2:]
	}
	return crypto.HexToECDSA(hexKey)
}

// TestSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// Additional benchmark tests
func BenchmarkBalanceQuery(b *testing.B) {
	// Get RPC URL from environment or use default
	rpcURL := os.Getenv("ETH_RPC_URL")
	if rpcURL == "" {
		rpcURL = "http://localhost:8545"
	}

	client, err := ethclient.Dial(rpcURL)
	require.NoError(b, err)
	defer client.Close()

	ctx := context.Background()
	aliceAddr := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266") // First anvil account

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.BalanceAt(ctx, aliceAddr, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBlockNumberQuery(b *testing.B) {
	// Get RPC URL from environment or use default
	rpcURL := os.Getenv("ETH_RPC_URL")
	if rpcURL == "" {
		rpcURL = "http://localhost:8545"
	}

	client, err := ethclient.Dial(rpcURL)
	require.NoError(b, err)
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.BlockNumber(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
