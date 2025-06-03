import { WebSocketServer } from 'ws';
import WebSocket from 'ws';
import { randomBytes } from 'crypto';
import { Room, SnakeWebSocket } from '../interfaces/index.ts';
import { getRoom, addRoom, removeRoom, getAllRooms } from './stateService.ts';
import {
  generateRoomId,
  generateFood,
  initializePlayer,
  gameTick,
  clearNetRPC,
  initializeBroadcastFunction
} from './gameService.ts';
import { createAppSession, closeAppSession } from './brokerService.ts';
import { Hex } from 'viem';

// Global reference to the WebSocket server
let webSocketServer: WebSocketServer;

// Store clients subscribed to room updates
const roomSubscribers = new Set<SnakeWebSocket>();

// Setup WebSocket handlers
export function setupWebSocketHandlers(wss: WebSocketServer): void {
  webSocketServer = wss;

  // Initialize the broadcast function in gameService
  initializeBroadcastFunction(broadcastGameState);

  wss.on('connection', (ws: WebSocket) => {
    console.log('Client connected');

    const snakeWs = ws as SnakeWebSocket;
    snakeWs.playerId = randomBytes(8).toString('hex');

    snakeWs.on('message', async (message: WebSocket.RawData) => {
      try {
        const data = JSON.parse(message.toString());
        await handleWebSocketMessage(snakeWs, data);
      } catch (error) {
        console.error('Error handling message:', error);
      }
    });

    snakeWs.on('close', async () => {
      // Remove from room subscribers if they were subscribed
      roomSubscribers.delete(snakeWs);
      await handleDisconnect(snakeWs);
    });
  });
}

// Get WebSocketServer instance
export function getWebSocketServer(): WebSocketServer {
  return webSocketServer;
}

// Broadcast game state to all clients in a room
export function broadcastGameState(roomId: string, gameState: any): void {
  console.log(`[websocketService] Broadcasting game state to room ${roomId} at ${Date.now()}`);
  console.log(`[websocketService] Game state version: ${gameState.stateVersion}`);
  console.log(`[websocketService] Game over status: ${gameState.isGameOver}`);

  let clientCount = 0;
  webSocketServer.clients.forEach(client => {
    const snakeClient = client as SnakeWebSocket;
    if (snakeClient.roomId === roomId && snakeClient.readyState === WebSocket.OPEN) {
      clientCount++;
      console.log(`[websocketService] Sending to client ${snakeClient.playerId} at ${Date.now()}`);
      snakeClient.send(JSON.stringify(gameState));
    }
  });
  console.log(`[websocketService] Broadcast complete. Sent to ${clientCount} clients at ${Date.now()}`);
}

// Handle WebSocket message
async function handleWebSocketMessage(ws: SnakeWebSocket, data: any): Promise<void> {
  switch (data.type) {
    case 'createRoom': {
      await handleCreateRoom(ws, data);
      break;
    }

    case 'joinRoom': {
      await handleJoinRoom(ws, data);
      break;
    }

    case 'changeDirection': {
      await handleChangeDirection(ws, data);
      break;
    }

    case 'playAgain': {
      await handlePlayAgain(data);
      break;
    }

    case 'finalizeGame': {
      await handleFinalizeGame(data);
      break;
    }

    case 'subscribeRooms': {
      handleSubscribeRooms(ws);
      break;
    }

    case 'unsubscribeRooms': {
      handleUnsubscribeRooms(ws);
      break;
    }
  }
}

