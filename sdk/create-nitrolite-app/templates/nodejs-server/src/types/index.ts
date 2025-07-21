export type WSStatus = 'disconnected' | 'connecting' | 'connected' | 'reconnecting' | 'failed' | 'pending_auth';

export interface CryptoKeypair {
  privateKey: string;
  address: string;
}

export interface SessionKey {
  privateKey: string;
  address: string;
}

export interface NitroliteAuthContext {
  walletAddress: string;
  sessionKey: SessionKey;
  privateKey: string;
}

export interface NitroliteConfig {
  wsUrl: string;
  pingInterval: number;
  reconnectDelay: number;
  maxRetries: number;
  requestTimeout: number;
}

export interface NitroliteConnectionCallbacks {
  onMessage?: (data: any) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: Error) => void;
  onAuthSuccess?: () => void;
  onAuthFailed?: (error: string) => void;
  onChallengeReceived?: (challengeData: any) => void;
  onVerifyFailed?: (error: string) => void;
}

export class UserRejectedError extends Error {
  constructor(message: string = 'User rejected the signing request') {
    super(message);
    this.name = 'UserRejectedError';
  }

  static isUserRejection(error: Error): boolean {
    const msg = error.message.toLowerCase();
    return (
      msg.includes('user rejected') ||
      msg.includes('user denied') ||
      msg.includes('user cancelled') ||
      msg.includes('rejected by user') ||
      msg.includes('user canceled') ||
      msg.includes('request rejected')
    );
  }
}

export interface RequestInfo {
  requestId: string;
  method: string;
  params: any;
  timestamp: number;
}
