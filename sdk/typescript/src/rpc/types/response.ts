import { Address, Hex } from 'viem';
import {
    RPCMethod,
    GenericRPCMessage,
    RPCChannelStatus,
    RPCNetworkInfo,
    RPCAppSession,
    RPCAsset,
    RPCTransaction,
    RPCChannel,
    RPCSessionKey,
    RPCState,
    PaginationMetadata,
    RPCBalanceEntry,
} from '.';

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
 * Represents the response structure for the 'node.v1.ping' RPC method.
 */
export interface PingResponse extends GenericRPCMessage {
    method: RPCMethod.Ping;
    params: {};
}

/**
 * Represents the response structure for the 'node.v1.get_config' RPC method.
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
 * Represents the response structure for the 'node.v1.get_assets' RPC method.
 */
export interface GetAssetsResponse extends GenericRPCMessage {
    method: RPCMethod.GetAssets;
    params: {
        /** List of assets available in the clearnode. */
        assets: RPCAsset[];
    };
}

/**
 * Represents the response structure for the 'user.v1.get_balances' RPC method.
 */
export interface GetBalancesResponse extends GenericRPCMessage {
    method: RPCMethod.GetBalances;
    params: {
        /** List of user balances */
        balances: RPCBalanceEntry[];
    };
}

/**
 * Represents the response structure for the 'user.v1.get_transactions' RPC method.
 */
export interface GetTransactionsResponse extends GenericRPCMessage {
    method: RPCMethod.GetTransactions;
    params: {
        /** List of transactions */
        transactions: RPCTransaction[];
        /** Pagination metadata */
        metadata?: PaginationMetadata;
    };
}

/**
 * Represents the response structure for the 'channels.v1.get_home_channel' RPC method.
 */
export interface GetHomeChannelResponse extends GenericRPCMessage {
    method: RPCMethod.GetHomeChannel;
    params: {
        /** Home channel information */
        channel: RPCChannel;
    };
}

/**
 * Represents the response structure for the 'channels.v1.get_escrow_channel' RPC method.
 */
export interface GetEscrowChannelResponse extends GenericRPCMessage {
    method: RPCMethod.GetEscrowChannel;
    params: {
        /** Escrow channel information */
        channel: RPCChannel;
    };
}

/**
 * Represents the response structure for the 'channels.v1.get_channels' RPC method.
 */
export interface GetChannelsResponse extends GenericRPCMessage {
    method: RPCMethod.GetChannels;
    params: {
        /** List of channels. */
        channels: RPCChannel[];
        /** Pagination metadata */
        metadata?: PaginationMetadata;
    };
}

/**
 * Represents the response structure for the 'channels.v1.get_latest_state' RPC method.
 */
export interface GetLatestStateResponse extends GenericRPCMessage {
    method: RPCMethod.GetLatestState;
    params: {
        /** Latest state for the user's asset */
        state: RPCState;
    };
}

/**
 * Represents the response structure for the 'channels.v1.get_states' RPC method.
 */
export interface GetStatesResponse extends GenericRPCMessage {
    method: RPCMethod.GetStates;
    params: {
        /** List of states */
        states: RPCState[];
        /** Pagination metadata */
        metadata?: PaginationMetadata;
    };
}

/**
 * Represents the response structure for the 'channels.v1.request_creation' RPC method.
 */
export interface CreateChannelResponse extends GenericRPCMessage {
    method: RPCMethod.CreateChannel;
    params: {
        /** Node's signature for the state */
        signature: Hex;
    };
}

/**
 * Represents the response structure for the 'channels.v1.submit_state' RPC method.
 */
export interface SubmitStateResponse extends GenericRPCMessage {
    method: RPCMethod.SubmitState;
    params: {
        /** Server's signature for the state */
        signature: Hex;
    };
}

/**
 * Represents the response structure for the 'app_sessions.v1.get_app_definition' RPC method.
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
 * Represents the response structure for the 'app_sessions.v1.get_app_sessions' RPC method.
 */
export interface GetAppSessionsResponse extends GenericRPCMessage {
    method: RPCMethod.GetAppSessions;
    params: {
        appSessions: RPCAppSession[];
    };
}

/**
 * Represents the response structure for the 'app_sessions.v1.create_app_session' RPC method.
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
 * Represents the response structure for the 'app_sessions.v1.submit_app_state' RPC method.
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
 * Represents the response structure for the 'app_sessions.v1.submit_deposit_state' RPC method.
 */
export interface SubmitDepositStateResponse extends GenericRPCMessage {
    method: RPCMethod.SubmitDepositState;
    params: {
        /** Server's signature for the deposit */
        signature: Hex;
    };
}

