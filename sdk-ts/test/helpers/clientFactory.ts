import { NitroliteClient } from "../../src/client/NitroliteClient";
import { PublicClient, WalletClient } from "viem";

/**
 * Creates a mock public client for testing
 * @returns Mock public client
 */
export function createMockPublicClient(): PublicClient {
    return {
        chain: { id: 1 },
        readContract: jest.fn(),
        simulateContract: jest.fn().mockResolvedValue({
            request: {},
        }),
        waitForTransactionReceipt: jest.fn().mockResolvedValue({}),
    } as unknown as PublicClient;
}

/**
 * Creates a mock wallet client for testing
 * @returns Mock wallet client
 */
export function createMockWalletClient(): WalletClient {
    return {
        writeContract: jest.fn().mockResolvedValue("0xTRANSACTION_HASH"),
    } as unknown as WalletClient;
}

/**
 * Creates a mock account for testing
 * @param address Optional address to use, defaults to a test address
 * @returns Mock account
 */
export function createMockAccount(address: string = "0x1111111111111111111111111111111111111111") {
    return { address };
}

/**
 * Creates a test Nitrolite client for testing
 * @param options Optional configuration options
 * @returns Mock Nitrolite client
 */
export function createTestClient(
    options: {
        publicClient?: PublicClient;
        walletClient?: WalletClient;
        account?: { address: string };
        custodyAddress?: string;
        adjudicatorAddress?: string;
    } = {}
): NitroliteClient {
    const mockCustodyAddress = options.custodyAddress || "0x2222222222222222222222222222222222222222";
    const mockAdjudicatorAddress = options.adjudicatorAddress || "0x3333333333333333333333333333333333333333";

    return new NitroliteClient({
        publicClient: options.publicClient || createMockPublicClient(),
        walletClient: options.walletClient || createMockWalletClient(),
        account: options.account || createMockAccount(),
        chainId: 1,
        addresses: {
            custody: mockCustodyAddress,
            adjudicators: {
                base: mockAdjudicatorAddress,
                numeric: "0x4444444444444444444444444444444444444444",
                sequential: "0x5555555555555555555555555555555555555555",
            },
        },
    });
}
