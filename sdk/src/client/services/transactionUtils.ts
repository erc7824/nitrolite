import { PublicClient, Hash, TransactionReceipt } from 'viem';
import { Errors } from '../../errors';

/**
 * Waits for a transaction to be mined and returns the receipt.
 * @param publicClient - A Viem PublicClient instance.
 * @param hash - The transaction hash to wait for.
 * @returns The transaction receipt.
 * @throws {TransactionError} If the transaction fails, reverts, or waiting times out.
 */
export async function waitForTransaction(publicClient: PublicClient, hash: Hash): Promise<TransactionReceipt> {
    const operationName = 'waitForTransaction';

    try {
        const receipt = await publicClient.waitForTransactionReceipt({ hash });

        if (receipt.status === 'reverted') {
            throw new Error(`Transaction reverted`);
        }

        return receipt;
    } catch (error: any) {
        if (error instanceof Errors.NitroliteError) throw error;
        throw new Errors.TransactionError(operationName, error, { hash });
    }
}
