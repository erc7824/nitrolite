import { Address, Hex } from 'viem';
import {
    RPCMethod,
    GenericRPCMessage,
    RPCAppDefinition,
    RPCChannelStatus,
    RPCTransferAllocation,
    RPCAppSessionAllocation,
    RPCAllowance,
    GetLedgerTransactionsFilters,
    RPCAppStateIntent,
    RPCProtocolVersion,
} from '.';

/**
 * Represents the request structure for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeRequest extends GenericRPCMessage {
    method: RPCMethod.AuthChallenge;
    params: {
        /** The challenge message to be signed by the client for authentication. */
        challenge_message: string;
    };
}

/**
 * Represents the request structure for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRequest extends GenericRPCMessage {
    method: RPCMethod.AuthVerify;
    params:
        | {
              /** JSON Web Token for authentication. */
              jwt: string;
          }
        | {
              /** The challenge token received from auth_challenge response. Used to verify the client's signature and prevent replay attacks. */
              challenge: string;
          };
}

/**
 * Represents the request structure for the 'get_config' RPC method.
 */
export interface GetConfigRequest extends GenericRPCMessage {
    method: RPCMethod.GetConfig;
    params: {};
}

/**
 * Represents the request structure for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRequest extends GenericRPCMessage {
    method: RPCMethod.GetLedgerBalances;
    params: {
        /** Optional account ID to filter balances. */
        account_id?: string;
    };
}

/**
 * Represents the request structure for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesRequest extends GenericRPCMessage {
    method: RPCMethod.GetLedgerEntries;
    params: {
        /** The account ID to filter ledger entries. */
        account_id?: string;
        /** The asset symbol to filter ledger entries. */
        asset?: string;
        /** Optional wallet address to filter ledger entries. If provided, overrides the authenticated wallet. */
        wallet?: Address;
    };
}

/**
 * Represents the request structure for the 'get_transactions' RPC method.
 */
export interface GetLedgerTransactionsRequest extends GenericRPCMessage {
    method: RPCMethod.GetLedgerTransactions;
    params: GetLedgerTransactionsFilters & {
        account_id: string;
    };
}

/**
 * Represents the request structure for the 'get_user_tag' RPC method.
 */
export interface GetUserTagRequest extends GenericRPCMessage {
    method: RPCMethod.GetUserTag;
    params: {};
}

/**
 * Represents the request structure for the 'get_session_keys' RPC method.
 */
export interface GetSessionKeysRequest extends GenericRPCMessage {
    method: RPCMethod.GetSessionKeys;
    params: {};
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
        /** The detailed definition of the application being created, including protocol, participants, weights, and quorum. */
        definition: RPCAppDefinition;
        /** The initial allocation distribution among participants. Each participant must have sufficient balance for their allocation. */
        allocations: RPCAppSessionAllocation[];
        /** Optional session data as a JSON string that can store application-specific state or metadata. */
        session_data?: string;
    };
}

/**
 * Represents the request structure for the 'submit_app_state' RPC method.
 */
export interface SubmitAppStateRequest extends GenericRPCMessage {
    method: RPCMethod.SubmitAppState;
    params: SubmitAppStateRequestParamsV02 | SubmitAppStateRequestParamsV04;
}

/**
 * Represents the request structure for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRequest extends GenericRPCMessage {
    method: RPCMethod.CloseAppSession;
    params: {
        /** The unique identifier of the application session to close. */
        app_session_id: Hex;
        /** The final allocation distribution among participants upon closing. Must include all participants and maintain total balance. */
        allocations: RPCAppSessionAllocation[];
        /** Optional session data as a JSON string that can store application-specific state or metadata. */
        session_data?: string;
    };
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
        /** Optional, The participant address to filter application sessions. */
        participant?: Address;
        /** Optional, The status to filter application sessions (e.g., "open", "closed"). */
        status?: RPCChannelStatus;
    };
}

/**
 * Represents the request structure for the 'create_channel' RPC method.
 * @remarks
 * Now it is impossible to request the Node for a non-zero initial user allocation as
 * channels are created with zero deposit and must be funded separately via resize_channel.
 * This constraint ensures proper funding sequencing and will be refined in the next major release.
 */
export interface CreateChannelRequest extends GenericRPCMessage {
    method: RPCMethod.CreateChannel;
    params: {
        /** The blockchain network ID where the channel should be created. */
        chain_id: number;
        /** The token contract address for the channel. */
        token: Address;
        /** NOTE: NO initial allocation for the user */
    };
}

/**
 * Represents the request structure for the 'resize_channel' RPC method.
 */
export interface ResizeChannelRequest extends GenericRPCMessage {
    method: RPCMethod.ResizeChannel;
    params: {
        /** The unique identifier of the channel to resize. */
        channel_id: Hex;
        /** Amount to resize the channel by (can be positive or negative). Required if allocate_amount is not provided. */
        resize_amount?: bigint;
        /** Amount to allocate from the unified balance to the channel. Required if resize_amount is not provided. */
        allocate_amount?: bigint;
        /** The address where the resized funds will be sent. */
        funds_destination: Address;
    };
}