/**
 * Represents the response structure for the 'app_sessions.v1.rebalance_app_sessions' RPC method.
 */
export interface RebalanceAppSessionsResponse extends GenericRPCMessage {
    method: RPCMethod.RebalanceAppSessions;
    params: {
        /** Unique batch ID for the rebalance operation */
        batchId: string;
    };
}

/**
 * Represents the response structure for the 'session_keys.v1.register' RPC method.
 */
export interface RegisterResponse extends GenericRPCMessage {
    method: RPCMethod.Register;
    params: {
        /** Challenge message to sign for authorization */
        challengeMessage: string;
    };
}

/**
 * Represents the response structure for the 'session_keys.v1.get_session_keys' RPC method.
 */
export interface GetSessionKeysResponse extends GenericRPCMessage {
    method: RPCMethod.GetSessionKeys;
    params: {
        /** Array of active session keys for the authenticated user. */
        sessionKeys: RPCSessionKey[];
    };
}

/**
 * Represents the response structure for the 'session_keys.v1.revoke_session_key' RPC method.
 */
export interface RevokeSessionKeyResponse extends GenericRPCMessage {
    method: RPCMethod.RevokeSessionKey;
    params: {
        /** The session key address that was revoked. */
        sessionKey: Address;
    };
}

/**
 * Represents the response structure for the 'message' server push event.
 */
export interface MessageResponse extends GenericRPCMessage {
    method: RPCMethod.Message;
    params: {};
}

/**
 * Represents the response structure for the 'assets' server push event.
 */
export interface AssetsResponse extends GenericRPCMessage {
    method: RPCMethod.Assets;
    params: {
        /** List of assets available in the clearnode. */
        assets: RPCAsset[];
    };
}

/**
 * Represents the response structure for the 'bu' server push event.
 */
export interface BalanceUpdateResponse extends GenericRPCMessage {
    method: RPCMethod.BalanceUpdate;
    params: {
        /** List of balance updates. */
        balanceUpdates: RPCBalanceEntry[];
    };
}

/**
 * Represents the response structure for the 'channels' server push event.
 */
export interface ChannelsUpdateResponse extends GenericRPCMessage {
    method: RPCMethod.ChannelsUpdate;
    params: {
        /** List of channels. */
        channels: RPCChannel[];
    };
}

/**
 * Represents the response structure for the 'cu' server push event.
 */
export interface ChannelUpdateResponse extends GenericRPCMessage {
    method: RPCMethod.ChannelUpdate;
    params: RPCChannel;
}

/**
 * Represents the response structure for the 'tr' server push event.
 */
export interface TransferNotificationResponse extends GenericRPCMessage {
    method: RPCMethod.TransferNotification;
    params: {
        /** List of transactions representing transfers. */
        transactions: RPCTransaction[];
    };
}

/**
 * Server push notification for app session updates.
 */
export interface AppSessionUpdateResponse extends GenericRPCMessage {
    method: RPCMethod.AppSessionUpdate;
    params: {
        /** Updated app session information */
        appSession: RPCAppSession;
    };
}

/**
 * Union type for all possible RPC response types.
 * This allows for type-safe handling of different response structures.
 */
export type RPCResponse =
    // Node
    | ErrorResponse
    | GetConfigResponse
    | GetAssetsResponse
    | PingResponse
    // User
    | GetBalancesResponse
    | GetTransactionsResponse
    // Channels
    | GetChannelsResponse
    | GetHomeChannelResponse
    | GetEscrowChannelResponse
    | GetLatestStateResponse
    | GetStatesResponse
    | CreateChannelResponse
    | SubmitStateResponse
    // App Sessions
    | GetAppDefinitionResponse
    | GetAppSessionsResponse
    | CreateAppSessionResponse
    | SubmitAppStateResponse
    | SubmitDepositStateResponse
    | RebalanceAppSessionsResponse
    // Session Keys
    | RegisterResponse
    | GetSessionKeysResponse
    | RevokeSessionKeyResponse
    // Server Push
    | AssetsResponse
    | MessageResponse
    | BalanceUpdateResponse
    | ChannelsUpdateResponse
    | ChannelUpdateResponse
    | TransferNotificationResponse
    | AppSessionUpdateResponse;

/** Represents the parameters for the 'error' RPC method. */
export type ErrorResponseParams = ErrorResponse['params'];

/** Represents the parameters for the 'node.v1.ping' RPC method. */
export type PingResponseParams = PingResponse['params'];

