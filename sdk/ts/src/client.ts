/**
 * Main Nitrolite SDK Client
 * Provides a unified interface for interacting with Nitrolite payment channels.
 * Combines both high-level operations (Deposit, Withdraw, Transfer) and
 * low-level RPC access for advanced use cases.
 */

import { Address, Hex, createPublicClient, createWalletClient, http, custom } from 'viem';
import Decimal from 'decimal.js';
import * as core from './core/types';
import * as app from './app/types';
import * as API from './rpc/api';
import { StateV1, ChannelDefinitionV1 } from './rpc/types';
import { RPCClient } from './rpc/client';
import { WebsocketDialer } from './rpc/dialer';
import { ClientAssetStore } from './asset_store';
import { Config, DefaultConfig, Option } from './config';
import {
  generateNonce,
  transformNodeConfig,
  transformAssets,
  transformBalances,
  transformChannel,
  transformState,
  transformTransaction,
  transformPaginationMetadata,
  transformAppDefinitionToRPC,
  transformAppStateUpdateToRPC,
  transformSignedAppStateUpdateToRPC,
} from './utils';
import * as blockchain from './blockchain';
import { nextState, applyChannelCreation, applyHomeDepositTransition, applyHomeWithdrawalTransition, applyTransferSendTransition, applyFinalizeTransition, applyCommitTransition } from './core/state';
import { newVoidState } from './core/types';
import { packState } from './core/state_packer';
import { StateSigner, TransactionSigner } from './signers';

/**
 * Default challenge period for channels (1 day in seconds)
 */
export const DEFAULT_CHALLENGE_PERIOD = 86400;

// Re-export signer interfaces for convenience
export type { StateSigner, TransactionSigner };

/**
 * Client provides a unified interface for interacting with Nitrolite.
 * It combines both high-level operations (Deposit, Withdraw, Transfer) and
 * low-level RPC access for advanced use cases.
 *
 * @example
 * ```typescript
 * import { Client, withBlockchainRPC } from '@nitrolite/sdk';
 * import { privateKeyToAccount } from 'viem/accounts';
 *
 * // Create signers
 * const account = privateKeyToAccount('0x...');
 * const stateSigner = new WalletStateSigner(walletClient);
 *
 * // Create client
 * const client = await Client.create(
 *   'wss://node.nitrolite.com/ws',
 *   stateSigner,
 *   walletClient,
 *   withBlockchainRPC(80002n, 'https://polygon-amoy.alchemy.com/v2/KEY')
 * );
 *
 * // High-level operations
 * const txHash = await client.deposit(80002n, 'usdc', new Decimal(100));
 * const txId = await client.transfer('0xRecipient...', 'usdc', new Decimal(50));
 *
 * // Low-level operations
 * const config = await client.getConfig();
 * const balances = await client.getBalances('0x1234...');
 * ```
 */
export class Client {
  private rpcClient: RPCClient;
  private config: Config;
  private exitPromise: Promise<void>;
  private exitResolve?: () => void;
  private blockchainClients: Map<bigint, blockchain.evm.Client>;
  private homeBlockchains: Map<string, bigint>;
  private stateSigner: StateSigner;
  private txSigner: TransactionSigner;
  private assetStore: ClientAssetStore;

  private constructor(
    rpcClient: RPCClient,
    config: Config,
    stateSigner: StateSigner,
    txSigner: TransactionSigner,
    assetStore: ClientAssetStore
  ) {
    this.rpcClient = rpcClient;
    this.config = config;
    this.stateSigner = stateSigner;
    this.txSigner = txSigner;
    this.assetStore = assetStore;
    this.blockchainClients = new Map();
    this.homeBlockchains = new Map();

    // Create exit promise
    this.exitPromise = new Promise((resolve) => {
      this.exitResolve = resolve;
    });
  }

  /**
   * Create a new Nitrolite client with both high-level and low-level methods.
   * This is the recommended constructor for most use cases.
   *
   * @param wsURL - WebSocket URL of the Nitrolite server (e.g., "wss://node.nitrolite.com/ws")
   * @param stateSigner - Signer for signing channel states (EthereumMsgSigner)
   * @param txSigner - Signer for blockchain transactions (EthereumRawSigner)
   * @param opts - Optional configuration (withBlockchainRPC, withHandshakeTimeout, etc.)
   * @returns Configured Client ready for operations
   *
   * @example
   * ```typescript
   * import { createSigners } from '@nitrolite/sdk';
   *
   * const { stateSigner, txSigner } = createSigners('0x...');
   * const client = await Client.create(
   *   'wss://node.nitrolite.com/ws',
   *   stateSigner,
   *   txSigner,
   *   withBlockchainRPC(80002n, 'https://polygon-amoy.alchemy.com/v2/KEY')
   * );
   * ```
   */
  static async create(
    wsURL: string,
    stateSigner: StateSigner,
    txSigner: TransactionSigner,
    ...opts: Option[]
  ): Promise<Client> {
    // Build config starting with defaults
    const config: Config = {
      url: wsURL,
      handshakeTimeout: DefaultConfig.handshakeTimeout,
      pingInterval: DefaultConfig.pingInterval,
      errorHandler: DefaultConfig.errorHandler,
      blockchainRPCs: DefaultConfig.blockchainRPCs || new Map(),
    };

    // Apply user options
    for (const opt of opts) {
      opt(config);
    }

    // Create WebSocket dialer
    const dialer = new WebsocketDialer();
    const rpcClient = new RPCClient(dialer);

    // Declare client variable for use in asset store
    let client: Client;

    // Create asset store (will be initialized with client methods)
    const assetStore = new ClientAssetStore(async () => {
      return await client.getAssets();
    });

    // Create client instance
    client = new Client(rpcClient, config, stateSigner, txSigner, assetStore);

    // Error handler wrapper
    const handleError = (err?: Error) => {
      if (err && config.errorHandler) {
        config.errorHandler(err);
      }
      client.exitResolve?.();
    };

    // Establish connection
    await rpcClient.start(wsURL, handleError);

    return client;
  }

