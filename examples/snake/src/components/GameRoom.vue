<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue';
import gameService from '../services/GameService';

const props = defineProps<{
  roomId: string;
  playerId: string;
  nickname: string;
}>();

const emit = defineEmits(['exit-game']);

interface Player {
  id: string;
  nickname: string;
  segments: Array<{ x: number; y: number }>;
  score: number;
  isDead: boolean;
}

interface GameState {
  players: Player[];
  food: { x: number; y: number };
  gridSize: { width: number; height: number };
  isGameOver: boolean;
}

const canvasRef = ref<HTMLCanvasElement | null>(null);
const ctx = ref<CanvasRenderingContext2D | null>(null);
const gameState = ref<GameState | null>(null);
const cellSize = ref(15); // Size of each cell in pixels
const isGameStarted = ref(false);
const waitingForPlayer = ref(false);
const gameOver = ref(false);
const copySuccess = ref(false);

// Handle incoming WebSocket messages
const handleMessage = (event: MessageEvent) => {
  try {
    const data = JSON.parse(event.data);

    if (data.type === 'gameState') {
      gameState.value = data;
      isGameStarted.value = true;
      waitingForPlayer.value = false;

      // Update game over state from server
      if (data.isGameOver) {
        gameOver.value = true;
      }
    }
  } catch (error) {
    console.error('Error parsing message:', error);
  }
};

const handleKeyDown = (event: KeyboardEvent) => {
  const ws = gameService.getWebSocket();
  if (!ws || !isGameStarted.value) return;

  if (gameOver.value || gameState.value?.isGameOver) return;

  const currentPlayer = gameState.value?.players.find(player => player.id === props.playerId);
  if (currentPlayer?.isDead) return;

  let direction: string | null = null;

  switch (event.key) {
    case 'ArrowUp':
    case 'w':
    case 'W':
      direction = 'up';
      break;
    case 'ArrowDown':
    case 's':
    case 'S':
      direction = 'down';
      break;
    case 'ArrowLeft':
    case 'a':
    case 'A':
      direction = 'left';
      break;
    case 'ArrowRight':
    case 'd':
    case 'D':
      direction = 'right';
      break;
  }

  if (direction) {
    ws.send(JSON.stringify({
      type: 'changeDirection',
      direction
    }));
  }
};

// Draw game on canvas
const drawGame = () => {
  if (!ctx.value || !gameState.value || !canvasRef.value) return;

  const { width, height } = gameState.value.gridSize;
  const canvas = canvasRef.value;

  // Set canvas size
  canvas.width = width * cellSize.value;
  canvas.height = height * cellSize.value;

  // Clear canvas
  ctx.value.clearRect(0, 0, canvas.width, canvas.height);

  // Draw grid (optional)
  ctx.value.strokeStyle = '#eee';
  ctx.value.lineWidth = 0.5;

  for (let x = 0; x <= width; x++) {
    ctx.value.beginPath();
    ctx.value.moveTo(x * cellSize.value, 0);
    ctx.value.lineTo(x * cellSize.value, height * cellSize.value);
    ctx.value.stroke();
  }

  for (let y = 0; y <= height; y++) {
    ctx.value.beginPath();
    ctx.value.moveTo(0, y * cellSize.value);
    ctx.value.lineTo(width * cellSize.value, y * cellSize.value);
    ctx.value.stroke();
  }

  // Draw food
  const { food } = gameState.value;
  ctx.value.fillStyle = '#e91e63';
  ctx.value.beginPath();
  ctx.value.arc(
    (food.x + 0.5) * cellSize.value,
    (food.y + 0.5) * cellSize.value,
    cellSize.value / 2,
    0,
    Math.PI * 2
  );
  ctx.value.fill();

  // Draw players
  const colors = ['#4CAF50', '#2196F3']; // Green for player 1, blue for player 2

  gameState.value.players.forEach((player, index) => {
    const color = colors[index % colors.length];
    const isCurrentPlayer = player.id === props.playerId;

    // Only draw if player isn't dead
    if (!player.isDead) {
      // Draw snake segments
      player.segments.forEach((segment, segIndex) => {
        ctx.value!.fillStyle = color;

        // Draw head slightly larger and with a border if it's the current player
        if (segIndex === 0) {
          if (isCurrentPlayer) {
            // Draw border for head of current player
            ctx.value!.fillRect(
              segment.x * cellSize.value,
              segment.y * cellSize.value,
              cellSize.value,
              cellSize.value
            );

            // Draw white eyes
            ctx.value!.fillStyle = 'white';
            ctx.value!.beginPath();
            ctx.value!.arc(
              (segment.x + 0.3) * cellSize.value,
              (segment.y + 0.3) * cellSize.value,
              cellSize.value / 8,
              0,
              Math.PI * 2
            );
            ctx.value!.arc(
              (segment.x + 0.7) * cellSize.value,
              (segment.y + 0.3) * cellSize.value,
              cellSize.value / 8,
              0,
              Math.PI * 2
            );
            ctx.value!.fill();
          } else {
            // Regular head for other player
            ctx.value!.fillRect(
              segment.x * cellSize.value,
              segment.y * cellSize.value,
              cellSize.value,
              cellSize.value
            );
          }
        } else {
          // Regular body segment
          ctx.value!.fillRect(
            segment.x * cellSize.value + 1,
            segment.y * cellSize.value + 1,
            cellSize.value - 2,
            cellSize.value - 2
          );
        }
      });
    }

    // Draw player nickname and score (for all players, even dead ones)
    const status = player.isDead ? " (DEAD)" : "";
    ctx.value!.fillStyle = player.isDead ? "#999" : color;
    ctx.value!.font = '14px Arial';
    ctx.value!.textAlign = 'left';
    const scoreY = index === 0 ? 20 : 40;
    ctx.value!.fillText(`${player.nickname}${status}: ${player.score}`, 10, scoreY);
  });
};

