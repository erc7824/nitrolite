import { useCallback } from 'react';
import { WebSocketClient } from '@/websocket';
import { useMessageService } from '../useMessageService';
import { usePingPongBenchmark } from './usePingPongBenchmark';

// Custom hook to handle message processing
export function useMessageHandler(
    clientRef: React.RefObject<WebSocketClient>,
    pingPongBenchmark: ReturnType<typeof usePingPongBenchmark>,
    messageService: ReturnType<typeof useMessageService>
) {
    const { addSystemMessage, addErrorMessage } = messageService;
    const { benchmarkInProgress, handlePongResponse, handleGuestPingMessage, handleGuestPongMessage } = pingPongBenchmark;
    
    const handleMessage = useCallback((message: any) => {
        console.log('Received message:', message);
        const timestamp = new Date().toISOString();
        
        // Check if it's a pong message from server
        if (message && typeof message === 'object' && 'type' in message && message.type === 'pong') {
            handlePongResponse(timestamp);
            return;
        }
        
        // Check if it's a channel message
        if (
            message &&
            typeof message === 'object' &&
            'type' in message &&
            message.type === 'channel_message' &&
            'data' in message &&
            message.data &&
            typeof message.data === 'object' &&
            'content' in message.data &&
            'sender' in message.data
        ) {
            handleChannelMessage(message, timestamp);
        } else {
            // Non-channel messages
            addSystemMessage(`[${timestamp}] Received message of type: ${message.type || 'unknown'}`);
        }
    }, [handlePongResponse, addSystemMessage]);
    
    const handleChannelMessage = useCallback((message: any, timestamp: string) => {
        const content = String(message.data.content).toLowerCase().trim();
        const sender = String(message.data.sender);
        const userKey = clientRef.current?.getShortenedPublicKey();
        const isFromCurrentUser = sender === userKey;
        
        if (content === 'ping') {
            if (!isFromCurrentUser) {
                handleGuestPingMessage(sender, timestamp);
            }
        } else if (content === 'pong') {
            if (!isFromCurrentUser) {
                handleGuestPongMessage(sender, timestamp);
            }
        } else if (!isFromCurrentUser) {
            // Normal channel message (not ping/pong)
            addSystemMessage(`[${timestamp}] ${sender}: ${message.data.content}`);
        }
    }, [clientRef, addSystemMessage, handleGuestPingMessage, handleGuestPongMessage]);
    
    return { handleMessage };
}