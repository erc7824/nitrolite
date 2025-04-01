import {
  Account,
  Address,
  PublicClient,
  WalletClient,
  Abi
} from 'viem';
import { ContractAddresses } from '../abis';
import { Logger } from '../config';

/**
 * Configuration options for the Nitrolite client
 */
export interface NitroliteClientConfig {
  /**
   * Public client for reading from the blockchain
   * Required to interact with the blockchain
   */
  publicClient: PublicClient;
  
  /**
   * Wallet client for sending transactions
   * Required for operations that modify the blockchain state
   */
  walletClient?: WalletClient;
  
  /**
   * Signer account
   * Required for operations that modify the blockchain state
   */
  account?: Account;
  
  /**
   * Chain ID of the network being used
   * If not provided, will use the chain from the publicClient
   */
  chainId?: number;
  
  /**
   * Contract custody address (legacy parameter)
   * Use 'addresses' property instead when possible
   */
  custodyAddress?: Address;
  
  /**
   * Contract addresses for the Nitrolite infrastructure
   * Either this or custodyAddress must be provided
   */
  addresses?: ContractAddresses;
  
  /**
   * Custom adjudicator ABIs by type
   * If provided, these will override the default AdjudicatorAbi
   */
  adjudicatorAbis?: Record<string, Abi>;

  /**
   * Logger instance for client logs
   * If not provided, the default console logger will be used
   */
  logger?: Logger;
}