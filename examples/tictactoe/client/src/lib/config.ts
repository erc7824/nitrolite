import { connectorsForWallets } from "@rainbow-me/rainbowkit";
import { createConfig, http } from "wagmi";
import { mainnet } from "wagmi/chains";

import { toPrivyWallet } from "@privy-io/cross-app-connect/rainbow-kit";

export const connectors = connectorsForWallets(
    [
        {
            groupName: "Recommended",
            wallets: [
                toPrivyWallet({
                    id: "cmbevh1e30022l80nhp974z8m",
                    name: "Yellow Wallet",
                    iconUrl: "https://yellow.com/favicon.ico",
                }),
            ],
        },
    ],
    {
        appName: "Tic Tac Toe",
        projectId: "Demo",
    }
);

export const config = createConfig({
    chains: [mainnet],
    transports: {
        [mainnet.id]: http(),
    },
    connectors,
    ssr: true,
});
