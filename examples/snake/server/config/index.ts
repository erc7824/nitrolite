import { ContractAddresses } from "@erc7824/nitrolite";
import { Hex } from "viem";

export const BROKER_WS_URL = process.env.BROKER_WS_URL as string;
export const SERVER_PRIVATE_KEY = process.env.SERVER_PRIVATE_KEY as Hex;
export const WALLET_PRIVATE_KEY = process.env.WALLET_PRIVATE_KEY as Hex;
export const POLYGON_RPC_URL = process.env.POLYGON_RPC_URL as string;

// Contract addresses
export const CONTRACT_ADDRESSES: ContractAddresses = {
    custody: (process.env.CUSTODY_ADDRESS || "0x1096644156Ed58BF596e67d35827Adc97A25D940") as Hex,
    adjudicator: (process.env.ADJUDICATOR_ADDRESS || "0xa3f2f64455c9f8D68d9dCAeC2605D64680FaF898") as Hex,
    tokenAddress: (process.env.TOKEN_ADDRESS || "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359") as Hex,
    guestAddress: (process.env.GUEST_ADDRESS || "0x3c93C321634a80FB3657CFAC707718A11cA57cBf") as Hex, // broker channel address is used here
};
