<script setup lang="ts">
import { defineEmits, defineProps, ref, watch } from 'vue';
import WalletConnect from './WalletConnect.vue';
import ChannelSetup from './ChannelSetup.vue';
import gameService from '../services/GameService';

const props = defineProps<{
  nickname: string;
  roomId: string;
  errorMessage: string;
}>();

const emit = defineEmits([
  'update:nickname',
  'update:roomId',
  'update:errorMessage',
  'create-room',
  'join-room'
]);

const isWalletConnected = ref(false);
const isChannelCreated = ref(false);
const isCreatingRoom = ref(false);
const isJoiningRoom = ref(false);
const walletAddress = ref('');
const channelInfo = ref<any>(null);

// Subscribe to GameService state
const gameError = gameService.getErrorMessage();

// Watch for error messages from GameService
watch(gameError, (newError) => {
  if (newError) {
    emit('update:errorMessage', newError);
  }
});

const updateNickname = (e: Event) => {
  emit('update:nickname', (e.target as HTMLInputElement).value);
};

const updateRoomId = (e: Event) => {
  emit('update:roomId', (e.target as HTMLInputElement).value);
};

const createRoom = () => {
  if (isChannelCreated.value) {
    isCreatingRoom.value = true;
    gameService.createRoom(props.nickname, channelInfo.value.channelId, walletAddress.value);
    emit('create-room');
  } else {
    console.log("createRoom", isChannelCreated.value);
    emit('update:errorMessage', 'Please create a channel first');
  }
};

const joinRoom = () => {
  if (isChannelCreated.value) {
    isJoiningRoom.value = true;
    gameService.joinRoom(props.roomId, props.nickname, channelInfo.value.channelId, walletAddress.value);
    emit('join-room');
  } else {
    emit('update:errorMessage', 'Please join a channel first');
  }
};

// Handle wallet connection events
const onWalletConnected = (data: { address: string, balance: bigint }) => {
  isWalletConnected.value = true;
  walletAddress.value = data.address;
};

const onWalletDisconnected = () => {
  isWalletConnected.value = false;
  walletAddress.value = '';
  isChannelCreated.value = false;
  channelInfo.value = null;
};

// Handle channel creation events
const onChannelCreated = async (data: any) => {
  console.log("onChannelCreated", data);
  isChannelCreated.value = true;
  channelInfo.value = data;
};

const onError = (error: string) => {
  emit('update:errorMessage', error);
};

// Generate a random room ID that resembles a SHA-256 hash
function generateRoomId(): string {
  // Generate a random array of bytes
  const randomBytes = new Uint8Array(32);
  window.crypto.getRandomValues(randomBytes);

  // Convert to hex string
  return Array.from(randomBytes)
    .map(b => b.toString(16).padStart(2, '0'))
    .join('');
}
</script>

<template>
  <div class="lobby">
    <!-- Wallet Connection Card -->
    <WalletConnect @wallet-connected="onWalletConnected" @wallet-disconnected="onWalletDisconnected" @error="onError" />

    <!-- Username Card -->
    <div class="form-container">
      <h2>Set Username</h2>

      <div class="form-group">
        <label for="nickname">Your Nickname:</label>
        <input id="nickname" type="text" :value="nickname" @input="updateNickname" placeholder="Enter your nickname"
          :disabled="isCreatingRoom || isJoiningRoom" />
      </div>
    </div>

    <!-- Channel Setup -->
    <ChannelSetup v-if="isWalletConnected && nickname" :isWalletConnected="isWalletConnected"
      :roomId="roomId || generateRoomId()" :roomCreator="!roomId" @channel-created="onChannelCreated"
      @error="onError" />

    <!-- Game Actions Card -->
    <div class="form-container">
      <h2>Game Options</h2>

      <div class="actions">
        <div class="action-group">
          <button @click="createRoom" class="btn primary"
            :disabled="!nickname || !isWalletConnected || !isChannelCreated || isCreatingRoom">
            {{ isCreatingRoom ? 'Creating Room...' : 'Create New Room' }}
          </button>

          <div class="requirements" v-if="!isWalletConnected || !nickname || !isChannelCreated">
            <div v-if="!isWalletConnected" class="requirement">⚠️ Connect wallet first</div>
            <div v-if="!nickname" class="requirement">⚠️ Set a nickname first</div>
            <div v-if="isWalletConnected && nickname && !isChannelCreated" class="requirement">⚠️ Create a channel first
            </div>
          </div>
        </div>

        <div class="divider">OR</div>

        <div class="action-group">
          <div class="form-group">
            <label for="roomId">Room ID:</label>
            <input id="roomId" type="text" :value="roomId" @input="updateRoomId" placeholder="Enter room ID"
              :disabled="isJoiningRoom" />
          </div>
          <button @click="joinRoom" class="btn secondary"
            :disabled="!nickname || !roomId || !isWalletConnected || !isChannelCreated || isJoiningRoom">
            {{ isJoiningRoom ? 'Joining Room...' : 'Join Existing Room' }}
          </button>

          <div class="requirements" v-if="!isWalletConnected || !nickname || !roomId || !isChannelCreated">
            <div v-if="!isWalletConnected" class="requirement">⚠️ Connect wallet first</div>
            <div v-if="!nickname" class="requirement">⚠️ Set a nickname first</div>
            <div v-if="!roomId" class="requirement">⚠️ Enter a room ID</div>
            <div v-if="isWalletConnected && nickname && roomId && !isChannelCreated" class="requirement">⚠️ Join a
              channel first</div>
          </div>
        </div>
      </div>

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
}

.form-container {
  background-color: #f8f8f8;
  border-radius: 8px;
  padding: 30px;
  width: 100%;
  max-width: 500px;
  box-shadow: 0 3px 6px rgba(0, 0, 0, 0.1);
}

h2 {
  text-align: center;
  margin-top: 0;
  margin-bottom: 25px;
  color: #333;
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: block;
  margin-bottom: 6px;
  font-weight: 600;
  color: #555;
}

input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 16px;
}

input:focus {
  border-color: #4CAF50;
  outline: none;
}

.actions {
  margin-top: 25px;
}

.action-group {
  margin-bottom: 20px;
}

.btn {
  width: 100%;
  padding: 12px;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
}

.primary {
  background-color: #4CAF50;
  color: white;
}

.primary:hover {
  background-color: #388E3C;
}

.secondary {
  background-color: #2196F3;
  color: white;
}

.secondary:hover {
  background-color: #1976D2;
}

.divider {
  text-align: center;
  margin: 15px 0;
  color: #888;
  position: relative;
}

.divider::before,
.divider::after {
  content: '';
  position: absolute;
  top: 50%;
  width: 45%;
  height: 1px;
  background-color: #ddd;
}

.divider::before {
  left: 0;
}

.divider::after {
  right: 0;
}

.error-message {
  background-color: #ffebee;
  color: #c62828;
  padding: 10px;
  border-radius: 4px;
  margin-top: 20px;
  text-align: center;
}

.requirements {
  margin-top: 10px;
  font-size: 0.85em;
}

.requirement {
  color: #f44336;
  margin-bottom: 4px;
}
</style>