// Handle create room message
async function handleCreateRoom(ws: SnakeWebSocket, data: any): Promise<void> {
  console.log('[websocketService] Creating room with data:', data);
  
  // Check if player is already in a room
  if (ws.roomId) {
    console.log(`[websocketService] Player ${ws.playerId} already in room ${ws.roomId}, ignoring create room request`);
    ws.send(JSON.stringify({
      type: 'error',
      message: 'You are already in a room'
    }));
    return;
  }
  
  const roomId = generateRoomId();
  const { nickname, channelId, walletAddress } = data;
  const gridSize = { width: 40, height: 30 };

  // Create player
  const player = initializePlayer(ws.playerId, nickname, gridSize);

  // Create room with channel support
  const room: Room = {
    id: roomId,
    players: new Map([[player.id, player]]),
    food: generateFood(gridSize, new Map([[player.id, player]])),
    gameInterval: null,
    gridSize,
    channelIds: new Set(),
    playerAddresses: new Map([[player.id, walletAddress]]),
    currentState: null,
    stateVersion: 0,
    createdAt: Date.now()
  };

  // Add channelId if provided
  if (channelId) {
    room.channelIds.add(channelId);
    ws.channelId = channelId;
  }

  // Store the room
  addRoom(roomId, room);

  ws.roomId = roomId;

  // Respond with room info
  ws.send(JSON.stringify({
    type: 'roomCreated',
    roomId,
    playerId: player.id
  }));

  console.log(`[websocketService] Room created: ${roomId}, Player: ${player.id}, Address: ${walletAddress}`);
  
  // Notify subscribers about new room
  broadcastRoomUpdate(roomId);
}

// Handle join room message
async function handleJoinRoom(ws: SnakeWebSocket, data: any): Promise<void> {
  console.log('[websocketService] Joining room with data:', data);
  
  // Check if player is already in a room
  if (ws.roomId) {
    console.log(`[websocketService] Player ${ws.playerId} already in room ${ws.roomId}, ignoring join room request`);
    ws.send(JSON.stringify({
      type: 'error',
      message: 'You are already in a room'
    }));
    return;
  }
  
  const { roomId, nickname, channelId, walletAddress } = data;
  const room = getRoom(roomId);
  console.log('[websocketService] Room lookup result:', room ? 'found' : 'not found');

  if (!room) {
    ws.send(JSON.stringify({
      type: 'error',
      message: 'Room not found'
    }));
    return;
  }

  if (room.players.size >= 2) {
    ws.send(JSON.stringify({
      type: 'error',
      message: 'Room is full'
    }));
    return;
  }

  // Create player
  const player = initializePlayer(ws.playerId, nickname, room.gridSize);

  // Add player to room
  room.players.set(player.id, player);
  room.playerAddresses.set(player.id, walletAddress);

  ws.roomId = roomId;

  // Add channelId if provided
  if (channelId) {
    room.channelIds.add(channelId);
    ws.channelId = channelId;
  }

  // Respond with room info
  ws.send(JSON.stringify({
    type: 'roomJoined',
    roomId,
    playerId: player.id
  }));

  console.log(`Player joined room: ${roomId}, Player: ${player.id}, Address: ${walletAddress}`);
  console.log('Room data:', room);
  
  // Notify subscribers about room update
  broadcastRoomUpdate(roomId);

  // If we have 2 players and a channel, create the app session
  if (room.players.size === 2 && room.channelIds.size > 0) {
    try {
      const players = Array.from(room.players.values());

      // Create the app session
      console.log(`[websocketService] Creating app session for room ${roomId} with players:`, {
        player1: { id: players[0].id, address: room.playerAddresses.get(players[0].id) },
        player2: { id: players[1].id, address: room.playerAddresses.get(players[1].id) }
      });

      const createdAppId = await createAppSession(
        room.playerAddresses.get(players[0].id) as Hex,
        room.playerAddresses.get(players[1].id) as Hex
      );

      if (!createdAppId) {
        throw new Error("Failed to create app session - no app ID returned");
      }

      room.appId = createdAppId as Hex;
      console.log(`[websocketService] Created app session ${createdAppId} for room ${roomId}`);

      // Start the game
      room.gameInterval = setInterval(async () => {
        await gameTick(roomId);
      }, 150);

      // Initial game state broadcast
      const gameState = {
        type: 'gameState',
        players: Array.from(room.players.values()).map(p => ({
          id: p.id,
          nickname: p.nickname,
          segments: p.segments,
          score: p.score,
          isDead: p.isDead || false
        })),
        food: room.food,
        gridSize: room.gridSize,
        isGameOver: room.isGameOver || false,
        stateVersion: ++room.stateVersion,
        timestamp: Date.now()
      };

      // Store the current state in the room
      room.currentState = gameState;

      // Broadcast to all players in the room
      broadcastGameState(roomId, gameState);
    } catch (error: unknown) {
      console.error(`[websocketService] Error creating app session for room ${roomId}:`, error);

      // Clean up any partial state
      if (room.appId) {
        try {
          console.log(`[websocketService] Cleaning up failed app session ${room.appId}`);
          const players = Array.from(room.players.values());
          await closeAppSession(
            room.appId,
            room.playerAddresses.get(players[0].id) as Hex,
            room.playerAddresses.get(players[1].id) as Hex);
        } catch (closeError) {
          console.error(`[websocketService] Error cleaning up failed app session:`, closeError);
        }
        room.appId = undefined; // Clear the app ID after successful closure
      }

      ws.send(JSON.stringify({
        type: 'error',
        message: 'Failed to create app session: ' + (error instanceof Error ? error.message : 'Unknown error')
      }));
    }
  }
}

