import { useEffect } from 'react';
import { useSnapshot } from 'valtio';
import { createPublicClient, createWalletClient, custom, http } from 'viem';
import { NitroliteClient } from '@erc7824/nitrolite';
import WalletStore from '../store/WalletStore';
import NitroliteStore from '../store/NitroliteStore';
import { chains } from '@/config/chains';

// You might want to move these to a config file
const CONTRACTS = {
    custody: '',
    adjudicators: {
        base: '0x5fbdb2315678afecb367f032d93f642f64180aa3',
    },
};

export function useNitroliteClient() {
    const walletState = useSnapshot(WalletStore.state);

    useEffect(() => {
        if (!walletState.connected || !walletState.account || !walletState.chainId) {
            return;
        }

        console.log('Initializing Nitrolite client...');

        const chain = chains.find((chain) => chain.id === walletState.chainId);

        try {
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

            console.log("wallet client", walletClient)

            // Create Nitrolite client
            const client = new NitroliteClient({
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
    }, [walletState.connected, walletState.account, walletState.chainId]);

    return;
}
