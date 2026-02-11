import {
    Account,
    Address,
    Chain,
    ParseAccount,
    SimulateContractReturnType,
    Transport,
    WalletClient,
    zeroAddress,
} from 'viem';
import { ContractAddresses } from '../abis';
import * as Errors from '../errors';
import { Erc20Service, NitroliteService } from './services';
import {
    _prepareAndSignChallengeState,
    _prepareAndSignFinalState,
    _prepareAndSignInitialState,
} from './state';
import {
    ChallengeChannelParams,
    CheckpointChannelParams,
    CloseChannelParams,
    CreateChannelParams,
} from './types';
import { StateSigner } from './signer';

/**
 * Represents the data needed to construct a transaction or UserOperation call.
 * Derived from viem's SimulateContractReturnType['request'].
 */
export type PreparedTransaction = SimulateContractReturnType['request'];

/**
 * @dev Note: `stateSigner.signState` function should NOT add an EIP-191 prefix to the message signed as
 * the contract expects the raw message to be signed.
 */
export interface PreparerDependencies {
    nitroliteService: NitroliteService;
    erc20Service: Erc20Service;
    addresses: ContractAddresses;
    account: ParseAccount<Account>;
    walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;
    stateSigner: StateSigner;
    challengeDuration: number;
    chainId: number;
}

/**
 * Handles the preparation of transaction data for various Nitrolite V1 operations,
 * suitable for use with Account Abstraction (UserOperations) or manual transaction sending.
 * It simulates transactions but does not execute them.
 */
export class NitroliteTransactionPreparer {
    private readonly deps: PreparerDependencies;

    /**
     * Creates an instance of NitroliteTransactionPreparer.
     * @param dependencies - The services and configuration needed for preparation. See {@link PreparerDependencies}.
     */
    constructor(dependencies: PreparerDependencies) {
        this.deps = dependencies;
    }

    /**
     * Prepares the transactions data necessary for a deposit operation (to vault),
     * including ERC20 approval if required.
     * @param node The address of the node.
     * @param tokenAddress The address of the token to deposit.
     * @param amount The amount of tokens/ETH to deposit.
     * @returns An array of PreparedTransaction objects (approve + deposit, or just deposit).
     */
    async prepareDepositTransactions(
        node: Address,
        tokenAddress: Address,
        amount: bigint,
    ): Promise<PreparedTransaction[]> {
        const transactions: PreparedTransaction[] = [];
        const spender = this.deps.addresses.custody;
        const owner = this.deps.account.address;

        if (tokenAddress !== zeroAddress) {
            const allowance = await this.deps.erc20Service.getTokenAllowance(tokenAddress, owner, spender);
            if (allowance < amount) {
                try {
                    const approveTx = await this.deps.erc20Service.prepareApprove(tokenAddress, spender, amount);
                    transactions.push(approveTx);
                } catch (err) {
                    throw new Errors.ContractCallError('prepareApprove (for deposit)', err as Error, {
                        tokenAddress,
                        spender,
                        amount,
                    });
                }
            }
        }

        try {
            const depositTx = await this.deps.nitroliteService.prepareDeposit(node, tokenAddress, amount);
            transactions.push(depositTx);
        } catch (err) {
            throw new Errors.ContractCallError('prepareDeposit', err as Error, { node, tokenAddress, amount });
        }

        return transactions;
    }

    /**
     * Prepares the transaction data for creating a new state channel.
     * Handles internal state construction and signing.
     * @param params Parameters for channel creation. See {@link CreateChannelParams}.
     * @returns The prepared transaction data ({ to, data, value }).
     */
    async prepareCreateChannelTransaction(params: CreateChannelParams): Promise<PreparedTransaction> {
        try {
            const { initialState } = await _prepareAndSignInitialState(this.deps, params);

            return await this.deps.nitroliteService.prepareCreateChannel(params.definition, initialState);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareCreateChannelTransaction', err as Error, { params });
        }
    }

