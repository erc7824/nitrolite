import { z } from 'zod';
import { Address, Hex } from 'viem';
import {
    RPCMethod,
    CreateAppSessionResponseParams,
    SubmitAppStateResponseParams,
    SubmitDepositStateResponseParams,
    RebalanceAppSessionsResponseParams,
    GetAppDefinitionResponseParams,
    GetAppSessionsResponseParams,
    AppSessionUpdateResponseParams,
    RPCAppSession,
    RPCAppParticipant,
    RPCAppSessionAllocation,
} from '../types';
import { hexSchema, addressSchema, ParamsParser } from './common';

// App participant schema
const AppParticipantSchema = z
    .object({
        wallet_address: addressSchema,
        signature_weight: z.number(),
    })
    .transform(
        (raw): RPCAppParticipant => ({
            walletAddress: raw.wallet_address,
            signatureWeight: raw.signature_weight,
        }),
    );

// App session allocation schema
const AppSessionAllocationSchema = z
    .object({
        asset: z.string(),
        amount: z.string(),
        participant: addressSchema,
    })
    .transform(
        (raw): RPCAppSessionAllocation => ({
            asset: raw.asset,
            amount: raw.amount,
            participant: raw.participant,
        }),
    );

// App session schema
const AppSessionSchema = z
    .object({
        app_session_id: z.string(),
        status: z.string(),
        participants: z.array(AppParticipantSchema),
        session_data: z.string().optional(),
        quorum: z.number(),
        version: z.number(),
        nonce: z.number(),
        allocations: z.array(AppSessionAllocationSchema),
    })
    .transform(
        (raw): RPCAppSession => ({
            appSessionId: raw.app_session_id,
            status: raw.status,
            participants: raw.participants,
            sessionData: raw.session_data,
            quorum: raw.quorum,
            version: raw.version,
            nonce: raw.nonce,
            allocations: raw.allocations,
        }),
    );

// create_app_session response parser
const CreateAppSessionParamsSchema = z
    .object({
        app_session_id: hexSchema,
        version: z.number(),
        status: z.string(),
    })
    .transform(
        (raw): CreateAppSessionResponseParams => ({
            appSessionId: raw.app_session_id,
            version: raw.version,
            status: raw.status as any, // RPCChannelStatus enum
        }),
    );

// submit_app_state response parser
const SubmitAppStateParamsSchema = z
    .object({
        app_session_id: hexSchema,
        version: z.number(),
        status: z.string(),
    })
    .transform(
        (raw): SubmitAppStateResponseParams => ({
            appSessionId: raw.app_session_id,
            version: raw.version,
            status: raw.status as any, // RPCChannelStatus enum
        }),
    );

// submit_deposit_state response parser
const SubmitDepositStateParamsSchema = z
    .object({
        signature: hexSchema,
    })
    .transform(
        (raw): SubmitDepositStateResponseParams => ({
            signature: raw.signature,
        }),
    );

// rebalance_app_sessions response parser
const RebalanceAppSessionsParamsSchema = z
    .object({
        batch_id: z.string(),
    })
    .transform(
        (raw): RebalanceAppSessionsResponseParams => ({
            batchId: raw.batch_id,
        }),
    );

// get_app_definition response parser
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

// get_app_sessions response parser
const GetAppSessionsParamsSchema = z
    .object({
        app_sessions: z.array(AppSessionSchema),
    })
    .transform(
        (raw): GetAppSessionsResponseParams => ({
            appSessions: raw.app_sessions,
        }),
    );

// app_session_update event parser (server push)
const AppSessionUpdateParamsSchema = z
    .object({
        app_session: AppSessionSchema,
    })
    .transform(
        (raw): AppSessionUpdateResponseParams => ({
            appSession: raw.app_session,
        }),
    );

export const appParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.CreateAppSession]: (params) => CreateAppSessionParamsSchema.parse(params),
    [RPCMethod.SubmitAppState]: (params) => SubmitAppStateParamsSchema.parse(params),
    [RPCMethod.SubmitDepositState]: (params) => SubmitDepositStateParamsSchema.parse(params),
    [RPCMethod.RebalanceAppSessions]: (params) => RebalanceAppSessionsParamsSchema.parse(params),
    [RPCMethod.GetAppDefinition]: (params) => GetAppDefinitionParamsSchema.parse(params),
    [RPCMethod.GetAppSessions]: (params) => GetAppSessionsParamsSchema.parse(params),
    [RPCMethod.AppSessionUpdate]: (params) => AppSessionUpdateParamsSchema.parse(params),
};
