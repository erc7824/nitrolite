import { proxy } from 'valtio';
import { type Transaction as LedgerTransaction } from '@erc7824/nitrolite';

export interface ILedgerTransactionsState {
    transactions: LedgerTransaction[];
    areTransactionsFetched: boolean;
    loading: boolean;
    error: string | null;
}

const state = proxy<ILedgerTransactionsState>({
    transactions: [],
    areTransactionsFetched: true,
    loading: false,
    error: null,
});

const LedgerHistoryStore = {
    state,
    setHistory(transactions: LedgerTransaction[]): void {
        state.transactions = Array.isArray(transactions) ? [...transactions] : [];
        state.loading = false;
        state.error = null;
    },
    appendHistory(transactions: LedgerTransaction[]): void {
        // Make sure to avoid duplicates by using a Map
        const idMap = new Map<number, LedgerTransaction>();
        [...state.transactions, ...transactions].forEach((item) => {
            idMap.set(item.id, item);
        });
        state.transactions = Array.from(idMap.values());

        state.loading = false;
        state.error = null;
        if (transactions.length == 0) {
            state.areTransactionsFetched = true; // No more transactions to fetch
        } else {
            state.areTransactionsFetched = false; // More transactions may be available
        }
    },
    setLoading(loading: boolean): void {
        state.loading = loading;
    },
};

export default LedgerHistoryStore;
