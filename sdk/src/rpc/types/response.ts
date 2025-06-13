import { Address, Hex } from 'viem';
import { RPCMethod, GenericRPCMessage, AppDefinition, RPCChannelStatus, AuthVerifyRequestParams, TransferAllocation } from '.';

/**
 * Represents the parameters for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeResponseParams {
    /** The challenge message to be signed by the client for authentication. */
    challenge_message: string;
}

/**
 * Represents the response structure for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeResponse extends GenericRPCMessage {
    method: RPCMethod.AuthChallenge;
    params: [AuthChallengeResponseParams];
}

/**
 * Represents the parameters for the 'auth_verify' RPC method.
 */
export interface AuthVerifyResponseParams {
    address: Address;
    /** Available only if challenge auth method was used in {@link AuthVerifyRequestParams} during the call to {@link RPCMethod.AuthRequest} */
  jwt_token?: string;
    session_key: Address;
    success: boolean;
}

/**
 * Represents the parameters for the 'error' RPC method.
 */
export interface ErrorResponseParams {
    /** The error message describing what went wrong. */
    error: string;
}

/**
 * Represents the network information for the 'get_config' RPC method.
 */
export interface NetworkInfo {
  /** The name of the network (e.g., "Ethereum", "Polygon"). */
  name: string;
  /** The chain ID of the network. */
  chain_id: number;
  /** The custody contract address for the network. */
  custody_address: Address;
  /** The adjudicator contract address for the network. */
  adjudicator_address: Address;
}

/**
 * Represents the parameters for the 'get_config' RPC method.
 */
export interface GetConfigResponseParams {
  /** The Ethereum address of the broker. */
  broker_address: Address;
  /** List of supported networks and their configurations. */
  networks: NetworkInfo[];
}

/**
 * Represents the parameters for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesResponseParams {
    /** The asset symbol (e.g., "ETH", "USDC"). */
    asset: string;
    /** The balance amount as a string. */
    amount: string;
}

/**
 * Represents the parameters for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesResponseParams {
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
export interface CreateAppSessionResponseParams {
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
export interface SubmitStateResponseParams {
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
export interface CloseAppSessionResponseParams {
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
export interface GetAppDefinitionResponseParams extends AppDefinition {
  /** The protocol identifier for the application (e.g., "payment", "swap"). */
  protocol: string;
  /** List of Ethereum addresses of participants in the application session. */
  participants: Address[];
  /** Array of signature weights for each participant, used for quorum calculations. */
  weights: number[];
  /** The minimum number of signatures required for state updates. */
  quorum: number;
  /** The challenge period in seconds for state updates. */
  challenge: number;
  /** A unique nonce value for the application session to prevent replay attacks. */
  nonce: number;
}

/**
 * Represents the parameters for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsResponseParams {
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

export interface ServerSignature {
  /** The recovery value of the signature. */
  v: string;
  r: string;
  s: string;
}

export interface RPCAllocation {
  /** The destination address for the allocation. */
  destination: Address;
  /** The token contract address. */
  token: Address;
  /** The amount to allocate as a string. */
  amount: string;
}

/**
 * Represents the parameters for the 'resize_channel' RPC method.
 */
export interface ResizeChannelResponseParams {
    /** The unique identifier for the channel. */
    channel_id: Hex;
    /** The encoded state data for the channel. */
    state_data: string;
    /** The intent type for the state update. */
    intent: number;
    /** The version number of the channel. */
    version: number;
    /** The list of allocations for the channel. */
    allocations: RPCAllocation[];
    /** The hash of the channel state. */
    state_hash: string;
    /** The server's signature for the state update. */
    server_signature: ServerSignature;
}

/**
 * Represents the parameters for the 'close_channel' RPC method.
 */
export interface CloseChannelResponseParams {
    /** The unique identifier for the channel. */
    channel_id: Hex;
    /** The intent type for the state update. */
    intent: number;
    /** The version number of the channel. */
    version: number;
    /** The encoded state data for the channel. */
    state_data: string;
    /** The list of final allocations for the channel. */
    allocations: RPCAllocation[];
    /** The hash of the channel state. */
    state_hash: string;
    /** The server's signature for the state update. */
    server_signature: ServerSignature;
}

/**
 * Represents the parameters for the 'get_channels' RPC method.
 */
