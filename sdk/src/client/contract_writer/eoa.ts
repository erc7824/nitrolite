import {
    Account,
    Chain,
    Client,
    Hash,
    Hex,
    ParseAccount,
    publicActions,
    PublicClient,
    TransactionReceipt,
    Transport,
    WalletActions,
    WalletClient,
} from 'viem';
import { CallsDetails, ContractCallParams, ContractWriter, WriteResult } from './types';
import Errors from '../../errors';

export type EOAContractWriterConfig = {
    publicClient: PublicClient;
    walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;
};

export class EOAContractWriter implements ContractWriter {
    public readonly publicClient: PublicClient;
    public readonly walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;
    public readonly account: ParseAccount<Account>;

    constructor(config: EOAContractWriterConfig) {
        if (!config.publicClient) throw new Errors.MissingParameterError('publicClient');
        if (!config.walletClient) throw new Errors.MissingParameterError('walletClient');
        if (!config.walletClient.account) throw new Errors.MissingParameterError('walletClient.account');

        this.publicClient = config.publicClient;
        this.walletClient = config.walletClient;
        this.account = this.walletClient.account;
    }

    async write(callsDetails: CallsDetails): Promise<WriteResult> {
        if (callsDetails.calls.length < 1) {
            throw new Error('No calls provided');
        }

        const result: WriteResult = { txHashes: [] };

        // EOA writer does not support batching, so we execute calls sequentially
        for (const call of callsDetails.calls) {
            const txHash = await this._writeCall(call);
            await this.waitForTransaction(txHash);

            result.txHashes.push(txHash);
        }

        return result;
    }

    getAccount(): Account {
        return this.account;
    }

    private async _writeCall(callParams: ContractCallParams): Promise<Hex> {
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
}
