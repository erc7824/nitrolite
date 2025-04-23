import { describe, test, expect, jest, beforeEach } from "@jest/globals";
import { NitroliteClient } from "../../src/client/index";
import { Errors } from "../../src/errors";
import { Account, Address, Hash, Hex, ParseAccount, PublicClient, WalletClient, zeroAddress } from "viem";
import { ContractAddresses } from "../../src/abis";
import {
    AccountInfo,
    ChallengeChannelParams,
    ChannelId,
    CheckpointChannelParams,
    CloseChannelParams,
    CreateChannelParams,
    Signature,
    State,
} from "../../src/client/types";
import * as stateModule from "../../src/client/state";
import { NitroliteTransactionPreparer } from "../../src/client/prepare";

// Mock data
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
    writeContract: jest.fn(() => Promise.resolve("0xTransactionHash" as Hash)),
} as unknown as WalletClient;

const mockPublicClient = {
    readContract: jest.fn(),
    simulateContract: jest.fn(),
} as unknown as PublicClient;

const mockChannelId = "0xChannelId" as ChannelId;
const mockSignature: Signature = { v: 27, r: "0x00" as Hex, s: "0x00" as Hex };

const mockState: State = {
    data: "0x00" as Hex,
    allocations: [
        { destination: "0xParticipant1" as Address, token: "0xTokenAddress" as Address, amount: BigInt(100) },
        { destination: "0xParticipant2" as Address, token: "0xTokenAddress" as Address, amount: BigInt(0) },
    ],
    sigs: [mockSignature],
};

// Mock services
jest.mock("../../src/client/services/NitroliteService", () => {
    return {
        NitroliteService: jest.fn().mockImplementation(() => ({
            deposit: jest.fn(() => Promise.resolve("0xDepositHash" as Hash)),
            createChannel: jest.fn(() => Promise.resolve("0xCreateHash" as Hash)),
            checkpoint: jest.fn(() => Promise.resolve("0xCheckpointHash" as Hash)),
            challenge: jest.fn(() => Promise.resolve("0xChallengeHash" as Hash)),
            close: jest.fn(() => Promise.resolve("0xCloseHash" as Hash)),
            withdraw: jest.fn(() => Promise.resolve("0xWithdrawHash" as Hash)),
            getAccountChannels: jest.fn(() => Promise.resolve(["0xChannel1", "0xChannel2"])),
            getAccountInfo: jest.fn(() =>
                Promise.resolve({
                    available: BigInt(100),
                    locked: BigInt(200),
                    channelCount: BigInt(2),
                })
            ),
        })),
    };
});

jest.mock("../../src/client/services/Erc20Service", () => {
    return {
        Erc20Service: jest.fn().mockImplementation(() => ({
            getTokenBalance: jest.fn(() => Promise.resolve(BigInt(1000))),
            getTokenAllowance: jest.fn(() => Promise.resolve(BigInt(500))),
            approve: jest.fn(() => Promise.resolve("0xApproveHash" as Hash)),
        })),
    };
});

// Mock state utilities
jest.mock("../../src/client/state", () => ({
    _prepareAndSignInitialState: jest.fn(),
    _prepareAndSignFinalState: jest.fn(),
}));

