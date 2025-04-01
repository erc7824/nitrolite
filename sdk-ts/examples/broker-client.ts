/**
 * Broker Client Example
 *
 * This example demonstrates how to use the Hachi SDK to interact with the Virtual Ledger Broker.
 * It shows how to:
 * 1. Establish an on-chain state channel with the broker
 * 2. Connect to the broker via WebSocket
 * 3. Create and use virtual channels for communication with other participants
 * 4. Handle state updates and signatures
 * 5. Properly manage error handling for a robust application
 */

import { HachiClient, TokenError, NetworkError, generateChannelNonce, getStateHash, verifySignature } from "../src";

import { RPCClient, LVCI } from "../src/rpc";
import { RPCChannelManager } from "../src/rpc/channel";
import { createRPCChannelContext } from "../src/rpc/integration";
import { VirtualChannelError } from "../src/errors";

import { createPublicClient, createWalletClient, http, parseEther, formatEther, type Hex, type Address } from "viem";
import { privateKeyToAccount } from "viem/accounts";
import { sepolia } from "viem/chains";
import WebSocket from "isows";

// Broker WebSocket provider implementation for Hachi SDK
class BrokerWebSocketProvider {
    private ws: WebSocket | null = null;
    private address: Address;
    private brokerUrl: string;
    private messageHandlers: Array<(from: Address, message: any) => void> = [];
    private connected = false;
    private channelId: string | null = null;

    constructor(address: Address, brokerUrl: string) {
        this.address = address;
        this.brokerUrl = brokerUrl;
    }

    async connect(): Promise<void> {
        if (this.connected) {
            return;
        }

        return new Promise((resolve, reject) => {
            try {
                this.ws = new WebSocket(this.brokerUrl);

                this.ws.onopen = () => {
                    console.log("Connected to broker websocket");
                    this.connected = true;
                    this.ws?.send(
                        JSON.stringify({
                            type: "connect",
                            address: this.address,
                        })
                    );
                    resolve();
                };

                this.ws.onmessage = (event) => {
                    try {
                        const data = JSON.parse(event.data as string);

                        // Handle broker specific messages
                        if (data.type === "connection_established") {
                            console.log("Connection established with broker");
                            return;
                        }

                        if (data.type === "channel_created") {
                            this.channelId = data.channelId;
                            console.log(`Channel created: ${this.channelId}`);
                            return;
                        }

                        // Handle RPC messages
                        if (data.req || data.res) {
                            const from = data.sender as Address;
                            this.messageHandlers.forEach((handler) => handler(from, data));
                        }
                    } catch (error) {
                        console.error("Error parsing message:", error);
                    }
                };

                this.ws.onerror = (error) => {
                    console.error("WebSocket error:", error);
                    reject(new Error("WebSocket connection error"));
                };

                this.ws.onclose = () => {
                    console.log("WebSocket connection closed");
                    this.connected = false;
                };
            } catch (error) {
                reject(error);
            }
        });
    }

    async disconnect(): Promise<void> {
        if (!this.connected || !this.ws) {
            return;
        }

        return new Promise((resolve) => {
            if (this.ws) {
                this.ws.onclose = () => {
                    this.connected = false;
                    resolve();
                };
                this.ws.close();
            } else {
                resolve();
            }
        });
    }

    async send(recipient: Address, message: any): Promise<void> {
        if (!this.connected || !this.ws) {
            throw new NetworkError("Not connected to broker");
        }

        // Add recipient to message
        const messageWithRecipient = {
            ...message,
            recipient,
            sender: this.address,
            channelId: this.channelId,
        };

        this.ws.send(JSON.stringify(messageWithRecipient));
    }

    onMessage(handler: (from: Address, message: any) => void): () => void {
        this.messageHandlers.push(handler);

        // Return function to unregister handler
        return () => {
            const index = this.messageHandlers.indexOf(handler);
            if (index !== -1) {
                this.messageHandlers.splice(index, 1);
            }
        };
    }

    isConnected(): boolean {
        return this.connected;
    }

