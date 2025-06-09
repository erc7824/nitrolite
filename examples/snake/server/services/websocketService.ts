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
  initializeBroadcastFunction
} from './gameService.ts';
import { closeAppSession, sendToBroker } from './brokerService.ts';
import {
  generateAppSessionMessage,
  addAppSessionSignature,
  serverSigner,
  serverWallet
} from './appSessionService.ts';
import { Hex } from 'viem';
import { CloseAppSessionRequest, createCloseAppSessionMessage } from '@erc7824/nitrolite';

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

// Broadcast vote update to all clients in a room
function broadcastVoteUpdate(roomId: string): void {
  const room = getRoom(roomId);
  if (!room) return;

  const voteUpdate = {
    type: 'playAgainVoteUpdate',
    playAgainVotes: room.playAgainVotes ? Array.from(room.playAgainVotes) : [],
    totalPlayers: room.players.size,
    votesNeeded: room.players.size - (room.playAgainVotes?.size || 0)
  };

  console.log(`[broadcastVoteUpdate] Broadcasting vote update to room ${roomId}:`, voteUpdate);

  webSocketServer.clients.forEach(client => {
    const snakeClient = client as SnakeWebSocket;
    if (snakeClient.roomId === roomId && snakeClient.readyState === WebSocket.OPEN) {
      snakeClient.send(JSON.stringify(voteUpdate));
    }
  });
}

// Broadcast player disconnect notification to all clients in a room
function broadcastPlayerDisconnect(roomId: string, playerNickname: string): void {
  const disconnectNotification = {
    type: 'playerDisconnected',
    playerNickname,
    message: `${playerNickname} left the room`
  };

  console.log(`[broadcastPlayerDisconnect] Broadcasting disconnect notification to room ${roomId}:`, disconnectNotification);

  webSocketServer.clients.forEach(client => {
    const snakeClient = client as SnakeWebSocket;
    if (snakeClient.roomId === roomId && snakeClient.readyState === WebSocket.OPEN) {
      snakeClient.send(JSON.stringify(disconnectNotification));
    }
  });
}

