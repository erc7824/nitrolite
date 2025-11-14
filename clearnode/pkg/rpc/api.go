// Package rpc provides the RPC API types for the ClearNode broker service.
//
// The ClearNode RPC API enables interaction with the broker for payment channel
// management, virtual application sessions, and ledger operations. All monetary
// amounts use decimal.Decimal for arbitrary precision arithmetic.
package rpc

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

// ===========================================================================
// Protocol Versioning
// ============================================================================

// Version represents the protocol version.
// This is used to provide backward compatibility as the API evolves.
type Version string

const (
	// VersionNitroRPCv0_2 is the initial supported version of the NitroRPC protocol.
	VersionNitroRPCv0_2 Version = "NitroRPC/0.2"
	// VersionNitroRPCv0_4 introduces intents for application session state updates.
	VersionNitroRPCv0_4 Version = "NitroRPC/0.4"
)

var (
	// supportedProtocolVersions lists all protocol versions supported by the broker.
	supportedProtocolVersions = map[string]bool{
		VersionNitroRPCv0_2.String(): true,
		VersionNitroRPCv0_4.String(): true,
	}
)

// String returns the string representation of the version.
func (v Version) String() string {
	return string(v)
}

// IsSupportedVersion checks if the given version is supported by the broker.
func IsSupportedVersion(version Version) bool {
	return supportedProtocolVersions[version.String()]
}

// ============================================================================
// RPC Method Constants
// ============================================================================

// Method represents an RPC method name that can be called on the ClearNode broker.
// These constants define all available methods in the RPC API.
type Method string

const (
	// PingMethod is a simple method to check connectivity and liveness.
	PingMethod Method = "ping"
	// PongMethod is the response to a ping request.
	PongMethod Method = "pong"
	// ErrorMethod is the identifier for error responses.
	ErrorMethod Method = "error"
	// GetConfigMethod returns the broker configuration and supported networks.
	GetConfigMethod Method = "get_config"
	// GetAssetsMethod returns the list of supported assets/tokens.
	GetAssetsMethod Method = "get_assets"
	// GetAppDefinitionMethod returns the definition of an application session.
	GetAppDefinitionMethod Method = "get_app_definition"
	// GetAppSessionsMethod returns a list of application sessions.
	GetAppSessionsMethod Method = "get_app_sessions"
	// GetChannelsMethod returns a list of payment channels.
	GetChannelsMethod Method = "get_channels"
	// GetLedgerEntriesMethod returns ledger entries (double-entry bookkeeping).
	GetLedgerEntriesMethod Method = "get_ledger_entries"
	// GetLedgerTransactionsMethod returns ledger transactions.
	GetLedgerTransactionsMethod Method = "get_ledger_transactions"
	// AuthRequestMethod initiates authentication with challenge generation.
	AuthRequestMethod Method = "auth_request"
	// AuthChallengeMethod is the response to an auth request with the challenge.
	AuthChallengeMethod Method = "auth_challenge"
	// AuthVerifyMethod verifies authentication via signature or JWT.
	AuthVerifyMethod Method = "auth_verify"
	// GetUserTagMethod returns the human-readable tag for a wallet (auth required).
	GetUserTagMethod Method = "get_user_tag"
	// GetLedgerBalancesMethod returns account balances (auth required).
	GetLedgerBalancesMethod Method = "get_ledger_balances"
	// GetRPCHistoryMethod returns RPC call history (auth required).
	GetRPCHistoryMethod Method = "get_rpc_history"
	// GetSessionKeysMethod returns session keys with allowances (auth required).
	GetSessionKeysMethod Method = "get_session_keys"
	// CreateChannelMethod creates a new payment channel (auth required).
	CreateChannelMethod Method = "create_channel"
	// ResizeChannelMethod resizes an existing channel (auth required).
	ResizeChannelMethod Method = "resize_channel"
	// CloseChannelMethod closes a payment channel (auth required).
	CloseChannelMethod Method = "close_channel"
	// TransferMethod transfers funds between accounts (auth required).
	TransferMethod Method = "transfer"
	// CreateAppSessionMethod creates a virtual application session (auth required).
	CreateAppSessionMethod Method = "create_app_session"
	// SubmitAppStateMethod updates application state (auth required).
	SubmitAppStateMethod Method = "submit_app_state"
	// CloseAppSessionMethod closes an application session (auth required).
	CloseAppSessionMethod Method = "close_app_session"
	// CleanupSessionKeyCacheMethod clears the session key cache (test mode only).
	CleanupSessionKeyCacheMethod Method = "cleanup_session_key_cache"
)

