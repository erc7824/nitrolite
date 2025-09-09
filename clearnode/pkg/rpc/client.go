// Package rpc provides client-side functionality for the Clearnode RPC protocol.
// This file contains the Client implementation that handles RPC method calls,
// event subscriptions, and manages the underlying connection through a Dialer.
package rpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/erc7824/nitrolite/clearnode/pkg/log"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"github.com/google/uuid"
)

// Client provides a high-level interface for interacting with a Clearnode RPC server.
// It wraps a Dialer implementation (e.g., WebSocket) and provides type-safe methods
// for all RPC operations, automatic event handling, and thread-safe event handler
// registration.
//
// The Client is designed to be used concurrently from multiple goroutines and
// maintains a registry of event handlers for processing unsolicited server events.
//
// Example usage:
//
//	dialer := rpc.NewWebsocketDialer(rpc.DefaultWebsocketDialerConfig)
//	client := rpc.NewClient(dialer)
//
//	// Register event handlers before connecting
//	client.HandleBalanceUpdateEvent(func(ctx context.Context, notif BalanceUpdateNotification, sigs []sign.Signature) {
//	    fmt.Printf("Balance updated: %v\n", notif.BalanceUpdates)
//	})
//
//	// Connect and start listening for events
//	go dialer.Dial(ctx, "ws://localhost:8080/ws", handleClosure)
//	go client.ListenEvents(ctx, handleClosure)
//
//	// Make RPC calls
//	config, sigs, err := client.GetConfig(ctx)
type Client struct {
	Dialer
	eventHandlers map[Event]any
	mu            sync.RWMutex // protects eventHandlers
}

// NewClient creates a new RPC client with the provided dialer.
// The dialer is responsible for establishing and maintaining the connection
// to the RPC server. Common dialer implementations include WebsocketDialer.
//
// Parameters:
//   - dialer: The Dialer implementation to use for communication
//
// Returns:
//   - *Client: A new client instance ready for use
//
// Example:
//
//	cfg := rpc.DefaultWebsocketDialerConfig
//	dialer := rpc.NewWebsocketDialer(cfg)
//	client := rpc.NewClient(dialer)
func NewClient(dialer Dialer) *Client {
	return &Client{
		Dialer:        dialer,
		eventHandlers: make(map[Event]any),
	}
}

// BalanceUpdateEventHandler is the callback function type for balance update events.
// It receives the context, notification data, and server signatures for verification.
type BalanceUpdateEventHandler func(ctx context.Context, notif BalanceUpdateNotification, resSig []sign.Signature)

// ChannelUpdateEventHandler is the callback function type for channel state change events.
// It receives the context, notification data, and server signatures for verification.
type ChannelUpdateEventHandler func(ctx context.Context, notif ChannelUpdateNotification, resSig []sign.Signature)

// TransferEventHandler is the callback function type for transfer events.
// It receives the context, notification data, and server signatures for verification.
type TransferEventHandler func(ctx context.Context, notif TransferNotification, resSig []sign.Signature)

// ListenEvents starts listening for unsolicited events from the server.
// This method should be called in a separate goroutine after establishing
// the connection through the dialer. It will continuously process events
// until the context is cancelled or the connection is closed.
//
// The method automatically routes events to their registered handlers based
// on the event type. Unknown events are logged but do not cause errors.
//
// Parameters:
//   - ctx: Context for cancellation and logging
//   - handleClosure: Callback invoked when the event loop exits (err is nil for clean shutdown)
//
// Example usage:
//
//	// Start event listener in background
//	go client.ListenEvents(ctx, func(err error) {
//	    if err != nil {
//	        log.Error("Event listener stopped with error", "error", err)
//	    } else {
//	        log.Info("Event listener stopped cleanly")
//	    }
//	})
//
// Note: Make sure to register event handlers before calling this method.
func (c *Client) ListenEvents(ctx context.Context, handleClosure func(err error)) {
	logger := log.FromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			handleClosure(nil)
			return
		case event := <-c.EventCh():
			if event == nil {
				handleClosure(nil)
				return
			}

			switch event.Res.Method {
			case BalanceUpdateEvent.String():
				c.handleBalanceUpdateEvent(ctx, event)
			case ChannelUpdateEvent.String():
				c.handleChannelUpdateEvent(ctx, event)
			case TransferEvent.String():
				c.handleTransferEvent(ctx, event)
			default:
				logger.Warn("unknown event received", "method", event.Res.Method)
			}
		}
	}
}

