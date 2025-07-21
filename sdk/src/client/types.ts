import { Account, Hex, PublicClient, WalletClient, Chain, Transport, ParseAccount, Address } from 'viem';
import { ContractAddresses } from '../abis';

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
 * Allocation structure representing fund distribution
 */
export interface Allocation {
    destination: Address; // Where funds are sent on channel closure
    token: Address; // ERC-20 token address (zero address for ETH)
    amount: bigint; // Token amount allocated
}

/**
 * Channel configuration structure
 */
export interface Channel {
    participants: Address[]; // List of participants in the channel
    adjudicator: Address; // Address of the contract that validates final states
    challenge: bigint; // Duration in seconds for challenge period (uint64 in contract)
    nonce: bigint; // Unique per channel with same participants and adjudicator (uint64 in contract)
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
 * Channel data structure - contains all information about a channel
 */
export interface ChannelData {
    channel: Channel; // Channel configuration
    status: ChannelStatus; // Current status of the channel
    wallets: [Address, Address]; // List of participant wallet addresses
    challengeExpiry: bigint; // Timestamp when the challenge period ends
    lastValidState: State; // Last valid state of the channel recorded on-chain
}

interface UnsignedState {
    intent: StateIntent; // Intent of the state (uint8 enum in contract)
    version: bigint; // Version of the state (uint256 in contract)
    data: Hex; // Application data encoded (bytes in contract)
    allocations: Allocation[]; // Asset allocation array
}

/**
 * Channel state structure - matches the contract State struct
 */
export interface State extends UnsignedState {
    sigs: Signature[]; // State signatures array
}

/**
 * Extended state structure with channel ID and server signature to close the channel
 */
export interface FinalState extends UnsignedState {
    channelId: ChannelId;
    serverSignature: Signature;
}

// Legacy types for backward compatibility
export interface LegacyChannel {
    participants: [Address, Address]; // Legacy: exactly 2 participants
    adjudicator: Address;
    challenge: bigint;
    nonce: bigint;
    chainId: number; // Legacy field not in contract
}

export interface LegacyState {
    intent: StateIntent;
    version: bigint;
    data: Hex;
    allocations: [Allocation, Allocation]; // Legacy: exactly 2 allocations
    sigs: Signature[];
}

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
     * Optional: A separate viem WalletClient used *only* for signing off-chain state updates (`signMessage`).
     * Provide this if you want to use a different key (e.g., a "hot" key from localStorage)
     * for state signing than the one used for on-chain transactions.
     * If omitted, `walletClient` will be used for state signing.
     * @dev Note that the client's `signMessage` function should NOT add an EIP-191 prefix to the message signed. See {@link SignMessageFn} for details.
     * viem's `signMessage` can operate in `raw` mode, which suffice.
     */
    stateWalletClient?: WalletClient<Transport, Chain, ParseAccount<Account>>;

    /** Contract addresses required by the SDK. */
    addresses: ContractAddresses;

    /** Chain ID for the channel */
    chainId: number;

    /** Default challenge duration (in seconds) for new channels. */
    challengeDuration: bigint;
}

/**
 * Parameters required for creating a new state channel.
 */
export interface CreateChannelParams {
    initialAllocationAmounts: [bigint, bigint];
    stateData?: Hex;
}

/**
 * Parameters required for collaboratively closing a state channel.
 */
export interface CloseChannelParams {
    stateData?: Hex;
    finalState: FinalState;
}

/**
 * Parameters required for challenging a state channel.
 */
export interface ChallengeChannelParams {
    channelId: ChannelId;
    candidateState: State;
    proofStates?: State[];
}

/**
 * Parameters required for resizing a state channel.
 */
export interface ResizeChannelParams {
    resizeState: FinalState;
    proofStates: State[];
}

/**
 * Parameters required for checkpointing a state on-chain.
 */
export interface CheckpointChannelParams {
    channelId: ChannelId;
    candidateState: State;
    proofStates?: State[];
}
