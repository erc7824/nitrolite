import { Router, Request, Response } from 'express';
import { config } from '../config/index.js';
import { isAuthenticatedWithBroker } from '../services/nitrolite/client.js';
import { getConnectionStats } from '../services/websocket.js';

export const healthRouter = Router();

/**
 * Basic health check endpoint
 */
healthRouter.get('/', (req: Request, res: Response) => {
  res.json({
    status: 'healthy',
    server: '{{projectName}}',
    version: '0.1.0',
    timestamp: new Date().toISOString(),
    uptime: process.uptime()
  });
});

/**
 * Detailed health check with service status
 */
healthRouter.get('/detailed', (req: Request, res: Response) => {
  const connectionStats = getConnectionStats();
  
  res.json({
    status: 'healthy',
    server: '{{projectName}}',
    version: '0.1.0',
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    environment: config.isDev ? 'development' : 'production',
    services: {
      nitrolite: {
        connected: isAuthenticatedWithBroker(),
        brokerUrl: config.yellowWsUrl
      },
      websocket: {
        totalConnections: connectionStats.total,
        authenticatedConnections: connectionStats.authenticated
      }
    },
    memory: {
      used: Math.round(process.memoryUsage().heapUsed / 1024 / 1024),
      total: Math.round(process.memoryUsage().heapTotal / 1024 / 1024),
      unit: 'MB'
    }
  });
});

/**
 * Readiness probe endpoint
 */
healthRouter.get('/ready', (req: Request, res: Response) => {
  const isReady = isAuthenticatedWithBroker();
  
  if (isReady) {
    res.status(200).json({
      status: 'ready',
      message: 'Service is ready to accept connections'
    });
  } else {
    res.status(503).json({
      status: 'not_ready',
      message: 'Service is not ready - Nitrolite client not connected'
    });
  }
});

/**
 * Liveness probe endpoint
 */
healthRouter.get('/live', (req: Request, res: Response) => {
  res.status(200).json({
    status: 'alive',
    timestamp: new Date().toISOString()
  });
});