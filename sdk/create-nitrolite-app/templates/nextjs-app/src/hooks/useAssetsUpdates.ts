import AssetsStore, { type LedgerBalance, type TAsset } from '../store/AssetsStore';
import {
    type GetAssetsResponseParams,
    type BalanceUpdateResponseParams,
    type GetLedgerTransactionsResponseParams,
    type UserTagParams,
} from '@erc7824/nitrolite';
import LedgerHistoryStore from '../store/LedgerHistoryStore';
import UserTagStore from '../store/UserTagStore';

/**
 * Processes a balance update message
 */
export function handleBalanceUpdate(balanceUpdates: BalanceUpdateResponseParams[]) {
    // Check if balanceUpdates is an array with elements
    console.log('Received balance updates:', balanceUpdates);

    if (!Array.isArray(balanceUpdates) || balanceUpdates.length === 0) {
        AssetsStore.setLedgerBalancesFirstLoading();
        console.warn('No balance updates received or invalid format:', balanceUpdates);
        return;
    }

    // Handle both possible formats: direct array of balances or nested array
    let balances: LedgerBalance[];

    if (Array.isArray(balanceUpdates[0])) {
        // Format: [[ {balance1}, {balance2}, ... ]]
        balances = balanceUpdates[0] as LedgerBalance[];
    } else if (typeof balanceUpdates[0] === 'object' && balanceUpdates[0] !== null) {
        // Format: [ {balance1}, {balance2}, ... ]
        balances = balanceUpdates as LedgerBalance[];
    } else {
        return;
    }

    // Update store with new balances
    AssetsStore.setLedgerBalances(balances);

    // Clear loading state after successful update
    AssetsStore.setLedgerBalancesLoading(false);
}

export function handleAssetsUpdate(assetsUpdates: GetAssetsResponseParams[]) {
    if (!Array.isArray(assetsUpdates) || assetsUpdates.length === 0) {
        return;
    }

    let assets: TAsset[];

    if (Array.isArray(assetsUpdates[0])) {
        // Format: [[ {asset1}, {asset2}, ... ]]
        assets = assetsUpdates[0].map((asset: any) => ({
            token: asset.token,
            chainId: asset.chain_id,
            symbol: asset.symbol,
            decimals: asset.decimals,
        }));
    } else if (typeof assetsUpdates[0] === 'object' && assetsUpdates[0] !== null) {
        // Format: [ {asset1}, {asset2}, ... ]
        assets = assetsUpdates.map((asset: any) => ({
            token: asset.token,
            chainId: asset.chain_id,
            symbol: asset.symbol,
            decimals: asset.decimals,
        }));
    } else {
        return;
    }

    AssetsStore.setAssets(assets);
}

export function handleLedgerTransactionsUpdate(txUpdates: GetLedgerTransactionsResponseParams) {
    // Handle both possible formats: direct array of balances or nested array
    let transactions: GetLedgerTransactionsResponseParams = [];

    if (Array.isArray(txUpdates[0])) {
        // Format: [[ {tx1}, {tx2}, ... ]]
        transactions = txUpdates[0] as GetLedgerTransactionsResponseParams;
    } else if (typeof txUpdates[0] === 'object' && txUpdates[0] !== null) {
        // Format: [ {tx1}, {tx2}, ... ]
        transactions = txUpdates as GetLedgerTransactionsResponseParams;
    } else if (!Array.isArray(txUpdates)) {
        console.error('Invalid format for ledger transactions update:', txUpdates);
        return;
    }

    LedgerHistoryStore.appendHistory(transactions);
}

/**
 * Processes a user tag response
 */
export function handleGetUserTag(userTagResponse: UserTagParams) {
    console.log('Received user tag update:', userTagResponse);

    if (userTagResponse && typeof userTagResponse.tag === 'string') {
        UserTagStore.setUserTag(userTagResponse.tag);
    } else {
        console.error('Invalid user tag response format:', userTagResponse);
        UserTagStore.setError('Invalid user tag response format');
    }
}
