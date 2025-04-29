import { jest } from "@jest/globals";
jest.mock("../../src/client/state", () => ({
    _prepareAndSignInitialState: jest.fn(),
    _prepareAndSignFinalState: jest.fn(),
}));
import { describe, test, expect, beforeEach } from "@jest/globals";
import { Hex, zeroAddress } from "viem";
import { NitroliteTransactionPreparer } from "../../src/client/prepare";
import { _prepareAndSignInitialState, _prepareAndSignFinalState } from "../../src/client/state";
import { NitroliteService, Erc20Service } from "../../src/client/services";
import { ContractAddresses } from "../../src/abis";
import * as Errors from "../../src/errors";
import { CreateChannelParams, CheckpointChannelParams, ChallengeChannelParams, CloseChannelParams, Allocation } from "../../src/client/types";

// TODO: remove ts-ignore
describe("NitroliteTransactionPreparer", () => {
    const tokenAddress = "0xTOK" as const;
    const custody = "0xCUST" as const;
    const accountAddress = "0xACC" as const;
    const guestAddress = "0xGUEST" as const;

    const adjudicator = "0xADD" as const;

    const addresses: ContractAddresses = {
        tokenAddress,
        custody,
        guestAddress,
        adjudicator,
    };
    const account = { address: accountAddress };

    let mockNitro: jest.Mocked<NitroliteService>;
    let mockERC20: jest.Mocked<Erc20Service>;
    let deps: any;
    let prep: NitroliteTransactionPreparer;

    beforeEach(() => {
        mockNitro = {
            prepareDeposit: jest.fn(),
            prepareCreateChannel: jest.fn(),
            prepareCheckpoint: jest.fn(),
            prepareChallenge: jest.fn(),
            prepareClose: jest.fn(),
            prepareWithdraw: jest.fn(),
        } as any;
        mockERC20 = {
            getTokenAllowance: jest.fn(),
            prepareApprove: jest.fn(),
        } as any;
        deps = {
            nitroliteService: mockNitro,
            erc20Service: mockERC20,
            addresses: addresses,
            account,
            walletClient: {},
            stateWalletClient: {},
            challengeDuration: 100n,
        };
        prep = new NitroliteTransactionPreparer(deps);
    });

    describe("prepareDepositTransactions", () => {
        test("ERC20 no approval needed", async () => {
            mockERC20.getTokenAllowance.mockResolvedValue(100n);
            mockNitro.prepareDeposit.mockResolvedValue({ to: "0xA", data: "0xA" } as any);
            const txs = await prep.prepareDepositTransactions(50n);
            expect(mockERC20.getTokenAllowance).toHaveBeenCalledWith(tokenAddress, accountAddress, custody);
            expect(txs).toHaveLength(1);
        });

        test("ERC20 needs approval", async () => {
            mockERC20.getTokenAllowance.mockResolvedValue(10n);
            mockERC20.prepareApprove.mockResolvedValue({ to: "0xA", data: "0xA" } as any);
            mockNitro.prepareDeposit.mockResolvedValue({ to: "0xD", data: "0xD" } as any);
            const txs = await prep.prepareDepositTransactions(50n);
            expect(mockERC20.prepareApprove).toHaveBeenCalledWith(tokenAddress, custody, 50n);
            expect(txs).toHaveLength(2);
        });

        test("skip approval for ETH", async () => {
            deps.addresses.tokenAddress = zeroAddress;
            mockNitro.prepareDeposit.mockResolvedValue({ to: "0xD", data: "0xD" } as any);
            const txs = await prep.prepareDepositTransactions(20n);
            expect(mockERC20.getTokenAllowance).not.toHaveBeenCalled();
            expect(txs).toHaveLength(1);
        });

        test("prepareDeposit error wraps", async () => {
            mockERC20.getTokenAllowance.mockResolvedValue(100n);
            mockNitro.prepareDeposit.mockRejectedValue(new Error("fail"));
            await expect(prep.prepareDepositTransactions(10n)).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe("prepareCreateChannelTransaction", () => {
        const params: CreateChannelParams = { initialAllocationAmounts: [1n, 2n], stateData: "0xA" as any };
        test("success", async () => {
            const fake = { channel: {}, initialState: {} };
            // @ts-ignore
            (_prepareAndSignInitialState as jest.Mock).mockResolvedValue(fake);
            mockNitro.prepareCreateChannel.mockResolvedValue({ to: "0xC", data: "0xC" } as any);
            const tx = await prep.prepareCreateChannelTransaction(params);
            expect(_prepareAndSignInitialState).toHaveBeenCalledWith(deps, params);
            expect(tx).toEqual({ to: "0xC", data: "0xC" });
        });

        test("wraps non-NitroliteError", async () => {
            // @ts-ignore
            (_prepareAndSignInitialState as jest.Mock).mockRejectedValue(new Error("oops"));
            await expect(prep.prepareCreateChannelTransaction(params)).rejects.toThrow(Errors.ContractCallError);
        });

        test("rethrows NitroliteError from state", async () => {
            const nitroliteError = new Errors.MissingParameterError("x");
            // @ts-ignore
            (_prepareAndSignInitialState as jest.Mock).mockRejectedValueOnce(nitroliteError);
            await expect(prep.prepareCreateChannelTransaction(params)).rejects.toBe(nitroliteError);
        });
    });

    describe("prepareDepositAndCreateChannelTransactions", () => {
        test("combines flows", async () => {
            mockERC20.getTokenAllowance.mockResolvedValue(100n);
            mockNitro.prepareDeposit.mockResolvedValue({ to: "0xD", data: "0xD" } as any);
            // @ts-ignore
            (_prepareAndSignInitialState as jest.Mock).mockResolvedValue({ channel: {}, initialState: {} });
            mockNitro.prepareCreateChannel.mockResolvedValue({ to: "0xC", data: "0xC" } as any);
            const all = await prep.prepareDepositAndCreateChannelTransactions(10n, {} as any);
            expect(all).toHaveLength(2);
        });

        test("rethrows NitroliteError from deposit prepare", async () => {
            const ne = new Errors.MissingParameterError("d");
            mockERC20.getTokenAllowance.mockResolvedValue(100n);
            mockNitro.prepareDeposit.mockRejectedValueOnce(ne);
            await expect(prep.prepareDepositAndCreateChannelTransactions(10n, {} as any)).rejects.toThrow(Errors.ContractCallError);
        });

        test("rethrows NitroliteError from createChannel prepare", async () => {
            mockERC20.getTokenAllowance.mockResolvedValue(100n);
            mockNitro.prepareDeposit.mockResolvedValue({ to: "0xD", data: "0xD" } as any);
            // @ts-ignore
            (_prepareAndSignInitialState as jest.Mock).mockResolvedValue({ channel: {}, initialState: {} });
            const ne = new Errors.MissingParameterError("y");
            mockNitro.prepareCreateChannel.mockRejectedValueOnce(ne);
            await expect(prep.prepareDepositAndCreateChannelTransactions(10n, {} as any)).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe("prepareCheckpointChannelTransaction", () => {
        const good: CheckpointChannelParams = { channelId: "0xcid" as any, candidateState: { sigs: [1, 2] } as any, proofStates: [] };
        const bad: CheckpointChannelParams = { channelId: "0xcid" as any, candidateState: { sigs: [1] } as any };
        test("valid", async () => {
            mockNitro.prepareCheckpoint.mockResolvedValue({ to: "0xK", data: "0xK" } as any);
            await expect(prep.prepareCheckpointChannelTransaction(good)).resolves.toEqual({ to: "0xK", data: "0xK" } as any);
        });
        test("invalid sigs", async () => {
            await expect(prep.prepareCheckpointChannelTransaction(bad)).rejects.toThrow(Errors.InvalidParameterError);
        });
    });

    describe("prepareChallengeChannelTransaction", () => {
        test("success and wrap", async () => {
            mockNitro.prepareChallenge.mockResolvedValue({ to: "0xH", data: "0xH" } as any);
            await expect(
                // @ts-ignore
                prep.prepareChallengeChannelTransaction({ channelId: "0xcid" as any, candidateState: {}, proofStates: [] })
            ).resolves.toBeDefined();
        });
    });

    describe("prepareCloseChannelTransaction", () => {
        test("success", async () => {
            // @ts-ignore
            (_prepareAndSignFinalState as jest.Mock).mockResolvedValue({ finalStateWithSigs: {}, channelId: "0xcid" });
            mockNitro.prepareClose.mockResolvedValue({ to: "0xX", data: "0xX" } as any);
            await expect(
                prep.prepareCloseChannelTransaction({
                    finalState: {
                        stateData: "0xA" as any,
                        channelId: "0xcid" as any,
                        allocations: [
                            { destination: "0x0" as Hex, token: tokenAddress, amount: 10n },
                            { destination: "0x0" as Hex, token: tokenAddress, amount: 10n },
                        ] as [Allocation, Allocation],
                        version: 0n,
                        serverSignature: [] as any,
                    },
                })
            ).resolves.toEqual({ to: "0xX", data: "0xX" });
        });

        test("wraps final state error", async () => {
            // @ts-ignore
            (_prepareAndSignFinalState as jest.Mock).mockRejectedValueOnce(new Error("state fail"));
            await expect(
                prep.prepareCloseChannelTransaction({
                    finalState: {
                        stateData: "0xA" as any,
                        channelId: "0xcid" as any,
                        allocations: [
                            { destination: "0x0" as Hex, token: tokenAddress, amount: 10n },
                            { destination: "0x0" as Hex, token: tokenAddress, amount: 10n },
                        ] as [Allocation, Allocation],
                        version: 0n,
                        serverSignature: [] as any,
                    },
                })
            ).rejects.toThrow(Errors.ContractCallError);
        });

        test("wraps non-NitroliteError from close", async () => {
            // @ts-ignore
            (_prepareAndSignFinalState as jest.Mock).mockResolvedValue({ finalStateWithSigs: {}, channelId: "0xcid" });
            const err = new Error("oops");
            mockNitro.prepareClose.mockRejectedValueOnce(err);
            await expect(
                prep.prepareCloseChannelTransaction({
                    finalState: {
                        stateData: "0xA" as any,
                        channelId: "0xcid" as any,
                        allocations: [
                            { destination: "0x0" as Hex, token: tokenAddress, amount: 10n },
                            { destination: "0x0" as Hex, token: tokenAddress, amount: 10n },
                        ] as [Allocation, Allocation],
                        version: 0n,
                        serverSignature: [] as any,
                    },
                })
            ).rejects.toThrow(Errors.ContractCallError);
        });

        test("rethrows NitroliteError from close", async () => {
            // @ts-ignore
            (_prepareAndSignFinalState as jest.Mock).mockResolvedValue({ finalStateWithSigs: {}, channelId: "0xcid" });
            const ne = new Errors.MissingParameterError("z");
            mockNitro.prepareClose.mockRejectedValueOnce(ne);
            await expect(
                prep.prepareCloseChannelTransaction({
                    finalState: {
                        stateData: "0xA" as any,
                        channelId: "0xcid" as any,
                        allocations: [
                            { destination: "0x0" as Hex, token: tokenAddress, amount: 10n },
                            { destination: "0x0" as Hex, token: tokenAddress, amount: 10n },
                        ] as [Allocation, Allocation],
                        version: 0n,
                        serverSignature: [] as any,
                    },
                })
            ).rejects.toBe(ne);
        });
    });

    describe("prepareWithdrawalTransaction", () => {
        test("success", async () => {
            mockNitro.prepareWithdraw.mockResolvedValue({ to: "0xW", data: "0xW" } as any);
            await expect(prep.prepareWithdrawalTransaction(5n)).resolves.toEqual({ to: "0xW", data: "0xW" } as any);
        });
    });

    describe("prepareApproveTokensTransaction", () => {
        test("ETH error", async () => {
            deps.addresses.tokenAddress = zeroAddress;
            await expect(prep.prepareApproveTokensTransaction(1n)).rejects.toThrow(Errors.InvalidParameterError);
        });
        test("success", async () => {
            deps.addresses.tokenAddress = tokenAddress;
            mockERC20.prepareApprove.mockResolvedValue({ to: "0xA", data: "0xA" } as any);
            await expect(prep.prepareApproveTokensTransaction(7n)).resolves.toEqual({ to: "0xA", data: "0xA" } as any);
        });
        test("rethrows NitroliteError from prepareApproveTokensTransaction", async () => {
            deps.addresses.tokenAddress = tokenAddress;
            const ne = new Errors.MissingParameterError("a");
            mockERC20.prepareApprove.mockRejectedValueOnce(ne);
            await expect(prep.prepareApproveTokensTransaction(7n)).rejects.toBe(ne);
        });
    });
});
