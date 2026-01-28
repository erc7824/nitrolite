package sdk

import (
	"context"
	"fmt"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

// Client provides a simple, high-level interface for interacting with a Clearnode instance.
// It wraps the underlying RPC client to provide a cleaner API with SDK-specific types.
//
// Example usage:
//
//	client, err := sdk.NewClient("ws://localhost:7824/ws")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	ctx := context.Background()
//	if err := client.Ping(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	config, err := client.GetConfig(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Node: %s (v%s)\n", config.NodeAddress, config.NodeVersion)
type Client struct {
	rpcDialer rpc.Dialer
	rpcClient *rpc.Client
	config    Config
	exitCh    chan struct{}
}

// NewClient creates a new Clearnode client and establishes a connection to the server.
// The connection is established immediately and an error is returned if the connection fails.
//
// Parameters:
//   - wsURL: The WebSocket URL of the clearnode server (e.g., "ws://localhost:7824/ws")
//   - opts: Optional configuration options (e.g., WithHandshakeTimeout, WithPingInterval)
//
// Example:
//
//	client, err := sdk.NewClient(
//	    "ws://localhost:7824/ws",
//	    sdk.WithHandshakeTimeout(10*time.Second),
//	    sdk.WithErrorHandler(func(err error) {
//	        log.Printf("Connection error: %v", err)
//	    }),
//	)
func NewClient(wsURL string, opts ...Option) (*Client, error) {
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
		rpcDialer: dialer,
		rpcClient: rpcClient,
		config:    config,
		exitCh:    make(chan struct{}),
	}

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

// Ping checks connectivity to the clearnode server.
// This is useful for health checks and verifying the connection is active.
//
// Example:
//
//	ctx := context.Background()
//	if err := client.Ping(ctx); err != nil {
//	    log.Printf("Server is unreachable: %v", err)
//	}
func (c *Client) Ping(ctx context.Context) error {
	if err := c.rpcClient.NodeV1Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	return nil
}

// GetConfig retrieves the clearnode configuration including node identity and supported blockchains.
//
// Returns:
//   - NodeConfig containing the node address, version, and list of supported blockchain networks
//   - Error if the request fails
//
// Example:
//
//	config, err := client.GetConfig(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Node: %s (v%s)\n", config.NodeAddress, config.NodeVersion)
//	for _, bc := range config.Blockchains {
//	    fmt.Printf("  - %s (ID: %d)\n", bc.Name, bc.BlockchainID)
//	}
func (c *Client) GetConfig(ctx context.Context) (*core.NodeConfig, error) {
	resp, err := c.rpcClient.NodeV1GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	return transformNodeConfig(resp), nil
}

// GetBlockchains retrieves the list of supported blockchain networks.
// This is a convenience method that calls GetConfig and extracts the blockchains list.
//
// Returns:
//   - Slice of Blockchain containing name, chain ID, and contract address for each network
//   - Error if the request fails
//
// Example:
//
//	blockchains, err := client.GetBlockchains(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, bc := range blockchains {
//	    fmt.Printf("%s: %s\n", bc.Name, bc.ContractAddress)
//	}
func (c *Client) GetBlockchains(ctx context.Context) ([]core.Blockchain, error) {
	config, err := c.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get blockchains: %w", err)
	}
	return config.Blockchains, nil
}

// GetAssets retrieves all supported assets with optional blockchain filter.
//
// Parameters:
//   - blockchainID: Optional blockchain ID to filter assets (pass nil for all assets)
//
// Returns:
//   - Slice of Asset containing asset information and token implementations
//   - Error if the request fails
//
// Example:
//
//	// Get all assets
//	assets, err := client.GetAssets(ctx, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, asset := range assets {
//	    fmt.Printf("%s (%s): %d tokens\n", asset.Name, asset.Symbol, len(asset.Tokens))
//	}
//
//	// Get assets for specific blockchain
//	chainID := uint64(80002)
//	assets, err := client.GetAssets(ctx, &chainID)
func (c *Client) GetAssets(ctx context.Context, blockchainID *uint64) ([]core.Asset, error) {
	req := rpc.NodeV1GetAssetsRequest{
		BlockchainID: blockchainID,
	}
	resp, err := c.rpcClient.NodeV1GetAssets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}
	return transformAssets(resp.Assets), nil
}

