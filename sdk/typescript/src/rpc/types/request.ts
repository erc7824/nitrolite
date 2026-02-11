import { Address, Hex } from 'viem';
import {
    RPCMethod,
    GenericRPCMessage,
    RPCAppDefinition,
    RPCAppSessionAllocation,
    RPCAppStateIntent,
    RPCState,
    RPCAppStateUpdate,
    RPCSignedAppStateUpdate,
} from '.';

/**
 * Represents the request structure for the 'get_config' RPC method.
 */
export interface GetConfigRequest extends GenericRPCMessage {
    method: RPCMethod.GetConfig;
    params: {};
}

/**
 * Represents the request structure for the 'get_session_keys' RPC method.
 */
export interface GetSessionKeysRequest extends GenericRPCMessage {
    method: RPCMethod.GetSessionKeys;
    params: {
        /** User's wallet address */
        wallet: Address;
    };
}

/**
 * Represents the request structure for the 'revoke_session_key' RPC method.
 */
export interface RevokeSessionKeyRequest extends GenericRPCMessage {
    method: RPCMethod.RevokeSessionKey;
    params: {
        /** The session key address to revoke */
        session_key: Address;
    };
}

/**
 * Represents the request structure for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRequest extends GenericRPCMessage {
    method: RPCMethod.CreateAppSession;
    params: {
        /** Application definition including participants and quorum */
        definition: RPCAppDefinition;
        /** Optional JSON stringified session data */
        session_data?: string;
        /** App Session creation signatures */
        quorum_sigs: Hex[];
    };
}

/**
 * Represents the request structure for the 'submit_app_state' RPC method.
 */
export interface SubmitAppStateRequest extends GenericRPCMessage {
    method: RPCMethod.SubmitAppState;
    params: SubmitAppStateRequestParams;
}

/**
 * Represents the request structure for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRequest extends GenericRPCMessage {
    method: RPCMethod.GetAppDefinition;
    params: {
        /** The unique identifier of the application session to retrieve the definition for. */
        app_session_id: Hex;
    };
}

/**
 * Represents the request structure for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRequest extends GenericRPCMessage {
    method: RPCMethod.GetAppSessions;
    params: {
        /** Filter by application session ID */
        app_session_id?: string;
        /** Filter by participant wallet address */
        participant?: Address;
        /** Filter by status (open/closed) */
        status?: string;
        /** Pagination parameters */
        pagination?: {
            /** Number of items to skip */
            offset?: number;
            /** Number of items to return */
            limit?: number;
            /** Sort order */
            sort?: 'asc' | 'desc';
        };
    };
}

/**
 * Represents the request structure for the 'request_creation' RPC method.
 * Requests the node to create a new channel with the provided state and definition.
 */
export interface CreateChannelRequest extends GenericRPCMessage {
    method: RPCMethod.CreateChannel;
    params: {
        /** The state to be submitted */
        state: RPCState;
        /** Definition of the channel to be created */
        channel_definition: {
            /** A unique number to prevent replay attacks */
            nonce: number;
            /** Challenge period for the channel in seconds */
            challenge: number;
        };
    };
}

/**
 * Represents the request structure for the 'get_channels' RPC method.
 */
export interface GetChannelsRequest extends GenericRPCMessage {
    method: RPCMethod.GetChannels;
    params: {
        /** User's wallet address */
        wallet: Address;
        /** Filter by asset */
        asset?: string;
        /** Filter by status */
        status?: string;
        /** Pagination parameters */
        pagination?: {
            /** Number of items to skip */
            offset?: number;
            /** Number of items to return */
            limit?: number;
            /** Sort order */
            sort?: 'asc' | 'desc';
        };
    };
}

/**
 * Represents the request structure for the 'get_assets' RPC method.
 */
export interface GetAssetsRequest extends GenericRPCMessage {
    method: RPCMethod.GetAssets;
    params: {
        /** Optional chain ID to filter assets by network. If not provided, returns assets from all networks. */
        chain_id?: number;
    };
}