// Handle change direction message
async function handleChangeDirection(ws: SnakeWebSocket, data: any): Promise<void> {
  const roomId = ws.roomId;
  const { direction } = data;

  if (!roomId) return;

  const room = getRoom(roomId);
  if (!room) return;

  // Don't process direction changes if game is over
  if (room.isGameOver) return;

  const player = room.players.get(ws.playerId);
  if (!player) return;

  // Don't allow dead players to change direction
  if (player.isDead) return;

  // Prevent 180 degree turns
  if (player.direction === 'up' && direction === 'down') return;
  if (player.direction === 'down' && direction === 'up') return;
  if (player.direction === 'left' && direction === 'right') return;
  if (player.direction === 'right' && direction === 'left') return;

  player.direction = direction;
}

// Handle play again message
async function handlePlayAgain(data: any): Promise<void> {
  const { roomId } = data;
  if (!roomId) return;

  const room = getRoom(roomId);
  if (!room) return;

  // Reset game state
  room.isGameOver = false;

  // Reset players
  for (const player of room.players.values()) {
    const { width, height } = room.gridSize;
    const x = Math.floor(Math.random() * (width - 10)) + 5;
    const y = Math.floor(Math.random() * (height - 10)) + 5;

    player.position = { x, y };
    player.direction = ['up', 'down', 'left', 'right'][Math.floor(Math.random() * 4)] as 'up' | 'down' | 'left' | 'right';
    player.segments = [{ x, y }];
    player.score = 0;
    player.isDead = false;
  }

  // Create new food
  room.food = generateFood(room.gridSize, room.players);

  // Reset state version
  room.stateVersion = 0;

  // Restart game interval if needed
  if (!room.gameInterval) {
    room.gameInterval = setInterval(async () => {
      await gameTick(roomId);
    }, 150);
  }

  // Create and broadcast initial game state
  const gameState = {
    type: 'gameState',
    players: Array.from(room.players.values()).map(p => ({
      id: p.id,
      nickname: p.nickname,
      segments: p.segments,
      score: p.score,
      isDead: p.isDead || false
    })),
    food: room.food,
    gridSize: room.gridSize,
    isGameOver: room.isGameOver || false,
    stateVersion: ++room.stateVersion,
    timestamp: Date.now()
  };

  // Store the current state in the room
  room.currentState = gameState;

  // Broadcast to all players in the room
  broadcastGameState(roomId, gameState);
}

