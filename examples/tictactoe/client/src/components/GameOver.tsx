import { useEffect } from 'react';
import type { GameOver as GameOverType, PlayerSymbol } from '../types';
import { Button } from './ui/button';
import { cn } from '../lib/utils';
import { Trophy, Medal, CircleSlash } from 'lucide-react';
import { useSoundEffects } from '../hooks/useSoundEffects';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from './ui/dialog';

interface GameOverProps {
  gameOver: GameOverType;
  playerSymbol: PlayerSymbol | null;
  onPlayAgain: () => void;
}

export function GameOver({ gameOver, playerSymbol, onPlayAgain }: GameOverProps) {
  const { winner } = gameOver;
  const { playSound } = useSoundEffects();
  
  // Play appropriate sound effect when component mounts
  useEffect(() => {
    if (winner === playerSymbol) {
      playSound('win', 0.5);
    } else if (winner) {
      playSound('game-over', 0.5);
    } else {
      playSound('draw', 0.5);
    }
  }, [winner, playerSymbol, playSound]);
  
  // Determine message and styling based on game outcome
  const getMessage = () => {
    if (!winner) {
      return "It's a Draw!";
    }
    
    return winner === playerSymbol ? "You Won!" : "You Lost!";
  };
  
  // Get appropriate icon for result
  const ResultIcon = !winner ? CircleSlash : (winner === playerSymbol ? Trophy : Medal);
  
  // Styles based on winner
  const iconColor = winner === 'X' ? 'text-cyan-400' : winner === 'O' ? 'text-fuchsia-400' : 'text-gray-400';
  const bgGradient = winner === 'X' 
    ? 'from-cyan-900/30 to-gray-900/90' 
    : winner === 'O' 
      ? 'from-fuchsia-900/30 to-gray-900/90' 
      : 'from-gray-800/30 to-gray-900/90';
  
  return (
    <Dialog open={true} modal={true}>
      <DialogContent 
        className="max-w-md w-full border-gray-700 shadow-2xl relative overflow-hidden"
        style={{
          boxShadow: '0 0 30px rgba(0, 0, 0, 0.4)',
          maxWidth: '28rem'
        }}>
        {/* Background gradient */}
        <div className={cn(
          "absolute inset-0 bg-gradient-to-b",
          bgGradient,
          "z-0"
        )}></div>
        
        {/* Particle effects */}
        {winner === playerSymbol && (
          <div className="absolute inset-0 overflow-hidden">
            <div className="absolute w-full h-[200%] top-[-50%] left-0 bg-[radial-gradient(circle,_white_1px,_transparent_1px)] bg-[length:20px_20px] opacity-[0.03] animate-sparkle"></div>
          </div>
        )}
        
        {/* Content */}
        <div className="relative z-10">
          <DialogHeader className="text-center pb-2">
            <div className="mx-auto bg-gray-800/50 p-3 rounded-full mb-2">
              <ResultIcon className={cn("h-12 w-12", iconColor)} />
            </div>
            <DialogTitle className={cn(
              'text-4xl font-bold',
              winner === 'X' && 'text-cyan-400',
              winner === 'O' && 'text-fuchsia-400',
              !winner && 'text-gray-300'
            )}>
              {getMessage()}
            </DialogTitle>
          </DialogHeader>
          
          <div className="text-center">
            <p className="text-gray-300 mb-4">
              {winner ? `${winner} has won the game!` : "No more moves available."}
            </p>
          </div>
          
          <DialogFooter className="flex justify-center pb-2 mt-4">
            <Button
              onClick={onPlayAgain}
              type="button"
              variant={winner === 'X' ? 'cyan' : winner === 'O' ? 'magenta' : 'default'}
              size="lg"
              className="px-8"
            >
              Play Again
            </Button>
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>
  );
}