    getChannelId(): string | null {
        return this.channelId;
    }
}

// Main application class
class BrokerClientApp {
    private hachiClient: HachiClient;
    private rpcClient: RPCClient;
    private channelManager: RPCChannelManager;
    private provider: BrokerWebSocketProvider;
    private address: Address;
    private brokerAddress: Address;
    private tokenAddress: Address;

    constructor(
        privateKey: Hex,
        brokerUrl: string,
        brokerAddress: Address,
        tokenAddress: Address,
        contractAddresses: {
            custody: Address;
            adjudicators: Record<string, Address>;
        }
    ) {
        // Create viem account from private key
        const account = privateKeyToAccount(privateKey);
        this.address = account.address;
        this.brokerAddress = brokerAddress;
        this.tokenAddress = tokenAddress;

        // Create Ethereum clients
        const publicClient = createPublicClient({
            chain: sepolia,
            transport: http("https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY"),
        });

        const walletClient = createWalletClient({
            account,
            chain: sepolia,
            transport: http("https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY"),
        });

        // Create Hachi client for on-chain interactions
        this.hachiClient = new HachiClient({
            publicClient,
            walletClient,
            account,
            addresses: contractAddresses,
            logger: console,
        });

        // Create broker websocket provider
        this.provider = new BrokerWebSocketProvider(this.address, brokerUrl);

        // Create RPC client for off-chain communication
        this.rpcClient = new RPCClient({
            provider: this.provider,
            address: this.address,
            signer: (message) => account.signMessage({ message }),
            requestTimeoutMs: 60000,
            maxRequestRetries: 3,
            logger: console,
        });

        // Create channel manager
        this.channelManager = new RPCChannelManager(this.rpcClient);
    }

    /**
     * Initialize the application
     */
    async initialize(): Promise<void> {
        try {
            // Connect to the broker
            await this.provider.connect();
            console.log("Connected to broker");

            // Register RPC methods
            this.registerRPCMethods();

            // Get current timestamp from broker
            const timestamp = await this.getServerTime();
            console.log(`Broker timestamp: ${timestamp}`);
        } catch (error) {
            console.error("Failed to initialize:", error);
            throw error;
        }
    }

    /**
     * Establish a direct channel with the broker
     */
    async createDirectChannel(amount: bigint): Promise<string> {
        try {
            // Check token balance
            const balance = await this.hachiClient.getTokenBalance(this.tokenAddress, this.address);
            console.log(`Token balance: ${formatEther(balance)}`);

            if (balance < amount) {
                throw new TokenError("Insufficient token balance to fund channel", "INSUFFICIENT_BALANCE", 400, "Add more tokens to your wallet", {
                    required: amount,
                    actual: balance,
                    tokenAddress: this.tokenAddress,
                });
            }

            // Check token allowance
            const allowance = await this.hachiClient.getTokenAllowance(this.tokenAddress, this.address, this.hachiClient.addresses.custody);

            console.log(`Current allowance: ${formatEther(allowance)}`);

            // Approve tokens if needed
            if (allowance < amount) {
                console.log(`Approving ${formatEther(amount)} tokens`);
                await this.hachiClient.approveTokens(this.tokenAddress, amount, this.hachiClient.addresses.custody);
                console.log("Token approval successful");
            }

            // Create a channel with the broker
            const nonce = generateChannelNonce(this.address);

            // Configure the channel
            const channel = {
                participants: [this.address, this.brokerAddress] as [Address, Address],
                adjudicator: this.hachiClient.addresses.adjudicators.sequential,
                challenge: BigInt(86400), // 1 day challenge period
                nonce,
            };

            // Initial state with allocations
            const initialState = {
                data: "0x00", // Empty data for sequential adjudicator
                allocations: [
                    {
                        destination: this.address,
                        amount,
                        token: this.tokenAddress,
                    },
                    {
                        destination: this.brokerAddress,
                        amount: BigInt(0),
                        token: this.tokenAddress,
                    },
                ],
                sigs: [],
            };

            // Open the channel
            console.log("Opening direct channel with broker...");
            const channelId = await this.hachiClient.openChannel(channel, initialState);
            console.log(`Channel opened with ID: ${channelId}`);

            return channelId;
        } catch (error) {
            console.error("Failed to create direct channel:", error);
            throw error;
        }
    }

