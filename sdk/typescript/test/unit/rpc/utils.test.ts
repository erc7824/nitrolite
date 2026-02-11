import { describe, test, expect, jest } from '@jest/globals';
import {
    getCurrentTimestamp,
    generateRequestId,
    getRequestId,
    getMethod,
    getParams,
    getTimestamp,
    getMessageType,
    isRequest,
    isResponse,
    isEvent,
    isErrorResponse,
    toBytes,
    isValidResponseTimestamp,
    isValidResponseRequestId,
} from '../../../src/rpc/utils';
import { RPCMessage, RPCMessageType } from '../../../src/rpc/types';

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

    describe('getMessageType', () => {
        test('should extract message type from wire format', () => {
            const message: RPCMessage = [RPCMessageType.Request, 123, 'ping', {}, 456];
            expect(getMessageType(message)).toBe(RPCMessageType.Request);
        });
    });

    describe('getRequestId', () => {
        test('should extract request ID from wire format', () => {
            const message: RPCMessage = [RPCMessageType.Request, 123, 'ping', {}, 456];
            expect(getRequestId(message)).toBe(123);
        });
    });

    describe('getMethod', () => {
        test('should extract method from wire format', () => {
            const message: RPCMessage = [RPCMessageType.Request, 123, 'ping', {}, 456];
            expect(getMethod(message)).toBe('ping');
        });
    });

    describe('getParams', () => {
        test('should extract params from wire format', () => {
            const params = { param1: 'value1', param2: 'value2' };
            const message: RPCMessage = [RPCMessageType.Request, 123, 'ping', params, 456];
            expect(getParams(message)).toEqual(params);
        });

        test('should return params even if empty object', () => {
            const message: RPCMessage = [RPCMessageType.Request, 123, 'ping', {}, 456];
            expect(getParams(message)).toEqual({});
        });
    });

    describe('getTimestamp', () => {
        test('should extract timestamp from wire format', () => {
            const message: RPCMessage = [RPCMessageType.Request, 123, 'ping', {}, 456];
            expect(getTimestamp(message)).toBe(456);
        });
    });

    describe('isRequest', () => {
        test('should return true for request message', () => {
            const message: RPCMessage = [RPCMessageType.Request, 123, 'ping', {}, 456];
            expect(isRequest(message)).toBe(true);
        });

        test('should return false for response message', () => {
            const message: RPCMessage = [RPCMessageType.Response, 123, 'ping', {}, 456];
            expect(isRequest(message)).toBe(false);
        });
    });

    describe('isResponse', () => {
        test('should return true for response message', () => {
            const message: RPCMessage = [RPCMessageType.Response, 123, 'ping', {}, 456];
            expect(isResponse(message)).toBe(true);
        });

        test('should return false for request message', () => {
            const message: RPCMessage = [RPCMessageType.Request, 123, 'ping', {}, 456];
            expect(isResponse(message)).toBe(false);
        });
    });

    describe('isEvent', () => {
        test('should return true for event message', () => {
            const message: RPCMessage = [RPCMessageType.Event, 123, 'channelUpdate', {}, 456];
            expect(isEvent(message)).toBe(true);
        });

        test('should return false for request message', () => {
            const message: RPCMessage = [RPCMessageType.Request, 123, 'ping', {}, 456];
            expect(isEvent(message)).toBe(false);
        });
    });

    describe('isErrorResponse', () => {
        test('should return true for error response message', () => {
            const message: RPCMessage = [RPCMessageType.ErrorResponse, 123, 'error', { code: 500, message: 'Error' }, 456];
            expect(isErrorResponse(message)).toBe(true);
        });

        test('should return false for response message', () => {
            const message: RPCMessage = [RPCMessageType.Response, 123, 'ping', {}, 456];
            expect(isErrorResponse(message)).toBe(false);
        });
    });

    describe('toBytes', () => {
        test('should convert string values to hex', () => {
            const values = ['test'];
            const result = toBytes(values);
            expect(result).toHaveLength(1);
            expect(result[0]).toMatch(/^0x[0-9a-fA-F]+$/);
        });

        test('should convert object values to JSON then hex', () => {
            const values = [{ key: 'value' }];
            const result = toBytes(values);
            expect(result).toHaveLength(1);
            expect(result[0]).toMatch(/^0x[0-9a-fA-F]+$/);
        });
    });

    describe('isValidResponseTimestamp', () => {
        test('should return true if response timestamp is greater than request timestamp', () => {
            const request: RPCMessage = [RPCMessageType.Request, 123, 'ping', {}, 100];
            const response: RPCMessage = [RPCMessageType.Response, 123, 'ping', {}, 200];
            expect(isValidResponseTimestamp(request, response)).toBe(true);
        });

        test('should return false if response timestamp is not greater than request timestamp', () => {
            const request: RPCMessage = [RPCMessageType.Request, 123, 'ping', {}, 200];
            const response: RPCMessage = [RPCMessageType.Response, 123, 'ping', {}, 100];
            expect(isValidResponseTimestamp(request, response)).toBe(false);
        });
    });

    describe('isValidResponseRequestId', () => {
        test('should return true if response request ID matches request ID', () => {
            const request: RPCMessage = [RPCMessageType.Request, 123, 'ping', {}, 100];
            const response: RPCMessage = [RPCMessageType.Response, 123, 'ping', {}, 200];
            expect(isValidResponseRequestId(request, response)).toBe(true);
        });

        test('should return false if response request ID does not match request ID', () => {
            const request: RPCMessage = [RPCMessageType.Request, 123, 'ping', {}, 100];
            const response: RPCMessage = [RPCMessageType.Response, 456, 'ping', {}, 200];
            expect(isValidResponseRequestId(request, response)).toBe(false);
        });
    });
});
