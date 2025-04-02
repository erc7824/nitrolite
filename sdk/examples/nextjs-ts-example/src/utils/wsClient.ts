import { Hex } from "viem";
import { NitroliteRPC } from "@erc7824/nitrolite";
import { ethers } from "ethers";
import { WSStatus, Channel } from "@/types";

export interface CryptoKeypair {
    publicKey: string;
    privateKey: string;
    address?: string;
}

export interface WalletSigner {
    publicKey: string;
    address?: string;
    sign: (message: string) => Promise<Hex>;
}

// Utility function to derive Ethereum address from public key
export const getAddressFromPublicKey = (publicKey: string): string => {
    try {
        // Remove '0x' prefix if it exists and make sure it's a compressed public key
        const cleanPublicKey = publicKey.startsWith("0x") ? publicKey.slice(2) : publicKey;

        // Keccak hash of the public key
        const hash = ethers.utils.keccak256("0x" + cleanPublicKey);

        // Take the last 20 bytes of the hash and prefix with '0x' to get the address
        const address = "0x" + hash.slice(-40);

        // Return checksummed address
        return ethers.utils.getAddress(address);
    } catch (error) {
        console.error("Error deriving address from public key:", error);
        return "0x0000000000000000000000000000000000000000";
    }
};

// Create signer from ethers wallet with Keccak support
export const createEthersSigner = (privateKey: string): WalletSigner => {
    try {
        // Create ethers wallet from private key
        const wallet = new ethers.Wallet(privateKey);

        return {
            publicKey: wallet.publicKey,
            address: wallet.address,
            sign: async (message: string): Promise<Hex> => {
                // Hash the message with Keccak before signing for additional security
                const messageHash = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(message));
                const signature = await wallet.signMessage(ethers.utils.arrayify(messageHash));
                return signature as Hex;
            },
        };
    } catch (error) {
        console.error("Error creating ethers signer, using mock instead:", error);
        return createMockSigner();
    }
};

// Generate a random keypair using ethers with Keccak hashing
export const generateKeyPair = async (): Promise<CryptoKeypair> => {
    try {
        // Create random wallet
        const wallet = ethers.Wallet.createRandom();

        // Hash the private key with Keccak256 for additional security
        const privateKeyHash = ethers.utils.keccak256(wallet.privateKey);

        // Derive public key from hashed private key to create a new wallet
        const walletFromHashedKey = new ethers.Wallet(privateKeyHash);

        const address = walletFromHashedKey.address;

        return {
            privateKey: privateKeyHash,
            publicKey: walletFromHashedKey.publicKey,
            address: address,
        };
    } catch (error) {
        console.error("Error generating keypair, using mock instead:", error);
        // Fallback mock implementation
        const randomHex = ethers.utils.randomBytes(32);
        const privateKey = ethers.utils.keccak256(randomHex);
        const wallet = new ethers.Wallet(privateKey);

        return {
            privateKey: privateKey,
            publicKey: wallet.publicKey,
            address: wallet.address,
        };
    }
};

// Mock signer for development using Keccak
export const createMockSigner = (): WalletSigner => {
    try {
        // Create a deterministic but unique mock wallet
        const seed = ethers.utils.id("mock-wallet-" + Date.now().toString());
        const privateKey = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(seed));
        const wallet = new ethers.Wallet(privateKey);

        return {
            publicKey: wallet.publicKey,
            address: wallet.address,
            sign: async (msg: string): Promise<Hex> => {
                // Real signing with the mock wallet
                const messageHash = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(msg));
                const signature = await wallet.signMessage(ethers.utils.arrayify(messageHash));
                return signature as Hex;
            },
        };
    } catch (error) {
        console.error("Error creating mock signer:", error);
        // Very basic fallback if ethers fails
        const randomKey = "0x" + Array.from({ length: 42 }, (_, i) => (i === 0 ? "" : Math.floor(Math.random() * 16).toString(16))).join("");
        const mockAddress = "0x" + Array.from({ length: 40 }, () => Math.floor(Math.random() * 16).toString(16)).join("");
        return {
            publicKey: randomKey,
            address: mockAddress,
            sign: async (msg: string): Promise<Hex> => {
                // Use keccak for consistent approach even in fallback
                try {
                    const mockSignature = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(msg + randomKey));
                    return mockSignature as Hex;
                } catch (e) {
                    return ("0x" + Array.from({ length: 130 }, () => Math.floor(Math.random() * 16).toString(16)).join("")) as Hex;
                }
            },
        };
    }
};

// Default mock signer
export const mockSigner = createMockSigner();

export enum WebSocketReadyState {
    CONNECTING = 0,
    OPEN = 1,
    CLOSING = 2,
    CLOSED = 3,
}