/**
 * Represents the request structure for the 'message' RPC method.
 */
export interface MessageRequest extends GenericRPCMessage {
    method: RPCMethod.Message;
    /** The message parameters are handled by the virtual application */
    params: Record<string, unknown>;
}

/**
 * Represents the request structure for the 'ping' RPC method.
 */
export interface PingRequest extends GenericRPCMessage {
    method: RPCMethod.Ping;
    params: {};
}

/**
 * Represents the request parameters for the 'get_config' RPC method.
 */
export type GetConfigRequestParams = GetConfigRequest['params'];

/**
 * Represents the request parameters for the 'get_session_keys' RPC method.
 */
export type GetSessionKeysRequestParams = GetSessionKeysRequest['params'];

/**
 * Represents the request parameters for the 'revoke_session_key' RPC method.
 */
export type RevokeSessionKeyRequestParams = RevokeSessionKeyRequest['params'];

/**
 * Represents the request parameters for the 'create_app_session' RPC method.
 */
export type CreateAppSessionRequestParams = CreateAppSessionRequest['params'];

/**
 * Represents the request parameters for the 'submit_app_state' RPC method.
 */
export type SubmitAppStateRequestParams = {
    /** The application session state update */
    app_state_update: {
        /** The unique identifier of the application session to update */
        app_session_id: Hex;
        /** The intent of the state update */
        intent: RPCAppStateIntent;
        /** The state version number */
        version: number;
        /** The new allocation distribution among participants */
        allocations: RPCAppSessionAllocation[];
        /** Optional session data as a JSON string */
        session_data?: string;
    };
    /** Quorum signatures from participants */
    quorum_sigs: Hex[];
};

/**
 * Represents the request parameters for the 'get_app_definition' RPC method.
 */
export type GetAppDefinitionRequestParams = GetAppDefinitionRequest['params'];

/**
 * Represents the request parameters for the 'get_app_sessions' RPC method.
 */
export type GetAppSessionsRequestParams = GetAppSessionsRequest['params'];

/**
 * Represents the request parameters for the 'create_channel' RPC method.
 */
export type CreateChannelRequestParams = CreateChannelRequest['params'];

/**
 * Represents the request parameters for the 'get_channels' RPC method.
 */
export type GetChannelsRequestParams = GetChannelsRequest['params'];

/**
 * Represents the request parameters for the 'get_assets' RPC method.
 */
export type GetAssetsRequestParams = GetAssetsRequest['params'];

/**
 * Represents the request parameters for the 'message' RPC method.
 */
export type MessageRequestParams = MessageRequest['params'];

/**
 * Represents the request parameters for the 'ping' RPC method.
 */
export type PingRequestParams = PingRequest['params'];

/**
 * Request to get home channel information.
 */
export interface GetHomeChannelRequest extends GenericRPCMessage {
    method: RPCMethod.GetHomeChannel;
    params: {
        /** User's wallet address */
        wallet: Address;
        /** Asset symbol */
        asset: string;
    };
}

export type GetHomeChannelRequestParams = GetHomeChannelRequest['params'];

/**
 * Request to get escrow channel information.
 */
export interface GetEscrowChannelRequest extends GenericRPCMessage {
    method: RPCMethod.GetEscrowChannel;
    params: {
        /** Escrow channel ID */
        escrow_channel_id: string;
    };
}

export type GetEscrowChannelRequestParams = GetEscrowChannelRequest['params'];

/**
 * Request to get the latest state for a user's asset.
 */
export interface GetLatestStateRequest extends GenericRPCMessage {
    method: RPCMethod.GetLatestState;
    params: {
        /** User's wallet address */
        wallet: Address;
        /** Asset symbol */
        asset: string;
        /** Get only signed states */
        only_signed: boolean;
    };
}

export type GetLatestStateRequestParams = GetLatestStateRequest['params'];

