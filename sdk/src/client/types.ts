import { Address, Hex } from 'viem';

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
    token: Address; // ERC-20 token contract address
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
 * Channel state structure
 */
export interface State {
    data: Hex; // Application data encoded, decoded by the adjudicator for business logic
    allocations: [Allocation, Allocation]; // Combined asset allocation and destination for each participant
    sigs: Signature[]; // stateHash signatures
}

/**
 * Adjudicator status enum
 */
export enum AdjudicatorStatus {
    VOID = 0, // Channel was never active or have an anomaly
    PARTIAL = 1, // Partial funding waiting for other participants
    ACTIVE = 2, // Channel fully funded using open or state are valid
    INVALID = 3, // Channel state is invalid
    FINAL = 4, // This is the FINAL State channel can be closed
}

/**
 * Participant roles in a channel
 */
export enum Role {
    UNDEFINED = -1,
    HOST = 0,
    GUEST = 1,
}
