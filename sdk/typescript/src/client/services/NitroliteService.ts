import { Account, Address, PublicClient, WalletClient, Hash, zeroAddress, Hex } from 'viem';
import { custodyAbi } from '../../abis/generated';
import { ContractAddresses } from '../../abis';
import { Errors } from '../../errors';
import { ChannelData, ChannelDefinition, ChannelId, State, Ledger } from '../types';

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
    return walletClient.writeContract({
        ...request,
        account,
    } as any);
};

/**
 * Service for interacting with the Nitrolite V1 Custody smart contract.
 * Provides methods for channel management, deposits, withdrawals, and escrow operations.
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
     * Converts ChannelDefinition to format expected by V1 ABI
     * REQUIRED: All fields must be readonly for ABI compatibility
     */
    private convertChannelDefinitionForABI(definition: ChannelDefinition) {
        return {
            challengeDuration: definition.challengeDuration,
            user: definition.user,
            node: definition.node,
            nonce: definition.nonce,
            metadata: definition.metadata,
        } as const;
    }

    /**
     * Converts Ledger to format expected by V1 ABI
     * REQUIRED: All fields must be readonly for ABI compatibility
     */
    private convertLedgerForABI(ledger: Ledger) {
        return {
            chainId: ledger.chainId,
            token: ledger.token,
            decimals: ledger.decimals,
            userAllocation: ledger.userAllocation,
            userNetFlow: ledger.userNetFlow,
            nodeAllocation: ledger.nodeAllocation,
            nodeNetFlow: ledger.nodeNetFlow,
        } as const;
    }

    /**
     * Converts State to format expected by V1 ABI
     * REQUIRED: All nested structures must be readonly for ABI compatibility
     */
    private convertStateForABI(state: State) {
        return {
            version: state.version,
            intent: state.intent as number,
            metadata: state.metadata,
            homeState: this.convertLedgerForABI(state.homeState),
            nonHomeState: this.convertLedgerForABI(state.nonHomeState),
            userSig: state.userSig,
            nodeSig: state.nodeSig,
        } as const;
    }

    /**
     * Converts contract ChannelDefinition result to SDK type
     */
    private convertChannelDefinitionFromContract(contractDef: any): ChannelDefinition {
        return {
            challengeDuration: contractDef.challengeDuration,
            user: contractDef.user,
            node: contractDef.node,
            nonce: contractDef.nonce,
            metadata: contractDef.metadata,
        };
    }

    /**
     * Converts contract Ledger result to SDK type
     */
    private convertLedgerFromContract(contractLedger: any): Ledger {
        return {
            chainId: contractLedger.chainId,
            token: contractLedger.token,
            decimals: contractLedger.decimals,
            userAllocation: contractLedger.userAllocation,
            userNetFlow: contractLedger.userNetFlow,
            nodeAllocation: contractLedger.nodeAllocation,
            nodeNetFlow: contractLedger.nodeNetFlow,
        };
    }

    /**
     * Converts contract State result to SDK type
     */
    private convertStateFromContract(contractState: any): State {
        return {
            version: contractState.version,
            intent: contractState.intent,
            metadata: contractState.metadata,
            homeState: this.convertLedgerFromContract(contractState.homeState),
            nonHomeState: this.convertLedgerFromContract(contractState.nonHomeState),
            userSig: contractState.userSig,
            nodeSig: contractState.nodeSig,
        };
    }

    /**
     * Prepares the request data for depositing to the vault.
     * Useful for batching multiple calls in a single UserOperation.
     * @param node Address of the node.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to deposit.
     * @returns The prepared transaction request object.
     */
    async prepareDeposit(node: Address, tokenAddress: Address, amount: bigint): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareDeposit';

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'depositToVault',
                args: [node, tokenAddress, amount],
                account: account,
                value: tokenAddress === zeroAddress ? amount : 0n,
            });

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { node, tokenAddress, amount });
        }
    }

    /**
     * Deposit tokens or ETH into the vault.
     * This method simulates and executes the transaction directly.
     * @param node Address of the node.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to deposit.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async deposit(node: Address, tokenAddress: Address, amount: bigint): Promise<Hash> {
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'deposit';

        try {
            const request = await this.prepareDeposit(node, tokenAddress, amount);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { node, tokenAddress, amount });
        }
    }

    /**
     * Prepares the request data for withdrawing from the vault.
     * Useful for batching multiple calls in a single UserOperation.
     * @param to Recipient address.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to withdraw.
     * @returns The prepared transaction request object.
     */
    async prepareWithdraw(
        to: Address,
        tokenAddress: Address,
        amount: bigint,
    ): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareWithdraw';

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'withdrawFromVault',
                args: [to, tokenAddress, amount],
                account: account,
            });

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { to, tokenAddress, amount });
        }
    }

    /**
     * Withdraw funds from the vault.
     * This method simulates and executes the transaction directly.
     * @param to Recipient address.
     * @param tokenAddress Address of the token (use zeroAddress for ETH).
     * @param amount Amount to withdraw.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async withdraw(to: Address, tokenAddress: Address, amount: bigint): Promise<Hash> {
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'withdraw';

        try {
            const request = await this.prepareWithdraw(to, tokenAddress, amount);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { to, tokenAddress, amount });
        }
    }

    /**
     * Prepares the request data for creating a new channel.
     * Useful for batching multiple calls in a single UserOperation.
     * @param definition Channel definition. See {@link ChannelDefinition} for details.
     * @param initialState Initial state. See {@link State} for details.
     * @returns The prepared transaction request object.
     */
    async prepareCreateChannel(
        definition: ChannelDefinition,
        initialState: State,
    ): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareCreateChannel';

        try {
            const abiDefinition = this.convertChannelDefinitionForABI(definition);
            const abiState = this.convertStateForABI(initialState);

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'createChannel',
                args: [abiDefinition, abiState],
                account: account,
            });

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { definition, initialState });
        }
    }

    /**
     * Create a new channel.
     * This method simulates and executes the transaction directly.
     * @param definition Channel definition. See {@link ChannelDefinition} for details.
     * @param initialState Initial state. See {@link State} for details.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async createChannel(definition: ChannelDefinition, initialState: State): Promise<Hash> {
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'createChannel';

        try {
            const request = await this.prepareCreateChannel(definition, initialState);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { definition, initialState });
        }
    }

    /**
     * Prepares the request data for depositing to a channel.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channelId Channel ID.
     * @param candidate New state after deposit. See {@link State} for details.
     * @returns The prepared transaction request object.
     */
    async prepareDepositToChannel(channelId: ChannelId, candidate: State): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareDepositToChannel';

        try {
            const abiState = this.convertStateForABI(candidate);

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'depositToChannel',
                args: [channelId, abiState],
                account: account,
            });

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { channelId, candidate });
        }
    }

    /**
     * Deposit to an existing channel.
     * This method simulates and executes the transaction directly.
     * @param channelId Channel ID.
     * @param candidate New state after deposit. See {@link State} for details.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async depositToChannel(channelId: ChannelId, candidate: State): Promise<Hash> {
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'depositToChannel';

        try {
            const request = await this.prepareDepositToChannel(channelId, candidate);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId, candidate });
        }
    }

    /**
     * Prepares the request data for withdrawing from a channel.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channelId Channel ID.
     * @param candidate New state after withdrawal. See {@link State} for details.
     * @returns The prepared transaction request object.
     */
    async prepareWithdrawFromChannel(channelId: ChannelId, candidate: State): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareWithdrawFromChannel';

        try {
            const abiState = this.convertStateForABI(candidate);

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'withdrawFromChannel',
                args: [channelId, abiState],
                account: account,
            });

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { channelId, candidate });
        }
    }

    /**
     * Withdraw from an existing channel.
     * This method simulates and executes the transaction directly.
     * @param channelId Channel ID.
     * @param candidate New state after withdrawal. See {@link State} for details.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async withdrawFromChannel(channelId: ChannelId, candidate: State): Promise<Hash> {
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'withdrawFromChannel';

        try {
            const request = await this.prepareWithdrawFromChannel(channelId, candidate);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId, candidate });
        }
    }

    /**
     * Prepares the request data for checkpointing a state.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channelId Channel ID.
     * @param candidate State to checkpoint. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @returns The prepared transaction request object.
     */
    async prepareCheckpointChannel(
        channelId: ChannelId,
        candidate: State,
        proofs: State[] = [],
    ): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareCheckpointChannel';

        try {
            const abiCandidate = this.convertStateForABI(candidate);
            const abiProofs = proofs.map((proof) => this.convertStateForABI(proof));

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'checkpointChannel',
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
     * @param channelId Channel ID.
     * @param candidate State to checkpoint. See {@link State} for details.
     * @param proofs Supporting proofs if required by adjudicator. See {@link State} for details.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async checkpointChannel(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<Hash> {
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'checkpointChannel';

        try {
            const request = await this.prepareCheckpointChannel(channelId, candidate, proofs);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId });
        }
    }

    /**
     * Prepares the request data for challenging a channel.
     * Useful for batching multiple calls in a single UserOperation.
     * @param channelId Channel ID.
     * @param candidate State being challenged. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @param challengerSig Challenger's signature.
     * @returns The prepared transaction request object.
     */
    async prepareChallengeChannel(
        channelId: ChannelId,
        candidate: State,
        proofs: State[] = [],
        challengerSig: Hex,
    ): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareChallengeChannel';

        try {
            const abiCandidate = this.convertStateForABI(candidate);
            const abiProofs = proofs.map((proof) => this.convertStateForABI(proof));

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'challengeChannel',
                args: [channelId, abiCandidate, abiProofs, challengerSig],
                account: account,
            });

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { channelId });
        }
    }

    /**
     * Challenge a channel state on-chain.
     * This method simulates and executes the transaction directly.
     * @param channelId Channel ID.
     * @param candidate State being challenged. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @param challengerSig Challenger's signature.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async challengeChannel(channelId: ChannelId, candidate: State, proofs: State[] = [], challengerSig: Hex): Promise<Hash> {
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'challengeChannel';

        try {
            const request = await this.prepareChallengeChannel(channelId, candidate, proofs, challengerSig);
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
    async prepareCloseChannel(
        channelId: ChannelId,
        candidate: State,
        proofs: State[] = [],
    ): Promise<PreparedContractRequest> {
        const account = this.ensureAccount();
        const operationName = 'prepareCloseChannel';

        try {
            const abiCandidate = this.convertStateForABI(candidate);
            const abiProofs = proofs.map((proof) => this.convertStateForABI(proof));

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: 'closeChannel',
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
     * @param channelId Channel ID.
     * @param candidate Final state. See {@link State} for details.
     * @param proofs Supporting proofs. See {@link State} for details.
     * @returns Transaction hash.
     * @error Throws ContractCallError | TransactionError
     */
    async closeChannel(channelId: ChannelId, candidate: State, proofs: State[] = []): Promise<Hash> {
        const walletClient = this.ensureWalletClient();
        const account = this.ensureAccount();
        const operationName = 'closeChannel';

        try {
            const request = await this.prepareCloseChannel(channelId, candidate, proofs);
            return await executeWriteContract(walletClient, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { channelId });
        }
    }

    /**
     * Get the list of open channel IDs for a user
     * @param user Address of the user
     * @returns Array of Channel IDs
     * @error Throws ContractReadError if the read operation fails
     */
    async getOpenChannels(user: Address): Promise<ChannelId[]> {
        const functionName = 'getOpenChannels';

        try {
            const result = await this.publicClient.readContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: functionName,
                args: [user],
            });

            return result as ChannelId[];
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractReadError(functionName, error, { user });
        }
    }

    /**
     * Get all channel IDs for a user (both open and closed)
     * @param user Address of the user
     * @returns Array of Channel IDs
     * @error Throws ContractReadError if the read operation fails
     */
    async getChannelIds(user: Address): Promise<ChannelId[]> {
        const functionName = 'getChannelIds';

        try {
            const result = await this.publicClient.readContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: functionName,
                args: [user],
            });

            return result as ChannelId[];
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractReadError(functionName, error, { user });
        }
    }

    /**
     * Get account balance in the vault for a specific node and token
     * @param node Address of the node
     * @param token Address of the token (use zeroAddress for ETH)
     * @returns Balance as bigint
     * @error Throws ContractReadError if the read operation fails
     */
    async getAccountBalance(node: Address, token: Address): Promise<bigint> {
        const functionName = 'getAccountBalance';

        try {
            const result = await this.publicClient.readContract({
                address: this.custodyAddress,
                abi: custodyAbi,
                functionName: functionName,
                args: [node, token],
            });

            return result as bigint;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractReadError(functionName, error, { node, token });
        }
    }

    /**
     * Get channel data for a specific channel ID.
     * @param channelId ID of the channel to retrieve data for.
     * @returns ChannelData object containing definition, status, last state, and challenge expiry.
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
                status: result[0],
                definition: this.convertChannelDefinitionFromContract(result[1]),
                lastState: this.convertStateFromContract(result[2]),
                challengeExpiry: result[3],
            };
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractReadError(functionName, error, { channelId });
        }
    }
}
