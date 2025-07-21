import type { MessageSigner, RequestData, ResponsePayload } from '@erc7824/nitrolite';

/**
 * Wallet signer interface
 */
export interface WalletSigner {
  /** Public key in hexadecimal format */
  publicKey: string;
  /** Ethereum address derived from the public key */
  address: `0x${string}`;
  /** Function to sign a message and return a hex signature */
  sign: MessageSigner;
}

/**
 * WebSocket message types for the application
 */
export interface WebSocketMessage {
  type: string;
  payload?: any;
  timestamp?: number;
}

/**
 * Authentication message
 */
export interface AuthMessage extends WebSocketMessage {
  type: 'auth';
  payload: {
    walletAddress: string;
    signature: string;
    message: string;
  };
}

/**
 * Generic application message
 */
export interface AppMessage extends WebSocketMessage {
  type: 'app_message';
  payload: {
    action: string;
    data?: any;
  };
}

/**
 * Error message
 */
export interface ErrorMessage extends WebSocketMessage {
  type: 'error';
  payload: {
    code: string;
    message: string;
    details?: any;
  };
}

/**
 * Success response message
 */
export interface SuccessMessage extends WebSocketMessage {
  type: 'success';
  payload: {
    message: string;
    data?: any;
  };
}

/**
 * Connection information
 */
export interface ConnectionInfo {
  id: string;
  walletAddress?: string;
  isAuthenticated: boolean;
  connectedAt: Date;
  lastActivity: Date;
}

/**
 * App session information
 */
export interface AppSession {
  id: string;
  participants: string[];
  createdAt: Date;
  isActive: boolean;
  data?: any;
}