import { z } from 'zod';
import { Address } from 'viem';
import {
    RPCMethod,
    GetLedgerBalancesResponseParams,
    GetLedgerEntriesResponseParams,
    BalanceUpdateResponseParams,
    GetLedgerTransactionsResponseParams,
    TxType,
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

const txTypeEnum = z.enum(Object.values(TxType) as [string, ...string[]]);

const GetLedgerTransactionsParamsSchema = z
    .array(
        z.array(
            z
                .object({
                    id: z.number(),
                    tx_type: txTypeEnum,
                    from_account: addressSchema,
                    to_account: addressSchema,
                    asset: z.string(),
                    amount: z.union([z.string(), z.number()]).transform((a) => BigInt(a)),
                    created_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
                })
                .strict()
                .transform(
                    (raw) =>
                        ({
                            id: raw.id,
                            txType: raw.tx_type as TxType,
                            fromAccount: raw.from_account as Address,
                            toAccount: raw.to_account as Address,
                            asset: raw.asset,
                            amount: raw.amount,
                            createdAt: raw.created_at,
                        }) as GetLedgerTransactionsResponseParams,
                ),
        ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0] as GetLedgerTransactionsResponseParams[]);

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

export const ledgerParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.GetLedgerBalances]: (params) => GetLedgerBalancesParamsSchema.parse(params),
    [RPCMethod.GetLedgerEntries]: (params) => GetLedgerEntriesParamsSchema.parse(params),
    [RPCMethod.GetLedgerTransactions]: (params) => GetLedgerTransactionsParamsSchema.parse(params),
    [RPCMethod.BalanceUpdate]: (params) => BalanceUpdateParamsSchema.parse(params),
};
