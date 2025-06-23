/**
 * Nitrolite RPC (WebSocket) client
 * This file handles all WebSocket communication with Nitrolite server
 */
import { 
    createAuthRequestMessage, 
    createAuthVerifyMessage, 
    createEIP712AuthMessageSigner,
    createPingMessage, 
    NitroliteRPC,
    parseRPCResponse
} from "@erc7824/nitrolite";
import dotenv from "dotenv";
import { ethers } from "ethers";
import WebSocket from "ws";

import logger from "../utils/logger.js";

import { getWalletClient } from "./nitroliteOnChain.js";

/**
 * EIP-712 domain for auth_verify challenge
 */
const getAuthDomain = () => {
    return {
        name: "Viper Duel",
    };
};

const expire = String(Math.floor(Date.now() / 1000) + 24 * 60 * 60);

// Load environment variables
dotenv.config();

// Connection status
export const WSStatus = {
    CONNECTED: "connected",
    CONNECTING: "connecting",
    DISCONNECTED: "disconnected",
    RECONNECTING: "reconnecting",
    RECONNECT_FAILED: "reconnect_failed",
    AUTH_FAILED: "auth_failed",
    AUTHENTICATING: "authenticating",
};

// Server-side WebSocket client with authentication
export class NitroliteRPCClient {
    constructor(url, privateKey) {
        this.url = url;
        this.privateKey = privateKey;
        this.ws = null;
        this.status = WSStatus.DISCONNECTED;
        this.channel = null;
        this.wallet = new ethers.Wallet(privateKey);
        this.address = this.wallet.address;
        this.pendingRequests = new Map();
        this.nextRequestId = 1;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
        this.reconnectTimeout = null;
        this.onMessageCallbacks = [];
        this.onStatusChangeCallbacks = [];
        this.walletClient = null;

        logger.system(`RPC client initialized with address: ${this.address}`);
    }

    // Register message callback
    onMessage(callback) {
        this.onMessageCallbacks.push(callback);
    }

    // Register status change callback
    onStatusChange(callback) {
        this.onStatusChangeCallbacks.push(callback);
    }

    // Connect to WebSocket server
    async connect() {
        if (this.status === WSStatus.CONNECTED || this.status === WSStatus.CONNECTING) {
            logger.ws("Already connected or connecting...");
            return;
        }

        try {
            logger.ws(`Connecting to ${this.url}...`);
            this.setStatus(WSStatus.CONNECTING);

            this.ws = new WebSocket(this.url);

            this.ws.on("open", async () => {
                logger.ws("WebSocket connection established");
                this.setStatus(WSStatus.AUTHENTICATING);
                try {
                    await this.authenticate();
                    logger.auth("Successfully authenticated with the WebSocket server");
                    this.reconnectAttempts = 0;
                    this.startPingInterval();
                } catch (error) {
                    logger.error("Authentication failed:", error);
                    this.setStatus(WSStatus.AUTH_FAILED);
                    this.ws.close();
                }
            });

            this.ws.on("message", (data) => {
                this.handleMessage(data);
            });

            this.ws.on("error", (error) => {
                logger.error("WebSocket error:", error);
            });

            this.ws.on("close", () => {
                logger.ws("WebSocket connection closed");
                this.setStatus(WSStatus.DISCONNECTED);
                clearInterval(this.pingInterval);
                this.handleReconnect();
            });
        } catch (error) {
            logger.error("Failed to connect:", error);
            this.setStatus(WSStatus.DISCONNECTED);
            this.handleReconnect();
        }
    }

    // Update status and notify listeners
    setStatus(status) {
        const prevStatus = this.status;
        this.status = status;
        logger.ws(`Status changed: ${prevStatus} -> ${status}`);
        this.onStatusChangeCallbacks.forEach((callback) => callback(status));
    }

    // Sign message function for non-auth requests
    async signMessage(data) {
        const messageStr = typeof data === "string" ? data : JSON.stringify(data);
        const digestHex = ethers.id(messageStr);
        const messageBytes = ethers.getBytes(digestHex);
        const { serialized: signature } = this.wallet.signingKey.sign(messageBytes);
        return signature;
    }