// String returns the string representation of the method.
func (m Method) String() string {
	return string(m)
}

// ============================================================================
// Notification Event Types
// ============================================================================

// Event represents a notification event type sent by the broker.
// Events are unsolicited notifications sent to connected clients.
type Event string

const (
	// BalanceUpdateEvent notifies clients of balance changes.
	BalanceUpdateEvent Event = "bu"
	// ChannelUpdateEvent notifies clients of channel state changes.
	ChannelUpdateEvent Event = "cu"
	// TransferEvent notifies clients of incoming transfers.
	TransferEvent Event = "tr"
	// AppSessionUpdateEvent notifies clients of app session state changes.
	AppSessionUpdateEvent Event = "asu"
)

// String returns the string representation of the event.
func (e Event) String() string {
	return string(e)
}

// ============================================================================
// Public API Types - No Authentication Required
// ============================================================================

// GetConfigResponse returns the broker's configuration including supported networks.
// This is typically the first call made by clients to discover available chains.
type GetConfigResponse BrokerConfig

// GetAssetsRequest filters the list of supported assets.
type GetAssetsRequest struct {
	// ChainID optionally filters assets by blockchain network
	ChainID *uint32 `json:"chain_id,omitempty"`
}

// GetAssetsResponse contains the list of supported assets/tokens.
type GetAssetsResponse struct {
	Assets []Asset `json:"assets"`
}

// GetAppDefinitionRequest queries for a specific application session's definition.
type GetAppDefinitionRequest struct {
	// AppSessionID is the unique identifier for the application session
	AppSessionID string `json:"app_session_id"`
}

// GetAppDefinitionResponse returns the application's protocol definition.
type GetAppDefinitionResponse AppDefinition

// GetAppSessionsRequest queries for application sessions with optional filters.
type GetAppSessionsRequest struct {
	ListOptions
	// Participant filters sessions by wallet address
	Participant string `json:"participant,omitempty"`
	// Status filters by session state (e.g., "open", "closed")
	Status string `json:"status,omitempty"`
}

// GetAppSessionsResponse contains the filtered list of application sessions.
type GetAppSessionsResponse struct {
	AppSessions []AppSession `json:"app_sessions"`
}

// GetChannelsRequest queries for payment channels with optional filters.
type GetChannelsRequest struct {
	ListOptions
	// Participant filters channels by wallet address
	Participant string `json:"participant,omitempty"`
	// Status filters by channel state (e.g., "open", "closed", "challenged")
	Status string `json:"status,omitempty"`
}

// GetChannelsResponse contains the filtered list of payment channels.
type GetChannelsResponse struct {
	Channels []Channel `json:"channels"`
}

// GetLedgerEntriesRequest queries double-entry bookkeeping entries.
type GetLedgerEntriesRequest struct {
	ListOptions
	// AccountID filters entries by the account identifier
	AccountID string `json:"account_id,omitempty"`
	// Asset filters by token/asset symbol
	Asset string `json:"asset,omitempty"`
	// Wallet filters by participant wallet address
	Wallet string `json:"wallet,omitempty"`
}

// GetLedgerEntriesResponse contains the filtered ledger entries.
type GetLedgerEntriesResponse struct {
	LedgerEntries []LedgerEntry `json:"ledger_entries"`
}

