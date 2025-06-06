/**
 * Game and WebSocket types for Viper Duel
 */

import type { CreateAppSessionRequest } from '@erc7824/nitrolite';

// Snake direction
export type Direction = 'UP' | 'DOWN' | 'LEFT' | 'RIGHT';

// Bet amounts (in USDC)
export type BetAmount = 0 | 0.01 | 0.1 | 1 | 2;

// Bet option for UI display
export interface BetOption {
  value: BetAmount;
  label: string;
  disabled?: boolean;
}

// Position on the grid
export interface Position {
  x: number;
  y: number;
}

// Snake data
export interface Snake {
  body: Position[];
  direction: Direction;
  alive: boolean;
  score: number;
}

// Players in the game
export interface Players {
  player1: string; // EOA address of player 1 (host)
  player2: string; // EOA address of player 2 (guest)
}

// Game state from server
export interface GameState {
  roomId: string;
  snakes: {
    player1: Snake;
    player2: Snake;
  };
  food: Position[];
  players: Players;
  gameTime: number;
  betAmount: BetAmount;
}

// Game over state
export interface GameOver {
  winner: string | null; // 'player1', 'player2', or null for tie
  finalScores: {
    player1: number;
    player2: number;
  };
  gameTime: number;
}

// Room join payload
export interface JoinRoomPayload {
  roomId?: string | undefined; // Explicitly marked as optional
  eoa: string;
  betAmount: BetAmount;
}

// Direction change payload
export interface DirectionPayload {
  roomId: string;
  direction: Direction;
}

// WebSocket message types
export type WebSocketMessageType = 
  | 'joinRoom'
  | 'startGame'
  | 'changeDirection'
  | 'getAvailableRooms'
  | 'room:state'
  | 'room:ready'
  | 'room:created'
  | 'room:available'
  | 'game:started'
  | 'game:over'
  | 'game:update'
  | 'onlineUsers'
  | 'players:count'
  | 'error'
  | 'appSession:signatureRequest'
  | 'appSession:startGameRequest'
  | 'appSession:signatureConfirmed'
  | 'appSession:signature'
  | 'appSession:startGame';

// Base WebSocket message
export interface WebSocketMessage {
  type: WebSocketMessageType;
}

// Client -> Server messages

export interface JoinRoomMessage extends WebSocketMessage {
  type: 'joinRoom';
  payload: JoinRoomPayload;
}

export interface StartGamePayload {
  roomId: string;
}

export interface StartGameMessage extends WebSocketMessage {
  type: 'startGame';
  payload: StartGamePayload;
}

export interface DirectionChangeMessage extends WebSocketMessage {
  type: 'changeDirection';
  payload: DirectionPayload;
}

// Server -> Client messages

export interface RoomStateMessage extends WebSocketMessage, GameState {
  type: 'room:state';
}

export interface RoomReadyMessage extends WebSocketMessage {
  type: 'room:ready';
  roomId: string;
}

export interface RoomCreatedMessage extends WebSocketMessage {
  type: 'room:created';
  roomId: string;
  role: 'host' | 'guest';
}

export interface GameStartedMessage extends WebSocketMessage {
  type: 'game:started';
  roomId: string;
}

export interface GameUpdateMessage extends WebSocketMessage, GameState {
  type: 'game:update';
}

export interface GameOverMessage extends WebSocketMessage, GameOver {
  type: 'game:over';
}

export interface ErrorMessage extends WebSocketMessage {
  type: 'error';
  code: string;
  msg: string;
}

// Available Room type
export interface AvailableRoom {
  roomId: string;
  hostAddress: string;
  createdAt: number;
  betAmount: BetAmount;
}

export interface AvailableRoomsMessage extends WebSocketMessage {
  type: 'room:available';
  rooms: AvailableRoom[];
}

export interface GetAvailableRoomsMessage extends WebSocketMessage {
  type: 'getAvailableRooms';
}

export interface OnlineUsersMessage extends WebSocketMessage {
  type: 'onlineUsers' | 'players:count';
  count: number;
}

// App Session related messages

export interface AppSessionSignatureRequestMessage extends WebSocketMessage {
  type: 'appSession:signatureRequest';
  roomId: string;
  appSessionData: CreateAppSessionRequest[];
  appDefinition: unknown;
  participants: string[];
  requestToSign: unknown[];
}

export interface AppSessionStartGameRequestMessage extends WebSocketMessage {
  type: 'appSession:startGameRequest';
  roomId: string;
  appSessionData: CreateAppSessionRequest[];
  appDefinition: unknown;
  participants: string[];
  requestToSign: unknown[];
}

export interface AppSessionSignatureConfirmedMessage extends WebSocketMessage {
  type: 'appSession:signatureConfirmed';
  roomId: string;
}

export interface AppSessionSignatureMessage extends WebSocketMessage {
  type: 'appSession:signature';
  payload: {
    roomId: string;
    signature: string;
  };
}

export interface AppSessionStartGameMessage extends WebSocketMessage {
  type: 'appSession:startGame';
  payload: {
    roomId: string;
    signature: string;
  };
}

// Union type for all WebSocket messages
export type WebSocketMessages =
  | JoinRoomMessage
  | StartGameMessage
  | DirectionChangeMessage
  | RoomStateMessage
  | RoomReadyMessage
  | RoomCreatedMessage
  | GameStartedMessage
  | GameUpdateMessage
  | GameOverMessage
  | AvailableRoomsMessage
  | GetAvailableRoomsMessage
  | OnlineUsersMessage
  | ErrorMessage
  | AppSessionSignatureRequestMessage
  | AppSessionStartGameRequestMessage
  | AppSessionSignatureConfirmedMessage
  | AppSessionSignatureMessage
  | AppSessionStartGameMessage;

// MetaMask Ethereum Provider
export interface MetaMaskEthereumProvider {
  isMetaMask?: boolean;
  request: (request: { method: string; params?: Array<any> }) => Promise<any>;
  on: (event: string, listener: (...args: any[]) => void) => void;
  removeListener: (event: string, listener: (...args: any[]) => void) => void;
  selectedAddress?: string;
  isConnected?: () => boolean;
}

// Add type definition for window.ethereum
declare global {
  interface Window {
    ethereum?: MetaMaskEthereumProvider;
  }
}