  // ============================================================================
  // Home Blockchain Management
  // ============================================================================

  /**
   * SetHomeBlockchain configures the primary blockchain network for a specific asset.
   * This is required for operations like Transfer which may trigger channel creation
   * but do not accept a blockchain ID as a parameter.
   *
   * @param asset - The asset symbol (e.g., "usdc")
   * @param blockchainId - The chain ID to associate with the asset (e.g., 80002n)
   *
   * @example
   * ```typescript
   * // Set USDC to settle on Polygon Amoy
   * await client.setHomeBlockchain('usdc', 80002n);
   * ```
   */
  async setHomeBlockchain(asset: string, blockchainId: bigint): Promise<void> {
    const existingBlockchainId = this.homeBlockchains.get(asset);
    if (existingBlockchainId !== undefined) {
      throw new Error(
        `home blockchain is already set for asset ${asset} to ${existingBlockchainId}, please use Migrate() if you want to change home blockchain`
      );
    }

    const exists = await this.assetStore.assetExistsOnBlockchain(blockchainId, asset);
    if (!exists) {
      throw new Error(`asset ${asset} not supported on blockchain ${blockchainId}`);
    }

    this.homeBlockchains.set(asset, blockchainId);
  }

  // ============================================================================
  // Connection & Lifecycle Methods
  // ============================================================================

  /**
   * Close cleanly shuts down the client connection.
   * It's recommended to call this when done using the client.
   *
   * @example
   * ```typescript
   * await client.close();
   * ```
   */
  async close(): Promise<void> {
    this.exitResolve?.();
  }

  /**
   * WaitForClose returns a promise that resolves when the connection is lost or closed.
   * This is useful for monitoring connection health in long-running applications.
   *
   * @example
   * ```typescript
   * client.waitForClose().then(() => {
   *   console.log('Connection closed');
   * });
   * ```
   */
  waitForClose(): Promise<void> {
    return this.exitPromise;
  }

  // ============================================================================
  // Shared Helper Methods
  // ============================================================================

  /**
   * SignState signs a channel state by packing it, hashing it, and signing the hash.
   * Returns the signature as a hex-encoded string (with 0x prefix).
   *
   * This is a low-level method exposed for advanced users who want to manually
   * construct and sign states. Most users should use the high-level methods like
   * transfer, deposit, and withdraw instead.
   */
  async signState(state: core.State): Promise<Hex> {
    // Pack the state into ABI-encoded bytes
    const packed = await packState(state, this.assetStore);

    // Sign the packed state using the state signer (adds Ethereum message prefix and hashes internally)
    const signature = await this.stateSigner.signMessage(packed);

    return signature;
  }

  /**
   * GetUserAddress returns the Ethereum address associated with the signer.
   * This is useful for identifying the current user's wallet address.
   */
  getUserAddress(): Address {
    return this.stateSigner.getAddress();
  }

  /**
   * SignAndSubmitState is a helper that signs a state and submits it to the node.
   * Returns the node's signature.
   */
  private async signAndSubmitState(state: core.State): Promise<Hex> {
    // Sign state
    const sig = await this.signState(state);
    state.userSig = sig;

    // Submit to node
    const nodeSig = await this.submitState(state);

    // Update state with node signature
    state.nodeSig = nodeSig as Hex;

    return nodeSig as Hex;
  }

  // ============================================================================
  // High-Level Operations
  // ============================================================================