// GetLedgerTransactionsRequest queries ledger transactions with optional filters.
type GetLedgerTransactionsRequest struct {
	ListOptions
	// AccountID filters transactions by account
	AccountID string `json:"account_id,omitempty"`
	// Asset filters by token/asset symbol
	Asset string `json:"asset,omitempty"`
	// TxType filters by transaction type
	TxType string `json:"tx_type,omitempty"`
}

// GetLedgerTransactionsResponse contains the filtered ledger transactions.
type GetLedgerTransactionsResponse struct {
	LedgerTransactions []LedgerTransaction `json:"ledger_transactions"`
}

// ============================================================================
// Authentication API Types
// ============================================================================

// AuthRequestRequest initiates wallet-based authentication flow.
type AuthRequestRequest struct {
	// Address is the wallet address requesting authentication
	Address string `json:"address"`
	// SessionKey is a unique key for this authentication session
	SessionKey string `json:"session_key"`
	// Application identifies the application requesting authentication
	Application string `json:"application"`
	// Allowances define spending limits for the authenticated session
	Allowances []Allowance `json:"allowances"`
	// ExpiresAt defines when the authentication expires (Unix timestamp)
	ExpiresAt uint64 `json:"expires_at"`
	// Scope defines the permission scope for the session
	Scope string `json:"scope"`
}

// AuthRequestResponse contains the challenge for wallet signature.
type AuthRequestResponse struct {
	// ChallengeMessage is the UUID that must be signed by the wallet
	ChallengeMessage uuid.UUID `json:"challenge_message"`
}

// AuthSigVerifyRequest verifies a signed authentication challenge.
type AuthSigVerifyRequest struct {
	// Challenge is the UUID that was signed
	Challenge uuid.UUID `json:"challenge"`
}

// AuthSigVerifyResponse contains the authentication result and session token.
type AuthSigVerifyResponse struct {
	// Address is the authenticated wallet address
	Address string `json:"address"`
	// SessionKey echoes the session key from the request
	SessionKey string `json:"session_key"`
	// JwtToken is the JWT for authenticated API calls
	JwtToken string `json:"jwt_token"`
	// Success indicates if authentication succeeded
	Success bool `json:"success"`
}

// AuthJWTVerifyRequest verifies an existing JWT token.
type AuthJWTVerifyRequest struct {
	// JWT is the token to verify
	JWT string `json:"jwt"`
}

// AuthJWTVerifyResponse contains the JWT verification result.
type AuthJWTVerifyResponse struct {
	// Address is the wallet address from the JWT
	Address string `json:"address"`
	// SessionKey is the session identifier from the JWT
	SessionKey string `json:"session_key"`
	// Success indicates if the JWT is valid
	Success bool `json:"success"`
}

// GetSessionKeysRequest queries for session keys associated with the authenticated wallet.
type GetSessionKeysRequest struct {
	// No parameters - returns all session keys for the authenticated user
}

// GetSessionKeysResponse contains the list of active session keys.
type GetSessionKeysResponse struct {
	SessionKeys []SessionKeyResponse `json:"session_keys"`
}

// SessionKeyResponse represents a single session key with its allowances and usage.
type SessionKeyResponse struct {
	// ID is the internal database identifier
	ID uint `json:"id"`
	// SessionKey is the public key/address of the session key
	SessionKey string `json:"session_key"`
	// Application is the name or identifier of the application this key is for
	Application string `json:"application"`
	// Allowances contains spending limits per asset with usage tracking
	Allowances []AllowanceUsage `json:"allowances"`
	// Scope defines the permission scope for the session (e.g., "app.create", "ledger.readonly")
	Scope string `json:"scope,omitempty"`
	// ExpiresAt is when the session key expires
	ExpiresAt time.Time `json:"expires_at"`
	// CreatedAt is when the session key was created
	CreatedAt time.Time `json:"created_at"`
}

// AllowanceUsage represents an asset allowance with usage tracking.
type AllowanceUsage struct {
	// Asset is the token/asset symbol
	Asset string `json:"asset"`
	// Allowance is the total spending limit for this asset
	Allowance decimal.Decimal `json:"allowance"`
	// Used is how much of the allowance has been spent
	Used decimal.Decimal `json:"used"`
}

