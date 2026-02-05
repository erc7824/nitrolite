package sdk

import (
	"context"
	"fmt"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/shopspring/decimal"
)

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
//	for _, session := range sessions {
//	    fmt.Printf("Session %s: %d participants\n", session.AppSessionID, len(session.Participants))
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

	appSessions, err := transformAppSessions(resp.AppSessions)
	if err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to transform app sessions: %w", err)
	}

	return appSessions, transformPaginationMetadata(resp.Metadata), nil
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
//	fmt.Printf("App: %s, Quorum: %d\n", def.Application, def.Quorum)
func (c *Client) GetAppDefinition(ctx context.Context, appSessionID string) (*app.AppDefinitionV1, error) {
	if appSessionID == "" {
		return nil, fmt.Errorf("app session ID required")
	}
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
//	def := app.AppDefinitionV1{
//	    Application: "chess-v1",
//	    Participants: []app.AppParticipantV1{...},
//	    Quorum: 2,
//	    Nonce: 1,
//	}
//	sessionID, version, status, err := client.CreateAppSession(ctx, def, "{}", []string{"sig1", "sig2"})
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

// SubmitAppSessionDeposit submits a deposit to an app session.
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
//	appUpdate := app.AppStateUpdateV1{
//	    AppSessionID: "session123",
//	    Intent: app.AppStateUpdateIntentDeposit,
//	    Version: 2,
//	    Allocations: []app.AppAllocationV1{...},
//	}
//	nodeSig, err := client.SubmitAppSessionDeposit(ctx, appUpdate, []string{"sig1"}, userState)
func (c *Client) SubmitAppSessionDeposit(ctx context.Context, appStateUpdate app.AppStateUpdateV1, quorumSigs []string, asset string, depositAmount decimal.Decimal) (string, error) {
	// TODO: Would be good to only have appStateUpdate and quorumSigs here, as userState can be built inside.
	appUpdate := transformAppStateUpdateToRPC(appStateUpdate)

	currentState, err := c.GetLatestState(ctx, c.GetUserAddress(), asset, false)
	if err != nil {
		return "", fmt.Errorf("failed to get latest state: %w", err)
	}

	nextState := currentState.NextState()

	_, err = nextState.ApplyCommitTransition(appUpdate.AppSessionID, depositAmount)
	if err != nil {
		return "", fmt.Errorf("failed to apply commit transition: %w", err)
	}

	stateSig, err := c.SignState(nextState)
	if err != nil {
		return "", fmt.Errorf("failed to sign state: %w", err)
	}

	nextState.UserSig = &stateSig
	req := rpc.AppSessionsV1SubmitDepositStateRequest{
		AppStateUpdate: appUpdate,
		QuorumSigs:     quorumSigs,
		UserState:      transformStateToRPC(*nextState),
	}

	resp, err := c.rpcClient.AppSessionsV1SubmitDepositState(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to submit deposit state: %w", err)
	}
	return resp.StateNodeSig, nil
}

// SubmitAppState submits an app session state update.
// This method handles operate, withdraw, and close intents.
// For deposits, use SubmitAppSessionDeposit instead.
//
// Parameters:
//   - appStateUpdate: The app state update (intent: operate, withdraw, or close)
//   - quorumSigs: Participant signatures for the app state update
//
// Returns:
//   - Error if the request fails
//
// Example:
//
//	appUpdate := app.AppStateUpdateV1{
//	    AppSessionID: "session123",
//	    Intent: app.AppStateUpdateIntentOperate,
//	    Version: 3,
//	    Allocations: []app.AppAllocationV1{...},
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
//	updates := []app.SignedAppStateUpdateV1{...}
//	batchID, err := client.RebalanceAppSessions(ctx, updates)
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
