import { Address, Hex } from 'viem';
import { RPCMethod, RequestID, Timestamp, AppDefinition, AppSessionAllocation, RPCChannelStatus as RPCChannelStatus, Allowance, TransferAllocation } from '.';

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
export interface AuthChallengeRequestParams {
    /** The challenge message to be signed by the client for authentication. */
    challengeMessage: string;
}

/**
 * Represents the request structure for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeRequest extends GenericRPCMessage {
    method: RPCMethod.AuthChallenge;
    params: AuthChallengeRequestParams;
}

/**
 * Represents the request parameters for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRequestParams {
    /** The challenge token received from auth_challenge response. Used to verify the client's signature and prevent replay attacks. */
    challenge: string,
    /** JSON Web Token for authentication, if provided. */
    jwt: string;
}

/**
 * Represents the request structure for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRequest extends GenericRPCMessage {
    method: RPCMethod.AuthVerify;
    params: AuthVerifyRequestParams;
}

/**
 * Represents the request parameters for the 'get_config' RPC method.
 */
export interface GetConfigRequestParams {
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
export interface GetConfigRequest extends GenericRPCMessage {
    method: RPCMethod.GetConfig;
    params: GetConfigRequestParams;
}

/**
 * Represents the request parameters for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRequestParams {
    /** Optional participant address to filter balances. If not provided, uses the authenticated wallet address. */
    participant?: Address;
    /** Optional account ID to filter balances. If provided, overrides the participant address. */
    account_id?: string;
}

/**
 * Represents the request structure for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRequest extends GenericRPCMessage {
    method: RPCMethod.GetLedgerBalances;
    params: GetLedgerBalancesRequestParams;
}

/**
 * Represents the request parameters for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesRequestParams {
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
export interface GetLedgerEntriesRequest extends GenericRPCMessage {
    method: RPCMethod.GetLedgerEntries;
    params: GetLedgerEntriesRequestParams;
}

/**
 * Represents the request parameters for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRequestParams {
    /** The detailed definition of the application being created, including protocol, participants, weights, and quorum. */
    definition: AppDefinition;
    /** The initial allocation distribution among participants. Each participant must have sufficient balance for their allocation. */
    allocations: AppSessionAllocation[];
}

/**
 * Represents the request structure for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRequest extends GenericRPCMessage {
    method: RPCMethod.CreateAppSession;
    params: CreateAppSessionRequestParams;
}

/**
 * Represents the request parameters for the 'submit_state' RPC method.
 */
export interface SubmitStateRequestParams {
    /** The unique identifier of the application session to update. */
    app_session_id: Hex;
    /** The new allocation distribution among participants. Must include all participants and maintain total balance. */
    allocations: AppSessionAllocation[];
}

/**
 * Represents the request structure for the 'submit_state' RPC method.
 */
export interface SubmitStateRequest extends GenericRPCMessage {
    method: RPCMethod.SubmitState;
    params: SubmitStateRequestParams;
}

/**
 * Represents the request parameters for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRequestParams {
    /** The unique identifier of the application session to close. */
    app_session_id: Hex;
    /** The final allocation distribution among participants upon closing. Must include all participants and maintain total balance. */
    allocations: AppSessionAllocation[];
}

/**
 * Represents the request structure for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRequest extends GenericRPCMessage {
    method: RPCMethod.CloseAppSession;
    params: CloseAppSessionRequestParams;
}

/**
 * Represents the request parameters for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRequestParams {
    /** The unique identifier of the application session to retrieve. */
    app_session_id: Hex;
}

/**
 * Represents the request structure for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRequest extends GenericRPCMessage {
    method: RPCMethod.GetAppDefinition;
    params: GetAppDefinitionRequestParams;
}

/**
 * Represents the request parameters for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRequestParams {
    /** Optional participant address to filter application sessions. If not provided, returns all sessions. */
    participant?: Address;
    /** Optional status to filter application sessions (e.g., "open", "closed"). If not provided, returns sessions of all statuses. */
    status?: RPCChannelStatus;
}

/**
 * Represents the request structure for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRequest extends GenericRPCMessage {
    method: RPCMethod.GetAppSessions;
    params: GetAppSessionsRequestParams;
}

/**
 * Represents the request parameters for the 'resize_channel' RPC method.
 */
export interface ResizeChannelRequestParams {
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
export interface ResizeChannelRequest extends GenericRPCMessage {
    method: RPCMethod.ResizeChannel;
    params: ResizeChannelRequestParams;
}

/**
 * Represents the request parameters for the 'close_channel' RPC method.
 */
export interface CloseChannelRequestParams {
    /** The unique identifier of the channel to close. */
    channel_id: Hex;
    /** The address where the channel funds will be sent upon closing. */
    funds_destination: Address;
}

/**
 * Represents the request structure for the 'close_channel' RPC method.
 */
export interface CloseChannelRequest extends GenericRPCMessage {
    method: RPCMethod.CloseChannel;
    params: CloseChannelRequestParams;
}

/**
 * Represents the request parameters for the 'get_channels' RPC method.
 */
export interface GetChannelsRequestParams {
    /** Optional participant address to filter channels. If not provided, returns all channels. */
    participant?: Address;
    /** Optional status to filter channels (e.g., "open", "closed"). If not provided, returns channels of all statuses. */
    status?: RPCChannelStatus;
}

/**
 * Represents the request structure for the 'get_channels' RPC method.
 */
export interface GetChannelsRequest extends GenericRPCMessage {
    method: RPCMethod.GetChannels;
    params: GetChannelsRequestParams;
}

/**
 * Represents the request parameters for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRequestParams {
    /** The participant address to retrieve RPC history for. Must be the authenticated wallet address. */
    participant: Address;
}