// Ping sends a ping request to the server to check connectivity and liveness.
// The server should respond with a pong message. This method can be used
// for health checks or to keep the connection alive.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//
// Returns:
//   - []sign.Signature: Server signatures on the pong response
//   - error: Error if the ping failed or received unexpected response
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	sigs, err := client.Ping(ctx)
//	if err != nil {
//	    log.Error("Server not responding", "error", err)
//	}
func (c *Client) Ping(ctx context.Context) ([]sign.Signature, error) {
	var resSig []sign.Signature
	res, err := c.call(ctx, PingMethod, nil)
	if err != nil {
		return resSig, err
	}

	if res.Res.Method != string(PongMethod) {
		return resSig, fmt.Errorf("unexpected response method: %s", res.Res.Method)
	}

	return resSig, nil
}

// GetConfig retrieves the current server configuration.
// This includes network parameters and other operational settings.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//
// Returns:
//   - GetConfigResponse: Server configuration data
//   - []sign.Signature: Server signatures for verification
//   - error: Error if the request failed
//
// Example:
//
//	config, sigs, err := client.GetConfig(ctx)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Network: %+v\n", config.Networks)
func (c *Client) GetConfig(ctx context.Context) (GetConfigResponse, []sign.Signature, error) {
	var resParams GetConfigResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetConfigMethod, nil)
	if err != nil {
		return resParams, resSig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, resSig, nil
}

// GetAssets retrieves information about supported assets.
// This includes asset addresses, decimals, symbols, and other metadata.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Request parameters for filtering assets
//
// Returns:
//   - GetAssetsResponse: Asset information
//   - []sign.Signature: Server signatures for verification
//   - error: Error if the request failed
//
// Example:
//
//	req := GetAssetsRequest{
//	    AssetIDs: []string{"ETH", "USDC"},
//	}
//	assets, sigs, err := client.GetAssets(ctx, req)
//	if err != nil {
//	    return err
//	}
//	for _, asset := range assets.Assets {
//	    fmt.Printf("%s: %s (decimals: %d)\n", asset.Symbol, asset.Address, asset.Decimals)
//	}
func (c *Client) GetAssets(ctx context.Context, reqParams GetAssetsRequest) (GetAssetsResponse, []sign.Signature, error) {
	var resParams GetAssetsResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetAssetsMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, resSig, nil
}

