import { describe, test, expect, beforeEach, jest } from '@jest/globals';
import { NitroliteClient } from '../../src/client/index';
import { Errors } from '../../src/errors';
import { Address, Hash, Hex } from 'viem';
import * as stateModule from '../../src/client/state';
import { Allocation, Channel, StateIntent } from '../../src/client/types';

describe('NitroliteClient', () => {
    let client: NitroliteClient;
    const mockPublicClient = {
        // @ts-ignore
        waitForTransactionReceipt: jest.fn().mockResolvedValue({ status: 'success' }),
    } as any;
    const mockAccount = { address: '0xOWNER' as Address };
    const mockWalletClient = { account: mockAccount } as any;
    const mockAddresses = {
        custody: '0xCUST' as Address,
        adjudicator: '0xADJ' as Address,
        guestAddress: '0xGUEST' as Address,
        tokenAddress: '0xTOKEN' as Address,
    };
    const challengeDuration = 3600n;

    let mockNitroService: any;
    let mockErc20Service: any;

    beforeEach(() => {
        jest.restoreAllMocks();
        client = new NitroliteClient({
            publicClient: mockPublicClient,
            walletClient: mockWalletClient,
            addresses: mockAddresses,
            challengeDuration,
        });
        mockNitroService = {
            deposit: jest.fn(),
            createChannel: jest.fn(),
            checkpoint: jest.fn(),
            challenge: jest.fn(),
            close: jest.fn(),
            withdraw: jest.fn(),
            getAccountChannels: jest.fn(),
            getAccountInfo: jest.fn(),
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

            const tx = await client.deposit(50n);

            expect(mockErc20Service.getTokenAllowance).toHaveBeenCalledWith(
                mockAddresses.tokenAddress,
                mockAccount.address,
                mockAddresses.custody,
            );
            expect(mockNitroService.deposit).toHaveBeenCalledWith(mockAddresses.tokenAddress, 50n);
            expect(tx).toBe('0xDEP');
        });

        test('ERC20 needs approval', async () => {
            mockErc20Service.getTokenAllowance.mockResolvedValue(10n);
            mockErc20Service.approve.mockResolvedValue('0xAPP' as Hash);
            mockNitroService.deposit.mockResolvedValue('0xDEP' as Hash);

            const tx = await client.deposit(50n);

            expect(mockErc20Service.approve).toHaveBeenCalledWith(
                mockAddresses.tokenAddress,
                mockAddresses.custody,
                50n,
            );
            expect(tx).toBe('0xDEP');
        });

        test('approve failure throws TokenError', async () => {
            mockErc20Service.getTokenAllowance.mockResolvedValue(0n);
            mockErc20Service.approve.mockRejectedValue(new Error('fail'));

            await expect(client.deposit(10n)).rejects.toThrow(Errors.TokenError);
        });

        test('deposit failure throws ContractCallError', async () => {
            mockErc20Service.getTokenAllowance.mockResolvedValue(100n);
            mockNitroService.deposit.mockRejectedValue(new Error('fail'));

            await expect(client.deposit(10n)).rejects.toThrow(Errors.ContractCallError);
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

            const result = await client.createChannel(params);

            expect(stateModule._prepareAndSignInitialState).toHaveBeenCalledWith(expect.anything(), params);
            expect(mockNitroService.createChannel).toHaveBeenCalledWith(channel, initialState);
            expect(result).toEqual({
                channelId,
                initialState,
                txHash: '0xCRE',
            });
        });

        test('failure throws ContractCallError', async () => {
            jest.spyOn(stateModule, '_prepareAndSignInitialState').mockRejectedValue(new Error('fail'));
            await expect(client.createChannel(params)).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe('depositAndCreateChannel', () => {
        test('combines deposit and create', async () => {
            jest.spyOn(client, 'deposit').mockResolvedValueOnce('0xDEP' as Hash);
            jest.spyOn(client, 'createChannel').mockResolvedValueOnce({
                channelId: '0xcid' as Hex,
                initialState: {} as any,
                txHash: '0xCRE' as Hash,
            });

            const res = await client.depositAndCreateChannel(10n, {
                initialAllocationAmounts: [1n, 2n],
            } as any);

            expect(client.deposit).toHaveBeenCalledWith(10n);
            expect(client.createChannel).toHaveBeenCalledWith(expect.any(Object));
            expect(res).toEqual({
                channelId: '0xcid' as Hex,
                initialState: {},
                depositTxHash: '0xDEP',
                createChannelTxHash: '0xCRE',
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
            const tx = await client.withdrawal(20n);
            expect(mockNitroService.withdraw).toHaveBeenCalledWith(mockAddresses.tokenAddress, 20n);
            expect(tx).toBe('0xWDL');
        });

        test('failure throws ContractCallError', async () => {
            mockNitroService.withdraw.mockRejectedValue(new Error('fail'));
            await expect(client.withdrawal(20n)).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe('getAccountChannels', () => {
        test('success', async () => {
            mockNitroService.getAccountChannels.mockResolvedValue(['0xc1', '0xc2'] as Address[]);
            const res = await client.getAccountChannels();
            expect(res).toEqual(['0xc1', '0xc2']);
            expect(mockNitroService.getAccountChannels).toHaveBeenCalledWith(mockAccount.address);
        });
    });

    describe('getAccountInfo', () => {
        test('success', async () => {
            const info = {
                available: 1n,
                locked: 2n,
                channelCount: 3n,
            };
            mockNitroService.getAccountInfo.mockResolvedValue(info);
            const res = await client.getAccountInfo();
            expect(res).toEqual(info);
            expect(mockNitroService.getAccountInfo).toHaveBeenCalledWith(
                mockAccount.address,
                mockAddresses.tokenAddress,
            );
        });
    });

    describe('approveTokens', () => {
        test('success', async () => {
            mockErc20Service.approve.mockResolvedValue('0xAPP' as Hash);
            const tx = await client.approveTokens(30n);
            expect(mockErc20Service.approve).toHaveBeenCalledWith(
                mockAddresses.tokenAddress,
                mockAddresses.custody,
                30n,
            );
            expect(tx).toBe('0xAPP');
        });

        test('failure throws TokenError', async () => {
            mockErc20Service.approve.mockRejectedValue(new Error('fail'));
            await expect(client.approveTokens(30n)).rejects.toThrow(Errors.TokenError);
        });
    });

    describe('getTokenAllowance', () => {
        test('success', async () => {
            mockErc20Service.getTokenAllowance.mockResolvedValue(500n);
            const v = await client.getTokenAllowance();
            expect(v).toBe(500n);
            expect(mockErc20Service.getTokenAllowance).toHaveBeenCalledWith(
                mockAddresses.tokenAddress,
                mockAccount.address,
                mockAddresses.custody,
            );
        });
    });

    describe('getTokenBalance', () => {
        test('success', async () => {
            mockErc20Service.getTokenBalance.mockResolvedValue(1000n);
            const v = await client.getTokenBalance();
            expect(v).toBe(1000n);
            expect(mockErc20Service.getTokenBalance).toHaveBeenCalledWith(
                mockAddresses.tokenAddress,
                mockAccount.address,
            );
        });
    });
});
