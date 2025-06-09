import { ref } from 'vue';
import type { Ref } from 'vue';
import clearNetService from './ClearNetService';
import { GAMESERVER_WS_URL } from '../config';
import { createEthersSigner } from '../crypto';

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

export interface Room {
  id: string;
  name: string;
  players: Array<{
    id: string;
    nickname: string;
  }>;
  maxPlayers: number;
  isGameActive: boolean;
  createdAt: number;
}

class GameService {
  private ws: WebSocket | null = null;
  private isConnected: Ref<boolean> = ref(false);
  private playerId: Ref<string> = ref('');
  private roomId: Ref<string> = ref('');
  private errorMessage: Ref<string> = ref('');
  private gameState: Ref<GameState | null> = ref(null);
  private availableRooms: Ref<Room[]> = ref([]);
  private signatureStatus: Ref<string> = ref(''); // For UI feedback
  private isSigningAppSession: Ref<boolean> = ref(false);
  private messageHandlers: Map<string, (data: any) => void> = new Map();
  private connectionPromise: Promise<void> | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 3;
  private isConnecting = false;
  private reconnectTimeout: number | null = null;
  private isCreatingRoom = false; // Add flag to track room creation state

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

    this.messageHandlers.set('roomsList', (data) => {
      console.log('[GameService] Rooms list received:', data);
      this.availableRooms.value = data.rooms || [];
    });

    this.messageHandlers.set('roomUpdated', (data) => {
      console.log('[GameService] Room updated:', data);
      // Update the specific room in the list
      const roomIndex = this.availableRooms.value.findIndex(room => room.id === data.room.id);
      if (roomIndex !== -1) {
        this.availableRooms.value[roomIndex] = data.room;
      } else {
        // If room doesn't exist, add it
        this.availableRooms.value.push(data.room);
      }
    });

    this.messageHandlers.set('roomRemoved', (data) => {
      console.log('[GameService] Room removed:', data);
      this.availableRooms.value = this.availableRooms.value.filter(room => room.id !== data.roomId);
    });

    this.messageHandlers.set('playAgainVoteUpdate', (data) => {
      console.log('[GameService] Play again vote update:', data);
      // Handle play again vote updates if needed
    });

    this.messageHandlers.set('playerDisconnected', (data) => {
      console.log('[GameService] Player disconnected:', data);
      // Handle player disconnect notifications if needed
    });

    // App session signature collection handlers
    this.messageHandlers.set('appSession:signatureRequest', (data) => {
      console.log('[GameService] App session signature request:', data);
      this.handleAppSessionSignatureRequest(data);
    });

    this.messageHandlers.set('appSession:startGameRequest', (data) => {
      console.log('[GameService] App session start game request:', data);
      this.handleAppSessionStartGameRequest(data);
    });

    this.messageHandlers.set('appSession:signatureConfirmed', (data) => {
      console.log('[GameService] App session signature confirmed:', data);
      this.signatureStatus.value = 'Waiting for host to start game...';
    });

    this.messageHandlers.set('appSession:gameStarted', (data) => {
      this.handleGameStarted(data);
    });

    this.messageHandlers.set('appSession:closeRequest', (data) => {
      this.handleCloseSessionSignatureRequest(data);
    });

    this.messageHandlers.set('appSession:closed', () => {
      console.log('[GameService] App session closed, clearing room state');
      this.clearRoomState();
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
            console.log('[GameService] Received WebSocket message:', event.data);
            const data = JSON.parse(event.data);
            const handler = this.messageHandlers.get(data.type);
            if (handler) {
              handler(data);
            } else {
              console.log('[GameService] No handler found for message type:', data.type);
            }
          } catch (error) {
            console.error('[GameService] Error parsing message:', error);
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
    }, 1000) as unknown as number;
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
      // Check if room creation is already in progress
      if (this.isCreatingRoom) {
        console.log('[GameService] Room creation already in progress, ignoring duplicate request');
        return;
      }