// Animation loop
const animationFrame = ref<number>(0);
const animate = () => {
  drawGame();
  animationFrame.value = requestAnimationFrame(animate);
};

// Watch for changes to socket and re-attach event listener
watch(() => gameService.getIsConnected().value, (_) => {
  const ws = gameService.getWebSocket();
  if (ws) {
    ws.addEventListener('message', handleMessage);
  }
}, { immediate: true });

// Set up game when component is mounted
onMounted(() => {
  console.log('[GameRoom] Component mounted');

  window.addEventListener('keydown', handleKeyDown);

  if (canvasRef.value) {
    ctx.value = canvasRef.value.getContext('2d');

    // Start animation loop
    animationFrame.value = requestAnimationFrame(animate);
  }

  // If we don't receive a game state soon, show waiting message
  setTimeout(() => {
    if (!isGameStarted.value) {
      waitingForPlayer.value = true;
    }
  }, 1000);
});

// Clean up when component is unmounted
onUnmounted(() => {
  console.log('[GameRoom] Component unmounting');

  window.removeEventListener('keydown', handleKeyDown);

  if (animationFrame.value) {
    cancelAnimationFrame(animationFrame.value);
  }
});

// Helper to determine if a player has the highest score
const isHighestScore = (player: Player): boolean => {
  if (!gameState.value) return false;
  const highestScore = Math.max(...gameState.value.players.map(p => p.score));
  return player.score === highestScore;
};

// Helper to determine if there's a tie (multiple players with highest score)
const isTie = (): boolean => {
  if (!gameState.value) return false;
  const highestScore = Math.max(...gameState.value.players.map(p => p.score));
  return gameState.value.players.filter(p => p.score === highestScore).length > 1;
};

// Copy room ID to clipboard
const copyRoomId = () => {
  navigator.clipboard.writeText(props.roomId)
    .then(() => {
      copySuccess.value = true;
      setTimeout(() => {
        copySuccess.value = false;
      }, 2000);
    })
    .catch(err => {
      console.error('Failed to copy room ID: ', err);
    });
};

// Close channel and withdraw funds
const closeChannel = async () => {
  try {
    // Notify the server to finalize the game and close the app session
    const ws = gameService.getWebSocket();
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: 'finalizeGame',
        roomId: props.roomId,
        playerId: props.playerId
      }));
    }

    // Exit game and return to lobby
    emit('exit-game');
  } catch (error) {
    console.error('Error closing channel:', error);
  }
};

// Play another game with the same channel
const playAgain = () => {
  // Reset game state
  gameOver.value = false;

  // Notify server that we want to play again
  const ws = gameService.getWebSocket();
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({
      type: 'playAgain',
      roomId: props.roomId,
      playerId: props.playerId
    }));
  }
};
</script>

