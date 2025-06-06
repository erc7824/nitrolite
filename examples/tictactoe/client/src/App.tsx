import { useState, useEffect, useCallback } from "react";
import { useWebSocket } from "./hooks/useWebSocket";
import { useGameState } from "./hooks/useGameState";
import { GameScreen } from "./components/GameScreen";
import { ErrorModal } from "./components/ErrorModal";
import { BackgroundAnimation } from "./components/BackgroundAnimation";
import { GameLobbyIntegrated } from "./components/GameLobbyIntegrated";
import type { JoinRoomPayload, AvailableRoom, AvailableRoomsMessage, PlayerSymbol } from "./types";
import "./App.css";
import { useWebSocketNitrolite } from "./hooks/useWebSocketNitrolite";
import { useNitroliteIntegration } from "./hooks/useNitroliteIntegration";
import { useNitrolite } from "./context/NitroliteClientWrapper";
import { WalletButton } from "@rainbow-me/rainbowkit";
import "@rainbow-me/rainbowkit/styles.css";

function App() {
    // Player's Ethereum address - now managed by useMetaMask hook in Lobby
    const [eoaAddress, setEoaAddress] = useState<string>("");

    // Game view state
    const [gameView, setGameView] = useState<"lobby" | "game">("lobby");

    // WebSocket connection
    const {
        error: wsError,
        lastMessage,
        joinRoom,
        makeMove,
        startGame,
        getAvailableRooms,
        sendAppSessionSignature,
        sendAppSessionStartGame,
    } = useWebSocket();
    useWebSocketNitrolite();
    const { client, loading: nitroliteLoading, error: nitroliteError } = useNitrolite();

    // Initialize the Nitrolite integration
    const { initializeNitroliteClient } = useNitroliteIntegration();

    // When the Nitrolite client is available, initialize it
    useEffect(() => {
        if (client && !nitroliteLoading && !nitroliteError) {
            console.log("Initializing Nitrolite client in App component");
            initializeNitroliteClient(client);
        } else if (nitroliteError) {
            console.error("Nitrolite client error:", nitroliteError);
        }
    }, [client, nitroliteLoading, nitroliteError, initializeNitroliteClient]);

    // Removed this reference as we're now using destructuring above

    // Available rooms state
    const [availableRooms, setAvailableRooms] = useState<AvailableRoom[]>([]);
    const [onlineUsers, setOnlineUsers] = useState<number>(1);

    // Game state
    const {
        gameState,
        gameOver,
        roomId,
        errorMessage,
        isRoomReady,
        isGameStarted,
        isHost,
        playerSymbol,
        isPlayerTurn,
        formatShortAddress,
        getOpponentAddress,
        resetGame,
        awaitingHostStart,
        signAndStartGame,
        isSigningInProgress,
        signatureError,
    } = useGameState(lastMessage, eoaAddress, sendAppSessionSignature, sendAppSessionStartGame);

    // Handle errors
    const [showError, setShowError] = useState<boolean>(false);
    const [errorDisplay, setErrorDisplay] = useState<string | null>(null);

    useEffect(() => {
        // Combine all possible error sources
        const combinedError = wsError || errorMessage || nitroliteError || signatureError;

        if (combinedError) {
            console.log("Error detected:", combinedError);

            // Don't show error modal for MetaMask connection message
            if (combinedError === "MetaMask not connected. Please connect your wallet.") {
                setShowError(false);
                setErrorDisplay(null);
            } else {
                setShowError(true);
                setErrorDisplay(combinedError);
            }
        } else {
            setShowError(false);
            setErrorDisplay(null);
        }
    }, [wsError, errorMessage, nitroliteError, signatureError]);

    // Process available rooms from websocket messages
    useEffect(() => {
        if (lastMessage && lastMessage.type === "room:available") {
            const roomsMessage = lastMessage as AvailableRoomsMessage;
            setAvailableRooms(roomsMessage.rooms);
        }

        if (lastMessage && lastMessage.type === "onlineUsers") {
            setOnlineUsers(lastMessage.count);
        }
    }, [lastMessage]);

    // Handle fetching available rooms
    const handleGetAvailableRooms = useCallback(() => {
        getAvailableRooms();
    }, [getAvailableRooms]);

    // Handle joining a room
    const handleJoinRoom = (payload: JoinRoomPayload) => {
        setEoaAddress(payload.eoa);

        // If creating a new room, mark as host
        if (payload.roomId === undefined) {
            console.log("Creating new room as host, payload:", payload);
        } else {
            console.log("Joining existing room:", payload.roomId, "payload:", payload);
        }

        // Join room via WebSocket - pass the payload directly
        console.log("Sending WebSocket joinRoom with payload:", {
            roomId: payload.roomId,
            eoa: payload.eoa,
        });

        joinRoom({
            roomId: payload.roomId,
            eoa: payload.eoa,
        });

        // Switch to game view
        setGameView("game");
    };

    // Handle cell click
    const handleCellClick = (position: number) => {
        if (!roomId || !isPlayerTurn || gameOver) return;

        makeMove({
            roomId,
            pos: position,
        });
    };

    // Handle starting the game (host only)
    const handleStartGame = () => {
        if (!roomId || !isHost) {
            console.error("Cannot start game: not host or no room ID");
            return;
        }

        // If we're awaiting host signature for app session, sign and start
        if (awaitingHostStart) {
            console.log("Signing app session and starting game for room:", roomId);
            signAndStartGame();
        } else {
            console.log("Starting game as host for room:", roomId);
            startGame(roomId);
        }
    };

    // Handle play again
    const handlePlayAgain = () => {
        // For now, just reload the page
        window.location.reload();

        // TODO: Implement proper reset logic when @erc7824/nitrolite is integrated
    };

    // Handle error close
    const handleErrorClose = () => {
        setShowError(false);
        resetGame();
        setGameView("lobby");
    };

    return (
        <div className="min-h-screen w-full flex flex-col justify-center items-center p-4 relative overflow-hidden">
            {/* Background particles */}
            <BackgroundAnimation />

            {/* Background grid pattern */}
            <div className="fixed inset-0 bg-grid-pattern opacity-10 z-0"></div>

            {/* Decorative glow effects */}
            <div className="fixed top-[-50%] left-[-20%] w-[140%] h-[140%] bg-gradient-radial from-cyan-900/5 to-transparent opacity-30 blur-3xl z-0"></div>
            <div className="fixed bottom-[-50%] right-[-20%] w-[140%] h-[140%] bg-gradient-radial from-fuchsia-900/5 to-transparent opacity-30 blur-3xl z-0"></div>

            {/* Only show the app header in game view */}
            {gameView === "game" && (
                <div className="absolute top-6 left-0 right-0 flex justify-center pointer-events-none">
                    <div className="text-center">
                        <h1 className="text-3xl sm:text-4xl font-bold">
                            <span className="text-glow-cyan">Nitro</span>
                            <span className="text-glow-magenta ml-1">Aura</span>
                        </h1>
                        <p className="text-gray-400 text-sm mt-1">Light speed, neon bleed.</p>
                    </div>
                </div>
            )}

            {/* Main Content */}
            <div className="max-w-xl w-full relative z-10">
                <WalletButton wallet="cmbevh1e30022l80nhp974z8m" />

                {gameView === "lobby" ? (
                    <GameLobbyIntegrated
                        onJoinRoom={handleJoinRoom}
                        availableRooms={availableRooms}
                        onGetAvailableRooms={handleGetAvailableRooms}
                        onlineUsers={onlineUsers}
                    />
                ) : (
                    <GameScreen
                        gameState={gameState}
                        playerSymbol={playerSymbol as PlayerSymbol}
                        isPlayerTurn={isPlayerTurn}
                        isRoomReady={isRoomReady}
                        isGameStarted={isGameStarted}
                        isHost={isHost}
                        gameOver={gameOver}
                        playerAddress={eoaAddress}
                        opponentAddress={getOpponentAddress()}
                        roomId={roomId}
                        formatShortAddress={formatShortAddress}
                        onCellClick={handleCellClick}
                        onPlayAgain={handlePlayAgain}
                        onStartGame={handleStartGame}
                        awaitingHostStart={awaitingHostStart}
                        isSigningInProgress={isSigningInProgress}
                    />
                )}

                {showError && <ErrorModal message={errorDisplay || "An unknown error occurred"} onClose={handleErrorClose} />}
            </div>
        </div>
    );
}

export default App;
