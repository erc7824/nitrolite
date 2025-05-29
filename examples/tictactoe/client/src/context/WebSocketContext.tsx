import React, { createContext, useContext, useState, useEffect, useCallback, useMemo, type ReactNode, useRef } from "react";
import { WebSocketClient, createWebSocketClient, type WSStatus, type WalletSigner, getAddressFromPublicKey } from "../websocket";
import type { Channel } from "@erc7824/nitrolite";
import APP_CONFIG from "./app";
import { generateKeyPair, createEthersSigner } from "./createSigner";
import { WalletStore } from "../store";

// Interface for key pairs
export interface CryptoKeypair {
    privateKey: string;
    publicKey?: string;
    address?: string;
}

const CRYPTO_KEYPAIR_KEY = "crypto_keypair";
const WS_URL = APP_CONFIG.WEBSOCKET.URL;

interface WebSocketContextProps {
    client: WebSocketClient | null;
    status: WSStatus;
    keyPair: CryptoKeypair | null;
    wsChannel: Channel | null;
    currentNitroliteChannel: Channel | null;
    isConnected: boolean;
    hasKeys: boolean;
    generateKeys: () => Promise<CryptoKeypair | null>;
    connect: () => Promise<boolean>;
    disconnect: () => void;
    setNitroliteChannel: (channel: Channel) => void;
    clearKeys: () => void;
    sendPing: () => Promise<void>;
    sendRequest: (payload: string) => Promise<unknown>;
}

const WebSocketContext = createContext<WebSocketContextProps | undefined>(undefined);

