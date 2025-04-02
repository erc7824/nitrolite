export const getInfuraEndpointNameByChainId = (chainId: number) => {
    switch (chainId) {
        case 1:
            return 'mainnet';
        case 11155111:
            return 'sepolia';
        case 137:
            return 'polygon-mainnet';
        case 59141:
            return 'linea-sepolia';
        case 59144:
            return 'linea-mainnet';
        case 80001:
            return 'polygon-amoy';
        case 8453:
            return 'base-mainnet';
        case 56:
            return 'bsc-mainnet';
        case 100:
            return 'gnosis';
        case 10:
            return 'optimism-mainnet';
        case 534352:
            return 'scroll';
        case 42161:
            return 'arbitrum-mainnet';
        default:
            return 'unknown';
    }
};
