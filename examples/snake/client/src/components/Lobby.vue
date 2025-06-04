<script setup lang="ts">
import { defineEmits, defineProps, ref, watch, onMounted, onUnmounted } from 'vue';
import gameService, { type Room } from '../services/GameService';
import clearNetService from '../services/ClearNetService';
import { Hex } from 'viem';

const props = defineProps<{
  nickname: string;
  roomId: string;
  errorMessage: string;
  walletAddress: string;
}>();

const emit = defineEmits([
  'update:nickname',
  'update:roomId',
  'update:errorMessage',
  'create-room',
  'join-room'
]);

const isCreatingRoom = ref(false);
const isJoiningRoom = ref(false);
const channelInfo = ref<Hex | null>(null);
const activeTab = ref<'create' | 'join'>('create');
const availableRooms = gameService.getAvailableRooms();

onMounted(async () => {
  try {
    channelInfo.value = await clearNetService.getActiveChannel();
    // Subscribe to room updates when component mounts
    await gameService.subscribeToRooms();
  } catch (error) {
    console.error('Failed to load active channel:', error);
    emit('update:errorMessage', 'Failed to load active channel');
  }
});

onUnmounted(async () => {
  // Unsubscribe from room updates when component unmounts
  await gameService.unsubscribeFromRooms();
});

const gameError = gameService.getErrorMessage();

watch(gameError, (newError) => {
  if (newError) {
    emit('update:errorMessage', newError);
  }
});

const updateNickname = (e: Event) => {
  emit('update:nickname', (e.target as HTMLInputElement).value);
};

const createRoom = () => {
  if (isCreatingRoom.value) return;
  
  if (channelInfo.value) {
    isCreatingRoom.value = true;
    gameService.createRoom(props.nickname, channelInfo.value, props.walletAddress);
    emit('create-room');
  } else {
    console.error("No active channel found");
    emit('update:errorMessage', 'No active channel found');
  }
};


const setActiveTab = (tab: 'create' | 'join') => {
  activeTab.value = tab;
};

const joinRoomFromList = async (room: Room) => {
  if (isJoiningRoom.value) return;
  
  if (!channelInfo.value) {
    emit('update:errorMessage', 'No active channel found');
    return;
  }

  isJoiningRoom.value = true;
  try {
    await gameService.joinRoomById(room.id, props.nickname, channelInfo.value, props.walletAddress);
    emit('join-room');
  } catch (error) {
    console.error('Error joining room:', error);
    isJoiningRoom.value = false;
  }
};

const formatAddress = (address: string): string => {
  return `${address.slice(0, 6)}...${address.slice(-4)}`;
};


</script>

