import {
    Account,
    Address,
    PublicClient,
    WalletClient,
    Hash,
    zeroAddress
} from 'viem';
import { custodyAbi } from '../../abis/generated';
import { ContractAddresses } from '../../abis';
import { Errors } from '../../errors';
import { Channel, ChannelId, Signature, State } from '../types';

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
    walletClient: WalletClient,
    request: PreparedContractRequest,
    account: Account | Address
): Promise<Hash> => {
    // The request from simulateContract contains all required parameters for writeContract.
    // We safely spread the request and add the account. This is type-safe because:
    // 1. simulateContract validates the contract call against the ABI
    // 2. The returned request contains the exact parameters needed by writeContract
    // 3. We only add the account parameter which is required by writeContract
    //
    // Note: Type assertion is necessary due to viem's complex union types for transaction parameters.
    // The runtime behavior is correct - simulateContract returns compatible parameters for writeContract.
    return walletClient.writeContract({
        ...request,
        account,
    } as any);
};

/**
 * Service for interacting directly with the Nitrolite Custody smart contract.
 * Provides methods for channel management, deposits, and withdrawals specific to the Custody contract.
 */
export class NitroliteService {
    private readonly publicClient: PublicClient;
    private readonly walletClient?: WalletClient;
    private readonly account?: Account | Address;
    private readonly addresses: ContractAddresses;

    constructor(
        publicClient: PublicClient,
        addresses: ContractAddresses,
        walletClient?: WalletClient,
        account?: Account | Address,
    ) {
        if (!publicClient) {
            throw new Errors.MissingParameterError('publicClient');
        }

        if (!addresses || !addresses.custody) {
            throw new Errors.MissingParameterError('addresses.custody');
        }

        this.publicClient = publicClient;
        this.walletClient = walletClient;
        this.account = account || walletClient?.account;
        this.addresses = addresses;
    }

