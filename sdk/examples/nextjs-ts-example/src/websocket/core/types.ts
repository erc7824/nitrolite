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
}

/**
 * Interface for RPC request parameters
 */
export interface RPCRequest {
    method: string;
    params: unknown[];
}
