import { describe, test, expect, jest, beforeAll, afterAll } from '@jest/globals';
import { Address, Hex } from 'viem';
import { NitroliteRPC } from '../../../src/rpc/nitrolite';
import {
    RPCMessage,
    RPCMessageType,
    StateSigner,
    StateVerifier,
    MultiStateVerifier,
} from '../../../src/rpc/types';

describe('NitroliteRPC', () => {
    beforeAll(() => {
        jest.spyOn(console, 'error').mockImplementation(() => {});
    });
    afterAll(() => {
        (console.error as jest.Mock).mockRestore();
    });

    describe('createRequest', () => {
        test('should create a valid request message in wire format', () => {
            const requestId = 12345;
            const method = 'ping';
            const params = {
                param1: 'value1',
                param2: 'value2',
            };
            const timestamp = 1619876543210;

            const result = NitroliteRPC.createRequest({
                requestId,
                method,
                params,
                timestamp,
            });

            expect(result).toEqual([RPCMessageType.Request, requestId, method, params, timestamp]);
        });

        test('should use default values when not provided', () => {
            jest.spyOn(global.Date, 'now').mockReturnValue(1619876543210);
            const method = 'ping';
            const result = NitroliteRPC.createRequest({
                method,
                params: {},
            });

            expect(result[0]).toBe(RPCMessageType.Request);
            expect(result[1]).toBeGreaterThan(0);
            expect(result[2]).toBe(method);
            expect(result[3]).toEqual({});
            expect(result[4]).toBe(1619876543210);
        });
    });

    describe('parseMessage', () => {
        test('should parse a valid wire format message', () => {
            const message: RPCMessage = [RPCMessageType.Request, 12345, 'ping', { test: 'data' }, 1619876543210];

            const result = NitroliteRPC.parseMessage(message);

            expect(result).toEqual({
                type: RPCMessageType.Request,
                requestId: 12345,
                method: 'ping',
                params: { test: 'data' },
                timestamp: 1619876543210,
            });
        });

        test('should throw on invalid message format', () => {
            const invalidMessage = [1, 2, 3] as any;

            expect(() => NitroliteRPC.parseMessage(invalidMessage)).toThrow('Invalid RPC message format');
        });
    });

    describe('signStateData', () => {
        test('should sign state data using provided signer', async () => {
            const mockSigner = jest.fn<StateSigner>().mockResolvedValue('0xsignature' as Hex);
            const data = { test: 'data' };

            const result = await NitroliteRPC.signStateData(data, mockSigner);

            expect(mockSigner).toHaveBeenCalledWith(data);
            expect(result).toBe('0xsignature');
        });
    });

    describe('verifyStateSignature', () => {
        test('should verify a state signature correctly', async () => {
            const mockVerifier = jest.fn<StateVerifier>().mockResolvedValue(true);
            const data = { test: 'data' };
            const signature = '0xsignature' as Hex;
            const expectedSigner = '0xsigner' as Address;

            const result = await NitroliteRPC.verifyStateSignature(data, signature, expectedSigner, mockVerifier);

            expect(mockVerifier).toHaveBeenCalledWith(data, signature, expectedSigner);
            expect(result).toBe(true);
        });

        test('should return false on verification error', async () => {
            const mockVerifier = jest.fn<StateVerifier>().mockImplementation(() => {
                throw new Error('Verification error');
            });
            const data = { test: 'data' };
            const signature = '0xsignature' as Hex;
            const expectedSigner = '0xsigner' as Address;

            const result = await NitroliteRPC.verifyStateSignature(data, signature, expectedSigner, mockVerifier);

            expect(mockVerifier).toHaveBeenCalled();
            expect(console.error).toHaveBeenCalled();
            expect(result).toBe(false);
        });
    });

    describe('verifyMultipleStateSignatures', () => {
        test('should verify multiple state signatures correctly', async () => {
            const mockVerifier = jest.fn<MultiStateVerifier>().mockResolvedValue(true);
            const data = { test: 'data' };
            const signatures = ['0xsig1' as Hex, '0xsig2' as Hex];
            const expectedSigners = ['0xsigner1' as Address, '0xsigner2' as Address];

            const result = await NitroliteRPC.verifyMultipleStateSignatures(data, signatures, expectedSigners, mockVerifier);

            expect(mockVerifier).toHaveBeenCalledWith(data, signatures, expectedSigners);
            expect(result).toBe(true);
        });

        test('should return false on verification error', async () => {
            const mockVerifier = jest.fn<MultiStateVerifier>().mockImplementation(() => {
                throw new Error('Verification error');
            });
            const data = { test: 'data' };
            const signatures = ['0xsig1' as Hex, '0xsig2' as Hex];
            const expectedSigners = ['0xsigner1' as Address, '0xsigner2' as Address];

            const result = await NitroliteRPC.verifyMultipleStateSignatures(data, signatures, expectedSigners, mockVerifier);

            expect(mockVerifier).toHaveBeenCalled();
            expect(console.error).toHaveBeenCalled();
            expect(result).toBe(false);
        });
    });
});
