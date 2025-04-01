/**
 * Virtual channel operations
 */

import { Address } from 'viem';
import { LightVirtualChannelIdentifier, VirtualChannelState } from '../types';
import { State } from '../../types';
import { RPCMethod } from '../types';
import { LVCI } from '../virtual';
import Errors from '../../errors';
import { Logger } from '../../config';
import { validateRequiredParams } from '../utils';

/**
 * Virtual channel helper that handles routing and operations
 */
export class VirtualChannelHelper {
  private logger: Logger;
  private address: Address;
  private sendRequest: <T extends any[] = any[]>(
    recipient: Address, 
    method: string, 
    params: any[]
  ) => Promise<T>;

  constructor(
    logger: Logger,
    address: Address,
    sendRequest: <T extends any[] = any[]>(
      recipient: Address, 
      method: string, 
      params: any[]
    ) => Promise<T>
  ) {
    this.logger = logger;
    this.address = address;
    this.sendRequest = sendRequest;
  }

  /**
   * Create a virtual channel through intermediaries
   * @param lvci The light virtual channel identifier
   * @param state Initial state
   * @returns Promise that resolves with the created virtual channel state
   */
  async createVirtualChannel(
    lvci: LightVirtualChannelIdentifier, 
    state: State
  ): Promise<VirtualChannelState> {
    // Validate parameters
    validateRequiredParams({ lvci, state });
    
    // Make sure we're part of the channel
    if (!LVCI.isParticipant(lvci, this.address)) {
      throw new Errors.UnauthorizedError(
        'Local address is not a participant in the virtual channel',
        { address: this.address, lvci: LVCI.toString(lvci) }
      );
    }
    
    this.logger.debug('Creating virtual channel', { 
      lvci: LVCI.toString(lvci),
      participantCount: LVCI.getPath(lvci).length
    });
    
    // Get the next hop
    const nextHop = LVCI.getNextHop(lvci, this.address);
    
    if (!nextHop) {
      throw new Errors.NoNextHopError(this.address, {
        lvci: LVCI.toString(lvci),
        position: LVCI.getPosition(lvci, this.address)
      });
    }
    
    try {
      // Send the create request to the next hop
      const result = await this.sendRequest<[VirtualChannelState]>(
        nextHop,
        RPCMethod.CREATE_VIRTUAL_CHANNEL,
        [lvci, state]
      );
      
      this.logger.info('Virtual channel created', { 
        lvci: LVCI.toString(lvci),
        nextHop
      });
      
      return result[0];
    } catch (error) {
      this.logger.error('Failed to create virtual channel', { 
        lvci: LVCI.toString(lvci),
        nextHop,
        error
      });
      
      throw new Errors.VirtualChannelError(
        `Failed to create virtual channel: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'VIRTUAL_CHANNEL_CREATE_FAILED',
        500,
        'Check that all participants in the path are online and properly configured',
        { cause: error, lvci: LVCI.toString(lvci), nextHop }
      );
    }
  }
  
  /**
   * Relay a state update through a virtual channel
   * @param lvci The light virtual channel identifier
   * @param state The new state
   * @returns Promise that resolves with the relayed state (with signatures)
   */
  async relayStateUpdate(
    lvci: LightVirtualChannelIdentifier, 
    state: State
  ): Promise<VirtualChannelState> {
    // Validate parameters
    validateRequiredParams({ lvci, state });
    
    // Make sure we're part of the channel
    if (!LVCI.isParticipant(lvci, this.address)) {
      throw new Errors.UnauthorizedError(
        'Local address is not a participant in the virtual channel',
        { address: this.address, lvci: LVCI.toString(lvci) }
      );
    }
    
    // Get the position of this node in the virtual channel path
    const position = LVCI.getPosition(lvci, this.address);
    const path = LVCI.getPath(lvci);
    
    this.logger.debug('Relaying state update in virtual channel', { 
      lvci: LVCI.toString(lvci),
      position,
      pathLength: path.length
    });
    
    // Determine if we're at an endpoint (origin or destination)
    const isOrigin = position === 0;
    const isDestination = position === path.length - 1;
    
    // Implement virtual channel relay protocol:
    // 1. Updates flow from origin toward destination first
    // 2. Once reached destination, they flow back toward origin
    // 3. This creates a full round-trip ensuring all nodes have the latest state
    
    // Relay logic based on position and state flow direction
    let forward: boolean;
    
    if (isDestination) {
      // We're at the destination, relay back toward the origin
      forward = false;
      
      this.logger.debug('At destination - relaying back toward origin', {
        lvci: LVCI.toString(lvci),
        position
      });
    } else if (isOrigin) {
      // We're at the origin, relay toward the destination
      forward = true;
      
      this.logger.debug('At origin - relaying toward destination', {
        lvci: LVCI.toString(lvci),
        position
      });
    } else {
      // We're an intermediary node
      // Check if we're on a state update's path toward the destination
      // or on the return path toward the origin
      // Approach: We determine this from the state's metadata or from configuration
      
      // Get the isInbound flag from state metadata if available
      // Otherwise, default to forwarding toward destination
      const isInbound = (state as any)?.metadata?.isInbound === true;
      
      if (isInbound) {
        // State is flowing back from destination to origin
        forward = false;
        this.logger.debug('State flowing back toward origin', {
          lvci: LVCI.toString(lvci),
          position
        });
      } else {
        // State is flowing from origin to destination
        forward = true;
        this.logger.debug('State flowing toward destination', {
          lvci: LVCI.toString(lvci),
          position
        });
      }
    }
    const nextHop = LVCI.getNextHop(lvci, this.address, forward);
    
    if (!nextHop) {
      throw new Errors.NoNextHopError(this.address, {
        lvci: LVCI.toString(lvci),
        position,
        forward
      });
    }
    
    try {
      // Send the relay request to the next hop
      const result = await this.sendRequest<[VirtualChannelState]>(
        nextHop,
        RPCMethod.RELAY_STATE_UPDATE,
        [lvci, state]
      );
      
      this.logger.debug('State update relayed successfully', { 
        lvci: LVCI.toString(lvci),
        nextHop
      });
      
      return result[0];
    } catch (error) {
      this.logger.error('Failed to relay state update', { 
        lvci: LVCI.toString(lvci),
        nextHop,
        error
      });
      
      throw new Errors.RelayError(
        `Failed to relay state update: ${error instanceof Error ? error.message : 'Unknown error'}`,
        { cause: error, lvci: LVCI.toString(lvci), nextHop }
      );
    }
  }
}