// GetBalances retrieves the balance information for a user.
//
// Parameters:
//   - wallet: The user's wallet address
//
// Returns:
//   - Slice of Balance containing asset balances
//   - Error if the request fails
//
// Example:
//
//	balances, err := client.GetBalances(ctx, "0x1234567890abcdef...")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, balance := range balances {
//	    fmt.Printf("%s: %s\n", balance.Asset, balance.Amount)
//	}
func (c *Client) GetBalances(ctx context.Context, wallet string) ([]core.BalanceEntry, error) {
	req := rpc.UserV1GetBalancesRequest{
		Wallet: wallet,
	}
	resp, err := c.rpcClient.UserV1GetBalances(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get balances: %w", err)
	}
	return transformBalances(resp.Balances), nil
}

// GetChannelsOptions contains optional filters for GetChannels.
type GetChannelsOptions struct {
	// Status filters by channel status (e.g., "active", "closed")
	Status *string

	// Asset filters by asset symbol
	Asset *string

	// Pagination parameters
	Pagination *core.PaginationParams
}

// GetChannels retrieves all channels for a user with optional filtering.
//
// Parameters:
//   - wallet: The user's wallet address
//   - opts: Optional filters (pass nil for no filters)
//
// Returns:
//   - Slice of Channel
//   - core.PaginationMetadata with pagination information
//   - Error if the request fails
//
// Example:
//
//	channels, meta, err := client.GetChannels(ctx, "0x1234...", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Found %d channels (page %d of %d)\n", len(channels), meta.Page, meta.PageCount)
//
//	// With filters
//	status := "active"
//	limit := uint32(10)
//	opts := &sdk.GetChannelsOptions{
//	    Status: &status,
//	    Pagination: &sdk.PaginationParams{Limit: &limit},
//	}
//	channels, meta, err := client.GetChannels(ctx, "0x1234...", opts)
func (c *Client) GetChannels(ctx context.Context, wallet string, opts *GetChannelsOptions) ([]core.Channel, core.PaginationMetadata, error) {
	req := rpc.ChannelsV1GetChannelsRequest{
		Wallet: wallet,
	}
	if opts != nil {
		req.Status = opts.Status
		req.Asset = opts.Asset
		req.Pagination = transformPaginationParams(opts.Pagination)
	}
	resp, err := c.rpcClient.ChannelsV1GetChannels(ctx, req)
	if err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to get channels: %w", err)
	}
	return transformChannels(resp.Channels), transformPaginationMetadata(resp.Metadata), nil
}

// GetTransactionsOptions contains optional filters for GetTransactions.
type GetTransactionsOptions struct {
	// Asset filters by asset symbol
	Asset *string

	// Pagination parameters
	Pagination *core.PaginationParams
}

// GetTransactions retrieves transaction history for a user with optional filtering.
//
// Parameters:
//   - wallet: The user's wallet address
//   - opts: Optional filters (pass nil for no filters)
//
// Returns:
//   - Slice of Transaction
//   - core.PaginationMetadata with pagination information
//   - Error if the request fails
//
// Example:
//
//	txs, meta, err := client.GetTransactions(ctx, "0x1234...", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, tx := range txs {
//	    fmt.Printf("%s: %s → %s (%s %s)\n", tx.TxType, tx.FromAccount, tx.ToAccount, tx.Amount, tx.Asset)
//	}
//
//	// With pagination
//	limit := uint32(20)
//	opts := &sdk.GetTransactionsOptions{
//	    Pagination: &sdk.PaginationParams{Limit: &limit},
//	}
//	txs, meta, err := client.GetTransactions(ctx, "0x1234...", opts)
func (c *Client) GetTransactions(ctx context.Context, wallet string, opts *GetTransactionsOptions) ([]core.Transaction, core.PaginationMetadata, error) {
	req := rpc.UserV1GetTransactionsRequest{
		Wallet: wallet,
	}
	if opts != nil {
		req.Asset = opts.Asset
		req.Pagination = transformPaginationParams(opts.Pagination)
	}
	resp, err := c.rpcClient.UserV1GetTransactions(ctx, req)
	if err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to get transactions: %w", err)
	}
	return transformTransactions(resp.Transactions), transformPaginationMetadata(resp.Metadata), nil
}

// ============================================================================
// core.State Management Methods
// ============================================================================

// GetLatestState retrieves the latest state for a user's asset.
//
// Parameters:
//   - wallet: The user's wallet address
//   - asset: The asset symbol (e.g., "usdc")
//   - onlySigned: If true, returns only the latest signed state
//
// Returns:
//   - core.State containing all state information
//   - Error if the request fails
//
// Example:
//
//	state, err := client.GetLatestState(ctx, "0x1234...", "usdc", false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("State ID: %s, Version: %s\n", state.ID, state.Version)
//	fmt.Printf("Balance: %s\n", state.HomeLedger.UserBalance)
func (c *Client) GetLatestState(ctx context.Context, wallet, asset string, onlySigned bool) (*core.State, error) {
	req := rpc.ChannelsV1GetLatestStateRequest{
		Wallet:     wallet,
		Asset:      asset,
		OnlySigned: onlySigned,
	}
	resp, err := c.rpcClient.ChannelsV1GetLatestState(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest state: %w", err)
	}
	state, err := transformState(resp.State)
	if err != nil {
		return nil, fmt.Errorf("failed to transform state: %w", err)
	}
	return &state, nil
}

