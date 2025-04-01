/**
 * RPC utilities and helper functions
 */

import { Hex, Address, hashMessage } from 'viem';
import { Logger } from '../config';
import Errors from '../errors';

/**
 * Create a payload for signing from any data
 * @param data The data to sign
 * @returns The payload as a hex string
 */
export function createPayload(data: any): Hex {
  // Convert the data to a string and hash it
  return hashMessage(JSON.stringify(data));
}

/**
 * Verify a signature
 * @param payload The payload that was signed
 * @param signature The signature
 * @param expectedSigner The expected signer address
 * @param logger Optional logger for debugging
 * @returns True if the signature is valid
 */
export async function verifySignature(
  payload: Hex, 
  signature: Hex, 
  expectedSigner: Address, 
  logger?: Logger
): Promise<boolean> {
  try {
    if (logger) {
      logger.debug('Verifying signature', { 
        expectedSigner,
        payloadLength: payload.length
      });
    }
    
    // Import verification utilities from viem
    // Use the verifyMessage function to check if the signature is valid
    // The function confirms that the recovered address matches the expected signer
    const { verifyMessage } = await import('viem');
    
    // Verify that the signature was created by the expected signer
    const valid = await verifyMessage({
      address: expectedSigner,
      message: payload,
      signature: signature,
    });
    
    if (logger) {
      logger.debug('Signature verification result', {
        valid,
        expectedSigner
      });
    }
    
    return valid;
  } catch (error) {
    if (logger) {
      logger.error('Error verifying signature', { 
        expectedSigner, 
        error 
      });
    }
    return false;
  }
}

/**
 * Validate that an RPC client is connected
 * @param handlerRegistered Whether a handler is registered
 * @throws {ProviderNotConnectedError} If the client is not connected
 */
export function validateConnection(handlerRegistered: boolean): void {
  if (!handlerRegistered) {
    throw new Errors.ProviderNotConnectedError('RPC Client');
  }
}

/**
 * Validate required parameters
 * @param params Object containing parameters to validate
 */
export function validateRequiredParams(params: Record<string, any>): void {
  for (const [name, value] of Object.entries(params)) {
    if (!value) {
      throw new Errors.MissingParameterError(name);
    }
  }
}