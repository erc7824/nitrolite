'use client';

import { useEffect } from 'react';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useMessageService } from '@/hooks/useMessageService';
import WalletStore from '@/store/WalletStore';
import { fetchAssets } from '@/store/AssetsStore';
import { Address, encodeAbiParameters, keccak256, parseSignature } from 'viem';
// import { generateKeyPair } from "@/websocket/crypto";

// Components
import { Header } from '@/components/Header';
import { ChannelStatus } from '@/components/ChannelStatus';
import { AuthKeyDisplay } from '@/components/AuthKeyDisplay';
import { MessageList } from '@/components/MessageList';
import { RequestForm } from '@/components/RequestForm';
import { InfoSection } from '@/components/InfoSection';
import MetaMaskConnect from '@/components/MetaMaskConnect';
import NitroliteStore from '@/store/NitroliteStore';
import { CounterApp } from './apps/counter';
import { useNitroliteClient } from '@/hooks/useNitroliteClient';
import { privateKeyToAccount } from 'viem/accounts';

export default function Home() {
    const { status, addSystemMessage } = useMessageService();

    const {
        keyPair,
        currentChannel,
        isConnected,
        generateKeys,
        connect,
        disconnect,
        subscribeToChannel,
        sendMessage,
        sendPing,
        checkBalance,
        sendRequest,
    } = useWebSocket('ws://localhost:8000/ws');

    // Load assets and add initial message when component mounts
    useEffect(() => {
        fetchAssets();

        // Add an initial system message
        addSystemMessage('Application initialized - Welcome to Nitrolite!');
    }, [addSystemMessage]);

    useNitroliteClient();

    // Function to handle channel opening
    const handleOpenChannel = async (tokenAddress: string, amount: string) => {
        // Add system message about channel opening
        addSystemMessage(
            `Opening channel with token ${tokenAddress.substring(0, 6)}...${tokenAddress.substring(38)} and amount ${amount}`,
        );

        const app = new CounterApp();

        NitroliteStore.setChannelContext(currentChannel, '0x70997970C51812dc3A010C7d01b50e0d17dc79C8', app);

        const appState = { type: 'system', text: '' };
        //TODO:
        const stateHash = NitroliteStore.state.channelContext[currentChannel].getStateHash(
            appState,
            tokenAddress as Address,
            [BigInt(amount), BigInt(0)] as [bigint, bigint],
        );

        const account = privateKeyToAccount('0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80');
        const signature = await account.signMessage({
            message: { raw: stateHash },
        });
        const parsedSig = parseSignature(signature);

        await NitroliteStore.deposit(currentChannel, tokenAddress as Address, amount);
        await NitroliteStore.openChannel(
            currentChannel,
            { type: 'system', text: '' },
            tokenAddress as Address,
            [BigInt(amount), BigInt(0)] as [bigint, bigint],
            [
                {
                    r: parsedSig.r,
                    s: parsedSig.s,
                    v: +parsedSig.v.toString(),
                },
            ],
        );

        // Generate keys and connect to websocket in a sequential flow
        try {
            // Step 1: Generate keys if not present
            let currentKeyPair = keyPair;

            if (!currentKeyPair) {
                addSystemMessage('Generating new key pair...');
                currentKeyPair = await generateKeys();
                if (!currentKeyPair) {
                    const errorMsg = 'Failed to generate keys';

                    addSystemMessage(errorMsg);
                    throw new Error(errorMsg);
                }
                addSystemMessage('Key pair generated successfully');
            }

            // Step 2: Connect to the broker websocket only after we have keys
            if (status === 'disconnected' && currentKeyPair) {
                try {
                    addSystemMessage('Connecting to WebSocket server...');
                    await connect();
                    addSystemMessage('WebSocket connection established');
                } catch (error) {
                    addSystemMessage(
                        'WebSocket connection error: Make sure the WebSocket server is running at ws://localhost:8000/ws',
                        error,
                    );
                }
            }
        } catch (error) {
            addSystemMessage(
                `Error in channel opening sequence: ${error instanceof Error ? error.message : String(error)}`,
            );
        }
    };

    // Handle wallet disconnection
    const handleDisconnect = async () => {
        // First disconnect from WebSocket if connected
        if (status === 'connected') {
            disconnect(); // This is the WebSocket disconnect
        }

        // Then disconnect from MetaMask
        const { disconnectWallet } = await import('@/hooks/useMetaMask');

        await disconnectWallet();
    };

    const isChannelOpen = WalletStore.state.channelOpen;

    return (
        <div className="min-h-screen bg-gradient-to-br from-white to-gray-100 text-gray-800 p-6">
            <div className="max-w-6xl mx-auto">
                <Header onDisconnect={handleDisconnect} wsConnected={isConnected} />

                {isChannelOpen ? (
                    <>
                        <div className="flex gap-3 mb-2 flex-col md:flex-row">
                            <ChannelStatus status={status} />
                            <AuthKeyDisplay keyPair={keyPair} status={status} />
                        </div>

                        <MessageList />

                        <RequestForm
                            isConnected={isConnected}
                            currentChannel={currentChannel}
                            onSendRequest={sendRequest}
                            onSendMessage={sendMessage}
                            onSubscribeToChannel={subscribeToChannel}
                            onSendPing={sendPing}
                            onCheckBalance={checkBalance}
                        />

                        <InfoSection />
                    </>
                ) : (
                    <MetaMaskConnect onChannelOpen={handleOpenChannel} />
                )}
            </div>
        </div>
    );
}
