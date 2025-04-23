import { describe, test, expect, jest, beforeEach } from "@jest/globals";
import { NitroliteTransactionPreparer, PreparerDependencies } from "../../src/client/prepare";
import { NitroliteService } from "../../src/client/services/NitroliteService";
import { Erc20Service } from "../../src/client/services/Erc20Service";
import { Errors } from "../../src/errors";
import { Account, Address, Hash, Hex, ParseAccount, WalletClient, zeroAddress } from "viem";
import { ContractAddresses } from "../../src/abis";
import {
    ChallengeChannelParams,
    Channel,
    ChannelId,
    CheckpointChannelParams,
    CloseChannelParams,
    CreateChannelParams,
    Signature,
    State,
} from "../../src/client/types";

// Mock dependencies and data
const mockAddresses: ContractAddresses = {
    custody: "0xCustodyAddress" as Address,
    adjudicators: {
        default: "0xAdjudicatorAddress" as Address,
    },
    guestAddress: "0xGuestAddress" as Address,
    tokenAddress: "0xTokenAddress" as Address,
};

const mockAccount = {
    address: "0xUserAddress" as Address,
};

const mockWalletClient = {
    account: mockAccount,
    signMessage: jest.fn(() => Promise.resolve("0xSignature" as Hex)),
} as unknown as WalletClient;

const mockStateWalletClient = {
    account: mockAccount,
    signMessage: jest.fn(() => Promise.resolve("0xStateSignature" as Hex)),
} as unknown as WalletClient;

const mockNitroliteService = {
    prepareDeposit: jest.fn(),
    prepareCreateChannel: jest.fn(),
    prepareCheckpoint: jest.fn(),
    prepareChallenge: jest.fn(),
    prepareClose: jest.fn(),
    prepareWithdraw: jest.fn(),
    deposit: jest.fn(),
    createChannel: jest.fn(),
    challenge: jest.fn(),
    checkpoint: jest.fn(),
    close: jest.fn(),
    withdraw: jest.fn(),
} as unknown as NitroliteService;

const mockErc20Service = {
    prepareApprove: jest.fn(),
    getTokenAllowance: jest.fn(),
    approve: jest.fn(),
} as unknown as Erc20Service;

const mockDependencies: PreparerDependencies = {
    nitroliteService: mockNitroliteService,
    erc20Service: mockErc20Service,
    addresses: mockAddresses,
    account: mockAccount as ParseAccount<Account>,
    walletClient: mockWalletClient,
    stateWalletClient: mockStateWalletClient,
    challengeDuration: BigInt(86400),
};

const mockChannelId = "0xChannelId" as ChannelId;
const mockState: State = {
    data: "0x00" as Hex,
    allocations: [
        { destination: "0xParticipant1" as Address, token: "0xTokenAddress" as Address, amount: BigInt(100) },
        { destination: "0xParticipant2" as Address, token: "0xTokenAddress" as Address, amount: BigInt(0) },
    ],
    sigs: [{ v: 27, r: "0x00" as Hex, s: "0x00" as Hex }],
};
const mockSignature: Signature = { v: 27, r: "0x00" as Hex, s: "0x00" as Hex };

