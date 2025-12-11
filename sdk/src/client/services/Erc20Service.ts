import { Account, Address, PublicClient, WalletClient, Hash } from 'viem';
import { Erc20Abi } from '../../abis/token';
import { Errors } from '../../errors';
import { ContractCallParams, ContractWriter } from '../contract_writer/types';
import { EOAContractWriter } from '../contract_writer/eoa';

/**
 * Type utility to properly type the request object from simulateContract
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
 * Service for interacting with ERC20 token contracts.
 * Provides methods for reading balance/allowance and approving spending.
 */
export class Erc20Service {
    private readonly publicClient: PublicClient;
    private readonly account?: Account | Address;
    private readonly contractWriter?: ContractWriter;

    constructor(
        publicClient: PublicClient,
        walletClient?: WalletClient,
        account?: Account | Address,
        contractWriter?: ContractWriter,
    ) {
        if (!publicClient) {
            throw new Errors.MissingParameterError('publicClient');
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

    /**
     * Get the token balance for a specific account.
     * @param tokenAddress Address of the ERC20 token.
     * @param account Address of the account to check balance for.
     * @returns The token balance as a bigint.
     * @error Throws ContractReadError if the read operation fails.
     */
    async getTokenBalance(tokenAddress: Address, account: Address): Promise<bigint> {
        const functionName = 'balanceOf';

        try {
            const balance = await this.publicClient.readContract({
                address: tokenAddress,
                abi: Erc20Abi,
                functionName: functionName,
                args: [account],
            });

            return balance as bigint;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractReadError(functionName, error, { tokenAddress, account });
        }
    }

    /**
     * Get the allowance granted by an owner to a spender.
     * @param tokenAddress Address of the ERC20 token.
     * @param owner Address of the token owner.
     * @param spender Address of the spender.
     * @returns The allowance amount as a bigint.
     * @error Throws ContractReadError if the read operation fails.
     */
    async getTokenAllowance(tokenAddress: Address, owner: Address, spender: Address): Promise<bigint> {
        const functionName = 'allowance';

        try {
            const allowance = await this.publicClient.readContract({
                address: tokenAddress,
                abi: Erc20Abi,
                functionName: functionName,
                args: [owner, spender],
            });

            return allowance as bigint;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractReadError(functionName, error, { tokenAddress, owner, spender });
        }
    }

    // TODO: comment
    prepareApproveCallParams(tokenAddress: Address, spender: Address, amount: bigint): ContractCallParams {
        const account = this.ensureAccount();

        return {
            address: tokenAddress,
            abi: Erc20Abi,
            functionName: 'approve',
            args: [spender, amount],
            account: account,
        };
    }

    /**
     * Prepares the request data for an ERC20 approve transaction.
     * Useful for batching multiple calls in a single UserOperation.
     * @param tokenAddress Address of the ERC20 token.
     * @param spender Address of the spender.
     * @param amount Amount to approve.
     * @returns The prepared transaction request object.
     * @throws {ContractCallError} If simulation fails.
     * @throws {AccountRequiredError} If no account is available for simulation.
     */
    async prepareApprove(tokenAddress: Address, spender: Address, amount: bigint): Promise<PreparedContractRequest> {
        const operationName = 'prepareApprove';

        try {
            const params = this.prepareApproveCallParams(tokenAddress, spender, amount);
            const { request } = await this.publicClient.simulateContract(params);

            return request;
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.ContractCallError(operationName, error, { tokenAddress, spender, amount });
        }
    }

    /**
     * Executes an ERC20 approve transaction.
     * This method simulates and executes the transaction directly.
     * You do not need to call `prepareApprove` separately unless batching operations.
     * @param tokenAddress Address of the ERC20 token.
     * @param spender Address of the spender.
     * @param amount Amount to approve.
     * @returns The transaction hash.
     * @throws {ContractCallError} If simulation fails.
     * @throws {TransactionError} If sending the transaction fails.
     * @throws {WalletClientRequiredError | AccountRequiredError} If wallet/account is missing.
     */
    async approve(tokenAddress: Address, spender: Address, amount: bigint): Promise<Hash> {
        const contractWriter = this.ensureContractWriter();
        const account = this.ensureAccount();
        const operationName = 'approve';
        try {
            const request = await this.prepareApprove(tokenAddress, spender, amount);
            return await executeWriteContract(contractWriter, request, account);
        } catch (error: any) {
            if (error instanceof Errors.NitroliteError) throw error;
            throw new Errors.TransactionError(operationName, error, { tokenAddress, spender, amount });
        }
    }
}
