/**
 * Wallet signers that work with viem WalletClient (MetaMask, etc.)
 * These adapt WalletClient to work with the Nitrolite SDK signer interfaces
 */

import type { WalletClient } from 'viem';
import type { Address, Hex } from 'viem';

/**
 * WalletStateSigner implements StateSigner using a viem WalletClient
 * This is suitable for MetaMask and other browser wallets
 */
export class WalletStateSigner {
  private walletClient: WalletClient;

  constructor(walletClient: WalletClient) {
    this.walletClient = walletClient;
  }

  getAddress(): Address {
    if (!this.walletClient.account?.address) {
      throw new Error('Wallet client does not have an account address');
    }
    return this.walletClient.account.address;
  }

  /**
   * Sign a message hash using EIP-191 (with Ethereum message prefix)
   */
  async signMessage(hash: Hex): Promise<Hex> {
    if (!this.walletClient.account) {
      throw new Error('Wallet client does not have an account');
    }

    const signature = await this.walletClient.signMessage({
      account: this.walletClient.account,
      message: { raw: hash },
    });

    return signature;
  }
}

/**
 * WalletTransactionSigner implements TransactionSigner using a viem WalletClient
 * This is suitable for MetaMask and other browser wallets
 */
export class WalletTransactionSigner {
  private walletClient: WalletClient;

  constructor(walletClient: WalletClient) {
    this.walletClient = walletClient;
  }

  getAddress(): Address {
    if (!this.walletClient.account?.address) {
      throw new Error('Wallet client does not have an account address');
    }
    return this.walletClient.account.address;
  }

  /**
   * Send a transaction to the blockchain
   */
  async sendTransaction(_tx: any): Promise<Hex> {
    throw new Error('sendTransaction requires a wallet client - use the blockchain client instead');
  }

  /**
   * Sign a message (required by TransactionSigner interface)
   * This wraps the signRaw functionality
   */
  async signMessage(message: { raw: Hex }): Promise<Hex> {
    return await this.signRaw(message.raw);
  }

  /**
   * Sign a message (raw bytes without prefix)
   */
  async signRaw(hash: Hex): Promise<Hex> {
    if (!this.walletClient.account) {
      throw new Error('Wallet client does not have an account');
    }

    const signature = await this.walletClient.signTypedData({
      account: this.walletClient.account,
      domain: {
        name: 'Nitrolite',
        version: '1',
        chainId: 1,
      },
      types: {
        Message: [{ name: 'data', type: 'bytes32' }],
      },
      primaryType: 'Message',
      message: {
        data: hash,
      },
    });

    return signature;
  }
}
