import { ref } from 'vue';
import type { Ref } from 'vue';
import clearNetService from './ClearNetService';
import { GAMESERVER_WS_URL } from '../config';

export interface GameState {
  type: string;
  players: Array<{
    id: string;
    nickname: string;
    segments: Array<{ x: number; y: number }>;
    score: number;
    isDead: boolean;
  }>;
  food: { x: number; y: number };
  gridSize: { width: number; height: number };
  isGameOver: boolean;
  stateVersion: number;
  timestamp: number;
}

class GameService {
  private ws: WebSocket | null = null;
  private isConnected: Ref<boolean> = ref(false);
  private playerId: Ref<string> = ref('');
  private roomId: Ref<string> = ref('');
  private errorMessage: Ref<string> = ref('');
  private gameState: Ref<GameState | null> = ref(null);
  private messageHandlers: Map<string, (data: any) => void> = new Map();
  private connectionPromise: Promise<void> | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 3;
  private isConnecting = false;
  private reconnectTimeout: NodeJS.Timeout | null = null;

  constructor() {
    this.setupMessageHandlers();
    // Initialize connection on service creation
    this.connect();
  }

  private setupMessageHandlers() {
    this.messageHandlers.set('roomCreated', (data) => {
      console.log('[GameService] Room created:', data);
      this.roomId.value = data.roomId;
      this.playerId.value = data.playerId;
    });

    this.messageHandlers.set('roomJoined', (data) => {
      console.log('[GameService] Room joined:', data);
      this.roomId.value = data.roomId;
      this.playerId.value = data.playerId;
    });

    this.messageHandlers.set('error', (data) => {
      console.log('[GameService] Error received:', data);
      this.errorMessage.value = data.message;
    });

    this.messageHandlers.set('gameState', (data) => {
      this.gameState.value = data;
    });

    this.messageHandlers.set('signState', (data) => {
      this.handleStateSignRequest(data.channelId, data.state, data.stateId);
    });

    this.messageHandlers.set('channelFinalized', (data) => {
      console.log(`[GameService] Channel ${data.channelId} has been finalized`);
    });
  }

  connect() {
    console.log('[GameService] connect() called, current state:', {
      wsState: this.ws?.readyState,
      isConnecting: this.isConnecting,
      hasConnectionPromise: !!this.connectionPromise,
      reconnectAttempts: this.reconnectAttempts,
      roomId: this.roomId.value
    });

    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('[GameService] WebSocket already connected, returning');
      return Promise.resolve();
    }

    if (this.connectionPromise) {
      console.log('[GameService] Connection attempt already in progress, returning existing promise');
      return this.connectionPromise;
    }

    if (this.isConnecting) {
      console.log('[GameService] Already connecting, returning');
      return Promise.resolve();
    }

