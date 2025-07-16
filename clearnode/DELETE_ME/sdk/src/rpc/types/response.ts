// Auto-generated response types.
// Generated from JSON schemas.

import type { Address, Hex } from 'viem';
import {RPCMethod, GenericRPCMessage} from '.';

export type TransactionType = "transfer" | "deposit" | "withdrawal" | "app_deposit" | "app_withdrawal";

/**
 * Represents the response structure for the {@link RPCMethod.GetLedgerTransactions} RPC method.
 */
export interface GetLedgerTransactionsResponse extends GenericRPCMessage {
    method: RPCMethod.GetLedgerTransactions;
    params: GetLedgerTransactionsResponseParams;
}

export interface GetLedgerTransactionsResponseParams {
  amount: bigint,
  asset: string,
  createdAt: Date,
  fromAccount: Address,
  fromAccountTag?: string,
  id: number,
  toAccount: Address,
  toAccountTag?: string,
  txType: TransactionType
}

/**
 * Union type for all possible RPC response types.
 * This allows for type-safe handling of different response structures.
 */
export type RPCResponse =
    | GetLedgerTransactionsResponse
;

/**
 * Maps RPC methods to their corresponding parameter types.
 */
export type ExtractResponseByMethod<M extends RPCMethod> = Extract<RPCResponse, { method: M }>;

// Helper type to extract the response type for a given method
export type RPCResponseParams = ExtractResponseByMethod<RPCMethod>['params'];

export type RPCResponseParamsByMethod = {
    [M in RPCMethod]: ExtractResponseByMethod<M>['params'];
};

