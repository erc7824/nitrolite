import { Account, Hex, PublicClient, WalletClient, Chain, Transport, ParseAccount, Address } from 'viem';
import { ContractAddresses } from '../abis';
import { StateSigner } from './signer';

/**
 * Channel identifier
 */
export type ChannelId = Hex;

/**
 * State hash
 */
export type StateHash = Hex;

/**
 * Signature type used when signing states
 * @dev Hex is used to support EIP-1271 and EIP-6492 signatures.
 */
export type Signature = Hex;

/**
 * Ledger structure - represents token allocations and flows on a specific chain
 * Used in V1 state structure for tracking balances across home and non-home chains
 */
export interface Ledger {
    chainId: bigint; // Chain ID where this ledger exists (uint64 in contract)
    token: Address; // Token contract address (zero address for ETH)
    decimals: number; // Token decimals (uint8 in contract)
    userAllocation: bigint; // User's allocated amount
    userNetFlow: bigint; // User's net flow (can be negative, int256 in contract)
    nodeAllocation: bigint; // Node's allocated amount
    nodeNetFlow: bigint; // Node's net flow (can be negative, int256 in contract)
}

/**
 * Channel definition structure for V1 contracts
 * Defines the parameters of a payment channel
 */
export interface ChannelDefinition {
    challengeDuration: number; // Duration in seconds for challenge period (uint32 in contract)
    user: Address; // User's wallet address
    node: Address; // Node's wallet address
    nonce: bigint; // Unique nonce for channel creation (uint64 in contract)
    metadata: Hex; // Additional metadata (bytes32 in contract)
}


/**
 * Channel status enum - represents the various states a channel can be in
 */
export enum ChannelStatus {
    VOID, // Channel was not created, State.version must be 0
    INITIAL, // Channel is created and in funding process, State.version must be 0
    ACTIVE, // Channel fully funded and operational, State.version is greater than 0
    DISPUTE, // Challenge period is active
    FINAL, // Final state, channel can be closed
}

/**
 * Channel status enum - matches the StateIntent enum in the contract
 */
export enum StateIntent {
    OPERATE = 0, // Operate the state application
    INITIALIZE = 1, // Initial funding state
    RESIZE = 2, // Resize state
    FINALIZE = 3, // Final closing state
}

/**
 * Channel data structure - contains all information about a channel (V1)
 */
export interface ChannelData {
    definition: ChannelDefinition; // Channel definition
    status: ChannelStatus; // Current status of the channel
    lastState: State; // Last state of the channel recorded on-chain
    challengeExpiry: bigint; // Timestamp when the challenge period ends (0 if not challenged)
}

/**
 * V1 State structure - represents the channel state with home and non-home ledgers
 * This is the primary state structure used in V1 contracts
 */
export interface State {
    version: bigint; // State version number (uint64 in contract)
    intent: StateIntent; // Intent of the state (uint8 enum in contract)
    metadata: Hex; // Additional metadata (bytes32 in contract)
    homeState: Ledger; // Home chain ledger state
    nonHomeState: Ledger; // Non-home chain ledger state (can be zero for single-chain)
    userSig: Hex; // User's signature
    nodeSig: Hex; // Node's signature
}

/**
 * Unsigned portion of V1 State - used for signing
 * Omits the signature fields from State
 */
export type UnsignedStateV1 = Omit<State, 'userSig' | 'nodeSig'>;


/**
 * Configuration for initializing the NitroliteClient.
 */
export interface NitroliteClientConfig {
    /** The viem PublicClient for reading blockchain data. */
    publicClient: PublicClient;

    /**
     * The viem WalletClient used for:
     * 1. Sending on-chain transactions in direct execution methods (e.g., `client.deposit`).
     * 2. Providing the 'account' context for transaction preparation (`client.txPreparer`).
     * 3. Signing off-chain states *if* `stateWalletClient` is not provided.
     * @dev Note that the client's `signMessage` function should NOT add an EIP-191 prefix to the message signed. See {@link SignMessageFn} for details.
     * viem's `signMessage` can operate in `raw` mode, which suffice.
     */
    walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;

    /**
     * Implementation of the StateSigner interface used for signing protocol states.
     */
    stateSigner: StateSigner;

    /** Contract addresses required by the SDK. */
    addresses: ContractAddresses;

    /** Chain ID for the channel */
    chainId: number;

    /** Default challenge duration (in seconds) for new channels. */
    challengeDuration: number;
}

/**
 * Parameters required for creating a new state channel (V1).
 */
export interface CreateChannelParams {
    definition: ChannelDefinition;
    initialState: State;
}

/**
 * Parameters required for collaboratively closing a state channel (V1).
 */
export interface CloseChannelParams {
    channelId: ChannelId;
    finalState: State;
    proofs?: State[];
}

/**
 * Parameters required for challenging a state channel (V1).
 */
export interface ChallengeChannelParams {
    channelId: ChannelId;
    candidateState: State;
    proofs?: State[];
    challengerSig: Hex;
}

/**
 * Parameters required for checkpointing a state on-chain (V1).
 */
export interface CheckpointChannelParams {
    channelId: ChannelId;
    candidateState: State;
    proofs?: State[];
}

/**
 * Parameters required for depositing to a channel (V1).
 */
export interface DepositToChannelParams {
    channelId: ChannelId;
    candidate: State;
}

/**
 * Parameters required for withdrawing from a channel (V1).
 */
export interface WithdrawFromChannelParams {
    channelId: ChannelId;
    candidate: State;
}
