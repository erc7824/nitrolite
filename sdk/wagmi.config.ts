import { defineConfig } from "@wagmi/cli";
import { foundry } from "@wagmi/cli/plugins";

export default defineConfig({
  out: "src/generated.ts",
  contracts: [],
  plugins: [
    foundry({
      project: "../contract",
      include: [
        // Include only the main contracts we need
        "Custody.sol/**",
        "Dummy.sol/**",
        // Include adjudicators
        "Consensus.sol/**",
        "Counter.sol/**",
        "Remittance.sol/**",
      ],
      exclude: [
        // Exclude test files and OpenZeppelin dependencies
        "*.t.sol/**",
        "*.s.sol/**",
        "forge-std/**",
        "openzeppelin-contracts/**",
      ],
      // Add more verbose output
      forge: {
        build: true,
        rebuild: true,
      },
    }),
  ],
});
