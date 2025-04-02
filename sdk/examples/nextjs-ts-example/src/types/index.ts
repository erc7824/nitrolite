import { MessageType } from '@/hooks/useMessageStyles';

export type Message = {
  text: string;
  type: MessageType;
  sender?: string;
  timestamp?: number;
};

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