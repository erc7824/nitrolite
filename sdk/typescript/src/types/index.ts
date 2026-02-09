import { Address, Hex } from 'viem';
import { ChannelDefinition, State } from '../client/types';

/**
 * Participant roles in a channel (V1: User and Node)
 */
export enum Role {
    UNDEFINED = -1,
    USER = 0, // User participant
    NODE = 1, // Node participant
}

/**
 * Standard app data types (Example structures)
 */
export namespace AppDataTypes {
    export interface NumericState {
        value: bigint;
    }

    export interface SequentialState {
        sequence: bigint;
        value: bigint;
    }

    export interface TurnBasedState {
        data: any;
        turn: number;
        status: number;
        isComplete: boolean;
    }
}

/**
 * Channel metadata (V1)
 * Represents off-chain context or cached information about a channel.
 */
export interface Metadata {
    definition: ChannelDefinition; // The channel definition
    challengeExpire?: bigint; // Optional: Calculated expiry timestamp based on last challenge
    lastValidState?: State; // Optional: The last known valid state off-chain
}

/**
 * Generic application logic interface
 */
export interface AppLogic<T = unknown> {
    /**
     * Encode application data to bytes (Hex string)
     * @param data Application-specific data structure
     * @returns Hex-encoded data for the State.metadata field
     */
    encode: (data: T) => Hex;

    /**
     * Decode application data from bytes (Hex string)
     * @param encoded Hex-encoded data from the State.metadata field
     * @returns Application-specific data structure
     */
    decode: (encoded: Hex) => T;

    /**
     * Validate a state transition based on application logic.
     * @param definition The channel definition.
     * @param prevState The application-specific data of the previous state.
     * @param nextState The application-specific data of the next state to validate.
     * @returns Whether the transition is valid according to app rules.
     */
    validateTransition?: (definition: ChannelDefinition, prevState: T, nextState: T) => boolean;

    /**
     * Define what historical states (proofs) are needed for an on-chain operation (e.g., challenge, close).
     * @param definition The channel definition.
     * @param state The application-specific data of the state requiring proofs.
     * @param previousStates Array of historical full State objects.
     * @returns Array of full State objects required as proofs.
     */
    provideProofs?: (definition: ChannelDefinition, state: T, previousStates: State[]) => State[];

    /**
     * Check if the application state represents a final or terminal condition.
     * This often corresponds to setting State.isFinal = true.
     * @param state Application-specific state data.
     * @returns Whether the state is final according to app rules.
     */
    isFinal?: (state: T) => boolean;

    /**
     * Get the specific adjudicator contract address associated with this application logic.
     * @returns Contract address of the adjudicator.
     */
    getAdjudicatorAddress: () => Address;

    /**
     * Get adjudicator type identifier (optional, for potential future use).
     * @returns String identifier for the adjudicator type.
     */
    getAdjudicatorType?: () => string;
}

/**
 * Application configuration used when initializing or interacting with channels.
 */
export interface AppConfig<T = unknown> {
    /**
     * Application-specific logic implementation.
     */
    appLogic: AppLogic<T>;

    /**
     * Initial application state data (used for creating the initial State.appData).
     */
    initialState?: T;
}