  /**
   * Deposit adds funds to the user's channel by depositing from the blockchain.
   * This method handles two scenarios automatically:
   * 1. If no channel exists: Creates a new channel with the initial deposit
   * 2. If channel exists: Checkpoints the deposit to the existing channel
   *
   * @param blockchainId - The blockchain network ID (e.g., 80002n for Polygon Amoy)
   * @param asset - The asset symbol to deposit (e.g., "usdc")
   * @param amount - The amount to deposit
   * @returns Transaction hash of the blockchain transaction
   *
   * @example
   * ```typescript
   * const txHash = await client.deposit(80002n, 'usdc', new Decimal(100));
   * console.log('Deposit transaction:', txHash);
   * ```
   */
  async deposit(blockchainId: bigint, asset: string, amount: Decimal): Promise<string> {
    const userWallet = this.getUserAddress();

    // Initialize blockchain client if needed
    await this.initializeBlockchainClient(blockchainId);
    const blockchainClient = this.blockchainClients.get(blockchainId)!;

    // Get node address
    const nodeAddress = await this.getNodeAddress();
    if (!nodeAddress) {
      throw new Error('node address is undefined - ensure node config is properly loaded');
    }

    // Get token address for this asset on this blockchain
    const tokenAddress = await this.assetStore.getTokenAddress(asset, blockchainId);
    if (!tokenAddress) {
      throw new Error(`token address not found for asset ${asset} on blockchain ${blockchainId}`);
    }

    // Try to get latest state to determine if channel exists
    let state: core.State | null = null;
    try {
      state = await this.getLatestState(userWallet, asset, false);
    } catch (err) {
      // Channel doesn't exist, will create it
    }

    // Scenario A: Channel doesn't exist - create it
    if (!state || !state.homeChannelId) {
      // Create channel definition
      const channelDef: core.ChannelDefinition = {
        nonce: generateNonce(),
        challenge: DEFAULT_CHALLENGE_PERIOD,
      };

      if (!state) {
        state = newVoidState(asset, userWallet);
      }
      const newState = nextState(state!);

      applyChannelCreation(newState, channelDef, blockchainId, tokenAddress as Address, nodeAddress);
      applyHomeDepositTransition(newState, amount);

      // Sign state
      const sig = await this.signState(newState);
      newState.userSig = sig;

      // Request channel creation from node
      const nodeSig = await this.requestChannelCreation(newState, channelDef);
      newState.nodeSig = nodeSig as Hex;

      // Create channel on blockchain
      const txHash = await blockchainClient.create(channelDef, newState);

      return txHash;
    }

    // Scenario B: Channel exists - checkpoint deposit
    const newState = nextState(state);
    applyHomeDepositTransition(newState, amount);

    // Sign and submit state to node
    await this.signAndSubmitState(newState);

    // Checkpoint on blockchain
    const txHash = await blockchainClient.checkpoint(newState);

    return txHash;
  }

  /**
   * Withdraw removes funds from the user's channel and returns them to the blockchain wallet.
   * This operation handles two scenarios automatically:
   * 1. If no channel exists: Creates a new channel and executes the withdrawal in one transaction
   * 2. If channel exists: Checkpoints the withdrawal to the existing channel
   *
   * @param blockchainId - The blockchain network ID (e.g., 80002n for Polygon Amoy)
   * @param asset - The asset symbol to withdraw (e.g., "usdc")
   * @param amount - The amount to withdraw
   * @returns Transaction hash of the blockchain transaction
   *
   * @example
   * ```typescript
   * const txHash = await client.withdraw(80002n, 'usdc', new Decimal(25));
   * console.log('Withdrawal transaction:', txHash);
   * ```
   */
  async withdraw(blockchainId: bigint, asset: string, amount: Decimal): Promise<string> {
    const userWallet = this.getUserAddress();

    // Initialize blockchain client if needed
    await this.initializeBlockchainClient(blockchainId);
    const blockchainClient = this.blockchainClients.get(blockchainId)!;

    // Get node address
    const nodeAddress = await this.getNodeAddress();
    if (!nodeAddress) {
      throw new Error('node address is undefined - ensure node config is properly loaded');
    }

    // Get token address for this asset on this blockchain
    const tokenAddress = await this.assetStore.getTokenAddress(asset, blockchainId);
    if (!tokenAddress) {
      throw new Error(`token address not found for asset ${asset} on blockchain ${blockchainId}`);
    }

    // Try to get latest state to determine if channel exists
    let state: core.State | null = null;
    try {
      state = await this.getLatestState(userWallet, asset, false);
    } catch (err) {
      // Channel doesn't exist, will create it
    }

    // Channel doesn't exist - create it and withdraw
    if (!state || !state.homeChannelId) {
      // Create channel definition
      const channelDef: core.ChannelDefinition = {
        nonce: generateNonce(),
        challenge: DEFAULT_CHALLENGE_PERIOD,
      };

      if (!state) {
        state = newVoidState(asset, userWallet);
      }
      const newState = nextState(state!);

      applyChannelCreation(newState, channelDef, blockchainId, tokenAddress as Address, nodeAddress);
      applyHomeWithdrawalTransition(newState, amount);

      // Sign state
      const sig = await this.signState(newState);
      newState.userSig = sig;

      // Request channel creation from node
      const nodeSig = await this.requestChannelCreation(newState, channelDef);
      newState.nodeSig = nodeSig as Hex;

      // Create channel on blockchain (Smart Contract handles Creation + Withdrawal)
      const txHash = await blockchainClient.create(channelDef, newState);

      return txHash;
    }

    // Create next state
    const newState = nextState(state);
    applyHomeWithdrawalTransition(newState, amount);

    // Sign and submit state to node
    await this.signAndSubmitState(newState);

    // Checkpoint on blockchain
    const txHash = await blockchainClient.checkpoint(newState);

    return txHash;
  }

