import { useState, useEffect } from 'react';
// Cell component for individual grid cells (legacy from tic-tac-toe)
import { cn } from '../lib/utils';
import { CellAuraEffect } from './CellAuraEffect';
import { CellMarkEffect } from './CellMarkEffect';
import { useSoundEffects } from '../hooks/useSoundEffects';

interface CellProps {
  value: string | null;
  position: number;
  isPlayerTurn: boolean;
  onClick: (position: number) => void;
  gameOver: boolean;
}

export function Cell({ value, position, isPlayerTurn, onClick, gameOver }: CellProps) {
  // Determine if cell is clickable
  const isClickable = !value && isPlayerTurn && !gameOver;
  
  // Track when a symbol is placed for animation purposes
  const [activated, setActivated] = useState(false);
  const [prevValue, setPrevValue] = useState<string | null>(null);
  
  // Sound effects
  const { playSound } = useSoundEffects();
  
  // Detect when a new symbol is placed
  useEffect(() => {
    if (value && value !== prevValue) {
      setActivated(true);
      setPrevValue(value);
      
      // Play sound effect based on the symbol
      playSound(value === 'X' ? 'mark-x' : 'mark-o', 0.4);
      
      // Dispatch custom event for background particles to react
      const rect = document.getElementById(`cell-${position}`)?.getBoundingClientRect();
      if (rect) {
        const cellMarkedEvent = new CustomEvent('cellMarked', {
          detail: {
            x: rect.left + rect.width / 2,
            y: rect.top + rect.height / 2,
            isX: value === 'X'
          }
        });
        window.dispatchEvent(cellMarkedEvent);
      }
    }
  }, [value, prevValue, position]);

  return (
    <div 
      id={`cell-${position}`}
      className={cn(
        // Base styles for all cells
        "flex items-center justify-center", 
        "h-16 w-16 sm:h-20 sm:w-20 md:h-24 md:w-24", 
        "border rounded-md sm:rounded-lg text-4xl sm:text-5xl font-bold",
        "transition-all duration-300 relative z-10",
        
        // State-specific styles
        value ? "bg-gray-900/70 backdrop-blur" : "bg-gray-900/30",
        value === 'X' && "text-cyan-400 cell-x border-cyan-900/50",
        value === 'O' && "text-fuchsia-400 cell-o border-fuchsia-900/50",
        !value && "border-gray-700/50",
        
        // Interaction states
        isClickable && "hover:bg-gray-800/40 hover:border-cyan-500/30 hover:scale-[1.02] cursor-pointer",
        !isClickable && !value && "opacity-70"
      )}
      onClick={() => isClickable ? onClick(position) : undefined}
      role="button"
      aria-label={value ? `Cell ${position + 1}, ${value}` : `Cell ${position + 1}, empty`}
      aria-disabled={!isClickable}
      tabIndex={isClickable ? 0 : -1}
    >
      {/* Enhanced aura effects - only appears when cell has a value */}
      {value && (
        <>
          {/* Inner glow */}
          <div className={cn(
            "absolute inset-0 opacity-30 rounded-lg", 
            value === 'X' ? "bg-cyan-500/10" : "bg-fuchsia-500/10"
          )}></div>
          
          {/* Outer glow */}
          <div className={cn(
            "absolute -inset-1 opacity-10 blur-md rounded-xl z-0",
            value === 'X' ? "bg-cyan-500" : "bg-fuchsia-500"
          )}></div>
          
          {/* Static particle effect (background) */}
          <div className="absolute inset-0 opacity-20 overflow-hidden rounded-lg">
            <div className={cn(
              "absolute w-[200%] h-[200%] top-[-50%] left-[-50%]",
              "bg-[radial-gradient(circle,_white_1px,_transparent_1px)]",
              "bg-[length:12px_12px]",
              "animate-sparkle"
            )}></div>
          </div>
          
          {/* Animated pulsing effect */}
          <div className={cn(
            "absolute inset-0 rounded-lg z-0 animate-[pulse_2s_ease-in-out_infinite]",
            value === 'X' ? "bg-cyan-500/5" : "bg-fuchsia-500/5"
          )}></div>
          
          {/* Dynamic canvas-based particle effect */}
          <CellAuraEffect value={value} activated={activated} />
          
          {/* Dramatic mark effect when cell is activated */}
          <CellMarkEffect value={value} activated={activated} />
        </>
      )}
      
      {/* Hover effect for empty cells */}
      {!value && isPlayerTurn && !gameOver && (
        <div className="absolute inset-0 bg-cyan-500/0 hover:bg-cyan-500/5 rounded-lg transition-colors duration-200"></div>
      )}
      
      {/* The X or O with text shadow */}
      <span className={cn(
        "relative z-10",
        value === 'X' && "text-shadow-cyan",
        value === 'O' && "text-shadow-magenta"
      )}>
        {value}
      </span>
    </div>
  );
}