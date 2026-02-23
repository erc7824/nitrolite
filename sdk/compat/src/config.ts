import { type Option, withBlockchainRPC } from '@erc7824/nitrolite';

export interface CompatClientConfig {
    wsURL: string;
    blockchainRPCs?: Record<number, string>;
}

export function buildClientOptions(config: CompatClientConfig): Option[] {
    const opts: Option[] = [];

    if (config.blockchainRPCs) {
        for (const [chainId, rpcUrl] of Object.entries(config.blockchainRPCs)) {
            opts.push(withBlockchainRPC(BigInt(chainId), rpcUrl));
        }
    }

    return opts;
}

/**
 * Reads blockchain RPC URLs from NEXT_PUBLIC_BLOCKCHAIN_RPCS env var.
 * Format: "chainId:rpcUrl,chainId:rpcUrl"
 * Example: "11155111:https://rpc.sepolia.io"
 */
export function blockchainRPCsFromEnv(): Record<number, string> {
    const raw = process.env.NEXT_PUBLIC_BLOCKCHAIN_RPCS ?? '';
    return raw
        .split(',')
        .filter(Boolean)
        .reduce<Record<number, string>>((acc, pair) => {
            const idx = pair.indexOf(':');
            if (idx === -1) return acc;
            const chainId = Number(pair.slice(0, idx).trim());
            const rpcUrl = pair.slice(idx + 1).trim();
            if (Number.isFinite(chainId) && rpcUrl) acc[chainId] = rpcUrl;
            return acc;
        }, {});
}
