import { z } from 'zod';
import { Address, Hex } from 'viem';
import {
    RPCMethod,
    ChannelOperationResponseParams,
    CreateChannelResponseParams,
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
    amount: z.string().transform((a) => BigInt(a)),
});

const ChannelOperationParamsSchema = z
    .array(
        z
            .object({
                channel_id: hexSchema,
                state: z.object({
                    intent: z.number(),
                    version: z.number(),
                    state_data: z.string().transform((data) => data as Hex),
                    allocations: z.array(RPCAllocationSchema),
                }),
                server_signature: hexSchema,
            })
            .transform(
                (raw) =>
                    ({
                        channelId: raw.channel_id as Hex,
                        state: {
                            intent: raw.state.intent,
                            version: raw.state.version,
                            stateData: raw.state.state_data,
                            allocations: raw.state.allocations.map((a) => ({
                                destination: a.destination as Address,
                                token: a.token as Address,
                                amount: a.amount,
                            })),
                        },
                        serverSignature: raw.server_signature,
                    }) as ChannelOperationResponseParams,
            ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const CreateChannelParamsSchema = ChannelOperationParamsSchema.transform(
    (params) => params as CreateChannelResponseParams,
);

const ResizeChannelParamsSchema = ChannelOperationParamsSchema.transform(
    (params) => params as ResizeChannelResponseParams,
);

const CloseChannelParamsSchema = ChannelOperationParamsSchema.transform(
    (params) => params as CloseChannelResponseParams,
);

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
        nonce: z.union([z.string(), z.number()]).transform((n) => BigInt(n)),
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

const GetChannelsParamsSchema = z
    .array(z.array(ChannelUpdateObjectSchema))
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0])
    .transform((arr) => arr as GetChannelsResponseParams);

const ChannelUpdateParamsSchema = z
    .array(ChannelUpdateObjectSchema)
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const ChannelsUpdateParamsSchema = z
    .array(z.array(ChannelUpdateObjectSchema))
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

export const channelParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.CreateChannel]: (params) => CreateChannelParamsSchema.parse(params),
    [RPCMethod.ResizeChannel]: (params) => ResizeChannelParamsSchema.parse(params),
    [RPCMethod.CloseChannel]: (params) => CloseChannelParamsSchema.parse(params),
    [RPCMethod.GetChannels]: (params) => GetChannelsParamsSchema.parse(params),
    [RPCMethod.ChannelUpdate]: (params) => ChannelUpdateParamsSchema.parse(params),
    [RPCMethod.ChannelsUpdate]: (params) => ChannelsUpdateParamsSchema.parse(params),
};
