import { proxy } from 'valtio';
import { Message, Channel } from '@/types';
import { MessageType } from '@/hooks/useMessageStyles';

export interface IMessageState {
  messages: Message[];
  activeChannel: Channel;
}

const state = proxy<IMessageState>({
  messages: [],
  activeChannel: 'public',
});

const MessageStore = {
  state,

  addMessage(message: Message) {
    state.messages.push({
      ...message,
      timestamp: message.timestamp || Date.now(),
    });
  },

  addSystemMessage(text: string) {
    this.addMessage({
      text,
      type: 'system',
      timestamp: Date.now(),
    });
  },

  addErrorMessage(text: string) {
    this.addMessage({
      text,
      type: 'error',
      timestamp: Date.now(),
    });
  },

  addSentMessage(text: string, sender?: string) {
    this.addMessage({
      text,
      type: 'sent',
      sender,
      timestamp: Date.now(),
    });
  },

  addReceivedMessage(text: string, sender?: string) {
    this.addMessage({
      text,
      type: 'received',
      sender,
      timestamp: Date.now(),
    });
  },

  clearMessages() {
    state.messages = [];
  },

  setActiveChannel(channel: Channel) {
    state.activeChannel = channel;
    this.addSystemMessage(`Switched to ${channel} channel`);
  }
};

export default MessageStore;