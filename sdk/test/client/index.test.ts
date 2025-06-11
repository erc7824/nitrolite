import { describe, test, expect, beforeEach, jest } from '@jest/globals';
import { NitroliteClient } from '../../src/client/index';
import { Errors } from '../../src/errors';
import { Address, Hash, Hex } from 'viem';
import * as stateModule from '../../src/client/state';
import { Allocation, Channel, ChannelId, ChannelStatus, State, StateIntent } from '../../src/client/types';

describe('NitroliteClient', () => {
    let client: NitroliteClient;
    const mockPublicClient = {
        waitForTransactionReceipt: jest.fn().mockResolvedValue({ status: 'success' }),
    } as any;
    const mockAccount = { address: '0xOWNER' as Address };
    const mockWalletClient = { account: mockAccount } as any;
    const mockAddresses = {
        custody: '0xCUST' as Address,
        adjudicator: '0xADJ' as Address,
        guestAddress: '0xGUEST' as Address,
    };
    const tokenAddress = '0xTOKEN' as Address;
    const challengeDuration = 3600n;
    const chainId = 1;

    let mockNitroService: any;
    let mockErc20Service: any;

    beforeEach(() => {
        jest.restoreAllMocks();
        client = new NitroliteClient({
            publicClient: mockPublicClient,
            walletClient: mockWalletClient,
            addresses: mockAddresses,
            challengeDuration,
            chainId: chainId,
        });
        mockNitroService = {
            deposit: jest.fn(),
            createChannel: jest.fn(),
            depositAndCreateChannel: jest.fn(),
            checkpoint: jest.fn(),
            challenge: jest.fn(),
            close: jest.fn(),
            withdraw: jest.fn(),
            getOpenChannels: jest.fn(),
            getAccountBalance: jest.fn(),
            getChannelBalance: jest.fn(),
            getChannelData: jest.fn(),
        };
        mockErc20Service = {
            getTokenAllowance: jest.fn(),
            approve: jest.fn(),
            getTokenBalance: jest.fn(),
        };
        // override private services
        // @ts-ignore
        client.nitroliteService = mockNitroService;
        // @ts-ignore
        client.erc20Service = mockErc20Service;
    });

    describe('deposit', () => {
        test('ERC20 no approval needed', async () => {
            mockErc20Service.getTokenAllowance.mockResolvedValue(100n);
            mockNitroService.deposit.mockResolvedValue('0xDEP' as Hash);

            const tx = await client.deposit(tokenAddress, 50n);

            expect(mockErc20Service.getTokenAllowance).toHaveBeenCalledWith(
                tokenAddress,
                mockAccount.address,
                mockAddresses.custody,
            );
            expect(mockNitroService.deposit).toHaveBeenCalledWith(tokenAddress, 50n);
            expect(tx).toBe('0xDEP');
        });

        test('ERC20 needs approval', async () => {
            mockErc20Service.getTokenAllowance.mockResolvedValue(10n);
            mockErc20Service.approve.mockResolvedValue('0xAPP' as Hash);
            mockNitroService.deposit.mockResolvedValue('0xDEP' as Hash);

            const tx = await client.deposit(tokenAddress, 50n);

            expect(mockErc20Service.approve).toHaveBeenCalledWith(
                tokenAddress,
                mockAddresses.custody,
                50n,
            );
            expect(tx).toBe('0xDEP');
        });

        test('approve failure throws TokenError', async () => {
            mockErc20Service.getTokenAllowance.mockResolvedValue(0n);
            mockErc20Service.approve.mockRejectedValue(new Error('fail'));

            await expect(client.deposit(tokenAddress, 10n)).rejects.toThrow(Errors.TokenError);
        });

        test('deposit failure throws ContractCallError', async () => {
            mockErc20Service.getTokenAllowance.mockResolvedValue(100n);
            mockNitroService.deposit.mockRejectedValue(new Error('fail'));

            await expect(client.deposit(tokenAddress, 10n)).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe('createChannel', () => {
        const params = {
            initialAllocationAmounts: [1n, 2n] as [bigint, bigint],
            stateData: '0x00' as any,
        };

        test('success', async () => {
            const channel: Channel = {
                participants: ['0x0', '0x1'], // List of participants in the channel [Host, Guest]
                adjudicator: mockAddresses.adjudicator, // Address of the contract that validates final states
                challenge: challengeDuration, // Duration in seconds for challenge period
                nonce: 1n, // Unique per channel with same participants and adjudicator
            };

            const initialState = {
                data: '0x00' as Hex,
                intent: StateIntent.INITIALIZE,
                allocations: [
                    { destination: '0x0' as Hex, token: '0x0', amount: 1n } as Allocation,
                    { destination: '0x1' as Hex, token: '0x0', amount: 2n } as Allocation,
                ] as [Allocation, Allocation],
                version: 0n,
                sigs: [],
            };
            const channelId = '0xcid' as Hex;
            jest.spyOn(stateModule, '_prepareAndSignInitialState').mockResolvedValue({
                channel,
                initialState,
                channelId,
            });
            mockNitroService.createChannel.mockResolvedValue('0xCRE' as Hash);

            const result = await client.createChannel(tokenAddress, params);

            expect(stateModule._prepareAndSignInitialState).toHaveBeenCalledWith(tokenAddress, expect.anything(), params);
            expect(mockNitroService.createChannel).toHaveBeenCalledWith(channel, initialState);
            expect(result).toEqual({
                channelId,
                initialState,
                txHash: '0xCRE',
            });
        });

        test('failure throws ContractCallError', async () => {
            jest.spyOn(stateModule, '_prepareAndSignInitialState').mockRejectedValue(new Error('fail'));
            await expect(client.createChannel(tokenAddress, params)).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe('depositAndCreateChannel', () => {
        test('combines deposit and create', async () => {
            mockNitroService.depositAndCreateChannel.mockResolvedValue('0xDEPandCRE' as Hash);

            const res = await client.depositAndCreateChannel(tokenAddress, 10n, {
                initialAllocationAmounts: [1n, 2n],
                stateData: '0x00' as any,
            } as any);

            expect(client.deposit).toHaveBeenCalledWith(tokenAddress, 10n);
            expect(client.createChannel).toHaveBeenCalledWith(tokenAddress, expect.any(Object));
            expect(res).toEqual({
                channelId: '0xcid' as Hex,
                initialState: {},
                txHash: '0xDEPandCRE' as Hash,
            });
        });
    });

    describe('checkpointChannel', () => {
        test('success', async () => {
            const params = {
                channelId: '0xcid' as Hex,
                candidateState: { sigs: ['s1', 's2'] } as any,
                proofStates: [],
            };
            mockNitroService.checkpoint.mockResolvedValue('0xCHK' as Hash);

            const tx = await client.checkpointChannel(params);
            expect(mockNitroService.checkpoint).toHaveBeenCalledWith(
                params.channelId,
                params.candidateState,
                params.proofStates,
            );
            expect(tx).toBe('0xCHK');
        });

        test('insufficient sigs throws InvalidParameterError', async () => {
            const params = {
                channelId: '0xcid' as Hex,
                candidateState: { sigs: ['s1'] } as any,
            };
            await expect(client.checkpointChannel(params)).rejects.toThrow(Errors.InvalidParameterError);
        });
    });

    describe('challengeChannel', () => {
        test('success', async () => {
            const params = {
                channelId: '0xcid' as Hex,
                candidateState: {} as any,
                proofStates: [],
            };
            mockNitroService.challenge.mockResolvedValue('0xCHL' as Hash);
            const tx = await client.challengeChannel(params);
            expect(mockNitroService.challenge).toHaveBeenCalledWith(
                params.channelId,
                params.candidateState,
                params.proofStates,
            );
            expect(tx).toBe('0xCHL');
        });

        test('failure throws ContractCallError', async () => {
            const params = {
                channelId: '0xcid' as Hex,
                candidateState: {} as any,
                proofStates: [],
            };
            mockNitroService.challenge.mockRejectedValue(new Error('fail'));
            await expect(client.challengeChannel(params)).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe('closeChannel', () => {
        test('success', async () => {
            jest.spyOn(stateModule, '_prepareAndSignFinalState').mockResolvedValue({
                finalStateWithSigs: {} as any,
                channelId: '0xcid' as Hex,
            });
            mockNitroService.close.mockResolvedValue('0xCLS' as Hash);

            const tx = await client.closeChannel({
                finalState: {
                    channelId: '0xcid',
                    allocations: [],
                    version: 0n,
                    serverSignature: [] as any,
                } as any,
            });
            expect(stateModule._prepareAndSignFinalState).toHaveBeenCalledWith(expect.anything(), expect.any(Object));
            expect(mockNitroService.close).toHaveBeenCalledWith('0xcid', {} as any);
            expect(tx).toBe('0xCLS');
        });

        test('failure throws ContractCallError', async () => {
            jest.spyOn(stateModule, '_prepareAndSignFinalState').mockRejectedValue(new Error('fail'));
            await expect(client.closeChannel({ finalState: { channelId: '0xcid' } as any } as any)).rejects.toThrow(
                Errors.ContractCallError,
            );
        });
    });

    describe('withdrawal', () => {
        test('success', async () => {
            mockNitroService.withdraw.mockResolvedValue('0xWDL' as Hash);
            const tx = await client.withdrawal(tokenAddress, 20n);
            expect(mockNitroService.withdraw).toHaveBeenCalledWith(tokenAddress, 20n);
            expect(tx).toBe('0xWDL');
        });

        test('failure throws ContractCallError', async () => {
            mockNitroService.withdraw.mockRejectedValue(new Error('fail'));
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
            expect(mockNitroService.getAccountBalance).toHaveBeenCalledWith(
                mockAccount.address,
                tokenAddress,
            );
        });
    });

    describe('getChannelBalances', () => {
        test('success', async () => {
            const balances = 42n;
            mockNitroService.getChannelBalance.mockResolvedValue(balances);
            const res = await client.getChannelBalance('0xcid' as ChannelId, tokenAddress);
            expect(res).toEqual(balances);
            expect(mockNitroService.getChannelBalance).toHaveBeenCalledWith(
                '0xcid' as ChannelId,
                tokenAddress,
            );
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
            expect(mockNitroService.getChannelData).toHaveBeenCalledWith(
                '0xcid' as ChannelId
            );
        });
    });

    describe('approveTokens', () => {
        test('success', async () => {
            mockErc20Service.approve.mockResolvedValue('0xAPP' as Hash);
            const tx = await client.approveTokens(tokenAddress, 30n);
            expect(mockErc20Service.approve).toHaveBeenCalledWith(
                tokenAddress,
                mockAddresses.custody,
                30n,
            );
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
            expect(mockErc20Service.getTokenBalance).toHaveBeenCalledWith(
                tokenAddress,
                mockAccount.address,
            );
        });
    });
});
