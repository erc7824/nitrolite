import { ethers } from 'ethers';
import { logger } from '../../utils/logger.js';

/**
 * Verify a signature against a message and expected signer
 */
export function verifySignature(message: string, signature: string, expectedAddress: string): boolean {
  try {
    const recoveredAddress = ethers.verifyMessage(message, signature);
    return recoveredAddress.toLowerCase() === expectedAddress.toLowerCase();
  } catch (error) {
    logger.error('Error verifying signature:', error);
    return false;
  }
}

/**
 * Verify wallet authentication for incoming requests
 */
export function authenticateWallet(walletAddress: string, signature: string, message: string): boolean {
  if (!walletAddress || !signature || !message) {
    logger.warn('Missing required authentication parameters');
    return false;
  }

  try {
    return verifySignature(message, signature, walletAddress);
  } catch (error) {
    logger.error('Wallet authentication failed:', error);
    return false;
  }
}

/**
 * Generate authentication challenge for wallet
 */
export function generateAuthChallenge(walletAddress: string): string {
  const timestamp = Date.now();
  const nonce = Math.random().toString(36).substring(2, 15);
  
  return `Please sign this message to authenticate with {{projectName}}:
Address: ${walletAddress}
Timestamp: ${timestamp}
Nonce: ${nonce}`;
}

/**
 * Validate authentication challenge
 */
export function validateAuthChallenge(challenge: string, maxAge: number = 300000): boolean {
  try {
    const timestampMatch = challenge.match(/Timestamp: (\d+)/);
    if (!timestampMatch) {
      return false;
    }

    const timestamp = parseInt(timestampMatch[1]);
    const age = Date.now() - timestamp;
    
    return age <= maxAge; // Default 5 minutes
  } catch (error) {
    logger.error('Error validating auth challenge:', error);
    return false;
  }
}