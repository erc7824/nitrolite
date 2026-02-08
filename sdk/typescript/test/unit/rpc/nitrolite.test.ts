import { describe, test, expect, jest, beforeAll, afterAll } from '@jest/globals';
import { Address, Hex } from 'viem';
import { NitroliteRPC } from '../../../src/rpc/nitrolite';
import {
    NitroliteRPCMessage,
    MessageSigner,
    SingleMessageVerifier,
    MultiMessageVerifier,
    RPCMethod,
    RPCData,
} from '../../../src/rpc/types';

describe('NitroliteRPC', () => {
    beforeAll(() => {
        jest.spyOn(console, 'error').mockImplementation(() => {});
    });
    afterAll(() => {
        (console.error as jest.Mock).mockRestore();
    });
    describe('createRequest', () => {
        test('should create a valid request message', () => {
            const requestId = 12345;
            const method = RPCMethod.Ping;
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

            expect(result).toEqual({
                req: [requestId, method, params, timestamp],
                sig: [],
            });
        });

        test('should use default values when not provided', () => {
            jest.spyOn(global.Date, 'now').mockReturnValue(1619876543210);
            const result = NitroliteRPC.createRequest({
                method: RPCMethod.Ping,
                params: {},
            });

            expect(result.req).toBeDefined();
            expect(result.req![0]).toBeGreaterThan(0);
            expect(result.req![1]).toBe(RPCMethod.Ping);
            expect(result.req![2]).toEqual({});
            expect(result.req![3]).toBe(1619876543210);
        });
    });

    describe('createAppRequest', () => {
        test('should create a valid application request message', () => {
            const requestId = 12345;
            const method = RPCMethod.Ping;
            const params = {
                param1: 'value1',
                param2: 'value2',
            };
            const timestamp = 1619876543210;
            const accountId = '0xaccountId' as Hex;

            const result = NitroliteRPC.createAppRequest(
                {
                    requestId,
                    method,
                    params,
                    timestamp,
                },
                accountId,
            );

            expect(result).toEqual({
                req: [requestId, method, params, timestamp],
                sid: accountId,
            });
        });
    });

    describe('signRequestMessage', () => {
        test('should sign a request message and add signature to the message', async () => {
            const mockSigner = jest
                .fn<(data: RPCData) => Promise<Hex>>()
                .mockResolvedValue('0xsignature' as Hex);
            const request: NitroliteRPCMessage = {
                req: [12345, RPCMethod.Ping, ['param1', 'param2'], 1619876543210],
            };

            const result = await NitroliteRPC.signRequestMessage(request, mockSigner);

            expect(mockSigner).toHaveBeenCalledWith(request.req);
            expect(result).toEqual({
                req: [12345, RPCMethod.Ping, ['param1', 'param2'], 1619876543210],
                sig: ['0xsignature'],
            });
            expect(result).toBe(request);
        });

        test("should throw if message doesn't have req field", async () => {
            const mockSigner: MessageSigner = jest.fn<MessageSigner>().mockResolvedValue('0xsignature' as Hex);
            const invalidMessage = { res: [1, 2, 3, 4] } as unknown as NitroliteRPCMessage;

            await expect(NitroliteRPC.signRequestMessage(invalidMessage, mockSigner)).rejects.toThrow(
                "signRequestMessage can only sign request messages containing 'req'.",
            );
        });
    });

    describe('verifySingleSignature', () => {
        test('should verify a correctly signed message', async () => {
            const mockVerifier = jest.fn<SingleMessageVerifier>().mockResolvedValue(true);
            const message: NitroliteRPCMessage = {
                req: [12345, RPCMethod.Ping, ['param1', 'param2'], 1619876543210],
                sig: ['0xsignature' as Hex],
            };
            const expectedSigner = '0xsigner' as Address;

            const result = await NitroliteRPC.verifySingleSignature(message, expectedSigner, mockVerifier);

            expect(mockVerifier).toHaveBeenCalledWith(message.req, '0xsignature', expectedSigner);
            expect(result).toBe(true);
        });

        test('should return false for message with no signatures', async () => {
            const mockVerifier = jest.fn<SingleMessageVerifier>();
            const message: NitroliteRPCMessage = {
                req: [12345, RPCMethod.Ping, ['param1', 'param2'], 1619876543210],
            };
            const expectedSigner = '0xsigner' as Address;

            const result = await NitroliteRPC.verifySingleSignature(message, expectedSigner, mockVerifier);

            expect(mockVerifier).not.toHaveBeenCalled();
            expect(result).toBe(false);
        });

        test('should handle verification errors', async () => {
            const mockVerifier = jest.fn<SingleMessageVerifier>().mockImplementation(() => {
                throw new Error('Verification error');
            });
            const message: NitroliteRPCMessage = {
                req: [12345, RPCMethod.Ping, ['param1', 'param2'], 1619876543210],
                sig: ['0xsignature' as Hex],
            };
            const expectedSigner = '0xsigner' as Address;

            const result = await NitroliteRPC.verifySingleSignature(message, expectedSigner, mockVerifier);

            expect(mockVerifier).toHaveBeenCalled();
            expect(result).toBe(false);
        });

        test('should verify a response message', async () => {
            const mockVerifier = jest.fn<SingleMessageVerifier>().mockResolvedValue(true);
            const message: NitroliteRPCMessage = {
                res: [12345, RPCMethod.Ping, ['result1', 'result2'], 1619876543210],
                sig: ['0xsignature' as Hex],
            };
            const expectedSigner = '0xsigner' as Address;

            const result = await NitroliteRPC.verifySingleSignature(message, expectedSigner, mockVerifier);

            expect(mockVerifier).toHaveBeenCalledWith(message.res, '0xsignature', expectedSigner);
            expect(result).toBe(true);
        });

        test('should verify only the first signature when multiple are present', async () => {
            console.warn = jest.fn();
            const mockVerifier = jest.fn<SingleMessageVerifier>().mockResolvedValue(true);
            const message: NitroliteRPCMessage = {
                req: [12345, RPCMethod.Ping, ['param1', 'param2'], 1619876543210],
                sig: ['0xsignature1' as Hex, '0xsignature2' as Hex],
            };
            const expectedSigner = '0xsigner' as Address;

            const result = await NitroliteRPC.verifySingleSignature(message, expectedSigner, mockVerifier);

            expect(mockVerifier).toHaveBeenCalledWith(message.req, '0xsignature1', expectedSigner);
            expect(console.error).toHaveBeenCalled();
            expect(result).toBe(true);
        });
    });

    describe('verifyMultipleSignatures', () => {
        test('should verify multiple signatures correctly', async () => {
            const mockVerifier = jest.fn<MultiMessageVerifier>().mockResolvedValue(true);
            const message: NitroliteRPCMessage = {
                req: [12345, RPCMethod.Ping, ['param1', 'param2'], 1619876543210],
                sig: ['0xsig1' as Hex, '0xsig2' as Hex],
            };
            const expectedSigners = ['0xsigner1' as Address, '0xsigner2' as Address];

            const result = await NitroliteRPC.verifyMultipleSignatures(message, expectedSigners, mockVerifier);

            expect(mockVerifier).toHaveBeenCalledWith(message.req, message.sig, expectedSigners);
            expect(result).toBe(true);
        });

        test('should return false for message with no signatures', async () => {
            const mockVerifier = jest.fn<MultiMessageVerifier>();
            const message: NitroliteRPCMessage = {
                req: [12345, RPCMethod.Ping, ['param1', 'param2'], 1619876543210],
            };
            const expectedSigners = ['0xsigner1' as Address, '0xsigner2' as Address];

            const result = await NitroliteRPC.verifyMultipleSignatures(message, expectedSigners, mockVerifier);

            expect(mockVerifier).not.toHaveBeenCalled();
            expect(result).toBe(false);
        });

        test('should handle verification errors', async () => {
            const mockVerifier = jest.fn<MultiMessageVerifier>().mockImplementation(() => {
                throw new Error('Verification error');
            });
            const message: NitroliteRPCMessage = {
                req: [12345, RPCMethod.Ping, ['param1', 'param2'], 1619876543210],
                sig: ['0xsig1' as Hex, '0xsig2' as Hex],
            };
            const expectedSigners = ['0xsigner1' as Address, '0xsigner2' as Address];

            const result = await NitroliteRPC.verifyMultipleSignatures(message, expectedSigners, mockVerifier);

            expect(mockVerifier).toHaveBeenCalled();
            expect(result).toBe(false);
        });
    });
});
