/**
 * @file Tests for src/client/state.ts
 */
import { describe, test, expect, beforeEach, jest } from '@jest/globals';
import { Address, Hex, zeroAddress, zeroHash } from 'viem';
import { _prepareAndSignInitialState, _prepareAndSignFinalState } from '../../../src/client/state';
import { Errors } from '../../../src/errors';
import { State, CreateChannelParams, StateIntent, CloseChannelParams, ChannelDefinition } from '../../../src/client/types';

describe('_prepareAndSignInitialState', () => {
    let deps: any;
    const nodeAddress = '0x5555555555555555555555555555555555555555' as Address;
    const userAddress = '0x1234567890123456789012345678901234567890' as Address;
    const tokenAddress = '0x4444444444444444444444444444444444444444' as Address;

    const definition: ChannelDefinition = {
        user: userAddress,
        node: nodeAddress,
        nonce: 1n,
        challengeDuration: 3600,
        metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
    };

    const initialState: State = {
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

    beforeEach(() => {
        deps = {
            account: { address: userAddress },
            addresses: {
                custody: '0x1111111111111111111111111111111111111111' as Address,
            },
            challengeDuration: 3600,
            chainId: 1,
        };
    });

    test('success with valid params', async () => {
        const params: CreateChannelParams = {
            definition,
            initialState,
        };
        const result = await _prepareAndSignInitialState(deps, params);

        expect(result.channelId).toBe(zeroHash);
        expect(result.initialState).toEqual(initialState);
    });

    test('throws if no definition', async () => {
        const params = {
            initialState,
        } as any;

        await expect(_prepareAndSignInitialState(deps, params)).rejects.toThrow(Errors.MissingParameterError);
    });

    test('throws if no initialState', async () => {
        const params = {
            definition,
        } as any;

        await expect(_prepareAndSignInitialState(deps, params)).rejects.toThrow(Errors.MissingParameterError);
    });
});

describe('_prepareAndSignFinalState', () => {
    let deps: any;
    const nodeAddress = '0x5555555555555555555555555555555555555555' as Address;
    const userAddress = '0x1234567890123456789012345678901234567890' as Address;
    const tokenAddress = '0x4444444444444444444444444444444444444444' as Address;
    const channelId = '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex;

    const finalState: State = {
        version: 1n,
        intent: StateIntent.FINALIZE,
        metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
        homeState: {
            chainId: 1n,
            token: tokenAddress,
            decimals: 18,
            userAllocation: 50n,
            userNetFlow: 0n,
            nodeAllocation: 50n,
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

    beforeEach(() => {
        deps = {
            account: { address: userAddress },
            addresses: {
                custody: '0x1111111111111111111111111111111111111111' as Address,
            },
            challengeDuration: 3600,
            chainId: 1,
        };
    });

    test('success with valid params', async () => {
        const params: CloseChannelParams = {
            channelId,
            finalState,
            proofs: [],
        };
        const result = await _prepareAndSignFinalState(deps, params);

        expect(result.channelId).toBe(channelId);
        expect(result.finalState).toEqual(finalState);
        expect(result.proofs).toEqual([]);
    });

    test('throws if no finalState', async () => {
        const params = {
            channelId,
        } as any;

        await expect(_prepareAndSignFinalState(deps, params)).rejects.toThrow(Errors.MissingParameterError);
    });
});
