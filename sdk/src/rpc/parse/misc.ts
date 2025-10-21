import { z } from 'zod';
import {
    RPCMethod,
    GetConfigResponseParams,
    ErrorResponseParams,
    GetRPCHistoryResponseParams,
    RPCNetworkInfo,
    RPCHistoryEntry,
    GetUserTagResponseParams,
    GetSessionKeysResponseParams,
    RPCSessionKey,
    RPCAllowance,
} from '../types';
import { hexSchema, addressSchema, ParamsParser } from './common';

const NetworkInfoObjectSchema = z
    .object({
        chain_id: z.number(),
        custody_address: addressSchema,
        adjudicator_address: addressSchema,
    })
    .transform(
        (raw): RPCNetworkInfo => ({
            chainId: raw.chain_id,
            custodyAddress: raw.custody_address,
            adjudicatorAddress: raw.adjudicator_address,
        }),
    );

const GetConfigParamsSchema = z
    .object({ broker_address: addressSchema, networks: z.array(NetworkInfoObjectSchema) })
    .strict()
    .transform(
        (raw): GetConfigResponseParams => ({
            brokerAddress: raw.broker_address,
            networks: raw.networks,
        }),
    );

const ErrorParamsSchema = z
    .object({ error: z.string() })
    // Validate received type with linter
    .transform((raw): ErrorResponseParams => raw);

const RPCEntryObjectSchema = z
    .object({
        id: z.number(),
        sender: addressSchema,
        req_id: z.number(),
        method: z.string(),
        params: z.string(),
        timestamp: z.number(),
        req_sig: z.array(hexSchema),
        res_sig: z.array(hexSchema),
        response: z.string(),
    })
    .transform(
        (raw): RPCHistoryEntry => ({
            id: raw.id,
            sender: raw.sender,
            reqId: raw.req_id,
            method: raw.method,
            params: raw.params,
            timestamp: raw.timestamp,
            reqSig: raw.req_sig,
            resSig: raw.res_sig,
            response: raw.response,
        }),
    );

const GetRPCHistoryParamsSchema = z
    .object({
        rpc_entries: z.array(RPCEntryObjectSchema),
    })
    .transform(
        (raw): GetRPCHistoryResponseParams => ({
            rpcEntries: raw.rpc_entries,
        }),
    );

const GetUserTagParamsSchema = z
    .object({
        tag: z.string(),
    })
    .strict()
    // Validate received type with linter
    .transform((raw): GetUserTagResponseParams => raw);

const AllowanceObjectSchema = z
    .object({
        asset: z.string(),
        amount: z.string(),
    })
    .transform(
        (raw): RPCAllowance => ({
            asset: raw.asset,
            amount: raw.amount,
        }),
    );

const SessionKeyObjectSchema = z
    .object({
        id: z.number(),
        session_key: addressSchema,
        app_name: z.string().optional(),
        app_address: z.string().optional(),
        allowance: z.array(AllowanceObjectSchema),
        used_allowance: z.array(AllowanceObjectSchema),
        scope: z.string().optional(),
        expires_at: z.string().optional(),
        created_at: z.string(),
    })
    .transform(
        (raw): RPCSessionKey => ({
            id: raw.id,
            sessionKey: raw.session_key,
            appName: raw.app_name,
            appAddress: raw.app_address,
            allowance: raw.allowance,
            usedAllowance: raw.used_allowance,
            scope: raw.scope,
            expiresAt: raw.expires_at ? new Date(raw.expires_at) : undefined,
            createdAt: new Date(raw.created_at),
        }),
    );

const GetSessionKeysParamsSchema = z
    .object({
        session_keys: z.array(SessionKeyObjectSchema),
    })
    .transform(
        (raw): GetSessionKeysResponseParams => ({
            sessionKeys: raw.session_keys,
        }),
    );

const parseMessageParams: ParamsParser<unknown> = (params) => {
    return params;
};

export const miscParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.GetConfig]: (params) => GetConfigParamsSchema.parse(params),
    [RPCMethod.Error]: (params) => ErrorParamsSchema.parse(params),
    [RPCMethod.GetRPCHistory]: (params) => GetRPCHistoryParamsSchema.parse(params),
    [RPCMethod.GetUserTag]: (params) => GetUserTagParamsSchema.parse(params),
    [RPCMethod.GetSessionKeys]: (params) => GetSessionKeysParamsSchema.parse(params),
    [RPCMethod.Message]: parseMessageParams,
};
