import { z } from 'zod';
import {
    RPCMethod,
    GetHomeChannelResponseParams,
    GetEscrowChannelResponseParams,
    GetChannelsResponseParams,
    GetLatestStateResponseParams,
    GetStatesResponseParams,
    CreateChannelResponseParams,
    SubmitStateResponseParams,
    ChannelUpdateResponseParams,
    ChannelsUpdateResponseParams,
} from '../types';
import { hexSchema, channelSchema, stateSchema, paginationMetadataSchema, ParamsParser } from './common';

// get_home_channel response parser
const GetHomeChannelParamsSchema = z
    .object({
        channel: channelSchema,
    })
    .transform(
        (raw): GetHomeChannelResponseParams => ({
            channel: raw.channel,
        }),
    );

// get_escrow_channel response parser
const GetEscrowChannelParamsSchema = z
    .object({
        channel: channelSchema,
    })
    .transform(
        (raw): GetEscrowChannelResponseParams => ({
            channel: raw.channel,
        }),
    );

// get_channels response parser
const GetChannelsParamsSchema = z
    .object({
        channels: z.array(channelSchema),
        metadata: paginationMetadataSchema.optional(),
    })
    .transform(
        (raw): GetChannelsResponseParams => ({
            channels: raw.channels,
            metadata: raw.metadata,
        }),
    );

// get_latest_state response parser
const GetLatestStateParamsSchema = z
    .object({
        state: stateSchema,
    })
    .transform(
        (raw): GetLatestStateResponseParams => ({
            state: raw.state,
        }),
    );

// get_states response parser
const GetStatesParamsSchema = z
    .object({
        states: z.array(stateSchema),
        metadata: paginationMetadataSchema.optional(),
    })
    .transform(
        (raw): GetStatesResponseParams => ({
            states: raw.states,
            metadata: raw.metadata,
        }),
    );

// request_creation response parser (CreateChannel method)
const CreateChannelParamsSchema = z
    .object({
        signature: hexSchema,
    })
    .transform(
        (raw): CreateChannelResponseParams => ({
            signature: raw.signature,
        }),
    );

// submit_state response parser
const SubmitStateParamsSchema = z
    .object({
        signature: hexSchema,
    })
    .transform(
        (raw): SubmitStateResponseParams => ({
            signature: raw.signature,
        }),
    );

// channel_update event parser (server push) - note: params is RPCChannel directly, not wrapped
const ChannelUpdateParamsSchema = channelSchema.transform((raw): ChannelUpdateResponseParams => raw);

// channels_update event parser (server push)
const ChannelsUpdateParamsSchema = z
    .object({
        channels: z.array(channelSchema),
    })
    .transform(
        (raw): ChannelsUpdateResponseParams => ({
            channels: raw.channels,
        }),
    );

export const channelParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.GetHomeChannel]: (params) => GetHomeChannelParamsSchema.parse(params),
    [RPCMethod.GetEscrowChannel]: (params) => GetEscrowChannelParamsSchema.parse(params),
    [RPCMethod.GetChannels]: (params) => GetChannelsParamsSchema.parse(params),
    [RPCMethod.GetLatestState]: (params) => GetLatestStateParamsSchema.parse(params),
    [RPCMethod.GetStates]: (params) => GetStatesParamsSchema.parse(params),
    [RPCMethod.CreateChannel]: (params) => CreateChannelParamsSchema.parse(params),
    [RPCMethod.SubmitState]: (params) => SubmitStateParamsSchema.parse(params),
    [RPCMethod.ChannelUpdate]: (params) => ChannelUpdateParamsSchema.parse(params),
    [RPCMethod.ChannelsUpdate]: (params) => ChannelsUpdateParamsSchema.parse(params),
};
