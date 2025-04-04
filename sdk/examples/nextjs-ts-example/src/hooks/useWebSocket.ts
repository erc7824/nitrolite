import { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import { createWebSocketClient, createEthersSigner, WalletSigner, WebSocketClient } from '@/websocket';
import { Channel } from '@/types';
import { useMessageService } from './useMessageService';
import { useKeyPair, usePingPongBenchmark, useMessageHandler } from './websocket';

/**
 * Custom hook to manage WebSocket connection and operations
 */
export function useWebSocket(url: string) {
    const [status, setStatus] = useState<string>('disconnected');
    const [currentChannel, setCurrentChannel] = useState<Channel | null>(null);
    const [startBenchmarkFlag, setStartBenchmarkFlag] = useState<boolean>(false);
    const [currentSigner, setCurrentSigner] = useState<WalletSigner | null>(null);

    // Use our key pair management hook
    const { keyPair, clearKeys, generateKeys, hasKeys } = useKeyPair();

    // Use our message service
    const messageService = useMessageService();
    const { setStatus: setMessageStatus, addSystemMessage, addErrorMessage, addPingMessage } = messageService;

    // Update both statuses
    const updateStatus = useCallback(
        (newStatus: string) => {
            setStatus(newStatus);
            setMessageStatus(newStatus);
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
            maxReconnectAttempts: 3,
            requestTimeout: 10000,
            pingChannel: 'public',
        });
    }, [url, currentSigner]);

    const clientRef = useRef<WebSocketClient | null>(null);

    // Update the client reference when the client changes
    useEffect(() => {
        clientRef.current = client;
    }, [client]);

    // Use our ping-pong benchmark hook
    const pingPongBenchmark = usePingPongBenchmark(clientRef, messageService);
    const { runPingPongBenchmark } = pingPongBenchmark;

    // Use our message handler hook
    const { handleMessage } = useMessageHandler(clientRef, pingPongBenchmark, messageService);

    // Run ping-pong benchmark when flag is set
    useEffect(() => {
        if (startBenchmarkFlag && clientRef.current?.isConnected) {
            addSystemMessage('Starting 1000 ping-pong benchmark...');
            runPingPongBenchmark(1000).catch((err) => {
                console.error('Benchmark error:', err);
                addErrorMessage(`Benchmark error: ${err instanceof Error ? err.message : String(err)}`);
            });
            setStartBenchmarkFlag(false);
        }
    }, [startBenchmarkFlag, runPingPongBenchmark, addSystemMessage, addErrorMessage]);

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
        client.onMessage(handleMessage);

        // Add initial system message
        addSystemMessage('WebSocket listeners initialized successfully');

        return () => {
            addSystemMessage('Cleaning up WebSocket connection');
            client.close();
        };
    }, [updateStatus, handleMessage, addSystemMessage, addErrorMessage]);

    // Generate a new key pair with error handling
    const generateKeysWithErrorHandling = useCallback(async () => {
        try {
            const newKeyPair = await generateKeys();

            if (newKeyPair) {
                // Create a new signer with the generated private key
                const newSigner = createEthersSigner(newKeyPair.privateKey);
                setCurrentSigner(newSigner);
                return newKeyPair;
            }
            return null;
        } catch (error) {
            const errorMsg = `Error generating keys: ${error instanceof Error ? error.message : String(error)}`;
            addErrorMessage(errorMsg);
            return null;
        }
    }, [generateKeys, addErrorMessage]);

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
            setStartBenchmarkFlag(true);
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
            addPingMessage(`>user: PING`, 'user');
        } catch (error) {
            console.error('Ping error:', error);
            addErrorMessage(`Ping error: ${error instanceof Error ? error.message : String(error)}`);
        }
    }, [addPingMessage, addErrorMessage]);

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

            return await clientRef.current.sendRequest(methodName, params);
        } catch (error) {
            console.error('Request error:', error);
        }
    }, []);

    // Get latest stats
    const stats = pingPongBenchmark.getTotalStats();

    return {
        // State
        status,
        keyPair,
        currentChannel,

        // Computed values
        isConnected: clientRef.current?.isConnected || false,
        hasKeys,
        benchmarkInProgress: pingPongBenchmark.benchmarkInProgress.current,

        // Message counts
        userPingCount: stats.user.pings,
        userPongCount: stats.user.pongs,
        userTotal: stats.user.total,
        guestPingCount: stats.guest.pings,
        guestPongCount: stats.guest.pongs,
        guestTotal: stats.guest.total,
        totalMessages: stats.total,

        // Actions
        generateKeys: generateKeysWithErrorHandling,
        connect,
        disconnect,
        subscribeToChannel,
        sendMessage,
        sendPing,
        checkBalance,
        sendRequest,
        clearKeys,
        runPingPongBenchmark,
    };
}
