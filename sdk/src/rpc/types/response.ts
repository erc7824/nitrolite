import { Address, Hex } from 'viem';
import {
    RPCMethod,
    GenericRPCMessage,
    AppDefinition,
    RPCChannelStatus,
    AuthVerifyRequestParams,
    TransferAllocation,
    ChannelUpdate,
} from '.';

/**
 * Represents the parameters for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeResponseParams {
    /** The challenge message to be signed by the client for authentication. */
    challengeMessage: string;
}
export type AuthChallengeRPCResponseParams = AuthChallengeResponseParams; // for backward compatibility

/**
 * Represents the response structure for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeResponse extends GenericRPCMessage {
    method: RPCMethod.AuthChallenge;
    params: AuthChallengeResponseParams;
}

/**
 * Represents the parameters for the 'auth_verify' RPC method.
 */
export type AuthVerifyResponseParams =
    | {
          address: Address;
          sessionKey: Address;
          success: boolean;
      }
    & {
          /** Available only if challenge auth method was used in {@link AuthVerifyRequestParams} during the call to {@link RPCMethod.AuthRequest} */
          jwtToken: string;
      };
export type AuthVerifyRPCResponseParams = AuthVerifyResponseParams; // for backward compatibility

/**
 * Represents the parameters for the 'error' RPC method.
 */
export interface ErrorResponseParams {
    /** The error message describing what went wrong. */
    error: string;
}
export type ErrorRPCResponseParams = ErrorResponseParams; // for backward compatibility

/**
 * Represents the network information for the 'get_config' RPC method.
 */
export interface NetworkInfo {
    /** The name of the network (e.g., "Ethereum", "Polygon"). */
    name: string;
    /** The chain ID of the network. */
    chainId: number;
    /** The custody contract address for the network. */
    custodyAddress: Address;
    /** The adjudicator contract address for the network. */
    adjudicatorAddress: Address;
}

/**
 * Represents the parameters for the 'get_config' RPC method.
 */
export interface GetConfigResponseParams {
    /** The Ethereum address of the broker. */
    brokerAddress: Address;
    /** List of supported networks and their configurations. */
    networks: NetworkInfo[];
}
export type GetConfigRPCResponseParams = GetConfigResponseParams; // for backward compatibility

/**
 * Represents the parameters for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesResponseParams {
    /** The asset symbol (e.g., "ETH", "USDC"). */
    asset: string;
    /** The balance amount. */
    amount: string;
}
export type GetLedgerBalancesRPCResponseParams = GetLedgerBalancesResponseParams; // for backward compatibility

/**
 * Represents the parameters for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesResponseParams {
    /** Unique identifier for the ledger entry. */
    id: number;
    /** The account identifier associated with the entry. */
    accountId: string;
    /** The type of account (e.g., "wallet", "channel"). */
    accountType: string;
    /** The asset symbol for the entry. */
    asset: string;
    /** The Ethereum address of the participant. */
    participant: Address;
    /** The credit amount. */
    credit: string;
    /** The debit amount. */
    debit: string;
    /** The timestamp when the entry was created. */
    createdAt: Date;
}
export type GetLedgerEntriesRPCResponseParams = GetLedgerEntriesResponseParams; // for backward compatibility

/**
 * Represents the parameters for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionResponseParams {
    /** The unique identifier for the application session. */
    appSessionId: Hex;
    /** The version number of the session. */
    version: number;
    /** The current status of the channel (e.g., "open", "closed"). */
    status: RPCChannelStatus;
}
export type CreateAppSessionRPCResponseParams = CreateAppSessionResponseParams; // for backward compatibility

/**
 * Represents the parameters for the 'submit_state' RPC method.
 */
export interface SubmitStateResponseParams {
    /** The unique identifier for the application session. */
    appSessionId: Hex;
    /** The version number of the session. */
    version: number;
    /** The current status of the channel (e.g., "open", "closed"). */
    status: RPCChannelStatus;
}
export type SubmitStateRPCResponseParams = SubmitStateResponseParams; // for backward compatibility

/**
 * Represents the parameters for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionResponseParams {
    /** The unique identifier for the application session. */
    appSessionId: Hex;
    /** The version number of the session. */
    version: number;
    /** The current status of the channel (e.g., "open", "closed"). */
    status: RPCChannelStatus;
}
export type CloseAppSessionRPCResponseParams = CloseAppSessionResponseParams; // for backward compatibility

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
export type GetAppDefinitionRPCResponseParams = GetAppDefinitionResponseParams; // for backward compatibility

/**
 * Represents the parameters for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsResponseParams {
    /** The unique identifier for the application session. */
    appSessionId: Hex;
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
    createdAt: Date;
    /** The timestamp when the session was last updated. */
    updatedAt: Date;
}
export type GetAppSessionsRPCResponseParams = GetAppSessionsResponseParams; // for backward compatibility

export interface ServerSignature {
    /** The recovery value of the signature. */
    v: number;
    r: Hex;
    s: Hex;
}

