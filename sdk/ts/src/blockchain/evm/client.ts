/**
 * EVM Blockchain Client
 * Main client for interacting with ChannelHub contract
 */

import { Address, Hex, hexToBytes } from 'viem';
import Decimal from 'decimal.js';
import * as core from '../../core/types';
import { decimalToBigInt } from '../../core/utils';
import { AssetStore, EVMClient, WalletSigner } from './interface';
import { ChannelHubAbi } from './channel_hub_abi';
import {
  coreDefToContractDef,
  coreStateToContractState,
  contractStateToCoreState,
} from './utils';
import { newERC20 } from './erc20';

/**
 * ClientOptions for configuring the blockchain client
 */
export interface ClientOptions {
  requireCheckAllowance?: boolean;
  requireCheckBalance?: boolean;
}

/**
 * Client provides methods to interact with the ChannelHub contract
 */
export class Client {
  private contractAddress: Address;
  private evmClient: EVMClient;
  private walletSigner: WalletSigner;
  private blockchainId: bigint;
  private nodeAddress: Address;
  private assetStore: AssetStore;

  private requireCheckAllowance: boolean;
  private requireCheckBalance: boolean;

  constructor(
    contractAddress: Address,
    evmClient: EVMClient,
    walletSigner: WalletSigner,
    blockchainId: bigint,
    nodeAddress: Address,
    assetStore: AssetStore,
    options?: ClientOptions
  ) {
    this.contractAddress = contractAddress;
    this.evmClient = evmClient;
    this.walletSigner = walletSigner;
    this.blockchainId = blockchainId;
    this.nodeAddress = nodeAddress;
    this.assetStore = assetStore;

    this.requireCheckAllowance = options?.requireCheckAllowance ?? true;
    this.requireCheckBalance = options?.requireCheckBalance ?? true;
  }

  private hexToBytes32(s: string): `0x${string}` {
    const bytes = hexToBytes(s as Hex);
    if (bytes.length !== 32) {
      throw new Error(`invalid length: expected 32 bytes, got ${bytes.length}`);
    }
    // Convert Uint8Array back to hex string
    return `0x${Array.from(bytes)
      .map((b) => b.toString(16).padStart(2, '0'))
      .join('')}` as `0x${string}`;
  }

  // ========= Getters - IVault =========

  async getAccountsBalances(accounts: Address[], tokens: Address[]): Promise<Decimal[][]> {
    if (accounts.length === 0 || tokens.length === 0) {
      return [];
    }

    const result: Decimal[][] = [];
    for (const account of accounts) {
      const accountBalances: Decimal[] = [];
      for (const token of tokens) {
        const balance = (await this.evmClient.readContract({
          address: this.contractAddress,
          abi: ChannelHubAbi,
          functionName: 'getAccountBalance',
          args: [account, token],
        })) as bigint;
        accountBalances.push(new Decimal(balance.toString()));
      }
      result.push(accountBalances);
    }

    return result;
  }

  private async getAllowance(asset: string, owner: Address): Promise<Decimal> {
    const tokenAddress = await this.assetStore.getTokenAddress(asset, this.blockchainId);
    const erc20 = newERC20(tokenAddress, this.evmClient);
    const allowance = await erc20.allowance(owner, this.contractAddress);

    const decimals = await this.assetStore.getTokenDecimals(this.blockchainId, tokenAddress);
    return new Decimal(allowance.toString()).div(Decimal.pow(10, decimals));
  }

  private async getTokenBalance(asset: string, account: Address): Promise<Decimal> {
    const tokenAddress = await this.assetStore.getTokenAddress(asset, this.blockchainId);
    const erc20 = newERC20(tokenAddress, this.evmClient);
    const balance = await erc20.balanceOf(account);

    const decimals = await this.assetStore.getTokenDecimals(this.blockchainId, tokenAddress);
    return new Decimal(balance.toString()).div(Decimal.pow(10, decimals));
  }

  // ========= Getters - ChannelHub =========

  async getNodeBalance(token: Address): Promise<Decimal> {
    const balance = (await this.evmClient.readContract({
      address: this.contractAddress,
      abi: ChannelHubAbi,
      functionName: 'getAccountBalance',
      args: [this.nodeAddress, token],
    })) as bigint;

    const decimals = await this.assetStore.getTokenDecimals(this.blockchainId, token);
    return new Decimal(balance.toString()).div(Decimal.pow(10, decimals));
  }

  async getOpenChannels(user: Address): Promise<string[]> {
    const channelIds = (await this.evmClient.readContract({
      address: this.contractAddress,
      abi: ChannelHubAbi,
      functionName: 'getOpenChannels',
      args: [user],
    })) as `0x${string}`[];
    return channelIds.map((id) => id);
  }

