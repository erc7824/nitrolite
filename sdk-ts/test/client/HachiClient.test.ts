import { NitroliteClient } from "../../src/client/NitroliteClient";
import { PublicClient, WalletClient, createPublicClient, http, Abi } from "viem";
import { Channel, State } from "../../src/types";

// Mock dependencies
jest.mock("viem", () => {
    const originalModule = jest.requireActual("viem");

    return {
        ...originalModule,
        createPublicClient: jest.fn(),
        http: jest.fn(),
    };
});

// Mock PublicClient
const mockPublicClient = {
    chain: { id: 1 },
    readContract: jest.fn(),
    simulateContract: jest.fn(),
    waitForTransactionReceipt: jest.fn(),
} as unknown as PublicClient;

// Mock WalletClient
const mockWalletClient = {
    writeContract: jest.fn(),
} as unknown as WalletClient;

// Mock account
const mockAccount = {
    address: "0x1111111111111111111111111111111111111111",
};

// Test data
const mockCustodyAddress = "0x2222222222222222222222222222222222222222";
const mockAdjudicatorAddress = "0x3333333333333333333333333333333333333333";

describe("NitroliteClient", () => {
    beforeEach(() => {
        jest.clearAllMocks();
        (createPublicClient as jest.Mock).mockReturnValue(mockPublicClient);
        (http as jest.Mock).mockReturnValue("mock-http-transport");

        mockPublicClient.simulateContract = jest.fn().mockResolvedValue({
            request: {},
        });
        mockWalletClient.writeContract = jest.fn().mockResolvedValue("0xTRANSACTION_HASH");
        mockPublicClient.waitForTransactionReceipt = jest.fn().mockResolvedValue({});
    });

    describe("constructor", () => {
        it("should initialize with minimal config", () => {
            const client = new NitroliteClient({
                publicClient: mockPublicClient,
                chainId: 1,
                addresses: {
                    custody: mockCustodyAddress,
                    adjudicators: {
                        base: mockAdjudicatorAddress,
                    },
                },
            });

            expect(client.publicClient).toBe(mockPublicClient);
            expect(client.chainId).toBe(1);
            expect(client.custodyAddress).toBe(mockCustodyAddress);
        });

        it("should initialize with wallet client and account", () => {
            const client = new NitroliteClient({
                publicClient: mockPublicClient,
                walletClient: mockWalletClient,
                account: mockAccount,
                chainId: 1,
                addresses: {
                    custody: mockCustodyAddress,
                    adjudicators: {
                        base: mockAdjudicatorAddress,
                    },
                },
            });

            expect(client.publicClient).toBe(mockPublicClient);
            expect(client.walletClient).toBe(mockWalletClient);
            expect(client.account).toBe(mockAccount);
        });

        it("should detect chain ID from public client if not provided", () => {
            const client = new NitroliteClient({
                publicClient: mockPublicClient,
                addresses: {
                    custody: mockCustodyAddress,
                    adjudicators: {
                        base: mockAdjudicatorAddress,
                    },
                },
            });

            expect(client.chainId).toBe(1); // From mockPublicClient.chain.id
        });

        it("should throw error if no addresses provided", () => {
            expect(() => {
                new NitroliteClient({
                    publicClient: mockPublicClient,
                    chainId: 1,
                });
            }).toThrow();
        });
    });

    describe("adjudicator management", () => {
        let client: NitroliteClient;

        beforeEach(() => {
            client = new NitroliteClient({
                publicClient: mockPublicClient,
                chainId: 1,
                addresses: {
                    custody: mockCustodyAddress,
                    adjudicators: {
                        base: mockAdjudicatorAddress,
                        numeric: "0x4444444444444444444444444444444444444444",
                    },
                },
            });
        });

        it("should get adjudicator address by type", () => {
            const address = client.getAdjudicatorAddress("base");
            expect(address).toBe(mockAdjudicatorAddress);

            const numericAddress = client.getAdjudicatorAddress("numeric");
            expect(numericAddress).toBe("0x4444444444444444444444444444444444444444");
        });

        it("should fall back to base adjudicator if type not found", () => {
            const address = client.getAdjudicatorAddress("unknown", true);
            expect(address).toBe(mockAdjudicatorAddress);
        });

        it("should throw if adjudicator type not found and no fallback", () => {
            expect(() => {
                client.getAdjudicatorAddress("unknown", false);
            }).toThrow();
        });

        it("should register and retrieve custom adjudicator ABIs", () => {
            const mockAbi: Abi = [{ type: "function", name: "test" }];
            client.registerAdjudicatorAbi("custom", mockAbi);

            const retrievedAbi = client.getAdjudicatorAbi("custom");
            expect(retrievedAbi).toBe(mockAbi);
        });

        it("should get base adjudicator ABI as fallback", () => {
            const abi = client.getAdjudicatorAbi("unknown");
            expect(abi).toBeTruthy(); // Should return the default ABI
        });
    });

    describe("channel operations", () => {
        let client: NitroliteClient;

        beforeEach(() => {
            client = new NitroliteClient({
                publicClient: mockPublicClient,
                walletClient: mockWalletClient,
                account: mockAccount,
                chainId: 1,
                addresses: {
                    custody: mockCustodyAddress,
                    adjudicators: {
                        base: mockAdjudicatorAddress,
                        numeric: "0x4444444444444444444444444444444444444444",
                    },
                },
            });
        });

        it("should create a numeric channel", () => {
            const channel = client.createNumericChannel({
                participants: ["0x1111111111111111111111111111111111111111", "0x5555555555555555555555555555555555555555"],
            });

            expect(channel).toBeDefined();
            expect(channel.getChannel().adjudicator).toBe("0x4444444444444444444444444444444444444444");
        });

        it("should create a sequential channel", () => {
            const channel = client.createSequentialChannel({
                participants: ["0x1111111111111111111111111111111111111111", "0x5555555555555555555555555555555555555555"],
            });

            expect(channel).toBeDefined();
        });

        it("should create a custom channel", () => {
            const channel = client.createCustomChannel({
                participants: ["0x1111111111111111111111111111111111111111", "0x5555555555555555555555555555555555555555"],
                encode: (data) => "0x1234",
                decode: (encoded) => ({ value: 42 }),
            });

            expect(channel).toBeDefined();
            expect(channel.getChannel().adjudicator).toBe(mockAdjudicatorAddress);
        });

        it("should create a custom channel with custom adjudicator", () => {
            const customAdjudicator = "0x6666666666666666666666666666666666666666";
            const channel = client.createCustomChannel({
                participants: ["0x1111111111111111111111111111111111111111", "0x5555555555555555555555555555555555555555"],
                adjudicatorAddress: customAdjudicator,
                encode: (data) => "0x1234",
                decode: (encoded) => ({ value: 42 }),
            });

            expect(channel).toBeDefined();
            expect(channel.getChannel().adjudicator).toBe(customAdjudicator);
        });
    });
});
