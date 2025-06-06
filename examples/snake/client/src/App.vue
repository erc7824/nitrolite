<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue';
import GameRoom from './components/GameRoom.vue';
import Lobby from './components/Lobby.vue';
import clearNetService from './services/ClearNetService';
import gameService from './services/GameService';
import { createWalletClient, custom, Hex } from 'viem';
import { polygon } from 'viem/chains';
import { CryptoKeypair } from './crypto';
import { ethers } from "ethers";

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
  if (!clearNetService) {
    errorMessage.value = 'ClearNet service not initialized';
    return;
  }

  // Check if we have an active channel
  let activeChannelId = await clearNetService.getActiveChannel();
  if (!activeChannelId) {
    errorMessage.value = 'Please create a channel first';
    return;
  }

  try {
    const walletAddress = clearNetService.walletClient?.account.address as Hex;
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
  if (!clearNetService) {
    errorMessage.value = 'ClearNet service not initialized';
    return;
  }
  const activeChannelId = await clearNetService.getActiveChannel();
  if (!activeChannelId) {
    errorMessage.value = 'Please join a channel first';
    return;
  }

  try {
    const walletAddress = clearNetService.walletClient?.account.address as Hex;
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

    // Get or create session key
    let keyPair: CryptoKeypair = await clearNetService.getOrCreateKeyPair();
    const wallet = new ethers.Wallet(keyPair.privateKey);
    const stateWalletClient = {
      ...wallet,
      account: { address: wallet.address, },
      signMessage: async ({ message: { raw } }: { message: { raw: string } }) => {
        const { serialized: signature } = wallet.signingKey.sign(raw as ethers.BytesLike);

        return signature as Hex;
      },
    };

    console.log('[App] Initializing ClearNetService...');
    // @ts-ignore
    await clearNetService.initialize(walletClient, stateWalletClient);
    console.log('[App] ClearNetService initialized successfully');

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
        <!-- Signature collection overlay -->
        <div v-if="gameService.getIsSigningAppSession().value" class="signature-overlay">
          <div class="signature-container">
            <div class="signature-spinner"></div>
            <h3>App Session Creation</h3>
            <p>{{ gameService.getSignatureStatus().value }}</p>
            <small>Please check your wallet for signature requests</small>
          </div>
        </div>

        <Lobby v-if="currentScreen === 'lobby'" v-model:nickname="nickname" v-model:roomId="roomId"
          :socket="gameService.getWebSocket()" :walletAddress="walletAddress"
          :errorMessage="gameService.getErrorMessage().value" @create-room="createRoom" @join-room="joinRoom" />

        <GameRoom v-else-if="currentScreen === 'game'" :roomId="gameService.getRoomId().value" :walletAddress="walletAddress"
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

.signature-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.signature-container {
  background: white;
  padding: 40px;
  border-radius: 12px;
  text-align: center;
  max-width: 400px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.signature-container h3 {
  color: #4CAF50;
  margin-bottom: 20px;
  font-size: 1.5rem;
}

.signature-container p {
  color: #333;
  margin-bottom: 20px;
  font-size: 1.1rem;
  font-weight: 500;
}

.signature-container small {
  color: #666;
  font-size: 0.9rem;
}

.signature-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #4CAF50;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 20px;
}
</style>