// GetStatesOptions contains optional filters for GetStates.
type GetStatesOptions struct {
	// Epoch filters by user epoch index
	Epoch *uint64

	// ChannelID filters by Home/Escrow Channel ID
	ChannelID *string

	// OnlySigned returns only signed states
	OnlySigned bool

	// Pagination parameters
	Pagination *core.PaginationParams
}

// GetStates retrieves state history for a user with optional filtering.
//
// Parameters:
//   - wallet: The user's wallet address
//   - asset: The asset symbol
//   - opts: Optional filters (pass nil for no filters)
//
// Returns:
//   - Slice of core.State
//   - core.PaginationMetadata with pagination information
//   - Error if the request fails
//
// Example:
//
//	states, meta, err := client.GetStates(ctx, "0x1234...", "usdc", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Found %d states\n", len(states))
//	for _, state := range states {
//	    fmt.Printf("  Version %s: Balance %s\n", state.Version, state.HomeLedger.UserBalance)
//	}
func (c *Client) GetStates(ctx context.Context, wallet, asset string, opts *GetStatesOptions) ([]core.State, core.PaginationMetadata, error) {
	req := rpc.ChannelsV1GetStatesRequest{
		Wallet: wallet,
		Asset:  asset,
	}
	if opts != nil {
		req.Epoch = opts.Epoch
		req.ChannelID = opts.ChannelID
		req.OnlySigned = opts.OnlySigned
		req.Pagination = transformPaginationParams(opts.Pagination)
	}
	resp, err := c.rpcClient.ChannelsV1GetStates(ctx, req)
	if err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to get states: %w", err)
	}
	states, err := transformStates(resp.States)
	if err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to transform states: %w", err)
	}
	return states, transformPaginationMetadata(resp.Metadata), nil
}

// SubmitState submits a signed state update to the node.
// The state must be properly signed by the user before submission.
//
// Parameters:
//   - state: The state to submit (must include valid signatures and transitions)
//
// Returns:
//   - Node's signature of the state
//   - Error if the request fails
//
// Example:
//
//	// Assuming you have a properly signed state
//	nodeSig, err := client.SubmitState(ctx, myState)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("State submitted, node signature: %s\n", nodeSig)
func (c *Client) SubmitState(ctx context.Context, state core.State) (string, error) {
	req := rpc.ChannelsV1SubmitStateRequest{
		State: transformStateToRPC(state),
	}
	resp, err := c.rpcClient.ChannelsV1SubmitState(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to submit state: %w", err)
	}
	return resp.Signature, nil
}

// RequestChannelCreation requests the node to sign a channel creation.
// This is typically the first step when creating a new payment channel.
//
// Parameters:
//   - state: The initial state for the channel
//   - channelDef: The channel definition with nonce and challenge period
//
// Returns:
//   - Node's signature for the channel creation
//   - Error if the request fails
//
// Example:
//
//	initialState := State{
//	    Asset: "usdc",
//	    UserWallet: "0x1234...",
//	    // ... other fields
//	}
//	channelDef := core.ChannelDefinition{
//	    Nonce: 1,
//	    Challenge: 3600, // 1 hour challenge period
//	}
//	nodeSig, err := client.RequestChannelCreation(ctx, initialState, channelDef)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Channel creation approved, signature: %s\n", nodeSig)
func (c *Client) RequestChannelCreation(ctx context.Context, state core.State, channelDef core.ChannelDefinition) (string, error) {
	req := rpc.ChannelsV1RequestCreationRequest{
		State:             transformStateToRPC(state),
		ChannelDefinition: transformChannelDefinitionToRPC(channelDef),
	}
	resp, err := c.rpcClient.ChannelsV1RequestCreation(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to request channel creation: %w", err)
	}
	return resp.Signature, nil
}

// ============================================================================
// App Session Methods
// ============================================================================

// GetAppSessionsOptions contains optional filters for GetAppSessions.
type GetAppSessionsOptions struct {
	// AppSessionID filters by application session ID
	AppSessionID *string

	// Participant filters by participant wallet address
	Participant *string

	// Status filters by status ("open" or "closed")
	Status *string

	// Pagination parameters
	Pagination *core.PaginationParams
}

