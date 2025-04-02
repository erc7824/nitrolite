import { createAppLogic, encoders, StateValidators, AppStatus } from '../../src/utils';
import { AppLogic } from '../../src/types';

describe('Application Logic Utilities', () => {
  const testAdjudicatorAddress = '0x1111111111111111111111111111111111111111';
  
  describe('createAppLogic', () => {
    it('should create a basic AppLogic instance', () => {
      const encode = jest.fn().mockReturnValue('0x1234');
      const decode = jest.fn().mockReturnValue({ value: 123 });
      
      const appLogic = createAppLogic({
        adjudicatorAddress: testAdjudicatorAddress,
        encode,
        decode
      });
      
      expect(appLogic).toBeDefined();
      expect(appLogic.encode).toBe(encode);
      expect(appLogic.decode).toBe(decode);
      expect(appLogic.getAdjudicatorAddress()).toBe(testAdjudicatorAddress);
      expect(appLogic.validateTransition).toBeUndefined();
      expect(appLogic.isFinal).toBeUndefined();
    });
    
    it('should include optional functions when provided', () => {
      const encode = jest.fn().mockReturnValue('0x1234');
      const decode = jest.fn().mockReturnValue({ value: 123 });
      const validateTransition = jest.fn().mockReturnValue(true);
      const isFinal = jest.fn().mockReturnValue(false);
      
      const appLogic = createAppLogic({
        adjudicatorAddress: testAdjudicatorAddress,
        encode,
        decode,
        validateTransition,
        isFinal
      });
      
      expect(appLogic.validateTransition).toBe(validateTransition);
      expect(appLogic.isFinal).toBe(isFinal);
    });
    
    it('should set adjudicator type when provided', () => {
      const encode = jest.fn().mockReturnValue('0x1234');
      const decode = jest.fn().mockReturnValue({ value: 123 });
      const adjudicatorType = 'custom';
      
      const appLogic = createAppLogic({
        adjudicatorAddress: testAdjudicatorAddress,
        adjudicatorType,
        encode,
        decode
      });
      
      expect(appLogic.getAdjudicatorType).toBeDefined();
      expect(appLogic.getAdjudicatorType!()).toBe(adjudicatorType);
    });
  });
  
  describe('encoders', () => {
    it('should encode numeric values', () => {
      const encoded = encoders.numeric(BigInt(42));
      expect(encoded).toBeDefined();
      expect(typeof encoded).toBe('string');
      expect(encoded.startsWith('0x')).toBe(true);
    });
    
    it('should encode sequential values', () => {
      const encoded = encoders.sequential(BigInt(1), BigInt(42));
      expect(encoded).toBeDefined();
      expect(typeof encoded).toBe('string');
      expect(encoded.startsWith('0x')).toBe(true);
    });
    
    it('should encode turn-based values', () => {
      const encoded = encoders.turnBased({}, 1, 2, true);
      expect(encoded).toBeDefined();
      expect(typeof encoded).toBe('string');
      expect(encoded.startsWith('0x')).toBe(true);
    });
    
    it('should create empty state', () => {
      const empty = encoders.empty();
      expect(empty).toBe('0x');
    });
  });
  
  describe('StateValidators', () => {
    describe('turnBased', () => {
      it('should validate turn transitions', () => {
        const getTurn = jest.fn()
          .mockReturnValueOnce(0) // prevState turn = 0
          .mockReturnValueOnce(1); // nextState turn = 1
        
        const validator = StateValidators.turnBased(getTurn);
        
        const signer = '0x2222222222222222222222222222222222222222';
        const roles = [
          '0x1111111111111111111111111111111111111111', // role 0
          '0x2222222222222222222222222222222222222222'  // role 1
        ];
        
        const result = validator({}, {}, signer, roles);
        
        expect(result).toBe(true);
        expect(getTurn).toHaveBeenCalledTimes(2);
      });
      
      it('should reject invalid turn transitions', () => {
        const getTurn = jest.fn()
          .mockReturnValueOnce(0) // prevState turn = 0
          .mockReturnValueOnce(0); // nextState turn = 0 (should be 1)
        
        const validator = StateValidators.turnBased(getTurn);
        
        const signer = '0x2222222222222222222222222222222222222222';
        const roles = [
          '0x1111111111111111111111111111111111111111', // role 0
          '0x2222222222222222222222222222222222222222'  // role 1
        ];
        
        const result = validator({}, {}, signer, roles);
        
        expect(result).toBe(false);
      });
      
      it('should reject when wrong signer', () => {
        const getTurn = jest.fn()
          .mockReturnValueOnce(0) // prevState turn = 0
          .mockReturnValueOnce(1); // nextState turn = 1
        
        const validator = StateValidators.turnBased(getTurn);
        
        const signer = '0x3333333333333333333333333333333333333333'; // wrong signer
        const roles = [
          '0x1111111111111111111111111111111111111111', // role 0
          '0x2222222222222222222222222222222222222222'  // role 1
        ];
        
        const result = validator({}, {}, signer, roles);
        
        expect(result).toBe(false);
      });
    });
    
    describe('sequential', () => {
      it('should validate sequential transitions', () => {
        const getSequence = jest.fn()
          .mockReturnValueOnce(BigInt(1)) // prevState sequence = 1
          .mockReturnValueOnce(BigInt(2)); // nextState sequence = 2
        
        const getValue = jest.fn()
          .mockReturnValueOnce(BigInt(10)) // prevState value = 10
          .mockReturnValueOnce(BigInt(20)); // nextState value = 20
        
        const validator = StateValidators.sequential(getSequence, getValue);
        
        const signer = '0x1111111111111111111111111111111111111111';
        const initiator = '0x1111111111111111111111111111111111111111';
        
        const result = validator({}, {}, signer, initiator);
        
        expect(result).toBe(true);
        expect(getSequence).toHaveBeenCalledTimes(2);
        expect(getValue).toHaveBeenCalledTimes(2);
      });
      
      it('should reject when sequence does not increase', () => {
        const getSequence = jest.fn()
          .mockReturnValueOnce(BigInt(2)) // prevState sequence = 2
          .mockReturnValueOnce(BigInt(2)); // nextState sequence = 2 (not increased)
        
        const getValue = jest.fn()
          .mockReturnValueOnce(BigInt(10)) // prevState value = 10
          .mockReturnValueOnce(BigInt(20)); // nextState value = 20
        
        const validator = StateValidators.sequential(getSequence, getValue);
        
        const signer = '0x1111111111111111111111111111111111111111';
        const initiator = '0x1111111111111111111111111111111111111111';
        
        const result = validator({}, {}, signer, initiator);
        
        expect(result).toBe(false);
      });
      
      it('should reject when value decreases', () => {
        const getSequence = jest.fn()
          .mockReturnValueOnce(BigInt(1)) // prevState sequence = 1
          .mockReturnValueOnce(BigInt(2)); // nextState sequence = 2
        
        const getValue = jest.fn()
          .mockReturnValueOnce(BigInt(20)) // prevState value = 20
          .mockReturnValueOnce(BigInt(10)); // nextState value = 10 (decreased)
        
        const validator = StateValidators.sequential(getSequence, getValue);
        
        const signer = '0x1111111111111111111111111111111111111111';
        const initiator = '0x1111111111111111111111111111111111111111';
        
        const result = validator({}, {}, signer, initiator);
        
        expect(result).toBe(false);
      });
      
      it('should reject when wrong signer', () => {
        const getSequence = jest.fn()
          .mockReturnValueOnce(BigInt(1)) // prevState sequence = 1
          .mockReturnValueOnce(BigInt(2)); // nextState sequence = 2
        
        const getValue = jest.fn()
          .mockReturnValueOnce(BigInt(10)) // prevState value = 10
          .mockReturnValueOnce(BigInt(20)); // nextState value = 20
        
        const validator = StateValidators.sequential(getSequence, getValue);
        
        const signer = '0x2222222222222222222222222222222222222222';
        const initiator = '0x1111111111111111111111111111111111111111';
        
        const result = validator({}, {}, signer, initiator);
        
        expect(result).toBe(false);
      });
    });
  });
});