// RevokeSessionKeyRequest contains the parameters for revoking a session key.
type RevokeSessionKeyRequest struct {
	// SessionKey is the address of the session key to revoke
	SessionKey string `json:"session_key"`
}

// RevokeSessionKeyResponse indicates successful revocation of a session key.
type RevokeSessionKeyResponse struct {
	// SessionKey is the address of the revoked session key
	SessionKey string `json:"session_key"`
}

// ============================================================================
// Private API Types - Authentication Required
// ============================================================================

// GetUserTagResponse returns the human-readable tag for a wallet address.
type GetUserTagResponse struct {
	Tag string `json:"tag"`
}

// GetLedgerBalancesRequest queries account balances.
type GetLedgerBalancesRequest struct {
	// AccountID optionally filters balances by account
	AccountID string `json:"account_id,omitempty"`
}

// GetLedgerBalancesResponse contains the account balances by asset.
type GetLedgerBalancesResponse struct {
	LedgerBalances []LedgerBalance `json:"ledger_balances"`
}

// GetRPCHistoryRequest queries the RPC call history for the authenticated user.
type GetRPCHistoryRequest struct {
	ListOptions
}

// GetRPCHistoryResponse contains the RPC call history entries.
type GetRPCHistoryResponse struct {
	RPCEntries []HistoryEntry `json:"rpc_entries"`
}

// CreateChannelRequest opens a new payment channel with the broker.
type CreateChannelRequest struct {
	// ChainID identifies the blockchain network
	ChainID uint32 `json:"chain_id" validate:"required"`
	// Token is the asset/token address for the channel
	Token string `json:"token" validate:"required"`
	// SessionKey optionally specifies a custom session identifier
	SessionKey *string `json:"session_key,omitempty" validate:"omitempty"`
}

// CreateChannelResponse contains the created channel details and initial state.
type CreateChannelResponse ChannelOperationResponse

// ResizeChannelRequest modifies the funding in an existing channel.
type ResizeChannelRequest struct {
	// ChannelID identifies the channel to resize
	ChannelID string `json:"channel_id" validate:"required"`
	// AllocateAmount adds funds to the channel (mutually exclusive with ResizeAmount)
	AllocateAmount *decimal.Decimal `json:"allocate_amount,omitempty" validate:"omitempty,required_without=ResizeAmount,bigint"`
	// ResizeAmount sets the new total channel size (mutually exclusive with AllocateAmount)
	ResizeAmount *decimal.Decimal `json:"resize_amount,omitempty" validate:"omitempty,required_without=AllocateAmount,bigint"`
	// FundsDestination is where to send funds if reducing channel size
	FundsDestination string `json:"funds_destination" validate:"required"`
}

// ResizeChannelResponse contains the updated channel state after resizing.
type ResizeChannelResponse ChannelOperationResponse

// CloseChannelRequest specifies channel closure parameters.
type CloseChannelRequest struct {
	// ChannelID identifies the channel to close
	ChannelID string `json:"channel_id"`
	// FundsDestination is where to send the channel funds
	FundsDestination string `json:"funds_destination"`
}

// CloseChannelResponse contains the final channel state after closure.
type CloseChannelResponse ChannelOperationResponse

// TransferRequest moves funds between accounts or to external addresses.
type TransferRequest struct {
	// Destination is the recipient wallet address
	Destination string `json:"destination"`
	// DestinationUserTag is the recipient's human-readable tag
	DestinationUserTag string `json:"destination_user_tag"`
	// Allocations specifies amounts per asset to transfer
	Allocations []TransferAllocation `json:"allocations"`
}

// TransferResponse contains the ledger transactions created by the transfer.
type TransferResponse struct {
	Transactions []LedgerTransaction `json:"transactions"`
}

