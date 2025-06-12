import { Address, Hex } from 'viem';
import { RPCMethod, RequestID, Timestamp, AppDefinition, RPCChannelStatus } from '.';

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
 * Represents the parameters for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeRPCResponseParams {
  /** The challenge message to be signed by the client for authentication. */
  challengeMessage: string;
}

/**
 * Represents the response structure for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeRPCResponse extends GenericRPCMessage {
  method: RPCMethod.AuthChallenge;
  params: AuthChallengeRPCResponseParams;
}

/**
 * Represents the parameters for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRPCResponseParams {
  address: Address;
  jwtToken: string;
  sessionKey: Address;
  success: boolean;
}

/**
 * Represents the parameters for the 'error' RPC method.
 */
export interface ErrorRPCResponseParams {
  /** The error message describing what went wrong. */
  error: string;
}

/**
 * Represents the parameters for the 'get_config' RPC method.
 */
export interface GetConfigRPCResponseParams {
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
 * Represents the parameters for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRPCResponseParams {
  /** The asset symbol (e.g., "ETH", "USDC"). */
  asset: string;
  /** The balance amount as a string. */
  amount: string;
}

/**
 * Represents the parameters for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesRPCResponseParams {
  /** Unique identifier for the ledger entry. */
  id: number;
  /** The account identifier associated with the entry. */
  account_id: string;
  /** The type of account (e.g., "wallet", "channel"). */
  account_type: string;
  /** The asset symbol for the entry. */
  asset: string;
  /** The Ethereum address of the participant. */
  participant: Address;
  /** The credit amount as a string. */
  credit: string;
  /** The debit amount as a string. */
  debit: string;
  /** The timestamp when the entry was created. */
  created_at: string;
}

/**
 * Represents the parameters for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRPCResponseParams {
  /** The unique identifier for the application session. */
  app_session_id: Hex;
  /** The version number of the session. */
  version: number;
  /** The current status of the channel (e.g., "open", "closed"). */
  status: RPCChannelStatus;
}

/**
 * Represents the parameters for the 'submit_state' RPC method.
 */
export interface SubmitStateRPCResponseParams {
  /** The unique identifier for the application session. */
  app_session_id: Hex;
  /** The version number of the session. */
  version: number;
  /** The current status of the channel (e.g., "open", "closed"). */
  status: RPCChannelStatus;
}

/**
 * Represents the parameters for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRPCResponseParams {
  /** The unique identifier for the application session. */
  app_session_id: Hex;
  /** The version number of the session. */
  version: number;
  /** The current status of the channel (e.g., "open", "closed"). */
  status: RPCChannelStatus;
}

/**
 * Represents the parameters for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRPCResponseParams extends AppDefinition { }

/**
 * Represents the parameters for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRPCResponseParams {
  /** The unique identifier for the application session. */
  app_session_id: Hex;
  /** The current status of the channel (e.g., "open", "closed"). */
  status: RPCChannelStatus;
  /** List of participant Ethereum addresses. */
  participants: Address[];
  /** The protocol identifier for the application. */
  protocol: string;
  /** The challenge period in seconds. */
  challenge: number;
  /** The signature weights for each participant. */
  weights: number[];
  /** The minimum number of signatures required for state updates. */
  quorum: number;
  /** The version number of the session. */
  version: number;
  /** The nonce value for the session. */
  nonce: number;
  /** The timestamp when the session was created. */
  created_at: string;
  /** The timestamp when the session was last updated. */
  updated_at: string;
}

/**
 * Represents the parameters for the 'resize_channel' RPC method.
 */
export interface ResizeChannelRPCResponseParams {
  /** The unique identifier for the channel. */
  channel_id: Hex;
  /** The encoded state data for the channel. */
  state_data: string;
  /** The intent type for the state update. */
  intent: number;
  /** The version number of the channel. */
  version: number;
  /** The list of allocations for the channel. */
  allocations: {
    /** The destination address for the allocation. */
    destination: Address;
    /** The token contract address. */
    token: Address;
    /** The amount to allocate as a string. */
    amount: string;
  }[];
  /** The hash of the channel state. */
  state_hash: string;
  /** The server's signature for the state update. */
  server_signature: {
    /** The recovery value of the signature. */
    v: string;
    /** The r value of the signature. */
    r: string;
    /** The s value of the signature. */
    s: string;
  };
}

