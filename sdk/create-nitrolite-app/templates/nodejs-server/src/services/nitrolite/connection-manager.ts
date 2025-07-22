import { createPingMessage, createECDSAMessageSigner } from '@erc7824/nitrolite';
import type { NitroliteConfig, SessionKey } from '../../types/index.js';
import { UserRejectedError } from '../../types/index.js';
import { logger } from '../../utils/logger.js';
import { EventEmitter } from './event-emitter.js';

export interface ConnectionEvents {
    reconnectScheduled: number;
    maxRetriesReached: void;
    pingFailed: void;
}

export class ConnectionManager {
    private retryCount = 0;
    private userRejectedAuth = false;
    private isDestroyed = false;
    private pingInterval: NodeJS.Timeout | null = null;
    private reconnectTimeout: NodeJS.Timeout | null = null;
    
    private readonly reconnectScheduledEmitter = new EventEmitter<number>();
    private readonly maxRetriesReachedEmitter = new EventEmitter<void>();
    private readonly pingFailedEmitter = new EventEmitter<void>();

    constructor(private config: NitroliteConfig) {}

    get currentRetryCount(): number {
        return this.retryCount;
    }

    get hasUserRejectedAuth(): boolean {
        return this.userRejectedAuth;
    }

    get destroyed(): boolean {
        return this.isDestroyed;
    }

    shouldReconnect(): boolean {
        return !this.userRejectedAuth && !this.isDestroyed && this.retryCount < this.config.maxRetries;
    }

    handleConnectionError(error: Error): void {
        if (error instanceof UserRejectedError || UserRejectedError.isUserRejection(error)) {
            this.userRejectedAuth = true;
            logger.info('User rejected authentication - disabling reconnection');
        }
    }

    scheduleReconnect(reconnectCallback: () => Promise<void>): void {
        if (this.isDestroyed || !this.shouldReconnect()) {
            this.maxRetriesReachedEmitter.emit();
            return;
        }

        this.retryCount++;
        const delay = this.config.reconnectDelay * Math.pow(2, this.retryCount - 1);
        
        logger.warn(`Scheduling reconnect attempt ${this.retryCount}/${this.config.maxRetries} in ${delay}ms`);
        this.reconnectScheduledEmitter.emit(delay);
        
        this.clearReconnectTimeout();
        this.reconnectTimeout = setTimeout(async () => {
            try {
                await reconnectCallback();
            } catch (error) {
                logger.warn('Reconnection attempt failed:', error instanceof Error ? error.message : error);
            }
        }, delay);
    }

    resetRetryCount(): void {
        this.retryCount = 0;
    }

    resetUserRejection(): void {
        this.userRejectedAuth = false;
    }

    startPingInterval(sessionKey: SessionKey | null, wsSend: (data: string) => void): void {
        this.clearPingInterval();
        
        this.pingInterval = setInterval(async () => {
            if (!sessionKey) {
                logger.warn('Cannot ping - no session key available');
                this.pingFailedEmitter.emit();
                return;
            }

            try {
                const sessionSigner = createECDSAMessageSigner(sessionKey.privateKey as `0x${string}`);
                const pingMessage = await createPingMessage(sessionSigner);
                wsSend(pingMessage);
            } catch (error) {
                logger.warn('Ping failed:', error);
                this.pingFailedEmitter.emit();
            }
        }, this.config.pingInterval);
    }

    private clearPingInterval(): void {
        if (this.pingInterval) {
            clearInterval(this.pingInterval);
            this.pingInterval = null;
        }
    }

    private clearReconnectTimeout(): void {
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }
    }

    cleanup(): void {
        this.clearPingInterval();
        this.clearReconnectTimeout();
    }

    destroy(): void {
        this.isDestroyed = true;
        this.cleanup();
        this.reconnectScheduledEmitter.clear();
        this.maxRetriesReachedEmitter.clear();
        this.pingFailedEmitter.clear();
    }

    onReconnectScheduled(listener: (delay: number) => void): () => void {
        return this.reconnectScheduledEmitter.add(listener);
    }

    onMaxRetriesReached(listener: () => void): () => void {
        return this.maxRetriesReachedEmitter.add(listener);
    }

    onPingFailed(listener: () => void): () => void {
        return this.pingFailedEmitter.add(listener);
    }
}