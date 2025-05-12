import { describe, test, expect, beforeEach, jest } from "@jest/globals";
import { Address, Hex, SimulateContractReturnType, zeroAddress } from "viem";
import { NitroliteService } from "../../../src/client/services/NitroliteService";
import { Errors } from "../../../src/errors";
import { CustodyAbi, ContractAddresses } from "../../../src/abis";
import { Channel, ChannelId, Signature, State } from "../../../src/client/types";

describe("NitroliteService", () => {
    const custodyAddress = "0x0000000000000000000000000000000000000001" as Address;
    const addresses: ContractAddresses = { custody: custodyAddress } as any;
    const account = "0x0000000000000000000000000000000000000002" as Address;
    const chaindId = 1;

    // Dummy data for channel methods
    const channelConfig = {} as Channel;
    const initialState = {} as State;
    const channelId = "0x1" as ChannelId;
    const participantIndex = 0n;
    const participantSig = "0xsig" as unknown as Signature;
    const candidateState = {} as State;
    const proofs = [{} as State];
    const newChannelConfig = {} as Channel;
    const newDepositState = {} as State;

    let mockPublicClient: any;
    let mockWalletClient: any;
    let service: NitroliteService;

    beforeEach(() => {
        mockPublicClient = {
            simulateContract: jest.fn(),
            readContract: jest.fn(),
        };
        mockWalletClient = {
            writeContract: jest.fn(),
            account,
        };
        service = new NitroliteService(mockPublicClient, addresses, mockWalletClient, account);
    });

    describe("constructor", () => {
        test("throws if publicClient missing", () => {
            expect(() => new NitroliteService(undefined as any, addresses)).toThrow(Errors.MissingParameterError);
        });
        test("throws if addresses.custody missing", () => {
            expect(() => new NitroliteService(mockPublicClient, {} as any, mockWalletClient, account)).toThrow(Errors.MissingParameterError);
        });
    });

    // Helper to generate a fake request object
    function fakeRequest(): SimulateContractReturnType["request"] {
        return { to: "0x", data: "0x" } as unknown as SimulateContractReturnType["request"];
    }

    // List of all prepare/execute pairs
    const methodDefs = [
        {
            prepareName: "prepareDeposit",
            execName: "deposit",
            prepare: () => service.prepareDeposit(custodyAddress, 123n),
            exec: () => service.deposit(custodyAddress, 123n),
            fn: "deposit",
            extra: (args: any) => ({ value: zeroAddress === args[0] ? args[1] : 0n }),
        },
        {
            prepareName: "prepareCreateChannel",
            execName: "createChannel",
            prepare: () => service.prepareCreateChannel(channelConfig, initialState),
            exec: () => service.createChannel(channelConfig, initialState),
            fn: "create",
        },
        {
            prepareName: "prepareJoinChannel",
            execName: "joinChannel",
            prepare: () => service.prepareJoinChannel(channelId, participantIndex, participantSig),
            exec: () => service.joinChannel(channelId, participantIndex, participantSig),
            fn: "join",
        },
        {
            prepareName: "prepareCheckpoint",
            execName: "checkpoint",
            prepare: () => service.prepareCheckpoint(channelId, candidateState, proofs),
            exec: () => service.checkpoint(channelId, candidateState, proofs),
            fn: "checkpoint",
        },
        {
            prepareName: "prepareChallenge",
            execName: "challenge",
            prepare: () => service.prepareChallenge(channelId, candidateState, proofs),
            exec: () => service.challenge(channelId, candidateState, proofs),
            fn: "challenge",
        },
        {
            prepareName: "prepareClose",
            execName: "close",
            prepare: () => service.prepareClose(channelId, candidateState, proofs),
            exec: () => service.close(channelId, candidateState, proofs),
            fn: "close",
        },
        {
            prepareName: "prepareResize",
            execName: "resize",
            prepare: () => service.prepareResize(channelId, candidateState, proofs),
            exec: () => service.resize(channelId, candidateState, proofs),
            fn: "resize",
        },
        {
            prepareName: "prepareWithdraw",
            execName: "withdraw",
            prepare: () => service.prepareWithdraw(custodyAddress, 456n),
            exec: () => service.withdraw(custodyAddress, 456n),
            fn: "withdraw",
        },
    ];

    for (const def of methodDefs) {
        describe(def.prepareName, () => {
            test("success", async () => {
                const req = fakeRequest();
                (mockPublicClient.simulateContract as any).mockResolvedValue({ request: req, result: {} });
                const res = await def.prepare();
                expect(mockPublicClient.simulateContract).toHaveBeenCalledWith(
                    expect.objectContaining({
                        address: custodyAddress,
                        abi: CustodyAbi,
                        functionName: def.fn,
                        args: expect.any(Array),
                        account,
                        ...(def.extra ? def.extra([custodyAddress, 123n]) : {}),
                    })
                );
                expect(res).toBe(req);
            });
            test("ContractCallError", async () => {
                (mockPublicClient.simulateContract as any).mockRejectedValue(new Error("fail"));
                await expect(def.prepare()).rejects.toThrow(Errors.ContractCallError);
            });
            test("rethrow NitroliteError", async () => {
                const ne = new Errors.MissingParameterError("x");
                (mockPublicClient.simulateContract as any).mockRejectedValue(ne);
                await expect(def.prepare()).rejects.toThrow(ne);
            });
        });

        describe(def.execName, () => {
            test("success", async () => {
                const req = fakeRequest();
                (mockPublicClient.simulateContract as any).mockResolvedValue({ request: req, result: {} });
                (mockWalletClient.writeContract as any).mockResolvedValue("0xhash");
                const hash = await def.exec();
                expect(mockWalletClient.writeContract).toHaveBeenCalledWith({ ...req, account });
                expect(hash).toBe("0xhash");
            });
            test("TransactionError", async () => {
                const req = fakeRequest();
                (mockPublicClient.simulateContract as any).mockResolvedValue({ request: req, result: {} });
                (mockWalletClient.writeContract as any).mockRejectedValue(new Error("oops"));
                await expect(def.exec()).rejects.toThrow(Errors.TransactionError);
            });
            test("rethrow NitroliteError", async () => {
                (mockPublicClient.simulateContract as any).mockResolvedValue({ request: {} as any, result: {} });
                const ne = new Errors.WalletClientRequiredError();
                (mockWalletClient.writeContract as any).mockRejectedValue(ne);
                await expect(def.exec()).rejects.toThrow(ne);
            });
        });
    }

    describe("getAccountChannels", () => {
        test("success", async () => {
            const arr = ["0xA", "0xB"];
            (mockPublicClient.readContract as any).mockResolvedValue(arr);
            const out = await service.getAccountChannels(account);
            expect(out).toEqual(arr);
            expect(mockPublicClient.readContract).toHaveBeenCalledWith({
                address: custodyAddress,
                abi: CustodyAbi,
                functionName: "getAccountChannels",
                args: [account],
            });
        });
        test("ContractReadError", async () => {
            (mockPublicClient.readContract as any).mockRejectedValue(new Error("fail"));
            await expect(service.getAccountChannels(account)).rejects.toThrow(Errors.ContractReadError);
        });
    });

    describe("getAccountInfo", () => {
        test("success", async () => {
            const data = [1n, 3n] as [bigint, bigint];
            (mockPublicClient.readContract as any).mockResolvedValue(data);
            const info = await service.getAccountInfo(account, custodyAddress);
            expect(info).toEqual({ available: 1n, channelCount: 3n });
            expect(mockPublicClient.readContract).toHaveBeenCalledWith({
                address: custodyAddress,
                abi: CustodyAbi,
                functionName: "getAccountInfo",
                args: [account, custodyAddress],
            });
        });
        test("ContractReadError", async () => {
            (mockPublicClient.readContract as any).mockRejectedValue(new Error("err"));
            await expect(service.getAccountInfo(account, custodyAddress)).rejects.toThrow(Errors.ContractReadError);
        });
    });
});
