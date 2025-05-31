/**
 * Tic Tac Toe game engine
 */
import { ethers } from 'ethers';

/**
 * @typedef {Object} GameState
 * @property {Array<string|null>} board - 9 elements representing the board (null or player symbol)
 * @property {string} nextTurn - The player whose turn is next ('X' or 'O')
 * @property {string|null} winner - The winner of the game ('X', 'O', or null if no winner yet)
 * @property {boolean} isGameOver - Whether the game is over
 * @property {Object} players - Object with player information
 * @property {string} players.X - EOA address of player X
 * @property {string} players.O - EOA address of player O
 */

/**
 * Creates a new game state
 * @param {string} hostEoa - Host's Ethereum address (X player)
 * @param {string} guestEoa - Guest's Ethereum address (O player)
 * @returns {GameState} Initial game state
 */
export function createGame(hostEoa, guestEoa) {
  // Format addresses to proper checksum format
  const formattedHostEoa = ethers.getAddress(hostEoa);
  const formattedGuestEoa = ethers.getAddress(guestEoa);
  
  return {
    board: Array(9).fill(null),
    nextTurn: 'X', // X goes first
    winner: null,
    isGameOver: false,
    players: {
      X: formattedHostEoa,
      O: formattedGuestEoa
    }
  };
}

/**
 * Makes a move on the board
 * @param {GameState} gameState - Current game state
 * @param {number} position - Position to place the piece (0-8)
 * @param {string} playerEoa - Player's Ethereum address
 * @returns {Object} Result with updated game state or error
 */
export function makeMove(gameState, position, playerEoa) {
  // Format player address to proper checksum format
  const formattedPlayerEoa = ethers.getAddress(playerEoa);
  
  // Check if the game is already over
  if (gameState.isGameOver) {
    return { success: false, error: 'Game is already over' };
  }

  // Check if it's the player's turn
  const playerSymbol = gameState.players.X === formattedPlayerEoa ? 'X' : 'O';
  if (playerSymbol !== gameState.nextTurn) {
    return { success: false, error: 'Not your turn' };
  }

  // Check if the position is valid and empty
  if (position < 0 || position > 8 || gameState.board[position] !== null) {
    return { success: false, error: 'Invalid move: position is occupied or out of bounds' };
  }

  // Create a new board with the move applied
  const newBoard = [...gameState.board];
  newBoard[position] = playerSymbol;

  // Check for win or draw
  const winner = checkWinner(newBoard);
  const isDraw = !winner && newBoard.every(cell => cell !== null);

  // Create updated game state
  const updatedGameState = {
    ...gameState,
    board: newBoard,
    nextTurn: playerSymbol === 'X' ? 'O' : 'X',
    winner: winner,
    isGameOver: !!winner || isDraw
  };

  return { 
    success: true, 
    gameState: updatedGameState
  };
}

/**
 * Checks if there's a winner on the board
 * @param {Array<string|null>} board - The current board state
 * @returns {string|null} The winner ('X' or 'O') or null if no winner
 */
export function checkWinner(board) {
  // Winning patterns: rows, columns, and diagonals
  const lines = [
    [0, 1, 2], [3, 4, 5], [6, 7, 8], // Rows
    [0, 3, 6], [1, 4, 7], [2, 5, 8], // Columns
    [0, 4, 8], [2, 4, 6]             // Diagonals
  ];

  for (const [a, b, c] of lines) {
    if (board[a] && board[a] === board[b] && board[a] === board[c]) {
      return board[a]; // Return the winner symbol ('X' or 'O')
    }
  }

  return null; // No winner
}

/**
 * Formats game state for client consumption
 * @param {GameState} gameState - Current game state
 * @param {string} roomId - Room ID
 * @returns {Object} Formatted game state for client
 */
export function formatGameState(gameState, roomId) {
  return {
    roomId,
    board: gameState.board,
    nextTurn: gameState.nextTurn,
    players: gameState.players
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
    board: gameState.board
  };
}