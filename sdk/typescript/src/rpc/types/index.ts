import { Address, Hex } from 'viem';

export * from './request';
export * from './response';
export * from './common';

/**
 * Message type identifier for RPC protocol.
 * Indicates whether the message is a request, response, event, or error.
 */
export enum RPCMessageType {
    /** Request message (client to server) */
    Request = 1,
    /** Response message (server to client, success) */
    Response = 2,
    /** Event message (server-initiated notification) */
    Event = 3,
    /** Error response message (server to client, error) */
    ErrorResponse = 4,
}

/** Type alias for Request ID (uint64) */
export type RequestID = number;

/** Type alias for Timestamp (uint64 in milliseconds) */
export type Timestamp = number;

/** Type alias for Account ID (channelId or appId) */
export type AccountID = Hex;

/**
 * Represents the wire format for RPC messages as transmitted over WebSocket.
 * Format: [type, requestId, method, params, timestamp]
 *
 * Example request: [1, 100, "node.v1.ping", {}, 1770559919268]
 * Example response: [2, 100, "node.v1.ping", {}, 1770559919300]
 */
export type RPCMessage = [RPCMessageType, RequestID, string, Record<string, unknown>, Timestamp];

/**
 * Represents a generic RPC message structure that includes common fields.
 * This interface is extended by specific RPC request and response types.
 */
export interface GenericRPCMessage {
    requestId?: RequestID;
    timestamp?: Timestamp;
}

/**
 * Defines standard error codes for the Nitrolite RPC protocol.
 * Includes standard JSON-RPC codes and custom codes for specific errors.
 */
export enum NitroliteErrorCode {
    PARSE_ERROR = -32700,
    INVALID_REQUEST = -32600,
    METHOD_NOT_FOUND = -32601,
    INVALID_PARAMS = -32602,
    INTERNAL_ERROR = -32603,
    AUTHENTICATION_FAILED = -32000,
    INVALID_SIGNATURE = -32003,
    INVALID_TIMESTAMP = -32004,
    INVALID_REQUEST_ID = -32005,
    INSUFFICIENT_FUNDS = -32007,
    ACCOUNT_NOT_FOUND = -32008,
    APPLICATION_NOT_FOUND = -32009,
    INVALID_INTENT = -32010,
    INSUFFICIENT_SIGNATURES = -32006,
    CHALLENGE_EXPIRED = -32011,
    INVALID_CHALLENGE = -32012,
}

/**
 * Defines the function signature for signing state data.
 * Used for signing states that get embedded in request params (e.g., user_sig field).
 *
 * Example implementation:
 * - Using signMessage: (data) => walletClient.signMessage({ message: JSON.stringify(data) })
 * - Using signTypedData: (data) => walletClient.signTypedData({ domain, types, message: data })
 *
 * @param data - The data object to sign (e.g., RPCState)
 * @returns A Promise that resolves to the cryptographic signature as a Hex string.
 */
export type StateSigner = (data: unknown) => Promise<Hex>;

/**
 * Defines the function signature for signing challenge state data.
 * This signer is specifically used for signing state challenges in the form of keccak256(abi.encodePacked(packedState, 'challenge')).
 *
 * @param stateHash - The state hash as a Hex string
 * @returns A Promise that resolves to the cryptographic signature as a Hex string.
 */
export type ChallengeStateSigner = (stateHash: Hex) => Promise<Hex>;

/**
 * Defines the function signature for verifying a signature against state data.
 * @param data - The state data that was signed
 * @param signature - The signature (Hex string) to verify
 * @param address - The Ethereum address of the expected signer
 * @returns A Promise that resolves to true if the signature is valid, false otherwise.
 */
export type StateVerifier = (data: unknown, signature: Hex, address: Address) => Promise<boolean>;

/**
 * Defines the function signature for verifying multiple signatures against state data.
 * This is used for operations requiring consensus from multiple parties.
 * @param data - The state data that was signed
 * @param signatures - An array of signature strings (Hex) to verify
 * @param expectedSigners - An array of Ethereum addresses of the required signers
 * @returns A Promise that resolves to true if all required signatures are present and valid, false otherwise.
 */
export type MultiStateVerifier = (data: unknown, signatures: Hex[], expectedSigners: Address[]) => Promise<boolean>;

/**
 * RPC methods supported by the Clearnode API.
 *
 * @see {@link https://github.com/erc7824/nitrolite/blob/main/docs/api.yaml API Specification}
 */
export enum RPCMethod {
    /** Error response from the server */
    Error = 'error',
    /** Health check to verify connection is alive */
    Ping = 'node.v1.ping',
    /** Get node configuration and supported blockchains */
    GetConfig = 'node.v1.get_config',
    /** Get list of supported assets, optionally filtered by blockchain */
    GetAssets = 'node.v1.get_assets',

    // User Group (user.v1)
    /** Get user's asset balances */
    GetBalances = 'user.v1.get_balances',
    /** Get user's transaction history with optional filtering and pagination */
    GetTransactions = 'user.v1.get_transactions',

    // Channels Group (channels.v1)
    /** Get list of channels with optional filtering and pagination */
    GetChannels = 'channels.v1.get_channels',
    /** Get home channel information for a specific wallet and asset */
    GetHomeChannel = 'channels.v1.get_home_channel',
    /** Get escrow channel information by channel ID */
    GetEscrowChannel = 'channels.v1.get_escrow_channel',
    /** Get the latest state for a user's asset */
    GetLatestState = 'channels.v1.get_latest_state',
    /** Get state history with optional filtering and pagination */
    GetStates = 'channels.v1.get_states',
    /** Create a new channel */
    CreateChannel = 'channels.v1.request_creation',
    /** Submit a cross-chain state transition */
    SubmitState = 'channels.v1.submit_state',

    // App Sessions Group (app_sessions.v1)
    /** Get application definition for a session */
    GetAppDefinition = 'app_sessions.v1.get_app_definition',
    /** Get list of application sessions with optional filtering and pagination */
    GetAppSessions = 'app_sessions.v1.get_app_sessions',
    /** Create a new application session */
    CreateAppSession = 'app_sessions.v1.create_app_session',
    /** Submit an application session state update */
    SubmitAppState = 'app_sessions.v1.submit_app_state',
    /** Submit an application session deposit state update */
    SubmitDepositState = 'app_sessions.v1.submit_deposit_state',
    /** Atomically rebalance multiple application sessions */
    RebalanceAppSessions = 'app_sessions.v1.rebalance_app_sessions',

    // Session Keys Group (session_keys.v1)
    /** Register a new session key with allowances and expiration */
    Register = 'session_keys.v1.register',
    /** Get all active session keys for a user */
    GetSessionKeys = 'session_keys.v1.get_session_keys',
    /** Revoke an existing session key */
    RevokeSessionKey = 'session_keys.v1.revoke_session_key',

    // Server Push Events (no group prefix for events)
    /** Server-initiated asset list update */
    Assets = 'assets',
    /** Application-scoped message */
    Message = 'message',
    /** Balance change notification */
    BalanceUpdate = 'bu',
    /** Channel list update notification */
    ChannelsUpdate = 'channels',
    /** Single channel update notification */
    ChannelUpdate = 'cu',
    /** Transfer notification */
    TransferNotification = 'tr',
    /** Application session update notification */
    AppSessionUpdate = 'asu',
}