/**
 * Represents the parameters for the 'close_channel' RPC method.
 */
export interface CloseChannelRPCResponseParams {
  /** The unique identifier for the channel. */
  channel_id: Hex;
  /** The intent type for the state update. */
  intent: number;
  /** The version number of the channel. */
  version: number;
  /** The encoded state data for the channel. */
  state_data: string;
  /** The list of final allocations for the channel. */
  allocations: {
    /** The destination address for the allocation. */
    destination: Address;
    /** The token contract address. */
    token: Address;
    /** The amount to allocate as a string. */
    amount: string;
  }[];
  /** The hash of the channel state. */
  state_hash: string;
  /** The server's signature for the state update. */
  server_signature: {
    /** The recovery value of the signature. */
    v: string;
    /** The r value of the signature. */
    r: string;
    /** The s value of the signature. */
    s: string;
  };
}

/**
 * Represents the parameters for the 'get_channels' RPC method.
 */
export interface GetChannelsRPCResponseParams {
  /** The unique identifier for the channel. */
  channel_id: Hex;
  /** The Ethereum address of the participant. */
  participant: Address;
  /** The current status of the channel (e.g., "open", "closed"). */
  status: RPCChannelStatus;
  /** The token contract address. */
  token: Address;
  /** The wallet address associated with the channel. */
  wallet: Address;
  /** The total amount in the channel as a string. */
  amount: string;
  /** The chain ID where the channel exists. */
  chain_id: number;
  /** The adjudicator contract address. */
  adjudicator: Address;
  /** The challenge period in seconds. */
  challenge: number;
  /** The nonce value for the channel. */
  nonce: number;
  /** The version number of the channel. */
  version: number;
  /** The timestamp when the channel was created. */
  created_at: string;
  /** The timestamp when the channel was last updated. */
  updated_at: string;
}

/**
 * Represents the parameters for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRPCResponseParams {
  /** Unique identifier for the RPC entry. */
  id: number;
  /** The Ethereum address of the sender. */
  sender: Address;
  /** The request ID for the RPC call. */
  req_id: number;
  /** The RPC method name. */
  method: string;
  /** The JSON string of the request parameters. */
  params: string;
  /** The timestamp of the RPC call. */
  timestamp: number;
  /** Array of request signatures. */
  req_sig: Hex[];
  /** Array of response signatures. */
  res_sig: Hex[];
  /** The JSON string of the response. */
  response: string;
}

/**
 * Represents the parameters for the 'get_assets' RPC method.
 */
export interface GetAssetsRPCResponseParams {
  /** The token contract address. */
  token: Address;
  /** The chain ID where the asset exists. */
  chain_id: number;
  /** The asset symbol (e.g., "ETH", "USDC"). */
  symbol: string;
  /** The number of decimal places for the asset. */
  decimals: number;
}

/**
 * Represents the response structure for an error response.
 */
export interface ErrorRPCResponse extends GenericRPCMessage {
  method: RPCMethod.Error;
  params: ErrorRPCResponseParams;
}

/**
 * Represents the response structure for the 'get_config' RPC method.
 */
export interface GetConfigRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetConfig;
  params: GetConfigRPCResponseParams;
}

/**
 * Represents the response structure for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetLedgerBalances;
  params: GetLedgerBalancesRPCResponseParams[];
}

/**
 * Represents the response structure for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetLedgerEntries;
  params: GetLedgerEntriesRPCResponseParams[];
}

/**
 * Represents the response structure for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRPCResponse extends GenericRPCMessage {
  method: RPCMethod.CreateAppSession;
  params: CreateAppSessionRPCResponseParams;
}

/**
 * Represents the response structure for the 'submit_state' RPC method.
 */
export interface SubmitStateRPCResponse extends GenericRPCMessage {
  method: RPCMethod.SubmitState;
  params: SubmitStateRPCResponseParams;
}

