import { Account, Address, Chain, Hash, ParseAccount, PublicClient, Transport, WalletClient, zeroAddress } from 'viem';

import { ContractAddresses } from '../abis';
import * as Errors from '../errors';
import { NitroliteTransactionPreparer, PreparerDependencies } from './prepare';
import { Erc20Service, NitroliteService, waitForTransaction } from './services';
import {
    _prepareAndSignChallengeState,
    _prepareAndSignFinalState,
    _prepareAndSignInitialState,
    _prepareAndSignResizeState,
} from './state';
import {
    ChallengeChannelParams,
    ChannelData,
    ChannelId,
    CheckpointChannelParams,
    CloseChannelParams,
    CreateChannelParams,
    NitroliteClientConfig,
    ResizeChannelParams,
    State,
} from './types';
import { StateSigner } from './signer';

const CUSTODY_MIN_CHALLENGE_DURATION = 3600n;

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
    public readonly chainId: number;
    private readonly stateSigner: StateSigner;
    private readonly nitroliteService: NitroliteService;
    private readonly erc20Service: Erc20Service;
    private readonly sharedDeps: PreparerDependencies;

    constructor(config: NitroliteClientConfig) {
        if (!config.publicClient) throw new Errors.MissingParameterError('publicClient');
        if (!config.walletClient) throw new Errors.MissingParameterError('walletClient');
        if (!config.walletClient.account) throw new Errors.MissingParameterError('walletClient.account');
        if (!config.challengeDuration) throw new Errors.MissingParameterError('challengeDuration');
        if (config.challengeDuration < CUSTODY_MIN_CHALLENGE_DURATION)
            throw new Errors.InvalidParameterError(
                `The minimum challenge duration is ${CUSTODY_MIN_CHALLENGE_DURATION} seconds`,
            );
        if (!config.addresses?.custody) throw new Errors.MissingParameterError('addresses.custody');
        if (!config.addresses?.adjudicator) throw new Errors.MissingParameterError('addresses.adjudicator');
        if (!config.chainId) throw new Errors.MissingParameterError('chainId');

        this.publicClient = config.publicClient;
        this.walletClient = config.walletClient;
        this.stateSigner = config.stateSigner;
        this.account = config.walletClient.account;
        this.addresses = config.addresses;
        this.challengeDuration = config.challengeDuration;
        this.chainId = config.chainId;

        this.nitroliteService = new NitroliteService(
            this.publicClient,
            this.addresses,
            this.walletClient,
            this.account,
        );
        this.erc20Service = new Erc20Service(this.publicClient, this.walletClient);

        this.sharedDeps = {
            nitroliteService: this.nitroliteService,
            erc20Service: this.erc20Service,
            addresses: this.addresses,
            account: this.account,
            walletClient: this.walletClient,
            challengeDuration: this.challengeDuration,
            stateSigner: this.stateSigner,
            chainId: this.chainId,
        };

        this.txPreparer = new NitroliteTransactionPreparer(this.sharedDeps);
    }

    /**
     * Deposits tokens or ETH into the custody contract.
     * Handles ERC20 approval if necessary.
     * @param tokenAddress The address of the token to deposit.
     * @param amount The amount of tokens/ETH to deposit.
     * @returns The transaction hash.
     */
    async deposit(tokenAddress: Address, amount: bigint): Promise<Hash> {
        const owner = this.account.address;
        const spender = this.addresses.custody;

        if (tokenAddress !== zeroAddress) {
            const allowance = await this.erc20Service.getTokenAllowance(tokenAddress, owner, spender);
            if (allowance < amount) {
                try {
                    const hash = await this.erc20Service.approve(tokenAddress, spender, amount);
                    await waitForTransaction(this.publicClient, hash);
                } catch (err) {
                    const error = new Errors.TokenError('Failed to approve tokens for deposit');
                    throw error;
                }
            }
        }

        try {
            const depositHash = await this.nitroliteService.deposit(tokenAddress, amount);
            await waitForTransaction(this.publicClient, depositHash);

            return depositHash;
        } catch (err) {
            throw new Errors.ContractCallError('Failed to execute deposit on contract', err as Error);
        }
    }

    /**
     * Creates a new state channel on-chain.
     * Constructs the initial state, signs it, and calls the custody contract.
     * @param tokenAddress The address of the token for the channel.
     * @param params Parameters for channel creation. See {@link CreateChannelParams}.
     * @returns The channel ID, the signed initial state, and the transaction hash.
     */
    async createChannel(
        params: CreateChannelParams,
    ): Promise<{ channelId: ChannelId; initialState: State; txHash: Hash }> {
        try {
            const { initialState, channelId } = await _prepareAndSignInitialState(
                this.sharedDeps,
                params,
            );
            const txHash = await this.nitroliteService.createChannel(params.channel, initialState);

            return { channelId, initialState, txHash };
        } catch (err) {
            throw new Errors.ContractCallError('Failed to execute createChannel on contract', err as Error);
        }
    }

    /**
     * Deposits tokens and creates a new channel in a single operation.
     * Approves the custody contract to spend the tokens if necessary.
     * @param tokenAddress The address of the token to deposit and use for the channel.
     * @param depositAmount The amount of tokens to deposit.
     * @param params Parameters for channel creation. See {@link CreateChannelParams}.
     * @returns An object containing the channel ID, initial state, and the transaction hash.
     */
    async depositAndCreateChannel(
        tokenAddress: Address,
        depositAmount: bigint,
        params: CreateChannelParams,
    ): Promise<{ channelId: ChannelId; initialState: State; txHash: Hash }> {
        try {
            const owner = this.account.address;
            const spender = this.addresses.custody;
            const { initialState, channelId } = await _prepareAndSignInitialState(
                this.sharedDeps,
                params,
            );

            if (tokenAddress !== zeroAddress) {
                const allowance = await this.erc20Service.getTokenAllowance(tokenAddress, owner, spender);
                if (allowance < depositAmount) {
                    try {
                        const hash = await this.erc20Service.approve(tokenAddress, spender, depositAmount);
                        await waitForTransaction(this.publicClient, hash);
                    } catch (err) {
                        const error = new Errors.TokenError('Failed to approve tokens for deposit');
                        throw error;
                    }
                }
            }

            const txHash = await this.nitroliteService.depositAndCreateChannel(
                tokenAddress,
                depositAmount,
                params.channel,
                initialState,
            );

            return { channelId, initialState, txHash };
        } catch (err) {
            throw new Errors.ContractCallError('Failed to execute depositAndCreateChannel on contract', err as Error);
        }
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
            throw new Errors.InvalidParameterError(
                'Candidate state for checkpoint must be signed by both participants.',
            );
        }

        try {
            return await this.nitroliteService.checkpoint(channelId, candidateState, proofStates);
        } catch (err) {
            throw new Errors.ContractCallError('Failed to execute checkpointChannel on contract', err as Error);
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
        const { challengerSig } = await _prepareAndSignChallengeState(this.sharedDeps, params);

        try {
            return await this.nitroliteService.challenge(channelId, candidateState, proofStates, challengerSig);
        } catch (err) {
            throw new Errors.ContractCallError('Failed to execute challengeChannel on contract', err as Error);
        }
    }

    /**
     * Resize a channel on-chain using candidate state.
     * Requires the candidate state.
     * @param params Parameters for resizing the channel. See {@link ResizeChannelParams}.
     * @returns The transaction hash.
     */
    async resizeChannel(params: ResizeChannelParams): Promise<{resizeState: State; txHash: Hash}> {
        const { resizeStateWithSigs, proofs, channelId } = await _prepareAndSignResizeState(this.sharedDeps, params);

        try {
            return {resizeState: resizeStateWithSigs, txHash: await this.nitroliteService.resize(channelId, resizeStateWithSigs, proofs)};
        } catch (err) {
            throw new Errors.ContractCallError('Failed to execute resizeChannel on contract', err as Error);
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
            throw new Errors.ContractCallError('Failed to execute closeChannel on contract', err as Error);
        }
    }

    /**
     * Withdraws tokens previously deposited into the custody contract.
     * This does not withdraw funds locked in active channels.
     * @param tokenAddress The address of the token to withdraw.
     * @param amount The amount of tokens/ETH to withdraw.
     * @returns The transaction hash.
     */
    async withdrawal(tokenAddress: Address, amount: bigint): Promise<Hash> {
        try {
            return await this.nitroliteService.withdraw(tokenAddress, amount);
        } catch (err) {
            throw new Errors.ContractCallError('Failed to execute withdrawDeposit on contract', err as Error);
        }
    }

    /**
     * Retrieves a list of channel IDs associated with a specific account.
     * @returns An array of Channel IDs.
     */
    async getOpenChannels(): Promise<ChannelId[]> {
        try {
            return await this.nitroliteService.getOpenChannels(this.account.address);
        } catch (err) {
            throw err;
        }
    }

    /**
     * Retrieves deposit and lock information for an account regarding a specific token.
     * @param tokenAddress The address of the token to query.
     * @returns Account info including available, locked amounts and channel count.
     */
    async getAccountBalance(tokenAddress: Address): Promise<bigint>;
    async getAccountBalance(tokenAddress: Address[]): Promise<bigint[]>;
    async getAccountBalance(tokenAddress: Address | Address[]): Promise<bigint | bigint[]> {
        try {
            if (Array.isArray(tokenAddress)) {
                return await this.nitroliteService.getAccountBalance(this.account.address, tokenAddress);
            } else {
                return await this.nitroliteService.getAccountBalance(this.account.address, tokenAddress);
            }
        } catch (err) {
            throw err;
        }
    }

    /**
     * Retrieves the balances of all channels for a specific channel ID.
     * @param channelId The ID of the channel to query.
     * @param tokenAddress The address of the token to query balances for.
     * @returns An array of balances for the specified channel and token.
     */
    async getChannelBalance(channelId: ChannelId, tokenAddress: Address): Promise<bigint>;
    async getChannelBalance(channelId: ChannelId, tokenAddress: Address[]): Promise<bigint[]>;
    async getChannelBalance(channelId: ChannelId, tokenAddress: Address | Address[]): Promise<bigint | bigint[]> {
        try {
            if (Array.isArray(tokenAddress)) {
                return await this.nitroliteService.getChannelBalance(channelId, tokenAddress);
            } else {
                return await this.nitroliteService.getChannelBalance(channelId, tokenAddress);
            }
        } catch (err) {
            throw err;
        }
    }

    /**
     * Retrieves detailed channel data for a specific channel ID.
     * @param channelId The ID of the channel to query.
     * @returns An object containing channel data including participants, adjudicator, challenge duration, and allocations.
     */
    async getChannelData(channelId: ChannelId): Promise<ChannelData> {
        try {
            return await this.nitroliteService.getChannelData(channelId);
        } catch (err) {
            throw err;
        }
    }

    /**
     * Approves the custody contract to spend a specified amount of an ERC20 token.
     * @param tokenAddress The address of the ERC20 token to approve.
     * @param amount The amount to approve.
     * @returns The transaction hash.
     */
    async approveTokens(tokenAddress: Address, amount: bigint): Promise<Hash> {
        const spender = this.addresses.custody;

        try {
            return await this.erc20Service.approve(tokenAddress, spender, amount);
        } catch (err) {
            throw new Errors.TokenError(
                'Failed to approve tokens',
                undefined,
                undefined,
                undefined,
                undefined,
                err as Error,
            );
        }
    }

    /**
     * Gets the current allowance granted by an owner to a spender for a specific ERC20 token.
     * @param tokenAddress The address of the ERC20 token.
     * @returns The allowance amount as a bigint.
     */
    async getTokenAllowance(tokenAddress: Address): Promise<bigint> {
        const targetOwner = this.account.address;
        const targetSpender = this.addresses.custody;

        try {
            return await this.erc20Service.getTokenAllowance(tokenAddress, targetOwner, targetSpender);
        } catch (err) {
            throw new Errors.TokenError(
                'Failed to get token allowance',
                undefined,
                undefined,
                undefined,
                undefined,
                err as Error,
            );
        }
    }

    /**
     * Gets the balance of a specific ERC20 token for an account.
     * @param tokenAddress The address of the ERC20 token.
     * @returns The token balance as a bigint.
     */
    async getTokenBalance(tokenAddress: Address): Promise<bigint> {
        const targetAccount = this.account.address;
        try {
            return await this.erc20Service.getTokenBalance(tokenAddress, targetAccount);
        } catch (err) {
            throw new Errors.TokenError(
                'Failed to get token balance',
                undefined,
                undefined,
                undefined,
                undefined,
                err as Error,
            );
        }
    }
}
