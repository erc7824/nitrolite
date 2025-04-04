import { defineChain } from 'viem';
import {
    base,
    bsc,
    gnosis,
    linea,
    lineaSepolia,
    mainnet,
    optimism,
    polygonMumbai,
    scroll,
    arbitrum,
    sepolia,
    avalanche,
    localhost,
} from 'viem/chains';

const polygon = defineChain({
    id: 137,
    name: 'Polygon',
    nativeCurrency: { name: 'POL', symbol: 'POL', decimals: 18 },
    rpcUrls: {
        default: {
            http: ['https://polygon.drpc.org', 'https://polygon-rpc.com'],
        },
    },
    blockExplorers: {
        default: {
            name: 'PolygonScan',
            url: 'https://polygonscan.com',
            apiUrl: 'https://api.polygonscan.com/api',
        },
    },
    contracts: {
        multicall3: {
            address: '0xca11bde05977b3631167028862be2a173976ca11',
            blockCreated: 25770160,
        },
    },
});

const supportedChains = (
    process.env.NEXT_PUBLIC_SUPPORTED_CHAINS?.split(',').map((chain) => {
        const sanitizedChain = chain.trim().replace(/^\"|\"$/g, '');
        const number = Number(sanitizedChain);

        return !isNaN(number) && number !== 0 ? number : null;
    }) ?? [137, 80002]
).filter((chain) => chain !== null);

const chainsArray = [
    mainnet,
    defineChain({ ...bsc, rpcUrls: { default: { http: ['https://bsc.drpc.org'] } } }), // default rpcUrls for sepolia are not working
    defineChain({ ...sepolia, rpcUrls: { default: { http: ['https://sepolia.gateway.tenderly.co'] } } }),
    base,
    arbitrum,
    avalanche,
    polygon,
    polygonMumbai,
    optimism,
    gnosis,
    linea,
    lineaSepolia,
    scroll,
    localhost,
];

export const chains = chainsArray.filter((chain) => supportedChains.includes(chain.id));
