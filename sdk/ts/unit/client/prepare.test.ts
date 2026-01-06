import { jest } from '@jest/globals';
jest.mock('../../../src/client/state', () => ({
    _prepareAndSignInitialState: jest.fn(),
    _prepareAndSignFinalState: jest.fn(),
    _prepareAndSignChallengeState: jest.fn(),
}));
import { describe, test, expect, beforeEach } from '@jest/globals';
import { Hex, zeroAddress } from 'viem';
import { NitroliteTransactionPreparer } from '../../../src/client/prepare';
import {
    _prepareAndSignInitialState,
    _prepareAndSignFinalState,
    _prepareAndSignChallengeState,
} from '../../../src/client/state';
import { NitroliteService, Erc20Service } from '../../../src/client/services';
import { ContractAddresses } from '../../../src/abis';
import * as Errors from '../../../src/errors';
import { CreateChannelParams, CheckpointChannelParams, Allocation, StateIntent } from '../../../src/client/types';

// TODO: remove ts-ignore
describe('NitroliteTransactionPreparer', () => {
    const tokenAddress = '0x4444444444444444444444444444444444444444' as const;
    const custody = '0x1111111111111111111111111111111111111111' as const;
    const accountAddress = '0x1234567890123456789012345678901234567890' as const;
    const guestAddress = '0x3333333333333333333333333333333333333333' as const;
    const adjudicator = '0x2222222222222222222222222222222222222222' as const;

    const addresses: ContractAddresses = {
        custody,
        guestAddress,
        adjudicator,
    };
    const account = { address: accountAddress };

    let mockNitro: jest.Mocked<NitroliteService>;
    let mockERC20: jest.Mocked<Erc20Service>;
    let deps: any;
    let prep: NitroliteTransactionPreparer;

    const mockSignMessage = jest
        .fn()
        .mockResolvedValue(
            '0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1b',
        );

    beforeEach(() => {
        mockNitro = {
            prepareDeposit: jest.fn(),
            prepareCreateChannel: jest.fn(),
            prepareDepositAndCreateChannel: jest.fn(),
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
            walletClient: { signMessage: mockSignMessage },
            stateWalletClient: { signMessage: mockSignMessage },
            challengeDuration: 100n,
        };
        prep = new NitroliteTransactionPreparer(deps);
        // Mock _prepareAndSignChallengeState for challenge tests
        (_prepareAndSignChallengeState as jest.Mock) = jest
            .fn()
            .mockResolvedValue({ challengerSig: { r: '0x1', s: '0x2', v: 27 } });
    });

    describe('prepareDepositTransactions', () => {
        test('ERC20 no approval needed', async () => {
            mockERC20.getTokenAllowance.mockResolvedValue(100n);
            mockNitro.prepareDeposit.mockResolvedValue({ to: '0xA', data: '0xA' } as any);
            const txs = await prep.prepareDepositTransactions(tokenAddress, 50n);
            expect(mockERC20.getTokenAllowance).toHaveBeenCalledWith(tokenAddress, accountAddress, custody);
            expect(txs).toHaveLength(1);
        });

        test('ERC20 needs approval', async () => {
            mockERC20.getTokenAllowance.mockResolvedValue(10n);
            mockERC20.prepareApprove.mockResolvedValue({ to: '0xA', data: '0xA' } as any);
            mockNitro.prepareDeposit.mockResolvedValue({ to: '0xD', data: '0xD' } as any);
            const txs = await prep.prepareDepositTransactions(tokenAddress, 50n);
            expect(mockERC20.prepareApprove).toHaveBeenCalledWith(tokenAddress, custody, 50n);
            expect(txs).toHaveLength(2);
        });

        test('skip approval for ETH', async () => {
            mockNitro.prepareDeposit.mockResolvedValue({ to: '0xD', data: '0xD' } as any);
            const txs = await prep.prepareDepositTransactions(zeroAddress, 20n);
            expect(mockERC20.getTokenAllowance).not.toHaveBeenCalled();
            expect(txs).toHaveLength(1);
        });

        test('prepareDeposit error wraps', async () => {
            mockERC20.getTokenAllowance.mockResolvedValue(100n);
            mockNitro.prepareDeposit.mockRejectedValue(new Error('fail'));
            await expect(prep.prepareDepositTransactions(tokenAddress, 10n)).rejects.toThrow(Errors.ContractCallError);
        });
    });

    describe('prepareCreateChannelTransaction', () => {
        const params: CreateChannelParams = {
            channel: {
                participants: [accountAddress, guestAddress],
                adjudicator: '0xadj',
                challenge: 123n,
                nonce: 999n,
            },
            initialState: {
                data: '0xcustomData',
                intent: StateIntent.INITIALIZE,
                allocations: [
                    { destination: accountAddress, token: tokenAddress, amount: 10n },
                    { destination: guestAddress, token: tokenAddress, amount: 20n },
                ],
                version: 0n,
                sigs: [],
            },
        };
        test('success', async () => {
            const fake = { channel: {}, initialState: {} };
            // @ts-ignore
            (_prepareAndSignInitialState as jest.Mock).mockResolvedValue(fake);
            mockNitro.prepareCreateChannel.mockResolvedValue({ to: '0xC', data: '0xC' } as any);
            const tx = await prep.prepareCreateChannelTransaction(params);
            expect(_prepareAndSignInitialState).toHaveBeenCalledWith(deps, params);
            expect(tx).toEqual({ to: '0xC', data: '0xC' });
        });

        test('wraps non-NitroliteError', async () => {
            // @ts-ignore
            (_prepareAndSignInitialState as jest.Mock).mockRejectedValue(new Error('oops'));
            await expect(prep.prepareCreateChannelTransaction(params)).rejects.toThrow(Errors.ContractCallError);
        });

        test('rethrows NitroliteError from state', async () => {
            const nitroliteError = new Errors.MissingParameterError('x');
            // @ts-ignore
            (_prepareAndSignInitialState as jest.Mock).mockRejectedValueOnce(nitroliteError);
            await expect(prep.prepareCreateChannelTransaction(params)).rejects.toBe(nitroliteError);
        });
    });

    describe('prepareDepositAndCreateChannelTransactions', () => {
        const params: CreateChannelParams = {
            channel: {
                participants: [accountAddress, guestAddress],
                adjudicator: '0xadj',
                challenge: 123n,
                nonce: 999n,
            },
            initialState: {
                data: '0xcustomData',
                intent: StateIntent.INITIALIZE,
                allocations: [
                    { destination: accountAddress, token: tokenAddress, amount: 10n },
                    { destination: guestAddress, token: tokenAddress, amount: 20n },
                ],
                version: 0n,
                sigs: [],
            },
        };
        test('combines flows', async () => {
            // Test case where approval is needed
            mockERC20.getTokenAllowance.mockResolvedValue(5n); // Less than deposit amount
            mockERC20.prepareApprove.mockResolvedValue({ to: '0xA', data: '0xA' } as any);
            // @ts-ignore
            (_prepareAndSignInitialState as jest.Mock).mockResolvedValue({ channel: {}, initialState: {} });
            mockNitro.prepareDepositAndCreateChannel.mockResolvedValue({ to: '0xD', data: '0xD' });

            const all = await prep.prepareDepositAndCreateChannelTransactions(tokenAddress, 10n, params);

            expect(all).toHaveLength(2);
            expect(all[0]).toEqual({ to: '0xA', data: '0xA' }); // Approval transaction
            expect(all[1]).toEqual({ to: '0xD', data: '0xD' }); // Deposit and create channel transaction

            // Verify the correct calls were made
            expect(mockERC20.getTokenAllowance).toHaveBeenCalledWith(tokenAddress, accountAddress, custody);
            expect(mockERC20.prepareApprove).toHaveBeenCalledWith(tokenAddress, custody, 10n);
            expect(_prepareAndSignInitialState).toHaveBeenCalledWith(deps, params);
        });

        test('rethrows NitroliteError from deposit prepare', async () => {
            const ne = new Errors.MissingParameterError('d');
            mockERC20.getTokenAllowance.mockResolvedValue(100n);
            // @ts-ignore
            (_prepareAndSignInitialState as jest.Mock).mockResolvedValue({ channel: {}, initialState: {} });
            mockNitro.prepareDepositAndCreateChannel.mockRejectedValueOnce(ne);
            await expect(prep.prepareDepositAndCreateChannelTransactions(tokenAddress, 10n, {} as any)).rejects.toBe(
                ne,
            );
        });

        test('rethrows NitroliteError from createChannel prepare', async () => {
            const ne = new Errors.MissingParameterError('y');
            mockERC20.getTokenAllowance.mockResolvedValue(100n);
            // @ts-ignore
            (_prepareAndSignInitialState as jest.Mock).mockResolvedValue({ channel: {}, initialState: {} });
            mockNitro.prepareDepositAndCreateChannel.mockRejectedValueOnce(ne);
            await expect(prep.prepareDepositAndCreateChannelTransactions(tokenAddress, 10n, {} as any)).rejects.toBe(
                ne,
            );
        });
    });

    describe('prepareCheckpointChannelTransaction', () => {
        const good: CheckpointChannelParams = {
            channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
            candidateState: { sigs: [1, 2] } as any,
            proofStates: [],
        };
        const bad: CheckpointChannelParams = {
            channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
            candidateState: { sigs: [1] } as any,
        };
        test('valid', async () => {
            mockNitro.prepareCheckpoint.mockResolvedValue({ to: '0xK', data: '0xK' } as any);
            await expect(prep.prepareCheckpointChannelTransaction(good)).resolves.toEqual({
                to: '0xK',
                data: '0xK',
            } as any);
        });
        test('invalid sigs', async () => {
            await expect(prep.prepareCheckpointChannelTransaction(bad)).rejects.toThrow(Errors.InvalidParameterError);
        });
    });

    describe('prepareChallengeChannelTransaction', () => {
        test('success and wrap', async () => {
            mockNitro.prepareChallenge.mockResolvedValue({ to: '0xH', data: '0xH' } as any);
            await expect(
                // @ts-ignore
                prep.prepareChallengeChannelTransaction({
                    channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                    candidateState: {} as any,
                    proofStates: [],
                }),
            ).resolves.toBeDefined();
        });
    });

    describe('prepareCloseChannelTransaction', () => {
        test('success', async () => {
            // @ts-ignore
            (_prepareAndSignFinalState as jest.Mock).mockResolvedValue({
                finalStateWithSigs: {},
                channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
            });
            mockNitro.prepareClose.mockResolvedValue({ to: '0xX', data: '0xX' } as any);
            await expect(
                prep.prepareCloseChannelTransaction({
                    finalState: {
                        data: '0xA' as any,
                        channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                        allocations: [
                            { destination: '0x0' as Hex, token: tokenAddress, amount: 10n },
                            { destination: '0x0' as Hex, token: tokenAddress, amount: 10n },
                        ] as [Allocation, Allocation],
                        version: 0n,
                        serverSignature: [] as any,
                        intent: StateIntent.FINALIZE,
                    },
                }),
            ).resolves.toEqual({ to: '0xX', data: '0xX' });
        });

        test('wraps final state error', async () => {
            // @ts-ignore
            (_prepareAndSignFinalState as jest.Mock).mockRejectedValueOnce(new Error('state fail'));
            await expect(
                prep.prepareCloseChannelTransaction({
                    finalState: {
                        data: '0xA' as any,
                        channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                        allocations: [
                            { destination: '0x0' as Hex, token: tokenAddress, amount: 10n },
                            { destination: '0x0' as Hex, token: tokenAddress, amount: 10n },
                        ] as [Allocation, Allocation],
                        version: 0n,
                        serverSignature: [] as any,
                        intent: StateIntent.FINALIZE,
                    },
                }),
            ).rejects.toThrow(Errors.ContractCallError);
        });

        test('wraps non-NitroliteError from close', async () => {
            // @ts-ignore
            (_prepareAndSignFinalState as jest.Mock).mockResolvedValue({
                finalStateWithSigs: {},
                channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
            });
            const err = new Error('oops');
            mockNitro.prepareClose.mockRejectedValueOnce(err);
            await expect(
                prep.prepareCloseChannelTransaction({
                    finalState: {
                        data: '0xA' as any,
                        channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                        allocations: [
                            { destination: '0x0' as Hex, token: tokenAddress, amount: 10n },
                            { destination: '0x0' as Hex, token: tokenAddress, amount: 10n },
                        ] as [Allocation, Allocation],
                        version: 0n,
                        serverSignature: [] as any,
                        intent: StateIntent.FINALIZE,
                    },
                }),
            ).rejects.toThrow(Errors.ContractCallError);
        });

        test('rethrows NitroliteError from close', async () => {
            // @ts-ignore
            (_prepareAndSignFinalState as jest.Mock).mockResolvedValue({
                finalStateWithSigs: {},
                channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
            });
            const ne = new Errors.MissingParameterError('z');
            mockNitro.prepareClose.mockRejectedValueOnce(ne);
            await expect(
                prep.prepareCloseChannelTransaction({
                    finalState: {
                        data: '0xA' as any,
                        channelId: '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex,
                        allocations: [
                            { destination: '0x0' as Hex, token: tokenAddress, amount: 10n },
                            { destination: '0x0' as Hex, token: tokenAddress, amount: 10n },
                        ] as [Allocation, Allocation],
                        version: 0n,
                        serverSignature: [] as any,
                        intent: StateIntent.FINALIZE,
                    },
                }),
            ).rejects.toBe(ne);
        });
    });

    describe('prepareWithdrawalTransaction', () => {
        test('success', async () => {
            mockNitro.prepareWithdraw.mockResolvedValue({ to: '0xW', data: '0xW' } as any);
            await expect(prep.prepareWithdrawalTransaction(tokenAddress, 5n)).resolves.toEqual({
                to: '0xW',
                data: '0xW',
            } as any);
        });
    });

    describe('prepareApproveTokensTransaction', () => {
        test('ETH error', async () => {
            await expect(prep.prepareApproveTokensTransaction(zeroAddress, 1n)).rejects.toThrow(
                Errors.InvalidParameterError,
            );
        });
        test('success', async () => {
            mockERC20.prepareApprove.mockResolvedValue({ to: '0xA', data: '0xA' } as any);
            await expect(prep.prepareApproveTokensTransaction(tokenAddress, 7n)).resolves.toEqual({
                to: '0xA',
                data: '0xA',
            } as any);
        });
        test('rethrows NitroliteError from prepareApproveTokensTransaction', async () => {
            const ne = new Errors.MissingParameterError('a');
            mockERC20.prepareApprove.mockRejectedValueOnce(ne);
            await expect(prep.prepareApproveTokensTransaction(tokenAddress, 7n)).rejects.toBe(ne);
        });
    });
});
