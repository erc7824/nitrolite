import { Hex } from 'viem';
import { ethers } from 'ethers';
import dotenv from 'dotenv';
import { createAppSessionMessage, AppDefinition, CreateAppSessionRequest, MessageSigner } from '@erc7824/nitrolite';
import { SERVER_PRIVATE_KEY } from '../config/index.ts';
import { DEFAULT_PROTOCOL, sendToBroker } from './brokerService.ts';

// Load environment variables
dotenv.config();

// Types
interface AppSession {
  appId: string;
  participantA: Hex;
  participantB: Hex;
  serverAddress: Hex;
  tokenAddress: string;
  createdAt: number;
}

interface PendingAppSession {
  appSessionData: CreateAppSessionRequest[];
  appDefinition: AppDefinition;
  participantA: Hex;
  participantB: Hex;
  serverAddress: Hex;
  signatures: Map<Hex, string>;
  createdAt: number;
  nonce: number;
  requestToSign: any;
  originalSignedMessage: string;
}

// Maps to store app sessions and pending signatures
const roomAppSessions = new Map<string, AppSession>();
const pendingAppSessions = new Map<string, PendingAppSession>();

// Create server wallet for signing
export const serverWallet = new ethers.Wallet(SERVER_PRIVATE_KEY);

// Create a compatible signer function
export const serverSigner: MessageSigner = async (payload: any): Promise<Hex> => {
  const message = typeof payload === 'string' ? payload : JSON.stringify(payload);
  return await serverWallet.signMessage(message) as Hex;
};

/**
 * Generate app session message for multi-signature collection
 */
export async function generateAppSessionMessage(roomId: string, participantA: Hex, participantB: Hex): Promise<{
  appSessionData: CreateAppSessionRequest[];
  appDefinition: AppDefinition;
  participants: Hex[];
  requestToSign: any;
}> {
  try {
    // Format addresses to proper checksum format
    const formattedParticipantA = ethers.getAddress(participantA) as Hex;
    const formattedParticipantB = ethers.getAddress(participantB) as Hex;

    console.log(`Generating app session message for room ${roomId} with participants A: ${formattedParticipantA}, B: ${formattedParticipantB}`);

    // Check if we already have a pending session
    let pendingSession = pendingAppSessions.get(roomId);

    if (pendingSession) {
      console.log(`Using existing app session message for room ${roomId} - nonce: ${pendingSession.nonce}, requestToSign: ${JSON.stringify(pendingSession.requestToSign)}`);
      return {
        appSessionData: pendingSession.appSessionData,
        appDefinition: pendingSession.appDefinition,
        participants: [pendingSession.participantA, pendingSession.participantB, pendingSession.serverAddress],
        requestToSign: pendingSession.requestToSign
      };
    }

    // Get the server's address
    const serverAddress = serverWallet.address as Hex;

    // Create app definition with fixed nonce
    const nonce = Date.now();
    const appDefinition: AppDefinition = {
      protocol: DEFAULT_PROTOCOL,
      participants: [formattedParticipantA, formattedParticipantB, serverAddress],
      weights: [0, 0, 100],
      quorum: 100,
      challenge: 0,
      nonce: nonce,
    };

    const appSessionData: CreateAppSessionRequest[] = [{
      definition: appDefinition,
      allocations: [
        {
          participant: formattedParticipantA,
          asset: 'usdc',
          amount: '0.00001',
        },
        {
          participant: formattedParticipantB,
          asset: 'usdc',
          amount: '0.00001',
        },
        {
          participant: serverAddress,
          asset: 'usdc',
          amount: '0',
        },
      ]
    }];

    // Generate the complete request structure
    const signedMessage = await createAppSessionMessage(serverSigner, appSessionData);
    const parsedMessage = JSON.parse(signedMessage);

    // Extract the request structure that clients should sign
    const requestToSign = parsedMessage.req;

    console.debug(`Generated request structure for room ${roomId}:`, requestToSign);

    // Store the pending app session data
    pendingAppSessions.set(roomId, {
      appSessionData,
      appDefinition,
      participantA: formattedParticipantA,
      participantB: formattedParticipantB,
      serverAddress,
      signatures: new Map(),
      createdAt: Date.now(),
      nonce: nonce,
      requestToSign: requestToSign,
      originalSignedMessage: signedMessage
    });

    console.log(`App session message generated for room ${roomId} with nonce ${nonce}`);
    return {
      appSessionData,
      appDefinition,
      participants: [formattedParticipantA, formattedParticipantB, serverAddress],
      requestToSign: requestToSign
    };

  } catch (error) {
    console.error(`Error generating app session message for room ${roomId}:`, error);
    throw error;
  }
}

/**
 * Add a signature to the pending app session
 */
