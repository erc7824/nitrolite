/**
 * @file Tests for src/client/state.ts
 */
import { describe, test, expect, beforeEach, jest } from '@jest/globals';
import { Hex } from 'viem';
import { _prepareAndSignInitialState, _prepareAndSignFinalState } from '../../../src/client/state';
import * as utils from '../../../src/utils';
import { Errors } from '../../../src/errors';
import { Channel, CreateChannelParams, State, StateIntent } from '../../../src/client/types';
import { channel } from 'diagnostics_channel';

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
    let deps: any;
    let defaultChannel: Channel;
    let defaultState: State;
    const guestAddress = '0xGUEST' as Hex;
    const tokenAddress = '0xTOKEN' as Hex;
    const adjudicatorAddress = '0xADJ' as Hex;
    const challengeDuration = 123n;

    beforeEach(() => {
        deps = {
            account: { address: '0xOWNER' as Hex },
            stateWalletClient: {
                account: { address: '0xOWNER' as Hex },
                signMessage: async (_: string) => 'walletSig',
            },
            addresses: {
                guestAddress,
                adjudicator: adjudicatorAddress,
            },
            challengeDuration,
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
                { destination: deps.account.address, token: tokenAddress, amount: 10n },
                { destination: guestAddress, token: tokenAddress, amount: 20n },
            ],
            version: 0n,
            sigs: [],
        };
    });

    test('success with explicit stateData', async () => {
        const params: CreateChannelParams = {
            channel: defaultChannel,
            initialState: defaultState,
        };
        const { initialState, channelId } = await _prepareAndSignInitialState(deps, params);

        // channelId is stubbed
        expect(channelId).toBe('cid');
        // State fields
        expect(initialState).toEqual({
            data: '0xcustomData',
            intent: StateIntent.INITIALIZE,
            allocations: [
                { destination: deps.account.address, token: tokenAddress, amount: 10n },
                { destination: guestAddress, token: tokenAddress, amount: 20n },
            ],
            version: 0n,
            sigs: ['accSig'],
        });
        // Signs the state
        expect(utils.signState).toHaveBeenCalledWith(
            'cid',
            {
                data: '0xcustomData',
                intent: StateIntent.INITIALIZE,
                allocations: expect.any(Array),
                version: 0n,
                sigs: [],
            },
            deps.stateWalletClient.signMessage,
        );
    });

    test('throws if no adjudicator', async () => {
        const localChannel = { ...defaultChannel, adjudicator: undefined } as any;

        await expect(
            _prepareAndSignInitialState(deps, {
                channel: localChannel,
                initialState: defaultState,
            }),
        ).rejects.toThrow(Errors.MissingParameterError);
    });

    test('throws if bad allocations length', async () => {
        const localState = { ...defaultState, allocations: [] } as any;

        await expect(
            _prepareAndSignInitialState(deps, {
                channel: defaultChannel,
                initialState: localState,
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

    beforeEach(() => {
        deps = {
            stateWalletClient: {
                account: { address: '0xOWNER' as Hex },
                signMessage: async (_: string) => 'walletSig2',
            },
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
        expect(utils.signState).toHaveBeenCalledWith(
            'cid',
            {
                data: 'finalData',
                intent: StateIntent.FINALIZE,
                allocations,
                version,
                sigs: [],
            },
            deps.stateWalletClient.signMessage,
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