    /**
     * Authenticates with the WebSocket server using:
     * 1. auth_request: empty signature with wallet address
     * 2. auth_verify: EIP-712 signature for challenge verification
     *
     * @param timeout - Timeout in milliseconds for the entire process
     * @returns A Promise that resolves when authenticated
     */
    async authenticate(timeout = 10000) {
        if (!this.ws) {
            throw new Error("WebSocket not connected");
        }

        logger.auth("Starting authentication with SDK 0.2.11 flow...");
        logger.auth("- Wallet address:", this.address);
        logger.auth("- EIP-712 signature for auth_verify challenge");

        const authMessage = {
            wallet: this.address,
            participant: this.address,
            app_name: "Viper Duel",
            expire: expire, // 24 hours in seconds
            scope: "console",
            application: this.address,
            allowances: [],
        };

        return new Promise((resolve, reject) => {
            let authTimeoutId = null;

            const cleanup = () => {
                if (authTimeoutId) {
                    clearTimeout(authTimeoutId);
                    authTimeoutId = null;
                }
                this.ws.removeEventListener("message", handleAuthResponse);
            };

            const resetTimeout = () => {
                if (authTimeoutId) {
                    clearTimeout(authTimeoutId);
                }
                authTimeoutId = setTimeout(() => {
                    cleanup();
                    reject(new Error("Authentication timeout"));
                }, timeout);
            };

            const handleAuthResponse = async (event) => {
                const data = event.data || event;
                
                try {
                    const response = parseRPCResponse(data);

                    // Check for challenge response: {"res": [id, "auth_challenge", {"challenge": "uuid"}, timestamp]}
                    if (response.method === "auth_challenge") {
                        logger.auth("Received auth_challenge, preparing EIP-712 auth_verify...");
                        resetTimeout(); // Reset timeout while we process and send verify

                        try {
                            logger.auth("Creating EIP-712 signing function...");
                            
                            // Ensure we have a wallet client for EIP-712 signing
                            if (!this.walletClient) {
                                logger.auth("Initializing wallet client for EIP-712 signing...");
                                this.walletClient = await getWalletClient(this.privateKey);
                            }
                            
                            const eip712SigningFunction = createEIP712AuthMessageSigner(
                                this.walletClient,
                                {
                                    scope: authMessage.scope,
                                    application: authMessage.application,
                                    participant: authMessage.participant,
                                    expire: authMessage.expire,
                                    allowances: authMessage.allowances.map((allowance) => ({
                                        asset: allowance.symbol || allowance.asset,
                                        amount: allowance.amount.toString(),
                                    })),
                                },
                                getAuthDomain(),
                            );

                            logger.auth("Calling createAuthVerifyMessage...");
                            const authVerify = await createAuthVerifyMessage(eip712SigningFunction, response);

                            logger.auth("Sending auth_verify with EIP-712 signature");
                            this.ws.send(authVerify);
                            logger.auth("auth_verify sent successfully");
                        } catch (eip712Error) {
                            logger.error("Error creating EIP-712 auth_verify:", eip712Error);
                            logger.error("Error stack:", eip712Error.stack);

                            cleanup();
                            reject(new Error(`EIP-712 auth_verify failed: ${eip712Error.message}`));
                            return;
                        }
                    }
                    // Check for success response
                    else if (response.method === "auth_verify" && response.params.success) {
                        logger.auth("Authentication successful");

                        cleanup();
                        
                        // Set status to connected
                        this.setStatus(WSStatus.CONNECTED);

                        try {
                            // Request channel information for our address and check if we
                            // need to create one
                            const channels = await this.getChannelInfo();
                            // Check if we have valid channels
                            const hasValidChannel = channels && Array.isArray(channels) && channels.length > 0 && channels[0] !== null;

                            if (!hasValidChannel) {
                                logger.nitro("No valid channels found after authentication, will create one");
                            }
                        } catch (error) {
                            logger.error("Failed to get channel info, continuing anyway:", error);
                        }

                        resolve();
                    }
                    // Check for error response
                    else if (response.method === "error") {
                        const errorMsg = response.params.error || "Authentication failed";

                        logger.error("Authentication failed:", errorMsg);
                        cleanup();
                        reject(new Error(String(errorMsg)));
                    } else {
                        logger.auth("Received non-auth message during auth, continuing to listen:", response);
                        // Keep listening if it wasn't a final success/error
                    }
                } catch (error) {
                    // Ignore non-auth methods during authentication
                    if (error.message && error.message.includes("Unknown method:")) {
                        logger.auth("Ignoring non-auth message during authentication:", error.message);
                        return;
                    }
                    
                    logger.error("Error handling auth response:", error);
                    logger.error("Error stack:", error.stack);
                    cleanup();
                    reject(new Error(`Authentication error: ${error instanceof Error ? error.message : String(error)}`));
                }
            };

            // Step 1: Send auth_request
            const sendAuthRequest = async () => {
                try {
                    logger.auth("Sending auth_request...");
                    const authRequest = await createAuthRequestMessage(authMessage);
                    this.ws.send(authRequest);
                    logger.auth("auth_request sent successfully");
                } catch (requestError) {
                    logger.error("Error creating auth_request:", requestError);
                    cleanup();
                    reject(new Error(`Failed to create auth_request: ${requestError.message}`));
                }
            };

            this.ws.addEventListener("message", handleAuthResponse);
            resetTimeout(); // Start the initial timeout

            // Start authentication process
            sendAuthRequest();
        });
    }

