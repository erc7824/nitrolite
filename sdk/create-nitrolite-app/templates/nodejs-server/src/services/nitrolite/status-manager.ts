import type { WSStatus } from '../../types/index.js';
import { EventEmitter } from './event-emitter.js';

export class StatusManager {
    private status: WSStatus = 'disconnected';
    private readonly statusEmitter = new EventEmitter<WSStatus>();

    get currentStatus(): WSStatus {
        return this.status;
    }

    get isConnected(): boolean {
        return this.status === 'connected';
    }

    get isConnecting(): boolean {
        return this.status === 'connecting';
    }

    get isReconnecting(): boolean {
        return this.status === 'reconnecting';
    }

    get isFailed(): boolean {
        return this.status === 'failed';
    }

    get isPendingAuth(): boolean {
        return this.status === 'pending_auth';
    }

    get isDisconnected(): boolean {
        return this.status === 'disconnected';
    }

    setStatus(newStatus: WSStatus): void {
        if (this.status !== newStatus) {
            this.status = newStatus;
            this.statusEmitter.emit(newStatus);
        }
    }

    onStatusChange(listener: (status: WSStatus) => void): () => void {
        return this.statusEmitter.add(listener);
    }

    reset(): void {
        this.setStatus('disconnected');
    }

    destroy(): void {
        this.statusEmitter.clear();
    }
}