// CreateAppSessionRequest starts a new virtual application session.
type CreateAppSessionRequest struct {
	// Definition specifies the application protocol and participants
	Definition AppDefinition `json:"definition"`
	// Allocations defines the initial asset distribution
	Allocations []AppAllocation `json:"allocations"`
	// SessionData contains application-specific state data
	SessionData *string `json:"session_data"`
}

// CreateAppSessionResponse contains the created application session.
type CreateAppSessionResponse AppSession

// SubmitAppStateRequest updates an application session's state.
type SubmitAppStateRequest struct {
	// AppSessionID identifies the session to update
	AppSessionID string `json:"app_session_id"`
	// Intent indicates the purpose of the state update (operate, deposit, withdraw)
	// Required since protocol version: NitroRPC/0.4
	Intent AppSessionIntent `json:"intent"`
	// Version is the new state version number
	// Required since protocol version: NitroRPC/0.4
	Version uint64 `json:"version"`
	// Allocations defines the new asset distribution
	Allocations []AppAllocation `json:"allocations"`
	// SessionData contains the new application state
	SessionData *string `json:"session_data"`
}

// SubmitAppStateResponse contains the updated application session.
type SubmitAppStateResponse AppSession

// CloseAppSessionRequest specifies application session closure parameters.
type CloseAppSessionRequest struct {
	// AppSessionID identifies the session to close
	AppSessionID string `json:"app_session_id"`
	// SessionData contains the final application state
	SessionData *string `json:"session_data"`
	// Allocations defines the final asset distribution
	Allocations []AppAllocation `json:"allocations"`
}

// CloseAppSessionResponse contains the closed application session.
type CloseAppSessionResponse AppSession

// ============================================================================
// Notification Types - Unsolicited Events
// ============================================================================

// BalanceUpdateNotification is sent when account balances change.
// This notification is triggered by transfers, channel operations, or app session updates.
type BalanceUpdateNotification struct {
	// BalanceUpdates contains the new balances for affected accounts
	BalanceUpdates []LedgerBalance `json:"balance_updates"`
}

// ChannelUpdateNotification is sent when a channel's state changes.
// This includes channel creation, resizing, closure, or challenge events.
// The notification contains the full updated channel information.
type ChannelUpdateNotification Channel

// TransferNotification is sent when a transfer affects the user's account.
// This includes both incoming and outgoing transfers.
type TransferNotification struct {
	// Transactions contains the ledger transactions for the transfer
	Transactions []LedgerTransaction `json:"transactions"`
}

// AppSessionUpdateNotification is sent when an application session's state changes.
// This includes session creation, state updates, and session closure.
type AppSessionUpdateNotification struct {
	AppSession AppSession `json:"app_session"`
	// ParticipantAllocations contains each participant's asset allocations
	ParticipantAllocations []AppAllocation `json:"participant_allocations"`
}

// ============================================================================
// Common Types - Shared across multiple API calls
// ============================================================================

// ListOptions provides pagination and sorting for list endpoints.
type ListOptions struct {
	// Offset is the number of items to skip (for pagination)
	Offset uint32 `json:"offset,omitempty"`
	// Limit is the maximum number of items to return
	Limit uint32 `json:"limit,omitempty"`
	// Sort specifies the sort order (asc/desc)
	Sort *SortType `json:"sort,omitempty"`
}

// SortType defines the sort order for list results.
type SortType string

const (
	// SortTypeAscending sorts results in ascending order
	SortTypeAscending SortType = "asc"
	// SortTypeDescending sorts results in descending order
	SortTypeDescending SortType = "desc"
)

// ToString converts the sort type to uppercase string representation.
func (s SortType) ToString() string {
	return strings.ToUpper(string(s))
}

// ============================================================================
// Configuration Types
// ============================================================================

// BrokerConfig contains the broker's configuration and supported networks.
type BrokerConfig struct {
	// BrokerAddress is the wallet address of the broker
	BrokerAddress string `json:"broker_address"`
	// Networks lists all supported blockchain networks
	Blockchains []BlockchainInfo `json:"networks"` // TODO: rename to "blockchains"
}

