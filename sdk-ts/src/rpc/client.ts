import { Address, Hex, hashMessage } from 'viem';
import { EventEmitter } from 'events';
import { 
  RPCProvider, 
  RPCMessage, 
  RPCRequest, 
  RPCResponse, 
  RPCError, 
  RPCNotification, 
  RPCMessageType, 
  RPCMethodHandler, 
  RPCMethod, 
  RPCNotificationType 
} from './types';
import { LightVirtualChannelIdentifier, VirtualChannelState } from './types';
import { ChannelId, State } from '../types';
import { 
  DEFAULT_CONFIG, 
  RPC_ERROR_CODES, 
  SDKConfig, 
  getConfigWithDefaults, 
  Logger, 
  defaultLogger, 
  createFilteredLogger 
} from '../config';
import Errors from '../errors';
import { 
  RequestHandler, 
  ResponseHandler, 
  ErrorHandler, 
  NotificationHandler 
} from './handlers';
import { createPayload, verifySignature, validateConnection, validateRequiredParams } from './utils';
import { VirtualChannelHelper } from './channels/virtual';

/**
 * Configuration for the RPC Client
 * TODO: Ensure documentation clearly explains all options, including deprecated ones.
 */
export interface RPCClientConfig extends SDKConfig {
  /**
   * Provider for transport
   */
  provider: RPCProvider;
  
  /**
   * Local address
   */
  address: Address;
  
  /**
   * Signer function to sign messages
   */
  signer: (message: Hex) => Promise<Hex>;
  
  /**
   * Timeout for requests in milliseconds
   * @deprecated Use requestTimeoutMs instead
   */
  requestTimeout?: number;
  
  /**
   * Maximum number of retries for failed requests
   * @deprecated Use maxRequestRetries instead
   */
  maxRetries?: number;
}

/**
 * Pending request structure
 */
type PendingRequest = {
  resolve: (result: any[]) => void;
  reject: (error: Error) => void;
  timeout: NodeJS.Timeout;
  retries: number;
};

/**
 * RPC Client for off-chain communication
 * 
 * Implements the Nitro RPC protocol for communicating between participants
 * in a state channel network.
 */
export class RPCClient extends EventEmitter {
  private provider: RPCProvider;
  private address: Address;
  private signer: (message: Hex) => Promise<Hex>;
  private requestTimeout: number;
  private maxRetries: number;
  
  private methods: Map<string, RPCMethodHandler> = new Map();
  private pendingRequests: Map<number, PendingRequest> = new Map();
  
  private nextRequestId: number = 1;
  private lastServerTime: number = Date.now();
  private unregisterHandler: (() => void) | null = null;
  
  private logger: Logger;
  
  // Specialized handlers
  private requestHandler: RequestHandler;
  private responseHandler: ResponseHandler;
  private errorHandler: ErrorHandler;
  private notificationHandler: NotificationHandler;
  
  // Virtual channel helper
  private virtualChannelHelper: VirtualChannelHelper;
  
  /**
   * Create a new RPC Client
   * @param config Configuration options
   */
  constructor(config: RPCClientConfig) {
    super();
    
    // Apply defaults to configuration
    const fullConfig = getConfigWithDefaults(config);
    
    this.provider = config.provider;
    this.address = config.address;
    this.signer = config.signer;
    
    // Handle deprecated configuration options with fallbacks
    this.requestTimeout = config.requestTimeout || fullConfig.requestTimeoutMs;
    this.maxRetries = config.maxRetries || fullConfig.maxRequestRetries;
    
    // Create logger instance
    this.logger = fullConfig.logger || defaultLogger;
    
    // Use filtered logger if log level is specified
    if (fullConfig.logLevel) {
      this.logger = createFilteredLogger(fullConfig.logLevel, this.logger);
    }
    
    // Initialize message handlers
    this.requestHandler = new RequestHandler(
      this.logger,
      this.methods,
      this.signer,
      this.address,
      this.createPayload.bind(this),
      this.verifySignature.bind(this),
      this.sendResponse.bind(this),
      this.sendErrorResponse.bind(this)
    );
    
    this.responseHandler = new ResponseHandler(
      this.logger,
      this.pendingRequests
    );
    
    this.errorHandler = new ErrorHandler(
      this.logger,
      this.pendingRequests
    );
    
    this.notificationHandler = new NotificationHandler(
      this.logger,
      this.verifySignature.bind(this),
      this.createPayload.bind(this),
      this.emitEvent.bind(this)
    );
    
    // Initialize virtual channel helper
    this.virtualChannelHelper = new VirtualChannelHelper(
      this.logger,
      this.address,
      this.sendRequest.bind(this)
    );
    
    // Register standard methods
    this.registerMethod(RPCMethod.PING, async () => ["pong"]);
    this.registerMethod(RPCMethod.GET_TIME, async () => [this.lastServerTime]);
    
    this.logger.debug('RPC Client initialized', {
      address: this.address,
      requestTimeout: this.requestTimeout,
      maxRetries: this.maxRetries
    });
  }
  
