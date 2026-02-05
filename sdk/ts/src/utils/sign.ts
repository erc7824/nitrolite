/**
 * Signing utilities for raw ECDSA messages
 */

import { Hex } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';

/**
 * Sign a raw message using ECDSA with a private key
 * @param message - The message to sign (as hex string)
 * @param privateKey - The private key to use for signing
 * @returns The signature as a hex string
 */
export async function signRawECDSAMessage(message: Hex, privateKey: Hex): Promise<Hex> {
  const account = privateKeyToAccount(privateKey);
  const signature = await account.signMessage({
    message: { raw: message },
  });
  return signature;
}
