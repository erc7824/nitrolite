import WebSocket from "ws";
import { ethers } from "ethers";
import {
    createAuthRequestMessage,
    RequestData,
    ResponsePayload,
    MessageSigner,
    NitroliteClient,
    NitroliteClientConfig,
    AppDefinition,
    NitroliteRPC,
    createGetLedgerBalancesMessage,
    CreateAppSessionRequest,
    CloseAppSessionRequest,
    AuthRequest,
    getCurrentTimestamp,
} from "@erc7824/nitrolite";
import { BROKER_WS_URL, CONTRACT_ADDRESSES, POLYGON_RPC_URL, WALLET_PRIVATE_KEY } from "../config/index.ts";
import { setBrokerWebSocket, getBrokerWebSocket, addPendingRequest, getPendingRequest, clearPendingRequest } from "./stateService.ts";
import { Hex, createWalletClient, createPublicClient, http } from "viem";
import { polygon } from "viem/chains";
import { privateKeyToAccount } from "viem/accounts";

import util from 'util';
util.inspect.defaultOptions.depth = null;

const DEFAULT_PROTOCOL = "app_aura_nitrolite_v0";
const DEFAULT_WEIGHTS: number[] = [0, 0, 100]; // Alice: 0, Bob: 0, Server: 100
const DEFAULT_QUORUM: number = 100; // server alone decides the outcome

// Flag to indicate if we've authenticated with the broker
let isAuthenticated = false;
let client: NitroliteClient = createClient();

function createClient(): NitroliteClient {
    // Create the wallet client using the ethereum provider

    const wallet = privateKeyToAccount(WALLET_PRIVATE_KEY);
    const walletClient = createWalletClient({
        transport: http(process.env.POLYGON_RPC_URL),
        chain: polygon,
        account: wallet,
    });

    const publicClient = createPublicClient({
        transport: http(POLYGON_RPC_URL),
        chain: polygon,
    });

    // Create a dedicated client for signing state updates
    const stateWalletClient = createWalletClient({
        transport: http(process.env.POLYGON_RPC_URL),
        chain: polygon,
        account: wallet,
    });
    const config: NitroliteClientConfig = {
        publicClient,
        walletClient,
        stateWalletClient,
        addresses: CONTRACT_ADDRESSES,
        chainId: polygon.id,
        challengeDuration: BigInt(86400), // 1 day in seconds
    };
    const client = new NitroliteClient(config);

    return client;
}

async function createBrokerChannel(client: NitroliteClient): Promise<void> {
    // Create a channel with the broker
    const createChannelResponse = await client.createChannel({
        initialAllocationAmounts: [0n, 0n],
        stateData: "0x",
    });
    console.log("Created channel", createChannelResponse);

    // Check if broker joined the channel
    getChannels();
}

async function getChannels(): Promise<void> {
    const brokerWs = getBrokerWebSocket();
    if (!brokerWs || brokerWs.readyState !== WebSocket.OPEN) {
        throw new Error("WebSocket not connected");
    }

    const signer = createEthersSigner(WALLET_PRIVATE_KEY);
    const params = [{ participant: signer.address }];
    const request = NitroliteRPC.createRequest(10, "get_channels", params);
    const getChannelMessage = await NitroliteRPC.signRequestMessage(request, signer.sign);
    brokerWs.send(JSON.stringify(getChannelMessage));
}

// Connects to the Nitrolite broker
export function connectToBroker(): void {
    const brokerWs = getBrokerWebSocket();
    if (brokerWs && (brokerWs.readyState === WebSocket.OPEN || brokerWs.readyState === WebSocket.CONNECTING)) {
        console.log("WebSocket already connected or connecting. State:", brokerWs.readyState);
        return;
    }

    console.log(`Connecting to Nitrolite broker at ${BROKER_WS_URL}`);
    const ws = new WebSocket(BROKER_WS_URL);
    setBrokerWebSocket(ws);
    isAuthenticated = false;

    ws.on("open", async () => {
        console.log("Connected to Nitrolite broker");

        // Authenticate with the broker immediately upon connection
        try {
            await authenticateWithBroker();
            console.log("Successfully authenticated with broker");
        } catch (error) {
            console.error("Authentication with broker failed:", error);
        }
    });

    ws.on("message", (data) => {
        try {
            const message = JSON.parse(data.toString());
            console.log("Received message from broker:", {
                method: message.res?.[1],
                requestId: message.res?.[0],
                isAuthenticated
            });
            handleBrokerMessage(message);
        } catch (error) {
            console.error("Error parsing message from broker:", error);
        }
    });

    ws.on("close", (code, reason) => {
        console.log("Disconnected from Nitrolite broker:", {
            code,
            reason: reason.toString(),
            isAuthenticated
        });
        isAuthenticated = false;
        setTimeout(connectToBroker, 5000);
    });

    ws.on("error", (error) => {
        console.error("Error in broker WebSocket connection:", {
            error: error.message,
            isAuthenticated
        });
    });
}

