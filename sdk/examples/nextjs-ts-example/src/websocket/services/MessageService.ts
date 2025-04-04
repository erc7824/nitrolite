import { proxy, useSnapshot } from 'valtio';
import { Message, Channel, WSStatus } from '@/types';
// import { MessageType } from '@/hooks/useMessageStyles';

// Message state interface with minimized properties
export interface IMessageState {
    messages: Message[];
    activeChannel: Channel;
    status: WSStatus;
}

// Single proxy state for messages
const state = proxy<IMessageState>({
    messages: [],
    activeChannel: 'public',
    status: 'disconnected',
});

/**
 * Message Service - Centralized handling of all application messages
 * - Used by WebSocket and UI components
 * - Provides hooks and utility functions for message management
 */
const MessageService = {
    state,

    // Common channel operations
    channels: {
        setActive(channel: Channel) {
            state.activeChannel = channel;
            MessageService.system(`Switched to ${channel} channel`);
        },

        getActive: () => state.activeChannel,
    },

    // Connection status
    status: {
        set(status: WSStatus) {
            state.status = status;
            MessageService.system(`Connection status: ${status}`);
        },

        get: () => state.status,
    },

    // Message type shortcuts (for cleaner code elsewhere)
    system: (text: string) => MessageService.add({ text, type: 'system' }),
    error: (text: string) => MessageService.add({ text, type: 'error' }),
    sent: (text: string, sender?: string) => MessageService.add({ text, type: 'sent', sender }),
    received: (text: string, sender?: string) => MessageService.add({ text, type: 'received', sender }),
    success: (text: string) => MessageService.add({ text, type: 'success' }),

    // Core message handler
    add(message: Partial<Message>) {
        if (!message.text) return;

        state.messages.push({
            text: message.text,
            type: message.type || 'info',
            sender: message.sender,
            timestamp: message.timestamp || Date.now(),
        });

        // Limit message history to prevent memory issues (last 200 messages)
        if (state.messages.length > 200) {
            state.messages = state.messages.slice(-200);
        }
    },

    // Clear all messages
    clear() {
        state.messages = [];
    },

    // Parse and handle incoming WebSocket message
    handleWebSocketMessage(data: unknown) {
        if (!data || typeof data !== 'object') return;

        // Type guard function to verify object shape
        const hasProperty = <T extends object, K extends string>(obj: T, prop: K): obj is T & Record<K, unknown> => {
            return prop in obj;
        };

        if (hasProperty(data, 'type') && typeof data.type === 'string') {
            if (data.type === 'message' && hasProperty(data, 'data') && typeof data.data === 'object' && data.data) {
                const messageData = data.data;

                if (
                    hasProperty(messageData, 'message') &&
                    hasProperty(messageData, 'sender') &&
                    typeof messageData.message === 'string' &&
                    typeof messageData.sender === 'string'
                ) {
                    MessageService.received(messageData.message, messageData.sender);
                } else {
                    MessageService.received(`Received: ${JSON.stringify(messageData)}`);
                }
            } else if (
                data.type === 'pong' &&
                hasProperty(data, 'data') &&
                typeof data.data === 'object' &&
                data.data
            ) {
                const timestamp = hasProperty(data.data, 'timestamp') ? data.data.timestamp : 'no timestamp';

                MessageService.received(`Server responded with pong (${timestamp})`);
            } else if (
                data.type === 'channel_message' &&
                hasProperty(data, 'data') &&
                typeof data.data === 'object' &&
                data.data &&
                hasProperty(data.data, 'content') &&
                hasProperty(data.data, 'sender')
            ) {
                const content = String(data.data.content).toLowerCase();
                const sender = String(data.data.sender);
                
                if (content === 'ping' || content === 'pong') {
                    // These are handled by the useWebSocket hook
                    return;
                } else {
                    MessageService.received(`${sender}: ${data.data.content}`, sender);
                }
            } else if (
                data.type === 'rpc_response' &&
                hasProperty(data, 'data') &&
                typeof data.data === 'object' &&
                data.data
            ) {
                if (
                    hasProperty(data.data, 'method') &&
                    data.data.method === 'get_balance' &&
                    hasProperty(data.data, 'result')
                ) {
                    MessageService.success(`Balance: ${data.data.result || '0'} tokens`);
                }
            } else if (data.type === 'auth_success') {
                MessageService.system('Authentication successful');
            } else if (
                data.type === 'subscribe_success' &&
                hasProperty(data, 'data') &&
                typeof data.data === 'object' &&
                data.data &&
                hasProperty(data.data, 'channel') &&
                (data.data.channel === 'public' ||
                    data.data.channel === 'game' ||
                    data.data.channel === 'trade' ||
                    data.data.channel === 'private')
            ) {
                MessageService.channels.setActive(data.data.channel);
            }
        }
    },

    // Hook for components to use message data
    useMessages() {
        const { messages } = useSnapshot(state);

        return {
            messages,
            clear: MessageService.clear,
        };
    },

    // Hook for components to use connection status
    useConnectionStatus() {
        const { status } = useSnapshot(state);

        return status;
    },
};

export default MessageService;
