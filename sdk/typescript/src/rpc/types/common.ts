import { Hex, Address, Hash } from 'viem';

/** Represents the status of a channel. */
export enum RPCChannelStatus {
    Void = 'void',
    Open = 'open',
    Challenged = 'challenged',
    Closed = 'closed',
}

/**
 * Participant in an application session.
 */
export interface RPCAppParticipant {
    /** Participant's wallet address */
    walletAddress: Address;
    /** Signature weight for the participant */
    signatureWeight: number;
}

/**
 * Defines the structure of an application definition used when creating an application.
 */
export interface RPCAppDefinition {
    /** Application identifier from an app registry */
    application: string;
    /** List of participants in the app session */
    participants: RPCAppParticipant[];
    /** Quorum required for the app session */
    quorum: number;
    /** A unique number to prevent replay attacks */
    nonce: number;
}

/**
 * Represents the network information for the 'get_config' RPC method.
 */
export interface RPCNetworkInfo {
    /** Blockchain name */
    name: string;
    /** Blockchain network ID */
    blockchainId: number;
    /** Address of the main contract on this blockchain */
    contractAddress: Address;
}

export enum RPCAppStateIntent {
    /** Intent for a standard state update */
    Operate = 'operate',
    /** Intent for depositing funds into the app session */
    Deposit = 'deposit',
    /** Intent for withdrawing funds from the app session */
    Withdraw = 'withdraw',
    /** Intent for rebalancing multiple app sessions atomically */
    Rebalance = 'rebalance',
}

/**
 * Represents the app session information.
 */
export interface RPCAppSession {
    /** A unique application session identifier */
    appSessionId: string;
    /** Session status (open/closed) */
    status: string;
    /** List of participant wallet addresses with weights */
    participants: RPCAppParticipant[];
    /** JSON stringified session data */
    sessionData?: string;
    /** Quorum required for operations */
    quorum: number;
    /** Current version of the session state */
    version: number;
    /** Nonce for the session */
    nonce: number;
    /** List of allocations in the app state */
    allocations: RPCAppSessionAllocation[];
}

/**
 * Token information for a specific blockchain.
 */
export interface RPCToken {
    /** Token name */
    name: string;
    /** Token symbol */
    symbol: string;
    /** Token contract address */
    address: Address;
    /** Blockchain network ID */
    blockchainId: number;
    /** Number of decimal places */
    decimals: number;
}

/**
 * Asset information received from the clearnode.
 */
export interface RPCAsset {
    /** Asset name */
    name: string;
    /** Asset symbol (e.g., "eth", "usdc") */
    symbol: string;
    /** Supported tokens for this asset across different blockchains */
    tokens: RPCToken[];
}

/**
 * Represents an allowance for session key registration.
 */
