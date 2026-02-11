import { z } from 'zod';
import {
    RPCMethod,
    GetBalancesResponseParams,
    GetTransactionsResponseParams,
    BalanceUpdateResponseParams,
    TransferNotificationResponseParams,
    RPCTransactionType,
    RPCBalanceEntry,
    RPCTransaction,
} from '../types';
import { dateSchema, decimalSchema, paginationMetadataSchema, ParamsParser } from './common';

const BalanceEntrySchema = z
    .object({
        asset: z.string(),
        amount: decimalSchema,
    })
    .transform(
        (raw): RPCBalanceEntry => ({
            asset: raw.asset,
            amount: raw.amount,
        }),
    );

const TransactionSchema = z
    .object({
        id: z.string(),
        asset: z.string(),
        tx_type: z.nativeEnum(RPCTransactionType),
        from_account: z.string(),
        to_account: z.string(),
        sender_new_state_id: z.string().optional(),
        receiver_new_state_id: z.string().optional(),
        amount: decimalSchema,
        created_at: dateSchema,
    })
    .transform(
        (raw): RPCTransaction => ({
            id: raw.id,
            asset: raw.asset,
            txType: raw.tx_type,
            fromAccount: raw.from_account,
            toAccount: raw.to_account,
            senderNewStateId: raw.sender_new_state_id,
            receiverNewStateId: raw.receiver_new_state_id,
            amount: raw.amount,
            createdAt: raw.created_at,
        }),
    );

const GetBalancesParamsSchema = z
    .object({
        balances: z.array(BalanceEntrySchema),
    })
    .transform(
        (raw): GetBalancesResponseParams => ({
            balances: raw.balances,
        }),
    );

const GetTransactionsParamsSchema = z
    .object({
        transactions: z.array(TransactionSchema),
        metadata: paginationMetadataSchema.optional(),
    })
    .transform(
        (raw): GetTransactionsResponseParams => ({
            transactions: raw.transactions,
            metadata: raw.metadata,
        }),
    );

const BalanceUpdateParamsSchema = z
    .object({
        balance_updates: z.array(BalanceEntrySchema),
    })
    .transform(
        (raw): BalanceUpdateResponseParams => ({
            balanceUpdates: raw.balance_updates,
        }),
    );

const TransferNotificationParamsSchema = z
    .object({
        transactions: z.array(TransactionSchema),
    })
    .transform(
        (raw): TransferNotificationResponseParams => ({
            transactions: raw.transactions,
        }),
    );

export const ledgerParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.GetBalances]: (params) => GetBalancesParamsSchema.parse(params),
    [RPCMethod.GetTransactions]: (params) => GetTransactionsParamsSchema.parse(params),
    [RPCMethod.BalanceUpdate]: (params) => BalanceUpdateParamsSchema.parse(params),
    [RPCMethod.TransferNotification]: (params) => TransferNotificationParamsSchema.parse(params),
};
