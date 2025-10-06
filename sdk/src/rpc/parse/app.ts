import { z } from 'zod';
import { Address } from 'viem';
import {
    RPCMethod,
    CreateAppSessionResponseParams,
    SubmitAppStateResponseParams,
    CloseAppSessionResponseParams,
    GetAppDefinitionResponseParams,
    GetAppSessionsResponseParams,
    RPCAppSession,
} from '../types';
import { hexSchema, addressSchema, statusEnum, ParamsParser, dateSchema, protocolVersionEnum } from './common';

const AppSessionObject = z.object({
    app_session_id: hexSchema,
    status: statusEnum,
    participants: z.array(addressSchema),
    protocol: protocolVersionEnum,
    challenge: z.number(),
    weights: z.array(z.number()),
    quorum: z.number(),
    version: z.number(),
    nonce: z.number(),
    created_at: dateSchema,
    updated_at: dateSchema,
    session_data: z.string().optional(),
});

const AppSessionObjectSchema = AppSessionObject
    .transform(
        (raw): RPCAppSession => ({
            appSessionId: raw.app_session_id,
            status: raw.status,
            participants: raw.participants,
            protocol: raw.protocol,
            challenge: raw.challenge,
            weights: raw.weights,
            quorum: raw.quorum,
            version: raw.version,
            nonce: raw.nonce,
            createdAt: raw.created_at,
            updatedAt: raw.updated_at,
            sessionData: raw.session_data,
        }),
    );

const CreateAppSessionParamsSchema = z
    .object({ app_session_id: hexSchema, version: z.number(), status: statusEnum })
    .transform(
        (raw): CreateAppSessionResponseParams => ({
            appSessionId: raw.app_session_id,
            version: raw.version,
            status: raw.status,
        }),
    );

const SubmitAppStateParamsSchema = z
    .object({ app_session_id: hexSchema, version: z.number(), status: statusEnum })
    .transform(
        (raw): SubmitAppStateResponseParams => ({
            appSessionId: raw.app_session_id,
            version: raw.version,
            status: raw.status,
        }),
    );

const CloseAppSessionParamsSchema = z
    .object({ app_session_id: hexSchema, version: z.number(), status: statusEnum })
    .transform(
        (raw): CloseAppSessionResponseParams => ({
            appSessionId: raw.app_session_id,
            version: raw.version,
            status: raw.status,
        }),
    );

const GetAppDefinitionParamsSchema = z
    .object({
        protocol: protocolVersionEnum,
        participants: z.array(addressSchema),
        weights: z.array(z.number()),
        quorum: z.number(),
        challenge: z.number(),
        nonce: z.number(),
    })
    .transform(
        (raw): GetAppDefinitionResponseParams => ({
            protocol: raw.protocol,
            participants: raw.participants as Address[],
            weights: raw.weights,
            quorum: raw.quorum,
            challenge: raw.challenge,
            nonce: raw.nonce,
        }),
    );

const GetAppSessionsParamsSchema = z
    .object({
        app_sessions: z.array(AppSessionObjectSchema),
    })
    .transform(
        (raw): GetAppSessionsResponseParams => ({
            appSessions: raw.app_sessions,
        }),
    );

const AppSessionUpdateObjectSchema = z
    .object({
        ...AppSessionObject.shape,
        participant_allocations: z.record(z.string(), z.record(z.string(), z.string())), // participant -> asset -> amount
    })
    .transform((raw) => ({
        appSessionId: raw.app_session_id,
        status: raw.status,
        participants: raw.participants,
        sessionData: raw.session_data,
        protocol: raw.protocol,
        challenge: raw.challenge,
        weights: raw.weights,
        quorum: raw.quorum,
        version: raw.version,
        nonce: raw.nonce,
        createdAt: raw.created_at,
        updatedAt: raw.updated_at,
        participantAllocations: raw.participant_allocations,
    }));

const AppSessionUpdateParamsSchema = AppSessionUpdateObjectSchema;

export const appParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.CreateAppSession]: (params) => CreateAppSessionParamsSchema.parse(params),
    [RPCMethod.SubmitAppState]: (params) => SubmitAppStateParamsSchema.parse(params),
    [RPCMethod.CloseAppSession]: (params) => CloseAppSessionParamsSchema.parse(params),
    [RPCMethod.GetAppDefinition]: (params) => GetAppDefinitionParamsSchema.parse(params),
    [RPCMethod.GetAppSessions]: (params) => GetAppSessionsParamsSchema.parse(params),
    [RPCMethod.AppSessionUpdate]: (params) => AppSessionUpdateParamsSchema.parse(params),
};
