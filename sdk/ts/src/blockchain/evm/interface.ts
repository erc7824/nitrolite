/**
 * Blockchain EVM interfaces and type definitions
 */

import { Address, PublicClient, WalletClient } from 'viem';

/**
 * AssetStore provides methods to retrieve asset and token information
 */
export interface AssetStore {
  /**
   * GetAssetDecimals returns the decimal places for an asset
   */
  getAssetDecimals(asset: string): Promise<number>;

  /**
   * GetTokenDecimals returns the decimal places for a token on a specific blockchain
   */
  getTokenDecimals(blockchainId: bigint, tokenAddress: Address): Promise<number>;

  /**
   * GetTokenAddress returns the token contract address for an asset on a specific blockchain
   */
  getTokenAddress(asset: string, blockchainId: bigint): Promise<Address>;
}

/**
 * EVMClient wraps viem's PublicClient for blockchain interactions
 */
export interface EVMClient extends PublicClient {}

/**
 * WalletSigner wraps viem's WalletClient for transaction signing
 */
export interface WalletSigner extends WalletClient {}
