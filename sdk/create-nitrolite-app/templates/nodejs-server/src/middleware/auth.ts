import { Request, Response, NextFunction } from 'express';
import { logger } from '../utils/logger.js';
import { authenticateWallet, validateAuthChallenge } from '../services/nitrolite/auth.js';

/**
 * Extended Request interface with authentication info
 */
export interface AuthenticatedRequest extends Request {
  walletAddress?: string;
  isAuthenticated?: boolean;
}

/**
 * Middleware to authenticate wallet-signed requests
 */
export function authenticateWalletMiddleware(req: AuthenticatedRequest, res: Response, next: NextFunction): void {
  const walletAddress = req.headers['x-wallet-address'] as string;
  const signature = req.headers['x-signature'] as string;
  const message = req.headers['x-message'] as string;

  if (!walletAddress || !signature || !message) {
    res.status(401).json({
      error: 'Authentication required',
      code: 'MISSING_AUTH_HEADERS',
      message: 'Missing required authentication headers: x-wallet-address, x-signature, x-message'
    });
    return;
  }

  // Validate auth challenge
  if (!validateAuthChallenge(message)) {
    res.status(401).json({
      error: 'Authentication failed',
      code: 'EXPIRED_CHALLENGE',
      message: 'Authentication challenge expired'
    });
    return;
  }

  // Verify signature
  if (!authenticateWallet(walletAddress, signature, message)) {
    res.status(401).json({
      error: 'Authentication failed',
      code: 'INVALID_SIGNATURE',
      message: 'Invalid signature'
    });
    return;
  }

  // Add authentication info to request
  req.walletAddress = walletAddress;
  req.isAuthenticated = true;

  logger.debug(`HTTP request authenticated for wallet: ${walletAddress}`);
  next();
}

/**
 * Middleware to require authentication
 */
export function requireAuth(req: AuthenticatedRequest, res: Response, next: NextFunction): void {
  if (!req.isAuthenticated) {
    res.status(401).json({
      error: 'Authentication required',
      code: 'NOT_AUTHENTICATED',
      message: 'This endpoint requires authentication'
    });
    return;
  }
  next();
}