<template>
  <div class="game-container">
    <div class="room-info">
      <h2>
        Room:
        <span class="room-id" @click="copyRoomId" title="Click to copy">{{ roomId }}</span>
        <button class="copy-btn" @click="copyRoomId" title="Copy room ID">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
          </svg>
        </button>
      </h2>
      <p>Share this room ID with your friend to play together!</p>
      <div v-if="copySuccess" class="copy-success">Room ID copied to clipboard!</div>
    </div>

    <div class="channel-status">
      <span class="status-label">Channel Status:</span>
      <span class="status-value active">Active</span>
    </div>

    <div v-if="waitingForPlayer" class="waiting-message">
      Waiting for another player to join...
    </div>

    <div v-if="gameOver" class="game-over">
      <h2>Game Over!</h2>
      <div v-if="gameState">
        <div v-for="player in gameState.players" :key="player.id" class="player-result">
          <span :class="{ 'winner': isHighestScore(player) }">
            {{ player.nickname }}: {{ player.score }} points
            <span v-if="isHighestScore(player)" class="winner-badge">
              {{ isTie() ? 'TIE!' : 'WINNER!' }}
            </span>
          </span>
        </div>
      </div>

      <div class="game-over-actions">
        <button @click="closeChannel" class="close-channel-btn">Close Channel & Withdraw</button>
        <button @click="playAgain" class="play-again-btn">Play Again</button>
      </div>
    </div>

    <div class="game-board" :class="{ 'game-started': isGameStarted }">
      <canvas ref="canvasRef"></canvas>
    </div>

    <div class="controls-info">
      <h3>Controls</h3>
      <p>Use arrow keys or WASD to control your snake.</p>
    </div>
  </div>
</template>

<style scoped>
.game-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
}

.room-info {
  margin-bottom: 20px;
  text-align: center;
}

.room-info h2 {
  color: #333;
  margin-bottom: 5px;
}

.room-info p {
  color: #666;
  margin: 0;
}

.room-id {
  cursor: pointer;
  padding: 2px 5px;
  background-color: #f0f0f0;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.room-id:hover {
  background-color: #e0e0e0;
}

.copy-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: #666;
  vertical-align: middle;
  margin-left: 5px;
  padding: 3px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.copy-btn:hover {
  background-color: #e0e0e0;
}

.copy-success {
  margin-top: 5px;
  color: #4CAF50;
  font-size: 0.9em;
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.waiting-message {
  background-color: #fff8e1;
  color: #ff8f00;
  padding: 15px;
  border-radius: 8px;
  margin: 20px 0;
  text-align: center;
  font-weight: 600;
}

.game-board {
  display: flex;
  justify-content: center;
  align-items: center;
  border: 2px solid #ddd;
  border-radius: 4px;
  margin: 20px 0;
  overflow: hidden;
}

.game-board.game-started {
  border-color: #4CAF50;
}

canvas {
  display: block;
}

.controls-info {
  background-color: #f5f5f5;
  padding: 15px;
  border-radius: 8px;
  margin-top: 20px;
  text-align: center;
  max-width: 400px;
}

.controls-info h3 {
  margin-top: 0;
  color: #333;
}

.controls-info p {
  margin-bottom: 0;
  color: #666;
}

.game-over {
  background-color: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 20px;
  border-radius: 8px;
  margin: 20px 0;
  text-align: center;
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 100;
  min-width: 300px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.5);
}

.game-over h2 {
  color: #ff6b6b;
  margin-bottom: 20px;
}

.player-result {
  font-size: 18px;
  margin-bottom: 10px;
}

.winner {
  color: #ffd700;
  font-weight: bold;
}

.winner-badge {
  background-color: #ffd700;
  color: black;
  padding: 2px 8px;
  border-radius: 4px;
  margin-left: 10px;
  font-size: 14px;
  min-width: 60px;
  display: inline-block;
}

.room-id {
  cursor: pointer;
  padding: 2px 5px;
  background-color: #f0f0f0;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.room-id:hover {
  background-color: #e0e0e0;
}

.copy-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: #666;
  vertical-align: middle;
  margin-left: 5px;
  padding: 3px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.copy-btn:hover {
  background-color: #e0e0e0;
}

.copy-success {
  margin-top: 5px;
  color: #4CAF50;
  font-size: 0.9em;
  animation: fadeIn 0.3s ease-in-out;
}

.channel-status {
  background-color: #f8f9fa;
  border-radius: 4px;
  padding: 8px 12px;
  margin-bottom: 15px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-label {
  font-weight: 600;
  color: #666;
}

.status-value {
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 0.85em;
  font-weight: bold;
}

.status-value.active {
  background-color: #e8f5e9;
  color: #388e3c;
}

.game-over-actions {
  margin-top: 20px;
  display: flex;
  justify-content: center;
  gap: 15px;
}

.close-channel-btn, .play-again-btn {
  padding: 10px 16px;
  border: none;
  border-radius: 4px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
}

.close-channel-btn {
  background-color: #f44336;
  color: white;
}

.close-channel-btn:hover {
  background-color: #d32f2f;
}

.play-again-btn {
  background-color: #4CAF50;
  color: white;
}

.play-again-btn:hover {
  background-color: #388e3c;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>