<template>
  <div class="lobby">
    <!-- Enhanced Wallet Info Card -->
    <div class="wallet-card">
      <div class="wallet-header">
        <div class="wallet-icon">
          <!-- Official MetaMask Logo from Wikipedia -->
          <img 
            src="https://upload.wikimedia.org/wikipedia/commons/3/36/MetaMask_Fox.svg" 
            alt="MetaMask" 
            width="28" 
            height="28"
          />
        </div>
        <div class="wallet-title">
          <h3>Wallet Connected</h3>
          <div class="connection-status">
            <div class="status-dot connected"></div>
            <span>Active</span>
          </div>
        </div>
      </div>
      
      <div class="wallet-details">
        <div class="detail-row">
          <div class="detail-label">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 12c0 4.97-4.03 9-9 9s-9-4.03-9-9 4.03-9 9-9 9 4.03 9 9z"></path>
              <path d="M16 12h-4v4"></path>
            </svg>
            Address
          </div>
          <div class="detail-value wallet-address">
            {{ formatAddress(props.walletAddress) }}
          </div>
        </div>
        
        <div class="detail-row">
          <div class="detail-label">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
            </svg>
            Channel
          </div>
          <div class="detail-value channel-info">
            <div v-if="channelInfo" class="channel-connected">
              <span>{{ formatAddress(channelInfo) }}</span>
              <div class="status-indicator">
                <div class="status-dot connected pulse"></div>
                <span class="status-text">Ready</span>
              </div>
            </div>
            <div v-else class="channel-loading">
              <div class="loading-spinner"></div>
              <span>Connecting...</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Game Card with Tabs -->
    <div class="form-container">
      <!-- Tab Navigation -->
      <div class="tab-nav">
        <button 
          @click="setActiveTab('create')" 
          :class="{ active: activeTab === 'create' }"
          class="tab-button"
        >
          Create Room
        </button>
        <button 
          @click="setActiveTab('join')" 
          :class="{ active: activeTab === 'join' }"
          class="tab-button"
        >
          Join Room
        </button>
      </div>

      <!-- Username Input (always visible) -->
      <div class="form-group">
        <label for="nickname">Your Nickname:</label>
        <input 
          id="nickname" 
          type="text" 
          :value="nickname" 
          @input="updateNickname" 
          placeholder="Enter your nickname"
          :disabled="isCreatingRoom || isJoiningRoom" 
        />
      </div>

      <!-- Tab Content -->
      <div class="tab-content">
        <!-- Create Room Tab -->
        <div v-if="activeTab === 'create'" class="tab-panel">
          <button 
            @click="createRoom" 
            class="btn primary"
            :disabled="!nickname || isCreatingRoom || !channelInfo"
          >
            {{ isCreatingRoom ? 'Creating Room...' : 'Create New Room' }}
          </button>
          
          <div class="requirements" v-if="!nickname || !channelInfo">
            <div v-if="!nickname" class="requirement">⚠️ Enter a nickname first</div>
            <div v-if="!channelInfo" class="requirement">⚠️ Waiting for channel connection...</div>
          </div>
        </div>

        <!-- Join Room Tab -->
        <div v-if="activeTab === 'join'" class="tab-panel">
          <div class="rooms-list">
            <div v-if="availableRooms.length === 0" class="no-rooms">
              <div class="no-rooms-icon">🎮</div>
              <p>No available rooms found</p>
              <p class="no-rooms-subtitle">Create a new room or wait for others to create one</p>
            </div>
            
            <div v-else class="rooms-container">
              <h3>Available Rooms</h3>
              <div 
                v-for="room in availableRooms" 
                :key="room.id" 
                class="room-item"
                :class="{ disabled: room.players.length >= room.maxPlayers || room.isGameActive }"
              >
                <div class="room-info">
                  <div class="room-name">{{ room.name || `Room ${room.id.slice(0, 8)}` }}</div>
                  <div class="room-players">
                    {{ room.players.length }}/{{ room.maxPlayers }} players
                  </div>
                  <div class="room-status">
                    <span v-if="room.isGameActive" class="status-badge active">In Game</span>
                    <span v-else-if="room.players.length >= room.maxPlayers" class="status-badge full">Full</span>
                    <span v-else class="status-badge waiting">Waiting</span>
                  </div>
                </div>
                
                <button 
                  @click="joinRoomFromList(room)"
                  class="btn secondary room-join-btn"
                  :disabled="!nickname || room.players.length >= room.maxPlayers || room.isGameActive || isJoiningRoom || !channelInfo"
                >
                  {{ isJoiningRoom ? 'Joining...' : 'Join' }}
                </button>
              </div>
            </div>
          </div>

          <div class="requirements" v-if="!nickname || !channelInfo">
            <div v-if="!nickname" class="requirement">⚠️ Enter a nickname first</div>
            <div v-if="!channelInfo" class="requirement">⚠️ Waiting for channel connection...</div>
          </div>
        </div>
      </div>

      <!-- Error Message -->
      <div v-if="errorMessage" class="error-message">
        {{ errorMessage }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.lobby {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  min-height: 60vh;
  gap: 20px;
  padding: 20px;
}

.form-container {
  background: white;
  border: 2px solid #ddd;
  border-radius: 8px;
  padding: 30px;
  width: 100%;
  max-width: 500px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

h2 {
  text-align: center;
  margin-top: 0;
  margin-bottom: 25px;
  color: #333;
  font-weight: 600;
  font-size: 1.5rem;
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: block;
  margin-bottom: 6px;
  font-weight: 600;
  color: #333;
}

input {
  width: 100%;
  padding: 12px 15px;
  border: 2px solid #ddd;
  border-radius: 4px;
  font-size: 16px;
  background: white;
  color: #333;
  transition: border-color 0.2s;
}

input:focus {
  border-color: #4CAF50;
  outline: none;
}

input::placeholder {
  color: #999;
}

/* Tab Navigation */
.tab-nav {
  display: flex;
  border-bottom: 2px solid #ddd;
  margin-bottom: 25px;
  background: #f5f5f5;
  border-radius: 4px 4px 0 0;
  overflow: hidden;
}

.tab-button {
  flex: 1;
  padding: 15px 20px;
  border: none;
  background: #f5f5f5;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  color: #666;
  transition: all 0.2s;
}

.tab-button:hover {
  background: #e0e0e0;
  color: #333;
}

.tab-button.active {
  background: white;
  color: #333;
  font-weight: 600;
  border-bottom: 2px solid #4CAF50;
}

.tab-button:not(:last-child) {
  border-right: 1px solid #ddd;
}

/* Tab Content */
.tab-content {
  margin-top: 10px;
}

.tab-panel {
  margin-top: 20px;
}

/* Button Styles */
.btn {
  width: 100%;
  padding: 15px 20px;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.primary {
  background: #4CAF50;
  color: white;
}

.primary:hover:not(:disabled) {
  background: #388e3c;
}

.secondary {
  background: #2196F3;
  color: white;
}

.secondary:hover:not(:disabled) {
  background: #1976d2;
}

.error-message {
  background: #f44336;
  color: white;
  padding: 15px;
  border-radius: 4px;
  margin-top: 20px;
  text-align: center;
}

.requirements {
  margin-top: 15px;
  font-size: 0.9em;
  color: #666;
}

.requirement {
  margin-bottom: 6px;
}

/* Wallet Card Styles */
.wallet-card {
  background: white;
  border: 2px solid #ddd;
  border-radius: 8px;
  padding: 0;
  width: 100%;
  max-width: 500px;
  margin-bottom: 20px;
  overflow: hidden;
}

.wallet-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 24px 16px;
  border-bottom: 1px solid #eee;
  background: #f9f9f9;
}

.wallet-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: #f0f0f0;
  border-radius: 8px;
}

.wallet-title h3 {
  color: #333;
  margin: 0;
  font-size: 1rem;
  font-weight: bold;
  margin-bottom: 4px;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #666;
  font-size: 0.8rem;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #4CAF50;
}

.status-dot.connected {
  background-color: #4CAF50;
}

.status-dot.pulse {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.2);
    opacity: 0.7;
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

.wallet-details {
  padding: 16px 24px 20px;
}

.detail-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.detail-row:last-child {
  margin-bottom: 0;
}

.detail-label {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #666;
  font-size: 0.9rem;
  font-weight: 500;
}

.detail-value {
  color: #333;
  font-weight: 600;
  font-size: 0.9rem;
}

.wallet-address {
  font-family: monospace;
  background: #f0f0f0;
  padding: 4px 8px;
  border-radius: 4px;
  border: 1px solid #ddd;
}

.channel-connected {
  display: flex;
  align-items: center;
  gap: 12px;
}

.channel-connected span {
  font-family: monospace;
  background: #f0f0f0;
  padding: 4px 8px;
  border-radius: 4px;
  border: 1px solid #ddd;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
}

.status-text {
  font-size: 0.8rem;
  color: #4CAF50;
  font-weight: 500;
}

.channel-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #666;
}

