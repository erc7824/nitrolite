import { Address } from 'viem';
import { ChannelContext } from '../client/channels/ChannelContext';
import { AppLogic, Channel, ChannelId, State } from '../types';
import { RPCClient } from './client';
import { RPCChannelManager } from './channel';
import { RPCMethod, RPCNotificationType } from './types';

/**
 * Integrates the RPC protocol with the ChannelContext
 * 
 * This class enhances the existing ChannelContext to add RPC communication
 * capabilities, enabling off-chain state synchronization between channel participants.
 */
export class RPCChannelContext<T = unknown> extends ChannelContext<T> {
  private rpcClient: RPCClient;
  private rpcChannelManager: RPCChannelManager;
  private counterparty: Address;
  private onStateUpdateCallback?: (state: State) => void;
  
  /**
   * Create a new RPC-enabled channel context
   * @param channelContext The base channel context to enhance
   * @param rpcClient The RPC client for communication
   * @param rpcChannelManager The RPC channel manager
   */
  constructor(
    channelContext: ChannelContext<T>,
    rpcClient: RPCClient,
    rpcChannelManager: RPCChannelManager
  ) {
    // Call super with all the original parameters
    super(
      channelContext['client'],
      channelContext.getChannel(),
      channelContext['appLogic']
    );
    
    this.rpcClient = rpcClient;
    this.rpcChannelManager = rpcChannelManager;
    
    // Get the counterparty address
    this.counterparty = this.getOtherParticipant();
    
    // Register the channel with the RPC channel manager
    this.rpcChannelManager.registerChannel(
      this.getChannel(),
      this.counterparty,
      this.getCurrentState()
    );
    
    // Listen for RPC state updates for this channel
    this.listenForStateUpdates();
  }
  
  /**
   * Set a callback for state updates
   * @param callback Function to call when state is updated
   */
  onStateUpdate(callback: (state: State) => void): void {
    this.onStateUpdateCallback = callback;
  }
  
  /**
   * Override the updateAppState method to include RPC communication
   * @param newAppState New application state
   * @returns Promise with the updated state
   */
  async updateAppState(newAppState: T): Promise<State> {
    // First update local state using the parent method
    const newState = await super.updateAppState(newAppState);
    
    // Then send the state update via RPC
    try {
      const signedState = await this.rpcChannelManager.updateState(
        this.getChannelId(),
        newState
      );
      
      // Store the signed state
      // This should be handled by the channel context automatically
      
      return signedState;
    } catch (error) {
      console.error('Error sending state update via RPC:', error);
      throw error;
    }
  }
  
  /**
   * Override the close method to notify the counterparty
   * @param proofs Optional proofs for the close operation
   */
  async close(proofs: State[] = []): Promise<void> {
    // First close the channel using the parent method
    await super.close(proofs);
    
    // Then notify the counterparty
    try {
      const currentState = this.getCurrentState();
      if (currentState) {
        await this.rpcClient.notifyClosure(
          this.counterparty,
          this.getChannelId(),
          currentState
        );
      }
    } catch (error) {
      console.error('Error notifying counterparty of channel closure:', error);
      // Don't throw here, as the channel is already closed on-chain
    }
  }
  
  /**
   * Override the challenge method to notify the counterparty
   * @param proofs Optional proofs for the challenge
   */
  async challenge(proofs: State[] = []): Promise<void> {
    // First challenge the channel using the parent method
    await super.challenge(proofs);
    
    // Then notify the counterparty
    try {
      const currentState = this.getCurrentState();
      if (currentState) {
        // We'd need to get the expiration time from the transaction receipt
        // This is a placeholder value
        const expirationTime = Date.now() + (Number(this.getChannel().challenge) * 1000);
        
        await this.rpcClient.notifyChallenge(
          this.counterparty,
          this.getChannelId(),
          expirationTime,
          currentState
        );
      }
    } catch (error) {
      console.error('Error notifying counterparty of challenge:', error);
      // Don't throw here, as the challenge is already submitted on-chain
    }
  }
  
  /**
   * Setup listeners for state updates
   */
  private listenForStateUpdates(): void {
    const channelId = this.getChannelId();
    
    // Listen for state updates for this channel
    this.rpcClient.on('channel:state_updated', (updatedChannelId, state) => {
      if (updatedChannelId === channelId) {
        this.handleStateUpdate(state);
      }
    });
    
    // Listen for challenge events
    this.rpcClient.on('channel:challenged', (updatedChannelId, expirationTime, state) => {
      if (updatedChannelId === channelId) {
        this.handleChallengeNotification(expirationTime, state);
      }
    });
    
    // Listen for closure events
    this.rpcClient.on('channel:closed', (updatedChannelId, state) => {
      if (updatedChannelId === channelId) {
        this.handleClosureNotification(state);
      }
    });
  }
  
  /**
   * Handle an incoming state update
   * @param state The updated state
   */
  private async handleStateUpdate(state: State): Promise<void> {
    // Process the received state
    const success = await this.processReceivedState(state);
    
    if (success) {
      // Trigger any registered callback
      if (this.onStateUpdateCallback) {
        this.onStateUpdateCallback(state);
      }
    } else {
      console.warn('Invalid state update received from counterparty');
    }
  }
  
  /**
   * Handle a challenge notification
   * @param expirationTime When the challenge expires
   * @param state The challenged state
   */
  private handleChallengeNotification(expirationTime: number, state: State): void {
    // Process the challenged state
    // Here you would typically check if you need to respond to the challenge
    console.log(`Channel ${this.getChannelId()} was challenged, expiring at ${expirationTime}`);
  }
  
  /**
   * Handle a closure notification
   * @param state The final state
   */
  private handleClosureNotification(state: State): void {
    // Process the closure notification
    console.log(`Channel ${this.getChannelId()} was closed with final state:`, state);
  }
}

/**
 * Factory function to create an RPC-enabled channel context
 * @param channelContext The channel context to enhance
 * @param rpcClient The RPC client for communication
 * @param rpcChannelManager The RPC channel manager
 * @returns An RPC-enabled channel context
 */
export function createRPCChannelContext<T>(
  channelContext: ChannelContext<T>,
  rpcClient: RPCClient,
  rpcChannelManager: RPCChannelManager
): RPCChannelContext<T> {
  return new RPCChannelContext<T>(channelContext, rpcClient, rpcChannelManager);
}
