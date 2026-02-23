import type { WebSocketServer, WebSocket } from 'ws';
import { getNitroliteClient } from './nitrolite/client.js';
import { logger } from '../utils/logger.js';

interface ClientConnection {
  ws: WebSocket;
  id: string;
  isAlive: boolean;
}

class WebSocketService {
  private clients = new Map<string, ClientConnection>();
  private messageHandlers = new Map<string, (client: ClientConnection, data: any) => void>();

  constructor() {
    this.setupMessageHandlers();
  }

  private setupMessageHandlers(): void {
    // Handler for ping messages
    this.messageHandlers.set('ping', (client: ClientConnection, data: any) => {
      this.sendToClient(client.id, { type: 'pong', timestamp: Date.now() });
    });

    // Handler for Nitrolite message forwarding
    this.messageHandlers.set('nitrolite_message', (client: ClientConnection, data: any) => {
      const nitroliteClient = getNitroliteClient();
      
      if (!nitroliteClient || !nitroliteClient.isConnected) {
        this.sendToClient(client.id, {
          type: 'error',
          message: 'Not connected to Nitrolite network',
          code: 'NOT_CONNECTED'
        });
        return;
      }

      try {
        nitroliteClient.send(data.payload);
        logger.debug(`Forwarded message to Nitrolite from client ${client.id}`);
      } catch (error) {
        logger.error('Failed to forward message to Nitrolite:', error);
        this.sendToClient(client.id, {
          type: 'error',
          message: 'Failed to forward message to Nitrolite',
          code: 'FORWARD_FAILED',
          error: error instanceof Error ? error.message : 'Unknown error'
        });
      }
    });

    // Handler for status requests
    this.messageHandlers.set('status', (client: ClientConnection, data: any) => {
      const nitroliteClient = getNitroliteClient();
      
      this.sendToClient(client.id, {
        type: 'status',
        nitrolite: {
          connected: nitroliteClient?.isConnected || false,
          status: nitroliteClient?.currentStatus || 'disconnected',
          sessionAddress: nitroliteClient?.currentSessionAddress || null,
        },
        server: {
          uptime: process.uptime(),
          connectedClients: this.clients.size,
        }
      });
    });
  }

  setupWebSocketServer(wss: WebSocketServer): void {
    logger.info('Setting up WebSocket server...');

    // Setup Nitrolite message forwarding
    this.setupNitroliteForwarding();

    wss.on('connection', (ws: WebSocket) => {
      const clientId = this.generateClientId();
      
      const client: ClientConnection = {
        ws,
        id: clientId,
        isAlive: true,
      };

      this.clients.set(clientId, client);
      logger.info(`Client ${clientId} connected. Total clients: ${this.clients.size}`);

      // Send welcome message
      this.sendToClient(clientId, {
        type: 'welcome',
        clientId,
        timestamp: Date.now(),
        nitroliteStatus: {
          connected: getNitroliteClient()?.isConnected || false,
          status: getNitroliteClient()?.currentStatus || 'disconnected',
        }
      });

      // Handle incoming messages
      ws.on('message', (message: Buffer) => {
        try {
          const data = JSON.parse(message.toString());
          this.handleClientMessage(client, data);
        } catch (error) {
          logger.error(`Invalid message from client ${clientId}:`, error);
          this.sendToClient(clientId, {
            type: 'error',
            message: 'Invalid JSON message',
            code: 'INVALID_JSON'
          });
        }
      });

      // Handle pong responses for heartbeat
      ws.on('pong', () => {
        client.isAlive = true;
      });

      // Handle client disconnection
      ws.on('close', () => {
        this.clients.delete(clientId);
        logger.info(`Client ${clientId} disconnected. Total clients: ${this.clients.size}`);
      });

      // Handle WebSocket errors
      ws.on('error', (error) => {
        logger.error(`WebSocket error for client ${clientId}:`, error);
        this.clients.delete(clientId);
      });
    });

    // Setup heartbeat mechanism
    this.setupHeartbeat();

    logger.info('WebSocket server setup complete');
  }

