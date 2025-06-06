import type { GameState, Snake } from '../types';
import { cn } from '../lib/utils';

interface BoardProps {
  gameState: GameState;
  playerId: string | null; // 'player1' or 'player2'
}

const GRID_SIZE = 20;
const CELL_SIZE = 'w-4 h-4'; // Tailwind classes for cell size

export function Board({ 
  gameState, 
  playerId
}: BoardProps) {
  // Helper function to determine what's at a position
  const getCellContent = (x: number, y: number): 'empty' | 'player1' | 'player2' | 'food' => {
    // Check for food
    if (gameState.food.some(food => food.x === x && food.y === y)) {
      return 'food';
    }
    
    // Check for player1 snake
    if (gameState.snakes.player1.body.some(segment => segment.x === x && segment.y === y)) {
      return 'player1';
    }
    
    // Check for player2 snake
    if (gameState.snakes.player2.body.some(segment => segment.x === x && segment.y === y)) {
      return 'player2';
    }
    
    return 'empty';
  };
  
  // Helper function to check if position is snake head
  const isSnakeHead = (x: number, y: number, snake: Snake): boolean => {
    return snake.body.length > 0 && snake.body[0].x === x && snake.body[0].y === y;
  };

  return (
    <div className="relative w-full max-w-2xl mx-auto">
      {/* Glow effect based on player */}
      <div 
        className={cn(
          "absolute -inset-4 opacity-50 blur-xl rounded-lg",
          playerId === 'player1' ? "bg-viper-green/20" : "bg-viper-purple/20"
        )}
      ></div>
      
      {/* Board container */}
      <div 
        className="grid gap-0.5 bg-gray-900/40 p-4 rounded-xl backdrop-blur-sm border border-gray-800/50 shadow-xl relative overflow-hidden z-10"
        style={{ gridTemplateColumns: `repeat(${GRID_SIZE}, minmax(0, 1fr))` }}
        role="grid"
        aria-label="Snake Game Board"
      >
        {/* Background patterns and effects */}
        <div className="absolute inset-0 bg-gradient-to-br from-viper-green/5 to-viper-purple/5 z-0"></div>
        
        {/* Game cells */}
        {Array.from({ length: GRID_SIZE * GRID_SIZE }, (_, index) => {
          const x = index % GRID_SIZE;
          const y = Math.floor(index / GRID_SIZE);
          const content = getCellContent(x, y);
          
          return (
            <div
              key={`${x}-${y}`}
              className={cn(
                CELL_SIZE,
                "border border-gray-700/30 relative",
                {
                  'bg-viper-yellow': content === 'food',
                  'bg-viper-green': content === 'player1' && isSnakeHead(x, y, gameState.snakes.player1),
                  'bg-viper-green-dark': content === 'player1' && !isSnakeHead(x, y, gameState.snakes.player1),
                  'bg-viper-purple': content === 'player2' && isSnakeHead(x, y, gameState.snakes.player2),
                  'bg-viper-purple-dark': content === 'player2' && !isSnakeHead(x, y, gameState.snakes.player2),
                  'bg-viper-charcoal-light/20': content === 'empty'
                }
              )}
            >
              {/* Add subtle glow for snake heads */}
              {((content === 'player1' && isSnakeHead(x, y, gameState.snakes.player1)) ||
                (content === 'player2' && isSnakeHead(x, y, gameState.snakes.player2))) && (
                <div className="absolute inset-0 animate-pulse bg-white/20 rounded-sm"></div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}