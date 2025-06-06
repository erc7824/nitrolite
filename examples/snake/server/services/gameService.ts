import { randomBytes } from 'crypto';
import { Player } from '../interfaces/index.ts';
import { getRoom } from './stateService.ts';
import { closeAppSession } from './brokerService.ts';
import { Hex } from 'viem';

// Import websocketService to broadcast game state - imported at bottom of file to avoid circular dependency
let broadcastGameStateToClients: (roomId: string, gameState: any) => void;

// Generate a random room ID
export function generateRoomId(): string {
  return randomBytes(16).toString('hex');
}

// Generate a random food position
export function generateFood(
  gridSize: { width: number; height: number },
  players: Map<string, Player>
): { x: number; y: number } {
  let x: number = 0, y: number = 0;
  let validPosition = false;

  while (!validPosition) {
    x = Math.floor(Math.random() * gridSize.width);
    y = Math.floor(Math.random() * gridSize.height);

    validPosition = true;

    for (const player of players.values()) {
      for (const segment of player.segments) {
        if (segment.x === x && segment.y === y) {
          validPosition = false;
          break;
        }
      }
      if (!validPosition) break;
    }
  }

  return { x, y };
}

// Initialize a new player
export function initializePlayer(
  id: string,
  nickname: string,
  gridSize: { width: number; height: number }
): Player {
  const x = Math.floor(Math.random() * (gridSize.width - 10)) + 5;
  const y = Math.floor(Math.random() * (gridSize.height - 10)) + 5;

  return {
    id,
    nickname,
    position: { x, y },
    direction: ['up', 'down', 'left', 'right'][Math.floor(Math.random() * 4)] as 'up' | 'down' | 'left' | 'right',
    segments: [{ x, y }],
    score: 0
  };
}

// Initialize broadcast function
export function initializeBroadcastFunction(
  broadcastFn: (roomId: string, gameState: any) => void
): void {
  broadcastGameStateToClients = broadcastFn;
}

// Process one game tick for a room
export async function gameTick(roomId: string): Promise<void> {
  const room = getRoom(roomId);
  if (!room) return;

  // If game is over, don't process any more ticks
  if (room.isGameOver) {
    return;
  }

  const { players, food, gridSize } = room;
  const playersArray = Array.from(players.values());

  // Move each player
  for (const player of playersArray) {
    // Skip if player is dead
    if (player.isDead) continue;

    // Get current head position
    const head = { ...player.position };

    // Move head based on direction
    switch (player.direction) {
      case 'up':
        head.y = (head.y - 1 + gridSize.height) % gridSize.height;
        break;
      case 'down':
        head.y = (head.y + 1) % gridSize.height;
        break;
      case 'left':
        head.x = (head.x - 1 + gridSize.width) % gridSize.width;
        break;
      case 'right':
        head.x = (head.x + 1) % gridSize.width;
        break;
    }

    // Update player position
    player.position = head;

    // Check if player ate food
    if (head.x === food.x && head.y === food.y) {
      player.score += 10;
      // Add new segment (will be handled below - we don't remove the tail)
      // Generate new food
      room.food = generateFood(gridSize, players);
    } else {
      // Remove tail if didn't eat food
      player.segments.pop();
    }

    // Add new head to segments
    player.segments.unshift({ ...head });

    // Check for collisions with other players
    let isCollision = false;
    for (const otherPlayer of playersArray) {
      // Skip first segment of the current player
      const segments = otherPlayer === player ?
        otherPlayer.segments.slice(1) :
        otherPlayer.segments;

      for (const segment of segments) {
        if (head.x === segment.x && head.y === segment.y) {
          // Collision detected - mark player as dead
          console.log(`[gameTick] Collision detected for player ${player.id} at position (${head.x}, ${head.y})`);
          console.log(`[gameTick] Collided with segment from player ${otherPlayer.id} at (${segment.x}, ${segment.y})`);
          isCollision = true;
          player.isDead = true;
          break;
        }
      }
      if (isCollision) break;
    }
  }

  // Check if game is over (only one player left alive or all players dead)
  if (playersArray.length > 1) {
    const alivePlayers = playersArray.filter(p => !p.isDead);
    console.log(`[gameTick] Alive players count: ${alivePlayers.length}`);
    console.log(`[gameTick] Player states:`, playersArray.map(p => ({ id: p.id, isDead: p.isDead })));

    if (alivePlayers.length <= 1) {
      console.log(`[gameTick] Game over condition met at ${Date.now()}`);
      room.isGameOver = true;

      // If there's an interval, clear it to stop the game
      if (room.gameInterval) {
        console.log(`[gameTick] Clearing game interval`);
        clearInterval(room.gameInterval);
        room.gameInterval = null;
      }
    }
  }

  // Broadcast game state to all players in the room
  console.log(`[gameTick] Broadcasting game state at ${Date.now()}`);
  await broadcastGameState(roomId);
}

// Broadcast game state to all clients in a room
export async function broadcastGameState(roomId: string): Promise<void> {
  const room = getRoom(roomId);
  if (!room) {
    console.warn(`[broadcastGameState] Room ${roomId} not found`);
    return;
  }

  // Create game state
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

  // Broadcast to all players in the room if the broadcast function is initialized
  if (broadcastGameStateToClients) {
    console.log(`[broadcastGameState] Broadcasting to clients at ${Date.now()}`);
    broadcastGameStateToClients(roomId, gameState);
    console.log(`[broadcastGameState] Broadcast complete at ${Date.now()}`);
  } else {
    console.warn(`[broadcastGameState] No broadcast function available`);
  }
}

