import {
  Account,
  Address,
  PublicClient,
  WalletClient,
  Abi,
  Hex
} from 'viem';
import {
  Channel,
  State,
  ChannelId,
  AppDataTypes,
  AppLogic
} from '../types';
import { 
  AdjudicatorAbi,
  defaultAbiConfig,
  ContractAddresses
} from '../abis';
import Errors from '../errors'; // Import Errors
import { Logger, defaultLogger } from '../config';

import { HachiClientConfig } from './config';
import { ChannelOperations } from './operations';
import { 
  ChannelContext,
  createNumericChannel,
  createSequentialChannel,
  createCustomChannel
} from './channels';
import { generateChannelNonce } from '../utils';

/**
 * Main client for interacting with Hachi contracts
 */
export class HachiClient {
  public readonly publicClient: PublicClient;
  public readonly walletClient?: WalletClient;
  public readonly account?: Account;
  public readonly chainId: number;
  public readonly addresses: ContractAddresses;
  public readonly adjudicatorAbis: Record<string, Abi>;
  public readonly logger: Logger;
  
  private readonly operations: ChannelOperations;

  constructor(config: HachiClientConfig) {
    // TODO: Add more comprehensive configuration validation (e.g., address formats)
    if (!config.publicClient) {
      throw new Errors.MissingParameterError('publicClient');
    }
    
    // Use chain ID from the public client if not explicitly provided
    let chainId = config.chainId;
    if (!chainId) {
      chainId = config.publicClient.chain?.id;
      if (!chainId) {
        throw new Errors.MissingParameterError('chainId');
      }
    }
    
    // Prefer using the 'addresses' object, 'custodyAddress' is for backward compatibility
    if (!config.addresses && !config.custodyAddress) {
      throw new Errors.MissingParameterError('addresses or custodyAddress');
    }
    
    this.publicClient = config.publicClient;
    this.walletClient = config.walletClient;
    this.account = config.account;
    this.chainId = chainId;
    this.logger = config.logger || defaultLogger;
    
    // Use provided addresses or create from custody address
    if (config.addresses) {
      this.addresses = config.addresses;
    } else {
      // Backwards compatibility for custodyAddress
      this.addresses = {
        custody: config.custodyAddress as Address,
        adjudicators: {}
      };
    }
    
    // Make sure adjudicators object exists
    if (!this.addresses.adjudicators) {
      this.addresses.adjudicators = {};
    }
    
    // Initialize adjudicator ABIs with defaults
    this.adjudicatorAbis = {
      'base': AdjudicatorAbi,
      ...(config.adjudicatorAbis || {})
    };
    
    // Initialize channel operations
    this.operations = new ChannelOperations(
      this.publicClient,
      this.walletClient,
      this.account,
      this.custodyAddress,
      this.logger
    );
  }
  
  /**
   * Register a custom adjudicator ABI
   * @param type Adjudicator type name
   * @param abi Custom ABI for the adjudicator
   */
  registerAdjudicatorAbi(type: string, abi: Abi): void {
    this.adjudicatorAbis[type] = abi;
  }
  
  /**
   * Get an adjudicator ABI by type
   * @param type The adjudicator type
   * @returns The adjudicator ABI
   */
  getAdjudicatorAbi(type: string = 'base'): Abi {
    const abi = this.adjudicatorAbis[type];
    if (!abi) {
      // Fall back to base adjudicator ABI if specific type not found
      return this.adjudicatorAbis['base'] || AdjudicatorAbi;
    }
    return abi;
  }
  
  /**
   * Get the custody contract address
   */
  get custodyAddress(): Address {
    return this.addresses.custody;
  }
  
  /**
   * Get an adjudicator address by type
   * @param type The adjudicator type
   * @param fallbackToBase Whether to fall back to the base adjudicator if type not found
   * @returns The adjudicator address
   */
  getAdjudicatorAddress(type: string, fallbackToBase: boolean = true): Address {
    // First try to get the requested adjudicator type
    const address = this.addresses.adjudicators[type];
    if (address) {
      return address;
    }
    
    // If requested to fall back and base adjudicator exists, use it
    if (fallbackToBase && this.addresses.adjudicators['base']) {
      return this.addresses.adjudicators['base'];
    }
    
    // Otherwise throw an error with helpful message
    throw new Errors.ContractNotFoundError(
      `Adjudicator type: ${type}`,
      {
        availableTypes: Object.keys(this.addresses.adjudicators),
        requestedType: type
      }
    );
  }
  
