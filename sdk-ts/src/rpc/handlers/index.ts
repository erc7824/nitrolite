/**
 * RPC Message Handlers
 * 
 * Contains handler implementations for different message types
 */

import { Address, Hex } from 'viem'; // Import Hex
import { RPCRequest, RPCResponse, RPCError as RPCTypeError, RPCNotification, RPCMessageType } from '../types'; // Alias local RPCError type
import { Logger } from '../../config';
import {
  HachiError,
  ValidationError,
  InvalidSignatureError,
  UnauthorizedError,
  StateError,
  ChannelNotFoundError,
  MethodNotFoundError,
  TimeoutError,
  VirtualChannelError,
  InvalidRPCParamsError,
  RequestTimeoutError,
  RPCError as HachiRPCError // Alias to avoid conflict
} from '../../errors';
import { RPC_ERROR_CODES } from '../../config';

/**
 * Interface for message handlers
 */
export interface MessageHandler {
  /**
   * Handle a message and return a result
   */
  handle(from: Address, message: any): Promise<void>;
}

/**
 * Base class for all message handlers
 */
export abstract class BaseMessageHandler implements MessageHandler {
  protected logger: Logger;
  
  constructor(logger: Logger) {
    this.logger = logger;
  }
  
  abstract handle(from: Address, message: any): Promise<void>;
}

/**
 * Request message handler
 */
export class RequestHandler extends BaseMessageHandler {
  private methods: Map<string, (params: any[], from: Address) => Promise<any[]>>;
  private signer: (payload: any) => Promise<string>; // Signer likely returns Hex, but let's keep as string for now if it works elsewhere
  private address: Address;
  private createPayload: (data: any) => Hex; // Update property type
  private verifySignature: (payload: Hex, signature: Hex, expectedSigner: Address) => Promise<boolean>; // Update property type
  private sendResponse: (to: Address, requestId: number, method: string, result: any[]) => Promise<void>;
  private sendErrorResponse: (to: Address, requestId: number, code: number, message: string) => Promise<void>;
  
  constructor(
    logger: Logger,
    methods: Map<string, (params: any[], from: Address) => Promise<any[]>>,
    signer: (payload: any) => Promise<string>,
    address: Address,
    createPayload: (data: any) => Hex, // Expect Hex
    verifySignature: (payload: Hex, signature: Hex, expectedSigner: Address) => Promise<boolean>, // Expect Hex
    sendResponse: (to: Address, requestId: number, method: string, result: any[]) => Promise<void>,
    sendErrorResponse: (to: Address, requestId: number, code: number, message: string) => Promise<void>
  ) {
    super(logger);
    this.methods = methods;
    this.signer = signer;
    this.address = address;
    this.createPayload = createPayload;
    this.verifySignature = verifySignature;
    this.sendResponse = sendResponse;
    this.sendErrorResponse = sendErrorResponse;
  }
  
