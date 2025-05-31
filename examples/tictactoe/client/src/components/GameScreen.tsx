import { Board } from "./Board";
import { GameStatus } from "./GameStatus";
import { GameOver } from "./GameOver";
import { RoomInfo } from "./RoomInfo";
import type { GameState, GameOver as GameOverType, PlayerSymbol } from "../types";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Button } from "./ui/button";
import { useEffect, useState } from "react";

interface GameScreenProps {
    gameState: GameState;
    playerSymbol: PlayerSymbol | null;
    isPlayerTurn: boolean;
    isRoomReady: boolean;
    isGameStarted: boolean;
    isHost: boolean;
    gameOver: GameOverType | null;
    playerAddress: string;
    opponentAddress: string;
    roomId: string;
    formatShortAddress: (address: string) => string;
    onCellClick: (position: number) => void;
    onPlayAgain: () => void;
    onStartGame: () => void;
    awaitingHostStart?: boolean;
    isSigningInProgress?: boolean;
}

export function GameScreen({
    gameState,
    playerSymbol,
    isPlayerTurn,
    isRoomReady,
    isGameStarted,
    isHost,
    gameOver,
    playerAddress,
    opponentAddress,
    roomId,
    formatShortAddress,
    onCellClick,
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
        playerSymbol,
        opponentAddress,
        hasChannelId,
    });

    return (
        <div className="flex flex-col items-center w-full max-w-md mx-auto px-4 sm:px-6 py-4 sm:py-8">
            <Card className="w-full shadow-xl border-gray-800/50 bg-gray-900/80 backdrop-blur-sm">
                {/* Subtle background glow effect for the card */}
                <div className="absolute inset-0 bg-gradient-to-br from-cyan-900/5 via-transparent to-fuchsia-900/5 rounded-lg z-0"></div>

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
                                        <CardTitle className="text-lg sm:text-xl text-cyan-400 flex items-center justify-center gap-2">
                                            <div className="w-4 h-4 border-t-2 border-r-2 border-cyan-500 border-solid rounded-full animate-spin"></div>
                                            {hasChannelId ? "Waiting for another player to join..." : "Preparing your game..."}
                                        </CardTitle>
                                    </CardHeader>
                                    <CardContent>
                                        {roomId && hasChannelId && (
                                            <div className="text-sm bg-gray-900/70 text-gray-300 p-3 rounded-md border border-gray-800/80 mb-2 shadow-inner">
                                                <p className="font-medium text-gray-200 mb-1">Share this room ID:</p>
                                                <div className="bg-gray-900 p-2 rounded text-cyan-300 font-mono break-all select-all border border-gray-800/50">
                                                    {roomId}
                                                </div>
                                            </div>
                                        )}
                                        {hasChannelId ? (
                                            <p className="text-xs text-gray-500 mt-2">Players need this ID to join your game</p>
                                        ) : (
                                            <div className="flex flex-col items-center justify-center py-4">
                                                <div className="w-6 h-6 border-t-2 border-r-2 border-cyan-500 border-solid rounded-full animate-spin mb-3"></div>
                                                <p className="text-gray-300">Creating your channel...</p>
                                                <p className="text-sm text-gray-500 mt-1">Please wait while we set up your game</p>
                                            </div>
                                        )}
                                    </CardContent>
                                </Card>
                            ) : isHost ? (
                                <Card className="bg-gray-800/50 border-gray-800/70 shadow-md">
                                    <CardHeader className="pb-1">
                                        <CardTitle className="text-lg sm:text-xl text-cyan-400 flex items-center justify-center gap-2">
                                            <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
                                            Game ready! You are the host
                                        </CardTitle>
                                    </CardHeader>
                                    <CardContent>
                                        <Button
                                            onClick={onStartGame}
                                            variant="cyan"
                                            size="xl"
                                            className="w-full mt-2 shadow-lg shadow-cyan-900/20 animate-pulse"
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
                                        <CardTitle className="text-lg sm:text-xl text-cyan-400 flex items-center justify-center gap-2">
                                            <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
                                            Game ready!
                                        </CardTitle>
                                    </CardHeader>
                                    <CardContent>
                                        <p className="text-gray-300 mb-4">Waiting for host to start the game...</p>
                                        <div className="flex items-center justify-center">
                                            <div className="w-6 h-6 border-t-2 border-cyan-500 border-solid rounded-full animate-spin mr-2"></div>
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
                            <GameStatus
                                isPlayerTurn={isPlayerTurn}
                                playerSymbol={playerSymbol}
                                isRoomReady={isRoomReady}
                                isGameStarted={isGameStarted}
                                playerAddress={playerAddress}
                                opponentAddress={opponentAddress}
                                formatShortAddress={formatShortAddress}
                            />
                        </div>
                    )}

                    {/* Game Board - only show when game is started */}
                    {isGameStarted && (
                        <div className="flex justify-center my-4 transition-all duration-300">
                            <Board
                                gameState={gameState}
                                playerSymbol={playerSymbol}
                                isPlayerTurn={isPlayerTurn}
                                gameOver={gameOver !== null}
                                onCellClick={onCellClick}
                            />
                        </div>
                    )}
                </CardContent>
            </Card>

            {/* Game Over Modal */}
            {gameOver && <GameOver gameOver={gameOver} playerSymbol={playerSymbol} onPlayAgain={onPlayAgain} />}
        </div>
    );
}