/**
 * Represents the response structure for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRPCResponse extends GenericRPCMessage {
  method: RPCMethod.CloseAppSession;
  params: CloseAppSessionRPCResponseParams;
}

/**
 * Represents the response structure for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetAppDefinition;
  params: GetAppDefinitionRPCResponseParams;
}

/**
 * Represents the response structure for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetAppSessions;
  params: GetAppSessionsRPCResponseParams[];
}

/**
 * Represents the response structure for the 'resize_channel' RPC method.
 */
export interface ResizeChannelRPCResponse extends GenericRPCMessage {
  method: RPCMethod.ResizeChannel;
  params: ResizeChannelRPCResponseParams;
}

/**
 * Represents the response structure for the 'close_channel' RPC method.
 */
export interface CloseChannelRPCResponse extends GenericRPCMessage {
  method: RPCMethod.CloseChannel;
  params: CloseChannelRPCResponseParams;
}

/**
 * Represents the response structure for the 'get_channels' RPC method.
 */
export interface GetChannelsRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetChannels;
  params: GetChannelsRPCResponseParams[];
}

/**
 * Represents the response structure for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetRPCHistory;
  params: GetRPCHistoryRPCResponseParams[];
}

/**
 * Represents the response structure for the 'get_assets' RPC method.
 */
export interface GetAssetsRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetAssets;
  params: GetAssetsRPCResponseParams[];
}

/**
 * Represents the response structure for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRPCResponse extends GenericRPCMessage {
  method: RPCMethod.AuthVerify;
  params: AuthVerifyRPCResponseParams;
}

/**
 * Represents the parameters for the 'auth_request' RPC method.
 */
export interface AuthRequestRPCResponseParams {
  /** The challenge message to be signed by the client for authentication. */
  challengeMessage: string;
}

/**
 * Represents the response structure for the 'auth_request' RPC method.
 */
export interface AuthRequestRPCResponse extends GenericRPCMessage {
  method: RPCMethod.AuthRequest;
  params: AuthRequestRPCResponseParams;
}

/**
 * Represents the response parameters for the 'message' RPC method.
 */
export interface MessageRPCResponseParams {
  // Message response parameters are handled by the application
}

/**
 * Represents the response structure for the 'message' RPC method.
 */
export interface MessageRPCResponse extends GenericRPCMessage {
  method: RPCMethod.Message;
  params: MessageRPCResponseParams;
}

/**
 * Represents the parameters for the 'balance_update' RPC method.
 */
export interface BalanceUpdateRPCResponseParams {
  /** The asset symbol (e.g., "ETH", "USDC"). */
  asset: string;
  /** The balance amount as a string. */
  amount: string;
}

/**
 * Represents the response structure for the 'balance_update' RPC method.
 */
export interface BalanceUpdateRPCResponse extends GenericRPCMessage {
  method: RPCMethod.BalanceUpdate;
  params: BalanceUpdateRPCResponseParams[];
}

/**
 * Represents the response structure for the 'channels_update' RPC method.
 */
export interface ChannelsUpdateRPCResponse extends GenericRPCMessage {
  method: RPCMethod.ChannelsUpdate;
  params: ChannelUpdateRPCResponseParams[];
}

/**
 * Represents the parameters for the 'channel_update' RPC method.
 */
export interface ChannelUpdateRPCResponseParams {
  /** The unique identifier for the channel. */
  channel_id: Hex;
  /** The Ethereum address of the participant. */
  participant: Address;
  /** The current status of the channel (e.g., "open", "closed"). */
  status: RPCChannelStatus;
  /** The token contract address. */
  token: Address;
  /** The total amount in the channel as a string. */
  amount: string;
  /** The chain ID where the channel exists. */
  chain_id: number;
  /** The adjudicator contract address. */
  adjudicator: Address;
  /** The challenge period in seconds. */
  challenge: number;
  /** The nonce value for the channel. */
  nonce: number;
  /** The version number of the channel. */
  version: number;
  /** The timestamp when the channel was created. */
  created_at: string;
  /** The timestamp when the channel was last updated. */
  updated_at: string;
}

