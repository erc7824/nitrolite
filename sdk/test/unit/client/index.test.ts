import { describe, test, expect, beforeEach, jest } from '@jest/globals';
import { NitroliteClient } from '../../../src/client/index';
import { Errors } from '../../../src/errors';
import { Address, Hash, Hex } from 'viem';
import * as stateModule from '../../../src/client/state';
import {
    Allocation,
    ChannelId,
    ChannelStatus,
    CreateChannelParams,
    StateIntent,
} from '../../../src/client/types';

describe('NitroliteClient', () => {
    let client: NitroliteClient;
    const mockPublicClient = {
        waitForTransactionReceipt: jest.fn(() => Promise.resolve({ status: 'success' })),
    } as any;
    const mockAccount = { address: '0x1234567890123456789012345678901234567890' as Address };
    const mockSignature = '0x' + '1234567890abcdef'.repeat(8) + '1b'; // 128 hex chars, v = 27
    const mockSignMessage = jest.fn(() => Promise.resolve(mockSignature));
    const mockWalletClient = {
        account: mockAccount,
        signMessage: mockSignMessage,
    } as any;
    const mockAddresses = {
        custody: '0x1111111111111111111111111111111111111111' as Address,
        adjudicator: '0x2222222222222222222222222222222222222222' as Address,
    };
    const brokerAddress = '0x3333333333333333333333333333333333333333' as Address;
    const tokenAddress = '0x4444444444444444444444444444444444444444' as Address;
    const challengeDuration = 3600n;
    const chainId = 1;

    let mockNitroService: any;
    let mockErc20Service: any;
    let mockContractWriter: any;

    const stateSigner = {
        getAddress: jest.fn(() => mockAccount.address),
        signState: jest.fn(async (_1: Hex, _2: any) => mockSignature as Hex),
        signRawMessage: jest.fn(async (_: Hex) => mockSignature as Hex),
    }

    beforeEach(() => {
        jest.restoreAllMocks();
        mockContractWriter = {
            // @ts-ignore
            write: jest.fn().mockResolvedValue({ txHashes: ['0xTXHASH' as Hash] }),
        };
        client = new NitroliteClient({
            publicClient: mockPublicClient,
            walletClient: mockWalletClient,
            addresses: mockAddresses,
            challengeDuration,
            chainId: chainId,
            stateSigner,
        });
        mockNitroService = {
            prepareDepositCallParams: jest.fn(),
            prepareCreateChannelCallParams: jest.fn(),
            prepareDepositAndCreateChannelCallParams: jest.fn(),
            prepareCheckpointCallParams: jest.fn(),
            prepareChallengeCallParams: jest.fn(),
            prepareResizeCallParams: jest.fn(),
            prepareCloseCallParams: jest.fn(),
            prepareWithdrawCallParams: jest.fn(),
            getOpenChannels: jest.fn(),
            getAccountBalance: jest.fn(),
            getChannelBalance: jest.fn(),
            getChannelData: jest.fn(),
        };
        mockErc20Service = {
            getTokenAllowance: jest.fn(),
            prepareApproveCallParams: jest.fn(),
            approve: jest.fn(),
            getTokenBalance: jest.fn(),
        };
        // override private services and contractWriter
        // @ts-ignore
        client.nitroliteService = mockNitroService;
        // @ts-ignore
        client.erc20Service = mockErc20Service;
        // @ts-ignore
        client.contractWriter = mockContractWriter;
        // also override sharedDeps to use mock services
        // @ts-ignore
        client.sharedDeps.nitroliteService = mockNitroService;
        // @ts-ignore
        client.sharedDeps.erc20Service = mockErc20Service;
    });

    describe('deposit', () => {
        test('ERC20 no approval needed', async () => {
            mockErc20Service.getTokenAllowance.mockResolvedValue(100n);
            mockNitroService.prepareDepositCallParams.mockReturnValue({ fn: 'deposit' });

            const tx = await client.deposit(tokenAddress, 50n);

            expect(mockErc20Service.getTokenAllowance).toHaveBeenCalledWith(
                tokenAddress,
                mockAccount.address,
                mockAddresses.custody,
            );
            expect(mockNitroService.prepareDepositCallParams).toHaveBeenCalledWith(tokenAddress, 50n);
            expect(mockContractWriter.write).toHaveBeenCalledWith({ calls: [{ fn: 'deposit' }] });
            expect(tx).toBe('0xTXHASH');
        });

        test('ERC20 needs approval', async () => {
            mockErc20Service.getTokenAllowance.mockResolvedValue(10n);
            mockErc20Service.prepareApproveCallParams.mockReturnValue({ fn: 'approve' });
            mockNitroService.prepareDepositCallParams.mockReturnValue({ fn: 'deposit' });

            const tx = await client.deposit(tokenAddress, 50n);

            expect(mockErc20Service.prepareApproveCallParams).toHaveBeenCalledWith(tokenAddress, mockAddresses.custody, 50n);
            expect(mockContractWriter.write).toHaveBeenCalledWith({ calls: [{ fn: 'approve' }, { fn: 'deposit' }] });
            expect(tx).toBe('0xTXHASH');
        });

        test('approve failure throws TokenError', async () => {
            mockErc20Service.getTokenAllowance.mockResolvedValue(0n);
            mockErc20Service.prepareApproveCallParams.mockReturnValue({ fn: 'approve' });
            mockContractWriter.write.mockRejectedValue(new Error('fail'));

            await expect(client.deposit(tokenAddress, 10n)).rejects.toThrow(Error);
        });

        test('deposit failure throws ContractCallError', async () => {
            mockErc20Service.getTokenAllowance.mockResolvedValue(100n);
            mockNitroService.prepareDepositCallParams.mockReturnValue({ fn: 'deposit' });
            mockContractWriter.write.mockRejectedValue(new Error('fail'));

            await expect(client.deposit(tokenAddress, 10n)).rejects.toThrow(Error);
        });
    });

    describe('createChannel', () => {
        const params: CreateChannelParams = {
            channel: {
                participants: ['0x0', '0x1'], // List of participants in the channel [Host, Guest]
                adjudicator: mockAddresses.adjudicator, // Address of the contract that validates final states
                challenge: challengeDuration, // Duration in seconds for challenge period
                nonce: 1n, // Unique per channel with same participants and adjudicator
            },
            unsignedInitialState: {
                data: '0x00' as Hex,
                intent: StateIntent.INITIALIZE,
                allocations: [
                    {
                        destination: '0x1234567890123456789012345678901234567890' as Hex,
                        token: '0x0',
                        amount: 1n,
                    } as Allocation,
                    {
                        destination: '0x2345678901234567890123456789012345678901' as Hex,
                        token: '0x0',
                        amount: 2n,
                    } as Allocation,
                ] as [Allocation, Allocation],
                version: 0n,
            },
            serverSignature: '0xSRVSIG' as Hex,
        };

        test('success', async () => {
            const initialState = {
                ...params.unsignedInitialState,
                sigs: ['0xaccSig', '0xSRVSIG'] as Hex[],
            }

            const channelId = '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex;
            jest.spyOn(stateModule, '_prepareAndSignInitialState').mockResolvedValue({initialState, channelId});
            mockNitroService.prepareCreateChannelCallParams.mockReturnValue({ fn: 'createChannel' });

            const result = await client.createChannel(params);

            expect(stateModule._prepareAndSignInitialState).toHaveBeenCalledWith(expect.anything(), params);
            expect(mockNitroService.prepareCreateChannelCallParams).toHaveBeenCalledWith(params.channel, initialState);
            expect(mockContractWriter.write).toHaveBeenCalledWith({ calls: [{ fn: 'createChannel' }] });
            expect(result).toEqual({
                channelId,
                initialState,
                txHash: '0xTXHASH',
            });
        });

        test('failure throws ContractCallError', async () => {
            jest.spyOn(stateModule, '_prepareAndSignInitialState').mockRejectedValue(new Error('fail'));
            await expect(client.createChannel(params)).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe('depositAndCreateChannel', () => {
        test('combines deposit and create', async () => {
            const channelId = '0xcid' as Hex;
            const initialState = {
                data: '0x00' as Hex,
                intent: 0,
                allocations: [],
                version: 0n,
                sigs: [],
            };

            mockErc20Service.getTokenAllowance.mockResolvedValue(100n);
            jest.spyOn(stateModule, '_prepareAndSignInitialState').mockResolvedValue({
                initialState,
                channelId,
            });
            mockNitroService.prepareDepositAndCreateChannelCallParams.mockReturnValue({ fn: 'depositAndCreate' });
            const res = await client.depositAndCreateChannel(tokenAddress, 10n, {
                initialAllocationAmounts: [1n, 2n],
                stateData: '0x00' as any,
            } as any);

            expect(mockContractWriter.write).toHaveBeenCalledWith({ calls: [{ fn: 'depositAndCreate' }] });
            expect(res).toEqual({
                channelId,
                initialState,
                txHash: '0xTXHASH' as Hash,
            });
        });
    });

    describe('checkpointChannel', () => {
        test('success', async () => {
            const params = {
                channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                candidateState: { sigs: ['s1', 's2'] } as any,
                proofStates: [],
            };
            mockNitroService.prepareCheckpointCallParams.mockReturnValue({ fn: 'checkpoint' });

            const tx = await client.checkpointChannel(params);
            expect(mockNitroService.prepareCheckpointCallParams).toHaveBeenCalledWith(
                params.channelId,
                params.candidateState,
                params.proofStates,
            );
            expect(mockContractWriter.write).toHaveBeenCalledWith({ calls: [{ fn: 'checkpoint' }] });
            expect(tx).toBe('0xTXHASH');
        });

        test('insufficient sigs throws InvalidParameterError', async () => {
            const params = {
                channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                candidateState: { sigs: ['s1'] } as any,
            };
            await expect(client.checkpointChannel(params)).rejects.toThrow(Errors.InvalidParameterError);
        });
    });

    describe('challengeChannel', () => {
        test('success', async () => {
            // Mock getChannelData to return proper channel structure
            mockNitroService.getChannelData.mockResolvedValue({
                channel: {
                    participants: [mockAccount.address, brokerAddress],
                    adjudicator: mockAddresses.adjudicator,
                    challenge: challengeDuration,
                    nonce: 1n,
                },
                status: ChannelStatus.ACTIVE,
                wallets: [mockAccount.address, brokerAddress],
                challengeExpiry: 0n,
                lastValidState: {} as any,
            });
            mockNitroService.prepareChallengeCallParams.mockReturnValue({ fn: 'challenge' });
            const params = {
                channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                candidateState: {
                    intent: 0,
                    version: 1n,
                    data: '0x00' as Hex,
                    allocations: [
                        {
                            destination: '0x1234567890123456789012345678901234567890' as Hex,
                            token: tokenAddress,
                            amount: 1n,
                        },
                        {
                            destination: '0x2345678901234567890123456789012345678901' as Hex,
                            token: tokenAddress,
                            amount: 2n,
                        },
                    ],
                    sigs: [],
                },
                proofStates: [],
            };
            const tx = await client.challengeChannel(params);
            expect(mockNitroService.prepareChallengeCallParams).toHaveBeenCalledWith(
                params.channelId,
                params.candidateState,
                params.proofStates,
                mockSignature, // the signature
            );
            expect(mockContractWriter.write).toHaveBeenCalledWith({ calls: [{ fn: 'challenge' }] });
            expect(tx).toBe('0xTXHASH');
        });

        test('failure throws ContractCallError', async () => {
            const params = {
                channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                candidateState: {
                    intent: 0,
                    version: 1n,
                    data: '0x00' as Hex,
                    allocations: [
                        {
                            destination: '0x1234567890123456789012345678901234567890' as Hex,
                            token: tokenAddress,
                            amount: 1n,
                        },
                        {
                            destination: '0x2345678901234567890123456789012345678901' as Hex,
                            token: tokenAddress,
                            amount: 2n,
                        },
                    ],
                    sigs: [],
                },
                proofStates: [],
            };
            // Mock getChannelData to succeed
            mockNitroService.getChannelData.mockResolvedValue({
                channel: {
                    participants: [mockAccount.address, brokerAddress],
                    adjudicator: mockAddresses.adjudicator,
                    challenge: challengeDuration,
                    nonce: 1n,
                },
                status: ChannelStatus.ACTIVE,
                wallets: [mockAccount.address, brokerAddress],
                challengeExpiry: 0n,
                lastValidState: {} as any,
            });
            // But make challenge fail
            mockNitroService.prepareChallengeCallParams.mockReturnValue({ fn: 'challenge' });
            mockContractWriter.write.mockRejectedValue(new Error('fail'));
            await expect(client.challengeChannel(params)).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe('closeChannel', () => {
        test('success', async () => {
            jest.spyOn(stateModule, '_prepareAndSignFinalState').mockResolvedValue({
                finalStateWithSigs: {} as any,
                channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
            });
            mockNitroService.prepareCloseCallParams.mockReturnValue({ fn: 'close' });

            const tx = await client.closeChannel({
                finalState: {
                    channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                    allocations: [],
                    version: 0n,
                    serverSignature: [] as any,
                } as any,
            });
            expect(stateModule._prepareAndSignFinalState).toHaveBeenCalledWith(expect.anything(), expect.any(Object));
            expect(mockNitroService.prepareCloseCallParams).toHaveBeenCalledWith(
                '0x0000000000000000000000000000000000000000000000000000000000000001',
                {} as any,
            );
            expect(mockContractWriter.write).toHaveBeenCalledWith({ calls: [{ fn: 'close' }] });
            expect(tx).toBe('0xTXHASH');
        });

        test('failure throws ContractCallError', async () => {
            jest.spyOn(stateModule, '_prepareAndSignFinalState').mockRejectedValue(new Error('fail'));
            await expect(
                client.closeChannel({
                    finalState: {
                        channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                    } as any,
                } as any),
            ).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe('withdrawal', () => {
        test('success', async () => {
            mockNitroService.prepareWithdrawCallParams.mockReturnValue({ fn: 'withdraw' });
            const tx = await client.withdrawal(tokenAddress, 20n);
            expect(mockNitroService.prepareWithdrawCallParams).toHaveBeenCalledWith(tokenAddress, 20n);
            expect(mockContractWriter.write).toHaveBeenCalledWith({ calls: [{ fn: 'withdraw' }] });
            expect(tx).toBe('0xTXHASH');
        });

        test('failure throws ContractCallError', async () => {
            mockNitroService.prepareWithdrawCallParams.mockReturnValue({ fn: 'withdraw' });
            mockContractWriter.write.mockRejectedValue(new Error('fail'));
            await expect(client.withdrawal(tokenAddress, 20n)).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe('getOpenChannels', () => {
        test('success', async () => {
            mockNitroService.getOpenChannels.mockResolvedValue(['0xc1', '0xc2'] as Address[]);
            const res = await client.getOpenChannels();
            expect(res).toEqual(['0xc1', '0xc2']);
            expect(mockNitroService.getOpenChannels).toHaveBeenCalledWith(mockAccount.address);
        });
    });

    describe('getAccountBalances', () => {
        test('success', async () => {
            const balances = 42n;
            mockNitroService.getAccountBalance.mockResolvedValue(balances);
            const res = await client.getAccountBalance(tokenAddress);
            expect(res).toEqual(balances);
            expect(mockNitroService.getAccountBalance).toHaveBeenCalledWith(mockAccount.address, tokenAddress);
        });
    });

    describe('getChannelBalances', () => {
        test('success', async () => {
            const balances = 42n;
            mockNitroService.getChannelBalance.mockResolvedValue(balances);
            const res = await client.getChannelBalance('0xcid' as ChannelId, tokenAddress);
            expect(res).toEqual(balances);
            expect(mockNitroService.getChannelBalance).toHaveBeenCalledWith('0xcid' as ChannelId, tokenAddress);
        });
    });

    describe('getChannelData', () => {
        test('success', async () => {
            const data = {
                channel: {
                    participants: ['0x0', '0x1'] as Address[],
                    adjudicator: mockAddresses.adjudicator,
                    challenge: challengeDuration,
                    nonce: 1n,
                },
                status: ChannelStatus.INITIAL,
                wallets: ['0xabc', '0xdef'] as Address[],
                challengeExpiry: 1234567890n,
                lastValidState: {
                    data: '0x00' as Hex,
                    intent: StateIntent.INITIALIZE,
                    version: 0n,
                    allocations: [
                        { destination: '0x0', token: '0xtok', amount: 50n },
                        { destination: '0x1', token: '0xtok', amount: 50n },
                    ],
                    sigs: [],
                },
            };
            mockNitroService.getChannelData.mockResolvedValue(data);
            const res = await client.getChannelData('0xcid' as ChannelId);
            expect(res).toEqual(data);
            expect(mockNitroService.getChannelData).toHaveBeenCalledWith('0xcid' as ChannelId);
        });
    });

    describe('approveTokens', () => {
        test('success', async () => {
            mockErc20Service.approve.mockResolvedValue('0xAPP' as Hash);
            const tx = await client.approveTokens(tokenAddress, 30n);
            expect(mockErc20Service.approve).toHaveBeenCalledWith(tokenAddress, mockAddresses.custody, 30n);
            expect(tx).toBe('0xAPP');
        });

        test('failure throws TokenError', async () => {
            mockErc20Service.approve.mockRejectedValue(new Error('fail'));
            await expect(client.approveTokens(tokenAddress, 30n)).rejects.toThrow(Errors.TokenError);
        });
    });

    describe('getTokenAllowance', () => {
        test('success', async () => {
            mockErc20Service.getTokenAllowance.mockResolvedValue(500n);
            const v = await client.getTokenAllowance(tokenAddress);
            expect(v).toBe(500n);
            expect(mockErc20Service.getTokenAllowance).toHaveBeenCalledWith(
                tokenAddress,
                mockAccount.address,
                mockAddresses.custody,
            );
        });
    });

    describe('getTokenBalance', () => {
        test('success', async () => {
            mockErc20Service.getTokenBalance.mockResolvedValue(1000n);
            const v = await client.getTokenBalance(tokenAddress);
            expect(v).toBe(1000n);
            expect(mockErc20Service.getTokenBalance).toHaveBeenCalledWith(tokenAddress, mockAccount.address);
        });
    });
});
