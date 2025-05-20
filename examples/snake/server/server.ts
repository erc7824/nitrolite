import { createServer } from 'http';
import { WebSocketServer } from 'ws';
import dotenv from 'dotenv';
import express from 'express';
import bodyParser from 'body-parser';

// Load environment variables from .env file
dotenv.config();

// Create Express app
const app = express();
app.use(bodyParser.json());

// Create HTTP server
const server = createServer(app);

// Create WebSocket server
const wss = new WebSocketServer({ server });

// Export server objects
export { server, wss };

// Import after exports to avoid circular dependencies
import { connectToBroker } from './services/brokerService';
import { setupWebSocketHandlers } from './services/websocketService';
import { setupApiRoutes } from './routes/apiRoutes';
import { gracefulShutdown } from './utils/shutdown';

// Setup API routes
setupApiRoutes(app);

// Setup WebSocket handlers
setupWebSocketHandlers(wss);

// Initialize broker connection
connectToBroker();

// Start server
const PORT = process.env.PORT || 3001;
server.listen(PORT, () => {
  console.log(`WebSocket server is running on port ${PORT}`);
});

// Handle graceful shutdown
process.on('SIGTERM', () => gracefulShutdown(server, wss, 'SIGTERM'));
process.on('SIGINT', () => gracefulShutdown(server, wss, 'SIGINT'));