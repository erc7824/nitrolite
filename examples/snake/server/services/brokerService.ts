import WebSocket from "ws";
import { ethers } from "ethers";
import {
    createAuthRequestMessage,
    createAuthVerifyMessage,
    createAuthVerifyMessageWithJWT,
    createEIP712AuthMessageSigner,
    RequestData,
    ResponsePayload,
    MessageSigner,
    AppDefinition,
    NitroliteRPC,
    createGetLedgerBalancesMessage,
    CreateAppSessionRequest,
    createCloseAppSessionMessage,
    CloseAppSessionRequest,
    parseRPCResponse,
    RPCResponse,
} from "@erc7824/nitrolite";
import { BROKER_WS_URL, WALLET_PRIVATE_KEY } from "../config/index.ts";
import { setBrokerWebSocket, getBrokerWebSocket, addPendingRequest, getPendingRequest, clearPendingRequest } from "./stateService.ts";
import { Hex, createWalletClient, http } from "viem";
import { privateKeyToAccount } from "viem/accounts";
import { polygon } from "viem/chains";

import util from 'util';
util.inspect.defaultOptions.depth = null;

export const DEFAULT_PROTOCOL = "app_snake_nitrolite";
const DEFAULT_WEIGHTS: number[] = [0, 0, 100]; // Alice: 0, Bob: 0, Server: 100
const DEFAULT_QUORUM: number = 100; // server alone decides the outcome

// Flag to indicate if we've authenticated with the broker
let isAuthenticated = false;

