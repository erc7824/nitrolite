package sdk

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/pkg/blockchain/evm"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/erc7824/nitrolite/pkg/sign"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

// Client provides a unified interface for interacting with Clearnode.
// It combines both high-level operations (Deposit, Withdraw, Transfer) and
// low-level RPC access for advanced use cases.
//
// High-level example:
//
//	stateSigner, _ := sign.NewEthereumMsgSigner(privateKeyHex)
//	txSigner, _ := sign.NewEthereumRawSigner(privateKeyHex)
//	client, _ := sdk.NewClient(
//	    "wss://clearnode.example.com/ws",
//	    stateSigner,
//	    txSigner,
//	    sdk.WithBlockchainRPC(80002, "https://polygon-amoy.alchemy.com/v2/KEY"),
//	)
//	defer client.Close()
//
//	// High-level operations
//	txHash, _ := client.Deposit(ctx, 80002, "usdc", decimal.NewFromInt(100))
//	txID, _ := client.Transfer(ctx, "0xRecipient...", "usdc", decimal.NewFromInt(50))
//
//	// Low-level operations
//	config, _ := client.GetConfig(ctx)
//	balances, _ := client.GetBalances(ctx, walletAddress)
type Client struct {
	rpcDialer         rpc.Dialer
	rpcClient         *rpc.Client
	config            Config
	exitCh            chan struct{}
	blockchainClients map[uint64]*evm.Client
	homeBlockchains   map[string]uint64
	stateSigner       sign.Signer
	txSigner          sign.Signer
	assetStore        *clientAssetStore
}

// clientAssetStore implements core.AssetStore by fetching data from the Clearnode API.
type clientAssetStore struct {
	client *Client
	cache  map[string]core.Asset // asset symbol -> Asset
}

func newClientAssetStore(client *Client) *clientAssetStore {
	return &clientAssetStore{
		client: client,
		cache:  make(map[string]core.Asset),
	}
}

// GetAssetDecimals returns the decimals for an asset as stored in Clearnode.
func (s *clientAssetStore) GetAssetDecimals(asset string) (uint8, error) {
	// Check cache first
	if cached, ok := s.cache[asset]; ok {
		return cached.Decimals, nil
	}

	// Fetch from node
	assets, err := s.client.GetAssets(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch assets: %w", err)
	}

	// Update cache and find asset
	for _, a := range assets {
		s.cache[a.Symbol] = a
		if strings.EqualFold(a.Symbol, asset) {
			return a.Decimals, nil
		}
	}

	return 0, fmt.Errorf("asset %s not found", asset)
}

// GetTokenDecimals returns the decimals for a specific token on a blockchain.
func (s *clientAssetStore) GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error) {
	// Fetch all assets if cache is empty
	if len(s.cache) == 0 {
		assets, err := s.client.GetAssets(context.Background(), nil)
		if err != nil {
			return 0, fmt.Errorf("failed to fetch assets: %w", err)
		}
		for _, a := range assets {
			s.cache[a.Symbol] = a
		}
	}

	// Search through all assets for matching token
	tokenAddress = strings.ToLower(tokenAddress)
	for _, asset := range s.cache {
		for _, token := range asset.Tokens {
			if token.BlockchainID == blockchainID &&
				strings.EqualFold(token.Address, tokenAddress) {
				return token.Decimals, nil
			}
		}
	}

	return 0, fmt.Errorf("token %s on blockchain %d not found", tokenAddress, blockchainID)
}