  /**
   * Connect to the provider
   */
  async connect(): Promise<void> {
    try {
      await this.provider.connect();
      
      // Register message handler
      this.unregisterHandler = this.provider.onMessage(this.handleMessage.bind(this));
      
      this.logger.info('Connected to provider', { address: this.address });
    } catch (error) {
      this.logger.error('Failed to connect to provider', { error });
      throw new Errors.ConnectionError('Failed to connect to RPC provider', { 
        cause: error, 
        providerType: this.provider.constructor.name 
      });
    }
  }
  
  /**
   * Disconnect from the provider
   */
  async disconnect(): Promise<void> {
    try {
      // Unregister message handler
      if (this.unregisterHandler) {
        this.unregisterHandler();
        this.unregisterHandler = null;
      }
      
      // Clear all pending requests
      for (const [id, { timeout, reject }] of this.pendingRequests.entries()) {
        clearTimeout(timeout);
        reject(new Errors.ConnectionError('Client disconnected'));
        this.pendingRequests.delete(id);
      }
      
      await this.provider.disconnect();
      
      this.logger.info('Disconnected from provider', { address: this.address });
    } catch (error) {
      this.logger.error('Error during disconnect', { error });
      throw new Errors.ConnectionError('Failed to disconnect from RPC provider', { 
        cause: error, 
        providerType: this.provider.constructor.name 
      });
    }
  }
  
  /**
   * Register a method handler
   * @param method Method name
   * @param handler Handler function
   */
  registerMethod(method: string, handler: RPCMethodHandler): void {
    this.methods.set(method, handler);
    this.logger.debug(`Method registered: ${method}`);
  }
  
  /**
   * Unregister a method handler
   * @param method Method name
   */
  unregisterMethod(method: string): void {
    this.methods.delete(method);
    this.logger.debug(`Method unregistered: ${method}`);
  }
  
  /**
   * Send a request to a recipient
   * @param recipient Recipient address
   * @param method Method name
   * @param params Method parameters
   * @returns Promise that resolves with the result
   */
  async sendRequest<T extends any[] = any[]>(
    recipient: Address, 
    method: string, 
    params: any[] = []
  ): Promise<T> {
    // Ensure provider is connected
    validateConnection(!!this.unregisterHandler);
    
    // Validate parameters
    validateRequiredParams({ recipient, method });
    
    const requestId = this.nextRequestId++;
    const timestamp = this.lastServerTime || Date.now();
    
    // Create the request message
    const request: RPCRequest = {
      type: RPCMessageType.REQUEST,
      ts: timestamp,
      req: [requestId, method, params, timestamp]
    };
    
    try {
      // Sign the request payload
      const payload = this.createPayload(request.req);
      request.sig = await this.signer(payload);
      
      this.logger.debug('Sending request', { 
        requestId, 
        method, 
        recipient, 
        paramsLength: params.length 
      });
      
      // Send the request
      return new Promise<T>((resolve, reject) => {
        // Set timeout
        const timeout = setTimeout(() => {
          const pendingRequest = this.pendingRequests.get(requestId);
          if (!pendingRequest) return;
          
          // Check if we should retry
          if (pendingRequest.retries < this.maxRetries) {
            // Clear from pending requests
            this.pendingRequests.delete(requestId);
            
            // Increment retry counter and logging
            const retryCount = pendingRequest.retries + 1;
            this.logger.warn(`Request timed out, retrying (${retryCount}/${this.maxRetries})`, {
              requestId,
              method,
              recipient
            });
            
            // Retry the request
            this.sendRequest<T>(recipient, method, params)
              .then(resolve)
              .catch(reject);
          } else {
            // No more retries, reject with timeout
            this.pendingRequests.delete(requestId);
            this.logger.error('Request timed out after all retries', {
              requestId,
              method,
              recipient,
              retries: this.maxRetries
            });
            
            reject(new Errors.RequestTimeoutError(
              `Request timed out for method '${method}'`, 
              this.maxRetries,
              { requestId, method, recipient }
            ));
          }
        }, this.requestTimeout);
        
        // Store the pending request
        this.pendingRequests.set(requestId, {
          resolve: resolve as any,
          reject,
          timeout,
          retries: 0
        });
        
        // Send the request
        this.provider.send(recipient, request)
          .catch(error => {
            // Clear the timeout
            clearTimeout(timeout);
            
            // Clear from pending requests
            this.pendingRequests.delete(requestId);
            
            this.logger.error('Failed to send request', { 
              requestId, 
              method, 
              recipient, 
              error 
            });
            
            // Reject with appropriate error type
            reject(new Errors.ConnectionError(
              `Failed to send request for method '${method}'`,
              { cause: error, requestId, method, recipient }
            ));
          });
      });
    } catch (error) {
      this.logger.error('Error preparing request', { 
        requestId, 
        method, 
        recipient, 
        error 
      });
      
      throw new Errors.ConnectionError(
        `Failed to prepare request for method '${method}'`,
        { cause: error, requestId, method, recipient }
      );
    }
  }
  
