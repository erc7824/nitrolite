import { Address, Hex } from 'viem';

export * from './request';
export * from './response';

/** Type alias for Request ID (uint64) */
export type RequestID = number;

/** Type alias for Timestamp (uint64) */
export type Timestamp = number;

/** Type alias for Account ID (channelId or appId) */
export type AccountID = Hex;

/** Represents the data payload within a request message: [requestId, method, params, timestamp?]. */
export type RequestData = [RequestID, RPCMethod, object[], Timestamp?];

/** Represents the data payload within a successful response message: [requestId, method, result, timestamp?]. */
export type ResponseData = [RequestID, RPCMethod, object[], Timestamp?];

/** Represents the status of a channel. */
export enum RPCChannelStatus {
    Joining = 'joining',
    Open = 'open',
    Closed = 'closed',
    Challenged = 'challenged',
}

/**
 * Represents a generic RPC message structure that includes common fields.
 * This interface is extended by specific RPC request and response types.
 */
export interface GenericRPCMessage {
    requestId: RequestID;
    timestamp?: Timestamp;
    signatures?: Hex[];
}

/** Base type for asset allocations with common asset and amount fields. */
export type AssetAllocation = {
    /** The symbol of the asset (e.g., "USDC", "USDT", "ETH"). */
    asset: string;
    /** The amount of the asset. Must be a positive number. */
    amount: string;
};

/**
 * Represents a generic RPC message structure that includes common fields.
 * This interface is extended by specific RPC request and response types.
 */
export type Allowance = {
    /** The symbol of the asset (e.g., "USDC", "USDT"). */
    asset: string;
    /** The amount of the asset that is allowed to be spent. */
    amount: string;
};

/** Represents the allocation of assets within an application session.
 * This structure is used to define the initial allocation of assets among participants.
 * It includes the participant's address, the asset (usdc, usdt, etc) being allocated, and the amount.
 */
export type AppSessionAllocation = AssetAllocation & {
    /** The Ethereum address of the participant receiving the allocation. */
    participant: Address;
};

/** Represents the allocation of assets for a transfer.
 * This structure is used to define the asset and amount being transferred.
 */
export type TransferAllocation = AssetAllocation;

/**
 * Represents the structure of an error object within an error response payload.
 */
export interface NitroliteRPCErrorDetail {
    /** The error message describing what went wrong. */
    error: string;
}

/** Represents the data payload for an error response: [requestId, "error", [errorDetail], timestamp?]. */
export type ErrorResponseData = [RequestID, 'error', [NitroliteRPCErrorDetail], Timestamp?];

/** Union type for the 'res' payload, covering both success and error responses. */
export type ResponsePayload = ResponseData | ErrorResponseData;

/**
 * Defines the wire format for Nitrolite RPC messages, based on NitroRPC principles
 * as adapted for the Clearnet protocol.
 * This is the structure used for WebSocket communication.
 */