  /**
   * Transfer sends funds from the user to another wallet address.
   * This is an off-chain operation that doesn't require blockchain interaction.
   *
   * @param recipientWallet - The recipient's wallet address (e.g., "0x1234...")
   * @param asset - The asset symbol to transfer (e.g., "usdc")
   * @param amount - The amount to transfer
   * @returns Transaction ID for tracking
   *
   * @example
   * ```typescript
   * const txId = await client.transfer('0xRecipient...', 'usdc', new Decimal(50));
   * console.log('Transfer successful:', txId);
   * ```
   */
  async transfer(recipientWallet: string, asset: string, amount: Decimal): Promise<string> {
    const senderWallet = this.getUserAddress();

    // Get sender's latest state
    let state: core.State | null = null;
    try {
      state = await this.getLatestState(senderWallet, asset, false);
    } catch (err) {
      // Channel doesn't exist
    }

    if (!state || !state.homeChannelId) {
      // Create channel definition
      const channelDef: core.ChannelDefinition = {
        nonce: generateNonce(),
        challenge: DEFAULT_CHALLENGE_PERIOD,
      };

      if (!state) {
        state = newVoidState(asset, senderWallet);
      }
      const newState = nextState(state!);

      const blockchainId = this.homeBlockchains.get(asset);
      if (!blockchainId) {
        throw new Error(`home blockchain not set for asset ${asset}`);
      }

      // Get node address
      const nodeAddress = await this.getNodeAddress();
      if (!nodeAddress) {
        throw new Error('node address is undefined - ensure node config is properly loaded');
      }

      // Get token address for this asset on this blockchain
      const tokenAddress = await this.assetStore.getTokenAddress(asset, blockchainId);
      if (!tokenAddress) {
        throw new Error(`token address not found for asset ${asset} on blockchain ${blockchainId}`);
      }

      // Initialize blockchain client if needed
      await this.initializeBlockchainClient(blockchainId);
      const blockchainClient = this.blockchainClients.get(blockchainId)!;

      applyChannelCreation(newState, channelDef, blockchainId, tokenAddress as Address, nodeAddress);
      const transition = applyTransferSendTransition(newState, recipientWallet, amount);

      const sig = await this.signState(newState);
      newState.userSig = sig;

      // Request channel creation from node
      const nodeSig = await this.requestChannelCreation(newState, channelDef);
      newState.nodeSig = nodeSig as Hex;

      // Create channel on blockchain
      const txHash = await blockchainClient.create(channelDef, newState);

      return txHash;
    }

    // Create next state
    const newState = nextState(state);
    const transition = applyTransferSendTransition(newState, recipientWallet, amount);

    // Sign and submit state
    await this.signAndSubmitState(newState);

    // Return transaction ID from the transition
    return transition.txId;
  }

  /**
   * CloseHomeChannel finalizes and closes the user's channel for a specific asset.
   *
   * @param asset - The asset symbol (e.g., "usdc")
   * @returns Transaction hash of the blockchain transaction
   *
   * @example
   * ```typescript
   * const txHash = await client.closeHomeChannel('usdc');
   * console.log('Channel closed:', txHash);
   * ```
   */
  async closeHomeChannel(asset: string): Promise<string> {
    const senderWallet = this.getUserAddress();

    const state = await this.getLatestState(senderWallet, asset, false);

    if (!state.homeChannelId) {
      throw new Error(`no channel exists for asset ${asset}`);
    }

    const blockchainId = state.homeLedger.blockchainId;

    // Initialize blockchain client if needed
    await this.initializeBlockchainClient(blockchainId);
    const blockchainClient = this.blockchainClients.get(blockchainId)!;

    // Create next state
    const newState = nextState(state);
    applyFinalizeTransition(newState);

    // Sign and submit state
    await this.signAndSubmitState(newState);

    // Close on blockchain
    const txHash = await blockchainClient.close(newState);

    return txHash;
  }

  /**
   * Approve token spending for a specific chain and token
   * @param chainId - The blockchain ID
   * @param tokenAddress - The ERC20 token contract address
   * @param amount - Amount to approve (in smallest unit, e.g., wei)
   * @returns Transaction hash
   */
  async approveToken(chainId: bigint, tokenAddress: string, amount: bigint): Promise<string> {
    await this.initializeBlockchainClient(chainId);
    const blockchainClient = this.blockchainClients.get(chainId)!;

    return await blockchainClient.approveTokenByAddress(
      tokenAddress as `0x${string}`,
      amount
    );
  }

  /**
   * Check token allowance for a specific chain and token
   * @param chainId - The blockchain ID
   * @param tokenAddress - The ERC20 token contract address
   * @param owner - The owner address
   * @returns Current allowance amount (in smallest unit)
   */
  async checkTokenAllowance(
    chainId: bigint,
    tokenAddress: string,
    owner: string
  ): Promise<bigint> {
    await this.initializeBlockchainClient(chainId);
    const blockchainClient = this.blockchainClients.get(chainId)!;

    return await blockchainClient.checkAllowanceByAddress(
      tokenAddress as `0x${string}`,
      owner as `0x${string}`
    );
  }

  // ============================================================================
  // Node Information Methods
  // ============================================================================

  /**
   * Ping performs a health check on the node connection.
   *
   * @example
   * ```typescript
   * await client.ping();
   * console.log('Node is healthy');
   * ```
   */
  async ping(): Promise<void> {
    await this.rpcClient.nodeV1Ping();
  }

  /**
   * GetConfig retrieves the node configuration including supported blockchains.
   *
   * @returns Node configuration with blockchain list
   *
   * @example
   * ```typescript
   * const config = await client.getConfig();
   * console.log('Node version:', config.nodeVersion);
   * console.log('Supported blockchains:', config.blockchains);
   * ```
   */
  async getConfig(): Promise<core.NodeConfig> {
    const resp = await this.rpcClient.nodeV1GetConfig();
    return transformNodeConfig(resp);
  }

