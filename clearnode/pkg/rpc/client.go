// Package rpc provides a high-level client for interacting with the ClearNode RPC server.
//
// The Client type wraps a Dialer to provide convenient methods for all RPC operations,
// including authentication, channel management, application sessions, and event handling.
package rpc

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/google/uuid"

	"github.com/erc7824/nitrolite/clearnode/pkg/log"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

// Client provides a high-level interface for interacting with the ClearNode RPC server.
// It wraps a Dialer to provide convenient methods for all RPC operations and manages
// event handlers for asynchronous notifications from the server.
//
// The Client is safe for concurrent use and supports:
//   - All public RPC methods (no authentication required)
//   - Authentication via wallet signature or JWT
//   - Authenticated RPC methods (require prior authentication)
//   - Event handling for balance updates, channel updates, and transfers
//
// Example usage:
//
//	dialer := rpc.NewWebsocketDialer(rpc.DefaultWebsocketDialerConfig)
//	client := rpc.NewClient(dialer)
//
//	// Connect to the server
//	go dialer.Dial(ctx, "wss://server.example.com/ws", handleError)
//
//	// Set up event handlers
//	client.HandleBalanceUpdateEvent(func(ctx context.Context, notif BalanceUpdateNotification, sigs []sign.Signature) {
//	    fmt.Printf("Balance updated: %+v\n", notif.BalanceUpdates)
//	})
//
//	// Start listening for events
//	go client.ListenEvents(ctx, handleError)
//
//	// Make RPC calls
//	config, _, err := client.GetConfig(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
type Client struct {
	dialer        Dialer
	eventHandlers map[Event]any
	mu            sync.RWMutex // protects eventHandlers
}

// NewClient creates a new RPC client using the provided dialer.
// The dialer must be connected before making RPC calls.
//
// Example:
//
//	dialer := rpc.NewWebsocketDialer(rpc.DefaultWebsocketDialerConfig)
//	client := rpc.NewClient(dialer)
func NewClient(dialer Dialer) *Client {
	return &Client{
		dialer:        dialer,
		eventHandlers: make(map[Event]any),
	}
}

// Start establishes a connection to the RPC server and begins listening for events.
// This method combines connection establishment and event handling in a single call,
// simplifying the client initialization process.
//
// The method will:
// 1. Establish a WebSocket connection to the specified URL
// 2. Start listening for server events in the background
// 3. Return immediately after successful connection (non-blocking)
//
// Parameters:
//   - ctx: Context for the connection lifetime (canceling stops the connection)
//   - url: WebSocket URL to connect to (e.g., "wss://server.example.com/ws")
//   - handleClosure: Callback invoked when the connection closes (with error if any)
//
// Returns an error if the initial connection fails (e.g., invalid URL, network error).
// After a successful return, the connection runs in the background until the context
// is canceled or a connection error occurs.
//
// Example:
//
//	client := rpc.NewClient(dialer)
//
//	// Set up event handlers before starting
//	client.HandleBalanceUpdateEvent(handleBalanceUpdate)
//
//	// Start the client
//	err := client.Start(ctx, "wss://server.example.com/ws", func(err error) {
//	    if err != nil {
//	        log.Error("Connection closed", "error", err)
//	    }
//	})
//	if err != nil {
//	    log.Fatal("Failed to start client", "error", err)
//	}
//
//	// Now you can make RPC calls
//	config, _, err := client.GetConfig(ctx)
func (c *Client) Start(ctx context.Context, url string, handleClosure func(err error)) error {
	parentCtx, cancel := context.WithCancel(ctx)
	childHandleClosure := func(err error) {
		cancel()
		handleClosure(err)
	}

	if err := c.dialer.Dial(parentCtx, url, childHandleClosure); err != nil {
		return err
	}

	go c.listenEvents(parentCtx)

	return nil
}

// BalanceUpdateEventHandler processes balance update notifications from the server.
// These notifications are sent when account balances change due to transfers,
// channel operations, or application session updates.
type BalanceUpdateEventHandler func(ctx context.Context, notif BalanceUpdateNotification, resSig []sign.Signature)

// ChannelUpdateEventHandler processes channel update notifications from the server.
// These notifications are sent when a channel's state changes, including creation,
// resizing, closure, or challenge events.
type ChannelUpdateEventHandler func(ctx context.Context, notif ChannelUpdateNotification, resSig []sign.Signature)

// AppSessionUpdateEventHandler processes application session update notifications from the server.
// These notifications are sent when an application session's state changes.
type AppSessionUpdateEventHandler func(ctx context.Context, notif AppSessionUpdateNotification, resSig []sign.Signature)

// TransferEventHandler processes transfer notifications from the server.
// These notifications are sent when transfers affect the authenticated user's account,
// including both incoming and outgoing transfers.
type TransferEventHandler func(ctx context.Context, notif TransferNotification, resSig []sign.Signature)

// listenEvents is an internal method that listens for asynchronous events from the server.
// This method is automatically started by the Start() method and runs in a background goroutine.
// It blocks until the context is cancelled, continuously reading from the dialer's event channel
// and dispatching events to registered handlers based on their type.
//
// Events are received through a separate channel from RPC responses, allowing
// the server to push notifications without client requests. When an event is received,
// it's routed to the appropriate handler (balance update, channel update, or transfer).
//
// This method is not intended to be called directly by users. Event handling is
// automatically managed when calling Start().
func (c *Client) listenEvents(ctx context.Context) {
	logger := log.FromContext(ctx)
	eventCh := c.dialer.EventCh()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-eventCh:
			if event == nil {
				continue
			}

			switch event.Res.Method {
			case BalanceUpdateEvent.String():
				c.handleBalanceUpdateEvent(ctx, event)
			case ChannelUpdateEvent.String():
				c.handleChannelUpdateEvent(ctx, event)
			case AppSessionUpdateEvent.String():
				c.handleAppSessionUpdateEvent(ctx, event)
			case TransferEvent.String():
				c.handleTransferEvent(ctx, event)
			default:
				logger.Warn("unknown event received", "method", event.Res.Method)
			}
		}
	}
}

