import { WalletSigner } from './crypto';
import { WebSocketClientOptions } from './core/types';
import { WebSocketConnection } from './core/connection';
import { WSRequests } from './core/requests';

/**
 * Client for WebSocket communication with signature-based authentication
 */
export class WebSocketClient {
    private connection: WebSocketConnection;
    private requests: WSRequests;

    /**
     * Creates a new WebSocketClient
     *
     * @param url - The WebSocket URL to connect to
     * @param signer - The WalletSigner to use for authentication
     * @param options - Configuration options
     */
    constructor(
        url: string,
        private signer: WalletSigner,
        options: WebSocketClientOptions = {
            autoReconnect: true,
            reconnectDelay: 1000,
            maxReconnectAttempts: 5,
            requestTimeout: 10000,
        },
    ) {
        this.connection = new WebSocketConnection(url, signer, options);
        this.requests = new WSRequests(this.connection, signer);
    }

    // Forward connection methods
    onStatusChange(cb: (status: import('@/types').WSStatus) => void) {
        return this.connection.onStatusChange(cb);
    }
    onMessage(cb: (message: unknown) => void) {
        return this.connection.onMessage(cb);
    }
    onError(cb: (error: Error) => void) {
        return this.connection.onError(cb);
    }
    get readyState() {
        return this.connection.readyState;
    }
    get isConnected() {
        return this.connection.isConnected;
    }
    get currentSubscribedChannel() {
        return this.connection.currentSubscribedChannel;
    }
    getShortenedPublicKey() {
        return this.connection.getShortenedPublicKey();
    }
    connect() {
        return this.connection.connect();
    }
    close() {
        return this.connection.close();
    }

    // Forward request methods
    sendRequest(method: string, params: unknown[] = []) {
        return this.requests.sendRequest(method, params);
    }
    subscribe(channel: import('@/types').Channel) {
        return this.requests.subscribe(channel);
    }
    publishMessage(message: string) {
        return this.requests.publishMessage(message);
    }
    ping() {
        return this.requests.ping();
    }
    checkBalance(tokenAddress?: string) {
        return this.requests.checkBalance(tokenAddress);
    }
    sendBatch(requests: { method: string; params: unknown[] }[]) {
        return this.requests.sendBatch(requests);
    }
}

// Export types and helper functions from submodules
export * from './crypto';
export * from './core';

/**
 * Factory function to create a WebSocketClient
 *
 * @param url - The WebSocket URL to connect to
 * @param signer - The WalletSigner to use for authentication
 * @param options - Configuration options
 * @returns A new WebSocketClient instance
 */
export const createWebSocketClient = (
    url: string,
    signer: WalletSigner,
    options?: Partial<WebSocketClientOptions>,
): WebSocketClient => {
    return new WebSocketClient(url, signer, {
        autoReconnect: true,
        reconnectDelay: 1000,
        maxReconnectAttempts: 5,
        requestTimeout: 10000,
        ...options,
    });
};
