import { Address, Hex } from 'viem';
import { RPCMethod, GenericRPCMessage, AppDefinition, RPCChannelStatus, TransferAllocation } from '.';

/**
 * Represents the request parameters for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeRequestParams {
    /** The challenge message to be signed by the client for authentication. */
    challenge_message: string;
}
export type AuthChallengeRPCRequestParams = AuthChallengeRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeRequest extends GenericRPCMessage {
    method: RPCMethod.AuthChallenge;
    params: AuthChallengeRequestParams[];
}

/**
 * Represents the request parameters for the 'auth_verify' RPC method.
 * Either JWT or challenge must be provided. JWT takes precedence over challenge.
 */
export type AuthVerifyRequestParams =
    | {
          /** JSON Web Token for authentication. */
          jwt: string;
      }
    | {
          /** The challenge token received from auth_challenge response. Used to verify the client's signature and prevent replay attacks. */
          challenge: string;
      };
export type AuthVerifyRPCRequestParams = AuthVerifyRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRequest extends GenericRPCMessage {
    method: RPCMethod.AuthVerify;
    params: AuthVerifyRequestParams[];
}

/**
 * Represents the request structure for the 'get_config' RPC method.
 */
export interface GetConfigRequest extends GenericRPCMessage {
    method: RPCMethod.GetConfig;
    params: [];
}

/**
 * Represents the request parameters for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRequestParams {
    /** The participant address to filter balances. */
    participant: Address;
    /** Optional account ID to filter balances. If provided, overrides the participant address. */
    account_id?: string;
}
export type GetLedgerBalancesRPCRequestParams = GetLedgerBalancesRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRequest extends GenericRPCMessage {
    method: RPCMethod.GetLedgerBalances;
    params: [GetLedgerBalancesRequestParams];
}

/**
 * Represents the request parameters for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesRequestParams {
    /** The account ID to filter ledger entries. */
    account_id?: string;
    /** The asset symbol to filter ledger entries. */
    asset?: string;
    /** Optional wallet address to filter ledger entries. If provided, overrides the authenticated wallet. */
    wallet?: Address;
}
export type GetLedgerEntriesRPCRequestParams = GetLedgerEntriesRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesRequest extends GenericRPCMessage {
    method: RPCMethod.GetLedgerEntries;
    params: [GetLedgerEntriesRequestParams];
}

/**
 * Represents the request parameters for the 'get_transactions' RPC method.
 */
export enum TxType {
    Transfer = 'transfer',
    Deposit = 'deposit',
    Withdrawal = 'withdrawal',
    AppDeposit = 'app_deposit',
    AppWithdrawal = 'app_withdrawal',
}

/**
 * Represents the request parameters for the 'get_transactions' RPC method.
 */
export interface GetLedgerTransactionsFilters {
    /** The asset symbol to filter transactions. */
    asset?: string;
    /** The transaction type to filter transactions. */
    tx_type?: TxType;
    /** Pagination offset. */
    offset?: number;
    /** Number of transactions to return. */
    limit?: number;
    /** Sort order by created_at. */
    sort?: 'asc' | 'desc';
}

export interface GetLedgerTransactionsRequestParams extends GetLedgerTransactionsFilters {
    /** The account ID to filter transactions. */
    account_id: string;
}
export type GetLedgerTransactionsRPCRequestParams = GetLedgerTransactionsRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'get_transactions' RPC method.
 */
export interface GetLedgerTransactionsRequest extends GenericRPCMessage {
    method: RPCMethod.GetLedgerTransactions;
    params: GetLedgerTransactionsRequestParams;
}

/**
 * Represents the request parameters for the 'get_user_tag' RPC method.
 */
export interface GetUserTagRequestParams {
    // This method takes no parameters - empty object
}
export type GetUserTagRPCRequestParams = GetUserTagRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'get_user_tag' RPC method.
 */
export interface GetUserTagRequest extends GenericRPCMessage {
    method: RPCMethod.GetUserTag;
    params: [];
}

/** Represents the allocation of assets within an application session.
 * This structure is used to define allocation of assets among participants.
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
 * Represents the request parameters for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRequestParams {
    /** The detailed definition of the application being created, including protocol, participants, weights, and quorum. */
    definition: AppDefinition;
    /** The initial allocation distribution among participants. Each participant must have sufficient balance for their allocation. */
    allocations: AppSessionAllocation[];
    /** Optional session data as a JSON string that can store application-specific state or metadata. */
    session_data?: string;
}
export type CreateAppSessionRPCRequestParams = CreateAppSessionRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRequest extends GenericRPCMessage {
    method: RPCMethod.CreateAppSession;
    params: [CreateAppSessionRequestParams];
}

