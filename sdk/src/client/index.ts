import { Account, SimulateContractReturnType, PublicClient, WalletClient, Chain, Transport, ParseAccount, Hash, zeroAddress } from "viem";

import { NitroliteService, Erc20Service } from "./services";
import {
    Channel,
    State,
    ChannelId,
    NitroliteClientConfig,
    CreateChannelParams,
    CheckpointChannelParams,
    ChallengeChannelParams,
    CloseChannelParams,
    AccountInfo,
} from "./types";
import { getStateHash, generateChannelNonce, getChannelId, encoders, removeQuotesFromRS, signState } from "../utils";
import * as Errors from "../errors";
import { ContractAddresses } from "../abis";
import { MAGIC_NUMBERS } from "../config";
import { _prepareAndSignFinalState, _prepareAndSignInitialState } from "./state";

export type PreparedTransaction = SimulateContractReturnType["request"];

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
    private readonly nitroliteService: NitroliteService;
    private readonly erc20Service: Erc20Service;

    constructor(config: NitroliteClientConfig) {
        if (!config.publicClient) throw new Errors.MissingParameterError("publicClient");
        if (!config.walletClient) throw new Errors.MissingParameterError("walletClient");
        if (!config.walletClient.account) throw new Errors.MissingParameterError("walletClient.account");
        if (!config.addresses?.custody) throw new Errors.MissingParameterError("addresses.custody");
        if (!config.addresses?.adjudicators) throw new Errors.MissingParameterError("addresses.adjudicators");
        if (!config.addresses?.guestAddress) throw new Errors.MissingParameterError("addresses.guestAddress");
        if (!config.addresses?.tokenAddress) throw new Errors.MissingParameterError("addresses.tokenAddress");

        this.publicClient = config.publicClient;
        this.walletClient = config.walletClient;
        this.account = config.walletClient.account;
        this.addresses = config.addresses;
        this.challengeDuration = config.challengeDuration ?? 0n;

        this.nitroliteService = new NitroliteService(this.publicClient, this.addresses, this.walletClient, this.account);
        this.erc20Service = new Erc20Service(this.publicClient, this.walletClient);
    }

    /**
     * Prepares the transaction data necessary for a deposit operation,
     * including ERC20 approval if required.
     * Designed for use with Account Abstraction (UserOperations).
     * @param amount The amount of tokens/ETH to deposit.
     * @returns An array of PreparedTransaction objects (approve + deposit, or just deposit).
     */
    async prepareDepositTransactions(amount: bigint): Promise<PreparedTransaction[]> {
        const transactions: PreparedTransaction[] = [];
        const tokenAddress = this.addresses.tokenAddress;
        const spender = this.addresses.custody;
        const owner = this.account.address;

        if (tokenAddress !== zeroAddress) {
            const allowance = await this.erc20Service.getTokenAllowance(tokenAddress, owner, spender);
            if (allowance < amount) {
                try {
                    const approveTx = await this.erc20Service.prepareApprove(tokenAddress, spender, amount);
                    transactions.push(approveTx);
                } catch (err) {
                    // Throw a specific error indicating preparation failure
                    throw new Errors.ContractCallError("prepareApprove (for deposit)", err as Error, { tokenAddress, spender, amount });
                }
            }
        }

        try {
            const depositTx = await this.nitroliteService.prepareDeposit(tokenAddress, amount);
            transactions.push(depositTx);
        } catch (err) {
            throw new Errors.ContractCallError("prepareDeposit", err as Error, { tokenAddress, amount });
        }

        return transactions;
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
     * Prepares the transaction data for creating a new state channel.
     * Handles internal state construction and signing.
     * Designed for use with Account Abstraction (UserOperations).
     * @param params Parameters for channel creation.
     * @returns The prepared transaction data ({ to, data, value }).
     */
    async prepareCreateChannelTransaction(params: CreateChannelParams): Promise<PreparedTransaction> {
        try {
            const { channel, initialState } = await _prepareAndSignInitialState(this, params);

            const preparedTx = await this.nitroliteService.prepareCreateChannel(channel, initialState);

            return preparedTx;
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError("prepareCreateChannelTransaction", err as Error, { params });
        }
    }

    /**
     * Creates a new state channel on-chain.
     * Constructs the initial state, signs it, and calls the custody contract.
     * @param params Parameters for channel creation.
     * @returns The channel ID, the signed initial state, and the transaction hash.
     */
    async createChannel(params: CreateChannelParams): Promise<{ channelId: ChannelId; initialState: State; txHash: Hash }> {
        try {
            const { channel, initialState, channelId } = await _prepareAndSignInitialState(this, params);

            const txHash = await this.nitroliteService.createChannel(channel, initialState);

            return { channelId, initialState, txHash };
        } catch (err) {
            throw new Errors.ContractCallError("Failed to execute createChannel on contract", err as Error);
        }
    }

    /**
     * Prepares the transaction data for depositing funds and creating a channel in a single operation.
     * Includes potential ERC20 approval. Designed for batching with Account Abstraction (UserOperations).
     * @param depositAmount The amount to deposit.
     * @param params Parameters for channel creation.
     * @returns An array of PreparedTransaction objects (approve?, deposit, createChannel).
     */
    async prepareDepositAndCreateChannelTransactions(depositAmount: bigint, params: CreateChannelParams): Promise<PreparedTransaction[]> {
        let allTransactions: PreparedTransaction[] = [];

        try {
            const depositTxs = await this.prepareDepositTransactions(depositAmount);
            allTransactions = allTransactions.concat(depositTxs);
        } catch (err) {
            throw new Errors.ContractCallError("Failed to prepare deposit part of depositAndCreateChannel", err as Error);
        }

        try {
            const createChannelTx = await this.prepareCreateChannelTransaction(params);
            allTransactions.push(createChannelTx);
        } catch (err) {
            throw new Errors.ContractCallError("Failed to prepare createChannel part of depositAndCreateChannel", err as Error);
        }

        return allTransactions;
    }

    /**
     * Deposits funds and creates a channel by sending sequential transactions (Direct Execution).
     * Handles ERC20 approval if necessary for the deposit.
     * @param depositAmount The amount to deposit.
     * @param params Parameters for channel creation.
     * @returns An object containing the channel ID, initial state, and transaction hashes for deposit and creation.
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
     * Prepares the transaction data for checkpointing a state on-chain.
     * Requires the state to be signed by both participants.
     * Designed for use with Account Abstraction (UserOperations).
     * @param params Parameters for checkpointing the state.
     * @returns The prepared transaction data ({ to, data, value }).
     */
    async prepareCheckpointChannelTransaction(params: CheckpointChannelParams): Promise<PreparedTransaction> {
        const { channelId, candidateState, proofStates = [] } = params;

        if (!candidateState.sigs || candidateState.sigs.length < 2) {
            throw new Errors.InvalidParameterError("Candidate state for checkpoint must be signed by both participants.");
        }

        try {
            return await this.nitroliteService.prepareCheckpoint(channelId, candidateState, proofStates);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError("prepareCheckpointChannelTransaction", err as Error, { params });
        }
    }

    /**
     * Checkpoints a state on-chain.
     * Requires the state to be signed by both participants.
     * @param params Parameters for checkpointing the state.
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
     * Prepares the transaction data for challenging a channel on-chain.
     * Designed for use with Account Abstraction (UserOperations).
     * @param params Parameters for challenging the channel.
     * @returns The prepared transaction data ({ to, data, value }).
     */
    async prepareChallengeChannelTransaction(params: ChallengeChannelParams): Promise<PreparedTransaction> {
        const { channelId, candidateState, proofStates = [] } = params;

        try {
            return await this.nitroliteService.prepareChallenge(channelId, candidateState, proofStates);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError("prepareChallengeChannelTransaction", err as Error, { params });
        }
    }

    /**
     * Challenges a channel on-chain with a candidate state.
     * Used when the counterparty is unresponsive. Requires the candidate state to be signed by the challenger.
     * @param params Parameters for challenging the channel.
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
     * Prepares the transaction data for closing a channel collaboratively.
     * Handles internal state construction and signing.
     * Designed for use with Account Abstraction (UserOperations).
     * @param params Parameters for closing the channel.
     * @returns The prepared transaction data ({ to, data, value }).
     */
    async prepareCloseChannelTransaction(params: CloseChannelParams): Promise<PreparedTransaction> {
        try {
            const { finalStateWithSigs, channelId } = await _prepareAndSignFinalState(this, params);

            return await this.nitroliteService.prepareClose(channelId, finalStateWithSigs);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError("prepareCloseChannelTransaction", err as Error, { params });
        }
    }

    /**
     * Closes a channel on-chain using a mutually agreed final state.
     * Requires the final state signed by both participants.
     * @param params Parameters for closing the channel.
     * @returns The transaction hash.
     */
    async closeChannel(params: CloseChannelParams): Promise<Hash> {
        try {
            const { finalStateWithSigs, channelId } = await _prepareAndSignFinalState(this, params);

            return await this.nitroliteService.close(channelId, finalStateWithSigs);
        } catch (err) {
            throw new Errors.ContractError("Failed to execute closeChannel on contract", undefined, undefined, undefined, undefined, err as Error);
        }
    }

    /**
     * Prepares the transaction data for withdrawing deposited funds from the custody contract.
     * This does not withdraw funds locked in active channels.
     * Designed for use with Account Abstraction (UserOperations).
     * @param amount The amount of tokens/ETH to withdraw.
     * @returns The prepared transaction data ({ to, data, value }).
     */
    async prepareWithdrawalTransaction(amount: bigint): Promise<PreparedTransaction> {
        const tokenAddress = this.addresses.tokenAddress;

        try {
            return await this.nitroliteService.prepareWithdraw(tokenAddress, amount);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError("prepareWithdrawalTransaction", err as Error, { amount, tokenAddress });
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
            throw new Errors.ContractError("Failed to execute withdrawDeposit on contract", undefined, undefined, undefined, undefined, err as Error);
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