  /**
   * Send a notification to a recipient
   * @param recipient Recipient address
   * @param type Notification type
   * @param data Notification data
   */
  async sendNotification(recipient: Address, type: string, data: any[] = []): Promise<void> {
    // Ensure provider is connected
    validateConnection(!!this.unregisterHandler);
    
    // Validate parameters
    validateRequiredParams({ recipient, type });
    
    const timestamp = this.lastServerTime || Date.now();
    
    try {
      // Create the notification message
      const notification: RPCNotification = {
        type: RPCMessageType.NOTIFICATION,
        ts: timestamp,
        ntf: [type, data, timestamp]
      };
      
      // Sign the notification payload
      const payload = this.createPayload(notification.ntf);
      notification.sig = await this.signer(payload);
      
      this.logger.debug('Sending notification', { 
        type, 
        recipient, 
        dataLength: data.length 
      });
      
      // Send the notification
      await this.provider.send(recipient, notification);
    } catch (error) {
      this.logger.error('Failed to send notification', { 
        type, 
        recipient, 
        error 
      });
      
      throw new Errors.ConnectionError(
        `Failed to send notification of type '${type}'`,
        { cause: error, type, recipient }
      );
    }
  }
  
  /**
   * Emit an event based on notification
   * @param type Event type
   * @param data Event data
   * @param from Source address
   */
  private emitEvent(type: string, data: any[], from: Address): void {
    // Emit an event for the notification
    this.emit(type, data, from);
    
    // Emit a combined event for all notifications
    this.emit('notification', type, data, from);
  }
  
  /**
   * Handle an incoming message
   * @param from Sender address
   * @param message The message
   */
  private async handleMessage(from: Address, message: RPCMessage): Promise<void> {
    try {
      // Update the server time
      if (message.ts && message.ts > this.lastServerTime) {
        this.lastServerTime = message.ts;
      }
      
      this.logger.debug('Received message', { 
        type: message.type, 
        from,
        timestamp: message.ts
      });
      
      // Handle by message type
      switch (message.type) {
        case RPCMessageType.REQUEST:
          await this.requestHandler.handle(from, message as RPCRequest);
          break;
          
        case RPCMessageType.RESPONSE:
          await this.responseHandler.handle(from, message as RPCResponse);
          break;
          
        case RPCMessageType.ERROR:
          await this.errorHandler.handle(from, message as RPCError);
          break;
          
        case RPCMessageType.NOTIFICATION:
          await this.notificationHandler.handle(from, message as RPCNotification);
          break;
          
        default:
          this.logger.warn('Unknown message type', { type: message.type, from });
      }
    } catch (error) {
      this.logger.error('Error handling message', { 
        from, 
        messageType: message.type, 
        error 
      });
    }
  }
  
  /**
   * Create a payload for signing
   * @param data The data to sign
   * @returns The payload as a hex string
   */
  private createPayload(data: any): Hex {
    return createPayload(data);
  }
  
  /**
   * Verify a signature
   * @param payload The payload that was signed
   * @param signature The signature
   * @param expectedSigner The expected signer address
   * @returns True if the signature is valid
   */
  private async verifySignature(
    payload: Hex, 
    signature: Hex, 
    expectedSigner: Address
  ): Promise<boolean> {
    return verifySignature(payload, signature, expectedSigner, this.logger);
  }
  
