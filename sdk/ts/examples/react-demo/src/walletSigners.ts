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
    console.log('📍 WalletStateSigner.getAddress called');
    if (!this.walletClient.account?.address) {
      console.error('❌ No account address in wallet client');
      throw new Error('Wallet client does not have an account address');
    }
    console.log('  Address:', this.walletClient.account.address);
    return this.walletClient.account.address;
  }

  /**
   * Sign a message hash using EIP-191 (with Ethereum message prefix)
   */
  async signMessage(hash: Hex): Promise<Hex> {
    console.log('🔐 WalletStateSigner.signMessage called');
    console.log('  Hash to sign:', hash);

    if (!this.walletClient.account) {
      console.error('❌ No account in wallet client');
      throw new Error('Wallet client does not have an account');
    }

    console.log('  Account:', this.walletClient.account.address);
    console.log('⏳ Requesting signature from wallet...');

    const signature = await this.walletClient.signMessage({
      account: this.walletClient.account,
      message: { raw: hash },
    });

    console.log('✅ Message signed successfully');
    console.log('  Signature:', signature);
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
    console.log('📍 WalletTransactionSigner.getAddress called');
    if (!this.walletClient.account?.address) {
      console.error('❌ No account address in wallet client');
      throw new Error('Wallet client does not have an account address');
    }
    console.log('  Address:', this.walletClient.account.address);
    return this.walletClient.account.address;
  }

  /**
   * Send a transaction to the blockchain
   */
  async sendTransaction(tx: any): Promise<Hex> {
    console.log('⚠️ WalletTransactionSigner.sendTransaction called (this should not be used)');
    console.log('  Transaction data:', tx);
    throw new Error('sendTransaction requires a wallet client - use the blockchain client instead');
  }

  /**
   * Sign a message (required by TransactionSigner interface)
   * This wraps the signRaw functionality
   */
  async signMessage(message: { raw: Hex }): Promise<Hex> {
    console.log('🔐 WalletTransactionSigner.signMessage called');
    console.log('  Message hash:', message.raw);
    return await this.signRaw(message.raw);
  }

  /**
   * Sign a message (raw bytes without prefix)
   */
  async signRaw(hash: Hex): Promise<Hex> {
    console.log('🔐 WalletTransactionSigner.signRaw called');
    console.log('  Hash to sign:', hash);

    if (!this.walletClient.account) {
      console.error('❌ No account in wallet client');
      throw new Error('Wallet client does not have an account');
    }

    console.log('  Account:', this.walletClient.account.address);
    console.log('⏳ Requesting typed data signature from wallet...');

    // Sign the hash directly without EIP-191 prefix
    // MetaMask doesn't have a direct "sign raw" method, so we use signTypedData with a minimal schema
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

    console.log('✅ Typed data signed successfully');
    console.log('  Signature:', signature);
    return signature;
  }
}
