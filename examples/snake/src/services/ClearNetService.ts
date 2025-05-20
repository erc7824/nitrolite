import {
    NitroliteClient,
    createAuthRequestMessage,
    createAuthVerifyMessage,
    createGetLedgerBalancesMessage,
    type NitroliteClientConfig,
} from "@erc7824/nitrolite";
import { BROKER_WS_URL } from "../config";
import { createEthersSigner, generateKeyPair } from "../crypto";
import type { Hex } from "viem";

export interface ChannelData {
    channelId: string;
    state: any;
}

class ClearNetService {
    public client!: NitroliteClient;
    public config!: NitroliteClientConfig;
    private isConnected = false;
    private currentAddress: string | null = null;
    private activeChannel: ChannelData | null = null;
    private wsConnection: WebSocket | null = null;
    private readonly wsUrl = BROKER_WS_URL;
    private pendingRequests = new Map<
        string,
        {
            resolve: (value: any) => void;
            reject: (reason: Error) => void;
            timeout: NodeJS.Timeout;
        }
    >();
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private reconnectDelay = 1000;
    private reconnectTimeout: NodeJS.Timeout | null = null;
    private authenticationInProgress: Promise<void> | null = null;

    constructor() {
        // Try to restore channel from localStorage on initialization
        this.restoreChannelFromStorage();
    }

    public restoreChannelFromStorage(): void {
        try {
            const channelId = localStorage.getItem("nitro_channel_id");
            const channelState = localStorage.getItem("nitro_channel_state");

            if (channelId && channelState) {
                this.activeChannel = {
                    channelId,
                    state: JSON.parse(channelState, (_, value) => {
                        // Handle bigint values stored as strings
                        if (typeof value === 'string' && value.endsWith('n')) {
                            return BigInt(value.slice(0, -1));
                        }
                        return value;
                    })
                };
                console.log("Restored channel from storage:", this.activeChannel);
            }
        } catch (error) {
            console.error("Failed to restore channel from storage:", error);
            // Clear potentially corrupted storage
            this.clearChannelStorage();
        }
    }

    private clearChannelStorage() {
        try {
            localStorage.removeItem("nitro_channel_id");
            localStorage.removeItem("nitro_channel_state");
        } catch (error) {
            console.error("Failed to clear channel storage:", error);
        }
    }

    async initialize(config: NitroliteClientConfig): Promise<boolean> {
        try {
            // Validate the config
            if (!config) {
                throw new Error("Config object is required");
            }

            // Check for required config properties
            if (!config.walletClient) {
                throw new Error("walletClient is required in config");
            }

            if (!config.walletClient.account || !config.walletClient.account.address) {
                throw new Error("walletClient.account.address is required");
            }

            console.log("Initializing with wallet address:", config.walletClient.account.address);

            // Store the config for later use with wallet client
            this.config = config;

            // Initialize the Nitrolite client
            this.client = new NitroliteClient(config);
            console.log("Nitrolite client initialized", this.client);
            this.currentAddress = config.walletClient.account.address;
            console.log("Current wallet client address:", this.currentAddress);

            // Initialize WebSocket connection to ClearNet
            console.log("Initializing WebSocket connection...");
            await this.initializeWebSocket();

            this.isConnected = true;
            console.log("ClearNet client initialized successfully");
            return true;
        } catch (error) {
            console.error("Failed to initialize ClearNet client:", error);
            throw error; // Throw the error instead of returning false
        }
    }

    private initializeWebSocket(): Promise<void> {
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }

