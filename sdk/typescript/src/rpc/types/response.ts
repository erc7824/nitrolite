import { Address, Hex } from 'viem';
import {
    RPCMethod,
    GenericRPCMessage,
    RPCAppDefinition,
    RPCChannelStatus,
    RPCChannelUpdate,
    RPCNetworkInfo,
    RPCBalance,
    RPCLedgerEntry,
    RPCAppSession,
    RPCHistoryEntry,
    RPCAsset,
    RPCTransaction,
    RPCChannelUpdateWithWallet,
    RPCChannelOperation,
    RPCChannel,
    RPCSessionKey,
} from '.';

/**
 * Represents the response structure for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeResponse extends GenericRPCMessage {
    method: RPCMethod.AuthChallenge;
    params: {
        /** The challenge message to be signed by the client for authentication. */
        challengeMessage: string;
    };
}

/**
 * Represents the response structure for an error response.
 */
export interface ErrorResponse extends GenericRPCMessage {
    method: RPCMethod.Error;
    params: {
        /** The error message describing what went wrong. */
        error: string;
    };
}

/**
 * Represents the response structure for the 'get_config' RPC method.
 */
export interface GetConfigResponse extends GenericRPCMessage {
    method: RPCMethod.GetConfig;
    params: {
        /** The Ethereum address of the broker. */
        brokerAddress: Address;
        /** List of supported networks and their configurations. */
        networks: RPCNetworkInfo[];
    };
}

/**
 * Represents the response structure for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesResponse extends GenericRPCMessage {
    method: RPCMethod.GetLedgerBalances;
    params: {
        /** List of balances for each asset in the ledger. */
        ledgerBalances: RPCBalance[];
    };
}

/**
 * Represents the response structure for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesResponse extends GenericRPCMessage {
    method: RPCMethod.GetLedgerEntries;
    params: {
        /** List of ledger entries containing transaction details. */
        ledgerEntries: RPCLedgerEntry[];
    };
}

/**
 * Represents the response structure for the 'get_transactions' RPC method.
 */
export interface GetLedgerTransactionsResponse extends GenericRPCMessage {
    method: RPCMethod.GetLedgerTransactions;
    params: {
        /** List of transactions in the ledger. */
        ledgerTransactions: RPCTransaction[];
    };
}

/**
 * Represents the response structure for the 'get_user_tag' RPC method.
 */
export interface GetUserTagResponse extends GenericRPCMessage {
    method: RPCMethod.GetUserTag;
    params: {
        /** The user's unique tag identifier. */
        tag: string;
    };
}

/**
 * Represents the response structure for the 'get_session_keys' RPC method.
 */
export interface GetSessionKeysResponse extends GenericRPCMessage {
    method: RPCMethod.GetSessionKeys;
    params: {
        /** Array of active session keys for the authenticated user. */
        sessionKeys: RPCSessionKey[];
    };
}

/**
 * Represents the response structure for the 'revoke_session_key' RPC method.
 */
export interface RevokeSessionKeyResponse extends GenericRPCMessage {
    method: RPCMethod.RevokeSessionKey;
    params: {
        /** The session key address that was revoked. */
        sessionKey: Address;
    };
}

/**
 * Represents the response structure for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionResponse extends GenericRPCMessage {
    method: RPCMethod.CreateAppSession;
    params: {
        /** The unique identifier for the application session. */
        appSessionId: Hex;
        /** The version number of the session. */
        version: number;
        /** The current status of the channel (e.g., "open", "closed"). */
        status: RPCChannelStatus;
    };
}

/**
 * Represents the response structure for the 'submit_app_state' RPC method.
 */
export interface SubmitAppStateResponse extends GenericRPCMessage {
    method: RPCMethod.SubmitAppState;
    params: {
        /** The unique identifier for the application session. */
        appSessionId: Hex;
        /** The version number of the session. */
        version: number;
        /** The current status of the channel (e.g., "open", "closed"). */
        status: RPCChannelStatus;
    };
}