    /**
     * Create a virtual channel with another participant through the broker
     */
    async createVirtualChannel(
        counterparty: Address,
        initialAllocation: {
            myAmount: bigint;
            theirAmount: bigint;
            token: Address;
        }
    ): Promise<string> {
        try {
            // Check if connected to broker
            if (!this.provider.isConnected()) {
                throw new NetworkError("Not connected to broker");
            }

            // Create LVCI (logical virtual channel identifier)
            const lvci = LVCI.create(this.address, counterparty, [this.brokerAddress]);

            // Encode initial state for the virtual channel
            const initialState = {
                data: "0x00", // Empty data for sequential adjudicator
                allocations: [
                    {
                        destination: this.address,
                        amount: initialAllocation.myAmount,
                        token: initialAllocation.token,
                    },
                    {
                        destination: counterparty,
                        amount: initialAllocation.theirAmount,
                        token: initialAllocation.token,
                    },
                ],
                sigs: [],
            };

            // Request virtual channel creation through broker
            console.log(`Creating virtual channel with ${counterparty}...`);
            const vchanId = await this.rpcClient.createVirtualChannel(lvci, initialState);
            console.log(`Virtual channel created with ID: ${vchanId}`);

            return vchanId;
        } catch (error) {
            if (error instanceof VirtualChannelError) {
                console.error("Virtual channel error:", error.message);
                console.error("Suggestion:", error.suggestion);
                if (error.details?.lvci) {
                    console.error("LVCI:", error.details.lvci);
                }
            } else {
                console.error("Failed to create virtual channel:", error);
            }
            throw error;
        }
    }

    /**
     * Send a signed NitroRPC request to a counterparty using the broker
     */
    async sendRPCRequest(recipient: Address, method: string, params: any[], vchanId: string): Promise<any> {
        try {
            // Get current timestamp from broker
            const timestamp = await this.getServerTime();

            // Generate request ID
            const requestId = Math.floor(Math.random() * 10000) + 1000;

            // Create request payload
            const payload: [number, string, any[], number] = [requestId, method, params, timestamp];

            // Convert payload to hex string for signing
            const payloadHex = ("0x" + Buffer.from(JSON.stringify(payload)).toString("hex")) as Hex;

            // Sign the payload
            const signature = await this.rpcClient.config.signer(payloadHex);

            // Create NitroRPC request message
            const rpcMessage = {
                req: payload,
                sig: signature,
                vchanId, // Include virtual channel ID
            };

            console.log(`Sending RPC request: ${method}`);

            // Send request through broker
            const response = await this.rpcClient.sendRequest(recipient, method, params, {
                metadata: { vchanId, timestamp },
            });

            return response;
        } catch (error) {
            console.error(`Failed to send RPC request for method ${method}:`, error);
            throw error;
        }
    }

