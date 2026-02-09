import { Account, Address, Chain, Hash, ParseAccount, PublicClient, Transport, WalletClient, zeroAddress } from 'viem';

import { ContractAddresses } from '../abis';
import * as Errors from '../errors';
import { NitroliteTransactionPreparer, PreparerDependencies } from './prepare';
import { Erc20Service, NitroliteService, waitForTransaction } from './services';
import {
    _prepareAndSignChallengeState,
    _prepareAndSignFinalState,
    _prepareAndSignInitialState,
} from './state';
import {
    ChallengeChannelParams,
    ChannelData,
    ChannelId,
    CheckpointChannelParams,
    CloseChannelParams,
    CreateChannelParams,
    NitroliteClientConfig,
    State,
} from './types';
import { StateSigner } from './signer';

const CUSTODY_MIN_CHALLENGE_DURATION = 3600;

/**
 * The main client class for interacting with the Nitrolite V1 SDK.
 * Provides high-level methods for managing state channels and funds.
 */
export class NitroliteClient {
    /** Service for interacting with the Custody contract. */
    readonly nitroliteService: NitroliteService;
    /** Service for interacting with ERC20 tokens. */
    readonly erc20Service: Erc20Service;
    /** Transaction preparer for Account Abstraction and batching. */
    readonly txPreparer: NitroliteTransactionPreparer;

    private readonly publicClient: PublicClient;
    private readonly walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;
    private readonly stateSigner: StateSigner;
    private readonly addresses: ContractAddresses;
    private readonly chainId: number;
    private readonly challengeDuration: number;

    /**
     * Creates an instance of NitroliteClient.
     * @param config - Configuration for the client. See {@link NitroliteClientConfig}.
     */
    constructor(config: NitroliteClientConfig) {
        if (!config.publicClient) {
            throw new Errors.MissingParameterError('publicClient');
        }

        if (!config.walletClient) {
            throw new Errors.MissingParameterError('walletClient');
        }

        if (!config.stateSigner) {
            throw new Errors.MissingParameterError('stateSigner');
        }

        if (!config.addresses) {
            throw new Errors.MissingParameterError('addresses');
        }

        if (config.challengeDuration < CUSTODY_MIN_CHALLENGE_DURATION) {
            throw new Errors.InvalidParameterError(
                `Challenge duration must be at least ${CUSTODY_MIN_CHALLENGE_DURATION} seconds`,
            );
        }

        this.publicClient = config.publicClient;
        this.walletClient = config.walletClient;
        this.stateSigner = config.stateSigner;
        this.addresses = config.addresses;
        this.chainId = config.chainId;
        this.challengeDuration = config.challengeDuration;

        this.nitroliteService = new NitroliteService(
            this.publicClient,
            this.addresses,
            this.walletClient,
            this.walletClient.account,
        );

        this.erc20Service = new Erc20Service(this.publicClient, this.walletClient, this.walletClient.account);

        const preparerDeps: PreparerDependencies = {
            nitroliteService: this.nitroliteService,
            erc20Service: this.erc20Service,
            addresses: this.addresses,
            account: this.walletClient.account,
            walletClient: this.walletClient,
            stateSigner: this.stateSigner,
            challengeDuration: this.challengeDuration,
            chainId: this.chainId,
        };

        this.txPreparer = new NitroliteTransactionPreparer(preparerDeps);
    }

    /**
     * Deposit tokens or ETH into the vault.
     * @param node Address of the node.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to deposit.
     * @returns Transaction hash.
     */
    async deposit(node: Address, tokenAddress: Address, amount: bigint): Promise<Hash> {
        return this.nitroliteService.deposit(node, tokenAddress, amount);
    }

    /**
     * Create a new channel.
     * @param params Parameters for channel creation. See {@link CreateChannelParams}.
     * @returns Transaction hash.
     */
    async createChannel(params: CreateChannelParams): Promise<Hash> {
        const { initialState } = await _prepareAndSignInitialState(
            {
                nitroliteService: this.nitroliteService,
                erc20Service: this.erc20Service,
                addresses: this.addresses,
                account: this.walletClient.account,
                walletClient: this.walletClient,
                stateSigner: this.stateSigner,
                challengeDuration: this.challengeDuration,
                chainId: this.chainId,
            },
            params,
        );

        return await this.nitroliteService.createChannel(params.definition, initialState);
    }

    /**
     * Deposit tokens or ETH to vault and create a channel in a single flow.
     * This will handle token approval (if needed), deposit, and channel creation sequentially.
     * @param node Address of the node.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param depositAmount Amount to deposit.
     * @param params Parameters for channel creation. See {@link CreateChannelParams}.
     * @returns Transaction hash from the channel creation.
     */
    async depositAndCreateChannel(
        node: Address,
        tokenAddress: Address,
        depositAmount: bigint,
        params: CreateChannelParams,
    ): Promise<Hash> {
        // Prepare all transactions (approval + deposit + createChannel)
        const transactions = await this.txPreparer.prepareDepositAndCreateChannelTransactions(
            node,
            tokenAddress,
            depositAmount,
            params,
        );

        // Execute all transactions sequentially
        let lastTxHash: Hash = '0x' as Hash;
        for (const tx of transactions) {
            lastTxHash = await this.walletClient.writeContract(tx as any);
            await this.waitForTransaction(lastTxHash);
        }

        return lastTxHash;
    }

