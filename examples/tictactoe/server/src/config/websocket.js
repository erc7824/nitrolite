/**
 * WebSocket server configuration
 */
import { WebSocketServer } from 'ws';
import dotenv from 'dotenv';
import logger from '../utils/logger.js';

// Load environment variables
dotenv.config();

/**
 * Creates a new WebSocket server
 * @returns {WebSocketServer} The WebSocket server instance
 */
export function createWebSocketServer() {
  const port = process.env.PORT || 8080;
  logger.system(`Creating WebSocket server on port ${port}`);
  
  return new WebSocketServer({ 
    host: '0.0.0.0', 
    port 
  });
}

/**
 * Utility function to send an error message
 * @param {WebSocket} ws - WebSocket connection
 * @param {string} code - Error code
 * @param {string} msg - Error message
 */
export function sendError(ws, code, msg) {
  ws.send(JSON.stringify({
    type: 'error',
    code,
    msg
  }));
}

/**
 * Starts a ping interval to keep connections alive
 * @param {WebSocketServer} wss - The WebSocket server instance
 * @param {number} interval - Ping interval in milliseconds (default: 30000)
 * @returns {NodeJS.Timeout} The interval timer
 */
export function startPingInterval(wss, interval = 30000) {
  return setInterval(() => {
    wss.clients.forEach((client) => {
      if (client.readyState === 1) { // WebSocket.OPEN
        client.send(JSON.stringify({ type: 'ping' }));
      }
    });
  }, interval);
}