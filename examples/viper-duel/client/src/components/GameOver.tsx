import { useEffect } from 'react';
import type { GameOver as GameOverType } from '../types';
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
  playerId: string | null;
  onPlayAgain: () => void;
}

export function GameOver({ gameOver, playerId, onPlayAgain }: GameOverProps) {
  const { winner, finalScores, gameTime } = gameOver;
  const { playSound } = useSoundEffects();
  
  // Play appropriate sound effect when component mounts
  useEffect(() => {
    if (winner === playerId) {
      playSound('win', 0.5);
    } else if (winner) {
      playSound('game-over', 0.5);
    } else {
      playSound('draw', 0.5);
    }
  }, [winner, playerId, playSound]);
  
  // Determine message and styling based on game outcome
  const getMessage = () => {
    if (!winner) {
      return "It's a Tie!";
    }
    
    return winner === playerId ? "You Won!" : "You Lost!";
  };
  
  // Get appropriate icon for result
  const ResultIcon = !winner ? CircleSlash : (winner === playerId ? Trophy : Medal);
  
  // Styles based on winner
  const iconColor = winner === 'player1' ? 'text-viper-green' : winner === 'player2' ? 'text-viper-purple' : 'text-viper-grey';
  const bgGradient = winner === 'player1' 
    ? 'from-viper-green/30 to-viper-charcoal/90' 
    : winner === 'player2' 
      ? 'from-viper-purple/30 to-viper-charcoal/90' 
      : 'from-viper-charcoal-light/30 to-viper-charcoal/90';
  
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
        {winner === playerId && (
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
              'text-4xl font-bold font-pixel',
              winner === 'player1' && 'text-viper-green',
              winner === 'player2' && 'text-viper-purple',
              !winner && 'text-viper-grey'
            )}>
              {getMessage()}
            </DialogTitle>
          </DialogHeader>
          
          <div className="text-center">
            <div className="bg-gray-800/50 rounded-lg p-4 mb-4">
              <h3 className="text-lg font-semibold text-white mb-2">Final Scores</h3>
              <div className="flex justify-between items-center mb-2">
                <span className="text-viper-green font-mono">Player 1:</span>
                <span className="text-white font-bold font-mono">{finalScores.player1}</span>
              </div>
              <div className="flex justify-between items-center mb-2">
                <span className="text-viper-purple font-mono">Player 2:</span>
                <span className="text-white font-bold font-mono">{finalScores.player2}</span>
              </div>
              <div className="text-sm text-viper-grey mt-2 font-mono">
                Game Time: {gameTime} seconds
              </div>
            </div>
            <p className="text-gray-300 mb-4">
              {winner ? `${winner === 'player1' ? 'Player 1' : 'Player 2'} wins!` : "Both snakes died - it's a tie!"}
            </p>
          </div>
          
          <DialogFooter className="flex justify-center pb-2 mt-4">
            <Button
              onClick={onPlayAgain}
              type="button"
              className={cn(
                "px-8 py-3 font-bold rounded-lg transition-colors",
                winner === 'player1' && "bg-viper-green hover:bg-viper-green-dark text-viper-charcoal",
                winner === 'player2' && "bg-viper-purple hover:bg-viper-purple-dark text-white",
                !winner && "bg-viper-grey hover:bg-viper-grey-dark text-viper-charcoal"
              )}
            >
              Play Again
            </Button>
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>
  );
}