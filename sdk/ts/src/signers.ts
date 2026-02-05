/**
 * Signer implementations for Nitrolite SDK
 * Provides EthereumMsgSigner and EthereumRawSigner matching the Go SDK patterns
 */

import { Address, Hex } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';

/**
 * StateSigner interface for signing channel states
 * Used for signing off-chain state updates
 */
export interface StateSigner {
  /** Get the address of the signer */
  getAddress(): Address;
  /** Sign a message hash (used for EIP-191 message signing) */
  signMessage(hash: Hex): Promise<Hex>;
}

/**
 * TransactionSigner interface for signing blockchain transactions
 * Used for on-chain operations (deposits, withdrawals, etc.)
 */
export interface TransactionSigner {
  /** Get the address of the signer */
  getAddress(): Address;
  /** Send a transaction to the blockchain */
  sendTransaction(tx: any): Promise<Hex>;
  /** Sign a message (raw bytes) */
  signMessage(message: { raw: Hex }): Promise<Hex>;
}

/**
 * EthereumMsgSigner implements StateSigner using EIP-191 message signing
 * Corresponds to Go SDK's sign.NewEthereumMsgSigner
 *
 * This signer prepends "\x19Ethereum Signed Message:\n" before signing,
 * making it compatible with eth_sign and personal_sign RPC methods.
 *
 * @example
 * ```typescript
 * import { privateKeyToAccount } from 'viem/accounts';
 * import { EthereumMsgSigner } from '@nitrolite/sdk';
 *
 * const account = privateKeyToAccount('0x...');
 * const signer = new EthereumMsgSigner(account);
 * ```
 */
export class EthereumMsgSigner implements StateSigner {
  private account: ReturnType<typeof privateKeyToAccount>;

  constructor(privateKeyOrAccount: Hex | ReturnType<typeof privateKeyToAccount>) {
    if (typeof privateKeyOrAccount === 'string') {
      this.account = privateKeyToAccount(privateKeyOrAccount);
    } else {
      this.account = privateKeyOrAccount;
    }
  }

  getAddress(): Address {
    return this.account.address;
  }

  /**
   * Sign a message hash using EIP-191 (with Ethereum message prefix)
   * The message is automatically prefixed with "\x19Ethereum Signed Message:\n"
   */
  async signMessage(hash: Hex): Promise<Hex> {
    return await this.account.signMessage({
      message: { raw: hash },
    });
  }
}

/**
 * EthereumRawSigner implements TransactionSigner using raw ECDSA signing
 * Corresponds to Go SDK's sign.NewEthereumRawSigner
 *
 * This signer signs raw hashes directly without any prefix,
 * making it suitable for transaction signing and EIP-712 typed data.
 *
 * @example
 * ```typescript
 * import { privateKeyToAccount } from 'viem/accounts';
 * import { EthereumRawSigner } from '@nitrolite/sdk';
 *
 * const account = privateKeyToAccount('0x...');
 * const signer = new EthereumRawSigner(account);
 * ```
 */
export class EthereumRawSigner implements TransactionSigner {
  private account: ReturnType<typeof privateKeyToAccount>;

  constructor(privateKeyOrAccount: Hex | ReturnType<typeof privateKeyToAccount>) {
    if (typeof privateKeyOrAccount === 'string') {
      this.account = privateKeyToAccount(privateKeyOrAccount);
    } else {
      this.account = privateKeyOrAccount;
    }
  }

  getAddress(): Address {
    return this.account.address;
  }

  /**
   * Send a transaction to the blockchain
   */
  async sendTransaction(tx: any): Promise<Hex> {
    throw new Error('sendTransaction requires a wallet client - use the blockchain client instead');
  }

  /**
   * Sign a message (raw bytes without prefix)
   */
  async signMessage(message: { raw: Hex }): Promise<Hex> {
    return await this.account.sign({ hash: message.raw });
  }
}

/**
 * Helper function to create signers from a private key
 *
 * @param privateKey - Hex-encoded private key
 * @returns Object containing both state and transaction signers
 *
 * @example
 * ```typescript
 * import { createSigners } from '@nitrolite/sdk';
 *
 * const { stateSigner, txSigner } = createSigners('0x...');
 * const client = await Client.create(wsURL, stateSigner, txSigner);
 * ```
 */
export function createSigners(privateKey: Hex): {
  stateSigner: StateSigner;
  txSigner: TransactionSigner;
} {
  const account = privateKeyToAccount(privateKey);
  return {
    stateSigner: new EthereumMsgSigner(account),
    txSigner: new EthereumRawSigner(account),
  };
}
