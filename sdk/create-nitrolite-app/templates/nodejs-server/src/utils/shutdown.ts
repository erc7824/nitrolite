import { Server } from 'http';
import { WebSocketServer } from 'ws';
import { logger } from './logger.js';

/**
 * Setup graceful shutdown handlers
 */
export function setupGracefulShutdown(server: Server, wss: WebSocketServer): void {
  const gracefulShutdown = (signal: string) => {
    logger.info(`Received ${signal}, shutting down gracefully...`);

    // Stop accepting new connections
    server.close(() => {
      logger.info('HTTP server closed');
    });

    // Close all WebSocket connections
    wss.clients.forEach((ws) => {
      ws.terminate();
    });

    wss.close(() => {
      logger.info('WebSocket server closed');
    });

    // Force exit after timeout
    setTimeout(() => {
      logger.warn('Force closing server after timeout');
      process.exit(1);
    }, 10000); // 10 second timeout

    // Graceful exit
    setTimeout(() => {
      logger.info('Server shut down complete');
      process.exit(0);
    }, 1000);
  };

  // Handle different shutdown signals
  process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
  process.on('SIGINT', () => gracefulShutdown('SIGINT'));

  // Handle uncaught exceptions
  process.on('uncaughtException', (error) => {
    logger.error('Uncaught exception:', error);
    gracefulShutdown('UNCAUGHT_EXCEPTION');
  });

  // Handle unhandled promise rejections
  process.on('unhandledRejection', (reason, promise) => {
    logger.error('Unhandled rejection at:', promise, 'reason:', reason);
    gracefulShutdown('UNHANDLED_REJECTION');
  });
}
