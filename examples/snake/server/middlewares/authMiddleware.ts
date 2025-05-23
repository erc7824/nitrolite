import { ethers } from 'ethers';
import { verifySignature } from '../services/brokerService.ts';

// Simple challenge store - in production, this should use a proper database or Redis
const challenges = new Map<string, { challenge: string, timestamp: number }>();

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
