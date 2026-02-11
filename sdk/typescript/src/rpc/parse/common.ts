import { z } from 'zod';
import { RPCMethod, RPCTransitionType } from '../types';
import { Address, Hex } from 'viem';

// --- Shared Interfaces & Classes ---

export interface ParamsParser<T> {
    (params: object[]): T;
}

export class ParserParamsMissingError extends Error {
    constructor(method: RPCMethod) {
        super(`Missing params for ${method} parser`);
        this.name = 'ParserParamsMissingError';
    }
}

// --- Shared Zod Schemas ---

export const hexSchema = z
    .string()
    .refine((val) => /^0x[0-9a-fA-F]*$/.test(val), {
        message: 'Must be a 0x-prefixed hex string',
    })
    .transform((v: string) => v as Hex);

export const addressSchema = z
    .string()
    .refine((val) => /^0x[0-9a-fA-F]{40}$/.test(val), {
        message: 'Must be a 0x-prefixed hex string of 40 hex chars (EVM address)',
    })
    .transform((v: string) => v as Address);

// TODO: add more validation for bigints if needed
export const bigIntSchema = z.string();

export const dateSchema = z.union([z.string(), z.date()]).transform((v) => new Date(v));

export const decimalSchema = z
    .union([z.string(), z.number()])
    .transform((v) => v.toString())
    .refine((val) => /^[+-]?((\d+(\.\d*)?)|(\.\d+))$/.test(val), {
        message: 'Must be a valid decimal string',
    });

export const paginationMetadataSchema = z
    .object({
        page: z.number(),
        per_page: z.number(),
        total_count: z.number(),
        page_count: z.number(),
    })
    .transform((raw) => ({
        page: raw.page,
        perPage: raw.per_page,
        totalCount: raw.total_count,
        pageCount: raw.page_count,
    }));

export const transitionSchema = z
    .object({
        type: z.nativeEnum(RPCTransitionType),
        tx_hash: z.string().optional(),
        account_id: z.string(),
        amount: decimalSchema,
    })
    .transform((raw) => ({
        type: raw.type,
        txHash: raw.tx_hash,
        accountId: raw.account_id,
        amount: raw.amount,
    }));

export const ledgerSchema = z
    .object({
        token_address: addressSchema,
        blockchain_id: z.number(),
        user_balance: decimalSchema,
        user_net_flow: decimalSchema,
        node_balance: decimalSchema,
        node_net_flow: decimalSchema,
    })
    .transform((raw) => ({
        tokenAddress: raw.token_address,
        blockchainId: raw.blockchain_id,
        userBalance: raw.user_balance,
        userNetFlow: raw.user_net_flow,
        nodeBalance: raw.node_balance,
        nodeNetFlow: raw.node_net_flow,
    }));

export const stateSchema = z
    .object({
        id: z.string(),
        transitions: z.array(transitionSchema),
        asset: z.string(),
        user_wallet: addressSchema,
        epoch: z.number(),
        version: z.number(),
        home_channel_id: z.string().optional(),
        escrow_channel_id: z.string().optional(),
        home_ledger: ledgerSchema,
        escrow_ledger: ledgerSchema.optional(),
        user_sig: hexSchema.optional(),
        node_sig: hexSchema.optional(),
    })
    .transform((raw) => ({
        id: raw.id,
        transitions: raw.transitions,
        asset: raw.asset,
        userWallet: raw.user_wallet,
        epoch: raw.epoch,
        version: raw.version,
        homeChannelId: raw.home_channel_id,
        escrowChannelId: raw.escrow_channel_id,
        homeLedger: raw.home_ledger,
        escrowLedger: raw.escrow_ledger,
        userSig: raw.user_sig,
        nodeSig: raw.node_sig,
    }));

export const channelSchema = z
    .object({
        channel_id: z.string(),
        user_wallet: addressSchema,
        node_wallet: addressSchema,
        type: z.enum(['home', 'escrow']),
        blockchain_id: z.number(),
        token_address: addressSchema,
        challenge: z.number(),
        nonce: z.number(),
        status: z.enum(['void', 'open', 'challenged', 'closed']),
        state_version: z.number(),
    })
    .transform((raw) => ({
        channelId: raw.channel_id,
        userWallet: raw.user_wallet,
        nodeWallet: raw.node_wallet,
        type: raw.type,
        blockchainId: raw.blockchain_id,
        tokenAddress: raw.token_address,
        challenge: raw.challenge,
        nonce: raw.nonce,
        status: raw.status,
        stateVersion: raw.state_version,
    }));

// --- Shared Parser Functions ---

export const noop: ParamsParser<unknown> = (_) => {
    return {};
};
