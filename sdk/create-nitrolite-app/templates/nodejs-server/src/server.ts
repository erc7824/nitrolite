import express from 'express';
import { createServer } from 'http';
import { WebSocketServer } from 'ws';
import { config, isDevelopment } from './config/index.js';
import { initializeNitroliteClient } from './services/nitrolite/client.js';
import { logger } from './utils/logger.js';
import { setupGracefulShutdown } from './utils/shutdown.js';

async function startServer() {
  try {
    // Create Express app
    const app = express();
    app.use(express.json());

    // Create HTTP server
    const server = createServer(app);

    // Create WebSocket server
    const wss = new WebSocketServer({ server });

    // Initialize Nitrolite client
    logger.info('Initializing Nitrolite client...');
    await initializeNitroliteClient();
    logger.info('Nitrolite client initialized successfully');

    // Setup graceful shutdown
    setupGracefulShutdown(server, wss);

    // Start server
    server.listen(config.port, () => {
      logger.info(`🚀 Server started on port ${config.port}`);
      logger.info(`🔗 Connecting to Yellow network: ${config.yellowWsUrl}`);
      if (isDevelopment) {
        logger.info(`💡 Running in development mode`);
      }
    });
  } catch (error) {
    logger.error('Failed to start server:', error);
    process.exit(1);
  }
}

// Start the server
startServer();
