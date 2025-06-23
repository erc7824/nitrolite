/**
 * Nitrolite app sessions for game rooms
 * This file handles creating and closing app sessions for games
 */
import { createAppSessionMessage, createCloseAppSessionMessage } from '@erc7824/nitrolite';
import { ethers } from 'ethers';
import dotenv from 'dotenv';
import logger from '../utils/logger.js';
import { getRPCClient } from './nitroliteRPC.js';

// Load environment variables
dotenv.config();

// Map to store app sessions by room ID
const roomAppSessions = new Map();

// Map to store pending app session signatures by room ID
const pendingAppSessions = new Map();

/**
 * Generate app session message for multi-signature collection
 * @param {string} roomId - Room ID
 * @param {string} participantA - First player's address
 * @param {string} participantB - Second player's address
 * @param {number} betAmount - Bet amount for each player (0, 0.01, 0.1, 1, 2)
 * @returns {Promise<Object>} The unsigned app session message and app definition
 */
export async function generateAppSessionMessage(roomId, participantA, participantB, betAmount = 0) {
  try {
    // Format addresses to proper checksum format
    const formattedParticipantA = ethers.getAddress(participantA);
    const formattedParticipantB = ethers.getAddress(participantB);
    
    logger.nitro(`Generating app session message for room ${roomId} with participants A: ${formattedParticipantA}, B: ${formattedParticipantB}`);
    
    // Check if we already have a pending session (to ensure consistency)
    let pendingSession = pendingAppSessions.get(roomId);
    
    if (pendingSession) {
      logger.nitro(`Using existing app session message for room ${roomId} - nonce: ${pendingSession.nonce}, requestToSign: ${JSON.stringify(pendingSession.requestToSign)}`);
      return {
        appSessionData: pendingSession.appSessionData,
        appDefinition: pendingSession.appDefinition,
        participants: [pendingSession.participantA, pendingSession.participantB, pendingSession.serverAddress],
        requestToSign: pendingSession.requestToSign
      };
    }
    
    // Get the RPC client
    const rpcClient = await getRPCClient();
    if (!rpcClient) {
      throw new Error('RPC client not initialized');
    }
    
    // Get the server's address and format it
    const serverAddress = ethers.getAddress(rpcClient.address);
    
    // Create app definition with fixed nonce to ensure all participants sign the same message
    const nonce = Date.now();
    const appDefinition = {
      protocol: "viper_duel_nitrolite_v0",
      participants: [formattedParticipantA, formattedParticipantB, serverAddress],
      weights: [0, 0, 100],
      quorum: 100,
      challenge: 0,
      nonce: nonce,
    };
    
    const appSessionData = [{
      definition: appDefinition,
      allocations: [
        {
          participant: formattedParticipantA,
          asset: 'usdc',
          amount: betAmount.toString(),
        },
        {
          participant: formattedParticipantB,
          asset: 'usdc',
          amount: betAmount.toString(),
        },
        {
          participant: serverAddress,
          asset: 'usdc',
          amount: '0',
        },
      ]
    }];
    
    // Generate the complete request structure that everyone will sign
    const sign = rpcClient.signMessage.bind(rpcClient);
    const signedMessage = await createAppSessionMessage(sign, appSessionData);
    const parsedMessage = JSON.parse(signedMessage);
    
    // Extract the request structure that clients should sign (same as what server signs)
    const requestToSign = parsedMessage.req;
    
    logger.data(`Generated request structure for room ${roomId}:`, requestToSign);
    
    // Store the pending app session data including the request structure
    pendingAppSessions.set(roomId, {
      appSessionData,
      appDefinition,
      participantA: formattedParticipantA,
      participantB: formattedParticipantB,
      serverAddress,
      betAmount,
      signatures: new Map(),
      createdAt: Date.now(),
      nonce: nonce,
      requestToSign: requestToSign,
      originalSignedMessage: signedMessage
    });
    
    logger.nitro(`App session message generated for room ${roomId} with nonce ${nonce}`);
    return {
      appSessionData,
      appDefinition,
      participants: [formattedParticipantA, formattedParticipantB, serverAddress],
      requestToSign: requestToSign
    };
    
  } catch (error) {
    logger.error(`Error generating app session message for room ${roomId}:`, error);
    throw error;
  }
}