// GetAppDefinition retrieves the definition of a specific application.
// This includes the application's code, ABI, version, and deployment information.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Request containing the app ID to query
//
// Returns:
//   - GetAppDefinitionResponse: Application definition data
//   - []sign.Signature: Server signatures for verification
//   - error: Error if the request failed
//
// Example:
//
//	req := GetAppDefinitionRequest{
//	    AppID: "0xabc123...",
//	}
//	appDef, sigs, err := client.GetAppDefinition(ctx, req)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("App version: %s\n", appDef.Version)
//	fmt.Printf("Bytecode: %x\n", appDef.Bytecode)
func (c *Client) GetAppDefinition(ctx context.Context, reqParams GetAppDefinitionRequest) (GetAppDefinitionResponse, []sign.Signature, error) {
	var resParams GetAppDefinitionResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetAppDefinitionMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetAppSessions retrieves information about application sessions.
// This can be filtered by app ID, participants, or status.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Request parameters for filtering sessions
//
// Returns:
//   - GetAppSessionsResponse: Application session data
//   - []sign.Signature: Server signatures for verification
//   - error: Error if the request failed
//
// Example:
//
//	req := GetAppSessionsRequest{
//	    AppID: "0xabc123...",
//	    Status: "active",
//	}
//	sessions, sigs, err := client.GetAppSessions(ctx, req)
//	if err != nil {
//	    return err
//	}
//	for _, session := range sessions.Sessions {
//	    fmt.Printf("Session %s: %d participants\n", session.ID, len(session.Participants))
//	}
func (c *Client) GetAppSessions(ctx context.Context, reqParams GetAppSessionsRequest) (GetAppSessionsResponse, []sign.Signature, error) {
	var resParams GetAppSessionsResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetAppSessionsMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetChannels retrieves information about payment channels.
// This can be filtered by participant, status, or other criteria.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Request parameters for filtering channels
//
// Returns:
//   - GetChannelsResponse: Channel information
//   - []sign.Signature: Server signatures for verification
//   - error: Error if the request failed
//
// Example:
//
//	req := GetChannelsRequest{
//	    Participant: myAddress,
//	    Status: "open",
//	}
//	channels, sigs, err := client.GetChannels(ctx, req)
//	if err != nil {
//	    return err
//	}
//	for _, ch := range channels.Channels {
//	    fmt.Printf("Channel %s: %s ↔ %s\n", ch.ID, ch.ParticipantA, ch.ParticipantB)
//	}
func (c *Client) GetChannels(ctx context.Context, reqParams GetChannelsRequest) (GetChannelsResponse, []sign.Signature, error) {
	var resParams GetChannelsResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetChannelsMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetLedgerEntries retrieves ledger entries (debits and credits) for accounts.
// This provides a detailed transaction history with individual ledger operations.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Request parameters for filtering entries
//
// Returns:
//   - GetLedgerEntriesResponse: Ledger entry data
//   - []sign.Signature: Server signatures for verification
//   - error: Error if the request failed
//
// Example:
//
//	req := GetLedgerEntriesRequest{
//	    Account: myAddress,
//	    Limit: 100,
//	    Offset: 0,
//	}
//	entries, sigs, err := client.GetLedgerEntries(ctx, req)
//	if err != nil {
//	    return err
//	}
//	for _, entry := range entries.Entries {
//	    fmt.Printf("%s: %s %s\n", entry.Type, entry.Amount, entry.Asset)
//	}
func (c *Client) GetLedgerEntries(ctx context.Context, reqParams GetLedgerEntriesRequest) (GetLedgerEntriesResponse, []sign.Signature, error) {
	var resParams GetLedgerEntriesResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetLedgerEntriesMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetLedgerTransactions retrieves complete transaction records.
// Each transaction may contain multiple ledger entries.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Request parameters for filtering transactions
//
// Returns:
//   - GetLedgerTransactionsResponse: Transaction data
//   - []sign.Signature: Server signatures for verification
//   - error: Error if the request failed
//
// Example:
//
//	req := GetLedgerTransactionsRequest{
//	    Account: myAddress,
//	    FromDate: time.Now().AddDate(0, -1, 0), // Last month
//	    ToDate: time.Now(),
//	}
//	txs, sigs, err := client.GetLedgerTransactions(ctx, req)
//	if err != nil {
//	    return err
//	}
//	for _, tx := range txs.Transactions {
//	    fmt.Printf("Tx %s: %s at %v\n", tx.ID, tx.Type, tx.Timestamp)
//	}
func (c *Client) GetLedgerTransactions(ctx context.Context, reqParams GetLedgerTransactionsRequest) (GetLedgerTransactionsResponse, []sign.Signature, error) {
	var resParams GetLedgerTransactionsResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetLedgerTransactionsMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// AuthRequest initiates an authentication flow by requesting a challenge.
// The server returns a challenge that must be signed and submitted via AuthSigVerify.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Authentication request parameters
//
// Returns:
//   - AuthRequestResponse: Challenge data to be signed
//   - []sign.Signature: Server signatures for verification
//   - error: Error if the request failed
//
// Example:
//
//	req := AuthRequestRequest{
//	    Address: myAddress,
//	    Scope: []string{"read", "write"},
//	}
//	challenge, sigs, err := client.AuthRequest(ctx, req)
//	if err != nil {
//	    return err
//	}
//	// Sign the challenge with your private key
//	signature := signChallenge(challenge.Challenge)
//	// Submit the signature via AuthSigVerify
func (c *Client) AuthRequest(ctx context.Context, reqParams AuthRequestRequest) (AuthRequestResponse, []sign.Signature, error) {
	var resParams AuthRequestResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, AuthRequestMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}

	if res.Res.Method != string(AuthChallengeMethod) {
		return resParams, resSig, fmt.Errorf("unexpected response method: %s", res.Res.Method)
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// AuthSigVerify completes signature-based authentication by submitting
// a signed challenge. This must be called after AuthRequest with the
// challenge signed by the user's private key.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Verification request containing the original challenge
//   - reqSig: Signature over the challenge
//
// Returns:
//   - AuthSigVerifyResponse: Authentication token or session data
//   - []sign.Signature: Server signatures for verification
//   - error: Error if authentication failed
//
// Example:
//
//	// Assuming you have the challenge from AuthRequest
//	signature := signWithPrivateKey(challenge.Challenge)
//	verifyReq := AuthSigVerifyRequest{
//	    Challenge: challenge.Challenge,
//	    Address: myAddress,
//	}
//	authResp, sigs, err := client.AuthSigVerify(ctx, verifyReq, signature)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Authenticated! Token: %s\n", authResp.Token)
func (c *Client) AuthSigVerify(ctx context.Context, reqParams AuthSigVerifyRequest, reqSig sign.Signature) (AuthSigVerifyResponse, []sign.Signature, error) {
	var resParams AuthSigVerifyResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, AuthVerifyMethod, &reqParams, reqSig)
	if err != nil {
		return resParams, resSig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// AuthJWTVerify performs JWT-based authentication.
// This is an alternative to signature-based auth for clients that
// already have a valid JWT token from an external identity provider.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Request containing the JWT token
//
// Returns:
//   - AuthJWTVerifyResponse: Session data or access token
//   - []sign.Signature: Server signatures for verification
//   - error: Error if authentication failed
//
// Example:
//
//	req := AuthJWTVerifyRequest{
//	    Token: jwtTokenFromOAuth,
//	    Provider: "auth0",
//	}
//	authResp, sigs, err := client.AuthJWTVerify(ctx, req)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("JWT verified! User: %s\n", authResp.UserID)
func (c *Client) AuthJWTVerify(ctx context.Context, reqParams AuthJWTVerifyRequest) (AuthJWTVerifyResponse, []sign.Signature, error) {
	var resParams AuthJWTVerifyResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, AuthVerifyMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetUserTag retrieves the user tag (human-readable identifier) for the
// authenticated account. User tags are unique aliases that can be used
// instead of addresses for better UX.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//
// Returns:
//   - GetUserTagResponse: User tag information
//   - []sign.Signature: Server signatures for verification
//   - error: Error if the request failed
//
// Note: This method requires authentication.
//
// Example:
//
//	userTag, sigs, err := client.GetUserTag(ctx)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Your user tag: @%s\n", userTag.Tag)
func (c *Client) GetUserTag(ctx context.Context) (GetUserTagResponse, []sign.Signature, error) {
	var resParams GetUserTagResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetUserTagMethod, nil)
	if err != nil {
		return resParams, resSig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetLedgerBalances retrieves current account balances for specified assets.
// This provides the net balance after all debits and credits.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Request specifying accounts and assets to query
//
// Returns:
//   - GetLedgerBalancesResponse: Current balance information
//   - []sign.Signature: Server signatures for verification
//   - error: Error if the request failed
//
// Example:
//
//	req := GetLedgerBalancesRequest{
//	    Account: myAddress,
//	    Assets: []string{"ETH", "USDC"},
//	}
//	balances, sigs, err := client.GetLedgerBalances(ctx, req)
//	if err != nil {
//	    return err
//	}
//	for _, balance := range balances.Balances {
//	    fmt.Printf("%s: %s\n", balance.Asset, balance.Amount)
//	}
func (c *Client) GetLedgerBalances(ctx context.Context, reqParams GetLedgerBalancesRequest) (GetLedgerBalancesResponse, []sign.Signature, error) {
	var resParams GetLedgerBalancesResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetLedgerBalancesMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// GetRPCHistory retrieves the history of RPC calls made by the authenticated user.
// This can be useful for debugging, auditing, or replaying operations.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Request parameters for filtering history
//
// Returns:
//   - GetRPCHistoryResponse: Historical RPC data
//   - []sign.Signature: Server signatures for verification
//   - error: Error if the request failed
//
// Note: This method requires authentication.
//
// Example:
//
//	req := GetRPCHistoryRequest{
//	    Methods: []string{"Transfer", "CreateChannel"},
//	    Limit: 50,
//	    Since: time.Now().Add(-24 * time.Hour),
//	}
//	history, sigs, err := client.GetRPCHistory(ctx, req)
//	if err != nil {
//	    return err
//	}
//	for _, call := range history.Calls {
//	    fmt.Printf("%v: %s\n", call.Timestamp, call.Method)
//	}
func (c *Client) GetRPCHistory(ctx context.Context, reqParams GetRPCHistoryRequest) (GetRPCHistoryResponse, []sign.Signature, error) {
	var resParams GetRPCHistoryResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, GetRPCHistoryMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// CreateChannel creates a new payment channel between two participants.
// This operation requires a signature from the channel creator and may
// involve locking funds in the channel.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Channel creation parameters (participants, initial balances)
//   - reqSig: Signature authorizing the channel creation
//
// Returns:
//   - CreateChannelResponse: Created channel information
//   - []sign.Signature: Server signatures for verification
//   - error: Error if channel creation failed
//
// Example:
//
//	req := CreateChannelRequest{
//	    Participants: []string{myAddress, counterpartyAddress},
//	    InitialBalances: map[string]string{
//	        myAddress: "1000000000000000000", // 1 ETH
//	        counterpartyAddress: "0",
//	    },
//	    Asset: "ETH",
//	}
//	signature := signChannelCreation(req)
//	channel, sigs, err := client.CreateChannel(ctx, req, signature)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Channel created: %s\n", channel.ChannelID)
func (c *Client) CreateChannel(ctx context.Context, req *Request) (CreateChannelResponse, []sign.Signature, error) {
	if req == nil || req.Req.Method != string(CreateChannelMethod) {
		return CreateChannelResponse{}, nil, ErrInvalidRequestMethod
	}

	var resParams CreateChannelResponse
	var resSig []sign.Signature

	res, err := c.Call(ctx, req)
	if err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Error(); err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// ResizeChannel modifies the capacity of an existing channel by adding
// or removing funds. This operation requires signatures from all participants.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Resize parameters (channel ID, new balances)
//   - reqSig: Signature authorizing the resize
//
// Returns:
//   - ResizeChannelResponse: Updated channel information
//   - []sign.Signature: Server signatures for verification
//   - error: Error if resize failed
//
// Example:
//
//	req := ResizeChannelRequest{
//	    ChannelID: channelID,
//	    NewBalances: map[string]string{
//	        myAddress: "2000000000000000000", // Increase to 2 ETH
//	        counterpartyAddress: "0",
//	    },
//	}
//	signature := signChannelResize(req)
//	updatedChannel, sigs, err := client.ResizeChannel(ctx, req, signature)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Channel resized: new capacity %s\n", updatedChannel.TotalCapacity)
func (c *Client) ResizeChannel(ctx context.Context, req *Request) (ResizeChannelResponse, []sign.Signature, error) {
	if req == nil || req.Req.Method != string(ResizeChannelMethod) {
		return ResizeChannelResponse{}, nil, ErrInvalidRequestMethod
	}

	var resParams ResizeChannelResponse
	var resSig []sign.Signature

	res, err := c.Call(ctx, req)
	if err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Error(); err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// CloseChannel initiates the closing of a payment channel.
// This can be done cooperatively (with all participants' signatures)
// or unilaterally (which may trigger a challenge period).
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Close parameters (channel ID, final state)
//   - reqSig: Signature authorizing the close
//
// Returns:
//   - CloseChannelResponse: Close confirmation and timeline
//   - []sign.Signature: Server signatures for verification
//   - error: Error if close initiation failed
//
// Example:
//
//	req := CloseChannelRequest{
//	    ChannelID: channelID,
//	    FinalBalances: map[string]string{
//	        myAddress: "1500000000000000000",
//	        counterpartyAddress: "500000000000000000",
//	    },
//	    Cooperative: true,
//	}
//	signature := signChannelClose(req)
//	closeResp, sigs, err := client.CloseChannel(ctx, req, signature)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Channel closing initiated. Finalization at: %v\n", closeResp.FinalizationTime)
func (c *Client) CloseChannel(ctx context.Context, req *Request) (CloseChannelResponse, []sign.Signature, error) {
	if req == nil || req.Req.Method != string(CloseChannelMethod) {
		return CloseChannelResponse{}, nil, ErrInvalidRequestMethod
	}

	var resParams CloseChannelResponse
	var resSig []sign.Signature

	res, err := c.Call(ctx, req)
	if err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Error(); err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// Transfer moves funds between accounts. This can be a simple transfer
// or a more complex multi-party transaction. The transfer is atomic -
// either all operations succeed or all fail.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Transfer details (from, to, amount, asset)
//
// Returns:
//   - TransferResponse: Transfer confirmation and transaction ID
//   - []sign.Signature: Server signatures for verification
//   - error: Error if transfer failed
//
// Note: The request must be properly signed by the sender.
//
// Example:
//
//	req := TransferRequest{
//	    From: myAddress,
//	    To: recipientAddress,
//	    Amount: "1000000000000000000", // 1 ETH
//	    Asset: "ETH",
//	    Memo: "Payment for services",
//	}
//	txResp, sigs, err := client.Transfer(ctx, req)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Transfer completed: %s\n", txResp.TransactionID)
func (c *Client) Transfer(ctx context.Context, reqParams TransferRequest) (TransferResponse, []sign.Signature, error) {
	var resParams TransferResponse
	var resSig []sign.Signature

	res, err := c.call(ctx, TransferMethod, &reqParams)
	if err != nil {
		return resParams, resSig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// CreateAppSession creates a new application session for multi-party computation
// or state channel applications. All participants must sign the session creation.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Session parameters (app ID, participants, initial state)
//   - reqSigs: Signatures from all participants
//
// Returns:
//   - CreateAppSessionResponse: Created session information
//   - []sign.Signature: Server signatures for verification
//   - error: Error if session creation failed
//
// Example:
//
//	req := CreateAppSessionRequest{
//	    AppID: appID,
//	    Participants: []string{alice, bob, charlie},
//	    InitialState: initialGameState,
//	    Timeout: 3600, // 1 hour session timeout
//	}
//	// Collect signatures from all participants
//	sigs := []sign.Signature{aliceSig, bobSig, charlieSig}
//
//	session, serverSigs, err := client.CreateAppSession(ctx, req, sigs)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Session created: %s\n", session.SessionID)
func (c *Client) CreateAppSession(ctx context.Context, req *Request) (CreateAppSessionResponse, []sign.Signature, error) {
	if req == nil || req.Req.Method != string(CreateAppSessionMethod) {
		return CreateAppSessionResponse{}, nil, ErrInvalidRequestMethod
	}

	var resParams CreateAppSessionResponse
	var resSig []sign.Signature

	res, err := c.Call(ctx, req)
	if err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Error(); err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// SubmitAppState submits a new state update for an application session.
// State updates must be signed by the required participants according
// to the application's rules (e.g., all participants, majority, etc.).
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: State update (session ID, new state, sequence number)
//   - reqSigs: Signatures from required participants
//
// Returns:
//   - SubmitAppStateResponse: State update confirmation
//   - []sign.Signature: Server signatures for verification
//   - error: Error if state update failed
//
// Example:
//
//	req := SubmitAppStateRequest{
//	    SessionID: sessionID,
//	    NewState: updatedGameState,
//	    SequenceNumber: 42,
//	    StateHash: computeStateHash(updatedGameState),
//	}
//	// For a game requiring all players to sign
//	sigs := []sign.Signature{aliceSig, bobSig, charlieSig}
//
//	stateResp, serverSigs, err := client.SubmitAppState(ctx, req, sigs)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("State updated to sequence %d\n", stateResp.SequenceNumber)
func (c *Client) SubmitAppState(ctx context.Context, req *Request) (SubmitAppStateResponse, []sign.Signature, error) {
	if req == nil || req.Req.Method != string(SubmitAppStateMethod) {
		return SubmitAppStateResponse{}, nil, ErrInvalidRequestMethod
	}

	var resParams SubmitAppStateResponse
	var resSig []sign.Signature

	res, err := c.Call(ctx, req)
	if err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Error(); err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// CloseAppSession closes an application session and finalizes its state.
// This distributes any funds or assets according to the final state.
// Can be done cooperatively or through timeout/dispute resolution.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - reqParams: Close parameters (session ID, final state, outcome)
//   - reqSigs: Signatures from required participants
//
// Returns:
//   - CloseAppSessionResponse: Session closure confirmation
//   - []sign.Signature: Server signatures for verification
//   - error: Error if closure failed
//
// Example:
//
//	req := CloseAppSessionParams{
//	    SessionID: sessionID,
//	    FinalState: finalGameState,
//	    Outcome: map[string]string{
//	        alice: "1000000000000000000", // Alice wins 1 ETH
//	        bob: "500000000000000000",   // Bob gets 0.5 ETH
//	        charlie: "500000000000000000", // Charlie gets 0.5 ETH
//	    },
//	}
//	// All participants sign the final outcome
//	sigs := []sign.Signature{aliceSig, bobSig, charlieSig}
//
//	closeResp, serverSigs, err := client.CloseAppSession(ctx, req, sigs)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Session closed. Funds distributed per outcome.\n")
func (c *Client) CloseAppSession(ctx context.Context, req *Request) (CloseAppSessionResponse, []sign.Signature, error) {
	if req == nil || req.Req.Method != string(CloseAppSessionMethod) {
		return CloseAppSessionResponse{}, nil, ErrInvalidRequestMethod
	}

	var resParams CloseAppSessionResponse
	var resSig []sign.Signature

	res, err := c.Call(ctx, req)
	if err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Error(); err != nil {
		return resParams, res.Sig, err
	}

	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, resSig, err
	}

	return resParams, res.Sig, nil
}

// call is an internal helper method that constructs and sends RPC requests.
// It handles the common pattern of creating a request with proper timestamp,
// sending it through the dialer, and checking for errors in the response.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - method: The RPC method to invoke
//   - reqParams: Request parameters (can be nil for methods without params)
//   - sigs: Optional signatures to include with the request
//
// Returns:
//   - *Response: The server's response
//   - error: Error from transport, server, or response parsing
//
// This method:
// 1. Converts parameters to the Params type
// 2. Creates a timestamped payload with auto-generated request ID
// 3. Wraps payload and signatures in a Request
// 4. Sends via the Dialer's Call method
// 5. Checks for protocol errors in the response
func (c *Client) call(ctx context.Context, method Method, reqParams any, sigs ...sign.Signature) (*Response, error) {
	payload, err := c.PreparePayload(method, reqParams)
	if err != nil {
		return nil, err
	}

	req := NewRequest(
		payload,
		sigs...,
	)

	res, err := c.Call(ctx, &req)
	if err != nil {
		return nil, err
	}

	if err := res.Res.Params.Error(); err != nil {
		return nil, err
	}

	return res, nil
}

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
// The handler will be called whenever account balances change due to transfers,
// channel operations, or application session updates.
//
// Only one handler can be registered per event type. Calling this method
// multiple times will replace the previous handler.
//
// Parameters:
//   - handler: Callback function to handle balance updates
//
// Example usage:
//
//	client.HandleBalanceUpdateEvent(func(ctx context.Context, notif BalanceUpdateNotification, sigs []sign.Signature) {
//	    for _, balance := range notif.BalanceUpdates {
//	        fmt.Printf("Balance changed - %s: %s\n", balance.Asset, balance.Amount)
//	    }
//	})
func (c *Client) HandleBalanceUpdateEvent(handler BalanceUpdateEventHandler) {
	c.setEventHandler(BalanceUpdateEvent, handler)
}

// handleBalanceUpdateEvent is an internal method that processes balance update events.
// It retrieves the registered handler and invokes it with the notification data.
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

// HandleChannelUpdateEvent registers a handler for channel state change notifications.
// The handler will be called whenever a channel's state changes (created, resized,
// closed, or challenged).
//
// Only one handler can be registered per event type. Calling this method
// multiple times will replace the previous handler.
//
// Parameters:
//   - handler: Callback function to handle channel updates
//
// Example usage:
//
//	client.HandleChannelUpdateEvent(func(ctx context.Context, notif ChannelUpdateNotification, sigs []sign.Signature) {
//	    fmt.Printf("Channel %s updated - Status: %s\n", notif.ChannelID, notif.Status)
//	    if notif.Status == ChannelStatusChallenged {
//	        // Handle challenge scenario
//	    }
//	})
func (c *Client) HandleChannelUpdateEvent(handler ChannelUpdateEventHandler) {
	c.setEventHandler(ChannelUpdateEvent, handler)
}

// handleChannelUpdateEvent is an internal method that processes channel update events.
// It retrieves the registered handler and invokes it with the notification data.
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

// HandleTransferEvent registers a handler for transfer notifications.
// The handler will be called whenever a transfer affects the user's account,
// including both incoming and outgoing transfers.
//
// Only one handler can be registered per event type. Calling this method
// multiple times will replace the previous handler.
//
// Parameters:
//   - handler: Callback function to handle transfer notifications
//
// Example usage:
//
//	client.HandleTransferEvent(func(ctx context.Context, notif TransferNotification, sigs []sign.Signature) {
//	    for _, tx := range notif.Transactions {
//	        direction := "sent"
//	        if tx.ToAccount == myAccount {
//	            direction = "received"
//	        }
//	        fmt.Printf("Transfer %s: %s %s\n", direction, tx.Amount, tx.Asset)
//	    }
//	})
func (c *Client) HandleTransferEvent(handler TransferEventHandler) {
	c.setEventHandler(TransferEvent, handler)
}

// handleTransferEvent is an internal method that processes transfer events.
// It retrieves the registered handler and invokes it with the notification data.
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

// setEventHandler stores a handler for a specific event type.
// This method is thread-safe and can be called concurrently.
//
// Parameters:
//   - event: The event type to handle
//   - handler: The handler function (must match the appropriate type)
func (c *Client) setEventHandler(event Event, handler any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.eventHandlers[event] = handler
}

// getEventHandler retrieves the handler for a specific event type.
// This method is thread-safe and can be called concurrently.
//
// Parameters:
//   - event: The event type to get the handler for
//
// Returns:
//   - any: The handler function or nil if no handler is registered
func (c *Client) getEventHandler(event Event) any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.eventHandlers[event]
}
