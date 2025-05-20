/**
 * Nitrolite app sessions for game rooms
 * This file handles creating and closing app sessions for games
 */
import { NitroliteRPC, createAppSessionMessage, createCloseAppSessionMessage } from '@erc7824/nitrolite';
import dotenv from 'dotenv';
import logger from '../utils/logger.js';
import { getRPCClient } from './nitroliteRPC.js';

// Load environment variables
dotenv.config();

// Map to store app sessions by room ID
const roomAppSessions = new Map();

/**
 * Create an app session for a game room
 * @param {string} roomId - Room ID
 * @param {string} participantA - First player's address
 * @param {string} participantB - Second player's address
 * @returns {Promise<string>} The app session ID
 */
export async function createAppSession(roomId, participantA, participantB) {
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
              amount: amount,
            },
            {
              participant: participantB,
              asset: 'usdc',
              amount: '0',
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
 * Close an app session for a game room
 * @param {string} roomId - Room ID
 * @param {Array<number>} [allocations=[0,0,0]] - Final allocations
 * @returns {Promise<boolean>} Success status
 */
export async function closeAppSession(roomId, allocations = [0, 0, 0]) {
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
    
    // Final allocations and close request
    const closeRequest = {
      app_session_id: appId,
      allocations: finalAllocations,
    };
    const finalIntent = finalAllocations;
    
    // Use the RPC client's signMessage method for consistent signing
    const sign = rpcClient.signMessage.bind(rpcClient);
    
    // Create the signed message
    const signedMessage = await createCloseAppSessionMessage(
      sign, 
      [closeRequest], 
      finalIntent
    );
    
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
