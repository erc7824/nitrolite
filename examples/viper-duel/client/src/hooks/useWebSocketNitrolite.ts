import { useWebSocketContext } from "../context/WebSocketContext";

/**
 * Custom hook to manage WebSocket connection and operations using WebSocketContext.
 * This hook provides a simple interface to the WebSocket functionality
 * without exposing the implementation details.
 */
export function useWebSocketNitrolite() {
    const context = useWebSocketContext();

    return {
        // Connection status
        status: context.status,
        isConnected: context.isConnected,
        
        // Key management
        keyPair: context.keyPair,
        hasKeys: context.hasKeys,
        generateKeys: context.generateKeys,
        clearKeys: context.clearKeys,
        
        // Channel management
        wsChannel: context.wsChannel,
        currentNitroliteChannel: context.currentNitroliteChannel,
        setNitroliteChannel: context.setNitroliteChannel,
        
        // Connection management
        connect: context.connect,
        disconnect: context.disconnect,
        
        // Message handling
        sendPing: context.sendPing,
        sendRequest: context.sendRequest,
    };
}
