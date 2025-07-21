import { logger } from '../utils/logger.js';

interface StoredMessage {
  id: string;
  timestamp: number;
  type: string;
  data: any;
  source: 'nitrolite' | 'websocket' | 'api';
}

class MessageStoreClass {
  private messages: StoredMessage[] = [];
  private maxMessages = 1000; // Keep last 1000 messages
  private messageListeners = new Set<(message: StoredMessage) => void>();

  addMessage(type: string, data: any, source: 'nitrolite' | 'websocket' | 'api' = 'nitrolite'): string {
    const message: StoredMessage = {
      id: this.generateId(),
      timestamp: Date.now(),
      type,
      data,
      source,
    };

    this.messages.unshift(message);

    // Trim to max messages
    if (this.messages.length > this.maxMessages) {
      this.messages = this.messages.slice(0, this.maxMessages);
    }

    // Notify listeners
    this.messageListeners.forEach(listener => {
      try {
        listener(message);
      } catch (error) {
        logger.error('Error in message listener:', error);
      }
    });

    logger.debug(`Added message ${message.id} (${type}) from ${source}`);
    return message.id;
  }

  getMessage(id: string): StoredMessage | undefined {
    return this.messages.find(msg => msg.id === id);
  }

  getMessages(options: {
    limit?: number;
    offset?: number;
    type?: string;
    source?: 'nitrolite' | 'websocket' | 'api';
    since?: number;
  } = {}): StoredMessage[] {
    let filtered = this.messages;

    // Filter by type
    if (options.type) {
      filtered = filtered.filter(msg => msg.type === options.type);
    }

    // Filter by source
    if (options.source) {
      filtered = filtered.filter(msg => msg.source === options.source);
    }

    // Filter by timestamp
    if (options.since) {
      filtered = filtered.filter(msg => msg.timestamp >= options.since);
    }

    // Apply offset
    if (options.offset) {
      filtered = filtered.slice(options.offset);
    }

    // Apply limit
    if (options.limit) {
      filtered = filtered.slice(0, options.limit);
    }

    return filtered;
  }

  getRecentMessages(count: number = 10): StoredMessage[] {
    return this.messages.slice(0, count);
  }

  getMessageStats(): {
    totalMessages: number;
    messagesByType: Record<string, number>;
    messagesBySource: Record<string, number>;
    oldestMessage?: number;
    newestMessage?: number;
  } {
    const stats = {
      totalMessages: this.messages.length,
      messagesByType: {} as Record<string, number>,
      messagesBySource: {} as Record<string, number>,
      oldestMessage: undefined as number | undefined,
      newestMessage: undefined as number | undefined,
    };

    if (this.messages.length > 0) {
      stats.newestMessage = this.messages[0].timestamp;
      stats.oldestMessage = this.messages[this.messages.length - 1].timestamp;
    }

    // Count by type and source
    for (const message of this.messages) {
      stats.messagesByType[message.type] = (stats.messagesByType[message.type] || 0) + 1;
      stats.messagesBySource[message.source] = (stats.messagesBySource[message.source] || 0) + 1;
    }

    return stats;
  }

  clearMessages(): void {
    const count = this.messages.length;
    this.messages = [];
    logger.info(`Cleared ${count} stored messages`);
  }

  onMessage(listener: (message: StoredMessage) => void): () => void {
    this.messageListeners.add(listener);
    return () => this.messageListeners.delete(listener);
  }

  private generateId(): string {
    return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  setMaxMessages(max: number): void {
    this.maxMessages = max;
    
    // Trim existing messages if needed
    if (this.messages.length > max) {
      this.messages = this.messages.slice(0, max);
    }
  }

  getMaxMessages(): number {
    return this.maxMessages;
  }
}

const MessageStore = new MessageStoreClass();

export default MessageStore;
export type { StoredMessage };