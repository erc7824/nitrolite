import { describe, test, expect, jest, beforeEach } from "@jest/globals";
import { Erc20Service } from "../../../src/client/services/Erc20Service";
import { Errors } from "../../../src/errors";
import { Address, Hash, PublicClient, SimulateContractReturnType, WalletClient } from "viem";

// Mock data
const mockWalletClient = {
    account: {
        address: "0xUserAddress" as Address,
    },
    writeContract: jest.fn(() => Promise.resolve("0xTransactionHash" as Hash)),
} as unknown as WalletClient;

const mockPublicClient = {
    readContract: jest.fn(),
    simulateContract: jest.fn(),
} as unknown as PublicClient;

describe("Erc20Service", () => {
    let service: Erc20Service;
    const tokenAddress = "0xTokenAddress" as Address;
    const spender = "0xSpenderAddress" as Address;
    const owner = "0xOwnerAddress" as Address;
    const amount = BigInt(100);

    beforeEach(() => {
        jest.clearAllMocks();
        mockPublicClient.simulateContract.mockImplementation(() => Promise.resolve({ request: { to: "0x123", data: "0x456" } }));
        mockPublicClient.readContract.mockImplementation(() => Promise.resolve(BigInt(0)));

        service = new Erc20Service(mockPublicClient, mockWalletClient);
    });

    describe("constructor", () => {
        test("should throw when publicClient is missing", () => {
            expect(() => new Erc20Service(null as unknown as PublicClient, mockWalletClient)).toThrow(Errors.MissingParameterError);
        });

        test("should initialize correctly", () => {
            expect(service).toBeDefined();
        });
    });

    describe("ensureWalletClient", () => {
        test("should throw when walletClient is missing", () => {
            const serviceWithoutWallet = new Erc20Service(mockPublicClient);

            expect(() => {
                // @ts-ignore - Accessing private method for testing
                serviceWithoutWallet.ensureWalletClient();
            }).toThrow(Errors.WalletClientRequiredError);
        });
    });

    describe("getTokenBalance", () => {
        test("should call readContract with correct parameters", async () => {
            const mockBalance = BigInt(1000);
            mockPublicClient.readContract.mockImplementationOnce(() => Promise.resolve(mockBalance));

            const result = await service.getTokenBalance(tokenAddress, owner);

            expect(mockPublicClient.readContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    address: tokenAddress,
                    functionName: "balanceOf",
                    args: [owner],
                })
            );

            expect(result).toEqual(mockBalance);
        });

        test("should handle read errors", async () => {
            mockPublicClient.readContract.mockImplementationOnce(() => Promise.reject(new Error("Read failed")));

            await expect(service.getTokenBalance(tokenAddress, owner)).rejects.toThrow(Errors.ContractReadError);
        });
    });

    describe("getTokenAllowance", () => {
        test("should call readContract with correct parameters", async () => {
            const mockAllowance = BigInt(500);
            mockPublicClient.readContract.mockImplementationOnce(() => Promise.resolve(mockAllowance));

            const result = await service.getTokenAllowance(tokenAddress, owner, spender);

            expect(mockPublicClient.readContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    address: tokenAddress,
                    functionName: "allowance",
                    args: [owner, spender],
                })
            );

            expect(result).toEqual(mockAllowance);
        });
    });

    describe("approve", () => {
        test("should prepare and submit an approve transaction", async () => {
            await expect(service.approve(tokenAddress, spender, amount)).resolves.toBe("0xTransactionHash");

            expect(mockPublicClient.simulateContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    address: tokenAddress,
                    functionName: "approve",
                    args: [spender, amount],
                })
            );

            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });

        test("should handle simulation errors", async () => {
            mockPublicClient.simulateContract.mockImplementationOnce(() => Promise.reject(new Error("Simulation failed")));

            await expect(service.approve(tokenAddress, spender, amount)).rejects.toThrow(Errors.ContractCallError);
        });

        test("should handle transaction errors", async () => {
            mockWalletClient.writeContract.mockImplementationOnce(() => Promise.reject(new Error("Transaction failed")));

            await expect(service.approve(tokenAddress, spender, amount)).rejects.toThrow(Errors.TransactionError);
        });
    });

    describe("prepareApprove", () => {
        test("should return a valid request object", async () => {
            const mockRequest = { to: "0xTo", data: "0xData" };

            mockPublicClient.simulateContract.mockImplementationOnce(() => Promise.resolve({ request: mockRequest }));

            const result = await service.prepareApprove(tokenAddress, spender, amount);

            expect(result).toEqual(mockRequest);
        });
    });
});