  async handle(from: Address, request: RPCRequest): Promise<void> {
    const [requestId, method, params, timestamp] = request.req;
    
    this.logger.debug('Handling request', { 
      requestId, 
      method, 
      from,
      paramsLength: params.length 
    });
    
    try {
      // Verify the signature if present
      // Verify the signature if present (request.sig is Hex)
      if (request.sig) {
        const payload = this.createPayload(request.req); // payload is Hex
        const valid = await this.verifySignature(payload, request.sig, from);
        
        if (!valid) {
          this.logger.warn('Invalid signature for request', { 
            requestId, 
            method, 
            from 
          });
          
          // Invalid signature, send error response
          await this.sendErrorResponse(
            from, 
            requestId, 
            RPC_ERROR_CODES.INVALID_SIGNATURE, 
            'Invalid signature'
          );
          return;
        }
      }
      
      // Look up the method handler
      const handler = this.methods.get(method);
      
      if (!handler) {
        this.logger.warn('Method not found', { 
          requestId, 
          method, 
          from 
        });
        
        // Method not found, send error response
        await this.sendErrorResponse(
          from, 
          requestId, 
          RPC_ERROR_CODES.METHOD_NOT_FOUND, 
          `Method '${method}' not found`
        );
        return;
      }
      
      // Call the handler
      const result = await handler(params, from);
      
      // Send the response
      await this.sendResponse(from, requestId, method, result);
      
      this.logger.debug('Request handled successfully', { 
        requestId, 
        method, 
        from,
        resultLength: result.length 
      });
    } catch (error: any) {
      this.logger.error('Error handling request', { 
        requestId, 
        method, 
        from, 
        error 
      });
      
      // Map HachiError to appropriate RPC error code
      let errorCode = RPC_ERROR_CODES.INTERNAL_ERROR;
      let errorMessage = error.message || 'Internal error';
      
      if (error instanceof HachiError) { // Use direct class name
        // Map error types to appropriate RPC error codes
        if (error instanceof ValidationError) { // Use direct class name
          errorCode = RPC_ERROR_CODES.INVALID_PARAMS;
        } else if (error instanceof InvalidSignatureError) { // Use direct class name
          errorCode = RPC_ERROR_CODES.INVALID_SIGNATURE;
        } else if (error instanceof UnauthorizedError) { // Use direct class name
          errorCode = RPC_ERROR_CODES.UNAUTHORIZED;
        } else if (error instanceof StateError) { // Use direct class name
          errorCode = RPC_ERROR_CODES.INVALID_STATE;
        } else if (error instanceof ChannelNotFoundError) { // Use direct class name
          errorCode = RPC_ERROR_CODES.CHANNEL_NOT_FOUND;
        } else if (error instanceof MethodNotFoundError) { // Use direct class name
          errorCode = RPC_ERROR_CODES.METHOD_NOT_FOUND;
        } else if (error instanceof TimeoutError) { // Use direct class name
          errorCode = RPC_ERROR_CODES.TIMEOUT;
        } else if (error instanceof VirtualChannelError) { // Use direct class name
          errorCode = RPC_ERROR_CODES.VIRTUAL_CHANNEL_ERROR;
        }
        
        errorMessage = error.toString();
      }
      
      // Send error response
      await this.sendErrorResponse(
        from,
        requestId,
        errorCode,
        errorMessage
      );
    }
  }
}

/**
 * Response message handler
 * TODO: Verify if signatures on RESPONSE messages need to be checked according to protocol spec.
 */
export class ResponseHandler extends BaseMessageHandler {
  private pendingRequests: Map<number, {
    resolve: (result: any[]) => void;
    reject: (error: Error) => void;
    timeout: NodeJS.Timeout;
    retries: number;
  }>;
  
  constructor(
    logger: Logger,
    pendingRequests: Map<number, {
      resolve: (result: any[]) => void;
      reject: (error: Error) => void;
      timeout: NodeJS.Timeout;
      retries: number;
    }>
  ) {
    super(logger);
    this.pendingRequests = pendingRequests;
  }
  
  async handle(from: Address, response: RPCResponse): Promise<void> {
    const [requestId, method, result, timestamp] = response.res;
    
    // Look up the pending request
    const pendingRequest = this.pendingRequests.get(requestId);
    
    if (!pendingRequest) {
      // No pending request found, ignore
      this.logger.warn('Received response for unknown request', { 
        requestId, 
        method, 
        from 
      });
      return;
    }
    
    this.logger.debug('Received response', { 
      requestId, 
      method, 
      from,
      resultLength: result?.length
    });
    
    // Clear the timeout
    clearTimeout(pendingRequest.timeout);
    
    // Clear from pending requests
    this.pendingRequests.delete(requestId);
    
    // Resolve the promise
    pendingRequest.resolve(result);
  }
}

/**
 * Error message handler
 * TODO: Verify if signatures on ERROR messages need to be checked according to protocol spec.
 */
export class ErrorHandler extends BaseMessageHandler {
  private pendingRequests: Map<number, {
    resolve: (result: any[]) => void;
    reject: (error: Error) => void;
    timeout: NodeJS.Timeout;
    retries: number;
  }>;
  
  constructor(
    logger: Logger,
    pendingRequests: Map<number, {
      resolve: (result: any[]) => void;
      reject: (error: Error) => void;
      timeout: NodeJS.Timeout;
      retries: number;
    }>
  ) {
    super(logger);
    this.pendingRequests = pendingRequests;
  }
  
