import { z } from 'zod';
import { RPCMethod } from '../types';
import { AddressSchema, HexSchema } from './common_gen';
import type {
  Address,
  TransactionType,
  BigNumber,
  TransactionResponse,
} from '../types/response';
import {
  AddressSchema,
  TransactionTypeSchema
} from './common_gen';

// Response schemas with camelCase transforms

export const BigNumberSchema = z.string().transform((v) => BigInt(v));

export const TransactionResponseSchema = z.object({
  amount: BigNumberSchema,
  asset: z.string(),
  created_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
  from_account: AddressSchema,
  from_account_tag: z.string().optional(),
  id: z.number(),
  to_account: AddressSchema,
  to_account_tag: z.string().optional(),
  tx_type: TransactionTypeSchema
})
    .transform((raw) => ({
      amount: raw.amount,
      asset: raw.asset,
      createdAt: raw.created_at,
      fromAccount: raw.from_account,
      fromAccountTag: raw.from_account_tag,
      id: raw.id,
      toAccount: raw.to_account,
      toAccountTag: raw.to_account_tag,
      txType: raw.tx_type
    }) as TransactionResponse);

// Response parser mapping
export const responseParsers: Record<string, (params: any) => any> = {
  [RPCMethod.GetLedgerTransactions]: (params) => TransactionResponseSchema.parse(params),
};