/**
 * Represents the request structure for the 'close_channel' RPC method.
 */
export interface CloseChannelRequest extends GenericRPCMessage {
    method: RPCMethod.CloseChannel;
    params: {
        /** The unique identifier of the channel to close. */
        channel_id: Hex;
        /** The address where the channel funds will be sent upon closing. */
        funds_destination: Address;
    };
}

/**
 * Represents the request structure for the 'get_channels' RPC method.
 */
export interface GetChannelsRequest extends GenericRPCMessage {
    method: RPCMethod.GetChannels;
    params: {
        /** Optional, The participant address to filter channels. */
        participant?: Address;
        /** Optional, The status to filter channels (e.g., "open", "closed"). */
        status?: RPCChannelStatus;
    };
}

/**
 * Represents the request structure for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRequest extends GenericRPCMessage {
    method: RPCMethod.GetRPCHistory;
    params: {};
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
 * Represents the request structure for the 'auth_request' RPC method.
 */
export interface AuthRequest extends GenericRPCMessage {
    method: RPCMethod.AuthRequest;
    params: {
        /** The Ethereum address of the wallet being authorized. */
        address: Address;
        /** The session key address associated with the authentication attempt. */
        session_key: Address;
        /** The name of the application being authorized. */
        application: string;
        /** The allowances for the connection. */
        allowances: RPCAllowance[];
        /** The expiration timestamp for the authorization (Unix timestamp as bigint). */
        expires_at: bigint;
        /** The scope of the authorization. */
        scope: string;
    };
}

/**
 * Represents the request structure for the 'message' RPC method.
 */
export interface MessageRequest extends GenericRPCMessage {
    method: RPCMethod.Message;
    /** The message parameters are handled by the virtual application */
    params: any;
}

/**
 * Represents the request structure for the 'ping' RPC method.
 */
export interface PingRequest extends GenericRPCMessage {
    method: RPCMethod.Ping;
    params: {};
}

/**
 * Represents the request structure for the 'pong' RPC method.
 */
export interface PongRequest extends GenericRPCMessage {
    method: RPCMethod.Pong;
    params: {};
}

/**
 * Represents the request structure for the 'transfer' RPC method.
 */
export interface TransferRequest extends GenericRPCMessage {
    method: RPCMethod.Transfer;
    params: {
        /** The destination address to transfer assets to. Required if destination_user_tag is not provided. */
        destination?: Address;
        /** The destination user tag to transfer assets to. Required if destination is not provided. */
        destination_user_tag?: string;
        /** The assets and amounts to transfer. */
        allocations: RPCTransferAllocation[];
    };
}

/**
 * Represents the request structure for the 'cleanup_session_key_cache' RPC method.
 */
export interface CleanupSessionKeyCacheRequest extends GenericRPCMessage {
    method: RPCMethod.CleanupSessionKeyCache;
    params: {};
}

/** Represents the request parameters for the 'auth_challenge' RPC method. */
export type AuthChallengeRequestParams = AuthChallengeRequest['params'];

/**
 * Represents the request parameters for the 'auth_verify' RPC method.
 * Either JWT or challenge must be provided. JWT takes precedence over challenge.
 */
export type AuthVerifyRequestParams = AuthVerifyRequest['params'];

/**
 * Represents the request parameters for the 'get_config' RPC method.
 */
export type GetConfigRequestParams = GetConfigRequest['params'];

/**
 * Represents the request parameters for the 'get_ledger_balances' RPC method.
 */
export type GetLedgerBalancesRequestParams = GetLedgerBalancesRequest['params'];

/**
 * Represents the request parameters for the 'get_ledger_entries' RPC method.
 */
export type GetLedgerEntriesRequestParams = GetLedgerEntriesRequest['params'];

/**
 * Represents the request parameters for the 'get_ledger_transactions' RPC method.
 */
export type GetLedgerTransactionsRequestParams = GetLedgerTransactionsRequest['params'];

/**
 * Represents the request parameters for the 'get_user_tag' RPC method.
 */
export type GetUserTagRequestParams = GetUserTagRequest['params'];

/**
 * Represents the request parameters for the 'get_session_keys' RPC method.
 */
export type GetSessionKeysRequestParams = GetSessionKeysRequest['params'];

/**
 * Represents the request parameters for the 'create_app_session' RPC method.
 */
export type CreateAppSessionRequestParams = CreateAppSessionRequest['params'];

/**
 * Represents the request parameters for the 'submit_app_state' RPC method (NitroRPC/0.2).
 */
export type SubmitAppStateRequestParamsV02 = {
    /** The unique identifier of the application session to update. */
    app_session_id: Hex;
    /** The new allocation distribution among participants. Must include all participants and maintain total balance. */
    allocations: RPCAppSessionAllocation[];
    /** Optional session data as a JSON string that can store application-specific state or metadata. */
    session_data?: string;
};