// AssetExistsOnBlockchain checks if a specific asset is supported on a specific blockchain.
func (s *clientAssetStore) AssetExistsOnBlockchain(blockchainID uint64, asset string) (bool, error) {
	// 1. Search existing cache first (handling case-insensitivity)
	for _, a := range s.cache {
		if strings.EqualFold(a.Symbol, asset) {
			for _, token := range a.Tokens {
				if token.BlockchainID == blockchainID {
					return true, nil
				}
			}
			// Asset found in cache, but not on this chain
			return false, nil
		}
	}

	// 2. If not found in cache, fetch fresh data from node
	assets, err := s.client.GetAssets(context.Background(), nil)
	if err != nil {
		return false, fmt.Errorf("failed to fetch assets: %w", err)
	}

	// 3. Update cache and search again
	for _, a := range assets {
		s.cache[a.Symbol] = a
		if strings.EqualFold(a.Symbol, asset) {
			for _, token := range a.Tokens {
				if token.BlockchainID == blockchainID {
					return true, nil
				}
			}
			// Asset found after fetch, but not on this chain
			return false, nil
		}
	}

	// Asset symbol not found at all
	return false, nil
}

// NewClient creates a new Clearnode client with both high-level and low-level methods.
// This is the recommended constructor for most use cases.
//
// Parameters:
//   - wsURL: WebSocket URL of the Clearnode server (e.g., "wss://clearnode.example.com/ws")
//   - stateSigner: sign.Signer for signing channel states (use sign.NewEthereumMsgSigner)
//   - txSigner: sign.Signer for signing blockchain transactions (use sign.NewEthereumRawSigner)
//   - opts: Optional configuration (WithBlockchainRPC, WithHandshakeTimeout, etc.)
//
// Returns:
//   - Configured Client ready for operations
//   - Error if connection or initialization fails
//
// Example:
//
//	stateSigner, _ := sign.NewEthereumMsgSigner(privateKeyHex)
//	txSigner, _ := sign.NewEthereumRawSigner(privateKeyHex)
//	client, err := sdk.NewClient(
//	    "wss://clearnode.example.com/ws",
//	    stateSigner,
//	    txSigner,
//	    sdk.WithBlockchainRPC(80002, "https://polygon-amoy.alchemy.com/v2/KEY"),
//	)
func NewClient(wsURL string, stateSigner, txSigner sign.Signer, opts ...Option) (*Client, error) {
	// Build config starting with defaults
	config := DefaultConfig
	config.URL = wsURL

	// Apply user options
	for _, opt := range opts {
		opt(&config)
	}

	// Create WebSocket dialer with configuration
	dialerConfig := rpc.DefaultWebsocketDialerConfig
	dialerConfig.HandshakeTimeout = config.HandshakeTimeout
	dialerConfig.PingInterval = config.PingInterval

	dialer := rpc.NewWebsocketDialer(dialerConfig)
	rpcClient := rpc.NewClient(dialer)

	// Create client instance
	client := &Client{
		rpcDialer:         dialer,
		rpcClient:         rpcClient,
		config:            config,
		exitCh:            make(chan struct{}),
		blockchainClients: make(map[uint64]*evm.Client),
		homeBlockchains:   make(map[string]uint64),
		stateSigner:       stateSigner,
		txSigner:          txSigner,
	}

	// Create asset store
	client.assetStore = newClientAssetStore(client)

	// Error handler wrapper
	handleError := func(err error) {
		if config.ErrorHandler != nil {
			config.ErrorHandler(err)
		}
		close(client.exitCh)
	}

	// Establish connection
	err := rpcClient.Start(context.Background(), wsURL, handleError)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to clearnode: %w", err)
	}

	return client, nil
}

// SetHomeBlockchain configures the primary blockchain network for a specific asset.
// This is required for operations like Transfer which may trigger channel creation
// but do not accept a blockchain ID as a parameter.
//
// Validation:
//   - Checks if the asset is actually supported on the specified blockchain.
//   - Verifies that a home blockchain hasn't already been set for this asset.
//
// Constraints:
//   - This mapping is immutable once set for the client instance.
//   - To move an asset to a different blockchain, use the Migrate() method instead.
//
// Parameters:
//   - asset: The asset symbol (e.g., "usdc")
//   - blockchainId: The chain ID to associate with the asset (e.g., 80002)
//
// Example:
//
//	// Set USDC to settle on Polygon Amoy
//	if err := client.SetHomeBlockchain("usdc", 80002); err != nil {
//	    log.Fatal(err)
//	}
func (c *Client) SetHomeBlockchain(asset string, blockchainId uint64) error {
	blockchainID, homeBlockchainIsSet := c.homeBlockchains[asset]
	if homeBlockchainIsSet {
		return fmt.Errorf("home blockchain is already set for asset %s to %d, please use Migrate() if you want to change home blockchain", asset, blockchainID)
	}
	ok, err := c.assetStore.AssetExistsOnBlockchain(blockchainId, asset)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("asset %s not supported on blockchain %d", asset, blockchainId)
	}
	c.homeBlockchains[asset] = blockchainId
	return nil
}

