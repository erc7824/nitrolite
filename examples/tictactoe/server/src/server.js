/**
 * WebSocket server for Nitro Aura Tic Tac Toe game
 */

import { createWebSocketServer, sendError, startPingInterval } from './config/websocket.js';
import { initializeRPCClient, createRoomManager } from './services/index.js';
import { handleJoinRoom, handleGetAvailableRooms } from './routes/roomRoutes.js';
import { handleStartGame, handleMove } from './routes/gameRoutes.js';
import logger from './utils/logger.js';

// Create WebSocket server
const wss = createWebSocketServer();
const roomManager = createRoomManager();

// Track active connections
// TODO: Use @erc7824/nitrolite for connection tracking when available
const connections = new Map();

// Track online users count
let onlineUsersCount = 0;

// Function to broadcast online users count to all clients
const broadcastOnlineUsersCount = () => {
  const message = JSON.stringify({
    type: 'onlineUsers',
    count: onlineUsersCount
  });
  
  wss.clients.forEach((client) => {
    if (client.readyState === 1) { // WebSocket.OPEN
      client.send(message);
    }
  });
  
  logger.ws(`Broadcasting online users count: ${onlineUsersCount}`);
};

// Create context object to share between route handlers
const context = {
  roomManager,
  connections,
  sendError: (ws, code, msg) => sendError(ws, code, msg)
};

wss.on('connection', (ws) => {
  logger.ws('Client connected');
  
  // Increment online users count and broadcast to all clients
  onlineUsersCount++;
  broadcastOnlineUsersCount();
  
  // Handle client messages
  ws.on('message', async (message) => {
    let data;
    try {
      data = JSON.parse(message);
    } catch (e) {
      return sendError(ws, 'INVALID_JSON', 'Invalid JSON format');
    }

    // Process message based on type
    try {
      switch (data.type) {
        case 'joinRoom':
          await handleJoinRoom(ws, data.payload, context);
          break;
        case 'startGame':
          await handleStartGame(ws, data.payload, context);
          break;
        case 'move':
          await handleMove(ws, data.payload, context);
          break;
        case 'getAvailableRooms':
          await handleGetAvailableRooms(ws, context);
          break;
        default:
          sendError(ws, 'INVALID_MESSAGE_TYPE', 'Invalid message type');
      }
    } catch (error) {
      logger.error(`Error handling message type ${data.type}:`, error);
      sendError(ws, 'INTERNAL_ERROR', 'An internal error occurred');
    }
  });

  // Handle disconnection
  ws.on('close', () => {
    // Find and remove the player from any room
    for (const [eoa, connection] of connections.entries()) {
      if (connection.ws === ws) {
        const result = roomManager.leaveRoom(eoa);
        if (result.success && result.roomId) {
          roomManager.broadcastToRoom(result.roomId, 'room:state', {
            roomId: result.roomId,
            // Send updated room state here
          });
        }
        connections.delete(eoa);
        break;
      }
    }
    
    // Decrement online users count and broadcast to all clients
    onlineUsersCount = Math.max(0, onlineUsersCount - 1);
    broadcastOnlineUsersCount();
    
    logger.ws('Client disconnected');
  });
});

// Initialize Nitrolite client and channel when server starts
async function initializeNitroliteServices() {
  try {
    logger.nitro('Initializing Nitrolite services...');
    const rpcClient = await initializeRPCClient();
    logger.nitro('Nitrolite RPC client initialized successfully');
    
    // Check if we have an existing channel
    if (rpcClient.channel) {
      logger.nitro('Connected to existing channel');
      logger.data('Channel info', rpcClient.channel);
    } else {
      logger.warn('No channel established after initialization');
      logger.nitro('Channels will be created as needed via getChannelInfo');
    }
  } catch (error) {
    logger.error('Failed to initialize Nitrolite services:', error);
    logger.system('Continuing in mock mode without Nitrolite channel');
  }
}

// Start server
const port = process.env.PORT || 8080;
logger.system(`WebSocket server starting on port ${port}`);

// Initialize Nitrolite client and channel
initializeNitroliteServices().then(() => {
  logger.system('Server initialization complete');
}).catch(error => {
  logger.error('Server initialization failed:', error);
});

// Start keepalive mechanism
startPingInterval(wss);

// Broadcast online users count periodically to ensure all clients have the latest count
setInterval(() => {
  broadcastOnlineUsersCount();
}, 30000);