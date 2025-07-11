import { z } from 'zod';
import { Address } from 'viem';
import {
    RPCMethod,
    GetLedgerBalancesResponseParams,
    GetLedgerEntriesResponseParams,
    BalanceUpdateResponseParams,
    GetLedgerTransactionsResponseParams,
    TxType,
    Transaction,
    TransferNotificationResponseParams,
    TransferResponseParams,
} from '../types';
import { addressSchema, ParamsParser } from './common';

const GetLedgerBalancesParamsSchema = z
    .array(
        z.array(
            z.object({
                asset: z.string(),
                amount: z.union([z.string(), z.number()]).transform((a) => a.toString()),
            }),
        ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0])
    .transform((arr) => arr as GetLedgerBalancesResponseParams[]);

const GetLedgerEntriesParamsSchema = z
    .array(
        z.array(
            z
                .object({
                    id: z.number(),
                    account_id: z.string(),
                    account_type: z.number(),
                    asset: z.string(),
                    participant: addressSchema,
                    credit: z.union([z.string(), z.number()]).transform((v) => v.toString()),
                    debit: z.union([z.string(), z.number()]).transform((v) => v.toString()),
                    created_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
                })
                .transform(
                    (e) =>
                        ({
                            id: e.id,
                            accountId: e.account_id,
                            accountType: e.account_type,
                            asset: e.asset,
                            participant: e.participant as Address,
                            credit: e.credit,
                            debit: e.debit,
                            createdAt: e.created_at,
                        }) as GetLedgerEntriesResponseParams,
                ),
        ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0])
    .transform((arr) => arr as GetLedgerEntriesResponseParams[]);

export const txTypeEnum = z.nativeEnum(TxType);

export const TransactionSchema = z
    .object({
        id: z.number(),
        tx_type: txTypeEnum,
        from_account: addressSchema,
        from_account_tag: z.string().optional(),
        to_account: addressSchema,
        to_account_tag: z.string().optional(),
        asset: z.string(),
        amount: z.string(),
        created_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
    })
    .transform(
        (raw): Transaction => ({
            id: raw.id,
            txType: raw.tx_type,
            fromAccount: raw.from_account as Address,
            fromAccountTag: raw.from_account_tag,
            toAccount: raw.to_account as Address,
            toAccountTag: raw.to_account_tag,
            asset: raw.asset,
            amount: raw.amount,
            createdAt: raw.created_at,
        }),
    );

const GetLedgerTransactionsParamsSchema = z
    .array(z.array(TransactionSchema))
    .refine((arr) => arr.length === 1)
    .transform((arr): GetLedgerTransactionsResponseParams => arr[0]);

const BalanceUpdateParamsSchema = z
    .array(
        z.array(
            z
                .object({ asset: z.string(), amount: z.union([z.string(), z.number()]).transform((a) => a.toString()) })
                .transform((b) => ({ asset: b.asset, amount: b.amount }) as BalanceUpdateResponseParams),
        ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const TransferParamsSchema = z
    .array(z.array(TransactionSchema))
    .refine((arr) => arr.length === 1)
    .transform((arr): TransferResponseParams => arr[0]);

const TransferNotificationParamsSchema = z
    .array(z.array(TransactionSchema))
    .refine((arr) => arr.length === 1)
    .transform((arr): TransferNotificationResponseParams => arr[0]);

export const ledgerParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.GetLedgerBalances]: (params) => GetLedgerBalancesParamsSchema.parse(params),
    [RPCMethod.GetLedgerEntries]: (params) => GetLedgerEntriesParamsSchema.parse(params),
    [RPCMethod.GetLedgerTransactions]: (params) => GetLedgerTransactionsParamsSchema.parse(params),
    [RPCMethod.BalanceUpdate]: (params) => BalanceUpdateParamsSchema.parse(params),
    [RPCMethod.Transfer]: (params) => TransferParamsSchema.parse(params),
    [RPCMethod.TransferNotification]: (params) => TransferNotificationParamsSchema.parse(params),
};
