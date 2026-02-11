import { z } from 'zod';
import {
    RPCMethod,
    RegisterResponseParams,
    GetSessionKeysResponseParams,
    RevokeSessionKeyResponseParams,
    RPCSessionKey,
    RPCAllowanceUsage,
} from '../types';
import { addressSchema, dateSchema, ParamsParser } from './common';

// Allowance usage schema
const AllowanceUsageSchema = z
    .object({
        asset: z.string(),
        allowance: z.string(),
        used: z.string(),
    })
    .transform(
        (raw): RPCAllowanceUsage => ({
            asset: raw.asset,
            allowance: raw.allowance,
            used: raw.used,
        }),
    );

// Session key schema
const SessionKeySchema = z
    .object({
        id: z.number(),
        session_key: addressSchema,
        application: z.string(),
        allowances: z.array(AllowanceUsageSchema),
        scope: z.string().optional(),
        expires_at: dateSchema,
        created_at: dateSchema,
    })
    .transform(
        (raw): RPCSessionKey => ({
            id: raw.id,
            sessionKey: raw.session_key,
            application: raw.application,
            allowances: raw.allowances,
            scope: raw.scope,
            expiresAt: raw.expires_at,
            createdAt: raw.created_at,
        }),
    );

// register response parser
const RegisterParamsSchema = z
    .object({
        challenge_message: z.string(),
    })
    .transform(
        (raw): RegisterResponseParams => ({
            challengeMessage: raw.challenge_message,
        }),
    );

// get_session_keys response parser
const GetSessionKeysParamsSchema = z
    .object({
        session_keys: z.array(SessionKeySchema),
    })
    .transform(
        (raw): GetSessionKeysResponseParams => ({
            sessionKeys: raw.session_keys,
        }),
    );

// revoke_session_key response parser
const RevokeSessionKeyParamsSchema = z
    .object({
        session_key: addressSchema,
    })
    .transform(
        (raw): RevokeSessionKeyResponseParams => ({
            sessionKey: raw.session_key,
        }),
    );

export const authParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.Register]: (params) => RegisterParamsSchema.parse(params),
    [RPCMethod.GetSessionKeys]: (params) => GetSessionKeysParamsSchema.parse(params),
    [RPCMethod.RevokeSessionKey]: (params) => RevokeSessionKeyParamsSchema.parse(params),
};
