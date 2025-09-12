import type { Hex, WalletClient } from 'viem';

// Use viem WalletClient directly since we're now consistently using viem clients from Privy
export type CompatibleWalletClient = WalletClient;

export type WSStatus = 'disconnected' | 'connecting' | 'connected' | 'reconnecting' | 'failed' | 'pending_auth';

export interface SessionKey {
    privateKey: string;
    address: string;
}

export interface YellowAuthContext {
    walletAddress: string;
    sessionKey: SessionKey;
    walletClient?: CompatibleWalletClient;
    signTypedData: (args: { domain: any; types: any; primaryType: string; message: any }) => Promise<Hex>;
}

export interface YellowConfig {
    wsUrl: string;
    pingInterval: number;
    reconnectDelay: number;
    maxRetries: number;
    requestTimeout: number;
}

export interface YellowConnectionCallbacks {
    onMessage?: (data: any) => void;
    onConnect?: () => void;
    onDisconnect?: () => void;
    onError?: (error: Error) => void;
    onAuthSuccess?: () => void;
    onAuthFailed?: (error: string) => void;
    onChallengeReceived?: (challengeData: any) => void;
    onVerifyFailed?: (error: string) => void;
}

export const YellowMessageType = {
    AUTH_REQUEST: 'auth_request',
    AUTH_CHALLENGE: 'auth_challenge',
    AUTH_VERIFY: 'auth_verify',
    PING: 'ping',
    PONG: 'pong',
    ERROR: 'error',
    NOTIFICATION: 'notification',
    BALANCE_UPDATE: 'balance_update',
    TRANSACTION: 'transaction',
    ORDER_STATUS: 'order_status',
} as const;

export type YellowMessageTypeValue = (typeof YellowMessageType)[keyof typeof YellowMessageType];

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
