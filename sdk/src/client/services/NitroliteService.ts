import { Account, Address, PublicClient, WalletClient, Hash, zeroAddress, Hex, ContractFunctionParameters } from 'viem';
import { custodyAbi } from '../../abis/generated';
import { ContractAddresses } from '../../abis';
import { Errors } from '../../errors';
import { Channel, ChannelData, ChannelId, Signature, State } from '../types';
import { ContractCallParams, ContractWriter } from '../contract_writer/types';
import { EOAContractWriter } from '../contract_writer/eoa';

/**
 * Type utility to properly type the request object from simulateContract
 * This ensures type safety when passing the request to writeContract
 *
 * The SimulateContractReturnType['request'] contains all necessary parameters
 * for writeContract, but viem's complex union types make direct compatibility challenging.
 * We use a more practical approach with proper type comments explaining the safety.
 */
type PreparedContractRequest = any;

/**
 * Type-safe wrapper for writeContract calls using prepared requests.
 * This function handles the type compatibility between simulateContract result and writeContract params.
 *
 * @param walletClient - The wallet client to use for writing
 * @param request - The prepared request from simulateContract
 * @param account - The account to use for the transaction
 * @returns Promise<Hash> - The transaction hash
 */
const executeWriteContract = async (
    contractWriter: ContractWriter,
    request: PreparedContractRequest,
    account: Account | Address,
): Promise<Hash> => {
    // The request from simulateContract contains all required parameters for writeContract.
    // We safely spread the request and add the account. This is type-safe because:
    // 1. simulateContract validates the contract call against the ABI
    // 2. The returned request contains the exact parameters needed by writeContract
    // 3. We only add the account parameter which is required by writeContract
    //
    // Note: Type assertion is necessary due to viem's complex union types for transaction parameters.
    // The runtime behavior is correct - simulateContract returns compatible parameters for writeContract.
    const calls = [
        {
            ...request,
            account,
        },
    ];

    const result = await contractWriter.write({
        calls,
    });

    if (result.txHashes.length < 1) {
        throw new Error('No transaction hashes returned from write operation');
    }

    return result.txHashes[result.txHashes.length - 1];
};

/**
 * Service for interacting directly with the Nitrolite Custody smart contract.
 * Provides methods for channel management, deposits, and withdrawals specific to the Custody contract.
 */
export class NitroliteService {
    private readonly publicClient: PublicClient;
    private readonly account?: Account | Address;
    private readonly addresses: ContractAddresses;
    private readonly contractWriter?: ContractWriter;

    constructor(
        publicClient: PublicClient,
        addresses: ContractAddresses,
        walletClient?: WalletClient,
        account?: Account | Address,
        contractWriter?: ContractWriter,
    ) {
        if (!publicClient) {
            throw new Errors.MissingParameterError('publicClient');
        }

        if (!addresses || !addresses.custody) {
            throw new Errors.MissingParameterError('addresses.custody');
        }

        if (contractWriter) {
            this.contractWriter = contractWriter;
        } else if (walletClient) {
            this.contractWriter = new EOAContractWriter({
                publicClient,
                // @ts-ignore
                walletClient,
            });
        }

        this.publicClient = publicClient;
        this.account = account || walletClient?.account;
        this.addresses = addresses;
    }

    /** Ensures a ContractWriter is available for write operations. */
    private ensureContractWriter(): ContractWriter {
        if (!this.contractWriter) {
            throw new Errors.ContractWriterRequiredError();
        }
        return this.contractWriter;
    }

    /** Ensures an Account is available for write/simulation operations. */
    private ensureAccount(): Account | Address {
        if (!this.account) {
            throw new Errors.AccountRequiredError();
        }
        return this.account;
    }

    /** Get the custody contract address. */
    get custodyAddress(): Address {
        return this.addresses.custody;
    }