export async function addAppSessionSignature(roomId: string, participantAddress: Hex, signature: string): Promise<boolean> {
  try {
    // Format the participant address
    const formattedParticipantAddress = ethers.getAddress(participantAddress) as Hex;

    const pendingSession = pendingAppSessions.get(roomId);
    if (!pendingSession) {
      throw new Error(`No pending app session found for room ${roomId}`);
    }

    // Verify the participant is part of this session
    const isValidParticipant = [pendingSession.participantA, pendingSession.participantB].includes(formattedParticipantAddress);
    if (!isValidParticipant) {
      throw new Error(`Invalid participant ${formattedParticipantAddress} for room ${roomId}`);
    }

    // Store the signature
    pendingSession.signatures.set(formattedParticipantAddress, signature);

    console.log(`Added signature for ${formattedParticipantAddress} in room ${roomId} (${pendingSession.signatures.size}/2 collected)`);
    console.debug(`Signature details:`, { participantAddress: formattedParticipantAddress, signature: signature.substring(0, 10) + '...', signatureLength: signature.length });

    // Check if we have all participant signatures
    const allParticipantsSigned = pendingSession.signatures.has(pendingSession.participantA) &&
      pendingSession.signatures.has(pendingSession.participantB);

    return allParticipantsSigned;

  } catch (error) {
    console.error(`Error adding signature for room ${roomId}:`, error);
    throw error;
  }
}

/**
 * Create an app session with collected signatures
 */
export async function createAppSessionWithSignatures(roomId: string): Promise<string> {
  try {
    const pendingSession = pendingAppSessions.get(roomId);
    if (!pendingSession) {
      throw new Error(`No pending app session found for room ${roomId}`);
    }

    // Verify all signatures are collected
    const allSigned = pendingSession.signatures.has(pendingSession.participantA) &&
      pendingSession.signatures.has(pendingSession.participantB);

    if (!allSigned) {
      throw new Error(`Not all signatures collected for room ${roomId}`);
    }

    console.log(`Creating app session with collected signatures for room ${roomId}`);

    // Collect all signatures including server signature
    const participantASignature = pendingSession.signatures.get(pendingSession.participantA);
    const participantBSignature = pendingSession.signatures.get(pendingSession.participantB);

    console.debug(`Participant signatures for room ${roomId}:`, {
      participantA: pendingSession.participantA,
      participantB: pendingSession.participantB,
      participantASignature,
      participantBSignature,
      allStoredSignatures: Array.from(pendingSession.signatures.entries())
    });

    // Validate that we have all participant signatures
    if (!participantASignature) {
      throw new Error(`Missing signature from participant A: ${pendingSession.participantA}`);
    }
    if (!participantBSignature) {
      throw new Error(`Missing signature from participant B: ${pendingSession.participantB}`);
    }

    // Create a properly formatted message with all signatures
    const allSignatures = [participantASignature, participantBSignature];

    // Now let the server sign the same request structure as the clients
    const sign = serverSigner;

    console.debug(`Server signing request structure for room ${roomId}:`, pendingSession.requestToSign);

    // Sign the same request structure that clients signed
    const serverSignature = await sign(pendingSession.requestToSign);

    console.debug(`Server signature created:`, serverSignature);

    // Add server signature to complete the array
    allSignatures.push(serverSignature);

    console.debug(`Combined signatures for room ${roomId}:`, allSignatures);

    // Create the final message with all signatures
    const finalMessage = JSON.parse(pendingSession.originalSignedMessage);
    finalMessage.sig = allSignatures;

    console.debug(`Final message structure:`, {
      req: finalMessage.req,
      signatures: finalMessage.sig,
      participantsOrder: pendingSession.appSessionData[0].definition.participants,
      messageToSend: JSON.stringify(finalMessage)
    });

    console.log("[createAppSessionWithSignatures] Sending request:", finalMessage);
    const result = await sendToBroker(finalMessage);
    const appId = result.app_session_id || (typeof result[0] === "object" ? result[0].app_session_id : null);
    console.log(`[createAppSessionWithSignatures] Created app session ${appId}`);

    // Store the app ID for this room
    roomAppSessions.set(roomId, {
      appId,
      participantA: pendingSession.participantA,
      participantB: pendingSession.participantB,
      serverAddress: pendingSession.serverAddress,
      tokenAddress: process.env.USDC_TOKEN_ADDRESS || '',
      createdAt: Date.now()
    });

    // Clean up pending session
    pendingAppSessions.delete(roomId);

    console.log(`Created app session with ID ${appId} for room ${roomId}`);
    return appId;

  } catch (error) {
    console.error(`Error creating app session with signatures for room ${roomId}:`, error);
    throw error;
  }
}

/**
 * Get the app session for a room
 */
export function getAppSession(roomId: string): AppSession | null {
  return roomAppSessions.get(roomId) || null;
}

/**
 * Get existing pending app session message for a room
 */
export function getPendingAppSessionMessage(roomId: string): {
  appSessionData: CreateAppSessionRequest[];
  appDefinition: AppDefinition;
  participants: Hex[];
  requestToSign: any;
} | null {
  const pendingSession = pendingAppSessions.get(roomId);
  if (!pendingSession) {
    return null;
  }

  return {
    appSessionData: pendingSession.appSessionData,
    appDefinition: pendingSession.appDefinition,
    participants: [pendingSession.participantA, pendingSession.participantB, pendingSession.serverAddress],
    requestToSign: pendingSession.requestToSign
  };
}

/**
 * Check if a room has an app session
 */
export function hasAppSession(roomId: string): boolean {
  return roomAppSessions.has(roomId);
}

/**
 * Get all app sessions
 */
export function getAllAppSessions(): Map<string, AppSession> {
  return roomAppSessions;
}
