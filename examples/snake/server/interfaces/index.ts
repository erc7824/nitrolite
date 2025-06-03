import { Hex } from 'viem';
import { WebSocket } from 'ws';

export interface SnakeWebSocket extends WebSocket {
  playerId: string;
  roomId?: string;
  channelId?: string;
}

export interface Player {
  id: string;
  nickname: string;
  position: { x: number; y: number };
  direction: 'up' | 'down' | 'left' | 'right';
  segments: Array<{ x: number; y: number }>;
  score: number;
  isDead?: boolean;
}

export interface Room {
  id: string;
  players: Map<string, Player>;
  food: { x: number; y: number };
  gameInterval: NodeJS.Timeout | null;
  gridSize: { width: number; height: number };
  isGameOver?: boolean;
  channelIds: Set<string>;
  appId?: Hex;
  playerAddresses: Map<string, Hex>; // Maps playerId to their ethereum address
  currentState: any;
  stateVersion: number;
  createdAt: number; // Timestamp when the room was created
  isClosingAppSession?: boolean;
}

export interface PendingRequest {
  resolve: Function;
  reject: Function;
  timeout: NodeJS.Timeout;
}

export interface RPCRequest {
  req: [number, string, any[], number]; // [requestId, method, params, timestamp]
  sig: string[];
}

export interface RPCResponse {
  res: [number, string, any[], number]; // [requestId, method, params, timestamp]
  sig: string[];
}

export interface ChallengeData {
  challenge_message: string;
}