  /**
   * GetBlockchains retrieves the list of supported blockchains.
   *
   * @returns Array of supported blockchains
   *
   * @example
   * ```typescript
   * const blockchains = await client.getBlockchains();
   * for (const chain of blockchains) {
   *   console.log(`${chain.name} (${chain.blockchainId})`);
   * }
   * ```
   */
  async getBlockchains(): Promise<core.Blockchain[]> {
    const config = await this.getConfig();
    return config.blockchains;
  }

  /**
   * GetAssets retrieves the list of supported assets, optionally filtered by blockchain.
   *
   * @param blockchainId - Optional blockchain ID to filter assets
   * @returns Array of supported assets
   *
   * @example
   * ```typescript
   * // Get all assets
   * const allAssets = await client.getAssets();
   *
   * // Get assets for specific blockchain
   * const polygonAssets = await client.getAssets(80002n);
   * ```
   */
  async getAssets(blockchainId?: bigint): Promise<core.Asset[]> {
    const req: API.NodeV1GetAssetsRequest = {};
    if (blockchainId !== undefined) {
      req.blockchain_id = blockchainId;
    }
    const resp = await this.rpcClient.nodeV1GetAssets(req);
    return transformAssets(resp.assets);
  }

  // ============================================================================
  // User Query Methods
  // ============================================================================

  /**
   * GetBalances retrieves the balance information for a user's wallet.
   *
   * @param wallet - The user's wallet address
   * @returns Array of balance entries for each asset
   *
   * @example
   * ```typescript
   * const balances = await client.getBalances('0x1234...');
   * for (const entry of balances) {
   *   console.log(`${entry.asset}: ${entry.balance}`);
   * }
   * ```
   */
  async getBalances(wallet: Address): Promise<core.BalanceEntry[]> {
    const req: API.UserV1GetBalancesRequest = {
      wallet,
    };
    const resp = await this.rpcClient.userV1GetBalances(req);
    return transformBalances(resp.balances);
  }

  /**
   * GetTransactions retrieves the transaction history for a user's wallet.
   *
   * @param wallet - The user's wallet address
   * @param options - Optional filters (asset, pagination)
   * @returns Array of transactions and pagination metadata
   *
   * @example
   * ```typescript
   * const { transactions, metadata } = await client.getTransactions('0x1234...', {
   *   asset: 'usdc',
   *   page: 1,
   *   pageSize: 10,
   * });
   * ```
   */
  async getTransactions(
    wallet: Address,
    options?: {
      asset?: string;
      page?: number;
      pageSize?: number;
    }
  ): Promise<{ transactions: core.Transaction[]; metadata: core.PaginationMetadata }> {
    const req: API.UserV1GetTransactionsRequest = {
      wallet,
      asset: options?.asset,
      pagination: options?.page && options?.pageSize ? {
        offset: (options.page - 1) * options.pageSize,
        limit: options.pageSize,
      } : undefined,
    };
    const resp = await this.rpcClient.userV1GetTransactions(req);
    return {
      transactions: resp.transactions.map(transformTransaction),
      metadata: transformPaginationMetadata(resp.metadata),
    };
  }

  // ============================================================================
  // Channel Query Methods
  // ============================================================================

  /**
   * GetHomeChannel retrieves home channel information for a user's asset.
   *
   * @param wallet - The user's wallet address
   * @param asset - The asset symbol
   * @returns Channel information for the home channel
   *
   * @example
   * ```typescript
   * const channel = await client.getHomeChannel('0x1234...', 'usdc');
   * console.log(`Channel: ${channel.channelId} (Version: ${channel.stateVersion})`);
   * ```
   */
  async getHomeChannel(wallet: Address, asset: string): Promise<core.Channel> {
    const req: API.ChannelsV1GetHomeChannelRequest = {
      wallet,
      asset,
    };
    const resp = await this.rpcClient.channelsV1GetHomeChannel(req);
    return transformChannel(resp.channel);
  }

  /**
   * GetEscrowChannel retrieves escrow channel information for a specific channel ID.
   *
   * @param escrowChannelId - The escrow channel ID to query
   * @returns Channel information for the escrow channel
   *
   * @example
   * ```typescript
   * const channel = await client.getEscrowChannel('0x1234...');
   * console.log(`Channel: ${channel.channelId} (Version: ${channel.stateVersion})`);
   * ```
   */
  async getEscrowChannel(escrowChannelId: string): Promise<core.Channel> {
    const req: API.ChannelsV1GetEscrowChannelRequest = {
      escrow_channel_id: escrowChannelId,
    };
    const resp = await this.rpcClient.channelsV1GetEscrowChannel(req);
    return transformChannel(resp.channel);
  }

  /**
   * GetLatestState retrieves the latest state for a user's asset.
   *
   * @param wallet - The user's wallet address
   * @param asset - The asset symbol (e.g., "usdc")
   * @param onlySigned - If true, returns only the latest signed state
   * @returns State containing all state information
   *
   * @example
   * ```typescript
   * const state = await client.getLatestState('0x1234...', 'usdc', false);
   * console.log(`Version: ${state.version}, Balance: ${state.homeLedger.userBalance}`);
   * ```
   */
  async getLatestState(wallet: Address, asset: string, onlySigned: boolean): Promise<core.State> {
    const req: API.ChannelsV1GetLatestStateRequest = {
      wallet,
      asset,
      only_signed: onlySigned,
    };
    const resp = await this.rpcClient.channelsV1GetLatestState(req);
    return transformState(resp.state);
  }

