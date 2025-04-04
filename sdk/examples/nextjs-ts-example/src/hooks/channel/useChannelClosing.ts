import { useCallback } from 'react';
import { Address, createWalletClient, Hex, http, parseSignature } from 'viem';
import { useMessageService } from '@/hooks/ui/useMessageService';
import NitroliteStore from '@/store/NitroliteStore';
import WalletStore from '@/store/WalletStore';
import { MessageType } from '@erc7824/nitrolite/dist/relay';
import APP_CONFIG from '@/config/app';
import { privateKeyToAccount } from 'viem/accounts';
import { localhost } from 'viem/chains';

/**
 * Custom hook to manage channel opening logic
 */
export function useChannelClosing() {
    const { addSystemMessage } = useMessageService();

    const handleCloseChannel = useCallback(
        async (channelId: string, tokenAddress: string, amounts: [string, string]) => {
            addSystemMessage(
                `Closing channel ${channelId} with token ${tokenAddress.substring(0, 6)}...${tokenAddress.substring(38)} and amounts: [${amounts[0]}, ${amounts[1]}]`,
            );

            if (!NitroliteStore.state.client || !NitroliteStore.state.client.walletClient) {
                const errorMsg = 'Nitrolite client not initialized - please connect your wallet first';

                addSystemMessage(errorMsg);
                throw new Error(errorMsg);
            }

            try {
                // Sign the state hash using MetaMask
                addSystemMessage('Signing final state with MetaMask...');
                if (!window.ethereum) {
                    throw new Error('MetaMask is not installed');
                }

                const address = WalletStore.state.account;

                if (!address) {
                    throw new Error('No wallet connected');
                }

                // Create initial app state
                const appState = { type: 'system' as MessageType, text: '0', sequence: '0' };
                const channel = NitroliteStore.getChannelContext(channelId);

                // Get state hash for signing
                const stateHash = channel.getStateHash(
                    appState,
                    tokenAddress as Address,
                    [BigInt(amounts[0]), BigInt(amounts[1])] as [bigint, bigint],
                );

                const signature = await window.ethereum.request({
                    method: 'personal_sign',
                    params: [stateHash, address],
                });

                addSystemMessage('Closing channel...');

                // TODO:
                const pseudoGuestAccount = privateKeyToAccount(APP_CONFIG.CHANNEL.GUEST_KEY as Hex);
                const pseudoGuestClient = createWalletClient({
                    chain: localhost,
                    transport: http(),
                    account: pseudoGuestAccount,
                });
                const guestSignature = await pseudoGuestClient.signMessage({
                    account: pseudoGuestAccount,
                    message: { raw: stateHash },
                });

                const pss = [parseSignature(signature as Hex), parseSignature(guestSignature as Hex)];

                try {
                    await NitroliteStore.closeChannel(
                        channelId,
                        appState,
                        tokenAddress as Address,
                        [BigInt(amounts[0]), BigInt(amounts[1])],
                        [
                            {
                                r: pss[0].r,
                                s: pss[0].s,
                                v: +pss[0].v.toString(),
                            },
                            {
                                r: pss[1].r,
                                s: pss[1].s,
                                v: +pss[1].v.toString(),
                            },
                        ],
                    );
                } catch (error) {
                    addSystemMessage(
                        `Error closing channel: ${error instanceof Error ? error.message : String(error)}`,
                    );
                    throw error;
                }

                addSystemMessage('Channel closed successfully! Withdrawn funds should be on your wallet...');
            } catch (error) {
                addSystemMessage(`Error closing channel: ${error instanceof Error ? error.message : String(error)}`);
                WalletStore.setChannelOpen(false);
            }
        },
        [addSystemMessage],
    );

    return {
        handleCloseChannel,
    };
}