// Handle WebSocket message
async function handleWebSocketMessage(ws: SnakeWebSocket, data: any): Promise<void> {
  console.log(`[websocketService] Received message type: ${data.type}`, data);

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
      console.log('[websocketService] Handling finalizeGame message:', data);
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

    case 'appSession:signature': {
      await handleAppSessionSignature(ws, data);
      break;
    }

    case 'closeSessionSignature': {
      await handleCloseSessionSignature(data);
      break;
    }

    default: {
      console.log(`[websocketService] Unknown message type: ${data.type}`);
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
    createdAt: Date.now(),
    closeSessionSignatures: new Map()
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

  // If we have 2 players and a channel, initiate app session creation with signatures
  if (room.players.size === 2 && room.channelIds.size > 0) {
    try {
      const players = Array.from(room.players.values());
      const participantA = room.playerAddresses.get(players[0].id) as Hex;
      const participantB = room.playerAddresses.get(players[1].id) as Hex;

      console.log(`[websocketService] Initiating app session creation for room ${roomId} with players:`, {
        player1: { id: players[0].id, address: participantA },
        player2: { id: players[1].id, address: participantB }
      });

      // Generate the message structure for signing
      const { requestToSign } = await generateAppSessionMessage(roomId, participantA, participantB);

      // Start signature collection with guest player (participant B)
      const guestPlayerId = players[1].id; // Second player to join
      const guestClient = Array.from(webSocketServer.clients).find(client => {
        const snakeClient = client as SnakeWebSocket;
        return snakeClient.playerId === guestPlayerId && snakeClient.roomId === roomId;
      }) as SnakeWebSocket;

      if (guestClient && guestClient.readyState === WebSocket.OPEN) {
        guestClient.send(JSON.stringify({
          type: 'appSession:signatureRequest',
          roomId,
          requestToSign,
          participantAddress: participantB
        }));
        console.log(`[websocketService] Sent signature request to guest player ${guestPlayerId}`);
      } else {
        throw new Error("Guest player not found or not connected");
      }

    } catch (error: unknown) {
      console.error(`[websocketService] Error initiating app session for room ${roomId}:`, error);
      ws.send(JSON.stringify({
        type: 'error',
        message: 'Failed to initiate app session: ' + (error instanceof Error ? error.message : 'Unknown error')
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
  const { roomId, playerId } = data;
  if (!roomId || !playerId) return;

  const room = getRoom(roomId);
  if (!room) return;

  // Initialize votes set if it doesn't exist
  if (!room.playAgainVotes) {
    room.playAgainVotes = new Set<string>();
  }

  // Add this player's vote
  room.playAgainVotes.add(playerId);

  console.log(`[handlePlayAgain] Player ${playerId} voted to play again. Current votes: ${room.playAgainVotes.size}/${room.players.size}`);
  console.log(`[handlePlayAgain] Current players in room:`, Array.from(room.players.keys()));
  console.log(`[handlePlayAgain] Current voters:`, Array.from(room.playAgainVotes));

  // Broadcast vote update to all players in the room
  broadcastVoteUpdate(roomId);

  // Check if all players have voted to play again (and we have at least 2 players)
  if (room.playAgainVotes.size === room.players.size && room.players.size >= 2) {
    console.log(`[handlePlayAgain] All players voted to play again. Restarting game.`);

    // Clear votes for next time
    room.playAgainVotes.clear();

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
}

// Handle finalize game message
async function handleFinalizeGame(data: any): Promise<void> {
  const { roomId, playerId, walletAddress } = data;
  if (!roomId || !playerId || !walletAddress) return;

  const room = getRoom(roomId);
  if (!room || !room.appId) {
    console.log(`[websocketService] Cannot finalize game - room ${roomId} not found or no app session`);
    return;
  }

  console.log(`[websocketService] Finalizing game for room ${roomId} with app session ${room.appId}`);

  // Mark game as over
  room.isGameOver = true;

  // Get player addresses
  const playerAddresses = Array.from(room.playerAddresses.values());
  const server = serverWallet.address as Hex;

  try {
    // Create close message and sign with server
    const params: CloseAppSessionRequest[] = [{
      app_session_id: room.appId,
      allocations: [
        {
          participant: playerAddresses[0] as Hex,
          asset: "usdc",
          amount: "0.00001"
        },
        {
          participant: playerAddresses[1] as Hex,
          asset: "usdc",
          amount: "0.00001"
        },
        {
          participant: server,
          asset: "usdc",
          amount: "0"
        }
      ]
    }];

    // Create the message to be signed by all participants
    const closeRequestData = await createCloseAppSessionMessage(serverSigner, params);

    // Store the request in the room for later use
    const requestToSign = JSON.parse(closeRequestData);
    console.log('[websocketService] Close session request to sign:', requestToSign);
    room.closeSessionRequest = requestToSign;
    room.closeSessionSignatures = new Map();

    // Send signature request to all players
    console.log('[websocketService] Sending close request to all players');
    webSocketServer.clients.forEach(client => {
      const snakeClient = client as SnakeWebSocket;
      const closeRequest = {
        type: 'appSession:closeRequest',
        roomId,
        requestToSign,
        participantAddress: room.playerAddresses.get(playerId)
      };
      if (snakeClient.roomId === roomId && snakeClient.readyState === WebSocket.OPEN) {
        snakeClient.send(JSON.stringify(closeRequest));
      }
    });

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
  } catch (error) {
    console.error(`[websocketService] Error preparing close session request for room ${roomId}:`, error);
  }
}

// Handle close session signature
async function handleCloseSessionSignature(data: any): Promise<void> {
  const { roomId, playerId, signature, participantAddress } = data;
  if (!roomId || !playerId || !signature || !participantAddress) {
    console.log(`[websocketService] Missing required fields for close session signature`);
    return;
  }

  const room = getRoom(roomId);
  if (!room || !room.appId || !room.closeSessionRequest) {
    console.log(`[websocketService] Cannot process signature - room ${roomId} not found or no close session request`);
    return;
  }

  // Store the signature
  room.closeSessionSignatures.set(playerId, signature);
  console.log(`[websocketService] Received signature from player ${playerId} for room ${roomId}`);

  // Check if we have all required signatures
  if (room.closeSessionSignatures.size === room.players.size) {
    try {
      room.isClosingAppSession = true;
      console.log(`[websocketService] All player signatures received, closing app session ${room.appId}`);

      // Get server signature
      const serverSignature = await serverSigner(room.closeSessionRequest);
      console.log(`[websocketService] Server signature created for room ${roomId}`);

      // Combine all signatures
      const allSignatures = [
        ...Array.from(room.closeSessionSignatures.values()),
        serverSignature
      ];

      // Create final message with all signatures
      const finalMessage = {
        req: room.closeSessionRequest.req,
        sig: allSignatures
      };

      console.log(`[websocketService] Sending final close request to broker for room ${roomId}`);
      await sendToBroker(finalMessage);
      console.log(`[websocketService] App session ${room.appId} closed successfully`);

      // Notify all players that the session is closed
      webSocketServer.clients.forEach(client => {
        const snakeClient = client as SnakeWebSocket;
        if (snakeClient.roomId === roomId && snakeClient.readyState === WebSocket.OPEN) {
          snakeClient.send(JSON.stringify({
            type: 'appSession:closed'
          }));
        }
      });

      // Clean up room state
      room.appId = undefined;
      room.closeSessionRequest = undefined;
      room.closeSessionSignatures.clear();
    } catch (error) {
      console.error(`[websocketService] Error closing app session ${room.appId}:`, error);
    } finally {
      room.isClosingAppSession = false;
    }
  } else {
    console.log(`[websocketService] Waiting for more signatures for room ${roomId} (${room.closeSessionSignatures.size}/${room.players.size})`);
  }
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

  // Get player info before removing them
  const disconnectedPlayer = room.players.get(ws.playerId);
  const playerNickname = disconnectedPlayer?.nickname || 'Unknown player';

  // Remove the player from the room
  room.players.delete(ws.playerId);
  console.log(`[websocketService] Removed player ${ws.playerId} from room ${roomId}`);

  // Remove player's vote if they had one
  if (room.playAgainVotes) {
    room.playAgainVotes.delete(ws.playerId);
  }

  // Notify remaining players about the disconnect
  if (room.players.size > 0) {
    broadcastPlayerDisconnect(roomId, playerNickname);
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

// Handle app session signature submission from guest player
async function handleAppSessionSignature(ws: SnakeWebSocket, data: any): Promise<void> {
  const { roomId, signature, participantAddress } = data;

  if (!roomId || !signature || !participantAddress) {
    ws.send(JSON.stringify({
      type: 'error',
      message: 'Missing required fields for app session signature'
    }));
    return;
  }

  const room = getRoom(roomId);
  if (!room) {
    ws.send(JSON.stringify({
      type: 'error',
      message: 'Room not found'
    }));
    return;
  }

  try {
    // Add the signature
    const allSigned = addAppSessionSignature(roomId, participantAddress as Hex, signature);
    if (!allSigned) {
      // Send confirmation to the player who just signed
      ws.send(JSON.stringify({
        type: 'appSession:signatureConfirmed',
        roomId
      }));
      console.log(`[handleAppSessionSignature] Player ${ws.playerId} submitted signature for room ${roomId}, waiting for other player`);
      return;
    }

    console.log(`[handleAppSessionSignature] All signatures collected for room ${roomId}, creating app session`);

    // Create the app session with all signatures
    // const appId = await createAppSessionWithSignatures(roomId);
    const appId = '0x66ffb53d1b788f396dca76d79ef95a6203aba487d4d2bec1a3a7cf3c7a0a40cb';
    room.appId = appId as Hex;

    console.log(`[handleAppSessionSignature] Created app session ${appId} for room ${roomId}`);

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

    // Notify all players that the game has started
    const players = Array.from(room.players.values());
    for (const player of players) {
      const client = Array.from(webSocketServer.clients).find(client => {
        const snakeClient = client as SnakeWebSocket;
        return snakeClient.playerId === player.id && snakeClient.roomId === roomId;
      }) as SnakeWebSocket;

      if (client && client.readyState === WebSocket.OPEN) {
        client.send(JSON.stringify({
          type: 'appSession:gameStarted',
          roomId,
          appId
        }));
      }
    }

  } catch (error) {
    console.error(`[handleAppSessionSignature] Error handling signature for room ${roomId}:`, error);
    ws.send(JSON.stringify({
      type: 'error',
      message: 'Failed to process signature: ' + (error instanceof Error ? error.message : 'Unknown error')
    }));
  }
}
