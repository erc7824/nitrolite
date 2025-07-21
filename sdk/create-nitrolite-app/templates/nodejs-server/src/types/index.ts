// Re-export types from services
export type {
  WalletSigner,
  WebSocketMessage,
  AuthMessage,
  AppMessage,
  ErrorMessage,
  SuccessMessage,
  ConnectionInfo,
  AppSession
} from '../services/nitrolite/types.js';

import { WebSocket } from 'ws';

// Additional shared types
export interface ServerConfig {
  port: number;
  isDev: boolean;
  isProd: boolean;
  yellowWsUrl: string;
  asset: string;
  walletPrivateKey: string;
  vApp: {
    name: string;
    scope: string;
  };
}

export interface WebSocketConnection extends WebSocket {
  id?: string;
  walletAddress?: string;
  isAuthenticated?: boolean;
  connectedAt?: Date;
  lastActivity?: Date;
}