  // ============================================================================
  // App Session Methods
  // ============================================================================

  /**
   * GetAppSessions retrieves application sessions for the user.
   *
   * @param options - Optional filters (appSessionId, wallet, status, pagination)
   * @returns Array of app session info and pagination metadata
   *
   * @example
   * ```typescript
   * const { sessions, metadata } = await client.getAppSessions({
   *   wallet: '0x1234...',
   *   status: 'open',
   *   page: 1,
   *   pageSize: 10,
   * });
   * ```
   */
  async getAppSessions(options?: {
    appSessionId?: string;
    wallet?: Address;
    status?: string;
    page?: number;
    pageSize?: number;
  }): Promise<{ sessions: app.AppSessionInfoV1[]; metadata: core.PaginationMetadata }> {
    const req: API.AppSessionsV1GetAppSessionsRequest = {
      app_session_id: options?.appSessionId,
      participant: options?.wallet,
      status: options?.status,
      pagination: options?.page && options?.pageSize ? {
        offset: (options.page - 1) * options.pageSize,
        limit: options.pageSize,
      } : undefined,
    };
    const resp = await this.rpcClient.appSessionsV1GetAppSessions(req);
    return {
      sessions: resp.app_sessions,
      metadata: transformPaginationMetadata(resp.metadata),
    };
  }

  /**
   * GetAppDefinition retrieves the definition for a specific app session.
   *
   * @param appSessionId - The app session ID
   * @returns App session definition
   *
   * @example
   * ```typescript
   * const definition = await client.getAppDefinition('0x1234...');
   * console.log('Participants:', definition.participants);
   * ```
   */
  async getAppDefinition(appSessionId: string): Promise<app.AppDefinitionV1> {
    const req: API.AppSessionsV1GetAppDefinitionRequest = {
      app_session_id: appSessionId,
    };
    const resp = await this.rpcClient.appSessionsV1GetAppDefinition(req);
    return resp.definition; // Already in correct format
  }

  /**
   * CreateAppSession creates a new application session between participants.
   *
   * @param definition - The app definition with participants, quorum, application ID
   * @param sessionData - Optional JSON stringified session data
   * @param quorumSigs - Participant signatures for the app session creation
   * @returns Object with appSessionId, version, and status
   *
   * @example
   * ```typescript
   * const definition: app.AppDefinitionV1 = {
   *   application: 'chess-v1',
   *   participants: [
   *     { walletAddress: '0x1234...', signatureWeight: 1 },
   *     { walletAddress: '0x5678...', signatureWeight: 1 },
   *   ],
   *   quorum: 2,
   *   nonce: 1n,
   * };
   * const { appSessionId, version, status } = await client.createAppSession(
   *   definition,
   *   '{}',
   *   ['sig1', 'sig2']
   * );
   * console.log('Created session:', appSessionId);
   * ```
   */
  async createAppSession(
    definition: app.AppDefinitionV1,
    sessionData: string,
    quorumSigs: string[]
  ): Promise<{ appSessionId: string; version: string; status: string }> {
    const req: API.AppSessionsV1CreateAppSessionRequest = {
      definition: transformAppDefinitionToRPC(definition) as any, // RPC type
      session_data: sessionData,
      quorum_sigs: quorumSigs,
    };
    const resp = await this.rpcClient.appSessionsV1CreateAppSession(req);
    return {
      appSessionId: resp.app_session_id,
      version: resp.version,
      status: resp.status,
    };
  }

  /**
   * SubmitAppSessionDeposit submits a deposit to an app session.
   * This updates both the app session state and the user's channel state.
   *
   * @param appStateUpdate - The app state update with deposit intent
   * @param quorumSigs - Participant signatures for the app state update
   * @param asset - The asset to deposit
   * @param depositAmount - Amount to deposit
   * @returns Node's signature for the state
   *
   * @example
   * ```typescript
   * const appUpdate: app.AppStateUpdateV1 = {
   *   appSessionId: 'session123',
   *   intent: app.AppStateUpdateIntent.Deposit,
   *   version: 2n,
   *   allocations: [
   *     { participant: '0x1234...', asset: 'usdc', amount: new Decimal(100) },
   *   ],
   *   sessionData: '{}',
   * };
   * const nodeSig = await client.submitAppSessionDeposit(
   *   appUpdate,
   *   ['sig1'],
   *   'usdc',
   *   new Decimal(100)
   * );
   * ```
   */
  async submitAppSessionDeposit(
    appStateUpdate: app.AppStateUpdateV1,
    quorumSigs: string[],
    asset: string,
    depositAmount: Decimal
  ): Promise<string> {
    // Get current state
    const currentState = await this.getLatestState(this.getUserAddress(), asset, false);

    // Create next state with commit transition (use app session ID as account ID)
    const newState = nextState(currentState);
    applyCommitTransition(newState, appStateUpdate.appSessionId, depositAmount);

    // Transform to RPC format after applying the commit transition
    const appUpdate = transformAppStateUpdateToRPC(appStateUpdate);

    // Sign the state
    const stateSig = await this.signState(newState);
    newState.userSig = stateSig;

    // Submit deposit
    const req: API.AppSessionsV1SubmitDepositStateRequest = {
      app_state_update: appUpdate as any, // RPC type
      quorum_sigs: quorumSigs,
      user_state: this.transformStateToRPC(newState),
    };

    const resp = await this.rpcClient.appSessionsV1SubmitDepositState(req);
    return resp.state_node_sig;
  }

