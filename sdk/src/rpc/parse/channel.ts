import { z } from 'zod';
import {
    RPCMethod,
    ResizeChannelResponseParams,
    CloseChannelResponseParams,
    GetChannelsResponseParams,
    ChannelUpdateResponseParams,
    RPCChannel,
    ChannelsUpdateResponseParams,
    RPCChannelWithWallet,
} from '../types';
import { hexSchema, addressSchema, statusEnum, ParamsParser, bigIntSchema, dateSchema } from './common';

const RPCAllocationSchema = z.object({
    destination: addressSchema,
    token: addressSchema,
    amount: bigIntSchema,
});

const ResizeChannelParamsSchema = z
    .object({
        channel_id: hexSchema,
        state_data: hexSchema,
        intent: z.number(),
        version: z.number(),
        allocations: z.array(RPCAllocationSchema),
        server_signature: hexSchema,
    })
    .transform(
        (raw): ResizeChannelResponseParams => ({
            channelId: raw.channel_id,
            stateData: raw.state_data,
            intent: raw.intent,
            version: raw.version,
            allocations: raw.allocations,
            serverSignature: raw.server_signature,
        }),
    );

const CloseChannelParamsSchema = z
    .object({
        channel_id: hexSchema,
        state_data: hexSchema,
        intent: z.number(),
        version: z.number(),
        allocations: z.array(RPCAllocationSchema),
        server_signature: hexSchema,
    })
    .transform(
        (raw): CloseChannelResponseParams => ({
            channelId: raw.channel_id,
            stateData: raw.state_data,
            intent: raw.intent,
            version: raw.version,
            allocations: raw.allocations,
            serverSignature: raw.server_signature,
        }),
    );

const ChannelUpdateObject = z.object({
    channel_id: hexSchema,
    participant: addressSchema,
    status: statusEnum,
    token: addressSchema,
    amount: bigIntSchema,
    chain_id: z.number(),
    adjudicator: addressSchema,
    challenge: z.number(),
    nonce: z.number(),
    version: z.number(),
    created_at: dateSchema,
    updated_at: dateSchema,
});

const ChannelUpdateObjectSchema = ChannelUpdateObject.transform(
    (raw): RPCChannel => ({
        channelId: raw.channel_id,
        participant: raw.participant,
        status: raw.status,
        token: raw.token,
        amount: raw.amount,
        chainId: raw.chain_id,
        adjudicator: raw.adjudicator,
        challenge: raw.challenge,
        nonce: raw.nonce,
        version: raw.version,
        createdAt: raw.created_at,
        updatedAt: raw.updated_at,
    }),
);

const ChannelUpdateWithWalletObjectSchema = z.object({
    ...ChannelUpdateObject.shape,
    wallet: addressSchema,
}).transform((raw): RPCChannelWithWallet => ({
    ...ChannelUpdateObjectSchema.parse(raw),
    wallet: raw.wallet,
}));

const GetChannelsParamsSchema = z
    .object({
        channels: z.array(ChannelUpdateWithWalletObjectSchema),
    })
    // Validate received type with linter
    .transform((raw): GetChannelsResponseParams => raw);

const ChannelUpdateParamsSchema = ChannelUpdateObjectSchema
    // Validate received type with linter
    .transform((raw): ChannelUpdateResponseParams => raw);

const ChannelsUpdateParamsSchema = z
    .object({
        channels: z.array(ChannelUpdateObjectSchema),
    })
    // Validate received type with linter
    .transform((raw): ChannelsUpdateResponseParams => raw);

export const channelParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.ResizeChannel]: (params) => ResizeChannelParamsSchema.parse(params),
    [RPCMethod.CloseChannel]: (params) => CloseChannelParamsSchema.parse(params),
    [RPCMethod.GetChannels]: (params) => GetChannelsParamsSchema.parse(params),
    [RPCMethod.ChannelUpdate]: (params) => ChannelUpdateParamsSchema.parse(params),
    [RPCMethod.ChannelsUpdate]: (params) => ChannelsUpdateParamsSchema.parse(params),
};