    // Handle incoming WebSocket messages
    handleMessage(data) {
        try {
            // Ensure data is properly handled as string
            const rawData = typeof data === "string" ? data : data.toString();
            const message = JSON.parse(rawData);
            logger.data("Received message", message);

            // Notify callbacks first to allow for authentication handling
            this.onMessageCallbacks.forEach((callback) => callback(message));

            // Handle response to pending requests
            if (message.res && Array.isArray(message.res) && message.res.length >= 3) {
                const requestId = message.res[0];
                if (this.pendingRequests.has(requestId)) {
                    const { resolve } = this.pendingRequests.get(requestId);
                    resolve(message.res[2]);
                    this.pendingRequests.delete(requestId);
                }
            }

            // Handle errors
            if (message.err && Array.isArray(message.err) && message.err.length >= 3) {
                const requestId = message.err[0];
                if (this.pendingRequests.has(requestId)) {
                    const { reject } = this.pendingRequests.get(requestId);
                    reject(new Error(`Error ${message.err[1]}: ${message.err[2]}`));
                    this.pendingRequests.delete(requestId);
                }
            }

            // Handle channel-specific messages
            if (message.type === "channel_created") {
                logger.nitro("Channel created successfully");
                logger.data("Channel data", message.channel);
                this.channel = message.channel;
            }
        } catch (error) {
            logger.error("Error handling message:", error);
        }
    }

    // Send a request to the WebSocket server
    async sendRequest(method, params = {}) {
        if (!this.ws) {
            throw new Error("WebSocket instance not initialized");
        }

        if (this.ws.readyState !== WebSocket.OPEN) {
            logger.error(`WebSocket not in OPEN state. Current state: ${this.ws.readyState}, Status: ${this.status}`);
            throw new Error(`WebSocket not in OPEN state. Current readyState: ${this.ws.readyState}`);
        }

        if (this.status !== WSStatus.CONNECTED) {
            logger.warn(`WebSocket status is ${this.status}, should be ${WSStatus.CONNECTED}. Proceeding anyway.`);
            if (this.status === WSStatus.AUTHENTICATING) {
                logger.system("Fixing status to CONNECTED for authenticated connection");
                this.setStatus(WSStatus.CONNECTED);
            }
        }

        const requestId = this.nextRequestId++;
        const sign = this.signMessage.bind(this);

        return new Promise(async (resolve, reject) => {
            try {
                const request = NitroliteRPC.createRequest(requestId, method, params);
                const signedRequest = await NitroliteRPC.signRequestMessage(request, sign);

                logger.ws(`Sending request: ${JSON.stringify(signedRequest).slice(0, 100)}...`);

                this.pendingRequests.set(requestId, { resolve, reject });

                setTimeout(() => {
                    if (this.pendingRequests.has(requestId)) {
                        this.pendingRequests.delete(requestId);
                        reject(new Error("Request timeout"));
                    }
                }, 10000);

                this.ws.send(typeof signedRequest === "string" ? signedRequest : JSON.stringify(signedRequest));
            } catch (error) {
                logger.error("Error sending request:", error);
                this.pendingRequests.delete(requestId);
                reject(error);
            }
        });
    }

