import { z } from 'zod';
import { Address } from 'viem';

// Common schemas used by both requests and responses

export const AddressSchema = z.string().refine((val) => /^0x[0-9a-fA-F]{40}$/.test(val), {
  message: 'Must be a 0x-prefixed hex string of 40 hex chars (EVM address)',
});

export const TransactionTypeSchema = z.enum(["transfer", "deposit", "withdrawal", "app_deposit", "app_withdrawal"]);

