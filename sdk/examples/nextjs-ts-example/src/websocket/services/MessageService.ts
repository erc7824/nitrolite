import { proxy, useSnapshot } from 'valtio';
import { Message, Channel, WSStatus } from '@/types';
import { MessageType } from '@/hooks/useMessageStyles';

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
  status: 'disconnected'
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
    
    getActive: () => state.activeChannel
  },
  
  // Connection status
  status: {
    set(status: WSStatus) {
      state.status = status;
      MessageService.system(`Connection status: ${status}`);
    },
    
    get: () => state.status
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
      timestamp: message.timestamp || Date.now()
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
  handleWebSocketMessage(data: any) {
    if (!data) return;
    
    if (data.type === 'message' && data.data) {
      const messageData = data.data;
      if (messageData.message && messageData.sender) {
        MessageService.received(messageData.message, messageData.sender);
      } else {
        MessageService.received(`Received: ${JSON.stringify(messageData)}`);
      }
    } else if (data.type === 'pong' && data.data) {
      MessageService.received(`Server responded with pong (${data.data.timestamp || 'no timestamp'})`);
    } else if (data.type === 'rpc_response' && data.data) {
      if (data.data.method === 'get_balance') {
        MessageService.success(`Balance: ${data.data.result || '0'} tokens`);
      }
    } else if (data.type === 'auth_success') {
      MessageService.system('Authentication successful');
    } else if (data.type === 'subscribe_success' && data.data?.channel) {
      MessageService.channels.setActive(data.data.channel as Channel);
    }
  },
  
  // Hook for components to use message data
  useMessages() {
    const { messages } = useSnapshot(state);
    return {
      messages,
      clear: MessageService.clear
    };
  },
  
  // Hook for components to use connection status
  useConnectionStatus() {
    const { status } = useSnapshot(state);
    return status;
  }
};

export default MessageService;