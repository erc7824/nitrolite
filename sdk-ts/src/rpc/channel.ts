import { Address, Hex } from 'viem';
import { ChannelId, Channel, State, StateHash } from '../types';
import { getChannelId, getStateHash, verifySignature } from '../utils';
import { RPCClient } from './client';
import { RPCMethod, RPCNotificationType } from './types';

/**
 * Class for managing a channel via RPC
 */
export class RPCChannelManager {
  private rpcClient: RPCClient;
  private channels: Map<ChannelId, {
    channel: Channel,
    currentState?: State,
    counterparty: Address
  }> = new Map();
  
  /**
   * Initialize the RPC Channel Manager
   * @param rpcClient RPC client for communication
   */
  constructor(rpcClient: RPCClient) {
    this.rpcClient = rpcClient;
    
    // Register for state update notifications
    this.rpcClient.on(RPCNotificationType.STATE_UPDATE, this.handleStateUpdate.bind(this));
    this.rpcClient.on(RPCNotificationType.CHALLENGE_STARTED, this.handleChallengeNotification.bind(this));
    this.rpcClient.on(RPCNotificationType.CHANNEL_CLOSED, this.handleChannelClosedNotification.bind(this));
    
    // Register RPC methods for handling incoming requests
    this.registerRPCMethods();
  }
  
  /**
   * Register a channel for tracking
   * @param channel Channel configuration
   * @param counterparty Address of the counterparty
   * @param initialState Optional initial state
   * @returns Channel ID
   */
  registerChannel(channel: Channel, counterparty: Address, initialState?: State): ChannelId {
    const channelId = getChannelId(channel);
    
    // Store the channel
    this.channels.set(channelId, {
      channel,
      currentState: initialState,
      counterparty
    });
    
    return channelId;
  }
  
  /**
   * Unregister a channel
   * @param channelId Channel ID
   */
  unregisterChannel(channelId: ChannelId): void {
    this.channels.delete(channelId);
  }
  
  /**
   * Get a registered channel
   * @param channelId Channel ID
   * @returns The channel or undefined if not found
   */
  getChannel(channelId: ChannelId): { channel: Channel, currentState?: State, counterparty: Address } | undefined {
    return this.channels.get(channelId);
  }
  
  /**
   * Get all registered channel IDs
   * @returns Array of channel IDs
   */
  getChannelIds(): ChannelId[] {
    return Array.from(this.channels.keys());
  }
  
  /**
   * Update the state of a channel
   * @param channelId Channel ID
   * @param newState New state
   * @returns Promise that resolves with the response state (signed by counterparty)
   */
  async updateState(channelId: ChannelId, newState: State): Promise<State> {
    const channelInfo = this.channels.get(channelId);
    
    if (!channelInfo) {
      throw new Error(`Channel ${channelId} not registered`);
    }
    
    // Send the state update to the counterparty
    const signedState = await this.rpcClient.sendStateUpdate(
      channelInfo.counterparty,
      channelId,
      newState
    );
    
    // Update the current state
    this.channels.set(channelId, {
      ...channelInfo,
      currentState: signedState
    });
    
    return signedState;
  }
  
  /**
   * Request a signature for a state
   * @param channelId Channel ID
   * @param state State to sign
   * @returns Promise that resolves with the signed state
   */
  async requestSignature(channelId: ChannelId, state: State): Promise<State> {
    const channelInfo = this.channels.get(channelId);
    
    if (!channelInfo) {
      throw new Error(`Channel ${channelId} not registered`);
    }
    
    // Request signature from the counterparty
    const signedState = await this.rpcClient.requestStateSignature(
      channelInfo.counterparty,
      channelId,
      state
    );
    
    return signedState;
  }
  
  /**
   * Register RPC methods for handling incoming requests
   */
  private registerRPCMethods(): void {
    // Register UPDATE_STATE method
    this.rpcClient.registerMethod(RPCMethod.UPDATE_STATE, async (params, from) => {
      const [channelId, state] = params;
      
      // Try to handle the state update
      const result = await this.handleStateUpdateRequest(channelId, state, from);
      
      return [result];
    });
    
    // Register SIGN_STATE method
    this.rpcClient.registerMethod(RPCMethod.SIGN_STATE, async (params, from) => {
      const [channelId, state] = params;
      
      // Try to handle the signature request
      const result = await this.handleStateSignRequest(channelId, state, from);
      
      return [result];
    });
    
    // Register GET_CHANNELS method
    this.rpcClient.registerMethod(RPCMethod.GET_CHANNELS, async () => {
      // Return a list of all channels
      return [this.getChannelIds()];
    });
  }
  
