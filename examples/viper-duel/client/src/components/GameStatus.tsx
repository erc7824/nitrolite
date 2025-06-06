// Game status component (legacy from tic-tac-toe)
import { cn } from '../lib/utils';
import { Badge } from './ui/badge';
import { UserCheck, User, Clock } from 'lucide-react';

interface GameStatusProps {
  isPlayerTurn: boolean;
  playerSymbol: string | null;
  isRoomReady: boolean;
  isGameStarted: boolean;
  playerAddress: string;
  opponentAddress: string;
  formatShortAddress: (address: string) => string;
}

export function GameStatus({
  isPlayerTurn,
  playerSymbol,
  playerAddress,
  opponentAddress,
  formatShortAddress
}: GameStatusProps) {
  // This component is only shown when the game has started, so we don't need
  // to handle the waiting states here anymore - they're handled in GameScreen

  // Status message and styling
  const statusMessage = isPlayerTurn 
    ? "Your Move" 
    : "Opponent's Turn";
  
  const playerBadgeVariant = playerSymbol === 'X' ? 'cyan' : 'magenta';
  const opponentBadgeVariant = playerSymbol === 'X' ? 'magenta' : 'cyan';
  
  // Get player colors
  const playerColor = playerSymbol === 'X' ? 'text-cyan-400' : 'text-fuchsia-400';
  const opponentColor = playerSymbol === 'X' ? 'text-fuchsia-400' : 'text-cyan-400';
  
  // Glow classes
  const playerGlow = playerSymbol === 'X' ? 'glow-cyan' : 'glow-magenta';
  const opponentGlow = playerSymbol === 'X' ? 'glow-magenta' : 'glow-cyan';

  return (
    <div className="flex flex-col items-center space-y-4">
      {/* Turn indicator */}
      <Badge 
        variant={isPlayerTurn ? playerBadgeVariant : 'default'}
        className={cn(
          "px-4 py-1.5 text-base font-medium", 
          isPlayerTurn && "shadow-[0_0_10px_rgba(0,229,255,0.3)]",
          "relative overflow-hidden"
        )}
      >
        {/* Animated background for active player */}
        {isPlayerTurn && (
          <div className="absolute inset-0 opacity-30 overflow-hidden">
            <div className={cn(
              "absolute w-[200%] h-full -left-[50%] top-0",
              "bg-gradient-to-r from-transparent via-white to-transparent",
              "animate-[shimmer_2s_infinite]"
            )}></div>
          </div>
        )}
        
        <div className="flex items-center gap-1.5">
          {isPlayerTurn ? (
            <UserCheck className="h-4 w-4" />
          ) : (
            <Clock className="h-4 w-4 animate-pulse" />
          )}
          {statusMessage}
        </div>
      </Badge>
      
      {/* Player info */}
      <div className="grid grid-cols-3 w-full max-w-md gap-2 items-center justify-items-center text-sm bg-gray-900/40 p-2 rounded-md border border-gray-800/50">
        {/* Player (You) */}
        <div className={cn(
          "flex items-center justify-center rounded-md p-2 bg-gray-800/70 w-full border border-gray-700/30",
          isPlayerTurn && playerGlow,
          "relative overflow-hidden"
        )}>
          {/* Active player indicator */}
          {isPlayerTurn && (
            <div className="absolute left-0 top-0 bottom-0 w-1 bg-gradient-to-b from-transparent via-cyan-500 to-transparent animate-pulse"></div>
          )}
          
          <Badge variant={playerBadgeVariant} className="mr-1.5 flex items-center gap-1">
            <User className="h-3 w-3" />
            <span>{playerSymbol}</span>
          </Badge>
          <span className={cn(
            "text-gray-300",
            isPlayerTurn && playerColor
          )}>
            {formatShortAddress(playerAddress)}
          </span>
          <span className="ml-1 text-gray-500 text-xs">(you)</span>
        </div>
        
        {/* VS */}
        <div className="flex flex-col items-center justify-center h-full">
          <span className="text-gray-500 font-medium">VS</span>
          <span className={cn(
            "text-xs px-2 py-0.5 rounded bg-gray-800/50 mt-1",
            isPlayerTurn ? "text-gray-500" : "text-gray-400"
          )}>
            {isPlayerTurn ? "your turn" : "waiting"}
          </span>
        </div>
        
        {/* Opponent */}
        <div className={cn(
          "flex items-center justify-center rounded-md p-2 bg-gray-800/70 w-full border border-gray-700/30",
          !isPlayerTurn && opponentGlow,
          "relative overflow-hidden"
        )}>
          {/* Active player indicator */}
          {!isPlayerTurn && (
            <div className="absolute left-0 top-0 bottom-0 w-1 bg-gradient-to-b from-transparent via-fuchsia-500 to-transparent animate-pulse"></div>
          )}
          
          <Badge variant={opponentBadgeVariant} className="mr-1.5 flex items-center gap-1">
            <User className="h-3 w-3" />
            <span>{playerSymbol === 'X' ? 'O' : 'X'}</span>
          </Badge>
          <span className={cn(
            "text-gray-300",
            !isPlayerTurn && opponentColor
          )}>
            {formatShortAddress(opponentAddress)}
          </span>
        </div>
      </div>
    </div>
  );
}