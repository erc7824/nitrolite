import { RPCTxType } from './common';

export interface PaginationFilters {
    /** Pagination offset. */
    offset?: number;
    /** Number of transactions to return. */
    limit?: number;
    /** Sort order by created_at. */
    sort?: 'asc' | 'desc';
}

export interface GetLedgerTransactionsFilters extends PaginationFilters {
    /** Filter by transaction type. */
    tx_type?: RPCTxType;
    /** Filter by asset symbol. */
    asset?: string;
}