    /** Ensures a WalletClient is available for write operations. */
    private ensureWalletClient(): WalletClient {
        if (!this.walletClient) {
            throw new Errors.WalletClientRequiredError();
        }
        return this.walletClient;
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
            sigs: (state.sigs || []).map((sig) => ({
                v: sig.v,
                r: sig.r,
                s: sig.s,
            })) as readonly {
                v: number;
                r: `0x${string}`;
                s: `0x${string}`;
            }[],
        } as const;
    }

    /**
     * Converts Signature type to format expected by generated ABI
     * REQUIRED: Ensures proper readonly typing for viem compatibility
     */
    private convertSignatureForABI(signature: Signature) {
        return {
            v: signature.v,
            r: signature.r,
            s: signature.s,
        } as const;
    }

    /**
     * Prepares the request data for a deposit transaction.
     * Useful for batching multiple calls in a single UserOperation.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to deposit.
     * @returns The prepared transaction request object.
     */
    async prepareDeposit(tokenAddress: Address, amount: bigint): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareDeposit';
        const accountAddress = typeof account === 'string' ? account : account.address;

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'deposit',
                args: [accountAddress, tokenAddress, amount],
                account: account,
                value: tokenAddress === zeroAddress ? amount : 0n,
            });

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
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'deposit';

        try {
            const request = await this.prepareDeposit(tokenAddress, amount);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { tokenAddress, amount });
        }
    }

    /**
     * Prepares the request data for creating a new channel.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channel Channel configuration. See {@link Channel} for details.
     * @param initial Initial state. See {@link State} for details.
     * @returns The prepared transaction request object.
     */
    async prepareCreateChannel(channel: Channel, initial: State): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareCreateChannel';

        try {
            const abiChannel = this.convertChannelForABI(channel);
            const abiState = this.convertStateForABI(initial);

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'create',
                args: [abiChannel, abiState],
                account: account,
            });

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
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'createChannel';

        try {
            const request = await this.prepareCreateChannel(channel, initial);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channel, initial });
        }
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
    async prepareDepositAndCreateChannel(tokenAddress: Address, amount: bigint, channel: Channel, initial: State): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareDepositAndCreateChannel';
        const accountAddress = typeof account === 'string' ? account : account.address;

        try {
            const abiChannel = this.convertChannelForABI(channel);
            const abiState = this.convertStateForABI(initial);

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'depositAndCreate',
                args: [accountAddress, tokenAddress, amount, abiChannel, abiState],
                account: account,
                value: tokenAddress === zeroAddress ? amount : 0n,
            });

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
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'depositAndCreateChannel';

        try {
            const request = await this.prepareDepositAndCreateChannel(tokenAddress, amount, channel, initial);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { tokenAddress, amount, channel, initial });
        }
    }

    /**
     * Prepares the request data for joining an existing channel.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channelId ID of the channel.
     * @param index Participant index.
     * @param sig Participant signature.
     * @returns The prepared transaction request object.
     */
    async prepareJoinChannel(
        channelId: ChannelId,
        index: bigint,
        sig: Signature,
    ): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareJoinChannel';

        try {
            const abiSignature = this.convertSignatureForABI(sig);

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'join',
                args: [channelId, index, abiSignature],
                account: account,
            });

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
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'joinChannel';

        try {
            const request = await this.prepareJoinChannel(channelId, index, sig);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId, index });
        }
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
        const account = this.ensureAccount();
        const operationName = 'prepareCheckpoint';

        try {
            const abiCandidate = this.convertStateForABI(candidate);
            const abiProofs = proofs.map((proof) => this.convertStateForABI(proof));

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'checkpoint',
                args: [channelId, abiCandidate, abiProofs],
                account: account,
            });

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
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'checkpoint';

        try {
            const request = await this.prepareCheckpoint(channelId, candidate, proofs);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId });
        }
    }

    /**
     * Prepares the request data for challenging a state.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channelId Channel ID.
     * @param candidate State being challenged. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @returns The prepared transaction request object.
     */
    async prepareChallenge(
        channelId: ChannelId,
        candidate: State,
        proofs: State[] = [],
    ): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareChallenge';

        try {
            const abiCandidate = this.convertStateForABI(candidate);
            const abiProofs = proofs.map((proof) => this.convertStateForABI(proof));

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'challenge',
                args: [channelId, abiCandidate, abiProofs],
                account: account,
            });

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
    async challenge(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<Hash> {
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'challenge';

        try {
            const request = await this.prepareChallenge(channelId, candidate, proofs);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId });
        }
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
        const account = this.ensureAccount();
        const operationName = 'prepareResize';

        try {
            const abiCandidate = this.convertStateForABI(candidate);
            const abiProofs = proofs.map((proof) => this.convertStateForABI(proof));

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'resize',
                args: [channelId, abiCandidate, abiProofs],
                account: account,
            });

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
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'resize';

        try {
            const request = await this.prepareResize(channelId, candidate, proofs);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId });
        }
    }

    /**
     * Prepares the request data for closing a channel.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channelId Channel ID.
     * @param candidate Final state. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @returns The prepared transaction request object.
     */
    async prepareClose(
        channelId: ChannelId,
        candidate: State,
        proofs: State[] = [],
    ): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareClose';

        try {
            const abiCandidate = this.convertStateForABI(candidate);
            const abiProofs = proofs.map((proof) => this.convertStateForABI(proof));

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'close',
                args: [channelId, abiCandidate, abiProofs],
                account: account,
            });

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
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'close';

        try {
            const request = await this.prepareClose(channelId, candidate, proofs);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId });
        }
    }

    /**
     * Prepares the request data for withdrawing funds.
     * Useful for batching multiple calls in a single UserOperation.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to withdraw.
     * @returns The prepared transaction request object.
     */
    async prepareWithdraw(tokenAddress: Address, amount: bigint): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareWithdraw';

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'withdraw',
                args: [tokenAddress, amount],
                account: account,
            });

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
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'withdraw';

        try {
            const request = await this.prepareWithdraw(tokenAddress, amount);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { tokenAddress, amount });
        }
    }

    /**
     * Get the list of channels for a given account
     * @param account Address of the account
     * @returns List of channel IDs
     * @error Throws ContractReadError if the read operation fails
     */
    async getAccountChannels(account: Address): Promise<ChannelId[]> {
        const functionName = 'getAccountChannels';

        try {
            const result = await this.publicClient.readContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: functionName,
                args: [account],
            });

            return result as ChannelId[];
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractReadError(functionName, error, { account });
        }
    }

    /**
     * Get account balance information for a specific token.
     * @param user The address of the user.
     * @param token The address of the token.
     * @returns An object containing available balance, locked balance, and channel count.
     * @error Throws ContractReadError if the read operation fails.
     */
    async getAccountInfo(user: Address, token: Address): Promise<{ available: bigint; channelCount: bigint }> {
        const functionName = 'getAccountInfo';

        try {
            const result = await this.publicClient.readContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: functionName,
                args: [user, token],
            });

            const [available, channelCount] = result as [bigint, bigint];

            return {
                available: available,
                channelCount: channelCount,
            };
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractReadError(functionName, error, { user, token });
        }
    }
}