// BlockchainInfo describes a supported blockchain network.
type BlockchainInfo struct {
	// ID is the network's chain identifier
	ID uint32 `json:"chain_id"`
	// Name is the human-readable name of the blockchain
	Name string `json:"name"` // TODO: add to SDK
	// CustodyAddress is the custody contract address
	CustodyAddress string `json:"custody_address"`
	// AdjudicatorAddress is the adjudicator contract address
	AdjudicatorAddress string `json:"adjudicator_address"`
}

// ============================================================================
// Asset Types
// ============================================================================

// Asset represents a supported token/asset on a specific chain.
type Asset struct {
	// Token is the token contract address
	Token string `json:"token"`
	// ChainID identifies the blockchain network
	ChainID uint32 `json:"chain_id"`
	// Symbol is the token symbol (e.g., "USDC")
	Symbol string `json:"symbol"`
	// Decimals is the number of decimal places for the token
	Decimals uint8 `json:"decimals"`
}

// Allowance defines spending limits for authenticated sessions.
type Allowance struct {
	// Asset is the token symbol
	Asset string `json:"asset"`
	// Amount is the spending limit
	Amount string `json:"amount"`
}

// TransferAllocation specifies an amount to transfer for a specific asset.
type TransferAllocation struct {
	// AssetSymbol identifies the asset (e.g., "USDC")
	AssetSymbol string `json:"asset"`
	// Amount to transfer
	Amount decimal.Decimal `json:"amount"`
}

// ============================================================================
// Application Session Types
// ============================================================================

// AppDefinition defines the protocol for a multi-party application.
type AppDefinition struct {
	// Application is the identifier of the application
	Application string `json:"application"`
	// Protocol identifies the version of the application protocol
	Protocol Version `json:"protocol"`
	// ParticipantWallets lists the wallet addresses of all participants
	ParticipantWallets []string `json:"participants"`
	// Weights defines the signature weight for each participant
	Weights []int64 `json:"weights"`
	// Quorum is the minimum weight required for consensus
	Quorum uint64 `json:"quorum"`
	// Challenge is the timeout period for disputes (in seconds)
	Challenge uint64 `json:"challenge"`
	// Nonce ensures uniqueness of the application instance
	Nonce uint64 `json:"nonce"`
}

// AppSession represents an active virtual application session.
type AppSession struct {
	// AppSessionID is the unique session identifier
	AppSessionID string `json:"app_session_id"`
	// Application is the name of the application
	Application string `json:"application"`
	// Status indicates the session state (open/closed)
	Status string `json:"status"`
	// ParticipantWallets lists all participants
	ParticipantWallets []string `json:"participants"`
	// SessionData contains application-specific state
	SessionData string `json:"session_data,omitempty"`
	// Protocol identifies the version of the application protocol
	Protocol Version `json:"protocol"`
	// Challenge is the dispute timeout period
	Challenge uint64 `json:"challenge"`
	// Weights defines participant signature weights
	Weights []int64 `json:"weights"`
	// Quorum is the consensus threshold
	Quorum uint64 `json:"quorum"`
	// Version tracks state updates
	Version uint64 `json:"version"`
	// Nonce ensures uniqueness
	Nonce uint64 `json:"nonce"`
	// CreatedAt is when the session was created (RFC3339)
	CreatedAt string `json:"created_at"`
	// UpdatedAt is when the session was last modified (RFC3339)
	UpdatedAt string `json:"updated_at"`
}

// AppAllocation defines asset distribution for a participant in an app session.
type AppAllocation struct {
	// Participant is the recipient's address
	Participant string `json:"participant"`
	// AssetSymbol identifies the asset
	AssetSymbol string `json:"asset"`
	// Amount allocated to the participant
	Amount decimal.Decimal `json:"amount"`
}

// AppSessionIntent indicates the purpose of an application state update.
type AppSessionIntent string

