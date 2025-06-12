import { Address, Hex } from 'viem';
import { RPCMethod, RequestID, Timestamp, AppDefinition, AppSessionAllocation, ChannelStatus } from '.';

/**
 * Represents a generic RPC message structure that includes common fields.
 * This interface is extended by specific RPC request and response types.
 */
interface GenericRPCMessage {
  requestId: RequestID;
  timestamp?: Timestamp;
  signatures?: Hex[];
}

/**
 * Represents the request parameters for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeRPCRequestParams {
  /** The challenge message to be signed by the client for authentication. */
  challengeMessage: string;
}

/**
 * Represents the request structure for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeRPCRequest extends GenericRPCMessage {
  method: RPCMethod.AuthChallenge;
  params: AuthChallengeRPCRequestParams;
}

/**
 * Represents the request parameters for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRPCRequestParams {
  /** The Ethereum address of the client attempting to authenticate. */
  address: Address;
  /** JSON Web Token for authentication, if provided. */
  jwtToken: string;
  /** The session key address associated with the authentication attempt. */
  sessionKey: Address;
  /** Indicates whether the authentication attempt was successful. */
  success: boolean;
}

/**
 * Represents the request structure for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRPCRequest extends GenericRPCMessage {
  method: RPCMethod.AuthVerify;
  params: AuthVerifyRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_config' RPC method.
 */
export interface GetConfigRPCRequestParams {
  /** The Ethereum address of the broker. */
  broker_address: Address;
  /** List of supported networks and their configurations. */
  networks: {
    /** The name of the network (e.g., "Ethereum", "Polygon"). */
    name: string;
    /** The chain ID of the network. */
    chain_id: number;
    /** The custody contract address for the network. */
    custody_address: Address;
    /** The adjudicator contract address for the network. */
    adjudicator_address: Address;
  }[];
}

/**
 * Represents the request structure for the 'get_config' RPC method.
 */
export interface GetConfigRPCRequest extends GenericRPCMessage {
  method: RPCMethod.GetConfig;
  params: GetConfigRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRPCRequestParams {
  /** Optional participant address to filter balances. If not provided, uses the authenticated wallet address. */
  participant?: Address;
  /** Optional account ID to filter balances. If provided, overrides the participant address. */
  account_id?: string;
}

/**
 * Represents the request structure for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRPCRequest extends GenericRPCMessage {
  method: RPCMethod.GetLedgerBalances;
  params: GetLedgerBalancesRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesRPCRequestParams {
  /** Optional account ID to filter ledger entries. */
  account_id?: string;
  /** Optional asset symbol to filter ledger entries. */
  asset?: string;
  /** Optional wallet address to filter ledger entries. If provided, overrides the authenticated wallet. */
  wallet?: Address;
}

/**
 * Represents the request structure for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesRPCRequest extends GenericRPCMessage {
  method: RPCMethod.GetLedgerEntries;
  params: GetLedgerEntriesRPCRequestParams;
}

/**
 * Represents the request parameters for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRPCRequestParams {
  /** The detailed definition of the application being created, including protocol, participants, weights, and quorum. */
  definition: AppDefinition;
  /** The initial allocation distribution among participants. Each participant must have sufficient balance for their allocation. */
  allocations: AppSessionAllocation[];
}

/**
 * Represents the request structure for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRPCRequest extends GenericRPCMessage {
  method: RPCMethod.CreateAppSession;
  params: CreateAppSessionRPCRequestParams;
}

/**
 * Represents the request parameters for the 'submit_state' RPC method.
 */
export interface SubmitStateRPCRequestParams {
  /** The unique identifier of the application session to update. */
  app_session_id: Hex;
  /** The new allocation distribution among participants. Must include all participants and maintain total balance. */
  allocations: AppSessionAllocation[];
}

/**
 * Represents the request structure for the 'submit_state' RPC method.
 */
export interface SubmitStateRPCRequest extends GenericRPCMessage {
  method: RPCMethod.SubmitState;
  params: SubmitStateRPCRequestParams;
}