// Ping sends a ping request to the server to check connectivity and liveness.
// Returns the response signatures if successful.
//
// This method is useful for:
//   - Testing the connection is alive
//   - Measuring round-trip latency
//   - Keeping the connection active
//
// Example:
//
//	sigs, err := client.Ping(ctx)
//	if err != nil {
//	    log.Error("Ping failed", "error", err)
//	}
func (c *Client) Ping(ctx context.Context) ([]sign.Signature, error) {
	var resSig []sign.Signature
	res, err := c.call(ctx, PingMethod, nil)
	if err != nil {
		return resSig, err
	}
	resSig = res.Sig

	if res.Res.Method != string(PongMethod) {
		return resSig, fmt.Errorf("unexpected response method: %s", res.Res.Method)
	}

	return resSig, nil
}

// GetConfig retrieves the server's configuration including supported networks.
// This is typically the first call made by clients to discover available chains
// and the server's wallet address.
//
// No authentication required.
//
// Returns:
//   - GetConfigResponse containing server address and network configurations
//   - Response signatures for verification
//   - Error if the request fails
//
// Example:
//
//	config, sigs, err := client.GetConfig(ctx)
//	if err != nil {
//	    log.Fatal("Failed to get config", "error", err)
//	}
//	for _, network := range config.Networks {
//	    fmt.Printf("Chain %d: Custody=%s\n", network.ChainID, network.CustodyAddress)
//	}
func (c *Client) GetConfig(ctx context.Context) (GetConfigResponse, []sign.Signature, error) {
	var resParams GetConfigResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetConfigMethod, nil)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, resSig, nil
}

// GetAssets retrieves the list of supported assets/tokens from the server.
// Assets can be filtered by chain ID to get tokens for a specific network.
//
// No authentication required.
//
// Parameters:
//   - reqParams: Filter options (optional ChainID filter)
//
// Returns:
//   - GetAssetsResponse containing the list of supported assets
//   - Response signatures for verification
//   - Error if the request fails
//
// Example:
//
//	// Get all assets
//	assets, _, err := client.GetAssets(ctx, GetAssetsRequest{})
//
//	// Get assets for a specific chain
//	ethChainID := uint32(1)
//	ethAssets, _, err := client.GetAssets(ctx, GetAssetsRequest{
//	    ChainID: &ethChainID,
//	})
func (c *Client) GetAssets(ctx context.Context, reqParams GetAssetsRequest) (GetAssetsResponse, []sign.Signature, error) {
	var resParams GetAssetsResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetAssetsMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, resSig, nil
}