  /**
   * SubmitAppState submits an app session state update.
   * This method handles operate, withdraw, and close intents.
   * For deposits, use submitAppSessionDeposit instead.
   *
   * @param appStateUpdate - The app state update (intent: operate, withdraw, or close)
   * @param quorumSigs - Participant signatures for the app state update
   *
   * @example
   * ```typescript
   * const appUpdate: app.AppStateUpdateV1 = {
   *   appSessionId: 'session123',
   *   intent: app.AppStateUpdateIntent.Operate,
   *   version: 3n,
   *   allocations: [
   *     { participant: '0x1234...', asset: 'usdc', amount: new Decimal(50) },
   *     { participant: '0x5678...', asset: 'usdc', amount: new Decimal(50) },
   *   ],
   *   sessionData: '{"move": "e4"}',
   * };
   * await client.submitAppState(appUpdate, ['sig1', 'sig2']);
   * ```
   */
  async submitAppState(
    appStateUpdate: app.AppStateUpdateV1,
    quorumSigs: string[]
  ): Promise<void> {
    const appUpdate = transformAppStateUpdateToRPC(appStateUpdate);

    const req: API.AppSessionsV1SubmitAppStateRequest = {
      app_state_update: appUpdate as any, // RPC type
      quorum_sigs: quorumSigs,
    };

    await this.rpcClient.appSessionsV1SubmitAppState(req);
  }

  /**
   * RebalanceAppSessions rebalances multiple application sessions atomically.
   *
   * This method performs atomic rebalancing across multiple app sessions, ensuring
   * that funds are redistributed consistently without the risk of partial updates.
   *
   * @param signedUpdates - Array of signed app state updates to apply atomically
   * @returns BatchID for tracking the rebalancing operation
   *
   * @example
   * ```typescript
   * const updates: app.SignedAppStateUpdateV1[] = [
   *   {
   *     appStateUpdate: { appSessionId: 'session1', intent: app.AppStateUpdateIntent.Rebalance, ... },
   *     quorumSigs: ['sig1', 'sig2'],
   *   },
   *   {
   *     appStateUpdate: { appSessionId: 'session2', intent: app.AppStateUpdateIntent.Rebalance, ... },
   *     quorumSigs: ['sig3', 'sig4'],
   *   },
   * ];
   * const batchId = await client.rebalanceAppSessions(updates);
   * console.log('Rebalance batch ID:', batchId);
   * ```
   */
  async rebalanceAppSessions(
    signedUpdates: app.SignedAppStateUpdateV1[]
  ): Promise<string> {
    // Transform SDK types to RPC types
    const rpcUpdates = signedUpdates.map(transformSignedAppStateUpdateToRPC);

    const req: API.AppSessionsV1RebalanceAppSessionsRequest = {
      signed_updates: rpcUpdates as any, // RPC type
    };

    const resp = await this.rpcClient.appSessionsV1RebalanceAppSessions(req);
    return resp.batch_id;
  }

  // ============================================================================
  // Private Helper Methods
  // ============================================================================

  /**
   * Initialize a blockchain client for a specific chain.
   */
  private async initializeBlockchainClient(chainId: bigint): Promise<void> {
    // Check if already initialized
    if (this.blockchainClients.has(chainId)) {
      return;
    }

    // Get RPC URL from config
    const rpcUrl = this.config.blockchainRPCs?.get(chainId);
    if (!rpcUrl) {
      throw new Error(
        `blockchain RPC not configured for chain ${chainId} (use withBlockchainRPC)`
      );
    }

    // Get node config to find contract address
    const config = await this.getConfig();
    const blockchainInfo = config.blockchains.find((b) => b.id === chainId);
    if (!blockchainInfo) {
      throw new Error(`blockchain ${chainId} not supported by node`);
    }

    const contractAddress = blockchainInfo.contractAddress;
    const nodeAddress = config.nodeAddress;

    // Create viem clients
    const publicClient = createPublicClient({
      transport: http(rpcUrl),
    }) as blockchain.evm.EVMClient;

    // Create a minimal chain object for the wallet client
    // This is required for viem to know which chain to submit transactions to
    const chain = {
      id: Number(chainId),
      name: `Chain ${chainId}`,
      nativeCurrency: { name: 'ETH', symbol: 'ETH', decimals: 18 },
      rpcUrls: {
        default: { http: [rpcUrl] },
        public: { http: [rpcUrl] },
      },
    };

    console.log(`🔗 Creating blockchain client for chain ${chainId}:`, {
      chainId: chainId.toString(),
      rpcUrl,
      contractAddress,
      chainName: chain.name,
    });

    // Detect if we're in a browser with MetaMask/wallet provider
    // In browser: use wallet provider for transactions (supports signing)
    // In Node.js: use HTTP (requires private key, won't work for transactions)
    const isBrowser = typeof window !== 'undefined' && typeof (window as any).ethereum !== 'undefined';

    let walletClient: blockchain.evm.WalletSigner;

    if (isBrowser) {
      console.log('🦊 Browser detected - using MetaMask/wallet provider for transactions');
      // Use MetaMask/wallet provider which supports transaction signing
      walletClient = createWalletClient({
        chain,
        transport: custom((window as any).ethereum),
        account: this.txSigner.getAddress(),
      }) as blockchain.evm.WalletSigner;
    } else {
      console.warn('⚠️  Node.js environment - HTTP transport will NOT support transaction signing');
      console.warn('    Transactions will fail unless using a service that manages private keys');
      // Fallback to HTTP (will fail for transactions in most cases)
      walletClient = createWalletClient({
        chain,
        transport: http(rpcUrl),
        account: this.txSigner.getAddress(),
      }) as blockchain.evm.WalletSigner;
    }

    const blockchainClient = new blockchain.evm.Client(
      contractAddress,
      publicClient,
      walletClient,
      chainId,
      nodeAddress,
      this.assetStore
    );

    this.blockchainClients.set(chainId, blockchainClient);
  }

