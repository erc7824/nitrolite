import { Address, Hex } from 'viem';
import { ChannelId, State, StateHash } from '../types';

/**
 * RPC Message Type Enum
 */
export enum RPCMessageType {
  REQUEST = 'request',
  RESPONSE = 'response',
  ERROR = 'error',
  NOTIFICATION = 'notification'
}

/**
 * RPC Provider interface
 * Implement this interface to provide transport for RPC messages
 */
export interface RPCProvider {
  /**
   * Send a message to a specific recipient
   * @param recipient The address of the recipient
   * @param message The message to send
   * @returns Promise that resolves when the message is sent
   */
  send(recipient: Address, message: RPCMessage): Promise<void>;
  
  /**
   * Register a handler for incoming messages
   * @param handler The handler function to call when a message is received
   * @returns A function that unregisters the handler
   */
  onMessage(handler: (from: Address, message: RPCMessage) => void): () => void;
  
  /**
   * Connect to the transport
   * @returns Promise that resolves when connected
   */
  connect(): Promise<void>;
  
  /**
   * Disconnect from the transport
   * @returns Promise that resolves when disconnected
   */
  disconnect(): Promise<void>;
}

/**
 * Base RPC Message Interface
 */
export interface RPCMessage {
  type: RPCMessageType;
  ts: number; // Timestamp
  sig?: Hex; // Optional signature of the message payload
}

/**
 * RPC Request Message
 */
export interface RPCRequest extends RPCMessage {
  type: RPCMessageType.REQUEST;
  req: [
    number, // Request ID
    string, // Method name
    any[], // Parameters
    number // Timestamp
  ];
}

/**
 * RPC Response Message
 */
export interface RPCResponse extends RPCMessage {
  type: RPCMessageType.RESPONSE;
  res: [
    number, // Request ID (matching the request)
    string, // Method name (matching the request)
    any[], // Result values
    number // Timestamp
  ];
}

/**
 * RPC Error Message
 */
export interface RPCError extends RPCMessage {
  type: RPCMessageType.ERROR;
  err: [
    number, // Request ID (matching the request)
    number, // Error code
    string, // Error message
    number // Timestamp
  ];
}

/**
 * RPC Notification Message (server-initiated)
 */
export interface RPCNotification extends RPCMessage {
  type: RPCMessageType.NOTIFICATION;
  ntf: [
    string, // Notification type
    any[], // Notification data
    number // Timestamp
  ];
}

/**
 * Union type for all RPC messages
 */
export type RPCMessageUnion = RPCRequest | RPCResponse | RPCError | RPCNotification;

/**
 * RPC Method Handler function type
 */
export type RPCMethodHandler = (params: any[], from: Address) => Promise<any[]>;

/**
 * Standard RPC method names
 */
export enum RPCMethod {
  // Channel methods
  OPEN_CHANNEL = 'open_channel',
  UPDATE_STATE = 'update_state',
  SIGN_STATE = 'sign_state',
  CLOSE_CHANNEL = 'close_channel',
  CHALLENGE_CHANNEL = 'challenge_channel',
  CHECKPOINT_CHANNEL = 'checkpoint_channel',
  
  // Virtual channel methods
  CREATE_VIRTUAL_CHANNEL = 'create_virtual_channel',
  RELAY_STATE_UPDATE = 'relay_state_update',
  RELAY_SIGNATURE = 'relay_signature',
  CLOSE_VIRTUAL_CHANNEL = 'close_virtual_channel',
  
  // Meta methods
  PING = 'ping',
  GET_TIME = 'get_time',
  GET_CHANNELS = 'get_channels'
}

/**
 * Standard notification types
 */
export enum RPCNotificationType {
  STATE_UPDATE = 'state_update',
  CHALLENGE_STARTED = 'challenge_started',
  CHALLENGE_COMPLETED = 'challenge_completed',
  CHANNEL_CLOSED = 'channel_closed'
}

/**
 * Error codes for RPC errors
 * 
 * @deprecated Use the RPC_ERROR_CODES from src/config.ts instead 
 * which provides more detailed documentation and consistent naming.
 * 
 * This enum is kept for backward compatibility but will be removed in a future version.
 * New code should import and use RPC_ERROR_CODES from src/config.ts.
 * 
 * Example:
 * ```typescript
 * // Old usage (deprecated)
 * import { RPCErrorCode } from './types';
 * const code = RPCErrorCode.METHOD_NOT_FOUND;
 * 
 * // New usage (preferred)
 * import { RPC_ERROR_CODES } from '../config';
 * const code = RPC_ERROR_CODES.METHOD_NOT_FOUND;
 * ```
 */
export enum RPCErrorCode {
  // Standard JSON-RPC error codes
  PARSE_ERROR = -32700,        // Invalid JSON
  INVALID_REQUEST = -32600,    // Invalid Request object
  METHOD_NOT_FOUND = -32601,   // Method doesn't exist
  INVALID_PARAMS = -32602,     // Invalid method parameters
  INTERNAL_ERROR = -32603,     // Internal JSON-RPC error
  
  // Nitrolite-specific error codes
  UNAUTHORIZED = -32000,           // Not authorized to call method
  INVALID_STATE = -32001,          // Invalid state transition
  CHANNEL_NOT_FOUND = -32002,      // Channel not found
  INVALID_SIGNATURE = -32003,      // Invalid signature
  INVALID_TRANSITION = -32004,     // Invalid state transition
  VIRTUAL_CHANNEL_ERROR = -32005,  // Virtual channel error
  TIMEOUT = -32007                 // Operation timed out
}

/**
 * Light Virtual Channel Identifier
 * Represents a virtual channel routing path
 */
export interface LightVirtualChannelIdentifier {
  origin: Address;
  destination: Address;
  intermediaries: Address[];
  nonce: bigint;
}

/**
 * Virtual Channel State
 */
export interface VirtualChannelState {
  lvci: LightVirtualChannelIdentifier;
  state: State;
  signatures: Record<string, Hex>; // Map of addresses to signatures
}
