import {
    Address,
    PublicClient,
    WalletClient,
    Account,
    decodeEventLog, // Import decodeEventLog
    zeroAddress,
    parseEventLogs,
    encodeFunctionData,
} from 'viem';
import { Channel, State, ChannelId, Role } from '../types';
import { CustodyAbi, Erc20Abi } from '../abis';
import Errors from '../errors'; // Import Errors
import { Logger, defaultLogger } from '../config';
import { ChannelOpenedEvent } from '../abis/custody';

/**
 * Channel operations that interact with the blockchain
 */
export class ChannelOperations {
    private readonly logger: Logger;

    constructor(
        private readonly publicClient: PublicClient,
        private readonly walletClient: WalletClient | undefined,
        private readonly account: Account | undefined,
        private readonly custodyAddress: Address,
        logger?: Logger
    ) {
        this.logger = logger || defaultLogger;
    }

    /**
     * Check if the client is properly configured for writing transactions
     */
    private ensureWalletClient(): void {
        if (!this.walletClient || !this.account) {
            throw new Errors.NitroliteError(
                'Wallet client and account required for this operation',
                'MISSING_WALLET_CLIENT',
                400,
                'Ensure walletClient and account are provided in NitroliteClient configuration for write operations',
                { operation: 'write' }
            );
        }
    }