  async handle(from: Address, error: RPCTypeError): Promise<void> { // Use aliased type RPCTypeError
    const [requestId, code, message, timestamp] = error.err;
    
    // Look up the pending request
    const pendingRequest = this.pendingRequests.get(requestId);
    
    if (!pendingRequest) {
      // No pending request found, ignore
      this.logger.warn('Received error for unknown request', { 
        requestId, 
        code, 
        message, 
        from 
      });
      return;
    }
    
    this.logger.debug('Received error response', { 
      requestId, 
      code, 
      message, 
      from 
    });
    
    // Clear the timeout
    clearTimeout(pendingRequest.timeout);
    
    // Clear from pending requests
    this.pendingRequests.delete(requestId);
    
    // Create appropriate error type based on the RPC error code
    let hachiError: HachiError; // Use direct type
    
    switch (code) {
      case RPC_ERROR_CODES.INVALID_PARAMS:
        hachiError = new InvalidRPCParamsError(message, { // Use direct class name
          requestId, 
          from,
          code
        });
        break;
        
      case RPC_ERROR_CODES.INVALID_SIGNATURE:
        hachiError = new InvalidSignatureError(message, { // Use direct class name
          requestId, 
          from,
          code
        });
        break;
        
      case RPC_ERROR_CODES.UNAUTHORIZED:
        hachiError = new UnauthorizedError(message, { // Use direct class name
          requestId, 
          from,
          code
        });
        break;
        
      case RPC_ERROR_CODES.METHOD_NOT_FOUND:
        hachiError = new MethodNotFoundError(undefined, { // Use direct class name
          requestId, 
          from,
          code,
          message
        });
        break;
        
      case RPC_ERROR_CODES.CHANNEL_NOT_FOUND:
        hachiError = new ChannelNotFoundError(undefined, { // Use direct class name
          requestId, 
          from,
          code,
          message
        });
        break;
        
      case RPC_ERROR_CODES.TIMEOUT:
        hachiError = new RequestTimeoutError(message, undefined, { // Use direct class name
          requestId, 
          from,
          code
        });
        break;
        
      case RPC_ERROR_CODES.VIRTUAL_CHANNEL_ERROR:
        hachiError = new VirtualChannelError(message, undefined, undefined, undefined, { // Use direct class name
          requestId, 
          from,
          code
        });
        break;
        
      default:
        // For standard JSON-RPC errors and other errors
        hachiError = new HachiRPCError(message, code, { // Use aliased HachiRPCError
          requestId, 
          from 
        });
    }
    
    // Reject the promise with the appropriate error
    pendingRequest.reject(hachiError);
  }
}

/**
 * Notification message handler
 */
export class NotificationHandler extends BaseMessageHandler {
  private verifySignature: (payload: Hex, signature: Hex, expectedSigner: Address) => Promise<boolean>; // Expect Hex
  private createPayload: (data: any) => Hex; // Expect Hex
  private emitEvent: (type: string, data: any[], from: Address) => void;
  
  constructor(
    logger: Logger,
    verifySignature: (payload: Hex, signature: Hex, expectedSigner: Address) => Promise<boolean>, // Expect Hex
    createPayload: (data: any) => Hex, // Expect Hex
    emitEvent: (type: string, data: any[], from: Address) => void
  ) {
    super(logger);
    this.verifySignature = verifySignature;
    this.createPayload = createPayload;
    this.emitEvent = emitEvent;
  }
  
  async handle(from: Address, notification: RPCNotification): Promise<void> {
    const [type, data, timestamp] = notification.ntf;
    
    this.logger.debug('Received notification', { 
      type, 
      from,
      dataLength: data?.length
    });
    
    try {
      // Verify the signature if present (notification.sig is Hex)
      if (notification.sig) {
        const payload = this.createPayload(notification.ntf); // payload is Hex
        const valid = await this.verifySignature(payload, notification.sig, from);
        
        if (!valid) {
          // Invalid signature, ignore notification
          this.logger.warn('Invalid signature for notification, ignoring', { 
            type, 
            from 
          });
          return;
        }
      }
      
      // Emit an event for the notification
      this.emitEvent(type, data, from);
      
      this.logger.debug('Notification processed', { type, from });
    } catch (error) {
      this.logger.error('Error handling notification', { 
        type, 
        from, 
        error 
      });
      // No response for notifications, so just log the error
    }
  }
}