    // Start ping interval to keep connection alive
    startPingInterval() {
        clearInterval(this.pingInterval);
        this.pingInterval = setInterval(async () => {
            if (this.status === WSStatus.CONNECTED) {
                try {
                    const sign = this.signMessage.bind(this);
                    const pingMessage = await createPingMessage(sign);
                    this.ws.send(pingMessage);
                } catch (error) {
                    logger.error("Error sending ping:", error);
                }
            }
        }, 30000);
    }

    // Handle reconnection
    handleReconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            logger.ws("Maximum reconnect attempts reached");
            this.setStatus(WSStatus.RECONNECT_FAILED);
            return;
        }

        this.reconnectAttempts++;
        const delay = this.reconnectDelay * this.reconnectAttempts;

        logger.ws(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
        this.setStatus(WSStatus.RECONNECTING);

        clearTimeout(this.reconnectTimeout);
        this.reconnectTimeout = setTimeout(() => {
            this.connect();
        }, delay);
    }

    // Close connection
    close() {
        clearInterval(this.pingInterval);
        clearTimeout(this.reconnectTimeout);

        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }

        logger.ws("WebSocket connection closed manually");
        this.setStatus(WSStatus.DISCONNECTED);
    }

    // Get channel information
    async getChannelInfo() {
        try {
            logger.nitro("Requesting channel information...");
            const response = await this.sendRequest("get_channels", [{ participant: this.address }]);
            logger.data("Channel info received", response);

            logger.system("Debug - Raw channel response:");
            logger.system(`- response type: ${typeof response}`);
            logger.system(`- is array: ${Array.isArray(response)}`);
            logger.system(`- stringified: ${JSON.stringify(response)}`);

            let channels = response;

            if (Array.isArray(response) && response.length === 1 && response[0] === null) {
                logger.system("Debug - Got array with single null item");
            }

            if (channels && Array.isArray(channels) && channels.length > 0 && channels[0] !== null) {
                logger.nitro(`Found ${channels.length} valid existing channels`);
                this.channel = channels[0];
                return channels;
            }

            logger.nitro("No valid channels found");

            if (!this.walletClient) {
                logger.nitro("Getting wallet client...");
                this.walletClient = await getWalletClient(this.privateKey);
            }

            return [];
        } catch (error) {
            logger.error("Error getting channel info:", error);
            throw error;
        }
    }
}

// Initialize and export the client instance
let rpcClient = null;

export async function initializeRPCClient() {
    logger.system("Initializing Nitrolite RPC client...");
    if (rpcClient) {
        logger.system("Nitrolite RPC client already initialized, returning existing instance...");
        return rpcClient;
    }

    try {
        logger.system("Initializing new Nitrolite RPC client...");

        if (!process.env.SERVER_PRIVATE_KEY) {
            throw new Error("SERVER_PRIVATE_KEY environment variable is not set");
        }

        if (!process.env.WS_URL) {
            throw new Error("WS_URL environment variable is not set");
        }

        rpcClient = new NitroliteRPCClient(process.env.WS_URL, process.env.SERVER_PRIVATE_KEY);

        rpcClient.onMessage((message) => {
            logger.ws("RPC Message:", JSON.stringify(message, null, 2));
        });

        rpcClient.onStatusChange((status) => {
            logger.ws("RPC Status changed:", status);
        });

        // Initialize wallet client before connecting (needed for authentication)
        rpcClient.walletClient = await getWalletClient(process.env.SERVER_PRIVATE_KEY);
        
        await rpcClient.connect();

        logger.system("Checking for existing channels...");
        const channels = await rpcClient.getChannelInfo();

        const hasValidChannel = channels && Array.isArray(channels) && channels.length > 0 && channels[0] !== null;

        if (hasValidChannel) {
            logger.nitro(`Found ${channels.length} existing valid channels`);
            logger.data("Channel data", channels[0]);
            rpcClient.channel = channels[0];
        } else {
            logger.nitro("No valid channels found in initializeRPCClient");
        }
    } catch (error) {
        logger.error("Error during RPC client initialization:", error);
    }

    return rpcClient;
}

export function getRPCClient() {
    return rpcClient;
}
