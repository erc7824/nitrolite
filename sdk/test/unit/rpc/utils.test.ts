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
    parseRPCResponse,
} from '../../../src/rpc/utils';
import { NitroliteRPCMessage, RPCMethod, RPCChannelStatus } from '../../../src/rpc/types';
import {
    AuthVerifyResponseParams,
    GetConfigResponseParams,
    GetLedgerBalancesResponseParams,
    CreateAppSessionResponseParams,
    SubmitStateResponseParams,
    CloseAppSessionResponseParams,
    GetAppDefinitionResponseParams,
    GetAppSessionsResponseParams,
    ResizeChannelResponseParams,
    CloseChannelResponseParams,
    GetChannelsResponseParams,
    GetRPCHistoryResponseParams,
    GetAssetsResponseParams,
    BalanceUpdateResponseParams,
    ChannelUpdateResponseParams,
} from '../../../src/rpc/types/response';

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

            const request2: NitroliteRPCMessage = { req: [undefined, RPCMethod.Ping, [], 100] as any };
            const response2: NitroliteRPCMessage = { res: [123, RPCMethod.Ping, [], 200] };
            expect(isValidResponseRequestId(request2, response2)).toBe(false);

            const request3: NitroliteRPCMessage = { req: [123, RPCMethod.Ping, [], 100] };
            const response3: NitroliteRPCMessage = { res: [undefined, RPCMethod.Ping, [], 200] as any };
            expect(isValidResponseRequestId(request3 as any, response3)).toBe(false);
        });
    });

    describe('parseRPCResponse', () => {
        test('should parse auth_challenge response', () => {
            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.AuthChallenge, [{ challengeMessage: 'test-challenge' }], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.AuthChallenge);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual([{ challengeMessage: 'test-challenge' }]);
        });

        test('should parse get_ledger_balances response with array of balances', () => {
            const balances: GetLedgerBalancesResponseParams[] = [
                { asset: 'ETH', amount: '1.5' },
                { asset: 'USDC', amount: '1000' },
            ];

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.GetLedgerBalances, [balances], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.GetLedgerBalances);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual(balances);
        });

        test('should parse get_config response', () => {
            const config: GetConfigResponseParams = {
                broker_address: '0x1234567890123456789012345678901234567890',
                networks: [
                    {
                        name: 'Ethereum',
                        chain_id: 1,
                        custody_address: '0x1234567890123456789012345678901234567890',
                        adjudicator_address: '0x1234567890123456789012345678901234567890',
                    },
                ],
            };

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.GetConfig, [config], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.GetConfig);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual([config]);
        });

        test('should parse ping response with empty params', () => {
            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.Ping, [{}], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.Ping);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual([{}]);
        });

        test('should parse auth_verify response', () => {
            const params: AuthVerifyResponseParams = {
                address: '0x1234567890123456789012345678901234567890',
                jwt_token: 'test-jwt-token',
                session_key: '0x1234567890123456789012345678901234567890',
                success: true,
            };

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.AuthVerify, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.AuthVerify);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual([params]);
        });

        test('should parse create_app_session response', () => {
            const params: CreateAppSessionResponseParams = {
                app_session_id: '0x1234567890123456789012345678901234567890',
                version: 1,
                status: RPCChannelStatus.Open,
            };

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.CreateAppSession, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.CreateAppSession);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual([params]);
        });

        test('should parse submit_state response', () => {
            const params: SubmitStateResponseParams = {
                app_session_id: '0x1234567890123456789012345678901234567890',
                version: 1,
                status: RPCChannelStatus.Open,
            };

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.SubmitState, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.SubmitState);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual([params]);
        });

        test('should parse close_app_session response', () => {
            const params: CloseAppSessionResponseParams = {
                app_session_id: '0x1234567890123456789012345678901234567890',
                version: 1,
                status: RPCChannelStatus.Closed,
            };

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.CloseAppSession, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.CloseAppSession);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual([params]);
        });

        test('should parse get_app_definition response', () => {
            const params: GetAppDefinitionResponseParams = {
                protocol: 'test-protocol',
                participants: ['0x1234567890123456789012345678901234567890'],
                weights: [1, 1],
                quorum: 2,
                challenge: 3600,
            };

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.GetAppDefinition, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.GetAppDefinition);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual([params]);
        });

        test('should parse get_app_sessions response', () => {
            const params: GetAppSessionsResponseParams[] = [
                {
                    app_session_id: '0x1234567890123456789012345678901234567890',
                    status: RPCChannelStatus.Open,
                    participants: ['0x1234567890123456789012345678901234567890'],
                    protocol: 'test-protocol',
                    challenge: 3600,
                    weights: [1, 1],
                    quorum: 2,
                    version: 1,
                    nonce: 1,
                    created_at: '2024-01-01T00:00:00Z',
                    updated_at: '2024-01-01T00:00:00Z',
                },
            ];

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.GetAppSessions, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.GetAppSessions);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual(params);
        });

        test('should parse resize_channel response', () => {
            const params: ResizeChannelResponseParams = {
                channel_id: '0x1234567890123456789012345678901234567890',
                state_data: 'test-state-data',
                intent: 1,
                version: 1,
                allocations: [
                    {
                        destination: '0x1234567890123456789012345678901234567890',
                        token: '0x1234567890123456789012345678901234567890',
                        amount: '1000',
                    },
                ],
                state_hash: '0x1234567890123456789012345678901234567890',
                server_signature: {
                    v: '27',
                    r: '0x1234567890123456789012345678901234567890',
                    s: '0x1234567890123456789012345678901234567890',
                },
            };

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.ResizeChannel, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.ResizeChannel);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual([params]);
        });

        test('should parse close_channel response', () => {
            const params: CloseChannelResponseParams = {
                channel_id: '0x1234567890123456789012345678901234567890',
                intent: 1,
                version: 1,
                state_data: 'test-state-data',
                allocations: [
                    {
                        destination: '0x1234567890123456789012345678901234567890',
                        token: '0x1234567890123456789012345678901234567890',
                        amount: '1000',
                    },
                ],
                state_hash: '0x1234567890123456789012345678901234567890',
                server_signature: {
                    v: '27',
                    r: '0x1234567890123456789012345678901234567890',
                    s: '0x1234567890123456789012345678901234567890',
                },
            };

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.CloseChannel, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.CloseChannel);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual([params]);
        });

        test('should parse get_channels response', () => {
            const params: GetChannelsResponseParams[] = [
                {
                    channel_id: '0x1234567890123456789012345678901234567890',
                    participant: '0x1234567890123456789012345678901234567890',
                    status: RPCChannelStatus.Open,
                    token: '0x1234567890123456789012345678901234567890',
                    wallet: '0x1234567890123456789012345678901234567890',
                    amount: '1000',
                    chain_id: 1,
                    adjudicator: '0x1234567890123456789012345678901234567890',
                    challenge: 3600,
                    nonce: 1,
                    version: 1,
                    created_at: '2024-01-01T00:00:00Z',
                    updated_at: '2024-01-01T00:00:00Z',
                },
            ];

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.GetChannels, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.GetChannels);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual(params);
        });

        test('should parse get_rpc_history response', () => {
            const params: GetRPCHistoryResponseParams[] = [
                {
                    id: 1,
                    sender: '0x1234567890123456789012345678901234567890',
                    req_id: 123,
                    method: 'test_method',
                    params: '{"test": "params"}',
                    timestamp: 456,
                    req_sig: ['0x123'],
                    res_sig: ['0x123'],
                    response: '{"test": "response"}',
                },
            ];

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.GetRPCHistory, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.GetRPCHistory);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual(params);
        });

        test('should parse get_assets response', () => {
            const params: GetAssetsResponseParams[] = [
                {
                    token: '0x1234567890123456789012345678901234567890',
                    chain_id: 1,
                    symbol: 'TEST',
                    decimals: 18,
                },
            ];

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.GetAssets, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.GetAssets);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual(params);
        });

        test('should parse pong response', () => {
            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.Pong, [{}], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.Pong);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual([{}]);
        });

        test('should parse message response', () => {
            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.Message, [{}], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.Message);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual([{}]);
        });

        test('should parse balance_update response', () => {
            const params: BalanceUpdateResponseParams[] = [
                {
                    asset: 'ETH',
                    amount: '1.5',
                },
            ];

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.BalanceUpdate, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.BalanceUpdate);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual(params);
        });

        test('should parse channels_update response', () => {
            const params: ChannelUpdateResponseParams[] = [
                {
                    channel_id: '0x1234567890123456789012345678901234567890',
                    participant: '0x1234567890123456789012345678901234567890',
                    status: RPCChannelStatus.Open,
                    token: '0x1234567890123456789012345678901234567890',
                    amount: '1000',
                    chain_id: 1,
                    adjudicator: '0x1234567890123456789012345678901234567890',
                    challenge: 3600,
                    nonce: 1,
                    version: 1,
                    created_at: '2024-01-01T00:00:00Z',
                    updated_at: '2024-01-01T00:00:00Z',
                },
            ];

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.ChannelsUpdate, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.ChannelsUpdate);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual(params);
        });

        test('should parse channel_update response', () => {
            const params: ChannelUpdateResponseParams = {
                channel_id: '0x1234567890123456789012345678901234567890',
                participant: '0x1234567890123456789012345678901234567890',
                status: RPCChannelStatus.Open,
                token: '0x1234567890123456789012345678901234567890',
                amount: '1000',
                chain_id: 1,
                adjudicator: '0x1234567890123456789012345678901234567890',
                challenge: 3600,
                nonce: 1,
                version: 1,
                created_at: '2024-01-01T00:00:00Z',
                updated_at: '2024-01-01T00:00:00Z',
            };

            const rawResponse = JSON.stringify({
                res: [123, RPCMethod.ChannelUpdate, [params], 456],
                sig: ['0x123'],
            });

            const result = parseRPCResponse(rawResponse);
            expect(result.method).toBe(RPCMethod.ChannelUpdate);
            expect(result.requestId).toBe(123);
            expect(result.timestamp).toBe(456);
            expect(result.signatures).toEqual(['0x123']);
            expect(result.params).toEqual([params]);
        });

        test('should throw error for invalid response format', () => {
            const invalidResponse = JSON.stringify({
                res: [123, RPCMethod.Ping, 456], // Missing timestamp
            });

            expect(() => parseRPCResponse(invalidResponse)).toThrow('Invalid RPC response format');
        });

        test('should throw error for invalid JSON', () => {
            const invalidJSON = 'invalid json';
            expect(() => parseRPCResponse(invalidJSON)).toThrow('Failed to parse RPC response');
        });
    });
});
