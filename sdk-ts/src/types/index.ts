import { Address, Hex } from 'viem';

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
  destination: Address;  // Where funds are sent on channel closure
  token: Address;        // ERC-20 token contract address
  amount: bigint;        // Token amount allocated
}

/**
 * Channel configuration structure
 */
export interface Channel {
  participants: [Address, Address]; // List of participants in the channel [Host, Guest]
  adjudicator: Address;             // Address of the contract that validates final states
  challenge: bigint;                // Duration in seconds for challenge period
  nonce: bigint;                    // Unique per channel with same participants and adjudicator
}

/**
 * Channel state structure
 */
export interface State {
  data: Hex;                     // Application data encoded, decoded by the adjudicator for business logic
  allocations: [Allocation, Allocation]; // Combined asset allocation and destination for each participant
  sigs: Signature[];              // stateHash signatures
}

/**
 * Metadata for tracking channel states
 */
export interface Metadata {
  chan: Channel;             // Channel configuration
  challengeExpire: bigint;   // If non-zero channel will resolve to lastValidState when challenge Expires
  lastValidState: State;     // Last valid state when adjudicator was called
}

/**
 * Adjudicator status enum
 */
export enum AdjudicatorStatus {
  VOID = 0,     // Channel was never active or have an anomaly
  PARTIAL = 1,  // Partial funding waiting for other participants
  ACTIVE = 2,   // Channel fully funded using open or state are valid
  INVALID = 3,  // Channel state is invalid
  FINAL = 4     // This is the FINAL State channel can be closed
}

/**
 * Participant roles in a channel
 */
export enum Role {
  HOST = 0,
  GUEST = 1
}

/**
 * Channel identifier
 */
export type ChannelId = Hex;

/**
 * State hash
 */
export type StateHash = Hex;

/**
 * Generic application logic interface
 */
export interface AppLogic<T = unknown> {
  /**
   * Encode application data to bytes
   * @param data Application-specific data structure
   * @returns Hex-encoded data for the state
   */
  encode: (data: T) => Hex;
  
  /**
   * Decode application data from bytes
   * @param encoded Hex-encoded data from the state
   * @returns Application-specific data structure
   */
  decode: (encoded: Hex) => T;
  
  /**
   * Validate a state transition
   * @param prevState Previous application state
   * @param nextState Next application state
   * @param signer Address of participant who signed this update
   * @returns Whether the transition is valid
   */
  validateTransition?: (prevState: T, nextState: T, signer: Address) => boolean;
  
  /**
   * Check if application state is final
   * @param state Application state
   * @returns Whether the state is final
   */
  isFinal?: (state: T) => boolean;
  
  /**
   * Get adjudicator contract address
   * @returns Contract address of the adjudicator
   */
  getAdjudicatorAddress: () => Address;
  
  /**
   * Get adjudicator type identifier (optional)
   * @returns String identifier for the adjudicator type 
   */
  getAdjudicatorType?: () => string;
}

/**
 * Application configuration for creating a new app
 */
export interface AppConfig<T = unknown> {
  /**
   * Application-specific logic
   */
  appLogic: AppLogic<T>;
  
  /**
   * Initial application state
   */
  initialState?: T;
}

/**
 * Example generic app data types (for reference only)
 */
export namespace AppDataTypes {
  // Generic app state with a numeric value
  export interface NumericState {
    value: bigint;
  }
  
  // Generic app state with sequence number and value
  export interface SequentialState {
    sequence: bigint;
    value: bigint;
  }
  
  // Generic app state with turn-based structure
  export interface TurnBasedState {
    data: unknown;
    turn: number;
    status: number;
    isComplete: boolean;
  }
}

/**
 * Channel events
 */
export interface ChannelOpenedEvent {
  channelId: ChannelId;
  channel: Channel;
}

export interface ChannelChallengedEvent {
  channelId: ChannelId;
  expiration: bigint;
}

export interface ChannelCheckpointedEvent {
  channelId: ChannelId;
}

export interface ChannelClosedEvent {
  channelId: ChannelId;
}