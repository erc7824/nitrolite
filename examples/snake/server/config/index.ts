import { ContractAddresses } from "@erc7824/nitrolite";
import { Hex } from "viem";
import { getCustodyAddress, getAdjudicatorAddress } from "./contractAddresses.js";

export const BROKER_WS_URL = process.env.BROKER_WS_URL as string;
export const SERVER_PRIVATE_KEY = process.env.SERVER_PRIVATE_KEY as Hex;
export const WALLET_PRIVATE_KEY = process.env.WALLET_PRIVATE_KEY as Hex;
export const POLYGON_RPC_URL = process.env.POLYGON_RPC_URL as string;

// Contract addresses - read from latest broadcast files
export const CONTRACT_ADDRESSES: ContractAddresses = {
    custody: process.env.CUSTODY_ADDRESS as Hex || getCustodyAddress(),
    adjudicator: process.env.ADJUDICATOR_ADDRESS as Hex || getAdjudicatorAddress(),
    tokenAddress: (process.env.TOKEN_ADDRESS || "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359") as Hex,
    guestAddress: (process.env.GUEST_ADDRESS || "0x3c93C321634a80FB3657CFAC707718A11cA57cBf") as Hex, // broker channel address is used here
};
