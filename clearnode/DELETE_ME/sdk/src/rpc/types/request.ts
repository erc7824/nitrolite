// Auto-generated request types.
// Generated from JSON schemas.

import type { Address, Hex } from 'viem';
import {RPCMethod, GenericRPCMessage} from '.';

export type TransactionType = "transfer" | "deposit" | "withdrawal" | "app_deposit" | "app_withdrawal";

export type SortType = "asc" | "desc";

/**
 * Represents the request structure for the {@link RPCMethod.GetLedgerTransactions} RPC method.
 */
export interface GetLedgerTransactionsRequest extends GenericRPCMessage {
    method: RPCMethod.GetLedgerTransactions;
    params: GetLedgerTransactionsRequestParams;
}

export interface GetLedgerTransactionsRequestParams {
  accountId?: Address,
  asset?: string,
  limit?: number,
  offset?: number,
  sort?: SortType,
  txType?: TransactionType
}

/**
 * Union type for all possible RPC request types.
 * This allows for type-safe handling of different request structures.
 */
export type RPCRequest =
    | GetLedgerTransactionsRequest
;

/**
 * Maps RPC methods to their corresponding parameter types.
 */
export type ExtractRequestByMethod<M extends RPCMethod> = Extract<RPCRequest, { method: M }>;

// Helper type to extract the request type for a given method
export type RPCRequestParams = ExtractRequestByMethod<RPCMethod>['params'];

export type RPCRequestParamsByMethod = {
    [M in RPCMethod]: ExtractRequestByMethod<M>['params'];
};

