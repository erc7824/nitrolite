import { WebSocket } from 'ws';
import type { NitroliteConfig } from '../../types/index.js';
import { logger } from '../../utils/logger.js';
import { EventEmitter } from './event-emitter.js';

import type { MessageEvent as WSMessageEvent, CloseEvent as WSCloseEvent, ErrorEvent } from 'ws';

export interface WebSocketEvents {
    open: void;
    message: WSMessageEvent;
    close: WSCloseEvent;
    error: ErrorEvent;
}

export class WebSocketManager {
    private ws: WebSocket | null = null;
    private readonly openEmitter = new EventEmitter<void>();
    private readonly messageEmitter = new EventEmitter<WSMessageEvent>();
    private readonly closeEmitter = new EventEmitter<WSCloseEvent>();
    private readonly errorEmitter = new EventEmitter<ErrorEvent>();
    
    constructor(private config: NitroliteConfig) {}

    get isOpen(): boolean {
        return this.ws?.readyState === WebSocket.OPEN;
    }

    get isConnecting(): boolean {
        return this.ws?.readyState === WebSocket.CONNECTING;
    }

    get isClosed(): boolean {
        return !this.ws || this.ws.readyState === WebSocket.CLOSED;
    }

    get rawWebSocket(): WebSocket | null {
        return this.ws;
    }

    async connect(): Promise<void> {
        if (this.ws && (this.isOpen || this.isConnecting)) {
            return;
        }

        return new Promise((resolve, reject) => {
            try {
                logger.debug(`WebSocket connecting to: ${this.config.wsUrl.substring(0, 30)}...`);
                this.ws = new WebSocket(this.config.wsUrl);

                this.ws.onopen = () => {
                    logger.info('WebSocket connection opened');
                    this.openEmitter.emit();
                    resolve();
                };

                this.ws.onmessage = (event) => {
                    this.messageEmitter.emit(event);
                };

                this.ws.onclose = (event) => {
                    logger.info('WebSocket connection closed');
                    this.closeEmitter.emit(event);
                };

                this.ws.onerror = (error) => {
                    logger.error('WebSocket error:', error);
                    this.errorEmitter.emit(error);
                    reject(new Error('WebSocket connection failed'));
                };
            } catch (error) {
                reject(error);
            }
        });
    }

    send(data: string): void {
        if (!this.isOpen) {
            throw new Error('WebSocket is not connected');
        }
        this.ws!.send(data);
    }

    close(): void {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }

    onOpen(listener: () => void): () => void {
        return this.openEmitter.add(listener);
    }

    onMessage(listener: (event: WSMessageEvent) => void): () => void {
        return this.messageEmitter.add(listener);
    }

    onClose(listener: (event: WSCloseEvent) => void): () => void {
        return this.closeEmitter.add(listener);
    }

    onError(listener: (error: ErrorEvent) => void): () => void {
        return this.errorEmitter.add(listener);
    }

    destroy(): void {
        this.close();
        this.openEmitter.clear();
        this.messageEmitter.clear();
        this.closeEmitter.clear();
        this.errorEmitter.clear();
    }
}