  /**
   * Get the node address from the config.
   */
  private async getNodeAddress(): Promise<Address> {
    const config = await this.getConfig();
    return config.nodeAddress;
  }

  /**
   * Submit a signed state update to the node.
   */
  private async submitState(state: core.State): Promise<string> {
    const req: API.ChannelsV1SubmitStateRequest = {
      state: this.transformStateToRPC(state),
    };
    const resp = await this.rpcClient.channelsV1SubmitState(req);
    return resp.signature;
  }

  /**
   * Request the node to sign a channel creation.
   */
  private async requestChannelCreation(
    state: core.State,
    channelDef: core.ChannelDefinition
  ): Promise<string> {
    const req: API.ChannelsV1RequestCreationRequest = {
      state: this.transformStateToRPC(state),
      channel_definition: this.transformChannelDefinitionToRPC(channelDef),
    };
    const resp = await this.rpcClient.channelsV1RequestCreation(req);
    return resp.signature;
  }

  /**
   * Transform core State to RPC StateV1
   */
  private transformStateToRPC(state: core.State): StateV1 {
    // This is a simplified version - you'll need to implement the full transformation
    return {
      id: state.id,
      transitions: state.transitions.map((t) => ({
        type: t.type, // Keep as TransitionType enum
        tx_id: t.txId,
        account_id: t.accountId,
        amount: t.amount.toString(),
      })),
      asset: state.asset,
      user_wallet: state.userWallet,
      epoch: state.epoch.toString(), // Convert bigint to string
      version: state.version.toString(), // Convert bigint to string
      home_channel_id: state.homeChannelId,
      escrow_channel_id: state.escrowChannelId,
      home_ledger: {
        token_address: state.homeLedger.tokenAddress,
        blockchain_id: state.homeLedger.blockchainId.toString(),
        user_balance: state.homeLedger.userBalance.toString(),
        user_net_flow: state.homeLedger.userNetFlow.toString(),
        node_balance: state.homeLedger.nodeBalance.toString(),
        node_net_flow: state.homeLedger.nodeNetFlow.toString(),
      },
      escrow_ledger: state.escrowLedger
        ? {
            token_address: state.escrowLedger.tokenAddress,
            blockchain_id: state.escrowLedger.blockchainId.toString(),
            user_balance: state.escrowLedger.userBalance.toString(),
            user_net_flow: state.escrowLedger.userNetFlow.toString(),
            node_balance: state.escrowLedger.nodeBalance.toString(),
            node_net_flow: state.escrowLedger.nodeNetFlow.toString(),
          }
        : undefined,
      user_sig: state.userSig,
      node_sig: state.nodeSig,
    };
  }

  /**
   * Transform core ChannelDefinition to RPC ChannelDefinitionV1
   */
  private transformChannelDefinitionToRPC(def: core.ChannelDefinition): ChannelDefinitionV1 {
    return {
      nonce: def.nonce.toString(),
      challenge: def.challenge,
    };
  }

  /**
   * Convert transition type enum to string
   */
  private transitionTypeToString(type: core.TransitionType): string {
    const typeMap: Record<core.TransitionType, string> = {
      [core.TransitionType.HomeDeposit]: 'home_deposit',
      [core.TransitionType.HomeWithdrawal]: 'home_withdrawal',
      [core.TransitionType.EscrowDeposit]: 'escrow_deposit',
      [core.TransitionType.EscrowWithdraw]: 'escrow_withdraw',
      [core.TransitionType.TransferSend]: 'transfer_send',
      [core.TransitionType.TransferReceive]: 'transfer_receive',
      [core.TransitionType.Commit]: 'commit',
      [core.TransitionType.Release]: 'release',
      [core.TransitionType.Migrate]: 'migrate',
      [core.TransitionType.EscrowLock]: 'escrow_lock',
      [core.TransitionType.MutualLock]: 'mutual_lock',
      [core.TransitionType.Finalize]: 'finalize',
    };
    return typeMap[type] || 'home_deposit';
  }
}
