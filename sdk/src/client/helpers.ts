import { Account, Address, Hash } from 'viem';
import { WriteResult } from './contract_writer/types';

export function getLastTxHashFromWriteResult(writeResult: WriteResult): Hash {
    if (writeResult.txHashes.length < 1) {
        throw new Error('No transaction hashes returned from write operation');
    }

    return writeResult.txHashes[writeResult.txHashes.length - 1];
}

export function getAccountAddress(account: Account | Address): Address {
    if (typeof account === 'object' && 'address' in account) {
        return account.address;
    }
    return account;
}