/**
 * Represents the response structure for the 'channel_update' RPC method.
 */
export interface ChannelUpdateRPCResponse extends GenericRPCMessage {
  method: RPCMethod.ChannelUpdate;
  params: ChannelUpdateRPCResponseParams;
}

/**
 * Represents the parameters for the 'ping' RPC method.
 */
export interface PingRPCResponseParams {
  // No parameters needed for ping
}

/**
 * Represents the response structure for the 'ping' RPC method.
 */
export interface PingRPCResponse extends GenericRPCMessage {
  method: RPCMethod.Ping;
  params: PingRPCResponseParams;
}

/**
 * Represents the parameters for the 'pong' RPC method.
 */
export interface PongRPCResponseParams {
  // No parameters needed for pong
}

/**
 * Represents the response structure for the 'pong' RPC method.
 */
export interface PongRPCResponse extends GenericRPCMessage {
  method: RPCMethod.Pong;
  params: PongRPCResponseParams;
}

/**
 * Union type for all possible RPC response types.
 * This allows for type-safe handling of different response structures.
 */
export type RPCResponse =
  | AuthChallengeRPCResponse
  | AuthVerifyRPCResponse
  | AuthRequestRPCResponse
  | ErrorRPCResponse
  | GetConfigRPCResponse
  | GetLedgerBalancesRPCResponse
  | GetLedgerEntriesRPCResponse
  | CreateAppSessionRPCResponse
  | SubmitStateRPCResponse
  | CloseAppSessionRPCResponse
  | GetAppDefinitionRPCResponse
  | GetAppSessionsRPCResponse
  | ResizeChannelRPCResponse
  | CloseChannelRPCResponse
  | GetChannelsRPCResponse
  | GetRPCHistoryRPCResponse
  | GetAssetsRPCResponse
  | PingRPCResponse
  | PongRPCResponse
  | MessageRPCResponse
  | BalanceUpdateRPCResponse
  | ChannelsUpdateRPCResponse
  | ChannelUpdateRPCResponse;

/**
 * Maps RPC methods to their corresponding parameter types.
 */
export type RPCResponseParamsByMethod = {
  [RPCMethod.AuthChallenge]: AuthChallengeRPCResponseParams;
  [RPCMethod.AuthVerify]: AuthVerifyRPCResponseParams;
  [RPCMethod.AuthRequest]: AuthRequestRPCResponseParams;
  [RPCMethod.Error]: ErrorRPCResponseParams;
  [RPCMethod.GetConfig]: GetConfigRPCResponseParams;
  [RPCMethod.GetLedgerBalances]: GetLedgerBalancesRPCResponseParams[];
  [RPCMethod.GetLedgerEntries]: GetLedgerEntriesRPCResponseParams[];
  [RPCMethod.CreateAppSession]: CreateAppSessionRPCResponseParams;
  [RPCMethod.SubmitState]: SubmitStateRPCResponseParams;
  [RPCMethod.CloseAppSession]: CloseAppSessionRPCResponseParams;
  [RPCMethod.GetAppDefinition]: GetAppDefinitionRPCResponseParams;
  [RPCMethod.GetAppSessions]: GetAppSessionsRPCResponseParams[];
  [RPCMethod.ResizeChannel]: ResizeChannelRPCResponseParams;
  [RPCMethod.CloseChannel]: CloseChannelRPCResponseParams;
  [RPCMethod.GetChannels]: GetChannelsRPCResponseParams[];
  [RPCMethod.GetRPCHistory]: GetRPCHistoryRPCResponseParams[];
  [RPCMethod.GetAssets]: GetAssetsRPCResponseParams[];
  [RPCMethod.Ping]: PingRPCResponseParams;
  [RPCMethod.Pong]: PongRPCResponseParams;
  [RPCMethod.Message]: MessageRPCResponseParams;
  [RPCMethod.BalanceUpdate]: BalanceUpdateRPCResponseParams[];
  [RPCMethod.ChannelsUpdate]: ChannelUpdateRPCResponseParams[];
  [RPCMethod.ChannelUpdate]: ChannelUpdateRPCResponseParams;
};
