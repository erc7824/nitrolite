import { useEffect, useRef } from 'react';
import type { GameState, Snake } from '../types';

interface SnakeCanvasProps {
  gameState: GameState;
  playerId: string | null;
}

const GRID_SIZE = 20;
const CANVAS_SIZE = 400; // 400px canvas
const CELL_SIZE = CANVAS_SIZE / GRID_SIZE; // 20px per cell

export function SnakeCanvas({ gameState, playerId }: SnakeCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const gradientsRef = useRef<{ [key: string]: CanvasGradient }>({});

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Set canvas size and scale for high DPI
    const dpr = window.devicePixelRatio || 1;
    canvas.width = CANVAS_SIZE * dpr;
    canvas.height = CANVAS_SIZE * dpr;
    canvas.style.width = `${CANVAS_SIZE}px`;
    canvas.style.height = `${CANVAS_SIZE}px`;
    ctx.scale(dpr, dpr);

    // Clear canvas
    ctx.clearRect(0, 0, CANVAS_SIZE, CANVAS_SIZE);

    // Create cached gradients if not exists
    if (!gradientsRef.current.green) {
      const greenGradient = ctx.createRadialGradient(0, 0, 0, 0, 0, CELL_SIZE * 0.6);
      greenGradient.addColorStop(0, '#2AFF6B'); // Bright green center
      greenGradient.addColorStop(0.7, '#1EC64E'); // Medium green
      greenGradient.addColorStop(1, 'rgba(42,255,107,0)'); // Transparent edge
      gradientsRef.current.green = greenGradient;
    }

    if (!gradientsRef.current.purple) {
      const purpleGradient = ctx.createRadialGradient(0, 0, 0, 0, 0, CELL_SIZE * 0.6);
      purpleGradient.addColorStop(0, '#B425FF'); // Bright purple center
      purpleGradient.addColorStop(0.7, '#782BF5'); // Medium purple
      purpleGradient.addColorStop(1, 'rgba(180,37,255,0)'); // Transparent edge
      gradientsRef.current.purple = purpleGradient;
    }

    // Draw background grid (subtle)
    ctx.globalAlpha = 0.1;
    ctx.strokeStyle = '#4B5563';
    ctx.lineWidth = 0.5;
    for (let i = 0; i <= GRID_SIZE; i++) {
      const pos = (i * CELL_SIZE);
      ctx.beginPath();
      ctx.moveTo(pos, 0);
      ctx.lineTo(pos, CANVAS_SIZE);
      ctx.stroke();
      
      ctx.beginPath();
      ctx.moveTo(0, pos);
      ctx.lineTo(CANVAS_SIZE, pos);
      ctx.stroke();
    }
    ctx.globalAlpha = 1;

    // Draw food
    gameState.food.forEach(food => {
      const x = food.x * CELL_SIZE + CELL_SIZE / 2;
      const y = food.y * CELL_SIZE + CELL_SIZE / 2;
      
      // Food glow
      ctx.shadowBlur = 8;
      ctx.shadowColor = '#FACC15';
      ctx.fillStyle = '#FACC15';
      ctx.beginPath();
      ctx.arc(x, y, CELL_SIZE * 0.3, 0, Math.PI * 2);
      ctx.fill();
      
      // Food highlight
      ctx.shadowBlur = 0;
      ctx.globalAlpha = 0.6;
      ctx.fillStyle = '#FFFFFF';
      ctx.beginPath();
      ctx.arc(x - CELL_SIZE * 0.1, y - CELL_SIZE * 0.1, CELL_SIZE * 0.15, 0, Math.PI * 2);
      ctx.fill();
      ctx.globalAlpha = 1;
    });

    // Helper function to check if two segments are connected (not a teleport wrap)
    const isConnected = (seg1: { x: number; y: number }, seg2: { x: number; y: number }): boolean => {
      const dx = Math.abs(seg1.x - seg2.x);
      const dy = Math.abs(seg1.y - seg2.y);
      // Connected if distance is 1 in either direction (not wrapped around edges)
      return (dx <= 1 && dy <= 1) && !(dx > 1 || dy > 1);
    };

    // Helper function to draw a smooth snake
    const drawSnake = (snake: Snake, isPlayer1: boolean) => {
      if (snake.body.length === 0) return;

      const color = isPlayer1 ? '#2AFF6B' : '#B425FF';
      const bodyThickness = CELL_SIZE * 0.6;
      
      // Draw body segments individually to avoid wraparound lines
      for (let i = 0; i < snake.body.length - 1; i++) {
        const current = snake.body[i];
        const next = snake.body[i + 1];
        
        // Only draw line if segments are actually connected (not wrapped around)
        if (isConnected(current, next)) {
          const x1 = current.x * CELL_SIZE + CELL_SIZE / 2;
          const y1 = current.y * CELL_SIZE + CELL_SIZE / 2;
          const x2 = next.x * CELL_SIZE + CELL_SIZE / 2;
          const y2 = next.y * CELL_SIZE + CELL_SIZE / 2;
          
          // Draw segment with glow
          ctx.shadowBlur = 12;
          ctx.shadowColor = color;
          ctx.lineWidth = bodyThickness;
          ctx.lineCap = 'round';
          ctx.strokeStyle = color;
          ctx.beginPath();
          ctx.moveTo(x1, y1);
          ctx.lineTo(x2, y2);
          ctx.stroke();
          
          // Draw highlight pass
          ctx.shadowBlur = 0;
          ctx.globalAlpha = 0.3;
          ctx.lineWidth = bodyThickness * 0.3;
          ctx.strokeStyle = '#FFFFFF';
          ctx.beginPath();
          ctx.moveTo(x1, y1);
          ctx.lineTo(x2, y2);
          ctx.stroke();
          ctx.globalAlpha = 1;
        }
      }
      
      // Draw each body segment as a circle for better continuity
      snake.body.forEach((segment, index) => {
        const x = segment.x * CELL_SIZE + CELL_SIZE / 2;
        const y = segment.y * CELL_SIZE + CELL_SIZE / 2;
        const radius = bodyThickness / 2;
        
        // Skip head (we'll draw it separately)
        if (index === 0) return;
        
        // Body segment glow
        ctx.shadowBlur = 8;
        ctx.shadowColor = color;
        ctx.fillStyle = color;
        ctx.beginPath();
        ctx.arc(x, y, radius, 0, Math.PI * 2);
        ctx.fill();
        
        // Body segment highlight
        ctx.shadowBlur = 0;
        ctx.globalAlpha = 0.3;
        ctx.fillStyle = '#FFFFFF';
        ctx.beginPath();
        ctx.arc(x - radius * 0.3, y - radius * 0.3, radius * 0.4, 0, Math.PI * 2);
        ctx.fill();
        ctx.globalAlpha = 1;
      });
      
      // Draw head as enhanced design
      if (snake.body.length > 0) {
        const head = snake.body[0];
        const headX = head.x * CELL_SIZE + CELL_SIZE / 2;
        const headY = head.y * CELL_SIZE + CELL_SIZE / 2;
        
        // Determine head direction for rotation
        let direction = { x: 0, y: -1 }; // Default up
        if (snake.body.length > 1) {
          const neck = snake.body[1];
          direction.x = head.x - neck.x;
          direction.y = head.y - neck.y;
          
          // Handle wraparound cases
          if (Math.abs(direction.x) > 1) direction.x = direction.x > 0 ? -1 : 1;
          if (Math.abs(direction.y) > 1) direction.y = direction.y > 0 ? -1 : 1;
        }
        
        // Enhanced head design - larger and more distinctive
        const headRadius = CELL_SIZE * 0.5;
        
        ctx.save();
        ctx.translate(headX, headY);
        
        // Rotate based on direction
        let angle = 0;
        if (direction.x === 1) angle = Math.PI / 2; // Right
        else if (direction.x === -1) angle = -Math.PI / 2; // Left
        else if (direction.y === 1) angle = Math.PI; // Down
        ctx.rotate(angle);
        
        // Head outer glow
        ctx.shadowBlur = 20;
        ctx.shadowColor = color;
        
        // Main head circle
        ctx.fillStyle = color;
        ctx.beginPath();
        ctx.arc(0, 0, headRadius, 0, Math.PI * 2);
        ctx.fill();
        
        // Inner gradient for depth
        ctx.shadowBlur = 0;
        const headGradient = ctx.createRadialGradient(0, 0, 0, 0, 0, headRadius);
        headGradient.addColorStop(0, color);
        headGradient.addColorStop(0.6, color);
        headGradient.addColorStop(1, isPlayer1 ? '#1EC64E' : '#782BF5');
        ctx.fillStyle = headGradient;
        ctx.beginPath();
        ctx.arc(0, 0, headRadius, 0, Math.PI * 2);
        ctx.fill();
        
        // Eye-like highlights for direction
        ctx.globalAlpha = 0.8;
        ctx.fillStyle = '#FFFFFF';
        ctx.beginPath();
        ctx.ellipse(headRadius * 0.2, -headRadius * 0.3, headRadius * 0.15, headRadius * 0.25, 0, 0, Math.PI * 2);
        ctx.fill();
        
        ctx.beginPath();
        ctx.ellipse(-headRadius * 0.2, -headRadius * 0.3, headRadius * 0.15, headRadius * 0.25, 0, 0, Math.PI * 2);
        ctx.fill();
        
        // Shine effect
        ctx.globalAlpha = 0.4;
        ctx.fillStyle = '#FFFFFF';
        ctx.beginPath();
        ctx.arc(-headRadius * 0.3, -headRadius * 0.3, headRadius * 0.3, 0, Math.PI * 2);
        ctx.fill();
        
        ctx.globalAlpha = 1;
        ctx.restore();
      }
    };

    // Draw snakes
    if (gameState.snakes.player1.alive) {
      drawSnake(gameState.snakes.player1, true);
    }
    if (gameState.snakes.player2.alive) {
      drawSnake(gameState.snakes.player2, false);
    }

  }, [gameState, playerId]);

  return (
    <div className="relative w-full max-w-2xl mx-auto">
      {/* Glow effect based on player */}
      <div 
        className={`absolute -inset-4 opacity-50 blur-xl rounded-lg ${
          playerId === 'player1' ? "bg-viper-green/20" : "bg-viper-purple/20"
        }`}
      />
      
      {/* Canvas container */}
      <div className="bg-gray-900/40 p-4 rounded-xl backdrop-blur-sm border border-gray-800/50 shadow-xl relative overflow-hidden z-10">
        {/* Background patterns and effects */}
        <div className="absolute inset-0 bg-gradient-to-br from-viper-green/5 to-viper-purple/5 z-0" />
        
        <canvas
          ref={canvasRef}
          className="w-full h-auto block"
          role="img"
          aria-label="Snake Game Board"
        />
      </div>
    </div>
  );
}