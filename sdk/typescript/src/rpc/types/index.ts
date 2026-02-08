import { Address, Hex } from 'viem';
import { RPCAllowance } from './common';

export * from './request';
export * from './response';
export * from './filters';
export * from './common';

/** Type alias for Request ID (uint64) */
export type RequestID = number;

/** Type alias for Timestamp (uint64) */
export type Timestamp = number;

/** Type alias for Account ID (channelId or appId) */
export type AccountID = Hex;

/** Represents the data payload within a request or response message: [requestId, method, params, timestamp?]. */
export type RPCData = [RequestID, RPCMethod, object, Timestamp?];

/**
 * Represents a generic RPC message structure that includes common fields.
 * This interface is extended by specific RPC request and response types.
 */
export interface GenericRPCMessage {
    requestId?: RequestID;
    timestamp?: Timestamp;
    signatures?: Hex[];
}

/**
 * Defines the wire format for Nitrolite RPC messages, based on NitroRPC principles
 * as adapted for the Clearnet protocol.
 * This is the structure used for WebSocket communication.
 */
export interface NitroliteRPCMessage {
    /** Contains the request payload if this is a request message. */
    req?: RPCData;
    /** Contains the response or error payload if this is a response message. */
    res?: RPCData;
    /** Optional cryptographic signature(s) for message authentication. */
    sig?: Hex[];
}

/**
 * Defines the wire format for Nitrolite RPC messages sent within the context
 * of a specific application.
 */
export interface ApplicationRPCMessage extends NitroliteRPCMessage {
    /**
     * Application Session ID. Mandatory.
     * This field also serves as the destination pubsub topic for the message.
     */
    sid: Hex;
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
 * Defines the function signature for signing message payloads (req or res objects).
 * Implementations can use either signMessage or signStateData depending on the use case.
 * For general RPC messages, signMessage is typically used.
 * For state channel operations, signStateData may be more appropriate.
 *
 * Example implementations:
 * - Using signMessage: (payload) => walletClient.signMessage({ message: JSON.stringify(payload) })
 * - Using signStateData: (payload) => walletClient.signStateData({ data: encodeAbiParameters([...], payload) })
 *
 * @param payload - The RequestData or ResponsePayload object (array) to sign.
 * @returns A Promise that resolves to the cryptographic signature as a Hex string.
 */
export type MessageSigner = (payload: RPCData) => Promise<Hex>;

/**
 * Defines the function signature for signing challenge state data.
 * This signer is specifically used for signing state challenges in the form of keccak256(abi.encodePacked(packedState, 'challenge')).
 *
 * @param stateHash - The state hash as a Hex string
 * @returns A Promise that resolves to the cryptographic signature as a Hex string.
 */
export type ChallengeStateSigner = (stateHash: Hex) => Promise<Hex>;

/**
 * Defines the function signature for verifying a single message signature against its payload.
 * @param payload - The RequestData or ResponsePayload object (array) that was signed.
 * @param signature - The single signature (Hex string) to verify.
 * @param address - The Ethereum address of the expected signer.
 * @returns A Promise that resolves to true if the signature is valid for the given payload and address, false otherwise.
 */
export type SingleMessageVerifier = (payload: RPCData, signature: Hex, address: Address) => Promise<boolean>;

/**
 * Defines the function signature for verifying multiple message signatures against a payload.
 * This is used for operations requiring consensus from multiple parties (e.g., closing an application).
 * @param payload - The RequestData or ResponsePayload object (array) that was signed.
 * @param signatures - An array of signature strings (Hex) to verify.
 * @param expectedSigners - An array of Ethereum addresses of the required signers. The implementation determines if order matters.
 * @returns A Promise that resolves to true if all required signatures from the expected signers are present and valid, false otherwise.
 */
export type MultiMessageVerifier = (
    payload: RPCData,
    signatures: Hex[],
    expectedSigners: Address[],
) => Promise<boolean>;

/**
 * Represents a partial EIP-712 message for authorization.
 * This is used to define the structure of the authorization message
 * that will be signed by the user.
 */
export interface PartialEIP712AuthMessage {
    scope: string;
    session_key: Address;
    expires_at: bigint;
    allowances: RPCAllowance[];
}

/**
 * Represents a complete EIP-712 message for authorization.
 */
export interface EIP712AuthMessage extends PartialEIP712AuthMessage {
    wallet: Address;
    challenge: string;
}

/**
 * Represents the EIP-712 domain for authorization messages.
 * This is used to define the domain separator for EIP-712 signatures.
 */
export interface EIP712AuthDomain {
    name: string;
}

/**
 * Represents the EIP-712 types for authorization messages.
 */
export const EIP712AuthTypes = {
    Policy: [
        { name: 'challenge', type: 'string' },
        { name: 'scope', type: 'string' },
        { name: 'wallet', type: 'address' },
        { name: 'session_key', type: 'address' },
        { name: 'expires_at', type: 'uint64' },
        { name: 'allowances', type: 'Allowance[]' },
    ],
    Allowance: [
        { name: 'asset', type: 'string' },
        { name: 'amount', type: 'string' },
    ],
};

/**
 * Represents the RPC methods used in the Nitrolite protocol.
 */
export enum RPCMethod {
    AuthRequest = 'auth_request',
    AuthChallenge = 'auth_challenge',
    AuthVerify = 'auth_verify',
    Error = 'error',
    GetConfig = 'get_config',
    GetLedgerBalances = 'get_ledger_balances',
    GetLedgerEntries = 'get_ledger_entries',
    GetLedgerTransactions = 'get_ledger_transactions',
    GetUserTag = 'get_user_tag',
    GetSessionKeys = 'get_session_keys',
    RevokeSessionKey = 'revoke_session_key',
    CreateAppSession = 'create_app_session',
    SubmitAppState = 'submit_app_state',
    CloseAppSession = 'close_app_session',
    GetAppDefinition = 'get_app_definition',
    GetAppSessions = 'get_app_sessions',
    CreateChannel = 'create_channel',
    ResizeChannel = 'resize_channel',
    CloseChannel = 'close_channel',
    GetChannels = 'get_channels',
    GetRPCHistory = 'get_rpc_history',
    GetAssets = 'get_assets',
    CleanupSessionKeyCache = 'cleanup_session_key_cache',

    Assets = 'assets',
    Message = 'message',
    BalanceUpdate = 'bu',
    ChannelsUpdate = 'channels',
    ChannelUpdate = 'cu',
    Ping = 'ping',
    Pong = 'pong',
    Transfer = 'transfer',
    TransferNotification = 'tr',
    AppSessionUpdate = 'asu',
}
