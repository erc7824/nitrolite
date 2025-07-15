package main

// RPCMethod represents RPC method names as string constants
type RPCMethod string

const (
	// Authentication methods
	RPCMethodAuthRequest   RPCMethod = "auth_request"
	RPCMethodAuthChallenge RPCMethod = "auth_challenge"
	RPCMethodAuthVerify    RPCMethod = "auth_verify"

	// Public methods
	RPCMethodError                 RPCMethod = "error"
	RPCMethodGetConfig             RPCMethod = "get_config"
	RPCMethodGetAssets             RPCMethod = "get_assets"
	RPCMethodGetAppDefinition      RPCMethod = "get_app_definition"
	RPCMethodGetAppSessions        RPCMethod = "get_app_sessions"
	RPCMethodGetChannels           RPCMethod = "get_channels"
	RPCMethodGetLedgerEntries      RPCMethod = "get_ledger_entries"
	RPCMethodGetLedgerTransactions RPCMethod = "get_ledger_transactions"
	RPCMethodPing                  RPCMethod = "ping"
	RPCMethodPong                  RPCMethod = "pong"

	// Private methods (require authentication)
	RPCMethodGetUserTag        RPCMethod = "get_user_tag"
	RPCMethodGetLedgerBalances RPCMethod = "get_ledger_balances"
	RPCMethodGetRPCHistory     RPCMethod = "get_rpc_history"

	// Channel methods
	RPCMethodResizeChannel RPCMethod = "resize_channel"
	RPCMethodCloseChannel  RPCMethod = "close_channel"

	// App session methods
	RPCMethodTransfer         RPCMethod = "transfer"
	RPCMethodCreateAppSession RPCMethod = "create_app_session"
	RPCMethodSubmitAppState   RPCMethod = "submit_app_state"
	RPCMethodCloseAppSession  RPCMethod = "close_app_session"

	// Notification methods
	RPCMethodAssets               RPCMethod = "assets"
	RPCMethodMessage              RPCMethod = "message"
	RPCMethodBalanceUpdate        RPCMethod = "bu"
	RPCMethodChannelsUpdate       RPCMethod = "channels"
	RPCMethodChannelUpdate        RPCMethod = "cu"
	RPCMethodTransferNotification RPCMethod = "tr"
)

// String returns the string representation of the RPC method
func (r RPCMethod) String() string {
	return string(r)
}

// AllRPCMethods returns all defined RPC methods
func AllRPCMethods() []RPCMethod {
	return []RPCMethod{
		RPCMethodAuthRequest,
		RPCMethodAuthChallenge,
		RPCMethodAuthVerify,
		RPCMethodError,
		RPCMethodGetConfig,
		RPCMethodGetAssets,
		RPCMethodGetAppDefinition,
		RPCMethodGetAppSessions,
		RPCMethodGetChannels,
		RPCMethodGetLedgerEntries,
		RPCMethodGetLedgerTransactions,
		RPCMethodPing,
		RPCMethodPong,
		RPCMethodGetUserTag,
		RPCMethodGetLedgerBalances,
		RPCMethodGetRPCHistory,
		RPCMethodResizeChannel,
		RPCMethodCloseChannel,
		RPCMethodTransfer,
		RPCMethodCreateAppSession,
		RPCMethodSubmitAppState,
		RPCMethodCloseAppSession,
		RPCMethodAssets,
		RPCMethodMessage,
		RPCMethodBalanceUpdate,
		RPCMethodChannelsUpdate,
		RPCMethodChannelUpdate,
		RPCMethodTransferNotification,
	}
}
