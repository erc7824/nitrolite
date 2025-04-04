import { useCallback } from 'react';
import { useMessageService } from '@/hooks/ui/useMessageService';
import { disconnectWallet } from './useMetaMask';

/**
 * Custom hook to manage wallet connection and disconnection
 */
export function useWalletConnection(status: string, disconnect: () => void) {
    const { addSystemMessage } = useMessageService();

    /**
     * Handles disconnecting from wallet and websocket
     */
    const handleDisconnect = useCallback(async () => {
        // First disconnect from WebSocket if connected
        if (status === 'connected') {
            disconnect(); // This is the WebSocket disconnect
        }

        // Then disconnect from MetaMask
        await disconnectWallet();
        addSystemMessage('Disconnected from wallet');
    }, [status, disconnect, addSystemMessage]);

    return {
        handleDisconnect,
    };
}