/**
 * Request to get state history with filtering.
 */
export interface GetStatesRequest extends GenericRPCMessage {
    method: RPCMethod.GetStates;
    params: {
        /** User's wallet address */
        wallet: Address;
        /** Asset symbol */
        asset: string;
        /** Filter by user epoch index */
        epoch?: number;
        /** Filter by Home/Escrow Channel ID */
        channel_id?: string;
        /** Return only signed states */
        only_signed: boolean;
        /** Pagination parameters */
        pagination?: {
            /** Number of items to skip */
            offset?: number;
            /** Number of items to return */
            limit?: number;
            /** Sort order */
            sort?: 'asc' | 'desc';
        };
    };
}

export type GetStatesRequestParams = GetStatesRequest['params'];

/**
 * Request to submit a state transition.
 */
export interface SubmitStateRequest extends GenericRPCMessage {
    method: RPCMethod.SubmitState;
    params: {
        /** State to submit */
        state: RPCState;
    };
}

export type SubmitStateRequestParams = SubmitStateRequest['params'];

/**
 * Request to submit an app session deposit state.
 */
export interface SubmitDepositStateRequest extends GenericRPCMessage {
    method: RPCMethod.SubmitDepositState;
    params: {
        /** App state update */
        app_state_update: RPCAppStateUpdate;
        /** Quorum signatures */
        quorum_sigs: Hex[];
        /** Associated user state */
        user_state: RPCState;
    };
}

export type SubmitDepositStateRequestParams = SubmitDepositStateRequest['params'];

/**
 * Request to atomically rebalance multiple app sessions.
 */
export interface RebalanceAppSessionsRequest extends GenericRPCMessage {
    method: RPCMethod.RebalanceAppSessions;
    params: {
        /** List of signed updates with intent 'rebalance' */
        signed_updates: RPCSignedAppStateUpdate[];
    };
}

export type RebalanceAppSessionsRequestParams = RebalanceAppSessionsRequest['params'];

/**
 * Request to register a new session key.
 */
export interface RegisterRequest extends GenericRPCMessage {
    method: RPCMethod.Register;
    params: {
        /** User wallet address */
        address: Address;
        /** Session key address (optional) */
        session_key?: Address;
        /** Application name (optional) */
        application?: string;
        /** Asset allowances (optional) */
        allowances?: Array<{ asset: string; allowance: string }>;
        /** Permission scope (optional) */
        scope?: string;
        /** Expiration timestamp (optional) */
        expires_at?: number;
    };
}

export type RegisterRequestParams = RegisterRequest['params'];

/**
 * Request to get user balances.
 */
export interface GetBalancesRequest extends GenericRPCMessage {
    method: RPCMethod.GetBalances;
    params: {
        /** User's wallet address */
        wallet: Address;
    };
}

export type GetBalancesRequestParams = GetBalancesRequest['params'];

/**
 * Request to get transaction history.
 */
export interface GetTransactionsRequest extends GenericRPCMessage {
    method: RPCMethod.GetTransactions;
    params: {
        /** User's wallet address */
        wallet: Address;
        /** Asset filter (optional) */
        asset?: string;
        /** Transaction type filter (optional) */
        tx_type?: string;
        /** Start time filter (optional) */
        from_time?: number;
        /** End time filter (optional) */
        to_time?: number;
        /** Pagination parameters */
        pagination?: {
            /** Number of items to skip */
            offset?: number;
            /** Number of items to return */
            limit?: number;
            /** Sort order */
            sort?: 'asc' | 'desc';
        };
    };
}

export type GetTransactionsRequestParams = GetTransactionsRequest['params'];

/**
 * Optional filters for get_transactions request.
 */
export interface GetTransactionsOptions {
    asset?: string;
    tx_type?: string;
    from_time?: number;
    to_time?: number;
    pagination?: {
        offset?: number;
        limit?: number;
        sort?: 'asc' | 'desc';
    };
}

