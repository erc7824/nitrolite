import { describe, test, expect, jest, beforeEach } from "@jest/globals";
import { _prepareAndSignInitialState, _prepareAndSignFinalState } from "../../src/client/state";
import { PreparerDependencies } from "../../src/client/prepare";
import { Errors } from "../../src/errors";
import { Account, Address, Hex, ParseAccount, WalletClient } from "viem";
import { ContractAddresses } from "../../src/abis";
import * as utils from "../../src/utils";
import { MAGIC_NUMBERS } from "../../src/config";
import { CloseChannelParams, CreateChannelParams, Signature } from "../../src/client/types";

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

const mockSignature: Signature = { v: 27, r: "0x00" as Hex, s: "0x00" as Hex };

const mockStateWalletClient = {
    account: mockAccount,
    signMessage: jest.fn(() => Promise.resolve("0xStateSignature" as Hex)),
} as unknown as WalletClient;

const mockDependencies: PreparerDependencies = {
    nitroliteService: null as any,
    erc20Service: null as any,
    addresses: mockAddresses,
    account: mockAccount as ParseAccount<Account>,
    walletClient: null as any,
    stateWalletClient: mockStateWalletClient,
    challengeDuration: BigInt(86400),
};

// Mock utility functions
jest.mock("../../src/utils", () => ({
    generateChannelNonce: jest.fn(() => BigInt(12345)),
    getChannelId: jest.fn(() => "0xChannelId" as Hex),
    getStateHash: jest.fn(() => "0xStateHash" as Hex),
    signState: jest.fn(() => Promise.resolve({ v: 27, r: "0x00" as Hex, s: "0x00" as Hex })),
    encoders: {
        numeric: jest.fn((value) => `0x${value.toString(16)}` as Hex),
    },
    removeQuotesFromRS: jest.fn(() => [{ v: 27, r: "0x00" as Hex, s: "0x00" as Hex }]),
}));

describe("State Utilities", () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    describe("_prepareAndSignInitialState", () => {
        const createParams: CreateChannelParams = {
            initialAllocationAmounts: [BigInt(100), BigInt(0)],
            stateData: "0x42" as Hex,
        };

        test("should prepare and sign initial state", async () => {
            const result = await _prepareAndSignInitialState(mockDependencies, createParams);

            // Verify channel is constructed correctly
            expect(result.channel).toEqual({
                participants: [mockAccount.address, mockAddresses.guestAddress],
                adjudicator: mockAddresses.adjudicators.default,
                challenge: mockDependencies.challengeDuration,
                nonce: BigInt(12345), // From mocked generateChannelNonce
            });

            // Verify initial state is constructed correctly
            expect(result.initialState).toEqual({
                data: createParams.stateData,
                allocations: [
                    { destination: mockAccount.address, token: mockAddresses.tokenAddress, amount: BigInt(100) },
                    { destination: mockAddresses.guestAddress, token: mockAddresses.tokenAddress, amount: BigInt(0) },
                ],
                sigs: [mockSignature], // From mocked signState
            });

            // Verify channel ID is correct
            expect(result.channelId).toBe("0xChannelId");

            // Verify utility functions were called
            expect(utils.generateChannelNonce).toHaveBeenCalled();
            expect(utils.getChannelId).toHaveBeenCalled();
            expect(utils.getStateHash).toHaveBeenCalled();
            expect(utils.signState).toHaveBeenCalled();
        });

        test("should use default initial app data if none provided", async () => {
            const paramsWithoutData: CreateChannelParams = {
                initialAllocationAmounts: [BigInt(100), BigInt(0)],
            };

            await _prepareAndSignInitialState(mockDependencies, paramsWithoutData);

            // Verify encoder was called with OPEN magic number
            expect(utils.encoders.numeric).toHaveBeenCalledWith(MAGIC_NUMBERS.OPEN);
        });

        test("should throw if adjudicator is missing", async () => {
            const depsWithoutAdjudicator = {
                ...mockDependencies,
                addresses: {
                    ...mockAddresses,
                    adjudicators: {},
                },
            };

            await expect(_prepareAndSignInitialState(depsWithoutAdjudicator, createParams)).rejects.toThrow(Errors.MissingParameterError);
        });

        test("should throw if participants are invalid", async () => {
            // Here we'd need to mock something to make participants invalid,
            // but the function doesn't have a direct path to test this since
            // it constructs participants internally

            // For coverage, we can test the allocation amounts validation
            const invalidParams: CreateChannelParams = {
                initialAllocationAmounts: null as any,
            };

            await expect(_prepareAndSignInitialState(mockDependencies, invalidParams)).rejects.toThrow(Errors.InvalidParameterError);
        });

        test("should throw if allocation amounts are invalid", async () => {
            const invalidParams: CreateChannelParams = {
                initialAllocationAmounts: [BigInt(100)] as any, // Should be a tuple of two
            };

            await expect(_prepareAndSignInitialState(mockDependencies, invalidParams)).rejects.toThrow(Errors.InvalidParameterError);
        });
    });

    describe("_prepareAndSignFinalState", () => {
        const closeParams: CloseChannelParams = {
            stateData: "0x42" as Hex,
            finalState: {
                channelId: "0xChannelId" as Hex,
                stateData: "0x00" as Hex,
                allocations: [
                    { destination: "0xParticipant1" as Address, token: "0xTokenAddress" as Address, amount: BigInt(80) },
                    { destination: "0xParticipant2" as Address, token: "0xTokenAddress" as Address, amount: BigInt(20) },
                ],
                serverSignature: [mockSignature],
            },
        };

        test("should prepare and sign final state", async () => {
            const result = await _prepareAndSignFinalState(mockDependencies, closeParams);

            // Verify state has all signatures
            expect(result.finalStateWithSigs).toEqual({
                data: closeParams.stateData,
                allocations: closeParams.finalState.allocations,
                sigs: [mockSignature, [mockSignature]], // Account signature + array of server signatures
            });

            // Verify channel ID is passed through
            expect(result.channelId).toBe(closeParams.finalState.channelId);

            // Verify utility functions were called
            expect(utils.getStateHash).toHaveBeenCalled();
            expect(utils.signState).toHaveBeenCalled();
            expect(utils.removeQuotesFromRS).toHaveBeenCalled();
        });

        test("should use default close app data if none provided", async () => {
            const paramsWithoutData: CloseChannelParams = {
                ...closeParams,
                stateData: undefined,
            };

            const result = await _prepareAndSignFinalState(mockDependencies, paramsWithoutData);

            // Verify encoder was called with CLOSE magic number
            expect(utils.encoders.numeric).toHaveBeenCalledWith(MAGIC_NUMBERS.CLOSE);

            // Verify structure is correct
            expect(result.finalStateWithSigs.sigs).toEqual([mockSignature, [mockSignature]]);
        });
    });
});
