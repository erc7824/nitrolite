import { Address, Hex } from 'viem';

/** Type alias for Request ID (uint64) */
export type RequestID = number;

/** Type alias for Timestamp (uint64) */
export type Timestamp = number;

/** Type alias for Account ID (channelId or appId) */
export type AccountID = Hex;

/** Represents the data payload within a request message: [requestId, method, params, timestamp?]. */
export type RequestData = [RequestID, RPCMethod, any[], Timestamp?];

/** Represents the data payload within a successful response message: [requestId, method, result, timestamp?]. */
export type ResponseData = [RequestID, RPCMethod, any[], Timestamp?];

/** Represents the status of a channel. */
export enum ChannelStatus {
    Joining = 'joining',
    Open = 'open',
    Closed = 'closed',
};

/** Represents a single allowance for an asset, used in application sessions.
 * This structure defines the symbol of the asset and the amount that is allowed to be spent.
 */
export type Allowance = {
    /** The symbol of the asset (e.g., "USDC", "USDT"). */
    symbol: string;
    /** The amount of the asset that is allowed to be spent. */
    amount: string;
};

/** Represents the allocation of assets within an application session.
 * This structure is used to define the initial allocation of assets among participants.
 * It includes the participant's address, the asset (usdc, usdt, etc) being allocated, and the amount.
 */
export type AppSessionAllocation = {
    /** The Ethereum address of the participant receiving the allocation. */
    participant: Address;
    /** The symbol of the asset being allocated (e.g., "USDC", "USDT"). */
    asset: string;
    /** The amount of the asset being allocated. Must be a positive number. */
    amount: string;
};

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
 * Defines the parameters required for the 'auth_request' RPC method.
 */
export interface AuthRequest {
    /** The Ethereum address of the wallet being authorized. */
    wallet: Address;
    /** The public address of the application that is being authorized. */
    participant: Address;
    /** The scope of the authorization, defining what permissions are granted (e.g., "app.create", "ledger.readonly"). */
    scope?: string;
    /** The name of the application being authorized. */
    app_name: string;
    /** The public address of the application that is being authorized. */
    application?: Address;
    /** The expiration timestamp for the authorization, in seconds since the Unix epoch. */
    expire?: string;
    /** An array of allowances, each defining an asset and the amount that can be spent. */
    allowances: Allowance[];
}

/**
 * Defines the parameters required for the 'create_app_session' RPC method.
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
 * Defines the parameters required for the 'close_app_session' RPC method.
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
    AuthChallenge = 'auth_challenge',
    AuthVerify = 'auth_verify',
    Error = 'error',
    GetConfig = 'get_config',
    GetLedgerBalances = 'get_ledger_balances',
    GetLedgerEntries = 'get_ledger_entries',
    CreateAppSession = 'create_app_session',
    SubmitState = 'submit_state',
    CloseAppSession = 'close_app_session',
    GetAppDefinition = 'get_app_definition',
    GetAppSessions = 'get_app_sessions',
    ResizeChannel = 'resize_channel',
    CloseChannel = 'close_channel',
    GetChannels = 'get_channels',
    GetRPCHistory = 'get_rpc_history',
    GetAssets = 'get_assets',
    Ping = 'ping',
    Message = 'message'
}

/**
 * Represents the parameters for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeRPCParams {
    challengeMessage: string;
}

/**
 * Represents the parameters for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRPCParams {
    address: Address;
    jwtToken: string;
    sessionKey: Address;
    success: boolean;
}

/**
 * Represents the parameters for the 'error' RPC method.
 */
export interface ErrorRPCParams {
    error: string;
}

/**
 * Represents the parameters for the 'get_config' RPC method.
 */
export interface GetConfigRPCParams {
    broker_address: Address;
    networks: {
        name: string;
        chain_id: number;
        custody_address: Address;
        adjudicator_address: Address;
    }[];
}

/**
 * Represents the parameters for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRPCParams {
    asset: string;
    amount: string;
}

/**
 * Represents the parameters for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesRPCParams {
    id: number;
    account_id: string;
    account_type: string;
    asset: string;
    participant: Address;
    credit: string;
    debit: string;
    created_at: string;
}

/**
 * Represents the parameters for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRPCParams {
    app_session_id: Hex;
    version: number;
    status: ChannelStatus;
}

/**
 * Represents the parameters for the 'submit_state' RPC method.
 */
export interface SubmitStateRPCParams {
    app_session_id: Hex;
    version: number;
    status: ChannelStatus;
}

