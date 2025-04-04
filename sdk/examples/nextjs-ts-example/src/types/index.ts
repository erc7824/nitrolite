// For type imports
import { MessageType } from '@/hooks/ui/useMessageStyles';

/**
 * Message interface for application state and communication
 */
export interface Message {
    text: string;
    type: MessageType;
    sender?: string;
    timestamp?: number;
    sequence?: string; // Sequence number for state ordering
}

export type Channel = 'public' | 'game' | 'trade' | 'private';

export type WSStatus =
    | 'connected'
    | 'connecting'
    | 'disconnected'
    | 'reconnecting'
    | 'reconnect_failed'
    | 'auth_failed'
    | 'authenticating'
    | 'waiting';
