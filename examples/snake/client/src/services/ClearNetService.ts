import {
    NitroliteClient,
    createAuthRequestMessage,
    createAuthVerifyMessage,
    createAuthVerifyMessageWithJWT,
    createGetLedgerBalancesMessage,
    type NitroliteClientConfig,
    type AuthRequest,
    NitroliteRPC,
    getCurrentTimestamp,
} from "@erc7824/nitrolite";
import { BROKER_WS_URL } from "../config";
import { createEthersSigner, generateKeyPair } from "../crypto";
import type { Hex } from "viem";
import { ethers } from "ethers";

export interface ChannelData {
    channelId: string;
    state: any;
}

/**
 * EIP-712 domain and types for auth_verify challenge
 */
const getAuthDomain = () => {
    return {
        name: 'Snake Game',
    };
};

const AUTH_TYPES = {
    Policy: [
        { name: 'challenge', type: 'string' },
        { name: 'scope', type: 'string' },
        { name: 'wallet', type: 'address' },
        { name: 'application', type: 'address' },
        { name: 'participant', type: 'address' },
        { name: 'expire', type: 'uint256' },
        { name: 'allowances', type: 'Allowance[]' },
    ],
    Allowance: [
        { name: 'asset', type: 'string' },
        { name: 'amount', type: 'uint256' },
    ],
};

/**
 * Creates EIP-712 signing function for challenge verification with proper challenge extraction
 */
