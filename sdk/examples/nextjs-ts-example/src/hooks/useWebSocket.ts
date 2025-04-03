import { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import {
    createWebSocketClient,
    createEthersSigner,
    generateKeyPair,
    WalletSigner,
    CryptoKeypair,
    getAddressFromPublicKey,
    WebSocketClient,
} from '@/websocket';
import { Channel } from '@/types';
import { useMessageService } from './useMessageService';

/**
 * Custom hook to manage WebSocket connection and operations
 */
export function useWebSocket(url: string) {
    const [status, setStatus] = useState<string>('disconnected');
    const [keyPair, setKeyPair] = useState<CryptoKeypair | null>(() => {
        // Try to load keys from localStorage on initial render
        if (typeof window !== 'undefined') {
            const savedKeys = localStorage.getItem('crypto_keypair');

            if (savedKeys) {
                try {
                    const parsed = JSON.parse(savedKeys) as CryptoKeypair;

                    // If missing address property, derive it from the public key
                    if (parsed.publicKey && !parsed.address) {
                        parsed.address = getAddressFromPublicKey(parsed.publicKey);
                        localStorage.setItem('crypto_keypair', JSON.stringify(parsed));
                    }

                    return parsed;
                } catch (e) {
                    console.error('Failed to parse saved keys:', e);
                }
            }
        }
        return null;
    });

    const [currentSigner, setCurrentSigner] = useState<WalletSigner | null>(null);
    const [currentChannel, setCurrentChannel] = useState<Channel | null>(null);

    // Use our message service
    const { setStatus: setMessageStatus, addSystemMessage, addErrorMessage } = useMessageService();

    // Update both statuses
    const updateStatus = useCallback(
        (newStatus: string) => {
            setStatus(newStatus);
            setMessageStatus(newStatus);

            // Add a system message about status change
            addSystemMessage(`Connection status changed to: ${newStatus}`);
        },
        [setMessageStatus, addSystemMessage],
    );

    // Initialize signer from existing keys if available
    useEffect(() => {
        if (keyPair?.privateKey && !currentSigner) {
            try {
                setCurrentSigner(createEthersSigner(keyPair.privateKey));
            } catch (e) {
                console.error('Failed to create signer from saved keys:', e);
            }
        }
    }, [keyPair, currentSigner]);

    // Create WebSocket client with current signer
    const client = useMemo(() => {
        if (!currentSigner) return null;
        return createWebSocketClient(url, currentSigner, {
            autoReconnect: true,
            reconnectDelay: 1000,
            maxReconnectAttempts: 5,
            requestTimeout: 10000,
        });
    }, [url, currentSigner]);

    const clientRef = useRef<WebSocketClient | null>(null);

    // Update the client reference when the client changes
    useEffect(() => {
        clientRef.current = client;
    }, [client]);

    // Initialize WebSocket event listeners
    useEffect(() => {
        const client = clientRef.current;

        if (!client) {
            addSystemMessage('WebSocket client not initialized');
            return;
        }

        addSystemMessage('Setting up WebSocket event listeners');

        // Set up status change handler
        client.onStatusChange(updateStatus);

        // Set up error handler
        client.onError((error) => {
            addErrorMessage(`WebSocket error: ${error.message}`);
        });

        // Set up message handler
        client.onMessage((message) => {
            addSystemMessage(`Received message of type: ${message.type || 'unknown'}`);
        });

        // Add initial system message
        addSystemMessage('WebSocket listeners initialized successfully');

        return () => {
            addSystemMessage('Cleaning up WebSocket connection');
            client.close();
        };
    }, [updateStatus, addSystemMessage, addErrorMessage]);

    // Generate a new key pair
    const generateKeys = useCallback(async () => {
        try {
            const newKeyPair = await generateKeyPair();

            setKeyPair(newKeyPair);

            // Save to localStorage
            if (typeof window !== 'undefined') {
                localStorage.setItem('crypto_keypair', JSON.stringify(newKeyPair));
            }

            // Create a new signer with the generated private key
            const newSigner = createEthersSigner(newKeyPair.privateKey);

            setCurrentSigner(newSigner);

            return newKeyPair;
        } catch (error) {
            const errorMsg = `Error generating keys: ${error instanceof Error ? error.message : String(error)}`;

            addErrorMessage(errorMsg);
            return null;
        }
    }, [addErrorMessage]);

    // Connect to WebSocket
    const connect = useCallback(async () => {
        if (!keyPair) {
            const errorMsg = 'No key pair available for connection';

            addSystemMessage(errorMsg);
            throw new Error(errorMsg);
        }

        try {
            addSystemMessage('Attempting to connect to WebSocket...');

            await clientRef.current.connect();

            addSystemMessage('WebSocket connected successfully');
            return true;
        } catch (error) {
            const errorMsg = `Connection error: ${error instanceof Error ? error.message : String(error)}`;

            addErrorMessage(errorMsg);
            throw error;
        }
    }, [keyPair, addSystemMessage, addErrorMessage]);

    // Disconnect from WebSocket
    const disconnect = useCallback(() => {
        clientRef.current?.close();
    }, []);

    // Subscribe to a channel
    const subscribeToChannel = useCallback(async (channel: Channel) => {
        if (!clientRef.current?.isConnected) return;

        try {
            await clientRef.current.subscribe(channel);
            setCurrentChannel(channel);
        } catch (error) {
            console.error('Subscribe error:', error);
        }
    }, []);

    // Send a message to the current channel
    const sendMessage = useCallback(async (message: string) => {
        if (!clientRef.current?.isConnected || !clientRef.current.currentSubscribedChannel) return;

        try {
            await clientRef.current.publishMessage(message);
        } catch (error) {
            console.error('Send error:', error);
        }
    }, []);

    // Send a ping request
    const sendPing = useCallback(async () => {
        if (!clientRef.current?.isConnected) return;

        try {
            await clientRef.current.ping();
        } catch (error) {
            console.error('Ping error:', error);
        }
    }, []);

    // Check balance
    const checkBalance = useCallback(async (tokenAddress: string = '0xSHIB...') => {
        if (!clientRef.current?.isConnected) return;

        try {
            await clientRef.current.checkBalance(tokenAddress);
        } catch (error) {
            console.error('Balance check error:', error);
        }
    }, []);

    // Send a generic RPC request
    const sendRequest = useCallback(async (methodName: string, methodParams: string) => {
        if (!clientRef.current?.isConnected) return;

        try {
            let params: unknown[] = [];

            if (methodParams.trim()) {
                try {
                    params = JSON.parse(methodParams);
                    if (!Array.isArray(params)) params = [params];
                } catch (e) {
                    console.error('Error parsing params:', e);
                    return;
                }
            }

            const response = await clientRef.current.sendRequest(methodName, params);

            return response;
        } catch (error) {
            console.error('Request error:', error);
        }
    }, []);

    // Function to clear saved keys
    const clearKeys = useCallback(() => {
        if (typeof window !== 'undefined') {
            localStorage.removeItem('crypto_keypair');
        }
        setKeyPair(null);
        setCurrentSigner(null);
    }, []);

    return {
        // State
        status,
        keyPair,
        currentChannel,

        // Computed values
        isConnected: clientRef.current?.isConnected || false,
        hasKeys: !!keyPair,

        // Actions
        generateKeys,
        connect,
        disconnect,
        subscribeToChannel,
        sendMessage,
        sendPing,
        checkBalance,
        sendRequest,
        clearKeys,
    };
}
