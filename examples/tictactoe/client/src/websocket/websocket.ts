import { type Hex } from "viem";
import { ethers } from "ethers";
import { createAuthRequestMessage, NitroliteRPC, createAuthVerifyMessage, createPingMessage } from "@erc7824/nitrolite";
import type { Channel } from "@erc7824/nitrolite";

// ===== Types =====

/**
 * WebSocket ready states
 */
export const WebSocketReadyState = {
    CONNECTING: 0,
    OPEN: 1,
    CLOSING: 2,
    CLOSED: 3,
} as const;

export type WebSocketReadyState = (typeof WebSocketReadyState)[keyof typeof WebSocketReadyState];

/**
 * WebSocket connection status
 */
export type WSStatus = "connected" | "connecting" | "disconnected" | "reconnecting" | "reconnect_failed" | "auth_failed" | "authenticating";

/**
 * WebSocket client configuration options
 */
export interface WebSocketClientOptions {
    autoReconnect: boolean;
    reconnectDelay: number;
    maxReconnectAttempts: number;
    requestTimeout: number;
}

/**
 * Wallet signer interface
 */
export interface WalletSigner {
    address: Hex;
    sign: (payload: any) => Promise<Hex>;
}

/**
 * Gets address from a public key
 */
export const getAddressFromPublicKey = (publicKey: string): string => {
    const formattedKey = publicKey.startsWith("0x") ? publicKey : `0x${publicKey}`;
    const hash = ethers.keccak256(formattedKey);
    const address = `0x${hash.slice(-40)}`;
    return ethers.getAddress(address);
};

// ===== Connection =====

/**
 * Core WebSocket client for browser applications
 */
export class WebSocketClient {
    private ws: WebSocket | null = null;
    private pendingRequests = new Map<number, { resolve: (value: unknown) => void; reject: (reason: Error) => void }>();
    // private requestCounter = 0;
    private reconnectAttempts = 0;
    private reconnectTimeout: any = null;
    private statusHandlers: ((status: WSStatus) => void)[] = [];
    private messageHandlers: ((message: unknown) => void)[] = [];
    private errorHandlers: ((error: Error) => void)[] = [];
    private currentChannel: any = null;
    private nitroliteChannel: Channel | null = null;

    /**
     * Creates a new WebSocket client
     */
    private url: string;
    private signer: WalletSigner;
    private options: WebSocketClientOptions;

    constructor(
        url: string,
        signer: WalletSigner,
        options: WebSocketClientOptions = {
            autoReconnect: true,
            reconnectDelay: 1000,
            maxReconnectAttempts: 5,
            requestTimeout: 10000,
        }
    ) {
        this.url = url;
        this.signer = signer;
        this.options = options;
    }

    /**
     * Registers a status change callback
     */
    onStatusChange(callback: (status: WSStatus) => void): void {
        this.statusHandlers.push(callback);
    }

    /**
     * Registers a message handler
     */
    onMessage(callback: (message: unknown) => void): void {
        this.messageHandlers.push(callback);
    }

    /**
     * Registers an error handler
     */
    onError(callback: (error: Error) => void): void {
        this.errorHandlers.push(callback);
    }

    /**
     * Gets whether the client is connected
     */
    get isConnected(): boolean {
        return this.ws !== null && this.ws.readyState === WebSocketReadyState.OPEN;
    }

    /**
     * Gets the current WebSocket ready state
     */
    get readyState(): WebSocketReadyState {
        return this.ws ? (this.ws.readyState as WebSocketReadyState) : WebSocketReadyState.CLOSED;
    }

    /**
     * Gets the current channel
     */
    get currentSubscribedChannel(): any {
        return this.currentChannel;
    }

    /**
     * Gets the current Nitrolite channel
     */
    get currentNitroliteChannel(): Channel | null {
        return this.nitroliteChannel;
    }

    /**
     * Sets the Nitrolite channel
     */
    setNitroliteChannel(channel: Channel): void {
        this.nitroliteChannel = channel;
    }

    /**
     * Connects to the WebSocket server
     */
    async connect(): Promise<void> {
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }

        if (this.isConnected) return;