/** Represents the parameters for the 'node.v1.get_config' RPC method. */
export type GetConfigResponseParams = GetConfigResponse['params'];

/** Represents the parameters for the 'node.v1.get_assets' RPC method. */
export type GetAssetsResponseParams = GetAssetsResponse['params'];

/** Represents the parameters for the 'user.v1.get_balances' RPC method. */
export type GetBalancesResponseParams = GetBalancesResponse['params'];

/** Represents the parameters for the 'user.v1.get_transactions' RPC method. */
export type GetTransactionsResponseParams = GetTransactionsResponse['params'];

/** Represents the parameters for the 'channels.v1.get_home_channel' RPC method. */
export type GetHomeChannelResponseParams = GetHomeChannelResponse['params'];

/** Represents the parameters for the 'channels.v1.get_escrow_channel' RPC method. */
export type GetEscrowChannelResponseParams = GetEscrowChannelResponse['params'];

/** Represents the parameters for the 'channels.v1.get_channels' RPC method. */
export type GetChannelsResponseParams = GetChannelsResponse['params'];

/** Represents the parameters for the 'channels.v1.get_latest_state' RPC method. */
export type GetLatestStateResponseParams = GetLatestStateResponse['params'];

/** Represents the parameters for the 'channels.v1.get_states' RPC method. */
export type GetStatesResponseParams = GetStatesResponse['params'];

/** Represents the parameters for the 'channels.v1.request_creation' RPC method. */
export type CreateChannelResponseParams = CreateChannelResponse['params'];

/** Represents the parameters for the 'channels.v1.submit_state' RPC method. */
export type SubmitStateResponseParams = SubmitStateResponse['params'];

/** Represents the parameters for the 'app_sessions.v1.get_app_definition' RPC method. */
export type GetAppDefinitionResponseParams = GetAppDefinitionResponse['params'];

/** Represents the parameters for the 'app_sessions.v1.get_app_sessions' RPC method. */
export type GetAppSessionsResponseParams = GetAppSessionsResponse['params'];

/** Represents the parameters for the 'app_sessions.v1.create_app_session' RPC method. */
export type CreateAppSessionResponseParams = CreateAppSessionResponse['params'];

/** Represents the parameters for the 'app_sessions.v1.submit_app_state' RPC method. */
export type SubmitAppStateResponseParams = SubmitAppStateResponse['params'];

/** Represents the parameters for the 'app_sessions.v1.submit_deposit_state' RPC method. */
export type SubmitDepositStateResponseParams = SubmitDepositStateResponse['params'];

/** Represents the parameters for the 'app_sessions.v1.rebalance_app_sessions' RPC method. */
export type RebalanceAppSessionsResponseParams = RebalanceAppSessionsResponse['params'];

/** Represents the parameters for the 'session_keys.v1.register' RPC method. */
export type RegisterResponseParams = RegisterResponse['params'];

/** Represents the parameters for the 'session_keys.v1.get_session_keys' RPC method. */
export type GetSessionKeysResponseParams = GetSessionKeysResponse['params'];

/** Represents the parameters for the 'session_keys.v1.revoke_session_key' RPC method. */
export type RevokeSessionKeyResponseParams = RevokeSessionKeyResponse['params'];

/** Represents the parameters for the 'assets' server push event. */
export type AssetsResponseParams = AssetsResponse['params'];

/** Represents the parameters for the 'message' server push event. */
export type MessageResponseParams = MessageResponse['params'];

/** Represents the parameters for the 'bu' server push event. */
export type BalanceUpdateResponseParams = BalanceUpdateResponse['params'];

/** Represents the parameters for the 'channels' server push event. */
export type ChannelsUpdateResponseParams = ChannelsUpdateResponse['params'];

/** Represents the parameters for the 'cu' server push event. */
export type ChannelUpdateResponseParams = ChannelUpdateResponse['params'];

/** Represents the parameters for the 'tr' server push event. */
export type TransferNotificationResponseParams = TransferNotificationResponse['params'];

/** Represents the parameters for the 'asu' server push event. */
export type AppSessionUpdateResponseParams = AppSessionUpdateResponse['params'];

/**
 * Helper type to extract the response type for a given method.
 */
export type ExtractResponseByMethod<M extends RPCMethod> = Extract<RPCResponse, { method: M }>;

/**
 * Type representing all response params as a union.
 */
export type RPCResponseParams = ExtractResponseByMethod<RPCMethod>['params'];

/**
 * Maps RPC methods to their corresponding response parameter types.
 */
export type RPCResponseParamsByMethod = {
    [M in RPCMethod]: ExtractResponseByMethod<M>['params'];
};
