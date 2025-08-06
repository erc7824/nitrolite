import { z } from 'zod';
import {
    RPCMethod,
    ResizeChannelResponseParams,
    CloseChannelResponseParams,
    GetChannelsResponseParams,
    ChannelUpdateResponseParams,
    RPCChannelUpdate,
    ChannelsUpdateResponseParams,
    RPCChannelUpdateWithWallet,
    CreateChannelResponseParams,
    RPCChannelOperation,
} from '../types';
import { hexSchema, addressSchema, statusEnum, ParamsParser, bigIntSchema, dateSchema } from './common';

const RPCAllocationSchema = z.object({
    destination: addressSchema,
    token: addressSchema,
    amount: bigIntSchema,
});

const ChannelOperationObject = z.object({
    channel_id: hexSchema,
    state: z.object({
        intent: z.number(),
        version: z.number(),
        state_data: hexSchema,
        allocations: z.array(RPCAllocationSchema),
    }),
    server_signature: hexSchema,
});

const ChannelOperationObjectSchema = ChannelOperationObject.transform(
    (raw): RPCChannelOperation => ({
        channelId: raw.channel_id,
        state: {
            intent: raw.state.intent,
            version: raw.state.version,
            stateData: raw.state.state_data,
            allocations: raw.state.allocations.map((a) => ({
                destination: a.destination,
                token: a.token,
                amount: BigInt(a.amount),
            })),
        },
        serverSignature: raw.server_signature,
    }),
);

const CreateChannelParamsSchema = z
    .object({
        ...ChannelOperationObject.shape,
        channel: z.object({
            participants: z.array(addressSchema),
            adjudicator: addressSchema,
            challenge: z.number(),
            nonce: z.number(),
        }),
    })
    .transform(
        (raw): CreateChannelResponseParams => ({
            ...ChannelOperationObjectSchema.parse(raw),
            channel: {
                participants: raw.channel.participants,
                adjudicator: raw.channel.adjudicator,
                challenge: raw.channel.challenge,
                nonce: raw.channel.nonce,
            },
        }),
    );

const ResizeChannelParamsSchema = ChannelOperationObjectSchema
    // Validate received type with linter
    .transform((raw): ResizeChannelResponseParams => raw);

const CloseChannelParamsSchema = ChannelOperationObjectSchema
    // Validate received type with linter
    .transform((raw): CloseChannelResponseParams => raw);

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
    (raw): RPCChannelUpdate => ({
        channelId: raw.channel_id,
        participant: raw.participant,
        status: raw.status,
        token: raw.token,
        amount: BigInt(raw.amount),
        chainId: raw.chain_id,
        adjudicator: raw.adjudicator,
        challenge: raw.challenge,
        nonce: raw.nonce,
        version: raw.version,
        createdAt: raw.created_at,
        updatedAt: raw.updated_at,
    }),
);

const ChannelUpdateWithWalletObjectSchema = z
    .object({
        ...ChannelUpdateObject.shape,
        wallet: addressSchema,
    })
    .transform(
        (raw): RPCChannelUpdateWithWallet => ({
            ...ChannelUpdateObjectSchema.parse(raw),
            wallet: raw.wallet,
        }),
    );

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
    [RPCMethod.CreateChannel]: (params) => CreateChannelParamsSchema.parse(params),
    [RPCMethod.ResizeChannel]: (params) => ResizeChannelParamsSchema.parse(params),
    [RPCMethod.CloseChannel]: (params) => CloseChannelParamsSchema.parse(params),
    [RPCMethod.GetChannels]: (params) => GetChannelsParamsSchema.parse(params),
    [RPCMethod.ChannelUpdate]: (params) => ChannelUpdateParamsSchema.parse(params),
    [RPCMethod.ChannelsUpdate]: (params) => ChannelsUpdateParamsSchema.parse(params),
};
