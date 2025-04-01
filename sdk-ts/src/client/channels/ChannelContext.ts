import {
  Address,
  Hex,
} from 'viem';
import {
  Channel,
  State,
  ChannelId,
  Role,
  AppLogic,
  Allocation
} from '../../types';
import { 
  getChannelId
} from '../../utils';
import type { HachiClient } from '../HachiClient';

/**
 * Channel context for managing application state
 */
export class ChannelContext<T = unknown> {
  readonly channel: Channel;
  readonly channelId: ChannelId;
  readonly client: HachiClient;
  readonly appLogic: AppLogic<T>;
  
  private currentState?: State;
  private appState?: T;
  private role: Role;

  /**
   * Create a new channel context
   */
  constructor(
    client: HachiClient,
    channel: Channel,
    appLogic: AppLogic<T>,
    initialAppState?: T
  ) {
    this.client = client;
    this.channel = channel;
    this.channelId = getChannelId(channel);
    this.appLogic = appLogic;
    this.appState = initialAppState;
    
    // Determine role
    const accountAddress = client.account?.address;
    if (accountAddress === channel.participants[0]) {
      this.role = Role.HOST;
    } else if (accountAddress === channel.participants[1]) {
      this.role = Role.GUEST;
    } else {
      throw new Error('Account is not a participant in this channel');
    }
  }

  /**
   * Get the channel configuration
   */
  getChannel(): Channel {
    return this.channel;
  }

  /**
   * Get the channel ID
   */
  getChannelId(): ChannelId {
    return this.channelId;
  }

  /**
   * Get the current state
   */
  getCurrentState(): State | undefined {
    return this.currentState;
  }

  /**
   * Get the current application state
   */
  getCurrentAppState(): T | undefined {
    return this.appState;
  }

  /**
   * Get the role of the current account
   */
  getRole(): Role {
    return this.role;
  }

  /**
   * Get the other participant's address
   */
  getOtherParticipant(): Address {
    return this.channel.participants[this.role === Role.HOST ? Role.GUEST : Role.HOST];
  }

  /**
   * Create a channel state based on application state
   */
  private async createChannelState(tokenAddress: Address, amounts: [bigint, bigint]): Promise<State> {
    if (!this.appState) {
      throw new Error('Application state not initialized');
    }
    
    // Encode the app state
    const data = this.appLogic.encode(this.appState);
    
    // Create allocations
    const allocations: [Allocation, Allocation] = [
      {
        destination: this.channel.participants[0],
        token: tokenAddress,
        amount: amounts[0]
      },
      {
        destination: this.channel.participants[1],
        token: tokenAddress,
        amount: amounts[1]
      }
    ];
    
    // Create state (without signatures)
    return {
      data,
      allocations,
      sigs: []
    };
  }

  /**
   * Open a channel with initial funding
   */
  async open(tokenAddress: Address, amounts: [bigint, bigint]): Promise<ChannelId> {
    if (!this.appState) {
      throw new Error('Application state not initialized');
    }
    
    // Create initial state
    const initialState = await this.createChannelState(tokenAddress, amounts);
    
    // Save as current state
    this.currentState = initialState;
    
    // Open the channel
    return this.client.openChannel(this.channel, initialState);
  }

  /**
   * Update the application state
   */
  async updateAppState(newAppState: T): Promise<State> {
    if (!this.currentState) {
      throw new Error('Channel state not initialized');
    }
    
    // Validate state transition if the app logic provides a validator
    if (this.appLogic.validateTransition && this.appState) {
      const isValid = this.appLogic.validateTransition(
        this.appState,
        newAppState,
        this.client.account?.address || '0x0000000000000000000000000000000000000000'
      );
      
      if (!isValid) {
        throw new Error('Invalid state transition');
      }
    }
    
    // Update the app state
    this.appState = newAppState;
    
    // Encode the app state
    const data = this.appLogic.encode(newAppState);
    
    // Create new state with existing allocations
    const newState: State = {
      data,
      allocations: this.currentState.allocations,
      sigs: [] // Will be filled when signing
    };
    
    // Update the channel state
    this.currentState = newState;
    
    return newState;
  }

  /**
   * Check if the current state is final
   */
  isFinal(): boolean {
    if (!this.appState || !this.appLogic.isFinal) {
      return false;
    }
    
    return this.appLogic.isFinal(this.appState);
  }

  /**
   * Close the channel with the current state
   */
  async close(proofs: State[] = []): Promise<void> {
    if (!this.currentState) {
      throw new Error('No current state to close with');
    }
    
    return this.client.closeChannel(this.channelId, this.currentState, proofs);
  }

  /**
   * Challenge the channel with the current state
   */
  async challenge(proofs: State[] = []): Promise<void> {
    if (!this.currentState) {
      throw new Error('No current state to challenge with');
    }
    
    return this.client.challengeChannel(this.channelId, this.currentState, proofs);
  }

  /**
   * Checkpoint the current state
   */
  async checkpoint(proofs: State[] = []): Promise<void> {
    if (!this.currentState) {
      throw new Error('No current state to checkpoint');
    }
    
    return this.client.checkpointChannel(this.channelId, this.currentState, proofs);
  }

  /**
   * Reclaim funds after challenge period expires
   */
  async reclaim(): Promise<void> {
    return this.client.reclaimChannel(this.channelId);
  }

  /**
   * Process a state received from the other participant
   */
  async processReceivedState(receivedState: State): Promise<boolean> {
    if (!this.currentState || !this.appState) {
      // First state in the channel
      this.currentState = receivedState;
      this.appState = this.appLogic.decode(receivedState.data);
      return true;
    }
    
    // Decode app states
    const currentAppState = this.appState;
    const receivedAppState = this.appLogic.decode(receivedState.data);
    
    // Validate state transition if the app logic provides a validator
    if (this.appLogic.validateTransition) {
      const otherParticipant = this.getOtherParticipant();
      const isValid = this.appLogic.validateTransition(
        currentAppState,
        receivedAppState,
        otherParticipant
      );
      
      if (!isValid) {
        return false;
      }
    }
    
    // Update our state
    this.appState = receivedAppState;
    this.currentState = receivedState;
    
    return true;
  }
}