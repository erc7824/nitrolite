import { describe, test, expect, jest, beforeEach } from "@jest/globals";
import { NitroliteService } from "../../../src/client/services/NitroliteService";
import { Errors } from "../../../src/errors";
import { Account, Address, Hash, Hex, PublicClient, SimulateContractReturnType, WalletClient, zeroAddress } from "viem";
import { ContractAddresses } from "../../../src/abis";
import { Channel, ChannelId, Signature, State } from "../../../src/client/types";

// Mock data
const mockAddresses: ContractAddresses = {
    custody: "0xCustodyAddress" as Address,
    adjudicators: {
        default: "0xAdjudicatorAddress" as Address,
    },
    guestAddress: "0xGuestAddress" as Address,
    tokenAddress: "0xTokenAddress" as Address,
};

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

const mockChannel: Channel = {
    participants: ["0xParticipant1" as Address, "0xParticipant2" as Address],
    adjudicator: "0xAdjudicatorAddress" as Address,
    challenge: BigInt(86400),
    nonce: BigInt(1),
};

const mockState: State = {
    data: "0x00" as Hex,
    allocations: [
        { destination: "0xParticipant1" as Address, token: "0xTokenAddress" as Address, amount: BigInt(100) },
        { destination: "0xParticipant2" as Address, token: "0xTokenAddress" as Address, amount: BigInt(0) },
    ],
    sigs: [{ v: 27, r: "0x00" as Hex, s: "0x00" as Hex }],
};

const mockChannelId = "0xChannelId" as ChannelId;
const mockSignature: Signature = { v: 27, r: "0x00" as Hex, s: "0x00" as Hex };