describe("NitroliteClient", () => {
    let client: NitroliteClient;

    beforeEach(() => {
        jest.clearAllMocks();

        // Setup mock implementations for state utilities
        (stateModule._prepareAndSignInitialState as jest.Mock).mockResolvedValue({
            channel: {
                participants: [mockAccount.address, mockAddresses.guestAddress],
                adjudicator: mockAddresses.adjudicators.default,
                challenge: BigInt(86400),
                nonce: BigInt(1),
            },
            initialState: mockState,
            channelId: mockChannelId,
        });

        (stateModule._prepareAndSignFinalState as jest.Mock).mockResolvedValue({
            finalStateWithSigs: {
                ...mockState,
                sigs: [mockSignature, mockSignature], // Both signatures
            },
            channelId: mockChannelId,
        });

        // Initialize the client
        client = new NitroliteClient({
            publicClient: mockPublicClient,
            walletClient: mockWalletClient,
            addresses: mockAddresses,
            challengeDuration: BigInt(86400),
        });
    });

    describe("constructor", () => {
        test("should throw when publicClient is missing", () => {
            expect(
                () =>
                    new NitroliteClient({
                        publicClient: null as unknown as PublicClient,
                        walletClient: mockWalletClient,
                        addresses: mockAddresses,
                    })
            ).toThrow(Errors.MissingParameterError);
        });

        test("should throw when walletClient is missing", () => {
            expect(
                () =>
                    new NitroliteClient({
                        publicClient: mockPublicClient,
                        walletClient: null as unknown as WalletClient,
                        addresses: mockAddresses,
                    })
            ).toThrow(Errors.MissingParameterError);
        });

        test("should throw when walletClient.account is missing", () => {
            const walletWithoutAccount = { ...mockWalletClient, account: undefined };
            expect(
                () =>
                    new NitroliteClient({
                        publicClient: mockPublicClient,
                        walletClient: walletWithoutAccount as unknown as WalletClient,
                        addresses: mockAddresses,
                    })
            ).toThrow(Errors.MissingParameterError);
        });

        test("should throw when challengeDuration is missing", () => {
            expect(
                () =>
                    new NitroliteClient({
                        publicClient: mockPublicClient,
                        walletClient: mockWalletClient,
                        addresses: mockAddresses,
                        challengeDuration: undefined,
                    })
            ).toThrow(Errors.MissingParameterError);
        });

        test("should throw when addresses.custody is missing", () => {
            expect(
                () =>
                    new NitroliteClient({
                        publicClient: mockPublicClient,
                        walletClient: mockWalletClient,
                        addresses: { ...mockAddresses, custody: undefined as unknown as Address },
                        challengeDuration: BigInt(86400),
                    })
            ).toThrow(Errors.MissingParameterError);
        });

        test("should throw when addresses.adjudicators is missing", () => {
            expect(
                () =>
                    new NitroliteClient({
                        publicClient: mockPublicClient,
                        walletClient: mockWalletClient,
                        addresses: { ...mockAddresses, adjudicators: undefined as any },
                        challengeDuration: BigInt(86400),
                    })
            ).toThrow(Errors.MissingParameterError);
        });

        test("should throw when addresses.guestAddress is missing", () => {
            expect(
                () =>
                    new NitroliteClient({
                        publicClient: mockPublicClient,
                        walletClient: mockWalletClient,
                        addresses: { ...mockAddresses, guestAddress: undefined as unknown as Address },
                        challengeDuration: BigInt(86400),
                    })
            ).toThrow(Errors.MissingParameterError);
        });

        test("should throw when addresses.tokenAddress is missing", () => {
            expect(
                () =>
                    new NitroliteClient({
                        publicClient: mockPublicClient,
                        walletClient: mockWalletClient,
                        addresses: { ...mockAddresses, tokenAddress: undefined as unknown as Address },
                        challengeDuration: BigInt(86400),
                    })
            ).toThrow(Errors.MissingParameterError);
        });

        test("should use walletClient as stateWalletClient if not provided", () => {
            const clientWithoutStateWallet = new NitroliteClient({
                publicClient: mockPublicClient,
                walletClient: mockWalletClient,
                addresses: mockAddresses,
                challengeDuration: BigInt(86400),
            });

            // @ts-ignore - Access private property for testing
            expect(clientWithoutStateWallet.stateWalletClient).toBe(mockWalletClient);
        });

        test("should use provided stateWalletClient if available", () => {
            const mockStateWalletClient = { ...mockWalletClient, id: "stateWallet" };

            const clientWithStateWallet = new NitroliteClient({
                publicClient: mockPublicClient,
                walletClient: mockWalletClient,
                stateWalletClient: mockStateWalletClient as unknown as WalletClient,
                addresses: mockAddresses,
                challengeDuration: BigInt(86400),
            });

            // @ts-ignore - Access private property for testing
            expect(clientWithStateWallet.stateWalletClient).toBe(mockStateWalletClient);
        });

        test("should initialize txPreparer", () => {
            // Verify that txPreparer is created
            expect(client.txPreparer).toBeInstanceOf(NitroliteTransactionPreparer);
        });
    });

    describe("deposit", () => {
        test("should handle ERC20 deposits", async () => {
            const amount = BigInt(100);

            await expect(client.deposit(amount)).resolves.toBe("0xDepositHash");

            // Nitrolite service deposit should be called
            // @ts-ignore - Access nitroliteService
            expect(client.nitroliteService.deposit).toHaveBeenCalledWith(mockAddresses.tokenAddress, amount);
        });

        test("should handle service errors", async () => {
            // @ts-ignore - Access nitroliteService and mock it to throw
            client.nitroliteService.deposit.mockRejectedValueOnce(new Error("Service error"));

            await expect(client.deposit(BigInt(100))).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe("createChannel", () => {
        test("should prepare state and create channel", async () => {
            const params: CreateChannelParams = {
                initialAllocationAmounts: [BigInt(100), BigInt(0)],
                stateData: "0x00" as Hex,
            };

            const result = await client.createChannel(params);

            // Should call prepareAndSignInitialState
            expect(stateModule._prepareAndSignInitialState).toHaveBeenCalledWith(expect.anything(), params);

            // Should call nitroliteService.createChannel
            // @ts-ignore - Access nitroliteService
            expect(client.nitroliteService.createChannel).toHaveBeenCalled();

            // Should return correct result
            expect(result).toEqual({
                channelId: mockChannelId,
                initialState: mockState,
                txHash: "0xCreateHash",
            });
        });

        test("should handle service errors", async () => {
            // @ts-ignore - Access nitroliteService and mock it to throw
            client.nitroliteService.createChannel.mockRejectedValueOnce(new Error("Service error"));

            await expect(
                client.createChannel({
                    initialAllocationAmounts: [BigInt(100), BigInt(0)],
                })
            ).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe("depositAndCreateChannel", () => {
        test("should deposit and create channel", async () => {
            const depositAmount = BigInt(100);
            const params: CreateChannelParams = {
                initialAllocationAmounts: [BigInt(100), BigInt(0)],
            };

            // Mock the implementations for testing
            jest.spyOn(client, "deposit").mockResolvedValueOnce("0xDepositHash" as Hash);
            jest.spyOn(client, "createChannel").mockResolvedValueOnce({
                channelId: mockChannelId,
                initialState: mockState,
                txHash: "0xCreateHash" as Hash,
            });

            const result = await client.depositAndCreateChannel(depositAmount, params);

            // Should call deposit
            expect(client.deposit).toHaveBeenCalledWith(depositAmount);

            // Should call createChannel
            expect(client.createChannel).toHaveBeenCalledWith(params);

            // Should return correct result
            expect(result).toEqual({
                channelId: mockChannelId,
                initialState: mockState,
                depositTxHash: "0xDepositHash",
                createChannelTxHash: "0xCreateHash",
            });
        });
    });

    describe("checkpointChannel", () => {
        test("should checkpoint channel", async () => {
            const params: CheckpointChannelParams = {
                channelId: mockChannelId,
                candidateState: {
                    ...mockState,
                    sigs: [mockSignature, mockSignature], // Two signatures required
                },
                proofStates: [],
            };

            await expect(client.checkpointChannel(params)).resolves.toBe("0xCheckpointHash");

            // @ts-ignore - Access nitroliteService
            expect(client.nitroliteService.checkpoint).toHaveBeenCalledWith(params.channelId, params.candidateState, params.proofStates);
        });

        test("should throw if candidateState has insufficient signatures", async () => {
            const params: CheckpointChannelParams = {
                channelId: mockChannelId,
                candidateState: mockState, // Only one signature
            };

            await expect(client.checkpointChannel(params)).rejects.toThrow(Errors.InvalidParameterError);
        });
    });

    describe("challengeChannel", () => {
        test("should challenge channel", async () => {
            const params: ChallengeChannelParams = {
                channelId: mockChannelId,
                candidateState: mockState,
                proofStates: [],
            };

            await expect(client.challengeChannel(params)).resolves.toBe("0xChallengeHash");

            // @ts-ignore - Access nitroliteService
            expect(client.nitroliteService.challenge).toHaveBeenCalledWith(params.channelId, params.candidateState, params.proofStates);
        });
    });

    describe("closeChannel", () => {
        test("should prepare final state and close channel", async () => {
            const params: CloseChannelParams = {
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

            await expect(client.closeChannel(params)).resolves.toBe("0xCloseHash");

            // Should call prepareAndSignFinalState
            expect(stateModule._prepareAndSignFinalState).toHaveBeenCalledWith(expect.anything(), params);

            // @ts-ignore - Access nitroliteService
            expect(client.nitroliteService.close).toHaveBeenCalled();
        });
    });

    describe("withdrawal", () => {
        test("should withdraw funds", async () => {
            const amount = BigInt(100);

            await expect(client.withdrawal(amount)).resolves.toBe("0xWithdrawHash");

            // @ts-ignore - Access nitroliteService
            expect(client.nitroliteService.withdraw).toHaveBeenCalledWith(mockAddresses.tokenAddress, amount);
        });
    });

    describe("getAccountChannels", () => {
        test("should get account channels", async () => {
            await expect(client.getAccountChannels()).resolves.toEqual(["0xChannel1", "0xChannel2"]);

            // @ts-ignore - Access nitroliteService
            expect(client.nitroliteService.getAccountChannels).toHaveBeenCalledWith(mockAccount.address);
        });
    });

    describe("getAccountInfo", () => {
        test("should get account info", async () => {
            const expectedInfo: AccountInfo = {
                available: BigInt(100),
                locked: BigInt(200),
                channelCount: BigInt(2),
            };

            await expect(client.getAccountInfo()).resolves.toEqual(expectedInfo);

            // @ts-ignore - Access nitroliteService
            expect(client.nitroliteService.getAccountInfo).toHaveBeenCalledWith(mockAccount.address, mockAddresses.tokenAddress);
        });
    });

    describe("approveTokens", () => {
        test("should approve tokens", async () => {
            const amount = BigInt(100);

            await expect(client.approveTokens(amount)).resolves.toBe("0xApproveHash");

            // @ts-ignore - Access erc20Service
            expect(client.erc20Service.approve).toHaveBeenCalledWith(mockAddresses.tokenAddress, mockAddresses.custody, amount);
        });
    });

    describe("getTokenAllowance", () => {
        test("should get token allowance", async () => {
            await expect(client.getTokenAllowance()).resolves.toBe(BigInt(500));

            // @ts-ignore - Access erc20Service
            expect(client.erc20Service.getTokenAllowance).toHaveBeenCalledWith(
                mockAddresses.tokenAddress,
                mockAccount.address,
                mockAddresses.custody
            );
        });
    });

    describe("getTokenBalance", () => {
        test("should get token balance", async () => {
            await expect(client.getTokenBalance()).resolves.toBe(BigInt(1000));

            // @ts-ignore - Access erc20Service
            expect(client.erc20Service.getTokenBalance).toHaveBeenCalledWith(mockAddresses.tokenAddress, mockAccount.address);
        });
    });
});
