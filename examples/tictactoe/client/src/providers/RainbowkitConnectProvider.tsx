'use client';

import { lightTheme, RainbowKitProvider, connectorsForWallets } from '@rainbow-me/rainbowkit';
import { metaMaskWallet } from '@rainbow-me/rainbowkit/wallets';
import { WagmiProvider } from 'wagmi';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';
import { createConfig, http } from 'wagmi';
import { polygon } from 'viem/chains';

const connectors = connectorsForWallets(
    [
        {
            groupName: 'Recommended',
            wallets: [metaMaskWallet],
        },
    ],
    {
        appName: 'TicTacToe',
        projectId: import.meta.env?.NEXT_PUBLIC_PROJECT_ID ?? '014c1e90b23a969ce37a8444f1977fad',
    }
);

export const rainbowkitConfig = createConfig({
    connectors,
    chains: [polygon],
    transports: {
        [polygon.id]: http(),
    },
    ssr: true,
});

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            staleTime: 1000 * 60 * 5,
        },
    },
});

interface IRainbowKitConnectProvider {
    children: React.ReactNode;
}

export const RainbowKitConnectProvider: React.FC<
    IRainbowKitConnectProvider
> = ({ children }: IRainbowKitConnectProvider) => {
    return (
        <WagmiProvider config={rainbowkitConfig}>
            <QueryClientProvider client={queryClient}>
                <RainbowKitProvider
                    locale='en'
                    theme={lightTheme({
                        accentColor: '#00E5FF',
                        borderRadius: 'small',
                        overlayBlur: 'small',
                    })}
                    appInfo={{ appName: 'TicTacToe' }}
                >
                    {children}
                </RainbowKitProvider>
            </QueryClientProvider>
        </WagmiProvider>
    );
};