/**
 * Represents the request structure for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRequest extends GenericRPCMessage {
    method: RPCMethod.GetRPCHistory;
    params: GetRPCHistoryRequestParams;
}

/**
 * Represents the request parameters for the 'get_assets' RPC method.
 */
export interface GetAssetsRequestParams {
    /** Optional chain ID to filter assets by network. If not provided, returns assets from all networks. */
    chain_id?: number;
}

/**
 * Represents the request structure for the 'get_assets' RPC method.
 */
export interface GetAssetsRequest extends GenericRPCMessage {
    method: RPCMethod.GetAssets;
    params: GetAssetsRequestParams;
}

/**
 * Represents the request parameters for the 'auth_request' RPC method.
 */
export interface AuthRequestParams {
    /** The Ethereum address of the wallet being authorized. */
    wallet: Address;
    /** The session key address associated with the authentication attempt. */
    participant: Address;
    /** The name of the application being authorized. */
    app_name: string;
    /** The allowances for the connection. */
    allowances: Allowance[];
    /** The expiration timestamp for the authorization. */
    expire: string;
    /** The scope of the authorization. */
    scope: string;
    /** The application address being authorized. */
    application: Address;
}

/**
 * Represents the request structure for the 'auth_request' RPC method.
 */
export interface AuthRequest extends GenericRPCMessage {
    method: RPCMethod.AuthRequest;
    params: AuthRequestParams;
}

/**
 * Represents the request parameters for the 'message' RPC method.
 */
export interface MessageRequestParams {
    // Message parameters are handled by the virtual application
}

/**
 * Represents the request structure for the 'message' RPC method.
 */
export interface MessageRequest extends GenericRPCMessage {
    method: RPCMethod.Message;
    params: MessageRequestParams;
}

/**
 * Represents the request parameters for the 'ping' RPC method.
 */
export interface PingRequestParams {
    // No parameters needed for ping
}

/**
 * Represents the request structure for the 'ping' RPC method.
 */
export interface PingRequest extends GenericRPCMessage {
    method: RPCMethod.Ping;
    params: PingRequestParams;
}

/**
 * Represents the request parameters for the 'pong' RPC method.
 */
export interface PongRequestParams {
    // No parameters needed for pong
}

/**
 * Represents the request structure for the 'pong' RPC method.
 */
export interface PongRequest extends GenericRPCMessage {
    method: RPCMethod.Pong;
    params: PongRequestParams;
}

/**
 * Represents the request parameters for the 'transfer' RPC method.
 */
export interface TransferRPCRequestParams {
    /** The destination address to transfer assets to. */
    destination: Address;
    /** The assets and amounts to transfer. */
    allocations: TransferAllocation[];
}

/**
 * Represents the request structure for the 'transfer' RPC method.
 */
export interface TransferRPCRequest extends GenericRPCMessage {
    method: RPCMethod.Transfer;
    params: TransferRPCRequestParams;
}

/**
 * Union type for all possible RPC request types.
 * This allows for type-safe handling of different request structures.
 */
export type RPCRequest =
  | AuthChallengeRequest
  | AuthVerifyRequest
  | AuthRequest
  | GetConfigRequest
  | GetLedgerBalancesRequest
  | GetLedgerEntriesRequest
  | CreateAppSessionRequest
  | SubmitStateRequest
  | CloseAppSessionRequest
  | GetAppDefinitionRequest
  | GetAppSessionsRequest
  | ResizeChannelRequest
  | CloseChannelRequest
  | GetChannelsRequest
  | GetRPCHistoryRequest
  | GetAssetsRequest
  | PingRequest
  | PongRequest
  | MessageRequest
  | TransferRPCRequest;

/**
 * Maps RPC methods to their corresponding request parameter types.
 */
export type RPCRequestParamsByMethod = {
    [RPCMethod.AuthChallenge]: AuthChallengeRequestParams;
    [RPCMethod.AuthVerify]: AuthVerifyRequestParams;
    [RPCMethod.AuthRequest]: AuthRequestParams;
    [RPCMethod.GetConfig]: GetConfigRequestParams;
    [RPCMethod.GetLedgerBalances]: GetLedgerBalancesRequestParams;
    [RPCMethod.GetLedgerEntries]: GetLedgerEntriesRequestParams;
    [RPCMethod.CreateAppSession]: CreateAppSessionRequestParams;
    [RPCMethod.SubmitState]: SubmitStateRequestParams;
    [RPCMethod.CloseAppSession]: CloseAppSessionRequestParams;
    [RPCMethod.GetAppDefinition]: GetAppDefinitionRequestParams;
    [RPCMethod.GetAppSessions]: GetAppSessionsRequestParams;
    [RPCMethod.ResizeChannel]: ResizeChannelRequestParams;
    [RPCMethod.CloseChannel]: CloseChannelRequestParams;
    [RPCMethod.GetChannels]: GetChannelsRequestParams;
    [RPCMethod.GetRPCHistory]: GetRPCHistoryRequestParams;
    [RPCMethod.GetAssets]: GetAssetsRequestParams;
    [RPCMethod.Ping]: PingRequestParams;
    [RPCMethod.Pong]: PongRequestParams;
    [RPCMethod.Message]: MessageRequestParams;
    [RPCMethod.Transfer]: TransferRPCRequestParams;
};