describe("NitroliteTransactionPreparer", () => {
    let preparer: NitroliteTransactionPreparer;

    beforeEach(() => {
        jest.clearAllMocks();

        // Reset mock implementations
        mockNitroliteService.prepareDeposit.mockResolvedValue({ to: "0x123", data: "0x456" });
        mockNitroliteService.prepareCreateChannel.mockResolvedValue({ to: "0x123", data: "0x456" });
        mockNitroliteService.prepareCheckpoint.mockResolvedValue({ to: "0x123", data: "0x456" });
        mockNitroliteService.prepareChallenge.mockResolvedValue({ to: "0x123", data: "0x456" });
        mockNitroliteService.prepareClose.mockResolvedValue({ to: "0x123", data: "0x456" });
        mockNitroliteService.prepareWithdraw.mockResolvedValue({ to: "0x123", data: "0x456" });

        mockErc20Service.prepareApprove.mockResolvedValue({ to: "0x123", data: "0x456" });
        mockErc20Service.getTokenAllowance.mockResolvedValue(BigInt(0));

        preparer = new NitroliteTransactionPreparer(mockDependencies);
    });

    describe("prepareDepositTransactions", () => {
        test("should prepare deposit transaction without approval when using ETH", async () => {
            // Setup for ETH (zero address)
            const deps = {
                ...mockDependencies,
                addresses: {
                    ...mockAddresses,
                    tokenAddress: zeroAddress,
                },
            };

            const ethPreparer = new NitroliteTransactionPreparer(deps);
            const amount = BigInt(100);

            const result = await ethPreparer.prepareDepositTransactions(amount);

            expect(mockNitroliteService.prepareDeposit).toHaveBeenCalledWith(zeroAddress, amount);
            expect(mockErc20Service.getTokenAllowance).not.toHaveBeenCalled();
            expect(mockErc20Service.prepareApprove).not.toHaveBeenCalled();
            expect(result).toHaveLength(1); // Only deposit, no approval
        });

        test("should prepare deposit transaction with approval when using ERC20 with insufficient allowance", async () => {
            const amount = BigInt(100);
            mockErc20Service.getTokenAllowance.mockResolvedValue(BigInt(50)); // Less than amount

            const result = await preparer.prepareDepositTransactions(amount);

            expect(mockErc20Service.getTokenAllowance).toHaveBeenCalledWith(mockAddresses.tokenAddress, mockAccount.address, mockAddresses.custody);
            expect(mockErc20Service.prepareApprove).toHaveBeenCalledWith(mockAddresses.tokenAddress, mockAddresses.custody, amount);
            expect(mockNitroliteService.prepareDeposit).toHaveBeenCalledWith(mockAddresses.tokenAddress, amount);
            expect(result).toHaveLength(2); // Approval + deposit
        });

        test("should prepare only deposit transaction when ERC20 allowance is sufficient", async () => {
            const amount = BigInt(100);
            mockErc20Service.getTokenAllowance.mockResolvedValue(BigInt(200)); // More than amount

            const result = await preparer.prepareDepositTransactions(amount);

            expect(mockErc20Service.getTokenAllowance).toHaveBeenCalled();
            expect(mockErc20Service.prepareApprove).not.toHaveBeenCalled();
            expect(mockNitroliteService.prepareDeposit).toHaveBeenCalled();
            expect(result).toHaveLength(1); // Only deposit
        });

        test("should handle approval error correctly", async () => {
            mockErc20Service.prepareApprove.mockRejectedValue(new Error("Approval failed"));

            await expect(preparer.prepareDepositTransactions(BigInt(100))).rejects.toThrow(Errors.ContractCallError);
        });

        test("should handle deposit error correctly", async () => {
            mockNitroliteService.prepareDeposit.mockRejectedValue(new Error("Deposit failed"));

            await expect(preparer.prepareDepositTransactions(BigInt(100))).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe("prepareCreateChannelTransaction", () => {
        const mockCreateParams: CreateChannelParams = {
            initialAllocationAmounts: [BigInt(100), BigInt(0)],
            stateData: "0x00" as Hex,
        };

        test("should prepare create channel transaction", async () => {
            // Mock state preparation (which we'll test separately)
            jest.spyOn(require("../../src/client/state"), "_prepareAndSignInitialState").mockImplementation(() =>
                Promise.resolve({
                    channel: {
                        participants: [mockAccount.address, mockAddresses.guestAddress],
                        adjudicator: mockAddresses.adjudicators.default,
                        challenge: BigInt(86400),
                        nonce: BigInt(1),
                    },
                    initialState: mockState,
                    channelId: mockChannelId,
                })
            );

            const result = await preparer.prepareCreateChannelTransaction(mockCreateParams);

            expect(mockNitroliteService.prepareCreateChannel).toHaveBeenCalled();
            expect(result).toEqual({ to: "0x123", data: "0x456" });
        });

        test("should handle errors in state preparation", async () => {
            jest.spyOn(require("../../src/client/state"), "_prepareAndSignInitialState").mockImplementation(() =>
                Promise.reject(new Error("State preparation failed"))
            );

            await expect(preparer.prepareCreateChannelTransaction(mockCreateParams)).rejects.toThrow(Error);
        });

        test("should handle contract call errors", async () => {
            jest.spyOn(require("../../src/client/state"), "_prepareAndSignInitialState").mockImplementation(() =>
                Promise.resolve({
                    channel: {
                        participants: [mockAccount.address, mockAddresses.guestAddress],
                        adjudicator: mockAddresses.adjudicators.default,
                        challenge: BigInt(86400),
                        nonce: BigInt(1),
                    },
                    initialState: mockState,
                    channelId: mockChannelId,
                })
            );

            mockNitroliteService.prepareCreateChannel.mockRejectedValue(
                new Errors.ContractCallError("prepareCreateChannel", new Error("Contract call failed"))
            );

            await expect(preparer.prepareCreateChannelTransaction(mockCreateParams)).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe("prepareDepositAndCreateChannelTransactions", () => {
        test("should prepare both deposit and create channel transactions", async () => {
            const depositAmount = BigInt(100);
            const createParams: CreateChannelParams = {
                initialAllocationAmounts: [BigInt(100), BigInt(0)],
            };

            // Mock inner methods
            jest.spyOn(preparer, "prepareDepositTransactions").mockImplementation(() =>
                Promise.resolve([{ to: "0xDeposit", data: "0xDepositData" }])
            );

            jest.spyOn(preparer, "prepareCreateChannelTransaction").mockImplementation(() =>
                Promise.resolve({ to: "0xCreate", data: "0xCreateData" })
            );

            const result = await preparer.prepareDepositAndCreateChannelTransactions(depositAmount, createParams);

            expect(preparer.prepareDepositTransactions).toHaveBeenCalledWith(depositAmount);
            expect(preparer.prepareCreateChannelTransaction).toHaveBeenCalledWith(createParams);
            expect(result).toHaveLength(2);
            expect(result[0]).toEqual({ to: "0xDeposit", data: "0xDepositData" });
            expect(result[1]).toEqual({ to: "0xCreate", data: "0xCreateData" });
        });

        test("should handle deposit error", async () => {
            jest.spyOn(preparer, "prepareDepositTransactions").mockImplementation(() => Promise.reject(new Error("Deposit failed")));

            await expect(
                preparer.prepareDepositAndCreateChannelTransactions(BigInt(100), { initialAllocationAmounts: [BigInt(100), BigInt(0)] })
            ).rejects.toThrow(Errors.ContractCallError);
        });

        test("should handle create channel error", async () => {
            jest.spyOn(preparer, "prepareDepositTransactions").mockImplementation(() =>
                Promise.resolve([{ to: "0xDeposit", data: "0xDepositData" }])
            );

            jest.spyOn(preparer, "prepareCreateChannelTransaction").mockImplementation(() => Promise.reject(new Error("Create failed")));

            await expect(
                preparer.prepareDepositAndCreateChannelTransactions(BigInt(100), { initialAllocationAmounts: [BigInt(100), BigInt(0)] })
            ).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe("prepareCheckpointChannelTransaction", () => {
        test("should prepare checkpoint transaction", async () => {
            const params: CheckpointChannelParams = {
                channelId: mockChannelId,
                candidateState: {
                    ...mockState,
                    sigs: [mockSignature, mockSignature], // Two signatures required
                },
            };

            const result = await preparer.prepareCheckpointChannelTransaction(params);

            expect(mockNitroliteService.prepareCheckpoint).toHaveBeenCalledWith(params.channelId, params.candidateState, []);
            expect(result).toEqual({ to: "0x123", data: "0x456" });
        });

        test("should throw if candidateState has insufficient signatures", async () => {
            const params: CheckpointChannelParams = {
                channelId: mockChannelId,
                candidateState: mockState, // Only one signature
            };

            await expect(preparer.prepareCheckpointChannelTransaction(params)).rejects.toThrow(Errors.InvalidParameterError);
        });
    });

    describe("prepareChallengeChannelTransaction", () => {
        test("should prepare challenge transaction", async () => {
            const params: ChallengeChannelParams = {
                channelId: mockChannelId,
                candidateState: mockState,
                proofStates: [],
            };

            const result = await preparer.prepareChallengeChannelTransaction(params);

            expect(mockNitroliteService.prepareChallenge).toHaveBeenCalledWith(params.channelId, params.candidateState, params.proofStates);
            expect(result).toEqual({ to: "0x123", data: "0x456" });
        });
    });

    describe("prepareCloseChannelTransaction", () => {
        const mockCloseParams: CloseChannelParams = {
            finalState: {
                channelId: mockChannelId,
                stateData: "0x00" as Hex,
                allocations: [
                    { destination: "0xParticipant1" as Address, token: "0xTokenAddress" as Address, amount: BigInt(80) },
                    { destination: "0xParticipant2" as Address, token: "0xTokenAddress" as Address, amount: BigInt(20) },
                ],
                serverSignature: [mockSignature],
            },
        };

        test("should prepare close channel transaction", async () => {
            // Mock state preparation
            jest.spyOn(require("../../src/client/state"), "_prepareAndSignFinalState").mockImplementation(() =>
                Promise.resolve({
                    finalStateWithSigs: {
                        ...mockState,
                        sigs: [mockSignature, mockSignature], // Both signatures
                    },
                    channelId: mockChannelId,
                })
            );

            const result = await preparer.prepareCloseChannelTransaction(mockCloseParams);

            expect(mockNitroliteService.prepareClose).toHaveBeenCalled();
            expect(result).toEqual({ to: "0x123", data: "0x456" });
        });
    });

    describe("prepareWithdrawalTransaction", () => {
        test("should prepare withdrawal transaction", async () => {
            const amount = BigInt(100);

            const result = await preparer.prepareWithdrawalTransaction(amount);

            expect(mockNitroliteService.prepareWithdraw).toHaveBeenCalledWith(mockAddresses.tokenAddress, amount);
            expect(result).toEqual({ to: "0x123", data: "0x456" });
        });
    });

    describe("prepareApproveTokensTransaction", () => {
        test("should prepare approve transaction", async () => {
            const amount = BigInt(100);

            const result = await preparer.prepareApproveTokensTransaction(amount);

            expect(mockErc20Service.prepareApprove).toHaveBeenCalledWith(mockAddresses.tokenAddress, mockAddresses.custody, amount);
            expect(result).toEqual({ to: "0x123", data: "0x456" });
        });

        test("should throw for ETH approval", async () => {
            // Setup for ETH (zero address)
            const deps = {
                ...mockDependencies,
                addresses: {
                    ...mockAddresses,
                    tokenAddress: zeroAddress,
                },
            };

            const ethPreparer = new NitroliteTransactionPreparer(deps);

            await expect(ethPreparer.prepareApproveTokensTransaction(BigInt(100))).rejects.toThrow(Errors.InvalidParameterError);
        });
    });
});
