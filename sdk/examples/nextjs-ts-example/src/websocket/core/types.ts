// import { WSStatus, Channel } from "@/types";

/**
 * Enum representing the possible states of a WebSocket connection
 */
export enum WebSocketReadyState {
    CONNECTING = 0,
    OPEN = 1,
    CLOSING = 2,
    CLOSED = 3,
}

/**
 * Interface for WebSocketClient configuration options
 */
export interface WebSocketClientOptions {
    /** Whether to automatically reconnect on disconnection */
    autoReconnect: boolean;
    /** Base delay between reconnection attempts in milliseconds */
    reconnectDelay: number;
    /** Maximum number of reconnection attempts */
    maxReconnectAttempts: number;
    /** Timeout for requests in milliseconds */
    requestTimeout: number;
    /** Number of pings to send during verification (defaults to 1000) */
    pingVerificationCount?: number;
    /** Size of ping batches to send at once (defaults to 10) */
    pingBatchSize?: number;
    /** Delay between ping batches in milliseconds (defaults to 10) */
    pingBatchDelay?: number;
    /** Required success rate for ping verification (0-1, defaults to 0.95) */
    pingSuccessThreshold?: number;
    /** Channel to use for P2P ping-pong verification (defaults to 'public') */
    pingChannel?: string;
}

/**
 * Interface for RPC request parameters
 */
export interface RPCRequest {
    method: string;
    params: unknown[];
}
