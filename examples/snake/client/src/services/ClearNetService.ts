import {
    NitroliteClient,
    createGetLedgerBalancesMessage,
    type NitroliteClientConfig,
} from "@erc7824/nitrolite";
import { BROKER_WS_URL } from "../config";
import { createEthersSigner, generateKeyPair } from "../crypto";
import type { Hex } from "viem";
import { authenticate } from "./authentication";

class ClearNetService {
    public client!: NitroliteClient;
    public config!: NitroliteClientConfig;
    private isConnected = false;
    private currentAddress: string | null = null;
    private activeChannel: Hex | null = null;
    private wsConnection: WebSocket | null = null;
    private readonly wsUrl = BROKER_WS_URL;
    private pendingRequests = new Map<
        string,
        {
            resolve: (value: any) => void;
            reject: (reason: Error) => void;
            timeout: number;
        }
    >();
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private reconnectDelay = 1000;
    private reconnectTimeout: number | null = null;
    private authenticationInProgress: Promise<void> | null = null;

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
        }, delay) as unknown as number;
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

        // Verify we have wallet client for authentication
        const eip712SignerWalletClient = this.config?.walletClient;
        if (!eip712SignerWalletClient) {
            throw new Error('No main wallet client (e.g., MetaMask) available for EIP-712 authentication');
        }

        // Get or create session key signer
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

        // Create and store the authentication promise
        this.authenticationInProgress = authenticate(this.wsConnection, eip712SignerWalletClient, signer, 15000)
            .then(async () => {
                console.log("Authentication successful, sending get_balances");

                // Send get_balances message after successful authentication
                const nitroChannelId = localStorage.getItem("nitro_channel_id");
                if (nitroChannelId && this.wsConnection) {
                    const getBalancesMsg = await createGetLedgerBalancesMessage(
                        signer.sign,
                        nitroChannelId as Hex
                    );
                    this.wsConnection.send(getBalancesMsg);
                }
            })
            .finally(() => {
                this.authenticationInProgress = null;
            });

        return this.authenticationInProgress;
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

    async getActiveChannel(): Promise<Hex | null> {
        if (this.activeChannel) {
            return this.activeChannel;
        }

        const channels = await this.client.getAccountChannels();
        if (channels.length <= 0) {
            throw new Error('No active channel found. Please open a channel at apps.yellow.com');
        }
        this.activeChannel = channels[0];

        return this.activeChannel;
    }
}

export const clearNetService = new ClearNetService();
export default clearNetService;