        return new Promise((resolve, reject) => {
            try {
                if (this.wsConnection && this.wsConnection.readyState === WebSocket.OPEN) {
                    return resolve();
                }

                // Check if wallet address is available
                if (!this.currentAddress) {
                    console.error("Cannot initialize WebSocket: No wallet address available");
                    return reject(new Error("No wallet address available"));
                }

                // Check ethereum provider availability
                const { ethereum } = window as any;
                if (!ethereum) {
                    console.error("Cannot initialize WebSocket: No ethereum provider found");
                    return reject(new Error("No ethereum provider found"));
                }

                console.log("Creating WebSocket connection to:", this.wsUrl);
                this.wsConnection = new WebSocket(this.wsUrl);

                let connectTimeout = setTimeout(() => {
                    console.error("WebSocket connection timeout");
                    reject(new Error("WebSocket connection timeout"));
                }, 10000);

                this.wsConnection.onopen = async () => {
                    clearTimeout(connectTimeout);
                    console.log("WebSocket connection established");

                    try {
                        // Log wallet client details for debugging
                        console.log("Wallet client account:", this.config?.walletClient?.account);
                        console.log("Current address:", this.currentAddress);

                        // Authenticate with the broker
                        await this.authenticateWithBroker();
                        this.isConnected = true;
                        this.reconnectAttempts = 0;
                        resolve();
                    } catch (error) {
                        console.error("Authentication failed:", error);
                        this.wsConnection?.close();
                        reject(error);
                    }
                };

                this.wsConnection.onerror = (error) => {
                    console.error("WebSocket connection error:", error);
                    reject(error);
                };

                this.wsConnection.onclose = () => {
                    console.log("WebSocket connection closed");
                    this.isConnected = false;
                    this.handleReconnect();
                };

                this.wsConnection.onmessage = (event) => {
                    try {
                        const message = JSON.parse(event.data);
                        this.handleWebSocketMessage(message);
                    } catch (error) {
                        console.error("Error parsing WebSocket message:", error);
                    }
                };
            } catch (error) {
                console.error("Error initializing WebSocket:", error);
                reject(error);
            }
        });
    }

    private handleReconnect(): void {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.log("Max reconnect attempts reached");
            return;
        }

        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
        }

        this.reconnectAttempts++;
        const delay = this.reconnectDelay * Math.pow(1.5, this.reconnectAttempts - 1);

        console.log(`Reconnecting in ${delay}ms...`);

        this.reconnectTimeout = setTimeout(() => {
            this.initializeWebSocket().catch(() => {
                console.log("Reconnect attempt failed");
            });
        }, delay);
    }

    private async authenticateWithBroker(): Promise<void> {
        // If authentication is already in progress, return the existing promise
        if (this.authenticationInProgress) {
            console.log("Authentication already in progress, reusing existing authentication flow");
            return this.authenticationInProgress;
        }

        if (!this.wsConnection || this.wsConnection.readyState !== WebSocket.OPEN) {
            throw new Error("WebSocket not connected");
        }

        /**
         * Gets or creates a wallet signer with a private key stored in localStorage
         */
        let keyPair = null;
        const savedKeys = localStorage.getItem("crypto_keypair");

        if (savedKeys) {
            try {
                keyPair = JSON.parse(savedKeys);
            } catch (error) {
                keyPair = null;
            }
        }

        if (!keyPair) {
            keyPair = await generateKeyPair();
            if (typeof window !== "undefined") {
                localStorage.setItem("crypto_keypair", JSON.stringify(keyPair));
            }
        }

        const signer = createEthersSigner(keyPair.privateKey);

        // Create a new authentication promise and store it
        const authPromise = new Promise<void>((resolve, reject) => {
            let authTimeout: NodeJS.Timeout;

            // Create a one-time message handler for authentication
            const authMessageHandler = async (event: MessageEvent) => {
                try {
                    const message = JSON.parse(event.data);
                    console.log("Auth process message received:", message);

                    // Check for auth_challenge response
                    if (message.res && message.res[1] === "auth_challenge") {
                        console.log("Received auth_challenge, preparing auth_verify...");
                        // Log the exact structure of the response to help us debug
                        console.log("Challenge full response:", message);
                        console.log("Challenge res array:", message.res);
                        console.log("Challenge data:", message.res[2]);

                        try {
                            // Let's try to extract the challenge directly from the raw response
                            const rawData = event.data;
                            console.log("Raw challenge response:", rawData);

                            // Extract the challenge from the response - more safely
                            let challenge = null;
                            const responseData = message.res[2];

                            if (Array.isArray(responseData) && responseData.length > 0) {
                                if (typeof responseData[0] === "object") {
                                    // Try both challenge and challenge_message fields
                                    challenge = responseData[0]?.challenge || responseData[0]?.challenge_message;
                                } else if (typeof responseData[0] === "string") {
                                    challenge = responseData[0];
                                }
                            } else if (typeof responseData === "object") {
                                challenge = responseData.challenge || responseData.challenge_message;
                            } else if (typeof responseData === "string") {
                                challenge = responseData;
                            }

                            console.log("Extracted challenge:", challenge);

                            if (!challenge) {
                                throw new Error("No challenge received in auth_challenge response");
                            }

                            console.log("Challenge received:", challenge);

                            // Use raw challenge data directly for verification
                            console.log("Using raw challenge response for verification");

                            // Pass the raw challenge response to createAuthVerifyMessage
                            // This should match what the server is doing
                            const authVerify = await createAuthVerifyMessage(
                                signer.sign,
                                rawData, // Raw challenge response as a string
                                signer.address
                            );

                            console.log("Auth verify created by nitrolite:", authVerify);
                            console.log("Sending auth_verify:", authVerify);
                            this.wsConnection?.send(authVerify);

                            setTimeout(async () => {
                                const nitroChannelId = localStorage.getItem("nitro_channel_id");

                                if (nitroChannelId) {
                                    // Send get_balances message to the broker
                                    const getBalancesMsg = await createGetLedgerBalancesMessage(signer.sign, nitroChannelId as Hex);

                                    this.wsConnection?.send(getBalancesMsg);
                                }
                            }, 2000);
                        } catch (error) {
                            console.error("Error creating auth verify message:", error);
                            cleanup();
                            reject(new Error("Failed to create auth verify message"));
                        }
                    }
                    // Check for auth_verify success response
                    else if (message.res && message.res[1] === "auth_verify") {
                        console.log("Authentication successful");
                        cleanup();
                        resolve();
                    }
                    // Check for error responses
                    else if (message.res && message.res[1] === "error") {
                        const errorMessage = message.res[2] && message.res[2][0]?.error ? message.res[2][0].error : "Unknown authentication error";
                        console.error("Authentication error:", errorMessage);
                        cleanup();
                        reject(new Error(errorMessage));
                    }
                } catch (error) {
                    console.error("Error processing authentication message:", error);
                    // Don't reject yet, it might be an unrelated message
                }
            };

            // Clean up function to remove listeners and clear timeout
            const cleanup = () => {
                this.wsConnection?.removeEventListener("message", authMessageHandler);
                clearTimeout(authTimeout);
                this.authenticationInProgress = null; // Reset authentication in progress
            };

            // Set timeout for auth process
            authTimeout = setTimeout(() => {
                cleanup();
                reject(new Error("Authentication timeout"));
            }, 15000); // 15 second timeout

            // Add temporary listener for authentication messages
            this.wsConnection?.addEventListener("message", authMessageHandler);

            // Use nitrolite's createAuthRequestMessage directly
            console.log("Starting authentication with address:", signer.address);

            // Use the same approach as the server
            createAuthRequestMessage(signer.sign, signer.address)
                .then((authRequest) => {
                    console.log("Sending auth_request:", authRequest);
                    this.wsConnection?.send(authRequest);
                })
                .catch((error) => {
                    console.error("Error creating auth request:", error);
                    cleanup();
                    reject(new Error(`Failed to create auth request: ${error.message}`));
                });
        });

        // Store the promise and return it
        this.authenticationInProgress = authPromise;
        return authPromise;
    }

    private handleWebSocketMessage(message: any): void {
        console.log("Received WebSocket message:", message);

        // Check if it's a response to a pending request
        if (message.id && this.pendingRequests.has(message.id)) {
            const { resolve, reject, timeout } = this.pendingRequests.get(message.id)!;
            clearTimeout(timeout);
            this.pendingRequests.delete(message.id);

            if (message.error) {
                reject(new Error(message.error.message || "Unknown error"));
            } else {
                resolve(message.result || message.res?.[2]);
            }
            return;
        }

        // Handle other message types
        if (message.method) {
            switch (message.method) {
                case "channel_update":
                    // Handle channel state update
                    if (this.activeChannel && message.params?.channel_id === this.activeChannel.channelId) {
                        this.activeChannel.state = message.params.state;
                    }
                    break;

                case "app_update":
                    // Handle application update
                    console.log("Received app update:", message.params);
                    break;
            }
        }
    }

    async signState(stateData: any, stateId: string, channelId: string) {
        if (!this.client || !this.isConnected) {
            console.error("ClearNet client not initialized");
            return null;
        }

        try {
            // We need to properly format the state according to Nitrolite SDK specs
            // The state must include the channelId, version, and any allocations
            const state = {
                channelId,
                stateData: JSON.stringify(stateData),
                version: BigInt(Math.floor(Date.now() / 1000)),
                allocations: this.activeChannel?.state?.allocations || [],
                stateId,
            };

            // Use the state wallet client if available, otherwise fall back to regular wallet client
            // This creates a cryptographic signature that proves this state update was authorized
            const stateHash = await this.getStateHash(state);

            // Choose which wallet client to use for signing
            const signingClient = this.config.stateWalletClient || this.config.walletClient;
            const signature = await signingClient.signMessage({
                message: { raw: stateHash },
            });

            return {
                signature,
                stateId,
                channelId,
                playerId: this.currentAddress,
            };
        } catch (error) {
            console.error("Failed to sign state:", error);
            return null;
        }
    }

    async getAccountChannels() {
        if (!this.client || !this.isConnected) {
            console.error("ClearNet client not initialized");
            return [];
        }

        try {
            return await this.client.getAccountChannels();
        } catch (error) {
            console.error("Failed to get account channels:", error);
            return [];
        }
    }

    async joinChannel(channelId: string, depositAmount: bigint, stateData: string): Promise<ChannelData | null> {
        if (!this.client || !this.isConnected) {
            console.error("ClearNet client not initialized");
            return null;
        }

        try {
            console.log("Joining channel:", channelId);
            console.log("Depositing amount:", depositAmount.toString());
            console.log("Join data:", stateData);

            // First, deposit funds using the Nitrolite client
            const depositTxHash = await this.client.deposit(depositAmount);

            if (!depositTxHash) {
                throw new Error("Failed to deposit funds");
            }

            console.log(`Deposit transaction hash: ${depositTxHash}`);

            // Get channel information to validate it exists
            const channelInfo = await this.client.getChannelInfo(channelId);

            if (!channelInfo) {
                throw new Error("Failed to get channel information");
            }

            console.log("Retrieved channel information:", channelInfo);

            // Join the channel by signing the state
            const initialState = channelInfo.state;

            // Sign the state hash
            const stateHash = await this.getStateHash(initialState);
            const signingClient = this.config.stateWalletClient || this.config.walletClient;

            // Sign the state hash
            const signature = await signingClient.signMessage({
                message: { raw: stateHash },
            });

            console.log(`Signed channel state with signature: ${signature}`);

            // Store channel in localStorage for persistence
            try {
                localStorage.setItem("nitro_channel_id", channelId);
                localStorage.setItem(
                    "nitro_channel_state",
                    JSON.stringify(initialState, (_, value) => (typeof value === "bigint" ? value.toString() + "n" : value))
                );
            } catch (error) {
                console.error("Failed to save channel to localStorage:", error);
            }

            // Store the active channel
            this.activeChannel = {
                channelId,
                state: initialState,
            };

            console.log(`Successfully joined channel ${channelId}`);
            return this.activeChannel;
        } catch (error) {
            console.error("Failed to join channel:", error);
            return null;
        }
    }

    // Helper method to hash a state with the Nitrolite protocol standard
    private async getStateHash(state: any): Promise<Hex> {
        if (!this.client) {
            throw new Error("ClearNet client not initialized");
        }

        // Format the state as required by the ERC-7824 specification
        const stateString = JSON.stringify(state);

        try {
            // Add the nitro protocol prefix for state hashing
            const prefixedState = `nitro-state:${stateString}`;

            // Convert to Uint8Array for hashing
            const encoder = new TextEncoder();
            const data = encoder.encode(prefixedState);

            // Use the browser's crypto API to create the state hash
            // This follows the ERC-7824 state hashing specification
            const hashBuffer = await window.crypto.subtle.digest("SHA-256", data);

            // Convert hash to hex string
            const hashArray = Array.from(new Uint8Array(hashBuffer));
            const hashHex = hashArray.map((b) => b.toString(16).padStart(2, "0")).join("");

            // Add the 0x prefix for Ethereum compatibility
            return "0x" + hashHex as Hex;
        } catch (error) {
            console.error("Failed to hash state:", error);
            // Return a mock hash if there's an error
            return "0x" + Array(64).fill("0").join("") as Hex;
        }
    }

    getActiveChannel(): ChannelData | null {
        // If we don't have an active channel but have one in storage, try to restore it
        if (!this.activeChannel) {
            this.restoreChannelFromStorage();
        }
        return this.activeChannel;
    }
}

export const clearNetService = new ClearNetService();
export default clearNetService;
