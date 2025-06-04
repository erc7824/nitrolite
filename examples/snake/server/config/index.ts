import { Hex } from "viem";

export const BROKER_WS_URL = process.env.BROKER_WS_URL as string;
export const SERVER_PRIVATE_KEY = process.env.SERVER_PRIVATE_KEY as Hex;
export const WALLET_PRIVATE_KEY = process.env.WALLET_PRIVATE_KEY as Hex;
export const CHAIN_ID = parseInt(process.env.CHAIN_ID || "137");
