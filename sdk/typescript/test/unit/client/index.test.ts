import { describe, test, expect, beforeEach, jest } from '@jest/globals';
import { NitroliteClient } from '../../../src/client/index';
import * as Errors from '../../../src/errors';
import { Address, Hash, Hex } from 'viem';
import * as stateModule from '../../../src/client/state';
import {
    ChannelId,
    ChannelStatus,
    ChannelData,
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
        writeContract: jest.fn(() => Promise.resolve('0xTX' as Hash)),
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

    const stateSigner = {
        getAddress: jest.fn(() => mockAccount.address),
        signState: jest.fn(async (_1: Hex, _2: any) => mockSignature as Hex),
        signRawMessage: jest.fn(async (_: Hex) => mockSignature as Hex),
    }

    beforeEach(() => {
        jest.restoreAllMocks();
        client = new NitroliteClient({
            publicClient: mockPublicClient,
            walletClient: mockWalletClient,
            addresses: mockAddresses,
            challengeDuration,
            chainId: chainId,
            stateSigner,
        });
        mockNitroService = {
            deposit: jest.fn(),
            createChannel: jest.fn(),
            checkpointChannel: jest.fn(),
            challengeChannel: jest.fn(),
            closeChannel: jest.fn(),
            withdraw: jest.fn(),
            getOpenChannels: jest.fn(),
            getAccountBalance: jest.fn(),
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
        // override txPreparer's dependencies as well
        // @ts-ignore
        client.txPreparer.nitroliteService = mockNitroService;
        // @ts-ignore
        client.txPreparer.erc20Service = mockErc20Service;
    });

    describe('deposit', () => {
        test('ERC20 deposit', async () => {
            mockNitroService.deposit.mockResolvedValue('0xDEP' as Hash);

            const tx = await client.deposit(brokerAddress, tokenAddress, 50n);

            expect(mockNitroService.deposit).toHaveBeenCalledWith(brokerAddress, tokenAddress, 50n);
            expect(tx).toBe('0xDEP');
        });

        test('deposit failure throws error', async () => {
            mockNitroService.deposit.mockRejectedValue(new Error('fail'));

            await expect(client.deposit(brokerAddress, tokenAddress, 10n)).rejects.toThrow();
        });
    });

    describe('createChannel', () => {
        const params: CreateChannelParams = {
            definition: {
                challengeDuration: 3600,
                user: mockAccount.address,
                node: brokerAddress,
                nonce: 1n,
                metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
            },
            initialState: {
                version: 0n,
                intent: StateIntent.INITIALIZE,
                metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
                homeState: {
                    chainId: 1n,
                    token: tokenAddress,
                    decimals: 18,
                    userAllocation: 100n,
                    userNetFlow: 0n,
                    nodeAllocation: 100n,
                    nodeNetFlow: 0n,
                },
                nonHomeState: {
                    chainId: 0n,
                    token: '0x0000000000000000000000000000000000000000' as Address,
                    decimals: 0,
                    userAllocation: 0n,
                    userNetFlow: 0n,
                    nodeAllocation: 0n,
                    nodeNetFlow: 0n,
                },
                userSig: '0xUSRSIG' as Hex,
                nodeSig: '0xNODSIG' as Hex,
            },
        };

        test('success', async () => {
            jest.spyOn(stateModule, '_prepareAndSignInitialState').mockResolvedValue({
                initialState: params.initialState,
            });
            mockNitroService.createChannel.mockResolvedValue('0xCRE' as Hash);

            const result = await client.createChannel(params);

            expect(stateModule._prepareAndSignInitialState).toHaveBeenCalledWith(expect.anything(), params);
            expect(mockNitroService.createChannel).toHaveBeenCalledWith(params.definition, params.initialState);
            expect(result).toBe('0xCRE');
        });

        test('failure throws error', async () => {
            jest.spyOn(stateModule, '_prepareAndSignInitialState').mockRejectedValue(new Error('fail'));
            await expect(client.createChannel(params)).rejects.toThrow();
        });
    });

    describe('depositAndCreateChannel', () => {
        test('combines deposit and create', async () => {
            const params: CreateChannelParams = {
                definition: {
                    challengeDuration: 3600,
                    user: mockAccount.address,
                    node: brokerAddress,
                    nonce: 1n,
                    metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
                },
                initialState: {
                    version: 0n,
                    intent: StateIntent.INITIALIZE,
                    metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
                    homeState: {
                        chainId: 1n,
                        token: tokenAddress,
                        decimals: 18,
                        userAllocation: 100n,
                        userNetFlow: 0n,
                        nodeAllocation: 100n,
                        nodeNetFlow: 0n,
                    },
                    nonHomeState: {
                        chainId: 0n,
                        token: '0x0000000000000000000000000000000000000000' as Address,
                        decimals: 0,
                        userAllocation: 0n,
                        userNetFlow: 0n,
                        nodeAllocation: 0n,
                        nodeNetFlow: 0n,
                    },
                    userSig: '0xUSRSIG' as Hex,
                    nodeSig: '0xNODSIG' as Hex,
                },
            };

            jest.spyOn(stateModule, '_prepareAndSignInitialState').mockResolvedValue({
                initialState: params.initialState,
            });
            // Mock txPreparer to return dummy transaction
            jest.spyOn(client.txPreparer, 'prepareDepositAndCreateChannelTransactions').mockResolvedValue([
                {
                    address: mockAddresses.custody,
                    abi: [],
                    functionName: 'deposit',
                    args: [],
                } as any,
            ]);

            const res = await client.depositAndCreateChannel(brokerAddress, tokenAddress, 10n, params);

            expect(mockWalletClient.writeContract).toHaveBeenCalled();
            expect(res).toBeDefined();
        });
    });

    describe('checkpointChannel', () => {
        test('success', async () => {
            const candidateState = {
                version: 1n,
                intent: StateIntent.OPERATE,
                metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
                homeState: {
                    chainId: 1n,
                    token: tokenAddress,
                    decimals: 18,
                    userAllocation: 90n,
                    userNetFlow: 0n,
                    nodeAllocation: 110n,
                    nodeNetFlow: 0n,
                },
                nonHomeState: {
                    chainId: 0n,
                    token: '0x0000000000000000000000000000000000000000' as Address,
                    decimals: 0,
                    userAllocation: 0n,
                    userNetFlow: 0n,
                    nodeAllocation: 0n,
                    nodeNetFlow: 0n,
                },
                userSig: '0xUSRSIG' as Hex,
                nodeSig: '0xNODSIG' as Hex,
            };
            const params = {
                channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                candidateState,
                proofs: [],
            };
            mockNitroService.checkpointChannel.mockResolvedValue('0xCHK' as Hash);

            const tx = await client.checkpointChannel(params);
            expect(mockNitroService.checkpointChannel).toHaveBeenCalledWith(
                params.channelId,
                params.candidateState,
                params.proofs,
            );
            expect(tx).toBe('0xCHK');
        });
    });

    describe('challengeChannel', () => {
        test('success', async () => {
            const candidateState = {
                version: 1n,
                intent: StateIntent.OPERATE,
                metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
                homeState: {
                    chainId: 1n,
                    token: tokenAddress,
                    decimals: 18,
                    userAllocation: 90n,
                    userNetFlow: 0n,
                    nodeAllocation: 110n,
                    nodeNetFlow: 0n,
                },
                nonHomeState: {
                    chainId: 0n,
                    token: '0x0000000000000000000000000000000000000000' as Address,
                    decimals: 0,
                    userAllocation: 0n,
                    userNetFlow: 0n,
                    nodeAllocation: 0n,
                    nodeNetFlow: 0n,
                },
                userSig: '0xUSRSIG' as Hex,
                nodeSig: '0xNODSIG' as Hex,
            };
            const params = {
                channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                candidateState,
                proofs: [],
            };

            jest.spyOn(stateModule, '_prepareAndSignChallengeState').mockResolvedValue({
                channelId: params.channelId,
                candidateState,
                proofs: [],
                challengerSig: mockSignature as Hex,
            });
            mockNitroService.challengeChannel.mockResolvedValue('0xCHL' as Hash);

            const tx = await client.challengeChannel(params);
            expect(mockNitroService.challengeChannel).toHaveBeenCalledWith(
                params.channelId,
                candidateState,
                [],
                mockSignature,
            );
            expect(tx).toBe('0xCHL');
        });

        test('failure throws error', async () => {
            const candidateState = {
                version: 1n,
                intent: StateIntent.OPERATE,
                metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
                homeState: {
                    chainId: 1n,
                    token: tokenAddress,
                    decimals: 18,
                    userAllocation: 90n,
                    userNetFlow: 0n,
                    nodeAllocation: 110n,
                    nodeNetFlow: 0n,
                },
                nonHomeState: {
                    chainId: 0n,
                    token: '0x0000000000000000000000000000000000000000' as Address,
                    decimals: 0,
                    userAllocation: 0n,
                    userNetFlow: 0n,
                    nodeAllocation: 0n,
                    nodeNetFlow: 0n,
                },
                userSig: '0xUSRSIG' as Hex,
                nodeSig: '0xNODSIG' as Hex,
            };
            const params = {
                channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                candidateState,
                proofs: [],
            };

            jest.spyOn(stateModule, '_prepareAndSignChallengeState').mockRejectedValue(new Error('fail'));
            await expect(client.challengeChannel(params)).rejects.toThrow();
        });
    });

    describe('closeChannel', () => {
        test('success', async () => {
            const finalState = {
                version: 2n,
                intent: StateIntent.FINALIZE,
                metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
                homeState: {
                    chainId: 1n,
                    token: tokenAddress,
                    decimals: 18,
                    userAllocation: 100n,
                    userNetFlow: 0n,
                    nodeAllocation: 100n,
                    nodeNetFlow: 0n,
                },
                nonHomeState: {
                    chainId: 0n,
                    token: '0x0000000000000000000000000000000000000000' as Address,
                    decimals: 0,
                    userAllocation: 0n,
                    userNetFlow: 0n,
                    nodeAllocation: 0n,
                    nodeNetFlow: 0n,
                },
                userSig: '0xUSRSIG' as Hex,
                nodeSig: '0xNODSIG' as Hex,
            };

            jest.spyOn(stateModule, '_prepareAndSignFinalState').mockResolvedValue({
                channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                finalState,
                proofs: [],
            });
            mockNitroService.closeChannel.mockResolvedValue('0xCLS' as Hash);

            const tx = await client.closeChannel({
                channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                finalState,
            });
            expect(stateModule._prepareAndSignFinalState).toHaveBeenCalledWith(expect.anything(), expect.any(Object));
            expect(mockNitroService.closeChannel).toHaveBeenCalledWith(
                '0x0000000000000000000000000000000000000000000000000000000000000001',
                finalState,
                [],
            );
            expect(tx).toBe('0xCLS');
        });

        test('failure throws error', async () => {
            const finalState = {
                version: 2n,
                intent: StateIntent.FINALIZE,
                metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
                homeState: {
                    chainId: 1n,
                    token: tokenAddress,
                    decimals: 18,
                    userAllocation: 100n,
                    userNetFlow: 0n,
                    nodeAllocation: 100n,
                    nodeNetFlow: 0n,
                },
                nonHomeState: {
                    chainId: 0n,
                    token: '0x0000000000000000000000000000000000000000' as Address,
                    decimals: 0,
                    userAllocation: 0n,
                    userNetFlow: 0n,
                    nodeAllocation: 0n,
                    nodeNetFlow: 0n,
                },
                userSig: '0xUSRSIG' as Hex,
                nodeSig: '0xNODSIG' as Hex,
            };

            jest.spyOn(stateModule, '_prepareAndSignFinalState').mockRejectedValue(new Error('fail'));
            await expect(
                client.closeChannel({
                    channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                    finalState,
                }),
            ).rejects.toThrow();
        });
    });

    describe('withdraw', () => {
        test('success', async () => {
            mockNitroService.withdraw.mockResolvedValue('0xWDL' as Hash);
            const tx = await client.withdraw(mockAccount.address, tokenAddress, 20n);
            expect(mockNitroService.withdraw).toHaveBeenCalledWith(mockAccount.address, tokenAddress, 20n);
            expect(tx).toBe('0xWDL');
        });

        test('failure throws error', async () => {
            mockNitroService.withdraw.mockRejectedValue(new Error('fail'));
            await expect(client.withdraw(mockAccount.address, tokenAddress, 20n)).rejects.toThrow();
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

    describe('getAccountBalance', () => {
        test('success', async () => {
            const balances = 42n;
            mockNitroService.getAccountBalance.mockResolvedValue(balances);
            const res = await client.getAccountBalance(brokerAddress, tokenAddress);
            expect(res).toEqual(balances);
            expect(mockNitroService.getAccountBalance).toHaveBeenCalledWith(brokerAddress, tokenAddress);
        });
    });

    describe('getChannelData', () => {
        test('success', async () => {
            const data: ChannelData = {
                definition: {
                    challengeDuration: 3600,
                    user: mockAccount.address,
                    node: brokerAddress,
                    nonce: 1n,
                    metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
                },
                status: ChannelStatus.INITIAL,
                challengeExpiry: 1234567890n,
                lastState: {
                    version: 0n,
                    intent: StateIntent.INITIALIZE,
                    metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
                    homeState: {
                        chainId: 1n,
                        token: tokenAddress,
                        decimals: 18,
                        userAllocation: 100n,
                        userNetFlow: 0n,
                        nodeAllocation: 100n,
                        nodeNetFlow: 0n,
                    },
                    nonHomeState: {
                        chainId: 0n,
                        token: '0x0000000000000000000000000000000000000000' as Address,
                        decimals: 0,
                        userAllocation: 0n,
                        userNetFlow: 0n,
                        nodeAllocation: 0n,
                        nodeNetFlow: 0n,
                    },
                    userSig: '0x' as Hex,
                    nodeSig: '0x' as Hex,
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

        test('failure throws error', async () => {
            mockErc20Service.approve.mockRejectedValue(new Error('fail'));
            await expect(client.approveTokens(tokenAddress, 30n)).rejects.toThrow();
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
