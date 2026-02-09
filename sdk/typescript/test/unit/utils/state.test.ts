import { describe, test, expect, jest } from '@jest/globals';
import { getStateHash, verifySignature } from '../../../src/utils/state';
import { type UnsignedStateV1, type Signature, StateIntent } from '../../../src/client/types';
import { Hex, Address, recoverMessageAddress, encodeAbiParameters, keccak256 } from 'viem';

jest.mock('viem', () => ({
    encodeAbiParameters: jest.fn(() => '0xencoded'),
    keccak256: jest.fn(() => '0xhash'),
    recoverMessageAddress: jest.fn(async () => '0xSignerAddress'),
}));

beforeAll(() => {
    jest.spyOn(console, 'error').mockImplementation(() => {});
});
afterAll(() => {
    (console.error as jest.Mock).mockRestore();
});

describe('getStateHash', () => {
    test('encodes state and hashes', () => {
        const channelId = '0xChannelId' as Hex;
        const state: UnsignedStateV1 = {
            version: 1n,
            intent: StateIntent.INITIALIZE,
            metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
            homeState: {
                chainId: 1n,
                token: '0xT' as Address,
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
        };
        const hash = getStateHash(channelId, state);

        const ledgerComponents = [
            { name: 'chainId', type: 'uint64' },
            { name: 'token', type: 'address' },
            { name: 'decimals', type: 'uint8' },
            { name: 'userAllocation', type: 'uint256' },
            { name: 'userNetFlow', type: 'int256' },
            { name: 'nodeAllocation', type: 'uint256' },
            { name: 'nodeNetFlow', type: 'int256' },
        ];

        expect(encodeAbiParameters).toHaveBeenCalledWith(
            [
                { name: 'channelId', type: 'bytes32' },
                { name: 'version', type: 'uint64' },
                { name: 'intent', type: 'uint8' },
                { name: 'metadata', type: 'bytes32' },
                {
                    name: 'homeState',
                    type: 'tuple',
                    components: ledgerComponents,
                },
                {
                    name: 'nonHomeState',
                    type: 'tuple',
                    components: ledgerComponents,
                },
            ],
            [
                channelId,
                state.version,
                state.intent,
                state.metadata,
                {
                    chainId: state.homeState.chainId,
                    token: state.homeState.token,
                    decimals: state.homeState.decimals,
                    userAllocation: state.homeState.userAllocation,
                    userNetFlow: state.homeState.userNetFlow,
                    nodeAllocation: state.homeState.nodeAllocation,
                    nodeNetFlow: state.homeState.nodeNetFlow,
                },
                {
                    chainId: state.nonHomeState.chainId,
                    token: state.nonHomeState.token,
                    decimals: state.nonHomeState.decimals,
                    userAllocation: state.nonHomeState.userAllocation,
                    userNetFlow: state.nonHomeState.userNetFlow,
                    nodeAllocation: state.nonHomeState.nodeAllocation,
                    nodeNetFlow: state.nonHomeState.nodeNetFlow,
                },
            ],
        );
        expect(keccak256).toHaveBeenCalledWith('0xencoded');
        expect(hash).toBe('0xhash');
    });
});

describe('verifySignature', () => {
    const channelId = '0xChannelId' as Hex;
    const state: UnsignedStateV1 = {
        version: 1n,
        intent: StateIntent.INITIALIZE,
        metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
        homeState: {
            chainId: 1n,
            token: '0xT' as Address,
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
    };
    const stateHash = getStateHash(channelId, state);
    const signature: Signature = "0xr0xs1b" as Signature;
    const expectedSigner = '0xSignerAddress' as Address;

    test('recovers address', async () => {
        const result = await verifySignature(channelId, state, signature, expectedSigner);
        expect(recoverMessageAddress).toHaveBeenCalledWith({
            message: { raw: stateHash },
            signature: signature as Hex,
        });
        expect(result).toBe(true);
    });

    test('returns false on recover error', async () => {
        const viemMock = jest.requireMock('viem');
        // @ts-ignore
        viemMock.recoverMessageAddress.mockRejectedValueOnce(new Error('fail'));
        const res = await verifySignature(channelId, state, signature, expectedSigner);
        expect(res).toBe(false);
    });

    test('returns false on mismatched address', async () => {
        const res = await verifySignature(channelId, state, signature, '0xOther' as Address);
        expect(res).toBe(false);
    });
});
