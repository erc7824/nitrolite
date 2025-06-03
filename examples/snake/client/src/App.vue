<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue';
import GameRoom from './components/GameRoom.vue';
import Lobby from './components/Lobby.vue';
import clearNetService from './services/ClearNetService';
import gameService from './services/GameService';
import { createWalletClient, createPublicClient, custom, http } from 'viem';
import { polygon } from 'viem/chains';
import { CryptoKeypair, generateKeyPair } from './crypto';
import { BROKER_WS_URL, CONTRACT_ADDRESSES } from './config';
import { NitroliteClientConfig } from '@erc7824/nitrolite';
import { privateKeyToAccount } from 'viem/accounts';

const nickname = ref('');
const roomId = ref('');
const currentScreen = ref('lobby'); // 'lobby' or 'game'
const errorMessage = ref('');
const isConnecting = ref(true);
const walletAddress = ref('');

// Create a new game room
const createRoom = async () => {
  if (!nickname.value.trim()) {
    errorMessage.value = 'Please enter a nickname';
    return;
  }

  // Check if we have an active channel
  let activeChannelId = await clearNetService.getActiveChannel();
  if (!activeChannelId) {
    errorMessage.value = 'Please create a channel first';
    return;
  }

  try {
    const walletAddress = clearNetService.client.walletClient.account.address;
    await gameService.createRoom(
      nickname.value.trim(),
      activeChannelId,
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
  const activeChannelId = await clearNetService.getActiveChannel();
  if (!activeChannelId) {
    errorMessage.value = 'Please join a channel first';
    return;
  }

  try {
    const walletAddress = clearNetService.client.walletClient.account.address;
    await gameService.joinRoom(
      roomId.value.trim(),
      nickname.value.trim(),
      activeChannelId,
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

// Auto-connect wallet and broker on page load
const autoConnect = async () => {
  isConnecting.value = true;

  try {
    // Check for MetaMask
    const { ethereum } = window as any;
    if (!ethereum) {
      throw new Error('MetaMask is required. Please install MetaMask extension.');
    }

    // Request accounts
    console.log('[App] Requesting MetaMask accounts...');
    const accounts = await ethereum.request({ method: 'eth_requestAccounts' });
    if (!accounts || accounts.length === 0) {
      throw new Error('No accounts found. Please connect your MetaMask wallet.');
    }

    walletAddress.value = accounts[0];
    console.log('[App] Wallet connected:', walletAddress.value);

    // Create wallet client
    const walletClient = createWalletClient({
      account: accounts[0],
      chain: polygon,
      transport: custom(ethereum)
    });

    // Create public client for reading blockchain data
    const publicClient = createPublicClient({
      chain: polygon,
      transport: http()
    });

    // Get or create session key
    let keyPair: CryptoKeypair | null = null;
    const savedKeys = localStorage.getItem('crypto_keypair');

    if (savedKeys) {
      try {
        keyPair = JSON.parse(savedKeys);
      } catch (error) {
        console.error('[App] Failed to parse saved keypair, generating new one');
        keyPair = null;
      }
    }

    if (!keyPair) {
      keyPair = await generateKeyPair();
      localStorage.setItem('crypto_keypair', JSON.stringify(keyPair));
    }

    const stateWalletClient = createWalletClient({
      account: privateKeyToAccount(keyPair.privateKey),
      chain: polygon,
      transport: custom(ethereum)
    });

    // Initialize ClearNetService with Nitrolite configuration
    const config: NitroliteClientConfig = {
      // @ts-ignore
      walletClient,
      publicClient,
      stateWalletClient,
      chainId: polygon.id,
      addresses: {
        custody: CONTRACT_ADDRESSES.custody,
        adjudicator: CONTRACT_ADDRESSES.adjudicator,
        tokenAddress: CONTRACT_ADDRESSES.tokenAddress,
        guestAddress: CONTRACT_ADDRESSES.guestAddress
      },
      brokerUrl: BROKER_WS_URL,
      challengeDuration: 3600n // 1 hour in seconds
    };

    console.log('[App] Initializing ClearNetService...');
    await clearNetService.initialize(config);
    console.log('[App] ClearNetService initialized successfully');

    // Check for existing channel
    const activeChannel = await clearNetService.getActiveChannel();
    if (!activeChannel) {
      throw new Error('No active channel found. Please open a channel at apps.yellow.com');
    }

    console.log('[App] Active channel found:', activeChannel);
    isConnecting.value = false;
    errorMessage.value = '';

  } catch (error) {
    console.error('[App] Auto-connect failed:', error);
    errorMessage.value = error instanceof Error ? error.message : 'Failed to connect wallet and broker';
    isConnecting.value = false;
    // Crash the app if no channel is available
    if (error instanceof Error && error.message.includes('No active channel')) {
      throw error;
    }
  }
};

onMounted(async () => {
  console.log('[App] Component mounted');
  // Auto-connect wallet and broker first
  await autoConnect();
  // Then initialize WebSocket connection
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
      <div v-if="isConnecting" class="loading-container">
        <div class="loading-spinner"></div>
        <p>Connecting to wallet and broker...</p>
      </div>

      <div v-else-if="errorMessage" class="error-container">
        <div class="error-icon">⚠️</div>
        <h2>Connection Failed</h2>
        <p>{{ errorMessage }}</p>
        <button @click="autoConnect" class="retry-btn">Retry Connection</button>
      </div>

      <div v-else>
        <Lobby v-if="currentScreen === 'lobby'" v-model:nickname="nickname" v-model:roomId="roomId"
          :socket="gameService.getWebSocket()" :walletAddress="walletAddress"
          :errorMessage="gameService.getErrorMessage().value" @create-room="createRoom" @join-room="joinRoom" />

        <GameRoom v-else-if="currentScreen === 'game'" :roomId="gameService.getRoomId().value"
          :playerId="gameService.getPlayerId().value" :nickname="nickname" :socket="gameService.getWebSocket()"
          @exit-game="currentScreen = 'lobby'" />
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

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  gap: 20px;
}

.loading-spinner {
  width: 50px;
  height: 50px;
  border: 5px solid #f3f3f3;
  border-top: 5px solid #4CAF50;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  text-align: center;
  padding: 20px;
}

.error-icon {
  font-size: 4rem;
  margin-bottom: 20px;
}

.error-container h2 {
  color: #d32f2f;
  margin-bottom: 10px;
}

.error-container p {
  color: #666;
  margin-bottom: 30px;
  max-width: 500px;
}

.retry-btn {
  padding: 12px 24px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
}

.retry-btn:hover {
  background-color: #388E3C;
}
</style>
