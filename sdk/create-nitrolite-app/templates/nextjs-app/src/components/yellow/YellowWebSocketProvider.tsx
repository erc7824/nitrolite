import { createContext, useContext, useEffect, useRef } from 'react';
import type { Children } from 'react';
import { usePrivy, useWallets } from '@privy-io/react-auth';
import { useYellowWebSocket } from '../../hooks/useYellowWebSocket';
import { useLoginState } from '../../hooks/useLoginState';
import {
    handleAssetsUpdate,
    handleBalanceUpdate,
    handleLedgerTransactionsUpdate,
    handleGetUserTag,
} from '../../hooks/useAssetsUpdates';
import { parseAnyRPCResponse, RPCMethod, type RPCResponse } from '@erc7824/nitrolite';
import { handleTransferNotification } from '../../hooks/useNotifications';

interface YellowWebSocketContextType {
    isConnected: boolean;
    isConnecting: boolean;
    isAuthenticated: boolean;
    error: string | null;
    sessionAddress: string | null;
    sessionSigner: any;
    send: (data: any) => void;
    sendWithResponse: (data: any, options?: { timeout?: number }) => Promise<{ requestInfo: any; response: any }>;
    connect: (walletAddress: string) => Promise<void>;
    disconnect: () => void;
    ping: () => Promise<void>;
    client: any;
    status: string;
}

const YellowWebSocketContext = createContext<YellowWebSocketContextType | null>(null);

interface YellowWebSocketProviderProps {
    children: typeof Children;
}

