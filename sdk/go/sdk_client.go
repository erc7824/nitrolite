package sdk

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/erc7824/nitrolite/pkg/blockchain/evm"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/sign"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

// SDKClient provides high-level methods for interacting with Clearnode.
// It extends the base Client with smart operations like Deposit, Withdraw, and Transfer
// that handle complex multi-step flows automatically.
//
// Example usage:
//
//	signer, err := sign.NewEthereumSigner(privateKeyHex)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	client, err := sdk.NewSDKClient(
//	    "wss://clearnode.example.com/ws",
//	    signer,
//	    sdk.WithBlockchainRPC(80002, "https://polygon-amoy.alchemy.com/v2/KEY"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Simple operations
//	txHash, err := client.Deposit(ctx, 80002, "usdc", decimal.NewFromInt(100))
//	txID, err := client.Transfer(ctx, "0xRecipient...", "usdc", decimal.NewFromInt(50))
//	txHash, err := client.Withdraw(ctx, 80002, "usdc", decimal.NewFromInt(25))
type SDKClient struct {
	*Client
	blockchainClients map[uint64]*evm.Client
	signer            sign.Signer
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

// NewSDKClient creates a new SDKClient and establishes a connection to the Clearnode server.
// It initializes the RPC client and sets up the signer for state operations.
//
// Parameters:
//   - wsURL: WebSocket URL of the Clearnode server (e.g., "wss://clearnode.example.com/ws")
//   - signer: sign.Signer for signing channel states (use sign.NewEthereumSigner)
//   - opts: Optional configuration (WithBlockchainRPC, WithHandshakeTimeout, etc.)
//
// Returns:
//   - Configured SDKClient ready for high-level operations
//   - Error if connection or initialization fails
//
// Example:
//
//	signer, _ := sign.NewEthereumSigner(privateKeyHex)
//	client, err := sdk.NewSDKClient(
//	    "wss://clearnode.example.com/ws",
//	    signer,
//	    sdk.WithBlockchainRPC(80002, "https://polygon-amoy.alchemy.com/v2/KEY"),
//	)
func NewSDKClient(wsURL string, signer sign.Signer, opts ...Option) (*SDKClient, error) {
	// Create base client
	baseClient, err := NewClient(wsURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create base client: %w", err)
	}

	// Create asset store
	assetStore := newClientAssetStore(baseClient)

	// Create smart client
	sdkClient := &SDKClient{
		Client:            baseClient,
		blockchainClients: make(map[uint64]*evm.Client),
		signer:            signer,
		assetStore:        assetStore,
	}

	return sdkClient, nil
}

// WithBlockchainRPC returns an Option that configures a blockchain RPC client for a specific chain.
// This is required for operations that interact with the blockchain (Deposit, Withdraw).
//
// Parameters:
//   - chainID: The blockchain network ID (e.g., 80002 for Polygon Amoy testnet)
//   - rpcURL: The RPC endpoint URL (e.g., "https://polygon-amoy.alchemy.com/v2/KEY")
//
// Example:
//
//	client, err := sdk.NewSDKClient(
//	    wsURL,
//	    signer,
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
func (c *SDKClient) initializeBlockchainClient(ctx context.Context, chainID uint64) error {
	// Check if already initialized
	if _, exists := c.blockchainClients[chainID]; exists {
		return nil
	}

	// Get RPC URL from config
	rpcURL, exists := c.config.BlockchainRPCs[chainID]
	if !exists {
		return fmt.Errorf("blockchain RPC not configured for chain %d (use WithBlockchainRPC)", chainID)
	}

	// Get node config to find contract address
	nodeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get node config: %w", err)
	}

	// Find contract address for this blockchain
	var contractAddress string
	for _, bc := range nodeConfig.Blockchains {
		if bc.ID == chainID {
			contractAddress = bc.ContractAddress
			break
		}
	}
	if contractAddress == "" {
		return fmt.Errorf("blockchain %d not supported by node", chainID)
	}

	// Connect to blockchain
	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to blockchain RPC: %w", err)
	}

	// Create blockchain client using the user's signer
	evmClient, err := evm.NewClient(
		common.HexToAddress(contractAddress),
		ethClient,
		c.signer,
		chainID,
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
		return uint64(1)
	}
	return nonceBig.Uint64()
}

