import { Hex, Address, Hash } from 'viem';

/** Represents the status of a channel. */
export enum RPCChannelStatus {
    Open = 'open',
    Closed = 'closed',
    Resizing = 'resizing',
    Challenged = 'challenged',
}

/**
 * Represents the request parameters for the 'get_transactions' RPC method.
 */
export enum RPCTxType {
    Transfer = 'transfer',
    Deposit = 'deposit',
    Withdrawal = 'withdrawal',
    AppDeposit = 'app_deposit',
    AppWithdrawal = 'app_withdrawal',
    EscrowLock = 'escrow_lock',
    EscrowUnlock = 'escrow_unlock',
}

/**
 * Represents the protocol version and is used to provide backward compatibility as the API evolves.
 */
export enum RPCProtocolVersion {
    // NitroRPC_0_2 is the initial supported version of the NitroRPC protocol
    NitroRPC_0_2 = 'NitroRPC/0.2',
    // NitroRPC_0_4 adds support for App Session deposits and withdrawals
    NitroRPC_0_4 = 'NitroRPC/0.4',
}

/**
 * Defines the structure of an application definition used when creating an application.
 */
export interface RPCAppDefinition {
    /** Application identifier */
    application: string;
    /** Protocol identifies the version of the application protocol */
    protocol: RPCProtocolVersion;
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
export interface RPCChannelUpdate {
    /** The unique identifier for the channel. */
    channelId: Hex;
    /** The Ethereum address of the participant. */
    participant: Address;
    /** The current status of the channel (e.g., "open", "closed"). */
    status: RPCChannelStatus;
    /** The token contract address. */
    token: Address;
    /** The total amount in the channel. */
    amount: BigInt;
    /** The chain ID where the channel exists. */
    chainId: number;
    /** The adjudicator contract address. */
    adjudicator: Address;
    /** The challenge period in seconds. */
    challenge: number;
    /** The nonce value for the channel. */
    nonce: number;
    /** The version number of the channel. */
    version: number;
    /** The timestamp when the channel was created. */
    createdAt: Date;
    /** The timestamp when the channel was last updated. */
    updatedAt: Date;
}

export interface RPCChannelUpdateWithWallet extends RPCChannelUpdate {
    /** The Ethereum address of the wallet associated with the channel. */
    wallet: Address;
}

/**
 * Represents the network information for the 'get_config' RPC method.
 */
export interface RPCNetworkInfo {
    /** The chain ID of the network. */
    chainId: number;
    /** The name of the blockchain (e.g., "polygon_amoy", "base_sepolia"). */
    name: string;
    /** The custody contract address for the network. */
    custodyAddress: Address;
    /** The adjudicator contract address for the network. */
    adjudicatorAddress: Address;
}

export enum RPCAppStateIntent {
    /** Intent for a standard state update */
    Operate = 'operate',
    /** Intent for depositing funds into the app session */
    Deposit = 'deposit',
    /** Intent for withdrawing funds from the app session */
    Withdraw = 'withdraw',
}

/**
 * Represents the app session information.
 */
export interface RPCAppSession {
    /** The unique identifier for the application session. */
    appSessionId: Hex;
    /** Application identifier */
    application: string;
    /** The current status of the channel (e.g., "open", "closed"). */
    status: RPCChannelStatus;
    /** List of participant Ethereum addresses. */
    participants: Address[];
    /** Protocol identifies the version of the application protocol */
    protocol: RPCProtocolVersion;
    /** The challenge period in seconds. */
    challenge: number;
    /** The signature weights for each participant. */
    weights: number[];
    /** The minimum number of signatures required for state updates. */
    quorum: number;
    /** The version number of the session. */
    version: number;
    /** The nonce value for the session. */
    nonce: number;
    /** The timestamp when the session was created. */
    createdAt: Date;
    /** The timestamp when the session was last updated. */
    updatedAt: Date;
    /** Optional session data as a JSON string that stores application-specific state or metadata. */
    sessionData?: string;
}

/**
 * Represents RPC entry in the history.
 */
export interface RPCHistoryEntry {
    /** Unique identifier for the RPC entry. */
    id: number;
    /** The Ethereum address of the sender. */
    sender: Address;
    /** The request ID for the RPC call. */
    reqId: number;
    /** The RPC method name. */
    method: string;
    /** The JSON string of the request parameters. */
    params: string;
    /** The timestamp of the RPC call. */
    timestamp: number;
    /** Array of request signatures. */
    reqSig: Hex[];
    /** Array of response signatures. */
    resSig: Hex[];
    /** The JSON string of the response. */
    response: string;
}

/**
 * Represents Asset information received from the clearnode.
 */
export interface RPCAsset {
    /** The token contract address. */
    token: Address;
    /** The chain ID where the asset exists. */
    chainId: number;
    /** The asset symbol (e.g., "eth", "usdc"). */
    symbol: string;
    /** The number of decimal places for the asset. */
    decimals: number;
}

/**
 * Represents the balance information from clearnode.
 */
export interface RPCBalance {
    /** The asset symbol (e.g., "eth", "usdc"). */
    asset: string;
    /** The balance amount. */
    amount: string;
}

export type LedgerAccountType = Address | Hash;

/**
 * Represents a single entry in the ledger.
 */
export interface RPCLedgerEntry {
    /** Unique identifier for the ledger entry. */
    id: number;
    /** The account identifier associated with the entry. */
    accountId: LedgerAccountType;
    /** The type of account (e.g., "wallet", "channel"). */
    accountType: number;
    /** The asset symbol for the entry. */
    asset: string;
    /** The Ethereum address of the participant. */
    participant: Address;
    /** The credit amount. */
    credit: string;
    /** The debit amount. */
    debit: string;
    /** The timestamp when the entry was created. */
    createdAt: Date;
}

/**
 * Represents the parameters for the transfer transaction.
 */
export interface RPCTransaction {
    /** Unique identifier for the transfer. */
    id: number;
    /** The type of transaction. */
    txType: RPCTxType;
    /** The source address from which assets were transferred. */
    fromAccount: LedgerAccountType;
    /** The user tag for the source account (optional). */
    fromAccountTag?: string;
    /** The destination address to which assets were transferred. */
    toAccount: LedgerAccountType;
    /** The user tag for the destination account (optional). */
    toAccountTag?: string;
    /** The asset symbol that was transferred. */
    asset: string;
    /** The amount that was transferred. */
    amount: string;
    /** The timestamp when the transfer was created. */
    createdAt: Date;
}

/**
 * Represents a generic RPC message structure that includes common fields.
 * This interface is extended by specific RPC request and response types.
 */
export interface RPCAllowance {
    /** The symbol of the asset (e.g., "eth", "usdc"). */
    asset: string;
    /** The amount of the asset that is allowed to be spent. */
    amount: string;
}

/**
 * Represents an allowance with usage tracking, combining both limit and used amount.
 */
export interface RPCAllowanceUsage {
    /** The symbol of the asset (e.g., "eth", "usdc"). */
    asset: string;
    /** The maximum amount of the asset that is allowed to be spent. */
    allowance: string;
    /** The amount of the asset that has been used. */
    used: string;
}

/**
 * Represents a session key with its allowances and usage tracking.
 */
export interface RPCSessionKey {
    /** Unique identifier for the session key record. */
    id: number;
    /** The address of the session key that can sign transactions. */
    sessionKey: Address;
    /** Name of the application this session key is authorized for. */
    application: string;
    /** Array of asset allowances with usage tracking. */
    allowances: RPCAllowanceUsage[];
    /** Permission scope for this session key (e.g., "app.create", "ledger.readonly"). */
    scope?: string;
    /** When this session key expires. */
    expiresAt: Date;
    /** When the session key was created. */
    createdAt: Date;
}

// TODO: create single domain allocation type

/**
 * Represents the allocation of assets within an application session.
 * This structure is used to define the initial allocation of assets among participants.
 * It includes the participant's address, the asset (usdc, usdt, etc) being allocated, and the amount.
 */
export interface RPCAppSessionAllocation {
    /** The symbol of the asset (e.g., "eth", "usdc"). */
    asset: string;
    /** The amount of the asset. Must be a positive number. */
    amount: string;
    /** The Ethereum address of the participant receiving the allocation. */
    participant: Address;
}

/**
 * Represents the allocation of assets for an RPC transfer.
 * This structure is used to define the asset and amount being transferred to a specific destination address.
 */
export interface RPCChannelAllocation {
    /** The destination address for the allocation. */
    destination: Address;
    /** The token contract address for the asset being allocated. */
    token: Address;
    /** The amount of the asset being allocated. */
    amount: bigint;
}

/**
 * Represents the allocation of assets for an RPC transfer.
 * This structure is used to define the asset and amount being transferred.
 */
export interface RPCTransferAllocation {
    /** The symbol of the asset (e.g., "eth", "usdc"). */
    asset: string;
    /** The amount of the asset being transferred. */
    amount: string;
}

/**
 * Represents the state of a channel operation.
 * This structure is used to define the intent, version, state data, and allocations for a channel operation.
 */
export interface RPCChannelOperationState {
    /** The intent type for the state update. */
    intent: number;
    /** The version number of the channel. */
    version: number;
    /** The encoded state data for the channel. */
    stateData: Hex;
    /** The list of allocations for the channel. */
    allocations: RPCChannelAllocation[];
}

export interface RPCChannelOperation {
    /** The unique identifier for the channel. */
    channelId: Hex;
    /** The channel state object. */
    state: RPCChannelOperationState;
    /** The server's signature for the state update. */
    serverSignature: Hex;
}

/**
 * Represents the fixed part of a channel, containing essential metadata.
 */
export interface RPCChannel {
    participants: Address[];
    adjudicator: Address;
    challenge: number;
    nonce: number;
}