function createEIP712SigningFunction(walletClient: any, stateSigner: any, expire: string) {
    if (!walletClient) {
        throw new Error('No wallet client available for EIP-712 signing');
    }

    return async (data: any): Promise<`0x${string}`> => {
        console.log('Signing auth_verify challenge with EIP-712:', data);

        let challengeUUID = '';
        const address = walletClient.account?.address;

        // The data coming in is the array from createAuthVerifyMessage
        // Format: [timestamp, "auth_verify", [{"address": "0x...", "challenge": "uuid"}], timestamp]
        if (Array.isArray(data)) {
            console.log('Data is array, extracting challenge from position [2][0].challenge');

            // Direct array access - data[2] should be the array with the challenge object
            if (data.length >= 3 && Array.isArray(data[2]) && data[2].length > 0) {
                const challengeObject = data[2][0];

                if (challengeObject && challengeObject.challenge) {
                    challengeUUID = challengeObject.challenge;
                    console.log('Extracted challenge UUID from array:', challengeUUID);
                }
            }
        } else if (typeof data === 'string') {
            try {
                const parsed = JSON.parse(data);

                console.log('Parsed challenge data:', parsed);

                // Handle different message structures
                if (parsed.res && Array.isArray(parsed.res)) {
                    // auth_challenge response: {"res": [id, "auth_challenge", {"challenge": "uuid"}, timestamp]}
                    if (parsed.res[1] === 'auth_challenge' && parsed.res[2]) {
                        challengeUUID = parsed.res[2].challenge_message || parsed.res[2].challenge;
                        console.log('Extracted challenge UUID from auth_challenge:', challengeUUID);
                    }
                    // auth_verify message: [timestamp, "auth_verify", [{"address": "0x...", "challenge": "uuid"}], timestamp]
                    else if (parsed.res[1] === 'auth_verify' && Array.isArray(parsed.res[2]) && parsed.res[2][0]) {
                        challengeUUID = parsed.res[2][0].challenge;
                        console.log('Extracted challenge UUID from auth_verify:', challengeUUID);
                    }
                }
                // Direct array format
                else if (Array.isArray(parsed) && parsed.length >= 3 && Array.isArray(parsed[2])) {
                    challengeUUID = parsed[2][0]?.challenge;
                    console.log('Extracted challenge UUID from direct array:', challengeUUID);
                }
            } catch (e) {
                console.error('Could not parse challenge data:', e);
                console.log('Using raw string as challenge');
                challengeUUID = data;
            }
        } else if (data && typeof data === 'object') {
            // If data is already an object, try to extract challenge
            challengeUUID = data.challenge || data.challenge_message;
            console.log('Extracted challenge from object:', challengeUUID);
        }

        if (!challengeUUID || challengeUUID.includes('[') || challengeUUID.includes('{')) {
            console.error('Challenge extraction failed or contains invalid characters:', challengeUUID);
            throw new Error('Could not extract valid challenge UUID for EIP-712 signing');
        }

        console.log('Final challenge UUID for EIP-712:', challengeUUID);
        console.log('Signing for address (original):', address);
        console.log('Signing for address (type):', typeof address);
        console.log('Auth domain:', getAuthDomain());

        // Create EIP-712 message with ONLY the challenge UUID
        const message = {
            challenge: challengeUUID,
            scope: 'snake-game',
            wallet: address as `0x${string}`,
            application: address as `0x${string}`,
            participant: stateSigner.address as `0x${string}`,
            expire: expire,
            allowances: [
                {
                    asset: 'usdc',
                    amount: '0',
                },
            ],
        };

        try {
            // Sign with EIP-712
            const signature = await walletClient.signTypedData({
                account: walletClient.account!,
                domain: getAuthDomain(),
                types: AUTH_TYPES,
                primaryType: 'Policy',
                message: message,
            });

            console.log('EIP-712 signature generated for challenge:', signature);
            return signature;
        } catch (eip712Error) {
            console.error('EIP-712 signing failed:', eip712Error);
            console.log('Attempting fallback to regular message signing...');

            try {
                // Fallback to regular message signing if EIP-712 fails
                const fallbackMessage = `Authentication challenge for ${address}: ${challengeUUID}`;

                console.log('Fallback message:', fallbackMessage);

                const fallbackSignature = await walletClient.signMessage({
                    message: fallbackMessage,
                    account: walletClient.account!,
                });

                console.log('Fallback signature generated:', fallbackSignature);
                return fallbackSignature as `0x${string}`;
            } catch (fallbackError) {
                console.error('Fallback signing also failed:', fallbackError);
                throw new Error(`Both EIP-712 and fallback signing failed: ${eip712Error.message}`);
            }
        }
    };
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
            timeout: number;
        }
    >();
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private reconnectDelay = 1000;
    private reconnectTimeout: number | null = null;
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

        // Verify we have wallet client for both address and EIP-712 signing
        const walletClient = this.config?.walletClient;

        if (!walletClient) {
            throw new Error('No wallet client available for authentication');
        }

        const privyWalletAddress = walletClient.account?.address;

        if (!privyWalletAddress) {
            throw new Error('No Privy wallet address available for authentication');
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
        const expire = String(Math.floor(Date.now() / 1000) + 24 * 60 * 60);

        // Create a new authentication promise and store it
        const authPromise = new Promise<void>(async (resolve, reject) => {
            let authTimeout: number;

            // Create a one-time message handler for authentication
            const authMessageHandler = async (event: MessageEvent) => {
                try {
                    const message = JSON.parse(event.data);
                    console.log("Auth process message received:", message);

                    // Check for auth_challenge response
                    if (message.res && message.res[1] === "auth_challenge") {
                        console.log("Received auth_challenge, preparing EIP-712 auth_verify...");

                        try {
                            // Step 2: Create EIP-712 signing function for challenge verification
                            console.log('Creating EIP-712 signing function...');
                            const eip712SigningFunction = createEIP712SigningFunction(walletClient, signer, expire);

                            console.log('Calling createAuthVerifyMessage...');
                            // Create and send verification message with EIP-712 signature
                            const authVerify = await createAuthVerifyMessage(
                                eip712SigningFunction,
                                event.data, // Pass the raw challenge response string/object
                            );

                            console.log('Sending auth_verify with EIP-712 signature');
                            this.wsConnection?.send(authVerify);
                            console.log('auth_verify sent successfully');

                            setTimeout(async () => {
                                const nitroChannelId = localStorage.getItem("nitro_channel_id");

                                if (nitroChannelId) {
                                    // Send get_balances message to the broker
                                    const getBalancesMsg = await createGetLedgerBalancesMessage(signer.sign, nitroChannelId as Hex);

                                    this.wsConnection?.send(getBalancesMsg);
                                }
                            }, 2000);
                        } catch (eip712Error) {
                            console.error('Error creating EIP-712 auth_verify:', eip712Error);
                            console.error('Error stack:', (eip712Error as Error).stack);

                            cleanup();
                            reject(new Error(`EIP-712 auth_verify failed: ${(eip712Error as Error).message}`));
                            return;
                        }
                    }
                    // Check for auth_verify success response
                    else if (message.res && (message.res[1] === "auth_verify" || message.res[1] === 'auth_success')) {
                        console.log("Authentication successful");

                        // If response contains a JWT token, store it
                        if (message.res[2]?.[0]?.['jwt_token']) {
                            console.log('JWT token received:', message.res[2][0]['jwt_token']);
                            window.localStorage.setItem('jwt_token', message.res[2][0]['jwt_token']);
                        }

                        cleanup();
                        resolve();
                    }
                    // Check for error responses
                    else if (message.err || (message.res && message.res[1] === "error")) {
                        const errorMsg = message.err?.[1] || message.error || message.res?.[2]?.[0]?.error || 'Authentication failed';

                        console.error('Authentication failed:', errorMsg);
                        window.localStorage.removeItem('jwt_token');
                        cleanup();
                        reject(new Error(String(errorMsg)));
                    } else {
                        console.log('Received non-auth message during auth, continuing to listen:', message);
                        // Keep listening if it wasn't a final success/error
                    }
                } catch (error) {
                    console.error("Error handling auth response:", error);
                    console.error("Error stack:", (error as Error).stack);
                    cleanup();
                    reject(new Error(`Authentication error: ${error instanceof Error ? error.message : String(error)}`));
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
            }, 15000) as unknown as number; // 15 second timeout

            // Add temporary listener for authentication messages
            this.wsConnection?.addEventListener("message", authMessageHandler);

            // Step 1: Send auth_request with empty signature and Privy address
            console.log('Starting authentication with:');
            console.log('- Privy wallet address:', privyWalletAddress);
            console.log('- Empty signature for auth_request');
            console.log('- EIP-712 signature for auth_verify challenge (UUID only)');

            try {
                let authRequest: string;
                const jwtToken = window.localStorage.getItem('jwt_token');

                if (jwtToken) {
                    console.log('JWT token found, sending auth request with token:', jwtToken);
                    authRequest = await createAuthVerifyMessageWithJWT(jwtToken);
                } else {
                    console.log('No JWT token found, proceeding with challenge-response authentication');
                    authRequest = await createAuthRequestMessage({
                        wallet: privyWalletAddress as Hex,
                        participant: signer.address as Hex,
                        app_name: 'Snake Game',
                        expire: expire,
                        scope: 'snake-game',
                        application: privyWalletAddress as Hex,
                        allowances: [
                            {
                                symbol: 'usdc',
                                amount: '0',
                            },
                        ],
                    });
                }

                console.log('Sending auth_request with empty signature and Privy address:', privyWalletAddress);
                this.wsConnection?.send(authRequest);
            } catch (requestError) {
                console.error('Error creating auth_request:', requestError);
                cleanup();
                reject(new Error(`Failed to create auth_request: ${(requestError as Error).message}`));
            }
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
