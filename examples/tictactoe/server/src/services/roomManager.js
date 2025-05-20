/**
 * Room manager for the tic-tac-toe game
 */

import { v4 as uuidv4 } from 'uuid';
import { makeMove } from './index.js';

/**
 * @typedef {Object} Room
 * @property {string} id - Unique room identifier
 * @property {Object} players - Map of player roles
 * @property {string|null} players.host - Host's Ethereum address (X player)
 * @property {string|null} players.guest - Guest's Ethereum address (O player)
 * @property {Map<string, Object>} connections - Map of player connections by EOA
 * @property {Object|null} gameState - Current game state
 * @property {boolean} isReady - Whether the room is ready to start
 */

/**
 * @typedef {Object} RoomManager
 * @property {Map<string, Room>} rooms - Map of active rooms by ID
 * @property {Map<string, string>} addressToRoom - Map of addresses to room IDs
 */

/**
 * Creates a new room manager
 * @returns {RoomManager} Room manager instance
 */
export function createRoomManager() {
  // In-memory storage for rooms and address-to-room mapping
  const rooms = new Map();
  const addressToRoom = new Map();

  /**
   * Creates a new room
   * @returns {string} Room ID
   */
  function createRoom() {
    const roomId = uuidv4();
    rooms.set(roomId, {
      id: roomId,
      players: {
        host: null,
        guest: null
      },
      connections: new Map(),
      gameState: null,
      isReady: false,
      createdAt: Date.now()
    });
    return roomId;
  }

  /**
   * Adds a player to a room
   * @param {string} roomId - Room ID
   * @param {string} eoa - Player's Ethereum address
   * @param {Object} ws - WebSocket connection
   * @returns {Object} Result with success flag and additional info
   */
  function joinRoom(roomId, eoa, ws) {
    // Check if player is already in another room
    if (addressToRoom.has(eoa)) {
      return { 
        success: false, 
        error: 'Player already in another room' 
      };
    }

    // Get the room - if roomId doesn't exist, return error
    if (!rooms.has(roomId)) {
      return {
        success: false,
        error: 'Room not found'
      };
    }
    
    let room = rooms.get(roomId);

    // Check if room is full
    if (room.players.host && room.players.guest) {
      return { 
        success: false, 
        error: 'Room is full' 
      };
    }

    // Assign player to available role
    let role;
    if (!room.players.host) {
      room.players.host = eoa;
      role = 'host';
    } else if (!room.players.guest) {
      room.players.guest = eoa;
      role = 'guest';
    }

    // Store connection and map address to room
    room.connections.set(eoa, { ws, role });
    addressToRoom.set(eoa, roomId);

    // Check if room has both players (ready to potentially start)
    const isRoomFull = room.players.host && room.players.guest;
    
    // Mark as ready when both players are present, but don't auto-start the game
    if (isRoomFull && !room.isReady) {
      room.isReady = true;
      console.log(`Room ${roomId} is ready with players: ${room.players.host} (host) and ${room.players.guest} (guest)`);
      // The game will be created when the host clicks "Start Game"
    }

    return { 
      success: true, 
      roomId, 
      role,
      isRoomReady: isRoomFull
    };
  }

  /**
   * Process a move in a game
   * @param {string} roomId - Room ID
   * @param {number} position - Position on the board (0-8)
   * @param {string} eoa - Player's Ethereum address
   * @returns {Object} Result with success flag and additional info
   */
  function processMove(roomId, position, eoa) {
    if (!rooms.has(roomId)) {
      return { 
        success: false, 
        error: 'Room not found' 
      };
    }

    const room = rooms.get(roomId);
    
    // Check if player is in this room
    if (!room.connections.has(eoa)) {
      return { 
        success: false, 
        error: 'Player not in this room' 
      };
    }

    // Check if the game has started
    if (!room.gameState) {
      return { 
        success: false, 
        error: 'Game has not started' 
      };
    }

    // Make the move
    const result = makeMove(room.gameState, position, eoa);
    if (!result.success) {
      return result;
    }

    // Update game state
    room.gameState = result.gameState;
    
    return {
      success: true,
      gameState: room.gameState,
      isGameOver: room.gameState.isGameOver
    };
  }

  /**
   * Removes a player from a room
   * @param {string} eoa - Player's Ethereum address
   * @returns {Object} Result with success flag and removed room info
   */
  function leaveRoom(eoa) {
    if (!addressToRoom.has(eoa)) {
      return { 
        success: false, 
        error: 'Player not in any room' 
      };
    }

    const roomId = addressToRoom.get(eoa);
    const room = rooms.get(roomId);
    
    // Clean up player connections
    if (room) {
      room.connections.delete(eoa);
      
      // Update player list
      if (room.players.host === eoa) {
        room.players.host = null;
      } else if (room.players.guest === eoa) {
        room.players.guest = null;
      }
      
      // Clean up room if empty
      if (!room.players.host && !room.players.guest) {
        rooms.delete(roomId);
      }
    }
    
    addressToRoom.delete(eoa);
    
    return { 
      success: true, 
      roomId 
    };
  }

  /**
   * Broadcasts a message to all players in a room
   * @param {string} roomId - Room ID
   * @param {string} type - Message type
   * @param {Object} data - Message data
   */
  function broadcastToRoom(roomId, type, data) {
    if (!rooms.has(roomId)) return;
    
    const room = rooms.get(roomId);
    const message = JSON.stringify({ type, ...data });
    
    for (const connection of room.connections.values()) {
      if (connection.ws.readyState === 1) { // WebSocket.OPEN
        connection.ws.send(message);
      }
    }
  }

  /**
   * Closes a room and notifies all players
   * @param {string} roomId - Room ID
   */
  function closeRoom(roomId) {
    if (!rooms.has(roomId)) return;
    
    const room = rooms.get(roomId);
    
    // Remove all players from the room
    for (const eoa of room.connections.keys()) {
      addressToRoom.delete(eoa);
    }
    
    // Delete the room
    rooms.delete(roomId);
  }

  // Return public API
  return {
    rooms,
    addressToRoom,
    createRoom,
    joinRoom,
    processMove,
    leaveRoom,
    broadcastToRoom,
    closeRoom
  };
}