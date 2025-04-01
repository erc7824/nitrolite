import { 
  createPayload,
  getCurrentTimestamp,
  generateRequestId,
  getRequestId,
  getMethod,
  getParams,
  getResult,
  getError
} from '../../src/rpc/utils';
import { Hex } from 'viem';
import { NitroliteRPCMessage } from '../../src/rpc/types';

describe('NitroliteRPC Utils', () => {
  // Test constants
  const TEST_DATA = { test: 'value', numeric: 42 };
  const TEST_REQUEST_ID = 12345;
  const TEST_METHOD = 'test_method';
  const TEST_PARAMS = [1, 'test', { key: 'value' }];
  const TEST_RESULT = [42, { success: true }];
  const TEST_ERROR_CODE = -32601;
  const TEST_ERROR_MESSAGE = 'Method not found';

  describe('createPayload', () => {
    it('should convert data to hex string', () => {
      const result = createPayload(TEST_DATA);
      
      expect(typeof result).toBe('string');
      expect(result.startsWith('0x')).toBe(true);
      
      // Verify we can convert back to original data
      const hexWithout0x = result.slice(2);
      const buffer = Buffer.from(hexWithout0x, 'hex');
      const decoded = JSON.parse(buffer.toString());
      
      expect(decoded).toEqual(TEST_DATA);
    });
  });

  describe('getCurrentTimestamp', () => {
    it('should return a number representing current time', () => {
      const before = Date.now();
      const timestamp = getCurrentTimestamp();
      const after = Date.now();
      
      expect(typeof timestamp).toBe('number');
      expect(timestamp).toBeGreaterThanOrEqual(before);
      expect(timestamp).toBeLessThanOrEqual(after);
    });
  });

  describe('generateRequestId', () => {
    it('should generate a numeric request ID', () => {
      const requestId = generateRequestId();
      
      expect(typeof requestId).toBe('number');
    });

    it('should generate different IDs on subsequent calls', () => {
      const id1 = generateRequestId();
      const id2 = generateRequestId();
      
      expect(id1).not.toBe(id2);
    });
  });

  describe('Message extraction utilities', () => {
    let requestMessage: NitroliteRPCMessage;
    let responseMessage: NitroliteRPCMessage;
    let errorMessage: NitroliteRPCMessage;

    beforeEach(() => {
      requestMessage = {
        req: [TEST_REQUEST_ID, TEST_METHOD, TEST_PARAMS, Date.now()]
      };
      
      responseMessage = {
        res: [TEST_REQUEST_ID, TEST_METHOD, TEST_RESULT, Date.now()]
      };
      
      errorMessage = {
        err: [TEST_REQUEST_ID, TEST_ERROR_CODE, TEST_ERROR_MESSAGE, Date.now()]
      };
    });

    describe('getRequestId', () => {
      it('should extract requestId from request message', () => {
        expect(getRequestId(requestMessage)).toBe(TEST_REQUEST_ID);
      });

      it('should extract requestId from response message', () => {
        expect(getRequestId(responseMessage)).toBe(TEST_REQUEST_ID);
      });

      it('should extract requestId from error message', () => {
        expect(getRequestId(errorMessage)).toBe(TEST_REQUEST_ID);
      });

      it('should return undefined for invalid message', () => {
        expect(getRequestId({})).toBeUndefined();
      });
    });

    describe('getMethod', () => {
      it('should extract method from request message', () => {
        expect(getMethod(requestMessage)).toBe(TEST_METHOD);
      });

      it('should extract method from response message', () => {
        expect(getMethod(responseMessage)).toBe(TEST_METHOD);
      });

      it('should return undefined for error message', () => {
        expect(getMethod(errorMessage)).toBeUndefined();
      });

      it('should return undefined for invalid message', () => {
        expect(getMethod({})).toBeUndefined();
      });
    });

    describe('getParams', () => {
      it('should extract params from request message', () => {
        expect(getParams(requestMessage)).toEqual(TEST_PARAMS);
      });

      it('should return empty array for non-request messages', () => {
        expect(getParams(responseMessage)).toEqual([]);
        expect(getParams(errorMessage)).toEqual([]);
        expect(getParams({})).toEqual([]);
      });
    });

    describe('getResult', () => {
      it('should extract result from response message', () => {
        expect(getResult(responseMessage)).toEqual(TEST_RESULT);
      });

      it('should return empty array for non-response messages', () => {
        expect(getResult(requestMessage)).toEqual([]);
        expect(getResult(errorMessage)).toEqual([]);
        expect(getResult({})).toEqual([]);
      });
    });

    describe('getError', () => {
      it('should extract error details from error message', () => {
        const error = getError(errorMessage);
        expect(error).toBeDefined();
        expect(error?.code).toBe(TEST_ERROR_CODE);
        expect(error?.message).toBe(TEST_ERROR_MESSAGE);
      });

      it('should return undefined for non-error messages', () => {
        expect(getError(requestMessage)).toBeUndefined();
        expect(getError(responseMessage)).toBeUndefined();
        expect(getError({})).toBeUndefined();
      });
    });
  });
});