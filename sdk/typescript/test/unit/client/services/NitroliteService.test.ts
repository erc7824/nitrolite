import { describe, test, expect, beforeEach, jest } from '@jest/globals';
import { Address, Hex, zeroAddress } from 'viem';
import { NitroliteService } from '../../../../src/client/services/NitroliteService';
import * as Errors from '../../../../src/errors';
import { ContractAddresses } from '../../../../src/abis';
import { ChannelDefinition, ChannelId, State, StateIntent } from '../../../../src/client/types';

describe('NitroliteService', () => {
    const custodyAddress = '0x0000000000000000000000000000000000000001' as Address;
    const addresses: ContractAddresses = { custody: custodyAddress } as any;
    const account = '0x0000000000000000000000000000000000000002' as Address;
    const nodeAddress = '0x0000000000000000000000000000000000000003' as Address;

    // Dummy data for methods
    const channelDefinition: ChannelDefinition = {
        challengeDuration: 3600,
        user: account,
        node: nodeAddress,
        nonce: 1n,
        metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
    };

    const initialState: State = {
        version: 0n,
        intent: StateIntent.INITIALIZE,
        metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
        homeState: {
            chainId: 1n,
            token: '0x0000000000000000000000000000000000000004' as Address,
            decimals: 18,
            userAllocation: 100n,
            userNetFlow: 0n,
            nodeAllocation: 100n,
            nodeNetFlow: 0n,
        },
        nonHomeState: {
            chainId: 0n,
            token: zeroAddress,
            decimals: 0,
            userAllocation: 0n,
            userNetFlow: 0n,
            nodeAllocation: 0n,
            nodeNetFlow: 0n,
        },
        userSig: '0x' as Hex,
        nodeSig: '0x' as Hex,
    };

    const channelId = '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef' as ChannelId;
    const tokenAddress = '0x0000000000000000000000000000000000000004' as Address;

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
            account: { address: account },
        };
        service = new NitroliteService(mockPublicClient, addresses, mockWalletClient, { address: account });
    });

    describe('constructor', () => {
        test('throws if publicClient missing', () => {
            expect(() => new NitroliteService(undefined as any, addresses)).toThrow(Errors.MissingParameterError);
        });
        test('throws if addresses.custody missing', () => {
            expect(() => new NitroliteService(mockPublicClient, {} as any, mockWalletClient, { address: account })).toThrow(
                Errors.MissingParameterError,
            );
        });
    });

    describe('deposit', () => {
        test('success with ERC20', async () => {
            mockPublicClient.simulateContract.mockResolvedValue({
                request: { to: custodyAddress, data: '0x' },
            });
            mockWalletClient.writeContract.mockResolvedValue('0xhash');

            const hash = await service.deposit(nodeAddress, tokenAddress, 100n);
            expect(hash).toBe('0xhash');
            expect(mockPublicClient.simulateContract).toHaveBeenCalled();
            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });

        test('success with ETH', async () => {
            mockPublicClient.simulateContract.mockResolvedValue({
                request: { to: custodyAddress, data: '0x' },
            });
            mockWalletClient.writeContract.mockResolvedValue('0xhash');

            const hash = await service.deposit(nodeAddress, zeroAddress, 100n);
            expect(hash).toBe('0xhash');
        });
    });

    describe('createChannel', () => {
        test('success', async () => {
            mockPublicClient.simulateContract.mockResolvedValue({
                request: { to: custodyAddress, data: '0x' },
            });
            mockWalletClient.writeContract.mockResolvedValue('0xhash');

            const hash = await service.createChannel(channelDefinition, initialState);
            expect(hash).toBe('0xhash');
            expect(mockPublicClient.simulateContract).toHaveBeenCalled();
            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });

        test('TransactionError on write failure', async () => {
            mockPublicClient.simulateContract.mockResolvedValue({
                request: { to: custodyAddress, data: '0x' },
            });
            mockWalletClient.writeContract.mockRejectedValue(new Error('oops'));

            await expect(service.createChannel(channelDefinition, initialState)).rejects.toThrow(Errors.TransactionError);
        });
    });

    describe('checkpointChannel', () => {
        test('success', async () => {
            mockPublicClient.simulateContract.mockResolvedValue({
                request: { to: custodyAddress, data: '0x' },
            });
            mockWalletClient.writeContract.mockResolvedValue('0xhash');

            const hash = await service.checkpointChannel(channelId, initialState, []);
            expect(hash).toBe('0xhash');
            expect(mockPublicClient.simulateContract).toHaveBeenCalled();
            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });
    });

    describe('challengeChannel', () => {
        test('success', async () => {
            mockPublicClient.simulateContract.mockResolvedValue({
                request: { to: custodyAddress, data: '0x' },
            });
            mockWalletClient.writeContract.mockResolvedValue('0xhash');

            const hash = await service.challengeChannel(channelId, initialState, [], '0xchallengerSig' as Hex);
            expect(hash).toBe('0xhash');
            expect(mockPublicClient.simulateContract).toHaveBeenCalled();
            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });
    });

    describe('closeChannel', () => {
        test('success', async () => {
            mockPublicClient.simulateContract.mockResolvedValue({
                request: { to: custodyAddress, data: '0x' },
            });
            mockWalletClient.writeContract.mockResolvedValue('0xhash');

            const hash = await service.closeChannel(channelId, initialState, []);
            expect(hash).toBe('0xhash');
            expect(mockPublicClient.simulateContract).toHaveBeenCalled();
            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });
    });

    describe('withdraw', () => {
        test('success', async () => {
            mockPublicClient.simulateContract.mockResolvedValue({
                request: { to: custodyAddress, data: '0x' },
            });
            mockWalletClient.writeContract.mockResolvedValue('0xhash');

            const hash = await service.withdraw(account, tokenAddress, 50n);
            expect(hash).toBe('0xhash');
            expect(mockPublicClient.simulateContract).toHaveBeenCalled();
            expect(mockWalletClient.writeContract).toHaveBeenCalled();
        });
    });

    describe('getOpenChannels', () => {
        test('success', async () => {
            const channels = ['0xABC', '0xDEF'] as ChannelId[];
            mockPublicClient.readContract.mockResolvedValue(channels);

            const result = await service.getOpenChannels(account);
            expect(result).toEqual(channels);
            expect(mockPublicClient.readContract).toHaveBeenCalled();
        });

        test('ContractReadError', async () => {
            mockPublicClient.readContract.mockRejectedValue(new Error('fail'));

            await expect(service.getOpenChannels(account)).rejects.toThrow(Errors.ContractReadError);
        });
    });

    describe('getAccountBalance', () => {
        test('success', async () => {
            mockPublicClient.readContract.mockResolvedValue(1000n);

            const balance = await service.getAccountBalance(nodeAddress, tokenAddress);
            expect(balance).toBe(1000n);
            expect(mockPublicClient.readContract).toHaveBeenCalled();
        });

        test('ContractReadError', async () => {
            mockPublicClient.readContract.mockRejectedValue(new Error('fail'));

            await expect(service.getAccountBalance(nodeAddress, tokenAddress)).rejects.toThrow(Errors.ContractReadError);
        });
    });

    describe('getChannelData', () => {
        test('success', async () => {
            const channelData = {
                status: 1,
                definition: channelDefinition,
                lastState: initialState,
                challengeExpiry: 0n,
            };
            // Mock return value order: [status, definition, lastState, challengeExpiry]
            mockPublicClient.readContract.mockResolvedValue([
                1,
                channelDefinition,
                initialState,
                0n,
            ]);

            const result = await service.getChannelData(channelId);
            expect(result.definition).toEqual(channelDefinition);
            expect(mockPublicClient.readContract).toHaveBeenCalled();
        });

        test('ContractReadError', async () => {
            mockPublicClient.readContract.mockRejectedValue(new Error('fail'));

            await expect(service.getChannelData(channelId)).rejects.toThrow(Errors.ContractReadError);
        });
    });
});