const (
	// AppSessionIntentOperate is for normal application operation
	AppSessionIntentOperate AppSessionIntent = "operate"
	// AppSessionIntentDeposit is for adding funds to the session
	AppSessionIntentDeposit AppSessionIntent = "deposit"
	// AppSessionIntentWithdraw is for removing funds from the session
	AppSessionIntentWithdraw AppSessionIntent = "withdraw"
)

// ============================================================================
// Channel Types
// ============================================================================

// Channel represents a payment channel between a user and the broker.
type Channel struct {
	// ChannelID is the unique channel identifier
	ChannelID string `json:"channel_id"`
	// Participant is the user's wallet address
	Participant string `json:"participant"`
	// Status indicates the channel state
	Status ChannelStatus `json:"status"`
	// Token is the asset contract address
	Token string `json:"token"`
	// Wallet is the participant's wallet address
	Wallet string `json:"wallet"`
	// RawAmount is the total channel capacity (user + broker funds)
	RawAmount decimal.Decimal `json:"amount"`
	// ChainID identifies the blockchain network
	ChainID uint32 `json:"chain_id"`
	// Adjudicator is the dispute resolution contract
	Adjudicator string `json:"adjudicator"`
	// Challenge is the dispute timeout period
	Challenge uint64 `json:"challenge"`
	// Nonce ensures channel uniqueness
	Nonce uint64 `json:"nonce"`
	// Version tracks state updates
	Version uint64 `json:"version"`
	// CreatedAt is when the channel was opened (RFC3339)
	CreatedAt string `json:"created_at"`
	// UpdatedAt is when the channel was last modified (RFC3339)
	UpdatedAt string `json:"updated_at"`
}

// ChannelStatus represents the current state of a payment channel.
type ChannelStatus string

var (
	// ChannelStatusOpen indicates an active channel
	ChannelStatusOpen ChannelStatus = "open"
	// ChannelStatusResizing indicates a channel being resized
	ChannelStatusResizing ChannelStatus = "resizing"
	// ChannelStatusClosed indicates a finalized channel
	ChannelStatusClosed ChannelStatus = "closed"
	// ChannelStatusChallenged indicates a channel in dispute
	ChannelStatusChallenged ChannelStatus = "challenged"
)

// ChannelOperationResponse is returned by channel create/resize/close operations.
type ChannelOperationResponse struct {
	// ChannelID is the channel identifier
	ChannelID string `json:"channel_id"`
	// Channel contains the on-chain channel parameters
	Channel *struct {
		Participants [2]string `json:"participants"`
		Adjudicator  string    `json:"adjudicator"`
		Challenge    uint64    `json:"challenge"`
		Nonce        uint64    `json:"nonce"`
	} `json:"channel,omitempty"`
	// State is the new channel state
	State UnsignedState `json:"state"`
	// StateSignature is the broker's signature on the state
	StateSignature sign.Signature `json:"server_signature"`
}

// ============================================================================
// State Types
// ============================================================================

// UnsignedState represents a channel or application state before signatures.
type UnsignedState struct {
	// Intent indicates the state's purpose
	Intent StateIntent `json:"intent"`
	// Version is the state sequence number
	Version uint64 `json:"version"`
	// Data contains application-specific state data
	Data string `json:"state_data"`
	// Allocations defines asset distribution
	Allocations []StateAllocation `json:"allocations"`
}

// Value implements driver.Valuer interface for database storage.
func (u UnsignedState) Value() (driver.Value, error) {
	return json.Marshal(u)
}

// Scan implements sql.Scanner interface for database retrieval.
func (u *UnsignedState) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into UnsignedState", value)
	}

	return json.Unmarshal(bytes, u)
}

// StateIntent indicates the purpose of a state transition.
type StateIntent uint8

const (
	// StateIntentOperate is for normal application operation
	StateIntentOperate StateIntent = 0
	// StateIntentInitialize is for initial channel funding
	StateIntentInitialize StateIntent = 1
	// StateIntentResize is for channel resize operations
	StateIntentResize StateIntent = 2
	// StateIntentFinalize is for channel closure
	StateIntentFinalize StateIntent = 3
)