/**
 * Represents the response structure for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionResponse extends GenericRPCMessage {
    method: RPCMethod.CloseAppSession;
    params: {
        /** The unique identifier for the application session. */
        appSessionId: Hex;
        /** The version number of the session. */
        version: number;
        /** The current status of the channel (e.g., "open", "closed"). */
        status: RPCChannelStatus;
    };
}

/**
 * Represents the response structure for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionResponse extends GenericRPCMessage {
    method: RPCMethod.GetAppDefinition;
    params: {
        /** Protocol identifies the version of the application protocol */
        protocol: string;
        /** An array of participant addresses (Ethereum addresses) involved in the application. Must have at least 2 participants. */
        participants: Address[];
        /** An array representing the relative weights or stakes of participants, often used for dispute resolution or allocation calculations. Order corresponds to the participants array. */
        weights: number[];
        /** The number of participants required to reach consensus or approve state updates. */
        quorum: number;
        /** A parameter related to the challenge period or mechanism within the application's protocol, in seconds. */
        challenge: number;
        /** A unique number used once, often for preventing replay attacks or ensuring uniqueness of the application instance. Must be non-zero. */
        nonce: number;
    };
}

/**
 * Represents the response structure for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsResponse extends GenericRPCMessage {
    method: RPCMethod.GetAppSessions;
    params: {
        appSessions: RPCAppSession[];
    };
}

/**
 * Represents the response structure for the 'create_channel' RPC method.
 */
export interface CreateChannelResponse extends GenericRPCMessage {
    method: RPCMethod.CreateChannel;
    params: RPCChannelOperation & {
        channel: RPCChannel;
    };
}

/**
 * Represents the response structure for the 'resize_channel' RPC method.
 */
export interface ResizeChannelResponse extends GenericRPCMessage {
    method: RPCMethod.ResizeChannel;
    params: RPCChannelOperation;
}

/**
 * Represents the response structure for the 'close_channel' RPC method.
 */
export interface CloseChannelResponse extends GenericRPCMessage {
    method: RPCMethod.CloseChannel;
    params: RPCChannelOperation;
}

/**
 * Represents the response structure for the 'get_channels' RPC method.
 */
export interface GetChannelsResponse extends GenericRPCMessage {
    method: RPCMethod.GetChannels;
    params: {
        /** List of channel updates containing information about each channel. */
        channels: RPCChannelUpdateWithWallet[];
    };
}

/**
 * Represents the response structure for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryResponse extends GenericRPCMessage {
    method: RPCMethod.GetRPCHistory;
    params: {
        /** List of RPC entries containing historical RPC calls and their responses. */
        rpcEntries: RPCHistoryEntry[];
    };
}

/**
 * Represents the response structure for the 'get_assets' RPC method.
 */
export interface GetAssetsResponse extends GenericRPCMessage {
    method: RPCMethod.GetAssets;
    params: {
        /** List of assets available in the clearnode. */
        assets: RPCAsset[];
    };
}

/**
 * Represents the response structure for the 'assets' RPC method.
 */
export interface AssetsResponse extends GenericRPCMessage {
    method: RPCMethod.Assets;
    params: {
        /** List of assets available in the clearnode. */
        assets: RPCAsset[];
    };
}

/**
 * Represents the response structure for the 'auth_verify' RPC method.
 */
export interface AuthVerifyResponse extends GenericRPCMessage {
    method: RPCMethod.AuthVerify;
    params: {
        address: Address;
        sessionKey: Address;
        success: boolean;
        /** Available only if challenge auth method was used in {@link AuthVerifyRequest} during the call to {@link RPCMethod.AuthRequest} */
        jwtToken?: string;
    };
}

/**
 * Represents the response structure for the 'auth_request' RPC method.
 */
export interface AuthRequestResponse extends GenericRPCMessage {
    method: RPCMethod.AuthRequest;
    params: {
        /** The challenge message to be signed by the client for authentication. */
        challengeMessage: string;
    };
}

/**
 * Represents the response structure for the 'message' RPC method.
 */
export interface MessageResponse extends GenericRPCMessage {
    method: RPCMethod.Message;
    params: {};
}

