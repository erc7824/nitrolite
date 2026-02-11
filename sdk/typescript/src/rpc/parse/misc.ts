import { z } from 'zod';
import { RPCMethod, GetConfigResponseParams, ErrorResponseParams, RPCNetworkInfo } from '../types';
import { addressSchema, ParamsParser } from './common';

const NetworkInfoSchema = z
    .object({
        name: z.string(),
        blockchain_id: z.number(),
        contract_address: addressSchema,
    })
    .transform(
        (raw): RPCNetworkInfo => ({
            name: raw.name,
            blockchainId: raw.blockchain_id,
            contractAddress: raw.contract_address,
        }),
    );

const GetConfigParamsSchema = z
    .object({
        broker_address: addressSchema,
        networks: z.array(NetworkInfoSchema),
    })
    .transform(
        (raw): GetConfigResponseParams => ({
            brokerAddress: raw.broker_address,
            networks: raw.networks,
        }),
    );

const ErrorParamsSchema = z
    .object({
        error: z.string(),
    })
    .transform(
        (raw): ErrorResponseParams => ({
            error: raw.error,
        }),
    );

const parseMessageParams: ParamsParser<unknown> = (params) => {
    return params;
};

export const miscParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.GetConfig]: (params) => GetConfigParamsSchema.parse(params),
    [RPCMethod.Error]: (params) => ErrorParamsSchema.parse(params),
    [RPCMethod.Message]: parseMessageParams,
};