export interface RPCAllocation {
    /** The destination address for the allocation. */
    destination: Address;
    /** The token contract address. */
    token: Address;
    /** The amount to allocate. */
    amount: BigInt;
}

/**
 * Represents the parameters for the 'resize_channel' RPC method.
 */
export interface ResizeChannelResponseParams {
    /** The unique identifier for the channel. */
    channelId: Hex;
    /** The encoded state data for the channel. */
    stateData: Hex;
    /** The intent type for the state update. */
    intent: number;
    /** The version number of the channel. */
    version: number;
    /** The list of allocations for the channel. */
    allocations: RPCAllocation[];
    /** The hash of the channel state. */
    stateHash: Hex;
    /** The server's signature for the state update. */
    serverSignature: ServerSignature;
}
export type ResizeChannelRPCResponseParams = ResizeChannelResponseParams; // for backward compatibility

/**
 * Represents the parameters for the 'close_channel' RPC method.
 */
export interface CloseChannelResponseParams {
    /** The unique identifier for the channel. */
    channelId: Hex;
    /** The intent type for the state update. */
    intent: number;
    /** The version number of the channel. */
    version: number;
    /** The encoded state data for the channel. */
    stateData: Hex;
    /** The list of final allocations for the channel. */
    allocations: RPCAllocation[];
    /** The hash of the channel state. */
    stateHash: Hex;
    /** The server's signature for the state update. */
    serverSignature: ServerSignature;
}
export type CloseChannelRPCResponseParams = CloseChannelResponseParams; // for backward compatibility

/**
 * Represents the parameters for the 'get_channels' RPC method.
 */
export type GetChannelsResponseParams = ChannelUpdate[];
export type GetChannelsRPCResponseParams = GetChannelsResponseParams; // for backward compatibility

/**
 * Represents the parameters for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryResponseParams {
    /** Unique identifier for the RPC entry. */
    id: number;
    /** The Ethereum address of the sender. */
    sender: Address;
    /** The request ID for the RPC call. */
    reqId: number;
    /** The RPC method name. */
    method: string;
    /** The JSON string of the request parameters. */
    params: string;
    /** The timestamp of the RPC call. */
    timestamp: number;
    /** Array of request signatures. */
    reqSig: Hex[];
    /** Array of response signatures. */
    resSig: Hex[];
    /** The JSON string of the response. */
    response: string;
}
export type GetRPCHistoryRPCResponseParams = GetRPCHistoryResponseParams; // for backward compatibility

/**
 * Represents the parameters for the 'get_assets' RPC method.
 */
export interface GetAssetsResponseParams {
    /** The token contract address. */
    token: Address;
    /** The chain ID where the asset exists. */
    chainId: number;
    /** The asset symbol (e.g., "ETH", "USDC"). */
    symbol: string;
    /** The number of decimal places for the asset. */
    decimals: number;
}
export type GetAssetsRPCResponseParams = GetAssetsResponseParams; // for backward compatibility

/**
 * Represents the response structure for an error response.
 */
export interface ErrorResponse extends GenericRPCMessage {
    method: RPCMethod.Error;
    params: ErrorResponseParams;
}

/**
 * Represents the response structure for the 'get_config' RPC method.
 */
export interface GetConfigResponse extends GenericRPCMessage {
    method: RPCMethod.GetConfig;
    params: GetConfigResponseParams;
}

/**
 * Represents the response structure for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesResponse extends GenericRPCMessage {
    method: RPCMethod.GetLedgerBalances;
    params: GetLedgerBalancesResponseParams[];
}

/**
 * Represents the response structure for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesResponse extends GenericRPCMessage {
    method: RPCMethod.GetLedgerEntries;
    params: GetLedgerEntriesResponseParams[];
}

/**
 * Represents the response structure for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionResponse extends GenericRPCMessage {
    method: RPCMethod.CreateAppSession;
    params: CreateAppSessionResponseParams;
}

/**
 * Represents the response structure for the 'submit_state' RPC method.
 */
export interface SubmitStateResponse extends GenericRPCMessage {
    method: RPCMethod.SubmitState;
    params: SubmitStateResponseParams;
}

/**
 * Represents the response structure for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionResponse extends GenericRPCMessage {
    method: RPCMethod.CloseAppSession;
    params: CloseAppSessionResponseParams;
}

/**
 * Represents the response structure for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionResponse extends GenericRPCMessage {
    method: RPCMethod.GetAppDefinition;
    params: GetAppDefinitionResponseParams;
}

/**
 * Represents the response structure for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsResponse extends GenericRPCMessage {
    method: RPCMethod.GetAppSessions;
    params: GetAppSessionsResponseParams[];
}

/**
 * Represents the response structure for the 'resize_channel' RPC method.
 */
export interface ResizeChannelResponse extends GenericRPCMessage {
    method: RPCMethod.ResizeChannel;
    params: ResizeChannelResponseParams;
}