      this.isCreatingRoom = true;
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
    } finally {
      // Reset the flag after a short delay to prevent race conditions
      setTimeout(() => {
        this.isCreatingRoom = false;
      }, 1000);
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
      console.error('[GameService] Cannot finalize game - WebSocket not connected');
      return;
    }

    console.log('[GameService] Sending finalizeGame message with data:', {
      roomId: this.roomId.value,
      playerId: this.playerId.value,
      walletAddress: window.ethereum.selectedAddress
    });

    this.ws.send(JSON.stringify({
      type: 'finalizeGame',
      roomId: this.roomId.value,
      playerId: this.playerId.value,
      walletAddress: window.ethereum.selectedAddress
    }));
  }

  clearRoomState() {
    console.log('[GameService] Clearing room state');
    this.roomId.value = '';
    this.playerId.value = '';
    this.gameState.value = null;
    this.errorMessage.value = '';
    this.isSigningAppSession.value = false;
    this.signatureStatus.value = '';
  }

  async subscribeToRooms() {
    try {
      await this.ensureConnected();

      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        throw new Error('Not connected to server');
      }

      console.log('[GameService] Subscribing to rooms updates');
      this.ws.send(JSON.stringify({
        type: 'subscribeRooms'
      }));
    } catch (error) {
      console.error('Error subscribing to rooms:', error);
    }
  }

  async unsubscribeFromRooms() {
    try {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        return; // Connection is already closed
      }

      console.log('[GameService] Unsubscribing from rooms updates');
      this.ws.send(JSON.stringify({
        type: 'unsubscribeRooms'
      }));
    } catch (error) {
      console.error('Error unsubscribing from rooms:', error);
    }
  }

  async joinRoomById(roomId: string, nickname: string, channelId: string, walletAddress: string) {
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

  private async handleAppSessionSignatureRequest(data: any) {
    try {
      const { roomId, requestToSign, participantAddress } = data;

      console.log('[GameService] Signing app session request for room:', roomId);
      console.log('[GameService] Request to sign:', requestToSign);

      this.isSigningAppSession.value = true;
      this.signatureStatus.value = 'Please sign the app session creation request...';

      const signature = await this.signAppSessionRequest(requestToSign);
      this.signatureStatus.value = 'Signature submitted, processing...';

      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({
          type: 'appSession:signature',
          roomId,
          signature,
          participantAddress
        }));
        console.log('[GameService] Sent app session signature for room:', roomId);
      }
    } catch (error) {
      console.error('[GameService] Error handling app session signature request:', error);
      this.isSigningAppSession.value = false;
      this.signatureStatus.value = '';
      this.errorMessage.value = 'Failed to sign app session request. Please try again.';
    }
  }

  private async handleAppSessionStartGameRequest(data: any) {
    try {
      const { roomId, requestToSign, participantAddress } = data;

      console.log('[GameService] Signing start game request for room:', roomId);
      console.log('[GameService] Request to sign:', requestToSign);

      this.isSigningAppSession.value = true;
      this.signatureStatus.value = 'Please sign to start the game...';

      const signature = await this.signAppSessionRequest(requestToSign);
      this.signatureStatus.value = 'Starting game...';

      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({
          type: 'appSession:startGame',
          roomId,
          signature,
          participantAddress
        }));
        console.log('[GameService] Sent start game signature for room:', roomId);
      }
    } catch (error) {
      console.error('[GameService] Error handling start game request:', error);
      this.isSigningAppSession.value = false;
      this.signatureStatus.value = '';
      this.errorMessage.value = 'Failed to sign start game reqest. Please try again.';
    }
  }

  private async signAppSessionRequest(requestToSign: any): Promise<string> {
    try {
      const messageString = JSON.stringify(requestToSign);
      console.log('[GameService] Signing request:', messageString);
      const keyPair = await clearNetService.getOrCreateKeyPair();
      const signer = createEthersSigner(keyPair.privateKey);
      console.log('[GameService] Successfully signed with session key');
      return signer.sign(requestToSign);
    } catch (error) {
      console.error('[GameService] Error signing app session request:', error);
      throw error;
    }
  }

  private async handleCloseSessionSignatureRequest(data: any) {
    try {
      const { roomId, requestToSign, participantAddress } = data;

      console.log('[GameService] Signing app session close request for room:', roomId);
      console.log('[GameService] Request to sign:', requestToSign);

      this.isSigningAppSession.value = true;
      this.signatureStatus.value = 'Please sign to close the game session...';

      const signature = await this.signAppSessionRequest(requestToSign);
      this.signatureStatus.value = 'Signature submitted, closing session...';

      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({
          type: 'closeSessionSignature',
          roomId,
          playerId: this.playerId.value,
          signature,
          participantAddress
        }));
        console.log('[GameService] Sent app session close signature for room:', roomId);
      }
    } catch (error) {
      console.error('[GameService] Error handling app session close request:', error);
      this.signatureStatus.value = '';
      this.errorMessage.value = 'Failed to sign app session close request. Please try again.';
    }
    this.isSigningAppSession.value = false;
  }

  private handleGameStarted(data: any) {
    console.log('[GameService] App session game started:', data);
    this.isSigningAppSession.value = false;
    this.signatureStatus.value = '';
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

  getAvailableRooms(): Ref<Room[]> {
    return this.availableRooms;
  }

  getSignatureStatus(): Ref<string> {
    return this.signatureStatus;
  }

  getIsSigningAppSession(): Ref<boolean> {
    return this.isSigningAppSession;
  }

  // Get the WebSocket instance
  getWebSocket(): WebSocket | null {
    return this.ws;
  }
}

// Create a singleton instance
const gameService = new GameService();
export default gameService;