/**
 * Represents the request parameters for the 'submit_app_state' RPC method (NitroRPC/0.4).
 */
export type SubmitAppStateRequestParamsV04 = {
    /** The unique identifier of the application session to update. */
    app_session_id: Hex;
    /** The intent of the state update */
    intent: RPCAppStateIntent;
    /** The state version number */
    version: number;
    /** The new allocation distribution among participants. Must include all participants and maintain total balance. */
    allocations: RPCAppSessionAllocation[];
    /** Optional session data as a JSON string that can store application-specific state or metadata. */
    session_data?: string;
};

/**
 * Maps protocol versions to their corresponding submit_app_state parameter types.
 */
export type SubmitAppStateParamsPerProtocol = {
    [RPCProtocolVersion.NitroRPC_0_2]: SubmitAppStateRequestParamsV02;
    [RPCProtocolVersion.NitroRPC_0_4]: SubmitAppStateRequestParamsV04;
};

/**
 * Represents the request parameters for the 'submit_app_state' RPC method.
 */
export type SubmitAppStateRequestParams = SubmitAppStateRequestParamsV02 | SubmitAppStateRequestParamsV04;

/**
 * Represents the request parameters for the 'close_app_session' RPC method.
 */
export type CloseAppSessionRequestParams = CloseAppSessionRequest['params'];

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
 * Represents the request parameters for the 'resize_channel' RPC method.
 */
export type ResizeChannelRequestParams = ResizeChannelRequest['params'];

/**
 * Represents the request parameters for the 'close_channel' RPC method.
 */
export type CloseChannelRequestParams = CloseChannelRequest['params'];

/**
 * Represents the request parameters for the 'get_channels' RPC method.
 */
export type GetChannelsRequestParams = GetChannelsRequest['params'];

/**
 * Represents the request parameters for the 'get_rpc_history' RPC method.
 */
export type GetRPCHistoryParams = GetRPCHistoryRequest['params'];

/**
 * Represents the request parameters for the 'get_assets' RPC method.
 */
export type GetAssetsRequestParams = GetAssetsRequest['params'];

/**
 * Represents the request parameters for the 'auth_request' RPC method.
 */
export type AuthRequestParams = AuthRequest['params'];

/**
 * Represents the request parameters for the 'message' RPC method.
 */
export type MessageRequestParams = MessageRequest['params'];

/**
 * Represents the request parameters for the 'ping' RPC method.
 */
export type PingRequestParams = PingRequest['params'];

/**
 * Represents the request parameters for the 'pong' RPC method.
 */
export type PongRequestParams = PongRequest['params'];

/**
 * Represents the request parameters for the 'transfer' RPC method.
 */
export type TransferRequestParams = TransferRequest['params'];

/**
 * Represents the request parameters for the 'cleanup_session_key_cache' RPC method.
 */
export type CleanupSessionKeyCacheRequestParams = CleanupSessionKeyCacheRequest['params'];

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
    | GetLedgerTransactionsRequest
    | GetUserTagRequest
    | GetSessionKeysRequest
    | RevokeSessionKeyRequest
    | CreateAppSessionRequest
    | SubmitAppStateRequest
    | CloseAppSessionRequest
    | GetAppDefinitionRequest
    | GetAppSessionsRequest
    | CreateChannelRequest
    | ResizeChannelRequest
    | CloseChannelRequest
    | GetChannelsRequest
    | GetRPCHistoryRequest
    | GetAssetsRequest
    | PingRequest
    | PongRequest
    | MessageRequest
    | TransferRequest
    | CleanupSessionKeyCacheRequest;

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
    [RPCMethod.GetLedgerTransactions]: GetLedgerTransactionsRequestParams;
    [RPCMethod.GetUserTag]: GetUserTagRequestParams;
    [RPCMethod.GetSessionKeys]: GetSessionKeysRequestParams;
    [RPCMethod.CreateAppSession]: CreateAppSessionRequestParams;
    [RPCMethod.SubmitAppState]: SubmitAppStateRequestParams;
    [RPCMethod.CloseAppSession]: CloseAppSessionRequestParams;
    [RPCMethod.GetAppDefinition]: GetAppDefinitionRequestParams;
    [RPCMethod.GetAppSessions]: GetAppSessionsRequestParams;
    [RPCMethod.CreateChannel]: CreateChannelRequestParams;
    [RPCMethod.ResizeChannel]: ResizeChannelRequestParams;
    [RPCMethod.CloseChannel]: CloseChannelRequestParams;
    [RPCMethod.GetChannels]: GetChannelsRequestParams;
    [RPCMethod.GetRPCHistory]: GetRPCHistoryParams;
    [RPCMethod.GetAssets]: GetAssetsRequestParams;
    [RPCMethod.Ping]: PingRequestParams;
    [RPCMethod.Pong]: PongRequestParams;
    [RPCMethod.Message]: any;
    [RPCMethod.Transfer]: TransferRequestParams;
    [RPCMethod.CleanupSessionKeyCache]: CleanupSessionKeyCacheRequestParams;
};