/**
 * Represents the parameters for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRPCParams {
    app_session_id: Hex;
    version: number;
    status: ChannelStatus;
}

/**
 * Represents the parameters for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRPCParams extends AppDefinition { }

/**
 * Represents the parameters for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRPCParams {
    app_session_id: Hex;
    status: ChannelStatus;
    participants: Address[];
    protocol: string;
    challenge: number;
    weights: number[];
    quorum: number;
    version: number;
    nonce: number;
    created_at: string;
    updated_at: string;
}

/**
 * Represents the parameters for the 'resize_channel' RPC method.
 */
export interface ResizeChannelRPCParams {
    channel_id: Hex;
    state_data: string;
    intent: number;
    version: number;
    allocations: {
        destination: Address;
        token: Address;
        amount: string;
    }[];
    state_hash: string;
    server_signature: {
        v: string;
        r: string;
        s: string;
    };
}

/**
 * Represents the parameters for the 'close_channel' RPC method.
 */
export interface CloseChannelRPCParams {
    channel_id: Hex;
    intent: number;
    version: number;
    state_data: string;
    allocations: {
        destination: Address;
        token: Address;
        amount: string;
    }[];
    state_hash: string;
    server_signature: {
        v: string;
        r: string;
        s: string;
    };
}

/**
 * Represents the parameters for the 'get_channels' RPC method.
 */
export interface GetChannelsRPCParams {
    channel_id: Hex;
    participant: Address;
    status: ChannelStatus;
    token: Address;
    wallet: Address;
    amount: string;
    chain_id: number;
    adjudicator: Address;
    challenge: number;
    nonce: number;
    version: number;
    created_at: string;
    updated_at: string;
}

/**
 * Represents the parameters for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRPCParams {
    id: number;
    sender: Address;
    req_id: number;
    method: string;
    params: string;
    timestamp: number;
    req_sig: Hex[];
    res_sig: Hex[];
    response: string;
}

/**
 * Represents the parameters for the 'get_assets' RPC method.
 */
export interface GetAssetsRPCParams {
    token: Address;
    chain_id: number;
    symbol: string;
    decimals: number;
}

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
 * Represents the request structure for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeRPCRequest extends GenericRPCMessage {
    method: RPCMethod.AuthChallenge;
    params: AuthChallengeRPCRequestParams;
}

/**
 * Represents the response structure for the 'auth_challenge'
 */
export interface AuthChallengeRPCResponse extends GenericRPCMessage {
    method: RPCMethod.AuthChallenge;
    params: AuthChallengeRPCParams;
}

/**
 * Represents the request parameters for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeRPCRequestParams {
    /** The challenge message to be signed by the client for authentication. */
    challengeMessage: string;
}

/**
 * Represents the request parameters for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRPCRequestParams {
    /** The Ethereum address of the client attempting to authenticate. */
    address: Address;
    /** JSON Web Token for authentication, if provided. */
    jwtToken: string;
    /** The session key address associated with the authentication attempt. */
    sessionKey: Address;
    /** Indicates whether the authentication attempt was successful. */
    success: boolean;
}

/**
 * Represents the request structure for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRPCRequest extends GenericRPCMessage {
    method: RPCMethod.AuthVerify;
    params: AuthVerifyRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_config' RPC method.
 */
export interface GetConfigRPCRequestParams {
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
export interface GetConfigRPCRequest extends GenericRPCMessage {
    method: RPCMethod.GetConfig;
    params: GetConfigRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRPCRequestParams {
    /** Optional participant address to filter balances. If not provided, uses the authenticated wallet address. */
    participant?: Address;
    /** Optional account ID to filter balances. If provided, overrides the participant address. */
    account_id?: string;
}

/**
 * Represents the request structure for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRPCRequest extends GenericRPCMessage {
    method: RPCMethod.GetLedgerBalances;
    params: GetLedgerBalancesRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesRPCRequestParams {
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
export interface GetLedgerEntriesRPCRequest extends GenericRPCMessage {
    method: RPCMethod.GetLedgerEntries;
    params: GetLedgerEntriesRPCRequestParams;
}

/**
 * Represents the request parameters for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRPCRequestParams {
    /** The detailed definition of the application being created, including protocol, participants, weights, and quorum. */
    definition: AppDefinition;
    /** The initial allocation distribution among participants. Each participant must have sufficient balance for their allocation. */
    allocations: AppSessionAllocation[];
}

/**
 * Represents the request structure for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRPCRequest extends GenericRPCMessage {
    method: RPCMethod.CreateAppSession;
    params: CreateAppSessionRPCRequestParams;
}

/**
 * Represents the request parameters for the 'submit_state' RPC method.
 */
export interface SubmitStateRPCRequestParams {
    /** The unique identifier of the application session to update. */
    app_session_id: Hex;
    /** The new allocation distribution among participants. Must include all participants and maintain total balance. */
    allocations: AppSessionAllocation[];
}

/**
 * Represents the request structure for the 'submit_state' RPC method.
 */
export interface SubmitStateRPCRequest extends GenericRPCMessage {
    method: RPCMethod.SubmitState;
    params: SubmitStateRPCRequestParams;
}

/**
 * Represents the request parameters for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRPCRequestParams {
    /** The unique identifier of the application session to close. */
    app_session_id: Hex;
    /** The final allocation distribution among participants upon closing. Must include all participants and maintain total balance. */
    allocations: AppSessionAllocation[];
}

