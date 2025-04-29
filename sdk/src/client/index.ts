import { Account, PublicClient, WalletClient, Chain, Transport, ParseAccount, Hash, zeroAddress } from "viem";

import { NitroliteService, Erc20Service } from "./services";
import {
    State,
    ChannelId,
    NitroliteClientConfig,
    CreateChannelParams,
    CheckpointChannelParams,
    ChallengeChannelParams,
    CloseChannelParams,
    AccountInfo,
} from "./types";
import * as Errors from "../errors";
import { ContractAddresses } from "../abis";
import { _prepareAndSignFinalState, _prepareAndSignInitialState } from "./state";
import { NitroliteTransactionPreparer, PreparerDependencies } from "./prepare";

/**
 * The main client class for interacting with the Nitrolite SDK.
 * Provides high-level methods for managing state channels and funds.
 */
export class NitroliteClient {
    public readonly publicClient: PublicClient;
    public readonly walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;
    public readonly account: ParseAccount<Account>;
    public readonly addresses: ContractAddresses;
    public readonly challengeDuration: bigint;
    public readonly txPreparer: NitroliteTransactionPreparer;
    private readonly stateWalletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;
    private readonly nitroliteService: NitroliteService;
    private readonly erc20Service: Erc20Service;
    private readonly sharedDeps: PreparerDependencies;

    constructor(config: NitroliteClientConfig) {
        if (!config.publicClient) throw new Errors.MissingParameterError("publicClient");
        if (!config.walletClient) throw new Errors.MissingParameterError("walletClient");
        if (!config.walletClient.account) throw new Errors.MissingParameterError("walletClient.account");
        if (!config.challengeDuration) throw new Errors.MissingParameterError("challengeDuration");
        if (!config.addresses?.custody) throw new Errors.MissingParameterError("addresses.custody");
        if (!config.addresses?.adjudicator) throw new Errors.MissingParameterError("addresses.adjudicator");
        if (!config.addresses?.guestAddress) throw new Errors.MissingParameterError("addresses.guestAddress");
        if (!config.addresses?.tokenAddress) throw new Errors.MissingParameterError("addresses.tokenAddress");

        this.publicClient = config.publicClient;
        this.walletClient = config.walletClient;
        // Determine which wallet client to use for state signing
        this.stateWalletClient = config.stateWalletClient ?? config.walletClient;
        this.account = config.walletClient.account;
        this.addresses = config.addresses;
        this.challengeDuration = config.challengeDuration;

        this.nitroliteService = new NitroliteService(this.publicClient, this.addresses, this.walletClient, this.account);
        this.erc20Service = new Erc20Service(this.publicClient, this.walletClient);

        this.sharedDeps = {
            nitroliteService: this.nitroliteService,
            erc20Service: this.erc20Service,
            addresses: this.addresses,
            account: this.account,
            walletClient: this.walletClient,
            challengeDuration: this.challengeDuration,
            stateWalletClient: this.stateWalletClient,
        };

        this.txPreparer = new NitroliteTransactionPreparer(this.sharedDeps);
    }

    /**
     * Deposits tokens or ETH into the custody contract.
     * Handles ERC20 approval if necessary.
     * @param amount The amount of tokens/ETH to deposit.
     * @returns The transaction hash.
     */
    async deposit(amount: bigint): Promise<Hash> {
        const owner = this.account.address;
        const spender = this.addresses.custody;
        const tokenAddress = this.addresses.tokenAddress;

        if (tokenAddress !== zeroAddress) {
            const allowance = await this.erc20Service.getTokenAllowance(tokenAddress, owner, spender);
            if (allowance < amount) {
                try {
                    await this.erc20Service.approve(tokenAddress, spender, amount);
                } catch (err) {
                    const error = new Errors.TokenError("Failed to approve tokens for deposit");
                    throw error;
                }
            }
        }

        try {
            return await this.nitroliteService.deposit(tokenAddress, amount);
        } catch (err) {
            throw new Errors.ContractCallError("Failed to execute deposit on contract", err as Error);
        }
    }

    /**
     * Creates a new state channel on-chain.
     * Constructs the initial state, signs it, and calls the custody contract.
     * @param params Parameters for channel creation. See {@link CreateChannelParams}.
     * @returns The channel ID, the signed initial state, and the transaction hash.
     */
    async createChannel(params: CreateChannelParams): Promise<{ channelId: ChannelId; initialState: State; txHash: Hash }> {
        try {
            const { channel, initialState, channelId } = await _prepareAndSignInitialState(this.sharedDeps, params);
            const txHash = await this.nitroliteService.createChannel(channel, initialState);

            return { channelId, initialState, txHash };
        } catch (err) {
            throw new Errors.ContractCallError("Failed to execute createChannel on contract", err as Error);
        }
    }

    /**
     * Deposits tokens and creates a new channel in a single operation.
     * @param depositAmount The amount of tokens to deposit.
     * @param params Parameters for channel creation. See {@link CreateChannelParams}.
     * @returns An object containing the channel ID, initial state, deposit transaction hash, and create channel transaction hash.
     */
    async depositAndCreateChannel(
        depositAmount: bigint,
        params: CreateChannelParams
    ): Promise<{ channelId: ChannelId; initialState: State; depositTxHash: Hash; createChannelTxHash: Hash }> {
        const depositTxHash = await this.deposit(depositAmount);
        const { channelId, initialState, txHash } = await this.createChannel(params);

        return { channelId, initialState, depositTxHash: depositTxHash, createChannelTxHash: txHash };
    }

