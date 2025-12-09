import {
    Account,
    Chain,
    Hash,
    Hex,
    ParseAccount,
    PublicClient,
    TransactionReceipt,
    Transport,
    WalletClient,
} from 'viem';
import { CallsDetails, ContractCallParams, ContractWriter, WriteResult } from './types';

export type EOAContractWriterConfigs = {
    publicClient: PublicClient;
    walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;
};

export class EOAContractWriter implements ContractWriter {
    public readonly publicClient: PublicClient;
    public readonly walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;
    public readonly account: ParseAccount<Account>;

    constructor(configs: EOAContractWriterConfigs) {
        // TODO: add validations
        this.publicClient = configs.publicClient;
        this.walletClient = configs.walletClient;
        this.account = this.walletClient.account;
    }

    async write(callsDetails: CallsDetails): Promise<WriteResult> {
        if (callsDetails.calls.length < 1) {
            throw new Error('No calls provided');
        }

        const result: WriteResult = { txHashes: [] };

        // EOA writer does not support batching, so we execute calls sequentially
        callsDetails.calls.forEach(async (call) => {
            const txHash = await this._writeCall(call);
            await this.waitForTransaction(txHash);

            result.txHashes.push(txHash);
        });

        return result;
    }

    private async _writeCall(callParams: ContractCallParams): Promise<Hex> {
        // TODO: add error handling
        const { request } = await this.publicClient.simulateContract({
            ...callParams,
            account: this.account,
        });

        return this.walletClient.writeContract({
            ...request,
            account: this.account,
        } as any);
    }

    async waitForTransaction(hash: Hash): Promise<TransactionReceipt> {
        const receipt = await this.publicClient.waitForTransactionReceipt({ hash });

        if (receipt.status === 'reverted') {
            throw new Error(`Transaction reverted`);
        }

        return receipt;
    }

    withBatch(): boolean {
        return false;
    }
}
