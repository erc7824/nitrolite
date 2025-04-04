'use client';

import { useEffect } from 'react';

// Import hooks from organized structure
import { useChannelOpening, useNitroliteClient } from '@/hooks/channel';
import { useMessageService } from '@/hooks/ui';
import { useWalletConnection } from '@/hooks/wallet';
import { useWebSocket } from '@/hooks/websocket';

// Components
import { ConnectedView } from '@/components/ConnectedView';
import { Header } from '@/components/Header';
import MetaMaskConnect from '@/components/MetaMaskConnect';

// Stores & Config
import WalletStore from '@/store/WalletStore';
import { fetchAssets } from '@/store/AssetsStore';
import APP_CONFIG from '@/config/app';

export default function Home() {
    const { status, addSystemMessage } = useMessageService();

    const { keyPair, wsChannel, currentNitroliteChannel, isConnected, connect, disconnect, generateKeys } =
        useWebSocket(APP_CONFIG.WEBSOCKET.URL);

    const { handleOpenChannel } = useChannelOpening(currentNitroliteChannel, connect, status, generateKeys);

    const { handleDisconnect } = useWalletConnection(status, disconnect);

    // Load assets and add initial message when component mounts
    useEffect(() => {
        fetchAssets();
        addSystemMessage('Application initialized - Welcome to Nitrolite!');
    }, [addSystemMessage]);

    useNitroliteClient();

    const isChannelOpen = WalletStore.state.channelOpen;

    return (
        <div className="min-h-screen bg-gradient-to-br from-white to-gray-100 text-gray-800 p-6">
            <div className="max-w-6xl mx-auto">
                <Header onDisconnect={handleDisconnect} wsConnected={isConnected} />

                {isChannelOpen ? (
                    <ConnectedView
                        status={status}
                        keyPair={keyPair}
                        isConnected={isConnected}
                        wsChannel={wsChannel}
                        onGenerateKeys={generateKeys}
                        onConnect={connect}
                        onDisconnect={disconnect}
                    />
                ) : (
                    <MetaMaskConnect onChannelOpen={handleOpenChannel} />
                )}
            </div>
        </div>
    );
}
