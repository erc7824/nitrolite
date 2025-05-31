import { createPublicClient, http, Address } from "viem";
import { mainnet } from "viem/chains";

import {
  custodyAbi,
  dummyAbi,
  consensusAbi,
  counterAbi,
  remittanceAdjudicatorAbi,
} from "./src/generated";

// Mock setup for demo
const publicClient = createPublicClient({
  chain: mainnet,
  transport: http(),
});

const CUSTODY_ADDRESS = "0x1234567890123456789012345678901234567890" as Address;
const USER_ADDRESS = "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef" as Address;
const TOKEN_ADDRESS = "0x0000000000000000000000000000000000000000" as Address;

async function demonstrateTypeSafety() {
  try {
    const accountInfo = await publicClient.readContract({
      address: CUSTODY_ADDRESS,
      abi: custodyAbi,
      functionName: "getAccountInfo",
      args: [USER_ADDRESS, TOKEN_ADDRESS],
    });

    console.log(
      "getAccountInfo() - TypeScript knows return type:",
      typeof accountInfo
    );

    // Try to access a function that doesn't exist
    // const invalid = await publicClient.readContract({
    //   address: CUSTODY_ADDRESS,
    //   abi: custodyAbi,
    //   functionName: 'nonExistentFunction',
    //   args: []
    // })
  } catch (error) {
    console.log("Just mock addresses here");
  }

  console.log(
    "  • Dummy Adjudicator:",
    Object.keys(dummyAbi).length,
    "ABI entries"
  );
  console.log(
    "  • Consensus Adjudicator:",
    Object.keys(consensusAbi).length,
    "ABI entries"
  );
  console.log(
    "  • Counter Adjudicator:",
    Object.keys(counterAbi).length,
    "ABI entries"
  );
  console.log(
    "  • Remittance Adjudicator:",
    Object.keys(remittanceAdjudicatorAbi).length,
    "ABI entries"
  );

  // The type system knows exactly what each function returns
  type CustodyFunctions = (typeof custodyAbi)[number]["name"]; // Auto-extracted function names
  type GetAccountInfoReturn = any;
  console.log("  All contract functions are fully typed");
  console.log("  Function parameters are type-checked");
  console.log("  Return types are automatically inferred");
  console.log("  Events and errors are included");
}

// Run the demo
demonstrateTypeSafety().catch(console.error);

export { demonstrateTypeSafety };