  /**
   * Send a response to a request
   * @param to Recipient address
   * @param requestId Request ID
   * @param method Method name
   * @param result Result data
   */
  private async sendResponse(
    to: Address, 
    requestId: number, 
    method: string, 
    result: any[]
  ): Promise<void> {
    try {
      const timestamp = Date.now();
      
      this.logger.debug('Sending response', { 
        requestId, 
        method, 
        to,
        resultLength: result?.length 
      });
      
      // Create the response message
      const response: RPCResponse = {
        type: RPCMessageType.RESPONSE,
        ts: timestamp,
        res: [requestId, method, result, timestamp]
      };
      
      // Sign the response payload
      const payload = this.createPayload(response.res);
      response.sig = await this.signer(payload);
      
      // Send the response
      await this.provider.send(to, response);
    } catch (error) {
      this.logger.error('Failed to send response', { 
        requestId, 
        method, 
        to, 
        error 
      });
      // No way to inform requester of this error since we're having trouble sending responses
      // Just log it and continue
    }
  }
  
  /**
   * Send an error response to a request
   * @param to Recipient address
   * @param requestId Request ID
   * @param code Error code
   * @param message Error message
   */
  private async sendErrorResponse(
    to: Address, 
    requestId: number, 
    code: number, 
    message: string
  ): Promise<void> {
    try {
      const timestamp = Date.now();
      
      this.logger.debug('Sending error response', { 
        requestId, 
        code, 
        to,
        message 
      });
      
      // Create the error message
      const error: RPCError = {
        type: RPCMessageType.ERROR,
        ts: timestamp,
        err: [requestId, code, message, timestamp]
      };
      
      // Sign the error payload
      const payload = this.createPayload(error.err);
      error.sig = await this.signer(payload);
      
      // Send the error
      await this.provider.send(to, error);
    } catch (error) {
      this.logger.error('Failed to send error response', { 
        requestId, 
        code, 
        to, 
        message, 
        error 
      });
      // No way to inform requester of this error since we're having trouble sending responses
      // Just log it and continue
    }
  }
  
  // Channel-specific RPC methods
  
  /**
   * Send a state update to another participant
   * @param recipient Recipient address
   * @param channelId Channel ID
   * @param state New state
   * @returns Promise that resolves with the response (typically a signed state)
   */
  async sendStateUpdate(recipient: Address, channelId: ChannelId, state: State): Promise<State> {
    const result = await this.sendRequest<[State]>(recipient, RPCMethod.UPDATE_STATE, [channelId, state]);
    return result[0];
  }
  
  /**
   * Request a signature for a state
   * @param recipient Recipient address
   * @param channelId Channel ID
   * @param state State to sign
   * @returns Promise that resolves with the signed state
   */
  async requestStateSignature(recipient: Address, channelId: ChannelId, state: State): Promise<State> {
    const result = await this.sendRequest<[State]>(recipient, RPCMethod.SIGN_STATE, [channelId, state]);
    return result[0];
  }
  
  /**
   * Notify a participant about a challenge
   * @param recipient Recipient address
   * @param channelId Channel ID
   * @param expirationTime When the challenge expires
   * @param challengeState The state used for the challenge
   */
  async notifyChallenge(
    recipient: Address, 
    channelId: ChannelId, 
    expirationTime: number, 
    challengeState: State
  ): Promise<void> {
    await this.sendNotification(
      recipient,
      RPCNotificationType.CHALLENGE_STARTED,
      [channelId, expirationTime, challengeState]
    );
  }
  
  /**
   * Notify a participant about a channel closure
   * @param recipient Recipient address
   * @param channelId Channel ID
   * @param finalState The final state
   */
  async notifyClosure(recipient: Address, channelId: ChannelId, finalState: State): Promise<void> {
    await this.sendNotification(
      recipient,
      RPCNotificationType.CHANNEL_CLOSED,
      [channelId, finalState]
    );
  }
  
  // Virtual channel methods
  
  /**
   * Create a virtual channel through intermediaries
   * @param lvci The light virtual channel identifier
   * @param state Initial state
   * @returns Promise that resolves with the created virtual channel state
   */
  async createVirtualChannel(lvci: LightVirtualChannelIdentifier, state: State): Promise<VirtualChannelState> {
    return this.virtualChannelHelper.createVirtualChannel(lvci, state);
  }
  
  /**
   * Relay a state update through a virtual channel
   * @param lvci The light virtual channel identifier
   * @param state The new state
   * @returns Promise that resolves with the relayed state (with signatures)
   */
  async relayStateUpdate(lvci: LightVirtualChannelIdentifier, state: State): Promise<VirtualChannelState> {
    return this.virtualChannelHelper.relayStateUpdate(lvci, state);
  }
}
