import { Address, Hex } from 'viem';
import { Channel, State } from '../client';

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
   * @param channel Channel in context of what, the state is being validated
   * @param prevState Previous application state
   * @param nextState Next application state
   * @returns Whether the transition is valid
   */
  validateTransition?: (channel: Channel, prevState: T, nextState: T) => boolean;

  /**
   * Define what proofs are needed for a state to be supported on SC
   * @param channel Channel in context of what, the state is being validated
   * @param state Application state, that requires proofs
   * @param previousStates Previous channel states
   * @returns Array of states that are needed to be supported on SC
   */
  provideProofs?: (channel: Channel, state: T, previousStates: State[]) => State[];
  
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