    /**
     * Open a new channel or join an existing one
     * @param channel Channel configuration
     * @param deposit Initial state and allocation
     * @returns Channel ID extracted from event logs
     */
    async openChannel(
        channel: Channel,
        deposit: State,
        participantIndex: Role = Role.UNDEFINED
    ): Promise<void> {
        this.ensureWalletClient();

        if (participantIndex === Role.UNDEFINED) {
            throw new Errors.NitroliteError(
                'Participant index is required to open a channel',
                'MISSING_PARTICIPANT_INDEX',
                400,
                'Specify the participant index (0 or 1) when opening a channel',
                { operation: 'openChannel' }
            );
        }

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: 'open',
                args: [channel, deposit],
                account: this.account!,
            });

            const hash = await this.walletClient!.writeContract(request);
            const receipt = await this.publicClient.waitForTransactionReceipt({
                hash,
            });

            if (receipt.status !== 'success') {
                throw new Errors.TransactionError(
                    'Channel opening transaction failed',
                    { receipt }
                );
            }
        } catch (error: any) {
            throw new Errors.ContractCallError(
                `Failed to open channel: ${error.message}`,
                { cause: error, channel, deposit }
            );
        }
    }

    /**
     * Close a channel with a mutually signed state
     * @param channelId Channel identifier
     * @param candidate Latest valid state
     * @param proofs Previous states required for validation
     */
    async closeChannel(
        channelId: ChannelId,
        candidate: State,
        proofs: State[] = []
    ): Promise<void> {
        this.ensureWalletClient();

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: 'close',
                args: [channelId, candidate, proofs],
                account: this.account!,
            });

            const hash = await this.walletClient!.writeContract(request);

            // Wait for transaction to be mined
            const receipt = await this.publicClient.waitForTransactionReceipt({
                hash,
            });

            if (receipt.status !== 'success') {
                throw new Errors.TransactionError(
                    'Channel close transaction failed',
                    { receipt }
                );
            }
        } catch (error: any) {
            const code =
                error instanceof Errors.TransactionError
                    ? 'TRANSACTION_FAILED'
                    : 'CONTRACT_CALL_FAILED';
            // Pass only message and details to ContractCallError constructor
            throw new Errors.ContractCallError(
                `Failed to close channel ${channelId}: ${error.message}`,
                { cause: error, channelId, candidate, proofs, code } // Include original code in details
            );
        }
    }

    /**
     * Challenge a channel when the counterparty is unresponsive
     * @param channelId Channel identifier
     * @param candidate Latest valid state
     * @param proofs Previous states required for validation
     */
    async challengeChannel(
        channelId: ChannelId,
        candidate: State,
        proofs: State[] = []
    ): Promise<void> {
        this.ensureWalletClient();

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: 'challenge',
                args: [channelId, candidate, proofs],
                account: this.account!,
            });

            const hash = await this.walletClient!.writeContract(request);

            // Wait for transaction to be mined
            const receipt = await this.publicClient.waitForTransactionReceipt({
                hash,
            });

            if (receipt.status !== 'success') {
                throw new Errors.TransactionError(
                    'Channel challenge transaction failed',
                    { receipt }
                );
            }
        } catch (error: any) {
            const code =
                error instanceof Errors.TransactionError
                    ? 'TRANSACTION_FAILED'
                    : 'CONTRACT_CALL_FAILED';
            // Pass only message and details to ContractCallError constructor
            throw new Errors.ContractCallError(
                `Failed to challenge channel ${channelId}: ${error.message}`,
                { cause: error, channelId, candidate, proofs, code } // Include original code in details
            );
        }
    }

    /**
     * Checkpoint a state to store it on-chain
     * @param channelId Channel identifier
     * @param candidate Latest valid state
     * @param proofs Previous states required for validation
     */
    async checkpointChannel(
        channelId: ChannelId,
        candidate: State,
        proofs: State[] = []
    ): Promise<void> {
        this.ensureWalletClient();

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: 'checkpoint',
                args: [channelId, candidate, proofs],
                account: this.account!,
            });

            const hash = await this.walletClient!.writeContract(request);

            // Wait for transaction to be mined
            const receipt = await this.publicClient.waitForTransactionReceipt({
                hash,
            });

            if (receipt.status !== 'success') {
                throw new Errors.TransactionError(
                    'Channel checkpoint transaction failed',
                    { receipt }
                );
            }
        } catch (error: any) {
            const code =
                error instanceof Errors.TransactionError
                    ? 'TRANSACTION_FAILED'
                    : 'CONTRACT_CALL_FAILED';
            // Pass only message and details to ContractCallError constructor
            throw new Errors.ContractCallError(
                `Failed to checkpoint channel ${channelId}: ${error.message}`,
                { cause: error, channelId, candidate, proofs, code } // Include original code in details
            );
        }
    }

    /**
     * Reclaim funds after challenge period expires
     * @param channelId Channel identifier
     */
    async reclaimChannel(channelId: ChannelId): Promise<void> {
        this.ensureWalletClient();

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: 'reclaim',
                args: [channelId],
                account: this.account!,
            });

            const hash = await this.walletClient!.writeContract(request);

            // Wait for transaction to be mined
            const receipt = await this.publicClient.waitForTransactionReceipt({
                hash,
            });

            if (receipt.status !== 'success') {
                throw new Errors.TransactionError(
                    'Channel reclaim transaction failed',
                    { receipt }
                );
            }
        } catch (error: any) {
            const code =
                error instanceof Errors.TransactionError
                    ? 'TRANSACTION_FAILED'
                    : 'CONTRACT_CALL_FAILED';
            // Pass only message and details to ContractCallError constructor
            throw new Errors.ContractCallError(
                `Failed to reclaim channel ${channelId}: ${error.message}`,
                { cause: error, channelId, code } // Include original code in details
            );
        }
    }

    async resetChannel(
        channelId: ChannelId,
        candidate: State,
        proofs: State[],
        newChannel: Channel,
        newDeposit: State
    ): Promise<void> {
        this.ensureWalletClient();

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: 'reset',
                args: [channelId, candidate, proofs, newChannel, newDeposit],
                account: this.account!,
            });

            const hash = await this.walletClient!.writeContract(request);
            const receipt = await this.publicClient.waitForTransactionReceipt({
                hash,
            });

            if (receipt.status !== 'success') {
                throw new Errors.TransactionError(
                    'Channel reset transaction failed',
                    { receipt }
                );
            }
        } catch (error: any) {
            throw new Errors.ContractCallError(
                `Failed to reset channel: ${error.message}`,
                {
                    cause: error,
                    channelId,
                    candidate,
                    proofs,
                    newChannel,
                    newDeposit,
                }
            );
        }
    }

    async deposit(tokenAddress: Address, amount: bigint): Promise<void> {
        this.ensureWalletClient();

        try {
            // If depositing tokens, approve first
            if (tokenAddress !== zeroAddress) {
                await this.approveTokens(
                    tokenAddress,
                    amount,
                    this.custodyAddress
                );
            }

            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: 'deposit',
                args: [tokenAddress, amount],
                account: this.account!,
            });

            const hash = await this.walletClient!.writeContract(request);

            const receipt = await this.publicClient.waitForTransactionReceipt({
                hash,
            });

            if (receipt.status !== 'success') {
                throw new Errors.TransactionError(
                    'Deposit transaction failed',
                    { receipt }
                );
            }
        } catch (error: any) {
            throw new Errors.ContractCallError(
                `Failed to deposit tokens: ${error.message}`,
                { cause: error, tokenAddress, amount }
            );
        }
    }

    /**
     * Withdraw tokens from the custody contract
     */
    async withdraw(tokenAddress: Address, amount: bigint): Promise<void> {
        this.ensureWalletClient();

        try {
            const { request } = await this.publicClient.simulateContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: 'withdraw',
                args: [tokenAddress, amount],
                account: this.account!,
            });

            const hash = await this.walletClient!.writeContract(request);
            const receipt = await this.publicClient.waitForTransactionReceipt({
                hash,
            });

            if (receipt.status !== 'success') {
                throw new Errors.TransactionError(
                    'Withdrawal transaction failed',
                    { receipt }
                );
            }
        } catch (error: any) {
            throw new Errors.ContractCallError(
                `Failed to withdraw tokens: ${error.message}`,
                { cause: error, tokenAddress, amount }
            );
        }
    }

    /**
     * Get available balance for an account
     */
    async getAvailableBalance(
        account: Address,
        tokenAddress: Address
    ): Promise<bigint> {
        try {
            return (await this.publicClient.readContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: 'getAvailableBalance',
                args: [account, tokenAddress],
            })) as bigint;
        } catch (error: any) {
            throw new Errors.ContractCallError(
                `Failed to get available balance: ${error.message}`,
                { cause: error, account, tokenAddress }
            );
        }
    }

    /**
     * Get channels associated with an account for a specific token
     */
    async getAccountChannels(
        account: Address,
        tokenAddress: Address
    ): Promise<ChannelId[]> {
        try {
            return (await this.publicClient.readContract({
                address: this.custodyAddress,
                abi: CustodyAbi,
                functionName: 'getAccountChannels',
                args: [account, tokenAddress],
            })) as ChannelId[];
        } catch (error: any) {
            throw new Errors.ContractCallError(
                `Failed to get account channels: ${error.message}`,
                { cause: error, account, tokenAddress }
            );
        }
    }

    /**
     * Get account information
     */
    async getAccountInfo(
        account: Address,
        tokenAddress: Address
    ): Promise<{
        deposited: bigint;
        locked: bigint;
        channelCount: number;
    }> {
        try {
            const [deposited, locked, channelCount] =
                (await this.publicClient.readContract({
                    address: this.custodyAddress,
                    abi: CustodyAbi,
                    functionName: 'getAccountInfo',
                    args: [account, tokenAddress],
                })) as [bigint, bigint, bigint];

            return {
                deposited,
                locked,
                channelCount: Number(channelCount),
            };
        } catch (error: any) {
            throw new Errors.ContractCallError(
                `Failed to get account info: ${error.message}`,
                { cause: error, account, tokenAddress }
            );
        }
    }

    /**
     * Approve tokens for the custody contract
     * @param tokenAddress ERC20 token address
     * @param amount Amount to approve
     * @param spender Address to approve (usually custody contract)
     */
    async approveTokens(
        tokenAddress: Address,
        amount: bigint,
        spender: Address
    ): Promise<void> {
        this.ensureWalletClient();

        try {
            const { request } = await this.publicClient.simulateContract({
                address: tokenAddress,
                abi: Erc20Abi,
                functionName: 'approve',
                args: [spender, amount],
                account: this.account!,
            });

            console.log(request)

            const hash = await this.walletClient!.writeContract(request);

            // Wait for transaction to be mined
            const receipt = await this.publicClient.waitForTransactionReceipt({
                hash,
            });

            if (receipt.status !== 'success') {
                throw new Errors.TransactionError(
                    'Token approval transaction failed',
                    { receipt }
                );
            }
        } catch (error: any) {
            const code =
                error instanceof Errors.TransactionError
                    ? 'TRANSACTION_FAILED'
                    : 'CONTRACT_CALL_FAILED';
            // Pass only message and details to ContractCallError constructor
            throw new Errors.ContractCallError(
                `Failed to approve token ${tokenAddress}: ${error.message}`,
                { cause: error, tokenAddress, amount, spender, code } // Include original code in details
            );
        }
    }

    /**
     * Get token allowance
     * @param tokenAddress ERC20 token address
     * @param owner Token owner
     * @param spender Address allowed to spend
     * @returns Allowance amount
     */
    async getTokenAllowance(
        tokenAddress: Address,
        owner: Address,
        spender: Address
    ): Promise<bigint> {
        return this.publicClient.readContract({
            address: tokenAddress,
            abi: Erc20Abi,
            functionName: 'allowance',
            args: [owner, spender],
        }) as Promise<bigint>; // Cast result to bigint
    }

    /**
     * Get token balance
     * @param tokenAddress ERC20 token address
     * @param account Account to check balance for
     * @returns Token balance
     */
    async getTokenBalance(
        tokenAddress: Address,
        account: Address
    ): Promise<bigint> {
        return this.publicClient.readContract({
            address: tokenAddress,
            abi: Erc20Abi,
            functionName: 'balanceOf',
            args: [account],
        }) as Promise<bigint>; // Cast result to bigint
    }
}
