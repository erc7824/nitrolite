import { describe, test, expect } from '@jest/globals';
import { NitroliteErrorCode } from '../../../src/rpc/types';

describe('NitroliteErrorCode enum', () => {
    const entries: Array<[keyof typeof NitroliteErrorCode, number]> = [
        ['PARSE_ERROR', -32700],
        ['INVALID_REQUEST', -32600],
        ['METHOD_NOT_FOUND', -32601],
        ['INVALID_PARAMS', -32602],
        ['INTERNAL_ERROR', -32603],
        ['AUTHENTICATION_FAILED', -32000],
        ['INVALID_SIGNATURE', -32003],
        ['INVALID_TIMESTAMP', -32004],
        ['INVALID_REQUEST_ID', -32005],
        ['INSUFFICIENT_FUNDS', -32007],
        ['ACCOUNT_NOT_FOUND', -32008],
        ['APPLICATION_NOT_FOUND', -32009],
        ['INVALID_INTENT', -32010],
        ['INSUFFICIENT_SIGNATURES', -32006],
        ['CHALLENGE_EXPIRED', -32011],
        ['INVALID_CHALLENGE', -32012],
    ];

    test.each(entries)('%s should equal %d', (key, expected) => {
        expect(NitroliteErrorCode[key]).toBe(expected);
    });

    test('reverse mapping works', () => {
        // For a numeric enum, accessing by value returns the key string
        const code = NitroliteErrorCode.PARSE_ERROR;
        expect(NitroliteErrorCode[code]).toBe('PARSE_ERROR');
    });
});