// ============================================================================
// Connection & Lifecycle Methods
// ============================================================================

// Close cleanly shuts down the client connection.
// It's recommended to defer this call after creating the client.
//
// Example:
//
//	client, err := NewClient(...)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
func (c *Client) Close() error {
	// The dialer handles connection cleanup internally
	select {
	case <-c.exitCh:
		// Already closed
	default:
		close(c.exitCh)
	}
	return nil
}

// WaitCh returns a channel that closes when the connection is lost or closed.
// This is useful for monitoring connection health in long-running applications.
//
// Example:
//
//	go func() {
//	    <-client.WaitCh()
//	    log.Println("Connection closed")
//	}()
func (c *Client) WaitCh() <-chan struct{} {
	return c.exitCh
}

// ============================================================================
// Shared Helper Methods
// ============================================================================

// SignState signs a channel state by packing it, hashing it, and signing the hash.
// Returns the signature as a hex-encoded string (with 0x prefix).
//
// This is a low-level method exposed for advanced users who want to manually
// construct and sign states. Most users should use the high-level methods like
// Transfer, Deposit, and Withdraw instead.
func (c *Client) SignState(state *core.State) (string, error) {
	if state == nil {
		return "", fmt.Errorf("state cannot be nil")
	}

	// Pack the state into ABI-encoded bytes
	packedState, err := core.PackState(*state, c.assetStore)
	if err != nil {
		return "", fmt.Errorf("failed to pack state: %w", err)
	}

	// Sign the hash
	signature, err := c.stateSigner.Sign(packedState)
	if err != nil {
		return "", fmt.Errorf("failed to sign state hash: %w", err)
	}

	// Return hex-encoded signature with 0x prefix
	return hexutil.Encode(signature), nil
}

// GetUserAddress returns the Ethereum address associated with the signer.
// This is useful for identifying the current user's wallet address.
func (c *Client) GetUserAddress() string {
	return c.stateSigner.PublicKey().Address().String()
}

// signAndSubmitState is a helper that signs a state and submits it to the node.
// It returns the node's signature.
func (c *Client) signAndSubmitState(ctx context.Context, state *core.State) (string, error) {
	// Sign state
	sig, err := c.SignState(state)
	if err != nil {
		return "", fmt.Errorf("failed to sign state: %w", err)
	}
	state.UserSig = &sig

	// Submit to node
	nodeSig, err := c.submitState(ctx, *state)
	if err != nil {
		return "", fmt.Errorf("failed to submit state: %w", err)
	}

	// Update state with node signature
	state.NodeSig = &nodeSig

	return nodeSig, nil
}

// ============================================================================
// High-Level Operations (Blockchain Interaction)
// ============================================================================

// WithBlockchainRPC returns an Option that configures a blockchain RPC client for a specific chain.
// This is required for operations that interact with the blockchain (Deposit, Withdraw).
//
// Parameters:
//   - chainID: The blockchain network ID (e.g., 80002 for Polygon Amoy testnet)
//   - rpcURL: The RPC endpoint URL (e.g., "https://polygon-amoy.alchemy.com/v2/KEY")
//
// Example:
//
//	client, err := sdk.NewClient(
//	    wsURL,
//	    stateSigner,
//	    txSigner,
//	    sdk.WithBlockchainRPC(80002, "https://polygon-amoy.alchemy.com/v2/KEY"),
//	    sdk.WithBlockchainRPC(84532, "https://base-sepolia.alchemy.com/v2/KEY"),
//	)
func WithBlockchainRPC(chainID uint64, rpcURL string) Option {
	return func(c *Config) {
		// Store blockchain RPC config for later initialization
		if c.BlockchainRPCs == nil {
			c.BlockchainRPCs = make(map[uint64]string)
		}
		c.BlockchainRPCs[chainID] = rpcURL
	}
}

