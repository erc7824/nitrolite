import { type Hex } from "viem";
import { ethers } from "ethers";
import {
    createAuthRequestMessage,
    NitroliteRPC,
    createAuthVerifyMessage,
    createPingMessage,
    createAuthVerifyMessageWithJWT,
} from "@erc7824/nitrolite";
import type { Channel } from "@erc7824/nitrolite";
import { WalletStore } from "../store";

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

/**
 * EIP-712 domain and types for auth_verify challenge
 */
const getAuthDomain = () => {
    return {
        name: "Viper Duel",
    };
}; 

const AUTH_TYPES = {
    Policy: [
        { name: "challenge", type: "string" },
        { name: "scope", type: "string" },
        { name: "wallet", type: "address" },
        { name: "application", type: "address" },
        { name: "participant", type: "address" },
        { name: "expire", type: "uint256" },
        { name: "allowances", type: "Allowance[]" },
    ],
    Allowance: [
        { name: "asset", type: "string" },
        { name: "amount", type: "uint256" },
    ],
};

const expire = String(Math.floor(Date.now() / 1000) + 24 * 60 * 60);

/**
 * Creates EIP-712 signing function for challenge verification with proper challenge extraction
 */
function createEIP712SigningFunction(stateSigner: WalletSigner) {
    const walletClient = WalletStore.getWalletClient();

    if (!walletClient) {
        throw new Error("No wallet client available for EIP-712 signing");
    }

    return async (data: any): Promise<`0x${string}`> => {
        console.log("Signing auth_verify challenge with EIP-712:", data);

        let challengeUUID = "";
        const address = walletClient.account?.address;

        // The data coming in is the array from createAuthVerifyMessage
        // Format: [timestamp, "auth_verify", [{"address": "0x...", "challenge": "uuid"}], timestamp]
        if (Array.isArray(data)) {
            console.log("Data is array, extracting challenge from position [2][0].challenge");

            // Direct array access - data[2] should be the array with the challenge object
            if (data.length >= 3 && Array.isArray(data[2]) && data[2].length > 0) {
                const challengeObject = data[2][0];

                if (challengeObject && challengeObject.challenge) {
                    challengeUUID = challengeObject.challenge;
                    console.log("Extracted challenge UUID from array:", challengeUUID);
                }
            }
        } else if (typeof data === "string") {
            try {
                const parsed = JSON.parse(data);

                console.log("Parsed challenge data:", parsed);

                // Handle different message structures
                if (parsed.res && Array.isArray(parsed.res)) {
                    // auth_challenge response: {"res": [id, "auth_challenge", {"challenge": "uuid"}, timestamp]}
                    if (parsed.res[1] === "auth_challenge" && parsed.res[2]) {
                        challengeUUID = parsed.res[2].challenge_message || parsed.res[2].challenge;
                        console.log("Extracted challenge UUID from auth_challenge:", challengeUUID);
                    }
                    // auth_verify message: [timestamp, "auth_verify", [{"address": "0x...", "challenge": "uuid"}], timestamp]
                    else if (parsed.res[1] === "auth_verify" && Array.isArray(parsed.res[2]) && parsed.res[2][0]) {
                        challengeUUID = parsed.res[2][0].challenge;
                        console.log("Extracted challenge UUID from auth_verify:", challengeUUID);
                    }
                }
                // Direct array format
                else if (Array.isArray(parsed) && parsed.length >= 3 && Array.isArray(parsed[2])) {
                    challengeUUID = parsed[2][0]?.challenge;
                    console.log("Extracted challenge UUID from direct array:", challengeUUID);
                }
            } catch (e) {
                console.error("Could not parse challenge data:", e);
                console.log("Using raw string as challenge");
                challengeUUID = data;
            }
        } else if (data && typeof data === "object") {
            // If data is already an object, try to extract challenge
            challengeUUID = data.challenge || data.challenge_message;
            console.log("Extracted challenge from object:", challengeUUID);
        }

        if (!challengeUUID || challengeUUID.includes("[") || challengeUUID.includes("{")) {
            console.error("Challenge extraction failed or contains invalid characters:", challengeUUID);
            throw new Error("Could not extract valid challenge UUID for EIP-712 signing");
        }

        console.log("Final challenge UUID for EIP-712:", challengeUUID);
        console.log("Signing for address:", address);
        console.log("Auth domain:", getAuthDomain());

        // Create EIP-712 message
        const message = {
            challenge: challengeUUID,
            scope: "app.nitro.aura",
            wallet: address as `0x${string}`,
            application: address as `0x${string}`,
            participant: stateSigner.address as `0x${string}`,
            expire: expire,
            allowances: [],
        };

        console.log("EIP-712 message to sign:", message);

        try {
            // Sign with EIP-712
            const signature = await walletClient.signTypedData({
                account: walletClient.account!,
                domain: getAuthDomain(),
                types: AUTH_TYPES,
                primaryType: "Policy",
                message: message,
            });

            console.log("EIP-712 signature generated for challenge:", signature);
            return signature;
        } catch (eip712Error) {
            console.error("EIP-712 signing failed:", eip712Error);
            console.log("Attempting fallback to regular message signing...");

            try {
                // Fallback to regular message signing if EIP-712 fails
                const fallbackMessage = `Authentication challenge for ${address}: ${challengeUUID}`;

                console.log("Fallback message:", fallbackMessage);

                const fallbackSignature = await walletClient.signMessage({
                    message: fallbackMessage,
                    account: walletClient.account!,
                });

                console.log("Fallback signature generated:", fallbackSignature);
                return fallbackSignature as `0x${string}`;
            } catch (fallbackError) {
                console.error("Fallback signing also failed:", fallbackError);
                throw new Error(`Both EIP-712 and fallback signing failed: ${(eip712Error as Error)?.message}`);
            }
        }
    };
}

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
    private pingInterval: any = null;

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
                        this.startPingInterval();
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
                    this.stopPingInterval();

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
     * Waits for wallet client to be available
     */
    private async waitForWalletClient(timeout: number = 10000): Promise<any> {
        const startTime = Date.now();

        while (Date.now() - startTime < timeout) {
            const walletClient = WalletStore.getWalletClient();
            if (walletClient?.account?.address) {
                return walletClient;
            }

            // Wait 100ms before checking again
            await new Promise((resolve) => setTimeout(resolve, 100));
        }

        throw new Error("Timeout waiting for wallet client to be available");
    }

    /**
     * Authenticates with the WebSocket server
     */
    private async authenticate(): Promise<void> {
        // Wait for wallet client to be available
        const walletClient = await this.waitForWalletClient();
        console.log("Authenticating with wallet client:", walletClient);
        if (!this.ws) throw new Error("WebSocket not connected");

        if (!walletClient?.account?.address) throw new Error("Wallet client not initialized or address not available");

        const privyWalletAddress = walletClient.account.address;

        console.log("Starting authentication with:");
        console.log("- Privy wallet address:", privyWalletAddress);

        // Check for JWT token first
        const jwtToken = typeof window !== "undefined" ? window.localStorage?.getItem("jwtToken") : null;

        let authRequest: string;

        if (jwtToken) {
            console.log("JWT token found, sending auth request with token");
            authRequest = await createAuthVerifyMessageWithJWT(jwtToken);
        } else {
            console.log("No JWT token found, proceeding with challenge-response authentication");
            authRequest = await createAuthRequestMessage({
                wallet: ethers.getAddress(privyWalletAddress) as `0x${string}`, // wallet
                participant: this.signer.address, //session key
                app_name: "Viper Duel",
                expire: expire,
                scope: "app.nitro.aura",
                application: ethers.getAddress(privyWalletAddress) as `0x${string}`,
                allowances: [],
            });
        }

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
                        // walletClient is already available from the authenticate method scope
                        const eip712SigningFunction = createEIP712SigningFunction(this.signer);

                        console.log("Calling createAuthVerifyMessage...");
                        // Create and send verification message with EIP-712 signature
                        const authVerify = await createAuthVerifyMessage(
                            eip712SigningFunction,
                            event.data // Pass the raw challenge response string/object
                        );

                        this.ws?.send(authVerify);
                    } else if (response.res && (response.res[1] === "auth_verify" || response.res[1] === "auth_success")) {
                        console.log("Authentication successful");

                        // If response contains a JWT token, store it
                        if (response.res[2]?.[0]?.["jwt_token"]) {
                            console.log("JWT token received:", response.res[2][0]["jwt_token"]);
                            if (typeof window !== "undefined") {
                                window.localStorage?.setItem("jwtToken", response.res[2][0]["jwt_token"]);
                            }
                        }

                        // Authentication successful
                        const paramsForChannels = [{ participant: ethers.getAddress(privyWalletAddress) as `0x${string}` }];
                        const getChannelsMessage = NitroliteRPC.createRequest(10, "get_channels", paramsForChannels);
                        const getChannelMessage = await NitroliteRPC.signRequestMessage(getChannelsMessage, this.signer.sign);
                        console.log("getChannelMessage", getChannelMessage);
                        this.ws?.send(JSON.stringify(getChannelMessage));
                        clearTimeout(authTimeout);
                        this.ws?.removeEventListener("message", handleAuthResponse);
                        resolve();
                    } else if (response.err || (response.res && response.res[1] === "error")) {
                        // Authentication error
                        const errorMsg = response.err?.[2] || response.error || response.res?.[2]?.[0]?.error || "Authentication failed";
                        console.error("Authentication failed:", errorMsg);
                        if (typeof window !== "undefined") {
                            window.localStorage?.removeItem("jwtToken");
                        }
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
     * Starts ping interval to keep connection alive
     */
    private startPingInterval(): void {
        this.stopPingInterval();
        this.pingInterval = setInterval(async () => {
            if (this.isConnected) {
                try {
                    await this.ping();
                } catch (error) {
                    console.error("Error sending ping:", error);
                }
            }
        }, 20000);
    }

    /**
     * Stops the ping interval
     */
    private stopPingInterval(): void {
        if (this.pingInterval) {
            clearInterval(this.pingInterval);
            this.pingInterval = null;
        }
    }

    /**
     * Closes the WebSocket connection
     */
    close(): void {
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }

        this.stopPingInterval();

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