export class WebSocketClient {
    private ws: WebSocket | null = null;
    private pendingRequests = new Map<number, { resolve: Function; reject: Function }>();
    private reconnectTimeout: NodeJS.Timeout | null = null;
    private reconnectAttempts = 0;
    private onStatusChangeCallback?: (status: WSStatus) => void;
    private onMessageCallback?: (message: any) => void;
    private onErrorCallback?: (error: Error) => void;
    private currentChannel: Channel | null = null;

    constructor(
        private url: string,
        private signer: WalletSigner,
        private options = {
            autoReconnect: true,
            reconnectDelay: 1000,
            maxReconnectAttempts: 5,
            requestTimeout: 10000,
        }
    ) {}

    onStatusChange(cb: (status: WSStatus) => void): void {
        this.onStatusChangeCallback = cb;
    }
    onMessage(cb: (message: any) => void): void {
        this.onMessageCallback = cb;
    }
    onError(cb: (error: Error) => void): void {
        this.onErrorCallback = cb;
    }
    get readyState(): WebSocketReadyState {
        return this.ws ? this.ws.readyState : WebSocketReadyState.CLOSED;
    }
    get isConnected(): boolean {
        return this.ws !== null && this.ws.readyState === WebSocketReadyState.OPEN;
    }
    get currentSubscribedChannel(): Channel | null {
        return this.currentChannel;
    }

    getShortenedPublicKey(): string {
        return this.signer.publicKey.substring(0, 8) + "..." + this.signer.publicKey.substring(this.signer.publicKey.length - 4);
    }

    connect(): Promise<void> {
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }

        return new Promise((resolve, reject) => {
            try {
                if (this.isConnected) return resolve();

                this.ws = new WebSocket(this.url);
                this.onStatusChangeCallback?.("connecting");

                this.ws.onopen = async () => {
                    try {
                        this.onStatusChangeCallback?.("authenticating");
                        await this.authenticate();
                        this.onStatusChangeCallback?.("connected");
                        this.reconnectAttempts = 0;
                        resolve();
                    } catch (error) {
                        this.onStatusChangeCallback?.("auth_failed");
                        this.onErrorCallback?.(error instanceof Error ? error : new Error(String(error)));
                        reject(error);
                        this.close();
                        this.handleReconnect();
                    }
                };

                this.ws.onmessage = (event) => {
                    let response;

                    try {
                        response = JSON.parse(event.data);
                    } catch (error) {
                        console.error("Error parsing message:", error);
                        console.log("Raw message:", event.data);
                        // Notify about message parsing error but don't break the connection
                        this.onErrorCallback?.(new Error("Failed to parse server message"));
                        return;
                    }

                    try {
                        // Notify callback about received message
                        this.onMessageCallback?.(response);

                        // Handle standard NitroRPC responses
                        if (response.res) {
                            const requestId = response.res[0];
                            if (this.pendingRequests.has(requestId)) {
                                this.pendingRequests.get(requestId)!.resolve(response.res[2]);
                                this.pendingRequests.delete(requestId);
                            }
                        }
                        // Handle error responses
                        else if (response.err) {
                            const requestId = response.err[0];
                            if (this.pendingRequests.has(requestId)) {
                                this.pendingRequests.get(requestId)!.reject(new Error(`Error ${response.err[1]}: ${response.err[2]}`));
                                this.pendingRequests.delete(requestId);
                            }
                        }
                        // Handle legacy/custom responses
                        else if (response.type) {
                            if (response.type === "auth_success") {
                                // Authentication handled separately
                            } else if (response.type === "subscribe_success" && response.data?.channel) {
                                this.currentChannel = response.data.channel as Channel;
                            }

                            // For all other responses with a requestId, resolve any pending requests
                            const requestId = response.requestId;
                            if (requestId && this.pendingRequests.has(requestId)) {
                                this.pendingRequests.get(requestId)!.resolve(response.data || response);
                                this.pendingRequests.delete(requestId);
                            }
                        }
                    } catch (error) {
                        console.error("Error handling message:", error);
                        this.onErrorCallback?.(new Error(`Error processing message: ${error instanceof Error ? error.message : String(error)}`));
                    }
                };

                this.ws.onerror = () => {
                    this.onErrorCallback?.(new Error("WebSocket connection error"));
                    reject(new Error("WebSocket connection error"));
                };

                this.ws.onclose = () => {
                    this.onStatusChangeCallback?.("disconnected");
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

    private handleReconnect(): void {
        if (!this.options.autoReconnect || this.reconnectAttempts >= this.options.maxReconnectAttempts) {
            if (this.reconnectAttempts >= this.options.maxReconnectAttempts) {
                this.onStatusChangeCallback?.("reconnect_failed");
            }
            return;
        }

        if (this.reconnectTimeout) clearTimeout(this.reconnectTimeout);

        this.reconnectAttempts++;
        const delay = this.options.reconnectDelay * Math.pow(1.5, this.reconnectAttempts - 1);
        this.onStatusChangeCallback?.("reconnecting");

        this.reconnectTimeout = setTimeout(() => {
            this.connect().catch(() => {});
        }, delay);
    }

    private async authenticate(): Promise<void> {
        if (!this.ws) throw new Error("WebSocket not connected");

        const authRequest = NitroliteRPC.createRequest("auth", [this.signer.publicKey]);
        const signedAuthRequest = await NitroliteRPC.signMessage(authRequest, this.signer.sign);

        return new Promise((resolve, reject) => {
            if (!this.ws) return reject(new Error("WebSocket not connected"));

            const authTimeout = setTimeout(() => {
                this.ws?.removeEventListener("message", handleAuthResponse);
                reject(new Error("Authentication timeout"));
            }, this.options.requestTimeout);

            const handleAuthResponse = (event: MessageEvent) => {
                let response;

                try {
                    response = JSON.parse(event.data);
                } catch (error) {
                    console.error("Error parsing auth response:", error);
                    console.log("Raw auth message:", event.data);
                    return; // Continue waiting for valid responses
                }

                try {
                    if ((response.res && response.res[1] === "auth") || response.type === "auth_success") {
                        clearTimeout(authTimeout);
                        this.ws!.removeEventListener("message", handleAuthResponse);
                        resolve();
                    } else if ((response.err && response.err[1]) || response.type === "auth_error") {
                        clearTimeout(authTimeout);
                        this.ws!.removeEventListener("message", handleAuthResponse);
                        const errorMsg = response.err ? response.err[2] : response.error || "Authentication failed";
                        reject(new Error(errorMsg));
                    }
                } catch (error) {
                    console.error("Error handling auth response:", error);
                    clearTimeout(authTimeout);
                    this.ws!.removeEventListener("message", handleAuthResponse);
                    reject(new Error(`Authentication error: ${error instanceof Error ? error.message : String(error)}`));
                }
            };

            this.ws.addEventListener("message", handleAuthResponse);
            this.ws.send(JSON.stringify(signedAuthRequest));
        });
    }

    async sendRequest(method: string, params: any[] = []): Promise<any> {
        if (!this.isConnected) throw new Error("WebSocket not connected");
        return this.sendSignedRequest(NitroliteRPC.createRequest(method, params));
    }

    // Channel subscription
    async subscribe(channel: Channel): Promise<void> {
        if (!this.isConnected) throw new Error("WebSocket not connected");

        const request = NitroliteRPC.createRequest("subscribe", [channel]);
        await this.sendSignedRequest(request);
        this.currentChannel = channel;
    }

    // Send message to a channel
    async publishMessage(message: string): Promise<void> {
        if (!this.isConnected) throw new Error("WebSocket not connected");
        if (!this.currentChannel) throw new Error("Not subscribed to any channel");

        const shortenedKey = this.getShortenedPublicKey();
        // Use createRequest instead of the missing createPublishRequest
        const request = NitroliteRPC.createRequest("publish", [this.currentChannel, message, shortenedKey]);
        await this.sendRequestDirect(await NitroliteRPC.signMessage(request, this.signer.sign));
    }

    // Send a ping request
    async ping(): Promise<any> {
        return this.sendSignedRequest(NitroliteRPC.createRequest("ping", []));
    }

    // Check balance
    async checkBalance(tokenAddress: string = "0xSHIB..."): Promise<any> {
        return this.sendSignedRequest(NitroliteRPC.createRequest("balance", [tokenAddress]));
    }

    // Helper method to sign and send a request
    private async sendSignedRequest(request: any): Promise<any> {
        const signedRequest = await NitroliteRPC.signMessage(request, this.signer.sign);
        return this.sendRequestDirect(signedRequest);
    }

    // Helper method to send a pre-constructed request
    private async sendRequestDirect(signedRequest: any): Promise<any> {
        if (!this.isConnected) throw new Error("WebSocket not connected");

        return new Promise((resolve, reject) => {
            const requestId = signedRequest.req[0];
            const timeout = setTimeout(() => {
                if (this.pendingRequests.has(requestId)) {
                    this.pendingRequests.delete(requestId);
                    reject(new Error(`Request timeout: ${signedRequest.req[1]}`));
                }
            }, this.options.requestTimeout);

            this.pendingRequests.set(requestId, {
                resolve: (result: any) => {
                    clearTimeout(timeout);
                    resolve(result);
                },
                reject: (error: Error) => {
                    clearTimeout(timeout);
                    reject(error);
                },
            });

            this.ws!.send(JSON.stringify(signedRequest));
        });
    }

    async sendBatch(requests: { method: string; params: any[] }[]): Promise<any[]> {
        return Promise.all(requests.map((req) => this.sendRequest(req.method, req.params)));
    }

    close(): void {
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }

        if (this.ws && [WebSocketReadyState.OPEN, WebSocketReadyState.CONNECTING].includes(this.ws.readyState)) {
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
        this.onStatusChangeCallback?.("disconnected");
    }
}

export const createWebSocketClient = (url: string, signer: WalletSigner, options?: any) => new WebSocketClient(url, signer, options);
