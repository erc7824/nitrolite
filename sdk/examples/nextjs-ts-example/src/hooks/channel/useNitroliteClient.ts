import { useEffect } from 'react';
import { useSnapshot } from 'valtio';
import { createPublicClient, createWalletClient, custom, http } from 'viem';
import { NitroliteClient } from '@erc7824/nitrolite';
import WalletStore from '@/store/WalletStore';
import NitroliteStore from '@/store/NitroliteStore';
import { chains } from '@/config/chains';
import CONTRACTS from '@/config/contracts';

export function useNitroliteClient() {
    const walletState = useSnapshot(WalletStore.state);

    useEffect(() => {
        if (!walletState.connected || !walletState.account || !walletState.chainId) {
            return;
        }

        const initializeClient = async () => {
            try {
                const chain = chains.find((chain) => chain.id === walletState.chainId);

                if (!chain) {
                    throw new Error(`Unsupported chain ID: ${walletState.chainId}`);
                }

                // Create public client
                const publicClient = createPublicClient({
                    transport: http(),
                    chain,
                });

                // Create wallet client using window.ethereum
                const walletClient = createWalletClient({
                    transport: custom(window.ethereum),
                    chain,
                    account: walletState.account,
                });

                // Create Nitrolite client
                const client = new NitroliteClient({
                    // @ts-ignore
                    publicClient,
                    walletClient,
                    account: walletClient.account,
                    chainId: walletState.chainId,
                    addresses: CONTRACTS,
                });

                // Save client to store
                NitroliteStore.setClient(client);
            } catch (error) {
                console.error('Failed to initialize Nitrolite client:', error);
                WalletStore.setError('Failed to initialize Nitrolite client');
            }
        };

        initializeClient();
    }, [walletState.connected, walletState.account, walletState.chainId]);

    return;
}