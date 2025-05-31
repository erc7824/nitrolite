/**
 * Game-related WebSocket message handlers
 */

import { validateMovePayload } from '../utils/validators.js';
import { 
  formatGameState, 
  formatGameOverMessage, 
  createGame, 
  createAppSession,
  closeAppSession,
  hasAppSession,
  generateAppSessionMessage,
  addAppSessionSignature,
  createAppSessionWithSignatures
} from '../services/index.js';
import logger from '../utils/logger.js';

/**
 * Handles a start game request
 * @param {WebSocket} ws - WebSocket connection
 * @param {Object} payload - Request payload
 * @param {Object} context - Application context containing roomManager and connections
 */
export async function handleStartGame(ws, payload, { roomManager, connections, sendError }) {
  if (!payload || typeof payload !== 'object') {
    return sendError(ws, 'INVALID_PAYLOAD', 'Invalid payload format');
  }

  const { roomId } = payload;

  if (!roomId) {
    return sendError(ws, 'INVALID_PAYLOAD', 'Room ID is required');
  }

  // Find the player trying to start the game
  let playerEoa = null;
  for (const [eoa, connection] of connections.entries()) {
    if (connection.ws === ws) {
      playerEoa = eoa;
      break;
    }
  }

  if (!playerEoa) {
    return sendError(ws, 'NOT_AUTHENTICATED', 'Player not authenticated');
  }

  // Get the room
  const room = roomManager.rooms.get(roomId);
  if (!room) {
    return sendError(ws, 'ROOM_NOT_FOUND', 'Room not found');
  }

  // Only the host can start the game
  if (room.players.host !== playerEoa) {
    return sendError(ws, 'NOT_AUTHORIZED', 'Only the host can start the game');
  }

  // Need both players
  if (!room.players.host || !room.players.guest) {
    return sendError(ws, 'ROOM_NOT_FULL', 'Room must have two players to start the game');
  }

  // Initialize game state if not already done
  if (!room.gameState) {
    room.gameState = createGame(room.players.host, room.players.guest);
  }

  // Create an app session for this game if not already created
  if (!hasAppSession(roomId)) {
    try {
      logger.nitro(`Creating app session for room ${roomId}`);
      const appId = await createAppSession(roomId, room.players.host, room.players.guest);
      logger.nitro(`App session created with ID ${appId}`);
      
      // Store the app ID in the room object
      room.appId = appId;
    } catch (error) {
      logger.error(`Failed to create app session for room ${roomId}:`, error);
      // Continue with the game even if app session creation fails
      // This allows the game to work in a fallback mode
    }
  }

  // Broadcast game started
  roomManager.broadcastToRoom(
    roomId,
    'game:started',
    { roomId, firstTurn: 'X' }
  );

  // Send the initial game state
  roomManager.broadcastToRoom(
    roomId, 
    'room:state', 
    formatGameState(room.gameState, roomId)
  );
}

/**
 * Handles a move request
 * @param {WebSocket} ws - WebSocket connection
 * @param {Object} payload - Request payload
 * @param {Object} context - Application context containing roomManager and connections
 */
export async function handleMove(ws, payload, { roomManager, connections, sendError }) {
  // Validate payload
  const validation = validateMovePayload(payload);
  if (!validation.success) {
    return sendError(ws, 'INVALID_PAYLOAD', validation.error);
  }

  const { roomId, pos } = payload;
  
  // Find the player making the move
  let playerEoa = null;
  for (const [eoa, connection] of connections.entries()) {
    if (connection.ws === ws) {
      playerEoa = eoa;
      break;
    }
  }

  if (!playerEoa) {
    return sendError(ws, 'NOT_AUTHENTICATED', 'Player not authenticated');
  }

  // Process the move
  const result = roomManager.processMove(roomId, pos, playerEoa);
  if (!result.success) {
    return sendError(ws, 'MOVE_FAILED', result.error);
  }

  // Broadcast updated game state
  roomManager.broadcastToRoom(
    roomId, 
    'room:state', 
    formatGameState(result.gameState, roomId)
  );

  // Handle game over condition
  if (result.isGameOver) {
    roomManager.broadcastToRoom(
      roomId, 
      'game:over', 
      formatGameOverMessage(result.gameState)
    );

    // Close the app session if one was created
    try {
      const room = roomManager.rooms.get(roomId);
      
      // First check if the room has an appId directly
      if (room && room.appId) {
        logger.nitro(`Closing app session with ID ${room.appId} for room ${roomId}`);
        
        // Determine winner based on game result
        let winnerId = null;
        if (result.gameState.winner === 'X') {
          winnerId = 'A'; // X is player A (host)
        } else if (result.gameState.winner === 'O') {
          winnerId = 'B'; // O is player B (guest)
        }
        // null winner means tie
        
        // Calculate allocations based on winner
        let finalAllocations;
        if (winnerId === 'A') {
          // Player A wins - gets all the funds
          finalAllocations = ['0.02', '0', '0']; // A gets both initial allocations
        } else if (winnerId === 'B') {
          // Player B wins - gets all the funds
          finalAllocations = ['0', '0.02', '0']; // B gets both initial allocations
        } else {
          // Tie or no winner - split evenly
          finalAllocations = ['0.01', '0.01', '0'];
        }
        
        await closeAppSession(roomId, finalAllocations);
        logger.nitro(`App session closed for room ${roomId}`);
      } 
      // Otherwise check the app sessions storage
      else if (hasAppSession(roomId)) {
        logger.nitro(`Closing app session from storage for room ${roomId}`);
        
        // Determine winner based on game result
        let winnerId = null;
        if (result.gameState.winner === 'X') {
          winnerId = 'A'; // X is player A (host)
        } else if (result.gameState.winner === 'O') {
          winnerId = 'B'; // O is player B (guest)
        }
        // null winner means tie
        
        // Calculate allocations based on winner
        let finalAllocations;
        if (winnerId === 'A') {
          // Player A wins - gets all the funds
          finalAllocations = ['0.02', '0', '0']; // A gets both initial allocations
        } else if (winnerId === 'B') {
          // Player B wins - gets all the funds
          finalAllocations = ['0', '0.02', '0']; // B gets both initial allocations
        } else {
          // Tie or no winner - split evenly
          finalAllocations = ['0.01', '0.01', '0'];
        }
        
        await closeAppSession(roomId, finalAllocations);
        logger.nitro(`App session closed for room ${roomId}`);
      }
    } catch (error) {
      logger.error(`Failed to close app session for room ${roomId}:`, error);
      // Continue with room cleanup even if app session closure fails
    }

    // Clean up the room after a short delay
    setTimeout(() => {
      roomManager.closeRoom(roomId);
    }, 5000);
  }
}