/**
 * Represents the request parameters for the 'submit_app_state' RPC method.
 */
export interface SubmitAppStateRequestParams {
    /** The unique identifier of the application session to update. */
    app_session_id: Hex;
    /** The new allocation distribution among participants. Must include all participants and maintain total balance. */
    allocations: AppSessionAllocation[];
    /** Optional session data as a JSON string that can store application-specific state or metadata. */
    session_data?: string;
}
export type SubmitAppStateRPCRequestParams = SubmitAppStateRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'submit_app_state' RPC method.
 */
export interface SubmitAppStateRequest extends GenericRPCMessage {
    method: RPCMethod.SubmitAppState;
    params: [SubmitAppStateRequestParams];
}

/**
 * Represents the request parameters for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRequestParams {
    /** The unique identifier of the application session to close. */
    app_session_id: Hex;
    /** The final allocation distribution among participants upon closing. Must include all participants and maintain total balance. */
    allocations: AppSessionAllocation[];
    /** Optional session data as a JSON string that can store application-specific state or metadata. */
    session_data?: string;
}
export type CloseAppSessionRPCRequestParams = CloseAppSessionRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRequest extends GenericRPCMessage {
    method: RPCMethod.CloseAppSession;
    params: [CloseAppSessionRequestParams];
}

/**
 * Represents the request parameters for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRequestParams {
    /** The unique identifier of the application session to retrieve. */
    app_session_id: Hex;
}
export type GetAppDefinitionRPCRequestParams = GetAppDefinitionRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRequest extends GenericRPCMessage {
    method: RPCMethod.GetAppDefinition;
    params: [GetAppDefinitionRequestParams];
}

/**
 * Represents the request parameters for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRequestParams {
    /** The participant address to filter application sessions. */
    participant: Address;
    /** The status to filter application sessions (e.g., "open", "closed"). */
    status: RPCChannelStatus;
}
export type GetAppSessionsRPCRequestParams = GetAppSessionsRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRequest extends GenericRPCMessage {
    method: RPCMethod.GetAppSessions;
    params: [GetAppSessionsRequestParams];
}

/**
 * Represents the request parameters for the 'resize_channel' RPC method.
 */
export type ResizeChannelRequestParams = {
    /** The unique identifier of the channel to resize. */
    channel_id: Hex;
    /** Amount to resize the channel by (can be positive or negative). Required if allocate_amount is not provided. */
    resize_amount?: bigint;
    /** Amount to allocate from the unified balance to the channel. Required if resize_amount is not provided. */
    allocate_amount?: bigint;
    /** The address where the resized funds will be sent. */
    funds_destination: Address;
};
export type ResizeChannelRPCRequestParams = ResizeChannelRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'resize_channel' RPC method.
 */
export interface ResizeChannelRequest extends GenericRPCMessage {
    method: RPCMethod.ResizeChannel;
    params: [ResizeChannelRequestParams];
}

/**
 * Represents the request parameters for the 'close_channel' RPC method.
 */
export interface CloseChannelRequestParams {
    /** The unique identifier of the channel to close. */
    channel_id: Hex;
    /** The address where the channel funds will be sent upon closing. */
    funds_destination: Address;
}
export type CloseChannelRPCRequestParams = CloseChannelRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'close_channel' RPC method.
 */
export interface CloseChannelRequest extends GenericRPCMessage {
    method: RPCMethod.CloseChannel;
    params: [CloseChannelRequestParams];
}

/**
 * Represents the request parameters for the 'get_channels' RPC method.
 */
export interface GetChannelsRequestParams {
    /** The participant address to filter channels. */
    participant: Address;
    /** The status to filter channels (e.g., "open", "closed"). */
    status: RPCChannelStatus;
}
export type GetChannelsRPCRequestParams = GetChannelsRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'get_channels' RPC method.
 */
export interface GetChannelsRequest extends GenericRPCMessage {
    method: RPCMethod.GetChannels;
    params: [GetChannelsRequestParams];
}

/**
 * Represents the request structure for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRequest extends GenericRPCMessage {
    method: RPCMethod.GetRPCHistory;
    params: [];
}

/**
 * Represents the request parameters for the 'get_assets' RPC method.
 */
export interface GetAssetsRequestParams {
    /** Optional chain ID to filter assets by network. If not provided, returns assets from all networks. */
    chain_id?: number;
}
export type GetAssetsRPCRequestParams = GetAssetsRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'get_assets' RPC method.
 */
export interface GetAssetsRequest extends GenericRPCMessage {
    method: RPCMethod.GetAssets;
    params: [GetAssetsRequestParams];
}

