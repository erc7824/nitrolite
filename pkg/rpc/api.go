// Package rpc provides the RPC API types for the Nitrolite Node service.
//
// This file implements the API request and response definitions specified in api.yaml
// with versioned types organized by functional groups. All types follow the naming
// convention: {Group}{Version}{Name}{Request|Response}.
package rpc

import "github.com/erc7824/nitrolite/pkg/core"

// ============================================================================
// Channels Group - V1 API
// ============================================================================

// ChannelsV1GetHomeChannelRequest retrieves current on-chain home channel information.
type ChannelsV1GetHomeChannelRequest struct {
	// UserWallet is the user's wallet address
	Wallet string `json:"wallet"`
	// Asset is the asset symbol
	Asset string `json:"asset"`
}

// ChannelsV1GetHomeChannelResponse returns the on-chain channel information.
type ChannelsV1GetHomeChannelResponse struct {
	// Channel is the on-chain channel information
	Channel ChannelV1 `json:"channel"`
}

// ChannelsV1GetEscrowChannelRequest retrieves current on-chain escrow channel information.
type ChannelsV1GetEscrowChannelRequest struct {
	// EscrowChannelID is the escrow channel ID
	EscrowChannelID string `json:"escrow_channel_id"`
}

// ChannelsV1GetEscrowChannelResponse returns the on-chain channel information.
type ChannelsV1GetEscrowChannelResponse struct {
	// Channel is the on-chain channel information
	Channel ChannelV1 `json:"channel"`
}

// ChannelsV1GetChannelsRequest retrieves all channels for a user with optional filtering.
type ChannelsV1GetChannelsRequest struct {
	// Wallet filters by user's wallet address
	Wallet string `json:"wallet"`
	// Status filters by status
	Status *string `json:"status,omitempty"`
	// Asset filters by asset
	Asset *string `json:"asset,omitempty"`
	// Pagination contains pagination parameters (offset, limit, sort)
	Pagination *PaginationParamsV1 `json:"pagination,omitempty"`
}

// ChannelsV1GetChannelsResponse returns the list of channels.
type ChannelsV1GetChannelsResponse struct {
	// Channels is the list of channels
	Channels []ChannelV1 `json:"channels"`
	// Metadata contains pagination information
	Metadata PaginationMetadataV1 `json:"metadata"`
}

// ChannelsV1GetLatestStateRequest retrieves the current state of the user stored on the Node.
type ChannelsV1GetLatestStateRequest struct {
	// UserWallet is the user's wallet address
	Wallet string `json:"wallet"`
	// Asset is the asset symbol
	Asset string `json:"asset"`
	// OnlySigned can be enabled to get the latest signed state to know what is the current pending transition
	OnlySigned bool `json:"only_signed"`
}

// ChannelsV1GetLatestStateResponse returns the current state of the user.
type ChannelsV1GetLatestStateResponse struct {
	// State is the current state of the user
	State StateV1 `json:"state"`
}

// ChannelsV1GetStatesRequest retrieves state history for a user with optional filtering.
type ChannelsV1GetStatesRequest struct {
	// Wallet is the user's wallet address
	Wallet string `json:"wallet"`
	// Asset filters by asset symbol
	Asset string `json:"asset"`
	// Epoch filters by user epoch index
	Epoch *uint64 `json:"epoch,omitempty"`
	// ChannelID filters by Home/Escrow Channel ID
	ChannelID *string `json:"channel_id,omitempty"`
	// OnlySigned returns only signed states
	OnlySigned bool `json:"only_signed"`
	// Pagination contains pagination parameters (offset, limit, sort)
	Pagination *PaginationParamsV1 `json:"pagination,omitempty"`
}

// ChannelsV1GetStatesResponse returns the list of states.
type ChannelsV1GetStatesResponse struct {
	// States is the list of states
	States []StateV1 `json:"states"`
	// Metadata contains pagination information
	Metadata PaginationMetadataV1 `json:"metadata"`
}

// ChannelsV1RequestCreationRequest requests the creation of a channel from Node.
type ChannelsV1RequestCreationRequest struct {
	// State is the state to be submitted
	State StateV1 `json:"state"`
	// ChannelDefinition is the definition of the channel to be created
	ChannelDefinition ChannelDefinitionV1 `json:"channel_definition"`
}

// ChannelsV1RequestCreationResponse returns the Node's signature for the state.
type ChannelsV1RequestCreationResponse struct {
	// Signature is the Node's signature for the state
	Signature string `json:"signature"`
}

// ChannelsV1SubmitStateRequest submits a cross-chain state.
type ChannelsV1SubmitStateRequest struct {
	// State is the state to be submitted
	State StateV1 `json:"state"`
}

