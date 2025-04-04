import { useCallback } from 'react';
import { useSnapshot } from 'valtio';
import MessageService from '@/websocket/services/MessageService';
import { Channel, WSStatus } from '@/types';
import { MessageType } from './useMessageStyles';

/**
 * Custom hook that provides access to the MessageService with
 * reactive state updates using valtio
 */
export function useMessageService() {
    const { messages, activeChannel, status } = useSnapshot(MessageService.state);

    // Message functionality
    const addMessage = useCallback((text: string, type: MessageType = 'info', sender?: string) => {
        MessageService.add({ text, type, sender });
    }, []);

    const addSystemMessage = useCallback((text: string) => {
        MessageService.system(text);
    }, []);

    const addErrorMessage = useCallback((text: string) => {
        MessageService.error(text);
    }, []);

    const addSentMessage = useCallback((text: string, sender?: string) => {
        MessageService.sent(text, sender);
    }, []);

    const addReceivedMessage = useCallback((text: string, sender?: string) => {
        MessageService.received(text, sender);
    }, []);

    const clearMessages = useCallback(() => {
        MessageService.clear();
    }, []);

    // Channel functionality
    const setActiveChannel = useCallback((channel: Channel) => {
        MessageService.channels.setActive(channel);
    }, []);

    // Status functionality
    const setStatus = useCallback((newStatus: string) => {
        MessageService.status.set(newStatus as WSStatus);
    }, []);

    // Add specialized message types for ping-pong with different parties
    const addPingMessage = useCallback((text: string, party?: 'user' | 'guest') => {
        MessageService.add({ text, type: party === 'guest' ? 'guest-ping' : 'user-ping', sender: party });
    }, []);

    const addPongMessage = useCallback((text: string, party?: 'user' | 'guest') => {
        MessageService.add({ text, type: party === 'guest' ? 'guest-pong' : 'user-pong', sender: party });
    }, []);

    return {
        // State
        messages,
        activeChannel,
        status,

        // Message methods
        addMessage,
        addSystemMessage,
        addErrorMessage,
        addSentMessage,
        addReceivedMessage,
        addPingMessage,
        addPongMessage,
        clearMessages,

        // Channel methods
        setActiveChannel,

        // Status methods
        setStatus,
    };
}
