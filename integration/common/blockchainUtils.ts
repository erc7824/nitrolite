import {
    WaitForTransactionReceiptParameters,
    Hash,
    createPublicClient,
    http,
    Address,
    erc20Abi,
    formatUnits,
    createTestClient,
} from 'viem';
import { chain } from './setup';

export class BlockchainUtils {
    private client = null;
    private testClient = null;
    private lastSnapshotId: string | null = null;

    constructor() {
        this.client = createPublicClient({
            chain,
            transport: http(),
        });

        this.testClient = createTestClient({
            chain,
            transport: http(),
            mode: 'anvil',
        });
    }

    async waitForTransaction(
        txHash: Hash,
        timeoutMs: number = 5000,
        confirmations: number = 0
    ): Promise<WaitForTransactionReceiptParameters> {
        try {
            const timeoutPromise = new Promise((_, reject) => {
                setTimeout(() => {
                    reject(new Error(`Transaction wait timeout after ${timeoutMs}ms`));
                }, timeoutMs);
            });

            const receiptPromise = this.client.waitForTransactionReceipt({
                hash: txHash,
                confirmations,
            });

            const receipt = await Promise.race([receiptPromise, timeoutPromise]);
            return receipt as WaitForTransactionReceiptParameters;
        } catch (error) {
            throw new Error(`Error waiting for transaction: ${error.message}`);
        }
    }

    async getBalance(address: `0x${string}`): Promise<bigint> {
        try {
            const balance = await this.client.getBalance({ address });
            return balance;
        } catch (error) {
            throw new Error(`Error getting balance: ${error.message}`);
        }
    }

    async getErc20Balance(
        tokenAddress: Address,
        userAddress: Address,
        decimals?: number
    ): Promise<{ rawBalance: bigint; formattedBalance: string }> {
        try {
            // Get balance
            const balance = await this.client.readContract({
                address: tokenAddress,
                abi: erc20Abi,
                functionName: 'balanceOf',
                args: [userAddress],
            });

            const tokenDecimals =
                decimals ??
                (await this.client.readContract({
                    address: tokenAddress,
                    abi: erc20Abi,
                    functionName: 'decimals',
                }));

            return {
                rawBalance: balance,
                formattedBalance: formatUnits(balance, tokenDecimals),
            };
        } catch (error) {
            throw new Error(`Error getting ERC20 balance: ${error.message}`);
        }
    }

    async makeSnapshot(): Promise<string> {
        try {
            const snapshotId = await this.testClient.snapshot();
            this.lastSnapshotId = snapshotId;

            return snapshotId;
        } catch (error) {
            throw new Error(`Error making snapshot: ${error.message}`);
        }
    }

    async resetSnapshot(snapshotId?: string): Promise<void> {
        try {
            if (!snapshotId && !this.lastSnapshotId) {
                throw new Error('No snapshot ID provided and no last snapshot available');
            }

            snapshotId = snapshotId || this.lastSnapshotId!;
            await this.testClient.revert({ id: snapshotId });
        } catch (error) {
            throw new Error(`Error resetting snapshot: ${error.message}`);
        }
    }
}