  async getHomeChannelData(homeChannelId: string): Promise<core.HomeChannelDataResponse> {
    const channelIdBytes = this.hexToBytes32(homeChannelId);

    const data = (await this.evmClient.readContract({
      address: this.contractAddress,
      abi: ChannelHubAbi,
      functionName: 'getChannelData',
      args: [channelIdBytes],
    })) as any;

    const lastState = contractStateToCoreState(data.lastState, homeChannelId);

    return {
      definition: {
        nonce: data.definition.nonce,
        challenge: data.definition.challengeDuration,
      },
      node: data.definition.node,
      lastState,
      challengeExpiry: data.challengeExpiry,
    };
  }

  // Note: Escrow methods would need additional contract methods in the ABI
  // These are placeholders based on the Go SDK structure
  async getEscrowDepositData(_escrowChannelId: string): Promise<core.EscrowDepositDataResponse> {
    throw new Error('getEscrowDepositData not implemented - needs contract ABI update');
  }

  async getEscrowWithdrawalData(
    _escrowChannelId: string
  ): Promise<core.EscrowWithdrawalDataResponse> {
    throw new Error('getEscrowWithdrawalData not implemented - needs contract ABI update');
  }

  // ========= IVault Functions =========

  async deposit(node: Address, token: Address, amount: Decimal): Promise<string> {
    const decimals = await this.assetStore.getTokenDecimals(this.blockchainId, token);
    const amountBig = decimalToBigInt(amount, decimals);

    const hash = await this.walletSigner.writeContract({
      address: this.contractAddress,
      abi: ChannelHubAbi,
      functionName: 'depositToVault',
      args: [node, token, amountBig],
      chain: undefined,
    } as any);
    return hash;
  }

  async withdraw(node: Address, token: Address, amount: Decimal): Promise<string> {
    const decimals = await this.assetStore.getTokenDecimals(this.blockchainId, token);
    const amountBig = decimalToBigInt(amount, decimals);

    const hash = await this.walletSigner.writeContract({
      address: this.contractAddress,
      abi: ChannelHubAbi,
      functionName: 'withdrawFromVault',
      args: [node, token, amountBig],
      chain: undefined,
    } as any);
    return hash;
  }

  // ========= Channel Lifecycle =========

  async create(def: core.ChannelDefinition, initState: core.State): Promise<string> {
    const contractDef = coreDefToContractDef(
      def,
      initState.asset,
      initState.userWallet,
      this.nodeAddress
    );

    const contractState = await coreStateToContractState(initState, (blockchainId, tokenAddress) =>
      this.assetStore.getTokenDecimals(blockchainId, tokenAddress)
    );

    // Check allowance and balance for deposits
    const lastTransition = initState.transitions[initState.transitions.length - 1];
    if (
      lastTransition &&
      (lastTransition.type === core.TransitionType.HomeDeposit ||
        lastTransition.type === core.TransitionType.EscrowDeposit)
    ) {
      if (this.requireCheckAllowance) {
        const allowance = await this.getAllowance(initState.asset, initState.userWallet);
        if (allowance.lessThan(lastTransition.amount)) {
          throw new Error('Allowance is not sufficient to cover the deposit amount');
        }
      }

      if (this.requireCheckBalance) {
        const balance = await this.getTokenBalance(initState.asset, initState.userWallet);
        if (balance.lessThan(lastTransition.amount)) {
          throw new Error('Balance is not sufficient to cover the deposit amount');
        }
      }
    }

    const hash = await this.walletSigner.writeContract({
      address: this.contractAddress,
      abi: ChannelHubAbi,
      functionName: 'createChannel',
      args: [contractDef, contractState],
      chain: undefined,
    } as any);
    return hash;
  }

