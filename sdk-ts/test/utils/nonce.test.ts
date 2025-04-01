import { generateChannelNonce } from '../../src/utils';
import { Address } from 'viem';

describe('Nonce Generation', () => {
  describe('generateChannelNonce', () => {
    it('should generate a BigInt value', () => {
      const nonce = generateChannelNonce();
      expect(typeof nonce).toBe('bigint');
    });
    
    it('should generate different values on consecutive calls', () => {
      const nonce1 = generateChannelNonce();
      const nonce2 = generateChannelNonce();
      expect(nonce1).not.toEqual(nonce2);
    });
    
    it('should incorporate address data when provided', () => {
      const address = '0x1234567890123456789012345678901234567890' as Address;
      
      const withAddress = generateChannelNonce(address);
      const withoutAddress = generateChannelNonce();
      
      expect(withAddress).not.toEqual(withoutAddress);
      
      // Same address should produce different nonces due to randomness and timestamp
      const withSameAddress = generateChannelNonce(address);
      expect(withSameAddress).not.toEqual(withAddress);
    });
    
    it('should generate values greater than zero', () => {
      const nonce = generateChannelNonce();
      expect(nonce > 0n).toBe(true);
    });
  });
});