// Authenticate with the broker using server's wallet and nitrolite package
async function authenticateWithBroker(): Promise<void> {
    const brokerWs = getBrokerWebSocket();
    if (!brokerWs || brokerWs.readyState !== WebSocket.OPEN) {
        throw new Error("WebSocket not connected");
    }

    // Create the wallet signer using our factory
    const signer = createEthersSigner(WALLET_PRIVATE_KEY);
    const serverAddress = signer.address;
    if (!serverAddress) {
        throw new Error("Server address not found");
    }

    return new Promise((resolve, reject) => {
        let authTimeout: NodeJS.Timeout;

        // Clean up function to remove listeners and clear timeout
        const cleanup = () => {
            brokerWs.removeListener("message", authMessageHandler);
            clearTimeout(authTimeout);
        };

        // Create a one-time message handler for authentication
        const authMessageHandler = async (data: WebSocket.RawData) => {
            try {
                const message = JSON.parse(data.toString());
                console.log("Auth process message received:", message);

                // Check for auth_challenge response (response to our auth_request)
                if (message.res && message.res[1] === "auth_challenge") {
                    console.log("Received auth_challenge, preparing auth_verify...");

                    try {
                        // Parse the challenge from the response
                        const parsedResponse = NitroliteRPC.parseResponse(data.toString());
                        if (!parsedResponse.isValid || parsedResponse.method !== "auth_challenge") {
                            throw new Error("Invalid auth_challenge response");
                        }
                        
                        const challengeData = parsedResponse.data as any[];
                        const challenge = challengeData[0]?.challenge_message;
                        if (!challenge) {
                            throw new Error("No challenge in auth_challenge response");
                        }

                        // Create auth_verify request with EIP-712 signature
                        const authVerifyRequest = await createAuthVerifyWithEIP712(
                            serverAddress,
                            challenge,
                            serverAddress, // session_key
                            "snake-game-server", // app_name
                            [] // allowances
                        );

                        console.log("Sending auth_verify:", authVerifyRequest);
                        brokerWs.send(authVerifyRequest);

                        // Send additional requests
                        const getBalances = await createGetLedgerBalancesMessage(
                            signer.sign,
                            serverAddress
                        );
                        brokerWs.send(getBalances);
                        await getChannels();
                    } catch (error: unknown) {
                        console.error("Error creating auth verify message:", error);
                        cleanup();
                        reject(new Error(`Failed to create auth verify message: ${error}`));
                    }
                }
                // Check for auth_verify success response
                else if (message.res && message.res[1] === "auth_verify") {
                    isAuthenticated = true;
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
            } catch (error: unknown) {
                console.error("Error processing authentication message:", error);
                cleanup();
                if (error instanceof Error) {
                    reject(new Error(`Authentication processing error: ${error.message}`));
                } else {
                    reject(new Error('Authentication processing error: Unknown error'));
                }
            }
        };

        // Set timeout for auth process
        authTimeout = setTimeout(() => {
            cleanup();
            reject(new Error("Authentication timeout"));
        }, 15000); // 15 second timeout

        // Add temporary listener for authentication messages
        brokerWs.on("message", authMessageHandler);

        // Create and send the auth request using nitrolite
        console.log("Starting authentication with address:", serverAddress);
        console.log("Server wallet address:", serverAddress);
        console.log("Private key used (first 4 chars):", WALLET_PRIVATE_KEY.substring(0, 6) + "...");

        // Create the auth request parameters for the new API
        const authRequest: AuthRequest = {
            address: serverAddress,
            session_key: serverAddress, // Using same address as session key for server
            app_name: "snake-game-server",
            allowances: [] // Empty allowances for server
        };

        // Generate the auth request using nitrolite and our properly typed signer
        createAuthRequestMessage(authRequest)
            .then((authRequest) => {
                console.log("Sending auth_request:", authRequest);
                brokerWs.send(authRequest);
            })
            .catch((error) => {
                console.error("Error creating auth request:", error);
                cleanup();
                reject(new Error(`Failed to create auth request: ${error.message}`));
            });
    });
}

// Handles messages received from the broker
export function handleBrokerMessage(message: any): void {
    try {
        // Log the raw message for debugging
        console.log("Received message from broker:", message);

        const requestId = message.res[0];
        const method = message.res[1];
        const payload = message.res[2];
        // Handle RPC format (new format with 'res' array)
        if (message.res && Array.isArray(message.res)) {
            // Check if it's an error message
            if (method === "error") {
                console.log("Received error from broker:", payload);

                // Check if it's a response to a pending request
                if (typeof requestId === "string" || typeof requestId === "number") {
                    const pendingRequest = getPendingRequest(requestId.toString());
                    if (pendingRequest) {
                        const { reject, timeout } = pendingRequest;
                        clearTimeout(timeout);
                        clearPendingRequest(requestId.toString());

                        const errorMessage = payload && payload[0]?.error ? payload[0].error : "Unknown error";
                        reject(new Error(errorMessage));
                    }
                }
                return;
            }
            else if (method === "get_channels" && payload.length === 0) {
                createBrokerChannel(client);
            }

            // Handle successful response to a pending request
            if (typeof requestId === "string" || typeof requestId === "number") {
                const pendingRequest = getPendingRequest(requestId.toString());
                if (pendingRequest) {
                    const { resolve, timeout } = pendingRequest;
                    clearTimeout(timeout);
                    clearPendingRequest(requestId.toString());

                    // For successful responses, return the result data (typically in res[2])
                    const resultData = payload || [];
                    resolve(resultData.length === 1 ? resultData[0] : resultData);
                    return;
                }
            }
        }

        // Legacy JSON-RPC response format (should rarely be used with new broker)
        if (message.id && typeof message.id === "string") {
            const pendingRequest = getPendingRequest(message.id);
            if (pendingRequest) {
                const { resolve, reject, timeout } = pendingRequest;
                clearTimeout(timeout);
                clearPendingRequest(message.id);

                if (message.error) {
                    reject(new Error(message.error.message || "Unknown error"));
                } else {
                    resolve(message.result || message);
                }
                return;
            }
        }

        // Handle other message types like notifications
        // (in a real implementation, you might want to emit events for these)
    } catch (error) {
        console.error("Error handling broker message:", error);
    }
}

// Check authentication status
export function isAuthenticatedWithBroker(): boolean {
    return isAuthenticated;
}

// Re-export the authentication function for external use
export { authenticateWithBroker };

// Sends a request to the broker and returns a promise
export async function sendToBroker(request: any): Promise<any> {
    // Check authentication first before creating the Promise
    if (!isAuthenticated && !(request.req && request.req[1] === "auth_request") && !(request.req && request.req[1] === "auth_verify")) {
        try {
            console.log("Not authenticated with broker, authenticating first...");
            await authenticateWithBroker();
        } catch (error) {
            console.error("Authentication failed:", error);
            if (error instanceof Error) {
                throw new Error(`Authentication failed: ${error.message}`);
            } else {
                throw new Error('Authentication failed: Unknown error');
            }
        }
    }

    return new Promise((resolve, reject) => {
        const brokerWs = getBrokerWebSocket();
        if (!brokerWs || brokerWs.readyState !== WebSocket.OPEN) {
            console.error("WebSocket not connected or not open. State:", brokerWs?.readyState);
            reject(new Error("Not connected to broker"));
            return;
        }

        console.log("Sending request to broker:", {
            method: request.req?.[1],
            requestId: request.req?.[0],
            isAuthenticated,
            wsState: brokerWs.readyState
        });

        // Prepare the request using a Promise chain
        const prepareRequest = async (): Promise<{ req: any; requestId: string | number }> => {
            let requestId: string | number;
            let preparedRequest = request;

            // Check if the request is in the new format
            if (request.req && Array.isArray(request.req)) {
                requestId = request.req[0] || Date.now();
                preparedRequest.req[0] = requestId;

                // If the signature is empty or missing, add it
                if (!preparedRequest.sig || preparedRequest.sig.length === 0 || !preparedRequest.sig[0]) {
                    const signature = await signRpcRequest(preparedRequest.req);
                    preparedRequest.sig = [signature];
                }
            } else {
                // Legacy format - convert to new format
                requestId = request.id || `req-${Date.now()}`;
                const reqData = [requestId, request.method, request.params ? [request.params] : [], Date.now()];

                // Sign the request
                const signature = await signRpcRequest(reqData);

                preparedRequest = {
                    req: reqData,
                    sig: [signature],
                };
            }

            return { req: preparedRequest, requestId };
        };

        // Execute the async preparation outside the Promise executor
        prepareRequest()
            .then(({ req, requestId }) => {
                // Convert requestId to string for tracking
                const requestIdStr = requestId.toString();

                const timeout = setTimeout(() => {
                    console.error("Request timed out:", {
                        requestId: requestIdStr,
                        method: req.req[1],
                        isAuthenticated,
                        wsState: brokerWs.readyState
                    });
                    clearPendingRequest(requestIdStr);
                    reject(new Error("Request timeout"));
                }, 10000); // 10 second timeout

                addPendingRequest(requestIdStr, resolve, reject, timeout);
                brokerWs.send(JSON.stringify(req));
            })
            .catch((error) => {
                console.error("Failed to prepare request:", error);
                reject(new Error(`Failed to prepare request: ${error.message}`));
            });
    });
}

// Creates an application session in the broker
export async function createAppSession(participantA: Hex, participantB: Hex): Promise<string> {
    // Ensure we're authenticated before creating an app session
    if (!isAuthenticated) {
        try {
            await authenticateWithBroker();
        } catch (error) {
            console.error(`Authentication failed before creating app session:`, error);
            if (error instanceof Error) {
                throw new Error(`Authentication required to create app session: ${error.message}`);
            } else {
                throw new Error('Authentication required to create app session: Unknown error');
            }
        }
    }

    // Get the server's wallet address
    const signer = createEthersSigner(WALLET_PRIVATE_KEY);
    if (!signer.address) {
        throw new Error("Server wallet address not found");
    }

    // Prepare the request object
    const participants = [participantA, participantB, signer.address as Hex];
    console.log("[createAppSession] Creating app session with:", {
        participants,
        signerAddress: signer.address
    });

    const requestId = Date.now();
    const appDefinition: AppDefinition = {
        protocol: DEFAULT_PROTOCOL,
        participants,
        weights: DEFAULT_WEIGHTS,
        quorum: DEFAULT_QUORUM,
        challenge: 0,
        nonce: Date.now(),
    };
    const params: CreateAppSessionRequest[] = [{
        definition: appDefinition,
        allocations: participants.map((participant) => ({
            participant,
            asset: CONTRACT_ADDRESSES.tokenAddress as Hex,
            amount: "0",
        }))
    }]
    const timestamp = Date.now();

    // Create the request with properly formatted parameters
    const request: { req: [number, string, CreateAppSessionRequest[], number] } = {
        req: [requestId, "create_app_session", params, timestamp],
    };

    console.log("[createAppSession] Sending request:", request);
    const result = await sendToBroker(request);
    const appId = result.app_session_id || (typeof result[0] === "object" ? result[0].app_session_id : null);
    console.log(`[createAppSession] Created app session ${appId}`);
    return appId;
}

// Closes an application session in the broker
export async function closeAppSession(appId: Hex, participantA: Hex, participantB: Hex): Promise<void> {
    // Ensure we're authenticated before closing an app session
    if (!isAuthenticated) {
        try {
            await authenticateWithBroker();
        } catch (error) {
            console.error(`Authentication failed before closing app session:`, error);
            if (error instanceof Error) {
                throw new Error(`Authentication required to close app session: ${error.message}`);
            } else {
                throw new Error('Authentication required to close app session: Unknown error');
            }
        }
    }

    // Get the server's wallet address
    const signer = createEthersSigner(WALLET_PRIVATE_KEY);
    if (!signer.address) {
        throw new Error("Server wallet address not found");
    }

    // Verify the app session exists before trying to close it
    try {
        const requestId = Date.now();
        const timestamp = Date.now();
        const request: { req: [number, string, { app_session_id: string }[], number] } = {
            req: [requestId, "get_app_definition", [{ app_session_id: appId }], timestamp]
        };
        console.log("[closeAppSession] Verifying app session exists:", appId);
        await sendToBroker(request);
        console.log("[closeAppSession] App session exists, proceeding with close");
    } catch (error) {
        console.error(`[closeAppSession] App session ${appId} not found or already closed:`, error);
        if (error instanceof Error) {
            throw new Error(`App session ${appId} not found or already closed: ${error.message}`);
        } else {
            throw new Error(`App session ${appId} not found or already closed: Unknown error`);
        }
    }

    // Prepare the request
    const requestId = Date.now();
    const params: CloseAppSessionRequest[] = [{
        app_session_id: appId,
        allocations: [participantA, participantB, signer.address].map((participant) => ({
            participant,
            asset: CONTRACT_ADDRESSES.tokenAddress as Hex,
            amount: "0",
        })),
    }];
    const timestamp = Date.now();

    // Create the request with properly formatted parameters
    const request: { req: [number, string, CloseAppSessionRequest[], number] } = {
        req: [requestId, "close_app_session", params, timestamp]
    };

    console.log("[closeAppSession] Sending close request:", request);
    await sendToBroker(request);
    console.log(`[closeAppSession] Closed app session ${appId}`);
}

// Helper function to sign state data with the server's private key
export async function signStateData(stateData: string): Promise<{ signature: string; address: Hex }> {
    const signer = createEthersSigner(WALLET_PRIVATE_KEY);
    return {
        signature: await signer.sign(stateData as unknown as RequestData),
        address: signer.address as Hex,
    };
}

/**
 * Interface for a wallet signer that can sign messages
 */
export interface WalletSigner {
    /** Public key in hexadecimal format */
    publicKey: string;
    /** Optional Ethereum address derived from the public key */
    address?: Hex;
    /** Function to sign a message and return a hex signature */
    sign: MessageSigner;
}

/**
 * Creates a signer from a private key using ethers.js
 *
 * @param privateKey - The private key to create the signer from
 * @returns A WalletSigner object that can sign messages
 * @throws Error if signer creation fails
 */
export function createEthersSigner(privateKey: string): WalletSigner {
    try {
        // Create ethers wallet from private key
        const wallet = new ethers.Wallet(privateKey);
        return {
            publicKey: wallet.publicKey,
            address: wallet.address as Hex,
            sign: async (payload: RequestData | ResponsePayload): Promise<Hex> => {
                try {
                    const messageBytes = ethers.utils.arrayify(ethers.utils.id(JSON.stringify(payload)));
                    const flatSignature = wallet._signingKey().signDigest(messageBytes);
                    const signature = ethers.utils.joinSignature(flatSignature);
                    return signature as Hex;
                } catch (error) {
                    console.error("Error signing message:", error);
                    throw error;
                }
            },
        };
    } catch (error) {
        console.error("Error creating ethers signer:", error);
        throw error;
    }
}

// Helper function to sign RPC request data for the broker
export async function signRpcRequest(requestData: any[]): Promise<string> {
    const signer = createEthersSigner(WALLET_PRIVATE_KEY);
    return signer.sign(requestData as RequestData);
}

// Verify a signature against a message and expected signer
export function verifySignature(message: string, signature: string, expectedAddress: string): boolean {
    try {
        // Use standard Ethereum message verification
        const recoveredAddress = ethers.utils.verifyMessage(message, signature);

        // Check if the recovered address matches the expected address
        return recoveredAddress.toLowerCase() === expectedAddress.toLowerCase();
    } catch (error) {
        console.error("Error verifying signature:", error);
        return false;
    }
}

// Create auth_verify message with EIP-712 signature
async function createAuthVerifyWithEIP712(
    address: string,
    challenge: string,
    sessionKey: string,
    appName: string,
    allowances: Array<{ asset: string; amount: string }>
): Promise<string> {
    const domain = {
        name: appName,
    };

    const types = {
        AuthVerify: [
            { name: "address", type: "address" },
            { name: "challenge", type: "string" },
            { name: "session_key", type: "address" },
            { name: "allowances", type: "Allowance[]" },
        ],
        Allowance: [
            { name: "asset", type: "string" },
            { name: "amount", type: "uint256" },
        ],
    };

    const value = {
        address: address,
        challenge: challenge,
        session_key: sessionKey,
        allowances: allowances.map(a => ({
            asset: a.asset,
            amount: ethers.BigNumber.from(a.amount || "0"),
        })),
    };

    // Create the wallet to sign
    const wallet = new ethers.Wallet(WALLET_PRIVATE_KEY);
    
    // Sign the typed data
    const signature = await wallet._signTypedData(domain, types, value);

    // Create the auth_verify request
    const requestId = Date.now();
    const timestamp = getCurrentTimestamp();
    const request = NitroliteRPC.createRequest(requestId, "auth_verify", [{ challenge }], timestamp);
    
    // Add the EIP-712 signature
    request.sig = [signature as Hex];

    return JSON.stringify(request);
}
