import { describe, test, expect, beforeEach, jest } from '@jest/globals';
import type { PublicClient, Hash, TransactionReceipt } from 'viem';
import { waitForTransaction } from '../../../src/client/services/transactionUtils';
import Errors from '../../../src/errors';

describe('waitForTransaction', () => {
    const hash = '0xdeadbeef' as Hash;
    let mockPublicClient: jest.Mocked<PublicClient>;

    beforeEach(() => {
        mockPublicClient = {
            waitForTransactionReceipt: jest.fn(),
        } as any;
    });

    test('returns receipt on success', async () => {
        const fakeReceipt = { status: 'success' } as unknown as TransactionReceipt;
        mockPublicClient.waitForTransactionReceipt.mockResolvedValue(fakeReceipt);

        const result = await waitForTransaction(mockPublicClient, hash);

        expect(mockPublicClient.waitForTransactionReceipt).toHaveBeenCalledWith({ hash });
        expect(result).toBe(fakeReceipt);
    });

    test("throws TransactionError when status is 'reverted'", async () => {
        const fakeReceipt = { status: 'reverted' } as unknown as TransactionReceipt;
        mockPublicClient.waitForTransactionReceipt.mockResolvedValue(fakeReceipt);

        await expect(waitForTransaction(mockPublicClient, hash)).rejects.toBeInstanceOf(Errors.TransactionError);
    });

    test('propagates NitroliteError from publicClient', async () => {
        const nitroErr = new Errors.ContractError('nogo');
        mockPublicClient.waitForTransactionReceipt.mockRejectedValueOnce(nitroErr);

        await expect(waitForTransaction(mockPublicClient, hash)).rejects.toBe(nitroErr);
    });

    test('wraps generic errors in TransactionError', async () => {
        mockPublicClient.waitForTransactionReceipt.mockRejectedValueOnce(new Error('oops'));

        await expect(waitForTransaction(mockPublicClient, hash)).rejects.toBeInstanceOf(Errors.TransactionError);
    });
});