/** Represents a single allowance for an asset, used in application sessions.
 * This structure defines the symbol of the asset and the amount that is allowed to be spent.
 */
export type Allowance = {
    /** The symbol of the asset (e.g., "USDC", "USDT"). */
    asset: string;
    /** The amount of the asset that is allowed to be spent. */
    amount: string;
};

/**
 * Represents the request parameters for the 'auth_request' RPC method.
 */
export interface AuthRequestParams {
    /** The Ethereum address of the wallet being authorized. */
    wallet: Address;
    /** The session key address associated with the authentication attempt. */
    participant: Address;
    /** The name of the application being authorized. */
    app_name: string;
    /** The allowances for the connection. */
    allowances: Allowance[];
    /** The expiration timestamp for the authorization. */
    expire: string;
    /** The scope of the authorization. */
    scope: string;
    /** The application address being authorized. */
    application: Address;
}
export type AuthRequestRPCRequestParams = AuthRequestParams; // for backward compatibility

/**
 * Represents the request structure for the 'auth_request' RPC method.
 */
export interface AuthRequest extends GenericRPCMessage {
    method: RPCMethod.AuthRequest;
    params: AuthRequestParams[];
}

/**
 * Represents the request structure for the 'message' RPC method.
 */
export interface MessageRequest extends GenericRPCMessage {
    method: RPCMethod.Message;
    /** The message parameters are handled by the virtual application */
    params: any[];
}

/**
 * Represents the request structure for the 'ping' RPC method.
 */
export interface PingRequest extends GenericRPCMessage {
    method: RPCMethod.Ping;
    /** No parameters needed for ping */
    params: [];
}

/**
 * Represents the request structure for the 'pong' RPC method.
 */
export interface PongRequest extends GenericRPCMessage {
    method: RPCMethod.Pong;
    /** No parameters needed for pong */
    params: [];
}

/**
 * Represents the request parameters for the 'transfer' RPC method.
 */
export interface TransferRequestParams {
    /** The destination address to transfer assets to. Required if destination_user_tag is not provided. */
    destination?: Address;
    /** The destination user tag to transfer assets to. Required if destination is not provided. */
    destination_user_tag?: string;
    /** The assets and amounts to transfer. */
    allocations: TransferAllocation[];
}

/**
 * Represents the request structure for the 'transfer' RPC method.
 */
export interface TransferRequest extends GenericRPCMessage {
    method: RPCMethod.Transfer;
    params: TransferRequestParams;
}
export type TransferRPCRequestParams = TransferRequestParams; // for backward compatibility

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
    | CreateAppSessionRequest
    | SubmitAppStateRequest
    | CloseAppSessionRequest
    | GetAppDefinitionRequest
    | GetAppSessionsRequest
    | ResizeChannelRequest
    | CloseChannelRequest
    | GetChannelsRequest
    | GetRPCHistoryRequest
    | GetAssetsRequest
    | PingRequest
    | PongRequest
    | MessageRequest
    | TransferRequest;

/**
 * Maps RPC methods to their corresponding request parameter types.
 */
export type RPCRequestParamsByMethod = {
    [RPCMethod.AuthChallenge]: AuthChallengeRequestParams;
    [RPCMethod.AuthVerify]: AuthVerifyRequestParams;
    [RPCMethod.AuthRequest]: AuthRequestParams;
    [RPCMethod.GetConfig]: [];
    [RPCMethod.GetLedgerBalances]: GetLedgerBalancesRequestParams;
    [RPCMethod.GetLedgerEntries]: GetLedgerEntriesRequestParams;
    [RPCMethod.GetLedgerTransactions]: GetLedgerTransactionsRequestParams;
    [RPCMethod.GetUserTag]: [];
    [RPCMethod.CreateAppSession]: CreateAppSessionRequestParams;
    [RPCMethod.SubmitAppState]: SubmitAppStateRequestParams;
    [RPCMethod.CloseAppSession]: CloseAppSessionRequestParams;
    [RPCMethod.GetAppDefinition]: GetAppDefinitionRequestParams;
    [RPCMethod.GetAppSessions]: GetAppSessionsRequestParams;
    [RPCMethod.ResizeChannel]: ResizeChannelRequestParams;
    [RPCMethod.CloseChannel]: CloseChannelRequestParams;
    [RPCMethod.GetChannels]: GetChannelsRequestParams;
    [RPCMethod.GetRPCHistory]: [];
    [RPCMethod.GetAssets]: GetAssetsRequestParams;
    [RPCMethod.Ping]: [];
    [RPCMethod.Pong]: [];
    [RPCMethod.Message]: any[];
    [RPCMethod.Transfer]: TransferRequestParams;
};
