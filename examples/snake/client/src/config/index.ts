import type { ContractAddresses } from "@erc7824/nitrolite";
import type { Hex } from "viem";

export const BROKER_WS_URL = "wss://clearnode-multichain-production.up.railway.app/ws";
export const GAMESERVER_WS_URL = "ws://localhost:3001";

const getEnvVar = (key: string, defaultValue: string): string => {
    try {
        const envValue = (import.meta.env?.[`VITE_${key}`] || window?.__ENV__?.[key] || null) as string | null;
        return envValue || defaultValue;
    } catch (e) {
        console.warn(`Could not access environment variable ${key}, using default value`);
        return defaultValue;
    }
};

export const CONTRACT_ADDRESSES: ContractAddresses = {
    custody: getEnvVar("CUSTODY_ADDRESS", "0x1096644156Ed58BF596e67d35827Adc97A25D940") as Hex,
    adjudicator: getEnvVar("ADJUDICATOR_ADDRESS", "0xa3f2f64455c9f8D68d9dCAeC2605D64680FaF898") as Hex,
    tokenAddress: getEnvVar("TOKEN_ADDRESS", "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359") as Hex,
    guestAddress: getEnvVar("GUEST_ADDRESS", "0x3c93C321634a80FB3657CFAC707718A11cA57cBf") as Hex,
};