    /**
     * Converts Channel type to format expected by generated ABI
     * REQUIRED: participants array must be readonly for ABI compatibility
     */
    private convertChannelForABI(channel: Channel) {
        return {
            participants: (channel.participants || []) as readonly Address[],
            adjudicator: channel.adjudicator,
            challenge: channel.challenge,
            nonce: channel.nonce,
        } as const;
    }

    /**
     * Converts State type to format expected by generated ABI
     * REQUIRED:
     * - StateIntent enum -> number conversion
     * - Mutable arrays -> readonly arrays
     * - Proper type constraints for viem compatibility
     */
    private convertStateForABI(state: State) {
        return {
            intent: state.intent as number, // StateIntent enum maps to uint8
            version: state.version,
            data: state.data,
            allocations: (state.allocations || []).map((alloc) => ({
                destination: alloc.destination,
                token: alloc.token,
                amount: alloc.amount,
            })) as readonly {
                destination: Address;
                token: Address;
                amount: bigint;
            }[],
            sigs: state.sigs || ([] as readonly Hex[]),
        } as const;
    }

    /**
     * Converts contract Channel result to SDK Channel type
     */
    private convertChannelFromContract(contractChannel: any): Channel {
        return {
            participants: [...contractChannel.participants],
            adjudicator: contractChannel.adjudicator,
            challenge: contractChannel.challenge,
            nonce: contractChannel.nonce,
        };
    }

    /**
     * Converts contract State result to SDK State type
     */
    private convertStateFromContract(contractState: any): State {
        return {
            intent: contractState.intent,
            version: contractState.version,
            data: contractState.data,
            allocations: contractState.allocations.map((alloc: any) => ({
                destination: alloc.destination,
                token: alloc.token,
                amount: alloc.amount,
            })),
            sigs: contractState.sigs,
        };
    }

    /**
     * Prepares contract call parameters for a deposit operation.
     * Returns parameters that can be used with ContractWriter for batching operations.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to deposit.
     * @returns Contract call parameters ready for execution.
     */
    prepareDepositCallParams(tokenAddress: Address, amount: bigint): ContractCallParams {
        const account = this.ensureAccount();
        const accountAddress = typeof account === 'string' ? account : account.address;

        return {
            address: this.custodyAddress,
            abi: custodyAbi,
            functionName: 'deposit',
            args: [accountAddress, tokenAddress, amount],
            account: account,
            value: tokenAddress === zeroAddress ? amount : 0n,
        };
    }