export interface RPCAllowance {
    /** The symbol of the asset (e.g., "eth", "usdc"). */
    asset: string;
    /** The maximum amount of the asset that is allowed to be spent. */
    allowance: string;
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
 * Represents an application session state update.
 */
export interface RPCAppStateUpdate {
    /** A unique application session identifier */
    app_session_id: string;
    /** The intent of the app session update */
    intent: RPCAppStateIntent;
    /** Version of the app state */
    version: number;
    /** List of allocations in the app state */
    allocations: RPCAppSessionAllocation[];
    /** JSON stringified session data */
    session_data?: string;
}

/**
 * Represents a signed application session state update.
 */
export interface RPCSignedAppStateUpdate {
    /** The application session state update */
    app_state_update: RPCAppStateUpdate;
    /** The signature quorum for the application session */
    quorum_sigs: Hex[];
}

/**
 * Transition types for state management.
 * Represents different types of state transitions that can occur in channels.
 */
export enum RPCTransitionType {
    TransferReceive = 'transfer_receive',
    TransferSend = 'transfer_send',
    Release = 'release',
    Commit = 'commit',
    HomeDeposit = 'home_deposit',
    HomeWithdrawal = 'home_withdrawal',
    MutualLock = 'mutual_lock',
    EscrowDeposit = 'escrow_deposit',
    EscrowLock = 'escrow_lock',
    EscrowWithdraw = 'escrow_withdraw',
    Migrate = 'migrate',
}

/**
 * Represents a single state transition.
 */
export interface RPCTransition {
    /** Type of the state transition */
    type: RPCTransitionType;
    /** Associated blockchain transaction hash (optional) */
    txHash?: string;
    /** Account identifier for the transition */
    accountId: string;
    /** Amount involved in the transition (decimal string) */
    amount: string;
}

/**
 * Represents ledger balances for a channel.
 * Tracks user and node balances with their net flows.
 */
export interface RPCLedger {
    /** Token contract address */
    tokenAddress: Address;
    /** Blockchain network ID */
    blockchainId: number;
    /** User's current balance (decimal string) */
    userBalance: string;
    /** User's net flow (decimal string) */
    userNetFlow: string;
    /** Node's current balance (decimal string) */
    nodeBalance: string;
    /** Node's net flow (decimal string) */
    nodeNetFlow: string;
}

/**
 * Represents a complete state.
 * States track all transitions and current balances for a user's asset.
 */
export interface RPCState {
    /** Deterministic hash identifier for the state */
    id: string;
    /** List of state transitions */
    transitions: RPCTransition[];
    /** Asset symbol (e.g., "usdc", "eth") */
    asset: string;
    /** User's wallet address */
    userWallet: Address;
    /** User epoch index */
    epoch: number;
    /** State version number */
    version: number;
    /** Home channel ID (optional) */
    homeChannelId?: string;
    /** Escrow channel ID (optional) */
    escrowChannelId?: string;
    /** Home channel ledger */
    homeLedger: RPCLedger;
    /** Escrow channel ledger (optional) */
    escrowLedger?: RPCLedger;
    /** User's signature (optional) */
    userSig?: Hex;
    /** Node's signature (optional) */
    nodeSig?: Hex;
}

/**
 * Represents channel information.
 * Channels can be either home (user-node) or escrow (multi-party).
 */
export interface RPCChannel {
    /** Unique channel identifier */
    channelId: string;
    /** User's wallet address */
    userWallet: Address;
    /** Node's wallet address */
    nodeWallet: Address;
    /** Channel type: home or escrow */
    type: 'home' | 'escrow';
    /** Blockchain network ID */
    blockchainId: number;
    /** Token contract address */
    tokenAddress: Address;
    /** Challenge period in seconds */
    challenge: number;
    /** Channel nonce for uniqueness */
    nonce: number;
    /** Channel status */
    status: 'void' | 'open' | 'challenged' | 'closed';
    /** On-chain state version */
    stateVersion: number;
}

/**
 * Channel definition for creating new channels.
 */
export interface RPCChannelDefinition {
    /** Unique number for replay protection */
    nonce: number;
    /** Challenge period in seconds */
    challenge: number;
}

/**
 * Transaction types supported by the API.
 */
export enum RPCTransactionType {
    Transfer = 'transfer',
    Release = 'release',
    Commit = 'commit',
    HomeDeposit = 'home_deposit',
    HomeWithdrawal = 'home_withdrawal',
    MutualLock = 'mutual_lock',
    EscrowDeposit = 'escrow_deposit',
    EscrowLock = 'escrow_lock',
    EscrowWithdraw = 'escrow_withdraw',
    Migrate = 'migrate',
}

/**
 * Represents a transaction.
 */
export interface RPCTransaction {
    /** Transaction reference ID */
    id: string;
    /** Asset symbol */
    asset: string;
    /** Transaction type */
    txType: RPCTransactionType;
    /** Sender account identifier */
    fromAccount: string;
    /** Recipient account identifier */
    toAccount: string;
    /** Sender's new state ID after transaction (optional) */
    senderNewStateId?: string;
    /** Receiver's new state ID after transaction (optional) */
    receiverNewStateId?: string;
    /** Transaction amount (decimal string) */
    amount: string;
    /** Transaction creation timestamp */
    createdAt: Date;
}

/**
 * Balance entry for user balances.
 */
export interface RPCBalanceEntry {
    /** Asset symbol */
    asset: string;
    /** Balance amount (decimal string) */
    amount: string;
}

/**
 * Pagination parameters for list queries.
 */
export interface PaginationParams {
    /** Number of items to skip */
    offset?: number;
    /** Maximum number of items to return */
    limit?: number;
    /** Sort order: ascending or descending */
    sort?: 'asc' | 'desc';
}

/**
 * Pagination metadata returned with paginated responses.
 */
export interface PaginationMetadata {
    /** Current page number */
    page: number;
    /** Items per page */
    perPage: number;
    /** Total number of items */
    totalCount: number;
    /** Total number of pages */
    pageCount: number;
}