  /**
   * Handle an incoming state update request
   * @param channelId Channel ID
   * @param state Proposed state
   * @param from Sender address
   * @returns Signed state if accepted
   */
  private async handleStateUpdateRequest(channelId: ChannelId, state: State, from: Address): Promise<State> {
    const channelInfo = this.channels.get(channelId);
    
    if (!channelInfo) {
      throw new Error(`Channel ${channelId} not registered`);
    }
    
    // Make sure the sender is the counterparty
    if (from.toLowerCase() !== channelInfo.counterparty.toLowerCase()) {
      throw new Error(`Unauthorized state update from ${from}`);
    }
    
    // Verify the state is valid
    // This would call into validation logic from the ChannelContext
    
    // For now, just accept the state
    const signedState = state;
    
    // TODO: Add signature to the state
    
    // Update the current state
    this.channels.set(channelId, {
      ...channelInfo,
      currentState: signedState
    });
    
    return signedState;
  }
  
  /**
   * Handle an incoming state signature request
   * @param channelId Channel ID
   * @param state State to sign
   * @param from Sender address
   * @returns Signed state
   */
  private async handleStateSignRequest(channelId: ChannelId, state: State, from: Address): Promise<State> {
    const channelInfo = this.channels.get(channelId);
    
    if (!channelInfo) {
      throw new Error(`Channel ${channelId} not registered`);
    }
    
    // Make sure the sender is the counterparty
    if (from.toLowerCase() !== channelInfo.counterparty.toLowerCase()) {
      throw new Error(`Unauthorized signature request from ${from}`);
    }
    
    // Verify the state is valid
    // This would call into validation logic from the ChannelContext
    
    // For now, just accept the state
    const signedState = state;
    
    // TODO: Add signature to the state
    
    return signedState;
  }
  
  /**
   * Handle an incoming state update notification
   * @param data Notification data
   * @param from Sender address
   */
  private async handleStateUpdate(data: any[], from: Address): Promise<void> {
    const [channelId, state] = data;
    
    const channelInfo = this.channels.get(channelId);
    
    if (!channelInfo) {
      console.warn(`Received state update for unknown channel ${channelId}`);
      return;
    }
    
    // Make sure the sender is the counterparty
    if (from.toLowerCase() !== channelInfo.counterparty.toLowerCase()) {
      console.warn(`Received state update from unauthorized sender ${from}`);
      return;
    }
    
    // Update the current state
    this.channels.set(channelId, {
      ...channelInfo,
      currentState: state
    });
    
    // Emit an event
    this.rpcClient.emit('channel:state_updated', channelId, state);
  }
  
  /**
   * Handle an incoming challenge notification
   * @param data Notification data
   * @param from Sender address
   */
  private async handleChallengeNotification(data: any[], from: Address): Promise<void> {
    const [channelId, expirationTime, challengeState] = data;
    
    const channelInfo = this.channels.get(channelId);
    
    if (!channelInfo) {
      console.warn(`Received challenge notification for unknown channel ${channelId}`);
      return;
    }
    
    // Make sure the sender is the counterparty
    if (from.toLowerCase() !== channelInfo.counterparty.toLowerCase()) {
      console.warn(`Received challenge notification from unauthorized sender ${from}`);
      return;
    }
    
    // Emit an event
    this.rpcClient.emit('channel:challenged', channelId, expirationTime, challengeState);
  }
  
  /**
   * Handle an incoming channel closed notification
   * @param data Notification data
   * @param from Sender address
   */
  private async handleChannelClosedNotification(data: any[], from: Address): Promise<void> {
    const [channelId, finalState] = data;
    
    const channelInfo = this.channels.get(channelId);
    
    if (!channelInfo) {
      console.warn(`Received closure notification for unknown channel ${channelId}`);
      return;
    }
    
    // Make sure the sender is the counterparty
    if (from.toLowerCase() !== channelInfo.counterparty.toLowerCase()) {
      console.warn(`Received closure notification from unauthorized sender ${from}`);
      return;
    }
    
    // Update the current state to the final state
    this.channels.set(channelId, {
      ...channelInfo,
      currentState: finalState
    });
    
    // Emit an event
    this.rpcClient.emit('channel:closed', channelId, finalState);
  }
}