// GetAppSessions retrieves application sessions with optional filtering.
//
// Parameters:
//   - opts: Optional filters (pass nil for no filters)
//
// Returns:
//   - Slice of AppSession
//   - core.PaginationMetadata with pagination information
//   - Error if the request fails
//
// Example:
//
//	sessions, meta, err := client.GetAppSessions(ctx, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, session := range sessions {
//	    fmt.Printf("Session %s: %s (%d participants)\n",
//	        session.AppSessionID, session.Status, len(session.Participants))
//	}
func (c *Client) GetAppSessions(ctx context.Context, opts *GetAppSessionsOptions) ([]app.AppSessionInfoV1, core.PaginationMetadata, error) {
	req := rpc.AppSessionsV1GetAppSessionsRequest{}
	if opts != nil {
		req.AppSessionID = opts.AppSessionID
		req.Participant = opts.Participant
		req.Status = opts.Status
		req.Pagination = transformPaginationParams(opts.Pagination)
	}
	resp, err := c.rpcClient.AppSessionsV1GetAppSessions(ctx, req)
	if err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to get app sessions: %w", err)
	}
	return transformAppSessions(resp.AppSessions), transformPaginationMetadata(resp.Metadata), nil
}

// GetAppDefinition retrieves the definition for a specific app session.
//
// Parameters:
//   - appSessionID: The application session ID
//
// Returns:
//   - app.AppDefinitionV1 with participants, quorum, and application info
//   - Error if the request fails
//
// Example:
//
//	def, err := client.GetAppDefinition(ctx, "session123")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("App: %s, Quorum: %d\n", def.Application, def.Quorum)
func (c *Client) GetAppDefinition(ctx context.Context, appSessionID string) (*app.AppDefinitionV1, error) {
	req := rpc.AppSessionsV1GetAppDefinitionRequest{
		AppSessionID: appSessionID,
	}
	resp, err := c.rpcClient.AppSessionsV1GetAppDefinition(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get app definition: %w", err)
	}
	def := transformAppDefinition(resp.Definition)
	return &def, nil
}

// CreateAppSession creates a new application session between participants.
//
// Parameters:
//   - definition: The app definition with participants, quorum, application ID
//   - sessionData: Optional JSON stringified session data
//   - quorumSigs: Participant signatures for the app session creation
//
// Returns:
//   - AppSessionID of the created session
//   - Initial version of the session
//   - Status of the session
//   - Error if the request fails
//
// Example:
//
//	def := sdk.AppDefinition{
//	    Application: "chess-v1",
//	    Participants: []sdk.AppParticipant{
//	        {WalletAddress: "0x1234...", SignatureWeight: 1},
//	        {WalletAddress: "0x5678...", SignatureWeight: 1},
//	    },
//	    Quorum: 2,
//	    Nonce: 1,
//	}
//	sessionID, version, status, err := client.CreateAppSession(ctx, def, "{}", []string{"sig1", "sig2"})
//	// version is returned as a string
func (c *Client) CreateAppSession(ctx context.Context, definition app.AppDefinitionV1, sessionData string, quorumSigs []string) (string, string, string, error) {
	req := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition:  transformAppDefinitionToRPC(definition),
		SessionData: sessionData,
		QuorumSigs:  quorumSigs,
	}
	resp, err := c.rpcClient.AppSessionsV1CreateAppSession(ctx, req)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create app session: %w", err)
	}
	return resp.AppSessionID, resp.Version, resp.Status, nil
}

// SubmitDepositState submits a deposit to an app session.
// This updates both the app session state and the user's channel state.
//
// Parameters:
//   - appStateUpdate: The app state update with deposit intent
//   - quorumSigs: Participant signatures for the app state update
//   - userState: The user's updated channel state
//
// Returns:
//   - Node's signature for the state
//   - Error if the request fails
//
// Example:
//
//	appUpdate := sdk.AppStateUpdate{
//	    AppSessionID: "session123",
//	    Intent: sdk.AppStateIntentDeposit,
//	    Version: 2,
//	    Allocations: []sdk.AppAllocation{...},
//	}
//	nodeSig, err := client.SubmitDepositState(ctx, appUpdate, []string{"sig1"}, userState)
func (c *Client) SubmitDepositState(ctx context.Context, appStateUpdate app.AppStateUpdateV1, quorumSigs []string, userState core.State) (string, error) {
	appUpdate := transformAppStateUpdateToRPC(appStateUpdate)

	req := rpc.AppSessionsV1SubmitDepositStateRequest{
		AppStateUpdate: appUpdate,
		QuorumSigs:     quorumSigs,
		UserState:      transformStateToRPC(userState),
	}
	resp, err := c.rpcClient.AppSessionsV1SubmitDepositState(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to submit deposit state: %w", err)
	}
	return resp.StateNodeSig, nil
}

