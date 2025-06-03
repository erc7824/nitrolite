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
  background: 
    radial-gradient(circle at 20% 80%, rgba(255, 0, 255, 0.1) 0%, transparent 50%),
    radial-gradient(circle at 80% 20%, rgba(0, 255, 255, 0.1) 0%, transparent 50%),
    radial-gradient(circle at 40% 40%, rgba(255, 255, 0, 0.05) 0%, transparent 50%),
    linear-gradient(135deg, #000011 0%, #001122 25%, #000033 50%, #001122 75%, #000011 100%);
  padding: 20px;
  position: relative;
  overflow: hidden;
}

.lobby::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: 
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 50px,
      rgba(0, 255, 255, 0.02) 50px,
      rgba(0, 255, 255, 0.02) 52px
    ),
    repeating-linear-gradient(
      -45deg,
      transparent,
      transparent 50px,
      rgba(255, 0, 255, 0.02) 50px,
      rgba(255, 0, 255, 0.02) 52px
    );
  pointer-events: none;
}

.form-container {
  background: linear-gradient(45deg, #ff00ff 0%, #00ffff 25%, #ffff00 50%, #ff00ff 75%, #00ffff 100%);
  background-size: 400% 400%;
  animation: retroGradient 4s ease infinite;
  border: 4px solid #00ffff;
  border-radius: 0;
  padding: 30px;
  width: 100%;
  max-width: 500px;
  box-shadow: 
    0 0 20px #ff00ff,
    inset 0 0 20px rgba(0, 255, 255, 0.1),
    0 8px 0 #ff00ff,
    0 12px 0 #00ffff;
  position: relative;
  overflow: hidden;
  transition: all 0.3s ease;
  font-family: 'Courier New', monospace;
}

.form-container:hover {
  box-shadow: 
    0 0 30px #ff00ff,
    inset 0 0 30px rgba(0, 255, 255, 0.2),
    0 12px 0 #ff00ff,
    0 16px 0 #00ffff;
  transform: translateY(-2px);
}

@keyframes retroGradient {
  0% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
  100% { background-position: 0% 50%; }
}

.form-container::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: 
    repeating-linear-gradient(
      90deg,
      transparent,
      transparent 2px,
      rgba(0, 255, 255, 0.03) 2px,
      rgba(0, 255, 255, 0.03) 4px
    ),
    repeating-linear-gradient(
      0deg,
      transparent,
      transparent 2px,
      rgba(255, 0, 255, 0.03) 2px,
      rgba(255, 0, 255, 0.03) 4px
    );
  pointer-events: none;
}

h2 {
  text-align: center;
  margin-top: 0;
  margin-bottom: 25px;
  color: #00ffff;
  font-family: 'Courier New', monospace;
  font-weight: bold;
  font-size: 1.8rem;
  text-transform: uppercase;
  letter-spacing: 2px;
  text-shadow: 
    0 0 10px #00ffff,
    0 0 20px #00ffff,
    2px 2px 0 #ff00ff;
  animation: textGlow 2s ease-in-out infinite alternate;
}