// ChannelsV1SubmitStateResponse returns the Node's signature for the state.
type ChannelsV1SubmitStateResponse struct {
	// Signature is the Node's signature for the state
	Signature string `json:"signature"`
}

// ChannelsV1HomeChannelCreatedEvent is emitted when a home channel is created.
type ChannelsV1HomeChannelCreatedEvent struct {
	// Channel is the created home channel information
	Channel ChannelV1 `json:"channel"`
	// InitialState is the initial state of the home channel
	InitialState StateV1 `json:"initial_state"`
}

// ============================================================================
// App Sessions Group - V1 API
// ============================================================================

// AppSessionsV1SubmitDepositStateRequest submits an application session state update.
type AppSessionsV1SubmitDepositStateRequest struct {
	// AppStateUpdate is the application session state update to be submitted
	AppStateUpdate AppStateUpdateV1 `json:"app_state_update"`
	// AppStateSignatures is the list of participant signatures for the app state update
	AppStateSignatures []string `json:"app_state_signatures"`
	// SigQuorum is the signature quorum for the application session
	SigQuorum uint64 `json:"sig_quorum"`
	// UserState is the user state associated with the application session update
	UserState StateV1 `json:"user_state"`
}

// AppSessionsV1SubmitDepositStateResponse returns the Node's signature for the deposit state.
type AppSessionsV1SubmitDepositStateResponse struct {
	// StateNodeSig is the Node's signature for the deposit state
	StateNodeSig string `json:"signature"`
}

// AppSessionsV1SubmitAppStateRequest submits an application session state update.
type AppSessionsV1SubmitAppStateRequest struct {
	// AppStateUpdate is the application session state update to be submitted
	AppStateUpdate AppStateUpdateV1 `json:"app_state_update"`
	// Signatures is the signature quorum for the application session
	Signatures []string `json:"signatures"`
}

// AppSessionsV1SubmitAppStateResponse returns the Node's signature for the new User state.
type AppSessionsV1SubmitAppStateResponse struct {
	// Signature is the Node's signature for the new User state
	Signature string `json:"signature"`
}

// AppSessionsV1GetAppDefinitionRequest retrieves the application definition for a specific app session.
type AppSessionsV1GetAppDefinitionRequest struct {
	// AppSessionID is the application session ID
	AppSessionID string `json:"app_session_id"`
}

// AppSessionsV1GetAppDefinitionResponse returns the application definition.
type AppSessionsV1GetAppDefinitionResponse struct {
	// Definition is the application definition
	Definition AppDefinitionV1 `json:"definition"`
}

// AppSessionsV1GetAppSessionsRequest lists all application sessions for a participant with optional filtering.
type AppSessionsV1GetAppSessionsRequest struct {
	// Participant filters by participant wallet address
	Participant *string `json:"participant,omitempty"`
	// Status filters by status (open/closed)
	Status *string `json:"status,omitempty"`
	// Pagination contains pagination parameters (offset, limit, sort)
	Pagination *PaginationParamsV1 `json:"pagination,omitempty"`
}

// AppSessionsV1GetAppSessionsResponse returns the list of application sessions.
type AppSessionsV1GetAppSessionsResponse struct {
	// AppSessions is the list of application sessions
	AppSessions []AppSessionInfoV1 `json:"app_sessions"`
	// Metadata contains pagination information
	Metadata PaginationMetadataV1 `json:"metadata"`
}

// AppSessionsV1CreateAppSessionRequest creates a new application session between participants.
type AppSessionsV1CreateAppSessionRequest struct {
	// Definition is the application definition including participants and quorum
	Definition AppDefinitionV1 `json:"definition"`
	// SessionData is the optional JSON stringified session data
	SessionData string `json:"session_data"`

	Signatures []string `json:"signatures,omitempty"` // Participant signatures for the app session creation
}

// AppSessionsV1CreateAppSessionResponse returns the created application session information.
type AppSessionsV1CreateAppSessionResponse struct {
	// AppSessionID is the created application session ID
	AppSessionID string `json:"app_session_id"`
	// Version is the initial version of the session
	Version string `json:"version"`
	// IsClosed is the status of the session (open)
	IsClosed bool `json:"is_closed"`
}

// AppSessionsV1CloseAppSessionRequest closes an application session and redistributes funds.
type AppSessionsV1CloseAppSessionRequest struct {
	// AppSessionID is the application session ID to close
	AppSessionID string `json:"app_session_id"`
	// Allocations defines the final asset allocations when closing the session
	Allocations []AppAllocationV1 `json:"allocations"`
	// SessionData is the optional final JSON stringified session data
	SessionData *string `json:"session_data,omitempty"`
}

