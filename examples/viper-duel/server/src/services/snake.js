/**
 * Snake game engine
 */
import { ethers } from 'ethers';

const GRID_WIDTH = 20;
const GRID_HEIGHT = 20;
const INITIAL_SNAKE_LENGTH = 3;
const GAME_SPEED = 200; // ms between moves

/**
 * @typedef {Object} Position
 * @property {number} x - X coordinate
 * @property {number} y - Y coordinate
 */

/**
 * @typedef {Object} Snake
 * @property {Array<Position>} body - Array of positions representing snake segments
 * @property {string} direction - Current direction ('UP', 'DOWN', 'LEFT', 'RIGHT')
 * @property {boolean} alive - Whether the snake is alive
 * @property {number} score - Player's score
 */

/**
 * @typedef {Object} GameState
 * @property {Object} snakes - Object with snake data for each player
 * @property {Snake} snakes.player1 - Player 1's snake
 * @property {Snake} snakes.player2 - Player 2's snake
 * @property {Array<Position>} food - Array of food positions
 * @property {string|null} winner - The winner of the game ('player1', 'player2', or null if no winner yet)
 * @property {boolean} isGameOver - Whether the game is over
 * @property {Object} players - Object with player information
 * @property {string} players.player1 - EOA address of player 1
 * @property {string} players.player2 - EOA address of player 2
 * @property {number} gameTime - Game time in seconds
 */

/**
 * Creates initial snake at starting position
 * @param {number} startX - Starting X position
 * @param {number} startY - Starting Y position
 * @param {string} direction - Initial direction
 * @returns {Snake} Initial snake
 */
function createInitialSnake(startX, startY, direction) {
  const body = [];
  
  // Create snake segments based on direction
  for (let i = 0; i < INITIAL_SNAKE_LENGTH; i++) {
    switch (direction) {
      case 'RIGHT':
        body.push({ x: startX - i, y: startY });
        break;
      case 'LEFT':
        body.push({ x: startX + i, y: startY });
        break;
      case 'DOWN':
        body.push({ x: startX, y: startY - i });
        break;
      case 'UP':
        body.push({ x: startX, y: startY + i });
        break;
    }
  }
  
  return {
    body,
    direction,
    alive: true,
    score: 0
  };
}

/**
 * Creates a new game state
 * @param {string} hostEoa - Host's Ethereum address (player 1)
 * @param {string} guestEoa - Guest's Ethereum address (player 2)
 * @returns {GameState} Initial game state
 */
export function createGame(hostEoa, guestEoa) {
  // Format addresses to proper checksum format
  const formattedHostEoa = ethers.getAddress(hostEoa);
  const formattedGuestEoa = ethers.getAddress(guestEoa);
  
  return {
    snakes: {
      player1: createInitialSnake(3, 3, 'RIGHT'),
      player2: createInitialSnake(GRID_WIDTH - 4, GRID_HEIGHT - 4, 'LEFT')
    },
    food: [spawnFood(), spawnFood(), spawnFood()], // Start with 3 food items
    winner: null,
    isGameOver: false,
    players: {
      player1: formattedHostEoa,
      player2: formattedGuestEoa
    },
    gameTime: 0
  };
}

/**
 * Spawns food at a random empty position
 * @param {GameState} gameState - Current game state (optional)
 * @returns {Position} Food position
 */
function spawnFood(gameState = null) {
  let position;
  let attempts = 0;
  const maxAttempts = 100;
  
  do {
    position = {
      x: Math.floor(Math.random() * GRID_WIDTH),
      y: Math.floor(Math.random() * GRID_HEIGHT)
    };
    attempts++;
  } while (gameState && isPositionOccupied(position, gameState) && attempts < maxAttempts);
  
  return position;
}

/**
 * Checks if a position is occupied by a snake
 * @param {Position} position - Position to check
 * @param {GameState} gameState - Current game state
 * @returns {boolean} Whether position is occupied
 */
function isPositionOccupied(position, gameState) {
  for (const snake of Object.values(gameState.snakes)) {
    for (const segment of snake.body) {
      if (segment.x === position.x && segment.y === position.y) {
        return true;
      }
    }
  }
  return false;
}

/**
 * Changes snake direction
 * @param {GameState} gameState - Current game state
 * @param {string} direction - New direction ('UP', 'DOWN', 'LEFT', 'RIGHT')
 * @param {string} playerEoa - Player's Ethereum address
 * @returns {Object} Result with updated game state or error
 */
export function changeDirection(gameState, direction, playerEoa) {
  // Format player address to proper checksum format
  const formattedPlayerEoa = ethers.getAddress(playerEoa);
  
  // Check if the game is already over
  if (gameState.isGameOver) {
    return { success: false, error: 'Game is already over' };
  }

  // Determine which player is making the move
  const playerId = gameState.players.player1 === formattedPlayerEoa ? 'player1' : 'player2';
  if (!gameState.players[playerId]) {
    return { success: false, error: 'Player not in this game' };
  }

  const snake = gameState.snakes[playerId];
  if (!snake.alive) {
    return { success: false, error: 'Snake is dead' };
  }

  // Update direction only - automatic timer will handle movement
  const updatedGameState = {
    ...gameState,
    snakes: {
      ...gameState.snakes,
      [playerId]: {
        ...snake,
        direction
      }
    }
  };

  return { 
    success: true, 
    gameState: updatedGameState
  };
}

