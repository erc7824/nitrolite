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
const INITIAL_GAME_STATE: GameState = {
  roomId: '',
  snakes: {
    player1: {
      body: [],
      direction: 'RIGHT',
      alive: false,
      score: 0
    },
    player2: {
      body: [],
      direction: 'LEFT', 
      alive: false,
      score: 0
    }
  },
  food: [],
  players: { player1: '', player2: '' },
  gameTime: 0,
  betAmount: 0
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

  // Determine player's role (player1 or player2)
  const playerId = gameState.players.player1 === eoaAddress ? 'player1' : 
                   gameState.players.player2 === eoaAddress ? 'player2' : null;

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
          snakes: lastMessage.snakes,
          food: lastMessage.food,
          players: lastMessage.players,
          gameTime: lastMessage.gameTime,
          betAmount: lastMessage.betAmount || 0
        });
        
        // Set host status based on player role (player1 is always host)
        if (lastMessage.players.player1 === eoaAddress) {
          console.log("Player is host (player1)");
          setIsHost(true);
        } else {
          console.log("Player is guest (player2)");
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

      case 'game:update':
        console.log("🐍 GAME UPDATE:", {
          gameTime: lastMessage.gameTime,
          player1Pos: lastMessage.snakes?.player1?.body?.[0],
          player2Pos: lastMessage.snakes?.player2?.body?.[0],
          player1Alive: lastMessage.snakes?.player1?.alive,
          player2Alive: lastMessage.snakes?.player2?.alive
        });
        setGameState({
          roomId: lastMessage.roomId,
          snakes: lastMessage.snakes,
          food: lastMessage.food,
          players: lastMessage.players,
          gameTime: lastMessage.gameTime,
          betAmount: lastMessage.betAmount || 0
        });
        setErrorMessage(null);
        break;

      case 'game:over':
        setGameOver({
          winner: lastMessage.winner,
          finalScores: lastMessage.finalScores,
          gameTime: lastMessage.gameTime
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
    if (!playerId) return '';
    return playerId === 'player1' ? gameState.players.player2 : gameState.players.player1;
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
    playerId,
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