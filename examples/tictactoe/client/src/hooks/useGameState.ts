import { useState, useEffect, useCallback } from 'react';
import type { 
  GameState, 
  GameOver, 
  WebSocketMessages,
  AppSessionSignatureRequestMessage,
  AppSessionStartGameRequestMessage
} from '../types';
import { useAppSessionSignature } from './useAppSessionSignature';

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
  eoaAddress: string,
  sendAppSessionSignature?: (roomId: string, signature: string) => void,
  sendAppSessionStartGame?: (roomId: string, signature: string) => void
) {
  // Game state
  const [gameState, setGameState] = useState<GameState>(INITIAL_GAME_STATE);
  const [gameOver, setGameOver] = useState<GameOver | null>(null);
  const [roomId, setRoomId] = useState<string>('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isRoomReady, setIsRoomReady] = useState(false);
  const [isGameStarted, setIsGameStarted] = useState(false);
  const [isHost, setIsHost] = useState(false);
  const [pendingSignatureRequest, setPendingSignatureRequest] = useState<AppSessionSignatureRequestMessage | AppSessionStartGameRequestMessage | null>(null);
  const [awaitingHostStart, setAwaitingHostStart] = useState(false);

  // App session signature handling
  const { 
    isSigningInProgress, 
    signatureError, 
    handleParticipantBSignature, 
    handleParticipantASignature 
  } = useAppSessionSignature(sendAppSessionSignature, sendAppSessionStartGame);

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

      case 'appSession:signatureRequest':
        console.log("Received signature request for participant B:", lastMessage);
        setPendingSignatureRequest(lastMessage as AppSessionSignatureRequestMessage);
        
        // Automatically sign for participant B (guest)
        if (!isHost) {
          try {
            handleParticipantBSignature(lastMessage as AppSessionSignatureRequestMessage);
          } catch (error) {
            console.error('Failed to handle participant B signature:', error);
            setErrorMessage('Failed to sign app session message');
          }
        }
        break;

      case 'appSession:startGameRequest':
        console.log("Received start game request for participant A (host):", lastMessage);
        setPendingSignatureRequest(lastMessage as AppSessionStartGameRequestMessage);
        setAwaitingHostStart(true);
        break;

      case 'appSession:signatureConfirmed':
        console.log("App session signature confirmed:", lastMessage);
        setPendingSignatureRequest(null);
        setErrorMessage(null);
        break;

      case 'error':
        setErrorMessage(lastMessage.msg);
        break;

      default:
        // Ignore unknown message types
        break;
    }
  }, [lastMessage, eoaAddress, handleParticipantBSignature, isHost]);

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

  // Handle host signing and starting game
  const signAndStartGame = useCallback(async () => {
    if (!pendingSignatureRequest || pendingSignatureRequest.type !== 'appSession:startGameRequest') {
      console.error('No pending start game request');
      return;
    }

    try {
      await handleParticipantASignature(pendingSignatureRequest as AppSessionStartGameRequestMessage);
      setPendingSignatureRequest(null);
      setAwaitingHostStart(false);
    } catch (error) {
      console.error('Failed to sign and start game:', error);
      setErrorMessage('Failed to sign and start game');
    }
  }, [pendingSignatureRequest, handleParticipantASignature]);

  // Reset game state
  const resetGame = useCallback(() => {
    setGameState(INITIAL_GAME_STATE);
    setGameOver(null);
    setIsRoomReady(false);
    setIsGameStarted(false);
    setIsHost(false);
    setRoomId('');
    setErrorMessage(null);
    setPendingSignatureRequest(null);
    setAwaitingHostStart(false);
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
    resetGame,
    pendingSignatureRequest,
    awaitingHostStart,
    signAndStartGame,
    isSigningInProgress,
    signatureError
  };
}