// GetAppDefinition retrieves the protocol definition for a specific application session.
// This includes the participants, consensus rules, and protocol parameters.
//
// No authentication required.
//
// Parameters:
//   - reqParams: Contains the AppSessionID to query
//
// Returns:
//   - GetAppDefinitionResponse with the application protocol details
//   - Response signatures for verification
//   - Error if the request fails or session not found
//
// Example:
//
//	def, _, err := client.GetAppDefinition(ctx, GetAppDefinitionRequest{
//	    AppSessionID: "app123",
//	})
//	if err != nil {
//	    log.Error("Failed to get app definition", "error", err)
//	}
//	fmt.Printf("Protocol: %s, Participants: %v\n", def.Protocol, def.ParticipantWallets)
func (c *Client) GetAppDefinition(ctx context.Context, reqParams GetAppDefinitionRequest) (GetAppDefinitionResponse, []sign.Signature, error) {
	var resParams GetAppDefinitionResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetAppDefinitionMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetAppSessions retrieves a list of application sessions with optional filters.
// Sessions can be filtered by participant wallet address and status.
//
// No authentication required.
//
// Parameters:
//   - reqParams: Filter and pagination options:
//   - Participant: Filter by wallet address (optional)
//   - Status: Filter by session state (e.g., "open", "closed") (optional)
//   - ListOptions: Pagination (offset, limit) and sorting
//
// Returns:
//   - GetAppSessionsResponse containing the filtered list of sessions
//   - Response signatures for verification
//   - Error if the request fails
//
// Example:
//
//	// Get all open sessions for a specific participant
//	sessions, _, err := client.GetAppSessions(ctx, GetAppSessionsRequest{
//	    Participant: "0x1234...",
//	    Status: "open",
//	    ListOptions: ListOptions{Limit: 10},
//	})
func (c *Client) GetAppSessions(ctx context.Context, reqParams GetAppSessionsRequest) (GetAppSessionsResponse, []sign.Signature, error) {
	var resParams GetAppSessionsResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetAppSessionsMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetChannels retrieves a list of payment channels with optional filters.
// Channels can be filtered by participant wallet address and status.
//
// No authentication required.
//
// Parameters:
//   - reqParams: Filter and pagination options:
//   - Participant: Filter by wallet address (optional)
//   - Status: Filter by channel state (e.g., "open", "closed", "challenged") (optional)
//   - ListOptions: Pagination (offset, limit) and sorting
//
// Returns:
//   - GetChannelsResponse containing the filtered list of channels
//   - Response signatures for verification
//   - Error if the request fails
//
// Example:
//
//	// Get all open channels for a user
//	channels, _, err := client.GetChannels(ctx, GetChannelsRequest{
//	    Participant: userWallet,
//	    Status: "open",
//	})
//	for _, ch := range channels.Channels {
//	    fmt.Printf("Channel %s: %s on chain %d\n", ch.ChannelID, ch.Status, ch.ChainID)
//	}
func (c *Client) GetChannels(ctx context.Context, reqParams GetChannelsRequest) (GetChannelsResponse, []sign.Signature, error) {
	var resParams GetChannelsResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetChannelsMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetLedgerEntries retrieves double-entry bookkeeping entries from the ledger.
// Entries can be filtered by account ID, asset, and wallet address.
//
// No authentication required.
//
// Parameters:
//   - reqParams: Filter and pagination options:
//   - AccountID: Filter by account identifier (optional)
//   - Asset: Filter by token/asset symbol (optional)
//   - Wallet: Filter by participant wallet address (optional)
//   - ListOptions: Pagination (offset, limit) and sorting
//
// Returns:
//   - GetLedgerEntriesResponse containing the filtered ledger entries
//   - Response signatures for verification
//   - Error if the request fails
//
// Example:
//
//	entries, _, err := client.GetLedgerEntries(ctx, GetLedgerEntriesRequest{
//	    Asset: "usdc",
//	    Wallet: userWallet,
//	    ListOptions: ListOptions{Limit: 50},
//	})
func (c *Client) GetLedgerEntries(ctx context.Context, reqParams GetLedgerEntriesRequest) (GetLedgerEntriesResponse, []sign.Signature, error) {
	var resParams GetLedgerEntriesResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetLedgerEntriesMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetLedgerTransactions retrieves ledger transactions (transfers between accounts).
// Transactions can be filtered by account ID, asset, and transaction type.
//
// No authentication required.
//
// Parameters:
//   - reqParams: Filter and pagination options:
//   - AccountID: Filter by account (optional)
//   - Asset: Filter by token/asset symbol (optional)
//   - TxType: Filter by transaction type (optional)
//   - ListOptions: Pagination (offset, limit) and sorting
//
// Returns:
//   - GetLedgerTransactionsResponse containing the filtered transactions
//   - Response signatures for verification
//   - Error if the request fails
//
// Example:
//
//	// Get recent USDC transactions
//	txns, _, err := client.GetLedgerTransactions(ctx, GetLedgerTransactionsRequest{
//	    Asset: "usdc",
//	    ListOptions: ListOptions{
//	        Limit: 20,
//	        Sort: &SortTypeDescending,
//	    },
//	})
//	for _, tx := range txns.LedgerTransactions {
//	    fmt.Printf("%s: %s from %s to %s\n", tx.TxType, tx.Amount, tx.FromAccount, tx.ToAccount)
//	}
func (c *Client) GetLedgerTransactions(ctx context.Context, reqParams GetLedgerTransactionsRequest) (GetLedgerTransactionsResponse, []sign.Signature, error) {
	var resParams GetLedgerTransactionsResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetLedgerTransactionsMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// AuthWithSig performs wallet-based authentication using a signature.
// This method handles the complete authentication flow:
//  1. Sends an auth request to get a challenge
//  2. Signs the challenge using the provided signer
//  3. Verifies the signature and receives a JWT token
//
// The JWT token returned should be stored and used for subsequent authenticated calls.
//
// Parameters:
//   - reqParams: Authentication request containing:
//   - Address: Main wallet address requesting authentication (cold wallet)
//   - SessionKey: Address of a different key that will be used for signing during this session (hot wallet)
//   - Application: Name of the application
//   - Allowances: Spending limits for the session
//   - ExpiresAt: When the authentication expires (Unix timestamp)
//   - Scope: Permission scope (e.g., "trade", "view", or empty)
//   - signer: Signer interface to sign the challenge (should correspond to Address, not SessionKey)
//
// Returns:
//   - AuthSigVerifyResponse containing JWT token and success status
//   - Response signatures for verification
//   - Error if authentication fails
//
// Example:
//
//	walletSigner, _ := sign.NewEthereumSigner(walletPrivateKey)     // Main wallet
//	sessionSigner, _ := sign.NewEthereumSigner(sessionPrivateKey)   // Session key
//
//	authReq := AuthRequestRequest{
//	    Address:            walletSigner.PublicKey().Address().String(),   // Main wallet
//	    SessionKey:         sessionSigner.PublicKey().Address().String(),  // Different key for session
//	    Application:            "MyDApp",
//	    Allowances:         []Allowance{{Asset: "usdc", Amount: "1000"}},
//	}
//
//	// Sign with main wallet, but SessionKey will be used for subsequent operations
//	authRes, _, err := client.AuthWithSig(ctx, authReq, walletSigner)
//	if err != nil {
//	    log.Fatal("Authentication failed", "error", err)
//	}
//	jwtToken := authRes.JwtToken // Store this for authenticated calls
func (c *Client) AuthWithSig(ctx context.Context, reqParams AuthRequestRequest, signer sign.Signer) (AuthSigVerifyResponse, []sign.Signature, error) {
	challengeRes, _, err := c.authRequest(ctx, reqParams)
	if err != nil {
		return AuthSigVerifyResponse{}, nil, fmt.Errorf("authentication request failed: %w", err)
	}

	chSig, err := signChallenge(signer, reqParams, challengeRes.ChallengeMessage.String())
	if err != nil {
		return AuthSigVerifyResponse{}, nil, fmt.Errorf("failed to sign challenge: %w", err)
	}

	verifyReq := AuthSigVerifyRequest{
		Challenge: challengeRes.ChallengeMessage,
	}
	return c.authSigVerify(ctx, verifyReq, chSig)
}

func (c *Client) authRequest(ctx context.Context, reqParams AuthRequestRequest) (AuthRequestResponse, []sign.Signature, error) {
	var resParams AuthRequestResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, AuthRequestMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if res.Res.Method != string(AuthChallengeMethod) {
		return resParams, resSig, fmt.Errorf("unexpected response method: %s", res.Res.Method)
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

func (c *Client) authSigVerify(ctx context.Context, reqParams AuthSigVerifyRequest, reqSig sign.Signature) (AuthSigVerifyResponse, []sign.Signature, error) {
	var resParams AuthSigVerifyResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, AuthVerifyMethod, &reqParams, reqSig)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// AuthJWTVerify verifies an existing JWT token and returns the associated session info.
// This is useful for validating a stored JWT token before making authenticated calls.
//
// Parameters:
//   - reqParams: Contains the JWT token to verify
//
// Returns:
//   - AuthJWTVerifyResponse with address, session key, and success status
//   - Response signatures for verification
//   - Error if the JWT is invalid or expired
//
// Example:
//
//	verifyReq := AuthJWTVerifyRequest{JWT: storedJwtToken}
//	verifyRes, _, err := client.AuthJWTVerify(ctx, verifyReq)
//	if err != nil || !verifyRes.Success {
//	    // Token is invalid or expired, need to re-authenticate
//	    authRes, _, err = client.AuthWithSig(ctx, authReq, signer)
//	}
func (c *Client) AuthJWTVerify(ctx context.Context, reqParams AuthJWTVerifyRequest) (AuthJWTVerifyResponse, []sign.Signature, error) {
	var resParams AuthJWTVerifyResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, AuthVerifyMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetUserTag returns the human-readable tag for the authenticated wallet.
// User tags provide a friendly identifier for wallet addresses.
//
// Requires authentication.
//
// Returns:
//   - GetUserTagResponse containing the user's tag
//   - Response signatures for verification
//   - Error if not authenticated or request fails
//
// Example:
//
//	tag, _, err := client.GetUserTag(ctx)
//	if err != nil {
//	    log.Error("Failed to get user tag", "error", err)
//	} else {
//	    fmt.Printf("User tag: %s\n", tag.Tag)
//	}
func (c *Client) GetUserTag(ctx context.Context) (GetUserTagResponse, []sign.Signature, error) {
	var resParams GetUserTagResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetUserTagMethod, nil)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetSessionKeys retrieves session keys with allowances for the authenticated user.
// Returns all active session keys with their spending limits and usage tracking.
//
// Requires authentication.
//
// Parameters:
//   - reqParams: Optional filter options (e.g., pagination)
//
// Returns:
//   - GetSessionKeysResponse containing session keys with allowances
//   - Response signatures for verification
//   - Error if not authenticated or request fails
//
// Example:
//
//	sessionKeys, _, err := client.GetSessionKeys(ctx, GetSessionKeysRequest{})
//	if err != nil {
//	    log.Error("Failed to get session keys", "error", err)
//	}
//	for _, sk := range sessionKeys.SessionKeys {
//	    fmt.Printf("Session key: %s, Application: %s\n", sk.SessionKey, sk.Application)
//	    for _, allowance := range sk.Allowances {
//	        fmt.Printf("  %s: %s / %s used\n", allowance.Asset, allowance.Used, allowance.Allowance)
//	    }
//	}
func (c *Client) GetSessionKeys(ctx context.Context, reqParams GetSessionKeysRequest) (GetSessionKeysResponse, []sign.Signature, error) {
	var resParams GetSessionKeysResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetSessionKeysMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetLedgerBalances retrieves account balances for the authenticated user.
// Balances show the current amount of each asset held in the user's accounts.
//
// Requires authentication.
//
// Parameters:
//   - reqParams: Filter options:
//   - AccountID: Filter balances by specific account (optional)
//
// Returns:
//   - GetLedgerBalancesResponse containing balance information by asset
//   - Response signatures for verification
//   - Error if not authenticated or request fails
//
// Example:
//
//	balances, _, err := client.GetLedgerBalances(ctx, GetLedgerBalancesRequest{})
//	if err != nil {
//	    log.Fatal("Failed to get balances", "error", err)
//	}
//	for _, balance := range balances.LedgerBalances {
//	    fmt.Printf("%s: %s\n", balance.Asset, balance.Amount)
//	}
func (c *Client) GetLedgerBalances(ctx context.Context, reqParams GetLedgerBalancesRequest) (GetLedgerBalancesResponse, []sign.Signature, error) {
	var resParams GetLedgerBalancesResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetLedgerBalancesMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetRPCHistory retrieves the RPC call history for the authenticated user.
// History entries include method calls with their parameters, results, and signatures.
//
// Requires authentication.
//
// Parameters:
//   - reqParams: Pagination options (offset, limit, sort)
//
// Returns:
//   - GetRPCHistoryResponse containing historical RPC entries
//   - Response signatures for verification
//   - Error if not authenticated or request fails
//
// Example:
//
//	history, _, err := client.GetRPCHistory(ctx, GetRPCHistoryRequest{
//	    ListOptions: ListOptions{Limit: 100},
//	})
//	if err != nil {
//	    log.Error("Failed to get RPC history", "error", err)
//	}
//	for _, entry := range history.RPCEntries {
//	    fmt.Printf("[%d] %s by %s\n", entry.Timestamp, entry.Method, entry.Sender)
//	}
func (c *Client) GetRPCHistory(ctx context.Context, reqParams GetRPCHistoryRequest) (GetRPCHistoryResponse, []sign.Signature, error) {
	var resParams GetRPCHistoryResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetRPCHistoryMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// CreateChannel requests the server to create a new payment channel.
// The server validates the request and returns a signed channel state that you must
// then sign and submit to the blockchain yourself to open the channel on-chain.
//
// Requires authentication.
//
// Parameters:
//   - req: Prepared Request with CreateChannelMethod containing:
//   - ChainID: Blockchain network identifier
//   - Token: Asset/token address for the channel
//   - Amount: Initial funding amount
//   - SessionKey: Key that will control the channel (required for security)
//
// Returns:
//   - CreateChannelResponse with initial channel state and server signature
//   - Response signatures for verification
//   - Error if not authenticated, invalid request, or server rejects
//
// After receiving the response, you must:
//  1. Sign the state yourself
//  2. Submit both signatures to the blockchain to open the channel
//
// Example:
//
//	amount := decimal.NewFromInt(1000000)
//	createReq := CreateChannelRequest{
//	    ChainID: 1,
//	    Token: "0xUSDC",
//	    Amount: &amount,
//	    SessionKey: &sessionKeyAddress, // Required
//	}
//	payload, _ := client.PreparePayload(CreateChannelMethod, createReq)
//
//	hash, _ := payload.Hash()
//	sig, err := sessionSigner.Sign(hash)
//	if err != nil {
//	    log.Fatal("Failed to sign request", "error", err)
//	}
//	fullReq := rpc.NewRequest(payload, sig)
//
//	response, _, err := client.CreateChannel(ctx, &fullReq)
//	if err != nil {
//	    log.Fatal("Failed to create channel", "error", err)
//	}
//
//	// Now sign the state and submit to blockchain
//	stateHash := computeStateHash(response.State)
//	mySignature, _ := sessionSigner.Sign(stateHash)
//	// Submit response.StateSignature and mySignature to blockchain
func (c *Client) CreateChannel(ctx context.Context, req *Request) (CreateChannelResponse, []sign.Signature, error) {
	if req == nil || req.Req.Method != string(CreateChannelMethod) {
		return CreateChannelResponse{}, nil, ErrInvalidRequestMethod
	}

	var resParams CreateChannelResponse
	var resSig []sign.Signature

	res, err := c.dialer.Call(ctx, req)
	if err != nil {
		return resParams, res.Sig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Error(); err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// ResizeChannel requests the server to modify channel funding.
// The server returns a signed state update that you must sign and submit to the blockchain.
//
// Requires authentication.
//
// Parameters:
//   - req: Prepared Request with ResizeChannelMethod containing:
//   - ChannelID: ID of the channel to resize
//   - AllocateAmount: Amount to move from your unified balance on ClearNode to the channel (optional)
//   - ResizeAmount: Amount to move from custody ledger on Custody Smart Contract to the channel (optional)
//   - FundsDestination: Where to send funds if reducing channel size
//
// AllocateAmount and ResizeAmount are mutually exclusive - provide only one.
//
// Returns:
//   - ResizeChannelResponse with updated channel state and server signature
//   - Response signatures for verification
//   - Error if not authenticated, invalid request, or server rejects
//
// Example:
//
//	// Move 500 tokens from your ClearNode balance to the channel
//	allocateAmount := decimal.NewFromInt(500)
//	resizeReq := ResizeChannelRequest{
//	    ChannelID: "ch123",
//	    AllocateAmount: &allocateAmount,
//	    FundsDestination: walletAddress,
//	}
//	payload, _ := client.PreparePayload(ResizeChannelMethod, resizeReq)
//
//	hash, _ := payload.Hash()
//	sig, err := sessionSigner.Sign(hash)
//	if err != nil {
//	    log.Fatal("Failed to sign request", "error", err)
//	}
//	fullReq := rpc.NewRequest(payload, sig)
//
//	response, _, err := client.ResizeChannel(ctx, &fullReq)
//	if err != nil {
//	    log.Fatal("Failed to resize channel", "error", err)
//	}
//
//	// Sign and submit the new state to blockchain
func (c *Client) ResizeChannel(ctx context.Context, req *Request) (ResizeChannelResponse, []sign.Signature, error) {
	if req == nil || req.Req.Method != string(ResizeChannelMethod) {
		return ResizeChannelResponse{}, nil, ErrInvalidRequestMethod
	}

	var resParams ResizeChannelResponse
	var resSig []sign.Signature

	res, err := c.dialer.Call(ctx, req)
	if err != nil {
		return resParams, res.Sig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Error(); err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// CloseChannel requests the server to close a payment channel.
// The server returns a final signed state that you must sign and submit to the blockchain
// to close the channel on-chain and recover your funds.
//
// Requires authentication.
//
// Parameters:
//   - req: Prepared Request with CloseChannelMethod containing:
//   - ChannelID: ID of the channel to close
//   - FundsDestination: Where to send the channel funds after closure
//
// Returns:
//   - CloseChannelResponse with final channel state and server signature
//   - Response signatures for verification
//   - Error if not authenticated, invalid request, or server rejects
//
// Example:
//
//	closeReq := CloseChannelRequest{
//	    ChannelID: "ch123",
//	    FundsDestination: walletAddress,
//	}
//	payload, _ := client.PreparePayload(CloseChannelMethod, closeReq)
//
//	hash, _ := payload.Hash()
//	sig, err := sessionSigner.Sign(hash)
//	if err != nil {
//	    log.Fatal("Failed to sign request", "error", err)
//	}
//	fullReq := rpc.NewRequest(payload, sig)
//
//	response, _, err := client.CloseChannel(ctx, &fullReq)
//	if err != nil {
//	    log.Error("Failed to close channel", "error", err)
//	}
//
//	// Sign the final state and submit to blockchain to close channel
func (c *Client) CloseChannel(ctx context.Context, req *Request) (CloseChannelResponse, []sign.Signature, error) {
	if req == nil || req.Req.Method != string(CloseChannelMethod) {
		return CloseChannelResponse{}, nil, ErrInvalidRequestMethod
	}

	var resParams CloseChannelResponse
	var resSig []sign.Signature

	res, err := c.dialer.Call(ctx, req)
	if err != nil {
		return resParams, res.Sig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Error(); err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// Transfer moves funds between accounts within the ClearNode system.
// This is an off-chain transfer that updates balances on ClearNode without
// touching the blockchain. Transfers are instant and gas-free.
//
// Requires authentication.
//
// Parameters:
//   - reqParams: Transfer request containing:
//   - Destination: Recipient's ClearNode account (wallet address)
//   - DestinationUserTag: Recipient's human-readable tag (optional)
//   - Allocations: List of assets and amounts to transfer
//
// Returns:
//   - TransferResponse with created ledger transactions
//   - Response signatures for verification
//   - Error if not authenticated, insufficient balance, or transfer fails
//
// Example:
//
//	transferReq := TransferRequest{
//	    Destination: "0xRecipient...",
//	    Allocations: []TransferAllocation{
//	        {AssetSymbol: "usdc", Amount: decimal.NewFromInt(100)},
//	        {AssetSymbol: "eth", Amount: decimal.NewFromFloat(0.1)},
//	    },
//	}
//
//	response, _, err := client.Transfer(ctx, transferReq)
//	if err != nil {
//	    log.Fatal("Transfer failed", "error", err)
//	}
//	for _, tx := range response.Transactions {
//	    fmt.Printf("Transferred %s %s to %s\n", tx.Amount, tx.Asset, tx.ToAccount)
//	}
func (c *Client) Transfer(ctx context.Context, req *Request) (TransferResponse, []sign.Signature, error) {
	if req == nil || req.Req.Method != string(TransferMethod) {
		return TransferResponse{}, nil, ErrInvalidRequestMethod
	}

	var resParams TransferResponse
	var resSig []sign.Signature

	res, err := c.dialer.Call(ctx, req)
	if err != nil {
		return resParams, res.Sig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Error(); err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// CreateAppSession starts a new virtual application session.
// Application sessions enable multi-party state channel applications.
//
// Requires authentication and signatures from all participants.
//
// Parameters:
//   - req: Prepared Request with CreateAppSessionMethod containing:
//   - Definition: Application protocol and participants
//   - Allocations: Initial asset distribution
//   - SessionData: Application-specific initial state (optional)
//
// The request must be signed by all participants listed in the definition.
//
// Returns:
//   - CreateAppSessionResponse with session ID and details
//   - Response signatures for verification
//   - Error if not authenticated, missing signatures, or creation fails
//
// Example:
//
//	createSessReq := CreateAppSessionRequest{
//	    Definition: AppDefinition{
//	        Protocol: "game/v1",
//	        ParticipantWallets: []string{player1, player2},
//	        Weights: []int64{1, 1},
//	        Quorum: 2,
//	        Challenge: 3600,
//	        Nonce: uint64(uuid.New().ID()),
//	    },
//	    Allocations: []AppAllocation{
//	        {ParticipantWallet: player1, AssetSymbol: "USDC", Amount: decimal.NewFromInt(100)},
//	        {ParticipantWallet: player2, AssetSymbol: "USDC", Amount: decimal.NewFromInt(100)},
//	    },
//	}
//	payload, _ := client.PreparePayload(CreateAppSessionMethod, createSessReq)
//	hash, _ := payload.Hash()
//
//	// Both participants must sign
//	sig1, err := player1Signer.Sign(hash)
//	if err != nil {
//	    log.Fatal("Player 1 sign failed", "error", err)
//	}
//	sig2, err := player2Signer.Sign(hash)
//	if err != nil {
//	    log.Fatal("Player 2 sign failed", "error", err)
//	}
//	fullReq := rpc.NewRequest(payload, sig1, sig2)
//
//	response, _, err := client.CreateAppSession(ctx, &fullReq)
func (c *Client) CreateAppSession(ctx context.Context, req *Request) (CreateAppSessionResponse, []sign.Signature, error) {
	if req == nil || req.Req.Method != string(CreateAppSessionMethod) {
		return CreateAppSessionResponse{}, nil, ErrInvalidRequestMethod
	}

	var resParams CreateAppSessionResponse
	var resSig []sign.Signature

	res, err := c.dialer.Call(ctx, req)
	if err != nil {
		return resParams, res.Sig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Error(); err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// SubmitAppState updates an application session's state.
// Requires signatures from enough participants to meet the session's quorum requirement.
//
// Requires authentication and sufficient signatures to satisfy quorum.
//
// Parameters:
//   - req: Prepared Request with SubmitAppStateMethod containing:
//   - AppSessionID: ID of the session to update
//   - Allocations: New asset distribution
//   - SessionData: New application state (optional)
//
// The request must include signatures totaling at least the quorum weight.
// For example, if quorum is 2 and all participants have weight 1, you need
// signatures from at least 2 participants.
//
// Returns:
//   - SubmitAppStateResponse with updated session details and new version
//   - Response signatures for verification
//   - Error if not authenticated, insufficient signatures, or invalid state
//
// Example:
//
//	updateReq := SubmitAppStateRequest{
//	    AppSessionID: "app123",
//	    Allocations: []AppAllocation{
//	        {ParticipantWallet: winner, AssetSymbol: "USDC", Amount: decimal.NewFromInt(190)},
//	        {ParticipantWallet: loser, AssetSymbol: "USDC", Amount: decimal.NewFromInt(10)},
//	    },
//	    SessionData: &gameResultData,
//	}
//	payload, _ := client.PreparePayload(SubmitAppStateMethod, updateReq)
//	hash, _ := payload.Hash()
//
//	// Get signatures to meet quorum (e.g., both players if quorum=2)
//	sig1, err := player1Signer.Sign(hash)
//	if err != nil {
//	    log.Fatal("Failed to sign", "error", err)
//	}
//	sig2, err := player2Signer.Sign(hash)
//	if err != nil {
//	    log.Fatal("Failed to sign", "error", err)
//	}
//	fullReq := rpc.NewRequest(payload, sig1, sig2)
//
//	response, _, err := client.SubmitAppState(ctx, &fullReq)
func (c *Client) SubmitAppState(ctx context.Context, req *Request) (SubmitAppStateResponse, []sign.Signature, error) {
	if req == nil || req.Req.Method != string(SubmitAppStateMethod) {
		return SubmitAppStateResponse{}, nil, ErrInvalidRequestMethod
	}

	var resParams SubmitAppStateResponse
	var resSig []sign.Signature

	res, err := c.dialer.Call(ctx, req)
	if err != nil {
		return resParams, res.Sig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Error(); err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// CloseAppSession closes an application session and distributes final assets.
// Requires signatures from enough participants to meet the quorum requirement.
//
// Requires authentication and sufficient signatures to satisfy quorum.
//
// Parameters:
//   - req: Prepared Request with CloseAppSessionMethod containing:
//   - AppSessionID: ID of the session to close
//   - SessionData: Final application state (optional)
//   - Allocations: Final asset distribution
//
// The request must include signatures totaling at least the quorum weight.
// Typically, all participants should sign to agree on the final distribution.
//
// Returns:
//   - CloseAppSessionResponse with closed session details
//   - Response signatures for verification
//   - Error if not authenticated, insufficient signatures, or closure fails
//
// Example:
//
//	closeReq := CloseAppSessionRequest{
//	    AppSessionID: "app123",
//	    Allocations: []AppAllocation{
//	        {ParticipantWallet: player1, AssetSymbol: "USDC", Amount: decimal.NewFromInt(150)},
//	        {ParticipantWallet: player2, AssetSymbol: "USDC", Amount: decimal.NewFromInt(50)},
//	    },
//	}
//	payload, _ := client.PreparePayload(CloseAppSessionMethod, closeReq)
//	hash, _ := payload.Hash()
//
//	// Get signatures to meet quorum (typically all participants)
//	sig1, err := player1Signer.Sign(hash)
//	if err != nil {
//	    log.Fatal("Failed to sign", "error", err)
//	}
//	sig2, err := player2Signer.Sign(hash)
//	if err != nil {
//	    log.Fatal("Failed to sign", "error", err)
//	}
//	fullReq := rpc.NewRequest(payload, sig1, sig2)
//
//	response, _, err := client.CloseAppSession(ctx, &fullReq)
func (c *Client) CloseAppSession(ctx context.Context, req *Request) (CloseAppSessionResponse, []sign.Signature, error) {
	if req == nil || req.Req.Method != string(CloseAppSessionMethod) {
		return CloseAppSessionResponse{}, nil, ErrInvalidRequestMethod
	}

	var resParams CloseAppSessionResponse
	var resSig []sign.Signature

	res, err := c.dialer.Call(ctx, req)
	if err != nil {
		return resParams, res.Sig, err
	}
	resSig = res.Sig

	if err := res.Res.Params.Error(); err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// CleanupSessionKeyCache clears cached session key data on the server.
// This is applicable only when Clearnode is running in test mode.
//
// Example:
//
//	sigs, err := client.CleanupSessionKeyCache(ctx)
//	if err != nil {
//	    log.Error("failed to cleanup session key cache", "error", err)
//	}
func (c *Client) CleanupSessionKeyCache(ctx context.Context) ([]sign.Signature, error) {
	var resSig []sign.Signature
	res, err := c.call(ctx, CleanupSessionKeyCacheMethod, nil)
	if err != nil {
		return resSig, err
	}
	resSig = res.Sig

	return resSig, nil
}

func (c *Client) call(ctx context.Context, method Method, reqParams any, sigs ...sign.Signature) (*Response, error) {
	payload, err := c.PreparePayload(method, reqParams)
	if err != nil {
		return nil, err
	}

	req := NewRequest(
		payload,
		sigs...,
	)

	res, err := c.dialer.Call(ctx, &req)
	if err != nil {
		return nil, err
	}

	if err := res.Res.Params.Error(); err != nil {
		return nil, err
	}

	return res, nil
}

// PreparePayload creates a Payload for an RPC method call.
// This helper method generates a unique request ID and packages the parameters.
//
// Parameters:
//   - method: The RPC method to call
//   - reqParams: The request parameters (can be nil for methods without parameters)
//
// Returns:
//   - Payload ready to be wrapped in a Request
//   - Error if parameter marshaling fails
//
// Example:
//
//	req := GetAssetsRequest{ChainID: &chainID}
//	payload, err := client.PreparePayload(GetAssetsMethod, req)
//	if err != nil {
//	    log.Fatal("Failed to prepare payload", "error", err)
//	}
//	// Now create a Request with the payload and any required signatures
//	fullReq := rpc.NewRequest(payload)
func (c *Client) PreparePayload(method Method, reqParams any) (Payload, error) {
	params, err := NewParams(reqParams)
	if err != nil {
		return Payload{}, err
	}

	return NewPayload(
		uint64(uuid.New().ID()),
		string(method),
		params,
	), nil
}

// HandleBalanceUpdateEvent registers a handler for balance update notifications.
// The handler will be called whenever the server sends a balance update event.
// Only one handler can be registered at a time; subsequent calls override the previous handler.
//
// Example:
//
//	client.HandleBalanceUpdateEvent(func(ctx context.Context, notif BalanceUpdateNotification, sigs []sign.Signature) {
//	    for _, balance := range notif.BalanceUpdates {
//	        fmt.Printf("Balance changed: %s = %s\n", balance.Asset, balance.Amount)
//	    }
//	})
func (c *Client) HandleBalanceUpdateEvent(handler BalanceUpdateEventHandler) {
	c.setEventHandler(BalanceUpdateEvent, handler)
}

func (c *Client) handleBalanceUpdateEvent(ctx context.Context, event *Response) {
	logger := log.FromContext(ctx)
	handler, ok := c.getEventHandler(BalanceUpdateEvent).(BalanceUpdateEventHandler)
	if !ok {
		logger.Warn("no handler for event", "method", event.Res.Method)
		return
	}

	var notif BalanceUpdateNotification
	if err := event.Res.Params.Translate(&notif); err != nil {
		logger.Error("failed to translate event", "error", err, "method", event.Res.Method)
		return
	}

	handler(ctx, notif, event.Sig)
}

// HandleChannelUpdateEvent registers a handler for channel update notifications.
// The handler will be called whenever a channel's state changes.
// Only one handler can be registered at a time; subsequent calls override the previous handler.
//
// Example:
//
//	client.HandleChannelUpdateEvent(func(ctx context.Context, notif ChannelUpdateNotification, sigs []sign.Signature) {
//	    fmt.Printf("Channel %s updated: status=%s\n", notif.ChannelID, notif.Status)
//	})
func (c *Client) HandleChannelUpdateEvent(handler ChannelUpdateEventHandler) {
	c.setEventHandler(ChannelUpdateEvent, handler)
}

func (c *Client) handleChannelUpdateEvent(ctx context.Context, event *Response) {
	logger := log.FromContext(ctx)
	handler, ok := c.getEventHandler(ChannelUpdateEvent).(ChannelUpdateEventHandler)
	if !ok {
		logger.Warn("no handler for event", "method", event.Res.Method)
		return
	}

	var notif ChannelUpdateNotification
	if err := event.Res.Params.Translate(&notif); err != nil {
		logger.Error("failed to translate event", "error", err, "method", event.Res.Method)
		return
	}

	handler(ctx, notif, event.Sig)
}

// HandleAppSessionUpdateEvent registers a handler for application session update notifications.
// The handler will be called whenever an application session's state changes.
// Only one handler can be registered at a time; subsequent calls override the previous handler.
//
// Example:
//
//	client.HandleAppSessionUpdateEvent(func(ctx context.Context, notif AppSessionUpdateNotification, sigs []sign.Signature) {
//	    fmt.Printf("App Session %s updated: status=%s\n", notif.AppSession.AppSessionID, notif.AppSession.Status)
//	})
func (c *Client) HandleAppSessionUpdateEvent(handler AppSessionUpdateEventHandler) {
	c.setEventHandler(AppSessionUpdateEvent, handler)
}

func (c *Client) handleAppSessionUpdateEvent(ctx context.Context, event *Response) {
	logger := log.FromContext(ctx)
	handler, ok := c.getEventHandler(AppSessionUpdateEvent).(AppSessionUpdateEventHandler)
	if !ok {
		logger.Warn("no handler for event", "method", event.Res.Method)
		return
	}

	var notif AppSessionUpdateNotification
	if err := event.Res.Params.Translate(&notif); err != nil {
		logger.Error("failed to translate event", "error", err, "method", event.Res.Method)
		return
	}

	handler(ctx, notif, event.Sig)
}

// HandleTransferEvent registers a handler for transfer notifications.
// The handler will be called for both incoming and outgoing transfers.
// Only one handler can be registered at a time; subsequent calls override the previous handler.
//
// Example:
//
//	client.HandleTransferEvent(func(ctx context.Context, notif TransferNotification, sigs []sign.Signature) {
//	    for _, tx := range notif.Transactions {
//	        if tx.ToAccount == myAccount {
//	            fmt.Printf("Received %s %s from %s\n", tx.Amount, tx.Asset, tx.FromAccount)
//	        }
//	    }
//	})
func (c *Client) HandleTransferEvent(handler TransferEventHandler) {
	c.setEventHandler(TransferEvent, handler)
}

func (c *Client) handleTransferEvent(ctx context.Context, event *Response) {
	logger := log.FromContext(ctx)
	handler, ok := c.getEventHandler(TransferEvent).(TransferEventHandler)
	if !ok {
		logger.Warn("no handler for event", "method", event.Res.Method)
		return
	}

	var notif TransferNotification
	if err := event.Res.Params.Translate(&notif); err != nil {
		logger.Error("failed to translate event", "error", err, "method", event.Res.Method)
		return
	}

	handler(ctx, notif, event.Sig)
}

func (c *Client) setEventHandler(event Event, handler any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.eventHandlers[event] = handler
}

func (c *Client) getEventHandler(event Event) any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.eventHandlers[event]
}

func signChallenge(signer sign.Signer, req AuthRequestRequest, token string) (sign.Signature, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
			},
			"Policy": {
				{Name: "challenge", Type: "string"},
				{Name: "scope", Type: "string"},
				{Name: "wallet", Type: "address"},
				{Name: "session_key", Type: "address"},
				{Name: "expires_at", Type: "uint64"},
				{Name: "allowances", Type: "Allowance[]"},
			},
			"Allowance": {
				{Name: "asset", Type: "string"},
				{Name: "amount", Type: "string"},
			},
		},
		PrimaryType: "Policy",
		Domain: apitypes.TypedDataDomain{
			Name: req.Application,
		},
		Message: map[string]any{
			"challenge":   token,
			"scope":       req.Scope,
			"wallet":      req.Address,
			"session_key": req.SessionKey,
			"expires_at":  new(big.Int).SetUint64(req.ExpiresAt),
			"allowances":  req.Allowances,
		},
	}

	hash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return sign.Signature{}, err
	}

	signature, err := signer.Sign(hash)
	if err != nil {
		return sign.Signature{}, err
	}

	return signature, nil
}
