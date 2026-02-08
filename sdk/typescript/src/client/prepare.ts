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
    _prepareAndSignResizeState,
} from './state';
import {
    ChallengeChannelParams,
    CheckpointChannelParams,
    CloseChannelParams,
    CreateChannelParams,
    ResizeChannelParams,
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
    challengeDuration: bigint;
    chainId: number;
}

/**
 * Handles the preparation of transaction data for various Nitrolite operations,
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
     * Prepares the transactions data necessary for a deposit operation,
     * including ERC20 approval if required.
     * @param tokenAddress The address of the token to deposit.
     * @param amount The amount of tokens/ETH to deposit.
     * @returns An array of PreparedTransaction objects (approve + deposit, or just deposit).
     */
    async prepareDepositTransactions(tokenAddress: Address, amount: bigint): Promise<PreparedTransaction[]> {
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
            const depositTx = await this.deps.nitroliteService.prepareDeposit(tokenAddress, amount);
            transactions.push(depositTx);
        } catch (err) {
            throw new Errors.ContractCallError('prepareDeposit', err as Error, { tokenAddress, amount });
        }

        return transactions;
    }

    /**
     * Prepares the transaction data for creating a new state channel.
     * Handles internal state construction and signing.
     * @param tokenAddress The address of the token for the channel.
     * @param params Parameters for channel creation. See {@link CreateChannelParams}.
     * @returns The prepared transaction data ({ to, data, value }).
     */
    async prepareCreateChannelTransaction(params: CreateChannelParams): Promise<PreparedTransaction> {
        try {
            const { initialState } = await _prepareAndSignInitialState(this.deps, params);

            return await this.deps.nitroliteService.prepareCreateChannel(params.channel, initialState);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareCreateChannelTransaction', err as Error, { params });
        }
    }

    /**
     * Prepares the transaction data for depositing funds and creating a channel in a single operation.
     * Includes potential ERC20 approval. Designed for batching.
     * @param tokenAddress The address of the token to deposit and use for the channel.
     * @param depositAmount The amount to deposit.
     * @param params Parameters for channel creation. See {@link CreateChannelParams}.
     * @returns An array of PreparedTransaction objects (approve?, deposit, createChannel).
     */
    async prepareDepositAndCreateChannelTransactions(
        tokenAddress: Address,
        amount: bigint,
        params: CreateChannelParams,
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
            const { initialState } = await _prepareAndSignInitialState(this.deps, params);
            const depositTx = await this.deps.nitroliteService.prepareDepositAndCreateChannel(
                tokenAddress,
                amount,
                params.channel,
                initialState,
            );
            transactions.push(depositTx);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareDepositAndCreateChannel', err as Error, {
                tokenAddress,
                amount,
            });
        }

        return transactions;
    }

    /**
     * Prepares the transaction data for checkpointing a state on-chain.
     * Requires the state to be signed by both participants.
     * @param params Parameters for checkpointing the state. See {@link CheckpointChannelParams}.
     * @returns The prepared transaction data ({ to, data, value }).
     */
    async prepareCheckpointChannelTransaction(params: CheckpointChannelParams): Promise<PreparedTransaction> {
        const { channelId, candidateState, proofStates = [] } = params;

        if (!candidateState.sigs || candidateState.sigs.length < 2) {
            throw new Errors.InvalidParameterError(
                'Candidate state for checkpoint must be signed by both participants.',
            );
        }

        try {
            return await this.deps.nitroliteService.prepareCheckpoint(channelId, candidateState, proofStates);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareCheckpointChannelTransaction', err as Error, { params });
        }
    }

    /**
     * Prepares the transaction data for challenging a channel on-chain.
     * @param params Parameters for challenging the channel. See {@link ChallengeChannelParams}.
     * @returns The prepared transaction data ({ to, data, value }).
     */
    async prepareChallengeChannelTransaction(params: ChallengeChannelParams): Promise<PreparedTransaction> {
        const { channelId, candidateState, proofStates = [] } = params;
        const { challengerSig } = await _prepareAndSignChallengeState(this.deps, params);

        try {
            return await this.deps.nitroliteService.prepareChallenge(
                channelId,
                candidateState,
                proofStates,
                challengerSig,
            );
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareChallengeChannelTransaction', err as Error, { params });
        }
    }

    /**
     * Prepares the transaction data for resize a channel on-chain.
     * @param params Parameters for resizing the channel. See {@link ResizeChannelParams}.
     * @returns The prepared transaction data ({ to, data, value }).
     */
    async prepareResizeChannelTransaction(params: ResizeChannelParams): Promise<PreparedTransaction> {
        const { resizeStateWithSigs, channelId } = await _prepareAndSignResizeState(this.deps, params);

        try {
            return await this.deps.nitroliteService.prepareResize(channelId, resizeStateWithSigs);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareResizeChannelTransaction', err as Error, { params });
        }
    }

    /**
     * Prepares the transaction data for closing a channel collaboratively.
     * Handles internal state construction and signing.
     * @param params Parameters for closing the channel. See {@link CloseChannelParams}.
     * @returns The prepared transaction data ({ to, data, value }).
     */
    async prepareCloseChannelTransaction(params: CloseChannelParams): Promise<PreparedTransaction> {
        try {
            const { finalStateWithSigs, channelId } = await _prepareAndSignFinalState(this.deps, params);

            return await this.deps.nitroliteService.prepareClose(channelId, finalStateWithSigs);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareCloseChannelTransaction', err as Error, { params });
        }
    }

    /**
     * Prepares the transaction data for withdrawing deposited funds from the custody contract.
     * This does not withdraw funds locked in active channels.
     * @param tokenAddress The address of the token to withdraw.
     * @param amount The amount of tokens/ETH to withdraw.
     * @returns The prepared transaction data ({ to, data, value }).
     */
    async prepareWithdrawalTransaction(tokenAddress: Address, amount: bigint): Promise<PreparedTransaction> {
        try {
            return await this.deps.nitroliteService.prepareWithdraw(tokenAddress, amount);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareWithdrawalTransaction', err as Error, { amount, tokenAddress });
        }
    }

    /**
     * Prepares the transaction data for approving the custody contract to spend ERC20 tokens.
     * @param tokenAddress The address of the ERC20 token to approve.
     * @param amount The amount to approve.
     * @returns The prepared transaction data ({ to, data, value }).
     */
    async prepareApproveTokensTransaction(tokenAddress: Address, amount: bigint): Promise<PreparedTransaction> {
        const spender = this.deps.addresses.custody;

        if (tokenAddress === zeroAddress) {
            throw new Errors.InvalidParameterError('Cannot prepare approval for ETH (zero address)');
        }
        try {
            return await this.deps.erc20Service.prepareApprove(tokenAddress, spender, amount);
        } catch (err) {
            if (err instanceof Errors.NitroliteError) throw err;
            throw new Errors.ContractCallError('prepareApproveTokensTransaction', err as Error, {
                amount,
                tokenAddress,
                spender,
            });
        }
    }
}
