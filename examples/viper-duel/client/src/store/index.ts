import { Store } from './storeUtils';
import type { Address, WalletClient } from 'viem';

// Nitrolite Store
interface NitroliteStoreState {
    client: any; // This would be NitroliteClient from the package
    isInitialized: boolean;
}

export const NitroliteStore = {
    state: new Store<NitroliteStoreState>({
        client: null,
        isInitialized: false,
    }),

    setClient(client: any): void {
        this.state.set('client', client);
        this.state.set('isInitialized', true);
    },
};

// Wallet Store
interface WalletStoreState {
    walletAddress: string | null;
    chainId: number | null;
    isConnected: boolean;
    channelOpen: boolean;
    walletClient: WalletClient | null;
    channelToken: Address | null;
    channelAmount: string | null;
    error: string | null;
}

export const WalletStore = {
    state: new Store<WalletStoreState>({
        walletAddress: null,
        chainId: null,
        isConnected: false,
        channelOpen: false,
        walletClient: null,
        channelToken: null,
        channelAmount: null,
        error: null,
    }),

    connect(address: Address): void {
        this.state.setState({
            walletAddress: address,
            isConnected: true,
            error: null,
        });
    },

    setError(error: string | null): void {
        this.state.set('error', error);
    },

    setWalletAddress(address: string | null): void {
        this.state.set('walletAddress', address);
        this.state.set('isConnected', !!address);
    },

    setChainId(chainId: number | null): void {
        this.state.set('chainId', chainId);
    },

    setChannelOpen(isOpen: boolean): void {
        this.state.set('channelOpen', isOpen);
    },

    openChannel(token: Address, amount: string): void {
        this.state.setState({
            channelOpen: true,
            channelToken: token,
            channelAmount: amount,
        });
    },

    /**
     * Set wallet client
     * @param walletClient Wallet client instance
     */
    setWalletClient(walletClient: WalletClient | null): void {
        this.state.set('walletClient', walletClient);

        if (walletClient?.account?.address) {
            this.state.set('walletAddress', walletClient.account.address);
        }
    },

    /**
     * Get Wallet Client
     */
    getWalletClient(): WalletClient | null {
        return this.state.getState().walletClient;
    },

    closeChannel(): void {
        this.state.setState({
            channelOpen: false,
            channelToken: null,
            channelAmount: null,
        });
    },
};

// Settings Store
interface SettingsStoreState {
    activeChain: number;
}

export const SettingsStore = {
    state: new Store<SettingsStoreState>({
        activeChain: 137, // Default to Polygon
    }),

    setActiveChain(chainId: number): void {
        this.state.set('activeChain', chainId);
    },
};