// StateAllocation defines asset distribution in a channel/app state.
type StateAllocation struct {
	// Participant is the recipient wallet address
	Participant string `json:"destination"`
	// TokenAddress is the asset contract address
	TokenAddress string `json:"token"`
	// RawAmount allocated to the participant
	RawAmount decimal.Decimal `json:"amount"`
}

// ============================================================================
// Ledger Types - Double-Entry Bookkeeping
// ============================================================================

// LedgerEntry represents a double-entry bookkeeping entry.
// The system follows the DEADCLIC mnemonic for debit/credit rules:
// - DEAD: Debit to increase Expense, Asset, Drawing accounts
// - CLIC: Credit to increase Liability, Income, Capital accounts
type LedgerEntry struct {
	// ID is the unique entry identifier
	ID uint `json:"id"`
	// AccountID identifies the ledger account
	AccountID string `json:"account_id"`
	// AccountType categorizes the account
	AccountType AccountType `json:"account_type"`
	// Asset is the token/asset symbol
	Asset string `json:"asset"`
	// Participant is the wallet address involved
	Participant string `json:"participant"`
	// Credit amount (increases liability/income/capital)
	Credit decimal.Decimal `json:"credit"`
	// Debit amount (increases expense/asset/drawing)
	Debit decimal.Decimal `json:"debit"`
	// CreatedAt is when the entry was recorded
	CreatedAt time.Time `json:"created_at"`
}

// AccountType categorizes ledger accounts following standard accounting principles.
type AccountType uint16

const (
	// AssetDefault represents asset accounts (1000-1999)
	AssetDefault AccountType = 1000
	// LiabilityDefault represents liability accounts (2000-2999)
	LiabilityDefault AccountType = 2000
	// EquityDefault represents equity/capital accounts (3000-3999)
	EquityDefault AccountType = 3000
	// RevenueDefault represents revenue accounts (4000-4999)
	RevenueDefault AccountType = 4000
	// ExpenseDefault represents expense accounts (5000-5999)
	ExpenseDefault AccountType = 5000
)

// LedgerTransaction represents a transfer between accounts.
type LedgerTransaction struct {
	// Id is the unique transaction identifier
	Id uint `json:"id"`
	// TxType categorizes the transaction
	TxType string `json:"tx_type"`
	// FromAccount is the source account
	FromAccount string `json:"from_account"`
	// FromAccountTag is the human-readable source tag
	FromAccountTag string `json:"from_account_tag,omitempty"`
	// ToAccount is the destination account
	ToAccount string `json:"to_account"`
	// ToAccountTag is the human-readable destination tag
	ToAccountTag string `json:"to_account_tag,omitempty"`
	// Asset is the token/asset symbol
	Asset string `json:"asset"`
	// Amount transferred
	Amount decimal.Decimal `json:"amount"`
	// CreatedAt is when the transaction occurred
	CreatedAt time.Time `json:"created_at"`
}

// LedgerBalance represents an account balance for a specific asset.
type LedgerBalance struct {
	// Asset is the token/asset symbol
	Asset string `json:"asset"`
	// Amount is the current balance
	Amount decimal.Decimal `json:"amount"`
}

// ============================================================================
// History Types
// ============================================================================

// HistoryEntry records an RPC method call and response with signatures.
type HistoryEntry struct {
	// ID is the unique entry identifier
	ID uint `json:"id"`
	// Sender is the wallet address that made the call
	Sender string `json:"sender"`
	// ReqID is the request identifier
	ReqID uint64 `json:"req_id"`
	// Method is the RPC method name
	Method string `json:"method"`
	// Params contains the request parameters (JSON)
	Params string `json:"params"`
	// Timestamp is when the call was made (Unix timestamp)
	Timestamp uint64 `json:"timestamp"`
	// ReqSig contains request signatures
	ReqSig []sign.Signature `json:"req_sig"`
	// Result contains the response data (JSON)
	Result string `json:"response"`
	// ResSig contains response signatures
	ResSig []sign.Signature `json:"res_sig"`
}