// AppSessionsV1CloseAppSessionResponse returns the closed application session information.
type AppSessionsV1CloseAppSessionResponse struct {
	// AppSessionID is the closed application session ID
	AppSessionID string `json:"app_session_id"`
	// Version is the final version of the session
	Version string `json:"version"`
	// Status is the status of the session (closed)
	Status string `json:"status"`
}

// ============================================================================
// Session Keys Group - V1 API (DRAFT)
// ============================================================================

// SessionKeysV1RegisterRequest initiates session key registration.
type SessionKeysV1RegisterRequest struct {
	// Address is the user wallet address
	Address string `json:"address"`
	// SessionKey is the session key address for delegation
	SessionKey *string `json:"session_key,omitempty"`
	// Application is the application name for analytics
	Application *string `json:"application,omitempty"`
	// Allowances contains asset allowances for the session
	Allowances []AssetAllowanceV1 `json:"allowances,omitempty"`
	// Scope is the permission scope
	Scope *string `json:"scope,omitempty"`
	// ExpiresAt is the session expiration timestamp
	ExpiresAt *uint64 `json:"expires_at,omitempty"`
}

// SessionKeysV1RegisterResponse.
type SessionKeysV1RegisterResponse struct {
}

// SessionKeysV1RevokeSessionKeyRequest revokes a session key by immediately invalidating it.
type SessionKeysV1RevokeSessionKeyRequest struct {
	// SessionKey is the address of the session key to revoke
	SessionKey string `json:"session_key"`
}

// SessionKeysV1RevokeSessionKeyResponse returns the address of the revoked session key.
type SessionKeysV1RevokeSessionKeyResponse struct {
	// SessionKey is the address of the revoked session key
	SessionKey string `json:"session_key"`
}

// SessionKeysV1GetSessionKeysRequest retrieves all active session keys for the authenticated user.
type SessionKeysV1GetSessionKeysRequest struct {
	// Wallet is the user's wallet address
	Wallet string `json:"wallet"`
}

// SessionKeysV1GetSessionKeysResponse returns the list of active session keys.
type SessionKeysV1GetSessionKeysResponse struct {
	// SessionKeys is the list of active session keys
	SessionKeys []SessionKeyV1 `json:"session_keys"`
}

// ============================================================================
// User Group - V1 API
// ============================================================================

// UserV1GetBalancesRequest retrieves the balances of the user in YN.
type UserV1GetBalancesRequest struct {
	// Wallet is the user's wallet address
	Wallet string `json:"wallet"`
}

// UserV1GetBalancesResponse returns the list of asset balances.
type UserV1GetBalancesResponse struct {
	// Balances is the list of asset balances
	Balances []BalanceEntryV1 `json:"balances"`
}

// UserV1GetTransactionsRequest retrieves ledger transaction history with optional filtering.
type UserV1GetTransactionsRequest struct {
	// Wallet filters by user's wallet address
	Wallet string `json:"wallet"`
	// Asset filters by asset symbol
	Asset *string `json:"asset,omitempty"`
	// TxType filters by transaction type
	TxType *core.TransactionType `json:"tx_type,omitempty"`
	// Pagination contains pagination parameters (offset, limit, sort)
	Pagination *PaginationParamsV1 `json:"pagination,omitempty"`
	// FromTime is the start time filter (Unix timestamp)
	FromTime *uint64 `json:"from_time,omitempty"`
	// ToTime is the end time filter (Unix timestamp)
	ToTime *uint64 `json:"to_time,omitempty"`
}

// UserV1GetTransactionsResponse returns the list of transactions.
type UserV1GetTransactionsResponse struct {
	// Transactions is the list of transactions
	Transactions []TransactionV1 `json:"transactions"`
	// Metadata contains pagination information
	Metadata PaginationMetadataV1 `json:"metadata"`
}

// ============================================================================
// Node Group - V1 API
// ============================================================================

// NodeV1PingRequest is a simple connectivity check.
type NodeV1PingRequest struct{}

// NodeV1PingResponse is the response to a ping request.
type NodeV1PingResponse struct{}

// NodeV1GetConfigRequest retrieves broker configuration and supported networks.
type NodeV1GetConfigRequest struct{}

// NodeV1GetConfigResponse returns the broker configuration.
type NodeV1GetConfigResponse struct {
	// NodeAddress is the node wallet address
	NodeAddress string `json:"node_address"`
	// Blockchains is the list of supported networks
	Blockchains []BlockchainInfoV1 `json:"blockchains"`
}

// NodeV1GetAssetsRequest retrieves all supported assets with optional chain filter.
type NodeV1GetAssetsRequest struct {
	// ChainID filters by blockchain network ID
	ChainID *uint32 `json:"chain_id,omitempty"`
}

// NodeV1GetAssetsResponse returns the list of supported assets.
type NodeV1GetAssetsResponse struct {
	// Assets is the list of supported assets
	Assets []AssetV1 `json:"assets"`
}
