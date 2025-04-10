import { Account, Address, PublicClient, WalletClient, Abi } from "viem";
import { AdjudicatorAbi, ContractAddresses } from "../abis";
import Errors from "../errors"; // Import Errors
import { Logger, defaultLogger } from "../config";

import { NitroliteClientConfig } from "./config";
import { ChannelOperations } from "./operations";
import { State, ChannelId, Channel, Role, Signature } from "./types";

/**
 * Main client for interacting with Nitrolite contracts
 */
export class NitroliteClient {
    public readonly publicClient: PublicClient;
    public readonly walletClient?: WalletClient;
    public readonly account?: Account;
    public readonly chainId: number;
    public readonly addresses: ContractAddresses;
    public readonly adjudicatorAbis: Record<string, Abi>;
    public readonly logger: Logger;

    private readonly operations: ChannelOperations;

    constructor(config: NitroliteClientConfig) {
        // TODO: Add more comprehensive configuration validation (e.g., address formats)
        if (!config.publicClient) {
            throw new Errors.MissingParameterError("publicClient");
        }

        // Use chain ID from the public client if not explicitly provided
        let chainId = config.chainId;
        if (!chainId) {
            chainId = config.publicClient.chain?.id;
            if (!chainId) {
                throw new Errors.MissingParameterError("chainId");
            }
        }

        if (!config.addresses) {
            throw new Errors.MissingParameterError("addresses");
        }

        this.publicClient = config.publicClient;
        this.walletClient = config.walletClient;
        this.account = config.account;
        this.chainId = chainId;
        this.addresses = config.addresses;
        this.logger = config.logger || defaultLogger;

        // Make sure adjudicators object exists
        if (!this.addresses.adjudicators) {
            this.addresses.adjudicators = {};
        }

        // Initialize adjudicator ABIs with defaults
        this.adjudicatorAbis = {
            base: AdjudicatorAbi,
            ...(config.adjudicatorAbis || {}),
        };

        // Initialize channel operations
        this.operations = new ChannelOperations(this.publicClient, this.walletClient, this.account, this.custodyAddress, this.logger);
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
    getAdjudicatorAbi(type: string = "base"): Abi {
        const abi = this.adjudicatorAbis[type];
        if (!abi) {
            // Fall back to base adjudicator ABI if specific type not found
            return this.adjudicatorAbis["base"] || AdjudicatorAbi;
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
     * @returns The adjudicator address
     */
    getAdjudicatorAddress(type: string): Address {
        // First try to get the requested adjudicator type
        const address = this.addresses.adjudicators[type];
        if (address) {
            return address;
        }

        // Otherwise throw an error with helpful message
        throw new Errors.ContractNotFoundError(`Adjudicator type: ${type}`, {
            availableTypes: Object.keys(this.addresses.adjudicators),
            requestedType: type,
        });
    }

    async deposit(tokenAddress: Address, amount: bigint): Promise<void> {
        return this.operations.deposit(tokenAddress, amount);
    }

    async withdraw(tokenAddress: Address, amount: bigint): Promise<void> {
        return this.operations.withdraw(tokenAddress, amount);
    }

    async getAccountChannels(account: Address): Promise<ChannelId[]> {
        return this.operations.getAccountChannels(account);
    }

    async getAccountInfo(account: Address, tokenAddress: Address): Promise<{ deposited: bigint; locked: bigint; channelCount: number }> {
        return this.operations.getAccountInfo(account, tokenAddress);
    }

    /**
     * Create a new channel
     */
    async createChannel(channel: Channel, deposit: State): Promise<void> {
        return this.operations.createChannel(channel, deposit);
    }

    /**
     * Join an existing channel
     */
    async joinChannel(channelId: ChannelId, index: number, sig: Signature): Promise<void> {
        return this.operations.joinChannel(channelId, index, sig);
    }

    /**
     * Close a channel with a mutually signed state
     */
    async closeChannel(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<void> {
        return this.operations.closeChannel(channelId, candidate, proofs);
    }

    /**
     * Close an old state and open a new one with a possibly different deposit
     */
    async resetChannel(channelId: ChannelId, candidate: State, proofs: State[], newChannel: Channel, newDeposit: State): Promise<void> {
        return this.operations.resetChannel(channelId, candidate, proofs, newChannel, newDeposit);
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
}