// initializeBlockchainClient initializes a blockchain client for a specific chain.
// This is called lazily when a blockchain operation is needed.
func (c *Client) initializeBlockchainClient(ctx context.Context, chainID uint64) error {
	// Check if already initialized
	if _, exists := c.blockchainClients[chainID]; exists {
		return nil
	}

	// Get RPC URL from config
	rpcURL, exists := c.config.BlockchainRPCs[chainID]
	if !exists {
		return fmt.Errorf("blockchain RPC not configured for chain %d (use WithBlockchainRPC)", chainID)
	}

	// Get contract address for this blockchain
	contractAddress, err := c.getContractAddress(ctx, chainID)
	if err != nil {
		return err
	}

	// Get node address
	nodeAddress, err := c.getNodeAddress(ctx)
	if err != nil {
		return err
	}

	// Connect to blockchain
	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to blockchain RPC: %w", err)
	}

	// Create blockchain client using the user's signer and node address
	evmClient, err := evm.NewClient(
		common.HexToAddress(contractAddress),
		ethClient,
		c.txSigner,
		chainID,
		nodeAddress,
		c.assetStore,
	)

	if err != nil {
		return fmt.Errorf("failed to create blockchain client: %w", err)
	}

	c.blockchainClients[chainID] = evmClient
	return nil
}

// generateNonce generates a random 8-byte nonce for channel creation.
func generateNonce() uint64 {
	nonceBig, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 64))
	if err != nil {
		// Fallback to timestamp-based nonce if crypto/rand fails
		return uint64(time.Now().UnixNano())
	}
	return nonceBig.Uint64()
}

// getTokenAddress looks up the token address for an asset on a specific blockchain.
func (c *Client) getTokenAddress(ctx context.Context, blockchainID uint64, asset string) (string, error) {
	assets, err := c.GetAssets(ctx, &blockchainID)
	if err != nil {
		return "", fmt.Errorf("failed to get assets: %w", err)
	}

	for _, a := range assets {
		if strings.EqualFold(a.Symbol, asset) {
			// Find token for this blockchain
			for _, token := range a.Tokens {
				if token.BlockchainID == blockchainID {
					return token.Address, nil
				}
			}
			return "", fmt.Errorf("asset %s not available on blockchain %d", asset, blockchainID)
		}
	}

	return "", fmt.Errorf("asset %s not found", asset)
}

// getContractAddress retrieves the contract address for a specific blockchain from node config.
func (c *Client) getContractAddress(ctx context.Context, blockchainID uint64) (string, error) {
	nodeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get node config: %w", err)
	}

	for _, bc := range nodeConfig.Blockchains {
		if bc.ID == blockchainID {
			return bc.ContractAddress, nil
		}
	}

	return "", fmt.Errorf("blockchain %d not found in node config", blockchainID)
}

// getNodeAddress retrieves the node's Ethereum address from the node config.
func (c *Client) getNodeAddress(ctx context.Context) (string, error) {
	nodeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get node config: %w", err)
	}
	return nodeConfig.NodeAddress, nil
}

