/**
 * @file Tests for src/client/state.ts
 */
import { describe, test, expect, beforeEach, jest } from '@jest/globals';
import { Hex } from 'viem';
import { _prepareAndSignInitialState, _prepareAndSignFinalState } from '../../../src/client/state';
import { Errors } from '../../../src/errors';
import { State, Channel, CreateChannelParams, StateIntent } from '../../../src/client/types';

// Mock utils
jest.mock('../../../src/utils', () => ({
    generateChannelNonce: jest.fn(() => 999n),
    getChannelId: jest.fn(() => 'cid' as any),
    getStateHash: jest.fn(() => 'hsh'),
    getPackedState: jest.fn(() => '0xpacked' as Hex),
    signState: jest.fn(async () => 'accSig'),
    encoders: { numeric: jest.fn(() => 'encData') },
    removeQuotesFromRS: jest.fn((s: string) => s.replace(/"/g, '')),
}));

describe('_prepareAndSignInitialState', () => {
    let deps: any;
    let defaultChannel: Channel;
    let defaultState: State;
    const guestAddress = '0xGUEST' as Hex;
    const tokenAddress = '0xTOKEN' as Hex;
    const adjudicatorAddress = '0xADJ' as Hex;
    const challengeDuration = BigInt(123);
    const stateSigner = {
        getAddress: jest.fn(() => '0xOWNER' as Hex),
        signState: jest.fn(async (_1: Hex, _2: State) => 'accSig'),
        signRawMessage: jest.fn(async (_: Hex) => 'accSig'),
    };

    beforeEach(() => {
        deps = {
            account: { address: '0xOWNER' as Hex },
            stateSigner,
            walletClient: {
                account: { address: '0xWALLET' as Hex },
                signMessage: stateSigner.signRawMessage,
            },
            addresses: {
                guestAddress,
                adjudicator: adjudicatorAddress,
            },
            challengeDuration,
            chainId: 1,
        };

        defaultChannel = {
            participants: [deps.account.address, guestAddress],
            adjudicator: adjudicatorAddress,
            challenge: challengeDuration,
            nonce: 999n,
        };

        defaultState = {
            data: '0xcustomData',
            intent: StateIntent.INITIALIZE,
            allocations: [
                // NOTE: first allocation amount is zero
                { destination: deps.account.address, token: tokenAddress, amount: 0n },
                { destination: guestAddress, token: tokenAddress, amount: 20n },
            ],
            version: 0n,
            sigs: [],
        };
    });

    test('success with explicit stateData', async () => {
        const params: CreateChannelParams = {
            channel: defaultChannel,
            unsignedInitialState: defaultState,
            serverSignature: '0xSRVSIG',
        };
        const { initialState, channelId } = await _prepareAndSignInitialState(deps, params);

        // channelId is stubbed
        expect(channelId).toBe('cid');
        // State fields
        expect(initialState).toEqual({
            data: '0xcustomData',
            intent: StateIntent.INITIALIZE,
            allocations: [
                { destination: deps.account.address, token: tokenAddress, amount: 0n },
                { destination: guestAddress, token: tokenAddress, amount: 20n },
            ],
            version: 0n,
            sigs: ['accSig', '0xSRVSIG'],
        });
        // Signs the state
        expect(stateSigner.signState).toHaveBeenCalledWith(
            'cid',
            {
                data: '0xcustomData',
                intent: StateIntent.INITIALIZE,
                allocations: expect.any(Array),
                version: 0n,
                sigs: [],
            }
        );
    });

    test('throws if no adjudicator', async () => {
        const localChannel = { ...defaultChannel, adjudicator: undefined } as any;

        await expect(
            _prepareAndSignInitialState(deps, {
                channel: localChannel,
                unsignedInitialState: defaultState,
                serverSignature: '0xSRVSIG',
            }),
        ).rejects.toThrow(Errors.MissingParameterError);
    });

    test('throws if bad allocations length', async () => {
        const localState = { ...defaultState, allocations: [] } as any;

        await expect(
            _prepareAndSignInitialState(deps, {
                channel: defaultChannel,
                unsignedInitialState: localState,
                serverSignature: '0xSRVSIG',
            }),
        ).rejects.toThrow(Errors.InvalidParameterError);
    });

    test('throws if first allocation amount is NOT zero', async () => {
        const localState = { ...defaultState, allocations: [{ ...defaultState.allocations[0], amount: 1n }] } as any;

        await expect(
            _prepareAndSignInitialState(deps, {
                channel: defaultChannel,
                unsignedInitialState: localState,
                serverSignature: '0xSRVSIG',
            }),
        ).rejects.toThrow(Errors.InvalidParameterError);
    });
});

describe('_prepareAndSignFinalState', () => {
    let deps: any;
    const serverSig = 'srvSig';
    const channelIdArg = 'cid' as Hex;
    const allocations = [{ destination: '0xA' as Hex, token: '0xT' as Hex, amount: 5n }];
    const version = 7n;
    const stateSigner = {
        getAddress: jest.fn(async () => '0xOWNER' as Hex),
        signState: jest.fn(async (_1: Hex, _2: State) => 'accSig'),
        signRawMessage: jest.fn(async (_: Hex) => 'accSig'),
    };

    beforeEach(() => {
        deps = {
            stateSigner,
            walletClient: {
                account: { address: '0xWALLET' as Hex },
                signMessage: stateSigner.signRawMessage,
            },
            nitroliteService: {
                getChannelData: jest.fn(async () => ({
                    channel: {
                        participants: ['0xOWNER' as Hex, '0xGUEST' as Hex],
                        adjudicator: '0xADJ' as Hex,
                        challenge: 123n,
                        nonce: 999n,
                    },
                    status: 0,
                    wallets: ['0xW1' as Hex, '0xW2' as Hex],
                    challengeExpiry: 0n,
                    lastValidState: {} as any,
                })),
            },
            addresses: {
                /* not used */
            },
            account: {
                /* not used */
            },
            challengeDuration: 0,
            chainId: 1,
        };
    });

    test('success with explicit stateData', async () => {
        const params = {
            stateData: 'finalData',
            finalState: {
                intent: StateIntent.FINALIZE,
                channelId: channelIdArg,
                allocations,
                version,
                serverSignature: serverSig,
            },
        };
        const { finalStateWithSigs, channelId } = await _prepareAndSignFinalState(deps, params as any);

        expect(channelId).toBe(channelIdArg);
        // Data and allocations
        expect(finalStateWithSigs).toEqual({
            data: 'finalData',
            intent: StateIntent.FINALIZE,
            allocations,
            version,
            sigs: ['accSig', 'srvSig'],
        });
        expect(stateSigner.signState).toHaveBeenCalledWith(
            'cid',
            {
                data: 'finalData',
                intent: StateIntent.FINALIZE,
                allocations,
                version,
                sigs: [],
            }
        );
    });

    test('throws if no stateData', async () => {
        const params = {
            stateData: undefined,
            finalState: {
                channelId: channelIdArg,
                allocations,
                version,
                serverSignature: serverSig,
            },
        };
        await expect(_prepareAndSignFinalState(deps, params as any)).rejects.toThrow(Errors.MissingParameterError);
    });
});
