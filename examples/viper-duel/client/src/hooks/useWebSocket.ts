import { useEffect, useRef, useState, useCallback } from "react";
import type { WebSocketMessages, JoinRoomPayload, DirectionPayload } from "../types";

// WebSocket hook for connecting to the game server
export function useWebSocket() {
    const [isConnected, setIsConnected] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const webSocketRef = useRef<WebSocket | null>(null);
    const [lastMessage, setLastMessage] = useState<WebSocketMessages | null>(null);

    // WebSocket server URL (use environment variable if available)
    const wsUrl = import.meta.env.VITE_WS_URL || "ws://localhost:8080";

    // Initialize WebSocket connection
    useEffect(() => {
        const webSocket = new WebSocket(wsUrl);

        webSocket.onopen = () => {
            setIsConnected(true);
            setError(null);
        };

        webSocket.onclose = () => {
            setIsConnected(false);
        };

        webSocket.onerror = () => {
            setError("Failed to connect to game server");
            setIsConnected(false);
        };

        webSocket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);

                setLastMessage(message as WebSocketMessages);
            } catch (err) {
                console.error("Error parsing WebSocket message", err);
            }
        };

        webSocketRef.current = webSocket;

        // Cleanup on unmount
        return () => {
            webSocket.close();
        };
    }, [wsUrl]);

    // Send a message to the server
    const sendMessage = useCallback(
        (message: object) => {
            if (webSocketRef.current && isConnected) {
                webSocketRef.current.send(JSON.stringify(message));
            } else {
                setError("Not connected to server");
            }
        },
        [isConnected]
    );

    // Join a room
    const joinRoom = useCallback(
        (payload: JoinRoomPayload) => {
            sendMessage({
                type: "joinRoom",
                payload,
            });
        },
        [sendMessage]
    );

    // Change snake direction
    const changeDirection = useCallback(
        (payload: DirectionPayload) => {
            sendMessage({
                type: "changeDirection",
                payload,
            });
        },
        [sendMessage]
    );

    // Start the game (host only)
    const startGame = useCallback(
        (roomId: string) => {
            sendMessage({
                type: "startGame",
                payload: { roomId },
            });
        },
        [sendMessage]
    );

    // Get available rooms
    const getAvailableRooms = useCallback(() => {
        sendMessage({
            type: "getAvailableRooms",
        });
    }, [sendMessage]);

    // Send app session signature
    const sendAppSessionSignature = useCallback(
        (roomId: string, signature: string) => {
            sendMessage({
                type: "appSession:signature",
                payload: { roomId, signature },
            });
        },
        [sendMessage]
    );

    // Send app session start game with signature
    const sendAppSessionStartGame = useCallback(
        (roomId: string, signature: string) => {
            sendMessage({
                type: "appSession:startGame",
                payload: { roomId, signature },
            });
        },
        [sendMessage]
    );

    return {
        isConnected,
        error,
        lastMessage,
        joinRoom,
        changeDirection,
        startGame,
        getAvailableRooms,
        sendAppSessionSignature,
        sendAppSessionStartGame,
    };
}
