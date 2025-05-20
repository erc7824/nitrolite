import { useState, useEffect, useCallback } from 'react';
import type { 
  GameState, 
  GameOver, 
  WebSocketMessages
} from '../types';

// Initial empty game state
const EMPTY_BOARD = Array(9).fill(null);
const INITIAL_GAME_STATE: GameState = {
  roomId: '',
  board: EMPTY_BOARD,
  nextTurn: 'X',
  players: { X: '', O: '' }
};

// Game state hook that processes WebSocket messages
export function useGameState(
  lastMessage: WebSocketMessages | null,
  eoaAddress: string
) {
  // Game state
  const [gameState, setGameState] = useState<GameState>(INITIAL_GAME_STATE);
  const [gameOver, setGameOver] = useState<GameOver | null>(null);
  const [roomId, setRoomId] = useState<string>('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isRoomReady, setIsRoomReady] = useState(false);
  const [isGameStarted, setIsGameStarted] = useState(false);
  const [isHost, setIsHost] = useState(false);

  // Determine player's role (X or O)
  const playerSymbol = gameState.players.X === eoaAddress ? 'X' : 
                      gameState.players.O === eoaAddress ? 'O' : null;

  // Is it the player's turn?
  const isPlayerTurn = playerSymbol === gameState.nextTurn;

  // We don't need to generate room IDs client-side anymore
  // The server handles room creation

  // Process WebSocket messages to update game state
  useEffect(() => {
    if (!lastMessage) return;

    console.log("Received WebSocket message:", lastMessage.type, lastMessage);

    switch (lastMessage.type) {
      case 'room:created':
        console.log("Room created:", lastMessage.roomId, "role:", lastMessage.role);
        setRoomId(lastMessage.roomId);
        
        // Set host status based on role
        if (lastMessage.role === 'host') {
          console.log("Player is host");
          setIsHost(true);
        } else {
          console.log("Player is guest");
          setIsHost(false);
        }
        
        setErrorMessage(null);
        break;
        
      case 'room:state':
        console.log("Received room:state", lastMessage, "eoaAddress:", eoaAddress);
        
        setGameState({
          roomId: lastMessage.roomId,
          board: lastMessage.board,
          nextTurn: lastMessage.nextTurn,
          players: lastMessage.players
        });
        
        // Set host status based on player role (X is always host)
        if (lastMessage.players.X === eoaAddress) {
          console.log("Player is host (X)");
          setIsHost(true);
        } else {
          console.log("Player is guest (O)");
          setIsHost(false);
        }
        
        // Always update room ID when we get a room:state message
        if (lastMessage.roomId) {
          setRoomId(lastMessage.roomId);
        }
        
        setErrorMessage(null);
        break;

      case 'room:ready':
        setRoomId(lastMessage.roomId);
        setIsRoomReady(true);
        setErrorMessage(null);
        break;
        
      case 'game:started':
        setIsGameStarted(true);
        setErrorMessage(null);
        break;

      case 'game:over':
        setGameOver({
          winner: lastMessage.winner,
          board: lastMessage.board
        });
        setErrorMessage(null);
        break;

      case 'error':
        setErrorMessage(lastMessage.msg);
        break;

      default:
        // Ignore unknown message types
        break;
    }
  }, [lastMessage, eoaAddress]);

  // Helper to format short address display
  const formatShortAddress = (address: string): string => {
    if (!address) return '';
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
  };

  // Get opponent's address
  const getOpponentAddress = (): string => {
    if (!playerSymbol) return '';
    return playerSymbol === 'X' ? gameState.players.O : gameState.players.X;
  };

  // Reset game state
  const resetGame = useCallback(() => {
    setGameState(INITIAL_GAME_STATE);
    setGameOver(null);
    setIsRoomReady(false);
    setIsGameStarted(false);
    setIsHost(false);
    setRoomId('');
    setErrorMessage(null);
  }, []);

  // TODO: Add integration with @erc7824/nitrolite for persisting game state
  
  return {
    gameState,
    gameOver,
    roomId,
    errorMessage,
    isRoomReady,
    isGameStarted,
    isHost,
    playerSymbol,
    isPlayerTurn,
    formatShortAddress,
    getOpponentAddress,
    resetGame
  };
}