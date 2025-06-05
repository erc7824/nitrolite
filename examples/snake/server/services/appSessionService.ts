import { Hex } from 'viem';
import { Room } from '../interfaces/index.ts';
import { getRoom } from './stateService.ts';
import { createEthersSigner } from './brokerService.ts';
import { WALLET_PRIVATE_KEY } from '../config/index.ts';
import { CreateAppSessionRequest, AppDefinition, NitroliteRPC } from '@erc7824/nitrolite';

const DEFAULT_PROTOCOL = "app_snake_nitrolite";
const DEFAULT_WEIGHTS: number[] = [0, 0, 100]; // Alice: 0, Bob: 0, Server: 100
const DEFAULT_QUORUM: number = 100; // server alone decides the outcome

// Store pending app session creation data
const pendingAppSessions = new Map<string, {
  requestToSign: any;
  signatures: Map<string, string>;
  participants: Hex[];
  appDefinition: AppDefinition;
  allocations: Array<{ participant: Hex; asset: string; amount: string }>;
}>();

/**
 * Generate the message structure that all participants will sign for app session creation
 */
export function generateAppSessionMessage(roomId: string, participantA: Hex, participantB: Hex): any {
  const room = getRoom(roomId);
  if (!room) {
    throw new Error(`Room ${roomId} not found`);
  }

  const signer = createEthersSigner(WALLET_PRIVATE_KEY);
  const participants = [participantA, participantB, signer.address as Hex];

  const appDefinition: AppDefinition = {
    protocol: DEFAULT_PROTOCOL,
    participants,
    weights: DEFAULT_WEIGHTS,
    quorum: DEFAULT_QUORUM,
    challenge: 0,
    nonce: Date.now(),
  };

  const allocations = participants.map((participant, index) => ({
    participant,
    asset: "usdc",
    amount: index < 2 ? "0.00001" : "0", // Players get 0.00001, server gets 0
  }));

  const requestId = Date.now();
  const timestamp = Date.now();
  const params: CreateAppSessionRequest[] = [{
    definition: appDefinition,
    allocations
  }];

  const requestToSign = [requestId, "create_app_session", params, timestamp];

  // Store pending app session data
  pendingAppSessions.set(roomId, {
    requestToSign,
    signatures: new Map(),
    participants,
    appDefinition,
    allocations
  });

  return { requestToSign, participants };
}

/**
 * Add a signature for app session creation
 */
export function addAppSessionSignature(roomId: string, participantAddress: Hex, signature: string): boolean {
  const pending = pendingAppSessions.get(roomId);
  if (!pending) {
    console.error(`No pending app session found for room ${roomId}`);
    return false;
  }

  // Verify the participant is valid for this room
  if (!pending.participants.includes(participantAddress)) {
    console.error(`Participant ${participantAddress} not found in room ${roomId} participants`);
    return false;
  }

  // Store the signature
  pending.signatures.set(participantAddress, signature);
  console.log(`Added signature for ${participantAddress} in room ${roomId}. Total signatures: ${pending.signatures.size}/${pending.participants.length}`);

  return true;
}

/**
 * Create app session with all collected signatures
 */
export async function createAppSessionWithSignatures(roomId: string): Promise<string> {
  const pending = pendingAppSessions.get(roomId);
  if (!pending) {
    throw new Error(`No pending app session found for room ${roomId}`);
  }

  // Verify we have signatures from all participants except server
  const participantSignaturesNeeded = pending.participants.slice(0, -1); // Exclude server
  for (const participant of participantSignaturesNeeded) {
    if (!pending.signatures.has(participant)) {
      throw new Error(`Missing signature from participant ${participant}`);
    }
  }

  // Server signs the same request
  const signer = createEthersSigner(WALLET_PRIVATE_KEY);
  const serverSignature = await signer.sign(pending.requestToSign);

  // Combine all signatures in participant order
  const allSignatures: string[] = [];
  for (const participant of pending.participants) {
    if (participant === signer.address) {
      allSignatures.push(serverSignature);
    } else {
      const sig = pending.signatures.get(participant);
      if (!sig) {
        throw new Error(`Missing signature from participant ${participant}`);
      }
      allSignatures.push(sig);
    }
  }

  // Create the final signed request
  const signedRequest = {
    req: pending.requestToSign,
    sig: allSignatures
  };

  console.log(`[createAppSessionWithSignatures] Sending app session creation with ${allSignatures.length} signatures for room ${roomId}`);

  // Send to broker via WebSocket (we'll need to integrate this with the existing broker service)
  const response = await sendSignedRequestToBroker(signedRequest);

  // Extract app session ID from response
  const appId = response.app_session_id || (typeof response[0] === "object" ? response[0].app_session_id : null);

  if (!appId) {
    throw new Error("Failed to create app session - no app ID returned");
  }

  // Clean up pending data
  pendingAppSessions.delete(roomId);

  console.log(`[createAppSessionWithSignatures] Created app session ${appId} for room ${roomId}`);
  return appId;
}

/**
 * Generate close app session message structure
 */
export function createCloseAppSessionMessage(roomId: string, appId: Hex, participantA: Hex, participantB: Hex): any {
  const signer = createEthersSigner(WALLET_PRIVATE_KEY);
  const participants = [participantA, participantB, signer.address as Hex];

  const allocations = participants.map((participant, index) => ({
    participant,
    asset: "usdc",
    amount: index < 2 ? "0.00001" : "0", // Same allocations as creation
  }));

  const requestId = Date.now();
  const timestamp = Date.now();
  const params = [{
    app_session_id: appId,
    allocations
  }];

  return [requestId, "close_app_session", params, timestamp];
}

/**
 * Check if app session creation is pending for a room
 */
export function isAppSessionPending(roomId: string): boolean {
  return pendingAppSessions.has(roomId);
}

/**
 * Get pending app session data for a room
 */
export function getPendingAppSession(roomId: string) {
  return pendingAppSessions.get(roomId);
}

/**
 * Clear pending app session data for a room
 */
export function clearPendingAppSession(roomId: string): void {
  pendingAppSessions.delete(roomId);
}

// Helper function to send signed request to broker
// This will integrate with the existing broker service
async function sendSignedRequestToBroker(signedRequest: any): Promise<any> {
  // Import here to avoid circular dependency
  const { sendToBroker } = await import('./brokerService.ts');
  return sendToBroker(signedRequest);
}