  private setupNitroliteForwarding(): void {
    const nitroliteClient = getNitroliteClient();
    
    if (!nitroliteClient) {
      logger.warn('Nitrolite client not available for message forwarding');
      return;
    }

    // Forward Nitrolite messages to all connected clients
    nitroliteClient.onMessage((message: any) => {
      this.broadcastToClients({
        type: 'nitrolite_message',
        data: message,
        timestamp: Date.now(),
      });
    });

    // Forward Nitrolite status changes to all connected clients
    nitroliteClient.onStatusChange((status) => {
      this.broadcastToClients({
        type: 'nitrolite_status',
        status,
        timestamp: Date.now(),
      });
    });

    // Forward Nitrolite errors to all connected clients
    nitroliteClient.onError((error) => {
      this.broadcastToClients({
        type: 'nitrolite_error',
        error: error.message,
        timestamp: Date.now(),
      });
    });

    logger.info('Nitrolite message forwarding setup complete');
  }

  private handleClientMessage(client: ClientConnection, data: any): void {
    const { type } = data;

    if (!type) {
      this.sendToClient(client.id, {
        type: 'error',
        message: 'Message type is required',
        code: 'MISSING_TYPE'
      });
      return;
    }

    const handler = this.messageHandlers.get(type);
    if (handler) {
      try {
        handler(client, data);
      } catch (error) {
        logger.error(`Error handling message type ${type} from client ${client.id}:`, error);
        this.sendToClient(client.id, {
          type: 'error',
          message: 'Internal server error',
          code: 'HANDLER_ERROR'
        });
      }
    } else {
      logger.warn(`Unknown message type '${type}' from client ${client.id}`);
      this.sendToClient(client.id, {
        type: 'error',
        message: `Unknown message type: ${type}`,
        code: 'UNKNOWN_TYPE'
      });
    }
  }

  private sendToClient(clientId: string, data: any): void {
    const client = this.clients.get(clientId);
    if (!client || client.ws.readyState !== client.ws.OPEN) {
      return;
    }

    try {
      client.ws.send(JSON.stringify(data));
    } catch (error) {
      logger.error(`Failed to send message to client ${clientId}:`, error);
      this.clients.delete(clientId);
    }
  }

  private broadcastToClients(data: any): void {
    const message = JSON.stringify(data);
    
    this.clients.forEach((client, clientId) => {
      if (client.ws.readyState === client.ws.OPEN) {
        try {
          client.ws.send(message);
        } catch (error) {
          logger.error(`Failed to broadcast to client ${clientId}:`, error);
          this.clients.delete(clientId);
        }
      }
    });
  }

  private setupHeartbeat(): void {
    const interval = setInterval(() => {
      this.clients.forEach((client, clientId) => {
        if (!client.isAlive) {
          logger.info(`Terminating unresponsive client ${clientId}`);
          client.ws.terminate();
          this.clients.delete(clientId);
          return;
        }

        client.isAlive = false;
        if (client.ws.readyState === client.ws.OPEN) {
          try {
            client.ws.ping();
          } catch (error) {
            logger.error(`Failed to ping client ${clientId}:`, error);
            this.clients.delete(clientId);
          }
        }
      });
    }, 30000); // 30 seconds

    // Cleanup interval on process exit
    process.on('SIGINT', () => {
      clearInterval(interval);
    });

    process.on('SIGTERM', () => {
      clearInterval(interval);
    });
  }

  private generateClientId(): string {
    return `client_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  // Public methods for external use
  getConnectedClientsCount(): number {
    return this.clients.size;
  }

  getConnectedClientIds(): string[] {
    return Array.from(this.clients.keys());
  }

  sendToAllClients(data: any): void {
    this.broadcastToClients(data);
  }

  disconnectClient(clientId: string): boolean {
    const client = this.clients.get(clientId);
    if (client) {
      client.ws.close();
      this.clients.delete(clientId);
      return true;
    }
    return false;
  }
}

// Global WebSocket service instance
const webSocketService = new WebSocketService();

export function setupWebSocketHandlers(wss: WebSocketServer): void {
  webSocketService.setupWebSocketServer(wss);
}

export function getWebSocketService(): WebSocketService {
  return webSocketService;
}