    this.isConnecting = true;
    this.connectionPromise = new Promise((resolve, reject) => {
      try {
        console.log('[GameService] Creating new WebSocket connection');
        this.ws = new WebSocket(GAMESERVER_WS_URL);

        this.ws.onopen = () => {
          this.isConnected.value = true;
          this.reconnectAttempts = 0;
          this.isConnecting = false;
          console.log('[GameService] WebSocket connected successfully');
          resolve();
        };

        this.ws.onclose = (event) => {
          console.log('[GameService] WebSocket closed:', {
            wasClean: event.wasClean,
            code: event.code,
            reason: event.reason,
            currentState: {
              isConnecting: this.isConnecting,
              reconnectAttempts: this.reconnectAttempts,
              hasConnectionPromise: !!this.connectionPromise,
              roomId: this.roomId.value
            }
          });

          this.isConnecting = false;
          if (!event.wasClean) {
            this.handleDisconnection();
          } else {
            this.isConnected.value = false;
            this.connectionPromise = null;
            // Only schedule reconnect if this wasn't an intentional disconnect
            // and we don't have an active room
            if (event.reason !== 'Client disconnected' && !this.roomId.value) {
              this.scheduleReconnect();
            }
          }
        };

        this.ws.onerror = (error) => {
          this.isConnecting = false;
          console.error('WebSocket error:', error);
        };

        this.ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            const handler = this.messageHandlers.get(data.type);
            if (handler) {
              handler(data);
            }
          } catch (error) {
            console.error('Error parsing message:', error);
          }
        };
      } catch (error) {
        this.isConnecting = false;
        this.connectionPromise = null;
        reject(error);
      }
    });

    return this.connectionPromise;
  }

  private scheduleReconnect() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }

    this.reconnectTimeout = setTimeout(() => {
      if (!this.isConnected.value && !this.isConnecting) {
        console.log('Attempting to reconnect...');
        this.connect();
      }
    }, 1000);
  }

  private handleDisconnection() {
    this.isConnected.value = false;
    this.connectionPromise = null;
    this.isConnecting = false;

    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
      this.scheduleReconnect();
    } else {
      this.errorMessage.value = 'Connection lost. Please refresh the page.';
    }
  }

  disconnect() {
    console.log('[GameService] disconnect() called, current state:', {
      wsState: this.ws?.readyState,
      isConnecting: this.isConnecting,
      hasConnectionPromise: !!this.connectionPromise
    });

    if (this.ws) {
      this.ws.close(1000, 'Client disconnected');
      this.ws = null;
      this.connectionPromise = null;
      this.reconnectAttempts = 0;
      this.isConnecting = false;
      if (this.reconnectTimeout) {
        clearTimeout(this.reconnectTimeout);
        this.reconnectTimeout = null;
      }
      console.log('[GameService] WebSocket disconnected and cleaned up');
    }
  }

  private async ensureConnected() {
    if (!this.isConnected.value) {
      await this.connect();
    }
  }

  async createRoom(nickname: string, channelId: string, walletAddress: string) {
    try {
      await this.ensureConnected();

      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        throw new Error('Not connected to server');
      }

      this.ws.send(JSON.stringify({
        type: 'createRoom',
        nickname,
        channelId,
        walletAddress,
      }));
    } catch (error) {
      console.error('Error creating room:', error);
      this.errorMessage.value = 'Failed to create room. Please try again.';
      throw error;
    }
  }

  async joinRoom(roomId: string, nickname: string, channelId: string, walletAddress: string) {
    try {
      await this.ensureConnected();

      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        throw new Error('Not connected to server');
      }

      this.ws.send(JSON.stringify({
        type: 'joinRoom',
        roomId,
        nickname,
        channelId,
        walletAddress,
      }));
    } catch (error) {
      console.error('Error joining room:', error);
      this.errorMessage.value = 'Failed to join room. Please try again.';
      throw error;
    }
  }

  changeDirection(direction: 'up' | 'down' | 'left' | 'right') {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      return;
    }

    this.ws.send(JSON.stringify({
      type: 'changeDirection',
      direction
    }));
  }

  playAgain() {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      return;
    }

    this.ws.send(JSON.stringify({
      type: 'playAgain',
      roomId: this.roomId.value
    }));
  }

  finalizeGame() {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      return;
    }

    this.ws.send(JSON.stringify({
      type: 'finalizeGame',
      roomId: this.roomId.value
    }));
  }

  private async handleStateSignRequest(channelId: string, state: any, stateId: string) {
    try {
      const signatureData = await clearNetService.signState(state, stateId, channelId);

      if (signatureData && this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({
          type: 'stateSignature',
          channelId: signatureData.channelId,
          stateId: signatureData.stateId,
          signature: signatureData.signature,
          playerId: signatureData.playerId
        }));
      }
    } catch (error) {
      console.error('Error signing state:', error);
    }
  }

  // Getters for reactive state
  getIsConnected(): Ref<boolean> {
    return this.isConnected;
  }

  getPlayerId(): Ref<string> {
    return this.playerId;
  }

  getRoomId(): Ref<string> {
    return this.roomId;
  }

  getErrorMessage(): Ref<string> {
    return this.errorMessage;
  }

  getGameState(): Ref<GameState | null> {
    return this.gameState;
  }

  // Get the WebSocket instance
  getWebSocket(): WebSocket | null {
    return this.ws;
  }
}

// Create a singleton instance
const gameService = new GameService();
export default gameService;
