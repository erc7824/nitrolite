import { describe, test, expect, jest } from '@jest/globals';
import {
    getCurrentTimestamp,
    generateRequestId,
    getRequestId,
    getMethod,
    getParams,
    getResult,
    getTimestamp,
    getError,
    toBytes,
    isValidResponseTimestamp,
    isValidResponseRequestId,
} from '../../../src/rpc/utils';
import { parseAuthChallengeResponse, parseCreateAppSessionResponse, parseGetConfigResponse, parseGetLedgerBalancesResponse, parsePingResponse } from '../../../src/rpc/parse/parse';
import { NitroliteRPCMessage, RPCMethod, RPCChannelStatus } from '../../../src/rpc/types';

describe('RPC Utils', () => {
    describe('getCurrentTimestamp', () => {
        test('should return the current timestamp', () => {
            jest.spyOn(Date, 'now').mockReturnValue(1234567890);
            expect(getCurrentTimestamp()).toBe(1234567890);
        });
    });

    describe('generateRequestId', () => {
        test('should generate a unique request ID', () => {
            jest.spyOn(Date, 'now').mockReturnValue(1234567890);
            jest.spyOn(Math, 'random').mockReturnValue(0.5);
            expect(generateRequestId()).toBe(1234567890 + 5000);
        });
    });

    describe('getRequestId', () => {
        test('should extract request ID from req field', () => {
            const message = { req: [123, RPCMethod.Ping, [], 456] };
            expect(getRequestId(message)).toBe(123);
        });

        test('should extract request ID from res field', () => {
            const message = { res: [123, RPCMethod.Ping, [], 456] };
            expect(getRequestId(message)).toBe(123);
        });

        test('should extract request ID from err field', () => {
            const message = { err: [123, 'error', 'message', 456] };
            expect(getRequestId(message)).toBe(123);
        });

        test('should return undefined if no ID found', () => {
            const message = { other: 'value' };
            expect(getRequestId(message)).toBeUndefined();
        });
    });

    describe('getMethod', () => {
        test('should extract method from req field', () => {
            const message = { req: [123, RPCMethod.Ping, [], 456] };
            expect(getMethod(message)).toBe(RPCMethod.Ping);
        });

        test('should extract method from res field', () => {
            const message = { res: [123, RPCMethod.Ping, [], 456] };
            expect(getMethod(message)).toBe(RPCMethod.Ping);
        });

        test('should return undefined if no method found', () => {
            const message = { other: 'value' };
            expect(getMethod(message)).toBeUndefined();
        });
    });

    describe('getParams', () => {
        test('should extract params from req field', () => {
            const params = ['param1', 'param2'];
            const message = { req: [123, RPCMethod.Ping, params, 456] };
            expect(getParams(message)).toBe(params);
        });

        test('should return empty array if no params found', () => {
            const message = { req: [123, RPCMethod.Ping, null, 456] };
            expect(getParams(message)).toEqual([]);
        });

        test('should return empty array if no req field', () => {
            const message = { other: 'value' };
            expect(getParams(message)).toEqual([]);
        });
    });

    describe('getResult', () => {
        test('should extract result from res field', () => {
            const result = ['result1', 'result2'];
            const message = { res: [123, RPCMethod.Ping, result, 456] };
            expect(getResult(message)).toBe(result);
        });

        test('should return empty array if no result found', () => {
            const message = { res: [123, RPCMethod.Ping, null, 456] };
            expect(getResult(message)).toEqual([]);
        });

        test('should return empty array if no res field', () => {
            const message = { other: 'value' };
            expect(getResult(message)).toEqual([]);
        });
    });

    describe('getTimestamp', () => {
        test('should extract timestamp from req field', () => {
            const message = { req: [123, RPCMethod.Ping, [], 456] };
            expect(getTimestamp(message)).toBe(456);
        });

        test('should extract timestamp from res field', () => {
            const message = { res: [123, RPCMethod.Ping, [], 456] };
            expect(getTimestamp(message)).toBe(456);
        });

        test('should extract timestamp from err field', () => {
            const message = { err: [123, 'error', 'message', 456] };
            expect(getTimestamp(message)).toBe(456);
        });

        test('should return undefined if no timestamp found', () => {
            const message = { other: 'value' };
            expect(getTimestamp(message)).toBeUndefined();
        });
    });

    describe('getError', () => {
        test('should extract error details from err field', () => {
            const message = { err: [123, 400, 'Bad Request', 456] };
            expect(getError(message)).toEqual({
                code: 400,
                message: 'Bad Request',
            });
        });

        test('should return undefined if no err field', () => {
            const message = { other: 'value' };
            expect(getError(message)).toBeUndefined();
        });
    });

    describe('toBytes', () => {
        test('should convert string values to bytes', () => {
            const values = ['value1', 'value2'];
            const result = toBytes(values);
            expect(result[0]).toBe('0x76616c756531');
            expect(result[1]).toBe('0x76616c756532');
        });

        test('should convert non-string values to JSON bytes', () => {
            const values = [{ key: 'value' }, 123];
            const result = toBytes(values);
            expect(result[0]).toBe('0x7b226b6579223a2276616c7565227d');
            expect(result[1]).toBe('0x313233');
        });
    });

    describe('isValidResponseTimestamp', () => {
        test('should return true if response timestamp is greater than request timestamp', () => {
            const request: NitroliteRPCMessage = { req: [123, RPCMethod.Ping, [], 100] };
            const response: NitroliteRPCMessage = { res: [123, RPCMethod.Ping, [], 200] };
            expect(isValidResponseTimestamp(request, response)).toBe(true);
        });

        test('should return false if response timestamp is less than or equal to request timestamp', () => {
            const request: NitroliteRPCMessage = { req: [123, RPCMethod.Ping, [], 200] };
            const response: NitroliteRPCMessage = { res: [123, RPCMethod.Ping, [], 200] };
            expect(isValidResponseTimestamp(request, response)).toBe(false);

            const response2: NitroliteRPCMessage = { res: [123, RPCMethod.Ping, [], 100] };
            expect(isValidResponseTimestamp(request, response2)).toBe(false);
        });

        test('should return false if timestamps are missing', () => {
            const request: NitroliteRPCMessage = { req: [123, RPCMethod.Ping, []] };
            const response: NitroliteRPCMessage = { res: [123, RPCMethod.Ping, [], 200] };
            expect(isValidResponseTimestamp(request, response)).toBe(false);

            const request2: NitroliteRPCMessage = { req: [123, RPCMethod.Ping, [], 100] };
            const response2: NitroliteRPCMessage = { res: [123, RPCMethod.Ping, []] };
            expect(isValidResponseTimestamp(request2, response2)).toBe(false);
        });
    });

    describe('isValidResponseRequestId', () => {
        test('should return true if response request ID matches request ID', () => {
            const request: NitroliteRPCMessage = { req: [123, RPCMethod.Ping, [], 100] };
            const response: NitroliteRPCMessage = { res: [123, RPCMethod.Ping, [], 200] };
            expect(isValidResponseRequestId(request, response)).toBe(true);
        });

        test('should return false if response request ID does not match request ID', () => {
            const request: NitroliteRPCMessage = { req: [123, RPCMethod.Ping, [], 100] };
            const response: NitroliteRPCMessage = { res: [456, RPCMethod.Ping, [], 200] };
            expect(isValidResponseRequestId(request, response)).toBe(false);
        });
    });
});

