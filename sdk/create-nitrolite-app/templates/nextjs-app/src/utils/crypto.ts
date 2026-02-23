import { generatePrivateKey, privateKeyToAccount } from 'viem/accounts';
import { getAddress, keccak256, toBytes } from 'viem';

export interface CryptoKeypair {
  privateKey: string;
  address: string;
}

/**
 * Generates a random keypair using viem
 */
export const generateKeyPair = async (): Promise<CryptoKeypair> => {
  const privateKey = generatePrivateKey();
  const privateKeyHash = keccak256(toBytes(privateKey));
  const account = privateKeyToAccount(privateKeyHash);

  return {
    privateKey: privateKeyHash,
    address: getAddress(account.address),
  };
};