// Handle finalize game message
async function handleFinalizeGame(data: any): Promise<void> {
  const { roomId } = data;
  if (!roomId) return;

  const room = getRoom(roomId);
  if (!room || !room.appId) {
    console.log(`[websocketService] Cannot finalize game - room ${roomId} not found or no app session`);
    return;
  }

  console.log(`[websocketService] Finalizing game for room ${roomId} with app session ${room.appId}`);

  // For manual finalization - end the game immediately
  if (room.gameInterval) {
    clearInterval(room.gameInterval);
    room.gameInterval = null;
  }

  room.isGameOver = true;

  // Create final state with game results
  const finalState = {
    roomId,
    stateVersion: room.stateVersion,
    players: Array.from(room.players.values()).map(p => ({
      id: p.id,
      nickname: p.nickname,
      score: p.score,
      isDead: p.isDead || false
    })),
    isGameOver: true,
    finalizedAt: Date.now(),
    reason: 'game_ended'
  };

  // Finalize all channels associated with this room
  if (room.channelIds.size > 0) {
    const finalizePromises = Array.from(room.channelIds).map(id =>
      clearNetRPC.finalizeChannel(id, finalState)
    );

    try {
      await Promise.all(finalizePromises);
      console.log(`[websocketService] Finalized all channels for room ${roomId}`);
    } catch (error) {
      console.error(`[websocketService] Error finalizing channels for room ${roomId}:`, error);
    }
  }

  // Close the app session if not already being closed
  if (room.appId && !room.isClosingAppSession) {
    try {
      room.isClosingAppSession = true;
      console.log(`[websocketService] Closing app session ${room.appId} for room ${roomId}`);
      const players = Array.from(room.players.values());
      await closeAppSession(
        room.appId,
        room.playerAddresses.get(players[0].id) as Hex,
        room.playerAddresses.get(players[1].id) as Hex);
      console.log(`[websocketService] App session ${room.appId} closed successfully`);
      room.appId = undefined; // Clear the app ID after successful closure
    } catch (error) {
      console.error(`[websocketService] Error closing app session ${room.appId}:`, error);
      // Don't clear the app ID on error - it might still be valid
    } finally {
      room.isClosingAppSession = false;
    }
  }

  // Create and broadcast final game state
  const gameState = {
    type: 'gameState',
    players: Array.from(room.players.values()).map(p => ({
      id: p.id,
      nickname: p.nickname,
      segments: p.segments,
      score: p.score,
      isDead: p.isDead || false
    })),
    food: room.food,
    gridSize: room.gridSize,
    isGameOver: true,
    stateVersion: ++room.stateVersion,
    timestamp: Date.now()
  };

  // Store the current state in the room
  room.currentState = gameState;

  // Broadcast to all players in the room
  broadcastGameState(roomId, gameState);

  // Clean up the room after a short delay to allow clients to receive the final state
  setTimeout(() => {
    // Clear roomId from all connected clients in this room before deleting
    webSocketServer.clients.forEach(client => {
      const snakeClient = client as SnakeWebSocket;
      if (snakeClient.roomId === roomId) {
        console.log(`[websocketService] Clearing roomId for player ${snakeClient.playerId}`);
        snakeClient.roomId = undefined;
      }
    });
    
    removeRoom(roomId);
    console.log(`[websocketService] Room deleted: ${roomId}`);
    
    // Notify subscribers about room removal
    broadcastRoomRemoved(roomId);
  }, 2000);
}

