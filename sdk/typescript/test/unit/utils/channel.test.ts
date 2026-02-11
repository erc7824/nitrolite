import { describe, test, expect, jest, beforeEach, afterEach } from '@jest/globals';
import { getChannelId, generateChannelNonce } from '../../../src/utils/channel';
import { encodeAbiParameters, keccak256, Address, Hex } from 'viem';
import type { ChannelDefinition } from '../../../src/client/types';

jest.mock('viem', () => ({
    encodeAbiParameters: jest.fn(() => '0xdeadbeef'),
    keccak256: jest.fn(() => '0xabc123'),
}));

describe('getChannelId', () => {
    const definition: ChannelDefinition = {
        challengeDuration: 3600,
        user: '0x1111111111111111111111111111111111111111' as Address,
        node: '0x2222222222222222222222222222222222222222' as Address,
        nonce: 200n,
        metadata: '0x0000000000000000000000000000000000000000000000000000000000000000' as Hex,
    };

    const chainId = 1;

    test('encodes parameters and hashes correctly', () => {
        const id = getChannelId(definition, chainId);
        expect(encodeAbiParameters).toHaveBeenCalledWith(
            [
                { name: 'challengeDuration', type: 'uint32' },
                { name: 'user', type: 'address' },
                { name: 'node', type: 'address' },
                { name: 'nonce', type: 'uint64' },
                { name: 'metadata', type: 'bytes32' },
                { name: 'chainId', type: 'uint256' },
            ],
            [definition.challengeDuration, definition.user, definition.node, definition.nonce, definition.metadata, BigInt(chainId)],
        );
        expect(keccak256).toHaveBeenCalledWith('0xdeadbeef');
        expect(id).toBe('0xabc123');
    });
});

describe('generateChannelNonce', () => {
    beforeEach(() => {
        jest.spyOn(Date, 'now').mockReturnValue(1000);
        jest.spyOn(Math, 'random').mockReturnValue(0);
    });
    afterEach(() => {
        (Date.now as jest.MockedFunction<any>).mockRestore();
        (Math.random as jest.MockedFunction<any>).mockRestore();
    });

    test('produces deterministic nonce without address', () => {
        // timestamp = floor(1000/1000) = 1n => 1n << 32 = 4294967296n
        // randomComponent = floor(0 * 0xffffffff) = 0n
        // masked to fit int64 (& 0x7fffffffffffffff)
        expect(generateChannelNonce()).toBe(4294967296n);
    });

    test('mixes address component when provided', () => {
        const address = '0x00000000000000000000000000000010' as Address;
        // timestamp<<32 = 4294967296n, addressComponent = BigInt('0x10') = 16n
        // nonce = 4294967296n ^ 16n = 4294967312n (within int64 range)
        expect(generateChannelNonce(address)).toBe(4294967312n);
    });
});
