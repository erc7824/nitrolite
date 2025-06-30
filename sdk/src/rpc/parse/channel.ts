import { z } from 'zod';
import { Address, Hex } from 'viem';
import {
    RPCMethod,
    ResizeChannelResponseParams,
    CloseChannelResponseParams,
    GetChannelsResponseParams,
    ChannelUpdateResponseParams,
    RPCChannelStatus,
    ChannelUpdate,
} from '../types';
import { hexSchema, addressSchema, statusEnum, ParamsParser } from './common';

const RPCAllocationSchema = z.object({
    destination: addressSchema,
    token: addressSchema,
    amount: z.union([z.string(), z.number()]).transform((a) => BigInt(a)),
});

const ServerSignatureSchema = z.object({
    v: z.union([z.string(), z.number()]).transform((a) => Number(a)),
    // TODO: it should use hexScheme as provided, but for some reason R and S value
    // are hex strings inside escaped double quotes: '"0x1234"' instead of '0x1234'
    // r: hexSchema,
    // s: hexSchema,
    r: z.string(),
    s: z.string(),
});

const ResizeChannelParamsSchema = z
    .array(
        z
            .object({
                channel_id: hexSchema,
                state_data: hexSchema,
                intent: z.number(),
                version: z.number(),
                allocations: z.array(RPCAllocationSchema),
                state_hash: hexSchema,
                server_signature: ServerSignatureSchema,
            })
            .transform(
                (raw) =>
                    ({
                        channelId: raw.channel_id as Hex,
                        stateData: raw.state_data as Hex,
                        intent: raw.intent,
                        version: raw.version,
                        allocations: raw.allocations.map((a) => ({
                            destination: a.destination as Address,
                            token: a.token as Address,
                            amount: a.amount,
                        })),
                        stateHash: raw.state_hash as Hex,
                        serverSignature: {
                            v: +raw.server_signature.v,
                            r: raw.server_signature.r as Hex,
                            s: raw.server_signature.s as Hex,
                        },
                    }) as ResizeChannelResponseParams,
            ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const CloseChannelParamsSchema = z
    .array(
        z
            .object({
                channel_id: hexSchema,
                state_data: hexSchema,
                intent: z.number(),
                version: z.number(),
                allocations: z.array(RPCAllocationSchema),
                state_hash: hexSchema,
                server_signature: ServerSignatureSchema,
            })
            .transform(
                (raw) =>
                    ({
                        channelId: raw.channel_id as Hex,
                        stateData: raw.state_data as Hex,
                        intent: raw.intent,
                        version: raw.version,
                        allocations: raw.allocations.map((a) => ({
                            destination: a.destination as Address,
                            token: a.token as Address,
                            amount: a.amount,
                        })),
                        stateHash: raw.state_hash as Hex,
                        serverSignature: {
                            v: +raw.server_signature.v,
                            r: raw.server_signature.r as Hex,
                            s: raw.server_signature.s as Hex,
                        },
                    }) as CloseChannelResponseParams,
            ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const GetChannelsParamsSchema = z
    .array(
        z.array(
            z
                .object({
                    channel_id: hexSchema,
                    participant: addressSchema,
                    status: statusEnum,
                    token: addressSchema,
                    wallet: addressSchema,
                    amount: z.union([z.string(), z.number()]).transform((a) => BigInt(a)),
                    chain_id: z.number(),
                    adjudicator: addressSchema,
                    challenge: z.number(),
                    nonce: z.number(),
                    version: z.number(),
                    created_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
                    updated_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
                })
                .transform(
                    (c) =>
                        ({
                            channelId: c.channel_id as Hex,
                            participant: c.participant as Address,
                            status: c.status as RPCChannelStatus,
                            token: c.token as Address,
                            wallet: c.wallet as Address,
                            amount: c.amount,
                            chainId: c.chain_id,
                            adjudicator: c.adjudicator as Address,
                            challenge: c.challenge,
                            nonce: c.nonce,
                            version: c.version,
                            createdAt: c.created_at,
                            updatedAt: c.updated_at,
                        }) as ChannelUpdate,
                ),
        ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0])
    .transform((arr) => arr as GetChannelsResponseParams);

const ChannelUpdateObjectSchema = z
    .object({
        channel_id: hexSchema,
        participant: addressSchema,
        status: statusEnum,
        token: addressSchema,
        wallet: z.union([addressSchema, z.literal('')]),
        amount: z.union([z.string(), z.number()]).transform((a) => BigInt(a)),
        chain_id: z.number(),
        adjudicator: addressSchema,
        challenge: z.number(),
        nonce: z.number(),
        version: z.number(),
        created_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
        updated_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
    })
    .transform(
        (c) =>
            ({
                channelId: c.channel_id as Hex,
                participant: c.participant as Address,
                status: c.status as RPCChannelStatus,
                token: c.token as Address,
                wallet: c.wallet as Address,
                amount: c.amount,
                chainId: c.chain_id,
                adjudicator: c.adjudicator as Address,
                challenge: c.challenge,
                nonce: c.nonce,
                version: c.version,
                createdAt: c.created_at,
                updatedAt: c.updated_at,
            }) as ChannelUpdateResponseParams,
    );

const ChannelUpdateParamsSchema = z
    .array(ChannelUpdateObjectSchema)
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const ChannelsUpdateParamsSchema = z
    .array(z.array(ChannelUpdateObjectSchema))
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

export const channelParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.ResizeChannel]: (params) => ResizeChannelParamsSchema.parse(params),
    [RPCMethod.CloseChannel]: (params) => CloseChannelParamsSchema.parse(params),
    [RPCMethod.GetChannels]: (params) => GetChannelsParamsSchema.parse(params),
    [RPCMethod.ChannelUpdate]: (params) => ChannelUpdateParamsSchema.parse(params),
    [RPCMethod.ChannelsUpdate]: (params) => ChannelsUpdateParamsSchema.parse(params),
};