/**
 * Represents the request structure for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRPCRequest extends GenericRPCMessage {
    method: RPCMethod.CloseAppSession;
    params: CloseAppSessionRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRPCRequestParams {
    /** The unique identifier of the application session to retrieve. */
    app_session_id: Hex;
}

/**
 * Represents the request structure for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRPCRequest extends GenericRPCMessage {
    method: RPCMethod.GetAppDefinition;
    params: GetAppDefinitionRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRPCRequestParams {
    /** Optional participant address to filter application sessions. If not provided, returns all sessions. */
    participant?: Address;
    /** Optional status to filter application sessions (e.g., "open", "closed"). If not provided, returns sessions of all statuses. */
    status?: ChannelStatus;
}

/**
 * Represents the request structure for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRPCRequest extends GenericRPCMessage {
    method: RPCMethod.GetAppSessions;
    params: GetAppSessionsRPCRequestParams;
}

/**
 * Represents the request parameters for the 'resize_channel' RPC method.
 */
export interface ResizeChannelRPCRequestParams {
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
export interface ResizeChannelRPCRequest extends GenericRPCMessage {
    method: RPCMethod.ResizeChannel;
    params: ResizeChannelRPCRequestParams;
}

/**
 * Represents the request parameters for the 'close_channel' RPC method.
 */
export interface CloseChannelRPCRequestParams {
    /** The unique identifier of the channel to close. */
    channel_id: Hex;
    /** The address where the channel funds will be sent upon closing. */
    funds_destination: Address;
}

/**
 * Represents the request structure for the 'close_channel' RPC method.
 */
export interface CloseChannelRPCRequest extends GenericRPCMessage {
    method: RPCMethod.CloseChannel;
    params: CloseChannelRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_channels' RPC method.
 */
export interface GetChannelsRPCRequestParams {
    /** Optional participant address to filter channels. If not provided, returns all channels. */
    participant?: Address;
    /** Optional status to filter channels (e.g., "open", "closed"). If not provided, returns channels of all statuses. */
    status?: ChannelStatus;
}

/**
 * Represents the request structure for the 'get_channels' RPC method.
 */
export interface GetChannelsRPCRequest extends GenericRPCMessage {
    method: RPCMethod.GetChannels;
    params: GetChannelsRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRPCRequestParams {
    /** The participant address to retrieve RPC history for. Must be the authenticated wallet address. */
    participant: Address;
}

/**
 * Represents the request structure for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRPCRequest extends GenericRPCMessage {
    method: RPCMethod.GetRPCHistory;
    params: GetRPCHistoryRPCRequestParams;
}

/**
 * Represents the request parameters for the 'get_assets' RPC method.
 */
export interface GetAssetsRPCRequestParams {
    /** Optional chain ID to filter assets by network. If not provided, returns assets from all networks. */
    chain_id?: number;
}

/**
 * Represents the request structure for the 'get_assets' RPC method.
 */
export interface GetAssetsRPCRequest extends GenericRPCMessage {
    method: RPCMethod.GetAssets;
    params: GetAssetsRPCRequestParams;
}

/**
 * Union type for all possible RPC request types.
 * This allows for type-safe handling of different request structures.
 */
export type RPCRequest =
    | AuthChallengeRPCRequest
    | AuthVerifyRPCRequest
    | GetConfigRPCRequest
    | GetLedgerBalancesRPCRequest
    | GetLedgerEntriesRPCRequest
    | CreateAppSessionRPCRequest
    | SubmitStateRPCRequest
    | CloseAppSessionRPCRequest
    | GetAppDefinitionRPCRequest
    | GetAppSessionsRPCRequest
    | ResizeChannelRPCRequest
    | CloseChannelRPCRequest
    | GetChannelsRPCRequest
    | GetRPCHistoryRPCRequest
    | GetAssetsRPCRequest;

/**
 * Represents the response structure for an error response.
 */
export interface ErrorRPCResponse extends GenericRPCMessage {
    method: RPCMethod.Error;
    params: ErrorRPCParams;
}

/**
 * Represents the response structure for the 'get_config' RPC method.
 */
export interface GetConfigRPCResponse extends GenericRPCMessage {
    method: RPCMethod.GetConfig;
    params: GetConfigRPCParams;
}

/**
 * Represents the response structure for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRPCResponse extends GenericRPCMessage {
    method: RPCMethod.GetLedgerBalances;
    params: GetLedgerBalancesRPCParams[];
}

/**
 * Represents the response structure for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesRPCResponse extends GenericRPCMessage {
    method: RPCMethod.GetLedgerEntries;
    params: GetLedgerEntriesRPCParams[];
}

/**
 * Represents the response structure for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRPCResponse extends GenericRPCMessage {
    method: RPCMethod.CreateAppSession;
    params: CreateAppSessionRPCParams;
}

/**
 * Represents the response structure for the 'submit_state' RPC method.
 */
export interface SubmitStateRPCResponse extends GenericRPCMessage {
    method: RPCMethod.SubmitState;
    params: SubmitStateRPCParams;
}

/**
 * Represents the response structure for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRPCResponse extends GenericRPCMessage {
    method: RPCMethod.CloseAppSession;
    params: CloseAppSessionRPCParams;
}

/**
 * Represents the response structure for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRPCResponse extends GenericRPCMessage {
    method: RPCMethod.GetAppDefinition;
    params: GetAppDefinitionRPCParams;
}

/**
 * Represents the response structure for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRPCResponse extends GenericRPCMessage {
    method: RPCMethod.GetAppSessions;
    params: GetAppSessionsRPCParams[];
}

/**
 * Represents the response structure for the 'resize_channel' RPC method.
 */
export interface ResizeChannelRPCResponse extends GenericRPCMessage {
    method: RPCMethod.ResizeChannel;
    params: ResizeChannelRPCParams;
}

/**
 * Represents the response structure for the 'close_channel' RPC method.
 */
export interface CloseChannelRPCResponse extends GenericRPCMessage {
    method: RPCMethod.CloseChannel;
    params: CloseChannelRPCParams;
}

/**
 * Represents the response structure for the 'get_channels' RPC method.
 */
export interface GetChannelsRPCResponse extends GenericRPCMessage {
    method: RPCMethod.GetChannels;
    params: GetChannelsRPCParams[];
}

/**
 * Represents the response structure for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRPCResponse extends GenericRPCMessage {
    method: RPCMethod.GetRPCHistory;
    params: GetRPCHistoryRPCParams[];
}

/**
 * Represents the response structure for the 'get_assets' RPC method.
 */
export interface GetAssetsRPCResponse extends GenericRPCMessage {
    method: RPCMethod.GetAssets;
    params: GetAssetsRPCParams[];
}

/**
 * Represents the response structure for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRPCResponse extends GenericRPCMessage {
    method: RPCMethod.AuthVerify;
    params: AuthVerifyRPCParams;
}

/**
 * Union type for all possible RPC response types.
 * This allows for type-safe handling of different response structures.
 */
export type RPCResponse =
    | AuthChallengeRPCResponse
    | AuthVerifyRPCResponse
    | ErrorRPCResponse
    | GetConfigRPCResponse
    | GetLedgerBalancesRPCResponse
    | GetLedgerEntriesRPCResponse
    | CreateAppSessionRPCResponse
    | SubmitStateRPCResponse
    | CloseAppSessionRPCResponse
    | GetAppDefinitionRPCResponse
    | GetAppSessionsRPCResponse
    | ResizeChannelRPCResponse
    | CloseChannelRPCResponse
    | GetChannelsRPCResponse
    | GetRPCHistoryRPCResponse
    | GetAssetsRPCResponse;

/**
 * Maps RPC methods to their corresponding parameter types.
 */
export type RPCParamsByMethod = {
    [RPCMethod.AuthChallenge]: AuthChallengeRPCParams;
    [RPCMethod.AuthVerify]: AuthVerifyRPCParams;
    [RPCMethod.Error]: ErrorRPCParams;
    [RPCMethod.GetConfig]: GetConfigRPCParams;
    [RPCMethod.GetLedgerBalances]: GetLedgerBalancesRPCParams[];
    [RPCMethod.GetLedgerEntries]: GetLedgerEntriesRPCParams[];
    [RPCMethod.CreateAppSession]: CreateAppSessionRPCParams;
    [RPCMethod.SubmitState]: SubmitStateRPCParams;
    [RPCMethod.CloseAppSession]: CloseAppSessionRPCParams;
    [RPCMethod.GetAppDefinition]: GetAppDefinitionRPCParams;
    [RPCMethod.GetAppSessions]: GetAppSessionsRPCParams[];
    [RPCMethod.ResizeChannel]: ResizeChannelRPCParams;
    [RPCMethod.CloseChannel]: CloseChannelRPCParams;
    [RPCMethod.GetChannels]: GetChannelsRPCParams[];
    [RPCMethod.GetRPCHistory]: GetRPCHistoryRPCParams[];
    [RPCMethod.GetAssets]: GetAssetsRPCParams[];
};
