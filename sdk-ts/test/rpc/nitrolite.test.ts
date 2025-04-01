import { NitroliteRPC, NitroliteRPCMessage, NitroliteErrorCode } from '../../src/rpc';
import { Hex } from 'viem';

describe('NitroliteRPC', () => {
  // Test constants
  const TEST_METHOD = 'test_method';
  const TEST_PARAMS = [1, 'test', { key: 'value' }];
  const TEST_RESULT = [42, { success: true }];
  const TEST_REQUEST_ID = 12345;
  const TEST_TIMESTAMP = 1709584837123;
  const TEST_ERROR_MESSAGE = 'Test error message';

  // Mock signing function that just returns a fixed signature
  const mockSigner = async (payload: Hex): Promise<Hex> => {
    return '0xabc123def456' as Hex;
  };

  // Mock verification function that always returns true
  const mockVerifier = async (payload: Hex, signature: Hex, address: string): Promise<boolean> => {
    return signature === '0xabc123def456';
  };

  describe('createRequest', () => {
    it('should create a valid request message', () => {
      const request = NitroliteRPC.createRequest(
        TEST_METHOD,
        TEST_PARAMS,
        TEST_REQUEST_ID,
        TEST_TIMESTAMP
      );

      expect(request.req).toBeDefined();
      expect(request.req?.[0]).toBe(TEST_REQUEST_ID);
      expect(request.req?.[1]).toBe(TEST_METHOD);
      expect(request.req?.[2]).toEqual(TEST_PARAMS);
      expect(request.req?.[3]).toBe(TEST_TIMESTAMP);
    });

    it('should generate requestId and timestamp if not provided', () => {
      const request = NitroliteRPC.createRequest(TEST_METHOD, TEST_PARAMS);

      expect(request.req).toBeDefined();
      expect(typeof request.req?.[0]).toBe('number');
      expect(request.req?.[1]).toBe(TEST_METHOD);
      expect(request.req?.[2]).toEqual(TEST_PARAMS);
      expect(typeof request.req?.[3]).toBe('number');
    });
  });

  describe('createResponse', () => {
    it('should create a valid response message', () => {
      const response = NitroliteRPC.createResponse(
        TEST_REQUEST_ID,
        TEST_METHOD,
        TEST_RESULT,
        TEST_TIMESTAMP
      );

      expect(response.res).toBeDefined();
      expect(response.res?.[0]).toBe(TEST_REQUEST_ID);
      expect(response.res?.[1]).toBe(TEST_METHOD);
      expect(response.res?.[2]).toEqual(TEST_RESULT);
      expect(response.res?.[3]).toBe(TEST_TIMESTAMP);
    });

    it('should use current timestamp if not provided', () => {
      const response = NitroliteRPC.createResponse(
        TEST_REQUEST_ID,
        TEST_METHOD,
        TEST_RESULT
      );

      expect(response.res).toBeDefined();
      expect(response.res?.[0]).toBe(TEST_REQUEST_ID);
      expect(typeof response.res?.[3]).toBe('number');
    });
  });

  describe('createError', () => {
    it('should create a valid error message', () => {
      const error = NitroliteRPC.createError(
        TEST_REQUEST_ID,
        NitroliteErrorCode.METHOD_NOT_FOUND,
        TEST_ERROR_MESSAGE,
        TEST_TIMESTAMP
      );

      expect(error.err).toBeDefined();
      expect(error.err?.[0]).toBe(TEST_REQUEST_ID);
      expect(error.err?.[1]).toBe(NitroliteErrorCode.METHOD_NOT_FOUND);
      expect(error.err?.[2]).toBe(TEST_ERROR_MESSAGE);
      expect(error.err?.[3]).toBe(TEST_TIMESTAMP);
    });
  });

  describe('signMessage', () => {
    it('should sign a request message', async () => {
      const request = NitroliteRPC.createRequest(
        TEST_METHOD,
        TEST_PARAMS,
        TEST_REQUEST_ID,
        TEST_TIMESTAMP
      );

      const signedMessage = await NitroliteRPC.signMessage(request, mockSigner);
      
      expect(signedMessage.req).toEqual(request.req);
      expect(signedMessage.sig).toBe('0xabc123def456');
    });

    it('should sign a response message', async () => {
      const response = NitroliteRPC.createResponse(
        TEST_REQUEST_ID,
        TEST_METHOD,
        TEST_RESULT,
        TEST_TIMESTAMP
      );

      const signedMessage = await NitroliteRPC.signMessage(response, mockSigner);
      
      expect(signedMessage.res).toEqual(response.res);
      expect(signedMessage.sig).toBe('0xabc123def456');
    });

    it('should sign an error message', async () => {
      const error = NitroliteRPC.createError(
        TEST_REQUEST_ID,
        NitroliteErrorCode.METHOD_NOT_FOUND,
        TEST_ERROR_MESSAGE,
        TEST_TIMESTAMP
      );

      const signedMessage = await NitroliteRPC.signMessage(error, mockSigner);
      
      expect(signedMessage.err).toEqual(error.err);
      expect(signedMessage.sig).toBe('0xabc123def456');
    });

    it('should throw an error for invalid messages', async () => {
      const invalidMessage = {} as NitroliteRPCMessage;
      
      await expect(NitroliteRPC.signMessage(invalidMessage, mockSigner))
        .rejects
        .toThrow('Invalid message: must contain req, res, or err field');
    });
  });

  describe('verifyMessage', () => {
    it('should verify a signed request message', async () => {
      const request = NitroliteRPC.createRequest(
        TEST_METHOD,
        TEST_PARAMS,
        TEST_REQUEST_ID,
        TEST_TIMESTAMP
      );
      
      const signedMessage = await NitroliteRPC.signMessage(request, mockSigner);
      const isValid = await NitroliteRPC.verifyMessage(
        signedMessage,
        '0x1234567890123456789012345678901234567890',
        mockVerifier
      );
      
      expect(isValid).toBe(true);
    });

    it('should return false for a message with no signature', async () => {
      const request = NitroliteRPC.createRequest(
        TEST_METHOD,
        TEST_PARAMS,
        TEST_REQUEST_ID,
        TEST_TIMESTAMP
      );
      
      const isValid = await NitroliteRPC.verifyMessage(
        request,
        '0x1234567890123456789012345678901234567890',
        mockVerifier
      );
      
      expect(isValid).toBe(false);
    });

    it('should return false for an invalid message', async () => {
      const invalidMessage = { sig: '0xabc123def456' } as NitroliteRPCMessage;
      
      const isValid = await NitroliteRPC.verifyMessage(
        invalidMessage,
        '0x1234567890123456789012345678901234567890',
        mockVerifier
      );
      
      expect(isValid).toBe(false);
    });
  });
});