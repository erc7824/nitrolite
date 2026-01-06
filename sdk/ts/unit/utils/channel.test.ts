import { describe, test, expect, jest, beforeEach, afterEach } from '@jest/globals';
import { getChannelId, generateChannelNonce } from '../../../src/utils/channel';
import { encodeAbiParameters, keccak256, Address } from 'viem';
import type { Channel } from '../../../src/client/types';

jest.mock('viem', () => ({
    encodeAbiParameters: jest.fn(() => '0xdeadbeef'),
    keccak256: jest.fn(() => '0xabc123'),
}));

describe('getChannelId', () => {
    const channel: Channel = {
        participants: [
            '0x1111111111111111111111111111111111111111' as Address,
            '0x2222222222222222222222222222222222222222' as Address,
        ],
        adjudicator: '0x3333333333333333333333333333333333333333' as Address,
        challenge: 100n,
        nonce: 200n,
    };

    const chainId = 1;

    test('encodes parameters and hashes correctly', () => {
        const id = getChannelId(channel, chainId);
        expect(encodeAbiParameters).toHaveBeenCalledWith(
            [
                { name: 'participants', type: 'address[]' },
                { name: 'adjudicator', type: 'address' },
                { name: 'challenge', type: 'uint64' },
                { name: 'nonce', type: 'uint64' },
                { name: 'chainId', type: 'uint256' },
            ],
            [channel.participants, channel.adjudicator, channel.challenge, channel.nonce, chainId],
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
        expect(generateChannelNonce()).toBe(4294967296n);
    });

    test('mixes address component when provided', () => {
        const address = '0x00000000000000000000000000000010' as Address;
        // timestamp<<32 = 4294967296n, addressComponent = BigInt('0x10') = 16n
        // nonce = 4294967296n ^ 16n = 4294967312n
        expect(generateChannelNonce(address)).toBe(4294967312n);
    });
});