describe("NitroliteService", () => {
    let service: NitroliteService;

    beforeEach(() => {
        jest.clearAllMocks();
        mockPublicClient.simulateContract.mockImplementation(() => Promise.resolve({ request: { to: "0x123", data: "0x456" } }));
        mockPublicClient.readContract.mockImplementation(() => Promise.resolve([]));

        service = new NitroliteService(mockPublicClient, mockAddresses, mockWalletClient);
    });

    describe("constructor", () => {
        test("should throw when publicClient is missing", () => {
            expect(() => new NitroliteService(null as unknown as PublicClient, mockAddresses, mockWalletClient)).toThrow(
                Errors.MissingParameterError
            );
        });

        test("should throw when addresses.custody is missing", () => {
            expect(
                () => new NitroliteService(mockPublicClient, { ...mockAddresses, custody: undefined as unknown as Address }, mockWalletClient)
            ).toThrow(Errors.MissingParameterError);
        });

        test("should initialize correctly", () => {
            expect(service).toBeDefined();
            expect(service.custodyAddress).toBe(mockAddresses.custody);
        });
    });

    describe("ensureWalletClient", () => {
        test("should throw when walletClient is missing", () => {
            const serviceWithoutWallet = new NitroliteService(mockPublicClient, mockAddresses);

            expect(() => {
                // @ts-ignore - Accessing private method for testing
                serviceWithoutWallet.ensureWalletClient();
            }).toThrow(Errors.WalletClientRequiredError);
        });
    });

    describe("ensureAccount", () => {
        test("should throw when account is missing", () => {
            const walletWithoutAccount = { ...mockWalletClient, account: undefined };
            const serviceWithoutAccount = new NitroliteService(mockPublicClient, mockAddresses, walletWithoutAccount as unknown as WalletClient);

            expect(() => {
                // @ts-ignore - Accessing private method for testing
                serviceWithoutAccount.ensureAccount();
            }).toThrow(Errors.AccountRequiredError);
        });
    });

    describe("deposit", () => {
        test("should prepare and submit a deposit transaction", async () => {
            const tokenAddress = mockAddresses.tokenAddress;
            const amount = BigInt(100);

            await expect(service.deposit(tokenAddress, amount)).resolves.toBe("0xTransactionHash");

            expect(mockPublicClient.simulateContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    address: mockAddresses.custody,
                    functionName: "deposit",
                    args: [tokenAddress, amount],
                })
            );

            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });

        test("should handle ETH deposits correctly (with value)", async () => {
            const tokenAddress = zeroAddress;
            const amount = BigInt(100);

            await service.deposit(tokenAddress, amount);

            expect(mockPublicClient.simulateContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    value: amount,
                })
            );
        });

        test("should handle simulation errors", async () => {
            mockPublicClient.simulateContract.mockImplementationOnce(() => Promise.reject(new Error("Simulation failed")));

            await expect(service.deposit("0xToken" as Address, BigInt(100))).rejects.toThrow(Errors.ContractCallError);
        });

        test("should handle transaction errors", async () => {
            mockWalletClient.writeContract.mockImplementationOnce(() => Promise.reject(new Error("Transaction failed")));

            await expect(service.deposit("0xToken" as Address, BigInt(100))).rejects.toThrow(Errors.TransactionError);
        });
    });

    describe("createChannel", () => {
        test("should prepare and submit a createChannel transaction", async () => {
            await expect(service.createChannel(mockChannel, mockState)).resolves.toBe("0xTransactionHash");

            expect(mockPublicClient.simulateContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    address: mockAddresses.custody,
                    functionName: "create",
                    args: [mockChannel, mockState],
                })
            );

            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });
    });

    describe("joinChannel", () => {
        test("should prepare and submit a joinChannel transaction", async () => {
            const index = BigInt(1);

            await expect(service.joinChannel(mockChannelId, index, mockSignature)).resolves.toBe("0xTransactionHash");

            expect(mockPublicClient.simulateContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    address: mockAddresses.custody,
                    functionName: "join",
                    args: [mockChannelId, index, mockSignature],
                })
            );

            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });
    });

    describe("checkpoint", () => {
        test("should prepare and submit a checkpoint transaction", async () => {
            const proofStates: State[] = [];

            await expect(service.checkpoint(mockChannelId, mockState, proofStates)).resolves.toBe("0xTransactionHash");

            expect(mockPublicClient.simulateContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    address: mockAddresses.custody,
                    functionName: "checkpoint",
                    args: [mockChannelId, mockState, proofStates],
                })
            );

            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });
    });

    describe("challenge", () => {
        test("should prepare and submit a challenge transaction", async () => {
            const proofStates: State[] = [];

            await expect(service.challenge(mockChannelId, mockState, proofStates)).resolves.toBe("0xTransactionHash");

            expect(mockPublicClient.simulateContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    address: mockAddresses.custody,
                    functionName: "challenge",
                    args: [mockChannelId, mockState, proofStates],
                })
            );

            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });
    });

    describe("close", () => {
        test("should prepare and submit a close transaction", async () => {
            const proofStates: State[] = [];

            await expect(service.close(mockChannelId, mockState, proofStates)).resolves.toBe("0xTransactionHash");

            expect(mockPublicClient.simulateContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    address: mockAddresses.custody,
                    functionName: "close",
                    args: [mockChannelId, mockState, proofStates],
                })
            );

            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });
    });

    describe("reset", () => {
        test("should prepare and submit a reset transaction", async () => {
            const proofStates: State[] = [];
            const newChannel = { ...mockChannel, nonce: BigInt(2) };
            const newDeposit = { ...mockState, data: "0x01" as Hex };

            await expect(service.reset(mockChannelId, mockState, proofStates, newChannel, newDeposit)).resolves.toBe("0xTransactionHash");

            expect(mockPublicClient.simulateContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    address: mockAddresses.custody,
                    functionName: "reset",
                    args: [mockChannelId, mockState, proofStates, newChannel, newDeposit],
                })
            );

            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });
    });

    describe("withdraw", () => {
        test("should prepare and submit a withdraw transaction", async () => {
            const tokenAddress = mockAddresses.tokenAddress;
            const amount = BigInt(100);

            await expect(service.withdraw(tokenAddress, amount)).resolves.toBe("0xTransactionHash");

            expect(mockPublicClient.simulateContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    address: mockAddresses.custody,
                    functionName: "withdraw",
                    args: [tokenAddress, amount],
                })
            );

            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });
    });

    describe("getAccountChannels", () => {
        test("should call readContract with correct parameters", async () => {
            const account = "0xUserAddress" as Address;
            const mockChannels = ["0xChannel1", "0xChannel2"];

            mockPublicClient.readContract.mockImplementationOnce(() => Promise.resolve(mockChannels));

            const result = await service.getAccountChannels(account);

            expect(mockPublicClient.readContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    address: mockAddresses.custody,
                    functionName: "getAccountChannels",
                    args: [account],
                })
            );

            expect(result).toEqual(mockChannels);
        });

        test("should handle read errors", async () => {
            mockPublicClient.readContract.mockImplementationOnce(() => Promise.reject(new Error("Read failed")));

            await expect(service.getAccountChannels("0xAccount" as Address)).rejects.toThrow(Errors.ContractReadError);
        });
    });

    describe("getAccountInfo", () => {
        test("should call readContract with correct parameters and parse result", async () => {
            const user = "0xUserAddress" as Address;
            const token = "0xTokenAddress" as Address;
            const mockInfo = [BigInt(100), BigInt(200), BigInt(2)]; // [available, locked, channelCount]

            mockPublicClient.readContract.mockImplementationOnce(() => Promise.resolve(mockInfo));

            const result = await service.getAccountInfo(user, token);

            expect(mockPublicClient.readContract).toHaveBeenCalledWith(
                expect.objectContaining({
                    address: mockAddresses.custody,
                    functionName: "getAccountInfo",
                    args: [user, token],
                })
            );

            expect(result).toEqual({
                available: BigInt(100),
                locked: BigInt(200),
                channelCount: BigInt(2),
            });
        });
    });

    // Test prepare* methods
    describe("prepareDeposit", () => {
        test("should return a valid request object", async () => {
            const mockRequest = { to: "0xTo", data: "0xData" };

            mockPublicClient.simulateContract.mockImplementationOnce(() => Promise.resolve({ request: mockRequest }));

            const result = await service.prepareDeposit("0xToken" as Address, BigInt(100));

            expect(result).toEqual(mockRequest);
        });
    });

    // Add similar tests for other prepare* methods
});