export const WebSocketProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
    const [status, setStatus] = useState<WSStatus>("disconnected");
    const [keyPair, setKeyPair] = useState<CryptoKeypair | null>(null);
    const [currentSigner, setCurrentSigner] = useState<WalletSigner | null>(null);
    const [wsChannel, setWsChannel] = useState<Channel | null>(null);
    const [currentNitroliteChannel, setCurrentNitroliteChannel] = useState<Channel | null>(null);
    const clientRef = useRef<WebSocketClient | null>(null);

    useEffect(() => {
        if (typeof window !== "undefined") {
            const savedKeys = localStorage.getItem(CRYPTO_KEYPAIR_KEY);

            if (savedKeys) {
                try {
                    const parsed = JSON.parse(savedKeys) as CryptoKeypair;

                    if (parsed.publicKey && !parsed.address) {
                        parsed.address = getAddressFromPublicKey(parsed.publicKey);
                        localStorage.setItem(CRYPTO_KEYPAIR_KEY, JSON.stringify(parsed));
                    }
                    setKeyPair(parsed);
                    console.log("Loaded existing keys from storage");
                    // @ts-ignore
                } catch (e) {
                    console.error("Failed to parse saved keys - will generate new ones");
                    localStorage.removeItem(CRYPTO_KEYPAIR_KEY);
                    generateNewKeysAndStore();
                }
            } else {
                console.log("No saved keys found - generating new ones");
                generateNewKeysAndStore();
            }
        }
    }, []);

    const generateNewKeysAndStore = async () => {
        try {
            const newKeyPair = await generateKeyPair();
            setKeyPair(newKeyPair);
            localStorage.setItem(CRYPTO_KEYPAIR_KEY, JSON.stringify(newKeyPair));
            console.log("Generated and stored new crypto keys");
            return newKeyPair;
        } catch (error) {
            console.error("Error generating and storing keys:", error);
            return null;
        }
    };

    const generateKeys = useCallback(async () => {
        try {
            if (typeof window !== "undefined") {
                const savedKeys = localStorage.getItem(CRYPTO_KEYPAIR_KEY);

                if (savedKeys) {
                    try {
                        const parsed = JSON.parse(savedKeys) as CryptoKeypair;

                        if (parsed && typeof parsed.privateKey === "string" && typeof parsed.publicKey === "string") {
                            if (parsed.publicKey && !parsed.address) {
                                parsed.address = getAddressFromPublicKey(parsed.publicKey);
                                localStorage.setItem(CRYPTO_KEYPAIR_KEY, JSON.stringify(parsed));
                            }
                            setKeyPair(parsed);
                            const signer = createEthersSigner(parsed.privateKey);

                            setCurrentSigner(signer);
                            return parsed;
                        }
                    } catch (e) {
                        // Could not parse, fall through and generate keys.
                        console.error("Failed to parse saved keys during generateKeys:", e);
                        localStorage.removeItem(CRYPTO_KEYPAIR_KEY);
                    }
                }
            }
            // If no valid keys in storage, generate new ones.
            const newKeyPair = await generateKeyPair();

            setKeyPair(newKeyPair);
            if (typeof window !== "undefined") {
                localStorage.setItem(CRYPTO_KEYPAIR_KEY, JSON.stringify(newKeyPair));
            }
            const newSigner = createEthersSigner(newKeyPair.privateKey);

            setCurrentSigner(newSigner);
            console.log("Generated new cryptographic keys");
            return newKeyPair;
        } catch (error) {
            const errorMsg = `Error generating keys: ${error instanceof Error ? error.message : String(error)}`;
            console.error(errorMsg);
            return null;
        }
    }, []);

    const clearKeys = useCallback(() => {
        if (typeof window !== "undefined") {
            localStorage.removeItem(CRYPTO_KEYPAIR_KEY);
        }
        setKeyPair(null);
        setCurrentSigner(null);
        if (clientRef.current?.isConnected) {
            clientRef.current.close();
        }
        console.log("Cleared cryptographic keys");
    }, []);

    useEffect(() => {
        if (keyPair?.privateKey && !currentSigner) {
            try {
                const signer = createEthersSigner(keyPair.privateKey);
                setCurrentSigner(signer);
                console.log("Initialized signer from keys");
            } catch (e) {
                console.error(`Failed to create signer: ${e instanceof Error ? e.message : String(e)}`);
            }
        }
    }, [keyPair, currentSigner]);

    useEffect(() => {
        if (currentSigner && !clientRef.current) {
            const newClient = createWebSocketClient(WS_URL, currentSigner, {
                autoReconnect: true,
                reconnectDelay: 1000,
                maxReconnectAttempts: 5,
                requestTimeout: 10000,
            });

            clientRef.current = newClient;

            newClient.onStatusChange((newStatus) => {
                setStatus(newStatus);
                if (newStatus === "connected") {
                    setWsChannel(newClient.currentSubscribedChannel);
                    setCurrentNitroliteChannel(newClient.currentNitroliteChannel);
                } else if (newStatus === "disconnected" || newStatus === "reconnect_failed") {
                    setWsChannel(null);
                    setCurrentNitroliteChannel(null);
                }
            });

            newClient.onError((error) => {
                console.error(`WebSocket error: ${error.message}`);
            });

            newClient.onMessage((message) => {
                const hasType = (msg: unknown): msg is { type: unknown } => typeof msg === "object" && msg !== null && "type" in msg;
                const messageType = hasType(message) ? (typeof message.type === "string" ? message.type : String(message.type)) : "unknown";

                console.log(`Received message (type: ${messageType})`);
            });

            // Don't automatically connect - wait for MetaMask to be connected first
            console.log("WebSocket client initialized, waiting for MetaMask connection...");
        }

        return () => {
            if (clientRef.current) {
                clientRef.current.close();
                clientRef.current = null;
                setStatus("disconnected");
                setWsChannel(null);
                setCurrentNitroliteChannel(null);
            }
        };
    }, [currentSigner]);

    const connect = useCallback(async () => {
        if (!clientRef.current) {
            console.error("Cannot connect: WebSocket client not initialized (no signer?)");
            return false;
        }
        if (clientRef.current.isConnected) {
            console.log("Already connected");
            return true;
        }

        // Check if MetaMask wallet is connected first
        const walletClient = WalletStore.getWalletClient();
        if (!walletClient?.account?.address) {
            console.log("Cannot connect to WebSocket: MetaMask wallet not connected. Please connect MetaMask first.");
            return false;
        }

        try {
            console.log("MetaMask connected, connecting to WebSocket server...");
            await clientRef.current.connect();
            console.log("WebSocket connection established");
            return true;
        } catch (error) {
            const errorMsg = `Connection error: ${error instanceof Error ? error.message : String(error)}`;
            console.error(errorMsg);
            return false;
        }
    }, []);

    // Effect to automatically connect WebSocket when MetaMask becomes available
    useEffect(() => {
        const checkAndConnect = async () => {
            const walletClient = WalletStore.getWalletClient();
            if (walletClient?.account?.address && clientRef.current && !clientRef.current.isConnected && status === "disconnected") {
                console.log("MetaMask connected, attempting WebSocket connection...");
                await connect();
            }
        };

        // Check immediately
        checkAndConnect();

        // Set up an interval to check periodically (in case we miss the wallet connection)
        const interval = setInterval(checkAndConnect, 1000);

        return () => clearInterval(interval);
    }, [connect, status]);

    const disconnect = useCallback(() => {
        if (clientRef.current) {
            clientRef.current.close();
            console.log("WebSocket connection closed");
        } else {
            console.error("Cannot disconnect: WebSocket client not initialized");
        }
    }, []);

    const setNitroliteChannel = useCallback((nitroliteChannel: Channel) => {
        if (!clientRef.current) {
            console.error("Cannot set Nitrolite channel: Client not initialized");
            return;
        }
        setCurrentNitroliteChannel(nitroliteChannel);
        clientRef.current.setNitroliteChannel(nitroliteChannel);
        console.log(`Set Nitrolite channel: ${JSON.stringify(nitroliteChannel).slice(0, 50)}...`);
    }, []);

    const sendPing = useCallback(async () => {
        if (!clientRef.current?.isConnected) {
            console.error("Cannot ping: Not connected");
            return;
        }
        try {
            console.log("Sending ping to server...");
            await clientRef.current.ping();
            console.log("Ping successful");
        } catch (error) {
            console.error(`Ping error: ${error instanceof Error ? error.message : String(error)}`);
        }
    }, []);

    // @ts-ignore
    const sendRequest = useCallback(async (signedRequest) => {
        if (!clientRef.current?.isConnected) {
            const errorMsg = `Cannot send request: Not connected`;
            console.error(errorMsg);
            throw new Error("WebSocket not connected");
        }

        try {
            console.log("Sending request to server...");
            const response = await clientRef.current.sendRequest(signedRequest);
            console.log("Request successful");
            return response;
        } catch (error) {
            const errorMsg = `Request error: ${error instanceof Error ? error.message : String(error)}`;
            console.error(errorMsg);
            throw error;
        }
    }, []);

    const value = useMemo(
        () => ({
            client: clientRef.current,
            status,
            keyPair,
            wsChannel,
            currentNitroliteChannel,
            isConnected: status === "connected",
            hasKeys: !!keyPair,
            generateKeys,
            connect,
            disconnect,
            setNitroliteChannel,
            clearKeys,
            sendPing,
            sendRequest,
        }),
        [
            status,
            keyPair,
            wsChannel,
            currentNitroliteChannel,
            generateKeys,
            connect,
            disconnect,
            setNitroliteChannel,
            clearKeys,
            sendPing,
            sendRequest,
        ]
    );

    return <WebSocketContext.Provider value={value}>{children}</WebSocketContext.Provider>;
};

export const useWebSocketContext = (): WebSocketContextProps => {
    const context = useContext(WebSocketContext);

    if (context === undefined) {
        throw new Error("useWebSocketContext must be used within a WebSocketProvider");
    }
    return context;
};
