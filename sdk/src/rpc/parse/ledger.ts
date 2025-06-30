import { z } from 'zod';
import { Address } from 'viem';
import {
    RPCMethod,
    GetLedgerBalancesResponseParams,
    GetLedgerEntriesResponseParams,
    BalanceUpdateResponseParams,
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
                    account_type: z.string(),
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
    [RPCMethod.BalanceUpdate]: (params) => BalanceUpdateParamsSchema.parse(params),
};