// Deposit adds funds to the user's channel by depositing from the blockchain.
// This method handles two scenarios automatically:
//  1. If no channel exists: Creates a new channel with the initial deposit
//  2. If channel exists: Checkpoints the deposit to the existing channel
//
// Parameters:
//   - ctx: Context for the operation
//   - blockchainID: The blockchain network ID (e.g., 80002 for Polygon Amoy)
//   - asset: The asset symbol to deposit (e.g., "usdc")
//   - amount: The amount to deposit
//
// Returns:
//   - Transaction hash of the blockchain transaction
//   - Error if the operation fails
//
// Requirements:
//   - Blockchain RPC must be configured for the specified chain (use WithBlockchainRPC)
//   - User must have approved the token spend to the contract address
//   - User must have sufficient token balance in their wallet
//
// Example:
//
//	txHash, err := client.Deposit(ctx, 80002, "usdc", decimal.NewFromInt(100))
//	fmt.Printf("Deposit transaction: %s\n", txHash)
func (c *Client) Deposit(ctx context.Context, blockchainID uint64, asset string, amount decimal.Decimal) (string, error) {
	userWallet := c.GetUserAddress()

	// Initialize blockchain client if needed
	if err := c.initializeBlockchainClient(ctx, blockchainID); err != nil {
		return "", err
	}

	blockchainClient := c.blockchainClients[blockchainID]

	// Get node address
	nodeAddress, err := c.getNodeAddress(ctx)
	if err != nil {
		return "", err
	}

	// Get token address for this asset on this blockchain
	tokenAddress, err := c.getTokenAddress(ctx, blockchainID, asset)
	if err != nil {
		return "", err
	}

	// Try to get latest state to determine if channel exists
	state, err := c.GetLatestState(ctx, userWallet, asset, false)

	// Scenario A: Channel doesn't exist - create it
	if err != nil || state.HomeChannelID == nil {
		// Create channel definition
		channelDef := core.ChannelDefinition{
			Nonce:     generateNonce(),
			Challenge: 86400, // 1 day challenge period
		}

		if state == nil {
			state = core.NewVoidState(asset, userWallet)
		}
		newState := state.NextState()

		_, err := newState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
		if err != nil {
			return "", fmt.Errorf("failed to apply channel creation: %w", err)
		}

		// Apply deposit transition
		_, err = newState.ApplyHomeDepositTransition(amount)
		if err != nil {
			return "", fmt.Errorf("failed to apply deposit transition: %w", err)
		}

		// Sign state
		sig, err := c.SignState(newState)
		if err != nil {
			return "", fmt.Errorf("failed to sign state: %w", err)
		}
		newState.UserSig = &sig

		// Request channel creation from node
		nodeSig, err := c.requestChannelCreation(ctx, *newState, channelDef)
		if err != nil {
			return "", fmt.Errorf("failed to request channel creation: %w", err)
		}
		newState.NodeSig = &nodeSig

		// Create channel on blockchain
		txHash, err := blockchainClient.Create(channelDef, *newState)
		if err != nil {
			return "", fmt.Errorf("failed to create channel on blockchain: %w", err)
		}

		return txHash, nil
	}

	// Scenario B: Channel exists - checkpoint deposit
	// Create next state
	nextState := state.NextState()

	// Apply deposit transition
	_, err = nextState.ApplyHomeDepositTransition(amount)
	if err != nil {
		return "", fmt.Errorf("failed to apply deposit transition: %w", err)
	}

	// Sign and submit state to node
	_, err = c.signAndSubmitState(ctx, nextState)
	if err != nil {
		return "", err
	}

	// Checkpoint on blockchain
	txHash, err := blockchainClient.Checkpoint(*nextState, nil)
	if err != nil {
		return "", fmt.Errorf("failed to checkpoint on blockchain: %w", err)
	}

	return txHash, nil
}

