import { Account, Address, PublicClient, WalletClient, Hash, zeroAddress, SimulateContractReturnType } from "viem";
import { CustodyAbi, ContractAddresses } from "../../abis"; // Adjust path
import { Errors } from "../../errors"; // Use the namespace import
import { Channel, ChannelId, Signature, State } from "../types";

/**
 * Service for interacting directly with the Nitrolite Custody smart contract.
 * Provides methods for channel management, deposits, and withdrawals specific to the Custody contract.
 */
export class NitroliteService {
    private readonly publicClient: PublicClient;
    private readonly walletClient?: WalletClient;
    private readonly account?: Account | Address;
    private readonly addresses: ContractAddresses;

    constructor(publicClient: PublicClient, addresses: ContractAddresses, walletClient?: WalletClient, account?: Account | Address) {
        if (!publicClient) {
            throw new Errors.MissingParameterError("publicClient");
        }

        if (!addresses || !addresses.custody) {
            throw new Errors.MissingParameterError("addresses.custody");
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
     * Prepares the request data for a deposit transaction.
     * Useful for batching multiple calls in a single UserOperation.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to deposit.
     * @returns The prepared transaction request object.
     */
    async prepareDeposit(tokenAddress: Address, amount: bigint): Promise<SimulateContractReturnType["request"]> {
        const account = this.ensureAccount();
        const operationName = "prepareDeposit";

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: "deposit",
                args: [tokenAddress, amount],
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
        const operationName = "deposit";

        try {
            const request = await this.prepareDeposit(tokenAddress, amount);

            return await walletClient.writeContract({ ...request, account });
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
    async prepareCreateChannel(channel: Channel, initial: State): Promise<SimulateContractReturnType["request"]> {
        const account = this.ensureAccount();
        const operationName = "prepareCreateChannel";

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: "create",
                args: [channel, initial],
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
        const operationName = "createChannel";

        try {
            const request = await this.prepareCreateChannel(channel, initial);

            return await walletClient.writeContract({ ...request, account });
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channel, initial });
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
    async prepareJoinChannel(channelId: ChannelId, index: bigint, sig: Signature): Promise<SimulateContractReturnType["request"]> {
        const account = this.ensureAccount();
        const operationName = "prepareJoinChannel";

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: "join",
                args: [channelId, index, sig],
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
        const operationName = "joinChannel";

        try {
            const request = await this.prepareJoinChannel(channelId, index, sig);

            return await walletClient.writeContract({ ...request, account });
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
    async prepareCheckpoint(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<SimulateContractReturnType["request"]> {
        const account = this.ensureAccount();
        const operationName = "prepareCheckpoint";

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: "checkpoint",
                args: [channelId, candidate, proofs],
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
        const operationName = "checkpoint";

        try {
            const request = await this.prepareCheckpoint(channelId, candidate, proofs);
            return await walletClient.writeContract({ ...request, account });
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
    async prepareChallenge(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<SimulateContractReturnType["request"]> {
        const account = this.ensureAccount();
        const operationName = "prepareChallenge";

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: "challenge",
                args: [channelId, candidate, proofs],
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
        const operationName = "challenge";

        try {
            const request = await this.prepareChallenge(channelId, candidate, proofs);

            return await walletClient.writeContract({ ...request, account });
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
    async prepareClose(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<SimulateContractReturnType["request"]> {
        const account = this.ensureAccount();
        const operationName = "prepareClose";

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: "close",
                args: [channelId, candidate, proofs],
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
        const operationName = "close";

        try {
            const request = await this.prepareClose(channelId, candidate, proofs);

            return await walletClient.writeContract({ ...request, account });
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId });
        }
    }

    /**
     * Prepares the request data for resetting a channel.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channelId Old channel ID.
     * @param candidate Final state for old channel. See {@link State} for details.
     * @param proofs Supporting proofs for old channel close. See {@link State} for details.
     * @param newChannel Configuration for the new channel. See {@link Channel} for details.
     * @param newDeposit Initial state for the new channel. See {@link State} for details.
     * @returns The prepared transaction request object.
     */
    async prepareReset(
        channelId: ChannelId,
        candidate: State,
        proofs: State[],
        newChannel: Channel,
        newDeposit: State
    ): Promise<SimulateContractReturnType["request"]> {
        const account = this.ensureAccount();
        const operationName = "prepareReset";

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: "reset",
                args: [channelId, candidate, proofs, newChannel, newDeposit],
                account: account,
            });

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { channelId });
        }
    }

    /**
     * Reset a channel (close old, create new).
     * This method simulates and executes the transaction directly.
     * You do not need to call `prepareReset` separately unless batching operations.
     * @param channelId Old channel ID.
     * @param candidate Final state for old channel. See {@link State} for details.
     * @param proofs Supporting proofs for old channel close. See {@link State} for details.
     * @param newChannel Configuration for the new channel. See {@link Channel} for details.
     * @param newDeposit Initial state for the new channel. See {@link State} for details.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async reset(channelId: ChannelId, candidate: State, proofs: State[], newChannel: Channel, newDeposit: State): Promise<Hash> {
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = "reset";

        try {
            const request = await this.prepareReset(channelId, candidate, proofs, newChannel, newDeposit);

            return await walletClient.writeContract({ ...request, account });
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
    async prepareWithdraw(tokenAddress: Address, amount: bigint): Promise<SimulateContractReturnType["request"]> {
        const account = this.ensureAccount();
        const operationName = "prepareWithdraw";

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: "withdraw",
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
        const operationName = "withdraw";

        try {
            const request = await this.prepareWithdraw(tokenAddress, amount);

            return await walletClient.writeContract({ ...request, account });
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
        const functionName = "getAccountChannels";

        try {
            const result = await this.publicClient.readContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
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
    async getAccountInfo(user: Address, token: Address): Promise<{ available: bigint; locked: bigint; channelCount: bigint }> {
        const functionName = "getAccountInfo";

        try {
            const result = await this.publicClient.readContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: functionName,
                args: [user, token],
            });

            const [available, locked, channelCount] = result as [bigint, bigint, bigint];

            return {
                available: available,
                locked: locked,
                channelCount: channelCount,
            };
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractReadError(functionName, error, { user, token });
        }
    }
}