    /**
     * Prepares the request data for a deposit transaction.
     * Useful for batching multiple calls in a single UserOperation.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to deposit.
     * @returns The prepared transaction request object.
     */
    async prepareDeposit(tokenAddress: Address, amount: bigint): Promise<PreparedContractRequest> {
        const operationName = 'prepareDeposit';

        try {
            const params = this.prepareDepositCallParams(tokenAddress, amount);
            const { request } = await this.publicClient.simulateContract(params);

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { tokenAddress, amount });
        }
    }

    /**
     * Deposit tokens or ETH into the custody contract.
     * This method simulates and executes the transaction directly.
     * You do not need to call `prepareDeposit` separately unless batching operations.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to deposit.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async deposit(tokenAddress: Address, amount: bigint): Promise<Hash> {
        const contractWriter = this.ensureContractWriter();
        const account = this.ensureAccount();
        const operationName = 'deposit';

        try {
            const request = await this.prepareDeposit(tokenAddress, amount);
            return await executeWriteContract(contractWriter, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { tokenAddress, amount });
        }
    }

    /**
     * Prepares contract call parameters for creating a new channel.
     * Returns parameters that can be used with ContractWriter for batching operations.
     * @param channel Channel configuration. See {@link Channel} for details.
     * @param initial Initial state. See {@link State} for details.
     * @returns Contract call parameters ready for execution.
     */
    prepareCreateChannelCallParams(channel: Channel, initial: State): ContractCallParams {
        const account = this.ensureAccount();
        const abiChannel = this.convertChannelForABI(channel);
        const abiState = this.convertStateForABI(initial);

        return {
            address: this.custodyAddress,
            abi: custodyAbi,
            functionName: 'create',
            args: [abiChannel, abiState],
            account: account,
        };
    }

    /**
     * Prepares the request data for creating a new channel.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channel Channel configuration. See {@link Channel} for details.
     * @param initial Initial state. See {@link State} for details.
     * @returns The prepared transaction request object.
     */
    async prepareCreateChannel(channel: Channel, initial: State): Promise<PreparedContractRequest> {
        const operationName = 'prepareCreateChannel';

        try {
            const params = this.prepareCreateChannelCallParams(channel, initial);
            const { request } = await this.publicClient.simulateContract(params);

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { channel, initial });
        }
    }

    /**
     * Create a new channel.
     * This method simulates and executes the transaction directly.
     * You do not need to call `prepareCreateChannel` separately unless batching operations.
     * @param channel Channel configuration. See {@link Channel} for details.
     * @param initial Initial state. See {@link State} for details.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async createChannel(channel: Channel, initial: State): Promise<Hash> {
        const contractWriter = this.ensureContractWriter();
        const account = this.ensureAccount();
        const operationName = 'createChannel';

        try {
            const request = await this.prepareCreateChannel(channel, initial);
            return await executeWriteContract(contractWriter, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channel, initial });
        }
    }

    /**
     * Prepares contract call parameters for depositing funds and creating a channel in one operation.
     * Returns parameters that can be used with ContractWriter for batching operations.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to deposit.
     * @param channel Channel configuration. See {@link Channel} for details.
     * @param initial Initial state. See {@link State} for details.
     * @returns Contract call parameters ready for execution.
     */
    prepareDepositAndCreateChannelCallParams(
        tokenAddress: Address,
        amount: bigint,
        channel: Channel,
        initial: State,
    ): ContractCallParams {
        const account = this.ensureAccount();
        const abiChannel = this.convertChannelForABI(channel);
        const abiState = this.convertStateForABI(initial);

        return {
            address: this.custodyAddress,
            abi: custodyAbi,
            functionName: 'depositAndCreate',
            args: [tokenAddress, amount, abiChannel, abiState],
            account: account,
            value: tokenAddress === zeroAddress ? amount : 0n,
        };
    }

    /**
     * Prepares the request data for depositing funds and creating a new channel in one operation.
     * Useful for batching multiple calls in a single UserOperation.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to deposit.
     * @param channel Channel configuration. See {@link Channel} for details.
     * @param initial Initial state. See {@link State} for details.
     * @returns The prepared transaction request object.
     */
    async prepareDepositAndCreateChannel(
        tokenAddress: Address,
        amount: bigint,
        channel: Channel,
        initial: State,
    ): Promise<PreparedContractRequest> {
        const operationName = 'prepareDepositAndCreateChannel';

        try {
            const params = this.prepareDepositAndCreateChannelCallParams(tokenAddress, amount, channel, initial);
            const { request } = await this.publicClient.simulateContract(params);

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { tokenAddress, amount, channel, initial });
        }
    }

    /**
     * Deposits tokens or ETH and creates a new channel in one operation.
     * This method simulates and executes the transaction directly.
     * You do not need to call `prepareDepositAndCreateChannel` separately unless batching operations.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to deposit.
     * @param channel Channel configuration. See {@link Channel} for details.
     * @param initial Initial state. See {@link State} for details.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async depositAndCreateChannel(
        tokenAddress: Address,
        amount: bigint,
        channel: Channel,
        initial: State,
    ): Promise<Hash> {
        const contractWriter = this.ensureContractWriter();
        const account = this.ensureAccount();
        const operationName = 'depositAndCreateChannel';

        try {
            const request = await this.prepareDepositAndCreateChannel(tokenAddress, amount, channel, initial);
            return await executeWriteContract(contractWriter, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { tokenAddress, amount, channel, initial });
        }
    }

    /**
     * Prepares contract call parameters for joining an existing channel.
     * Returns parameters that can be used with ContractWriter for batching operations.
     * @param channelId ID of the channel.
     * @param index Participant index.
     * @param sig Participant signature.
     * @returns Contract call parameters ready for execution.
     */
    prepareJoinChannelCallParams(channelId: ChannelId, index: bigint, sig: Signature): ContractCallParams {
        const account = this.ensureAccount();

        return {
            address: this.custodyAddress,
            abi: custodyAbi,
            functionName: 'join',
            args: [channelId, index, sig],
            account: account,
        };
    }

    /**
     * Prepares the request data for joining an existing channel.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channelId ID of the channel.
     * @param index Participant index.
     * @param sig Participant signature.
     * @returns The prepared transaction request object.
     */
    async prepareJoinChannel(channelId: ChannelId, index: bigint, sig: Signature): Promise<PreparedContractRequest> {
        const operationName = 'prepareJoinChannel';

        try {
            const params = this.prepareJoinChannelCallParams(channelId, index, sig);
            const { request } = await this.publicClient.simulateContract(params);

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { channelId, index });
        }
    }

    /**
     * Join an existing channel.
     * This method simulates and executes the transaction directly.
     * You do not need to call `prepareJoinChannel` separately unless batching operations.
     * @param channelId ID of the channel.
     * @param index Participant index.
     * @param sig Participant signature.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async joinChannel(channelId: ChannelId, index: bigint, sig: Signature): Promise<Hash> {
        const contractWriter = this.ensureContractWriter();
        const account = this.ensureAccount();
        const operationName = 'joinChannel';

        try {
            const request = await this.prepareJoinChannel(channelId, index, sig);
            return await executeWriteContract(contractWriter, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId, index });
        }
    }

    /**
     * Prepares contract call parameters for checkpointing a state.
     * Returns parameters that can be used with ContractWriter for batching operations.
     * @param channelId Channel ID. See {@link ChannelId} for details.
     * @param candidate State to checkpoint. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @returns Contract call parameters ready for execution.
     */
    prepareCheckpointCallParams(channelId: ChannelId, candidate: State, proofs: State[] = []): ContractCallParams {
        const account = this.ensureAccount();
        const abiCandidate = this.convertStateForABI(candidate);
        const abiProofs = proofs.map((proof) => this.convertStateForABI(proof));

        return {
            address: this.custodyAddress,
            abi: custodyAbi,
            functionName: 'checkpoint',
            args: [channelId, abiCandidate, abiProofs],
            account: account,
        };
    }

    /**
     * Prepares the request data for checkpointing a state.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channelId Channel ID. See {@link ChannelId} for details.
     * @param candidate State to checkpoint. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @returns The prepared transaction request object.
     */
    async prepareCheckpoint(
        channelId: ChannelId,
        candidate: State,
        proofs: State[] = [],
    ): Promise<PreparedContractRequest> {
        const operationName = 'prepareCheckpoint';

        try {
            const params = this.prepareCheckpointCallParams(channelId, candidate, proofs);
            const { request } = await this.publicClient.simulateContract(params);

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { channelId });
        }
    }

    /**
     * Checkpoint a state on-chain.
     * This method simulates and executes the transaction directly.
     * You do not need to call `prepareCheckpoint` separately unless batching operations.
     * @param channelId Channel ID.
     * @param candidate State to checkpoint. See {@link State} for details.
     * @param proofs Supporting proofs if required by adjudicator. See {@link State} for details.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async checkpoint(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<Hash> {
        const contractWriter = this.ensureContractWriter();
        const account = this.ensureAccount();
        const operationName = 'checkpoint';

        try {
            const request = await this.prepareCheckpoint(channelId, candidate, proofs);
            return await executeWriteContract(contractWriter, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId });
        }
    }

    /**
     * Prepares contract call parameters for challenging a state.
     * Returns parameters that can be used with ContractWriter for batching operations.
     * @param channelId Channel ID.
     * @param candidate State being challenged. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @param challengerSig Challenger signature. See {@link Signature} for details.
     * @returns Contract call parameters ready for execution.
     */
    prepareChallengeCallParams(
        channelId: ChannelId,
        candidate: State,
        proofs: State[] = [],
        challengerSig: Signature,
    ): ContractCallParams {
        const account = this.ensureAccount();
        const abiCandidate = this.convertStateForABI(candidate);
        const abiProofs = proofs.map((proof) => this.convertStateForABI(proof));

        return {
            address: this.custodyAddress,
            abi: custodyAbi,
            functionName: 'challenge',
            args: [channelId, abiCandidate, abiProofs, challengerSig],
            account: account,
        };
    }

    /**
     * Prepares the request data for challenging a state.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channelId Channel ID.
     * @param candidate State being challenged. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @param challengerSig Challenger signature. See {@link Signature} for details.
     * @returns The prepared transaction request object.
     */
    async prepareChallenge(
        channelId: ChannelId,
        candidate: State,
        proofs: State[] = [],
        challengerSig: Signature,
    ): Promise<PreparedContractRequest> {
        const operationName = 'prepareChallenge';

        try {
            const params = this.prepareChallengeCallParams(channelId, candidate, proofs, challengerSig);
            const { request } = await this.publicClient.simulateContract(params);

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { channelId });
        }
    }

    /**
     * Challenge a state on-chain.
     * This method simulates and executes the transaction directly.
     * You do not need to call `prepareChallenge` separately unless batching operations.
     * @param channelId Channel ID.
     * @param candidate State being challenged. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async challenge(
        channelId: ChannelId,
        candidate: State,
        proofs: State[] = [],
        challengerSig: Signature,
    ): Promise<Hash> {
        const contractWriter = this.ensureContractWriter();
        const account = this.ensureAccount();
        const operationName = 'challenge';

        try {
            const request = await this.prepareChallenge(channelId, candidate, proofs, challengerSig);
            return await executeWriteContract(contractWriter, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId });
        }
    }

    /**
     * Prepares contract call parameters for resizing a channel.
     * Returns parameters that can be used with ContractWriter for batching operations.
     * @param channelId Channel ID.
     * @param candidate Candidate state for the resizing channel. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @returns Contract call parameters ready for execution.
     */
    prepareResizeCallParams(channelId: ChannelId, candidate: State, proofs: State[] = []): ContractCallParams {
        const account = this.ensureAccount();
        const abiCandidate = this.convertStateForABI(candidate);
        const abiProofs = proofs.map((proof) => this.convertStateForABI(proof));

        return {
            address: this.custodyAddress,
            abi: custodyAbi,
            functionName: 'resize',
            args: [channelId, abiCandidate, abiProofs],
            account: account,
        };
    }

    /**
     * Prepares the request data for resize a channel.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channelId Channel ID.
     * @param candidate Candidate state for the resizing channel. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @returns The prepared transaction request object.
     */
    async prepareResize(
        channelId: ChannelId,
        candidate: State,
        proofs: State[] = [],
    ): Promise<PreparedContractRequest> {
        const operationName = 'prepareResize';

        try {
            const params = this.prepareResizeCallParams(channelId, candidate, proofs);
            const { request } = await this.publicClient.simulateContract(params);

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { channelId });
        }
    }

    /**
     * Resize a channel.
     * This method simulates and executes the transaction directly.
     * You do not need to call `prepareResize` separately unless batching operations.
     * @param channelId Channel ID.
     * @param candidate Candidate state for the resizing channel. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async resize(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<Hash> {
        const contractWriter = this.ensureContractWriter();
        const account = this.ensureAccount();
        const operationName = 'resize';

        try {
            const request = await this.prepareResize(channelId, candidate, proofs);
            return await executeWriteContract(contractWriter, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId });
        }
    }

    /**
     * Prepares contract call parameters for closing a channel.
     * Returns parameters that can be used with ContractWriter for batching operations.
     * @param channelId Channel ID.
     * @param candidate Final state. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @returns Contract call parameters ready for execution.
     */
    prepareCloseCallParams(channelId: ChannelId, candidate: State, proofs: State[] = []): ContractCallParams {
        const account = this.ensureAccount();
        const abiCandidate = this.convertStateForABI(candidate);
        const abiProofs = proofs.map((proof) => this.convertStateForABI(proof));

        return {
            address: this.custodyAddress,
            abi: custodyAbi,
            functionName: 'close',
            args: [channelId, abiCandidate, abiProofs],
            account: account,
        };
    }

    /**
     * Prepares the request data for closing a channel.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channelId Channel ID.
     * @param candidate Final state. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @returns The prepared transaction request object.
     */
    async prepareClose(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<PreparedContractRequest> {
        const operationName = 'prepareClose';

        try {
            const params = this.prepareCloseCallParams(channelId, candidate, proofs);
            const { request } = await this.publicClient.simulateContract(params);

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { channelId });
        }
    }

    /**
     * Close a channel cooperatively or after challenge expiry.
     * This method simulates and executes the transaction directly.
     * You do not need to call `prepareClose` separately unless batching operations.
     * @param channelId Channel ID.
     * @param candidate Final state. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async close(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<Hash> {
        const contractWriter = this.ensureContractWriter();
        const account = this.ensureAccount();
        const operationName = 'close';

        try {
            const request = await this.prepareClose(channelId, candidate, proofs);
            return await executeWriteContract(contractWriter, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId });
        }
    }

    /**
     * Prepares contract call parameters for withdrawing funds.
     * Returns parameters that can be used with ContractWriter for batching operations.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to withdraw.
     * @returns Contract call parameters ready for execution.
     */
    prepareWithdrawCallParams(tokenAddress: Address, amount: bigint): ContractCallParams {
        const account = this.ensureAccount();

        return {
            address: this.custodyAddress,
            abi: custodyAbi,
            functionName: 'withdraw',
            args: [tokenAddress, amount],
            account: account,
        };
    }

    /**
     * Prepares the request data for withdrawing funds.
     * Useful for batching multiple calls in a single UserOperation.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to withdraw.
     * @returns The prepared transaction request object.
     */
    async prepareWithdraw(tokenAddress: Address, amount: bigint): Promise<PreparedContractRequest> {
        const operationName = 'prepareWithdraw';

        try {
            const params = this.prepareWithdrawCallParams(tokenAddress, amount);
            const { request } = await this.publicClient.simulateContract(params);

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { tokenAddress, amount });
        }
    }

    /**
     * Withdraw available funds from the custody contract.
     * This method simulates and executes the transaction directly.
     * You do not need to call `prepareWithdraw` separately unless batching operations.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to withdraw.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async withdraw(tokenAddress: Address, amount: bigint): Promise<Hash> {
        const contractWriter = this.ensureContractWriter();
        const account = this.ensureAccount();
        const operationName = 'withdraw';

        try {
            const request = await this.prepareWithdraw(tokenAddress, amount);
            return await executeWriteContract(contractWriter, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { tokenAddress, amount });
        }
    }

    /**
     * Get the list of open channels for specified accounts
     * @param account Address or addresses of the accounts
     * @returns Matrix of Channel IDs, where each sub-array corresponds to an account
     * @error Throws ContractReadError if the read operation fails
     */
    async getOpenChannels(account: Address): Promise<ChannelId[]>;
    async getOpenChannels(account: Address[]): Promise<ChannelId[][]>;
    async getOpenChannels(account: Address | Address[]): Promise<ChannelId[] | ChannelId[][]> {
        const functionName = 'getOpenChannels';

        const accountsArg = Array.isArray(account) ? account : [account];
        if (accountsArg.length === 0) {
            throw new Errors.MissingParameterError('accounts');
        }

        try {
            const result = await this.publicClient.readContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: functionName,
                args: [accountsArg],
            });

            if (Array.isArray(account)) {
                return result as ChannelId[][];
            } else {
                return result[0] as ChannelId[];
            }
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractReadError(functionName, error, { accountsArg });
        }
    }

    /**
     * Get balances for specified accounts and tokens
     * @param user Address or addresses of the accounts
     * @param token Address or addresses of the tokens (use zeroAddress for ETH)
     * @returns Matrix of balances, where each sub-array corresponds to an account and token
     * @error Throws ContractReadError if the read operation fails
     */
    async getAccountBalance(user: Address, token: Address): Promise<bigint>;
    async getAccountBalance(user: Address, token: Address[]): Promise<bigint[]>;
    async getAccountBalance(user: Address[], token: Address): Promise<bigint[]>;
    async getAccountBalance(user: Address[], token: Address[]): Promise<bigint[][]>;
    async getAccountBalance(
        user: Address | Address[],
        token: Address | Address[],
    ): Promise<bigint | bigint[] | bigint[][]> {
        const functionName = 'getAccountsBalances';

        const usersArg = Array.isArray(user) ? user : [user];
        const tokensArg = Array.isArray(token) ? token : [token];
        if (usersArg.length === 0 || tokensArg.length === 0) {
            throw new Errors.MissingParameterError('users or tokens');
        }

        try {
            const result = await this.publicClient.readContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: functionName,
                args: [usersArg, tokensArg],
            });

            if (Array.isArray(token)) {
                if (Array.isArray(user)) {
                    return result as bigint[][];
                } else {
                    return result[0] as bigint[];
                }
            } else {
                if (Array.isArray(user)) {
                    return result[0] as bigint[];
                } else {
                    return result[0][0] as bigint;
                }
            }
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractReadError(functionName, error, { usersArg, tokensArg });
        }
    }

    /**
     * Get the balances for a specific channel and token or tokens.
     * @param channelId ID of the channel to retrieve balances for.
     * @param token Address or addresses of the tokens (use zeroAddress for ETH).
     * @returns Array of balances for the specified tokens.
     * @error Throws ContractReadError if the read operation fails.
     */
    async getChannelBalance(channelId: ChannelId, token: Address): Promise<bigint>;
    async getChannelBalance(channelId: ChannelId, token: Address[]): Promise<bigint[]>;
    async getChannelBalance(channelId: ChannelId, token: Address | Address[]): Promise<bigint | bigint[]> {
        const functionName = 'getChannelBalances';

        const tokensArg = Array.isArray(token) ? token : [token];
        if (tokensArg.length === 0) {
            throw new Errors.MissingParameterError('tokens');
        }

        try {
            const result = await this.publicClient.readContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: functionName,
                args: [channelId, tokensArg],
            });

            if (Array.isArray(token)) {
                return result as bigint[];
            } else {
                return result[0] as bigint;
            }
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractReadError(functionName, error, { channelId, tokensArg });
        }
    }

    /**
     * Get channel data for a specific channel ID.
     * @param channelId ID of the channel to retrieve data for.
     * @returns ChannelData object containing participants and adjudicator address.
     * @error Throws ContractReadError if the read operation fails.
     */
    async getChannelData(channelId: ChannelId): Promise<ChannelData> {
        const functionName = 'getChannelData';

        try {
            const result = await this.publicClient.readContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: functionName,
                args: [channelId],
            });

            return {
                channel: this.convertChannelFromContract(result[0]),
                status: result[1],
                wallets: result[2] as [Address, Address],
                challengeExpiry: result[3],
                lastValidState: this.convertStateFromContract(result[4]),
            };
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractReadError(functionName, error, { channelId });
        }
    }
}
