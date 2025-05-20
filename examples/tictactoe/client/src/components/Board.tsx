import { Cell } from './Cell';
import type { GameState, PlayerSymbol } from '../types';
import { cn } from '../lib/utils';

interface BoardProps {
  gameState: GameState;
  playerSymbol: PlayerSymbol | null;
  isPlayerTurn: boolean;
  gameOver: boolean;
  onCellClick: (position: number) => void;
}

export function Board({ 
  gameState, 
  playerSymbol, 
  isPlayerTurn, 
  gameOver, 
  onCellClick 
}: BoardProps) {
  return (
    <div className="relative w-full max-w-md mx-auto">
      {/* Glow effect based on player */}
      <div 
        className={cn(
          "absolute -inset-4 opacity-50 blur-xl rounded-full",
          playerSymbol === 'X' ? "bg-cyan-900/20" : "bg-fuchsia-900/20"
        )}
      ></div>
      
      {/* Board container with enhanced styling */}
      <div 
        className="grid grid-cols-3 gap-2 sm:gap-3 bg-gray-900/40 p-3 sm:p-4 rounded-xl backdrop-blur-sm border border-gray-800/50 shadow-xl relative overflow-hidden z-10"
        role="grid"
        aria-label="Tic Tac Toe Board"
      >
        {/* Background patterns and effects */}
        <div className="absolute inset-0 bg-gradient-to-br from-cyan-900/5 to-fuchsia-900/5 z-0"></div>
        <div className="absolute inset-0 opacity-5">
          <div className="absolute inset-0 bg-grid-pattern"></div>
        </div>
        
        {/* Game cells */}
        {gameState.board.map((cell, index) => (
          <Cell
            key={index}
            value={cell}
            position={index}
            isPlayerTurn={isPlayerTurn}
            onClick={onCellClick}
            gameOver={gameOver}
          />
        ))}
      </div>
    </div>
  );
}