  async checkpoint(candidate: core.State): Promise<string> {
    if (!candidate.homeChannelId) {
      throw new Error('Candidate state must have a home channel ID');
    }

    const channelIdBytes = this.hexToBytes32(candidate.homeChannelId);

    const contractCandidate = await coreStateToContractState(
      candidate,
      (blockchainId, tokenAddress) => this.assetStore.getTokenDecimals(blockchainId, tokenAddress)
    );

    // Check for deposit intent
    const lastTransition = candidate.transitions[candidate.transitions.length - 1];
    if (lastTransition?.type === core.TransitionType.HomeDeposit) {
      if (this.requireCheckAllowance) {
        const allowance = await this.getAllowance(candidate.asset, candidate.userWallet);
        if (allowance.lessThan(lastTransition.amount)) {
          throw new Error('Allowance is not sufficient to cover the deposit amount');
        }
      }

      if (this.requireCheckBalance) {
        const balance = await this.getTokenBalance(candidate.asset, candidate.userWallet);
        if (balance.lessThan(lastTransition.amount)) {
          throw new Error('Balance is not sufficient to cover the deposit amount');
        }
      }

      const hash = await this.walletSigner.writeContract({
        address: this.contractAddress,
        abi: ChannelHubAbi,
        functionName: 'depositToChannel',
        args: [channelIdBytes, contractCandidate],
        chain: undefined,
      } as any);
      return hash;
    }

    // Check for withdrawal intent
    if (lastTransition?.type === core.TransitionType.HomeWithdrawal) {
      const hash = await this.walletSigner.writeContract({
        address: this.contractAddress,
        abi: ChannelHubAbi,
        functionName: 'withdrawFromChannel',
        args: [channelIdBytes, contractCandidate],
        chain: undefined,
      } as any);
      return hash;
    }

    // Default checkpoint
    const hash = await this.walletSigner.writeContract({
      address: this.contractAddress,
      abi: ChannelHubAbi,
      functionName: 'checkpointChannel',
      args: [channelIdBytes, contractCandidate, []],
      chain: undefined,
    } as any);
    return hash;
  }

  async challenge(candidate: core.State, challengerSig: `0x${string}`): Promise<string> {
    if (!candidate.homeChannelId) {
      throw new Error('Candidate state must have a home channel ID');
    }

    const channelIdBytes = this.hexToBytes32(candidate.homeChannelId);

    const contractCandidate = await coreStateToContractState(
      candidate,
      (blockchainId, tokenAddress) => this.assetStore.getTokenDecimals(blockchainId, tokenAddress)
    );

    const hash = await this.walletSigner.writeContract({
      address: this.contractAddress,
      abi: ChannelHubAbi,
      functionName: 'challengeChannel',
      args: [channelIdBytes, contractCandidate, [], challengerSig],
      chain: undefined,
    } as any);
    return hash;
  }

  async close(candidate: core.State): Promise<string> {
    if (!candidate.homeChannelId) {
      throw new Error('Candidate state must have a home channel ID');
    }

    const channelIdBytes = this.hexToBytes32(candidate.homeChannelId);

    const contractCandidate = await coreStateToContractState(
      candidate,
      (blockchainId, tokenAddress) => this.assetStore.getTokenDecimals(blockchainId, tokenAddress)
    );

    // Verify close intent
    const lastTransition = candidate.transitions[candidate.transitions.length - 1];
    if (lastTransition?.type !== core.TransitionType.Finalize) {
      throw new Error('Unsupported intent for close');
    }

    const hash = await this.walletSigner.writeContract({
      address: this.contractAddress,
      abi: ChannelHubAbi,
      functionName: 'closeChannel',
      args: [channelIdBytes, contractCandidate, []],
      chain: undefined,
    } as any);
    return hash;
  }

  // ========= Escrow Operations =========
  // Note: These would need the full escrow methods in the ABI

  async initiateEscrowDeposit(_def: core.ChannelDefinition, _initState: core.State): Promise<string> {
    throw new Error('initiateEscrowDeposit not implemented - needs contract ABI update');
  }

  async challengeEscrowDeposit(
    _candidate: core.State,
    _challengerSig: `0x${string}`
  ): Promise<string> {
    throw new Error('challengeEscrowDeposit not implemented - needs contract ABI update');
  }

  async finalizeEscrowDeposit(_candidate: core.State): Promise<string> {
    throw new Error('finalizeEscrowDeposit not implemented - needs contract ABI update');
  }

  async initiateEscrowWithdrawal(
    _def: core.ChannelDefinition,
    _initState: core.State
  ): Promise<string> {
    throw new Error('initiateEscrowWithdrawal not implemented - needs contract ABI update');
  }

  async challengeEscrowWithdrawal(
    _candidate: core.State,
    _challengerSig: `0x${string}`
  ): Promise<string> {
    throw new Error('challengeEscrowWithdrawal not implemented - needs contract ABI update');
  }

  async finalizeEscrowWithdrawal(_candidate: core.State): Promise<string> {
    throw new Error('finalizeEscrowWithdrawal not implemented - needs contract ABI update');
  }

  async migrateChannelHere(_def: core.ChannelDefinition, _candidate: core.State): Promise<string> {
    throw new Error('migrateChannelHere not implemented - needs contract ABI update');
  }
}

/**
 * Create a new blockchain client
 */
export function newClient(
  contractAddress: Address,
  evmClient: EVMClient,
  walletSigner: WalletSigner,
  blockchainId: bigint,
  nodeAddress: Address,
  assetStore: AssetStore,
  options?: ClientOptions
): Client {
  return new Client(
    contractAddress,
    evmClient,
    walletSigner,
    blockchainId,
    nodeAddress,
    assetStore,
    options
  );
}
