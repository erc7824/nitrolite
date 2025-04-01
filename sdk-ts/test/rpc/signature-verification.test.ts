import { verifySignature, createPayload } from '../../src/rpc/utils';
import { hexToBytes, stringToHex, toHex, createWalletClient, hashMessage } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { createLogger } from '../../src/config';

// Mock wallet for testing
const privateKey = '0x1234567890123456789012345678901234567890123456789012345678901234';
const account = privateKeyToAccount(privateKey);

// Create a logger that doesn't output during tests
const logger = createLogger({ level: 'none' });

describe('RPC Signature Verification', () => {
  it('should verify a valid signature', async () => {
    // Create a test payload
    const data = { test: 'data', value: 123 };
    const payload = createPayload(data);
    
    // Sign the payload
    const signature = await account.signMessage({ message: { raw: payload } });
    
    // Verify the signature
    const isValid = await verifySignature(
      payload,
      signature as `0x${string}`,
      account.address,
      logger
    );
    
    expect(isValid).toBe(true);
  });
  
  it('should reject an invalid signature', async () => {
    // Create a test payload
    const data = { test: 'data', value: 123 };
    const payload = createPayload(data);
    
    // Generate a different payload to get an invalid signature
    const differentData = { test: 'different', value: 456 };
    const differentPayload = createPayload(differentData);
    
    // Sign the different payload
    const invalidSignature = await account.signMessage({ message: { raw: differentPayload } });
    
    // Verify the signature against the original payload (should fail)
    const isValid = await verifySignature(
      payload,
      invalidSignature as `0x${string}`,
      account.address,
      logger
    );
    
    expect(isValid).toBe(false);
  });
  
  it('should reject if signer address is wrong', async () => {
    // Create a test payload
    const data = { test: 'data', value: 123 };
    const payload = createPayload(data);
    
    // Sign the payload
    const signature = await account.signMessage({ message: { raw: payload } });
    
    // Use a different address for verification
    const wrongAddress = '0x1111111111111111111111111111111111111111';
    
    // Verify the signature with the wrong address
    const isValid = await verifySignature(
      payload,
      signature as `0x${string}`,
      wrongAddress,
      logger
    );
    
    expect(isValid).toBe(false);
  });
});