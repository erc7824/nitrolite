import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { usePrivy, useWallets } from '@privy-io/react-auth';
import { createWalletClient, custom, type Address } from 'viem';
import { mainnet } from 'viem/chains';
import { createYellowWebSocketClient, type YellowWebSocketClient } from '../services/yellow/client';
import type { WSStatus, YellowConnectionCallbacks, YellowConfig } from '../services/yellow/types';

interface UseYellowWebSocketOptions extends YellowConnectionCallbacks {
    config?: Partial<YellowConfig>;
}

export const useYellowWebSocket = (options: UseYellowWebSocketOptions = {}) => {
    const { authenticated, ready } = usePrivy();
    const { wallets } = useWallets();
    const [status, setStatus] = useState<WSStatus>('disconnected');
    const [error, setError] = useState<string | null>(null);
    const clientRef = useRef<YellowWebSocketClient | null>(null);
    const [isAutoApprovingChallenge, setIsAutoApprovingChallenge] = useState(false);

    // Check if Privy embedded wallet is ready
    const privyWalletReady = useMemo(() => {
        const embeddedPrivyWallet = wallets.find((wallet) => wallet.walletClientType === 'privy');
        return ready && authenticated && !!embeddedPrivyWallet;
    }, [ready, authenticated, wallets]);

    // Auto-approve challenges when they are received (but only once per challenge)
    const challengeApprovedRef = useRef<boolean>(false);

    useEffect(() => {
        const autoApproveChallenge = async () => {
            if (
                clientRef.current &&
                clientRef.current.hasPendingChallenge &&
                status === 'pending_auth' &&
                !isAutoApprovingChallenge &&
                !challengeApprovedRef.current
            ) {
                console.log('Auto-approving Yellow WebSocket challenge...');
                challengeApprovedRef.current = true;
                setIsAutoApprovingChallenge(true);

                try {
                    await clientRef.current.approveChallenge();
                    console.log('Challenge approved automatically');
                } catch (error) {
                    console.error('Auto-challenge approval failed:', error);
                    setError(error instanceof Error ? error.message : 'Challenge approval failed');
                    challengeApprovedRef.current = false; // Reset on error to allow retry
                } finally {
                    setIsAutoApprovingChallenge(false);
                }
            }
        };

        autoApproveChallenge();
    }, [status, isAutoApprovingChallenge]);

    // Reset challenge approval flag when disconnected
    useEffect(() => {
        if (status === 'disconnected' || status === 'failed') {
            challengeApprovedRef.current = false;
        }
    }, [status]);

    // Initialize client on mount
    useEffect(() => {
        const client = createYellowWebSocketClient(options.config, {
            onConnect: () => {
                setError(null);
                options.onConnect?.();
            },
            onDisconnect: options.onDisconnect,
            onMessage: options.onMessage,
            onError: (err) => {
                setError(err.message);
                options.onError?.(err);
            },
            onAuthSuccess: options.onAuthSuccess,
            onAuthFailed: (errorMsg) => {
                setError(errorMsg);
                options.onAuthFailed?.(errorMsg);
            },
            onChallengeReceived: (challenge) => {
                console.log('Challenge received, will auto-approve...');
                options.onChallengeReceived?.(challenge);
            },
            onVerifyFailed: options.onVerifyFailed,
        });

        clientRef.current = client;

        // Listen to status changes
        const unsubscribe = client.onStatusChange(setStatus);

        return () => {
            unsubscribe();
            client.destroy();
            clientRef.current = null;
        };
    }, []);

    const connect = useCallback(
        async (walletAddress: string) => {
            if (!authenticated || !clientRef.current) {
                throw new Error('User not authenticated or client not available');
            }

            if (clientRef.current.isConnected) {
                return;
            }

            // Wait for Privy wallet to be ready
            if (!privyWalletReady) {
                throw new Error('Privy wallet not ready - please wait for wallet to initialize');
            }

            // Find the embedded Privy wallet
            const embeddedPrivyWallet = wallets.find((wallet) => wallet.walletClientType === 'privy');
            if (!embeddedPrivyWallet) {
                throw new Error('Embedded Privy wallet not found');
            }

            console.log('Creating viem wallet client from embedded Privy wallet...');

            // Switch to mainnet chain
            await embeddedPrivyWallet.switchChain(mainnet.id);

            // Get the EIP1193 provider from the embedded wallet
            const eip1193provider = await embeddedPrivyWallet.getEthereumProvider();

            // Create a proper viem wallet client using the embedded wallet
            const walletClient = createWalletClient({
                account: embeddedPrivyWallet.address as Address,
                chain: mainnet,
                transport: custom(eip1193provider),
            });

            console.log('Viem wallet client created successfully:', walletClient);

            // Create signing function that's compatible with the client
            const signTypedData = async (args: { domain: any; types: any; primaryType: string; message: any }) => {
                return await walletClient.signTypedData({
                    account: embeddedPrivyWallet.address as any,
                    domain: args.domain,
                    types: args.types,
                    primaryType: args.primaryType,
                    message: args.message,
                });
            };

            await clientRef.current.connect(walletAddress, signTypedData, walletClient);
        },
        [authenticated, wallets, privyWalletReady],
    );

    const disconnect = useCallback(() => {
        clientRef.current?.disconnect();
    }, []);

    const send = useCallback((data: any) => {
        if (!clientRef.current?.isConnected) {
            throw new Error('Not connected to Yellow WebSocket');
        }
        clientRef.current.send(data);
    }, []);

    const sendWithResponse = useCallback(async (data: any, options?: { timeout?: number }) => {
        if (!clientRef.current?.isConnected) {
            throw new Error('Not connected to Yellow WebSocket');
        }
        return await clientRef.current.sendWithResponse(data, options);
    }, []);

    const ping = useCallback(async () => {
        if (!clientRef.current) {
            throw new Error('Client not initialized');
        }
        await clientRef.current.ping();
    }, []);

    return {
        client: clientRef.current,
        status,
        connect,
        disconnect,
        send,
        sendWithResponse,
        ping,
        isConnected: status === 'connected',
        isConnecting: status === 'connecting',
        isAuthenticated: status === 'connected',
        error,
        sessionAddress: clientRef.current?.currentSessionAddress || null,
        retryCount: 0, // Could be extracted from client if needed
        privyWalletReady,
    };
};