/**
 * Moves a single snake forward and handles collisions
 * @param {GameState} gameState - Current game state
 * @param {string} playerId - The player whose snake to move
 * @returns {Object} Result with updated game state
 */
function moveSnake(gameState, playerId) {
  if (gameState.isGameOver) {
    return { success: true, gameState };
  }

  const snake = gameState.snakes[playerId];
  if (!snake.alive) {
    return { success: true, gameState };
  }

  const head = snake.body[0];
  let newHead;
  
  // Calculate new head position
  switch (snake.direction) {
    case 'UP':
      newHead = { x: head.x, y: head.y - 1 };
      break;
    case 'DOWN':
      newHead = { x: head.x, y: head.y + 1 };
      break;
    case 'LEFT':
      newHead = { x: head.x - 1, y: head.y };
      break;
    case 'RIGHT':
      newHead = { x: head.x + 1, y: head.y };
      break;
  }
  
  // Handle screen wraparound (no wall collision)
  if (newHead.x < 0) {
    newHead.x = GRID_WIDTH - 1; // Wrap to right side
  } else if (newHead.x >= GRID_WIDTH) {
    newHead.x = 0; // Wrap to left side
  }
  
  if (newHead.y < 0) {
    newHead.y = GRID_HEIGHT - 1; // Wrap to bottom
  } else if (newHead.y >= GRID_HEIGHT) {
    newHead.y = 0; // Wrap to top
  }
  
  // Wraparound completed
  
  // Check self collision
  const selfCollision = snake.body.some(segment => 
    segment.x === newHead.x && segment.y === newHead.y
  );
  if (selfCollision) {
    console.log(`🔄 ${playerId} collided with self at (${newHead.x}, ${newHead.y})`);
    const updatedGameState = {
      ...gameState,
      snakes: {
        ...gameState.snakes,
        [playerId]: { ...snake, alive: false }
      }
    };
    return { success: true, gameState: updatedGameState };
  }
  
  // Check collision with other snake
  const otherPlayerId = playerId === 'player1' ? 'player2' : 'player1';
  const otherSnake = gameState.snakes[otherPlayerId];
  const otherSnakeCollision = otherSnake.body.some(segment => 
    segment.x === newHead.x && segment.y === newHead.y
  );
  if (otherSnakeCollision) {
    console.log(`🐍 ${playerId} collided with ${otherPlayerId} at (${newHead.x}, ${newHead.y})`);
    const updatedGameState = {
      ...gameState,
      snakes: {
        ...gameState.snakes,
        [playerId]: { ...snake, alive: false }
      }
    };
    return { success: true, gameState: updatedGameState };
  }
  
  // Check food collision
  let updatedFood = [...gameState.food];
  const foodIndex = updatedFood.findIndex(food => 
    food.x === newHead.x && food.y === newHead.y
  );
  
  let newBody;
  let newScore = snake.score;
  if (foodIndex !== -1) {
    // Ate food - grow snake
    newBody = [newHead, ...snake.body];
    updatedFood.splice(foodIndex, 1);
    newScore = snake.score + 1;
    
    console.log(`🍎 ${playerId} ate food! New score: ${newScore}`);
    
    // Spawn new food to maintain at least 3 food items
    while (updatedFood.length < 3) {
      updatedFood.push(spawnFood({ ...gameState, food: updatedFood }));
    }
  } else {
    // Normal move - don't grow
    newBody = [newHead, ...snake.body.slice(0, -1)];
  }
  
  const updatedGameState = {
    ...gameState,
    snakes: {
      ...gameState.snakes,
      [playerId]: {
        ...snake,
        body: newBody,
        score: newScore
      }
    },
    food: updatedFood,
    gameTime: gameState.gameTime + 1
  };
  
  // Check for game over conditions
  const aliveSnakes = Object.values(updatedGameState.snakes).filter(s => s.alive);
  console.log(`🔍 Game state check: player1 alive=${updatedGameState.snakes.player1.alive}, player2 alive=${updatedGameState.snakes.player2.alive}, alive count=${aliveSnakes.length}`);
  
  if (aliveSnakes.length === 0) {
    console.log(`🏁 Game over - Both snakes died`);
    updatedGameState.isGameOver = true;
    updatedGameState.winner = null;
  } else if (aliveSnakes.length === 1) {
    const winningPlayerId = Object.keys(updatedGameState.snakes).find(
      id => updatedGameState.snakes[id].alive
    );
    console.log(`🏆 Game over - ${winningPlayerId} wins!`);
    updatedGameState.winner = winningPlayerId;
    updatedGameState.isGameOver = true;
  } else {
    console.log(`✅ Game continues - both snakes alive`);
  }
  
  return { success: true, gameState: updatedGameState };
}

