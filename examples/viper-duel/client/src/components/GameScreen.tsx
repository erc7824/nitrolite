import { SnakeCanvas } from "./SnakeCanvas";
import { GameOver } from "./GameOver";
import { RoomInfo } from "./RoomInfo";
import type { GameState, GameOver as GameOverType, Direction } from "../types";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Button } from "./ui/button";
import { useEffect, useState } from "react";

interface GameScreenProps {
    gameState: GameState;
    playerId: string | null;
    isRoomReady: boolean;
    isGameStarted: boolean;
    isHost: boolean;
    gameOver: GameOverType | null;
    opponentAddress: string;
    roomId: string;
    onDirectionChange: (direction: Direction) => void;
    onPlayAgain: () => void;
    onStartGame: () => void;
    awaitingHostStart?: boolean;
    isSigningInProgress?: boolean;
}

export function GameScreen({
    gameState,
    playerId,
    isRoomReady,
    isGameStarted,
    isHost,
    gameOver,
    opponentAddress,
    roomId,
    onDirectionChange,
    onPlayAgain,
    onStartGame,
    awaitingHostStart = false,
    isSigningInProgress = false,
}: GameScreenProps) {
    const [hasChannelId, setHasChannelId] = useState<boolean>(false);

    useEffect(() => {
        const channelId = localStorage.getItem("nitrolite_channel_id");
        setHasChannelId(!!channelId);
    }, []);

    // Debug information
    console.log("GameScreen state:", {
        roomId,
        isRoomReady,
        isGameStarted,
        isHost,
        playerId,
        opponentAddress,
        hasChannelId,
    });

    // Handle keyboard input for snake direction
    useEffect(() => {
        if (!isGameStarted || gameOver) return;

        const handleKeyPress = (event: KeyboardEvent) => {
            // Handle both WASD and Arrow keys
            let direction: Direction | null = null;
            
            // Check for WASD keys (case insensitive)
            const key = event.key.toLowerCase();
            if (key === 'w') direction = 'UP';
            else if (key === 's') direction = 'DOWN';
            else if (key === 'a') direction = 'LEFT';
            else if (key === 'd') direction = 'RIGHT';
            
            // Check for Arrow keys (case sensitive)
            else if (event.key === 'ArrowUp') direction = 'UP';
            else if (event.key === 'ArrowDown') direction = 'DOWN';
            else if (event.key === 'ArrowLeft') direction = 'LEFT';
            else if (event.key === 'ArrowRight') direction = 'RIGHT';
            
            if (direction) {
                event.preventDefault();
                console.log(`Key pressed: ${event.key}, Direction: ${direction}`); // Debug log
                onDirectionChange(direction);
            }
        };

        window.addEventListener('keydown', handleKeyPress);
        return () => window.removeEventListener('keydown', handleKeyPress);
    }, [isGameStarted, gameOver, onDirectionChange]);

    return (
        <div className="flex flex-col items-center w-full max-w-md mx-auto px-4 sm:px-6 py-4 sm:py-8">
            <Card className="w-full shadow-xl border-gray-800/50 bg-gray-900/80 backdrop-blur-sm">
                {/* Subtle background glow effect for the card */}
                <div className="absolute inset-0 bg-gradient-to-br from-viper-green/5 via-transparent to-viper-purple/5 rounded-lg z-0"></div>

                <CardHeader className="pb-2 relative z-10">
                    <RoomInfo roomId={roomId} />
                </CardHeader>

                <CardContent className="py-4 relative z-10">
                    {/* Waiting for players or game start */}
                    {(!isRoomReady || !isGameStarted) && (
                        <div className="my-2 text-center">
                            {!isRoomReady ? (
                                <Card className="bg-gray-800/50 border-gray-800/70 shadow-md transform transition-transform hover:scale-[1.01]">
                                    <CardHeader className="pb-1">
                                        <CardTitle className="text-lg sm:text-xl text-viper-green flex items-center justify-center gap-2">
                                            <div className="w-4 h-4 border-t-2 border-r-2 border-viper-green border-solid rounded-full animate-spin"></div>
                                            {hasChannelId ? "Waiting for another player to join..." : "Preparing your game..."}
                                        </CardTitle>
                                    </CardHeader>
                                    <CardContent>
                                        {roomId && hasChannelId && (
                                            <div className="text-sm bg-gray-900/70 text-gray-300 p-3 rounded-md border border-gray-800/80 mb-2 shadow-inner">
                                                <p className="font-medium text-gray-200 mb-1">Share this room ID:</p>
                                                <div className="bg-gray-900 p-2 rounded text-viper-green font-mono break-all select-all border border-gray-800/50">
                                                    {roomId}
                                                </div>
                                            </div>
                                        )}
                                        {hasChannelId ? (
                                            <p className="text-xs text-gray-500 mt-2">Players need this ID to join your game</p>
                                        ) : (
                                            <div className="flex flex-col items-center justify-center py-4">
                                                <div className="w-6 h-6 border-t-2 border-r-2 border-viper-green border-solid rounded-full animate-spin mb-3"></div>
                                                <p className="text-gray-300">Creating your channel...</p>
                                                <p className="text-sm text-gray-500 mt-1">Please wait while we set up your game</p>
                                            </div>
                                        )}
                                    </CardContent>
                                </Card>
                            ) : isHost ? (
                                <Card className="bg-gray-800/50 border-gray-800/70 shadow-md">
                                    <CardHeader className="pb-1">
                                        <CardTitle className="text-lg sm:text-xl text-viper-green flex items-center justify-center gap-2">
                                            <span className="w-2 h-2 bg-viper-green rounded-full animate-pulse"></span>
                                            Game ready! You are the host
                                        </CardTitle>
                                    </CardHeader>
                                    <CardContent>
                                        <Button
                                            onClick={onStartGame}
                                            variant="viperGreen"
                                            size="lg"
                                            className="w-full mt-2 animate-pulse"
                                            disabled={isSigningInProgress}
                                        >
                                            {isSigningInProgress ? (
                                                <div className="flex items-center gap-2">
                                                    <div className="w-4 h-4 border-t-2 border-r-2 border-white border-solid rounded-full animate-spin"></div>
                                                    Signing...
                                                </div>
                                            ) : awaitingHostStart ? (
                                                "Sign & Start Game"
                                            ) : (
                                                "Start Game"
                                            )}
                                        </Button>
                                    </CardContent>
                                </Card>
                            ) : (
                                <Card className="bg-gray-800/50 border-gray-800/70 shadow-md">
                                    <CardHeader className="pb-1">
                                        <CardTitle className="text-lg sm:text-xl text-viper-green flex items-center justify-center gap-2">
                                            <span className="w-2 h-2 bg-viper-green rounded-full animate-pulse"></span>
                                            Game ready!
                                        </CardTitle>
                                    </CardHeader>
                                    <CardContent>
                                        <p className="text-gray-300 mb-4">Waiting for host to start the game...</p>
                                        <div className="flex items-center justify-center">
                                            <div className="w-6 h-6 border-t-2 border-viper-green border-solid rounded-full animate-spin mr-2"></div>
                                            <span className="text-gray-400">Please wait</span>
                                        </div>
                                    </CardContent>
                                </Card>
                            )}
                        </div>
                    )}

                    {/* Game Status - only show when game is started */}
                    {isGameStarted && (
                        <div className="mt-2 mb-6">
                            <div className="flex justify-between items-center p-4 bg-gray-800/50 rounded-lg border border-gray-700/50">
                                <div className="flex flex-col items-center">
                                    <span className="text-sm text-gray-400">You ({playerId})</span>
                                    <span className="text-lg font-bold text-viper-green">
                                        Score: {playerId === 'player1' ? gameState.snakes.player1.score : gameState.snakes.player2.score}
                                    </span>
                                    <span className={`text-xs ${playerId === 'player1' ? (gameState.snakes.player1.alive ? 'text-green-400' : 'text-red-400') : (gameState.snakes.player2.alive ? 'text-green-400' : 'text-red-400')}`}>
                                        {playerId === 'player1' ? (gameState.snakes.player1.alive ? 'Alive' : 'Dead') : (gameState.snakes.player2.alive ? 'Alive' : 'Dead')}
                                    </span>
                                </div>
                                <div className="text-center">
                                    <span className="text-sm text-gray-400">Time</span>
                                    <div className="text-lg font-bold text-white">{gameState.gameTime}s</div>
                                    <div className="text-xs text-gray-500">Use WASD or arrows</div>
                                </div>
                                <div className="flex flex-col items-center">
                                    <span className="text-sm text-gray-400">Opponent</span>
                                    <span className="text-lg font-bold text-viper-purple">
                                        Score: {playerId === 'player1' ? gameState.snakes.player2.score : gameState.snakes.player1.score}
                                    </span>
                                    <span className={`text-xs ${playerId === 'player1' ? (gameState.snakes.player2.alive ? 'text-green-400' : 'text-red-400') : (gameState.snakes.player1.alive ? 'text-green-400' : 'text-red-400')}`}>
                                        {playerId === 'player1' ? (gameState.snakes.player2.alive ? 'Alive' : 'Dead') : (gameState.snakes.player1.alive ? 'Alive' : 'Dead')}
                                    </span>
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Game Board - only show when game is started */}
                    {isGameStarted && (
                        <div className="flex justify-center my-4 transition-all duration-300">
                            <SnakeCanvas
                                gameState={gameState}
                                playerId={playerId}
                            />
                        </div>
                    )}
                </CardContent>
            </Card>

            {/* Game Over Modal */}
            {gameOver && <GameOver gameOver={gameOver} playerId={playerId} onPlayAgain={onPlayAgain} />}
        </div>
    );
}
