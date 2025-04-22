import { CustodyAbi, AdjudicatorAbi, Erc20Abi, ContractAddresses, defaultAbiConfig } from "../../src/abis";

describe("ABIs", () => {
    describe("CustodyAbi", () => {
        it("should export a valid ABI", () => {
            expect(Array.isArray(CustodyAbi)).toBe(true);
            expect(CustodyAbi.length).toBeGreaterThan(0);

            // Check for expected functions
            const hasFunctions = CustodyAbi.some(
                (item) => item.type === "function" && ["open", "close", "challenge", "checkpoint", "reclaim"].includes(item.name as string)
            );

            expect(hasFunctions).toBe(true);
        });
    });

    describe("AdjudicatorAbi", () => {
        it("should export a valid ABI", () => {
            expect(Array.isArray(AdjudicatorAbi)).toBe(true);
            expect(AdjudicatorAbi.length).toBeGreaterThan(0);

            // Check for adjudicate function
            const hasAdjudicate = AdjudicatorAbi.some((item) => item.type === "function" && item.name === "adjudicate");

            expect(hasAdjudicate).toBe(true);
        });
    });

    describe("Erc20Abi", () => {
        it("should export a valid ABI", () => {
            expect(Array.isArray(Erc20Abi)).toBe(true);
            expect(Erc20Abi.length).toBeGreaterThan(0);

            // Check for expected functions
            const hasFunctions = Erc20Abi.some(
                (item) => item.type === "function" && ["approve", "transfer", "transferFrom", "balanceOf", "allowance"].includes(item.name as string)
            );

            expect(hasFunctions).toBe(true);
        });
    });

    describe("defaultAbiConfig", () => {
        it("should have correct structure", () => {
            expect(defaultAbiConfig).toBeDefined();
            expect(typeof defaultAbiConfig.chainId).toBe("number");
            expect(defaultAbiConfig.addresses).toBeDefined();
            expect(defaultAbiConfig.addresses.custody).toBeDefined();
            expect(defaultAbiConfig.addresses.adjudicators).toBeDefined();
        });

        it("should have placeholder addresses", () => {
            // These should be placeholders, not real addresses
            expect(defaultAbiConfig.addresses.custody).toMatch(/^0x0+$/);

            // Check all adjudicator addresses are placeholders
            Object.values(defaultAbiConfig.addresses.adjudicators).forEach((address) => {
                expect(address).toMatch(/^0x0+$/);
            });
        });

        it("should include warning about placeholders", () => {
            // Implementation check: verify the explanation above the config
            // This is a bit hacky for testing, but it helps ensure documentation is clear
            const moduleContent = require("fs").readFileSync(require("path").resolve(__dirname, "../../src/abis/index.ts"), "utf8");

            expect(moduleContent).toContain("not for use");
            expect(moduleContent).toContain("placeholder");
        });
    });

    describe("ContractAddresses", () => {
        it("should allow creating valid contract addresses object", () => {
            const addresses: ContractAddresses = {
                custody: "0x1111111111111111111111111111111111111111",
                adjudicators: {
                    base: "0x2222222222222222222222222222222222222222",
                    numeric: "0x3333333333333333333333333333333333333333",
                    custom: "0x4444444444444444444444444444444444444444",
                },
            };

            expect(addresses.custody).toBe("0x1111111111111111111111111111111111111111");
            expect(addresses.adjudicators.base).toBe("0x2222222222222222222222222222222222222222");
            expect(addresses.adjudicators.numeric).toBe("0x3333333333333333333333333333333333333333");
            expect(addresses.adjudicators.custom).toBe("0x4444444444444444444444444444444444444444");
        });
    });
});
