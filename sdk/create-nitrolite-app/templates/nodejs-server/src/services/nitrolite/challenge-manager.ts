import { createPingMessage, createECDSAMessageSigner } from '@erc7824/nitrolite';
import type { SessionKey } from '../../types/index.js';
import { logger } from '../../utils/logger.js';
import { EventEmitter } from './event-emitter.js';

export interface ChallengeEvents {
    received: any;
    timeout: void;
    cleared: void;
}

export class ChallengeManager {
    private pendingChallenge: any = null;
    private rawChallengeMessage: string | null = null;
    private challengeTimeout: NodeJS.Timeout | null = null;
    private challengeKeepAliveInterval: NodeJS.Timeout | null = null;
    
    private readonly receivedEmitter = new EventEmitter<any>();
    private readonly timeoutEmitter = new EventEmitter<void>();
    private readonly clearedEmitter = new EventEmitter<void>();

    get hasPendingChallenge(): boolean {
        return this.pendingChallenge !== null;
    }

    get challenge(): any {
        return this.pendingChallenge;
    }

    get rawMessage(): string | null {
        return this.rawChallengeMessage;
    }

    setChallenge(challenge: any, rawMessage?: string): void {
        this.clearChallenge();
        this.pendingChallenge = challenge;
        this.rawChallengeMessage = rawMessage || null;
        this.receivedEmitter.emit(challenge);
        this.startKeepAlive();
    }

    clearChallenge(): void {
        if (this.pendingChallenge || this.rawChallengeMessage) {
            this.pendingChallenge = null;
            this.rawChallengeMessage = null;
            this.clearedEmitter.emit();
        }
        
        this.clearTimeout();
        this.clearKeepAlive();
    }

    setTimeout(timeout: number): void {
        this.clearTimeout();
        this.challengeTimeout = setTimeout(() => {
            this.timeoutEmitter.emit();
            this.clearChallenge();
        }, timeout);
    }

    private clearTimeout(): void {
        if (this.challengeTimeout) {
            clearTimeout(this.challengeTimeout);
            this.challengeTimeout = null;
        }
    }

    startKeepAlive(sessionKey?: SessionKey, wsSend?: (data: string) => void): void {
        this.clearKeepAlive();
        
        if (sessionKey && wsSend) {
            this.challengeKeepAliveInterval = setInterval(async () => {
                try {
                    const sessionSigner = createECDSAMessageSigner(sessionKey.privateKey as `0x${string}`);
                    const pingMessage = await createPingMessage(sessionSigner);
                    wsSend(pingMessage);
                } catch (error) {
                    logger.warn('Challenge keep-alive ping failed:', error);
                    this.clearKeepAlive();
                }
            }, 30000);
        }
    }

    private clearKeepAlive(): void {
        if (this.challengeKeepAliveInterval) {
            clearInterval(this.challengeKeepAliveInterval);
            this.challengeKeepAliveInterval = null;
        }
    }

    onChallengeReceived(listener: (challenge: any) => void): () => void {
        return this.receivedEmitter.add(listener);
    }

    onChallengeTimeout(listener: () => void): () => void {
        return this.timeoutEmitter.add(listener);
    }

    onChallengeCleared(listener: () => void): () => void {
        return this.clearedEmitter.add(listener);
    }

    destroy(): void {
        this.clearChallenge();
        this.receivedEmitter.clear();
        this.timeoutEmitter.clear();
        this.clearedEmitter.clear();
    }
}