/**
 * Represents the request parameters for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRPCRequestParams {
  /** The unique identifier of the application session to close. */
  app_session_id: Hex;
  /** The final allocation distribution among participants upon closing. Must include all participants and maintain total balance. */
  allocations: AppSessionAllocation[];
}

/**
 * Represents the request structure for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRPCRequest extends GenericRPCMessage {
  method: RPCMethod.CloseAppSession;
  params: CloseAppSessionRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRPCRequestParams {
  /** The unique identifier of the application session to retrieve. */
  app_session_id: Hex;
}

/**
 * Represents the request structure for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRPCRequest extends GenericRPCMessage {
  method: RPCMethod.GetAppDefinition;
  params: GetAppDefinitionRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRPCRequestParams {
  /** Optional participant address to filter application sessions. If not provided, returns all sessions. */
  participant?: Address;
  /** Optional status to filter application sessions (e.g., "open", "closed"). If not provided, returns sessions of all statuses. */
  status?: ChannelStatus;
}

/**
 * Represents the request structure for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRPCRequest extends GenericRPCMessage {
  method: RPCMethod.GetAppSessions;
  params: GetAppSessionsRPCRequestParams;
}

/**
 * Represents the request parameters for the 'resize_channel' RPC method.
 */
export interface ResizeChannelRPCRequestParams {
  /** The unique identifier of the channel to resize. */
  channel_id: Hex;
  /** Optional amount to resize the channel by (can be positive or negative). Must be provided if allocate_amount is not. */
  resize_amount?: bigint;
  /** Optional amount to allocate from the unified balance to the channel. Must be provided if resize_amount is not. */
  allocate_amount?: bigint;
  /** The address where the resized funds will be sent. */
  funds_destination: Address;
}

/**
 * Represents the request structure for the 'resize_channel' RPC method.
 */
export interface ResizeChannelRPCRequest extends GenericRPCMessage {
  method: RPCMethod.ResizeChannel;
  params: ResizeChannelRPCRequestParams;
}

/**
 * Represents the request parameters for the 'close_channel' RPC method.
 */
export interface CloseChannelRPCRequestParams {
  /** The unique identifier of the channel to close. */
  channel_id: Hex;
  /** The address where the channel funds will be sent upon closing. */
  funds_destination: Address;
}

/**
 * Represents the request structure for the 'close_channel' RPC method.
 */