/**
 * Represents the response structure for the 'bu' RPC method.
 */
export interface BalanceUpdateResponse extends GenericRPCMessage {
    method: RPCMethod.BalanceUpdate;
    params: {
        /** List of balance updates. */
        balanceUpdates: RPCBalance[];
    };
}

/**
 * Represents the response structure for the 'channels' RPC method.
 */
export interface ChannelsUpdateResponse extends GenericRPCMessage {
    method: RPCMethod.ChannelsUpdate;
    params: {
        /** List of channel updates. */
        channels: RPCChannelUpdate[];
    };
}

/**
 * Represents the response structure for the 'cu' RPC method.
 */
export interface ChannelUpdateResponse extends GenericRPCMessage {
    method: RPCMethod.ChannelUpdate;
    params: RPCChannelUpdate;
}

/**
 * Represents the response structure for the 'ping' RPC method.
 */
export interface PingResponse extends GenericRPCMessage {
    method: RPCMethod.Ping;
    params: {};
}

/**
 * Represents the response structure for the 'pong' RPC method.
 */
export interface PongResponse extends GenericRPCMessage {
    method: RPCMethod.Pong;
    params: {};
}

/**
 * Represents the response structure for the 'transfer' RPC method.
 */
export interface TransferResponse extends GenericRPCMessage {
    method: RPCMethod.Transfer;
    params: {
        /** List of transactions representing transfers. */
        transactions: RPCTransaction[];
    };
}

/**
 * Represents the response structure for the 'cleanup_session_key_cache' RPC method.
 */
export interface CleanupSessionKeyCacheResponse extends GenericRPCMessage {
    method: RPCMethod.CleanupSessionKeyCache;
    params: {};
}

/**
 * Represents the response structure for the 'transfer_notification' RPC method.
 */
export interface TransferNotificationResponse extends GenericRPCMessage {
    method: RPCMethod.TransferNotification;
    params: {
        /** List of transactions representing transfers. */
        transactions: RPCTransaction[];
    };
}

/**
 * Represents the parameters for the 'auth_challenge' RPC method.
 */
export type AuthChallengeResponseParams = AuthChallengeResponse['params'];

/**
 * Represents the parameters for the 'auth_verify' RPC method.
 */
export type AuthVerifyResponseParams = AuthVerifyResponse['params'];

/**
 * Represents the parameters for the 'error' RPC method.
 */
export type ErrorResponseParams = ErrorResponse['params'];

/**
 * Represents the parameters for the 'get_config' RPC method.
 */
export type GetConfigResponseParams = GetConfigResponse['params'];

/**
 * Represents the parameters for the 'get_ledger_balances' RPC method.
 */
export type GetLedgerBalancesResponseParams = GetLedgerBalancesResponse['params'];

/**
 * Represents the parameters for the 'get_ledger_entries' RPC method.
 */
export type GetLedgerEntriesResponseParams = GetLedgerEntriesResponse['params'];

/**
 * Represents the parameters for the 'get_ledger_transactions' RPC method.
 */
export type GetLedgerTransactionsResponseParams = GetLedgerTransactionsResponse['params'];

/**
 * Represents the parameters for the 'get_user_tag' RPC method.
 */
export type GetUserTagResponseParams = GetUserTagResponse['params'];

/**
 * Represents the parameters for the 'get_session_keys' RPC method.
 */
export type GetSessionKeysResponseParams = GetSessionKeysResponse['params'];

/**
 * Represents the parameters for the 'revoke_session_key' RPC method.
 */
export type RevokeSessionKeyResponseParams = RevokeSessionKeyResponse['params'];

/**
 * Represents the parameters for the 'create_app_session' RPC method.
 */
export type CreateAppSessionResponseParams = CreateAppSessionResponse['params'];

/**
 * Represents the parameters for the 'submit_app_state' RPC method.
 */
export type SubmitAppStateResponseParams = SubmitAppStateResponse['params'];

/**
 * Represents the parameters for the 'close_app_session' RPC method.
 */
export type CloseAppSessionResponseParams = CloseAppSessionResponse['params'];

