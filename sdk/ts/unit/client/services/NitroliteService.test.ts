import { describe, test, expect, beforeEach, jest } from '@jest/globals';
import { Address, Hex, SimulateContractReturnType, zeroAddress } from 'viem';
import { NitroliteService } from '../../../../src/client/services/NitroliteService';
import { Errors } from '../../../../src/errors';
import { CustodyAbi, ContractAddresses } from '../../../../src/abis';
import { Channel, ChannelId, Signature, State } from '../../../../src/client/types';

describe('NitroliteService', () => {
    const custodyAddress = '0x0000000000000000000000000000000000000001' as Address;
    const addresses: ContractAddresses = { custody: custodyAddress } as any;
    const account = '0x0000000000000000000000000000000000000002' as Address;
    const chaindId = 1;

    // Dummy data for channel methods
    const channelConfig = {} as Channel;
    const initialState = {} as State;
    const channelId = '0x1' as ChannelId;
    const participantIndex = 0n;
    const participantSig = '0xsig' as unknown as Signature;
    const candidateState = {} as State;
    const proofs = [{} as State];
    const challengerSig = '0xchallengerSig' as unknown as Signature;
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

    describe('constructor', () => {
        test('throws if publicClient missing', () => {
            expect(() => new NitroliteService(undefined as any, addresses)).toThrow(Errors.MissingParameterError);
        });
        test('throws if addresses.custody missing', () => {
            expect(() => new NitroliteService(mockPublicClient, {} as any, mockWalletClient, account)).toThrow(
                Errors.MissingParameterError,
            );
        });
    });

    // Helper to generate a fake request object
    function fakeRequest(): SimulateContractReturnType['request'] {
        return { to: '0x', data: '0x' } as unknown as SimulateContractReturnType['request'];
    }

    // List of all prepare/execute pairs
    const methodDefs = [
        {
            prepareName: 'prepareDeposit',
            execName: 'deposit',
            prepare: () => service.prepareDeposit(custodyAddress, 123n),
            exec: () => service.deposit(custodyAddress, 123n),
            fn: 'deposit',
            extra: (args: any) => ({ value: zeroAddress === args[0] ? args[1] : 0n }),
        },
        {
            prepareName: 'prepareCreateChannel',
            execName: 'createChannel',
            prepare: () => service.prepareCreateChannel(channelConfig, initialState),
            exec: () => service.createChannel(channelConfig, initialState),
            fn: 'create',
        },
        {
            prepareName: 'prepareDepositAndCreateChannel',
            execName: 'depositAndCreateChannel',
            prepare: () => service.prepareDepositAndCreateChannel(custodyAddress, 123n, channelConfig, initialState),
            exec: () => service.depositAndCreateChannel(custodyAddress, 123n, channelConfig, initialState),
            fn: 'depositAndCreate',
        },
        {
            prepareName: 'prepareJoinChannel',
            execName: 'joinChannel',
            prepare: () => service.prepareJoinChannel(channelId, participantIndex, participantSig),
            exec: () => service.joinChannel(channelId, participantIndex, participantSig),
            fn: 'join',
        },
        {
            prepareName: 'prepareCheckpoint',
            execName: 'checkpoint',
            prepare: () => service.prepareCheckpoint(channelId, candidateState, proofs),
            exec: () => service.checkpoint(channelId, candidateState, proofs),
            fn: 'checkpoint',
        },
        {
            prepareName: 'prepareChallenge',
            execName: 'challenge',
            prepare: () => service.prepareChallenge(channelId, candidateState, proofs, challengerSig),
            exec: () => service.challenge(channelId, candidateState, proofs, challengerSig),
            fn: 'challenge',
        },
        {
            prepareName: 'prepareClose',
            execName: 'close',
            prepare: () => service.prepareClose(channelId, candidateState, proofs),
            exec: () => service.close(channelId, candidateState, proofs),
            fn: 'close',
        },
        {
            prepareName: 'prepareResize',
            execName: 'resize',
            prepare: () => service.prepareResize(channelId, candidateState, proofs),
            exec: () => service.resize(channelId, candidateState, proofs),
            fn: 'resize',
        },
        {
            prepareName: 'prepareWithdraw',
            execName: 'withdraw',
            prepare: () => service.prepareWithdraw(custodyAddress, 456n),
            exec: () => service.withdraw(custodyAddress, 456n),
            fn: 'withdraw',
        },
    ];

    for (const def of methodDefs) {
        describe(def.prepareName, () => {
            test('success', async () => {
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
                    }),
                );
                expect(res).toBe(req);
            });
            test('ContractCallError', async () => {
                (mockPublicClient.simulateContract as any).mockRejectedValue(new Error('fail'));
                await expect(def.prepare()).rejects.toThrow(Errors.ContractCallError);
            });
            test('rethrow NitroliteError', async () => {
                const ne = new Errors.MissingParameterError('x');
                (mockPublicClient.simulateContract as any).mockRejectedValue(ne);
                await expect(def.prepare()).rejects.toThrow(ne);
            });
        });

        describe(def.execName, () => {
            test('success', async () => {
                const req = fakeRequest();
                (mockPublicClient.simulateContract as any).mockResolvedValue({ request: req, result: {} });
                (mockWalletClient.writeContract as any).mockResolvedValue('0xhash');
                const hash = await def.exec();
                expect(mockWalletClient.writeContract).toHaveBeenCalledWith({ ...req, account });
                expect(hash).toBe('0xhash');
            });
            test('TransactionError', async () => {
                const req = fakeRequest();
                (mockPublicClient.simulateContract as any).mockResolvedValue({ request: req, result: {} });
                (mockWalletClient.writeContract as any).mockRejectedValue(new Error('oops'));
                await expect(def.exec()).rejects.toThrow(Errors.TransactionError);
            });
            test('rethrow NitroliteError', async () => {
                (mockPublicClient.simulateContract as any).mockResolvedValue({ request: {} as any, result: {} });
                const ne = new Errors.WalletClientRequiredError();
                (mockWalletClient.writeContract as any).mockRejectedValue(ne);
                await expect(def.exec()).rejects.toThrow(ne);
            });
        });
    }

    describe('getOpenChannels', () => {
        test('success', async () => {
            const arr = ['0xA', '0xB'];
            (mockPublicClient.readContract as any).mockResolvedValue([arr]);
            const out = await service.getOpenChannels(account);
            expect(out).toEqual(arr);
            expect(mockPublicClient.readContract).toHaveBeenCalledWith({
                address: custodyAddress,
                abi: CustodyAbi,
                functionName: 'getOpenChannels',
                args: [[account]],
            });
        });
        test('ContractReadError', async () => {
            (mockPublicClient.readContract as any).mockRejectedValue(new Error('fail'));
            await expect(service.getOpenChannels(account)).rejects.toThrow(Errors.ContractReadError);
        });
    });

    describe('getAccountBalance', () => {
        const tokens = ['0xtok1', '0xtok2'] as Address[];
        test('success', async () => {
            const data = [[1n, 3n]] as bigint[][];
            (mockPublicClient.readContract as any).mockResolvedValue(data);
            const info = await service.getAccountBalance(account, tokens);
            expect(info).toEqual([1n, 3n]);
            expect(mockPublicClient.readContract).toHaveBeenCalledWith({
                address: custodyAddress,
                abi: CustodyAbi,
                functionName: 'getAccountsBalances',
                args: [[account], tokens],
            });
        });
        test('ContractReadError', async () => {
            (mockPublicClient.readContract as any).mockRejectedValue(new Error('err'));
            await expect(service.getAccountBalance(account, tokens)).rejects.toThrow(Errors.ContractReadError);
        });
    });

    describe('getChannelBalance', () => {
        const tokens = ['0xtok1', '0xtok2'] as Address[];
        test('success', async () => {
            const data = [[1n, 3n]] as bigint[][];
            (mockPublicClient.readContract as any).mockResolvedValue(data);
            const info = await service.getChannelBalance(channelId, tokens);
            expect(info).toEqual(data);
            expect(mockPublicClient.readContract).toHaveBeenCalledWith({
                address: custodyAddress,
                abi: CustodyAbi,
                functionName: 'getChannelBalances',
                args: [channelId, tokens],
            });
        });
        test('ContractReadError', async () => {
            (mockPublicClient.readContract as any).mockRejectedValue(new Error('err'));
            await expect(service.getChannelBalance(channelId, tokens)).rejects.toThrow(Errors.ContractReadError);
        });
    });

    describe('getChannelData', () => {
        test('success', async () => {
            // Mock contract returns a tuple array matching the actual contract structure
            const contractResult = [
                // result[0] - channel
                {
                    participants: ['0xabc', '0xdef'] as readonly Address[],
                    adjudicator: '0x123' as Address,
                    challenge: 100n,
                    nonce: 1n,
                },
                // result[1] - status
                0, // ChannelStatus.INITIAL
                // result[2] - wallets
                ['0xabc', '0xdef'] as readonly Address[],
                // result[3] - challengeExpiry
                1234567890n,
                // result[4] - lastValidState
                {
                    intent: 0,
                    version: 1n,
                    data: '0x123' as Hex,
                    allocations: [] as readonly any[],
                    sigs: [] as readonly any[],
                },
            ];

            const expectedResult = {
                channel: {
                    participants: ['0xabc', '0xdef'],
                    adjudicator: '0x123',
                    challenge: 100n,
                    nonce: 1n,
                },
                status: 0,
                wallets: ['0xabc', '0xdef'],
                challengeExpiry: 1234567890n,
                lastValidState: {
                    intent: 0,
                    version: 1n,
                    data: '0x123',
                    allocations: [],
                    sigs: [],
                },
            };

            (mockPublicClient.readContract as any).mockResolvedValue(contractResult);
            const res = await service.getChannelData(channelId);
            expect(res).toEqual(expectedResult);
            expect(mockPublicClient.readContract).toHaveBeenCalledWith({
                address: custodyAddress,
                abi: CustodyAbi,
                functionName: 'getChannelData',
                args: [channelId],
            });
        });
        test('ContractReadError', async () => {
            (mockPublicClient.readContract as any).mockRejectedValue(new Error('fail'));
            await expect(service.getChannelData(channelId)).rejects.toThrow(Errors.ContractReadError);
        });
    });
});
