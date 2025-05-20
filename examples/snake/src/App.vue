<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue';
import GameRoom from './components/GameRoom.vue';
import LobbyScreen from './components/LobbyScreen.vue';
import clearNetService from './services/ClearNetService';
import gameService from './services/GameService';

const nickname = ref('');
const roomId = ref('');
const currentScreen = ref('lobby'); // 'lobby' or 'game'
const errorMessage = ref('');

// Create a new game room
const createRoom = async () => {
  if (!nickname.value.trim()) {
    errorMessage.value = 'Please enter a nickname';
    return;
  }

  // Check if we have an active channel
  let activeChannel = clearNetService.getActiveChannel();
  if (!activeChannel) {
    // Try to restore channel from storage first
    clearNetService.restoreChannelFromStorage();
    activeChannel = clearNetService.getActiveChannel();
    if (!activeChannel) {
      errorMessage.value = 'Please create a channel first';
      return;
    }
  }

  try {
    const walletAddress = clearNetService.client.walletClient.account.address;
    await gameService.createRoom(
      nickname.value.trim(),
      activeChannel.channelId,
      walletAddress
    );
  } catch (error) {
    console.error('Error creating room:', error);
    // Error message is already set in GameService
  }
};

// Join an existing game room
const joinRoom = async () => {
  if (!nickname.value.trim()) {
    errorMessage.value = 'Please enter a nickname';
    return;
  }

  if (!roomId.value.trim()) {
    errorMessage.value = 'Please enter a room ID';
    return;
  }

  // Check if we have an active channel
  const activeChannel = clearNetService.getActiveChannel();
  if (!activeChannel) {
    errorMessage.value = 'Please join a channel first';
    return;
  }

  try {
    const walletAddress = clearNetService.client.walletClient.account.address;
    await gameService.joinRoom(
      roomId.value.trim(),
      nickname.value.trim(),
      activeChannel.channelId,
      walletAddress
    );
  } catch (error) {
    console.error('Error joining room:', error);
    // Error message is already set in GameService
  }
};

// Watch for room events from GameService
watch(gameService.getRoomId(), (newRoomId) => {
  console.log('[App] Room ID changed:', { newRoomId, currentScreen: currentScreen.value });
  if (newRoomId) {
    // Switch to game screen when room is created or joined
    currentScreen.value = 'game';
    // Clear any error messages when successfully entering game
    errorMessage.value = '';
  }
});

// Watch for connection state changes
watch(gameService.getIsConnected(), (isConnected) => {
  console.log('[App] Connection state changed:', { isConnected, currentScreen: currentScreen.value });
  if (!isConnected && currentScreen.value === 'game') {
  // Only show error if we're in the game screen
    errorMessage.value = 'Connection lost. Please refresh the page.';
  } else if (isConnected) {
    errorMessage.value = '';
  }
});

// Watch for screen changes
watch(() => currentScreen.value, (newScreen, oldScreen) => {
  console.log('[App] Screen changed:', { oldScreen, newScreen });
  if (oldScreen === 'game' && newScreen === 'lobby') {
    // Game ended, handle channel closing
    console.log('[App] Game ended, finalizing game');
    gameService.finalizeGame();
  }
});

// Watch for channel state changes
watch(() => clearNetService.getActiveChannel(), (newChannel) => {
  if (!newChannel) {
    // If channel is lost, show error
    errorMessage.value = 'Channel connection lost. Please create a new channel.';
  } else {
    // Clear any channel-related errors
    if (errorMessage.value.includes('channel')) {
      errorMessage.value = '';
    }
  }
});

onMounted(() => {
  console.log('[App] Component mounted');
  // Initialize WebSocket connection - this will be handled by GameService
  gameService.connect();
});

onUnmounted(() => {
  console.log('[App] Component unmounting');
});
</script>

<template>
  <div class="container">
    <header>
      <h1>Nitro Snake</h1>
    </header>

    <main>
      <div v-if="!gameService.getIsConnected().value && currentScreen === 'lobby'" class="connection-error">
        {{ errorMessage }}
      </div>
      <div v-else>
        <LobbyScreen
          v-if="currentScreen === 'lobby'"
          v-model:nickname="nickname"
          v-model:roomId="roomId"
          :socket="gameService.getWebSocket()"
          :errorMessage="gameService.getErrorMessage().value" @create-room="createRoom" @join-room="joinRoom" />

        <GameRoom
          v-else-if="currentScreen === 'game'"
          :roomId="gameService.getRoomId().value"
          :playerId="gameService.getPlayerId().value"
          :nickname="nickname"
          :socket="gameService.getWebSocket()"
          @exit-game="currentScreen = 'lobby'"
        />
      </div>
    </main>
  </div>
</template>

<style scoped>
.container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
}

header {
  text-align: center;
  margin-bottom: 30px;
}

h1 {
  color: #4CAF50;
  margin: 0;
  font-size: 2.5rem;
}

.connection-error {
  text-align: center;
  padding: 20px;
  background-color: #f8f8f8;
  border-radius: 8px;
  color: #666;
}
</style>