export const YellowWebSocketProvider = ({ children }: YellowWebSocketProviderProps) => {
    const { isLoggedIn, walletAddress } = useLoginState();
    const { user, authenticated, ready } = usePrivy();
    const { wallets } = useWallets();
    const connectionAttempted = useRef(false);
    const lastWalletAddress = useRef<string | null>(null);
    const userRejectedSigning = useRef(false);

    const {
        client,
        status,
        isConnected,
        isConnecting,
        isAuthenticated,
        error,
        sessionAddress,
        send,
        sendWithResponse,
        connect,
        disconnect,
        ping,
        privyWalletReady,
    } = useYellowWebSocket({
        onMessage: handleYellowMessage,
        onConnect: () => {
            connectionAttempted.current = true;
        },
        onDisconnect: () => {
            connectionAttempted.current = false;
        },
        onError: (error) => {
            // Check if user rejected signing
            if (
                error.name === 'UserRejectedError' ||
                error.message.toLowerCase().includes('user rejected') ||
                error.message.toLowerCase().includes('user denied') ||
                error.message.toLowerCase().includes('user cancelled')
            ) {
                userRejectedSigning.current = true;
            }
        },
        onAuthSuccess: () => {
            // Authentication successful
        },
        onAuthFailed: (errorMsg) => {
            connectionAttempted.current = false;

            // Check if user rejected signing
            if (
                errorMsg.toLowerCase().includes('user rejected') ||
                errorMsg.toLowerCase().includes('user denied') ||
                errorMsg.toLowerCase().includes('user cancelled')
            ) {
                userRejectedSigning.current = true;
            } else {
                // Reset after delay for non-rejection errors
                setTimeout(() => {
                    connectionAttempted.current = false;
                }, 5000);
            }
        },
        onChallengeReceived: (challengeData) => {
            // Challenge received - button should now be visible
            if (import.meta.env.DEV) {
                console.log('Challenge received:', challengeData);
            }
        },
        onVerifyFailed: (error) => {
            // Verification failed - keep connection alive
            if (import.meta.env.DEV) {
                console.log('Verification failed:', error);
            }
        },
    });

    // Auto-connect when user logs in and wallet is ready
    useEffect(() => {
        const embeddedPrivyWallet = wallets.find((wallet) => wallet.walletClientType === 'privy');

        console.log('YellowWebSocket Auto-connect check:', {
            isLoggedIn,
            walletAddress,
            authenticated,
            ready,
            userWalletAddress: user?.wallet?.address,
            embeddedPrivyWallet: !!embeddedPrivyWallet,
            privyWalletReady,
            isConnected,
            isConnecting,
            connectionAttempted: connectionAttempted.current,
            userRejectedSigning: userRejectedSigning.current,
            lastWalletAddress: lastWalletAddress.current,
        });

        const shouldConnect =
            isLoggedIn &&
            walletAddress &&
            authenticated &&
            ready &&
            user?.wallet?.address &&
            privyWalletReady && // Use the hook's wallet ready state
            !isConnected &&
            !isConnecting &&
            !connectionAttempted.current &&
            !userRejectedSigning.current &&
            lastWalletAddress.current !== walletAddress;

        console.log('shouldConnect:', shouldConnect);

        if (shouldConnect) {
            lastWalletAddress.current = walletAddress;
            connectionAttempted.current = true;
            console.log('🔌 Attempting to connect to Yellow WebSocket with address:', walletAddress);
            connect(walletAddress).catch((error) => {
                console.error('❌ Failed to connect to Yellow WebSocket:', error);
                connectionAttempted.current = false;
            });
        }
    }, [
        isLoggedIn,
        walletAddress,
        authenticated,
        ready,
        user?.wallet?.address,
        wallets,
        privyWalletReady,
        isConnected,
        isConnecting,
        connect,
    ]);

    // Reset connection tracking when wallet changes or user logs out
    useEffect(() => {
        if (!isLoggedIn || !walletAddress) {
            connectionAttempted.current = false;
            lastWalletAddress.current = null;
            userRejectedSigning.current = false;
        }
    }, [isLoggedIn, walletAddress]);

    function handleYellowMessage(data: any) {
        let res: RPCResponse;
        try {
            const message = typeof data === 'string' ? data : JSON.stringify(data);
            res = parseAnyRPCResponse(message);
        } catch (error) {
            throw new Error(`Error processing WebSocket message: ${error}`);
        }

        // Handle different message types
        switch (res.method) {
            case RPCMethod.AuthChallenge:
                // Forward auth_challenge to the client for proper handling
                if (client) {
                    if (import.meta.env.DEV) {
                        console.log('YellowWebSocketProvider forwarding auth_challenge to client');
                    }
                    // Manually trigger the client's challenge handling
                    if (client.handleChallengeMessage) {
                        client.handleChallengeMessage(data);
                    }
                }
                break;
            case RPCMethod.BalanceUpdate:
                console.log('Processing balance update with params:', res.params);
                handleBalanceUpdate(res.params);
                break;
            case RPCMethod.GetAssets:
            case RPCMethod.Assets:
                console.log('Processing assets update with params:', res.params);
                handleAssetsUpdate(res.params);
                break;
            case RPCMethod.GetLedgerBalances:
                console.log('Processing ledger balances update with params:', res.params);
                handleBalanceUpdate(res.params);
                break;
            case RPCMethod.GetLedgerTransactions:
                console.log('Processing ledger transactions update with params:', res.params);
                handleLedgerTransactionsUpdate(res.params);
                break;
            case RPCMethod.GetUserTag:
                console.log('Processing user tag update with params:', res.params);
                handleGetUserTag(res.params);
                break;
            case RPCMethod.TransferNotification:
                console.log('Processing transfer notification with params:', res.params);
                handleTransferNotification(res.params);
                break;
            default:
                if (import.meta.env.DEV) {
                    console.log('Unknown Yellow message:', data);
                }
        }
    }

    const contextValue: YellowWebSocketContextType = {
        isConnected,
        isConnecting,
        isAuthenticated,
        error,
        sessionAddress,
        sessionSigner: client?.sessionSigner || null,
        send,
        sendWithResponse,
        connect,
        disconnect,
        ping,
        client,
        status,
    };

    return <YellowWebSocketContext.Provider value={contextValue}>{children}</YellowWebSocketContext.Provider>;
};

export const useYellowWebSocketContext = (): YellowWebSocketContextType => {
    const context = useContext(YellowWebSocketContext);
    if (!context) {
        throw new Error('useYellowWebSocketContext must be used within YellowWebSocketProvider');
    }
    return context;
};
