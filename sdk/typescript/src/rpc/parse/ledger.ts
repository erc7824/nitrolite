import { z } from 'zod';
import { Address, Hex, isAddress, isHex } from 'viem';
import {
    RPCMethod,
    GetLedgerBalancesResponseParams,
    GetLedgerEntriesResponseParams,
    BalanceUpdateResponseParams,
    GetLedgerTransactionsResponseParams,
    RPCTxType,
    RPCTransaction,
    TransferNotificationResponseParams,
    TransferResponseParams,
    RPCBalance,
    RPCLedgerEntry,
} from '../types';
import { addressSchema, dateSchema, decimalSchema, ParamsParser } from './common';

const BalanceObjectSchema = z
    .object({
        asset: z.string(),
        amount: decimalSchema,
    })
    .transform((b): RPCBalance => b);

const GetLedgerBalancesParamsSchema = z
    .object({
        ledger_balances: z.array(BalanceObjectSchema),
    })
    .transform(
        (raw): GetLedgerBalancesResponseParams => ({
            ledgerBalances: raw.ledger_balances,
        }),
    );

export const ledgerAccountSchema = z
    .string()
    .refine((val) => isAddress(val) || (isHex(val) && val.length === 66), {
        message: 'Must be a valid EVM address or a 0x-prefixed 64-char hex string',
    })
    .transform((v) => v as Hex);

const LedgerEntryObjectSchema = z
    .object({
        id: z.number(),
        account_id: ledgerAccountSchema,
        account_type: z.number(),
        asset: z.string(),
        participant: addressSchema,
        credit: decimalSchema,
        debit: decimalSchema,
        created_at: dateSchema,
    })
    .transform(
        (e): RPCLedgerEntry => ({
            id: e.id,
            accountId: e.account_id,
            accountType: e.account_type,
            asset: e.asset,
            participant: e.participant,
            credit: e.credit,
            debit: e.debit,
            createdAt: e.created_at,
        }),
    );

const GetLedgerEntriesParamsSchema = z
    .object({
        ledger_entries: z.array(LedgerEntryObjectSchema),
    })
    .transform(
        (raw): GetLedgerEntriesResponseParams => ({
            ledgerEntries: raw.ledger_entries,
        }),
    );

export const txTypeEnum = z.nativeEnum(RPCTxType);

export const TransactionSchema = z
    .object({
        id: z.number(),
        tx_type: txTypeEnum,
        from_account: ledgerAccountSchema,
        from_account_tag: z.string().optional(),
        to_account: ledgerAccountSchema,
        to_account_tag: z.string().optional(),
        asset: z.string(),
        amount: z.string(),
        created_at: dateSchema,
    })
    .transform(
        (raw): RPCTransaction => ({
            id: raw.id,
            txType: raw.tx_type,
            fromAccount: raw.from_account,
            fromAccountTag: raw.from_account_tag,
            toAccount: raw.to_account,
            toAccountTag: raw.to_account_tag,
            asset: raw.asset,
            amount: raw.amount,
            createdAt: raw.created_at,
        }),
    );

const GetLedgerTransactionsParamsSchema = z
    .object({
        ledger_transactions: z.array(TransactionSchema),
    })
    .transform(
        (raw): GetLedgerTransactionsResponseParams => ({
            ledgerTransactions: raw.ledger_transactions,
        }),
    );

const BalanceUpdateParamsSchema = z
    .object({
        balance_updates: z.array(BalanceObjectSchema),
    })
    .transform(
        (raw): BalanceUpdateResponseParams => ({
            balanceUpdates: raw.balance_updates,
        }),
    );

const TransferParamsSchema = z
    .object({
        transactions: z.array(TransactionSchema),
    })
    .transform(
        (raw): TransferResponseParams => ({
            transactions: raw.transactions,
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
    [RPCMethod.GetLedgerBalances]: (params) => GetLedgerBalancesParamsSchema.parse(params),
    [RPCMethod.GetLedgerEntries]: (params) => GetLedgerEntriesParamsSchema.parse(params),
    [RPCMethod.GetLedgerTransactions]: (params) => GetLedgerTransactionsParamsSchema.parse(params),
    [RPCMethod.BalanceUpdate]: (params) => BalanceUpdateParamsSchema.parse(params),
    [RPCMethod.Transfer]: (params) => TransferParamsSchema.parse(params),
    [RPCMethod.TransferNotification]: (params) => TransferNotificationParamsSchema.parse(params),
};
