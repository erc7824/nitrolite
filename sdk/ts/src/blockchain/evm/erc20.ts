/**
 * ERC20 token contract wrapper
 */

import { Address } from 'viem';
import { Erc20Abi } from './erc20_abi';
import { EVMClient } from './interface';

/**
 * ERC20 contract wrapper for token interactions
 */
export class ERC20 {
  private tokenAddress: Address;
  private client: EVMClient;

  constructor(tokenAddress: Address, client: EVMClient) {
    this.tokenAddress = tokenAddress;
    this.client = client;
  }

  /**
   * Get the token balance of an account
   */
  async balanceOf(account: Address): Promise<bigint> {
    return this.client.readContract({
      address: this.tokenAddress,
      abi: Erc20Abi,
      functionName: 'balanceOf',
      args: [account],
    }) as Promise<bigint>;
  }

  /**
   * Get the allowance granted by owner to spender
   */
  async allowance(owner: Address, spender: Address): Promise<bigint> {
    return this.client.readContract({
      address: this.tokenAddress,
      abi: Erc20Abi,
      functionName: 'allowance',
      args: [owner, spender],
    }) as Promise<bigint>;
  }

  /**
   * Get the decimals of the token
   */
  async decimals(): Promise<number> {
    // decimals() is a standard ERC20 function but not in our minimal ABI
    // We'll need to call it directly if needed
    throw new Error('decimals() not available in minimal ERC20 ABI');
  }
}

/**
 * Create a new ERC20 contract instance
 */
export function newERC20(tokenAddress: Address, client: EVMClient): ERC20 {
  return new ERC20(tokenAddress, client);
}
