import { describe, test, expect, jest } from '@jest/globals';
import { getStateHash, verifySignature } from '../../../src/utils/state';
import { type State, type Signature, type Allocation, StateIntent } from '../../../src/client/types';
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
        const state: State = {
            data: '0xdata' as Hex,
            version: 1n,
            intent: StateIntent.INITIALIZE,
            allocations: [
                { destination: '0xA' as Address, token: '0xT' as Address, amount: 10n },
                { destination: '0xB' as Address, token: '0xT' as Address, amount: 10n },
            ] as [Allocation, Allocation],
            sigs: [], // sigs not used by getStateHash
        };
        const hash = getStateHash(channelId, state);
        expect(encodeAbiParameters).toHaveBeenCalledWith(
            [
                { name: 'channelId', type: 'bytes32' },
                { name: 'intent', type: 'uint8' },
                { name: 'version', type: 'uint256' },
                { name: 'data', type: 'bytes' },
                {
                    name: 'allocations',
                    type: 'tuple[]',
                    components: [
                        { name: 'destination', type: 'address' },
                        { name: 'token', type: 'address' },
                        { name: 'amount', type: 'uint256' },
                    ],
                },
            ],
            [channelId, state.intent, state.version, state.data, state.allocations],
        );
        expect(keccak256).toHaveBeenCalledWith('0xencoded');
        expect(hash).toBe('0xhash');
    });
});

describe('verifySignature', () => {
    const channelId = '0xChannelId' as Hex;
    const state: State = {
        data: '0xdata' as Hex,
        intent: StateIntent.INITIALIZE,
        allocations: [
            { destination: '0xA' as Address, token: '0xT' as Address, amount: 10n },
            { destination: '0xB' as Address, token: '0xT' as Address, amount: 20n },
        ] as [Allocation, Allocation],
        version: 0n,
        sigs: [],
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