/**
 * Moves all snakes forward and handles collisions
 * @param {GameState} gameState - Current game state
 * @returns {Object} Result with updated game state
 */
export function updateGame(gameState) {
  if (gameState.isGameOver) {
    return { success: true, gameState };
  }

  const updatedSnakes = { ...gameState.snakes };
  let updatedFood = [...gameState.food];
  
  // Move each alive snake
  for (const [playerId, snake] of Object.entries(updatedSnakes)) {
    if (!snake.alive) continue;
    
    const head = snake.body[0];
    let newHead;
    
    // Calculate new head position
    switch (snake.direction) {
      case 'UP':
        newHead = { x: head.x, y: head.y - 1 };
        break;
      case 'DOWN':
        newHead = { x: head.x, y: head.y + 1 };
        break;
      case 'LEFT':
        newHead = { x: head.x - 1, y: head.y };
        break;
      case 'RIGHT':
        newHead = { x: head.x + 1, y: head.y };
        break;
    }
    
    // Handle screen wraparound (no wall collision)
    if (newHead.x < 0) {
      newHead.x = GRID_WIDTH - 1; // Wrap to right side
    } else if (newHead.x >= GRID_WIDTH) {
      newHead.x = 0; // Wrap to left side
    }
    
    if (newHead.y < 0) {
      newHead.y = GRID_HEIGHT - 1; // Wrap to bottom
    } else if (newHead.y >= GRID_HEIGHT) {
      newHead.y = 0; // Wrap to top
    }
    
    // Wraparound completed
    
    // Check self collision
    const selfCollision = snake.body.some(segment => 
      segment.x === newHead.x && segment.y === newHead.y
    );
    if (selfCollision) {
      console.log(`🔄 ${playerId} collided with self at (${newHead.x}, ${newHead.y})`);
      updatedSnakes[playerId] = { ...snake, alive: false };
      continue;
    }
    
    // Check collision with other snake
    const otherPlayerId = playerId === 'player1' ? 'player2' : 'player1';
    const otherSnake = updatedSnakes[otherPlayerId];
    const otherSnakeCollision = otherSnake.body.some(segment => 
      segment.x === newHead.x && segment.y === newHead.y
    );
    if (otherSnakeCollision) {
      console.log(`🐍 ${playerId} collided with ${otherPlayerId} at (${newHead.x}, ${newHead.y})`);
      updatedSnakes[playerId] = { ...snake, alive: false };
      continue;
    }
    
    // Check food collision
    const foodIndex = updatedFood.findIndex(food => 
      food.x === newHead.x && food.y === newHead.y
    );
    
    let newBody;
    if (foodIndex !== -1) {
      // Ate food - grow snake
      newBody = [newHead, ...snake.body];
      updatedFood.splice(foodIndex, 1);
      updatedSnakes[playerId] = {
        ...snake,
        body: newBody,
        score: snake.score + 1
      };
      
      // Spawn new food to maintain at least 3 food items
      const newGameState = {
        ...gameState,
        snakes: updatedSnakes,
        food: updatedFood
      };
      
      // Always keep at least 3 food items on the board
      while (updatedFood.length < 3) {
        updatedFood.push(spawnFood(newGameState));
      }
    } else {
      // Normal move - don't grow
      newBody = [newHead, ...snake.body.slice(0, -1)];
      updatedSnakes[playerId] = {
        ...snake,
        body: newBody
      };
    }
  }
  
  // Check for game over conditions
  const aliveSnakes = Object.values(updatedSnakes).filter(snake => snake.alive);
  let winner = null;
  let isGameOver = false;
  
  if (aliveSnakes.length === 0) {
    // Both snakes died - tie
    isGameOver = true;
  } else if (aliveSnakes.length === 1) {
    // One snake alive - winner
    const winningPlayerId = Object.keys(updatedSnakes).find(
      playerId => updatedSnakes[playerId].alive
    );
    winner = winningPlayerId;
    isGameOver = true;
  }
  
  const updatedGameState = {
    ...gameState,
    snakes: updatedSnakes,
    food: updatedFood,
    winner,
    isGameOver,
    gameTime: gameState.gameTime + 1
  };
  
  return {
    success: true,
    gameState: updatedGameState
  };
}

/**
 * Formats game state for client consumption
 * @param {GameState} gameState - Current game state
 * @param {string} roomId - Room ID
 * @returns {Object} Formatted game state for client
 */
export function formatGameState(gameState, roomId, betAmount = 0) {
  return {
    roomId,
    snakes: gameState.snakes,
    food: gameState.food,
    players: gameState.players,
    gameTime: gameState.gameTime,
    betAmount: betAmount
  };
}

/**
 * Formats game over message
 * @param {GameState} gameState - Current game state
 * @returns {Object} Game over message
 */
export function formatGameOverMessage(gameState) {
  return {
    winner: gameState.winner,
    finalScores: {
      player1: gameState.snakes.player1.score,
      player2: gameState.snakes.player2.score
    },
    gameTime: gameState.gameTime
  };
}