/**
 * Represents the parameters for the 'get_app_definition' RPC method.
 */
export type GetAppDefinitionResponseParams = GetAppDefinitionResponse['params'];

/**
 * Represents the parameters for the 'get_app_sessions' RPC method.
 */
export type GetAppSessionsResponseParams = GetAppSessionsResponse['params'];

/**
 * Represents the parameters for the 'create_channel' RPC method.
 */
export type CreateChannelResponseParams = CreateChannelResponse['params'];

/**
 * Represents the parameters for the 'resize_channel' RPC method.
 */
export type ResizeChannelResponseParams = ResizeChannelResponse['params'];

/**
 * Represents the parameters for the 'close_channel' RPC method.
 */
export type CloseChannelResponseParams = CloseChannelResponse['params'];

/**
 * Represents the parameters for the 'get_channels' RPC method.
 */
export type GetChannelsResponseParams = GetChannelsResponse['params'];

/**
 * Represents the parameters for the 'get_rpc_history' RPC method.
 */
export type GetRPCHistoryResponseParams = GetRPCHistoryResponse['params'];

/**
 * Represents the parameters for the 'get_assets' RPC method.
 */
export type GetAssetsResponseParams = GetAssetsResponse['params'];

/**
 * Represents the parameters for the 'assets' RPC method.
 */
export type AssetsResponseParams = AssetsResponse['params'];

/**
 * Represents the parameters for the 'auth_request' RPC method.
 */
export type AuthRequestResponseParams = AuthRequestResponse['params'];

/**
 * Represents the parameters for the 'message' RPC method.
 */
export type MessageResponseParams = MessageResponse['params'];

/**
 * Represents the parameters for the 'bu' RPC method.
 */
export type BalanceUpdateResponseParams = BalanceUpdateResponse['params'];

/**
 * Represents the parameters for the 'channels' RPC method.
 */
export type ChannelsUpdateResponseParams = ChannelsUpdateResponse['params'];

/**
 * Represents the parameters for the 'cu' RPC method.
 */
export type ChannelUpdateResponseParams = ChannelUpdateResponse['params'];

/**
 * Represents the parameters for the 'ping' RPC method.
 */
export type PingResponseParams = PingResponse['params'];

/**
 * Represents the parameters for the 'pong' RPC method.
 */
export type PongResponseParams = PongResponse['params'];

/**
 * Represents the parameters for the 'transfer' RPC method.
 */
export type TransferResponseParams = TransferResponse['params'];

/**
 * Represents the parameters for the 'cleanup_session_key_cache' RPC method.
 */
export type CleanupSessionKeyCacheResponseParams = CleanupSessionKeyCacheResponse['params'];

/**
 * Represents the parameters for the 'tr' RPC method.
 */
export type TransferNotificationResponseParams = TransferNotificationResponse['params'];

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
    | GetLedgerTransactionsResponse
    | GetUserTagResponse
    | GetSessionKeysResponse
    | RevokeSessionKeyResponse
    | CreateAppSessionResponse
    | SubmitAppStateResponse
    | CloseAppSessionResponse
    | GetAppDefinitionResponse
    | GetAppSessionsResponse
    | CreateChannelResponse
    | ResizeChannelResponse
    | CloseChannelResponse
    | GetChannelsResponse
    | GetRPCHistoryResponse
    | GetAssetsResponse
    | AssetsResponse
    | PingResponse
    | PongResponse
    | TransferResponse
    | CleanupSessionKeyCacheResponse
    | MessageResponse
    | BalanceUpdateResponse
    | ChannelsUpdateResponse
    | ChannelUpdateResponse
    | TransferNotificationResponse;

/**
 * Maps RPC methods to their corresponding parameter types.
 */
// Helper type to extract the response type for a given method
export type ExtractResponseByMethod<M extends RPCMethod> = Extract<RPCResponse, { method: M }>;

export type RPCResponseParams = ExtractResponseByMethod<RPCMethod>['params'];

export type RPCResponseParamsByMethod = {
    [M in RPCMethod]: ExtractResponseByMethod<M>['params'];
};