/**
 * Add a signature to the pending app session
 * @param {string} roomId - Room ID
 * @param {string} participantAddress - Address of the signing participant
 * @param {string} signature - The participant's signature
 * @returns {Promise<boolean>} Whether all signatures are collected
 */
export async function addAppSessionSignature(roomId, participantAddress, signature) {
  try {
    // Format the participant address to proper checksum format
    const formattedParticipantAddress = ethers.getAddress(participantAddress);
    
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
    
    logger.nitro(`Added signature for ${formattedParticipantAddress} in room ${roomId} (${pendingSession.signatures.size}/2 collected)`);
    logger.data(`Signature details:`, { participantAddress: formattedParticipantAddress, signature: signature.substring(0, 10) + '...', signatureLength: signature.length });
    
    // Check if we have all participant signatures (not including server)
    const allParticipantsSigned = pendingSession.signatures.has(pendingSession.participantA) && 
                                  pendingSession.signatures.has(pendingSession.participantB);
    
    return allParticipantsSigned;
    
  } catch (error) {
    logger.error(`Error adding signature for room ${roomId}:`, error);
    throw error;
  }
}

/**
 * Create an app session with collected signatures
 * @param {string} roomId - Room ID
 * @returns {Promise<string>} The app session ID
 */
export async function createAppSessionWithSignatures(roomId) {
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
    
    logger.nitro(`Creating app session with collected signatures for room ${roomId}`);
    
    // Get the RPC client
    const rpcClient = await getRPCClient();
    if (!rpcClient) {
      throw new Error('RPC client not initialized');
    }
    
    // Collect all signatures including server signature
    const participantASignature = pendingSession.signatures.get(pendingSession.participantA);
    const participantBSignature = pendingSession.signatures.get(pendingSession.participantB);
    
    logger.data(`Participant signatures for room ${roomId}:`, {
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
    
    // Don't create a new server signature - use the existing signed message structure
    // but replace the single server signature with all collected signatures
    
    // Create a properly formatted message with all signatures
    // The signatures should be in the same order as participants: [participantA, participantB, server]
    const allSignatures = [participantASignature, participantBSignature];
    
    // Now let the server sign the same request structure as the clients
    const sign = rpcClient.signMessage.bind(rpcClient);
    
    logger.data(`Server signing request structure for room ${roomId}:`, pendingSession.requestToSign);
    
    // Sign the same request structure that clients signed
    const serverSignature = await sign(pendingSession.requestToSign);
    
    logger.data(`Server signature created:`, serverSignature);
    
    // Add server signature to complete the array
    allSignatures.push(serverSignature);
    
    logger.data(`Combined signatures for room ${roomId}:`, allSignatures);
    
    // Send the message directly using ws.send
    logger.nitro(`Sending app session creation message for room ${roomId}`);
    
    if (!rpcClient.ws || rpcClient.ws.readyState !== 1) {
      throw new Error('WebSocket not connected or not in OPEN state');
    }
    
    // Create the final message with all signatures
    const finalMessage = JSON.parse(pendingSession.originalSignedMessage);
    finalMessage.sig = allSignatures;
    
    logger.data(`Final message structure:`, {
      req: finalMessage.req,
      signatures: finalMessage.sig,
      participantsOrder: pendingSession.appSessionData[0].definition.participants,
      messageToSend: JSON.stringify(finalMessage)
    });
    
    // Set up a promise to handle the response from the WebSocket
    const appSessionResponsePromise = new Promise((resolve, reject) => {
      const handleAppSessionResponse = (data) => {
        try {
          const rawData = typeof data === 'string' ? data : data.toString();
          const message = JSON.parse(rawData);
          
          logger.data(`Received app session creation response:`, message);
          
          if (message.res && (message.res[1] === 'create_app_session' || 
                             message.res[1] === 'app_session_created')) {
            rpcClient.ws.removeListener('message', handleAppSessionResponse);
            resolve(message.res[2]);
          }
          
          if (message.err) {
            rpcClient.ws.removeListener('message', handleAppSessionResponse);
            reject(new Error(`Error ${message.err[1]}: ${message.err[2]}`));
          }
        } catch (error) {
          logger.error('Error handling app session response:', error);
        }
      };
      
      rpcClient.ws.on('message', handleAppSessionResponse);
      
      setTimeout(() => {
        rpcClient.ws.removeListener('message', handleAppSessionResponse);
        reject(new Error('App session creation timeout'));
      }, 10000);
    });
    
    // Send the final message (convert to string)
    rpcClient.ws.send(JSON.stringify(finalMessage));
    
    // Wait for the response
    const response = await appSessionResponsePromise;
    
    logger.data(`App session creation response for room ${roomId}:`, response);
    
    const appId = response?.app_session_id || response?.[0]?.app_session_id;
    
    if (!appId) {
      throw new Error('Failed to get app ID from response');
    }
    
    // Store the app ID for this room
    roomAppSessions.set(roomId, {
      appId,
      participantA: pendingSession.participantA,
      participantB: pendingSession.participantB,
      serverAddress: pendingSession.serverAddress,
      tokenAddress: process.env.USDC_TOKEN_ADDRESS,
      betAmount: pendingSession.betAmount,
      createdAt: Date.now()
    });
    
    // Clean up pending session
    pendingAppSessions.delete(roomId);
    
    logger.nitro(`Created app session with ID ${appId} for room ${roomId}`);
    return appId;
    
  } catch (error) {
    logger.error(`Error creating app session with signatures for room ${roomId}:`, error);
    throw error;
  }
}

/**
 * Create an app session for a game room (original function for backward compatibility)
 * @param {string} roomId - Room ID
 * @param {string} participantA - First player's address
 * @param {string} participantB - Second player's address
 * @param {number} betAmount - Bet amount for each player (0, 0.01, 0.1, 1, 2)
 * @returns {Promise<string>} The app session ID
 */
export async function createAppSession(roomId, participantA, participantB, betAmount = 0) {
  try {
    logger.nitro(`Creating app session for room ${roomId}`);
    
    // Get the RPC client
    const rpcClient = await getRPCClient();
    if (!rpcClient) {
      throw new Error('RPC client not initialized');
    }
    
    // Get the server's address
    const serverAddress = rpcClient.address;
    
    // Check if token address is available
    const tokenAddress = process.env.USDC_TOKEN_ADDRESS;
    if (!tokenAddress) {
      throw new Error('Token address not set in environment variables');
    }
    
    // Define the deposit amount (use '0' for free games or actual amount for paid games)
    const amount = '0'; // Set this to the appropriate amount if needed
    
    // Create app definition
    const appDefinition = {
      protocol: "app_aura_nitrolite_v0",
      participants: [participantA, participantB, serverAddress],
      weights: [0, 0, 100],
      quorum: 100,
      challenge: 0,
      nonce: Date.now(),
    };
    
    // Use the RPC client's signMessage method for consistent signing
    const sign = rpcClient.signMessage.bind(rpcClient);
    
    // Create the signed message
    const signedMessage = await createAppSessionMessage(
      sign,
      [
        {
          definition: appDefinition,
          allocations: [
            {
              participant: participantA,
              asset: 'usdc',
              amount: betAmount.toString(),
            },
            {
              participant: participantB,
              asset: 'usdc',
              amount: betAmount.toString(),
            },
            {
              participant: serverAddress,
              asset: 'usdc',
              amount: '0',
            },
          ]
        },
      ]
    );
    logger.data(`Signed app session message for room ${roomId}:`, signedMessage);
    // Send the message directly using ws.send, similar to authentication
    logger.nitro(`Sending app session creation message for room ${roomId}`);
    
    if (!rpcClient.ws || rpcClient.ws.readyState !== 1) { // WebSocket.OPEN
      throw new Error('WebSocket not connected or not in OPEN state');
    }
    
    // Set up a promise to handle the response from the WebSocket
    const appSessionResponsePromise = new Promise((resolve, reject) => {
      // Create a one-time message handler for the app session response
      const handleAppSessionResponse = (data) => {
        try {
          const rawData = typeof data === 'string' ? data : data.toString();
          const message = JSON.parse(rawData);
          
          logger.data(`Received app session creation response:`, message);
          
          // Check if this is an app session response
          if (message.res && (message.res[1] === 'create_app_session' || 
                             message.res[1] === 'app_session_created')) {
            // Remove the listener once we get the response
            rpcClient.ws.removeListener('message', handleAppSessionResponse);
            resolve(message.res[2]); // The app session data should be in the 3rd position
          }
          
          // Also check for error responses
          if (message.err) {
            rpcClient.ws.removeListener('message', handleAppSessionResponse);
            reject(new Error(`Error ${message.err[1]}: ${message.err[2]}`));
          }
        } catch (error) {
          logger.error('Error handling app session response:', error);
        }
      };
      
      // Add the message handler
      rpcClient.ws.on('message', handleAppSessionResponse);
      
      // Set timeout to prevent hanging
      setTimeout(() => {
        rpcClient.ws.removeListener('message', handleAppSessionResponse);
        reject(new Error('App session creation timeout'));
      }, 10000);
    });
    
    // Send the signed message directly
    rpcClient.ws.send(signedMessage);
    
    // Wait for the response
    const response = await appSessionResponsePromise;
    
    // Log the response
    logger.data(`App session creation response for room ${roomId}:`, response);
    
    // The response structure might vary, adapt this based on actual response
    const appId = response?.app_session_id || response?.[0]?.app_session_id;
    
    if (!appId) {
      throw new Error('Failed to get app ID from response');
    }
    
    // Store the app ID for this room
    roomAppSessions.set(roomId, {
      appId,
      participantA,
      participantB,
      serverAddress,
      tokenAddress,
      betAmount,
      createdAt: Date.now()
    });
    
    logger.nitro(`Created app session with ID ${appId} for room ${roomId}`);
    return appId;
    
  } catch (error) {
    logger.error(`Error creating app session for room ${roomId}:`, error);
    throw error;
  }
}

/**
 * Close an app session with winner taking the allocation
 * @param {string} roomId - Room ID
 * @param {string} winnerId - Winner's participant ID ('A' or 'B'), null for tie
 * @returns {Promise<boolean>} Success status
 */
export async function closeAppSessionWithWinner(roomId, winnerId = null) {
  try {
    // Get the app session for this room
    const appSession = roomAppSessions.get(roomId);
    if (!appSession) {
      logger.warn(`No app session found for room ${roomId}`);
      return false;
    }

    const { participantA, participantB, betAmount = 0 } = appSession;
    
    // Calculate allocations based on winner and bet amount
    // Need to ensure exact redistribution of what was initially allocated
    const betAmountStr = betAmount.toString();
    const totalPot = (betAmount * 2).toString();
    
    let allocations;
    if (winnerId === 'A') {
      // Player A wins - gets all the funds
      allocations = [totalPot, '0', '0']; // A gets both initial allocations
      logger.nitro(`Player A (${participantA}) wins room ${roomId} - taking full allocation of ${totalPot}`);
    } else if (winnerId === 'B') {
      // Player B wins - gets all the funds
      allocations = ['0', totalPot, '0']; // B gets both initial allocations
      logger.nitro(`Player B (${participantB}) wins room ${roomId} - taking full allocation of ${totalPot}`);
    } else {
      // Tie or no winner - split evenly (return original amounts)
      allocations = [betAmountStr, betAmountStr, '0'];
      logger.nitro(`Tie in room ${roomId} - splitting allocation evenly: ${betAmountStr} each`);
    }
    
    logger.data(`Allocation calculation for room ${roomId}:`, {
      betAmount,
      betAmountStr,
      totalPot,
      winnerId,
      finalAllocations: allocations
    });

    // Use the existing closeAppSession function with calculated allocations
    return await closeAppSession(roomId, allocations);
    
  } catch (error) {
    logger.error(`Error closing app session with winner for room ${roomId}:`, error);
    return false;
  }
}

/**
 * Close an app session for a game room
 * @param {string} roomId - Room ID
 * @param {Array<number>} [allocations=[0,0,0]] - Final allocations
 * @returns {Promise<boolean>} Success status
 */
export async function closeAppSession(roomId, allocations) {
  try {
    // Get the app session for this room
    const appSession = roomAppSessions.get(roomId);
    if (!appSession) {
      logger.warn(`No app session found for room ${roomId}`);
      return false;
    }
    
    // Make sure appId exists and is properly extracted
    const appId = appSession.appId;
    if (!appId) {
      logger.error(`No appId found in app session for room ${roomId}`);
      return false;
    }
    
    logger.nitro(`Closing app session ${appId} for room ${roomId}`);
    
    // Get the RPC client
    const rpcClient = await getRPCClient();
    if (!rpcClient) {
      throw new Error('RPC client not initialized');
    }

    // Extract participant addresses from the stored app session
    const { participantA, participantB, serverAddress } = appSession;

    // Check if we have all the required participants
    if (!participantA || !participantB || !serverAddress) {
      throw new Error('Missing participant information in app session');
    }

    const finalAllocations = [
      {
        participant: participantA,
        asset: 'usdc',
        amount: allocations[0].toString(),
      },
      {
        participant: participantB,
        asset: 'usdc',
        amount: allocations[1].toString(),
      },
      {
        participant: serverAddress,
        asset: 'usdc',
        amount: allocations[2].toString(),
      },
    ];
    
    // Log the allocations for debugging
    const totalAmount = parseFloat(allocations[0]) + parseFloat(allocations[1]) + parseFloat(allocations[2]);
    const expectedTotal = (appSession.betAmount || 0) * 2;
    logger.data(`Final allocations for room ${roomId}:`, {
      participantA: `${participantA} = ${allocations[0]}`,
      participantB: `${participantB} = ${allocations[1]}`,
      serverAddress: `${serverAddress} = ${allocations[2]}`,
      totalAmount: totalAmount.toString(),
      expectedTotal: expectedTotal.toString()
    });
    
    // Final allocations and close request
    const closeRequest = {
      app_session_id: appId,
      allocations: finalAllocations,
    };
    
    // Use the RPC client's signMessage method for consistent signing
    const sign = rpcClient.signMessage.bind(rpcClient);
    
    // Create the signed message
    const signedMessage = await createCloseAppSessionMessage(
      sign, 
      [closeRequest], 
    );

    logger.data(`Signed app session close message for room ${roomId}:`, signedMessage);
    
    // Send the message directly using ws.send, similar to authentication
    logger.nitro(`Sending app session close message for room ${roomId}`);
    
    if (!rpcClient.ws || rpcClient.ws.readyState !== 1) { // WebSocket.OPEN
      throw new Error('WebSocket not connected or not in OPEN state');
    }
    
    // Set up a promise to handle the response from the WebSocket
    const closeSessionResponsePromise = new Promise((resolve, reject) => {
      // Create a one-time message handler for the close session response
      const handleCloseSessionResponse = (data) => {
        try {
          const rawData = typeof data === 'string' ? data : data.toString();
          const message = JSON.parse(rawData);
          
          logger.data(`Received close session response:`, message);
          
          // Check if this is a close session response
          if (message.res && (message.res[1] === 'close_app_session' || 
                             message.res[1] === 'app_session_closed')) {
            // Remove the listener once we get the response
            rpcClient.ws.removeListener('message', handleCloseSessionResponse);
            resolve(message.res[2]);
          }
          
          // Also check for error responses
          if (message.err) {
            rpcClient.ws.removeListener('message', handleCloseSessionResponse);
            reject(new Error(`Error ${message.err[1]}: ${message.err[2]}`));
          }
        } catch (error) {
          logger.error('Error handling close session response:', error);
        }
      };
      
      // Add the message handler
      rpcClient.ws.on('message', handleCloseSessionResponse);
      
      // Set timeout to prevent hanging
      setTimeout(() => {
        rpcClient.ws.removeListener('message', handleCloseSessionResponse);
        reject(new Error('Close session timeout'));
      }, 10000);
    });
    
    // Send the signed message directly
    rpcClient.ws.send(signedMessage);
    
    // Wait for the response
    const response = await closeSessionResponsePromise;
    
    // Log the response
    logger.data(`App session close response for room ${roomId}:`, response);
    
    // Remove the app session
    roomAppSessions.delete(roomId);
    
    logger.nitro(`Closed app session ${appId} for room ${roomId}`);
    return true;
    
  } catch (error) {
    logger.error(`Error closing app session for room ${roomId}:`, error);
    return false;
  }
}

/**
 * Get the app session for a room
 * @param {string} roomId - Room ID
 * @returns {Object|null} The app session or null if not found
 */
export function getAppSession(roomId) {
  return roomAppSessions.get(roomId) || null;
}

/**
 * Get existing pending app session message for a room
 * @param {string} roomId - Room ID
 * @returns {Object|null} The existing app session message or null if not found
 */
export function getPendingAppSessionMessage(roomId) {
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
 * @param {string} roomId - Room ID
 * @returns {boolean} Whether the room has an app session
 */
export function hasAppSession(roomId) {
  return roomAppSessions.has(roomId);
}

/**
 * Get all app sessions
 * @returns {Map} Map of all app sessions
 */
export function getAllAppSessions() {
  return roomAppSessions;
}
