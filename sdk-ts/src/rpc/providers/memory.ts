import { Address } from 'viem';
import { RPCProvider, RPCMessage } from '../types';
import Errors from '../../errors';
import { SDKConfig, getConfigWithDefaults, Logger, defaultLogger, createFilteredLogger } from '../../config';

/**
 * Configuration for MemoryRPCProvider
 */
export interface MemoryProviderConfig extends SDKConfig {
  /**
   * Optional delay to simulate network latency in milliseconds
   * Default: 0
   */
  simulatedLatencyMs?: number;
}

/**
 * In-memory provider for RPC communication
 * Useful for testing and examples
 */
export class MemoryRPCProvider implements RPCProvider {
  // Map of address to handler function
  private handlers: Map<string, (from: Address, message: RPCMessage) => void> = new Map();
  
  // Network of providers (global static map shared between all instances)
  private static network: Map<string, MemoryRPCProvider> = new Map();
  
  private address: Address;
  private connected: boolean = false;
  private simulatedLatencyMs: number;
  private logger: Logger;
  
  /**
   * Create a new in-memory provider
   * @param address The local address
   * @param config Configuration options
   */
  constructor(address: Address, config: MemoryProviderConfig = {}) {
    // Apply defaults to configuration
    const fullConfig = getConfigWithDefaults(config);
    
    this.address = address.toLowerCase() as Address;
    this.simulatedLatencyMs = config.simulatedLatencyMs || 0;
    this.logger = fullConfig.logger || defaultLogger;
    
    // Use filtered logger if log level is specified
    if (fullConfig.logLevel) {
      this.logger = createFilteredLogger(fullConfig.logLevel, this.logger);
    }
  }
  
  /**
   * Connect to the provider network
   */
  async connect(): Promise<void> {
    // Register in the network
    MemoryRPCProvider.network.set(this.address, this);
    this.connected = true;
    
    this.logger.debug('Memory provider connected', { 
      address: this.address,
      networkSize: MemoryRPCProvider.network.size
    });
  }
  
  /**
   * Disconnect from the provider network
   */
  async disconnect(): Promise<void> {
    // Unregister from the network
    MemoryRPCProvider.network.delete(this.address);
    this.connected = false;
    
    this.logger.debug('Memory provider disconnected', { 
      address: this.address 
    });
  }
  
  /**
   * Send a message to a recipient
   * @param recipient Recipient address
   * @param message Message to send
   */
  async send(recipient: Address, message: RPCMessage): Promise<void> {
    if (!this.connected) {
      throw new Errors.ProviderNotConnectedError('Memory Provider');
    }
    
    // Validate parameters
    if (!recipient) {
      throw new Errors.MissingParameterError('recipient');
    }
    
    if (!message) {
      throw new Errors.MissingParameterError('message');
    }
    
    // Find the recipient provider
    const recipientAddr = recipient.toLowerCase();
    const recipientProvider = MemoryRPCProvider.network.get(recipientAddr);
    
    if (!recipientProvider) {
      this.logger.error('Recipient not found in network', { 
        recipient: recipientAddr,
        messageType: message.type
      });
      
      throw new Errors.ConnectionError(
        `Recipient ${recipient} not found in the network`,
        { recipient, messageType: message.type }
      );
    }
    
    // Call the recipient's handler
    const handler = recipientProvider.handlers.get(recipientAddr);
    
    if (handler) {
      this.logger.debug('Sending message', { 
        from: this.address, 
        to: recipientAddr, 
        type: message.type,
        delay: this.simulatedLatencyMs
      });
      
      // Use setTimeout to simulate async behavior and network latency
      setTimeout(() => {
        handler(this.address, message);
      }, this.simulatedLatencyMs);
    } else {
      this.logger.warn('Recipient has no handler registered', { 
        recipient: recipientAddr
      });
    }
  }
  
  /**
   * Register a handler for incoming messages
   * @param handler Handler function
   * @returns Function to unregister the handler
   */
  onMessage(handler: (from: Address, message: RPCMessage) => void): () => void {
    this.handlers.set(this.address, handler);
    
    this.logger.debug('Message handler registered', { 
      address: this.address 
    });
    
    // Return unregister function
    return () => {
      this.handlers.delete(this.address);
      this.logger.debug('Message handler unregistered', { 
        address: this.address 
      });
    };
  }
  
  /**
   * Reset the entire network
   * Useful for testing
   */
  static resetNetwork(): void {
    MemoryRPCProvider.network.clear();
  }
  
  /**
   * Get all connected addresses
   * @returns Array of addresses
   */
  static getConnectedAddresses(): Address[] {
    return Array.from(MemoryRPCProvider.network.keys()) as Address[];
  }
  
  /**
   * Get the size of the network
   * @returns Number of connected providers
   */
  static getNetworkSize(): number {
    return MemoryRPCProvider.network.size;
  }
}