// SubmitAppState submits an app session state update.
// This method handles operate, withdraw, and close intents.
// For deposits, use SubmitDepositState instead.
//
// Parameters:
//   - appStateUpdate: The app state update (intent: operate, withdraw, or close)
//   - quorumSigs: Participant signatures for the app state update
//
// Returns:
//   - Error if the request fails
//
// Example (operate):
//
//	appUpdate := sdk.AppStateUpdate{
//	    AppSessionID: "session123",
//	    Intent: sdk.AppStateIntentOperate,
//	    Version: 3,
//	    Allocations: []sdk.AppAllocation{...}, // Redistributed allocations
//	    SessionData: "{}",
//	}
//	err := client.SubmitAppState(ctx, appUpdate, []string{"sig1", "sig2"})
//
// Example (close):
//
//	appUpdate := sdk.AppStateUpdate{
//	    AppSessionID: "session123",
//	    Intent: sdk.AppStateIntentClose,
//	    Version: 5,
//	    Allocations: []sdk.AppAllocation{...}, // Final allocations
//	}
//	err := client.SubmitAppState(ctx, appUpdate, []string{"sig1", "sig2"})
func (c *Client) SubmitAppState(ctx context.Context, appStateUpdate app.AppStateUpdateV1, quorumSigs []string) error {
	appUpdate := transformAppStateUpdateToRPC(appStateUpdate)

	req := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: appUpdate,
		QuorumSigs:     quorumSigs,
	}
	_, err := c.rpcClient.AppSessionsV1SubmitAppState(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to submit app state: %w", err)
	}
	return nil
}

// RebalanceAppSessions rebalances multiple application sessions atomically.
//
// This method performs atomic rebalancing across multiple app sessions, ensuring
// that funds are redistributed consistently without the risk of partial updates.
//
// Parameters:
//   - signedUpdates: Slice of signed app state updates to apply atomically
//
// Returns:
//   - BatchID for tracking the rebalancing operation
//   - Error if the request fails
//
// Example:
//
//	updates := []sdk.SignedAppStateUpdate{
//	    {
//	        app.AppStateUpdateV1: sdk.AppStateUpdate{
//	            AppSessionID: "session1",
//	            Intent: sdk.AppStateIntentOperate,
//	            Version: 5,
//	            Allocations: []sdk.AppAllocation{
//	                {Participant: "0x1234...", Asset: "usdc", Amount: "100"},
//	            },
//	            SessionData: "{}",
//	        },
//	        QuorumSigs: []string{"sig1", "sig2"},
//	    },
//	    {
//	        app.AppStateUpdateV1: sdk.AppStateUpdate{
//	            AppSessionID: "session2",
//	            Intent: sdk.AppStateIntentOperate,
//	            Version: 3,
//	            Allocations: []sdk.AppAllocation{
//	                {Participant: "0x1234...", Asset: "usdc", Amount: "50"},
//	            },
//	            SessionData: "{}",
//	        },
//	        QuorumSigs: []string{"sig1", "sig2"},
//	    },
//	}
//	batchID, err := client.RebalanceAppSessions(ctx, updates)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Rebalance batch ID: %s\n", batchID)
func (c *Client) RebalanceAppSessions(ctx context.Context, signedUpdates []app.SignedAppStateUpdateV1) (string, error) {
	// Transform SDK types to RPC types
	rpcUpdates := make([]rpc.SignedAppStateUpdateV1, 0, len(signedUpdates))
	for _, update := range signedUpdates {
		rpcUpdate := transformSignedAppStateUpdateToRPC(update)
		rpcUpdates = append(rpcUpdates, rpcUpdate)
	}

	req := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: rpcUpdates,
	}

	resp, err := c.rpcClient.AppSessionsV1RebalanceAppSessions(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to rebalance app sessions: %w", err)
	}

	return resp.BatchID, nil
}

// Close cleanly shuts down the client connection.
// It's recommended to defer this call after creating the client.
//
// Example:
//
//	client, err := NewClient("ws://localhost:7824/ws")
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
//	client, err := NewClient("ws://localhost:7824/ws")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	go func() {
//	    <-client.WaitCh()
//	    log.Println("Connection closed")
//	}()
func (c *Client) WaitCh() <-chan struct{} {
	return c.exitCh
}