.loading-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid #ddd;
  border-top: 2px solid #4CAF50;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* Room List Styles */
.rooms-list {
  margin-top: 20px;
}

.no-rooms {
  text-align: center;
  padding: 40px 20px;
  color: #666;
}

.no-rooms-icon {
  font-size: 3rem;
  margin-bottom: 16px;
}

.no-rooms p {
  margin: 8px 0;
  font-size: 1.1rem;
}

.no-rooms-subtitle {
  font-size: 0.9rem !important;
  color: #888;
}

.rooms-container h3 {
  color: #333;
  margin-bottom: 16px;
  font-size: 1.1rem;
  font-weight: 600;
}

.room-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  margin-bottom: 12px;
  border: 2px solid #ddd;
  border-radius: 4px;
  background: white;
  transition: border-color 0.2s;
}

.room-item:hover:not(.disabled) {
  border-color: #4CAF50;
}

.room-item.disabled {
  opacity: 0.5;
  background: #f9f9f9;
}

.room-info {
  flex: 1;
}

.room-name {
  font-weight: bold;
  font-size: 1rem;
  color: #333;
  margin-bottom: 4px;
}

.room-players {
  font-size: 0.9rem;
  color: #666;
  margin-bottom: 6px;
}

.room-status {
  margin-top: 4px;
}

.status-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: bold;
  text-transform: uppercase;
}

.status-badge.waiting {
  background-color: #fff8e1;
  color: #ff8f00;
  border: 1px solid #ff8f00;
}

.status-badge.active {
  background-color: #e8f5e8;
  color: #4CAF50;
  border: 1px solid #4CAF50;
}

.status-badge.full {
  background-color: #ffebee;
  color: #f44336;
  border: 1px solid #f44336;
}

.room-join-btn {
  width: auto;
  min-width: 80px;
  margin-left: 16px;
  padding: 8px 16px;
  font-size: 0.9rem;
}
</style>