// Withdraw removes funds from the user's channel and returns them to the blockchain wallet.
// This operation handles two scenarios automatically:
//  1. If no channel exists: Creates a new channel and executes the withdrawal in one transaction
//  2. If channel exists: Checkpoints the withdrawal to the existing channel
//
// Parameters:
//   - ctx: Context for the operation
//   - blockchainID: The blockchain network ID (e.g., 80002 for Polygon Amoy)
//   - asset: The asset symbol to withdraw (e.g., "usdc")
//   - amount: The amount to withdraw
//
// Returns:
//   - Transaction hash of the blockchain transaction
//   - Error if the operation fails
//
// Requirements:
//   - Channel must exist (user must have deposited first)
//   - Blockchain RPC must be configured for the specified chain (use WithBlockchainRPC)
//   - User must have sufficient balance in the channel
//
// Example:
//
//	txHash, err := client.Withdraw(ctx, 80002, "usdc", decimal.NewFromInt(25))
//	fmt.Printf("Withdrawal transaction: %s\n", txHash)
func (c *Client) Withdraw(ctx context.Context, blockchainID uint64, asset string, amount decimal.Decimal) (string, error) {
	userWallet := c.GetUserAddress()

	// Initialize blockchain client if needed
	if err := c.initializeBlockchainClient(ctx, blockchainID); err != nil {
		return "", err
	}

	blockchainClient := c.blockchainClients[blockchainID]

	// Get node address (Required for channel creation flow)
	nodeAddress, err := c.getNodeAddress(ctx)
	if err != nil {
		return "", err
	}

	// Get token address for this asset on this blockchain (Required for channel creation flow)
	tokenAddress, err := c.getTokenAddress(ctx, blockchainID, asset)
	if err != nil {
		return "", err
	}

	// Try to get latest state to determine if channel exists
	state, err := c.GetLatestState(ctx, userWallet, asset, false)

	// Channel doesn't exist - create it and withdraw
	if err != nil || state.HomeChannelID == nil {
		// Create channel definition
		channelDef := core.ChannelDefinition{
			Nonce:     generateNonce(),
			Challenge: 86400, // 1 day challenge period
		}

		if state == nil {
			state = core.NewVoidState(asset, userWallet)
		}
		newState := state.NextState()

		_, err := newState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
		if err != nil {
			return "", fmt.Errorf("failed to apply channel creation: %w", err)
		}

		// Apply withdrawal transition
		// Note: Ensure your core logic allows withdrawal on a fresh state
		// (assuming the smart contract handles the net balance check)
		_, err = newState.ApplyHomeWithdrawalTransition(amount)
		if err != nil {
			return "", fmt.Errorf("failed to apply withdrawal transition: %w", err)
		}

		// Sign state
		sig, err := c.SignState(newState)
		if err != nil {
			return "", fmt.Errorf("failed to sign state: %w", err)
		}
		newState.UserSig = &sig

		// Request channel creation from node
		nodeSig, err := c.requestChannelCreation(ctx, *newState, channelDef)
		if err != nil {
			return "", fmt.Errorf("failed to request channel creation: %w", err)
		}
		newState.NodeSig = &nodeSig

		// Create channel on blockchain (Smart Contract handles Creation + Withdrawal)
		txHash, err := blockchainClient.Create(channelDef, *newState)
		if err != nil {
			return "", fmt.Errorf("failed to create channel on blockchain: %w", err)
		}

		return txHash, nil
	}

	// Create next state
	nextState := state.NextState()

	// Apply withdrawal transition
	_, err = nextState.ApplyHomeWithdrawalTransition(amount)
	if err != nil {
		return "", fmt.Errorf("failed to apply withdrawal transition: %w", err)
	}

	// Sign and submit state to node
	_, err = c.signAndSubmitState(ctx, nextState)
	if err != nil {
		return "", err
	}

	// Checkpoint on blockchain
	txHash, err := blockchainClient.Checkpoint(*nextState, nil)
	if err != nil {
		return "", fmt.Errorf("failed to checkpoint withdrawal on blockchain: %w", err)
	}

	return txHash, nil
}

