"use client";

import { lightTheme, RainbowKitProvider, type AvatarComponent } from "@rainbow-me/rainbowkit";
import { WagmiProvider } from "wagmi";
import { QueryClientProvider, QueryClient } from "@tanstack/react-query";
import { connectorsForWallets } from "@rainbow-me/rainbowkit";
import { createConfig, http } from "wagmi";
import { polygon } from "viem/chains";
import { metaMaskWallet } from "@rainbow-me/rainbowkit/wallets";

const connectors = connectorsForWallets(
    [
        {
            groupName: "Recommended",
            wallets: [metaMaskWallet],
        },
    ],
    {
        appName: "Viper Duel",
        projectId: import.meta.env?.NEXT_PUBLIC_PROJECT_ID ?? "7e477e0ec531e636b5ab844fff3db798",
    }
);

export const rainbowkitConfig = createConfig({
    chains: [polygon],
    transports: {
        [polygon.id]: http(),
    },
    connectors,
    ssr: true,
});

const CustomAvatar: AvatarComponent = () => {
    return null;
};

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

export const RainbowKitConnectProvider: React.FC<IRainbowKitConnectProvider> = ({ children }: IRainbowKitConnectProvider) => {
    return (
        <WagmiProvider config={rainbowkitConfig}>
            <QueryClientProvider client={queryClient}>
                <RainbowKitProvider
                    avatar={CustomAvatar}
                    locale="en"
                    theme={lightTheme({
                        accentColor: "#ECAA00",
                        borderRadius: "small",
                        overlayBlur: "small",
                    })}
                    appInfo={{ appName: "Viper Duel" }}
                >
                    {children}
                </RainbowKitProvider>
            </QueryClientProvider>
        </WagmiProvider>
    );
};