    /**
     * Checkpoint a state on-chain.
     * @param params Parameters for checkpointing. See {@link CheckpointChannelParams}.
     * @returns Transaction hash.
     */
    async checkpointChannel(params: CheckpointChannelParams): Promise<Hash> {
        const { channelId, candidateState, proofs = [] } = params;

        return await this.nitroliteService.checkpointChannel(channelId, candidateState, proofs);
    }

    /**
     * Challenge a channel state.
     * @param params Parameters for challenging. See {@link ChallengeChannelParams}.
     * @returns Transaction hash.
     */
    async challengeChannel(params: ChallengeChannelParams): Promise<Hash> {
        const { channelId, candidateState, proofs, challengerSig } = await _prepareAndSignChallengeState(
            {
                nitroliteService: this.nitroliteService,
                erc20Service: this.erc20Service,
                addresses: this.addresses,
                account: this.walletClient.account,
                walletClient: this.walletClient,
                stateSigner: this.stateSigner,
                challengeDuration: this.challengeDuration,
                chainId: this.chainId,
            },
            params,
        );

        return await this.nitroliteService.challengeChannel(channelId, candidateState, proofs, challengerSig);
    }

    /**
     * Close a channel cooperatively or after challenge expiry.
     * @param params Parameters for closing. See {@link CloseChannelParams}.
     * @returns Transaction hash.
     */
    async closeChannel(params: CloseChannelParams): Promise<Hash> {
        const { channelId, finalState, proofs } = await _prepareAndSignFinalState(
            {
                nitroliteService: this.nitroliteService,
                erc20Service: this.erc20Service,
                addresses: this.addresses,
                account: this.walletClient.account,
                walletClient: this.walletClient,
                stateSigner: this.stateSigner,
                challengeDuration: this.challengeDuration,
                chainId: this.chainId,
            },
            params,
        );

        return await this.nitroliteService.closeChannel(channelId, finalState, proofs || []);
    }

    /**
     * Withdraw funds from the vault.
     * @param to Recipient address.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to withdraw.
     * @returns Transaction hash.
     */
    async withdraw(to: Address, tokenAddress: Address, amount: bigint): Promise<Hash> {
        return this.nitroliteService.withdraw(to, tokenAddress, amount);
    }

    /**
     * Get the list of open channel IDs for the current user.
     * @returns Array of Channel IDs.
     */
    async getOpenChannels(): Promise<ChannelId[]> {
        const userAddress = this.walletClient.account.address;
        return this.nitroliteService.getOpenChannels(userAddress);
    }

    /**
     * Get account balance in the vault for a specific node and token.
     * @param node Address of the node.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @returns Balance as bigint.
     */
    async getAccountBalance(node: Address, tokenAddress: Address): Promise<bigint> {
        return this.nitroliteService.getAccountBalance(node, tokenAddress);
    }

    /**
     * Get channel data for a specific channel ID.
     * @param channelId ID of the channel.
     * @returns Channel data including definition, status, last state, and challenge expiry.
     */
    async getChannelData(channelId: ChannelId): Promise<ChannelData> {
        return this.nitroliteService.getChannelData(channelId);
    }

    /**
     * Approve tokens for the custody contract.
     * @param tokenAddress Address of the token.
     * @param amount Amount to approve.
     * @returns Transaction hash.
     */
    async approveTokens(tokenAddress: Address, amount: bigint): Promise<Hash> {
        if (tokenAddress === zeroAddress) {
            throw new Errors.InvalidParameterError('ETH does not require approval.');
        }

        const spender = this.addresses.custody;
        return this.erc20Service.approve(tokenAddress, spender, amount);
    }

    /**
     * Get the current token allowance for the custody contract.
     * @param tokenAddress Address of the token.
     * @returns Current allowance as bigint.
     */
    async getTokenAllowance(tokenAddress: Address): Promise<bigint> {
        if (tokenAddress === zeroAddress) {
            throw new Errors.InvalidParameterError('ETH does not have an allowance.');
        }

        const owner = this.walletClient.account.address;
        const spender = this.addresses.custody;
        return this.erc20Service.getTokenAllowance(tokenAddress, owner, spender);
    }

    /**
     * Get the token balance for the current user.
     * @param tokenAddress Address of the token.
     * @returns Balance as bigint.
     */
    async getTokenBalance(tokenAddress: Address): Promise<bigint> {
        const owner = this.walletClient.account.address;
        return this.erc20Service.getTokenBalance(tokenAddress, owner);
    }

    /**
     * Wait for a transaction to be confirmed.
     * @param txHash Transaction hash.
     * @returns Transaction receipt.
     */
    async waitForTransaction(txHash: Hash) {
        return waitForTransaction(this.publicClient, txHash);
    }
}