    /**
     * Prepares the transaction data for checkpointing a state on-chain.
     * @param params Parameters for checkpointing. See {@link CheckpointChannelParams}.
     * @returns The prepared transaction data.
     */
    async prepareCheckpointChannelTransaction(params: CheckpointChannelParams): Promise<PreparedTransaction> {
        const { channelId, candidateState, proofs = [] } = params;

        try {
            return await this.deps.nitroliteService.prepareCheckpointChannel(channelId, candidateState, proofs);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareCheckpointChannelTransaction', err as Error, { params });
        }
    }

    /**
     * Prepares the transaction data for challenging a channel.
     * @param params Parameters for challenging. See {@link ChallengeChannelParams}.
     * @returns The prepared transaction data.
     */
    async prepareChallengeChannelTransaction(params: ChallengeChannelParams): Promise<PreparedTransaction> {
        try {
            const { channelId, candidateState, proofs, challengerSig } = await _prepareAndSignChallengeState(
                this.deps,
                params,
            );

            return await this.deps.nitroliteService.prepareChallengeChannel(
                channelId,
                candidateState,
                proofs,
                challengerSig,
            );
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareChallengeChannelTransaction', err as Error, { params });
        }
    }

    /**
     * Prepares the transaction data for closing a channel.
     * @param params Parameters for closing. See {@link CloseChannelParams}.
     * @returns The prepared transaction data.
     */
    async prepareCloseChannelTransaction(params: CloseChannelParams): Promise<PreparedTransaction> {
        try {
            const { channelId, finalState, proofs } = await _prepareAndSignFinalState(this.deps, params);

            return await this.deps.nitroliteService.prepareCloseChannel(channelId, finalState, proofs || []);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareCloseChannelTransaction', err as Error, { params });
        }
    }

    /**
     * Prepares the transaction data for withdrawing from the vault.
     * @param to Recipient address.
     * @param tokenAddress The address of the token.
     * @param amount The amount to withdraw.
     * @returns The prepared transaction data.
     */
    async prepareWithdrawTransaction(
        to: Address,
        tokenAddress: Address,
        amount: bigint,
    ): Promise<PreparedTransaction> {
        try {
            return await this.deps.nitroliteService.prepareWithdraw(to, tokenAddress, amount);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareWithdrawTransaction', err as Error, {
                to,
                tokenAddress,
                amount,
            });
        }
    }

    /**
     * Prepares the transactions for depositing and creating a channel in one flow.
     * Includes potential ERC20 approval. Designed for batching.
     * @param node The address of the node.
     * @param tokenAddress The address of the token to deposit.
     * @param depositAmount The amount to deposit.
     * @param params Parameters for channel creation. See {@link CreateChannelParams}.
     * @returns An array of PreparedTransaction objects (approve?, deposit, createChannel).
     */
    async prepareDepositAndCreateChannelTransactions(
        node: Address,
        tokenAddress: Address,
        depositAmount: bigint,
        params: CreateChannelParams,
    ): Promise<PreparedTransaction[]> {
        const transactions: PreparedTransaction[] = [];
        const spender = this.deps.addresses.custody;
        const owner = this.deps.account.address;

        // Add approval if needed
        if (tokenAddress !== zeroAddress) {
            const allowance = await this.deps.erc20Service.getTokenAllowance(tokenAddress, owner, spender);
            if (allowance < depositAmount) {
                try {
                    const approveTx = await this.deps.erc20Service.prepareApprove(tokenAddress, spender, depositAmount);
                    transactions.push(approveTx);
                } catch (err) {
                    throw new Errors.ContractCallError('prepareApprove (for deposit)', err as Error, {
                        tokenAddress,
                        spender,
                        depositAmount,
                    });
                }
            }
        }

        // Add deposit transaction
        try {
            const depositTx = await this.deps.nitroliteService.prepareDeposit(node, tokenAddress, depositAmount);
            transactions.push(depositTx);
        } catch (err) {
            throw new Errors.ContractCallError('prepareDeposit', err as Error, { node, tokenAddress, depositAmount });
        }

        // Add create channel transaction
        try {
            const { initialState } = await _prepareAndSignInitialState(this.deps, params);
            const createChannelTx = await this.deps.nitroliteService.prepareCreateChannel(params.definition, initialState);
            transactions.push(createChannelTx);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareCreateChannel', err as Error, { params });
        }

        return transactions;
    }
}
