import type { ContractAddresses } from "@erc7824/nitrolite";
import type { Hex } from "viem";
import { getCustodyAddress, getAdjudicatorAddress } from "./contractAddresses.js";

const getEnvVar = (key: string, defaultValue: string): string => {
    try {
        const envValue = (import.meta.env?.[`VITE_${key}`] || window?.__ENV__?.[key] || null) as string | null;
        return envValue || defaultValue;
    } catch (e) {
        console.warn(`Could not access environment variable ${key}, using default value`);
        return defaultValue;
    }
};

export const BROKER_WS_URL = getEnvVar("BROKER_WS_URL", "wss://clearnode-multichain-production.up.railway.app/ws");
export const GAMESERVER_WS_URL = getEnvVar("GAMESERVER_WS_URL", "ws://localhost:3001");

// Contract addresses - read from latest broadcast files
export const CONTRACT_ADDRESSES: ContractAddresses = {
    custody: (getEnvVar("CUSTODY_ADDRESS", "") as Hex) || getCustodyAddress(),
    adjudicator: (getEnvVar("ADJUDICATOR_ADDRESS", "") as Hex) || getAdjudicatorAddress(),
    tokenAddress: getEnvVar("TOKEN_ADDRESS", "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359") as Hex,
    guestAddress: getEnvVar("GUEST_ADDRESS", "0x3c93C321634a80FB3657CFAC707718A11cA57cBf") as Hex,
};

export const CHAIN_ID = parseInt(getEnvVar("CHAIN_ID", "137"), 10);
