/**
 * 1. Added getContractInfo() to Custody.sol
 * 2. Run forge build
 * 3. Run npm run codegen
 */

import { createPublicClient, http, Address } from "viem";
import { mainnet } from "viem/chains";
import { custodyAbi } from "./src/generated";

const publicClient = createPublicClient({
  chain: mainnet,
  transport: http(),
});

const CUSTODY_ADDRESS = "0x1234567890123456789012345678901234567890" as Address;

async function testNewFunction() {
  try {
    // This function was just added to the contract and is now fully typed
    const contractInfo = await publicClient.readContract({
      address: CUSTODY_ADDRESS,
      abi: custodyAbi,
      functionName: "getContractInfo",
      args: [],
    });
  } catch (error) {
    console.log("mock address, but types are fully generated");
  }
}

testNewFunction().catch(console.error);

export { testNewFunction };