@keyframes textGlow {
  from { text-shadow: 0 0 10px #00ffff, 0 0 20px #00ffff, 2px 2px 0 #ff00ff; }
  to { text-shadow: 0 0 15px #00ffff, 0 0 30px #00ffff, 2px 2px 0 #ff00ff; }
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: block;
  margin-bottom: 6px;
  font-weight: bold;
  color: #ffff00;
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
  letter-spacing: 1px;
  text-shadow: 0 0 5px #ffff00;
}

input {
  width: 100%;
  padding: 12px 15px;
  border: 2px solid #00ffff;
  border-radius: 0;
  font-size: 16px;
  background: #000;
  color: #00ff00;
  font-family: 'Courier New', monospace;
  box-shadow: 
    inset 0 0 10px rgba(0, 255, 255, 0.2),
    0 0 5px #00ffff;
  transition: all 0.3s ease;
}

input:focus {
  border-color: #ff00ff;
  outline: none;
  box-shadow: 
    inset 0 0 15px rgba(255, 0, 255, 0.3),
    0 0 10px #ff00ff;
  color: #ffff00;
}

input::placeholder {
  color: #666;
}

/* Tab Navigation */
.tab-nav {
  display: flex;
  border-bottom: 3px solid #00ffff;
  margin-bottom: 25px;
  background: linear-gradient(90deg, #ff00ff 0%, #00ffff 100%);
  padding: 0;
}

.tab-button {
  flex: 1;
  padding: 15px 20px;
  border: 2px solid #ffff00;
  border-bottom: none;
  background: #000;
  font-size: 14px;
  font-weight: bold;
  cursor: pointer;
  color: #00ffff;
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
  letter-spacing: 1px;
  transition: all 0.3s ease;
  position: relative;
  text-shadow: 0 0 5px #00ffff;
}

.tab-button:hover {
  background: #1a0033;
  color: #ff00ff;
  text-shadow: 0 0 8px #ff00ff;
  box-shadow: inset 0 0 10px rgba(255, 0, 255, 0.3);
}

.tab-button.active {
  background: linear-gradient(45deg, #ff00ff, #00ffff);
  color: #000;
  text-shadow: none;
  font-weight: bold;
  border-color: #ffff00;
  box-shadow: 
    0 0 10px #ffff00,
    inset 0 0 20px rgba(255, 255, 0, 0.2);
}

.tab-button:not(:last-child) {
  border-right: none;
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
  border: 3px solid #ff00ff;
  border-radius: 0;
  font-size: 16px;
  font-weight: bold;
  cursor: pointer;
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
  letter-spacing: 2px;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
  text-shadow: 0 0 5px currentColor;
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  box-shadow: none;
  animation: none;
}

.btn:not(:disabled):hover {
  transform: translateY(-2px);
  animation: buttonPulse 0.6s ease-in-out infinite;
}

@keyframes buttonPulse {
  0%, 100% { box-shadow: 0 0 5px currentColor; }
  50% { box-shadow: 0 0 20px currentColor, 0 0 30px currentColor; }
}

.primary {
  background: linear-gradient(45deg, #ff00ff, #ff0080);
  color: #ffff00;
  border-color: #ffff00;
  box-shadow: 
    0 0 10px #ff00ff,
    inset 0 0 10px rgba(255, 255, 0, 0.1);
}

.primary:hover:not(:disabled) {
  background: linear-gradient(45deg, #ff0080, #ff00ff);
  box-shadow: 
    0 0 20px #ff00ff,
    0 0 30px #ff00ff,
    inset 0 0 20px rgba(255, 255, 0, 0.2);
}

.secondary {
  background: linear-gradient(45deg, #00ffff, #0080ff);
  color: #000;
  border-color: #ffff00;
  box-shadow: 
    0 0 10px #00ffff,
    inset 0 0 10px rgba(255, 255, 0, 0.1);
}

.secondary:hover:not(:disabled) {
  background: linear-gradient(45deg, #0080ff, #00ffff);
  box-shadow: 
    0 0 20px #00ffff,
    0 0 30px #00ffff,
    inset 0 0 20px rgba(255, 255, 0, 0.2);
}

.error-message {
  background: #ff0000;
  color: #ffff00;
  padding: 15px;
  border: 2px solid #ffff00;
  border-radius: 0;
  margin-top: 20px;
  text-align: center;
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
  font-weight: bold;
  letter-spacing: 1px;
  text-shadow: 0 0 5px #ffff00;
  animation: errorBlink 1s ease-in-out infinite;
}

@keyframes errorBlink {
  0%, 50% { opacity: 1; }
  25%, 75% { opacity: 0.7; }
}

.requirements {
  margin-top: 15px;
  font-size: 0.85em;
  font-family: 'Courier New', monospace;
}

.requirement {
  color: #ffff00;
  margin-bottom: 6px;
  text-shadow: 0 0 3px #ffff00;
  text-transform: uppercase;
  font-weight: bold;
}

/* Enhanced Wallet Card Styles */
.wallet-card {
  background: linear-gradient(135deg, #000033 0%, #001133 100%);
  border: 2px solid #333;
  border-radius: 0;
  padding: 0;
  width: 100%;
  max-width: 500px;
  box-shadow: 
    0 0 10px rgba(0, 255, 255, 0.1),
    inset 0 0 10px rgba(0, 0, 0, 0.5);
  margin-bottom: 20px;
  overflow: hidden;
  position: relative;
  transition: all 0.3s ease;
  font-family: 'Courier New', monospace;
}

.wallet-card:hover {
  transform: translateY(-1px);
  box-shadow: 
    0 0 15px rgba(0, 255, 255, 0.2),
    inset 0 0 15px rgba(0, 0, 0, 0.5);
}

.wallet-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: 
    repeating-linear-gradient(
      0deg,
      transparent,
      transparent 1px,
      rgba(0, 255, 255, 0.05) 1px,
      rgba(0, 255, 255, 0.05) 2px
    );
  pointer-events: none;
}

.wallet-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 24px 16px;
  border-bottom: 1px solid #333;
  background: rgba(0, 0, 0, 0.3);
}

.wallet-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: rgba(0, 0, 0, 0.05);
  border-radius: 10px;
}

.wallet-title h3 {
  color: #00ffff;
  margin: 0;
  font-size: 1rem;
  font-weight: bold;
  margin-bottom: 4px;
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #666;
  font-size: 0.8rem;
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #4ade80;
  box-shadow: 0 0 8px rgba(74, 222, 128, 0.6);
}

.status-dot.connected {
  background-color: #10b981;
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
  color: #888;
  font-size: 0.85rem;
  font-weight: bold;
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
}

.detail-value {
  color: #00ffff;
  font-weight: bold;
  font-size: 0.85rem;
  font-family: 'Courier New', monospace;
}

.wallet-address {
  font-family: 'Courier New', monospace;
  background: rgba(0, 0, 0, 0.5);
  padding: 4px 8px;
  border: 1px solid #333;
  border-radius: 0;
}

.channel-connected {
  display: flex;
  align-items: center;
  gap: 12px;
}

.channel-connected span {
  font-family: 'Courier New', monospace;
  background: rgba(0, 0, 0, 0.5);
  padding: 4px 8px;
  border: 1px solid #333;
  border-radius: 0;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
}

.status-text {
  font-size: 0.8rem;
  color: #00ff00;
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
}

.channel-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #888;
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
}

.loading-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid #333;
  border-top: 2px solid #00ffff;
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
  color: #ffff00;
  margin-bottom: 16px;
  font-size: 1.1rem;
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
  letter-spacing: 1px;
  text-shadow: 0 0 5px #ffff00;
}

.room-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  margin-bottom: 12px;
  border: 2px solid #333;
  border-radius: 0;
  background: linear-gradient(90deg, #001122 0%, #002244 100%);
  transition: all 0.3s ease;
  font-family: 'Courier New', monospace;
}

.room-item:hover:not(.disabled) {
  border-color: #00ffff;
  box-shadow: 0 0 10px rgba(0, 255, 255, 0.3);
}

.room-item.disabled {
  opacity: 0.4;
  background: #000011;
}

.room-info {
  flex: 1;
}

.room-name {
  font-weight: bold;
  font-size: 1rem;
  color: #00ffff;
  margin-bottom: 4px;
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
}

.room-players {
  font-size: 0.9rem;
  color: #888;
  margin-bottom: 6px;
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
}

.room-status {
  margin-top: 4px;
}

.status-badge {
  display: inline-block;
  padding: 4px 10px;
  border: 1px solid;
  border-radius: 0;
  font-size: 0.7rem;
  font-weight: bold;
  text-transform: uppercase;
  font-family: 'Courier New', monospace;
  letter-spacing: 1px;
}

.status-badge.waiting {
  background-color: #003300;
  color: #00ff00;
  border-color: #00ff00;
  text-shadow: 0 0 3px #00ff00;
}

.status-badge.active {
  background-color: #330033;
  color: #ff00ff;
  border-color: #ff00ff;
  text-shadow: 0 0 3px #ff00ff;
}

.status-badge.full {
  background-color: #330000;
  color: #ff0000;
  border-color: #ff0000;
  text-shadow: 0 0 3px #ff0000;
}

.room-join-btn {
  width: auto;
  min-width: 80px;
  margin-left: 16px;
  padding: 8px 16px;
  font-size: 0.9rem;
}
</style>