    /**
     * Checkpoints a state on-chain.
     * Requires the state to be signed by both participants.
     * @param params Parameters for checkpointing the state. See {@link CheckpointChannelParams}.
     * @returns The transaction hash.
     */
    async checkpointChannel(params: CheckpointChannelParams): Promise<Hash> {
        const { channelId, candidateState, proofStates = [] } = params;

        if (!candidateState.sigs || candidateState.sigs.length < 2) {
            throw new Errors.InvalidParameterError("Candidate state for checkpoint must be signed by both participants.");
        }

        try {
            return await this.nitroliteService.checkpoint(channelId, candidateState, proofStates);
        } catch (err) {
            throw new Errors.ContractCallError("Failed to execute checkpointChannel on contract", err as Error);
        }
    }

    /**
     * Challenges a channel on-chain with a candidate state.
     * Used when the counterparty is unresponsive. Requires the candidate state to be signed by the challenger.
     * @param params Parameters for challenging the channel. See {@link CreateChannelParams}.
     * @returns The transaction hash.
     */
    async challengeChannel(params: ChallengeChannelParams): Promise<Hash> {
        const { channelId, candidateState, proofStates = [] } = params;

        try {
            return await this.nitroliteService.challenge(channelId, candidateState, proofStates);
        } catch (err) {
            throw new Errors.ContractCallError("Failed to execute challengeChannel on contract", err as Error);
        }
    }

    /**
     * Closes a channel on-chain using a mutually agreed final state.
     * Requires the final state signed by both participants.
     * @param params Parameters for closing the channel. See {@link CloseChannelParams}.
     * @returns The transaction hash.
     */
    async closeChannel(params: CloseChannelParams): Promise<Hash> {
        try {
            const { finalStateWithSigs, channelId } = await _prepareAndSignFinalState(this.sharedDeps, params);

            return await this.nitroliteService.close(channelId, finalStateWithSigs);
        } catch (err) {
            throw new Errors.ContractCallError("Failed to execute closeChannel on contract", err as Error);
        }
    }

    /**
     * Withdraws tokens previously deposited into the custody contract.
     * This does not withdraw funds locked in active channels.
     * @param amount The amount of tokens/ETH to withdraw.
     * @returns The transaction hash.
     */
    async withdrawal(amount: bigint): Promise<Hash> {
        const tokenAddress = this.addresses.tokenAddress;

        try {
            return await this.nitroliteService.withdraw(tokenAddress, amount);
        } catch (err) {
            throw new Errors.ContractCallError("Failed to execute withdrawDeposit on contract", err as Error);
        }
    }

    /**
     * Retrieves a list of channel IDs associated with a specific account.
     * @returns An array of Channel IDs.
     */
    async getAccountChannels(): Promise<ChannelId[]> {
        try {
            return await this.nitroliteService.getAccountChannels(this.account.address);
        } catch (err) {
            throw err;
        }
    }

    /**
     * Retrieves deposit and lock information for an account regarding a specific token.
     * @returns Account info including available, locked amounts and channel count.
     */
    async getAccountInfo(): Promise<AccountInfo> {
        try {
            return await this.nitroliteService.getAccountInfo(this.account.address, this.addresses.tokenAddress);
        } catch (err) {
            throw err;
        }
    }

    /**
     * Approves the custody contract to spend a specified amount of an ERC20 token.
     * @returns The transaction hash.
     */
    async approveTokens(amount: bigint): Promise<Hash> {
        const spender = this.addresses.custody;
        const tokenAddress = this.addresses.tokenAddress;

        try {
            return await this.erc20Service.approve(tokenAddress, spender, amount);
        } catch (err) {
            throw new Errors.TokenError("Failed to approve tokens", undefined, undefined, undefined, undefined, err as Error);
        }
    }

    /**
     * Gets the current allowance granted by an owner to a spender for a specific ERC20 token.
     * @returns The allowance amount as a bigint.
     */
    async getTokenAllowance(): Promise<bigint> {
        const tokenAddress = this.addresses.tokenAddress;
        const targetOwner = this.account.address;
        const targetSpender = this.addresses.custody;

        try {
            return await this.erc20Service.getTokenAllowance(tokenAddress, targetOwner, targetSpender);
        } catch (err) {
            throw new Errors.TokenError("Failed to get token allowance", undefined, undefined, undefined, undefined, err as Error);
        }
    }

    /**
     * Gets the balance of a specific ERC20 token for an account.
     * @returns The token balance as a bigint.
     */
    async getTokenBalance(): Promise<bigint> {
        const tokenAddress = this.addresses.tokenAddress;
        const targetAccount = this.account.address;
        try {
            return await this.erc20Service.getTokenBalance(tokenAddress, targetAccount);
        } catch (err) {
            throw new Errors.TokenError("Failed to get token balance", undefined, undefined, undefined, undefined, err as Error);
        }
    }
}