/**
 * Optional filters for get_channels request.
 */
export interface GetChannelsOptions {
    asset?: string;
    status?: string;
    pagination?: {
        offset?: number;
        limit?: number;
        sort?: 'asc' | 'desc';
    };
}

/**
 * Optional filters for get_states request.
 */
export interface GetStatesOptions {
    epoch?: number;
    channel_id?: string;
    pagination?: {
        offset?: number;
        limit?: number;
        sort?: 'asc' | 'desc';
    };
}

/**
 * Optional filters for get_app_sessions request.
 */
export interface GetAppSessionsOptions {
    app_session_id?: string;
    participant?: Address;
    status?: string;
    pagination?: {
        offset?: number;
        limit?: number;
        sort?: 'asc' | 'desc';
    };
}

/**
 * Optional configuration for session key registration.
 */
export interface RegisterOptions {
    session_key?: Address;
    application?: string;
    allowances?: Array<{ asset: string; allowance: string }>;
    scope?: string;
    expires_at?: number;
}

/**
 * Union type for all possible RPC request types.
 * This allows for type-safe handling of different request structures.
 */
export type RPCRequest =
    // Node
    | GetConfigRequest
    | GetAssetsRequest
    | PingRequest
    // User
    | GetBalancesRequest
    | GetTransactionsRequest
    // Channels
    | GetChannelsRequest
    | GetHomeChannelRequest
    | GetEscrowChannelRequest
    | GetLatestStateRequest
    | GetStatesRequest
    | CreateChannelRequest
    | SubmitStateRequest
    // App Sessions
    | GetAppDefinitionRequest
    | GetAppSessionsRequest
    | CreateAppSessionRequest
    | SubmitAppStateRequest
    | SubmitDepositStateRequest
    | RebalanceAppSessionsRequest
    // Session Keys
    | RegisterRequest
    | GetSessionKeysRequest
    | RevokeSessionKeyRequest
    // Server Push
    | MessageRequest;

/**
 * Maps RPC methods to their corresponding request parameter types.
 */
export type RPCRequestParamsByMethod = {
    // Node
    [RPCMethod.GetConfig]: GetConfigRequestParams;
    [RPCMethod.GetAssets]: GetAssetsRequestParams;
    [RPCMethod.Ping]: PingRequestParams;
    // User
    [RPCMethod.GetBalances]: GetBalancesRequestParams;
    [RPCMethod.GetTransactions]: GetTransactionsRequestParams;
    // Channels
    [RPCMethod.GetChannels]: GetChannelsRequestParams;
    [RPCMethod.GetHomeChannel]: GetHomeChannelRequestParams;
    [RPCMethod.GetEscrowChannel]: GetEscrowChannelRequestParams;
    [RPCMethod.GetLatestState]: GetLatestStateRequestParams;
    [RPCMethod.GetStates]: GetStatesRequestParams;
    [RPCMethod.CreateChannel]: CreateChannelRequestParams;
    [RPCMethod.SubmitState]: SubmitStateRequestParams;
    // App Sessions
    [RPCMethod.GetAppDefinition]: GetAppDefinitionRequestParams;
    [RPCMethod.GetAppSessions]: GetAppSessionsRequestParams;
    [RPCMethod.CreateAppSession]: CreateAppSessionRequestParams;
    [RPCMethod.SubmitAppState]: SubmitAppStateRequestParams;
    [RPCMethod.SubmitDepositState]: SubmitDepositStateRequestParams;
    [RPCMethod.RebalanceAppSessions]: RebalanceAppSessionsRequestParams;
    // Session Keys
    [RPCMethod.Register]: RegisterRequestParams;
    [RPCMethod.GetSessionKeys]: GetSessionKeysRequestParams;
    [RPCMethod.RevokeSessionKey]: RevokeSessionKeyRequestParams;
    // Server Push
    [RPCMethod.Message]: MessageRequestParams;
};