/**
 * Represents the response structure for the 'close_channel' RPC method.
 */
export interface CloseChannelResponse extends GenericRPCMessage {
    method: RPCMethod.CloseChannel;
    params: CloseChannelResponseParams;
}

/**
 * Represents the response structure for the 'get_channels' RPC method.
 */
export interface GetChannelsResponse extends GenericRPCMessage {
    method: RPCMethod.GetChannels;
    params: GetChannelsResponseParams;
}

/**
 * Represents the response structure for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryResponse extends GenericRPCMessage {
    method: RPCMethod.GetRPCHistory;
    params: GetRPCHistoryResponseParams[];
}

/**
 * Represents the response structure for the 'get_assets' RPC method.
 */
export interface GetAssetsResponse extends GenericRPCMessage {
    method: RPCMethod.GetAssets;
    params: GetAssetsResponseParams[];
}

export interface AssetsResponse extends GenericRPCMessage {
    method: RPCMethod.Assets;
    params: GetAssetsResponseParams[];
}

/**
 * Represents the response structure for the 'auth_verify' RPC method.
 */
export interface AuthVerifyResponse extends GenericRPCMessage {
    method: RPCMethod.AuthVerify;
    params: AuthVerifyResponseParams;
}

/**
 * Represents the parameters for the 'auth_request' RPC method.
 */
export interface AuthRequestResponseParams {
    /** The challenge message to be signed by the client for authentication. */
    challengeMessage: string;
}
export type AuthRequestRPCResponseParams = AuthRequestResponseParams; // for backward compatibility

/**
 * Represents the response structure for the 'auth_request' RPC method.
 */
export interface AuthRequestResponse extends GenericRPCMessage {
    method: RPCMethod.AuthRequest;
    params: AuthRequestResponseParams;
}

/**
 * Represents the response parameters for the 'message' RPC method.
 */
export interface MessageResponseParams {
    // Message response parameters are handled by the application
}
export type MessageRPCResponseParams = MessageResponseParams; // for backward compatibility

/**
 * Represents the response structure for the 'message' RPC method.
 */
export interface MessageResponse extends GenericRPCMessage {
    method: RPCMethod.Message;
    params: MessageResponseParams;
}

/**
 * Represents the parameters for the 'bu' RPC method.
 */
export interface BalanceUpdateResponseParams {
    /** The asset symbol (e.g., "ETH", "USDC"). */
    asset: string;
    /** The balance amount. */
    amount: string;
}
export type BalanceUpdateRPCResponseParams = BalanceUpdateResponseParams; // for backward compatibility

/**
 * Represents the response structure for the 'bu' RPC method.
 */
export interface BalanceUpdateResponse extends GenericRPCMessage {
    method: RPCMethod.BalanceUpdate;
    params: BalanceUpdateResponseParams[];
}

/**
 * Represents the parameters for the 'channels' RPC method.
 */
export type ChannelsUpdateResponseParams = ChannelUpdate;

/**
 * Represents the response structure for the 'channels_update' RPC method.
 */
export interface ChannelsUpdateResponse extends GenericRPCMessage {
    method: RPCMethod.ChannelsUpdate;
    params: ChannelsUpdateResponseParams;
}

/**
 * Represents the parameters for the 'cu' RPC method.
 */
export type ChannelUpdateResponseParams = ChannelUpdate;
export type ChannelUpdateRPCResponseParams = ChannelUpdateResponseParams; // for backward compatibility

/**
 * Represents the response structure for the 'cu' RPC method.
 */
export interface ChannelUpdateResponse extends GenericRPCMessage {
    method: RPCMethod.ChannelUpdate;
    params: ChannelUpdateResponseParams;
}

/**
 * Represents the parameters for the 'ping' RPC method.
 */
export interface PingResponseParams {
    // No parameters needed for ping
}
export type PingRPCResponseParams = PingResponseParams; // for backward compatibility

/**
 * Represents the response structure for the 'ping' RPC method.
 */
export interface PingResponse extends GenericRPCMessage {
    method: RPCMethod.Ping;
    params: PingResponseParams;
}

/**
 * Represents the parameters for the 'pong' RPC method.
 */
export interface PongResponseParams {
    // No parameters needed for pong
}
export type PongRPCResponseParams = PongResponseParams; // for backward compatibility

/**
 * Represents the response structure for the 'pong' RPC method.
 */
export interface PongResponse extends GenericRPCMessage {
    method: RPCMethod.Pong;
    params: PongResponseParams;
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
    createdAt: Date;
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
    | AssetsResponse
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
// Helper type to extract the response type for a given method
export type ExtractResponseByMethod<M extends RPCMethod> = Extract<RPCResponse, { method: M }>;

export type RPCResponseParams = ExtractResponseByMethod<RPCMethod>['params'];

export type RPCResponseParamsByMethod = {
    [M in RPCMethod]: ExtractResponseByMethod<M>['params'];
};
