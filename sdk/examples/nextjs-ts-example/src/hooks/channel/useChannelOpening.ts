import { useCallback } from 'react';
import { Address, Hex, parseSignature } from 'viem';
import { CounterApp } from '@/services/apps/counter';
import { useMessageService } from '@/hooks/ui/useMessageService';
import NitroliteStore from '@/store/NitroliteStore';
import WalletStore from '@/store/WalletStore';
import APP_CONFIG from '@/config/app';
import { MessageType } from '@erc7824/nitrolite/dist/relay';

/**
 * Custom hook to manage channel opening logic
 */
export function useChannelOpening(connect: () => Promise<boolean>, generateKeys: () => Promise<any>) {
    const { addSystemMessage } = useMessageService();

    /**
     * Handles opening a payment channel
     */
    const handleOpenChannel = useCallback(
        async (tokenAddress: string, amount: string) => {
            // Add system message about channel opening
            addSystemMessage(
                `Opening channel with token ${tokenAddress.substring(0, 6)}...${tokenAddress.substring(38)} and amount ${amount}`,
            );

            // Check if we have a valid Nitrolite client
            if (!NitroliteStore.state.client || !NitroliteStore.state.client.walletClient) {
                const errorMsg = 'Nitrolite client not initialized - please connect your wallet first';
                addSystemMessage(errorMsg);
                throw new Error(errorMsg);
            }

            // Create Counter application instance
            const app = new CounterApp();

            // Set the channel open flag first
            WalletStore.setChannelOpen(true);

            // Create a unique channel identifier (typically derived from transaction hash)
            // For demo purposes, using a timestamp-based ID
            const channelId = `channel_${Date.now()}`;

            try {
                // Initialize channel context with the app logic
                addSystemMessage('Initializing channel context...');

                // First set up the channel context with the counter app
                NitroliteStore.setChannelContext(channelId, APP_CONFIG.CHANNEL.DEFAULT_ADDRESS, app);

                // Sign the state hash using MetaMask
                addSystemMessage('Signing initial state with MetaMask...');
                if (!window.ethereum) {
                    throw new Error('MetaMask is not installed');
                }

                const address = WalletStore.state.account;
                if (!address) {
                    throw new Error('No wallet connected');
                }

                // Deposit tokens and open the channel
                addSystemMessage('Depositing tokens and opening channel...');

                try {
                    console.log('Depositing tokens...');
                    await NitroliteStore.deposit(channelId, tokenAddress as Address, amount);
                    console.log('Deposited tokens...');
                } catch (error) {
                    addSystemMessage(
                        `Error depositing tokens: ${error instanceof Error ? error.message : String(error)}`,
                    );
                    throw error;
                }

                try {
                    console.log('Opening channel...');

                    // Create initial app state
                    const appState = { type: 'system' as MessageType, text: '0', sequence: '0' };

                    // Get state hash for signing
                    const stateHash = NitroliteStore.state.channelContext[channelId].getStateHash(
                        appState,
                        tokenAddress as Address,
                        [BigInt(amount), BigInt(0)] as [bigint, bigint],
                    );

                    // const account = privateKeyToAccount("0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")

                    // // 4. Sign the hash
                    // const signature = await account.signMessage({
                    //     message: { raw: stateHash },
                    // })

                    const signature = await window.ethereum.request({
                        method: 'personal_sign',
                        // params: [`0x${Buffer.from(stateHash).toString('hex')}`, address],
                        params: [stateHash, address],

                    });

                    const parsedSig = parseSignature(signature as Hex);

                    await NitroliteStore.openChannel(
                        channelId,
                        appState,
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
                    console.log('Opened channel...');
                } catch (error) {
                    addSystemMessage(
                        `Error opening channel: ${error instanceof Error ? error.message : String(error)}`,
                    );
                    throw error;
                }

                addSystemMessage('Channel opened successfully! Now generating keys for WebSocket connection...');

                // Automatically generate keys after channel is opened
                try {
                    addSystemMessage('Automatically generating keys for the channel...');
                    const newKeyPair = await generateKeys();

                    if (newKeyPair) {
                        addSystemMessage('Keys generated successfully. Attempting to connect to WebSocket server...');

                        // Automatically attempt to connect to WebSocket after generating keys
                        try {
                            const connected = await connect();

                            if (connected) {
                                addSystemMessage('Successfully connected to WebSocket server.');
                            } else {
                                addSystemMessage(
                                    'Failed to connect to WebSocket server. Please try connecting manually.',
                                );
                            }
                        } catch (connectError) {
                            addSystemMessage(
                                `Error connecting to WebSocket: ${connectError instanceof Error ? connectError.message : String(connectError)}`,
                            );
                        }
                    } else {
                        addSystemMessage('Failed to generate keys automatically. Please try generating keys manually.');
                    }
                } catch (error) {
                    addSystemMessage(
                        `Error generating keys: ${error instanceof Error ? error.message : String(error)}`,
                    );
                }
            } catch (error) {
                addSystemMessage(`Error opening channel: ${error instanceof Error ? error.message : String(error)}`);
                WalletStore.setChannelOpen(false);
            }
        },
        [addSystemMessage, generateKeys],
    );

    return {
        handleOpenChannel,
    };
}
