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

const AppAllocationObject = z.object({
    participant: addressSchema,
    asset: z.string(),
    amount: z.string(),
});

const AppSessionObject = z.object({
    app_session_id: hexSchema,
    application: z.string(),
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
            application: raw.application,
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
        protocol: z.string(),
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
        app_session: AppSessionObject,
        participant_allocations: z.array(AppAllocationObject),
    })
    .transform((raw) => ({
        appSessionId: raw.app_session.app_session_id,
        application: raw.app_session.application,
        status: raw.app_session.status,
        participants: raw.app_session.participants,
        sessionData: raw.app_session.session_data,
        protocol: raw.app_session.protocol,
        challenge: raw.app_session.challenge,
        weights: raw.app_session.weights,
        quorum: raw.app_session.quorum,
        version: raw.app_session.version,
        nonce: raw.app_session.nonce,
        createdAt: raw.app_session.created_at,
        updatedAt: raw.app_session.updated_at,
        participantAllocations: raw.participant_allocations.map((a) => ({
                participant: a.participant,
                asset: a.asset,
                amount: a.amount,
            })),
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
