/**
 * @file Tests for src/client/state.ts
 */
import { describe, test, expect, beforeEach, jest } from '@jest/globals';
import { Hex } from 'viem';
import { _prepareAndSignInitialState, _prepareAndSignFinalState } from '../../../src/client/state';
import * as utils from '../../../src/utils';
import { Errors } from '../../../src/errors';
import { State, StateIntent } from '../../../src/client/types';

// Mock utils
jest.mock('../../../src/utils', () => ({
    generateChannelNonce: jest.fn(() => 999n),
    getChannelId: jest.fn(() => 'cid' as any),
    getStateHash: jest.fn(() => 'hsh'),
    signState: jest.fn(async () => 'accSig'),
    encoders: { numeric: jest.fn(() => 'encData') },
    removeQuotesFromRS: jest.fn((s: string) => s.replace(/"/g, '')),
}));

describe('_prepareAndSignInitialState', () => {
    let deps;
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
            addresses: {
                guestAddress,
                adjudicator: adjudicatorAddress,
            },
            challengeDuration,
        };
    });

    test('success with explicit stateData', async () => {
        const params = {
            initialAllocationAmounts: [10n, 20n],
            stateData: 'customData',
        };
        const { channel, initialState, channelId } = await _prepareAndSignInitialState(
            tokenAddress,
            deps,
            params as any,
        );

        // Channel fields
        expect(utils.generateChannelNonce).toHaveBeenCalledWith(deps.account.address);
        expect(channel).toEqual({
            participants: [deps.account.address, guestAddress],
            adjudicator: adjudicatorAddress,
            challenge: challengeDuration,
            nonce: 999n,
        });
        // channelId is stubbed
        expect(channelId).toBe('cid');
        // State fields
        expect(initialState).toEqual({
            data: 'customData',
            intent: StateIntent.INITIALIZE,
            allocations: [
                { destination: deps.account.address, token: tokenAddress, amount: 10n },
                { destination: guestAddress, token: tokenAddress, amount: 20n },
            ],
            version: 0n,
            sigs: ['accSig'],
        });
        // Signs the state
        expect(stateSigner.signState).toHaveBeenCalledWith(
            'cid',
            {
                data: 'customData',
                intent: StateIntent.INITIALIZE,
                allocations: expect.any(Array),
                version: 0n,
                sigs: [],
            }
        );
    });

    test('throws if no adjudicator', async () => {
        deps.addresses.adjudicator = undefined;
        await expect(
            _prepareAndSignInitialState(tokenAddress, deps, {
                initialAllocationAmounts: [1n, 2n],
                stateData: '0xdata',
            }),
        ).rejects.toThrow(Errors.MissingParameterError);
    });

    test('throws if bad allocations length', async () => {
        await expect(
            _prepareAndSignInitialState(tokenAddress, deps, {
                initialAllocationAmounts: [1n],
                stateData: 'd',
            } as any),
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
            addresses: {
                /* not used */
            },
            account: {
                /* not used */
            },
            challengeDuration: 0,
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
