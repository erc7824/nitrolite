import { Address, Hex } from "viem";

/** Type alias for Request ID (uint64) */
export type RequestID = number;

/** Type alias for Timestamp (uint64) */
export type Timestamp = number;

/** Type alias for Account ID (channelId or appId) */
export type AccountID = Hex;

/** Represents the data payload within a request message: [requestId, method, params, timestamp?]. */
export type RequestData = [RequestID, string, any[], Timestamp?];

/** Represents the data payload within a successful response message: [requestId, method, result, timestamp?]. */
export type ResponseData = [RequestID, string, any[], Timestamp?];

/** Represents the allocation of assets within an application session.
 * This structure is used to define the initial allocation of assets among participants.
 * It includes the participant's address, the asset (usdc, usdt, etc) being allocated, and the amount.
 */
export type AppSessionAllocation = {
    participant: Address;
    asset: string;
    amount: string;
};

/**
 * Represents the structure of an error object within an error response payload.
 */
export interface NitroliteRPCErrorDetail {
    error: string;
}

/** Represents the data payload for an error response: [requestId, "error", [errorDetail], timestamp?]. */
export type ErrorResponseData = [RequestID, "error", [NitroliteRPCErrorDetail], Timestamp?];

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
    /** Optional cryptographic signature(s). */
    sig?: Hex[] | [""];
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
    // Removed originalMessage field
    /** Indicates if the parsed response represents an error (method === "error"). Undefined if isValid is false. */
    isError?: boolean;
    /** The Request ID from the response payload. Undefined if structure is invalid. */
    requestId?: RequestID;
    /** The method name from the response payload. Undefined if structure is invalid. */
    method?: string;
    /** The extracted data payload (result array for success, error detail object for error). Undefined if structure is invalid or error payload malformed. */
    data?: any[] | NitroliteRPCErrorDetail;
    /** The Application Session ID from the message envelope. Undefined if structure is invalid. */
    sid?: Hex;
    /** The Timestamp from the response payload. Undefined if structure is invalid. */
    timestamp?: Timestamp;
}

/**
 * Defines the structure of an application definition used when creating an application.
 */
export interface AppDefinition {
    /** The protocol identifier or name for the application logic. */
    protocol: string;
    /** An array of participant addresses (Ethereum addresses) involved in the application. */
    participants: Hex[];
    /** An array representing the relative weights or stakes of participants, often used for dispute resolution or allocation calculations. Order corresponds to the participants array. */
    weights: number[];
    /** The number of participants required to reach consensus or approve state updates. */
    quorum: number;
    /** A parameter related to the challenge period or mechanism within the application's protocol. */
    challenge: number;
    /** A unique number used once, often for preventing replay attacks or ensuring uniqueness of the application instance. */
    nonce?: number;
}

/**
 * Defines the parameters required for the 'auth_request' RPC method.
 */
export interface AuthRequest {
    /** The Address of the connected account. */
    address: Address;
    // Session key for the account, used for application authentication.
    session_key: Address;
    /** The name of the application associated with the account, used for application-specific operations. */
    app_name: string;
    /** Allowances for token to spend: symbol->amount */
    allowances?: Record<string, string>;
}

/**
 * Defines the parameters required for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRequest {
    /** The Address of the connected account. */
    address: Address;
    // The challenge string that needs to be signed by the user.
    challenge: string;
    // Session key for the account, used for application authentication.
    session_key: Address;
    /** The name of the application associated with the account, used for application-specific operations. */
    app_name: string;
}

/**
 * Defines the parameters required for the 'create_application' RPC method.
 */
export interface CreateAppSessionRequest {
    /** The detailed definition of the application being created. 
     * Example:
     * {
        "protocol": "NitroRPC/0.2",
        "participants": [
            "0xAaBbCcDdEeFf0011223344556677889900aAbBcC",
            "0x00112233445566778899AaBbCcDdEeFf00112233"
        ],
        "weights": [100, 0],
        "quorum": 100,
        "challenge": 86400,
        "nonce": 1
        }
    */
    definition: AppDefinition;
    /** The initial allocation distribution among participants. Order corresponds to the participants array in the definition. */
    allocations: AppSessionAllocation[];
}

/**
 * Defines the parameters required for the 'close_application' RPC method.
 */
export interface CloseAppSessionRequest {
    /** The unique identifier of the application session to be closed. */
    app_session_id: Hex;
    /** The final allocation distribution among participants upon closing the application. Order corresponds to the participants array in the application's definition. */
    allocations: AppSessionAllocation[];
}

/**
 * Defines the parameters required for the 'update_allocation' RPC method.
 */
export interface ResizeChannel {
    channel_id: Hex; // The unique identifier of the channel to be resized.
    resize_amount?: bigint; // How much user wants to deposit or withdraw from a token-network specific channel.
    allocate_amount?: bigint; // How much more token user wants to allocate to this token-network specific channel from his unified balance.
    funds_destination: Hex; // The address where the resized funds will be sent.
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
 * Defines the function signature for verifying a single message signature against its payload.
 * @param payload - The RequestData or ResponsePayload object (array) that was signed.
 * @param signature - The single signature (Hex string) to verify.
 * @param address - The Ethereum address of the expected signer.
 * @returns A Promise that resolves to true if the signature is valid for the given payload and address, false otherwise.
 */
export type SingleMessageVerifier = (payload: RequestData | ResponsePayload, signature: Hex, address: Address) => Promise<boolean>;

/**
 * Defines the function signature for verifying multiple message signatures against a payload.
 * This is used for operations requiring consensus from multiple parties (e.g., closing an application).
 * @param payload - The RequestData or ResponsePayload object (array) that was signed.
 * @param signatures - An array of signature strings (Hex) to verify.
 * @param expectedSigners - An array of Ethereum addresses of the required signers. The implementation determines if order matters.
 * @returns A Promise that resolves to true if all required signatures from the expected signers are present and valid, false otherwise.
 */
export type MultiMessageVerifier = (payload: RequestData | ResponsePayload, signatures: Hex[], expectedSigners: Address[]) => Promise<boolean>;