    /**
     * Handle incoming RPC requests
     */
    private registerRPCMethods(): void {
        // Register a method handler for "add"
        this.rpcClient.registerMethod("add", async (params, sender) => {
            console.log(`Received 'add' request from ${sender} with params:`, params);

            // Simple addition implementation
            const a = typeof params[0] === "number" ? params[0] : 0;
            const b = typeof params[1] === "number" ? params[1] : 0;
            const result = a + b;

            return [result];
        });

        // Register a method handler for "subtract"
        this.rpcClient.registerMethod("subtract", async (params, sender) => {
            console.log(`Received 'subtract' request from ${sender} with params:`, params);

            // Simple subtraction implementation
            const a = typeof params[0] === "number" ? params[0] : 0;
            const b = typeof params[1] === "number" ? params[1] : 0;
            const result = a - b;

            return [result];
        });

        // Register a method handler for "multiply"
        this.rpcClient.registerMethod("multiply", async (params, sender) => {
            console.log(`Received 'multiply' request from ${sender} with params:`, params);

            // Simple multiplication implementation
            const a = typeof params[0] === "number" ? params[0] : 0;
            const b = typeof params[1] === "number" ? params[1] : 0;
            const result = a * b;

            return [result];
        });

        // Register handlers for channel-related methods
        this.rpcClient.registerMethod("update_state", async (params, sender, metadata) => {
            console.log(`Received state update from ${sender}`);

            // Extract state from params
            const state = params[0];
            const vchanId = metadata?.vchanId;

            if (!vchanId) {
                throw new Error("Missing virtual channel ID");
            }

            // Verify the state transition is valid
            // (in a real implementation, you'd check against current state)

            // Create state hash and sign it
            const stateHash = getStateHash(state);
            const signature = await this.rpcClient.config.signer(stateHash);

            // Return signature as acceptance
            return [{ accepted: true, signature }];
        });
    }

    /**
     * Get the current server time from the broker
     */
    private async getServerTime(): Promise<number> {
        try {
            const response = await this.rpcClient.sendRequest(this.brokerAddress, "get_time", []);

            return response[0] as number;
        } catch (error) {
            console.error("Failed to get server time:", error);
            // Return current time as fallback
            return Date.now();
        }
    }

    /**
     * Close all connections and clean up resources
     */
    async cleanup(): Promise<void> {
        try {
            await this.rpcClient.disconnect();
            console.log("Disconnected from broker");
        } catch (error) {
            console.error("Error during cleanup:", error);
        }
    }
}

/**
 * Example usage
 */
async function runBrokerExample() {
    // Configuration
    const PRIVATE_KEY = "0x0000000000000000000000000000000000000000000000000000000000000001" as Hex; // Replace with real private key
    const BROKER_URL = "ws://localhost:8080";
    const BROKER_ADDRESS = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266" as Address;
    const TOKEN_ADDRESS = "0xc778417E063141139Fce010982780140Aa0cD5Ab" as Address; // WETH on Sepolia
    const CONTRACT_ADDRESSES = {
        custody: "0x5FbDB2315678afecb367f032d93F642f64180aa3" as Address,
        adjudicators: {
            sequential: "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512" as Address,
            numeric: "0x9fE46736679d2D9a65F0992F2272dE9f3c7fa6e0" as Address,
            trivial: "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9" as Address,
        },
    };

    const app = new BrokerClientApp(PRIVATE_KEY, BROKER_URL, BROKER_ADDRESS, TOKEN_ADDRESS, CONTRACT_ADDRESSES);

    try {
        // Initialize app and connect to broker
        await app.initialize();

        // Create direct channel with broker with 1 ETH
        const directChannelId = await app.createDirectChannel(parseEther("1"));

        // Create virtual channel with another participant
        const COUNTERPARTY = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8" as Address;
        const vchanId = await app.createVirtualChannel(COUNTERPARTY, {
            myAmount: parseEther("0.1"),
            theirAmount: parseEther("0"),
            token: TOKEN_ADDRESS,
        });

        // Send RPC requests through the virtual channel
        const addResult = await app.sendRPCRequest(COUNTERPARTY, "add", [10, 20], vchanId);
        console.log("Add result:", addResult);

        const subtractResult = await app.sendRPCRequest(COUNTERPARTY, "subtract", [50, 30], vchanId);
        console.log("Subtract result:", subtractResult);

        const multiplyResult = await app.sendRPCRequest(COUNTERPARTY, "multiply", [5, 7], vchanId);
        console.log("Multiply result:", multiplyResult);

        // Keep the connection alive for a while
        await new Promise((resolve) => setTimeout(resolve, 10000));
    } catch (error) {
        console.error("Example failed:", error);
    } finally {
        // Clean up
        await app.cleanup();
    }
}

// Run the example if executed directly
if (require.main === module) {
    runBrokerExample().catch(console.error);
}

// Export for importing in other files
export { BrokerClientApp, BrokerWebSocketProvider };