export interface NitroliteRPCMessage {
    /** Contains the request payload if this is a request message. */
    req?: RequestData;
    /** Contains the response or error payload if this is a response message. */
    res?: ResponsePayload;
    /** Optional cryptographic signature(s) for message authentication. */
    sig?: Hex[] | [''];
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
 * Represents the result of parsing an incoming Nitrolite RPC response message.
 * Contains extracted fields and validation status.
 */
export interface ParsedResponse {
    /** Indicates if the message was successfully parsed and passed basic structural validation. */
    isValid: boolean;
    /** If isValid is false, contains a description of the parsing or validation error. */
    error?: string;
    /** Indicates if the parsed response represents an error (method === "error"). Undefined if isValid is false. */
    isError?: boolean;
    /** The Request ID from the response payload. Undefined if structure is invalid. */
    requestId?: RequestID;
    /** The method name from the response payload. Undefined if structure is invalid. */
    method?: RPCMethod;
    /** The extracted data payload (result array for success, error detail object for error). Undefined if structure is invalid or error payload malformed. */
    data?: object[] | NitroliteRPCErrorDetail;
    /** The Application Session ID from the message envelope. Undefined if structure is invalid. */
    sid?: Hex;
    /** The Timestamp from the response payload. Undefined if structure is invalid. */
    timestamp?: Timestamp;
}

/**
 * Defines the structure of an application definition used when creating an application.
 */
export interface AppDefinition {
    /** The protocol identifier or name for the application logic (e.g., "NitroRPC/0.2"). */
    protocol: string;
    /** An array of participant addresses (Ethereum addresses) involved in the application. Must have at least 2 participants. */
    participants: Hex[];
    /** An array representing the relative weights or stakes of participants, often used for dispute resolution or allocation calculations. Order corresponds to the participants array. */
    weights: number[];
    /** The number of participants required to reach consensus or approve state updates. */
    quorum: number;
    /** A parameter related to the challenge period or mechanism within the application's protocol, in seconds. */
    challenge: number;
    /** A unique number used once, often for preventing replay attacks or ensuring uniqueness of the application instance. Must be non-zero. */
    nonce?: number;
}

/**
 * Represents a channel update message sent over the RPC protocol.
 */
export interface ChannelUpdate {
    /** The unique identifier for the channel. */
    channelId: Hex;
    /** The Ethereum address of the participant. */
    participant: Address;
    /** The current status of the channel (e.g., "open", "closed"). */
    status: RPCChannelStatus;
    /** The token contract address. */
    token: Address;
    /** The wallet address associated with the channel. */
    wallet: Address;
    /** The total amount in the channel. */
    amount: BigInt;
    /** The chain ID where the channel exists. */
    chainId: number;
    /** The adjudicator contract address. */
    adjudicator: Address;
    /** The challenge period in seconds. */
    challenge: number;
    /** The nonce value for the channel. */
    nonce: BigInt;
    /** The version number of the channel. */
    version: number;
    /** The timestamp when the channel was created. */
    createdAt: Date;
    /** The timestamp when the channel was last updated. */
    updatedAt: Date;
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
export type MessageSigner = (payload: RequestData | ResponsePayload) => Promise<Hex>;

/**
 * Defines the function signature for signing challenge state data.
 * This signer is specifically used for signing state challenges in the form of keccak256(abi.encode(stateHash, 'challenge')).
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
export type SingleMessageVerifier = (
    payload: RequestData | ResponsePayload,
    signature: Hex,
    address: Address,
) => Promise<boolean>;

/**
 * Defines the function signature for verifying multiple message signatures against a payload.
 * This is used for operations requiring consensus from multiple parties (e.g., closing an application).
 * @param payload - The RequestData or ResponsePayload object (array) that was signed.
 * @param signatures - An array of signature strings (Hex) to verify.
 * @param expectedSigners - An array of Ethereum addresses of the required signers. The implementation determines if order matters.
 * @returns A Promise that resolves to true if all required signatures from the expected signers are present and valid, false otherwise.
 */
export type MultiMessageVerifier = (
    payload: RequestData | ResponsePayload,
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
    application: Address;
    participant: Address;
    expire: string;
    // TODO: use Allowance type after replacing symbol with asset
    allowances: {
        asset: string;
        amount: string;
    }[];
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
        { name: 'application', type: 'address' },
        { name: 'participant', type: 'address' },
        { name: 'expire', type: 'uint256' },
        { name: 'allowances', type: 'Allowance[]' },
    ],
    Allowance: [
        { name: 'asset', type: 'string' },
        { name: 'amount', type: 'uint256' },
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
    CreateAppSession = 'create_app_session',
    SubmitAppState = 'submit_app_state',
    CloseAppSession = 'close_app_session',
    GetAppDefinition = 'get_app_definition',
    GetAppSessions = 'get_app_sessions',
    ResizeChannel = 'resize_channel',
    CloseChannel = 'close_channel',
    GetChannels = 'get_channels',
    GetRPCHistory = 'get_rpc_history',
    GetAssets = 'get_assets',
    Assets = 'assets',
    Message = 'message',
    BalanceUpdate = 'bu',
    ChannelsUpdate = 'channels',
    ChannelUpdate = 'cu',
    Ping = 'ping',
    Pong = 'pong',
    Transfer = 'transfer',
    TransferNotification = 'tr',
}