// Transfer sends funds from the user to another wallet address.
// This is the simplest operation as it doesn't require any blockchain interaction.
//
// Parameters:
//   - ctx: Context for the operation
//   - recipientWallet: The recipient's wallet address (e.g., "0x1234...")
//   - asset: The asset symbol to transfer (e.g., "usdc")
//   - amount: The amount to transfer
//
// Returns:
//   - Transaction ID for tracking
//   - Error if the operation fails
//
// Errors:
//   - Returns error if channel doesn't exist (user must deposit first)
//   - Returns error if insufficient balance
//   - Returns error if state submission fails
//
// Example:
//
//	txID, err := client.Transfer(ctx, "0xRecipient...", "usdc", decimal.NewFromInt(50))
//	fmt.Printf("Transfer successful: %s\n", txID)
func (c *Client) Transfer(ctx context.Context, recipientWallet string, asset string, amount decimal.Decimal) (string, error) {
	// Get sender's latest state
	senderWallet := c.GetUserAddress()
	state, err := c.GetLatestState(ctx, senderWallet, asset, false)
	if err != nil || state.HomeChannelID == nil {
		// Create channel definition
		channelDef := core.ChannelDefinition{
			Nonce:     generateNonce(),
			Challenge: 86400, // 1 day challenge period
		}

		if state == nil {
			state = core.NewVoidState(asset, senderWallet)
		}
		newState := state.NextState()

		blockchainID, ok := c.homeBlockchains[asset]
		if !ok {
			return "", fmt.Errorf("home blockchain not set for asset %s", asset)
		}

		// Get node address (Required for channel creation flow)
		nodeAddress, err := c.getNodeAddress(ctx)
		if err != nil {
			return "", err
		}

		// Get token address for this asset on this blockchain
		tokenAddress, err := c.getTokenAddress(ctx, blockchainID, asset)
		if err != nil {
			return "", err
		}

		// Initialize blockchain client if needed
		if err := c.initializeBlockchainClient(ctx, blockchainID); err != nil {
			return "", err
		}

		blockchainClient := c.blockchainClients[blockchainID]

		_, err = newState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
		if err != nil {
			return "", fmt.Errorf("failed to apply channel creation: %w", err)
		}

		// &{0x1afffc764a989ba7879c9d68b83eaa0c76320f7d72a0e92c7c5e36e387cb3b77
		// 	[{transfer_send 0x33281ad4bb1830a7536b7d287b81b8e8a9414944ca89d30745513c3460332c5a
		// 		 0x053aEAD7d3eebE4359300fDE849bCD9E77384989 0.1}]
		// 		  usdc 0xaB5670b44cb4A3B5535BD637cb600DA572148c98 0 2 0x140004e3850 <nil>
		// 		  {0x6E2C4707DA119425dF2c722E2695300154652f56 11155111 0.3 0 0 0.3} <nil> <nil> <nil>}

		// Apply withdrawal transition
		// Note: Ensure your core logic allows withdrawal on a fresh state
		// (assuming the smart contract handles the net balance check)
		_, err = newState.ApplyTransferSendTransition(recipientWallet, amount)
		if err != nil {
			return "", fmt.Errorf("failed to apply withdrawal transition: %w", err)
		}
		fmt.Println(newState)
		// Sign state
		sig, err := c.SignState(newState)
		if err != nil {
			return "", fmt.Errorf("failed to sign state: %w", err)
		}
		newState.UserSig = &sig

		// Request channel creation from node
		nodeSig, err := c.requestChannelCreation(ctx, *newState, channelDef)
		if err != nil {
			return "", fmt.Errorf("failed to request channel creation: %w", err)
		}
		newState.NodeSig = &nodeSig

		// Create channel on blockchain (Smart Contract handles Creation + Withdrawal)
		txHash, err := blockchainClient.Create(channelDef, *newState)
		if err != nil {
			return "", fmt.Errorf("failed to create channel on blockchain: %w", err)
		}

		return txHash, nil
	}

	// Create next state
	nextState := state.NextState()

	// Apply transfer send transition
	transition, err := nextState.ApplyTransferSendTransition(recipientWallet, amount)
	if err != nil {
		return "", fmt.Errorf("failed to apply transfer transition: %w", err)
	}

	// Sign and submit state
	_, err = c.signAndSubmitState(ctx, nextState)
	if err != nil {
		return "", err
	}

	// Return transaction ID from the transition
	return transition.TxID, nil
}
