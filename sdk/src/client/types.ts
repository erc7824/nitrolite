import { Account, Hex, PublicClient, WalletClient, Chain, Transport, ParseAccount, Address } from "viem";
import { ContractAddresses } from "../abis";

/**
 * Channel identifier
 */
export type ChannelId = Hex;

/**
 * State hash
 */
export type StateHash = Hex;

/**
 * Signature structure used for state channel operations
 */
export interface Signature {
    v: number;
    r: Hex;
    s: Hex;
}

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
    participants: [Address, Address]; // List of participants in the channel [Host, Guest]
    adjudicator: Address; // Address of the contract that validates final states
    challenge: bigint; // Duration in seconds for challenge period
    nonce: bigint; // Unique per channel with same participants and adjudicator
}

/**
 * Channel status enum
 */
export enum StateIntent {
    OPERATE, // Operate the state application
    INITIALIZE, // Initial funding state
    RESIZE, // Resize state
    FINALIZE, // Final closing state
}

/**
 * Channel state structure
 */
export interface State {
    intent: StateIntent; // Intent of the state (e.g., INITIAL, ACTIVE, FINAL)
    version: bigint; // Version of the state, incremented for each update
    data: Hex; // Application data encoded, decoded by the adjudicator for business logic
    allocations: [Allocation, Allocation]; // Combined asset allocation and destination for each participant
    sigs: Signature[]; // stateHash signatures
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
     */
    walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;

    /**
     * Optional: A separate viem WalletClient used *only* for signing off-chain state updates (`signMessage`).
     * Provide this if you want to use a different key (e.g., a "hot" key from localStorage)
     * for state signing than the one used for on-chain transactions.
     * If omitted, `walletClient` will be used for state signing.
     */
    stateWalletClient?: WalletClient<Transport, Chain, ParseAccount<Account>>;

    /** Contract addresses required by the SDK. */
    addresses: ContractAddresses;

    /** Optional: Default challenge duration (in seconds) for new channels. Defaults to 0 if omitted. */
    challengeDuration?: bigint;
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
    finalState: {
        channelId: ChannelId;
        stateData: Hex;
        allocations: [Allocation, Allocation];
        version: bigint;
        serverSignature: Signature;
    };
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
 * Parameters required for checkpointing a state on-chain.
 */
export interface CheckpointChannelParams {
    channelId: ChannelId;
    candidateState: State;
    proofStates?: State[];
}

/**
 * Information about an account's funds in the custody contract.
 */
export interface AccountInfo {
    available: bigint;
    locked: bigint;
    channelCount: bigint;
}