        return new Promise((resolve, reject) => {
            try {
                this.ws = new WebSocket(this.url);
                this.emitStatus("connecting");

                this.ws.onopen = async () => {
                    try {
                        this.emitStatus("authenticating");
                        await this.authenticate();
                        this.emitStatus("connected");
                        this.reconnectAttempts = 0;
                        resolve();
                    } catch (error) {
                        this.emitStatus("auth_failed");
                        this.emitError(error instanceof Error ? error : new Error(String(error)));
                        reject(error);
                        this.close();
                        this.handleReconnect();
                    }
                };

                this.ws.onmessage = this.handleMessage.bind(this);

                this.ws.onerror = () => {
                    this.emitError(new Error("WebSocket connection error"));
                    reject(new Error("WebSocket connection error"));
                };

                this.ws.onclose = () => {
                    this.emitStatus("disconnected");
                    this.ws = null;
                    this.currentChannel = null;

                    this.pendingRequests.forEach(({ reject }) => reject(new Error("WebSocket connection closed")));
                    this.pendingRequests.clear();

                    this.handleReconnect();
                };
            } catch (error) {
                reject(error);
                this.handleReconnect();
            }
        });
    }

    /**
     * Authenticates with the WebSocket server
     */
    private async authenticate(): Promise<void> {
        if (!this.ws) throw new Error("WebSocket not connected");

        // Create and send auth request
        const authRequest = await createAuthRequestMessage(this.signer.sign, this.signer.address);
        this.ws.send(authRequest);

        return new Promise((resolve, reject) => {
            const authTimeout = setTimeout(() => {
                this.ws?.removeEventListener("message", handleAuthResponse);
                reject(new Error("Authentication timeout"));
            }, this.options.requestTimeout);

            const handleAuthResponse = async (event: MessageEvent) => {
                let response;

                try {
                    response = JSON.parse(event.data);
                } catch (error) {
                    // Skip invalid messages
                    return;
                }

                try {
                    if (response.res && response.res[1] === "auth_challenge") {
                        // Handle challenge response
                        const authVerify = await createAuthVerifyMessage(this.signer.sign, event.data, this.signer.address);
                        this.ws?.send(authVerify);
                    } else if (response.res && response.res[1] === "auth_verify") {
                        // Authentication successful
                        const paramsForChannels = [{ participant: this.signer.address }];
                        const getChannelsMessage = NitroliteRPC.createRequest(10, "get_channels", paramsForChannels);
                        const getChannelMessage = await NitroliteRPC.signRequestMessage(getChannelsMessage, this.signer.sign);
                        console.log("getChannelMessage", getChannelMessage);
                        this.ws?.send(JSON.stringify(getChannelMessage));
                        clearTimeout(authTimeout);
                        this.ws?.removeEventListener("message", handleAuthResponse);
                        resolve();
                    } else if (response.err) {
                        // Authentication error
                        const errorMsg = response.err[2] || "Authentication failed";
                        clearTimeout(authTimeout);
                        this.ws?.removeEventListener("message", handleAuthResponse);
                        reject(new Error(String(errorMsg)));
                    }
                } catch (error) {
                    clearTimeout(authTimeout);
                    this.ws?.removeEventListener("message", handleAuthResponse);
                    reject(new Error(`Authentication error: ${error instanceof Error ? error.message : String(error)}`));
                }
            };

            this.ws?.addEventListener("message", handleAuthResponse);
        });
    }

    /**
     * Handles reconnection logic
     */
    private handleReconnect(): void {
        if (!this.options.autoReconnect || this.reconnectAttempts >= this.options.maxReconnectAttempts) {
            if (this.reconnectAttempts >= this.options.maxReconnectAttempts) {
                this.emitStatus("reconnect_failed");
            }
            return;
        }

        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
        }

        this.reconnectAttempts++;
        const delay = this.options.reconnectDelay * this.reconnectAttempts;

        this.emitStatus("reconnecting");

        this.reconnectTimeout = setTimeout(() => {
            this.connect().catch(() => {
                // Silent catch to prevent unhandled rejections
            });
        }, delay);
    }

    /**
     * Closes the WebSocket connection
     */
    close(): void {
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }

        if (this.ws && (this.ws.readyState === WebSocketReadyState.OPEN || this.ws.readyState === WebSocketReadyState.CONNECTING)) {
            try {
                this.ws.close(1000, "Normal closure");
            } catch (err) {
                console.error("Error while closing WebSocket:", err);
            }
        }

        this.ws = null;
        this.currentChannel = null;

        this.pendingRequests.forEach(({ reject }) => reject(new Error("WebSocket connection closed by client")));
        this.pendingRequests.clear();
        this.emitStatus("disconnected");
    }

    /**
     * Emits a status change to all registered handlers
     */
    private emitStatus(status: WSStatus): void {
        this.statusHandlers.forEach((handler) => handler(status));
    }

    /**
     * Emits a message to all registered handlers
     */
    private emitMessage(message: unknown): void {
        this.messageHandlers.forEach((handler) => handler(message));
    }

    /**
     * Emits an error to all registered handlers
     */
    private emitError(error: Error): void {
        this.errorHandlers.forEach((handler) => handler(error));
    }

    /**
     * Handles incoming WebSocket messages
     */
    private handleMessage(event: MessageEvent): void {
        let message;

        try {
            message = JSON.parse(event.data);
        } catch (error) {
            this.emitError(new Error(`Failed to parse message: ${event.data}`));
            return;
        }

        try {
            // Notify message handlers
            this.emitMessage(message);

            if (typeof message !== "object" || message === null) {
                return;
            }

            // Type guard to check for property existence
            const hasProperty = <T extends object, K extends string>(obj: T, prop: K): obj is T & Record<K, unknown> => {
                return prop in obj;
            };

            // Handle standard RPC responses (success)
            if (hasProperty(message, "res") && Array.isArray(message.res) && message.res.length >= 3) {
                const requestId = typeof message.res[0] === "number" ? message.res[0] : -1;
                if (this.pendingRequests.has(requestId)) {
                    this.pendingRequests.get(requestId)!.resolve(message.res[2]);
                    this.pendingRequests.delete(requestId);
                }
                return;
            }

            // Handle error responses
            if (hasProperty(message, "err") && Array.isArray(message.err) && message.err.length >= 3) {
                const requestId = typeof message.err[0] === "number" ? message.err[0] : -1;
                const errorMessage = `Error ${message.err[1]}: ${message.err[2]}`;

                if (this.pendingRequests.has(requestId)) {
                    this.pendingRequests.get(requestId)!.reject(new Error(errorMessage));
                    this.pendingRequests.delete(requestId);
                }
                return;
            }

            // Handle typed messages
            if (hasProperty(message, "type") && typeof message.type === "string") {
                // Handle channel subscription
                if (
                    message.type === "subscribe_success" &&
                    hasProperty(message, "data") &&
                    typeof message.data === "object" &&
                    message.data &&
                    hasProperty(message.data, "channel")
                ) {
                    this.currentChannel = message.data.channel;
                }

                // Handle request responses with requestId
                if (hasProperty(message, "requestId") && typeof message.requestId === "number") {
                    const requestId = message.requestId;
                    if (this.pendingRequests.has(requestId)) {
                        const result = hasProperty(message, "data") ? message.data : message;
                        this.pendingRequests.get(requestId)!.resolve(result);
                        this.pendingRequests.delete(requestId);
                    }
                }
            }
        } catch (error) {
            this.emitError(new Error(`Error processing message: ${error instanceof Error ? error.message : String(error)}`));
        }
    }

    /**
     * Sends a request to the server
     */
    async sendRequest(signedRequest: string): Promise<unknown> {
        if (!this.isConnected || !this.ws) {
            throw new Error("WebSocket not connected");
        }

        let requestId: number;

        try {
            const parsedRequest = JSON.parse(signedRequest);

            if (
                !parsedRequest ||
                !parsedRequest.req ||
                !Array.isArray(parsedRequest.req) ||
                parsedRequest.req.length < 2 ||
                typeof parsedRequest.req[0] !== "number" ||
                typeof parsedRequest.req[1] !== "string"
            ) {
                throw new Error("Invalid request format");
            }

            requestId = parsedRequest.req[0];
        } catch (parseError) {
            throw new Error(`Failed to parse request: ${parseError instanceof Error ? parseError.message : String(parseError)}`);
        }

        return new Promise((resolve, reject) => {
            const requestTimeout = setTimeout(() => {
                if (this.pendingRequests.has(requestId)) {
                    this.pendingRequests.delete(requestId);
                    reject(new Error(`Request timeout`));
                }
            }, this.options.requestTimeout);

            this.pendingRequests.set(requestId, {
                resolve: (result: unknown) => {
                    clearTimeout(requestTimeout);
                    resolve(result);
                },
                reject: (error: Error) => {
                    clearTimeout(requestTimeout);
                    reject(error);
                },
            });

            try {
                if (!this.ws) {
                    throw new Error("WebSocket is not initialized");
                }
                this.ws.send(signedRequest);
            } catch (error) {
                clearTimeout(requestTimeout);
                this.pendingRequests.delete(requestId);
                reject(new Error(`Failed to send message: ${error instanceof Error ? error.message : String(error)}`));
            }
        });
    }

    /**
     * Sends a ping to the server
     */
    async ping(): Promise<unknown> {
        return this.sendRequest(await createPingMessage(this.signer.sign));
    }
}

/**
 * Creates a new WebSocket client
 */
export function createWebSocketClient(url: string, signer: WalletSigner, options?: Partial<WebSocketClientOptions>): WebSocketClient {
    return new WebSocketClient(url, signer, {
        autoReconnect: true,
        reconnectDelay: 1000,
        maxReconnectAttempts: 5,
        requestTimeout: 10000,
        ...options,
    });
}
