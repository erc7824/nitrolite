import {
  Address,
  PublicClient,
  WalletClient,
  Account,
  decodeEventLog // Import decodeEventLog
} from 'viem';
import {
  Channel,
  State,
  ChannelId
} from '../types';
import { 
  CustodyAbi,
  Erc20Abi
} from '../abis';
// Remove getChannelId import if no longer needed elsewhere
// import { getChannelId } from '../utils';
import Errors from '../errors'; // Import Errors
import { Logger, defaultLogger } from '../config';

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
    deposit: State
  ): Promise<ChannelId> {
    this.ensureWalletClient();

    // If allocation amount > 0, we need to approve the token first
    const hostAllocation = deposit.allocations[0];
    if (hostAllocation.amount > 0 && hostAllocation.token !== '0x0000000000000000000000000000000000000000') {
      try {
        // Check current allowance before approving
        const currentAllowance = await this.getTokenAllowance(
          hostAllocation.token,
          this.account!.address,
          this.custodyAddress
        );
        
        // Only approve if the current allowance is insufficient
        if (currentAllowance < hostAllocation.amount) {
          this.logger.info('Approving tokens for channel', {
            token: hostAllocation.token,
            amount: hostAllocation.amount,
            custodyContract: this.custodyAddress
          });
          
          await this.approveTokens(
            hostAllocation.token,
            hostAllocation.amount,
            this.custodyAddress
          );
        } else {
          this.logger.info('Token allowance sufficient, skipping approval', {
            token: hostAllocation.token,
            currentAllowance,
            requiredAmount: hostAllocation.amount
          });
        }
      } catch (error: any) {
        // Handle common ERC20 errors
        const message = error.message || 'Unknown token error';
        
        // Attempt to categorize the error
        let errorCode = 'TOKEN_APPROVAL_FAILED';
        let suggestion = 'Check token contract and permissions';
        
        // Specific error handling based on common patterns
        if (message.includes('insufficient allowance')) {
          errorCode = 'INSUFFICIENT_ALLOWANCE';
          suggestion = 'Increase token allowance for the custody contract';
        } else if (message.includes('insufficient balance')) {
          errorCode = 'INSUFFICIENT_BALANCE';
          suggestion = 'Ensure you have enough tokens in your wallet';
        } else if (message.includes('reverted')) {
          errorCode = 'CONTRACT_REVERTED';
          suggestion = 'The token contract rejected the approval transaction';
        }
        
        throw new Errors.TokenError(
          `Failed to approve tokens for channel: ${message}`,
          errorCode,
          400,
          suggestion,
          { 
            cause: error, 
            token: hostAllocation.token, 
            amount: hostAllocation.amount, 
            spender: this.custodyAddress 
          }
        );
      }
    }

    try {
      // Call the open function on the custody contract
      const { request } = await this.publicClient.simulateContract({
        address: this.custodyAddress,
        abi: CustodyAbi,
      functionName: 'open',
      args: [channel, deposit],
        account: this.account!
      });

      const hash = await this.walletClient!.writeContract(request);
      
      // Wait for transaction to be mined
      const receipt = await this.publicClient.waitForTransactionReceipt({
        hash,
      });

      // Check transaction status
    if (receipt.status !== 'success') {
      throw new Errors.TransactionError('Channel opening transaction failed', {
        receipt,
        channel,
        deposit
      });
    }

    // Extract channelId from logs
    let channelId: ChannelId | null = null;
    const eventName = 'ChannelOpened'; // Assuming this is the event name

    for (const log of receipt.logs) {
      try {
        // Check if the log originates from the custody contract
        if (log.address.toLowerCase() === this.custodyAddress.toLowerCase()) {
          const decodedLog = decodeEventLog({
            abi: CustodyAbi,
            data: log.data,
            topics: log.topics,
            strict: false // Allow extra data in log
          });

          if (decodedLog.eventName === eventName) {
            // Assuming the event has a 'channelId' argument
            channelId = (decodedLog.args as any)?.channelId;
            if (channelId) {
              break; // Found the channelId, exit loop
            }
          }
        }
      } catch (e) {
        // Ignore logs that don't match the ABI or event signature
        // console.warn('Could not decode log:', e);
      }
    }

    if (!channelId) {
      throw new Errors.ContractError(
        `Could not find ${eventName} event log in transaction receipt`,
        'EVENT_NOT_FOUND', // Specific error code
        500,
        'Check transaction receipt and contract event definitions', // Specific suggestion
        { receipt, custodyAddress: this.custodyAddress } // Details object
      );
    }

    return channelId;

    } catch (error: any) {
      // Catch simulation, write, or receipt errors
      const code = error instanceof Errors.TransactionError ? 'TRANSACTION_FAILED' : 'CONTRACT_CALL_FAILED';
      const suggestion = error instanceof Errors.TransactionError 
        ? 'Check transaction status and parameters'
        : 'Check contract simulation parameters and network status';

      // Pass only message and details to ContractCallError constructor
      throw new Errors.ContractCallError(
        `Failed to open channel: ${error.message}`,
        { cause: error, channel, deposit, code } // Include original code in details
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
        account: this.account!
      });

      const hash = await this.walletClient!.writeContract(request);
      
      // Wait for transaction to be mined
      const receipt = await this.publicClient.waitForTransactionReceipt({
        hash,
      });

      if (receipt.status !== 'success') {
        throw new Errors.TransactionError('Channel close transaction failed', { receipt });
      }
    } catch (error: any) {
      const code = error instanceof Errors.TransactionError ? 'TRANSACTION_FAILED' : 'CONTRACT_CALL_FAILED';
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
        account: this.account!
      });

      const hash = await this.walletClient!.writeContract(request);
      
      // Wait for transaction to be mined
      const receipt = await this.publicClient.waitForTransactionReceipt({
        hash,
      });

      if (receipt.status !== 'success') {
        throw new Errors.TransactionError('Channel challenge transaction failed', { receipt });
      }
    } catch (error: any) {
      const code = error instanceof Errors.TransactionError ? 'TRANSACTION_FAILED' : 'CONTRACT_CALL_FAILED';
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
        account: this.account!
      });

      const hash = await this.walletClient!.writeContract(request);
      
      // Wait for transaction to be mined
      const receipt = await this.publicClient.waitForTransactionReceipt({
        hash,
      });

      if (receipt.status !== 'success') {
        throw new Errors.TransactionError('Channel checkpoint transaction failed', { receipt });
      }
    } catch (error: any) {
      const code = error instanceof Errors.TransactionError ? 'TRANSACTION_FAILED' : 'CONTRACT_CALL_FAILED';
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
        account: this.account!
      });

      const hash = await this.walletClient!.writeContract(request);
      
      // Wait for transaction to be mined
      const receipt = await this.publicClient.waitForTransactionReceipt({
        hash,
      });

      if (receipt.status !== 'success') {
        throw new Errors.TransactionError('Channel reclaim transaction failed', { receipt });
      }
    } catch (error: any) {
      const code = error instanceof Errors.TransactionError ? 'TRANSACTION_FAILED' : 'CONTRACT_CALL_FAILED';
      // Pass only message and details to ContractCallError constructor
      throw new Errors.ContractCallError(
        `Failed to reclaim channel ${channelId}: ${error.message}`,
        { cause: error, channelId, code } // Include original code in details
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
        account: this.account!
      });

      const hash = await this.walletClient!.writeContract(request);
      
      // Wait for transaction to be mined
      const receipt = await this.publicClient.waitForTransactionReceipt({
        hash,
      });

      if (receipt.status !== 'success') {
        throw new Errors.TransactionError('Token approval transaction failed', { receipt });
      }
    } catch (error: any) {
      const code = error instanceof Errors.TransactionError ? 'TRANSACTION_FAILED' : 'CONTRACT_CALL_FAILED';
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
      args: [owner, spender]
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
      args: [account]
    }) as Promise<bigint>; // Cast result to bigint
  }
}
