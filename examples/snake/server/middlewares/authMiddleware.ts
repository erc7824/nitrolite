import { Request, Response, NextFunction } from 'express';
import { ethers } from 'ethers';
import { verifySignature } from '../services/brokerService';

// Simple challenge store - in production, this should use a proper database or Redis
const challenges = new Map<string, { challenge: string, timestamp: number }>();

// Check if a client is authenticated via signature
export function requireAuth(req: Request, res: Response, next: NextFunction) {
  const authHeader = req.headers.authorization;
  
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'Authentication required' });
  }
  
  const token = authHeader.split(' ')[1];
  
  try {
    // In a real implementation, verify JWT or session
    // For demo purposes, we'll assume a simple token format: address:signature
    const [address, signature] = token.split(':');
    
    if (!address || !signature || !ethers.utils.isAddress(address)) {
      return res.status(401).json({ error: 'Invalid authentication token' });
    }
    
    // Add the authenticated address to the request for use in route handlers
    (req as any).authenticatedAddress = address;
    
    next();
  } catch (error) {
    console.error('Auth middleware error:', error);
    res.status(401).json({ error: 'Invalid authentication token' });
  }
}

// Generate a challenge for a client
export function generateChallenge(address: string): { challenge: string, timestamp: number } {
  // Generate a random challenge
  const timestamp = Date.now();
  const randomBytes = ethers.utils.randomBytes(16);
  const randomHex = ethers.utils.hexlify(randomBytes);
  const challenge = `Sign this message to authenticate with Nitro Snake: ${randomHex} at ${timestamp}`;
  
  // Store the challenge
  challenges.set(address.toLowerCase(), { challenge, timestamp });
  
  return { challenge, timestamp };
}

// Verify a client's signature against a stored challenge
export function verifyChallengeSignature(address: string, signature: string): boolean {
  if (!ethers.utils.isAddress(address)) {
    return false;
  }
  
  const storedChallenge = challenges.get(address.toLowerCase());
  if (!storedChallenge) {
    return false;
  }
  
  // Check if the challenge has expired (valid for 5 minutes)
  const now = Date.now();
  if (now - storedChallenge.timestamp > 5 * 60 * 1000) {
    challenges.delete(address.toLowerCase());
    return false;
  }
  
  // Verify the signature
  const isValid = verifySignature(storedChallenge.challenge, signature, address);
  
  // Remove the challenge after verification to prevent reuse
  if (isValid) {
    challenges.delete(address.toLowerCase());
  }
  
  return isValid;
}