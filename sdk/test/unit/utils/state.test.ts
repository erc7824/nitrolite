import { describe, test, expect, jest } from '@jest/globals';
import { getStateHash, signState, removeQuotesFromRS, verifySignature } from '../../../src/utils/state';
import { type State, type Signature, type Allocation, StateIntent } from '../../../src/client/types';
import { Hex, Address, recoverMessageAddress, parseSignature, encodeAbiParameters, keccak256 } from 'viem';

jest.mock('viem', () => ({
    encodeAbiParameters: jest.fn(() => '0xencoded'),
    keccak256: jest.fn(() => '0xhash'),
    parseSignature: jest.fn(() => ({ r: '0xr', s: '0xs', v: 27 })),
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

describe('signState', () => {
    const fakeHash = '0xstatehash' as Hex;
    const signer = jest.fn(async ({ message }) => {
        if (message.raw === fakeHash) return '0xr0xs1b';
        throw new Error('sign fail');
    });

    test('successfully signs and parses signature', async () => {
        // @ts-ignore
        const sig = await signState(fakeHash, signer);
        expect(signer).toHaveBeenCalledWith({ message: { raw: fakeHash } });
        expect(parseSignature).toHaveBeenCalledWith('0xr0xs1b');
        expect(sig).toEqual({ r: '0xr', s: '0xs', v: 27 });
    });

    test('throws if parseSignature yields no v', async () => {
        const viemMock = jest.requireMock('viem');
        // @ts-ignore
        viemMock.parseSignature.mockReturnValueOnce({ r: '0xr', s: '0xs' });
        // @ts-ignore
        await expect(signState(fakeHash, signer)).rejects.toThrow(
            /Invalid signature format: missing v or yParity value/,
        );
    });

    test('throws on signer error', async () => {
        const badSigner = jest.fn(async () => {
            throw new Error('bad');
        });
        await expect(signState(fakeHash, badSigner)).rejects.toThrow(/Failed to sign state hash: bad/);
    });
});

describe('removeQuotesFromRS', () => {
    test('removes surrounding quotes from r and s', () => {
        const input: any = { r: '"0xr"', s: '"0xs"', v: 27 };
        const out = removeQuotesFromRS(input);
        expect(out).toEqual({ r: '0xr', s: '0xs', v: 27 });
    });
    test('leaves values without quotes untouched', () => {
        const input: Signature = { r: '0xr', s: '0xs', v: 28 };
        expect(removeQuotesFromRS(input)).toEqual(input);
    });
});

describe('verifySignature', () => {
    const stateHash = '0xstate' as Hex;
    const signature: Signature = { r: '0xr', s: '0xs', v: 0 }; // v < 27
    const expectedSigner = '0xSignerAddress' as Address;

    test('normalizes v and recovers address', async () => {
        const result = await verifySignature(stateHash, signature, expectedSigner);
        // vNormalized = 0 + 27 = 27 -> hex "1b"
        const sigHex = `${signature.r}${signature.s.slice(2)}1b`;
        expect(recoverMessageAddress).toHaveBeenCalledWith({
            message: { raw: stateHash },
            signature: sigHex as Hex,
        });
        expect(result).toBe(true);
    });

    test('handles v already >=27', async () => {
        const sig2: Signature = { r: '0xr2', s: '0xs2', v: 28 };
        const result = await verifySignature(stateHash, sig2, expectedSigner);
        const sigHex2 = `${sig2.r}${sig2.s.slice(2)}1c`; // 28 hex
        expect(recoverMessageAddress).toHaveBeenCalledWith({
            message: { raw: stateHash },
            signature: sigHex2 as Hex,
        });
        expect(result).toBe(true);
    });

    test('returns false on recover error', async () => {
        const viemMock = jest.requireMock('viem');
        // @ts-ignore
        viemMock.recoverMessageAddress.mockRejectedValueOnce(new Error('fail'));
        const res = await verifySignature(stateHash, signature, expectedSigner);
        expect(res).toBe(false);
    });

    test('returns false on mismatched address', async () => {
        const res = await verifySignature(stateHash, signature, '0xOther' as Address);
        expect(res).toBe(false);
    });
});