// Handle client disconnect
async function handleDisconnect(ws: SnakeWebSocket): Promise<void> {
  console.log(`[websocketService] Client disconnected: ${ws.playerId}`);

  // Find the room this player was in
  const roomId = ws.roomId;
  if (!roomId) {
    console.log(`[websocketService] No room found for disconnected player ${ws.playerId}`);
    return;
  }

  const room = getRoom(roomId);
  if (!room) {
    console.log(`[websocketService] Room ${roomId} not found for disconnected player ${ws.playerId}`);
    return;
  }

  // Remove the player from the room
  room.players.delete(ws.playerId);
  console.log(`[websocketService] Removed player ${ws.playerId} from room ${roomId}`);
  
  // If room still has players, notify subscribers about room update
  if (room.players.size > 0) {
    broadcastRoomUpdate(roomId);
  }

  // If room is empty and this wasn't an intentional disconnect, clean up
  if (room.players.size === 0 && ws.readyState === WebSocket.CLOSED) {
    console.log(`[websocketService] Room ${roomId} is empty, cleaning up`);

    // Stop the game interval if it's running
    if (room.gameInterval) {
      clearInterval(room.gameInterval);
      room.gameInterval = null;
    }

    // Mark game as over
    room.isGameOver = true;

    // Close the app session if not already being closed
    if (room.appId && !room.isClosingAppSession) {
      try {
        room.isClosingAppSession = true;
        console.log(`[websocketService] Closing app session ${room.appId} for room ${roomId}`);
        const players = Array.from(room.players.values());
        await closeAppSession(
          room.appId,
          room.playerAddresses.get(players[0].id) as Hex,
          room.playerAddresses.get(players[1].id) as Hex);
        console.log(`[websocketService] App session ${room.appId} closed successfully`);
        room.appId = undefined; // Clear the app ID after successful closure
      } catch (error) {
        console.error(`[websocketService] Error closing app session ${room.appId}:`, error);
        // Don't clear the app ID on error - it might still be valid
      } finally {
        room.isClosingAppSession = false;
      }
    }

    // Finalize all channels associated with this room
    if (room.channelIds.size > 0) {
      const finalState = {
        roomId,
        players: Array.from(room.players.values()).map(p => ({
          id: p.id,
          nickname: p.nickname,
          score: p.score,
          isDead: p.isDead || false
        })),
        isGameOver: true,
        finalizedAt: Date.now(),
        reason: 'player_disconnected'
      };

      const finalizePromises = Array.from(room.channelIds).map(id =>
        clearNetRPC.finalizeChannel(id, finalState)
      );

      try {
        await Promise.all(finalizePromises);
        console.log(`[websocketService] Finalized all channels for room ${roomId}`);
      } catch (error) {
        console.error(`[websocketService] Error finalizing channels for room ${roomId}:`, error);
      }
    }

    // Remove the room
    removeRoom(roomId);
    console.log(`[websocketService] Room ${roomId} removed`);
    
    // Notify subscribers about room removal
    broadcastRoomRemoved(roomId);
  }
}

// Handle room subscription
function handleSubscribeRooms(ws: SnakeWebSocket): void {
  console.log(`[websocketService] Client ${ws.playerId} subscribing to room updates`);
  roomSubscribers.add(ws);
  
  // Send current rooms list
  const availableRooms = getAvailableRoomsList();
  ws.send(JSON.stringify({
    type: 'roomsList',
    rooms: availableRooms
  }));
}

// Handle room unsubscription
function handleUnsubscribeRooms(ws: SnakeWebSocket): void {
  console.log(`[websocketService] Client ${ws.playerId} unsubscribing from room updates`);
  roomSubscribers.delete(ws);
}

// Get list of available rooms (not full and not in active game)
function getAvailableRoomsList(): Array<any> {
  const allRooms = getAllRooms();
  const availableRooms: Array<any> = [];
  
  allRooms.forEach((room, roomId) => {
    // Include rooms that are not full (less than 2 players) and not in active game
    if (room.players.size < 2 && !room.isGameOver) {
      availableRooms.push({
        id: roomId,
        name: `Room ${roomId.slice(0, 8)}`,
        players: Array.from(room.players.values()).map(p => ({
          id: p.id,
          nickname: p.nickname
        })),
        maxPlayers: 2,
        isGameActive: !!room.gameInterval && !room.isGameOver,
        createdAt: room.createdAt
      });
    }
  });
  
  return availableRooms;
}

// Broadcast room updates to all subscribers
export function broadcastRoomUpdate(roomId: string): void {
  const room = getRoom(roomId);
  if (!room) return;
  
  const roomData = {
    id: roomId,
    name: `Room ${roomId.slice(0, 8)}`,
    players: Array.from(room.players.values()).map(p => ({
      id: p.id,
      nickname: p.nickname
    })),
    maxPlayers: 2,
    isGameActive: !!room.gameInterval && !room.isGameOver,
    createdAt: room.createdAt
  };
  
  const updateMessage = JSON.stringify({
    type: 'roomUpdated',
    room: roomData
  });
  
  // Send to all subscribers
  roomSubscribers.forEach(client => {
    if (client.readyState === WebSocket.OPEN) {
      client.send(updateMessage);
    }
  });
}

// Broadcast room removal to all subscribers
export function broadcastRoomRemoved(roomId: string): void {
  const removeMessage = JSON.stringify({
    type: 'roomRemoved',
    roomId
  });
  
  // Send to all subscribers
  roomSubscribers.forEach(client => {
    if (client.readyState === WebSocket.OPEN) {
      client.send(removeMessage);
    }
  });
}