export interface CloseChannelRPCRequest extends GenericRPCMessage {
  method: RPCMethod.CloseChannel;
  params: CloseChannelRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_channels' RPC method.
 */
export interface GetChannelsRPCRequestParams {
  /** Optional participant address to filter channels. If not provided, returns all channels. */
  participant?: Address;
  /** Optional status to filter channels (e.g., "open", "closed"). If not provided, returns channels of all statuses. */
  status?: ChannelStatus;
}

/**
 * Represents the request structure for the 'get_channels' RPC method.
 */
export interface GetChannelsRPCRequest extends GenericRPCMessage {
  method: RPCMethod.GetChannels;
  params: GetChannelsRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRPCRequestParams {
  /** The participant address to retrieve RPC history for. Must be the authenticated wallet address. */
  participant: Address;
}

/**
 * Represents the request structure for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRPCRequest extends GenericRPCMessage {
  method: RPCMethod.GetRPCHistory;
  params: GetRPCHistoryRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_assets' RPC method.
 */
export interface GetAssetsRPCRequestParams {
  /** Optional chain ID to filter assets by network. If not provided, returns assets from all networks. */
  chain_id?: number;
}

/**
 * Represents the request structure for the 'get_assets' RPC method.
 */
export interface GetAssetsRPCRequest extends GenericRPCMessage {
  method: RPCMethod.GetAssets;
  params: GetAssetsRPCRequestParams;
}

/**
 * Represents the request parameters for the 'auth_request' RPC method.
 */
export interface AuthRequestRPCRequestParams {
  /** The Ethereum address of the wallet being authorized. */
  address: Address;
  /** The session key address associated with the authentication attempt. */
  sessionKey: Address;
  /** The name of the application being authorized. */
  appName: string;
  /** The allowances for the connection. */
  allowances: {
    /** The asset symbol (e.g., "ETH", "USDC"). */
    asset: string;
    /** The amount allowed as a string. */
    amount: string;
  }[];
  /** The expiration timestamp for the authorization. */
  expire: string;
  /** The scope of the authorization. */
  scope: string;
  /** The application address being authorized. */
  applicationAddress: Address;
}

/**
 * Represents the request structure for the 'auth_request' RPC method.
 */
export interface AuthRequestRPCRequest extends GenericRPCMessage {
  method: RPCMethod.AuthRequest;
  params: AuthRequestRPCRequestParams;
}

/**
 * Represents the request parameters for the 'message' RPC method.
 */
export interface MessageRPCRequestParams {
  // Message parameters are handled by the virtual application
}

/**
 * Represents the request structure for the 'message' RPC method.
 */
export interface MessageRPCRequest extends GenericRPCMessage {
  method: RPCMethod.Message;
  params: MessageRPCRequestParams;
}

/**
 * Represents the request parameters for the 'ping' RPC method.
 */
export interface PingRPCRequestParams {
  // No parameters needed for ping
}

/**
 * Represents the request structure for the 'ping' RPC method.
 */
export interface PingRPCRequest extends GenericRPCMessage {
  method: RPCMethod.Ping;
  params: PingRPCRequestParams;
}

/**
 * Represents the request parameters for the 'pong' RPC method.
 */
export interface PongRPCRequestParams {
  // No parameters needed for pong
}

/**
 * Represents the request structure for the 'pong' RPC method.
 */
export interface PongRPCRequest extends GenericRPCMessage {
  method: RPCMethod.Pong;
  params: PongRPCRequestParams;
}

/**
 * Union type for all possible RPC request types.
 * This allows for type-safe handling of different request structures.
 */
export type RPCRequest =
  | AuthChallengeRPCRequest
  | AuthVerifyRPCRequest
  | AuthRequestRPCRequest
  | GetConfigRPCRequest
  | GetLedgerBalancesRPCRequest
  | GetLedgerEntriesRPCRequest
  | CreateAppSessionRPCRequest
  | SubmitStateRPCRequest
  | CloseAppSessionRPCRequest
  | GetAppDefinitionRPCRequest
  | GetAppSessionsRPCRequest
  | ResizeChannelRPCRequest
  | CloseChannelRPCRequest
  | GetChannelsRPCRequest
  | GetRPCHistoryRPCRequest
  | GetAssetsRPCRequest
  | PingRPCRequest
  | PongRPCRequest
  | MessageRPCRequest;

/**
 * Maps RPC methods to their corresponding request parameter types.
 */
export type RPCRequestParamsByMethod = {
  [RPCMethod.AuthChallenge]: AuthChallengeRPCRequestParams;
  [RPCMethod.AuthVerify]: AuthVerifyRPCRequestParams;
  [RPCMethod.AuthRequest]: AuthRequestRPCRequestParams;
  [RPCMethod.GetConfig]: GetConfigRPCRequestParams;
  [RPCMethod.GetLedgerBalances]: GetLedgerBalancesRPCRequestParams;
  [RPCMethod.GetLedgerEntries]: GetLedgerEntriesRPCRequestParams;
  [RPCMethod.CreateAppSession]: CreateAppSessionRPCRequestParams;
  [RPCMethod.SubmitState]: SubmitStateRPCRequestParams;
  [RPCMethod.CloseAppSession]: CloseAppSessionRPCRequestParams;
  [RPCMethod.GetAppDefinition]: GetAppDefinitionRPCRequestParams;
  [RPCMethod.GetAppSessions]: GetAppSessionsRPCRequestParams;
  [RPCMethod.ResizeChannel]: ResizeChannelRPCRequestParams;
  [RPCMethod.CloseChannel]: CloseChannelRPCRequestParams;
  [RPCMethod.GetChannels]: GetChannelsRPCRequestParams;
  [RPCMethod.GetRPCHistory]: GetRPCHistoryRPCRequestParams;
  [RPCMethod.GetAssets]: GetAssetsRPCRequestParams;
  [RPCMethod.Ping]: PingRPCRequestParams;
  [RPCMethod.Pong]: PongRPCRequestParams;
  [RPCMethod.Message]: MessageRPCRequestParams;
};
