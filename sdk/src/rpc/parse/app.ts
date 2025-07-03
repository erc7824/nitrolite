import { z } from 'zod';
import { Address } from 'viem';
import {
    RPCMethod,
    CreateAppSessionResponseParams,
    SubmitAppStateResponseParams,
    CloseAppSessionResponseParams,
    GetAppDefinitionResponseParams,
    GetAppSessionsResponseParams,
    RPCChannelStatus,
} from '../types';
import { hexSchema, addressSchema, statusEnum, ParamsParser } from './common';

const CreateAppSessionParamsSchema = z
    .array(
        z.object({ app_session_id: hexSchema, version: z.number(), status: statusEnum }).transform(
            (raw) =>
                ({
                    appSessionId: raw.app_session_id as `0x${string}`,
                    version: raw.version,
                    status: raw.status as RPCChannelStatus,
                }) as CreateAppSessionResponseParams,
        ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const SubmitAppStateParamsSchema = z
    .array(
        z.object({ app_session_id: hexSchema, version: z.number(), status: statusEnum }).transform(
            (raw) =>
                ({
                    appSessionId: raw.app_session_id as `0x${string}`,
                    version: raw.version,
                    status: raw.status as RPCChannelStatus,
                }) as SubmitAppStateResponseParams,
        ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const CloseAppSessionParamsSchema = z
    .array(
        z.object({ app_session_id: hexSchema, version: z.number(), status: statusEnum }).transform(
            (raw) =>
                ({
                    appSessionId: raw.app_session_id as `0x${string}`,
                    version: raw.version,
                    status: raw.status as RPCChannelStatus,
                }) as CloseAppSessionResponseParams,
        ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const GetAppDefinitionParamsSchema = z
    .array(
        z
            .object({
                protocol: z.string(),
                participants: z.array(addressSchema),
                weights: z.array(z.number()),
                quorum: z.number(),
                challenge: z.number(),
                nonce: z.number(),
            })
            .transform(
                (raw) =>
                    ({
                        protocol: raw.protocol,
                        participants: raw.participants as Address[],
                        weights: raw.weights,
                        quorum: raw.quorum,
                        challenge: raw.challenge,
                        nonce: raw.nonce,
                    }) as GetAppDefinitionResponseParams,
            ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const GetAppSessionsParamsSchema = z
    .array(
        z.array(
            z
                .object({
                    app_session_id: hexSchema,
                    status: statusEnum,
                    participants: z.array(addressSchema),
                    protocol: z.string(),
                    challenge: z.number(),
                    weights: z.array(z.number()),
                    quorum: z.number(),
                    version: z.number(),
                    nonce: z.number(),
                    created_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
                    updated_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
                })
                .transform(
                    (s) =>
                        ({
                            appSessionId: s.app_session_id as `0x${string}`,
                            status: s.status as RPCChannelStatus,
                            participants: s.participants as Address[],
                            protocol: s.protocol,
                            challenge: s.challenge,
                            weights: s.weights,
                            quorum: s.quorum,
                            version: s.version,
                            nonce: s.nonce,
                            createdAt: s.created_at,
                            updatedAt: s.updated_at,
                        }) as GetAppSessionsResponseParams,
                ),
        ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0])
    .transform((arr) => arr as GetAppSessionsResponseParams[]);

export const appParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.CreateAppSession]: (params) => CreateAppSessionParamsSchema.parse(params),
    [RPCMethod.SubmitAppState]: (params) => SubmitAppStateParamsSchema.parse(params),
    [RPCMethod.CloseAppSession]: (params) => CloseAppSessionParamsSchema.parse(params),
    [RPCMethod.GetAppDefinition]: (params) => GetAppDefinitionParamsSchema.parse(params),
    [RPCMethod.GetAppSessions]: (params) => GetAppSessionsParamsSchema.parse(params),
};