// Store JWT token at file level for reuse
let jwtToken: string | null = null;

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
        jwtToken = null; // Clear JWT token on disconnect
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


    const expire = String(Math.floor(Date.now() / 1000) + 24 * 60 * 60);

    const authMessage = {
        wallet: serverAddress,
        participant: serverAddress,
        app_name: 'Snake Game',
        expire: expire,
        scope: 'snake-game',
        application: serverAddress,
        allowances: [
            {
                symbol: 'usdc',
                amount: '0',
            },
        ],
    };

    return new Promise(async (resolve, reject) => {
        let authTimeout: NodeJS.Timeout;

        // Clean up function to remove listeners and clear timeout
        const cleanup = () => {
            brokerWs.removeListener("message", authMessageHandler);
            clearTimeout(authTimeout);
        };

        // Create a one-time message handler for authentication
        const authMessageHandler = async (data: WebSocket.RawData) => {
            try {
                let message: RPCResponse;
                try {
                    message = parseRPCResponse(data.toString());
                } catch (error) {
                    console.warn("Error parsing auth message from broker, skipping:", error);
                    return;
                }
                console.log("Auth process message received:", message);

                // Check for auth_challenge response (response to our auth_request)
                if (message.method === "auth_challenge") {
                    console.log("Received auth_challenge, preparing EIP-712 auth_verify...");

                    try {
                        // Step 2: Create EIP-712 signing function for challenge verification
                        const account = privateKeyToAccount(WALLET_PRIVATE_KEY as Hex);
                        const walletClient = createWalletClient({
                            account,
                            chain: polygon,
                            transport: http()
                        });

                        if (!walletClient) {
                            throw new Error('No wallet client available for EIP-712 signing');
                        }

                        console.log('Creating EIP-712 signing function...');
                        // @ts-ignore
                        const eip712SigningFunction = createEIP712AuthMessageSigner(walletClient, {
                            scope: authMessage.scope,
                            application: authMessage.application,
                            participant: authMessage.participant,
                            expire: authMessage.expire,
                            allowances: authMessage.allowances.map((allowance) => ({
                                asset: allowance.symbol,
                                amount: allowance.amount.toString(),
                            })),
                        }, getAuthDomain());

                        // Create and send verification message with EIP-712 signature
                        console.log('Calling createAuthVerifyMessage...');
                        const authVerify = await createAuthVerifyMessage(eip712SigningFunction, message);

                        console.log('Sending auth_verify with EIP-712 signature');
                        brokerWs.send(authVerify);
                        console.log('auth_verify sent successfully');

                        // Send additional requests
                        const getBalances = await createGetLedgerBalancesMessage(
                            signer.sign,
                            serverAddress
                        );
                        brokerWs.send(getBalances);
                        await getChannels();
                    } catch (eip712Error) {
                        console.error('Error creating EIP-712 auth_verify:', eip712Error);
                        console.error('Error stack:', (eip712Error as Error).stack);

                        cleanup();
                        reject(new Error(`EIP-712 auth_verify failed: ${(eip712Error as Error).message}`));
                        return;
                    }
                }
                // Check for auth_verify success response
                else if (message.method === "auth_verify") {
                    console.log("Authentication successful");

                    // If response contains a JWT token, store it
                    if (message.params.jwtToken) {
                        console.log('JWT token received:', message.params.jwtToken);
                        jwtToken = message.params.jwtToken;
                    }

                    isAuthenticated = true;
                    cleanup();
                    resolve();
                }
                // Check for error responses
                else if (message.method === "error") {
                    const errorMsg = message.params.error || 'Authentication failed';
                    console.error('Authentication failed:', errorMsg);

                    // Check if this is a JWT authentication failure and fallback to signer auth
                    const errorString = String(errorMsg).toLowerCase();
                    if (errorString.includes('jwt') || errorString.includes('token') || errorString.includes('invalid') || errorString.includes('expired')) {
                        console.warn('JWT authentication failed on server, attempting fallback to signer authentication');
                        jwtToken = null; // Clear invalid JWT token

                        try {
                            // Restart authentication with signer
                            const fallbackAuthRequest = await createAuthRequestMessage(authMessage);

                            console.log('Sending fallback auth_request with signer:', fallbackAuthRequest);
                            brokerWs.send(fallbackAuthRequest);
                            // Reset timeout for the fallback attempt
                            clearTimeout(authTimeout);
                            authTimeout = setTimeout(() => {
                                cleanup();
                                reject(new Error("Authentication timeout"));
                            }, 15000);
                            return; // Continue listening for the fallback response
                        } catch (fallbackError) {
                            console.error('Fallback to signer authentication failed:', fallbackError);
                            cleanup();
                            reject(new Error(`Both JWT and signer authentication failed: ${fallbackError}`));
                            return;
                        }
                    }

                    jwtToken = null; // Clear JWT token on auth failure
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

        // Set timeout for auth process
        authTimeout = setTimeout(() => {
            cleanup();
            reject(new Error("Authentication timeout"));
        }, 15000); // 15 second timeout

        // Add temporary listener for authentication messages
        brokerWs.on("message", authMessageHandler);

        // Step 1: Send auth_request with JWT token if available
        console.log('Starting authentication with:');
        console.log('- Server wallet address:', serverAddress);
        console.log('- JWT token if available, otherwise EIP-712 signature for auth_verify challenge');

        try {
            let authRequest: string;
            let usingJWT = false;

            if (jwtToken) {
                console.log('JWT token found, attempting JWT authentication:', jwtToken);
                try {
                    authRequest = await createAuthVerifyMessageWithJWT(jwtToken);
                    usingJWT = true;
                } catch (jwtError) {
                    console.warn('JWT auth failed, falling back to signer authentication:', jwtError);
                    // Clear invalid JWT token
                    jwtToken = null;
                    authRequest = await createAuthRequestMessage(authMessage);
                    usingJWT = false;
                }
            } else {
                console.log('No JWT token found, proceeding with challenge-response authentication');
                authRequest = await createAuthRequestMessage(authMessage);
                usingJWT = false;
            }

            console.log(`Sending auth_request (${usingJWT ? 'JWT' : 'challenge-response'}):`, authRequest);
            brokerWs.send(authRequest);
        } catch (requestError) {
            console.error('Error creating auth_request:', requestError);
            cleanup();
            reject(new Error(`Failed to create auth_request: ${(requestError as Error).message}`));
        }
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

        // Handle ping messages
        if (method === "ping") {
            console.log("Received ping from broker, sending pong");
            const brokerWs = getBrokerWebSocket();
            if (brokerWs && brokerWs.readyState === WebSocket.OPEN) {
                brokerWs.send(JSON.stringify({ type: "pong" }));
            }
            return;
        }

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
                throw new Error("No channels found. Please open a channel at apps.yellow.com");
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
        allocations: participants.map((participant, index) => ({
            participant,
            asset: "usdc",
            amount: index < 2 ? "0.00001" : "0", // Players get 0.00001, server gets 0
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

// Closes an application session in the broker with server signature only
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

    // Create close message and sign with server
    const params: CloseAppSessionRequest[] = [{
        app_session_id: appId,
        allocations: [participantA, participantB, signer.address].map((participant, index) => ({
            participant,
            asset: "usdc",
            amount: index < 2 ? "0.00001" : "0", // Players get 0.00001, server gets 0
        }))
    }]
    const closeRequestData = await createCloseAppSessionMessage(signer.sign, params);
    const req = JSON.parse(closeRequestData);
    const serverSignature = await signer.sign(req);

    // Create the signed request with server signature only
    const signedRequest = {
        req: closeRequestData,
        sig: [serverSignature]
    };

    console.log("[closeAppSession] Sending close request with server signature:", signedRequest);
    await sendToBroker(signedRequest);
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
            publicKey: wallet.address,
            address: wallet.address as Hex,
            sign: async (data: RequestData | ResponsePayload): Promise<Hex> => {
                try {
                    const messageStr = typeof data === "string" ? data : JSON.stringify(data);

                    const digestHex = ethers.id(messageStr);
                    const messageBytes = ethers.getBytes(digestHex);

                    const { serialized: signature } = wallet.signingKey.sign(messageBytes);
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
        const recoveredAddress = ethers.verifyMessage(message, signature);

        // Check if the recovered address matches the expected address
        return recoveredAddress.toLowerCase() === expectedAddress.toLowerCase();
    } catch (error) {
        console.error("Error verifying signature:", error);
        return false;
    }
}

/**
 * EIP-712 domain and types for auth_verify challenge
 */
const getAuthDomain = () => {
    return {
        name: 'Snake Game',
    };
};