export interface GetChannelsResponseParams {
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
export interface GetRPCHistoryResponseParams {
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
export interface GetAssetsResponseParams {
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
export interface ErrorResponse extends GenericRPCMessage {
    method: RPCMethod.Error;
    params: ErrorResponseParams[];
}

/**
 * Represents the response structure for the 'get_config' RPC method.
 */
export interface GetConfigResponse extends GenericRPCMessage {
    method: RPCMethod.GetConfig;
    params: GetConfigResponseParams[];
}

/**
 * Represents the response structure for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesResponse extends GenericRPCMessage {
    method: RPCMethod.GetLedgerBalances;
    params: [GetLedgerBalancesResponseParams[]];
}

/**
 * Represents the response structure for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesResponse extends GenericRPCMessage {
    method: RPCMethod.GetLedgerEntries;
    params: [GetLedgerEntriesResponseParams[]];
}

/**
 * Represents the response structure for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionResponse extends GenericRPCMessage {
    method: RPCMethod.CreateAppSession;
    params: [CreateAppSessionResponseParams];
}

/**
 * Represents the response structure for the 'submit_state' RPC method.
 */
export interface SubmitStateResponse extends GenericRPCMessage {
    method: RPCMethod.SubmitState;
    params: [SubmitStateResponseParams];
}

/**
 * Represents the response structure for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionResponse extends GenericRPCMessage {
    method: RPCMethod.CloseAppSession;
    params: [CloseAppSessionResponseParams];
}

/**
 * Represents the response structure for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionResponse extends GenericRPCMessage {
    method: RPCMethod.GetAppDefinition;
    params: [GetAppDefinitionResponseParams];
}

/**
 * Represents the response structure for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsResponse extends GenericRPCMessage {
    method: RPCMethod.GetAppSessions;
    params: [GetAppSessionsResponseParams[]];
}

/**
 * Represents the response structure for the 'resize_channel' RPC method.
 */
export interface ResizeChannelResponse extends GenericRPCMessage {
    method: RPCMethod.ResizeChannel;
    params: [ResizeChannelResponseParams];
}

/**
 * Represents the response structure for the 'close_channel' RPC method.
 */
export interface CloseChannelResponse extends GenericRPCMessage {
    method: RPCMethod.CloseChannel;
    params: [CloseChannelResponseParams];
}

/**
 * Represents the response structure for the 'get_channels' RPC method.
 */
export interface GetChannelsResponse extends GenericRPCMessage {
    method: RPCMethod.GetChannels;
    params: [GetChannelsResponseParams[]];
}

/**
 * Represents the response structure for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryResponse extends GenericRPCMessage {
    method: RPCMethod.GetRPCHistory;
    params: [GetRPCHistoryResponseParams[]];
}

/**
 * Represents the response structure for the 'get_assets' RPC method.
 */
export interface GetAssetsResponse extends GenericRPCMessage {
    method: RPCMethod.GetAssets;
    params: [GetAssetsResponseParams[]];
}

/**
 * Represents the response structure for the 'auth_verify' RPC method.
 */
export interface AuthVerifyResponse extends GenericRPCMessage {
    method: RPCMethod.AuthVerify;
    params: [AuthVerifyResponseParams];
}

/**
 * Represents the parameters for the 'auth_request' RPC method.
 */
export interface AuthRequestResponseParams {
    /** The challenge message to be signed by the client for authentication. */
    challenge_message: string;
}

/**
 * Represents the response structure for the 'auth_request' RPC method.
 */
export interface AuthRequestResponse extends GenericRPCMessage {
    method: RPCMethod.AuthRequest;
    params: [AuthRequestResponseParams];
}

/**
 * Represents the response parameters for the 'message' RPC method.
 */
export interface MessageResponseParams {
    // Message response parameters are handled by the application
}

/**
 * Represents the response structure for the 'message' RPC method.
 */
export interface MessageResponse extends GenericRPCMessage {
    method: RPCMethod.Message;
    params: [MessageResponseParams];
}

/**
 * Represents the parameters for the 'balance_update' RPC method.
 */
export interface BalanceUpdateResponseParams {
    /** The asset symbol (e.g., "ETH", "USDC"). */
    asset: string;
    /** The balance amount as a string. */
    amount: string;
}

/**
 * Represents the response structure for the 'balance_update' RPC method.
 */
export interface BalanceUpdateResponse extends GenericRPCMessage {
    method: RPCMethod.BalanceUpdate;
    params: [BalanceUpdateResponseParams[]];
}

/**
 * Represents the response structure for the 'channels_update' RPC method.
 */
export interface ChannelsUpdateResponse extends GenericRPCMessage {
    method: RPCMethod.ChannelsUpdate;
    params: [ChannelUpdateResponseParams[]];
}

/**
 * Represents the parameters for the 'channel_update' RPC method.
 */
export interface ChannelUpdateResponseParams {
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
export interface ChannelUpdateResponse extends GenericRPCMessage {
    method: RPCMethod.ChannelUpdate;
    params: [ChannelUpdateResponseParams];
}

/**
 * Represents the parameters for the 'ping' RPC method.
 */
export interface PingResponseParams {
    // No parameters needed for ping
}

/**
 * Represents the response structure for the 'ping' RPC method.
 */
export interface PingResponse extends GenericRPCMessage {
    method: RPCMethod.Ping;
    params: [PingResponseParams];
}

/**
 * Represents the parameters for the 'pong' RPC method.
 */
export interface PongResponseParams {
    // No parameters needed for pong
}

/**
 * Represents the response structure for the 'pong' RPC method.
 */
export interface PongResponse extends GenericRPCMessage {
    method: RPCMethod.Pong;
    params: [PongResponseParams];
}

/**
 * Represents the parameters for the 'transfer' RPC method.
 */
export interface TransferRPCResponseParams {
    /** The source address from which assets were transferred. */
    from: Address;
    /** The destination address to which assets were transferred. */
    to: Address;
    /** The assets and amounts that were transferred. */
    allocations: TransferAllocation[];
    /** The timestamp when the transfer was created. */
    created_at: string;
}

/**
 * Represents the response structure for the 'transfer' RPC method.
 */
export interface TransferRPCResponse extends GenericRPCMessage {
    method: RPCMethod.Transfer;
    params: TransferRPCResponseParams;
}

/**
 * Union type for all possible RPC response types.
 * This allows for type-safe handling of different response structures.
 */
export type RPCResponse =
  | AuthChallengeResponse
  | AuthVerifyResponse
  | AuthRequestResponse
  | ErrorResponse
  | GetConfigResponse
  | GetLedgerBalancesResponse
  | GetLedgerEntriesResponse
  | CreateAppSessionResponse
  | SubmitStateResponse
  | CloseAppSessionResponse
  | GetAppDefinitionResponse
  | GetAppSessionsResponse
  | ResizeChannelResponse
  | CloseChannelResponse
  | GetChannelsResponse
  | GetRPCHistoryResponse
  | GetAssetsResponse
  | PingResponse
  | PongResponse
  | MessageResponse
  | BalanceUpdateResponse
  | ChannelsUpdateResponse
  | ChannelUpdateResponse
  | TransferRPCResponse;

/**
 * Maps RPC methods to their corresponding parameter types.
 */
export type RPCResponseParamsByMethod = {
    [RPCMethod.AuthChallenge]: AuthChallengeResponseParams;
    [RPCMethod.AuthVerify]: AuthVerifyResponseParams;
    [RPCMethod.AuthRequest]: AuthRequestResponseParams;
    [RPCMethod.Error]: ErrorResponseParams;
    [RPCMethod.GetConfig]: GetConfigResponseParams;
    [RPCMethod.GetLedgerBalances]: GetLedgerBalancesResponseParams[];
    [RPCMethod.GetLedgerEntries]: GetLedgerEntriesResponseParams[];
    [RPCMethod.CreateAppSession]: CreateAppSessionResponseParams;
    [RPCMethod.SubmitState]: SubmitStateResponseParams;
    [RPCMethod.CloseAppSession]: CloseAppSessionResponseParams;
    [RPCMethod.GetAppDefinition]: GetAppDefinitionResponseParams;
    [RPCMethod.GetAppSessions]: GetAppSessionsResponseParams[];
    [RPCMethod.ResizeChannel]: ResizeChannelResponseParams;
    [RPCMethod.CloseChannel]: CloseChannelResponseParams;
    [RPCMethod.GetChannels]: GetChannelsResponseParams[];
    [RPCMethod.GetRPCHistory]: GetRPCHistoryResponseParams[];
    [RPCMethod.GetAssets]: GetAssetsResponseParams[];
    [RPCMethod.Ping]: PingResponseParams;
    [RPCMethod.Pong]: PongResponseParams;
    [RPCMethod.Message]: MessageResponseParams;
    [RPCMethod.BalanceUpdate]: BalanceUpdateResponseParams[];
    [RPCMethod.ChannelsUpdate]: ChannelUpdateResponseParams[];
    [RPCMethod.ChannelUpdate]: ChannelUpdateResponseParams;
    [RPCMethod.Transfer]: TransferRPCResponseParams;
};
