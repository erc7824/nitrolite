import { WebSocketServer, WebSocket } from 'ws';
import { randomBytes } from 'crypto';
import { logger } from '../utils/logger.js';
import { authenticateWallet, generateAuthChallenge, validateAuthChallenge } from './nitrolite/auth.js';
import type { 
  WebSocketConnection, 
  WebSocketMessage, 
  AuthMessage, 
  ErrorMessage, 
  SuccessMessage,
  ConnectionInfo 
} from '../types/index.js';

// Store active connections
const connections = new Map<string, ConnectionInfo>();

/**
 * Setup WebSocket server handlers
 */
export function setupWebSocketHandlers(wss: WebSocketServer): void {
  logger.info('Setting up WebSocket handlers');

  wss.on('connection', (ws: WebSocket) => {
    const connection = ws as WebSocketConnection;
    connection.id = randomBytes(8).toString('hex');
    connection.isAuthenticated = false;
    connection.connectedAt = new Date();
    connection.lastActivity = new Date();

    // Store connection info
    connections.set(connection.id, {
      id: connection.id,
      isAuthenticated: false,
      connectedAt: connection.connectedAt,
      lastActivity: connection.lastActivity
    });

    logger.info(`Client connected: ${connection.id}`);

    // Send welcome message
    sendMessage(connection, {
      type: 'welcome',
      payload: {
        connectionId: connection.id,
        message: 'Welcome to {{projectName}}'
      }
    });

    // Handle incoming messages
    connection.on('message', async (data: Buffer) => {
      try {
        connection.lastActivity = new Date();
        const message: WebSocketMessage = JSON.parse(data.toString());
        await handleMessage(connection, message);
      } catch (error) {
        logger.error('Error parsing WebSocket message:', error);
        sendError(connection, 'INVALID_MESSAGE', 'Invalid message format');
      }
    });

    // Handle connection close
    connection.on('close', () => {
      logger.info(`Client disconnected: ${connection.id}`);
      connections.delete(connection.id!);
    });

    // Handle connection error
    connection.on('error', (error) => {
      logger.error(`WebSocket error for ${connection.id}:`, error);
    });
  });

  // Periodic cleanup of stale connections
  setInterval(() => {
    const staleThreshold = 5 * 60 * 1000; // 5 minutes
    const now = new Date().getTime();

    for (const [id, info] of connections.entries()) {
      if (now - info.lastActivity.getTime() > staleThreshold) {
        logger.info(`Removing stale connection: ${id}`);
        connections.delete(id);
      }
    }
  }, 60000); // Check every minute
}

/**
 * Handle incoming WebSocket messages
 */
async function handleMessage(connection: WebSocketConnection, message: WebSocketMessage): Promise<void> {
  logger.debug(`Handling message type: ${message.type} from ${connection.id}`);

  switch (message.type) {
    case 'ping':
      handlePing(connection);
      break;

    case 'auth':
      await handleAuth(connection, message as AuthMessage);
      break;

    case 'get_challenge':
      handleGetChallenge(connection, message);
      break;

    case 'app_message':
      if (connection.isAuthenticated) {
        await handleAppMessage(connection, message);
      } else {
        sendError(connection, 'NOT_AUTHENTICATED', 'Authentication required');
      }
      break;

    default:
      logger.warn(`Unknown message type: ${message.type} from ${connection.id}`);
      sendError(connection, 'UNKNOWN_MESSAGE_TYPE', `Unknown message type: ${message.type}`);
  }
}

/**
 * Handle ping messages
 */
function handlePing(connection: WebSocketConnection): void {
  sendMessage(connection, {
    type: 'pong',
    timestamp: Date.now()
  });
}

/**
 * Handle authentication requests
 */
async function handleAuth(connection: WebSocketConnection, message: AuthMessage): Promise<void> {
  const { walletAddress, signature, message: authMessage } = message.payload;

  if (!walletAddress || !signature || !authMessage) {
    sendError(connection, 'INVALID_AUTH', 'Missing required authentication fields');
    return;
  }

  // Validate auth challenge
  if (!validateAuthChallenge(authMessage)) {
    sendError(connection, 'EXPIRED_CHALLENGE', 'Authentication challenge expired');
    return;
  }

  // Verify signature
  if (!authenticateWallet(walletAddress, signature, authMessage)) {
    sendError(connection, 'INVALID_SIGNATURE', 'Invalid signature');
    return;
  }

  // Mark as authenticated
  connection.isAuthenticated = true;
  connection.walletAddress = walletAddress;

  // Update connection info
  const connInfo = connections.get(connection.id!);
  if (connInfo) {
    connInfo.isAuthenticated = true;
    connInfo.walletAddress = walletAddress;
  }

  logger.info(`Client authenticated: ${connection.id} (${walletAddress})`);

  sendSuccess(connection, 'Authentication successful', {
    walletAddress,
    isAuthenticated: true
  });
}

/**
 * Handle get challenge requests
 */
function handleGetChallenge(connection: WebSocketConnection, message: WebSocketMessage): void {
  const walletAddress = message.payload?.walletAddress;

  if (!walletAddress) {
    sendError(connection, 'MISSING_WALLET_ADDRESS', 'Wallet address required');
    return;
  }

  const challenge = generateAuthChallenge(walletAddress);

  sendMessage(connection, {
    type: 'auth_challenge',
    payload: {
      challenge,
      walletAddress
    }
  });
}

/**
 * Handle application-specific messages
 */
async function handleAppMessage(connection: WebSocketConnection, message: WebSocketMessage): Promise<void> {
  const { action, data } = message.payload;

  logger.info(`Handling app message: ${action} from ${connection.walletAddress}`);

  // Add your application-specific message handling here
  switch (action) {
    case 'get_status':
      sendMessage(connection, {
        type: 'app_response',
        payload: {
          action: 'status',
          data: {
            server: '{{projectName}}',
            version: '0.1.0',
            timestamp: Date.now(),
            connections: connections.size
          }
        }
      });
      break;

    default:
      logger.warn(`Unknown app action: ${action}`);
      sendError(connection, 'UNKNOWN_ACTION', `Unknown action: ${action}`);
  }
}

/**
 * Send a message to a WebSocket connection
 */
function sendMessage(connection: WebSocketConnection, message: WebSocketMessage): void {
  if (connection.readyState === WebSocket.OPEN) {
    connection.send(JSON.stringify({
      ...message,
      timestamp: message.timestamp || Date.now()
    }));
  }
}

/**
 * Send an error message
 */
function sendError(connection: WebSocketConnection, code: string, message: string, details?: any): void {
  const errorMessage: ErrorMessage = {
    type: 'error',
    payload: {
      code,
      message,
      details
    }
  };
  sendMessage(connection, errorMessage);
}

/**
 * Send a success message
 */
function sendSuccess(connection: WebSocketConnection, message: string, data?: any): void {
  const successMessage: SuccessMessage = {
    type: 'success',
    payload: {
      message,
      data
    }
  };
  sendMessage(connection, successMessage);
}

/**
 * Broadcast message to all authenticated connections
 */
export function broadcastToAuthenticated(message: WebSocketMessage): void {
  for (const [id, info] of connections.entries()) {
    if (info.isAuthenticated) {
      // Find the actual WebSocket connection
      // In a real implementation, you'd store WebSocket references
      logger.debug(`Broadcasting to ${id}`);
    }
  }
}

/**
 * Get connection statistics
 */
export function getConnectionStats(): { total: number; authenticated: number } {
  const total = connections.size;
  const authenticated = Array.from(connections.values()).filter(c => c.isAuthenticated).length;
  
  return { total, authenticated };
}