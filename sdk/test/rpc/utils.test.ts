/**
 * @jest-environment node
 * @skip
 */
import { 
  getCurrentTimestamp,
  generateRequestId,
  getRequestId,
  getMethod,
  getParams,
  getResult,
  getError,
  getTimestamp,
  toBytes,
  isValidResponseTimestamp,
  isValidResponseRequestId
} from '../../src/rpc/utils';
import { Hex, stringToHex } from 'viem';
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
  const TEST_TIMESTAMP = 1709584837123;
  const LATER_TIMESTAMP = TEST_TIMESTAMP + 1000;

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
        req: [TEST_REQUEST_ID, TEST_METHOD, TEST_PARAMS, TEST_TIMESTAMP]
      };
      
      responseMessage = {
        res: [TEST_REQUEST_ID, TEST_METHOD, TEST_RESULT, LATER_TIMESTAMP]
      };
      
      errorMessage = {
        err: [TEST_REQUEST_ID, TEST_ERROR_CODE, TEST_ERROR_MESSAGE, TEST_TIMESTAMP]
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

    describe('getTimestamp', () => {
      it('should extract timestamp from request message', () => {
        expect(getTimestamp(requestMessage)).toBe(TEST_TIMESTAMP);
      });

      it('should extract timestamp from response message', () => {
        expect(getTimestamp(responseMessage)).toBe(LATER_TIMESTAMP);
      });

      it('should extract timestamp from error message', () => {
        expect(getTimestamp(errorMessage)).toBe(TEST_TIMESTAMP);
      });

      it('should return undefined for invalid message', () => {
        expect(getTimestamp({})).toBeUndefined();
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
  
  describe('toBytes', () => {
    it('should convert string values to hex format', () => {
      const values = ['test', 'hello'];
      const result = toBytes(values);
      
      expect(result.length).toBe(2);
      expect(result[0]).toBe(stringToHex('test'));
      expect(result[1]).toBe(stringToHex('hello'));
    });
    
    it('should convert object values to hex formatted JSON strings', () => {
      const values = [{ foo: 'bar' }, 123, true];
      const result = toBytes(values);
      
      expect(result.length).toBe(3);
      expect(result[0]).toBe(stringToHex(JSON.stringify({ foo: 'bar' })));
      expect(result[1]).toBe(stringToHex(JSON.stringify(123)));
      expect(result[2]).toBe(stringToHex(JSON.stringify(true)));
    });
  });

  
  describe('Message validation utilities', () => {
    describe('isValidResponseTimestamp', () => {
      it('should return true when response timestamp is greater than request timestamp', () => {
        const request = { req: [TEST_REQUEST_ID, TEST_METHOD, TEST_PARAMS, TEST_TIMESTAMP] };
        const response = { res: [TEST_REQUEST_ID, TEST_METHOD, TEST_RESULT, LATER_TIMESTAMP] };
        
        expect(isValidResponseTimestamp(request, response)).toBe(true);
      });
      
      it('should return false when response timestamp is equal to request timestamp', () => {
        const request = { req: [TEST_REQUEST_ID, TEST_METHOD, TEST_PARAMS, TEST_TIMESTAMP] };
        const response = { res: [TEST_REQUEST_ID, TEST_METHOD, TEST_RESULT, TEST_TIMESTAMP] };
        
        expect(isValidResponseTimestamp(request, response)).toBe(false);
      });
      
      it('should return false when response timestamp is less than request timestamp', () => {
        const request = { req: [TEST_REQUEST_ID, TEST_METHOD, TEST_PARAMS, LATER_TIMESTAMP] };
        const response = { res: [TEST_REQUEST_ID, TEST_METHOD, TEST_RESULT, TEST_TIMESTAMP] };
        
        expect(isValidResponseTimestamp(request, response)).toBe(false);
      });
      
      it('should return false for invalid messages', () => {
        const invalidRequest = { x: 123 };
        const invalidResponse = { y: 456 };
        
        expect(isValidResponseTimestamp(invalidRequest as any, invalidResponse as any)).toBe(false);
      });
    });
    
    describe('isValidResponseRequestId', () => {
      it('should return true when response request ID matches request ID', () => {
        const request = { req: [TEST_REQUEST_ID, TEST_METHOD, TEST_PARAMS, TEST_TIMESTAMP] };
        const response = { res: [TEST_REQUEST_ID, TEST_METHOD, TEST_RESULT, LATER_TIMESTAMP] };
        
        expect(isValidResponseRequestId(request, response)).toBe(true);
      });
      
      it('should return false when response request ID does not match request ID', () => {
        const request = { req: [TEST_REQUEST_ID, TEST_METHOD, TEST_PARAMS, TEST_TIMESTAMP] };
        const response = { res: [TEST_REQUEST_ID + 1, TEST_METHOD, TEST_RESULT, LATER_TIMESTAMP] };
        
        expect(isValidResponseRequestId(request, response)).toBe(false);
      });
      
      it('should return false for invalid messages', () => {
        const invalidRequest = { x: 123 };
        const invalidResponse = { y: 456 };
        
        expect(isValidResponseRequestId(invalidRequest as any, invalidResponse as any)).toBe(false);
      });
    });
  });
});