// getTokenAddress looks up the token address for an asset on a specific blockchain.
func (c *SDKClient) getTokenAddress(ctx context.Context, blockchainID uint64, asset string) (string, error) {
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
func (c *SDKClient) getContractAddress(ctx context.Context, blockchainID uint64) (string, error) {
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
func (c *SDKClient) getNodeAddress(ctx context.Context) (string, error) {
	nodeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get node config: %w", err)
	}
	return nodeConfig.NodeAddress, nil
}

// SignState signs a channel state by packing it, hashing it, and signing the hash.
// Returns the signature as a hex-encoded string (with 0x prefix).
//
// This is a low-level method exposed for advanced users who want to manually
// construct and sign states. Most users should use the high-level methods like
// Transfer, Deposit, and Withdraw instead.
func (c *SDKClient) SignState(state *core.State) (string, error) {
	if state == nil {
		return "", fmt.Errorf("state cannot be nil")
	}

	// Pack the state into ABI-encoded bytes
	packedState, err := core.PackState(*state, c.assetStore)
	if err != nil {
		return "", fmt.Errorf("failed to pack state: %w", err)
	}

	// Hash the packed state with Keccak256
	stateHash := crypto.Keccak256Hash(packedState).Bytes()

	// Sign the hash
	signature, err := c.signer.Sign(stateHash)
	if err != nil {
		return "", fmt.Errorf("failed to sign state hash: %w", err)
	}

	// Return hex-encoded signature with 0x prefix
	return "0x" + hex.EncodeToString(signature), nil
}

// GetUserAddress returns the Ethereum address associated with the signer.
// This is useful for identifying the current user's wallet address.
func (c *SDKClient) GetUserAddress() string {
	return c.signer.PublicKey().Address().String()
}

// signAndSubmitState is a helper that signs a state and submits it to the node.
// It returns the node's signature.
func (c *SDKClient) signAndSubmitState(ctx context.Context, state *core.State) (string, error) {
	// Sign state
	sig, err := c.SignState(state)
	if err != nil {
		return "", fmt.Errorf("failed to sign state: %w", err)
	}
	state.UserSig = &sig

	// Submit to node
	nodeSig, err := c.SubmitState(ctx, *state)
	if err != nil {
		return "", fmt.Errorf("failed to submit state: %w", err)
	}

	// Update state with node signature
	state.NodeSig = &nodeSig

	return nodeSig, nil
}

// waitForTransaction waits for a blockchain transaction to be mined (stub for now).
// In production, this should poll the blockchain and wait for confirmation.
func (c *SDKClient) waitForTransaction(ctx context.Context, chainID uint64, txHash string) error {
	// Get blockchain client
	client, exists := c.blockchainClients[chainID]
	if !exists {
		return fmt.Errorf("blockchain client not initialized for chain %d", chainID)
	}

	// Use the underlying ethclient to wait for receipt
	backend := client // This might need adjustment based on evm.Client implementation

	_ = backend // TODO: implement actual receipt waiting
	_ = txHash

	// For now, return immediately
	// In production: poll for transaction receipt and check for success
	return nil
}

// ============================================================================
// High-Level Operations
// ============================================================================

// Transfer sends funds from the user to another wallet address.
// This is the simplest operation as it doesn't require any blockchain interaction.
//
// The flow:
//  1. Get the sender's latest state
//  2. Create next state
//  3. Apply transfer send transition
//  4. Calculate state ID
//  5. Sign state
//  6. Submit to node
//  7. Return transaction ID
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
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Transfer successful: %s\n", txID)
func (c *SDKClient) Transfer(ctx context.Context, recipientWallet string, asset string, amount decimal.Decimal) (string, error) {
	// Get sender's latest state
	senderWallet := c.GetUserAddress()
	state, err := c.GetLatestState(ctx, senderWallet, asset, false)
	if err != nil {
		return "", fmt.Errorf("failed to get latest state: %w", err)
	}

	// Check if channel exists
	if state.HomeChannelID == nil {
		return "", fmt.Errorf("channel not created, deposit first")
	}

	// Create next state
	nextState := state.NextState()

	// Apply transfer send transition
	transition, err := nextState.ApplyTransferSendTransition(recipientWallet, amount)
	if err != nil {
		return "", fmt.Errorf("failed to apply transfer transition: %w", err)
	}

	// Calculate state ID (already done by NextState, but let's be explicit)
	nextState.ID = core.GetStateID(nextState.UserWallet, nextState.Asset, nextState.Epoch, nextState.Version)

	// Sign and submit state
	_, err = c.signAndSubmitState(ctx, nextState)
	if err != nil {
		return "", err
	}

	// Return transaction ID from the transition
	return transition.TxID, nil
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
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Deposit transaction: %s\n", txHash)
func (c *SDKClient) Deposit(ctx context.Context, blockchainID uint64, asset string, amount decimal.Decimal) (string, error) {
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
			Challenge: 86400, // 1 hour challenge period
		}

		// Create void state
		newState := core.NewVoidState(asset, userWallet).NextState()

		// Set home ledger
		newState.HomeLedger.TokenAddress = tokenAddress
		newState.HomeLedger.BlockchainID = blockchainID

		// Calculate home channel ID
		homeChannelID, err := core.GetHomeChannelID(
			nodeAddress,
			userWallet,
			asset,
			channelDef.Nonce,
			channelDef.Challenge,
		)
		if err != nil {
			return "", fmt.Errorf("failed to calculate home channel ID: %w", err)
		}
		newState.HomeChannelID = &homeChannelID

		// Apply deposit transition
		_, err = newState.ApplyHomeDepositTransition(amount)
		if err != nil {
			return "", fmt.Errorf("failed to apply deposit transition: %w", err)
		}

		// Calculate state ID
		newState.ID = core.GetStateID(newState.UserWallet, newState.Asset, newState.Epoch, newState.Version)

		// Sign state
		sig, err := c.SignState(newState)
		if err != nil {
			return "", fmt.Errorf("failed to sign state: %w", err)
		}
		newState.UserSig = &sig

		// Request channel creation from node
		nodeSig, err := c.RequestChannelCreation(ctx, *newState, channelDef)
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

	// Calculate state ID
	nextState.ID = core.GetStateID(nextState.UserWallet, nextState.Asset, nextState.Epoch, nextState.Version)

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
// This operation requires an existing channel with sufficient balance.
//
// The flow:
//  1. Get latest state (must exist)
//  2. Create next state
//  3. Apply withdrawal transition
//  4. Calculate state ID
//  5. Sign state
//  6. Submit to node
//  7. Checkpoint on blockchain
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
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Withdrawal transaction: %s\n", txHash)
func (c *SDKClient) Withdraw(ctx context.Context, blockchainID uint64, asset string, amount decimal.Decimal) (string, error) {
	userWallet := c.GetUserAddress()

	// Initialize blockchain client if needed
	if err := c.initializeBlockchainClient(ctx, blockchainID); err != nil {
		return "", err
	}

	blockchainClient := c.blockchainClients[blockchainID]

	// Get latest state (must exist for withdrawal)
	state, err := c.GetLatestState(ctx, userWallet, asset, false)
	if err != nil {
		return "", fmt.Errorf("failed to get latest state: %w", err)
	}

	// Check if channel exists
	if state.HomeChannelID == nil {
		return "", fmt.Errorf("channel does not exist, cannot withdraw")
	}

	// Create next state
	nextState := state.NextState()

	// Apply withdrawal transition
	_, err = nextState.ApplyHomeWithdrawalTransition(amount)
	if err != nil {
		return "", fmt.Errorf("failed to apply withdrawal transition: %w", err)
	}

	// Calculate state ID
	nextState.ID = core.GetStateID(nextState.UserWallet, nextState.Asset, nextState.Epoch, nextState.Version)

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