describe('rpc response parsers', () => {
    test('should parse auth_challenge response correctly', () => {
        const rawResponse = JSON.stringify({
            res: [123, RPCMethod.AuthChallenge, { challenge_message: 'test-challenge' }, 456],
            sig: ['0x123'],
        });

        const result = parseAuthChallengeResponse(rawResponse);
        expect(result.method).toBe(RPCMethod.AuthChallenge);
        expect(result.requestId).toBe(123);
        expect(result.timestamp).toBe(456);
        expect(result.signatures).toEqual(['0x123']);
        expect(result.params).toEqual({ challengeMessage: 'test-challenge' });
    });

    test('should parse get_ledger_balances response correctly', () => {
        const balances = {
            ledger_balances: [
                { asset: 'eth', amount: 1.5 },
                { asset: 'usdc', amount: 1000 },
            ],
        };

        const rawResponse = JSON.stringify({
            res: [123, RPCMethod.GetLedgerBalances, balances, 456],
            sig: ['0x123'],
        });

        const result = parseGetLedgerBalancesResponse(rawResponse);
        expect(result.method).toBe(RPCMethod.GetLedgerBalances);
        expect(result.params).toEqual({
            ledgerBalances: [
                { asset: 'eth', amount: '1.5' },
                { asset: 'usdc', amount: '1000' },
            ],
        });
    });

    test('should parse get_config response correctly', () => {
        const config = {
            broker_address: '0x1234567890123456789012345678901234567890',
            networks: [
                {
                    chain_id: 1,
                    name: 'ethereum_mainnet',
                    custody_address: '0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
                    adjudicator_address: '0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb',
                },
            ],
        };

        const rawResponse = JSON.stringify({
            res: [123, RPCMethod.GetConfig, config, 456],
            sig: ['0x123'],
        });

        const result = parseGetConfigResponse(rawResponse);
        expect(result.method).toBe(RPCMethod.GetConfig);
        expect(result.params).toEqual({
            brokerAddress: '0x1234567890123456789012345678901234567890',
            networks: [
                {
                    chainId: 1,
                    name: 'ethereum_mainnet',
                    custodyAddress: '0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
                    adjudicatorAddress: '0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb',
                },
            ],
        });
    });

    test('should parse ping response correctly', () => {
        const rawResponse = JSON.stringify({
            res: [123, RPCMethod.Ping, {}, 456],
            sig: ['0x123'],
        });

        const result = parsePingResponse(rawResponse);
        expect(result.method).toBe(RPCMethod.Ping);
        expect(result.requestId).toBe(123);
        expect(result.params).toEqual({});
    });

    test('should parse create_app_session response correctly', () => {
        const params = {
            app_session_id: '0x1234567890123456789012345678901234567890',
            version: 1,
            status: RPCChannelStatus.Open,
        };

        const rawResponse = JSON.stringify({
            res: [123, RPCMethod.CreateAppSession, params, 456],
            sig: ['0x123'],
        });

        const result = parseCreateAppSessionResponse(rawResponse);

        expect(result.method).toBe(RPCMethod.CreateAppSession);
        expect(result.params).toEqual({
            appSessionId: '0x1234567890123456789012345678901234567890',
            version: 1,
            status: RPCChannelStatus.Open,
        });
    });

    test('should throw error for invalid response format', () => {
        const invalidResponse = JSON.stringify({
            res: [123, RPCMethod.Ping, 456], // Missing one element
        });

        expect(() => parsePingResponse(invalidResponse)).toThrow('Invalid RPC response format');
    });

    test('should throw error for invalid JSON', () => {
        const invalidJSON = 'this is not json';
        expect(() => parsePingResponse(invalidJSON)).toThrow(/Failed to parse RPC response/);
    });

    test('should throw error when parsing with the wrong method', () => {
        const rawResponse = JSON.stringify({
            res: [123, RPCMethod.AuthChallenge, { challenge_message: 'test-challenge' }, 456],
            sig: ['0x123'],
        });

        // Try to parse an auth_challenge response with the ping parser
        expect(() => parsePingResponse(rawResponse)).toThrow(
            "Expected RPC method to be 'ping', but received 'auth_challenge'",
        );
    });
});
