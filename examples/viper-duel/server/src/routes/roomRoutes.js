/**
 * Room-related WebSocket message handlers
 */

import { validateJoinRoomPayload } from '../utils/validators.js';
import { formatGameState, generateAppSessionMessage } from '../services/index.js';
import logger from '../utils/logger.js';

/**
 * Handles a request to join a room
 * @param {WebSocket} ws - WebSocket connection
 * @param {Object} payload - Request payload
 * @param {Object} context - Application context containing roomManager and connections
 */
export async function handleJoinRoom(ws, payload, { roomManager, connections, sendError }) {
  // Validate payload
  const validation = validateJoinRoomPayload(payload);
  if (!validation.success) {
    return sendError(ws, 'INVALID_PAYLOAD', validation.error);
  }

  const { roomId, eoa, betAmount = 0 } = payload;
  console.log(`Processing ${validation.isCreating ? 'CREATE' : 'JOIN'} request for EOA: ${eoa}, roomId: ${roomId || 'NEW'}, betAmount: ${betAmount}`);

  // Validate bet amount
  const validBetAmounts = [0, 0.01, 0.1, 1, 2];
  if (!validBetAmounts.includes(betAmount)) {
    return sendError(ws, 'INVALID_BET_AMOUNT', 'Invalid bet amount. Must be 0, 0.01, 0.1, 1, or 2');
  }

  // Check if address is already connected
  if (connections.has(eoa)) {
    return sendError(ws, 'ALREADY_CONNECTED', 'Address already connected');
  }

  let result;
  if (validation.isCreating) {
    // Creating a new room
    const newRoomId = roomManager.createRoom(betAmount);
    console.log(`Created new room with ID: ${newRoomId}, bet amount: ${betAmount}`);
    
    // Join the newly created room as host
    result = roomManager.joinRoom(newRoomId, eoa, ws, betAmount);
    
    if (result.success) {
      console.log(`New room created: ${newRoomId} for player (host): ${eoa}`);
      
      // Send room ID to client immediately so they can share it
      ws.send(JSON.stringify({
        type: 'room:created',
        roomId: newRoomId,
        role: 'host'
      }));
    }
  } else {
    // Joining an existing room
    result = roomManager.joinRoom(roomId, eoa, ws, betAmount);
    
    if (result.success) {
      console.log(`Player ${eoa} joined room: ${roomId} as ${result.role}`);
    }
  }
  
  if (!result.success) {
    return sendError(ws, 'JOIN_FAILED', result.error);
  }

  // Store connection
  connections.set(eoa, { ws, roomId: result.roomId });

  // Get room
  const room = roomManager.rooms.get(result.roomId);

  // Send room state to all players
  if (room.gameState) {
    roomManager.broadcastToRoom(
      result.roomId, 
      'room:state', 
      formatGameState(room.gameState, result.roomId, room.betAmount)
    );
  }

  // Notify all players that room is ready if applicable
  if (result.isRoomReady) {
    roomManager.broadcastToRoom(result.roomId, 'room:ready', { roomId: result.roomId });
    
    logger.nitro(`Room ${result.roomId} is ready - starting signature collection flow`);
    logger.data(`Room players:`, { host: room.players.host, guest: room.players.guest });
    
    // Generate app session message for signature collection when room becomes ready (both players joined)
    try {
      const appSessionMessage = await generateAppSessionMessage(
        result.roomId, 
        room.players.host, 
        room.players.guest,
        room.betAmount
      );
      
      logger.nitro(`Generated app session message for room ${result.roomId}`);
      
      // Send the message to participant B (guest) for signature
      const guestConnection = room.connections.get(room.players.guest);
      if (guestConnection && guestConnection.ws.readyState === 1) {
        guestConnection.ws.send(JSON.stringify({
          type: 'appSession:signatureRequest',
          roomId: result.roomId,
          appSessionData: appSessionMessage.appSessionData,
          appDefinition: appSessionMessage.appDefinition,
          participants: appSessionMessage.participants,
          requestToSign: appSessionMessage.requestToSign
        }));
      }
      
    } catch (error) {
      logger.error(`Failed to generate app session message for room ${result.roomId}:`, error);
    }
  }
}

/**
 * Handles a request to get available rooms
 * @param {WebSocket} ws - WebSocket connection
 * @param {Object} context - Application context containing roomManager
 */
export async function handleGetAvailableRooms(ws, { roomManager }) {
  // Filter rooms that are not full
  const availableRooms = [];
  
  // Get current timestamp
  const now = Date.now();
  
  // Iterate through all rooms and find available ones
  for (const [roomId, room] of roomManager.rooms.entries()) {
    // Room is available if it has a host but no guest, and game is not started
    if (room.players.host && !room.players.guest && !room.gameState) {
      availableRooms.push({
        roomId,
        hostAddress: room.players.host,
        createdAt: room.createdAt || now, // Use tracked creation time or fall back to now
        betAmount: room.betAmount || 0
      });
    }
  }
  
  // Send available rooms to client
  ws.send(JSON.stringify({
    type: 'room:available',
    rooms: availableRooms
  }));
}