  // ======== Channel operations ========
  
  /**
   * Open a new channel or join an existing one
   */
  async openChannel(channel: Channel, deposit: State): Promise<ChannelId> {
    return this.operations.openChannel(channel, deposit);
  }
  
  /**
   * Close a channel with a mutually signed state
   */
  async closeChannel(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<void> {
    return this.operations.closeChannel(channelId, candidate, proofs);
  }
  
  /**
   * Challenge a channel when the counterparty is unresponsive
   */
  async challengeChannel(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<void> {
    return this.operations.challengeChannel(channelId, candidate, proofs);
  }
  
  /**
   * Checkpoint a state to store it on-chain
   */
  async checkpointChannel(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<void> {
    return this.operations.checkpointChannel(channelId, candidate, proofs);
  }
  
  /**
   * Reclaim funds after challenge period expires
   */
  async reclaimChannel(channelId: ChannelId): Promise<void> {
    return this.operations.reclaimChannel(channelId);
  }
  
  /**
   * Approve tokens for the custody contract
   */
  async approveTokens(tokenAddress: Address, amount: bigint, spender: Address): Promise<void> {
    return this.operations.approveTokens(tokenAddress, amount, spender);
  }
  
  /**
   * Get token allowance
   */
  async getTokenAllowance(tokenAddress: Address, owner: Address, spender: Address): Promise<bigint> {
    return this.operations.getTokenAllowance(tokenAddress, owner, spender);
  }
  
  /**
   * Get token balance
   */
  async getTokenBalance(tokenAddress: Address, account: Address): Promise<bigint> {
    return this.operations.getTokenBalance(tokenAddress, account);
  }
  
  // ======== Channel creation methods ========
  
  /**
   * Create a channel context with a specific application logic
   * @param params Parameters for creating the channel context
   * @returns A new channel context
   */
  createChannel<T = unknown>(params: {
    participants: [Address, Address];
    challenge?: bigint;
    nonce?: bigint;
    appLogic: AppLogic<T>;
    initialAppState?: T;
  }): ChannelContext<T> {
    // Create the channel configuration
    const channel: Channel = {
      participants: params.participants,
      adjudicator: params.appLogic.getAdjudicatorAddress(),
      challenge: params.challenge || BigInt(86400), // Default: 1 day
      // Use a robust nonce generation strategy to prevent collisions
      nonce: params.nonce || generateChannelNonce(this.account?.address)
    };
    
    // Create and return the channel context
    return new ChannelContext<T>(
      this,
      channel,
      params.appLogic,
      params.initialAppState
    );
  }
  
  /**
   * Create a numeric value application
   */
  createNumericChannel(params: {
    participants: [Address, Address];
    adjudicatorAddress?: Address;
    adjudicatorAbi?: Abi;
    challenge?: bigint;
    nonce?: bigint;
    initialValue?: bigint;
    finalValue?: bigint;
  }): ChannelContext<AppDataTypes.NumericState> {
    return createNumericChannel(this, params);
  }
  
  /**
   * Create a sequential state application
   */
  createSequentialChannel(params: {
    participants: [Address, Address];
    adjudicatorAddress?: Address;
    adjudicatorAbi?: Abi;
    challenge?: bigint;
    nonce?: bigint;
    initialValue?: bigint;
  }): ChannelContext<AppDataTypes.SequentialState> {
    return createSequentialChannel(this, params);
  }
  
  /**
   * Create a custom application
   */
  createCustomChannel<T = unknown>(params: {
    participants: [Address, Address];
    challenge?: bigint;
    nonce?: bigint;
    encode: (data: T) => Hex;
    decode: (encoded: Hex) => T;
    validateTransition?: (prevState: T, nextState: T, signer: Address) => boolean;
    isFinal?: (state: T) => boolean;
    adjudicatorAddress?: Address;
    adjudicatorType?: string;
    adjudicatorAbi?: Abi;
    initialState?: T;
  }): ChannelContext<T> {
    return createCustomChannel<T>(this, params);
  }
}
